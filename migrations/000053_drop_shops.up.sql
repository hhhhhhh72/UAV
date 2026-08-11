-- 删除历史遗留 shops 模块（PRD 无商家/店铺概念，商家=已审核企业展示）
-- 小程序商家页数据来自 enterprises（home API），shops 表/后台 CRUD 无人消费
DROP TABLE IF EXISTS shops;
