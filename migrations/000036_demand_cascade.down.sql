-- 回退：需求关联外键恢复为默认行为
ALTER TABLE demand_bids DROP CONSTRAINT IF EXISTS demand_bids_demand_id_fkey;
ALTER TABLE demand_bids ADD CONSTRAINT demand_bids_demand_id_fkey FOREIGN KEY (demand_id) REFERENCES demands(id);
ALTER TABLE demand_intents DROP CONSTRAINT IF EXISTS demand_intents_demand_id_fkey;
