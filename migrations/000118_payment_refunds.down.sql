-- 回滚 000118：去掉线上退款单。
-- 注意：回滚后微信侧已发生的退款将没有本地记录，退款回调会因找不到单号而被拒。
DROP TABLE IF EXISTS payment_refunds;
ALTER TABLE payment_orders DROP CONSTRAINT IF EXISTS payment_orders_refunded_within_amount;
ALTER TABLE payment_orders DROP COLUMN IF EXISTS refunded_fen;