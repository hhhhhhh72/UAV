-- 回滚：删除种子数据（仅删除本迁移插入的固定 id 数据）
DELETE FROM venues WHERE id IN ('venue-001', 'venue-002', 'venue-003');
DELETE FROM resource_pools WHERE id IN ('pool-001', 'pool-002', 'pool-003');
DELETE FROM emergency_drills WHERE id IN ('drill-001', 'drill-002', 'drill-003');
DELETE FROM exhibitions WHERE id IN ('exhib-001', 'exhib-002', 'exhib-003');
DELETE FROM compliance_docs WHERE id IN ('cdoc-001', 'cdoc-002', 'cdoc-003', 'cdoc-004');
DELETE FROM posts WHERE id IN ('post-001', 'post-002', 'post-003');
DELETE FROM case_entries WHERE id IN ('case-001', 'case-002', 'case-003');
DELETE FROM experts WHERE id IN ('expert-001', 'expert-002', 'expert-003', 'expert-004');

-- 回滚状态回填（有损：无法精确区分业务态与原状态，仅还原枚举口径）
UPDATE training_courses SET status = 'draft' WHERE status = 'published';
UPDATE employment_requests SET status = '' WHERE status = 'pending';
UPDATE escrow_transactions SET status = '' WHERE status = 'pending';
