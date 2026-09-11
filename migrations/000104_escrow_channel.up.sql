-- 资金闭环（商城订单接入托管金）+ 真实支付接入预留
--
-- 1) escrow_transactions 增加支付渠道与外部单号：现在资金全是内部记账（channel=internal），
--    将来接微信支付时，入金/退款的微信支付单号写进 external_txn_id，用于对账与幂等。
-- 2) 补 (reference_type, reference_id) 索引：订单/报名的资金幂等查询（HasFrozen/HasReleased/
--    HasRefunded）与对账扫描都按这两列过滤，此前无索引。
ALTER TABLE escrow_transactions ADD COLUMN IF NOT EXISTS channel varchar(32) NOT NULL DEFAULT 'internal';
ALTER TABLE escrow_transactions ADD COLUMN IF NOT EXISTS external_txn_id varchar(128) NOT NULL DEFAULT '';
CREATE INDEX IF NOT EXISTS idx_escrow_ref ON escrow_transactions(reference_type, reference_id);
-- 3) (channel, external_txn_id) 唯一索引：真实支付入账的最终防线。
--    并发回调同时通过「查重」时，唯一索引保证同一笔外部支付只能入账一次（重复入账＝印钞）。
CREATE UNIQUE INDEX IF NOT EXISTS idx_escrow_external ON escrow_transactions(channel, external_txn_id) WHERE external_txn_id <> '';
