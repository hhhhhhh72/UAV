-- 需求完成双确认：持久化确认者列表（替代进程内 map）
ALTER TABLE demands ADD COLUMN IF NOT EXISTS confirmations JSONB NOT NULL DEFAULT '[]';
