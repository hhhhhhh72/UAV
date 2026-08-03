-- Batch1-3 模块表回滚
ALTER TABLE jobs DROP COLUMN IF EXISTS job_type;

DROP TABLE IF EXISTS association_members;
DROP TABLE IF EXISTS emergency_drills;
DROP TABLE IF EXISTS emergency_depts;
DROP TABLE IF EXISTS rescue_cases;
DROP TABLE IF EXISTS cooperation_programs;
DROP TABLE IF EXISTS resource_pool_members;
DROP TABLE IF EXISTS resource_pools;
DROP TABLE IF EXISTS test_site_bookings;
DROP TABLE IF EXISTS test_sites;
DROP TABLE IF EXISTS exhibition_booths;
DROP TABLE IF EXISTS exhibitions;
DROP TABLE IF EXISTS study_tours;
DROP TABLE IF EXISTS colleges;
