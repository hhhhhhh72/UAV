#!/bin/bash
# 平台经营日报：每天 19:05 把「今天平台发生了什么 + 系统好不好」推到群里。
#
# 为什么需要（2026-09-18）：平台上每天发生了什么，此前只能靠人登录后台翻页面 ——
# 而人不会天天翻。运维告警解决了「坏了要有人说」，但**没坏**的那些天里，
# 新注册了几个用户、有没有人发需求、托管里有多少钱，没有任何渠道会主动告诉负责人。
# 这份日报就是那个渠道：一条纯文本消息，全是数字，不点链接、不带附件。
#
# 覆盖窗口：默认「当日 00:00 → 运行时刻」；DAY=YYYY-MM-DD 可指定某一天（补发/演练）。
# 产物（都在 $DIR 下，0600 目录）：
#   daily-report.txt    人看的正文（也就是推进群里的那段）
#   daily-report.json   机器读的原始计数 + 投递时间，供以后做趋势
#   daily-report.log    运行日志（webhook key 一律打码）
#
# 关于「送达」而不是「生成」：
#   JSON 里的 delivered_epoch **只在推送成功后**才更新。ops-status.sh 检查的是它，
#   不是文件时间 —— 否则 webhook 被停用/换 key 时，脚本每天照常生成、文件天天新鲜，
#   群里却一条都收不到，正是「设置好了但没人确认它真的在跑」的第四次重演。
#
# 用法：
#   bash deploy/daily-report.sh                 # 生成并推送
#   bash deploy/daily-report.sh --dry-run       # 只生成不推送（也不更新 delivered_epoch）
#   DAY=2026-09-17 bash deploy/daily-report.sh --dry-run   # 补看某一天
#
# cron（服务器 root）：
#   5 19 * * * /root/UAV/deploy/daily-report.sh >> /root/UAV-db-backups/daily-report-cron.log 2>&1
set -uo pipefail

DIR=${DIR:-$HOME/UAV-db-backups}
OUT_TXT=${OUT_TXT:-$DIR/daily-report.txt}
OUT_JSON=${OUT_JSON:-$DIR/daily-report.json}
LOG=${LOG:-$DIR/daily-report.log}
ENVFILE=${ENVFILE:-/root/UAV/alert.env}
SNAPSHOT=${SNAPSHOT:-/var/www/ops-status.json}
CONTAINER=${CONTAINER:-uav-db-1}
DB_NAME=${DB_NAME:-drone_platform}
DAY=${DAY:-}
SEND=${SEND:-1}
VERBOSE=${VERBOSE:-0}

for a in "$@"; do
  case "$a" in
    --dry-run)    SEND=0; VERBOSE=1 ;;
    --send)       SEND=1 ;;
    -v|--verbose) VERBOSE=1 ;;
    *) echo "未知参数：$a（可用：--dry-run / --send / --verbose）" >&2; exit 2 ;;
  esac
done

# DAY 会被直接拼进 SQL，先卡死格式（只允许纯日期），杜绝注入。
case "$DAY" in
  '') ;;
  [0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9]) ;;
  *) echo "DAY 格式应为 YYYY-MM-DD，收到：$DAY" >&2; exit 2 ;;
esac

mkdir -p "$DIR"
# 通道配置与告警共用一份文件：同一把钥匙、同一套形状校验（见 lib-notify.sh）。
# DAILY_WEBHOOK 可单独指定（想把日报发到另一个群时用）。
[ -f "$ENVFILE" ] && . "$ENVFILE"
NOTIFY_WEBHOOK=${DAILY_WEBHOOK:-${ALERT_WEBHOOK:-}}
NOTIFY_FORMAT=${DAILY_FORMAT:-${ALERT_FORMAT:-wecom}}
# shellcheck source=deploy/lib-notify.sh
. "$(dirname "${BASH_SOURCE[0]}")/lib-notify.sh"

log() {
  local m="$(date -Iseconds) $*"
  echo "$m" >> "$LOG"
  [ "$VERBOSE" = 1 ] && echo "$m"
  return 0
}

esc_q() { docker exec "$CONTAINER" psql -U drone -d "$DB_NAME" -t -A -F'|' -c "$1" 2>/dev/null; }

# 时间窗在 bash 里选好再拼进 SQL，**不要**在 SQL 里写 CASE WHEN '$DAY' = '' THEN ... ELSE ''::date END：
# PostgreSQL 会在**生成计划时**折叠常量表达式，即使 CASE 分支根本不会走到，`''::date`
# 也会被求值并直接报 invalid input syntax for type date —— 表现是「指定 DAY 时正常、
# 不指定时（也就是每天真正跑的那条路径）psql 一行都不返回」。这个坑是干跑时抓到的。
if [ -n "$DAY" ]; then
  WIN0="'$DAY'::date"
  WIN1="'$DAY'::date + interval '1 day'"
else
  WIN0="date_trunc('day', now())"
  WIN1="now()"
fi

