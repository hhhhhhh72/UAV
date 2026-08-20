-- P2：赛事报名并发重复——同一用户对同一赛事唯一约束。
-- 此前依赖 service 层 ListRegs 预检（TOCTOU），并发重复报名可双双落库。
-- 先清理存量重复（保留最早一条，即 id 最小者），再建唯一索引，否则建索引会失败。
DELETE FROM competition_registrations a
USING competition_registrations b
WHERE a.id > b.id
  AND a.competition_id = b.competition_id
  AND a.user_id = b.user_id;

CREATE UNIQUE INDEX IF NOT EXISTS uniq_competition_regs_user_comp
  ON competition_registrations(competition_id, user_id);
