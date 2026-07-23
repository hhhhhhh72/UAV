# 无人机产业综合服务平台

面向微信小程序与 Web 管理后台的全栈服务平台，覆盖无人机产业链 16 大业务模块。

## 技术栈

| 层 | 技术 |
|------|------|
| 后端 API | Go 1.22+，标准库 net/http，130 条端点 |
| 数据库 | PostgreSQL 15+（生产） / 内存存储（开发），49 张表 |
| 部署 | Docker 多阶段构建 + docker-compose（PG + API 双容器） |
| CI/CD | GitHub Actions（build + vet + test + integration） |

## 项目结构

```
.
├── cmd/api/main.go               # 启动入口，组装依赖
├── internal/
│   ├── httpapi/                   # 路由、中间件、Handler（解析请求→调Service→respond）
│   │   ├── server.go              # Server struct + 路由注册 + 中间件链
│   │   ├── auth.go                # Token 签发/验证 + 鉴权中间件
│   │   ├── admin.html             # 遗留嵌入式管理后台 SPA
│   │   └── *.go                   # 各业务 Handler
│   ├── service/                   # 业务规则、权限校验、状态机流转
│   ├── repository/                # 数据持久化
│   │   ├── repositories.go        # Repository 接口定义（15 个 interface）
│   │   ├── postgres/              # PostgreSQL 实现
│   │   └── memory/                # 内存实现（开发用）
│   ├── domain/models.go           # 业务实体与常量（40 个 struct）
│   ├── config/config.go           # 集中配置 + 验证 + 脱敏打印
│   ├── logger/logger.go           # 结构化日志（slog + 每日文件轮转）
│   ├── cache/cache.go             # 内存 TTL 缓存
│   ├── middleware/middleware.go    # 输入消毒 + 统一错误格式
│   └── crypto/                    # AES-256-GCM 加密 + 脱敏函数
├── migrations/                    # 24 个迁移文件（12 组 up/down）
├── docs/                           # 项目文档（10 份）
├── uploads/                        # 文件上传目录
├── scripts/                        # 迁移脚本（WXSS 迁移工具）
└── docker-compose.yml
```

## 分层约束（铁律）

这是最重要的编码规范，每次修改必须遵守：

```
Handler     → 只做：解析请求、调 Service、调 respond/fail
              禁做：SQL 查询、业务规则判断、直接读写库

Service     → 只做：角色校验、归属校验、状态机迁移、调 Repository 接口
              禁做：http.Request、http.ResponseWriter、JSON 编解码、环境变量

Repository  → 只做：SQL 执行、数据映射、返回 Domain 对象
              禁做：业务规则、权限判断、HTTP 相关
```

**违反案例（禁止）：**
- Handler 里写 `db.Query(...)` — 越界
- Service 里写 `json.NewEncoder(w).Encode(...)` — 越界
- Repository 里写 `if a.Role != "admin"` — 越界

## 命名约定

| 层级 | 约定 | 示例 |
|------|------|------|
| 路由 | `/api/v1/资源复数` | `/api/v1/demands` |
| Handler | `func (s *Server) verbNoun(...)` | `createDemand` |
| Service | `type NounService struct` | `DemandService` |
| Repository 接口 | `type NounRepository interface` | `DemandRepository` |
| PG 实现 | `type nounRepo struct` (小写开头) | `demandRepo` |
| 迁移文件 | `00000X_description.up.sql` | `000001_init.up.sql` |

## 中间件链

```
请求 → 幂等去重(24h) → 限流(100/s) → 链路追踪 → Panic恢复 → 安全头 → CORS → Token鉴权 → 业务处理
```

## 角色权限（4 级 RBAC）

| 角色 | 权限范围 |
|------|------|
| `platform_admin` | 全部管理权限（需求/用工/合同审核、用户管理、数据导出） |
| `association_admin` | 企业入驻审核、会员管理、行业资讯、内容审核 |
| `enterprise` | 发布/承接需求、招聘、用工申请、合同创建 |
| `individual` | 接单、求职、二手交易、社区发布 |

## 响应规范

```go
// 成功 — 必须用这些函数，禁止手动拼 JSON
respond(w, r, statusCode, data)
paginatedRespond(w, r, items, totalCount)

// 失败
fail(w, r, statusCode, err)
```

## 验收标准

每个功能模块完成后必须：

```bash
go build ./...          # 编译通过
go vet ./...            # 零告警
go test ./internal/...  # 全部 PASS
```

## 环境变量

| 变量 | 必填 | 说明 |
|------|:--:|------|
| `AUTH_SECRET` | ✅ | JWT 签名密钥，至少 32 字节 |
| `DATABASE_URL` | — | PostgreSQL 连接串（不设则用内存存储） |
| `WECHAT_APPID` | — | 小程序 AppID |
| `WECHAT_APPSECRET` | — | 小程序 AppSecret |
| `ADMIN_DEV_MODE` | — | 设为 `true` 启用管理后台 |
| `ENCRYPTION_KEY` | — | AES-256-GCM 加密密钥 |
| `SIGNING_SECRET` | — | 签章 Webhook 签名密钥 |
| `LOG_DIR` | — | 日志目录，默认 `./logs` |
| `LOG_LEVEL` | — | 日志级别：debug/info/warn/error |
| `CORS_ORIGINS` | — | CORS 允许来源，逗号分隔 |
| `HTTP_ADDR` | — | 监听地址，默认 `:8080` |

## 本地开发

```bash
# 编译启动工具（一次性）
go build -o dev.exe ./cmd/cli

# 使用
dev          # 菜单
dev api      # 后端 → :8080
dev all      # 全部
dev stop     # 停止

# 或传统方式
go run ./cmd/api                    # 后端
```

## 文档同步

| 代码变更 | 必须同步的文档 |
|------|------|
| 新增/修改 API 端点 | [docs/api-contract.md](docs/api-contract.md) |
| 新增/修改数据表 | [docs/data-model.md](docs/data-model.md) + 写迁移文件 |
| 新增/修改角色权限 | [docs/architecture.md](docs/architecture.md) |
| 新增环境变量/启动参数 | [README.md](README.md) |
| 修复审计发现问题 | [docs/roadmap.md](docs/roadmap.md) 追加修复记录 |

## 已知待办

1. **架构分层修复** — 9 个服务层模块内嵌内存存储，需统一走 Repository
2. **测试覆盖率提升** — 服务层从 ~15% 提升到 60%+
3. **文档同步** — `api-contract.md` 缺 17 条端点，`data-model.md` 缺 11 张表
