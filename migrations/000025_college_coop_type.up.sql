-- 院校分域标注：科研合作（三系统）/ 人才培养（五系统）（功能方案修订版 v2 三·五 分域）
ALTER TABLE colleges ADD COLUMN IF NOT EXISTS coop_type TEXT NOT NULL DEFAULT 'both';
