-- 回滚：按依赖顺序删除
DROP TABLE IF EXISTS idempotency_keys;
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS contracts;
DROP TABLE IF EXISTS employment_requests;
DROP TABLE IF EXISTS demands;
DROP TABLE IF EXISTS enterprises;
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS users;
