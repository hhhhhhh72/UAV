-- 在 users 表添加 role 列，统一角色存储
ALTER TABLE users ADD COLUMN IF NOT EXISTS role TEXT NOT NULL DEFAULT 'individual';

-- 为现有用户补充默认角色
UPDATE users SET role = 'individual' WHERE role = '';
