-- 研发难题攻关要求（前端 detail 读 requirements，兼容字符串/数组两种形态）
ALTER TABLE rd_challenges ADD COLUMN IF NOT EXISTS requirements TEXT NOT NULL DEFAULT '';
