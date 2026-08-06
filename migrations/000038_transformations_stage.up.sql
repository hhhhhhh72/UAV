-- 成果转化表补 stage 列（lab/pilot/industrialized/listed）——此前建表遗漏，SQL 层也从未读写该字段
ALTER TABLE transformations ADD COLUMN IF NOT EXISTS stage TEXT NOT NULL DEFAULT 'lab';
