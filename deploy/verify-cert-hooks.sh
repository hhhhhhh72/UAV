#!/bin/bash
# 证书续期钩子（deploy/cert-hooks.sh）的回归演练。
#
# 为什么必须有它（2026-09-21）：这些钩子会在**无人值守**的续期过程中**主动停掉 nginx**。
#   写错一个分支的代价不是"续期失败"，而是"网站下线且没人知道" —— 与备份静默失败、
#   告警假绿是同一类问题。所以停服务之前先把"停→兜底→拉起"的每条路径都验一遍。
#
# 全程用**假的 systemctl**（$SYSTEMCTL 指向临时目录里的桩），不碰生产 nginx：
#   桩把状态记在文件里，脚本只读那个文件断言，生产 systemd 一次都不会被调用。
set -uo pipefail

T=$(mktemp -d /tmp/cert-hooks-drill.XXXXXX)
trap 'rm -rf "$T"' EXIT

pass=0; fail=0
ok()  { echo "  ok   $1"; pass=$((pass+1)); }
bad() { echo "  FAIL $1"; fail=$((fail+1)); }

HOOKS=${HOOKS:-/root/UAV/deploy/cert-hooks.sh}
[ -f "$HOOKS" ] || { echo "找不到 $HOOKS"; exit 1; }

# 断言"生产 nginx 没被动过"：演练前后各取一次真实 systemd 的状态
real_before=$(systemctl is-active nginx 2>/dev/null || echo unknown)

# ---- 假 systemctl ----
STATE="$T/state"; CALLS="$T/calls"; : > "$CALLS"; echo active > "$STATE"
# 必须 export：桩是**另一个进程**，看不到本 shell 的普通变量。
# （第一版漏了这两行，桩把状态写进了空路径，5 项断言因此"通过"或"失败"得毫无意义 ——
#   演练脚本自己先被演练了一遍。）
export STATE CALLS
cat > "$T/systemctl" <<'STUB'
#!/bin/bash
# 桩：is-active / start / stop / reload 四个动作，状态落在 $STATE，动作序列落在 $CALLS
act="$1"; shift
while [ "${1:-}" = "--quiet" ]; do shift; done
unit="${1:-}"
echo "$act $unit" >> "$CALLS"
case "$act" in
  is-active) [ "$(cat "$STATE" 2>/dev/null)" = active ] && exit 0 || exit 3 ;;
  start)     echo active  > "$STATE"; exit 0 ;;
  stop)      [ "${FAKE_STOP_FAIL:-0}" = "1" ] && exit 1
             echo stopped > "$STATE"; exit 0 ;;
  reload)    [ "$(cat "$STATE" 2>/dev/null)" = active ] || exit 1
             echo active  > "$STATE"; exit 0 ;;
  *) exit 0 ;;
esac
STUB
chmod +x "$T/systemctl"

hook() { # hook <logfile> <subcommand>
  CERT_HOOK_LOG="$1" SYSTEMCTL="$T/systemctl" NGINX_UNIT=nginx \
    DEADMAN_SECONDS="${DEADMAN_SECONDS_OVERRIDE:-2}" FAKE_STOP_FAIL="${FAKE_STOP_FAIL:-0}" \
    bash "$HOOKS" "$2"
}
state() { cat "$STATE" 2>/dev/null || echo missing; }

echo "演练目录：$T"
echo "生产 nginx 状态（真实 systemd）：$real_before"
echo
echo "== 续期钩子演练 =="

# ① 正常路径：pre 停服务 → post 拉回来
echo active > "$STATE"
hook "$T/1.log" pre
if [ "$(state)" = stopped ] && grep -q 'pre: 已停' "$T/1.log"; then
  ok "pre 停掉了 nginx 并留下日志"
else
  bad "pre 异常：状态=$(state) 日志=$(tr '\n' '|' < "$T/1.log")"
fi
hook "$T/1.log" post
if [ "$(state)" = active ] && grep -q 'post: 已把' "$T/1.log"; then
  ok "post 把 nginx 拉回来了"
else
  bad "post 异常：状态=$(state)"
fi

