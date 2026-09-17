DROP INDEX IF EXISTS idx_trade_orders_status_updated;
DROP INDEX IF EXISTS idx_trade_orders_status_created;
ALTER TABLE trade_orders DROP COLUMN IF EXISTS returned_at;
ALTER TABLE trade_orders DROP COLUMN IF EXISTS return_note;
ALTER TABLE trade_orders DROP COLUMN IF EXISTS return_tracking;
