#!/bin/bash
# 2026-09-15 商品模块 + 服务能力合并部署
# 迁移 000106–000111 随 api 启动执行（runner 在监听之前跑，失败即 exit 1，不会带病启动）
#
# ⚠ 本批次含**不可逆**迁移 000111（DROP TABLE service_listings），
#   所以第 1 步强制备份数据库 —— 不要跳过。
set -e
cd "$(dirname "$0")"

PG_CONTAINER=${PG_CONTAINER:-uav-db-1}
API_CONTAINER=${API_CONTAINER:-uav-api-1}
PG_USER=${PG_USER:-drone}
PG_DB=${PG_DB:-drone_platform}
STAMP=$(date +%Y%m%d_%H%M%S)
BACKUP=backup_before_$STAMP.sql

echo "== 0/6 部署前检查：确认 .env 存在 =="
if [ ! -f .env ]; then
  echo "ERROR: 部署目录没有 .env —— 打包时被漏掉或误删。"
  echo "       历史事故：tar 未排除 .env 导致本地空值覆盖生产，WECHAT_APPID 丢失。"
  echo "       请先从旧部署目录恢复 .env 再执行本脚本。"
  exit 1
fi
grep -q 'WECHAT_APPID' .env && echo '  .env 含 WECHAT_APPID OK' || echo '  WARN: .env 里没有 WECHAT_APPID，微信登录会退化'

echo "== 1/6 备份数据库（000111 会删表，必须先备份）=="
sudo docker exec "$PG_CONTAINER" pg_dump -U "$PG_USER" -d "$PG_DB" > "$BACKUP"
echo "  已备份到 $BACKUP ($(wc -c < "$BACKUP") 字节)"

echo "== 2/6 备份 .env =="
cp .env ".env.bak.$STAMP"
echo "  已备份到 .env.bak.$STAMP"

echo "== 3/6 记录迁移前的关键表状态（便于对比）=="
sudo docker exec "$PG_CONTAINER" psql -U "$PG_USER" -d "$PG_DB" -t -c \
  "SELECT 'service_listings=' || count(*) FROM service_listings" 2>/dev/null || echo "  service_listings 不存在（可能已迁移过）"
sudo docker exec "$PG_CONTAINER" psql -U "$PG_USER" -d "$PG_DB" -t -c \
  "SELECT 'drone_products=' || count(*) FROM drone_products"

echo "== 4/6 重建 api 容器（迁移随启动执行）=="
sudo docker compose up -d --build api

echo "== 5/6 等待健康检查 =="
ok=0
for i in $(seq 1 30); do
  code=$(curl -s -o /dev/null -w '%{http_code}' http://127.0.0.1:8080/healthz || true)
  if [ "$code" = "200" ]; then echo "  healthz OK"; ok=1; break; fi
  sleep 3
done
if [ "$ok" != "1" ]; then
  echo "ERROR: api 未就绪。若迁移失败，日志里会有 migrations failed。"
  sudo docker compose logs --tail 80 api
  echo "--- 回滚：psql -U $PG_USER -d $PG_DB < $BACKUP ---"
  exit 1
fi

echo "== 6/6 本批次迁移验证 =="
q() { sudo docker exec "$PG_CONTAINER" psql -U "$PG_USER" -d "$PG_DB" -t -c "$1"; }

echo "-- 000106 商品审核维度（want 4）"
q "SELECT count(*) FROM information_schema.columns WHERE table_name='drone_products' AND column_name IN ('check_status','check_reason','reviewed_at','reviewed_by')"
echo "   未回填的行（want 0）:"
q "SELECT count(*) FROM drone_products WHERE COALESCE(check_status,'')=''"

echo "-- 000107/000109 价格模式与交付方式（want 2）"
q "SELECT count(*) FROM information_schema.columns WHERE table_name='drone_products' AND column_name IN ('price_mode','delivery')"

echo "-- 000108 订单收货/发货字段（want 7）"
q "SELECT count(*) FROM information_schema.columns WHERE table_name='trade_orders' AND column_name IN ('receiver_name','receiver_phone','receiver_region','receiver_address','shipping_company','shipping_tracking','shipped_at')"

echo "-- 000110 服务类字段（want 3）+ 迁入的服务类商品数（want > 0）"
q "SELECT count(*) FROM information_schema.columns WHERE table_name='drone_products' AND column_name IN ('category','region','unit')"
q "SELECT count(*) FROM drone_products WHERE prod_type IN ('repair','aerial','test_fly','calibration','airspace')"

echo "-- 000111 旧表已删除（want 两行都输出 NULL）"
q "SELECT COALESCE(to_regclass('public.service_listings')::text,'NULL')"
q "SELECT COALESCE(to_regclass('public.service_listing_favorites')::text,'NULL')"

echo "-- 已应用的迁移版本（want 111）"
q "SELECT max(version) FROM schema_migrations"

echo "== 公开接口冒烟 =="
for u in /healthz /api/v1/products /api/v1/service-listings /api/v1/demands /api/v1/training-courses; do
  echo -n "  $u => "
  curl -s -o /dev/null -w '%{http_code}\n' "http://127.0.0.1:8080$u"
done
echo "  （/api/v1/service-listings 现在读商品表，应 200 且内容与服务类商品一致）"

echo "== 容器状态 =="
sudo docker ps --format '{{.Names}}  {{.Status}}'
echo "== DONE =="
echo ""
echo "人工复核：小程序「我的发布」「需求大厅-服务能力分段」「服务详情」「收藏」各点一遍。"
echo "确认无异常后可删除备份：$BACKUP"