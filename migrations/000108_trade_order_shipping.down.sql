ALTER TABLE trade_orders DROP COLUMN IF EXISTS shipped_at;
ALTER TABLE trade_orders DROP COLUMN IF EXISTS shipping_tracking;
ALTER TABLE trade_orders DROP COLUMN IF EXISTS shipping_company;
ALTER TABLE trade_orders DROP COLUMN IF EXISTS receiver_address;
ALTER TABLE trade_orders DROP COLUMN IF EXISTS receiver_region;
ALTER TABLE trade_orders DROP COLUMN IF EXISTS receiver_phone;
ALTER TABLE trade_orders DROP COLUMN IF EXISTS receiver_name;
