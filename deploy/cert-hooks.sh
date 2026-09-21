#!/bin/bash
# 证书续期钩子：pre / post / deploy，由 certbot 在续期时调用。
#
# 为什么需要它（2026-09-21）：
#   本机部署在腾讯云境内 CVM 上，域名 api.cqnarc.cn **未做 ICP 备案**，腾讯云会把
#   **境外来源的 80 端口**请求劫持到 dnspod.qcloud.com 的「备案提示」页。
#   实测（同一境外出口，只换端口）：
#     http://api.cqnarc.cn/                 → 302 到 https://dnspod.qcloud.com/static/webblock.html?d=api.cqnarc.cn
#     https://api.cqnarc.cn/ops-status.json → 200，真实内容
#   而 Let's Encrypt 的多视角校验节点全在境外 → **HTTP-01 永远失败**：
#     $ certbot certonly --webroot -w /var/www/acme-webroot -d api.cqnarc.cn
#       Detail: 43.174.225.201: Invalid response from
#               https://dnspod.qcloud.com/static/webblock.html?d=api.cqnarc.cn
#   443 没有被劫持，所以改用 **TLS-ALPN-01**（同样在 443 上做校验）。
#   但 certbot 2.9 的 nginx 插件只支持 HTTP-01（源码 supported_challenges 返回
#   [challenges.HTTP01]，实测 --nginx --preferred-challenges tls-alpn-01 直接报
#   "None of the preferred challenges are supported by the selected plugin"），
#   只能走 standalone —— 于是校验期间 443 要让给 certbot，nginx 必须**短暂停一下**。
#
# 因此这里的核心不是"停 nginx"，而是"**保证 nginx 一定回得来**"：
#   ① pre 在停 nginx **之前**先挂一个 5 分钟的死人开关（setsid 脱离父进程），
#      到点发现 nginx 没在跑就无条件拉起来 —— 即使 certbot 被 SIGKILL、
#      post 钩子根本没执行，站点也会自己活过来；
#   ② post 幂等地把 nginx 拉起来，并用 systemctl is-active 复核；
#   ③ deploy 重载 nginx 让新证书生效（证书文件换了但 nginx 内存里还是旧证书，
#      这正是"文件检查全绿、客户端却仍在用旧证书"的假绿来源）。
#
# 三个钩子都幂等，执行顺序（pre → 校验 → deploy → post，或 pre → 校验 → post → deploy）
# 无论哪种都收敛到同一个结果。
set -uo pipefail

LOG=${CERT_HOOK_LOG:-/root/UAV-db-backups/cert-hook.log}
NGINX_UNIT=${NGINX_UNIT:-nginx}
DEADMAN_SECONDS=${DEADMAN_SECONDS:-300}
# 演练用：把 systemctl 换成假的，即可在不碰生产 nginx 的前提下验证钩子逻辑。
SYSTEMCTL=${SYSTEMCTL:-systemctl}

log() { printf '[%s] %s\n' "$(date -Iseconds)" "$*" >>"$LOG" 2>/dev/null || true; }

nginx_running() { "$SYSTEMCTL" is-active --quiet "$NGINX_UNIT"; }

start_nginx() {
  "$SYSTEMCTL" start "$NGINX_UNIT" >/dev/null 2>&1
  nginx_running
}

case "${1:-}" in
  pre)
    log "pre: 停 $NGINX_UNIT 让出 443 供 tls-alpn-01 校验"
    # 死人开关：先安排好"万一没人管，5 分钟后自己把站点拉回来"，再去停服务。
    # 顺序不能颠倒 —— 先停再挂开关，就存在一段"停了但没人兜底"的窗口。
    setsid nohup bash -c "
      sleep $DEADMAN_SECONDS
      $SYSTEMCTL is-active --quiet $NGINX_UNIT || {
        $SYSTEMCTL start $NGINX_UNIT >/dev/null 2>&1
        echo \"[\$(date -Iseconds)] deadman: $NGINX_UNIT 在续期结束后仍未运行，已强制拉起\" >>'$LOG'
      }
    " >/dev/null 2>&1 </dev/null 9>&- &
    disown 2>/dev/null || true
    if ! "$SYSTEMCTL" stop "$NGINX_UNIT" >/dev/null 2>&1; then
      log "pre: 停 $NGINX_UNIT 失败，放弃本次续期（站点保持在线）"
      exit 1
    fi
    log "pre: 已停 $NGINX_UNIT，死人开关 $DEADMAN_SECONDS 秒已就位"
    ;;
  post)
    if nginx_running; then
      log "post: $NGINX_UNIT 已在运行（无需处理）"
    elif start_nginx; then
      log "post: 已把 $NGINX_UNIT 拉回"
    else
      log "post: ！重启 $NGINX_UNIT 失败 —— 站点可能已下线，请立刻人工介入"
      exit 1
    fi
    ;;
  deploy)
    # 续期成功后证书文件已换，但 nginx 内存里还是旧证书 —— 必须重载。
    if ! nginx_running; then
      log "deploy: $NGINX_UNIT 未在运行，跳过重载（等 post 拉起时会自然加载新证书）"
      exit 0
    fi
    if "$SYSTEMCTL" reload "$NGINX_UNIT" >/dev/null 2>&1; then
      log "deploy: 已 reload $NGINX_UNIT 以加载新证书"
    else
      log "deploy: reload 失败，退化为 restart"
      start_nginx || { log "deploy: restart 也失败，请立刻人工介入"; exit 1; }
    fi
    ;;
  *)
    echo "用法: $0 {pre|post|deploy}" >&2
    exit 2
    ;;
esac
