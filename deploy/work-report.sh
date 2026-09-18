#!/bin/bash
# 工作日报：**由服务器生成并发送**，不依赖开发机是否开机。
#
# 为什么从本机搬到服务器（2026-09-18，用户问「我云服务不关机每天都有？」）：
#   第一版在开发机上生成（内容来自本机 git），开发机不在线那天就没有日报 —— 而服务器
#   24 小时开着却帮不上忙。实测服务器能直连 GitHub（`git ls-remote` 成功），于是改成
#   服务器每天 fetch 一份裸库自己生成。开发机那份脚本退化成「预览 / 手动补发」的薄壳。
#
# 裸库用 --filter=blob:none：只要提交与目录、不要文件内容，**1.4MB**（本机 .git 是 524MB），
# 每天 fetch 1 秒内完成。
#
# 用法：
#   bash deploy/work-report.sh              # 生成并推送
#   bash deploy/work-report.sh --dry-run    # 只打印，不推送、不动心跳
#   DAY=2026-09-17 bash deploy/work-report.sh --dry-run
#
# cron（root）：
#   30 19 * * * /root/UAV/deploy/work-report.sh >> /root/UAV-db-backups/work-report-cron.log 2>&1
#   定 19:30 而不是 17:30：服务器一直开着，晚一点能把整个下午的活都盖进来，也仍在 20:00 前。
set -uo pipefail

REPO=${REPO:-/root/UAV-repo.git}
DIR=${DIR:-$HOME/UAV-db-backups}
LOG=${LOG:-$DIR/work-report.log}
OUT=${OUT:-$DIR/work-report.txt}
HB=${HB:-$DIR/.work-report-heartbeat}
ENVFILE=${ENVFILE:-/root/UAV/alert.env}
CONTAINER=${CONTAINER:-uav-api-1}
DAY=${DAY:-$(date +%F)}
MODE=${1:-}

[ -f "$ENVFILE" ] && . "$ENVFILE"
# shellcheck source=deploy/lib-notify.sh
. "$(dirname "${BASH_SOURCE[0]}")/lib-notify.sh"
mkdir -p "$DIR"
log() { echo "$(date -Iseconds) $*" >> "$LOG"; return 0; }

if [ ! -d "$REPO" ]; then
  log "裸库不存在：$REPO"
  log "  建法：git clone --bare --filter=blob:none https://github.com/hhhhhhh72/UAV.git $REPO"
  exit 1
fi

# 先取最新提交。取不到就**直接退出**，绝不用旧数据发一份看起来正常的日报 ——
# 那正是「假绿」：数字都在，却少了一整天的活，谁看得出来。
if ! timeout 90 git -C "$REPO" fetch --quiet origin '+refs/heads/*:refs/heads/*' >>"$LOG" 2>&1; then
  log "git fetch 失败（GitHub 不通或超时）—— 不发旧数据，等下一次"
  exit 1
fi

# 部署节点：今天的日报里写清楚今天的代码到底上没上生产。
# **补看历史日期时不写这一行** —— 「今日已部署 14:34」配在 09-17 的日报上就是错的
# （容器是今天起的，跟那一天没关系）。
deploy_note=''
if [ "$DAY" = "$(date +%F)" ]; then
started=$(docker inspect -f '{{.State.StartedAt}}' "$CONTAINER" 2>/dev/null || true)
if [ -n "$started" ]; then
  # 去掉纳秒再交给 date（GNU date 认不了 9 位小数）
  iso=$(printf '%s' "$started" | sed -E 's/\.[0-9]+Z$/Z/')
  epoch=$(date -d "$iso" +%s 2>/dev/null || true)
  if [ -n "$epoch" ]; then
    if [ "$(date -d "@$epoch" +%F)" = "$(date +%F)" ]; then
      deploy_note="今日已部署生产 $(date -d "@$epoch" +%H:%M)"
    else
      deploy_note="生产上次部署 $(date -d "@$epoch" '+%m-%d %H:%M')"
    fi
  fi
fi
[ -n "$deploy_note" ] || deploy_note='部署状态未取到'
fi

text=$(python3 "$(dirname "${BASH_SOURCE[0]}")/work_report.py" "$REPO" "$DAY" "$deploy_note")
if [ -z "$text" ]; then
  log "生成失败：python 没有输出正文"
  exit 1
fi

chars=${#text}
printf '%s\n' "$text" > "$OUT"

if [ "$MODE" = "--dry-run" ]; then
  log "干跑（$chars 字），未推送、心跳未动"
  exit 0
fi

# 正文也打到 stdout：cron 日志里留一份存档，本机只读预览时直接拿这段
printf '%s\n' "$text"

log "SEND 工作日报（$chars 字，$DAY）"
if notify_send "$text"; then
  date +%s > "$HB"
  log "已投递（$chars 字），心跳已前移"
  exit 0
fi
log "工作日报**未能送达**（$chars 字）；心跳不动，运维告警会跟出来"
exit 1