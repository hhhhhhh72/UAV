DROP INDEX IF EXISTS idx_resources_visibility;
ALTER TABLE industry_resources DROP COLUMN IF EXISTS visibility_level;
