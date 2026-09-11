-- 一次性数据修复：清理"作者账号已不存在"的历史遗留内容（2026-09-11 上线的注销内容处置之前产生的）
-- 口径与 service.UserContentCleanupPlan 一致：在架内容下架、个人信息擦除/删除、履约记录不动。
\set ON_ERROR_STOP on
BEGIN;

-- 孤儿判定：owner 列非空且 users 表里查不到该账号
CREATE TEMP TABLE orphan_users ON COMMIT DROP AS
SELECT d.publisher_id AS id FROM demands d WHERE coalesce(d.publisher_id,'')<>'' AND NOT EXISTS (SELECT 1 FROM users u WHERE u.id=d.publisher_id)
UNION SELECT p.author_id FROM posts p WHERE coalesce(p.author_id,'')<>'' AND NOT EXISTS (SELECT 1 FROM users u WHERE u.id=p.author_id)
UNION SELECT r.user_id FROM resumes r WHERE coalesce(r.user_id,'')<>'' AND NOT EXISTS (SELECT 1 FROM users u WHERE u.id=r.user_id)
UNION SELECT a.applicant_id FROM job_applications a WHERE coalesce(a.applicant_id,'')<>'' AND NOT EXISTS (SELECT 1 FROM users u WHERE u.id=a.applicant_id)
UNION SELECT m.sender_id FROM messages m WHERE coalesce(m.sender_id,'')<>'' AND NOT EXISTS (SELECT 1 FROM users u WHERE u.id=m.sender_id)
UNION SELECT m.receiver_id FROM messages m WHERE coalesce(m.receiver_id,'')<>'' AND NOT EXISTS (SELECT 1 FROM users u WHERE u.id=m.receiver_id)
UNION SELECT e.owner_user_id FROM enterprises e WHERE coalesce(e.owner_user_id,'')<>'' AND NOT EXISTS (SELECT 1 FROM users u WHERE u.id=e.owner_user_id);

SELECT '孤儿账号数' AS item, count(*)::text AS n FROM orphan_users;

UPDATE demands d SET status='cancelled'
 WHERE d.status='published' AND d.publisher_id IN (SELECT id FROM orphan_users);
SELECT '下架孤儿需求' AS item, count(*)::text AS n FROM demands d
 WHERE d.status='cancelled' AND d.publisher_id IN (SELECT id FROM orphan_users);

UPDATE demands d SET contact='', publisher_name=''
 WHERE d.publisher_id IN (SELECT id FROM orphan_users);
SELECT '擦除孤儿需求个人信息' AS item, count(*)::text AS n FROM demands d WHERE d.publisher_id IN (SELECT id FROM orphan_users);

UPDATE posts p SET status='draft' WHERE p.status='published' AND p.author_id IN (SELECT id FROM orphan_users);
SELECT '下架孤儿动态' AS item, count(*)::text AS n FROM posts p WHERE p.status='draft' AND p.author_id IN (SELECT id FROM orphan_users);

UPDATE enterprises e SET legal_person='', contact_person='', contact_phone='', email='', license_url=''
 WHERE e.owner_user_id IN (SELECT id FROM orphan_users);
SELECT '擦除孤儿企业联系人' AS item, count(*)::text AS n FROM enterprises e WHERE e.owner_user_id IN (SELECT id FROM orphan_users);

DELETE FROM resumes r WHERE r.user_id IN (SELECT id FROM orphan_users);
SELECT '删除孤儿简历' AS item, count(*)::text AS n FROM resumes r WHERE r.user_id IN (SELECT id FROM orphan_users);

DELETE FROM job_applications a WHERE a.applicant_id IN (SELECT id FROM orphan_users);
SELECT '删除孤儿投递' AS item, count(*)::text AS n FROM job_applications a WHERE a.applicant_id IN (SELECT id FROM orphan_users);

DELETE FROM messages m WHERE m.sender_id IN (SELECT id FROM orphan_users) OR m.receiver_id IN (SELECT id FROM orphan_users);
SELECT '剩余孤儿站内信' AS item, count(*)::text AS n FROM messages m WHERE m.sender_id IN (SELECT id FROM orphan_users) OR m.receiver_id IN (SELECT id FROM orphan_users);

DELETE FROM certified_pilots c WHERE c.user_id IN (SELECT id FROM orphan_users);
SELECT '剩余孤儿飞手名录' AS item, count(*)::text AS n FROM certified_pilots c WHERE c.user_id IN (SELECT id FROM orphan_users);

-- 履约记录刻意不动：工单/合同/资金/审计
SELECT '保留的孤儿工单（不动）' AS item, count(*)::text AS n FROM work_orders w WHERE w.publisher_id IN (SELECT id FROM orphan_users);

COMMIT;
