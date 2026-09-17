-- 回滚 000113：重建 association_members。
--
-- 注意：代码侧的 AssociationMember 实体、仓储、服务、handler 都已删除，
-- 重建表**不会**恢复功能——要让功能回来必须同时回滚那些代码改动。
-- 这里只保证 schema 可回滚，避免 down 迁移直接报"表已存在/不存在"。

CREATE TABLE IF NOT EXISTS association_members (
    id            TEXT PRIMARY KEY,
    user_id       TEXT NOT NULL,
    enterprise_id TEXT NOT NULL DEFAULT '',
    role          TEXT NOT NULL DEFAULT 'member',
    join_date     TIMESTAMPTZ DEFAULT NOW(),
    expire_date   TIMESTAMPTZ,
    status        TEXT NOT NULL DEFAULT 'active',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_assoc_members_user ON association_members(user_id);
