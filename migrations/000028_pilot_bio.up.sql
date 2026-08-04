-- 飞手简介：擅长领域/自我介绍（认证飞手申请表单）
ALTER TABLE certified_pilots ADD COLUMN IF NOT EXISTS bio TEXT NOT NULL DEFAULT '';
