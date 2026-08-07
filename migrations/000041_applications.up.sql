-- 服务申请单（小程序 /api/submit 生产存储，替代 dev-only JSON 文件 applications.json）
CREATE TABLE IF NOT EXISTS service_applications (
    id           TEXT PRIMARY KEY,
    user_id      TEXT NOT NULL DEFAULT '',
    service_id   TEXT NOT NULL DEFAULT '',
    service_name TEXT NOT NULL DEFAULT '',
    order_no     TEXT NOT NULL DEFAULT '',
    status       TEXT NOT NULL DEFAULT '待处理',
    apply_time   TEXT NOT NULL DEFAULT '',
    form_data    JSONB NOT NULL DEFAULT '{}',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_service_applications_user ON service_applications (user_id);
CREATE INDEX IF NOT EXISTS idx_service_applications_order_no ON service_applications (order_no);
