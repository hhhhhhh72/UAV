-- 回滚：删除种子赛事（仅删除本迁移写入的 comp-1~6）
DELETE FROM competitions WHERE id IN ('comp-1','comp-2','comp-3','comp-4','comp-5','comp-6');
