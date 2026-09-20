#!/bin/bash
# 通知通道公共实现：deploy/alert.sh（告警）与 deploy/daily-report.sh（日报）共用。
#
# 为什么单独抽出来（2026-09-18）：
#   本文件里的两步校验是「假绿」的唯一阻挡 ——
#     ① 形状校验：webhook 填错的后果不是报错，而是 HTTP 200 + 群里什么都没有；
#     ② errcode 校验：通道类错误（key 失效、机器人被移出群、触发频率超限）
#        全是 HTTP 200 且 errcode ≠ 0。
#   各写一份迟早漂移，而本项目已经因为「两份列表不同步」吃过一次亏
#   （alert.sh 的分项清单 vs ops-status.sh 实际输出的分项：漏项时只会兜成
#    没有指向性的「运维快照异常：top_level」，等于没说）。
#
# 用法：调用方 source 本文件，然后 notify_send "<文本>"。
#   依赖变量：NOTIFY_WEBHOOK（未设时回落到 ALERT_WEBHOOK）
#   可选变量：NOTIFY_FORMAT（wecom|feishu，未设时回落到 ALERT_FORMAT，默认 wecom）
#   可选函数：调用方的 log()（未定义时兜底打到 stderr）
#   返回：0 送达 / 1 未配置或未送达 / 2 webhook 形状不对
set -uo pipefail

# ops-status 快照的**分项清单**：告警与日报都靠它决定「哪一项失守」。
# 只此一份，新增分项时改这里即可。deploy/verify-alerts.sh 会拿生产快照里实际的
# 分项与它逐项比对，漏项直接演练失败 —— 把「记得同步两份列表」交给机器。
NOTIFY_STATUS_SECTIONS="disk backup restore_drill containers cert escrow jobs instances"

# 分项的中文名：推送里出现的是给人看的词，不是 JSON 的 key。
notify_section_label() {
  case "$1" in
    disk)          echo 磁盘 ;;
    backup)        echo 备份 ;;
    restore_drill) echo 还原演练 ;;
    containers)    echo 容器 ;;
    cert)          echo 证书 ;;
    escrow)        echo 资金 ;;
    jobs)          echo 定时任务 ;;
    instances)     echo 实例数 ;;
    top_level)     echo 汇总 ;;
    *)             echo "$1" ;;
  esac
}

# notify_failed_sections <快照JSON路径> —— 输出失守分项的 JSON 数组（如 ["disk","jobs"]）。
# 判定口径与 alert.sh 历史行为一致：分项自己没打 ok=true 即算失守；
# 分项全绿但顶层 ok=false 时说明清单漏项，报 top_level（提醒去改 NOTIFY_STATUS_SECTIONS）。
notify_failed_sections() {
  python3 - "$1" "$NOTIFY_STATUS_SECTIONS" <<'PY'
import json, sys
path, sections = sys.argv[1], sys.argv[2].split()
try:
    with open(path) as f:
        d = json.load(f)
except Exception:
    print('null'); raise SystemExit(0)
bad = [k for k in sections if (d.get(k) or {}).get('ok') is not True]
if d.get('ok') is not True and not bad:
    bad.append('top_level')
print(json.dumps(bad, ensure_ascii=False))
PY
}

# mask_url 打日志用：webhook 里的 key 等同于「往群里发消息的钥匙」，
# 日志文件可能被备份/上传，不能原样落盘。只留域名与 key 的前 6 位。
notify_mask_url() {
  printf '%s' "$1" | sed -E 's#(key=)([A-Za-z0-9_-]{0,6})[A-Za-z0-9_-]*#\1\2...#'
}

notify_send() {
  local text="$1"
  if ! declare -F log >/dev/null 2>&1; then
    log() { echo "$*" >&2; }
  fi
  local hook="${NOTIFY_WEBHOOK:-${ALERT_WEBHOOK:-}}"
  local fmt="${NOTIFY_FORMAT:-${ALERT_FORMAT:-wecom}}"
  if [ -z "$hook" ]; then
    log "通知通道未配置：在 ${ENVFILE:-/root/UAV/alert.env} 写入 ALERT_WEBHOOK=<群机器人 webhook> 即可启用推送"
    return 1
  fi
  # 形状校验：群机器人 webhook 有固定的域名与路径前缀。
  # 为什么必须有这一步：填错的后果不是报错，而是**假绿** —— 比如误把管理后台的
  # 机器人资料页链接（work.weixin.qq.com/wework_admin/common/openBotProfile/...）
  # 填进来，那个页面同样返回 HTTP 200，脚本会记一条「推送 HTTP 200」就以为发出去了，
  # 群里却什么都没有。与备份静默失败同一类问题：通道坏了但没人知道。
  case "$hook" in
    *qyapi.weixin.qq.com/cgi-bin/webhook/send*|*oapi.dingtalk.com/robot/send*|*open.feishu.cn/open-apis/bot*) ;;
    *)
      log "通知地址形状不对（不是群机器人 webhook，已跳过推送）：$(notify_mask_url "$hook")"
      log "  正确形态应为 https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=... （企业微信）；钉钉/飞书见 alert.env 里的说明"
      return 2
      ;;
  esac
  local payload
  if [ "$fmt" = feishu ]; then
    payload=$(python3 -c 'import json,sys;print(json.dumps({"msg_type":"text","content":{"text":sys.argv[1]}},ensure_ascii=False))' "$text")
  else
    payload=$(python3 -c 'import json,sys;print(json.dumps({"msgtype":"text","text":{"content":sys.argv[1]}},ensure_ascii=False))' "$text")
  fi
  # 只看 HTTP 码不够：必须看返回体里的 errcode —— 通道类错误（key 失效、机器人被移除、
  # 触达频率超限）通常都是 HTTP 200 + errcode != 0。
  local resp code body
  resp=$(curl -s --max-time 10 -X POST -H 'Content-Type: application/json' -d "$payload" -w '\n%{http_code}' "$hook" 2>/dev/null || true)
  code=$(printf '%s' "$resp" | tail -n 1)
  body=$(printf '%s' "$resp" | sed '$d')
  case "$body" in
    *'"errcode":0'*|*'"errcode": 0'*|*'"code":0'*|*'"code": 0'*)
      log "推送成功 HTTP $code"
      return 0
      ;;
    *)
      log "推送**可能没送到**：HTTP $code 返回体 $(printf '%s' "$body" | head -c 200)"
      return 1
      ;;
  esac
}