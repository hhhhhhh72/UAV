-- 000065 工单防并发双建单（B 批加固）：
--   work_orders 增加 intent_id 列 + 唯一索引：同一意向只能生成一张工单。
--   配合 intentRepo.UpdateStatus 的 CAS（仅 pending 可流转）消除并发双建单。
--   索引排除空串：存量/非意向工单 intent_id='' 不参与唯一约束。

ALTER TABLE work_orders ADD COLUMN IF NOT EXISTS intent_id TEXT;

CREATE UNIQUE INDEX IF NOT EXISTS uniq_work_orders_intent
    ON work_orders (intent_id) WHERE intent_id IS NOT NULL AND intent_id <> '';
