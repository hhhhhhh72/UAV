-- 接单派单闭环（PRD FR-6.2~6.5）：企业确认接单后生成作业订单
CREATE TABLE IF NOT EXISTS work_orders (
    id            TEXT PRIMARY KEY,
    order_no      TEXT NOT NULL UNIQUE,           -- 订单编号，如 WO2601xxxxxx
    demand_id     TEXT NOT NULL REFERENCES demands(id),
    publisher_id  TEXT NOT NULL REFERENCES users(id), -- 需求方（企业）
    publisher_name TEXT NOT NULL DEFAULT '',
    worker_id     TEXT NOT NULL REFERENCES users(id), -- 接单飞手
    worker_name   TEXT NOT NULL DEFAULT '',
    amount_fen    BIGINT NOT NULL DEFAULT 0,      -- 订单金额（企业确认接单时填写，面议为 0）
    status        TEXT NOT NULL DEFAULT 'pending',-- pending/ongoing/awaiting_accept/completed/cancelled
    result_photos JSONB NOT NULL DEFAULT '[]',    -- 作业成果照片（飞手确认完成时上传）
    rework_note   TEXT NOT NULL DEFAULT '',       -- 企业整改要求
    cancel_reason TEXT NOT NULL DEFAULT '',       -- 取消原因
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_work_orders_demand ON work_orders(demand_id);
CREATE INDEX IF NOT EXISTS idx_work_orders_publisher ON work_orders(publisher_id);
CREATE INDEX IF NOT EXISTS idx_work_orders_worker ON work_orders(worker_id);
CREATE INDEX IF NOT EXISTS idx_work_orders_status ON work_orders(status);
