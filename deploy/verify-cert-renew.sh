#!/bin/bash
# 证书续期驱动（deploy/cert-renew.sh）的回归演练。
#
# 为什么必须有它（2026-09-21）：cert-renew.sh 会在**无人值守**的凌晨**主动停掉 nginx**。
#   「该不该续期」判断错一次 → 每天停一次站；「续期失败后没收尾」错一次 → 站一直停着。
#   而它依赖 acme.sh / flock / openssl s_client / 钩子脚本四五个外部东西，
#   任何一环坏了都必须 **fail-closed**：宁可这次不续期，也绝不能把站点停在那儿。
#   这跟备份静默失败、告警假绿是同一类问题 —— 所以每条分支都要有断言。
#
# 全程用临时的证书 / 钩子 / acme 桩 + 一个临时端口的 openssl s_server：
#   不碰生产 nginx、不碰 /etc/nginx/certs、不碰真的 acme.sh。
set -uo pipefail

T=$(mktemp -d /tmp/cert-renew-drill.XXXXXX)
# 必须 export：桩脚本（start-served.sh / hooks.sh）在**子进程**里跑，看不到本 shell 的普通变量。
# 少了这一行，桩里的 $T 是空的 → python 去找 /tls-serve.py → 端点起不来，
# 六条用例一起假失败（第一版就是这样）。
export T
SERVED_PID=""
cleanup() { [ -f "$T/served.pid" ] && kill "$(cat "$T/served.pid")" 2>/dev/null; rm -rf "$T"; }
trap cleanup EXIT

pass=0; fail=0
ok()  { echo "  ok   $1"; pass=$((pass+1)); }
bad() { echo "  FAIL $1"; fail=$((fail+1)); }

RENEW=${RENEW:-/root/UAV/deploy/cert-renew.sh}
[ -f "$RENEW" ] || { echo "找不到 $RENEW"; exit 1; }

real_nginx_before=$(systemctl is-active nginx 2>/dev/null || echo unknown)
real_cert_before=$(sha256sum /etc/nginx/certs/api.cqnarc.cn.fullchain.crt 2>/dev/null | cut -d' ' -f1)

# ---- 临时端口上的"服务端证书" ----
PORT=$(( 18000 + RANDOM % 2000 ))
SERVED_CRT="$T/served.crt"; SERVED_KEY="$T/served.key"
CALLS="$T/calls"; : > "$CALLS"; export CALLS

mk_cert() { # mk_cert <crt> <key> <serial> <days>
  openssl req -x509 -newkey ec -pkeyopt ec_paramgen_curve:prime256v1 \
    -keyout "$2" -out "$1" -days "$4" -nodes -set_serial "$3" \
    -subj "/CN=drill-api" >/dev/null 2>&1
}
# 起/换"服务端证书"端点。必须是**先杀干净再绑**：s_server 退出前端口还占着，
# 直接再起一个会 bind 失败，而那会表现为"重启后 s_client 仍然读到旧证书"——
# 我第一版就是这样，于是"续期成功后服务端证书应已更换"这条断言假失败。
# 用 python 的 TLS 服务而不是 openssl s_server：s_server 默认**一次只服务一个连接**，
# 前一个 s_client 的会话没干净收尾时，下一个连接会一直挂着 —— 表现就是"隔一次读不到证书"。
# 这个坑在演练里真的踩到了（"服务端 serial=" 空），换成多连接的 python 服务后消失。
cat > "$T/tls-serve.py" <<'PY'
import socket, ssl, sys, threading
ctx = ssl.SSLContext(ssl.PROTOCOL_TLS_SERVER)
ctx.load_cert_chain(sys.argv[1], sys.argv[2])
srv = socket.socket()
srv.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
srv.bind(('127.0.0.1', int(sys.argv[3])))
srv.listen(32)
def handle(c):
    # 必须真的做 TLS 握手（wrap_socket server_side）—— 只对裸 socket recv 然后关掉，
    # 客户端握手会在中途收到 EOF（SSLEOFError），证书一张也拿不到。
    try:
        with ctx.wrap_socket(c, server_side=True) as ss:
            try:
                ss.recv(4096)
            except Exception:
                pass
    except Exception:
        pass
    finally:
        try: c.close()
        except Exception: pass
while True:
    try:
        c, _ = srv.accept()
    except OSError:
        break
    threading.Thread(target=handle, args=(c,), daemon=True).start()
PY

cat > "$T/start-served.sh" <<'STUB'
#!/bin/bash
# start-served.sh <crt> <key> <port> <pidfile>
crt="$1"; key="$2"; port="$3"; pidf="$4"
if [ -f "$pidf" ]; then
  old=$(cat "$pidf" 2>/dev/null)
  kill "$old" 2>/dev/null
  for _ in $(seq 1 30); do kill -0 "$old" 2>/dev/null || break; sleep 0.1; done
