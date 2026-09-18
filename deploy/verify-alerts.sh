#!/bin/bash
# 告警链路演练：在**临时快照 + 临时配置**上把各种故障场景跑一遍，确认告警真的会报、且报得对。
#
# ⚠️ 为什么会有这个脚本（2026-09-18 的真实教训）：
# 我手工演练时直接覆盖了 /var/www/ops-status.json —— 那是**生产快照**，而 cron 每 10 分钟
# 就会读它。结果一条**假警报**推送到了真实群里（运维快照异常：top_level），当时系统
# 一切正常。两个错误叠加：变异测试写了生产文件 + 当时 alert 的分项列表还没同步。
#
# 结论：演练**绝不能碰**生产快照与生产配置。把正确做法固化成脚本，比写一句
# 「演练时记得用临时文件」可靠得多 —— 后者正是我当时以为自己在做的事。
#
# 用法：
#   bash deploy/verify-alerts.sh            # 干跑：只验判定，不推送（默认，安全）
#   bash deploy/verify-alerts.sh --send     # 额外发一条真实测试消息，验证通道连通
set -uo pipefail

SEND_REAL=0
[ "${1:-}" = "--send" ] && SEND_REAL=1

T=$(mktemp -d /tmp/alert-drill.XXXXXX)
trap 'rm -rf "$T"' EXIT
PROD_SNAPSHOT=/var/www/ops-status.json
PROD_ENV=/root/UAV/alert.env

# 全部指向临时文件：演练期间生产快照与生产配置一个字节都不会动。
OUT=$T/ops.json
STATE=$T/.state
LOG=$T/alert.log
ENVFILE=/dev/null
[ "$SEND_REAL" = 1 ] && ENVFILE=$PROD_ENV

[ -f "$PROD_SNAPSHOT" ] || { echo "找不到生产快照 $PROD_SNAPSHOT"; exit 1; }
prod_before=$(stat -c %Y "$PROD_SNAPSHOT")

pass=0; fail=0
check() {
  local desc="$1" want="$2" got
  got=$(grep -o 'SEND \[.*\] .*' "$LOG" 2>/dev/null | tail -1 | sed 's/^SEND //')
  if [ "$want" = SILENT ]; then
    if [ -z "$got" ]; then echo "  ok   $desc（未告警）"; pass=$((pass+1));
    else echo "  FAIL $desc：期望静默，实际告警「$got」"; fail=$((fail+1)); fi
  else
    if printf '%s' "$got" | grep -q "$want"; then echo "  ok   $desc | $got"; pass=$((pass+1));
    else echo "  FAIL $desc：期望含「$want」，实际「$got」"; fail=$((fail+1)); fi
  fi
  # 只清日志。**不能连 STATE 一起清**：冷却与「故障恢复」都依赖它，
  # 清掉就等于每次都从零开始 —— 我第一版就是这么错的，于是冷却和恢复两项假失败。
  # 需要重置状态的场景自己 rm -f。
  rm -f "$LOG"
}

# 以生产快照为模板造变体：mk <输出> <node.field=值> ...
mk() {
  python3 - "$PROD_SNAPSHOT" "$@" <<'PY'
import json, sys
src = sys.argv[1]
out = sys.argv[2]
d = json.load(open(src))
for kv in sys.argv[3:]:
    k, v = kv.split('=', 1)
    node, _, field = k.rpartition('.')
    if node:
        d[node][field] = eval(v)
    else:
        d[k] = eval(v)
json.dump(d, open(out, 'w'))
PY
}

echo "演练目录：$T"
echo "生产快照：$PROD_SNAPSHOT（演练期间不会被改动）"
if [ "$SEND_REAL" = 1 ]; then echo "推送：真实（会发一条到群里）"; else echo "推送：关（只验判定）"; fi
echo
echo "== 判定演练 =="

cp "$PROD_SNAPSHOT" "$T/healthy.json"
mk "$T/backup_bad.json"     'backup.ok=False' 'backup.age_hours=52.0'
mk "$T/disk_bad.json"       'disk.ok=False' 'disk.used_percent=93'
mk "$T/escrow_bad.json"     'escrow.ok=False' 'escrow.dup_keys=3'
mk "$T/jobs_bad.json"       'jobs.ok=False' 'jobs.hygiene_log_age_hours=99.0'
mk "$T/cert_bad.json"       'cert.ok=False' 'cert.days_left=3'
mk "$T/containers_bad.json" 'containers.ok=False' 'containers.api="exited"'

# 快照整体过期：生成脚本自己挂了，正是备份那次的形态
python3 - "$PROD_SNAPSHOT" "$T/stale.json" <<'PY'
import json, sys, time
d = json.load(open(sys.argv[1]))
d['generated_epoch'] = int(time.time()) - 10800
json.dump(d, open(sys.argv[2], 'w'))
PY

run() { SNAPSHOT="$1" OUT="$OUT" STATE="$STATE" LOG="$LOG" ENVFILE="$ENVFILE" bash /root/UAV/deploy/alert.sh >/dev/null 2>&1; }

# 每个独立场景开头重置状态；冷却与恢复这两项**刻意不重置**（它们正是靠状态工作的）。
reset() { rm -f "$STATE" "$LOG"; }

reset; run "$T/healthy.json";        check '健康快照' SILENT
reset; run "$T/backup_bad.json";     check '备份过期' backup
       run "$T/backup_bad.json";     check '同一故障立刻重放（冷却生效）' SILENT
       run "$T/healthy.json";        check '故障恢复' '已恢复正常'
reset; run "$T/disk_bad.json";       check '磁盘将满' disk
reset; run "$T/escrow_bad.json";     check '资金不变量失守' escrow
reset; run "$T/jobs_bad.json";       check '定时任务停摆' jobs
reset; run "$T/cert_bad.json";       check '证书将到期' cert
reset; run "$T/containers_bad.json"; check '容器异常' containers
reset; run "$T/stale.json";          check '快照超过 60 分钟未更新' snapshot_stale
# 文件缺失时消息是人类可读句子，不含签名串（签名只存在 STATE 里）——断言要按文本写
reset; run "$T/missing.json";        check '快照文件不存在' '运维快照文件不存在'

echo
if [ "$SEND_REAL" = 1 ]; then
  echo "== 通道连通 =="
  reset
  run "$T/backup_bad.json"
  if grep -q '推送成功 HTTP 200' "$LOG"; then echo "  ok   真实推送成功（群里应收到一条）"; pass=$((pass+1));
  else echo "  FAIL 真实推送失败：$(tail -1 "$LOG")"; fail=$((fail+1)); fi
  rm -f "$LOG" "$STATE"
  echo
fi

echo "== 结论：$pass 项通过，$fail 项失败 =="
prod_after=$(stat -c %Y "$PROD_SNAPSHOT")
if [ "$prod_before" = "$prod_after" ]; then
  echo "生产快照未被改动（mtime 一致）ok"
else
  echo "FAIL 生产快照被改动了，演练脚本有 bug"; fail=$((fail+1))
fi
[ "$fail" = 0 ] || exit 1