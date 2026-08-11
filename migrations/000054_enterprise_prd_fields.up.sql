-- PRD FR-2.1 会员企业档案补齐字段：
-- contact_person 联系人 / email 邮箱 / founded_at 成立时间 / capability_tags 能力标签（逗号分隔，预设标签库多选）
-- 注意：迁移系统每次全量重跑，必须幂等（IF NOT EXISTS）
ALTER TABLE enterprises
    ADD COLUMN IF NOT EXISTS contact_person  TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS email           TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS founded_at      TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS capability_tags TEXT NOT NULL DEFAULT '';
