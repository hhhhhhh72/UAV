-- 培训报名详情（小程序报名表单 12 字段，原训练_enrollments 只有 5 列，资料无处存储）
ALTER TABLE training_enrollments ADD COLUMN IF NOT EXISTS name TEXT NOT NULL DEFAULT '';
ALTER TABLE training_enrollments ADD COLUMN IF NOT EXISTS phone TEXT NOT NULL DEFAULT '';
ALTER TABLE training_enrollments ADD COLUMN IF NOT EXISTS id_card TEXT NOT NULL DEFAULT '';
ALTER TABLE training_enrollments ADD COLUMN IF NOT EXISTS gender TEXT NOT NULL DEFAULT '';
ALTER TABLE training_enrollments ADD COLUMN IF NOT EXISTS birthday TEXT NOT NULL DEFAULT '';
ALTER TABLE training_enrollments ADD COLUMN IF NOT EXISTS email TEXT NOT NULL DEFAULT '';
ALTER TABLE training_enrollments ADD COLUMN IF NOT EXISTS education TEXT NOT NULL DEFAULT '';
ALTER TABLE training_enrollments ADD COLUMN IF NOT EXISTS experience TEXT NOT NULL DEFAULT '';
ALTER TABLE training_enrollments ADD COLUMN IF NOT EXISTS photo_url TEXT NOT NULL DEFAULT '';
ALTER TABLE training_enrollments ADD COLUMN IF NOT EXISTS id_card_image TEXT NOT NULL DEFAULT '';
ALTER TABLE training_enrollments ADD COLUMN IF NOT EXISTS no_crime TEXT NOT NULL DEFAULT '';
