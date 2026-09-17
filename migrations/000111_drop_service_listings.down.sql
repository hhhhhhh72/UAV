-- 不可逆迁移：旧表结构与其中数据无法从本迁移恢复。
-- 数据本身仍在 drone_products 里（000110 迁过去的），如需回退请回退 000110 的代码侧。
-- 这里只把表结构建回来，保证 down 不会因为"表不存在"而报错。
CREATE TABLE IF NOT EXISTS service_listings (
    id            TEXT PRIMARY KEY,
    provider_id   TEXT NOT NULL DEFAULT '',
    provider_name TEXT NOT NULL DEFAULT '',
    title         TEXT NOT NULL,
    category      TEXT NOT NULL DEFAULT '',
    description   TEXT NOT NULL DEFAULT '',
    region        TEXT NOT NULL DEFAULT '',
    price_fen     BIGINT NOT NULL DEFAULT 0,
    unit          TEXT NOT NULL DEFAULT '',
    image         TEXT NOT NULL DEFAULT '',
    status        TEXT NOT NULL DEFAULT 'published',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS service_listing_favorites (
    id         TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL,
    listing_id TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