fi
python3 "$T/tls-serve.py" "$crt" "$key" "$port" >>"$T/py.out" 2>&1 &
echo $! > "$pidf"
# 就绪判据是"**能读到证书 serial**"，不是 s_client 的退出码：
# 服务端握手后立刻关连接时 s_client 会以 1 退出（unexpected eof），但握手其实成功了。
# 第一版按退出码判断，于是端点明明起来了却一律判"起不来"，六条用例一起假失败。
for _ in $(seq 1 25); do
  got=$(echo | timeout 2 openssl s_client -connect "127.0.0.1:$port" 2>/dev/null | openssl x509 -noout -serial 2>/dev/null)
  [ -n "$got" ] && exit 0
  sleep 0.2
done
exit 1
STUB
chmod +x "$T/start-served.sh"

start_served() { bash "$T/start-served.sh" "$SERVED_CRT" "$SERVED_KEY" "$PORT" "$T/served.pid"; }

mk_cert "$T/certA90.crt"  "$T/certA90.key"  1001 90   # "已续期、还很新"
mk_cert "$T/certB_new.crt" "$T/certB_new.key" 2002 90 # 桩 acme 换上的新证书
mk_cert "$T/certOld10.crt" "$T/certOld10.key" 1003 10 # "快到期、该续了"

# ---- 桩：钩子 ----
cat > "$T/hooks.sh" <<'STUB'
#!/bin/bash
echo "$1" >> "$CALLS"
case "$1" in
  deploy)
    # 模拟"证书文件已更新 → nginx 重新加载后开始服务新证书"。
    # **证书和私钥都要拷**：只换证书不换私钥，TLS 服务会以 KEY_VALUES_MISMATCH 起不来 ——
    # 第一版就漏了私钥，表现为"续期成功后端点再也读不到证书"，排查了两轮才定位。
    cp "$INSTALLED" "$SERVED_CRT"
    cp "${INSTALLED%.crt}.key" "$SERVED_KEY"
    bash "$T/start-served.sh" "$SERVED_CRT" "$SERVED_KEY" "$PORT" "$T/served.pid"
    echo "deploy: start-served 退出码=$? serial=$(echo | timeout 2 openssl s_client -connect "127.0.0.1:$PORT" 2>/dev/null | openssl x509 -noout -serial 2>/dev/null)" >> "$T/diag"
    ;;
esac
exit 0
STUB
chmod +x "$T/hooks.sh"

# pre 失败用的钩子：模拟"443 没让出来"
printf '#!/bin/bash\nexit 1\n' > "$T/failing-hooks.sh"; chmod +x "$T/failing-hooks.sh"

# ---- 桩：acme.sh ----
cat > "$T/fake-acme.sh" <<'STUB'
#!/bin/bash
mode="${FAKE_ACME_MODE:-ok}"
fullchain=""; keyfile=""; reloadcmd=""
while [ $# -gt 0 ]; do
  case "$1" in
    --fullchain-file) fullchain="$2"; shift 2 ;;
    --key-file)       keyfile="$2";   shift 2 ;;
    --reloadcmd)      reloadcmd="$2"; shift 2 ;;
    *) shift ;;
  esac
done
echo "acme($mode)" >> "$CALLS"
[ "$mode" = fail ] && exit 1
cp "$FAKE_ACME_CERT" "$fullchain" || exit 1
[ -n "$keyfile" ] && { cp "$FAKE_ACME_KEY" "$keyfile" || exit 1; }
[ -n "$reloadcmd" ] && eval "$reloadcmd"
exit 0
STUB
chmod +x "$T/fake-acme.sh"
# 桩也要过哈希校验这关：acme.sh 的真实哈希是钉死在 cert-renew.sh 里的，
# 而 ${VAR:-默认值} 对**空串**同样会套默认值 —— 传 ACME_SHA256= 是关不掉校验的
# （第一版就是这么错的，于是前三条用例全被"哈希不符"提前挡掉）。
STUB_ACME_SHA=$(sha256sum "$T/fake-acme.sh" | cut -d' ' -f1)

