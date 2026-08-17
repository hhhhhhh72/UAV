-- 000064 幂等去重唯一索引（上线评审 P1）：
--   意向登记 (demand_id, intentor_id) WHERE status='pending' 部分唯一索引
--   （已确认/关闭的旧意向允许再次登记，只防重复提交待处理意向）；
--   培训报名 (user_id, course_id) 全量唯一索引（同一用户同一课程只报一次名）。
--   建索引前清理存量重复数据（保留最早一条）。

-- 清理存量重复 pending 意向（保留 id 最小的一条）
DELETE FROM demand_intents a
USING demand_intents b
WHERE a.id > b.id
  AND a.demand_id = b.demand_id
  AND a.intentor_id = b.intentor_id
  AND a.status = 'pending'
  AND b.status = 'pending';

-- 清理存量重复报名（保留 id 最小的一条）
DELETE FROM training_enrollments a
USING training_enrollments b
WHERE a.id > b.id
  AND a.user_id = b.user_id
  AND a.course_id = b.course_id;

CREATE UNIQUE INDEX IF NOT EXISTS uniq_demand_intents_pending
    ON demand_intents (demand_id, intentor_id) WHERE status = 'pending';

CREATE UNIQUE INDEX IF NOT EXISTS uniq_training_enrollments_user_course
    ON training_enrollments (user_id, course_id);
