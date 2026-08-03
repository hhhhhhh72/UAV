-- 评价审核：reviews 表补 status 列（000011 建表遗漏，代码/审核流程已引用）
ALTER TABLE reviews ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'pending';
CREATE INDEX IF NOT EXISTS idx_reviews_status ON reviews(status);
