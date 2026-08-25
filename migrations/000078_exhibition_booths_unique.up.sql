-- 展位防超卖/重复：同一参展商对同一展会仅能申请一次；展位号在展会内唯一。
-- 先清理存量重复（保留最早一条，即 id 最小者）。
DELETE FROM exhibition_booths a
USING exhibition_booths b
WHERE a.id > b.id
  AND a.exhibition_id = b.exhibition_id
  AND a.exhibitor_id = b.exhibitor_id;

CREATE UNIQUE INDEX IF NOT EXISTS uniq_exhibition_booths_exhibitor
  ON exhibition_booths(exhibition_id, exhibitor_id);

CREATE UNIQUE INDEX IF NOT EXISTS uniq_exhibition_booths_number
  ON exhibition_booths(exhibition_id, booth_number)
  WHERE booth_number <> '';
