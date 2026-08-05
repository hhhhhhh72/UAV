-- 需求对接意向登记（联系对接模式：记录"谁对哪条需求产生对接意向"，
-- 作为撮合服务费的数据依据与管理端登记成交金额的触发点）
CREATE TABLE IF NOT EXISTS demand_intents (
    id          TEXT PRIMARY KEY,
    demand_id   TEXT NOT NULL,
    intentor_id TEXT NOT NULL,
    intentor_name TEXT NOT NULL,
    contact     TEXT NOT NULL,
    remark      TEXT NOT NULL DEFAULT '',
    status      TEXT NOT NULL DEFAULT 'pending', -- pending / contacted / done / closed
    version     INT  NOT NULL DEFAULT 1,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_demand_intents_demand ON demand_intents(demand_id);
CREATE INDEX IF NOT EXISTS idx_demand_intents_intentor ON demand_intents(intentor_id);
