#!/bin/bash
# 证书自动续期（acme.sh + TLS-ALPN-01）。由 cron 每天跑一次。
#
# 为什么是 TLS-ALPN-01 而不是 HTTP-01（2026-09-21 查清）：
#   域名 api.cqnarc.cn **未做 ICP 备案**，腾讯云会把**境外来源的 80 端口**请求劫持到
#   dnspod.qcloud.com 的「备案提示」页（实测：同一境外出口访问 http:// 拿到跳转 dnspod，
#   访问 https://api.cqnarc.cn/ops-status.json 拿到 200 真实内容）。Let's Encrypt 的
#   多视角校验节点全在境外 → HTTP-01 必然失败：
#     $ certbot certonly --webroot -w /var/www/acme-webroot -d api.cqnarc.cn
#       Detail: 43.174.225.201: Invalid response from
#               https://dnspod.qcloud.com/static/webblock.html?d=api.cqnarc.cn
#   443 没有被劫持，所以在 443 上做 TLS-ALPN-01 校验。
#
# 为什么不是 certbot：certbot 2.9 的 nginx 插件和 standalone 插件都只认 HTTP-01
#   （源码 supported_challenges 返回 [challenges.HTTP01]，实测 --nginx / --standalone
#    配 --preferred-challenges tls-alpn-01 一律 "None of the preferred challenges are
#    supported by the selected plugin"）。lego 能做 ALPN，但它强制要求 --email。
#
# 为什么不用邮箱：负责人明确要求"不用邮箱，跟报错报告一样发企业微信就行"。
#   acme.sh 走 Let's Encrypt 时不强制 --accountemail，正好满足；到期提醒由
#   ops-status.sh 的 cert 一节（21 天阈值）+ alert.sh 推到企业微信群兜底。
#
# 为什么"不必要就不动 nginx"：ALPN 校验期间 443 要让给 acme.sh 的临时 openssl s_server，
#   nginx 必须停一下。实测一次完整续期约 30 秒（LE 在 3 秒内就校验完，剩下 27 秒
#   全是 acme.sh 自己的轮询/下载）。所以**先看证书还剩几天**，没到阈值就直接退出，
#   一次也不碰 nginx —— 正常情况每 60 天才停一次，而不是每天停一次。
set -uo pipefail

ACME=${ACME_SH:-/root/UAV/tools/acme.sh}
ACME_HOME=${ACME_HOME:-/root/UAV/tools/acme-home}
DOMAIN=${DOMAIN:-api.cqnarc.cn}
SERVER=${ACME_SERVER:-letsencrypt}
KEYLEN=${ACME_KEYLEN:-ec-256}
CERT=${CERT:-/etc/nginx/certs/api.cqnarc.cn.fullchain.crt}
KEY=${KEY:-/etc/nginx/certs/api.cqnarc.cn.key}
HOOKS=${CERT_HOOKS:-/root/UAV/deploy/cert-hooks.sh}
LOG=${CERT_RENEW_LOG:-/root/UAV-db-backups/cert-renew.log}
LOCK=${CERT_RENEW_LOCK:-/var/lock/uav-cert-renew.lock}
RENEW_BEFORE_DAYS=${RENEW_BEFORE_DAYS:-30}
CERT_HOST=${CERT_HOST:-127.0.0.1}
CERT_PORT=${CERT_PORT:-443}
FORCE=${FORCE:-0}
# acme.sh 是从 GitHub raw 取的单文件 shell 脚本、以 root 运行，所以钉死哈希。
# 升级时先核对新哈希、连同这里一起改，别让它悄悄变。
ACME_SHA256=${ACME_SHA256:-c7d68b021cfd6380ea83a82962abde5b484779fee0b97d38681dfa1396bbc8d7}
ALLOW_ACME_HASH_MISMATCH=${ALLOW_ACME_HASH_MISMATCH:-0}

log() { printf '[%s] %s\n' "$(date -Iseconds)" "$*" >>"$LOG" 2>/dev/null || printf '[%s] %s\n' "$(date -Iseconds)" "$*"; }

# 读一个证书文件的到期时间 → 剩余天数；读不到返回 -1
days_left_of_file() {
  local f="$1" end
  [ -f "$f" ] || { echo -1; return; }
  end=$(openssl x509 -in "$f" -noout -enddate 2>/dev/null | cut -d= -f2)
  [ -n "$end" ] || { echo -1; return; }
  echo $(( ( $(date -d "$end" +%s) - $(date +%s) ) / 86400 ))
}

