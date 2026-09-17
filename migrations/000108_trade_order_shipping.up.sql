-- 交付链路补齐：收货地址 + 发货单号。
--
-- 缺口：发布表单里明明有「交付方式：自提/同城配送/物流发货/可协商」，但订单表
-- 既没有收货地址也没有发货单号——卖家选"物流发货"卖出一台整机后，
-- 订单里不知道寄给谁，也没地方填快递单号，而订单详情页的"确认发货"按钮是存在的。
-- 全库检索"收货地址/shipping_address/receiver_address/address_id" 零命中。
--
-- 本迁移只加字段，不改任何既有行的可见语义：历史订单的地址与单号为空串，
-- 发货流程对它们不适用（它们本来就发不出去）。

-- 收货信息：下单时快照到订单，不建地址簿（一期按单件直填；地址簿留待后续）。
ALTER TABLE trade_orders ADD COLUMN IF NOT EXISTS receiver_name    TEXT NOT NULL DEFAULT '';
ALTER TABLE trade_orders ADD COLUMN IF NOT EXISTS receiver_phone   TEXT NOT NULL DEFAULT '';
ALTER TABLE trade_orders ADD COLUMN IF NOT EXISTS receiver_region  TEXT NOT NULL DEFAULT '';
ALTER TABLE trade_orders ADD COLUMN IF NOT EXISTS receiver_address TEXT NOT NULL DEFAULT '';

-- 发货信息：卖家发货时写入。shipping_tracking 是发货单号（与 return_tracking 退货单号区分开，
-- 二者此前只有退货那一个，出库方向完全没有留痕）。
ALTER TABLE trade_orders ADD COLUMN IF NOT EXISTS shipping_company  TEXT NOT NULL DEFAULT '';
ALTER TABLE trade_orders ADD COLUMN IF NOT EXISTS shipping_tracking TEXT NOT NULL DEFAULT '';
ALTER TABLE trade_orders ADD COLUMN IF NOT EXISTS shipped_at        TIMESTAMPTZ;
