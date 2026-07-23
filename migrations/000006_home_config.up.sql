-- 首页配置 + Banner

CREATE TABLE IF NOT EXISTS banners (
    id          TEXT PRIMARY KEY,
    image_url   TEXT NOT NULL,
    link_url    TEXT,
    sort_order  INT DEFAULT 0,
    status      TEXT NOT NULL DEFAULT 'active',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS home_quick_entries (
    id          TEXT PRIMARY KEY,
    key         TEXT NOT NULL,
    name        TEXT NOT NULL,
    icon_url    TEXT,
    link_url    TEXT,
    sort_order  INT DEFAULT 0,
    status      TEXT NOT NULL DEFAULT 'active'
);

-- Seed default entries
INSERT INTO home_quick_entries(id, key, name, sort_order) VALUES
  ('hq-1','demand','无人机需求',1),
  ('hq-2','enterprise','入驻商家',2),
  ('hq-3','community','同城社区',3),
  ('hq-4','jobs','求职招聘',4)
ON CONFLICT DO NOTHING;

INSERT INTO banners(id, image_url, link_url, sort_order) VALUES
  ('banner-1','https://img.example.com/banner1.jpg','/pages/demand/list',1),
  ('banner-2','https://img.example.com/banner2.jpg','/pages/training/list',2)
ON CONFLICT DO NOTHING;
