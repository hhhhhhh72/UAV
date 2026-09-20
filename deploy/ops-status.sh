#!/bin/bash
# 运维状态快照：每 10 分钟由 cron 生成一份 JSON，供 .github/workflows/healthcheck.yml 探活。
#
# 为什么需要它（2026-09-17）：定时备份因 deploy/db-backup.sh 丢失可执行位 + 被 CRLF 化，
# 从 9/16 起静默失败两天，cron 只在日志里留了两行 Permission denied，没有任何人知道。
# 原有的探活只看 API /healthz 与后台首页 —— 进程活着，但备份死了、磁盘要满了、证书要过期了，
# 它一概看不见。此脚本把这些指标暴露成一个可被外部读取的快照。
#
# 只输出**粗粒度**的运维指标（百分比/小时数/状态名），不含任何业务数据或凭据。
#
# 检查分两类，缺一不可：
#   - **结果类**：备份文件在不在、新不新鲜、能不能还原；磁盘够不够；证书还有几天
#   - **过程类**（jobs 一节）：那些定时任务**本身**有没有在跑 —— 结果类检查看不出
#     "任务从没跑过"，因为根本没有结果可看
set -uo pipefail

OUT=${OUT:-/var/www/ops-status.json}
BACKUP_DIR=${BACKUP_DIR:-$HOME/UAV-db-backups}
CERT=${CERT:-/etc/nginx/certs/api.cqnarc.cn.fullchain.crt}

# 阈值集中在此，探活侧只读 ok 标志，避免两边各写一套判断而漂移。
DISK_MAX_PERCENT=${DISK_MAX_PERCENT:-85}
BACKUP_MAX_AGE_HOURS=${BACKUP_MAX_AGE_HOURS:-30}
CERT_MIN_DAYS=${CERT_MIN_DAYS:-21}
DRILL_MAX_AGE_DAYS=${DRILL_MAX_AGE_DAYS:-8}

now_epoch=$(date +%s)
generated_at=$(date -Iseconds)

# ---- 磁盘 ----
disk_used=$(df -P / | awk 'NR==2 {gsub(/%/,"",$5); print $5}')
disk_used=${disk_used:-0}
disk_ok=false; [ "$disk_used" -lt "$DISK_MAX_PERCENT" ] && disk_ok=true

# ---- 备份新鲜度 + 完整性 ----
#
# 取「最近一次**完成**的备份」，而不是「目录里最新的文件」。
#
# 为什么（2026-09-20 03:02 的真实误报，而且是**每天必发**）：
#   crontab 里 `0 3 * * *` 的备份排在 `*/10 * * * *` 的快照前面，同一分钟两者同秒启动；
#   备份要边写边落盘（约 1 秒），而快照这一秒里 `ls -1t | head -1` 正好拿到**还没写完**
#   的那个文件，`gzip -t` 必然失败 → 判 corrupt → 03:02 告警；10 分钟后再跑文件已完整
#   → 03:12 恢复。09-19、09-20 连着两天各发一对「异常 + 恢复」，真正的故障会被这种
#   噪音淹掉。backup.log 里的 `OK <文件名>` 只在备份**完整落盘且非空**之后才写，
#   用它当完成标志最可靠。
BACKUP_LOG=${BACKUP_LOG:-$BACKUP_DIR/backup.log}
BACKUP_SETTLE_SECONDS=${BACKUP_SETTLE_SECONDS:-180}
latest=""
if [ -f "$BACKUP_LOG" ]; then
  okname=$(grep -oE 'OK[[:space:]]+uav-db-[0-9]{8}-[0-9]{6}\.sql\.gz' "$BACKUP_LOG" 2>/dev/null | tail -1 | awk '{print $2}')
  if [ -n "$okname" ] && [ -f "$BACKUP_DIR/$okname" ]; then latest="$BACKUP_DIR/$okname"; fi
fi
# 兜底（没有 backup.log 的老环境）：仍按时间取最新，但跳过可能仍在写入的文件
if [ -z "$latest" ]; then
  while IFS= read -r f; do
    [ -z "$f" ] && continue
    f_age=$(( now_epoch - $(stat -c %Y "$f" 2>/dev/null || echo 0) ))
    if [ "$f_age" -ge "$BACKUP_SETTLE_SECONDS" ]; then latest="$f"; break; fi
  done < <(ls -1t "$BACKUP_DIR"/uav-db-*.sql.gz 2>/dev/null)
