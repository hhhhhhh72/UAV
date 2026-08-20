-- 需求收藏（用户收藏需求，详情页收藏入口与后续"我的收藏"列表）
CREATE TABLE IF NOT EXISTS demand_favorites (
    id        TEXT PRIMARY KEY,
    user_id   TEXT NOT NULL,
    demand_id TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_demand_favorites_user_demand ON demand_favorites(user_id, demand_id);
CREATE INDEX IF NOT EXISTS idx_demand_favorites_user ON demand_favorites(user_id, created_at DESC);
