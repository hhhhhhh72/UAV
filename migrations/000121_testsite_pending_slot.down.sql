-- 000121 down：把场地预约的排他约束退回 000120 的口径（只约束 approved）。
ALTER TABLE test_site_bookings DROP CONSTRAINT IF EXISTS excl_testsite_occupied_slot;
ALTER TABLE test_site_bookings DROP COLUMN IF EXISTS occupied_slot;
ALTER TABLE test_site_bookings
  ADD COLUMN occupied_slot tstzrange
  GENERATED ALWAYS AS (
    CASE WHEN status = 'approved' THEN tstzrange(start_time, end_time, '[]') ELSE NULL END
  ) STORED;
ALTER TABLE test_site_bookings
  ADD CONSTRAINT excl_testsite_occupied_slot
  EXCLUDE USING gist (site_id WITH =, occupied_slot WITH &&);
