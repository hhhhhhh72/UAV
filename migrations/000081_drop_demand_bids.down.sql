-- 回滚：按 000012_demand_bids 原结构重建（无数据，仅供回滚可用）。
CREATE TABLE IF NOT EXISTS demand_bids (
    id TEXT PRIMARY KEY,
    demand_id TEXT NOT NULL,
    bidder_id TEXT NOT NULL,
    bidder_name TEXT NOT NULL DEFAULT '',
    amount_fen BIGINT NOT NULL DEFAULT 0,
    proposal TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'pending',
    version INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_demand_bids_demand_id ON demand_bids(demand_id);
CREATE INDEX IF NOT EXISTS idx_demand_bids_bidder_id ON demand_bids(bidder_id);
