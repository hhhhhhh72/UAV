-- 培训报名 + 交易订单

-- 培训报名记录
CREATE TABLE IF NOT EXISTS training_enrollments (
    id          TEXT PRIMARY KEY,
    course_id   TEXT NOT NULL,
    user_id     TEXT NOT NULL,
    status      TEXT NOT NULL DEFAULT 'enrolled',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(course_id, user_id)
);

-- 交易订单
CREATE TABLE IF NOT EXISTS trade_orders (
    id              TEXT PRIMARY KEY,
    product_id      TEXT NOT NULL,
    buyer_id        TEXT NOT NULL,
    seller_id       TEXT NOT NULL,
    amount_fen      BIGINT DEFAULT 0,
    status          TEXT NOT NULL DEFAULT 'pending',
    version         INT NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_trade_buyer ON trade_orders(buyer_id);
CREATE INDEX IF NOT EXISTS idx_trade_seller ON trade_orders(seller_id);