run_renew() { # run_renew <已安装证书> <served 证书> [额外 env...]
  local installed="$1" served="$2"; shift 2
  # 两个位置都要放："已安装的证书文件"（cert-renew 读它判断还剩几天）与"服务端正在服务的证书"。
  # 只更新其中一个会让下一条用例从错误的起点出发（⑥ 就是这么假通过的：上一轮续期把
  # installed.crt 换成了 90 天的新证书，于是"并发保护"那条根本没走到锁，而是走了"未到期"）。
  cp "$installed" "$SERVED_CRT";  cp "${installed%.crt}.key" "$SERVED_KEY"
  cp "$installed" "$T/installed.crt"; cp "${installed%.crt}.key" "$T/installed.key"
  start_served || { echo "   （临时 TLS 端点起不来，跳过）"; return 3; }
  # 日志要清空：它是**追加**写的，不清的话后一条用例的 grep 会匹配到前一条留下的行，
  # 断言就变成了"历史上出现过"而不是"这一次发生了"。
  : > "$CALLS"; : > "$T/renew.log"
  env "CERT=$T/installed.crt" "KEY=$T/installed.key" \
      "CERT_HOOKS=$T/hooks.sh" "ACME_SH=$T/fake-acme.sh" "ACME_SHA256=$STUB_ACME_SHA" \
      "ACME_HOME=$T/acme-home" "CERT_RENEW_LOG=$T/renew.log" \
      "CERT_RENEW_LOCK=$T/lock" "CERT_HOST=127.0.0.1" "CERT_PORT=$PORT" \
      "INSTALLED=$T/installed.crt" "SERVED_CRT=$SERVED_CRT" "SERVED_KEY=$SERVED_KEY" \
      "T=$T" "PORT=$PORT" "FAKE_ACME_CERT=$T/certB_new.crt" "FAKE_ACME_KEY=$T/certB_new.key" \
      "$@" bash "$RENEW"
}

echo "演练目录：$T"
echo "临时 TLS 端口：$PORT"
echo "生产 nginx 状态（真实 systemd）：$real_nginx_before"
echo
echo "== 续期判定演练 =="

# ① 没到期：**一次都不该碰 nginx**（这是"每天跑一次"能成立的前提）
cp "$T/certA90.crt" "$T/installed.crt"; cp "$T/certA90.key" "$T/installed.key"
run_renew "$T/certA90.crt" "$T/certA90.crt"; rc=$?
if [ "$rc" = "0" ] && [ ! -s "$CALLS" ] && grep -q '不续期' "$T/renew.log"; then
  ok "还有 90 天时直接退出，钩子/acme 一次都没被调用"
else
  bad "未到期路径异常：rc=$rc 调用=$(tr '\n' '|' < "$CALLS")"
fi

# ② 到期 + 续期成功：顺序必须是 pre → acme(deploy) → post，且**服务端确实换上了新证书**
cp "$T/certOld10.crt" "$T/installed.crt"; cp "$T/certOld10.key" "$T/installed.key"
run_renew "$T/certOld10.crt" "$T/certOld10.crt"; rc=$?
calls=$(tr '\n' ' ' < "$CALLS")
served_now=$(echo | timeout 5 openssl s_client -connect "127.0.0.1:$PORT" 2>/dev/null | openssl x509 -noout -serial 2>/dev/null | cut -d= -f2)
# 注意：openssl x509 -serial 打的是**十六进制**（2002 → 07D2），别拿十进制去比
if [ "$rc" = "0" ] && [ "$calls" = "pre acme(ok) deploy post " ] && [ "$served_now" = "07D2" ] \
   && grep -q '续期成功' "$T/renew.log"; then
  ok "到期时按 pre→acme→deploy→post 走完，服务端证书已从 03EB 换成 $served_now"
else
  bad "续期成功路径异常：rc=$rc 调用=[$calls] 服务端 serial=$served_now"
fi

# ②b 续期结束后，单实例锁必须已经释放。
#     子进程会继承 flock 的 fd —— acme.sh 会拉起 --reloadcmd 钩子、钩子又拉起别的进程，
#     任何一环把 fd 9 留在手里，锁就一直不放，之后每次续期都会被判成"已有进程在跑"而
#     静默跳过（生产上就是"设置了但从来没在跑"）。所以这里显式钉住它。
if ( exec 9>"$T/lock"; flock -n 9 ); then
  ok "续期结束后单实例锁已释放（子进程没有继承 fd 9）"
else
  bad "续期结束后锁仍被占用 —— 下一次续期会被静默跳过"
fi

# ③ 续期失败：必须**仍然把 nginx 拉回来**，证书保持原样
cp "$T/certOld10.crt" "$T/installed.crt"; cp "$T/certOld10.key" "$T/installed.key"
run_renew "$T/certOld10.crt" "$T/certOld10.crt" FAKE_ACME_MODE=fail; rc=$?
calls=$(tr '\n' ' ' < "$CALLS")
if [ "$rc" != "0" ] && printf '%s' "$calls" | grep -q 'post' && grep -q '续期失败' "$T/renew.log"; then
  ok "acme 失败时仍执行 post（nginx 被拉回），退出码非 0"
