-- 种子内置合同模板：P2-2 将 listContractTemplates 从硬编码改为入库读取。
-- ON CONFLICT DO NOTHING 保证重复执行幂等，不会覆盖管理员后续自定义内容。
INSERT INTO contract_templates (id, name, version, content, status, created_at, updated_at)
VALUES
  ('tpl-001', '标准无人机服务合同', 1, '', 'active', now(), now()),
  ('tpl-002', '无人机买卖协议', 1, '', 'active', now(), now())
ON CONFLICT (id) DO NOTHING;
