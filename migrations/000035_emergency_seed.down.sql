-- 移除应急初始数据
DELETE FROM rescue_cases WHERE id IN ('rcase-1','rcase-2','rcase-3','rcase-4');
DELETE FROM emergency_depts WHERE id IN ('dept-1','dept-2','dept-3','dept-4','dept-5');
DELETE FROM emergency_dispatches WHERE id IN ('dsp-1','dsp-2','dsp-3');
