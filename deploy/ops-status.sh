#!/bin/bash
# 运维状态快照：每 10 分钟由 cron 生成一份 JSON，供 .github/workflows/healthcheck.yml 探活。
#
# 为什么需要它（2026-09-17）：定时备份因 deploy/db-backup.sh 丢失可执行位 + 被 CRLF 化，
# 从 9/16 起静默失败两天，cron 只在日志里留了两行 Permission denied，没有任何人知道。
# 原有的探活只看 API /healthz 与后台首页 —— 进程活着，但备份死了、磁盘要满了、证书要过期了，
# 它一概看不见。此脚本把这些指标暴露成一个可被外部读取的快照。
#
# 只输出**粗粒度**的运维指标（百分比/小时数/状态名），不含任何业务数据或凭据。
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
latest=$(ls -1t "$BACKUP_DIR"/uav-db-*.sql.gz 2>/dev/null | head -1)
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

# ---- 数据库 schema 版本（失败不影响整体判定，仅留空）----
schema=$(docker exec uav-db-1 psql -U drone -d drone_platform -t -A \
  -c 'SELECT max(version) FROM schema_migrations' 2>/dev/null | tr -d '[:space:]')
schema=${schema:-unknown}

# ---- 汇总 ----
ok=false
if [ "$disk_ok" = true ] && [ "$backup_ok" = true ] && [ "$containers_ok" = true ] && [ "$cert_ok" = true ] && [ "$drill_ok" = true ] && [ "$escrow_ok" = true ]; then
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
  "schema_version": "$schema"
}
JSON
chmod 644 "$tmp"
mv -f "$tmp" "$OUT"
