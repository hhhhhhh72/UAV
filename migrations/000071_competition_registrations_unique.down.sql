-- 回滚：删除赛事报名唯一索引（重复防护随之解除）。
DROP INDEX IF EXISTS uniq_competition_regs_user_comp;