# 一次查询取全部指标：多次查询会在跨零点/并发写入时取到互相矛盾的数。
# 只取计数与合计金额，不含任何用户 ID、手机号、姓名。
METRIC_SQL=$(cat <<SQL
SELECT k, v FROM (
  SELECT 'new_users' k, count(*)::text v FROM users WHERE created_at >= $WIN0 AND created_at < $WIN1
  UNION ALL SELECT 'new_enterprises', count(*)::text FROM enterprises WHERE created_at >= $WIN0 AND created_at < $WIN1
  UNION ALL SELECT 'new_demands', count(*)::text FROM demands WHERE created_at >= $WIN0 AND created_at < $WIN1
  UNION ALL SELECT 'new_products', count(*)::text FROM drone_products WHERE created_at >= $WIN0 AND created_at < $WIN1
  UNION ALL SELECT 'new_orders', count(*)::text FROM trade_orders WHERE created_at >= $WIN0 AND created_at < $WIN1
  UNION ALL SELECT 'money_in_fen', COALESCE(sum(amount_fen),0)::text FROM escrow_transactions WHERE created_at >= $WIN0 AND created_at < $WIN1 AND tx_type='deposit'
  UNION ALL SELECT 'release_fen', COALESCE(sum(amount_fen),0)::text FROM escrow_transactions WHERE created_at >= $WIN0 AND created_at < $WIN1 AND tx_type='release'
  UNION ALL SELECT 'refund_fen', COALESCE(sum(amount_fen),0)::text FROM escrow_transactions WHERE created_at >= $WIN0 AND created_at < $WIN1 AND tx_type='refund'
  UNION ALL SELECT 'withdraw_fen', COALESCE(sum(amount_fen),0)::text FROM escrow_transactions WHERE created_at >= $WIN0 AND created_at < $WIN1 AND tx_type='withdraw'
  UNION ALL SELECT 'todo_enterprise', count(*)::text FROM enterprises WHERE status NOT IN ('approved','rejected') AND deleted_at IS NULL
  UNION ALL SELECT 'todo_intent', count(*)::text FROM demand_intents WHERE status='pending'
  UNION ALL SELECT 'todo_workorder', count(*)::text FROM work_orders WHERE status='pending'
  UNION ALL SELECT 'todo_aftersale', count(*)::text FROM trade_orders WHERE status='aftersale'
  UNION ALL SELECT 'todo_check', count(*)::text FROM drone_products WHERE check_status NOT IN ('passed','rejected')
  UNION ALL SELECT 'todo_report', count(*)::text FROM reports WHERE status='pending'
  UNION ALL SELECT 'tot_users', count(*)::text FROM users WHERE deleted_at IS NULL
  UNION ALL SELECT 'tot_enterprises', count(*)::text FROM enterprises WHERE status='approved' AND deleted_at IS NULL
  UNION ALL SELECT 'tot_demands', count(*)::text FROM demands WHERE deleted_at IS NULL
  UNION ALL SELECT 'tot_products', count(*)::text FROM drone_products WHERE deleted_at IS NULL
  UNION ALL SELECT 'tot_orders', count(*)::text FROM trade_orders
  UNION ALL SELECT 'esc_balance_fen', COALESCE(sum(balance_fen),0)::text FROM escrow_accounts
  UNION ALL SELECT 'esc_frozen_fen', COALESCE(sum(frozen_fen),0)::text FROM escrow_accounts
) m ORDER BY 1
SQL
)

raw=$(esc_q "$METRIC_SQL")
if [ -z "$raw" ]; then
  # 不盖 delivered_epoch、不写正文：让 ops-status 的「日报停发」检查把故障暴露出来。
  # 数据库连不上时运维告警本来就会报容器，这里再报一次只是重复，不去抢那个通道。
  log "取数失败：psql 无返回（容器 $CONTAINER 没跑？库 $DB_NAME 不对？）"
  exit 1
fi

metrics_tmp=$(mktemp) || exit 1
trap 'rm -f "$metrics_tmp"' EXIT
printf '%s\n' "$raw" > "$metrics_tmp"

# 系统好不好：判定复用 lib-notify.sh 的分项清单与中文名（与告警同一份，不另写一套）。
failed_json=$(notify_failed_sections "$SNAPSHOT")
failed_cn=""
if [ "$failed_json" = "null" ]; then
  ops_state=missing
elif [ "$failed_json" = "[]" ]; then
  ops_state=ok
else
  ops_state=bad
  for k in $(printf '%s' "$failed_json" | tr -d '[]"' | tr ',' ' '); do
    failed_cn="${failed_cn:+$failed_cn、}$(notify_section_label "$k")"
  done
fi

