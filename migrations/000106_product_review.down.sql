DROP INDEX IF EXISTS idx_drone_products_check_status;
ALTER TABLE drone_products DROP COLUMN IF EXISTS reviewed_by;
ALTER TABLE drone_products DROP COLUMN IF EXISTS reviewed_at;
ALTER TABLE drone_products DROP COLUMN IF EXISTS check_reason;
ALTER TABLE drone_products DROP COLUMN IF EXISTS check_status;
