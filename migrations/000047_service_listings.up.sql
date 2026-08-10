-- 企业服务能力展示（PRD ②-2 供给能力展示）：企业发布的巡检/航拍/测绘等可承接能力
CREATE TABLE IF NOT EXISTS service_listings (
    id            TEXT PRIMARY KEY,
    provider_id   TEXT NOT NULL DEFAULT '',  -- 企业用户 ID（管理端录入可为空）
    provider_name TEXT NOT NULL DEFAULT '',  -- 企业名称（展示用）
    title         TEXT NOT NULL,             -- 服务标题
    category      TEXT NOT NULL DEFAULT '',  -- 服务分类：巡检/航拍/测绘/应急 等
    description   TEXT NOT NULL DEFAULT '',
    region        TEXT NOT NULL DEFAULT '',  -- 服务区域
    price_fen     BIGINT NOT NULL DEFAULT 0, -- 报价（分），0 为面议
    unit          TEXT NOT NULL DEFAULT '',  -- 单位：次/天/公里 等
    image         TEXT NOT NULL DEFAULT '',  -- 封面图
    status        TEXT NOT NULL DEFAULT 'published', -- published / offline
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_service_listings_status ON service_listings(status);
CREATE INDEX IF NOT EXISTS idx_service_listings_category ON service_listings(category);
