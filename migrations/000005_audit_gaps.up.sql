-- 审计缺陷修补: roles / files / deleted_at / outbox_events

-- roles 表
CREATE TABLE IF NOT EXISTS roles (
    id   TEXT PRIMARY KEY,
    code TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL
);
INSERT INTO roles(id, code, name) VALUES
  ('role-1','platform_admin','平台管理员'),
  ('role-2','association_admin','协会管理员'),
  ('role-3','enterprise','企业'),
  ('role-4','individual','个人')
ON CONFLICT DO NOTHING;

-- files 表
CREATE TABLE IF NOT EXISTS files (
    id           TEXT PRIMARY KEY,
    storage_key  TEXT NOT NULL,
    sha256       TEXT,
    content_type TEXT,
    size_bytes   BIGINT DEFAULT 0,
    visibility   TEXT DEFAULT 'private',
    owner_id     TEXT NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_files_owner ON files(owner_id);

-- outbox_events 表
CREATE TABLE IF NOT EXISTS outbox_events (
    id              TEXT PRIMARY KEY,
    aggregate_type  TEXT NOT NULL,
    aggregate_id    TEXT NOT NULL,
    event_type      TEXT NOT NULL,
    payload         JSONB DEFAULT '{}',
    published_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- deleted_at 追加到所有业务表
ALTER TABLE users ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
ALTER TABLE enterprises ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
ALTER TABLE demands ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
ALTER TABLE employment_requests ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
ALTER TABLE contracts ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
ALTER TABLE resumes ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
ALTER TABLE posts ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
ALTER TABLE listings ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
ALTER TABLE labour_orders ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
ALTER TABLE drone_products ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
ALTER TABLE certificates ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
ALTER TABLE training_courses ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
ALTER TABLE insurance_policies ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
ALTER TABLE loan_applications ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
