-- 成果转化表
-- 修复：原文件对不存在的 transformations 表执行 ALTER，导致 PG 迁移必失败。
-- 改为幂等建表（含全部列），000014 未建该表，此处补齐。
CREATE TABLE IF NOT EXISTS transformations (
    id             TEXT PRIMARY KEY,
    title          TEXT NOT NULL DEFAULT '',
    achievement_id TEXT NOT NULL DEFAULT '',
    owner_id       TEXT NOT NULL DEFAULT '',
    progress       TEXT DEFAULT '',
    partner_id     TEXT DEFAULT '',
    status         TEXT NOT NULL DEFAULT 'active',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
