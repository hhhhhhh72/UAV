-- 回滚：删除飞手补齐字段
ALTER TABLE certified_pilots
    DROP COLUMN IF EXISTS avatar,
    DROP COLUMN IF EXISTS region;
