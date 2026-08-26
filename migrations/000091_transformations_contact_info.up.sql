-- 成果转化联系方式（track 页「联系发布方」复制门）
ALTER TABLE transformations ADD COLUMN IF NOT EXISTS contact_info TEXT NOT NULL DEFAULT '';
