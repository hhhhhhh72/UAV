-- 报名记录冻结金额：completeEnrollment 按报名时冻结金额释放/退款，
-- 与课程实时价格解耦（防改价后资金错付/滞留）
ALTER TABLE training_enrollments ADD COLUMN IF NOT EXISTS paid_amount_fen BIGINT NOT NULL DEFAULT 0;
