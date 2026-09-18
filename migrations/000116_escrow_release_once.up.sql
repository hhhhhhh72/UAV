-- 000116: 让「同一付款方对同一业务单只放款一次」成为**数据库级**约束。
--
-- 背景：escrow_transactions 上唯一的唯一索引是
--   idx_escrow_external UNIQUE (channel, external_txn_id) WHERE external_txn_id <> ''
-- 它只覆盖**真实渠道入金**（微信支付回调带外部单号）。release / transfer 走内部渠道，
-- external_txn_id 恒为空串，这条部分索引根本不生效——库层面完全不去重。
--
-- 于是「不重复放款」只剩应用层一道 check-then-act（EscrowService.Release 先查
-- HasReleased 再落库），而 EscrowService.Release 的注释自己写明了这一点：
--   「PG 的 Release 只校验 frozen_fen 足够、没有 (from,ref) 去重，本次查询是唯一防线」
--
-- 两个并发请求同时通过该查询时（典型场景：completeEnrollment 的「completed 幂等补齐」
-- 分支没有状态 CAS，两个并发重试都会走到这里），付款方若还有**其它**冻结资金，
-- 两次释放都会成功 → 机构双倍入账、付款方被多扣。
--
-- Release 的余额调整与流水写入在同一个事务里，所以这条索引一旦存在，
-- 第二次插入会带着整个事务回滚——双倍放款在数据库层被彻底堵死，
-- 不依赖应用层的查询时序。（postgres.Release/Transfer 会把 23505 翻译成幂等成功，
-- 调用方看到的是"已放款"而不是报错。）
--
-- 只覆盖 release/transfer：freeze/refund 的重复是设计允许的
-- （报名被拒 → 回滚退款 → 用户重试，生产上确有同一 (user, course) 多次 freeze/refund 的记录），
-- 对它们加唯一约束会直接挡住正常业务。
--
-- 建索引前已核对生产数据：release/transfer 按 (from_user, tx_type, reference_type, reference_id)
-- 无任何重复（全表仅 1 行 release、0 行 transfer），可以安全创建。
CREATE UNIQUE INDEX IF NOT EXISTS idx_escrow_once_per_ref
    ON escrow_transactions (from_user, tx_type, reference_type, reference_id)
    WHERE tx_type IN ('release', 'transfer') AND reference_id <> '';
