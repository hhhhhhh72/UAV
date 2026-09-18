-- 000118: 线上退款单。
--
-- 背景：internal/wechatpay 此前完全没有退款能力，真实支付一旦开通，用户微信付的钱在
-- 退款时只会把平台内余额加回去，真金白银不会回到微信钱包。
--
-- 为什么退款的对象是**充值单**而不是订单：平台的钱是以「充值进托管余额」的形式进来的，
-- 用户微信付的那笔钱对应一条 payment_orders。微信退款 API 本就要求 out_trade_no +
-- 累计退款不超过原单金额，与这个模型正好对上；一个订单则可能由多笔充值拼出来，
-- 反而不适合直接作为退款对象。
--
-- 「累计退款不超过原单金额」放在**库层面**保证（refunded_fen 条件更新 + CHECK 约束），
-- 不依赖应用层先查后写——这与 000116/000117 对放款/退款做的是同一件事：
-- 资金不变量应由数据库兜底。

ALTER TABLE payment_orders ADD COLUMN IF NOT EXISTS refunded_fen bigint NOT NULL DEFAULT 0;

-- CHECK 约束没有 IF NOT EXISTS 语法，用 DO 块保证可重复执行。
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'payment_orders_refunded_within_amount') THEN
        ALTER TABLE payment_orders ADD CONSTRAINT payment_orders_refunded_within_amount
            CHECK (refunded_fen >= 0 AND refunded_fen <= amount_fen);
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS payment_refunds (
    id            text PRIMARY KEY,
    out_refund_no text        NOT NULL,
    out_trade_no  text        NOT NULL,
    user_id       text        NOT NULL,
    amount_fen    bigint      NOT NULL CHECK (amount_fen > 0),
    reason        text        NOT NULL DEFAULT '',
    status        text        NOT NULL DEFAULT 'created',
    refund_id     text        NOT NULL DEFAULT '',
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now()
);

-- 退款单号是我们自己生成的，必须全局唯一：它是微信退款回调的唯一对账主键。
CREATE UNIQUE INDEX IF NOT EXISTS payment_refunds_out_refund_no_key ON payment_refunds (out_refund_no);
-- 微信退款单号唯一（部分索引）：同一笔微信退款不可能挂到两个本地退款单上。
CREATE UNIQUE INDEX IF NOT EXISTS idx_payment_refunds_wechat ON payment_refunds (refund_id) WHERE refund_id <> '';
CREATE INDEX IF NOT EXISTS idx_payment_refunds_order ON payment_refunds (out_trade_no, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_payment_refunds_user ON payment_refunds (user_id, created_at DESC);