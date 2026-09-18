-- 回滚 000117：去掉 trade_order 退款唯一约束。
-- 注意：回滚后「确认收到退货」的并发双击重新只剩应用层的 check-then-act 一道防线。
DROP INDEX IF EXISTS idx_escrow_refund_once_per_order;
