-- 存量库清理：000076/000078 建唯一点前未去重的版本已在早期存量库应用，
-- 此迁移兜底两类重复数据（历史版本只在无重复的库上能应用成功，此处幂等清理）。
DELETE FROM certificates a
USING certificates b
WHERE a.id > b.id
  AND a.cert_number = b.cert_number
  AND a.cert_number <> '';

DELETE FROM exhibition_booths a
USING exhibition_booths b
WHERE a.id > b.id
  AND a.exhibition_id = b.exhibition_id
  AND a.exhibitor_id = b.exhibitor_id;

DELETE FROM exhibition_booths a
USING exhibition_booths b
WHERE a.id > b.id
  AND a.exhibition_id = b.exhibition_id
  AND a.booth_number = b.booth_number
  AND a.booth_number <> '';
