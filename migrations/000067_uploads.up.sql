-- 000067 上传文件台账 + 按用户配额（收尾批次）：
--   每次上传落一条记录（id/owner/size/visibility/created_at），
--   配额 = SUM(size_bytes) 按 owner 当日累计（UPLOAD_DAILY_QUOTA_BYTES，默认 50MB/天）。
--   迁移前已上传的历史文件不计入台账（存量数据无需回填）。

CREATE TABLE IF NOT EXISTS uploads (
    id          TEXT PRIMARY KEY,
    owner_id    TEXT NOT NULL,
    size_bytes  BIGINT NOT NULL CHECK (size_bytes >= 0),
    visibility  TEXT NOT NULL DEFAULT 'public',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_uploads_owner_created
    ON uploads (owner_id, created_at);
