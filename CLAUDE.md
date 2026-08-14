# 无人机产业综合服务平台

面向微信小程序与 Web 管理后台的全栈服务平台，覆盖无人机产业链 7 大业务系统。

## 技术栈

| 层 | 技术 |
|------|------|
| 后端 API | Go 1.25+，标准库 net/http，443 条路由注册（生产 403，dev-only 40），154 个 Go 文件（100 源码 + 54 测试） |
| 数据库 | PostgreSQL 16（生产） / 内存存储（开发），85 张表（63 组迁移） |
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
│   │   ├── h5_compat.go           # H5/Vue 前端兼容层（auth 路由生产注册，JSON 文件路由仅 dev）
│   │   └── *.go                   # 各业务 Handler（batch1-3 + biz + phase3）
│   ├── service/                   # 业务规则、权限校验、状态机流转
│   ├── repository/                # 数据持久化
│   │   ├── repositories.go        # Repository 接口定义（59 interface）
│   │   ├── postgres/              # PostgreSQL 实现（pgxpool）
│   │   └── memory/                # 内存实现（开发用，sync.RWMutex）
│   ├── domain/models.go           # 业务实体与常量（90 个 struct，含 models_batch*/models_new）
│   ├── config/config.go           # 集中配置 + 验证 + 脱敏打印
│   ├── logger/logger.go           # 结构化日志（slog + 每日文件轮转）
│   ├── cache/cache.go             # 内存 TTL 缓存（60s 默认，5min 自动清理）
│   ├── middleware/middleware.go    # 输入消毒 + 统一错误格式
│   └── crypto/                    # AES-256-GCM 加密 + 脱敏函数
├── migrations/                    # 63 组迁移（126 个 SQL 文件，85 表）
├── docs/                          # 项目文档（27 份 Markdown，中文）
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
请求 → 限流(100/s, 按RemoteAddr) → 链路追踪(requestID) → Panic恢复 → 安全头
  → CORS → Token鉴权(白名单+可选解析) → 幂等去重(24h, actor命名空间)
  → 管理端门禁(/api/v1/admin/* 仅admin) → 输入消毒(JSON去HTML标签) → 业务处理
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

协会内部细分(8级): 会长/副会长/秘书长/部门负责人/普通会员/副会长单位/合作院校/访客

## 7 大业务系统

| 系统 | 核心子模块 |
|------|------|
| ①会员生态资源管控 | 会员注册/专家智库/产业资源台账/人才资源库/协会8级权限 |
| ②产业供需智能对接 | 需求大厅/供应展示/意向对接/工单闭环/智能匹配/资源池 |
| ③产学研协同创新 | 科技成果库/研发难题广场/课题攻关/测试预约/成果转化追踪 |
| ④合规政策服务 | 政策资讯/合规知识库/团体标准库/项目申报/企业案例库 |
| ⑤人才教育与产教融合 | 培训认证/赛事管理/招聘求职/院校展示/校企共建 |
| ⑥活动与品牌服务 | 活动管理/会员品牌展示/展会排期/行业报告发布 |
| ⑦低空应急资源协同 | 应急资源/一键调度/救援案例库/部门对接/联合演练 |

## Token 认证

- **格式**: 标准 JWT (HS256) 为主（`IssueJWT`），兼容旧式两段 `payload.sig` 格式
- **过期**: Access Token 15分钟，Refresh Token 7天
- **刷新**: 轮转刷新令牌——先落库新令牌、成功后再撤销旧令牌（防 Store 失败锁号），库中只存 SHA-256 哈希

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
| 微信小程序 | `miniprogram/` | uni-app + Vue3 `<script setup>` + 自研 u- 组件库 | 103 页，5 Tab，6 分包 |
| Web 管理后台 | `frontend/` | Vue 3 + Arco Design Vue + ECharts | Admin SPA（40 后台路由 + 8 聚合页） |

**小程序设计规范**:
- 品牌色 `#0A66C2`（深空蓝），辅色 `#1DD4A8`（青绿）
- 全局 CSS 变量定义在 `App.vue` 的 `page` 选择器中
- 输入框: `bg=#fafafa` `radius=24rpx`，按钮: `radius=50rpx` + `box-shadow`
- 不使用 emoji 图标，用 CSS 绘制或文字标签

**小程序 API 调用注意**:
- `request.js` 自动 unwrap Go 后端的 `{ data: {...} }` 响应包（分页响应保留 `total`）
- 微信静默登录用 `POST /api/v1/auth/wechat/login`（生产路由，返回蛇形 `{ access_token, refresh_token, user }`）
- 密码登录用 `POST /api/auth/login`（H5 兼容层，生产注册，bcrypt 校验）
- Token 存储: `authStorage.setTokens(accessToken, refreshToken)` + `uni.setStorageSync('user', ...)`；刷新轮转，须保存新 refreshToken
- 主流程是微信静默登录（`App.vue` → `wx.login()` → `/api/v1/auth/wechat/login`），密码登录只是备用

## 关键踩坑记录

| 问题 | 根因 | 修复 |
|------|------|------|
| 注册 500: `duplicate key violates unique constraint "users_wechat_openid_key"` | 手机注册时 `wechat_openid` 为空字符串，第二次 `""` 违反 UNIQUE | 设置 `WechatOpenID: "phone:"+body.Phone` 确保唯一 |
| H5 兼容路由 404 | `/api/auth/*` 曾只在 `ADMIN_DEV_MODE=true` 时注册，生产 404 | **已修复**：auth 路由（login/register/me/refresh/logout + GET services/config）已无条件生产注册，JSON 文件路由仍 dev-only |
| 登录 401 | 小程序调 `/api/v1/login` 不存在 | 改为 `/api/auth/login` |
| 登录后 Token 不匹配 | 旧代码只存 `token`，新 API 返回 `accessToken` | 改用 `authStorage.setTokens()` |
| 小程序全部请求 `ERR_CONNECTION_TIMED_OUT` | `miniprogram/utils/config.js` 的 `BASE_URL` 写死旧局域网 IP（192.168.5.141），DHCP 换网后 IP 变了 | 用 `ipconfig` 查当前 WLAN IP（如 192.168.5.19），改 `BASE_URL` 后重新编译；真机调试需手机与电脑同一 WiFi |
| 微信一键登录变成"同一个号/多出 dev-fixed 用户" | 打包 tar 未排除 `.env`，本地 `.env` 覆盖服务器 `.env` → `WECHAT_APPID/APPSECRET` 丢失 → code2Session 失败 → `adminDevMode()` 兜底 `openid=dev-fixed` → 每次失败都建共享账号 | 打包命令必须 `--exclude='.env'`；部署后验证 `docker exec uav-api-1 env \| grep WECHAT` 非空；恢复 `~/UAV/.env` 中 WECHAT_APPID/WECHAT_APPSECRET 后 `docker compose up -d api` |
| `GET /api/v1/me` 的 demand_count 是全平台计数 | `me()` 用 `List(空 filter)` 统计全部已发布需求 | **已修复**：新增 `DemandRepository.ListByPublisher` 只统计本人需求（含回归测试） |
| refresh 轮转可能锁号 | 旧实现先 `Revoke` 旧令牌再 `Store` 新令牌，Store 失败即账号锁死；且签发的是旧式两段 Token | **已修复**：先落库新令牌、成功后再撤旧；签发统一 `IssueJWT` |
| 短信验证码可在线爆破 | 错误码无尝试次数限制、比较非常量时间 | **已修复**：5 次错误作废验证码 + `subtle.ConstantTimeCompare` |
| h5ImageProxy 开放重定向 | 任意 http/https URL 直接 302 跳转 | **已修复**：白名单（localhost/127.0.0.1/BASE_URL 域名）外一律 403 |
| `middleware.SanitizeBody` 是空壳 | 只查 Method/Content-Type 就放行，且未挂载 | **已修复**：实现真实 JSON 消毒（去 HTML 标签、password 保真、1MiB 上限）并挂载进中间件链 |

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
| 项目简介 + 快速开始 | [README.md](README.md) |
| 架构 + 分层 + 中间件 | [docs/系统架构/架构总览.md](docs/系统架构/架构总览.md) |
| 7大业务系统详情 | [docs/业务系统/](docs/业务系统/) |
| 全部 API 契约 | [docs/接口文档/API契约.md](docs/接口文档/API契约.md) |
| 85 张表结构 | [docs/数据设计/数据模型.md](docs/数据设计/数据模型.md) |
| 编码规范 | [docs/开发规范/编码规范.md](docs/开发规范/编码规范.md) |
| Docker + CI | [docs/运维部署/Docker部署.md](docs/运维部署/Docker部署.md) |
| 开发计划 | [PRD-四人并行开发方案.md](docs/项目管理/PRD-四人并行开发方案.md) |