# ② 死人开关：pre 之后**任何钩子都不跑**（模拟 certbot 被 SIGKILL），
#    nginx 仍必须在 DEADMAN_SECONDS 后自己回来 —— 这是整条链路的最后一道保险
echo active > "$STATE"
DEADMAN_SECONDS_OVERRIDE=1 hook "$T/2.log" pre
[ "$(state)" = stopped ] || bad "死人开关前置条件不成立（nginx 没停）"
sleep 3
if [ "$(state)" = active ] && grep -q 'deadman:' "$T/2.log"; then
  ok "certbot 中途死掉、钩子不再执行时，死人开关把 nginx 拉了回来"
else
  bad "死人开关没生效：状态=$(state) 日志=$(tr '\n' '|' < "$T/2.log")"
fi

# ③ post 幂等：已经在跑就不重复 start
echo active > "$STATE"; : > "$CALLS"
hook "$T/3.log" post
if [ "$(state)" = active ] && ! grep -q 'start nginx' "$CALLS" && grep -q '已在运行' "$T/3.log"; then
  ok "nginx 已在运行时 post 不重复 start（幂等）"
else
  bad "post 不幂等：调用=$(tr '\n' '|' < "$CALLS")"
fi

# ④ pre 停不掉时必须**放弃续期**，而不是继续跑（站点保持在线优先于换证书）
echo active > "$STATE"
FAKE_STOP_FAIL=1 hook "$T/4.log" pre; rc=$?
if [ "$rc" != "0" ] && [ "$(state)" = active ] && grep -q '放弃本次续期' "$T/4.log"; then
  ok "停服务失败时中止本次续期，站点保持在线"
else
  bad "停服务失败却继续：rc=$rc 状态=$(state)"
fi

# ⑤ deploy 重载（证书文件换了但 nginx 内存里还是旧的 —— 不 reload 就是假绿）
echo active > "$STATE"; : > "$CALLS"
hook "$T/5.log" deploy
if grep -q 'reload nginx' "$CALLS" && grep -q 'deploy: 已 reload' "$T/5.log"; then
  ok "deploy 触发了 reload 让新证书生效"
else
  bad "deploy 没有 reload：调用=$(tr '\n' '|' < "$CALLS")"
fi

# ⑥ deploy 时 nginx 没在跑：不该报错（post 拉起时会自然加载新证书）
echo stopped > "$STATE"; : > "$CALLS"
hook "$T/6.log" deploy; rc=$?
if [ "$rc" = "0" ] && grep -q '跳过重载' "$T/6.log" && ! grep -q 'reload nginx' "$CALLS"; then
  ok "nginx 未运行时 deploy 安全跳过"
else
  bad "deploy 未运行时行为异常：rc=$rc"
fi

# ⑦ 未知子命令必须报用法并返回 2（避免 certbot 配错钩子时静默什么都不做）
bash "$HOOKS" bogus >/dev/null 2>&1; rc=$?
[ "$rc" = "2" ] && ok "未知子命令返回 2 并打印用法" || bad "未知子命令返回 $rc（应为 2）"

# ⑧ 死人开关**不能**继承续期锁（cert-renew.sh 用 flock fd 9 做单实例）。
#    第一版就是这样错的：deploy/cert-hooks.sh pre 里的 setsid 子进程把 fd 9 一起继承了，
#    于是续期结束后 5 分钟内再跑一次都会被判成"已有续期进程在跑"而静默跳过 ——
#    实测出来的（当天那次人工复核直接被跳过，日志里是"已有一个续期进程在跑"）。
#    修法是给死人开关显式关掉 fd 9，这条断言把它钉住。
LOCK="$T/lock"
(
  exec 9>"$LOCK"
  flock -n 9 || exit 1
  hook "$T/8.log" pre
) >/dev/null 2>&1
sleep 0.3
if ( exec 9>"$LOCK"; flock -n 9 ); then
  ok "死人开关没有继承续期锁（fd 9 已显式关闭）"
else
  bad "续期锁仍被死人开关占着 —— 续期后的一段时间内会静默跳过"
fi

echo
echo "== 结论：$pass 项通过，$fail 项失败 =="
real_after=$(systemctl is-active nginx 2>/dev/null || echo unknown)
if [ "$real_before" = "$real_after" ]; then
  echo "生产 nginx 未被本次演练影响（$real_before → $real_after）ok"
else
  echo "FAIL 生产 nginx 状态被改动：$real_before → $real_after"; fail=$((fail+1))
fi
[ "$fail" = 0 ] || exit 1
