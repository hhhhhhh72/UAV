-- uploads 补列：domain.FileRecord 的 StorageKey/SHA256/ContentType 此前被仓储静默丢弃
--（INSERT/FindByID 均未读写这三列，回读恒空）。
ALTER TABLE uploads ADD COLUMN IF NOT EXISTS storage_key TEXT NOT NULL DEFAULT '';
ALTER TABLE uploads ADD COLUMN IF NOT EXISTS sha256 TEXT NOT NULL DEFAULT '';
ALTER TABLE uploads ADD COLUMN IF NOT EXISTS content_type TEXT NOT NULL DEFAULT '';
