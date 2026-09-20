-- 000120 down：只撤约束与生成列。
-- 故意不 DROP EXTENSION btree_gist：扩展本身无害，且可能已被其它对象依赖。
ALTER TABLE test_site_bookings DROP CONSTRAINT IF EXISTS excl_testsite_occupied_slot;
ALTER TABLE test_site_bookings DROP COLUMN IF EXISTS occupied_slot;
ALTER TABLE venue_bookings DROP CONSTRAINT IF EXISTS excl_venue_occupied_slot;
ALTER TABLE venue_bookings DROP COLUMN IF EXISTS occupied_slot;
