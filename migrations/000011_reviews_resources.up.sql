-- 信用评价 + 资源共享 + 政策服务

-- 评价表
CREATE TABLE IF NOT EXISTS reviews (
    id              TEXT PRIMARY KEY,
    reviewer_id     TEXT NOT NULL,
    target_type     TEXT NOT NULL, -- enterprise/instructor/course/product
    target_id       TEXT NOT NULL,
    rating          INT NOT NULL DEFAULT 5, -- 1-5
    content         TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_reviews_target ON reviews(target_type, target_id);

-- 场地预约
CREATE TABLE IF NOT EXISTS venues (
    id          TEXT PRIMARY KEY,
    owner_id    TEXT NOT NULL,
    name        TEXT NOT NULL,
    venue_type  TEXT NOT NULL, -- training_field/takeoff_point/maintenance
    location    TEXT,
    price_fen   BIGINT DEFAULT 0,
    status      TEXT NOT NULL DEFAULT 'available',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS venue_bookings (
    id          TEXT PRIMARY KEY,
    venue_id    TEXT NOT NULL,
    user_id     TEXT NOT NULL,
    start_time  TIMESTAMPTZ NOT NULL,
    end_time    TIMESTAMPTZ NOT NULL,
    status      TEXT NOT NULL DEFAULT 'booked',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
