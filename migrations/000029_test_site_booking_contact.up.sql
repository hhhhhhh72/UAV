-- 场地预约联系人（小程序预约表单字段，原无处存储）
ALTER TABLE test_site_bookings ADD COLUMN IF NOT EXISTS contact_name TEXT NOT NULL DEFAULT '';
ALTER TABLE test_site_bookings ADD COLUMN IF NOT EXISTS contact_phone TEXT NOT NULL DEFAULT '';
