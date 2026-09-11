-- 企业案例：补齐独立的视频字段
--
-- 背景：小程序「企业案例 → 案例详情」卡片本来就有视频播放能力，但案例模型只有 images 图片数组，
-- 后台表单也只收图片，视频无处可放 → 详情里永远没有视频可播。
-- 这里给案例补上视频地址与可选封面；小程序按 video_url 渲染 <video>。
ALTER TABLE case_entries ADD COLUMN IF NOT EXISTS video_url varchar(500) NOT NULL DEFAULT '';
ALTER TABLE case_entries ADD COLUMN IF NOT EXISTS video_poster_url varchar(500) NOT NULL DEFAULT '';
