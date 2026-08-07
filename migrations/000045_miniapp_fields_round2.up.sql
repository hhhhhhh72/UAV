-- 前端待补后端字段清单（2026-08-07）第二轮补齐：
-- training_courses: remain / environment / course_types（courses/enroll 页"仅剩N个"徽章、环境图集、课程类型）
-- enterprises: business_hours / logo / cover_image（enroll 页机构信息区）
-- instructors: photo（字段清单前瞻预留）

ALTER TABLE training_courses ADD COLUMN IF NOT EXISTS remain INT NOT NULL DEFAULT 0;
ALTER TABLE training_courses ADD COLUMN IF NOT EXISTS environment JSONB NOT NULL DEFAULT '[]';   -- 培训环境图集（页面 environment || env_images）
ALTER TABLE training_courses ADD COLUMN IF NOT EXISTS course_types JSONB NOT NULL DEFAULT '[]';  -- 课程类型列表（页面 course_types）

ALTER TABLE enterprises ADD COLUMN IF NOT EXISTS business_hours TEXT NOT NULL DEFAULT '';
ALTER TABLE enterprises ADD COLUMN IF NOT EXISTS logo TEXT NOT NULL DEFAULT '';
ALTER TABLE enterprises ADD COLUMN IF NOT EXISTS cover_image TEXT NOT NULL DEFAULT '';

ALTER TABLE instructors ADD COLUMN IF NOT EXISTS photo TEXT NOT NULL DEFAULT '';
