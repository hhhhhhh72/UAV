-- 000119: 把「同一付款方对同一业务单只出账一次」的唯一索引扩展到 withdraw。
--
-- 000116 建 idx_escrow_once_per_ref 时只覆盖了 release/transfer；本次新增的 withdraw
-- （微信退款成功时把钱从用户冻结里扣掉、真正送出平台）与它们同属「一次性的出账动作」，
-- 同样需要库级去重：退款回调会重试，缺了这层就可能在同一次退款上重复扣减用户冻结。
--
-- 生产数据核对：escrow_transactions 中 tx_type='withdraw' 的行数为 0，
-- 且 release/transfer 按 (from_user, reference_type, reference_id) 无重复，可安全重建索引。
DROP INDEX IF EXISTS idx_escrow_once_per_ref;
CREATE UNIQUE INDEX IF NOT EXISTS idx_escrow_once_per_ref
    ON escrow_transactions (from_user, tx_type, reference_type, reference_id)
    WHERE tx_type IN ('release', 'transfer', 'withdraw') AND reference_id <> '';