else
  bad "失败路径没兜住：rc=$rc 调用=[$calls]"
fi

# ④ acme.sh 不存在：fail-closed，且**不能**先停 nginx 再发现工具没了
cp "$T/certOld10.crt" "$T/installed.crt"; cp "$T/certOld10.key" "$T/installed.key"
run_renew "$T/certOld10.crt" "$T/certOld10.crt" ACME_SH="$T/no-such-acme.sh"; rc=$?
if [ "$rc" != "0" ] && [ ! -s "$CALLS" ]; then
  ok "缺 acme.sh 时 fail-closed，且没有动过 nginx"
else
  bad "缺工具时仍动了 nginx：rc=$rc 调用=$(tr '\n' '|' < "$CALLS")"
fi

# ⑤ acme.sh 哈希不符：同样 fail-closed（它以 root 运行，且是从网上取的单个 shell 脚本）
run_renew "$T/certOld10.crt" "$T/certOld10.crt" ACME_SHA256=deadbeef; rc=$?
if [ "$rc" != "0" ] && [ ! -s "$CALLS" ] && grep -q '哈希与钉死值不一致' "$T/renew.log"; then
  ok "acme.sh 哈希不符时拒绝执行，且没有动过 nginx"
else
  bad "哈希校验没生效：rc=$rc 调用=$(tr '\n' '|' < "$CALLS")"
fi
run_renew "$T/certOld10.crt" "$T/certOld10.crt" ACME_SHA256=deadbeef ALLOW_ACME_HASH_MISMATCH=1; rc=$?
if [ "$rc" = "0" ] && printf '%s' "$(tr '\n' ' ' < "$CALLS")" | grep -q 'acme(ok)'; then
  ok "显式置 ALLOW_ACME_HASH_MISMATCH=1 后可继续（有意放行的通道）"
else
  bad "放行开关无效：rc=$rc 调用=$(tr '\n' '|' < "$CALLS")"
fi

# ⑥ 已有续期在跑：不能被第二个进程打断（两个进程抢 443 会双双失败）
( exec 9>"$T/lock"; flock -n 9; sleep 3 ) &
HOLDER=$!
sleep 0.5
run_renew "$T/certOld10.crt" "$T/certOld10.crt"; rc=$?
kill "$HOLDER" 2>/dev/null
if [ "$rc" = "0" ] && [ ! -s "$CALLS" ] && grep -q '已有一个续期进程在跑' "$T/renew.log"; then
  ok "锁被占用时静默跳过，不抢 443"
else
  bad "并发保护失效：rc=$rc 调用=$(tr '\n' '|' < "$CALLS")"
fi

# ⑦ pre 钩子失败（443 没让出来）：必须中止，不能带着 nginx 继续跑 acme
#    注意先把"已安装证书"重置回快到期那份 —— ⑤b 那次成功续期会把 installed.crt
#    换成 90 天的新证书，不重置的话这里会走"未到期"分支，断言看着通过、其实没测到。
cp "$T/certOld10.crt" "$T/installed.crt"; cp "$T/certOld10.key" "$T/installed.key"
: > "$CALLS"; : > "$T/renew.log"
env "CERT=$T/installed.crt" "KEY=$T/installed.key" "ACME_SH=$T/fake-acme.sh" "ACME_SHA256=" \
    "CERT_HOOKS=$T/failing-hooks.sh" "ACME_SHA256=$STUB_ACME_SHA" "CERT_RENEW_LOG=$T/renew.log" "CERT_RENEW_LOCK=$T/lock" \
    "CERT_HOST=127.0.0.1" "CERT_PORT=$PORT" bash "$RENEW"; rc=$?
if [ "$rc" != "0" ] && ! grep -q 'acme(' "$CALLS"; then
  ok "pre 失败时中止续期，没有在 nginx 还占着 443 的情况下硬跑 acme"
else
  bad "pre 失败后仍继续：rc=$rc 调用=$(tr '\n' '|' < "$CALLS")"
fi

echo
echo "== 结论：$pass 项通过，$fail 项失败 =="

real_nginx_after=$(systemctl is-active nginx 2>/dev/null || echo unknown)
real_cert_after=$(sha256sum /etc/nginx/certs/api.cqnarc.cn.fullchain.crt 2>/dev/null | cut -d' ' -f1)
if [ "$real_nginx_before" = "$real_nginx_after" ] && [ "$real_cert_before" = "$real_cert_after" ]; then
  echo "生产 nginx 与生产证书均未被本次演练影响（$real_nginx_before，证书哈希未变）ok"
else
  echo "FAIL 生产环境被改动：nginx $real_nginx_before→$real_nginx_after，证书哈希变了"
  fail=$((fail+1))
fi
[ "$fail" = 0 ] || exit 1
