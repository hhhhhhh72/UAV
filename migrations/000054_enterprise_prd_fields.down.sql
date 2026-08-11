-- 回滚：删除 PRD 补齐字段
ALTER TABLE enterprises
    DROP COLUMN IF EXISTS contact_person,
    DROP COLUMN IF EXISTS email,
    DROP COLUMN IF EXISTS founded_at,
    DROP COLUMN IF EXISTS capability_tags;
