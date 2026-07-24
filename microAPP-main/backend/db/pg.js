const { Pool } = require('pg');

const USE_POSTGRES = process.env.USE_POSTGRES === '1';
const DATABASE_URL = process.env.PG_URL || process.env.DATABASE_URL || '';
const PG_SSL = String(process.env.PG_SSL || '').toLowerCase() === 'true';

let pool = null;

function createPool() {
    if (!USE_POSTGRES) return null;
    if (pool) return pool;

    if (DATABASE_URL) {
        pool = new Pool({
            connectionString: DATABASE_URL,
            ssl: PG_SSL ? { rejectUnauthorized: false } : false
        });
        return pool;
    }

    pool = new Pool({
        host: process.env.PG_HOST || '127.0.0.1',
        port: Number(process.env.PG_PORT || 5432),
        user: process.env.PG_USER || 'postgres',
        password: process.env.PG_PASSWORD || '',
        database: process.env.PG_DATABASE || 'lowaltitude'
    });
    return pool;
}

async function query(text, params = []) {
    if (!USE_POSTGRES) {
        throw new Error('PostgreSQL is disabled. Set USE_POSTGRES=1 to enable.');
    }
    const client = createPool();
    return client.query(text, params);
}

async function ensureJsonStoreTable() {
    if (!USE_POSTGRES) return;
    await query(`
        CREATE TABLE IF NOT EXISTS json_store (
            key TEXT PRIMARY KEY,
            data JSONB NOT NULL,
            updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
        );
    `);
}

module.exports = {
    USE_POSTGRES,
    query,
    ensureJsonStoreTable
};

