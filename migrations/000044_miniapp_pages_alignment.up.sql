-- 小程序 123 包 9 页对齐：colleges / competitions / training_courses 补充页面字段
-- 页面读取 snake_case 变体（cover/image、district/region 等），列名与 JSON tag 保持一致

-- ============ colleges（院校展示 list/detail） ============
ALTER TABLE colleges ADD COLUMN IF NOT EXISTS city TEXT NOT NULL DEFAULT '';
ALTER TABLE colleges ADD COLUMN IF NOT EXISTS tags JSONB NOT NULL DEFAULT '[]';             -- 985/211/专科/高职 等
ALTER TABLE colleges ADD COLUMN IF NOT EXISTS short_name TEXT NOT NULL DEFAULT '';
ALTER TABLE colleges ADD COLUMN IF NOT EXISTS level_tags TEXT NOT NULL DEFAULT '';
ALTER TABLE colleges ADD COLUMN IF NOT EXISTS specialties JSONB NOT NULL DEFAULT '[]';      -- 特色专业（字符串列表）
ALTER TABLE colleges ADD COLUMN IF NOT EXISTS major_count INT NOT NULL DEFAULT 0;
ALTER TABLE colleges ADD COLUMN IF NOT EXISTS partner_count INT NOT NULL DEFAULT 0;
ALTER TABLE colleges ADD COLUMN IF NOT EXISTS teacher_count INT NOT NULL DEFAULT 0;
ALTER TABLE colleges ADD COLUMN IF NOT EXISTS student_count INT NOT NULL DEFAULT 0;
ALTER TABLE colleges ADD COLUMN IF NOT EXISTS graduate_rate TEXT NOT NULL DEFAULT '';
ALTER TABLE colleges ADD COLUMN IF NOT EXISTS partners JSONB NOT NULL DEFAULT '[]';         -- 合作企业 [{icon,name,type}]
ALTER TABLE colleges ADD COLUMN IF NOT EXISTS cover TEXT NOT NULL DEFAULT '';               -- 封面图（页面 cover||image||campus_image||cover_image）
ALTER TABLE colleges ADD COLUMN IF NOT EXISTS photos JSONB NOT NULL DEFAULT '[]';           -- 校园环境图
ALTER TABLE colleges ADD COLUMN IF NOT EXISTS phone TEXT NOT NULL DEFAULT '';
ALTER TABLE colleges ADD COLUMN IF NOT EXISTS website TEXT NOT NULL DEFAULT '';
ALTER TABLE colleges ADD COLUMN IF NOT EXISTS intro TEXT NOT NULL DEFAULT '';
ALTER TABLE colleges ADD COLUMN IF NOT EXISTS majors_detail JSONB NOT NULL DEFAULT '[]';    -- 专业对象数组 [{name,degree,duration,key,flagship}]

-- ============ competitions（赛事 list/detail/register） ============
ALTER TABLE competitions ADD COLUMN IF NOT EXISTS fee INT NOT NULL DEFAULT 0;               -- 报名费（分）
ALTER TABLE competitions ADD COLUMN IF NOT EXISTS tags JSONB NOT NULL DEFAULT '[]';
ALTER TABLE competitions ADD COLUMN IF NOT EXISTS poster TEXT NOT NULL DEFAULT '';
ALTER TABLE competitions ADD COLUMN IF NOT EXISTS deadline TIMESTAMPTZ;                     -- 报名截止
ALTER TABLE competitions ADD COLUMN IF NOT EXISTS organizer_sub TEXT NOT NULL DEFAULT '';
ALTER TABLE competitions ADD COLUMN IF NOT EXISTS min_fee INT NOT NULL DEFAULT 0;
ALTER TABLE competitions ADD COLUMN IF NOT EXISTS registration_status TEXT NOT NULL DEFAULT '';
ALTER TABLE competitions ADD COLUMN IF NOT EXISTS requirements JSONB NOT NULL DEFAULT '[]'; -- [{icon,name,desc,level}]
ALTER TABLE competitions ADD COLUMN IF NOT EXISTS events JSONB NOT NULL DEFAULT '[]';       -- [{name,type,format,fee}]
ALTER TABLE competitions ADD COLUMN IF NOT EXISTS prizes JSONB NOT NULL DEFAULT '[]';       -- [{level,amount,metal,medal}]

-- ============ training_courses（培训 courses/enroll/register） ============
ALTER TABLE training_courses ADD COLUMN IF NOT EXISTS org_name TEXT NOT NULL DEFAULT '';
ALTER TABLE training_courses ADD COLUMN IF NOT EXISTS rating TEXT NOT NULL DEFAULT '';
ALTER TABLE training_courses ADD COLUMN IF NOT EXISTS review_count INT NOT NULL DEFAULT 0;
ALTER TABLE training_courses ADD COLUMN IF NOT EXISTS image TEXT NOT NULL DEFAULT '';        -- 封面（页面 image||cover_image||image_url）
ALTER TABLE training_courses ADD COLUMN IF NOT EXISTS cover_image TEXT NOT NULL DEFAULT '';
ALTER TABLE training_courses ADD COLUMN IF NOT EXISTS tags JSONB NOT NULL DEFAULT '[]';
ALTER TABLE training_courses ADD COLUMN IF NOT EXISTS district TEXT NOT NULL DEFAULT '';
ALTER TABLE training_courses ADD COLUMN IF NOT EXISTS duration_days INT NOT NULL DEFAULT 0;
ALTER TABLE training_courses ADD COLUMN IF NOT EXISTS certificate TEXT NOT NULL DEFAULT '';  -- 证书/结业证书图
ALTER TABLE training_courses ADD COLUMN IF NOT EXISTS certificate_url TEXT NOT NULL DEFAULT '';
ALTER TABLE training_courses ADD COLUMN IF NOT EXISTS courses JSONB NOT NULL DEFAULT '[]';   -- 课程方案 [{name,price}]
ALTER TABLE training_courses ADD COLUMN IF NOT EXISTS prices JSONB NOT NULL DEFAULT '[]';    -- 价格方案 [{name,price}]
ALTER TABLE training_courses ADD COLUMN IF NOT EXISTS business_hours TEXT NOT NULL DEFAULT '';
ALTER TABLE training_courses ADD COLUMN IF NOT EXISTS phone TEXT NOT NULL DEFAULT '';
