-- 培训机构报名审核备注（通过/拒绝原因，拒绝必填由服务层校验）
ALTER TABLE enrollments ADD COLUMN IF NOT EXISTS review_note TEXT NOT NULL DEFAULT '';
