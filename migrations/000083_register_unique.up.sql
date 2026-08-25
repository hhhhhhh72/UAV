-- 赛事/活动报名防超卖与重复（并发兜底）：service 层判重是 check-then-insert，
-- 无唯一约束时并发双请求可重复报名——与 000078 展位同一套补丁。
-- 先清理存量重复（保留 id 最小即最早一条），再建唯一索引。

DELETE FROM competition_registrations a
USING competition_registrations b
WHERE a.id > b.id
  AND a.competition_id = b.competition_id
  AND a.user_id = b.user_id;

CREATE UNIQUE INDEX IF NOT EXISTS uniq_competition_regs_user
  ON competition_registrations(competition_id, user_id);

DELETE FROM event_registrations a
USING event_registrations b
WHERE a.id > b.id
  AND a.event_id = b.event_id
  AND a.user_id = b.user_id;

CREATE UNIQUE INDEX IF NOT EXISTS uniq_event_regs_user
  ON event_registrations(event_id, user_id);
