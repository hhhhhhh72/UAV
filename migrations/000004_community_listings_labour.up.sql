-- 社区 + 二手 + 用工

-- 帖子
CREATE TABLE IF NOT EXISTS posts (
    id          TEXT PRIMARY KEY,
    author_id   TEXT NOT NULL,
    title       TEXT NOT NULL,
    content     TEXT,
    images      JSONB DEFAULT '[]',
    city_code   TEXT,
    status      TEXT NOT NULL DEFAULT 'pending',
    version     INT NOT NULL DEFAULT 1,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_posts_status ON posts(status, created_at DESC);

-- 评论
CREATE TABLE IF NOT EXISTS comments (
    id          TEXT PRIMARY KEY,
    post_id     TEXT NOT NULL,
    author_id   TEXT NOT NULL,
    content     TEXT NOT NULL,
    status      TEXT NOT NULL DEFAULT 'active',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_comments_post ON comments(post_id);

-- 举报
CREATE TABLE IF NOT EXISTS reports (
    id              TEXT PRIMARY KEY,
    reporter_id     TEXT NOT NULL,
    resource_type   TEXT NOT NULL,
    resource_id     TEXT NOT NULL,
    reason          TEXT,
    status          TEXT NOT NULL DEFAULT 'pending',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 二手商品
CREATE TABLE IF NOT EXISTS listings (
    id          TEXT PRIMARY KEY,
    seller_id   TEXT NOT NULL,
    title       TEXT NOT NULL,
    description TEXT,
    category    TEXT,
    price_fen   BIGINT DEFAULT 0,
    images      JSONB DEFAULT '[]',
    district    TEXT,
    status      TEXT NOT NULL DEFAULT 'listed',
    version     INT NOT NULL DEFAULT 1,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_listings_status ON listings(status, created_at DESC);

-- 收藏
CREATE TABLE IF NOT EXISTS listing_favorites (
    listing_id  TEXT NOT NULL,
    user_id     TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY(listing_id, user_id)
);

-- 用工订单
CREATE TABLE IF NOT EXISTS labour_orders (
    id              TEXT PRIMARY KEY,
    employer_id     TEXT NOT NULL,
    title           TEXT NOT NULL,
    description     TEXT,
    worker_count    INT DEFAULT 1,
    start_date      TIMESTAMPTZ,
    end_date        TIMESTAMPTZ,
    budget_fen      BIGINT DEFAULT 0,
    status          TEXT NOT NULL DEFAULT 'draft',
    version         INT NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 报价
CREATE TABLE IF NOT EXISTS labour_quotes (
    id          TEXT PRIMARY KEY,
    order_id    TEXT NOT NULL,
    quoter_id   TEXT NOT NULL,
    quoter_name TEXT,
    amount_fen  BIGINT DEFAULT 0,
    proposal    TEXT,
    status      TEXT NOT NULL DEFAULT 'pending',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 派遣
CREATE TABLE IF NOT EXISTS assignments (
    id          TEXT PRIMARY KEY,
    order_id    TEXT NOT NULL,
    worker_id   TEXT NOT NULL,
    status      TEXT NOT NULL DEFAULT 'assigned',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
