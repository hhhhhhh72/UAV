-- 性能审查：列表/计数热路径索引。
-- 各 List* 按 (过滤列, created_at DESC) 排序 + COUNT 过滤，未命中索引时全表扫描。
-- 全部 IF NOT EXISTS 幂等，存量库可安全重放。
CREATE INDEX IF NOT EXISTS idx_jobs_status_created ON jobs(status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_labour_orders_created ON labour_orders(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_repair_orders_customer ON repair_orders(customer_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_loan_applications_created ON loan_applications(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_resumes_created ON resumes(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_training_enrollments_course ON training_enrollments(course_id);
CREATE INDEX IF NOT EXISTS idx_job_applications_applicant ON job_applications(applicant_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_enterprises_status_created ON enterprises(status, created_at DESC);
