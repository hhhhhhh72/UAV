-- C11: 团体标准库补 category 列。
-- 小程序标准页按 国家标准/行业标准/团体标准/企业标准 分类筛选，
-- 旧表无该列导致 PG 侧 WHERE category=$1 报错、分类 tab 永远空列表。
ALTER TABLE standard_docs ADD COLUMN IF NOT EXISTS category TEXT NOT NULL DEFAULT '';

-- 存量数据按标准编号前缀回填分类（GB→国家标准 / HB·MH·JT·YD→行业标准 / T/→团体标准 / Q/→企业标准）
UPDATE standard_docs SET category = CASE
    WHEN standard_no LIKE 'GB%' THEN '国家标准'
    WHEN standard_no LIKE 'HB%' OR standard_no LIKE 'MH%' OR standard_no LIKE 'JT%' OR standard_no LIKE 'YD%' THEN '行业标准'
    WHEN standard_no LIKE 'T/%' THEN '团体标准'
    WHEN standard_no LIKE 'Q/%' THEN '企业标准'
    ELSE category
END WHERE category = '';

CREATE INDEX IF NOT EXISTS idx_standard_docs_category ON standard_docs(category);