fi
backup_file=""; backup_age_hours=-1; backup_size=0; backup_integrity=missing; backup_ok=false
if [ -n "$latest" ]; then
  backup_file=$(basename "$latest")
  backup_size=$(stat -c %s "$latest" 2>/dev/null || echo 0)
  age_sec=$(( now_epoch - $(stat -c %Y "$latest" 2>/dev/null || echo "$now_epoch") ))
  backup_age_hours=$(awk -v s="$age_sec" 'BEGIN{printf "%.1f", s/3600}')
  # 完整性：能通过 gzip 校验才算一份可用的备份（防 pg_dump 半途失败留下截断文件）
  if gzip -t "$latest" 2>/dev/null; then backup_integrity=ok; else backup_integrity=corrupt; fi
  if [ "$backup_integrity" = ok ] && awk -v h="$backup_age_hours" -v m="$BACKUP_MAX_AGE_HOURS" 'BEGIN{exit !(h < m)}'; then
    backup_ok=true
  fi
fi

# ---- 恢复演练结果（deploy/restore-drill.sh 每周写入）----
# 「文件存在」不等于「能恢复」：pg_dump 半途失败、磁盘写满、gzip 截断都会留下
# 一份看起来正常却还原不出来的备份。演练结果不达标同样要报警。
DRILL=${DRILL:-$BACKUP_DIR/restore-drill.json}
drill_ok=false; drill_age_days=-1; drill_note=missing; drill_tables=-1
if [ -f "$DRILL" ]; then
  drill_epoch=$(sed -n 's/.*"epoch": *\([0-9]*\).*/\1/p' "$DRILL" | head -1)
  drill_tables=$(sed -n 's/.*"tables": *\(-*[0-9]*\).*/\1/p' "$DRILL" | head -1)
  drill_note=$(sed -n 's/.*"note": *"\([^"]*\)".*/\1/p' "$DRILL" | head -1)
  grep -q '"ok": true' "$DRILL" && drill_ok=true
  if [ -n "$drill_epoch" ]; then
    drill_age_days=$(awk -v s="$(( now_epoch - drill_epoch ))" 'BEGIN{printf "%.1f", s/86400}')
    # 演练是周任务：超过 $DRILL_MAX_AGE_DAYS 天没有成功记录即视为异常
    awk -v d="$drill_age_days" -v m="$DRILL_MAX_AGE_DAYS" 'BEGIN{exit !(d < m)}' || drill_ok=false
  else
    drill_ok=false
  fi
fi
drill_tables=${drill_tables:--1}
drill_note=${drill_note:-missing}

# ---- 容器 ----
api_state=$(docker inspect -f '{{.State.Status}}' uav-api-1 2>/dev/null || echo missing)
db_state=$(docker inspect -f '{{.State.Status}}' uav-db-1 2>/dev/null || echo missing)
containers_ok=false
[ "$api_state" = running ] && [ "$db_state" = running ] && containers_ok=true

# ---- 证书 ----
cert_days=-1; cert_ok=false
if [ -f "$CERT" ]; then
  end=$(openssl x509 -in "$CERT" -noout -enddate 2>/dev/null | cut -d= -f2)
  if [ -n "$end" ]; then
    cert_days=$(( ( $(date -d "$end" +%s) - now_epoch ) / 86400 ))
    [ "$cert_days" -gt "$CERT_MIN_DAYS" ] && cert_ok=true
  fi
fi

# ---- 托管金（资金）不变量 ----
#
# 为什么需要（2026-09-18）：资金侧的三条不变量此前只靠应用层的 check-then-act，
# 库层面没有约束；补上唯一索引（migration 000116/000117）之后，还得有人**定期确认
# 它们还在**——索引可能被误删、数据可能被手工改动、将来可能有人新增一种资金动作
# 让下面的账目恒等式失效。这一节就是那个"定期确认"。
#
# 只输出计数与状态名，不含任何用户 ID 或金额明细（与文件顶部约定一致）。
# ESCROW_DB 可覆盖：只为让这套检查**本身**能被演练/变异测试（在临时库上造出
# 索引缺失、账目不平、未知资金动作三种故障，确认它真的会报警）。
ESCROW_DB=${ESCROW_DB:-drone_platform}
esc_q() { docker exec uav-db-1 psql -U drone -d "$ESCROW_DB" -t -A -c "$1" 2>/dev/null | tr -d '[:space:]'; }

