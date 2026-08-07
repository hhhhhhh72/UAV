-- 用户昵称（profile 保存接后端 PATCH /api/v1/me）
ALTER TABLE users ADD COLUMN IF NOT EXISTS name TEXT NOT NULL DEFAULT '';
