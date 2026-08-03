# 无人机产业综合服务平台

面向微信小程序与 Web 管理后台的全栈服务平台，覆盖无人机产业链 16 大业务模块。

## 技术栈

| 层 | 技术 |
|------|------|
| 后端 API | Go 1.25+，标准库 net/http，212 条端点，116 源文件 |
| 数据库 | PostgreSQL 15+（生产） / 内存存储（开发），66 张表 |
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
│   │   ├── auth_wechat.go         # 微信 code2Session + Token 刷新
│   │   ├── admin_handler.go       # 管理后台 Token（ADMIN_DEV_MODE）
│   │   ├── compat_routes.go       # 旧版 API 兼容 (/api/auth/*)
│   │   ├── h5_compat.go           # H5/Vue 前端兼容层 (JSON 文件存储)
│   │   ├── admin.html             # 嵌入式管理后台 SPA（35KB，含 ECharts）
│   │   └── *.go                   # 各业务 Handler（batch1-3 + biz + phase3）
│   ├── service/                   # 业务规则、权限校验、状态机流转
│   ├── repository/                # 数据持久化
│   │   ├── repositories.go        # Repository 接口定义（50+ interface）
│   │   ├── postgres/              # PostgreSQL 实现（pgxpool）
│   │   └── memory/                # 内存实现（开发用，sync.RWMutex）
│   ├── domain/models.go           # 业务实体与常量（66 个 struct）
│   ├── config/config.go           # 集中配置 + 验证 + 脱敏打印
│   ├── logger/logger.go           # 结构化日志（slog + 每日文件轮转）
│   ├── cache/cache.go             # 内存 TTL 缓存（60s 默认，5min 自动清理）
│   ├── middleware/middleware.go    # 输入消毒 + 统一错误格式
│   └── crypto/                    # AES-256-GCM 加密 + 脱敏函数
├── migrations/                    # 16 组迁移（31 个 SQL 文件，15 组 up/down + shops）
├── docs/                          # 项目文档（22 份，中文）
├── icons/                         # 15 个 SVG 图标
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

## 双存储模式

| 环境 | 存储 | 触发条件 |
|------|------|------|
| 开发 | 内存 | `DATABASE_URL` 未设置 |
| 生产 | PostgreSQL | `DATABASE_URL` 已设置 |

## 角色权限（4 级 RBAC）

| 角色 | 权限 |
|------|------|
| `platform_admin` | 全部管理权限 |
| `association_admin` | 企业审核/内容管理 |
| `enterprise` | 发布需求/招聘/合同 |
| `individual` | 接单/求职/交易 |

协会内部细分(7级): 会长/副会长/秘书长/部门负责人/普通会员/合作院校/访客

## 7 大业务系统

| 系统 | 核心子模块 |
|------|------|
| ①会员生态资源管控 | 会员注册/专家智库/产业资源台账/人才资源库/协会7级权限 |
| ②产业供需智能对接 | 需求大厅/供应展示/竞标报价/智能匹配/资源池 |
| ③产学研协同创新 | 科技成果库/研发难题广场/课题攻关/测试预约/成果转化追踪 |
| ④合规政策服务 | 政策资讯/合规知识库/团体标准库/项目申报/企业案例库 |
| ⑤人才教育与产教融合 | 培训认证/赛事管理/招聘求职/院校展示/校企共建 |
| ⑥活动与品牌服务 | 活动管理/会员品牌展示/展会排期/行业报告发布 |
| ⑦低空应急资源协同 | 应急资源/一键调度/救援案例库/部门对接/联合演练 |

## Token 认证

- **格式**: HMAC-SHA256 Bearer Token，非标准 JWT
- **结构**: `base64(json_header).base64(signature)`
- **过期**: Access Token 15分钟，Refresh Token 7天
- **刷新**: 轮转刷新令牌，每次刷新旧令牌作废

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

## 前端项目

| 项目 | 位置 | 技术栈 | 规模 |
|------|------|--------|------|
| 微信小程序 | `miniprogram/` | uni-app + Vue3 `<script setup>` + Vant Weapp | 68 页，5 Tab |
| Web 管理后台 | `frontend/` | Vue 3 + Element Plus + ECharts | Admin SPA |
| 嵌入式后台 | `internal/httpapi/admin.html` | 单文件 SPA | 35KB 内嵌 |

**小程序设计规范**:
- 品牌色 `#0A66C2`（深空蓝），辅色 `#1DD4A8`（青绿）
- 全局 CSS 变量定义在 `App.vue` 的 `page` 选择器中
- 输入框: `bg=#fafafa` `radius=24rpx`，按钮: `radius=50rpx` + `box-shadow`
- 不使用 emoji 图标，用 CSS 绘制或文字标签

**小程序 API 调用注意**:
- `request.js` 自动 unwrap Go 后端的 `{ data: {...} }` 响应包
- 登录用 `POST /api/auth/login`（返回 `{ accessToken, refreshToken, user }`）
- 注册用 `POST /api/auth/register`（返回 `{ success: true }`）
- Token 存储: `authStorage.setTokens(accessToken, refreshToken)` + `uni.setStorageSync('user', ...)`
- 主流程是微信静默登录（`App.vue` → `wx.login()` → `/api/auth/wx-login`），密码登录只是备用

## 关键踩坑记录

| 问题 | 根因 | 修复 |
|------|------|------|
| 注册 500: `duplicate key violates unique constraint "users_wechat_openid_key"` | 手机注册时 `wechat_openid` 为空字符串，第二次 `""` 违反 UNIQUE | 设置 `WechatOpenID: "phone:"+body.Phone` 确保唯一 |
| H5 兼容路由 404 | `/api/auth/*` 只在 `ADMIN_DEV_MODE=true` 时注册 | `run_api.bat` 已设该变量 |
| 登录 401 | 小程序调 `/api/v1/login` 不存在 | 改为 `/api/auth/login` |
| 登录后 Token 不匹配 | 旧代码只存 `token`，新 API 返回 `accessToken` | 改用 `authStorage.setTokens()` |

## 本地开发

```bash
# 后端 API（推荐用 run_api.bat，自动设环境变量）
go build -o api.exe ./cmd/api && run_api.bat

# 或手动
go run ./cmd/api     # 后端 → :8080

# 前端 Admin
cd frontend && npm run dev   # → :5173

# 小程序（用 HBuilderX 打开 miniprogram/ 目录编译）
```

## 详细文档

| 想看... | 文档 |
|------|------|
| 项目简介 + 快速开始 | [docs/项目概述/](docs/项目概述/) |
| 架构 + 分层 + 中间件 | [docs/系统架构/架构总览.md](docs/系统架构/架构总览.md) |
| 7大业务系统详情 | [docs/业务系统/](docs/业务系统/) |
| 全部 API 契约 | [docs/接口文档/API契约.md](docs/接口文档/API契约.md) |
| 66 张表结构 | [docs/数据设计/数据模型.md](docs/数据设计/数据模型.md) |
| 编码规范 | [docs/开发规范/编码规范.md](docs/开发规范/编码规范.md) |
| Docker + CI | [docs/运维部署/Docker部署.md](docs/运维部署/Docker部署.md) |
| 开发计划 | [docs/项目管理/开发计划.md](docs/项目管理/开发计划.md) |
