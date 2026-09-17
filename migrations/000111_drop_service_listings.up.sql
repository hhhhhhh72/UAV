-- 删除 service_listings（服务能力并入商城的收尾）。
--
-- 到这一步：数据已在 000110 迁进 drone_products，读路径全部走适配层
-- （service.ServiceListingService 读写商品，HTTP 形状不变）。
--
-- ⚠ 这是**不可逆**的一步，所以先断言再删：
--   如果 service_listings 里还有任何一行没出现在 drone_products 中（说明 000110 的
--   迁移不完整，或者之后又有人往旧表写过），直接 RAISE EXCEPTION 让整个迁移失败回滚，
--   而不是把唯一一份数据删掉。迁移运行器逐文件在事务里执行，失败即整体回滚。
DO $$
DECLARE
    missing INT;
BEGIN
    SELECT count(*) INTO missing
      FROM service_listings sl
     WHERE NOT EXISTS (SELECT 1 FROM drone_products p WHERE p.id = sl.id);

    IF missing > 0 THEN
        RAISE EXCEPTION 'service_listings 还有 % 行未迁入 drone_products，拒绝删除旧表', missing;
    END IF;
END $$;

-- 收藏关系同理：旧的 service_listing_favorites 每一行都应已在 product_favorites 里有对应。
DO $$
DECLARE
    missing INT;
BEGIN
    SELECT count(*) INTO missing
      FROM service_listing_favorites f
     WHERE NOT EXISTS (
        SELECT 1 FROM product_favorites pf
         WHERE pf.user_id = f.user_id AND pf.product_id = f.listing_id
     );

    IF missing > 0 THEN
        RAISE EXCEPTION 'service_listing_favorites 还有 % 行未迁入 product_favorites，拒绝删除旧表', missing;
    END IF;
END $$;

DROP TABLE IF EXISTS service_listing_favorites;
DROP TABLE IF EXISTS service_listings;
