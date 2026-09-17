-- 回退第一步：移除迁移进来的收藏与商品，再删三个新增列。
-- 旧表 service_listings 未被本迁移改动，无需恢复。
DELETE FROM product_favorites WHERE id LIKE 'pfav-sl-%';
-- 只删"确实由 service_listings 迁来的行"：id 相同且三个服务字段逐字一致。
-- （ON CONFLICT DO NOTHING 保证既有商品不会被覆盖，因此这个条件不会误删既有商品。）
DELETE FROM drone_products p
USING service_listings sl
WHERE p.id = sl.id
  AND p.category = sl.category
  AND p.region = sl.region
  AND p.unit = sl.unit;
DROP INDEX IF EXISTS idx_drone_products_category;
ALTER TABLE drone_products DROP COLUMN IF EXISTS unit;
ALTER TABLE drone_products DROP COLUMN IF EXISTS region;
ALTER TABLE drone_products DROP COLUMN IF EXISTS category;
