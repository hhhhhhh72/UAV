-- 活动报名并发重复：同一用户对同一活动唯一约束（对齐 000071 赛事报名）。
-- 先清理存量重复（保留最早一条，即 id 最小者），再建唯一索引。
DELETE FROM event_registrations a
USING event_registrations b
WHERE a.id > b.id
  AND a.event_id = b.event_id
  AND a.user_id = b.user_id;

CREATE UNIQUE INDEX IF NOT EXISTS uniq_event_regs_user_event
  ON event_registrations(event_id, user_id);
