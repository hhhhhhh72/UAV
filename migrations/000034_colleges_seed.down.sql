-- 移除院校初始数据（按 seed 写入的 id）
DELETE FROM colleges WHERE id IN ('college-1','college-2','college-3','college-4','college-5','college-6','college-7','college-8');
