-- 行业资讯

CREATE TABLE IF NOT EXISTS articles (
    id          TEXT PRIMARY KEY,
    title       TEXT NOT NULL,
    content     TEXT,
    summary     TEXT,
    category    TEXT NOT NULL, -- policy/news/alert/standard
    source      TEXT,
    author      TEXT,
    is_pinned   BOOLEAN DEFAULT false,
    status      TEXT NOT NULL DEFAULT 'draft',
    version     INT NOT NULL DEFAULT 1,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_articles_cat ON articles(category, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_articles_status ON articles(status, created_at DESC);
