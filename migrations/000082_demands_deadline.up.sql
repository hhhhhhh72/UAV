-- 需求有效期截止日（发布时效管理）：YYYY-MM-DD 文本列，空串=长期有效；
-- 服务层校验格式且不得早于今天。NOT NULL DEFAULT ''：存量行扫描不会出现 NULL。
ALTER TABLE demands ADD COLUMN IF NOT EXISTS deadline TEXT NOT NULL DEFAULT '';
