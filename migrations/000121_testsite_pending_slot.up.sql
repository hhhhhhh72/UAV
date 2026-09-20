-- 000121: 场地预约的排他约束收紧到 pending+approved（2026-09-20）
--
-- 000120 只约束 approved，原因是当时生产存在 7 组活跃行时段重叠（6 组 pending×pending、
-- 1 组 pending×approved），全部是 2026-08-12/08-18 的历史数据；把那批 pending 一并纳入
-- 会让迁移当场失败，而 cmd/api/main.go:281-283 对迁移失败是 os.Exit(1)，API 起不来。
--
-- 那批数据已于同日清理：5 行**已过期且与同场地其它预约重叠**的 pending 置为 rejected
-- （approved 那行保留、未改动），清理后活跃行重叠为 0。于是这里按 service 层
-- lockByKey 的判定口径把约束收紧：pending / approved 均占位。
--
-- 生成列表达式不能 ALTER，只能 drop column 再重建（重建会重算，行数很小）。
ALTER TABLE test_site_bookings DROP CONSTRAINT IF EXISTS excl_testsite_occupied_slot;
ALTER TABLE test_site_bookings DROP COLUMN IF EXISTS occupied_slot;
ALTER TABLE test_site_bookings
  ADD COLUMN occupied_slot tstzrange
  GENERATED ALWAYS AS (
    CASE WHEN status IN ('approved','pending') THEN tstzrange(start_time, end_time, '[]') ELSE NULL END
  ) STORED;
ALTER TABLE test_site_bookings
  ADD CONSTRAINT excl_testsite_occupied_slot
  EXCLUDE USING gist (site_id WITH =, occupied_slot WITH &&);
