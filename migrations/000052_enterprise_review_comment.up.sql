-- 企业入驻审核意见持久化：驳回/需补充原因对用户端可见
ALTER TABLE enterprises ADD COLUMN IF NOT EXISTS review_comment TEXT NOT NULL DEFAULT '';
