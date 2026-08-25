-- 售后驳回还原：记录申请售后前的订单状态（paid/shipped/completed），
-- 驳回时恢复原状态而非强制 completed（未发货已付款订单曾被驳回后永久卡死）。
ALTER TABLE trade_orders ADD COLUMN IF NOT EXISTS aftersale_from TEXT NOT NULL DEFAULT '';
