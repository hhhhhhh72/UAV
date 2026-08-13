-- C13: 赛事报名扩展字段。
-- 小程序 register.vue 提交 name/phone/id_card/photo_url/id_card_image，
-- 旧表只有 team_name/member_count/contact_info，参赛人信息与证件影像被丢弃。
ALTER TABLE competition_registrations ADD COLUMN IF NOT EXISTS name TEXT NOT NULL DEFAULT '';
ALTER TABLE competition_registrations ADD COLUMN IF NOT EXISTS phone TEXT NOT NULL DEFAULT '';
ALTER TABLE competition_registrations ADD COLUMN IF NOT EXISTS id_card TEXT NOT NULL DEFAULT '';
ALTER TABLE competition_registrations ADD COLUMN IF NOT EXISTS photo_url TEXT NOT NULL DEFAULT '';
ALTER TABLE competition_registrations ADD COLUMN IF NOT EXISTS id_card_image TEXT NOT NULL DEFAULT '';
