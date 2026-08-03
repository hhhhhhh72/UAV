-- Batch1-3 模块建表 + jobs 补列
-- 修复：colleges/exhibitions/test_sites/study_tours/transformations 等 13 张表
-- 此前只有代码无建表脚本，PG 模式下这些模块全部缺表。

-- 招聘职位补列（000003 漏建 job_type，代码已引用）
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS job_type TEXT NOT NULL DEFAULT '';

-- 院校展示
CREATE TABLE IF NOT EXISTS colleges (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    region      TEXT DEFAULT '',
    description TEXT DEFAULT '',
    logo_url    TEXT DEFAULT '',
    status      TEXT NOT NULL DEFAULT 'active',
    majors      JSONB DEFAULT '[]',
    facilities  JSONB DEFAULT '[]',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 研学活动
CREATE TABLE IF NOT EXISTS study_tours (
    id          TEXT PRIMARY KEY,
    title       TEXT NOT NULL,
    destination TEXT DEFAULT '',
    duration    TEXT DEFAULT '',
    capacity    INTEGER DEFAULT 0,
    status      TEXT NOT NULL DEFAULT 'draft',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 产业展会
CREATE TABLE IF NOT EXISTS exhibitions (
    id              TEXT PRIMARY KEY,
    title           TEXT NOT NULL,
    category        TEXT DEFAULT '',
    description     TEXT DEFAULT '',
    location        TEXT DEFAULT '',
    start_date      TIMESTAMPTZ,
    end_date        TIMESTAMPTZ,
    booth_count     INTEGER DEFAULT 0,
    booth_price_fen BIGINT DEFAULT 0,
    organizer       TEXT DEFAULT '',
    status          TEXT NOT NULL DEFAULT 'draft',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 展会展位预订
CREATE TABLE IF NOT EXISTS exhibition_booths (
    id            TEXT PRIMARY KEY,
    exhibition_id TEXT NOT NULL,
    exhibitor_id  TEXT NOT NULL DEFAULT '',
    booth_number  TEXT DEFAULT '',
    exhibit_name  TEXT DEFAULT '',
    exhibit_desc  TEXT DEFAULT '',
    status        TEXT NOT NULL DEFAULT 'applied',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 测试环境
CREATE TABLE IF NOT EXISTS test_sites (
    id           TEXT PRIMARY KEY,
    name         TEXT NOT NULL,
    site_type    TEXT DEFAULT '',
    owner_id     TEXT NOT NULL DEFAULT '',
    location     TEXT DEFAULT '',
    booking_rule TEXT DEFAULT '',
    status       TEXT NOT NULL DEFAULT 'available',
    price_fen    BIGINT DEFAULT 0,
    facilities   JSONB DEFAULT '[]',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 测试环境预约
CREATE TABLE IF NOT EXISTS test_site_bookings (
    id          TEXT PRIMARY KEY,
    site_id     TEXT NOT NULL,
    user_id     TEXT NOT NULL,
    purpose     TEXT DEFAULT '',
    start_time  TIMESTAMPTZ,
    end_time    TIMESTAMPTZ,
    status      TEXT NOT NULL DEFAULT 'pending',
    review_note TEXT DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 产业资源池
CREATE TABLE IF NOT EXISTS resource_pools (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    pool_type   TEXT DEFAULT '',
    description TEXT DEFAULT '',
    owner_id    TEXT NOT NULL DEFAULT '',
    resources   JSONB DEFAULT '[]',
    status      TEXT NOT NULL DEFAULT 'active',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 资源池成员
CREATE TABLE IF NOT EXISTS resource_pool_members (
    id        TEXT PRIMARY KEY,
    pool_id   TEXT NOT NULL,
    res_id    TEXT NOT NULL,
    res_type  TEXT DEFAULT '',
    quantity  INTEGER DEFAULT 1,
    status    TEXT NOT NULL DEFAULT 'standby',
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 校企共建
CREATE TABLE IF NOT EXISTS cooperation_programs (
    id             TEXT PRIMARY KEY,
    title          TEXT NOT NULL,
    college_id     TEXT NOT NULL DEFAULT '',
    enterprise_id  TEXT NOT NULL DEFAULT '',
    coop_type      TEXT DEFAULT '',
    description    TEXT DEFAULT '',
    start_date     TIMESTAMPTZ,
    end_date       TIMESTAMPTZ,
    student_quota  INTEGER DEFAULT 0,
    status         TEXT NOT NULL DEFAULT 'proposed',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 救援案例库
CREATE TABLE IF NOT EXISTS rescue_cases (
    id          TEXT PRIMARY KEY,
    title       TEXT NOT NULL,
    event_type  TEXT DEFAULT '',
    location    TEXT DEFAULT '',
    date        TIMESTAMPTZ,
    drone_model TEXT DEFAULT '',
    team_name   TEXT DEFAULT '',
    summary     TEXT DEFAULT '',
    result      TEXT DEFAULT '',
    lessons     TEXT DEFAULT '',
    media_urls  JSONB DEFAULT '[]',
    source      TEXT DEFAULT '',
    status      TEXT NOT NULL DEFAULT 'draft',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 应急管理部门对接
CREATE TABLE IF NOT EXISTS emergency_depts (
    id             TEXT PRIMARY KEY,
    name           TEXT NOT NULL,
    dept_type      TEXT DEFAULT '',
    region         TEXT DEFAULT '',
    contact_name   TEXT DEFAULT '',
    contact_phone  TEXT DEFAULT '',
    protocol_url   TEXT DEFAULT '',
    status         TEXT NOT NULL DEFAULT 'active',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 联合演练
CREATE TABLE IF NOT EXISTS emergency_drills (
    id           TEXT PRIMARY KEY,
    dept_id      TEXT NOT NULL,
    title        TEXT NOT NULL,
    scenario     TEXT DEFAULT '',
    date         TIMESTAMPTZ,
    participants INTEGER DEFAULT 0,
    drone_count  INTEGER DEFAULT 0,
    result       TEXT DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 协会多级会员
CREATE TABLE IF NOT EXISTS association_members (
    id            TEXT PRIMARY KEY,
    user_id       TEXT NOT NULL,
    enterprise_id TEXT NOT NULL DEFAULT '',
    role          TEXT NOT NULL DEFAULT 'member',
    join_date     TIMESTAMPTZ DEFAULT NOW(),
    expire_date   TIMESTAMPTZ,
    status        TEXT NOT NULL DEFAULT 'active',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 常用查询索引
CREATE INDEX IF NOT EXISTS idx_cooperation_college ON cooperation_programs(college_id);
CREATE INDEX IF NOT EXISTS idx_booths_exhibition ON exhibition_booths(exhibition_id);
CREATE INDEX IF NOT EXISTS idx_site_bookings_site ON test_site_bookings(site_id);
CREATE INDEX IF NOT EXISTS idx_pool_members_pool ON resource_pool_members(pool_id);
CREATE INDEX IF NOT EXISTS idx_drills_dept ON emergency_drills(dept_id);
CREATE INDEX IF NOT EXISTS idx_assoc_members_user ON association_members(user_id);
CREATE INDEX IF NOT EXISTS idx_rescue_cases_type ON rescue_cases(event_type);