# 1) 两条防重复资金的唯一索引必须都在
esc_idx=$(esc_q "SELECT count(*) FROM pg_indexes WHERE indexname IN ('idx_escrow_once_per_ref','idx_escrow_refund_once_per_order')")
esc_idx=${esc_idx:-0}

# 2) 不允许存在重复的放款/转账键，也不允许 trade_order 维度的重复退款键
esc_dup=$(esc_q "SELECT count(*) FROM (
  SELECT 1 FROM escrow_transactions
  WHERE reference_id <> '' AND (
    tx_type IN ('release','transfer','withdraw')
    OR (tx_type = 'refund' AND reference_type = 'trade_order'))
  GROUP BY from_user, tx_type, reference_type, reference_id HAVING count(*) > 1) t")
esc_dup=${esc_dup:-0}

# 3) 账目恒等式：每个托管账户的余额与冻结额，必须等于流水推算出来的值。
#    推导依据（按每笔流水的真实资金方向）：
#      deposit  → to_user.balance   += 额
#      freeze   → from_user.balance -= 额, from_user.frozen += 额
#      release  → from_user.frozen  -= 额, to_user.balance   += 额
#      refund   → to_user.frozen    -= 额, to_user.balance   += 额   ← 动的是 to_user
#      transfer → from_user.balance -= 额, to_user.balance   += 额
#      withdraw → from_user.frozen  -= 额（钱离开平台，两边余额都不动）
#    该式已用生产数据验证过与实际完全一致（不是恒真的空检查）。
esc_mismatch=$(esc_q "
WITH ledger AS (
  SELECT a.user_id AS uid,
    COALESCE(SUM(CASE WHEN t.to_user = a.user_id AND t.tx_type IN ('deposit','release','refund','transfer') THEN t.amount_fen ELSE 0 END), 0)
  - COALESCE(SUM(CASE WHEN t.from_user = a.user_id AND t.tx_type IN ('freeze','transfer') THEN t.amount_fen ELSE 0 END), 0) AS exp_balance,
    COALESCE(SUM(CASE WHEN t.from_user = a.user_id AND t.tx_type = 'freeze' THEN t.amount_fen ELSE 0 END), 0)
  - COALESCE(SUM(CASE WHEN t.from_user = a.user_id AND t.tx_type = 'release' THEN t.amount_fen ELSE 0 END), 0)
  - COALESCE(SUM(CASE WHEN t.to_user = a.user_id AND t.tx_type = 'refund' THEN t.amount_fen ELSE 0 END), 0)
  - COALESCE(SUM(CASE WHEN t.from_user = a.user_id AND t.tx_type = 'withdraw' THEN t.amount_fen ELSE 0 END), 0) AS exp_frozen
  FROM escrow_accounts a
  LEFT JOIN escrow_transactions t ON (t.to_user = a.user_id OR t.from_user = a.user_id)
  GROUP BY a.user_id)
SELECT count(*) FROM escrow_accounts a JOIN ledger l ON l.uid = a.user_id
WHERE a.balance_fen <> l.exp_balance OR a.frozen_fen <> l.exp_frozen")
esc_mismatch=${esc_mismatch:-0}

# 4) 出现未知的资金动作类型时恒等式不再成立：必须报出来让人复核公式，
#    而不是让它悄悄算出一堆假偏差（这是"检查检查器本身"的信号）。
esc_unknown=$(esc_q "SELECT count(*) FROM escrow_transactions WHERE tx_type NOT IN ('deposit','freeze','release','refund','transfer','withdraw')")
esc_unknown=${esc_unknown:-0}

esc_accounts=$(esc_q "SELECT count(*) FROM escrow_accounts"); esc_accounts=${esc_accounts:-0}
esc_frozen=$(esc_q "SELECT COALESCE(sum(frozen_fen),0) FROM escrow_accounts"); esc_frozen=${esc_frozen:-0}

escrow_ok=false
if [ "$esc_idx" = 2 ] && [ "$esc_dup" = 0 ] && [ "$esc_mismatch" = 0 ] && [ "$esc_unknown" = 0 ]; then
  escrow_ok=true
fi

# ---- 定时任务心跳（"设置好了但没人确认它真的在跑"）----
#
# 与上面几节的区别：那些看的是**结果**（备份文件在不在、新不新鲜），这一节看的是
# **任务本身有没有在跑**。9/17 装好的 disk-hygiene 是周任务，到 9/18 一次都没跑过 ——
# 结果类检查完全看不出来，因为根本没有结果可看。依据是日志/心跳文件的 mtime：
# 任务每跑一次就该前移一次。
# 只给**没有结果产物**的任务做心跳：
#   - disk-hygiene：跑完只写日志，没有可检查的产物
#   - alert：健康时完全静默，更是什么都不留（另盖一个心跳文件）
#   - 工作日报：有产物，但检查的是**送达**心跳（post-work-report.sh 只在推送成功后
#     覆盖写它）——比「生成了没有」更强，所以同样归在这里。
# 有结果产物的任务不加：备份看 backup.file 的年龄、还原演练看 drill 的年龄、
# 快照自己看 generated_epoch —— 那些检查更强（证明**结果产出了**，不只是脚本跑了）。
# 而且**日志年龄对它们不可靠**：db-backup.sh 成功时不产出任何 stdout/stderr，
# 而 `>>` 打开文件不写入不会更新 mtime，于是日志停在几天前、看起来像"没跑"。
# （这条假阳性是变异测试时抓到的：cron.log 停在 9/17，但当天 03:00 的备份产物明明在。）
JOB_HYGIENE_LOG=${JOB_HYGIENE_LOG:-$HOME/UAV-db-backups/disk-hygiene.log}
JOB_ALERT_HB=${JOB_ALERT_HB:-$HOME/UAV-db-backups/.alert-heartbeat}
JOB_WORK_REPORT_HB=${JOB_WORK_REPORT_HB:-$HOME/UAV-db-backups/.work-report-heartbeat}
HYGIENE_MAX_HOURS=${HYGIENE_MAX_HOURS:-26}
ALERT_MAX_MINUTES=${ALERT_MAX_MINUTES:-60}
# 工作日报是**日**任务（本机 17:30 生成 → scp 到服务器 → 由 post-work-report.sh 发出）：
# 两次之间正好隔 24 小时，阈值必须留余量，否则每天都会在「上一次跑完」到「下一次该跑」
# 之间出现一段假告警窗口。30 小时 = 24 小时 + 6 小时抖动余量。
WORK_REPORT_MAX_HOURS=${WORK_REPORT_MAX_HOURS:-30}

# age_hours_of 输出文件年龄（小时，一位小数）；不存在输出 -1。
age_hours_of() {
  if [ ! -f "$1" ]; then echo -1; return; fi
  awk -v s="$(( now_epoch - $(stat -c %Y "$1" 2>/dev/null || echo 0) ))" 'BEGIN{printf "%.1f", s/3600}'
}
within() { awk -v a="$1" -v m="$2" 'BEGIN{exit !(a >= 0 && a < m)}'; }

job_hygiene=$(age_hours_of "$JOB_HYGIENE_LOG")
job_alerthb=$(age_hours_of "$JOB_ALERT_HB")
alert_max_hours=$(awk -v m="$ALERT_MAX_MINUTES" 'BEGIN{printf "%.2f", m/60}')

# 工作日报看的是这个心跳文件，而它**只在推送成功后**才被覆盖写（见 post-work-report.sh）。
# 这是有意选更强的那一项：文件时间只能证明「脚本跑了」，而 webhook 被停用/换 key 时
# 脚本照样天天生成、文件天天新鲜，群里却一条都收不到 —— 那正是「设置好了但没人确认
# 它真的在跑」。只在**送达**时前移的时间戳才挡得住这种假绿。
job_workreport=$(age_hours_of "$JOB_WORK_REPORT_HB")

jobs_ok=false
if within "$job_hygiene" "$HYGIENE_MAX_HOURS" && within "$job_alerthb" "$alert_max_hours" && within "$job_workreport" "$WORK_REPORT_MAX_HOURS"; then
  jobs_ok=true
fi

# ============================================================
# instances —— 只允许一个 API 实例（2026-09-20 新增）
# ------------------------------------------------------------
# 全仓 8 处 check-then-act（课程/研学/赛事/活动报名、测试场地/场馆预约、职位投递）
# 用的都是**进程内**键锁（service/intent.go:48 lockByKey）。它们只在单实例下成立：
# 多一个 API 进程，锁就各锁各的，名额类业务会一起**静默超发**（研学实测容量 10
# 被 200 并发收下 87 人），而且没有任何机制会告诉你。
# 这一节不试图"证明只有一个" —— 它把那个没人验证过的前提变成可验证的绊线：
#   ① 跑着几个 uav-api 容器
#   ② 有几个不同的 client_addr 连到本库（应用连接走 TCP；运维脚本走 docker exec
#      的本地 socket → client_addr 为 NULL，不参与计数。生产实测单实例 = 1）
# 取不到数据即判失守：宁可报，不可假绿（与备份/告警那两次同一个口径）。
# ============================================================
INSTANCE_MAX=${INSTANCE_MAX:-1}
api_replicas=$(docker ps --filter 'ancestor=uav-api' --format '{{.Names}}' 2>/dev/null | wc -l | tr -d ' ')
db_clients=$(docker exec uav-db-1 psql -U drone -d "$ESCROW_DB" -t -A \
  -c "SELECT count(DISTINCT client_addr) FROM pg_stat_activity WHERE datname='$ESCROW_DB' AND client_addr IS NOT NULL" 2>/dev/null | tr -d '[:space:]')
api_replicas=${api_replicas:-0}
instances_note=""
instances_ok=false
if [ -z "$db_clients" ]; then
  instances_note="无法统计数据库连接来源（psql 查询失败）"
elif [ "$api_replicas" -le "$INSTANCE_MAX" ] && [ "$db_clients" -le "$INSTANCE_MAX" ]; then
  instances_ok=true
else
  instances_note="检测到多实例：api 容器 $api_replicas 个、数据库连接来源 $db_clients 个（上限 $INSTANCE_MAX）—— 进程内键锁失效，名额类业务会超发"
  echo "WARN: $instances_note" >&2
fi

# ---- 数据库 schema 版本（失败不影响整体判定，仅留空）----
schema=$(docker exec uav-db-1 psql -U drone -d drone_platform -t -A \
  -c 'SELECT max(version) FROM schema_migrations' 2>/dev/null | tr -d '[:space:]')
schema=${schema:-unknown}

# ---- 汇总 ----
ok=false
if [ "$disk_ok" = true ] && [ "$backup_ok" = true ] && [ "$containers_ok" = true ] && [ "$cert_ok" = true ] && [ "$drill_ok" = true ] && [ "$escrow_ok" = true ] && [ "$jobs_ok" = true ] && [ "$instances_ok" = true ]; then
  ok=true
fi

# 原子写：探活可能正好在写入过程中读取，绝不能读到半个 JSON。
tmp="$OUT.tmp.$$"
cat > "$tmp" <<JSON
{
  "generated_at": "$generated_at",
  "generated_epoch": $now_epoch,
  "ok": $ok,
  "disk": {"used_percent": $disk_used, "max_percent": $DISK_MAX_PERCENT, "ok": $disk_ok},
  "backup": {"file": "$backup_file", "age_hours": $backup_age_hours, "size_bytes": $backup_size, "integrity": "$backup_integrity", "max_age_hours": $BACKUP_MAX_AGE_HOURS, "ok": $backup_ok},
  "restore_drill": {"age_days": $drill_age_days, "max_age_days": $DRILL_MAX_AGE_DAYS, "tables": $drill_tables, "note": "$drill_note", "ok": $drill_ok},
  "containers": {"api": "$api_state", "db": "$db_state", "ok": $containers_ok},
  "cert": {"days_left": $cert_days, "min_days": $CERT_MIN_DAYS, "ok": $cert_ok},
  "escrow": {"indexes": $esc_idx, "dup_keys": $esc_dup, "ledger_mismatch": $esc_mismatch, "unknown_tx_types": $esc_unknown, "accounts": $esc_accounts, "frozen_fen": $esc_frozen, "ok": $escrow_ok},
  "jobs": {"hygiene_log_age_hours": $job_hygiene, "hygiene_max_hours": $HYGIENE_MAX_HOURS, "alert_heartbeat_age_hours": $job_alerthb, "alert_max_minutes": $ALERT_MAX_MINUTES, "work_report_age_hours": $job_workreport, "work_report_max_hours": $WORK_REPORT_MAX_HOURS, "ok": $jobs_ok},
  "instances": {"api_containers": $api_replicas, "db_client_addrs": ${db_clients:-0}, "max_allowed": $INSTANCE_MAX, "detail": "$instances_note", "ok": $instances_ok},
  "schema_version": "$schema"
}
JSON
chmod 644 "$tmp"
mv -f "$tmp" "$OUT"
