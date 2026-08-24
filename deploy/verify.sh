#!/bin/bash
# 上线后验证：容器/迁移/环境变量/健康检查
set -e
echo "== containers =="
sudo docker ps | grep uav- || true
echo "== migrations log =="
sudo docker logs uav-api-1 2>&1 | grep -iE "migration|panic|fatal" | tail -8 || true
echo "== schema version =="
sudo docker exec -i uav-db-1 psql -U drone -d drone_platform -t -c "SELECT max(version) FROM schema_migrations;"
echo "== api env keys =="
sudo docker exec uav-api-1 printenv ENV SIGNING_SECRET WECHAT_APPID | wc -c
echo "== healthz =="
curl -s https://api.cqnarc.cn/healthz
echo
echo "== index check =="
curl -s -o /dev/null -w 'demands:%{http_code} ' https://api.cqnarc.cn/api/v1/demands
curl -s -o /dev/null -w 'admin-token:%{http_code}\n' https://api.cqnarc.cn/api/v1/admin/token
echo "== endpoints spot check =="
declare -A eps=(
  ["pilots"]="https://api.cqnarc.cn/api/v1/certified-pilots?page=1&page_size=2"
  ["instructors"]="https://api.cqnarc.cn/api/v1/instructors"
  ["my-orders"]="https://api.cqnarc.cn/api/v1/trade-orders/mine"
  ["admin-orders"]="https://api.cqnarc.cn/api/v1/admin/orders"
)
for k in pilots instructors my-orders admin-orders; do
  code=$(curl -s -o /dev/null -w '%{http_code}' "${eps[$k]}")
  echo "$k:$code"
done
