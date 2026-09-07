-- 需求发布补差字段（PRD FR-5.1）：预算下限 / 附件（PDF 图 ≤10MB）/ 机型要求 / 飞手数量
ALTER TABLE demands
    ADD COLUMN IF NOT EXISTS budget_min_fen BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS attachments JSONB NOT NULL DEFAULT '[]'::jsonb,
    ADD COLUMN IF NOT EXISTS aircraft JSONB NOT NULL DEFAULT '[]'::jsonb,
    ADD COLUMN IF NOT EXISTS pilot_count INT NOT NULL DEFAULT 0;
