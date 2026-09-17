-- 商城订单退货退款流程（一期）
--
-- 背景：此前 aftersale_type 只有 refund/return 两个取值，但代码从不读它——
-- 「退货退款」与「仅退款」完全同构，卖家点同意即打款，货是否寄回系统不管，
-- 存在"买家拿了钱又留着货"的风险。
--
-- 本迁移补齐退货环节所需的列，配合 aftersale_status 的新状态：
--   仅退款 refund ：pending(待审核) → approved(已退款) / rejected(已驳回)
--   退货退款 return：pending → returning(已同意退货，待买家寄回)
--                          → returned(买家已寄回，待卖家确认收到)
--                          → approved(确认收货后已退款) / rejected(驳回)
-- 退款只在最后一步发生，因此退货单在 buyer 寄回、seller 确认之前不会动钱。
ALTER TABLE trade_orders ADD COLUMN IF NOT EXISTS return_tracking VARCHAR(128) NOT NULL DEFAULT '';
ALTER TABLE trade_orders ADD COLUMN IF NOT EXISTS return_note     TEXT        NOT NULL DEFAULT '';
ALTER TABLE trade_orders ADD COLUMN IF NOT EXISTS returned_at     TIMESTAMPTZ;

-- 后台任务索引：支付超时自动取消按 (status='pending', created_at)、
-- 自动确认收货按 (status='shipped', updated_at) 扫描，此前无对应索引会全表扫。
CREATE INDEX IF NOT EXISTS idx_trade_orders_status_created ON trade_orders (status, created_at);
CREATE INDEX IF NOT EXISTS idx_trade_orders_status_updated ON trade_orders (status, updated_at);
