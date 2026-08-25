-- 令牌版本：删除/封禁/改角色时自增，使已签发 token 立即失效（authenticate 校验）。
ALTER TABLE users ADD COLUMN IF NOT EXISTS token_version BIGINT NOT NULL DEFAULT 0;
