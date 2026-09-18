-- 回滚 000116：去掉放款唯一约束。
-- 注意：回滚后「不重复放款」重新只剩应用层的 check-then-act 一道防线。
DROP INDEX IF EXISTS idx_escrow_once_per_ref;
