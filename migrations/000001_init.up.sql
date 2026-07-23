-- 无人机产业综合服务平台 初始化迁移
-- PostgreSQL 15+

-- 用户与认证
CREATE TABLE IF NOT EXISTS users (
    id              TEXT PRIMARY KEY,
    wechat_openid   TEXT UNIQUE,
    phone_ciphertext TEXT,
    status          TEXT NOT NULL DEFAULT 'active',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    version         INT NOT NULL DEFAULT 1
);
CREATE INDEX IF NOT EXISTS idx_users_openid ON users(wechat_openid);

CREATE TABLE IF NOT EXISTS user_roles (
    user_id     TEXT NOT NULL REFERENCES users(id),
    role_code   TEXT NOT NULL,
    scope_type  TEXT,
    scope_id    TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(user_id, role_code, scope_type, scope_id)
);

CREATE TABLE IF NOT EXISTS refresh_tokens (
    id          TEXT PRIMARY KEY,
    user_id     TEXT NOT NULL REFERENCES users(id),
    token_hash  TEXT UNIQUE NOT NULL,
    expires_at  TIMESTAMPTZ NOT NULL,
    revoked_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_refresh_user ON refresh_tokens(user_id, expires_at);

-- 企业
CREATE TABLE IF NOT EXISTS enterprises (
    id              TEXT PRIMARY KEY,
    owner_user_id   TEXT NOT NULL,
    name            TEXT NOT NULL,
    credit_code_ciphertext TEXT,
    license_url     TEXT,
    account_name    TEXT,
    status          TEXT NOT NULL DEFAULT 'pending',
    is_member       BOOLEAN NOT NULL DEFAULT false,
    city_code       TEXT,
    version         INT NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_enterprises_owner ON enterprises(owner_user_id);
CREATE INDEX IF NOT EXISTS idx_enterprises_status_city ON enterprises(status, city_code);

-- 需求
CREATE TABLE IF NOT EXISTS demands (
    id              TEXT PRIMARY KEY,
    publisher_id    TEXT NOT NULL,
    publisher_name  TEXT NOT NULL,
    contact         TEXT NOT NULL,
    district        TEXT,
    city_code       TEXT,
    project_type    TEXT,
    title           TEXT NOT NULL,
    description     TEXT,
    images          JSONB DEFAULT '[]',
    latitude        DOUBLE PRECISION DEFAULT 0,
    longitude       DOUBLE PRECISION DEFAULT 0,
    budget_fen      BIGINT DEFAULT 0,
    status          TEXT NOT NULL DEFAULT 'pending',
    accepted_application_id TEXT,
    version         INT NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_demands_status ON demands(status, city_code, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_demands_publisher ON demands(publisher_id);

-- 用工
CREATE TABLE IF NOT EXISTS employment_requests (
    id              TEXT PRIMARY KEY,
    enterprise_id   TEXT NOT NULL,
    position        TEXT NOT NULL,
    headcount       INT NOT NULL DEFAULT 1,
    start_date      TIMESTAMPTZ,
    end_date        TIMESTAMPTZ,
    status          TEXT NOT NULL DEFAULT 'pending',
    version         INT NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_employment_enterprise ON employment_requests(enterprise_id);

-- 合同
CREATE TABLE IF NOT EXISTS contracts (
    id              TEXT PRIMARY KEY,
    enterprise_id   TEXT NOT NULL,
    template_id     TEXT,
    sign_url        TEXT,
    status          TEXT NOT NULL DEFAULT 'draft',
    version         INT NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_contracts_enterprise ON contracts(enterprise_id);

-- 审计日志
CREATE TABLE IF NOT EXISTS audit_logs (
    id              TEXT PRIMARY KEY,
    actor_id        TEXT NOT NULL,
    action          TEXT NOT NULL,
    resource_type   TEXT NOT NULL,
    resource_id     TEXT NOT NULL,
    result          TEXT,
    request_id      TEXT,
    metadata        JSONB DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_audit_resource ON audit_logs(resource_type, resource_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_actor ON audit_logs(actor_id, created_at DESC);

-- 幂等键
CREATE TABLE IF NOT EXISTS idempotency_keys (
    user_id         TEXT NOT NULL,
    key             TEXT NOT NULL,
    response_status INT NOT NULL,
    response_body   TEXT,
    expires_at      TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY(user_id, key)
);
