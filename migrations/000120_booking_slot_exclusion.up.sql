-- 000120: 场地/场馆预约的时段重叠在库层面兜底（2026-09-20）
--
-- 为什么需要：冲突判定此前只在 service 层 —— lockByKey 是**进程内**键锁
-- （全仓 8 处同类键锁之一），多一个 API 进程就各锁各的，同一场地同一时段会被
-- 双重占用；而且那本来也只是 check-then-insert，库层面没有任何兜底
-- （唯一索引表达不了区间重叠）。
--
-- 为什么用「生成列 + EXCLUDE」而不是直接 EXCLUDE：
--   PostgreSQL 的排他约束**不支持 WHERE 部分约束**。若直接对
--   (site_id, tstzrange(start_time,end_time)) 建约束，被取消/驳回的历史行也会参与
--   互斥 —— 某个时段一旦被取消过就再也订不回来。生成列在"不占用的状态"下取 NULL，
--   而 NULL 在排他约束里不与其他行冲突（与唯一索引对 NULL 的处理一致），
--   正好表达「只有真正占用的行才互斥」。
--
-- 上界用 '[]'（闭区间）：与 service 层判定 `!(end < start_b || start > end_b)` 一致
-- ——前一场 12:00 结束、后一场 12:00 开始视为冲突。tstzrange 三参重载在 pg_proc 里
-- 是 immutable（已在生产实测），可以进生成列。
--
-- **场地预约只约束 approved**：生产实测有 7 组活跃行重叠（6 组 pending×pending、
-- 1 组 pending×approved），全部是 2026-08-12/08-18 的历史数据；而
-- 「pending/approved 均占位」这条规则是 8991dd2（2026-08-25）才加的，晚于那些数据。
-- 若把 pending 一并纳入，本迁移会当场失败 —— 而 main.go:283 对迁移失败是
-- os.Exit(1)，API 根本起不来。approved 之间的重叠实测为 0，所以先钉死最强的那条
-- 不变量：「任何两个已通过的预约不可能同时占用一个场地」。
-- pending 期间的互斥仍由进程内键锁 + 审批前复检（ReviewBooking）承担；
-- 清理掉那 7 组历史数据后，可把条件收紧为 status IN ('approved','pending')。
--
-- 场馆预约生产 0 行，直接按代码里的规则（booked/pending 均占用）约束。

CREATE EXTENSION IF NOT EXISTS btree_gist;

-- 场地预约：approved 才占用
ALTER TABLE test_site_bookings
  ADD COLUMN IF NOT EXISTS occupied_slot tstzrange
  GENERATED ALWAYS AS (
    CASE WHEN status = 'approved' THEN tstzrange(start_time, end_time, '[]') ELSE NULL END
  ) STORED;
ALTER TABLE test_site_bookings DROP CONSTRAINT IF EXISTS excl_testsite_occupied_slot;
ALTER TABLE test_site_bookings
  ADD CONSTRAINT excl_testsite_occupied_slot
  EXCLUDE USING gist (site_id WITH =, occupied_slot WITH &&);

-- 场馆预约：booked / pending 均占用（与 reviews_resources.go 的判定一致）
ALTER TABLE venue_bookings
  ADD COLUMN IF NOT EXISTS occupied_slot tstzrange
  GENERATED ALWAYS AS (
    CASE WHEN status IN ('booked','pending') THEN tstzrange(start_time, end_time, '[]') ELSE NULL END
  ) STORED;
ALTER TABLE venue_bookings DROP CONSTRAINT IF EXISTS excl_venue_occupied_slot;
ALTER TABLE venue_bookings
  ADD CONSTRAINT excl_venue_occupied_slot
  EXCLUDE USING gist (venue_id WITH =, occupied_slot WITH &&);
