-- 撤掉「视频封面」列：案例只需要一个视频地址，封面图沿用 images[0]（产品确认不需要独立视频封面）。
-- 000102 里新加的 video_poster_url 从未被写入过，直接删除。
ALTER TABLE case_entries DROP COLUMN IF EXISTS video_poster_url;
