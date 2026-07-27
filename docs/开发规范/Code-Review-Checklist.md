# Code Review Checklist

> 每次 Pull Request 必须通过以下检查项，Review 者逐项打勾。

## 分层约束（铁律）

- [ ] **Handler 层**：没有 SQL 查询、业务规则判断、直接读写库
- [ ] **Service 层**：没有 http.Request、ResponseWriter、JSON 编解码、环境变量
- [ ] **Repository 层**：没有业务规则、权限判断、HTTP 相关代码

## 代码风格

- [ ] 新文件 < 400 行，函数 < 50 行
- [ ] 嵌套层级 < 4 层，使用 early return
- [ ] 命名遵循 CLAUDE.md 约定（Handler: verbNoun, Service: NounService, PG Repo: nounRepo）
- [ ] 参照同层已有代码风格，保持一致性
- [ ] 创建新对象，不原地修改

## 错误处理

- [ ] 所有 error 返回值被检查，没有 `_` 忽略
- [ ] 使用 `fmt.Errorf("context: %w", err)` 包装错误
- [ ] 错误信息包含足够上下文（不暴露敏感信息）

## 响应规范

- [ ] 使用 `respond()` / `fail()` / `paginatedRespond()` 统一响应
- [ ] 没有手动拼 JSON 响应体
- [ ] 状态码正确（200/201/400/401/403/404/409/500）

## 安全审查

- [ ] 数据库查询使用参数化查询（pgx $1, $2）
- [ ] 用户输入经过消毒中间件
- [ ] 操作前进行权限校验（Actor 注入）
- [ ] 敏感字段已加密（AES-256-GCM）或脱敏
- [ ] Token 验证在中间件层完成

## 测试

- [ ] 新增 Service 层代码有对应单元测试
- [ ] 边界用例已覆盖（空值、非法状态、权限不足）
- [ ] 状态迁移有完整路径测试
- [ ] 本地执行 `go test ./internal/...` 全部 PASS

## 日志与审计

- [ ] 写操作（CREATE/UPDATE/DELETE）有审计日志
- [ ] 使用 `slog.Info` 记录关键操作（含操作人、操作对象、结果）
- [ ] 不使用 `fmt.Println` 或 `log.Println`

## 数据库

- [ ] 涉及表结构变更，有成对的 up/down 迁移文件
- [ ] 迁移命名：`0000XX_description.up.sql` / `.down.sql`
- [ ] 新表有合适的索引（主键、外键、常用查询列）

## 依赖管理

- [ ] `go.mod` 中新增依赖经过确认（不引入不必要的第三方库）
- [ ] 优先使用标准库，仅在必要时引入第三方库
