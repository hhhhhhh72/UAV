-- 回滚：删除服务能力种子（仅删除本迁移写入的 sl-1~6）
DELETE FROM service_listings WHERE id IN ('sl-1','sl-2','sl-3','sl-4','sl-5','sl-6');
