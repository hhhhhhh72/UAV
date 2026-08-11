-- 低空研学补齐字段（研学三件套-后端待补清单）：
-- cover_image 封面图 URL / price_fen 价格（分）/ schedule 行程安排（JSONB 数组）
-- 注意：迁移系统每次全量重跑，必须幂等（IF NOT EXISTS）
ALTER TABLE study_tours
    ADD COLUMN IF NOT EXISTS cover_image TEXT DEFAULT '',
    ADD COLUMN IF NOT EXISTS price_fen   BIGINT DEFAULT 0,
    ADD COLUMN IF NOT EXISTS schedule    JSONB DEFAULT '[]';
