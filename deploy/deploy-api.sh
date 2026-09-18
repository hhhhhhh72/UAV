#!/bin/bash
# 后端发布：解包 → 归一化行尾 → 构建 → 重启 → 健康检查 → **自清理**。
#
# 为什么要有这个脚本（2026-09-18）：此前每次发布都是临时拼一串命令，
# 于是 tar 包、解包目录一路堆在 /tmp；加上反复 docker compose build 攒下的构建缓存，
# 磁盘从 26% 涨到 34%。保洁脚本是周任务、又没覆盖 /tmp，等于没人收拾。
# 把它固化成脚本：发布流程自带清理，保洁只做兜底。
#
# 用法（服务器上，包已 scp 到 /tmp）：
#   bash deploy/deploy-api.sh /tmp/uav-xxx.tar.gz
set -euo pipefail

PKG=${1:?用法: deploy-api.sh <tar.gz 路径>}
[ -f "$PKG" ] || { echo "✗ 找不到发布包：$PKG"; exit 1; }
cd /root/UAV

before_schema=$(docker exec uav-db-1 psql -U drone -d drone_platform -t -A -c 'SELECT max(version) FROM schema_migrations' 2>/dev/null | tr -d '[:space:]')
before_disk=$(df -h / | awk 'NR==2{print $5}')
echo "== 解包（发布前 schema=$before_schema，磁盘 $before_disk）=="
tar -xzf "$PKG"
# Windows 打包会带 CRLF：*.go 带 \r 编译报错，*.sh 带 \r 会被 set -euo pipefail 读成 pipefail\r
find cmd internal -name '*.go' -print0 | xargs -0 -r sed -i 's/\r$//'
# Windows 打包还会丢可执行位，这一步不能省
chmod +x deploy/*.sh *.sh 2>/dev/null || true

echo "== 构建 =="
docker compose build api 2>&1 | tail -2
echo "== 重启 =="
docker compose up -d api 2>&1 | tail -2

echo "== 健康检查 =="
ok=false
for i in $(seq 1 40); do
  if [ "$(curl -s -o /dev/null -w '%{http_code}' http://127.0.0.1:8080/healthz || echo 000)" = "200" ]; then
    echo "  healthz 200（第 $i 次）"
    ok=true
    break
  fi
  sleep 2
done
[ "$ok" = true ] || { echo "  ✗ 健康检查未通过；发布包**保留**在 $PKG 供排查"; exit 1; }

after_schema=$(docker exec uav-db-1 psql -U drone -d drone_platform -t -A -c 'SELECT max(version) FROM schema_migrations' 2>/dev/null | tr -d '[:space:]')
echo "  schema: $before_schema -> $after_schema"
echo "  容器重启次数: $(docker inspect uav-api-1 --format '{{.RestartCount}}')"
echo "  启动期异常: $(docker logs uav-api-1 2>&1 | grep -icE 'panic|level=ERROR' || true) 条"

# 自清理：包用完即删。这是磁盘不再被部署残留吃掉的关键一步。
# 只在健康检查通过后才删——失败时留证据。
rm -f "$PKG" /tmp/admin-dist.tar.gz
echo "== 完成（发布包已删除，磁盘 $(df -h / | awk 'NR==2{print $5}') ）=="