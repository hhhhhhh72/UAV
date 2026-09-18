#!/bin/bash
# 运维告警：读 ops-status 快照，不健康时推送到 IM 群机器人。
#
# 为什么需要（2026-09-18）：ops-status.sh 已经把健康指标全算出来了，但**没有任何东西
# 会主动通知人** —— 它是「你想起去看才看得到」。9/16 那次备份静默失败两天，就算当时
# 已有这份快照也没人看；GitHub Actions 探活失败虽然会发通知，但那不是天天会打开的地方。
#
# 通道：企业微信 / 钉钉 / 飞书 群机器人 webhook（只要一个 URL，不用申请任何凭据）。
# 在 /root/UAV/alert.env 写一行 ALERT_WEBHOOK=https://... 即启用；
# 未配置时**明确记日志**「告警通道未配置」，绝不静默地什么都不做。
#
# 去重与冷却：同一个故障最多每 COOLDOWN_SEC 提醒一次（默认 6 小时），故障消失时立即
# 发一条「已恢复」。没有这层，一个持续故障会每 10 分钟刷屏一次，最后被所有人无视。
set -uo pipefail

SNAPSHOT=${SNAPSHOT:-/var/www/ops-status.json}
STATE=${STATE:-$HOME/UAV-db-backups/.alert-state}
LOG=${LOG:-$HOME/UAV-db-backups/alert.log}
ENVFILE=${ENVFILE:-/root/UAV/alert.env}
SNAPSHOT_MAX_AGE_MIN=${SNAPSHOT_MAX_AGE_MIN:-60}
COOLDOWN_SEC=${COOLDOWN_SEC:-21600}
ALERT_FORMAT=${ALERT_FORMAT:-wecom}

[ -f "$ENVFILE" ] && . "$ENVFILE"
mkdir -p "$(dirname "$STATE")" 2>/dev/null || true

log() { echo "$(date -Iseconds) $*" >> "$LOG"; }

# ---- 判定（用 python3 读 JSON：比 sed/grep 可靠，且能把「哪几项不达标」逐条列出来）----
eval "$(python3 - "$SNAPSHOT" "$SNAPSHOT_MAX_AGE_MIN" <<'PY'
import json, shlex, sys, time
path, max_age = sys.argv[1], float(sys.argv[2])
healthy, sig, text = True, '', ''
try:
    with open(path) as f:
        d = json.load(f)
    bad = []
    for k in ('disk', 'backup', 'restore_drill', 'containers', 'cert', 'escrow'):
        v = d.get(k) or {}
        if v.get('ok') is not True:
            bad.append(k)
    if d.get('ok') is not True and not bad:
        bad.append('top_level')
    epoch = d.get('generated_epoch')
    age_min = (time.time() - float(epoch)) / 60.0 if epoch else None
    if age_min is None:
        bad.append('snapshot_no_epoch')
    elif age_min > max_age:
        bad.append('snapshot_stale(%dmin)' % int(age_min))
    if bad:
        healthy = False
        sig = ','.join(sorted(set(b.split('(')[0] for b in bad)))
        text = '运维快照异常：' + '、'.join(bad)
except FileNotFoundError:
    healthy = False; sig = 'snapshot_missing'; text = '运维快照文件不存在：' + path
except Exception as exc:
    healthy = False; sig = 'snapshot_unreadable'; text = '运维快照无法解析：%s' % exc
print('ALERT_HEALTHY=%s' % ('yes' if healthy else 'no'))
print('ALERT_SIG=%s' % shlex.quote(sig))
print('ALERT_TEXT=%s' % shlex.quote(text))
PY
)"

if [ -z "${ALERT_SIG+x}" ]; then
  log "判定失败：python 未输出结果（快照损坏或 python3 不可用）"
  exit 1
fi

# ---- 发送 ----
send() {
  local text="$1"
  log "SEND $text"
  if [ -z "${ALERT_WEBHOOK:-}" ]; then
    log "告警通道未配置：在 $ENVFILE 写入 ALERT_WEBHOOK=<群机器人 webhook> 即可启用推送"
    return 0
  fi
  local payload
  if [ "$ALERT_FORMAT" = feishu ]; then
    payload=$(python3 -c 'import json,sys;print(json.dumps({"msg_type":"text","content":{"text":sys.argv[1]}},ensure_ascii=False))' "$text")
  else
    payload=$(python3 -c 'import json,sys;print(json.dumps({"msgtype":"text","text":{"content":sys.argv[1]}},ensure_ascii=False))' "$text")
  fi
  local code
  code=$(curl -s -o /dev/null -w '%{http_code}' --max-time 10 -X POST -H 'Content-Type: application/json' -d "$payload" "$ALERT_WEBHOOK")
  log "推送 HTTP $code"
}

# ---- 冷却与恢复 ----
now=$(date +%s)
prev_sig=''; prev_epoch=0
if [ -f "$STATE" ]; then
  prev_sig=$(cut -d'|' -f1 "$STATE")
  prev_epoch=$(cut -d'|' -f2 "$STATE")
fi
case "$prev_epoch" in ''|*[!0-9]*) prev_epoch=0 ;; esac

if [ "$ALERT_HEALTHY" = yes ]; then
  if [ -n "$prev_sig" ]; then
    send "[恢复] $prev_sig 已恢复正常"
  fi
  printf '|%s\n' "$now" > "$STATE"
  exit 0
fi

if [ "$ALERT_SIG" = "$prev_sig" ] && [ $(( now - prev_epoch )) -lt "$COOLDOWN_SEC" ]; then
  log "同一故障在冷却期内，跳过提醒：$ALERT_SIG"
  exit 0
fi

send "[告警] $ALERT_TEXT（$(date '+%m-%d %H:%M')）"
printf '%s|%s\n' "$ALERT_SIG" "$now" > "$STATE"