<!-- Generated: 2026-08-19 | Files scanned: 154 | Token estimate: ~900 -->

# 后端架构地图（Go 1.25, net/http）

## 路由注册分组（httpapi.Server.Router，443 条：生产 403 + dev-only 40）

```
registerCoreRoutes       需求/企业/就业/合同/职位/社区/二手/用工/培训/交易/保险/金融/消息/文件/admin/auth
registerPhase3Routes     报名/到期提醒/交易订单/托管/资讯/评价/场地/管理端用户
registerBatch1Routes     产业资源池/测试场地预约/展会与展位
registerBatch2Routes     成果转化/院校展示/校企共建
registerBatch3Routes     救援案例/应急部门演练/协会成员(8级)
registerBizRoutes        专家智库/合规/报告/品牌/成果/难题/攻关/申报/赛事/活动/资源/应急调度/智能匹配
registerAdminListRoutes  18 模块管理端 CRUD (/api/v1/admin/*)
registerPublicAPIRoutes  小程序公开聚合别名
registerH5AuthRoutes     生产无条件注册 (7 条 auth 路由)
[dev-only] registerCompatRoutes(3) + registerH5Compat(36) + /swagger/
```

## 中间件链（外→内）

```
rateLimit(100/s, RemoteAddr) → requestID → recoverPanic → securityHeaders
  → withCORS → authenticate(白名单+可选解析) → idempotencyCheck(24h, actor命名空间)
  → adminGate(/api/v1/admin/* 仅 admin 角色) → SanitizeBody(JSON去HTML标签) → Handler
```

## Handler → Service → Repository 映射（代表性模块）

| 业务 | Handler 文件 | Service | Repo 接口 |
|------|-------------|---------|-----------|
| 需求 | server.go / demand_fulfillment.go | DemandService | DemandRepository |
| 企业 | enterprise.go / export_handler.go | EnterpriseService | EnterpriseRepository |
| 合同+电子签 webhook | contract.go (HMAC 签名+event_id 去重) | ContractService | ContractRepository |
| 交易托管 | escrow.go (资金方法原子+哨兵错误) | EscrowService | EscrowRepository |
| 微信登录 | auth_wechat.go | AuthService | UserRepository+RefreshToken |
| 就业/职位 | jobs.go (投递快照) | JobService | JobRepository |
| 培训认证 | training.go (身份证脱敏) | TrainingService | CertificateRepository |
| 管理端 CRUD | admin_crud_stubs.go (泛型 adminListFilter/adminSlicePage) | — (直连 repo) | 各 Repository |
| H5 兼容层 | h5_compat.go (⚠PG+JSON 双写) | — | users.json + PG users |

## 关键机制

- **响应规范**：respond/paginatedRespond/fail，统一 `{data, request_id}`；分页 `{data,total,page,page_size}`
- **脱敏**：crypto.MaskPhone/MaskIDCard/MaskCreditCode 出口统一脱敏；密码 hash `json:"-"` 永不序列化
- **幂等**：actor 命名空间隔离（`a.ID+":"+key`），24h TTL
- **CAS 并发**：CompareAndSetStatus — PG 用 `UPDATE...WHERE id AND status` 原子条件更新
- **文件上传**：10MB 限制 + 类型白名单 + SHA256 → `/uploads/{file_id}`
- **输入消毒**：SanitizeBody 挂载于最内层——JSON 字符串值去 HTML 标签、password 字段保真、json.Number 保大整数精度、1MiB 上限
- **Token**：登录/刷新统一签发标准 JWT (IssueJWT)，Verify 兼容旧式两段；refresh 轮转先存新再撤旧（防锁号）
- **SMS**：验证码 5min TTL + 60s 防重发 + 5 次错误作废 + 常量时间比较

## 已知问题（改动时注意）

1. h5_compat.go 双写 PG+JSON；image proxy 仅白名单域名跳转（dev-only 路由）
2. memory/memory.go(2497行)、postgres/postgres.go(2305行)、biz_handlers.go(1610行)、h5_compat.go(1262行) 超限需拆分
3. batch2_handlers.go 批量路由无角色校验；adminCreateProduct 两次写库非原子
4. 测试装配 newServer(t) 与 main.go 手工重复（54 测试文件，覆盖率约 46%）
5. SMS 验证码存进程内存（无法水平扩展）、未接真实短信商（dev_code 回显）
6. race 检测仅 CI（Ubuntu）可跑，Windows 本机需 gcc

## 验收标准（每次改动后）

```
go build ./... && go vet ./... && go test ./internal/...
```
