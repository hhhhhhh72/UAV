-- 成果附件资料（detail_pd.html 原型附件区块：名称/大小/URL）
ALTER TABLE achievements ADD COLUMN IF NOT EXISTS attachments JSONB DEFAULT '[]';
