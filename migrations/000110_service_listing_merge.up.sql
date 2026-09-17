-- 服务能力并入商城商品（第一步：只做加法）。
--
-- 背景：同一个"航拍服务"，卖家可以走两条路发布——
--   service_listings  企业供给能力展示，无价格闭环、无下单
--   drone_products    可下单可交易（prod_type=aerial 等 4 类服务），走托管金结算
-- 两条路字段不同、状态机不同、前端入口不同，用户在「发布」页看到两个入口做同一件事。
--
-- 本迁移**只做加法**：给商品补上服务能力特有的三个字段，并把 service_listings 的行迁成商品。
-- **不 drop 旧表** —— 下一步会先把 /api/v1/service-listings 系列端点改成读商品
-- （API 形状保持不变，前端零改动即可切换），验证无回归后由后续迁移删除旧表。
-- 这样任何一步都能单独回退。

-- 1) 商品补三个服务类字段（纯加列，不改动任何已有行的语义）
ALTER TABLE drone_products ADD COLUMN IF NOT EXISTS category TEXT NOT NULL DEFAULT '';
ALTER TABLE drone_products ADD COLUMN IF NOT EXISTS region   TEXT NOT NULL DEFAULT '';
ALTER TABLE drone_products ADD COLUMN IF NOT EXISTS unit     TEXT NOT NULL DEFAULT '';

-- 2) service_listings → drone_products
--
-- 映射说明：
--   id           保持原值（不改前缀）——收藏关系靠它对应，改了就断
--   seller_id    provider_id；为空时落 'platform'（与管理端建商品的平台卖家 ID 一致），
--                否则订单的 seller_id 会变空串，归属与售后全断
--   prod_type    由 category 推断；推断不出时归 'repair'（服务类兜底，绝不会误判成实物）
--   price_mode   price_fen=0 → negotiable（旧表注释写明"0 为面议"）
--   delivery     'negotiable'（服务不涉及寄送；也保证下单时不强制收货地址）
--   images       单图 → 数组；无图则空数组
--   status       published → listed / 其余 → removed
--   check_status 'passed'（这些行此前就是公开可见的，说明已通过审核）
--   ON CONFLICT DO NOTHING：万一 id 与既有商品撞车，保留既有商品，不覆盖
INSERT INTO drone_products (
    id, seller_id, seller_name, prod_type, title, description,
    price_fen, price_mode, delivery, images, brand, model, condition,
    views, status, check_status, check_reason, reviewed_at, reviewed_by,
    category, region, unit, version, created_at, updated_at
)
SELECT
    sl.id,
    COALESCE(NULLIF(sl.provider_id, ''), 'platform'),
    COALESCE(NULLIF(sl.provider_name, ''), '平台自营'),
    CASE
        WHEN sl.category LIKE '%航拍%' THEN 'aerial'
        WHEN sl.category LIKE '%试飞%' THEN 'test_fly'
        WHEN sl.category LIKE '%检测%' OR sl.category LIKE '%标定%' THEN 'calibration'
        WHEN sl.category LIKE '%空域%' THEN 'airspace'
        ELSE 'repair'
    END,
    sl.title,
    sl.description,
    sl.price_fen,
    CASE WHEN sl.price_fen = 0 THEN 'negotiable' ELSE 'fixed' END,
    'negotiable',
    CASE WHEN COALESCE(sl.image, '') = '' THEN '[]'::jsonb ELSE jsonb_build_array(sl.image) END,
    '', '', 'new',
    0,
    CASE WHEN sl.status = 'published' THEN 'listed' ELSE 'removed' END,
    'passed', '', now(), '',
    sl.category, sl.region, sl.unit,
    1, sl.created_at, sl.updated_at
FROM service_listings sl
ON CONFLICT (id) DO NOTHING;

-- 3) 收藏关系跟着迁移：service_listing_favorites → product_favorites
--    id 加 'pfav-sl-' 前缀，便于 down 精确回退，也不会与既有收藏 id 冲突
INSERT INTO product_favorites (id, user_id, product_id, created_at)
SELECT 'pfav-sl-' || f.user_id || '-' || f.listing_id, f.user_id, f.listing_id, f.created_at
FROM service_listing_favorites f
WHERE EXISTS (SELECT 1 FROM drone_products p WHERE p.id = f.listing_id)
ON CONFLICT (user_id, product_id) DO NOTHING;

CREATE INDEX IF NOT EXISTS idx_drone_products_category ON drone_products(category);
