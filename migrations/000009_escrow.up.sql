-- 资金托管（第三方存管）

CREATE TABLE IF NOT EXISTS escrow_accounts (
    user_id     TEXT PRIMARY KEY,
    balance_fen BIGINT NOT NULL DEFAULT 0,
    frozen_fen  BIGINT NOT NULL DEFAULT 0,
    version     INT NOT NULL DEFAULT 1,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS escrow_transactions (
    id              TEXT PRIMARY KEY,
    from_user       TEXT NOT NULL,
    to_user         TEXT NOT NULL,
    amount_fen      BIGINT NOT NULL,
    tx_type         TEXT NOT NULL, -- deposit/freeze/release/refund
    reference_type  TEXT,          -- course/trade
    reference_id    TEXT,
    status          TEXT NOT NULL DEFAULT 'pending',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_escrow_from ON escrow_transactions(from_user);
CREATE INDEX IF NOT EXISTS idx_escrow_to ON escrow_transactions(to_user);
CREATE INDEX IF NOT EXISTS idx_escrow_ref ON escrow_transactions(reference_type, reference_id);
