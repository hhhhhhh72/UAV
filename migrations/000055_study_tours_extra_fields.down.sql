-- 回滚：删除研学补齐字段
ALTER TABLE study_tours
    DROP COLUMN IF EXISTS cover_image,
    DROP COLUMN IF EXISTS price_fen,
    DROP COLUMN IF EXISTS schedule;
