-- 展位防超卖/重复：同一参展商对同一展会仅能申请一次；展位号在展会内唯一。
-- 先清理存量重复：① 同参展商同展会（保留最早一条，即 id 最小者）；
-- ② 同展会同展位号（保留最早一条）——两类重复都会让唯一索引创建失败、
-- 迁移失败即中止服务（部署阻断），必须先去重。
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

CREATE UNIQUE INDEX IF NOT EXISTS uniq_exhibition_booths_exhibitor
  ON exhibition_booths(exhibition_id, exhibitor_id);

CREATE UNIQUE INDEX IF NOT EXISTS uniq_exhibition_booths_number
  ON exhibition_booths(exhibition_id, booth_number)
  WHERE booth_number <> '';
