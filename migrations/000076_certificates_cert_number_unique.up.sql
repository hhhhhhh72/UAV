-- 证书号唯一（非空）：防并发下 check-then-insert 撞号重复入库；
-- 空串不参与唯一（部分用户不上传证书号）。
CREATE UNIQUE INDEX certificates_cert_number_unique
    ON certificates (cert_number)
    WHERE cert_number <> '';
