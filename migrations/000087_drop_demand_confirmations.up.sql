-- demands.confirmations 死列（需求完成双确认功能未实现、零 Go 引用）：删列收尾。
ALTER TABLE demands DROP COLUMN IF EXISTS confirmations;
