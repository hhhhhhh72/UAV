ALTER TABLE demands
    DROP COLUMN IF EXISTS budget_min_fen,
    DROP COLUMN IF EXISTS attachments,
    DROP COLUMN IF EXISTS aircraft,
    DROP COLUMN IF EXISTS pilot_count;
