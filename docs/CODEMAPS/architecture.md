<!-- Generated: 2026-08-19 | Files scanned: 300+ | Token estimate: ~600 -->

# 无人机产业综合服务平台 — 架构地图

## 系统形态

单体仓库（monorepo）：1 个 Go 后端 + 2 个前端（Web 管理后台 / 微信小程序）+ 1 个 PostgreSQL。

```
微信小程序 (uni-app, 103页)          Web 管理后台 (Vue3+Arco, 40路由)
        │  HTTPS JSON                       │
        ▼                                  ▼
┌───────────────── Go API :8080 ─────────────────┐
│ cmd/api/main.go → httpapi.NewServer          │
│ 中间件链: rateLimit→requestID→recoverPanic→  │
│   securityHeaders→CORS→authenticate→idempotency│
│   →adminGate→SanitizeBody                    │
│ Handler(60+) → Service(约50) → Repository(59接口)│
└──────────────┬────────────────┬───────────────┘
               │                │
    PostgreSQL 16 (生产)   内存存储 (开发, 无 DATABASE_URL)
    + JSON 文件 (H5 兼容层, dev-only)
```

## 分层铁律（所有修改必须遵守）

| 层 | 只做 | 禁做 |
|----|------|------|
| Handler | 解析请求、调 Service、respond/fail | SQL、业务规则 |
| Service | 角色/归属校验、状态机、调 Repository | http.Request、JSON |
| Repository | SQL、数据映射、返回 Domain | 业务规则、权限 |

## 数据流示例（需求发布→审核）

```
POST /api/v1/demands → createDemand(handler) → DemandService.Create
  → role check (enterprise/individual) → DemandRepo.Insert
  → 状态机: pending → (Review) published | rejected
  → completed/cancelled (发布方) ；审核通知异步 msgSvc.Send
```

## 双存储切换

唯一开关：`DATABASE_URL` 环境变量（main.go）。
- 有值 → postgres.NewStore + RunMigrationsFromDir（启动自动迁移）
- 无值 → memory.NewXxxRepository（sync.RWMutex + []T，重启即丢）
- 审计日志仅 PG 模式写入（SetAuditWriter）

## 认证模型

- 标准 JWT (HS256, IssueJWT) 为主、兼容旧式两段格式，Access 15min / Refresh 7d 轮转（先存新再撤旧防锁号），refresh 只存 SHA-256 哈希
- RBAC 4 级主角色（platform_admin > association_admin > enterprise > individual，写入 token）
- 协会内部 8 级（association_members 表：president/vice_president/secretary/dept_head/member/partner/college/guest），与主角色分层叠加（admin=3 > partner=2 > member=1 > public=0 的可见性模型）

## 关键文件索引

| 路径 | 职责 |
|------|------|
| cmd/api/main.go | 依赖组装、双存储分支、种子管理员 |
| internal/httpapi/server.go | Server struct、中间件链、respond/fail |
| internal/httpapi/auth.go | Token 签发/验证、authenticate/adminGate |
| internal/service/ | 约 50 个 *Service，业务规则与状态机 |
| internal/repository/repositories.go | 59 个 Repository 接口 |
| internal/repository/postgres/ | pgxpool 实现（postgres.go 2305 行，最大文件） |
| internal/repository/memory/ | 内存实现（memory.go 2497 行） |
| internal/domain/ | 90 个业务实体 struct 与角色常量 |
