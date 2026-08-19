-- 赛事划线原价（对齐小程序赛事列表/详情页 original_fee 展示）
ALTER TABLE competitions ADD COLUMN IF NOT EXISTS original_fee INT NOT NULL DEFAULT 0;
