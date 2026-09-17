-- 上传台账记录图片宽高：供给大厅卡片要按每张图自己的比例预留位置。
--
-- 为什么需要（2026-09-17）：图片框此前写死 1:1 并用 aspectFill 硬裁，
-- 而卡片要「不裁图 + 不跳动」就必须在渲染前知道每张图的比例——
-- 只知道 URL 是无法预留空间的，加载完再撑开会造成列表跳动（CLS）。
-- 存量行为 0，由回填脚本按磁盘上的文件补齐；0 表示未知，前端退化为 1:1。
ALTER TABLE uploads ADD COLUMN IF NOT EXISTS width  INT NOT NULL DEFAULT 0;
ALTER TABLE uploads ADD COLUMN IF NOT EXISTS height INT NOT NULL DEFAULT 0;
