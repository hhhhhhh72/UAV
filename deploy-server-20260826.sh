#!/bin/bash
# 2026-08-26 字段契约修复部署：重建 api 容器（迁移 000088-000094 随启动执行）+ 验证
set -e
cd "$(dirname "$0")"

echo "== 1/4 备份 .env（若有）=="
if [ -f .env ]; then
  cp .env ".env.bak.$(date +%Y%m%d_%H%M%S)"
  echo "backed up .env"
else
  echo "WARN: no .env in deploy dir -- compose build will fail; copy from existing deployment dir"
fi

echo "== 2/4 重建 api 容器（迁移随启动执行，均幂等）=="
sudo docker compose up -d --build api

echo "== 3/4 等待健康检查 =="
ok=0
for i in $(seq 1 20); do
  code=$(curl -s -o /dev/null -w '%{http_code}' http://127.0.0.1:8080/healthz || true)
  if [ "$code" = "200" ]; then echo "healthz OK"; ok=1; break; fi
  sleep 3
done
[ "$ok" = "1" ] || { echo "ERROR: api not healthy"; sudo docker compose logs --tail 50 api; exit 1; }

echo "== 4/4 迁移与内核验证 =="
echo "-- achievements stats columns (want views/favs):"
sudo docker exec uav-db-1 psql -U drone -d drone_platform -t -c "SELECT column_name FROM information_schema.columns WHERE table_name='achievements' AND column_name IN ('views','favs') ORDER BY 1"
echo "-- bookings time_slots (want time_slots):"
sudo docker exec uav-db-1 psql -U drone -d drone_platform -t -c "SELECT column_name FROM information_schema.columns WHERE table_name='test_site_bookings' AND column_name='time_slots'"
echo "-- claims table (want rd_challenge_claims):"
sudo docker exec uav-db-1 psql -U drone -d drone_platform -t -c "SELECT to_regclass('public.rd_challenge_claims')"
echo "-- test_sites params (want 6):"
sudo docker exec uav-db-1 psql -U drone -d drone_platform -t -c "SELECT count(*) FROM information_schema.columns WHERE table_name='test_sites' AND column_name IN ('airspace_range','max_takeoff_weight','runway_length','max_flight_height','compatible_models','image_url')"

echo "== 公开接口冒烟 =="
for u in /healthz /api/v1/achievements /api/v1/portfolios /api/v1/portfolios/featured /api/v1/test-sites /api/v1/challenges; do
  echo -n "$u => "
  curl -s -o /dev/null -w '%{http_code}\n' "http://127.0.0.1:8080$u"
done

echo "== 容器状态 =="
docker ps --format '{{.Names}}  {{.Status}}'
echo "== DONE =="
