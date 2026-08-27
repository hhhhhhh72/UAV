-- 低空研学报名表（研学详情→报名→我的报名→管理端审核 闭环；此前报名误提交到"服务申请"）
CREATE TABLE IF NOT EXISTS study_tour_enrollments (
  id TEXT PRIMARY KEY,
  tour_id TEXT NOT NULL,
  user_id TEXT NOT NULL,
  name TEXT NOT NULL DEFAULT '',
  phone TEXT NOT NULL DEFAULT '',
  adult_count INTEGER NOT NULL DEFAULT 1,
  child_count INTEGER NOT NULL DEFAULT 0,
  remark TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL DEFAULT 'pending',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS study_tour_enrollments_tour ON study_tour_enrollments (tour_id, created_at DESC);
CREATE INDEX IF NOT EXISTS study_tour_enrollments_user ON study_tour_enrollments (user_id, created_at DESC);
