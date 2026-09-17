-- 把"面议"从魔法值变成显式字段。
--
-- 此前 price_fen=0 同时表示两件事："卖家选了面议" 与 "卖家填了 0 元"，
-- 后端无法区分。小程序发布页用 (Number(v.price) || 0) * 100 生成价格，
-- 留空即 0，于是"面议"和"标价 0 元"落库后完全一样；设备区卡片还把它显示成 ¥0。
--
-- 新增 price_mode：
--   fixed      —— 明码标价，price_fen 必须 > 0
--   negotiable —— 面议，price_fen 必须 = 0
-- 两个字段的取值组合由应用层强制（service.TradingService.createProduct / UpdateMyProduct）。

ALTER TABLE drone_products ADD COLUMN IF NOT EXISTS price_mode TEXT NOT NULL DEFAULT 'fixed';

-- 回填：历史 0 价一律视为面议（这正是它此前的展示语义），非 0 价视为明码标价。
UPDATE drone_products SET price_mode = 'negotiable' WHERE price_fen = 0;
UPDATE drone_products SET price_mode = 'fixed' WHERE price_fen > 0;
