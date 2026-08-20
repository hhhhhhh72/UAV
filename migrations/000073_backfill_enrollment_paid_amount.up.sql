-- 回填报名记录固化冻结金额（升级前付费报名学费滞留修复）
-- 背景：000070 只加列（DEFAULT 0）无回填，升级前按当时课程价冻结付费报名的
-- paid_amount_fen=0，completeEnrollment 跳过释放 → 学费滞留 escrow frozen。
-- 按 escrow 冻结流水回填：取同用户同课程最近一条 completed 的 freeze 金额
-- （冻结金额即报名时固化价，与课程后续改价无关）。
-- 幂等且保守：只填 paid_amount_fen=0 的记录，重复执行不覆盖已有值。
UPDATE training_enrollments e
SET paid_amount_fen = COALESCE((
    SELECT t.amount_fen FROM escrow_transactions t
    WHERE t.tx_type='freeze' AND t.reference_type='training_course'
      AND t.from_user = e.user_id AND t.reference_id = e.course_id
      AND t.status='completed'
    ORDER BY t.created_at DESC LIMIT 1
), 0)
WHERE e.paid_amount_fen = 0;
