-- 交付包字段补齐：培训课程划线原价/通过率/机构年限/规模/简介/Hero大图 + 赛事划线原价
-- （前端已兼容兜底，后端补齐后划线价/通过率等真实显示，前端零改动）

ALTER TABLE training_courses ADD COLUMN IF NOT EXISTS original_fee BIGINT NOT NULL DEFAULT 0;
ALTER TABLE training_courses ADD COLUMN IF NOT EXISTS pass_rate DOUBLE PRECISION NOT NULL DEFAULT 0;
ALTER TABLE training_courses ADD COLUMN IF NOT EXISTS years INT NOT NULL DEFAULT 0;
ALTER TABLE training_courses ADD COLUMN IF NOT EXISTS scale TEXT NOT NULL DEFAULT '';
ALTER TABLE training_courses ADD COLUMN IF NOT EXISTS intro TEXT NOT NULL DEFAULT '';
ALTER TABLE training_courses ADD COLUMN IF NOT EXISTS banner TEXT NOT NULL DEFAULT '';

ALTER TABLE competitions ADD COLUMN IF NOT EXISTS original_fee BIGINT NOT NULL DEFAULT 0;
