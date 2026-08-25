-- 证书号唯一（非空）：防并发下 check-then-insert 撞号重复入库；
-- 空串不参与唯一（部分用户不上传证书号）。
-- 先清理存量重复（保留 id 最小=最早一条），否则存量重复会让建索引失败、
-- 而迁移在启动时单事务执行、失败即中止服务（部署阻断）。
DELETE FROM certificates a
USING certificates b
WHERE a.id > b.id
  AND a.cert_number = b.cert_number
  AND a.cert_number <> '';

CREATE UNIQUE INDEX certificates_cert_number_unique
    ON certificates (cert_number)
    WHERE cert_number <> '';
