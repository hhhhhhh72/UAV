-- 回滚 000119：把唯一索引恢复成只覆盖 release/transfer（与 000116 一致）。
DROP INDEX IF EXISTS idx_escrow_once_per_ref;
CREATE UNIQUE INDEX IF NOT EXISTS idx_escrow_once_per_ref
    ON escrow_transactions (from_user, tx_type, reference_type, reference_id)
    WHERE tx_type IN ('release', 'transfer') AND reference_id <> '';