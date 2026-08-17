-- 000066 修正 000065 的唯一索引谓词（排除空串 intent_id）：
--   已应用 000065 的库，旧索引按 "IS NOT NULL" 会把 '' 视为冲突值；
--   重建为 "IS NOT NULL AND <> ''"。

DROP INDEX IF EXISTS uniq_work_orders_intent;

CREATE UNIQUE INDEX IF NOT EXISTS uniq_work_orders_intent
    ON work_orders (intent_id) WHERE intent_id IS NOT NULL AND intent_id <> '';
