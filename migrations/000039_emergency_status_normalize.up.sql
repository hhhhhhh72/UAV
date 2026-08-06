-- 应急模块历史数据状态枚举规范化（旧枚举 → 业务枚举，使管理端筛选可命中历史数据）
-- 调度记录：active → in_progress（执行中）、done → completed（已完成）
UPDATE emergency_dispatches SET status = 'in_progress' WHERE status = 'active';
UPDATE emergency_dispatches SET status = 'completed' WHERE status = 'done';
-- 应急资源：deployed（已部署）≈ 使用中
UPDATE emergency_resources SET status = 'in_use' WHERE status = 'deployed';
