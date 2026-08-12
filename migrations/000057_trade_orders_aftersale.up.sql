-- 交易订单售后契约（一期）：买家申请售后 → 平台审核 → 结案
ALTER TABLE trade_orders ADD COLUMN IF NOT EXISTS aftersale_type TEXT NOT NULL DEFAULT '';
ALTER TABLE trade_orders ADD COLUMN IF NOT EXISTS aftersale_reason TEXT NOT NULL DEFAULT '';
ALTER TABLE trade_orders ADD COLUMN IF NOT EXISTS aftersale_desc TEXT NOT NULL DEFAULT '';
ALTER TABLE trade_orders ADD COLUMN IF NOT EXISTS aftersale_amount_fen BIGINT NOT NULL DEFAULT 0;
ALTER TABLE trade_orders ADD COLUMN IF NOT EXISTS aftersale_status TEXT NOT NULL DEFAULT '';
ALTER TABLE trade_orders ADD COLUMN IF NOT EXISTS aftersale_time TIMESTAMPTZ NOT NULL DEFAULT 'epoch';
