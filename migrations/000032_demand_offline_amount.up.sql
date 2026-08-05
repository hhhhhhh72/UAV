-- 线下成交金额登记（联系对接模式：平台撮合价值度量）
ALTER TABLE demands ADD COLUMN IF NOT EXISTS offline_amount_fen BIGINT NOT NULL DEFAULT 0;
