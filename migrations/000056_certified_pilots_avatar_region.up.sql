-- 认证飞手补齐字段（认证飞手-后端待补清单）：
-- avatar 头像 URL / region 所在地区
ALTER TABLE certified_pilots
    ADD COLUMN IF NOT EXISTS avatar TEXT DEFAULT '',
    ADD COLUMN IF NOT EXISTS region TEXT DEFAULT '';
