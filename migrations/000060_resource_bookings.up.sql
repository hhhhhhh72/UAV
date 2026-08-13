-- C11: 产业资源预约表。
-- 小程序资源详情页预约弹窗 → POST /api/v1/industry-resources/{id}/book（旧后端无此路由，必 404）。
CREATE TABLE IF NOT EXISTS industry_resource_bookings (
    id TEXT PRIMARY KEY,
    resource_id TEXT NOT NULL,
    user_id TEXT NOT NULL DEFAULT '',
    booking_date TEXT NOT NULL DEFAULT '',
    purpose TEXT NOT NULL DEFAULT '',
    contact_name TEXT NOT NULL DEFAULT '',
    contact_phone TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_resource_bookings_resource ON industry_resource_bookings(resource_id);
CREATE INDEX IF NOT EXISTS idx_resource_bookings_user ON industry_resource_bookings(user_id);
