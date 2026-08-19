-- 飞手驳回理由：审核留痕（管理端驳回认证时记录原因）
ALTER TABLE certified_pilots ADD COLUMN reject_reason TEXT NOT NULL DEFAULT '';
