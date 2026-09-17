-- 商品审核维度从 status 里拆出来。
--
-- 此前 status 一个字段同时承担两件事："审核中/被驳回" 与 "在售/已售/已下架"。
-- 后果是管理端的「驳回」只能写 status='removed'，与「卖家自己下架」同值：
-- 卖家在"我的发布"里分不清自己是被协会驳回还是主动下架，驳回也没有原因可留，
-- 审核动作本身也没有审计（trading_insurance_finance.go 全文零 s.audit 调用）。
--
-- 拆分后两个维度各自独立：
--   check_status: pending(待审核) / passed(已通过) / rejected(已驳回)
--   status:       pending(未上架) / listed(在售) / sold(已售) / removed(已下架)
-- 公开商城的口径收紧为 check_status='passed' AND status='listed'，缺一不可。
--
-- ⚠ 拆分的必要条件：公开列表的过滤条件必须同步从"排除 pending"改成上面这个合取式，
--   否则待审核商品会直接出现在低空商城里。见 service.ProductVisibleInHall。

ALTER TABLE drone_products ADD COLUMN IF NOT EXISTS check_status TEXT NOT NULL DEFAULT 'pending';
ALTER TABLE drone_products ADD COLUMN IF NOT EXISTS check_reason TEXT NOT NULL DEFAULT '';
ALTER TABLE drone_products ADD COLUMN IF NOT EXISTS reviewed_at TIMESTAMPTZ;
ALTER TABLE drone_products ADD COLUMN IF NOT EXISTS reviewed_by TEXT NOT NULL DEFAULT '';

-- 历史回填：不改变任何商品当前的可见性，只把旧语义翻译到新维度。
-- 1) 空串状态归一为 listed。空串此前被公开列表当作"在售"（trading_insurance_finance.go
--    的 p.Status == "" || p.Status == "listed"），归一后口径变成显式 listed，
--    消掉"异常数据静默上架"这一类问题。
UPDATE drone_products SET status = 'listed' WHERE COALESCE(status, '') = '';

-- 2) 其余非 pending 的行此前都是公开可见（或曾经可见）的，说明已通过审核，回填 passed。
--    仍在 pending 的行保持 check_status='pending'，需要管理端重新审核通过才会上架——
--    它们的 status 不是 listed，公开列表本来也看不到，行为不变。
UPDATE drone_products SET check_status = 'passed' WHERE COALESCE(status, '') <> 'pending';

CREATE INDEX IF NOT EXISTS idx_drone_products_check_status ON drone_products(check_status);
