SELECT current_database(), now();
SELECT max(version) AS schema_version FROM schema_migrations;
SELECT count(*) AS intents FROM demand_intents;
SELECT count(*) AS work_orders FROM work_orders;
SELECT count(*) AS enrolls FROM training_enrollments;
SELECT count(*) AS dup_pending_intents FROM (SELECT demand_id,intentor_id FROM demand_intents WHERE status='pending' GROUP BY demand_id,intentor_id HAVING count(*)>1) t;
SELECT count(*) AS dup_enrolls FROM (SELECT user_id,course_id FROM training_enrollments GROUP BY user_id,course_id HAVING count(*)>1) t;
SELECT count(*) AS users FROM users;
SELECT count(*) AS reviews FROM reviews;
