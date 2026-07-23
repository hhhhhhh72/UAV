-- 企业入驻 + 需求承接 + 其他补全

-- 企业资质文件
CREATE TABLE IF NOT EXISTS enterprise_documents (
    id            TEXT PRIMARY KEY,
    enterprise_id TEXT NOT NULL,
    file_id       TEXT NOT NULL,
    document_type TEXT NOT NULL,
    review_status TEXT DEFAULT 'pending',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 审核记录（企业/需求/内容复用）
CREATE TABLE IF NOT EXISTS review_records (
    id            TEXT PRIMARY KEY,
    resource_type TEXT NOT NULL,
    resource_id   TEXT NOT NULL,
    action        TEXT NOT NULL,
    reason        TEXT,
    reviewer_id   TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_review_resource ON review_records(resource_type, resource_id);

-- 合同模板
CREATE TABLE IF NOT EXISTS contract_templates (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    version     INT NOT NULL DEFAULT 1,
    content     TEXT,
    status      TEXT DEFAULT 'active',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 合同事件（签章回调记录）
CREATE TABLE IF NOT EXISTS contract_events (
    id          TEXT PRIMARY KEY,
    contract_id TEXT NOT NULL,
    event_type  TEXT NOT NULL,
    event_id    TEXT UNIQUE NOT NULL,
    payload     JSONB DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 招聘职位
CREATE TABLE IF NOT EXISTS jobs (
    id           TEXT PRIMARY KEY,
    enterprise_id TEXT NOT NULL,
    title        TEXT NOT NULL,
    description  TEXT,
    location     TEXT,
    salary_fen   BIGINT DEFAULT 0,
    status       TEXT DEFAULT 'draft',
    version      INT NOT NULL DEFAULT 1,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 简历
CREATE TABLE IF NOT EXISTS resumes (
    id          TEXT PRIMARY KEY,
    user_id     TEXT NOT NULL,
    title       TEXT,
    content     TEXT,
    visibility  TEXT DEFAULT 'private',
    version     INT NOT NULL DEFAULT 1,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 投递记录
CREATE TABLE IF NOT EXISTS job_applications (
    id          TEXT PRIMARY KEY,
    job_id      TEXT NOT NULL,
    resume_id   TEXT NOT NULL,
    applicant_id TEXT NOT NULL,
    status      TEXT DEFAULT 'submitted',
    version     INT NOT NULL DEFAULT 1,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_app_unique ON job_applications(job_id, applicant_id);
