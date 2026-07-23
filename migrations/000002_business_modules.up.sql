-- 业务模块扩展迁移

-- 1. 改造 demands 表 (先删后建，开发期)
ALTER TABLE demands ADD COLUMN IF NOT EXISTS city_code TEXT;
ALTER TABLE demands ADD COLUMN IF NOT EXISTS biz_type TEXT DEFAULT 'other';
ALTER TABLE demands ADD COLUMN IF NOT EXISTS budget_fen BIGINT DEFAULT 0;
ALTER TABLE demands ADD COLUMN IF NOT EXISTS biz_fields JSONB DEFAULT '{}';

-- 2. 报价竞标
CREATE TABLE IF NOT EXISTS demand_bids (
    id          TEXT PRIMARY KEY,
    demand_id   TEXT NOT NULL REFERENCES demands(id),
    bidder_id   TEXT NOT NULL,
    bidder_name TEXT NOT NULL,
    amount_fen  BIGINT NOT NULL DEFAULT 0,
    proposal    TEXT,
    status      TEXT NOT NULL DEFAULT 'pending',
    version     INT NOT NULL DEFAULT 1,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_bids_demand ON demand_bids(demand_id);

-- 3. 培训证书
CREATE TABLE IF NOT EXISTS certificates (
    id          TEXT PRIMARY KEY,
    user_id     TEXT NOT NULL,
    cert_type   TEXT NOT NULL,
    cert_number TEXT,
    level       TEXT,
    issue_date  TIMESTAMPTZ,
    expire_date TIMESTAMPTZ,
    issuer_org  TEXT,
    image_url   TEXT,
    status      TEXT NOT NULL DEFAULT 'pending',
    version     INT NOT NULL DEFAULT 1,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_certs_user ON certificates(user_id);

-- 4. 培训课程
CREATE TABLE IF NOT EXISTS training_courses (
    id             TEXT PRIMARY KEY,
    org_id         TEXT NOT NULL,
    title          TEXT NOT NULL,
    cert_type      TEXT,
    description    TEXT,
    start_date     TIMESTAMPTZ,
    end_date       TIMESTAMPTZ,
    max_students   INT DEFAULT 0,
    enrolled_count INT DEFAULT 0,
    location       TEXT,
    price_fen      BIGINT DEFAULT 0,
    status         TEXT NOT NULL DEFAULT 'draft',
    version        INT NOT NULL DEFAULT 1,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 5. 教练
CREATE TABLE IF NOT EXISTS instructors (
    id          TEXT PRIMARY KEY,
    user_id     TEXT NOT NULL UNIQUE,
    name        TEXT NOT NULL,
    cert_types  JSONB DEFAULT '[]',
    bio         TEXT,
    org_id      TEXT,
    status      TEXT NOT NULL DEFAULT 'pending',
    version     INT NOT NULL DEFAULT 1,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 6. 认证飞手
CREATE TABLE IF NOT EXISTS certified_pilots (
    id             TEXT PRIMARY KEY,
    user_id        TEXT NOT NULL UNIQUE,
    real_name      TEXT NOT NULL,
    id_card        TEXT,
    cert_ids       JSONB DEFAULT '[]',
    flight_hours   INT DEFAULT 0,
    rating         DOUBLE PRECISION DEFAULT 0,
    completed_jobs INT DEFAULT 0,
    status         TEXT NOT NULL DEFAULT 'pending',
    version        INT NOT NULL DEFAULT 1,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 7. 无人机商品
CREATE TABLE IF NOT EXISTS drone_products (
    id          TEXT PRIMARY KEY,
    seller_id   TEXT NOT NULL,
    seller_name TEXT NOT NULL,
    prod_type   TEXT NOT NULL,
    title       TEXT NOT NULL,
    description TEXT,
    price_fen   BIGINT DEFAULT 0,
    images      JSONB DEFAULT '[]',
    brand       TEXT,
    model       TEXT,
    condition   TEXT DEFAULT 'new',
    status      TEXT NOT NULL DEFAULT 'listed',
    version     INT NOT NULL DEFAULT 1,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_products_seller ON drone_products(seller_id);
CREATE INDEX IF NOT EXISTS idx_products_type ON drone_products(prod_type);

-- 8. 维修订单
CREATE TABLE IF NOT EXISTS repair_orders (
    id           TEXT PRIMARY KEY,
    customer_id  TEXT NOT NULL,
    product_desc TEXT,
    fault_desc   TEXT,
    quote_fen    BIGINT DEFAULT 0,
    status       TEXT NOT NULL DEFAULT 'submitted',
    technician   TEXT,
    version      INT NOT NULL DEFAULT 1,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 9. 保险保单
CREATE TABLE IF NOT EXISTS insurance_policies (
    id          TEXT PRIMARY KEY,
    user_id     TEXT NOT NULL,
    drone_model TEXT,
    drone_sn    TEXT,
    policy_type TEXT,
    premium_fen BIGINT DEFAULT 0,
    coverage_fen BIGINT DEFAULT 0,
    start_date  TIMESTAMPTZ,
    end_date    TIMESTAMPTZ,
    insurer     TEXT,
    status      TEXT NOT NULL DEFAULT 'active',
    version     INT NOT NULL DEFAULT 1,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_policies_user ON insurance_policies(user_id);

-- 10. 年审
CREATE TABLE IF NOT EXISTS annual_inspections (
    id           TEXT PRIMARY KEY,
    user_id      TEXT NOT NULL,
    drone_model  TEXT,
    drone_sn     TEXT,
    inspect_date TIMESTAMPTZ,
    expire_date  TIMESTAMPTZ,
    result       TEXT,
    report_url   TEXT,
    status       TEXT NOT NULL DEFAULT 'pending',
    version      INT NOT NULL DEFAULT 1,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 11. 分期贷款
CREATE TABLE IF NOT EXISTS loan_applications (
    id             TEXT PRIMARY KEY,
    user_id        TEXT NOT NULL,
    amount_fen     BIGINT DEFAULT 0,
    term_months    INT DEFAULT 12,
    purpose        TEXT,
    status         TEXT NOT NULL DEFAULT 'submitted',
    approved_fen   BIGINT DEFAULT 0,
    monthly_pay_fen BIGINT DEFAULT 0,
    version        INT NOT NULL DEFAULT 1,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);
