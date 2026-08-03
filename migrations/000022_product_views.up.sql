-- 商品浏览量（详情访问计数，真实业务数据）
ALTER TABLE drone_products ADD COLUMN IF NOT EXISTS views INTEGER NOT NULL DEFAULT 0;
