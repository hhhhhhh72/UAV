-- 回滚：业务枚举 → 旧枚举（仅当需要还原历史数据时执行）
UPDATE emergency_dispatches SET status = 'active' WHERE status = 'in_progress';
UPDATE emergency_dispatches SET status = 'done' WHERE status = 'completed';
UPDATE emergency_resources SET status = 'deployed' WHERE status = 'in_use';