# 读 nginx **实际在服务**的那张证书（不是文件）→ "serial enddate"
# 为什么必须看这个：证书文件换了但 nginx 没 reload 时，文件检查全绿、
# 客户端拿到的却还是旧证书 —— 典型的假绿。
served_cert() {
  echo | timeout 10 openssl s_client -connect "$CERT_HOST:$CERT_PORT" -servername "$DOMAIN" 2>/dev/null \
    | openssl x509 -noout -serial -enddate 2>/dev/null | sed 's/^serial=//; s/^notAfter=//' | tr '\n' ' '
}

exec 9>"$LOCK" 2>/dev/null || true
if ! flock -n 9; then
  log "已有一个续期进程在跑，本次跳过"
  exit 0
fi

# ---- 0. 工具自检（宁可报，不可假绿：检查跑不起来时最不该沉默放过）----
if [ ! -x "$ACME" ]; then
  log "！找不到可执行的 acme.sh（$ACME），无法续期"
  exit 1
fi
if [ -n "$ACME_SHA256" ]; then
  actual=$(sha256sum "$ACME" | cut -d' ' -f1)
  if [ "$actual" != "$ACME_SHA256" ] && [ "$ALLOW_ACME_HASH_MISMATCH" != "1" ]; then
    log "！acme.sh 哈希与钉死值不一致（期望 $ACME_SHA256，实际 $actual）—— 拒绝执行；确认无误后置 ALLOW_ACME_HASH_MISMATCH=1"
    exit 1
  fi
fi

# ---- 1. 还没到期就直接退出，一次也不碰 nginx ----
left=$(days_left_of_file "$CERT")
if [ "$FORCE" != "1" ] && [ "$left" -gt "$RENEW_BEFORE_DAYS" ]; then
  log "证书还有 $left 天（阈值 $RENEW_BEFORE_DAYS 天），本次不续期"
  exit 0
fi

log "开始续期（当前证书剩余 $left 天，FORCE=$FORCE）"
before_serial=$(served_cert | awk '{print $1}')
before_left=$(days_left_of_file "$CERT")

# ---- 2. 让出 443 → 续期 → 无论成败都把 nginx 拉回来 ----
"$HOOKS" pre || { log "！pre 钩子失败（443 没让出来），放弃本次续期"; exit 1; }
restore_nginx() { "$HOOKS" post || log "！post 钩子失败，nginx 可能仍处于停止状态"; }
trap 'restore_nginx' EXIT INT TERM

# `9>&-` 不能省：fd 9 是本次续期的 flock 单实例锁。子进程会继承它，
# 而继承它的进程只要还活着，锁就一直不放 —— 实测过一次（drill 里一个后台 TLS 进程
# 继承了 fd 9，之后所有续期调用都被判成"已有进程在跑"而静默跳过）。
# 这里显式关掉，让"锁只由 cert-renew.sh 自己持有"变成结构上的保证，而不是靠子进程自觉。
"$ACME" --issue --alpn -d "$DOMAIN" 9>&- \
  --server "$SERVER" --home "$ACME_HOME" --keylength "$KEYLEN" --force \
  --key-file "$KEY" --fullchain-file "$CERT" \
  --reloadcmd "$HOOKS deploy" \
  --log "$LOG.acme" >>"$LOG.acme" 2>&1
rc=$?

restore_nginx
trap - EXIT INT TERM

if [ "$rc" != "0" ]; then
  log "！acme.sh 续期失败（exit=$rc），详见 $LOG.acme；nginx 已恢复，证书保持原样（还有 $before_left 天）"
  exit 1
fi

# ---- 3. 复核：文件换了、**nginx 也确实换上了** ----
chown www-data:www-data "$CERT" "$KEY" 2>/dev/null || true
chmod 644 "$CERT" 2>/dev/null || true
chmod 600 "$KEY" 2>/dev/null || true

after_left=$(days_left_of_file "$CERT")
after_serial=$(served_cert | awk '{print $1}')
if [ "$after_left" -lt 0 ]; then
  log "！续期后读不到证书文件，请人工检查 $CERT"
  exit 1
fi
# 空值必须先拦：如果这里读不到（握手失败），after_serial 是空串，
# 而空串 ≠ before_serial 会被下面的相等判断放过去，于是"续期成功"照报、实际什么都没验证 ——
# 正是本项目最忌讳的假绿。这条断言是演练里补出来的（第一版真的报了"续期成功 … →"）。
if [ -z "$after_serial" ]; then
  log "！续期后读不到 nginx 正在服务的证书（$CERT_HOST:$CERT_PORT 握手失败）—— 无法确认新证书已生效"
  exit 1
fi
if [ "$after_serial" = "$before_serial" ]; then
  log "！续期后 nginx 仍在服务旧证书（serial 未变 $after_serial）—— 证书文件已更新但没生效，需人工 reload"
  exit 1
fi
log "续期成功：新证书剩余 $after_left 天（serial $before_serial → $after_serial）"
