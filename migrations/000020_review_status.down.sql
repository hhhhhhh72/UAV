DROP INDEX IF EXISTS idx_reviews_status;
ALTER TABLE reviews DROP COLUMN IF EXISTS status;
