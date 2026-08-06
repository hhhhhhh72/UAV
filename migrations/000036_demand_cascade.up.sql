-- 需求删除时级联清理关联数据（demand_bids 报价 / demand_intents 对接意向）
-- 管理端删除已取消/已驳回需求时，避免外键约束阻止删除或残留孤儿数据
ALTER TABLE demand_bids DROP CONSTRAINT IF EXISTS demand_bids_demand_id_fkey;
ALTER TABLE demand_bids ADD CONSTRAINT demand_bids_demand_id_fkey
    FOREIGN KEY (demand_id) REFERENCES demands(id) ON DELETE CASCADE;

ALTER TABLE demand_intents DROP CONSTRAINT IF EXISTS demand_intents_demand_id_fkey;
ALTER TABLE demand_intents ADD CONSTRAINT demand_intents_demand_id_fkey
    FOREIGN KEY (demand_id) REFERENCES demands(id) ON DELETE CASCADE;
