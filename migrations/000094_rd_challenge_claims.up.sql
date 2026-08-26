-- 研发难题揭榜意向（challenges/{id}/claims：GET 动态列表 + POST 提交）
CREATE TABLE IF NOT EXISTS rd_challenge_claims (
  id TEXT PRIMARY KEY,
  challenge_id TEXT NOT NULL,
  user_id TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'submitted',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS rd_challenge_claims_challenge_user
  ON rd_challenge_claims (challenge_id, user_id);
