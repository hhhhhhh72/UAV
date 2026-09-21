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
#   DAY=2026-09-17 SKIP_FETCH=1 bash deploy/work-report.sh   # 补发（网络不通、但本地裸库已有数据）
#
# cron（root）：
#   30 17 * * * /root/UAV/deploy/work-report.sh >> /root/UAV-db-backups/work-report-cron.log 2>&1
#   2026-09-21 起定 **17:30**（负责人要求提前；此前是 19:30）。
#   **已知代价**：17:30 之后提交的活当天看不到，也**不会补到第二天**（第二天报的是第二天的提交）。
#   原定 19:30 就是为了盖住整个下午。要两头都占，可以再加一班次日早晨只报
#   「昨天 17:30 之后的新增提交」—— 需要给 work_report.py 加一个 --since 口径，还没做。
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
# SKIP_FETCH=1 跳过 git fetch，直接用本地裸库里已有的数据生成 —— **人工补发**用。
# 场景（2026-09-21 真实发生）：服务器连不上 github.com:443（HTTPS 主站被挡，api/codeload/
# ssh:443/22 都通），当天那一班按设计拒发；而本地裸库其实**已经有当天全部提交**
#（17:11 那次 fetch 是成功的）。补发只需要一个"别再 fetch"的开关。
SKIP_FETCH=${SKIP_FETCH:-0}
# 重试窗口：偶发封锁可能持续几十分钟，5 分钟就放弃会把一整天的汇报丢掉
#（2026-09-20 实测：三次重试在 19:35 全部失败，当天 16 个提交一条没汇报）。
# 注意窗口要能盖住"这一班"：17:30 起跑 + 9000 秒 = 20:00 收工，仍在当天。
# 每轮内部仍是 60/120/180 秒递进三次，轮与轮之间间隔 RETRY_GAP_SECONDS，
# 直到累计等待超过 RETRY_WINDOW_SECONDS 才放弃并通知。
RETRY_WINDOW_SECONDS=${RETRY_WINDOW_SECONDS:-9000}
RETRY_GAP_SECONDS=${RETRY_GAP_SECONDS:-300}
# 备用通道：2026-09-21 实测这台机器**只有 github.com:443 不通**（HTTPS 主站被挡），
# 而 api.github.com / codeload / raw / ssh.github.com:443 / github.com:22 全通。
# 裸库的 origin 正是被封的那个 https 地址，于是每天那一班只能干等。
# 这里加一条 SSH over 443 的备用通道（需要仓库侧加只读 Deploy key，见 .pub）。
REPORT_REMOTE_FALLBACK=${REPORT_REMOTE_FALLBACK:-ssh://git@ssh.github.com:443/hhhhhhh72/UAV.git}
REPORT_SSH_KEY=${REPORT_SSH_KEY:-/root/.ssh/uav_report_ed25519}
REPORT_SSH_OPTS=${REPORT_SSH_OPTS:-"ssh -i $REPORT_SSH_KEY -o StrictHostKeyChecking=accept-new -o ConnectTimeout=15"}

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

# 先取最新提交。**取不到就绝不发**，也绝不用旧数据发一份看起来正常的日报 ——
# 那正是「假绿」：数字都在，却少了一整天的活，谁看得出来。
#
# 为什么要重试（2026-09-18 实测）：这台机器到 GitHub 的线路**不稳**。连着测三次 fetch，
# 两次是 2-3 秒，一次跑了 **133 秒后失败**（exit=128）；同一时刻 curl github.com 也超时。
# 只试一次的话，这种偶发就会被记成「今天没有日报」，而且要到 30 小时后心跳告警才暴露。
# 所以按 60/120/180 秒递进重试三次：偶发抖动基本都能自愈。
fetch_ok=0
fetch_started=$(date +%s)
fetch_round=0
if [ "$SKIP_FETCH" = "1" ]; then
  fetch_ok=1
  log "SKIP_FETCH=1：跳过 fetch，用本地裸库现有数据生成（人工补发）"
fi
while [ "$fetch_ok" != 1 ]; do
  fetch_round=$((fetch_round + 1))
  for attempt in 1 2 3; do
    fetch_timeout=$(( attempt * 60 ))
    if timeout "$fetch_timeout" git -C "$REPO" fetch --quiet origin '+refs/heads/*:refs/heads/*' >>"$LOG" 2>&1; then
      fetch_ok=1
      [ "$fetch_round" -gt 1 ] && log "git fetch 第 ${fetch_round} 轮才成功（等了 $(( ($(date +%s) - fetch_started) / 60 )) 分钟）"
      [ "$fetch_round" -eq 1 ] && [ "$attempt" -gt 1 ] && log "git fetch 第 $attempt 次才成功"
      break
    fi
    log "git fetch 失败（https，第 ${fetch_round} 轮第 ${attempt} 次，超时 ${fetch_timeout}s）"
    # 备用通道：SSH over 443。https 通时根本不会走到这里。
    if GIT_SSH_COMMAND="$REPORT_SSH_OPTS" timeout "$fetch_timeout" \
         git -C "$REPO" fetch --quiet "$REPORT_REMOTE_FALLBACK" '+refs/heads/*:refs/heads/*' >>"$LOG" 2>&1; then
      fetch_ok=1
      log "https 不通 → 备用通道（SSH over 443）成功（第 ${fetch_round} 轮第 ${attempt} 次）"
      break
    fi
    log "备用通道（SSH）也失败（第 ${fetch_round} 轮第 ${attempt} 次）"
    sleep 5
  done
  [ "$fetch_ok" = 1 ] && break
  waited=$(( $(date +%s) - fetch_started ))
  if [ "$waited" -ge "$RETRY_WINDOW_SECONDS" ]; then break; fi
  log "本轮三次都失败（已等 $(( waited / 60 )) 分钟），${RETRY_GAP_SECONDS}s 后重试"
  sleep "$RETRY_GAP_SECONDS"
done

if [ "$fetch_ok" != 1 ]; then
  waited_min=$(( ($(date +%s) - fetch_started) / 60 ))
  log "git fetch 在 $waited_min 分钟内 ${fetch_round} 轮都失败 —— 不发旧数据"
  # 立刻让人知道，而不是等 30 小时后由心跳告警兜出来（那时已经隔了一天）。
  # 注意**不动心跳**：心跳的含义是「今天的日报送到了」，这条不是日报。
  if [ "$MODE" != "--dry-run" ]; then
    notify_send "【日报异常】$(date +%m-%d) 的工作日报没能生成：服务器连不上 GitHub（已重试 ${fetch_round} 轮、${waited_min} 分钟）。历史提交还在，只是今天这份取不到数据。" || true
  fi
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

# 正文写一份到文件，同时打到 stdout —— **两种模式都要打**：
#   cron 日志里留一份存档；本机预览走的正是 --dry-run，靠 stdout 拿正文。
#   （这个 printf 一度只放在发送分支里，结果 --dry-run 什么都不输出，本机预览直接空手而归。）
chars=${#text}
printf '%s\n' "$text" > "$OUT"
printf '%s\n' "$text"

if [ "$MODE" = "--dry-run" ]; then
  log "干跑（$chars 字），未推送、心跳未动"
  exit 0
fi

log "SEND 工作日报（$chars 字，$DAY）"
if notify_send "$text"; then
  date +%s > "$HB"
  log "已投递（$chars 字），心跳已前移"
  exit 0
fi
log "工作日报**未能送达**（$chars 字）；心跳不动，运维告警会跟出来"
exit 1