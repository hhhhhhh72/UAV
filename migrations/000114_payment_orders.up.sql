-- 线上支付订单：用户充值先落 created，微信到账确认后 CAS 置 paid 再入托管金。
-- out_trade_no 唯一：它是对账主键，微信回调只带回它，金额与用户由本表反查。
CREATE TABLE IF NOT EXISTS payment_orders (
    id              TEXT PRIMARY KEY,
    out_trade_no    TEXT        NOT NULL UNIQUE,
    user_id         TEXT        NOT NULL,
    amount_fen      BIGINT      NOT NULL CHECK (amount_fen > 0),
    channel         TEXT        NOT NULL DEFAULT 'wechat',
    status          TEXT        NOT NULL DEFAULT 'created',
    prepay_id       TEXT        NOT NULL DEFAULT '',
    transaction_id  TEXT        NOT NULL DEFAULT '',
    paid_at         TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 一人一单：同一用户同一商户订单号不可能重复（UNIQUE 已保证），此处仅为列表查询。
CREATE INDEX IF NOT EXISTS idx_payment_orders_user ON payment_orders (user_id, created_at DESC);

-- 回调幂等键：微信支付单号全局唯一，重复回调不得二次入账。
CREATE UNIQUE INDEX IF NOT EXISTS idx_payment_orders_txn
    ON payment_orders (transaction_id) WHERE transaction_id <> '';