render=$(python3 - "$metrics_tmp" "$SNAPSHOT" "$DAY" "$OUT_JSON" "$failed_cn" "$ops_state" <<'PY'
import json, sys, time

metrics_path, snapshot_path, day_arg, json_path, failed_cn, ops_state = sys.argv[1:7]

raw = {}
with open(metrics_path, encoding='utf-8') as f:
    for line in f:
        line = line.strip()
        if '|' in line:
            k, v = line.split('|', 1)
            raw[k.strip()] = v.strip()

def num(key):
    try:
        return int(raw.get(key) or 0)
    except (TypeError, ValueError):
        return 0

def yuan(fen):
    return '¥%.2f' % (fen / 100.0)

def counts(pairs):
    return ' '.join('%s%d' % (label, value) for label, value in pairs)

# 日期标签：默认今天；DAY 指定时用那一天（补发/演练）
t = time.strptime(day_arg, '%Y-%m-%d') if day_arg else time.localtime()
day_label = time.strftime('%Y-%m-%d', t)
head = '【平台日报 %s 周%s】' % (time.strftime('%m-%d', t), '一二三四五六日'[t.tm_wday])

snap = {}
try:
    with open(snapshot_path, encoding='utf-8') as f:
        snap = json.load(f)
except Exception:
    snap = {}

if ops_state == 'missing':
    ops_line = '系统 快照缺失（运维检查没在跑）'
elif ops_state == 'bad':
    ops_line = '系统 异常：' + (failed_cn or '未知')
else:
    ops_line = '系统 正常'

bits = []
disk = (snap.get('disk') or {}).get('used_percent')
backup_age = (snap.get('backup') or {}).get('age_hours')
cert_days = (snap.get('cert') or {}).get('days_left')
if isinstance(disk, (int, float)):
    bits.append('磁盘%d%%' % disk)
if isinstance(backup_age, (int, float)) and backup_age >= 0:
    bits.append('备份%.1fh' % backup_age)
if isinstance(cert_days, (int, float)) and cert_days >= 0:
    bits.append('证书%d天' % cert_days)
if bits:
    ops_line += '（' + ' '.join(bits) + '）'

new_pairs = [('用户', num('new_users')), ('企业', num('new_enterprises')), ('需求', num('new_demands')),
             ('商品', num('new_products')), ('订单', num('new_orders'))]
todo_pairs = [('企业待审', num('todo_enterprise')), ('意向', num('todo_intent')),
              ('工单', num('todo_workorder')), ('售后', num('todo_aftersale')),
              ('商品待审', num('todo_check')), ('举报', num('todo_report'))]
tot_pairs = [('用户', num('tot_users')), ('企业', num('tot_enterprises')), ('需求', num('tot_demands')),
             ('商品', num('tot_products')), ('订单', num('tot_orders'))]
flow = num('money_in_fen') + num('release_fen') + num('refund_fen') + num('withdraw_fen')

lines = [head]
if flow == 0 and all(v == 0 for _, v in new_pairs):
    lines.append('今日无新增。')
else:
    lines.append('新增 ' + counts(new_pairs))
    lines.append('资金 入金%s 放款%s 退款%s 提现%s' % (
        yuan(num('money_in_fen')), yuan(num('release_fen')),
        yuan(num('refund_fen')), yuan(num('withdraw_fen'))))
lines.append('待办 ' + counts(todo_pairs))
lines.append('累计 ' + counts(tot_pairs))
lines.append('托管 余额%s 冻结%s' % (yuan(num('esc_balance_fen')), yuan(num('esc_frozen_fen'))))
lines.append(ops_line)
text = '\n'.join(lines)

payload = {
    'generated_at': time.strftime('%Y-%m-%dT%H:%M:%S%z'),
    'generated_epoch': int(time.time()),
    'day': day_label,
    'chars': len(text),
    'ops': ops_state,
    'ops_failed_cn': failed_cn,
    'metrics': raw,
    'text': text,
    'delivered_at': '',
    'delivered_epoch': 0,
}
with open(json_path, 'w', encoding='utf-8') as f:
    json.dump(payload, f, ensure_ascii=False, indent=2)
print(text)
PY
)
if [ -z "$render" ]; then
  log "渲染失败：python 没有输出正文（$SNAPSHOT 损坏？）"
  exit 1
fi

tmp="$OUT_TXT.tmp.$$"
printf '%s\n' "$render" > "$tmp" && mv -f "$tmp" "$OUT_TXT"
chars=${#render}
if [ "$chars" -gt 300 ]; then
  log "提示：正文 $chars 字，超过 300 字预算（加字段时注意）"
fi

if [ "$SEND" != 1 ]; then
  log "干跑（--dry-run / SEND=0）：已生成 $OUT_TXT（$chars 字），未推送，delivered_epoch 不更新"
  printf '%s\n' "$render"
  exit 0
fi

log "SEND 平台日报（$chars 字）"
if notify_send "$render"; then
  # 只有真送达才更新投递时间：ops-status 的「日报停发」检查看的就是它。
  python3 -c 'import json,sys,time
p=sys.argv[1]
d=json.load(open(p,encoding="utf-8"))
d["delivered_epoch"]=int(time.time())
d["delivered_at"]=time.strftime("%Y-%m-%dT%H:%M:%S%z")
json.dump(d,open(p,"w",encoding="utf-8"),ensure_ascii=False,indent=2)' "$OUT_JSON"
  log "已投递：$OUT_TXT"
  exit 0
fi

log "日报未能送达（通道问题，正文已落盘 $OUT_TXT）；delivered_epoch 保持不动，运维告警会跟出来"
exit 1