-- 商品/服务能力/培训课程 三类收藏（用户收藏浏览内容，"我的收藏"列表数据源）
-- 与 demand_favorites(000074) 同构：id 主键 + (user_id, item_id) 唯一 + 用户收藏时间索引
CREATE TABLE IF NOT EXISTS product_favorites (
    id         TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL,
    product_id TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_product_favorites_user_product ON product_favorites(user_id, product_id);
CREATE INDEX IF NOT EXISTS idx_product_favorites_user ON product_favorites(user_id, created_at DESC);

CREATE TABLE IF NOT EXISTS service_listing_favorites (
    id         TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL,
    listing_id TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_service_listing_favorites_user_listing ON service_listing_favorites(user_id, listing_id);
CREATE INDEX IF NOT EXISTS idx_service_listing_favorites_user ON service_listing_favorites(user_id, created_at DESC);

CREATE TABLE IF NOT EXISTS training_course_favorites (
    id        TEXT PRIMARY KEY,
    user_id   TEXT NOT NULL,
    course_id TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_training_course_favorites_user_course ON training_course_favorites(user_id, course_id);
CREATE INDEX IF NOT EXISTS idx_training_course_favorites_user ON training_course_favorites(user_id, created_at DESC);
