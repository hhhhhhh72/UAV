-- 000117: 售后「确认收到退货」的退款也只能发生一次——把 trade_order 维度的退款去重下沉到数据库。
--
-- 背景（000116 关掉放款那一半之后剩下的唯一缝隙）：
-- ConfirmReturnReceived 是**先退款再改状态**的（internal/service/phase3.go 里
-- refundForAftersale 在 UpdateAftersale 之前），这是有意为之——钱动不了就保持
-- returned 供重试，否则会出现"状态已结案、钱还冻着"的死局。
-- 代价是并发确认（两个管理员/卖家同时点「确认收到退货」）时，两边都会读到
-- AftersaleStatus='returned'，并双双通过 refundForAftersale 里的
-- HasFrozen + HasRefunded 两道 check-then-act 查询。
--
-- 该函数两个分支里，钱已放给卖家那一支走 Transfer（已被 000116 的
-- idx_escrow_once_per_ref 兜住）；钱还在冻结里这一支走 Refund，此前没有任何库级兜底。
-- 后果不是凭空生钱，而是**把同一笔冻结款重复退回去**：付款方的 frozen 被多扣，
-- 其余仍处于"已付款冻结"状态的订单再也释放不出来（release 会因 frozen 不足失败）。
--
-- 为什么范围只限 reference_type='trade_order'：
-- 一个订单只有一个售后（AftersaleStatus 是 trade_orders 上的单个字段），
-- 所以「一单一退款」正是正确的不变量。而培训课程那条的键是 (user, course)，
-- 同一人可以合法地反复冻结/退款（报名被拒 → 回滚退款 → 用户重试），
-- 对它加唯一约束会挡住正常业务——这一点 service 层有明确注释（phase3.go:464）。
--
-- 建索引前已核对生产数据：trade_order 维度的退款流水分组无任何重复，可以安全创建。
CREATE UNIQUE INDEX IF NOT EXISTS idx_escrow_refund_once_per_order
    ON escrow_transactions (from_user, tx_type, reference_type, reference_id)
    WHERE tx_type = 'refund' AND reference_type = 'trade_order' AND reference_id <> '';
