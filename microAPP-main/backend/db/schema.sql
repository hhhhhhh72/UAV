-- Minimal JSON store schema for PostgreSQL version (phase 1)
CREATE TABLE IF NOT EXISTS json_store (
    key TEXT PRIMARY KEY,
    data JSONB NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Optional: add indexes when normalizing tables in future phases

