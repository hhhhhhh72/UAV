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
# HB 心跳文件：每次**正常跑完**覆盖写一次时间戳（覆盖而非追加，不增长）。
# ops-status 会检查它是否新鲜 —— 用来发现「告警脚本自己被停掉/没被 cron 调起」。
# 注意写"正常跑完"而不是任何退出：脚本自己挂了就不该盖心跳，否则故障被自己掩盖。
HB=${HB:-$HOME/UAV-db-backups/.alert-heartbeat}

[ -f "$ENVFILE" ] && . "$ENVFILE"
mkdir -p "$(dirname "$STATE")" 2>/dev/null || true

log() { echo "$(date -Iseconds) $*" >> "$LOG"; }

# mask_url 打日志用：webhook 地址里的 key 等同于"往群里发消息的钥匙"，
# 日志文件可能被备份/上传，不能原样落盘。只留域名与 key 的前 6 位。
# finish 正常收尾：盖心跳再退出。只有走到这里的运行才算"跑完了"。
finish() {
  echo "$(date +%s)" > "$HB" 2>/dev/null || true
  exit "${1:-0}"
}

mask_url() {
  printf '%s' "$1" | sed -E 's#(key=)([A-Za-z0-9_-]{0,6})[A-Za-z0-9_-]*#\1\2...#'
}

# ---- 判定（用 python3 读 JSON：比 sed/grep 可靠，且能把「哪几项不达标」逐条列出来）----
eval "$(python3 - "$SNAPSHOT" "$SNAPSHOT_MAX_AGE_MIN" <<'PY'
import json, shlex, sys, time
path, max_age = sys.argv[1], float(sys.argv[2])
healthy, sig, text = True, '', ''
try:
    with open(path) as f:
        d = json.load(f)
    bad = []
    # 这份列表必须与 ops-status.sh 输出的分项**保持同步**：漏掉一项时，该分项失守
    # 会让下面 `not bad` 的分支兜成 'top_level'——告警照发，但消息完全没有指向性
    # （"运维快照异常：top_level" 等于没说）。新增分项时两边都要改。
    for k in ('disk', 'backup', 'restore_drill', 'containers', 'cert', 'escrow', 'jobs'):
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
  # 形状校验：群机器人 webhook 有固定的域名与路径前缀。
  # 为什么必须有这一步：填错的后果不是报错，而是**假绿** —— 比如误把管理后台的
  # 机器人资料页链接（work.weixin.qq.com/wework_admin/common/openBotProfile/...）
  # 填进来，那个页面同样返回 HTTP 200，脚本会记一条「推送 HTTP 200」就以为发出去了，
  # 群里却什么都没有。与备份静默失败同一类问题：通道坏了但没人知道。
  case "$ALERT_WEBHOOK" in
    *qyapi.weixin.qq.com/cgi-bin/webhook/send*|*oapi.dingtalk.com/robot/send*|*open.feishu.cn/open-apis/bot*) ;;
    *)
      log "ALERT_WEBHOOK 形状不对（不是群机器人 webhook 地址，已跳过推送）：$(mask_url "$ALERT_WEBHOOK")"
      log "  正确形态应为 https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=... （企业微信）；钉钉/飞书见 alert.env 里的说明"
      return 1
      ;;
  esac
  # 只看 HTTP 码不够：必须看返回体里的 errcode —— 通道类错误（key 失效、机器人被移除、
  # 触达频率超限）通常都是 HTTP 200 + errcode != 0。
  local resp code body
  resp=$(curl -s --max-time 10 -X POST -H 'Content-Type: application/json' -d "$payload" -w '\n%{http_code}' "$ALERT_WEBHOOK" 2>/dev/null || true)
  code=$(printf '%s' "$resp" | tail -n 1)
  body=$(printf '%s' "$resp" | sed '$d')
  case "$body" in
    *'"errcode":0'*|*'"errcode": 0'*|*'"code":0'*|*'"code": 0'*)
      log "推送成功 HTTP $code"
      ;;
    *)
      log "推送**可能没送到**：HTTP $code 返回体 $(printf '%s' "$body" | head -c 200)"
      return 1
      ;;
  esac
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
  finish 0
fi

if [ "$ALERT_SIG" = "$prev_sig" ] && [ $(( now - prev_epoch )) -lt "$COOLDOWN_SEC" ]; then
  log "同一故障在冷却期内，跳过提醒：$ALERT_SIG"
  finish 0
fi

send "[告警] $ALERT_TEXT（$(date '+%m-%d %H:%M')）"
printf '%s|%s\n' "$ALERT_SIG" "$now" > "$STATE"
finish 0