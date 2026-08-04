-- 产业资源台账分级浏览（.doc 原始需求：协会管理员/副会长单位/普通会员/合作院校/政府访客分级浏览）
-- visibility_level: public(政府访客/所有人) < member(会员+) < partner(副会长单位+) < admin(仅协会管理员)
ALTER TABLE industry_resources ADD COLUMN IF NOT EXISTS visibility_level TEXT NOT NULL DEFAULT 'public';
CREATE INDEX IF NOT EXISTS idx_resources_visibility ON industry_resources(visibility_level);
