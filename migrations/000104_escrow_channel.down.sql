DROP INDEX IF EXISTS idx_escrow_external;
DROP INDEX IF EXISTS idx_escrow_ref;
ALTER TABLE escrow_transactions DROP COLUMN IF EXISTS external_txn_id;
ALTER TABLE escrow_transactions DROP COLUMN IF EXISTS channel;
