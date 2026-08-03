-- 密码登录：bcrypt hash 持久化到 users 表（此前仅存兼容层 users.json）
ALTER TABLE users ADD COLUMN IF NOT EXISTS password_hash TEXT NOT NULL DEFAULT '';
