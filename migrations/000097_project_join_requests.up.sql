-- 课题攻关参与申请（POST /api/v1/projects/{id}/join）：
-- 用户在课题详情页申请参与攻关，协会后台评估（pending→contacted→closed）
CREATE TABLE IF NOT EXISTS project_join_requests (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  user_id TEXT NOT NULL,
  org_name TEXT NOT NULL DEFAULT '',
  message TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL DEFAULT 'pending',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 同一用户对同一课题只有一条申请记录（closed 后允许同一记录重新提交，不新建）
CREATE UNIQUE INDEX IF NOT EXISTS project_join_requests_project_user
  ON project_join_requests (project_id, user_id);

CREATE INDEX IF NOT EXISTS project_join_requests_project
  ON project_join_requests (project_id);
