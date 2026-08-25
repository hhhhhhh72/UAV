-- 撤回后允许重投（jobs.go "撤回后允许重投"）：job_applications 唯一索引须排除
-- withdrawn——此前整表唯一，撤回后重投同一职位触发唯一键 500。
DROP INDEX IF EXISTS idx_app_unique;
CREATE UNIQUE INDEX IF NOT EXISTS idx_app_unique
  ON job_applications (job_id, applicant_id)
  WHERE status <> 'withdrawn';
