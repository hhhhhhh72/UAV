-- 回滚：删除性能审查新增索引。
DROP INDEX IF EXISTS idx_jobs_status_created;
DROP INDEX IF EXISTS idx_labour_orders_created;
DROP INDEX IF EXISTS idx_repair_orders_customer;
DROP INDEX IF EXISTS idx_loan_applications_created;
DROP INDEX IF EXISTS idx_resumes_created;
DROP INDEX IF EXISTS idx_training_enrollments_course;
DROP INDEX IF EXISTS idx_job_applications_applicant;
DROP INDEX IF EXISTS idx_enterprises_status_created;
