-- C2 死代码清理：DemandBid 报价流已被意向对接（DemandIntent）取代，
-- 无任何路由/服务引用（仅残留仓储与表）。删表收尾。
DROP TABLE IF EXISTS demand_bids;
