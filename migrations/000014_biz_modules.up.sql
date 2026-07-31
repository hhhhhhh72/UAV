-- New business modules per 重庆市无人机产业协会 requirements
-- Experts, Cases, Compliance, Standards, Project Applications,
-- Achievements, R&D Challenges, Research Projects, Competitions,
-- Events, Portfolios, Industry Reports, Industry Resources,
-- Emergency Resources and Dispatches

CREATE TABLE IF NOT EXISTS experts (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    title TEXT NOT NULL DEFAULT '',
    org TEXT NOT NULL DEFAULT '',
    field TEXT NOT NULL DEFAULT '',
    tags JSONB DEFAULT '[]',
    bio TEXT DEFAULT '',
    avatar_url TEXT DEFAULT '',
    status TEXT NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS case_entries (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    category TEXT NOT NULL DEFAULT '',
    description TEXT DEFAULT '',
    images JSONB DEFAULT '[]',
    client_name TEXT DEFAULT '',
    result TEXT DEFAULT '',
    status TEXT NOT NULL DEFAULT 'draft',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS compliance_docs (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    category TEXT NOT NULL DEFAULT '',
    publisher TEXT NOT NULL DEFAULT '',
    publish_date TIMESTAMPTZ DEFAULT NOW(),
    summary TEXT DEFAULT '',
    file_url TEXT NOT NULL DEFAULT '',
    tags JSONB DEFAULT '[]',
    status TEXT NOT NULL DEFAULT 'draft',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS standard_docs (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    standard_no TEXT NOT NULL DEFAULT '',
    publisher TEXT DEFAULT '',
    effective_date TIMESTAMPTZ DEFAULT NOW(),
    status TEXT NOT NULL DEFAULT 'draft',
    scope TEXT NOT NULL DEFAULT '',
    summary TEXT NOT NULL DEFAULT '',
    file_url TEXT DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS project_applications (
    id TEXT PRIMARY KEY,
    applicant_id TEXT NOT NULL,
    project_name TEXT NOT NULL,
    category TEXT DEFAULT '',
    budget_fen BIGINT DEFAULT 0,
    description TEXT DEFAULT '',
    attachments JSONB DEFAULT '[]',
    status TEXT NOT NULL DEFAULT 'draft',
    review_note TEXT DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS achievements (
    id TEXT PRIMARY KEY,
    owner_id TEXT NOT NULL,
    title TEXT NOT NULL,
    achieve_type TEXT DEFAULT '',
    description TEXT DEFAULT '',
    field TEXT DEFAULT '',
    stage TEXT DEFAULT 'lab',
    images JSONB DEFAULT '[]',
    contact_info TEXT DEFAULT '',
    status TEXT NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS rd_challenges (
    id TEXT PRIMARY KEY,
    poster_id TEXT NOT NULL,
    title TEXT NOT NULL,
    field TEXT DEFAULT '',
    description TEXT DEFAULT '',
    budget_fen BIGINT DEFAULT 0,
    deadline TIMESTAMPTZ,
    status TEXT NOT NULL DEFAULT 'open',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS research_projects (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    field TEXT DEFAULT '',
    description TEXT DEFAULT '',
    lead_org TEXT DEFAULT '',
    members JSONB DEFAULT '[]',
    budget_fen BIGINT DEFAULT 0,
    start_date TIMESTAMPTZ,
    end_date TIMESTAMPTZ,
    milestones TEXT DEFAULT '',
    status TEXT NOT NULL DEFAULT 'planning',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS competitions (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    category TEXT DEFAULT '',
    description TEXT DEFAULT '',
    location TEXT DEFAULT '',
    start_date TIMESTAMPTZ,
    end_date TIMESTAMPTZ,
    max_teams INTEGER DEFAULT 0,
    reg_count INTEGER DEFAULT 0,
    sponsor TEXT DEFAULT '',
    status TEXT NOT NULL DEFAULT 'draft',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS competition_registrations (
    id TEXT PRIMARY KEY,
    competition_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    team_name TEXT DEFAULT '',
    member_count INTEGER DEFAULT 1,
    contact_info TEXT DEFAULT '',
    status TEXT NOT NULL DEFAULT 'submitted',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS association_events (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    event_type TEXT DEFAULT '',
    description TEXT DEFAULT '',
    location TEXT DEFAULT '',
    start_time TIMESTAMPTZ,
    end_time TIMESTAMPTZ,
    max_attendees INTEGER DEFAULT 0,
    reg_count INTEGER DEFAULT 0,
    cover_url TEXT DEFAULT '',
    status TEXT NOT NULL DEFAULT 'draft',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS event_registrations (
    id TEXT PRIMARY KEY,
    event_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    name TEXT DEFAULT '',
    phone TEXT DEFAULT '',
    org TEXT DEFAULT '',
    status TEXT NOT NULL DEFAULT 'registered',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS member_portfolios (
    id TEXT PRIMARY KEY,
    enterprise_id TEXT NOT NULL,
    name TEXT NOT NULL,
    logo_url TEXT DEFAULT '',
    cover_url TEXT DEFAULT '',
    description TEXT DEFAULT '',
    products JSONB DEFAULT '[]',
    honors JSONB DEFAULT '[]',
    contact_info TEXT DEFAULT '',
    status TEXT NOT NULL DEFAULT 'draft',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS industry_reports (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    period TEXT DEFAULT '',
    category TEXT DEFAULT '',
    summary TEXT DEFAULT '',
    content TEXT DEFAULT '',
    file_url TEXT DEFAULT '',
    author TEXT DEFAULT '',
    status TEXT NOT NULL DEFAULT 'draft',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS industry_resources (
    id TEXT PRIMARY KEY,
    owner_id TEXT NOT NULL,
    name TEXT NOT NULL,
    res_type TEXT DEFAULT '',
    model TEXT DEFAULT '',
    specs TEXT DEFAULT '',
    location TEXT DEFAULT '',
    price_fen BIGINT DEFAULT 0,
    booking_info TEXT DEFAULT '',
    status TEXT NOT NULL DEFAULT 'available',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS emergency_resources (
    id TEXT PRIMARY KEY,
    owner_id TEXT NOT NULL,
    name TEXT NOT NULL,
    res_type TEXT DEFAULT '',
    specs TEXT DEFAULT '',
    quantity INTEGER DEFAULT 1,
    location TEXT DEFAULT '',
    contact_info TEXT DEFAULT '',
    status TEXT NOT NULL DEFAULT 'standby',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS emergency_dispatches (
    id TEXT PRIMARY KEY,
    resource_id TEXT NOT NULL,
    event_desc TEXT DEFAULT '',
    location TEXT DEFAULT '',
    start_time TIMESTAMPTZ,
    end_time TIMESTAMPTZ,
    commander TEXT DEFAULT '',
    result TEXT DEFAULT '',
    status TEXT NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for common queries
CREATE INDEX IF NOT EXISTS idx_experts_field ON experts(field);
CREATE INDEX IF NOT EXISTS idx_case_entries_category ON case_entries(category);
CREATE INDEX IF NOT EXISTS idx_compliance_docs_category ON compliance_docs(category);
CREATE INDEX IF NOT EXISTS idx_achievements_field ON achievements(field);
CREATE INDEX IF NOT EXISTS idx_rd_challenges_status ON rd_challenges(status);
CREATE INDEX IF NOT EXISTS idx_project_applications_applicant ON project_applications(applicant_id);
CREATE INDEX IF NOT EXISTS idx_competition_regs_comp ON competition_registrations(competition_id);
CREATE INDEX IF NOT EXISTS idx_event_regs_event ON event_registrations(event_id);
CREATE INDEX IF NOT EXISTS idx_portfolios_enterprise ON member_portfolios(enterprise_id);
CREATE INDEX IF NOT EXISTS idx_industry_resources_type ON industry_resources(res_type);
