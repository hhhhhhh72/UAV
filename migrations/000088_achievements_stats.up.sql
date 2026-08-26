-- 成果统计与状态值域：views/favs 计数（排序「最多浏览/最多收藏」依据）；status 列已存在（新增 hot/new/transformed 值域由后端白名单把关）。
ALTER TABLE achievements ADD COLUMN IF NOT EXISTS views INTEGER NOT NULL DEFAULT 0;
ALTER TABLE achievements ADD COLUMN IF NOT EXISTS favs INTEGER NOT NULL DEFAULT 0;
