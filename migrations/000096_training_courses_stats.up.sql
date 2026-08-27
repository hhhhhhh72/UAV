-- 培训课程补充统计字段：通过率/机构年限（小程序评分卡"通过考试/机构年限"数据源，
-- 此前后端无字段恒显"—"；管理后台同步补录入）
ALTER TABLE training_courses ADD COLUMN IF NOT EXISTS pass_rate TEXT NOT NULL DEFAULT '';
ALTER TABLE training_courses ADD COLUMN IF NOT EXISTS years INTEGER NOT NULL DEFAULT 0;
