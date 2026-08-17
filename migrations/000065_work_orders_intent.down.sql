DROP INDEX IF EXISTS uniq_work_orders_intent;
ALTER TABLE work_orders DROP COLUMN IF EXISTS intent_id;
