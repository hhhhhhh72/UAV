-- 回滚：第二轮小程序字段补齐（000045 up）
ALTER TABLE instructors DROP COLUMN IF EXISTS photo;

ALTER TABLE enterprises DROP COLUMN IF EXISTS cover_image;
ALTER TABLE enterprises DROP COLUMN IF EXISTS logo;
ALTER TABLE enterprises DROP COLUMN IF EXISTS business_hours;

ALTER TABLE training_courses DROP COLUMN IF EXISTS course_types;
ALTER TABLE training_courses DROP COLUMN IF EXISTS environment;
ALTER TABLE training_courses DROP COLUMN IF EXISTS remain;
