# 无人机产业综合服务平台

面向微信小程序与 Web 管理后台的全栈服务平台，覆盖无人机产业链 **7 大业务系统**。

> 📋 **团队协作**: 4人并行开发 | [PRD](项目管理/PRD-四人并行开发方案.md) | 小程序 78 页 + 后台 39 路由  
> 🟢 **代码质量**: P0/P1清零(0 JSON忽略 + 0裸error) | 覆盖率 45.8% | go vet ✅

## 技术栈

| 层 | 技术 |
|------|------|
| 后端 API | Go 1.25+，标准库 net/http，约 380 条路由注册（生产约 335） |
| 数据库 | PostgreSQL 16（生产） / 内存存储（开发），80 张表（37 组迁移） |
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
│   │   └── *.go                   # 各业务 Handler
│   ├── service/                   # 业务规则、权限校验、状态机流转
│   ├── repository/                # 数据持久化
│   │   ├── repositories.go        # Repository 接口定义（50+ interface）
│   │   ├── postgres/              # PostgreSQL 实现
│   │   └── memory/                # 内存实现（开发用）
│   ├── domain/models.go           # 业务实体与常量（75 个 struct，含 models_batch*/models_new）
│   ├── config/config.go           # 集中配置 + 验证 + 脱敏打印
│   ├── logger/logger.go           # 结构化日志（slog + 每日文件轮转）
│   ├── cache/cache.go             # 内存 TTL 缓存
│   ├── middleware/middleware.go    # 输入消毒 + 统一错误格式
│   └── crypto/                    # AES-256-GCM 加密 + 脱敏函数
├── migrations/                    # 数据库迁移脚本 (80 表, 37 组)
├── docs/                           # 项目文档
├── prototypes/                     # HTML 原型 (首页+商家页)
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
| `ADMIN_DEV_MODE` | — | 设为 `true` 启用开发令牌 |
| `ENCRYPTION_KEY` | — | AES-256-GCM 加密密钥 |
| `CORS_ORIGINS` | — | CORS 允许来源，逗号分隔 |
| `HTTP_ADDR` | — | 监听地址，默认 `:8080` |

## 🚨 GitHub Push 规则

**任何人(含 AI) push 代码到 GitHub 前，必须经 A 书面确认。**
AI 生成的代码必须先人工审核，确认无误后再由 A 决定是否 push。
本地 commit 随意，但 `git push` 必须有 A 的明确许可。

## 本地开发

```bash
go run ./cmd/api     # 后端 → :8080
bash Test/e2e.sh     # E2E 测试
```

## 代码质量红线 (新增)

```go
// ❌ 禁止忽略 JSON 序列化错误
img, _ := json.Marshal(d.Images)  

// ✅ 必须检查
img, err := json.Marshal(d.Images)
if err != nil { return domain.Demand{}, fmt.Errorf("marshal images: %w", err) }

// ❌ 禁止裸返回 error
return err

// ✅ 必须包装上下文
return fmt.Errorf("delete expert %s: %w", id, err)
```

## 详细文档

| 想看... | 文档 |
|------|------|
| 四人协作分工 + Git策略 + AI规范 | [docs/项目管理/PRD-四人并行开发方案.md](项目管理/PRD-四人并行开发方案.md) |
| 架构 + 分层 + 中间件链 | [docs/系统架构/架构总览.md](系统架构/架构总览.md) |
| 7大业务系统详情 | [docs/业务系统/](业务系统/) |
| 全部 API 概览 | [docs/接口文档/API契约.md](接口文档/API契约.md)（约 380 条注册，swagger 仅 dev 可用） |
| 80 张表结构 | [docs/数据设计/数据模型.md](数据设计/数据模型.md) |
| Code Review 检查清单 | [docs/开发规范/Code-Review-Checklist.md](开发规范/Code-Review-Checklist.md) |
| 编码规范 | [docs/开发规范/编码规范.md](开发规范/编码规范.md) |
| Docker 部署 | [docs/运维部署/Docker部署.md](运维部署/Docker部署.md) |
