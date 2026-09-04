-- 成果收藏（用户收藏成果，幂等；计数与 favorites 表联动，重复收藏不重复计数）
CREATE TABLE IF NOT EXISTS achievement_favorites (
    id             TEXT PRIMARY KEY,
    user_id        TEXT NOT NULL,
    achievement_id TEXT NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_achievement_favorites_user_ach ON achievement_favorites(user_id, achievement_id);
CREATE INDEX IF NOT EXISTS idx_achievement_favorites_user ON achievement_favorites(user_id, created_at DESC);
