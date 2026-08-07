-- 展会展位封面图（ExhibitionList 表单/详情使用，原列缺失）
ALTER TABLE exhibitions ADD COLUMN IF NOT EXISTS cover_url TEXT NOT NULL DEFAULT '';
