const fs = require('fs');
const path = require('path');
const { USE_POSTGRES, ensureJsonStoreTable, query } = require('./pg');

const DATA_KEYS = {
    applications: path.join(__dirname, '..', 'data.json'),
    cases: path.join(__dirname, '..', 'cases.json'),
    users: path.join(__dirname, '..', 'users.json'),
    services_config: path.join(__dirname, '..', 'services_config.json')
};

function readJsonFile(filePath, fallback) {
    try {
        if (!fs.existsSync(filePath)) return fallback;
        const raw = fs.readFileSync(filePath, 'utf-8');
        return JSON.parse(raw);
    } catch (err) {
        console.error(`[migrate] Failed to read ${filePath}:`, err);
        return fallback;
    }
}

async function upsertJsonStore(key, data) {
    await query(
        `
        INSERT INTO json_store (key, data, updated_at)
        VALUES ($1, $2, NOW())
        ON CONFLICT (key)
        DO UPDATE SET data = EXCLUDED.data, updated_at = NOW();
        `,
        [key, JSON.stringify(data)]
    );
}

async function migrate() {
    if (!USE_POSTGRES) {
        console.error('[migrate] USE_POSTGRES is not enabled. Aborting.');
        process.exit(1);
    }

    await ensureJsonStoreTable();

    const users = readJsonFile(DATA_KEYS.users, []);
    const cases = readJsonFile(DATA_KEYS.cases, []);
    const applications = readJsonFile(DATA_KEYS.applications, []);
    const servicesConfig = readJsonFile(DATA_KEYS.services_config, {});

    await upsertJsonStore('users', users);
    await upsertJsonStore('cases', cases);
    await upsertJsonStore('applications', applications);
    await upsertJsonStore('services_config', servicesConfig);

    console.log('[migrate] PostgreSQL json_store updated successfully.');
    process.exit(0);
}

migrate().catch(err => {
    console.error('[migrate] Failed:', err);
    process.exit(1);
});

