-- 场地预约扩展字段（booking 页提交 9 字段此前被后端静默丢弃）
ALTER TABLE test_site_bookings ADD COLUMN IF NOT EXISTS booking_type TEXT NOT NULL DEFAULT '';
ALTER TABLE test_site_bookings ADD COLUMN IF NOT EXISTS model TEXT NOT NULL DEFAULT '';
ALTER TABLE test_site_bookings ADD COLUMN IF NOT EXISTS license_url TEXT NOT NULL DEFAULT '';
ALTER TABLE test_site_bookings ADD COLUMN IF NOT EXISTS team_name TEXT NOT NULL DEFAULT '';
ALTER TABLE test_site_bookings ADD COLUMN IF NOT EXISTS people_count INTEGER NOT NULL DEFAULT 0;
ALTER TABLE test_site_bookings ADD COLUMN IF NOT EXISTS equipment_list TEXT NOT NULL DEFAULT '';
ALTER TABLE test_site_bookings ADD COLUMN IF NOT EXISTS qualification_url TEXT NOT NULL DEFAULT '';
ALTER TABLE test_site_bookings ADD COLUMN IF NOT EXISTS equipment_note TEXT NOT NULL DEFAULT '';
ALTER TABLE test_site_bookings ADD COLUMN IF NOT EXISTS time_slots TEXT NOT NULL DEFAULT '';
