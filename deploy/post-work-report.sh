#!/bin/bash
# 把本机（Windows 开发机）生成的工作日报发到群里，并记下「送达」心跳。
#
# 为什么发送这一步放在服务器、而不是本机直接 POST（2026-09-18 实测）：
#   本机的 schannel 拿不到 TLS 凭据 —— curl 与 Invoke-WebRequest 一律报
#   `(35) schannel: AcquireCredentialsHandle failed: SEC_E_NO_CREDENTIALS`，
#   任何 https 都发不出去；而服务器上这条通道早就验证过（errcode:0）。
#   所以本机只负责**生成文本**，用 scp 传过来；发送与送达检查都留在服务器这一侧，
#   与告警共用 deploy/lib-notify.sh —— 不做第二套实现（本项目已经因为两份实现
#   各自漂移吃过亏：告警分项列表 vs ops-status 分项）。
#
# 用法：bash deploy/post-work-report.sh /tmp/work-report.txt
#   文本经 scp 传文件而不是管道，是为了绕开 Windows PowerShell 向原生程序
#   写 stdin 的编码问题（PS 5.1 的 $OutputEncoding 默认 ASCII，中文会变问号）。
#
# 心跳：**只在推送成功后**才写 $HB。ops-status.sh 的 jobs 一节检查它的年龄，
# 于是「日报没发出去」会自己冒出来 —— 而不是等人某天想起来问「今天怎么没日报」。
set -uo pipefail

DIR=${DIR:-$HOME/UAV-db-backups}
FILE=${1:-}
LOG=${LOG:-$DIR/work-report.log}
HB=${HB:-$DIR/.work-report-heartbeat}
ENVFILE=${ENVFILE:-/root/UAV/alert.env}

[ -f "$ENVFILE" ] && . "$ENVFILE"
# shellcheck source=deploy/lib-notify.sh
. "$(dirname "${BASH_SOURCE[0]}")/lib-notify.sh"
mkdir -p "$DIR"
log() { echo "$(date -Iseconds) $*" >> "$LOG"; return 0; }

if [ -z "$FILE" ] || [ ! -f "$FILE" ]; then
  log "取不到日报文件：${FILE:-（没给路径）}"
  exit 1
fi

# 去掉可能的 UTF-8 BOM 与 CR：本机生成的文件经 scp 过来，字节是原样的，
# 但不同生成方式可能带 BOM；带了就会顶在正文最前面显示成一个乱码方块。
text=$(sed -e '1s/^\xEF\xBB\xBF//' "$FILE" | tr -d '\r')
if [ -z "$text" ]; then
  log "日报文件是空的：$FILE"
  exit 1
fi

chars=${#text}
log "SEND 工作日报（$chars 字）"
if notify_send "$text"; then
  date +%s > "$HB"
  log "已投递（$chars 字），心跳已前移"
  exit 0
fi

log "工作日报**未能送达**（$chars 字）；心跳不动，运维告警会把「日报停发」报出来"
exit 1