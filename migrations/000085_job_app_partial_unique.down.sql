DROP INDEX IF EXISTS idx_app_unique;
CREATE UNIQUE INDEX IF NOT EXISTS idx_app_unique
  ON job_applications (job_id, applicant_id);
