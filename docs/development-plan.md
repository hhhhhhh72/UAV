# 无人机产业综合服务平台 — 四人开发计划（详细版）

**团队**: A（后端）、B（测试+质量）、C（小程序前端）、D（管理后台+设计系统）
**总工期**: 约 7 周
**总交付物**: 130 条接口 | 49 张表 | 33 个页面 | 10 个后台页面 | 100+ 测试用例

---

## 团队分工

| 代号 | 角色 | 技术栈 | 产出文件 |
|:--:|------|------|------|
| **A** | 后端开发 | Go + PostgreSQL + Docker | `internal/**/*.go` + `cmd/**/*.go` + `migrations/*.sql` |
| **B** | 测试+质量 | Go test + GitHub Actions + 安全扫描 | `internal/**/*_test.go` + `.github/workflows/*.yml` |
| **C** | 小程序前端 | 微信原生 + Vant Weapp 1.11 | `miniprogram/**/*.{js,wxml,wxss,json}` |
| **D** | 管理后台+设计系统 | HTML/CSS/JS + ECharts 5 | `internal/httpapi/admin.html` + `docs/design-system.md` |

> **C 和 D 如何协作**：D 负责全局设计规范（`docs/design-system.md`），定义色板、字体、圆角、阴影、动画参数。C 在 `app.wxss` 中用 rpx 单位实现同一套规范。两人代码完全独立，零冲突。

---

## 第一阶段：基础设施（第 1-2 周）

### 第 1 周

#### A — 项目骨架 + 中间件 + Docker

| 序号 | 任务 | 具体产出 |
|:--:|------|------|
| 1 | 创建项目目录结构 | `cmd/api/main.go` `internal/domain/` `internal/httpapi/` `internal/service/` `internal/repository/memory/` `internal/repository/postgres/` `internal/config/` `internal/logger/` `internal/crypto/` `internal/middleware/` `migrations/` `docs/` |
| 2 | 配置模块 | `internal/config/config.go` — `Load()` 从环境变量读 10 个配置项、`Validate()` 校验必填项（AUTH_SECRET 至少 32 字节）、`Print()` 脱敏打印 |
| 3 | 日志模块 | `internal/logger/logger.go` — `Init(env)` 按环境设置 slog 级别（dev=debug/prod=info）、支持 `LOG_LEVEL` 环境变量覆盖、每日日志文件轮转 |
| 4 | 请求 ID 中间件 | `func requestID(next http.Handler) http.Handler` — 注入 X-Request-ID 到 context 和响应头 |
| 5 | Panic 恢复中间件 | `func recoverPanic(next http.Handler) http.Handler` — defer recover、记录堆栈、返回 500 |
| 6 | CORS 中间件 | `func withCORS(next http.Handler) http.Handler` — 从 `CORS_ORIGINS` 环境变量读取白名单、Vary:Origin |
| 7 | 安全头中间件 | `func securityHeaders(next http.Handler) http.Handler` — X-Content-Type-Options/X-Frame-Options/HSTS/Referrer-Policy |
| 8 | 限流中间件 | `rateLimiter` struct — Token Bucket 100次/秒 burst 200、按 X-Forwarded-For 或 RemoteAddr 限流、超限返回 429 |
| 9 | 幂等中间件 | `idempotencyStore` struct — 按 Idempotency-Key 去重、24 小时 TTL、定时清理过期条目 |
| 10 | 统一响应函数 | `respond(w,r,status,data)` `fail(w,r,status,err)` `paginatedRespond(w,r,items,total)` |
| 11 | 健康检查 | `GET /healthz` — 返回 `{"status":"ok","checks":{"server":"up","storage":"memory"}}` |
| 12 | 中间件链组装 | `Router()` 函数 — 幂等→限流→请求ID→Panic恢复→安全头→CORS→业务路由 |
| 13 | Dockerfile | 多阶段构建：`golang:1.24-alpine` builder → `alpine:3.21` runtime、CGO_ENABLED=0、HEALTHCHECK |
| 14 | docker-compose.yml | PostgreSQL 16 + API 双容器、健康检查依赖、volumes |

#### B — 测试框架 + CI

| 序号 | 任务 | 具体产出 |
|:--:|------|------|
| 1 | 测试辅助函数 | `internal/httpapi/server_test.go` — `newServer(t)` 创建测试服务器、`auth(t,role)` 签发令牌、`request(t,app,method,path,body,role)` 发起 HTTP 请求 |
| 2 | 配置模块测试 | 测试 `Validate()`：AUTH_SECRET 为空→报错、长度<32→报错、production 模式缺 WECHAT_APPID→报错 |
| 3 | 限流测试 | `TestRateLimiterAllowsAndBlocks` — burst 200 通过、第 201 次返回 429 |
| 4 | 限流独立 Key 测试 | `TestRateLimiterSeparateKeys` — 两个不同 IP 各自独立限流 |
| 5 | Panic 恢复测试 | `TestPanicRecovery` — 故意触发 panic、验证返回 500 且进程不退出 |
| 6 | CORS 预检测试 | `TestCORSPreflight` — OPTIONS 请求返回 204 + 正确 Origin |
| 7 | CORS 拦截测试 | `TestCORSBlockedOrigin` — 非白名单 Origin 不返回 CORS 头 |
| 8 | CI 工作流 | `.github/workflows/ci.yml` — 两个 job：build（go build + go vet + go test）+ integration（PG service container + go test ./internal/repository/postgres/） |

#### C — 小程序初始化 + 首页

| 序号 | 任务 | 具体产出 |
|:--:|------|------|
| 1 | 项目初始化 | `miniprogram/app.json` — 注册 10 个主包页面 + 7 个分包 + 40 个 Vant Weapp 全局组件 |
| 2 | 全局样式之变量 | `miniprogram/app.wxss` — `--primary` `--bg` `--text-primary` `--shadow` `--radius` 等 25 个 CSS 变量，`--font-xs` 到 `--font-xxl` 6 级字号 |
| 3 | 全局样式之工具类 | `.flex` `.flex-center` `.flex-between` `.card` `.tag` `.btn-primary` `.text-ellipsis` `.empty-wrap` 等 30+ 工具类 |
| 4 | 网络请求封装 | `utils/api.js` — `api.get(url,params)` `api.post(url,data)` `api.patch(url,data)` `api.upload(url,filePath)` — 自动注入 Bearer Token、Token 过期自动刷新（刷新锁防并发）、X-Request-ID 追踪、统一错误处理 |
| 5 | 认证模块 | `utils/auth.js` — `wxLogin()` 调微信登录换取令牌存 globalData+Storage、`ensureLogin()` 确认登录态有效、`logout()` 撤销令牌 |
| 6 | 全局常量 | `utils/constants.js` — `bizTypeMap`(6 种业务类型) `certTypeMap` `productTypeMap`、`fenToYuan()` `formatPrice()` `formatDate()` `timeAgo()` |
| 7 | 自定义 TabBar | `custom-tab-bar/index.*` — 5 栏（🏠首页 📂分类 🛒卖机 📋任务 👤我的），选中态高亮 |
| 8 | 首页 WXML 结构 | `pages/index/index.wxml` — 渐变头部(城市+搜索)+轮播图+公告滚动+10 个快捷入口+供需双广场+需求卡片流 |
| 9 | 首页 JS 逻辑 | `pages/index/index.js` — `onLoad` 调 `/home` 获取 banner/公告/快捷入口/需求列表、下拉刷新、筛选切换、分页加载更多、登录状态管理 |
| 10 | 首页 WXSS 样式 | `pages/index/index.wxss` — 头部渐变背景、banner 圆角阴影、快捷入口图标 9 色渐变、卡片间距、底部留白 |

#### D — 设计规范 + 管理后台骨架

| 序号 | 任务 | 具体产出 |
|:--:|------|------|
| 1 | 设计规范文档 | `docs/design-system.md` — 色板(主色#4f46e5/深色#4338ca/成功#10b981/警告#f59e0b/危险#ef4444)、字体(-apple-system/PingFang SC)、6级字号(22-40rpx)、圆角(sm=10rpx/md=16rpx/lg=20rpx/round=999rpx)、5级阴影、动画(150-300ms cubic-bezier) |
| 2 | 后台 HTML 骨架 | `internal/httpapi/admin.html` — `<!DOCTYPE html>` 到 `</html>`，侧边栏 `<aside>` + `<div class="main">` + 内容区 `<div class="content">` |
| 3 | 后台 CSS 变量 | 60+ CSS 变量 — `--bg: #f8f9fb` `--surface: #fff` `--text: #1a1d28` `--accent: #4f46e5` `--sidebar-w: 232px` 等 |
| 4 | 后台侧边栏样式 | `.sidebar` — 深色渐变背景 `#111827→#1a2332`、品牌区 `.sidebar-brand`、导航分组 `.nav-group-label`、导航项 `.nav-item` hover/active 态、选中态紫色发光线 |
| 5 | 后台顶部栏 | `.header` — 页面标题+用户信息、白色底+底部边框 |
| 6 | 后台指标卡片 | `.metric-card` — 4 列网格布局、hover 上浮+顶部渐变条、`.metric-label` `.metric-value` `.metric-sub` |
| 7 | 后台面板+表格 | `.panel` `.panel-header` `table` `th` `td` — 面板圆角阴影、表头大写间距、行 hover 变色 |
| 8 | 后台按钮+标签+表单 | `.btn`(primary/success/danger/outline) `.tag`(blue/green/orange/red) `select` `input` `textarea` |
| 9 | 后台弹窗+Toast | `.modal-overlay` `.modal` — 毛玻璃背景+缩放入场动画、`.toast` — 浮动提示+渐变背景 |
| 10 | 后台页面切换 JS | `switchPage(page,el)` — 侧边栏点击切换 `.page-section` 可见、数据看板/企业审核/需求管理/举报管理/行业资讯/评价审核/用户管理/服务配置/数据导出/Token工具 共 10 个 page-section |
| 11 | 数据看板 JS | `loadStats()` — 调用 `/healthz` 查存储类型、`/admin/dashboard` 查统计数据、ECharts 渲染趋势图和饼图 |
| 12 | 企业审核 JS | `loadEnterprises()` — 调用 `/admin/enterprises` 分页加载、`reviewEnt(id,action)` 通过/驳回操作 |

### 第 2 周

#### A — 数据库 + 认证

| 序号 | 任务 | 具体产出 |
|:--:|------|------|
| 1 | 领域模型 | `internal/domain/models.go` — `Actor` `User` `Role`(4个常量) `Enterprise` `Demand` `DemandBid` `Contract` `Job` `Resume` `Post` `Comment` `Report` `Listing` `LabourOrder` `Certificate` `TrainingCourse` 等 40 个 struct + 全部状态常量 |
| 2 | 仓储接口 | `internal/repository/repositories.go` — `UserRepository` `DemandRepository` `EnterpriseRepository` `EmploymentRepository` `ContractRepository` `BidRepository` `JobRepository` 等 15 个 interface |
| 3 | PG 连接池 | `internal/repository/postgres/postgres.go` — `NewStore(ctx,url,cipher)` 创建 pgxpool、MaxConns=20、Ping、`Close()` |
| 4 | PG Demand 仓储 | `demandRepo` — `Create` `FindByID` `List` `Search` `SetStatus` `CompareAndSetStatus`（CAS 原子更新） |
| 5 | PG Enterprise 仓储 | `enterpriseRepo` — `Create` `Update` `FindByID` `FindByOwner` `ListByStatus` `Pending` `Search` |
| 6 | PG 就业+合同仓储 | `employmentRepo` `contractRepo` — `Create` `ListByEnterprise` `ListAll` `FindByID` `UpdateStatus` |
| 7 | PG Bid 仓储 | `pgBidRepo` — `Create` `FindByID` `ListByDemand` `UpdateStatus` |
| 8 | PG User+Token 仓储 | `userRepo` `memRefreshRepo` — `FindByOpenID` `Create` `FindByID` `All` `UpdateRole` / `Store` `Find` `Revoke` |
| 9 | 内存全部仓储 | `internal/repository/memory/memory.go` — 同上 9 个内存实现，带 `sync.RWMutex` 锁保护 |
| 10 | 初始化迁移 | `migrations/000001_init.up.sql` — `users` `enterprises` `demands` `contracts` `employment_requests` `refresh_tokens` `audit_logs` 等核心表 + 索引 |
| 11 | Token 管理器 | `internal/httpapi/auth.go` — `NewTokenManager(secret)` `Issue(actor,ttl)` HMAC-SHA256 签署 `{sub,role,exp}`、`Verify(token)` 验签+查过期 |
| 12 | 公开路径白名单 | `isPublicPath(path)` — 16 个 GET 端点免登录（`/home` `/search` `/demands` `/posts` 等） |
| 13 | 鉴权中间件 | `func authenticate(next http.Handler)` — 无条件放行公开路径+`/auth/*`+`/webhooks/*`+`/uploads/*`、校验 Bearer Token、注入 actor 到 context |
| 14 | 微信登录 | `internal/service/auth.go` — `WeChatLogin(code,appID,appSecret)` 调 `api.weixin.qq.com/sns/jscode2session` 换取 openid |
| 15 | 登录 Handler | `internal/httpapi/auth_wechat.go` — `POST /auth/wechat/login` 查/建 user→签发 access token(15min)+refresh token(7d 存哈希)→返回 |
| 16 | 刷新/登出/Me | `POST /auth/refresh` `POST /auth/logout` `GET /me` `PATCH /me` — Token 轮换、撤销、用户信息聚合 |
| 17 | 审计日志 | `func audit(ctx,actorID,action,resourceType,resourceID,result)` — 写 `audit_logs` 表 |
| 18 | main.go | `cmd/api/main.go` — config.Load→logger.Init→选择 PG/内存→NewServer(全部 service)→ListenAndServe |

#### B — 认证+中间件测试

| 序号 | 任务 | 具体产出 |
|:--:|------|------|
| 1 | 授权拦截测试 | `TestAuthorizationIsRequired` — 无 Token 访问写接口→401 |
| 2 | 公开路径测试 | 验证 `/healthz` `/api/v1/home` 无需 Token 可访问 |
| 3 | Token 签发验证测试 | 使用 `TestSecret` 签发→验证→过期 Token 被拒 |
| 4 | 内存仓储并发测试 | `TestEnterpriseRepoConcurrency` `TestDemandRepoConcurrency` — `go test -race` 零告警 |

#### C — 小程序页面扩展

| 序号 | 任务 | 具体产出 |
|:--:|------|------|
| 1 | 首页完善 | 公告滚动动画、banner 点击跳转、快捷入口绑定事件、需求卡片按类型着色 |
| 2 | 搜索页 | `pages/search/search.*` — 搜索框+搜索历史+需求/企业双列表展示 |
| 3 | 我的页面 | `pages/mine/index.*` — 头像+用户名+角色标签、订单统计 4 图标、钱包卡片（余额+充值/提现按钮）、推广卡片、功能菜单 8 项、退出登录 |
| 4 | 卖机商城 | `pages/market/market.*` — 商家横向滚动列表、商品双列网格、分类 tab（全部/整机/配件/维修）、价格标签、下单按钮 |

#### D — 后台核心页面 + 设计规范对齐

| 序号 | 任务 | 具体产出 |
|:--:|------|------|
| 1 | C/D 设计对齐 | D 将 `design-system.md` 中的色板/圆角/阴影参数同步给 C，C 据此更新 `app.wxss` 中的 CSS 变量 |
| 2 | 企业审核页 | 表格列（企业名称/开户账号/状态/提交时间/操作）、状态筛选下拉框（待审核/已通过/已驳回/全部）、通过/驳回/补件按钮 |
| 3 | 需求管理页 | 表格列（标题/类型/状态/发布时间/操作）、通过/驳回按钮 |
| 4 | 表格通用组件 | 斑马纹、排序箭头、分页信息 |
| 5 | ECharts 图表 | `renderTrendChart` `renderBizTypeChart` `renderStatusChart` — 折线图+饼图、tooltip 格式化、响应式 resize |

### 第一阶段验收清单

```
□ go build ./... 通过（零编译错误）
□ go vet ./... 通过（零静态分析告警）
□ go test ./internal/... 通过（约 15 个测试用例）
□ curl http://localhost:8080/healthz → 200
□ curl http://localhost:8080/api/v1/home → 200
□ curl -X POST http://localhost:8080/api/v1/auth/wechat/login → 200 (dev模式假登录)
□ 小程序微信开发者工具编译通过，首页和我的页面可交互
□ 管理后台 http://localhost:8080/admin 可访问，侧边栏可切换
□ 设计规范 docs/design-system.md 已交付
□ C 的 app.wxss 变量与 D 的规范一致
□ CI (GitHub Actions) build + test 双绿
```

---

## 第二阶段：核心业务（第 3-4 周）

### 第 3 周

#### A — 企业入驻 + 需求大厅 + 竞标

| 序号 | 任务 | 具体接口/函数 |
|:--:|------|------|
| 1 | 企业入驻 Service | `EnterpriseSvc.Create(a,in)` — 生成 ID→设 draft→Create、`Update(a,id,in)` — 查归属→改字段、`Submit(a,id)` — 校验归属→draft→submitted、`Review(a,id,action,reason)` — 校验角色→approve/reject/supplement、`ListByStatus(a,status,offset,limit)` — 分页查询、`ListMine(a)` — 查自己的 |
| 2 | 企业入驻 Handler | `POST /api/v1/enterprises` `PATCH /api/v1/enterprises/{id}` `POST /api/v1/enterprises/{id}/submit` `GET /api/v1/admin/enterprises?status=&page=&page_size=` `POST /api/v1/admin/enterprises/{id}/review` `GET /api/v1/enterprises` |
| 3 | 企业批量审批 | `POST /api/v1/admin/enterprises/batch-review` — 接收 `{ids:[...], action, reason}`→逐条调 Review→返回 `{total,results}`，最多 50 条/批 |
| 4 | 企业搜索 | `GET /api/v1/admin/enterprises/search?q=xxx` — 按名称模糊搜索 |
| 5 | 需求大厅 Service | `DemandService.Create(a,in)` — 生成 ID→设 pending→Create、`List(f)` — 筛选(状态/区域/类型)+排序、`Review(a,id,action,reason)` — approve→published / reject→rejected |
| 6 | 需求大厅 Handler | `POST /api/v1/demands` `GET /api/v1/demands?biz_type=&sort=&page=&page_size=` `PATCH /api/v1/demands/{id}` `POST /api/v1/demands/{id}/submit` `POST /api/v1/admin/demands/{id}/review` |
| 7 | 竞标 Service | `CreateBid(a,demandID,amountFen,proposal)` — 校验角色+需求存在+需求 published+不是自己→Create、`SelectBid(a,demandID,bidID)` — 校验发布者+bid 归属+需求 published→CAS published→matched→bid→accepted |
| 8 | 竞标 Handler | `POST /api/v1/demands/{id}/applications` `POST /api/v1/demands/{id}/applications/{bidId}/select` |
| 9 | 双确认+争议 | `ConfirmComplete(a,id)` — 双方各确认一次→计数 2→completed、`Dispute(a,id,reason)` — 仅发布者可发起 |
| 10 | 文件上传 | `POST /api/v1/files/upload` — multipart、10MB 限制、JPEG/PNG/WebP/PDF、SHA256 去重、存储到 `uploads/` |

#### B — 核心流程测试

| 序号 | 测试用例 | 验证点 |
|:--:|------|------|
| 1 | `TestDemandService_Create` | 创建需求→状态 pending、标题为空→报错 |
| 2 | `TestDemandService_CreateBid` | 竞标→持久化、投自己→报错 |
| 3 | `TestDemandService_SelectBid` | 发布者选标→成功、非发布者选标→报错、bid 不属于该需求→报错 |
| 4 | `TestDemandService_SelectBid_AlreadyMatched` | 已匹配需求选标→报错（CAS 防护） |
| 5 | `TestDemandService_Review` | approve→published、reject→rejected、reject 没原因→报错、非管理员→报错 |
| 6 | `TestDemandService_ConfirmComplete_DualConfirm` | 首次确认→未完成、二次确认→completed |
| 7 | `TestDemandService_Dispute` | 发布者争议→成功、非发布者→报错 |
| 8 | `TestContractService_Create` | 企业可创建、个人→报错 |
| 9 | `TestContractService_UpdateStatus` | Valid transition→成功、Invalid→报错、非归属企业→报错 |
| 10 | `TestContractService_FullLifecycle` | draft→sent→signing→signed→voided 全路径→成功 |
| 11 | `TestEmploymentService_CreateAndList` | 创建+分页列表→值一致 |
| 12 | `TestBidCreateAndSelectFlow` | HTTP 端到端：创建 demand→approve→create bid→select bid→验证 200 |
| 13 | `TestContractWebhookFlow` | HTTP 端到端：创建 contract→POST webhook→状态更新→重复 webhook 被去重 |
| 14 | `TestEmploymentListPagination` | 分页参数→返回 total 字段 |

#### C — 小程序核心业务页面

| 序号 | 任务 | 页面文件 |
|:--:|------|------|
| 1 | 需求大厅列表 | `pages/demand/list.*` — 7 种业务类型 tab 横向滚动、3 种排序（最新/预算最高/最近距离）、分页加载更多、`service-card` 左图右文布局 |
| 2 | 需求详情 | `pages/demand/detail/detail.*` — 图片轮播、需求描述、预算价格、竞标列表（报价+方案）、选标按钮、双确认按钮、争议入口 |
| 3 | 需求发布 | `pages/demand/publish/publish.*` — 标题/描述/联系方式/业务类型选择/预算/图片上传/提交按钮 |
| 4 | 企业入驻 | `pages/enterprise/apply.*` — 企业名称/开户账号/营业执照上传/提交按钮 |

#### D — 后台业务页面扩展

| 序号 | 任务 | 具体产出 |
|:--:|------|------|
| 1 | 企业审核增强 | 搜索输入框→调用 `/admin/enterprises/search?q=`、全选复选框→`toggleSelectAllEnt`、批量通过/驳回按钮→`batchReviewEnterprises`、批量操作结果 Toast |
| 2 | 举报管理 | 表格列（类型/资源ID/原因/举报人/状态） |
| 3 | 行业资讯 | 新建按钮→`showArticleModal`→标题/分类/来源/内容→`createArticle`→`POST /articles` |
| 4 | 评价审核 | 表格列（评分/内容/对象/用户/状态/时间/操作）→通过/驳回/删除 |
| 5 | 用户管理 | 表格列（用户ID/角色/状态）→新建用户弹窗→修改角色 |

### 第 4 周

#### A — 合同 + 资金 + 招聘 + 社区

| 序号 | 任务 | 具体接口/函数 |
|:--:|------|------|
| 1 | 合同 Service | `ContractService.Create(a,v)` — draft→Create、`UpdateStatus(a,id,newStatus)` — 角色校验+归属校验+状态机校验→UpdateStatus |
| 2 | 合同 Handler | `GET /contract-templates` `POST /contracts` `POST /contracts/{id}/void` |
| 3 | 签章 Webhook | `POST /webhooks/signing` — 验签(SIGNING_SECRET)→去重(event_id)→状态映射(sent/signing/signed/voided)→UpdateStatus、失败不消费 event_id 允许重试 |
| 4 | 资金 Service | `EscrowService` — `Deposit` `Freeze` `Release` `Refund` `GetBalance` `GetTransactions` |
| 5 | 资金 Handler | `POST /escrow/deposit` `POST /escrow/freeze` `POST /escrow/release` `POST /escrow/refund` `GET /escrow/balance` `GET /escrow/transactions` |
| 6 | 招聘 Service | `JobService` — `Create` `Update` `FindByID` `ListPublished` `ListByEnterprise` `Publish` `Close`、`ResumeService`、`ApplicationService` |
| 7 | 招聘 Handler | `POST /jobs` `GET /jobs` `POST /jobs/{id}/publish` `POST /jobs/{id}/close` `POST /resumes` `PATCH /resumes/{id}` `POST /applications` `PATCH /applications/{id}/status` |
| 8 | 社区 Service | `CommunityService` — `CreatePost` `PublishPost` `RemovePost` `ListPublishedPosts` `CreateComment` `CreateReport` |
| 9 | 社区 Handler | `POST /posts` `GET /posts` `POST /posts/{id}/publish` `POST /posts/{id}/comments` `POST /reports` `GET /admin/reports` |
| 10 | 加密模块 | `internal/crypto/` — `NewCipher(key)` AES-256-GCM、`Encrypt(plaintext)` `Decrypt(ciphertext)`、`MaskPhone(s)` `MaskCreditCode(s)` `MaskIDCard(s)` |
| 11 | PG 迁移补全 | `migrations/000002_business_modules.up.sql` 到 `000012_demand_bids.up.sql` — 全部扩展表+索引+外键 |

#### B — 并发 + 权限测试

| 序号 | 测试用例 | 验证点 |
|:--:|------|------|
| 1 | `TestConcurrent_200EnterpriseRegistrations` | 200 goroutine 同时注册→全部成功、耗时 <5s |
| 2 | `TestConcurrent_100BidsOnOneDemand` | 100 人同时竞标→≥90 成功 |
| 3 | `TestConcurrent_SelectBidRace` | 50 并发选标→严格只有 1 个成功（CAS 原子性） |
| 4 | `TestConcurrent_DualConfirmRace` | 20 并发确认→最终 DemandCompleted |
| 5 | `TestContractCreatePermissions` | 企业→201、个人→403 |
| 6 | `TestPendingEnterpriseNeedsAssociationRole` | 个人查看待审企业→403 |
| 7 | `TestEmploymentListRequiresPermission` | 个人查用工列表→403、企业→200 |

#### C — 小程序合同+资金页面

| 序号 | 任务 | 页面文件 |
|:--:|------|------|
| 1 | 合同签约 | `pages/contract/list.*` — 合同卡片(status tag+模板+时间)、新建弹窗(模板选择)、作废确认 |
| 2 | 资金托管 | `pages/finance/wallet.*` — 余额展示(大数字+冻结金额)、4 个操作按钮(充值/冻结/释放/退款)、交易记录列表(图标+金额+时间) |
| 3 | 卖机商城完善 | 商家横滚卡片、商品详情跳转、下单确认弹窗 |

#### D — 后台页面收尾

| 序号 | 任务 | 具体产出 |
|:--:|------|------|
| 1 | 用户管理完善 | 新建用户弹窗（ID+角色选择）、修改角色下拉选择 |
| 2 | 服务配置 | `openHomeConfigPopup` — Banner 管理（上传/预览/删除/排序/链接）、公告编辑（增删改） |
| 3 | 数据导出 | CSV 导出卡片（需求数据/企业数据）→`GET /admin/export/demands` |
| 4 | Token 工具 | Token 生成（角色选择+有效期）、当前 Token 显示、清除按钮 |
| 5 | ECharts 配色 | `COLORS` 对象与设计系统色板对齐（blue=#4f46e5/green=#10b981/orange=#f59e0b 等） |

### 第二阶段验收清单

```
□ curl 企业入驻全流程：POST enterprises→submit→admin review approve→200
□ curl 需求全流程：POST demands→submit→approve→POST bid→select bid→complete×2→completed
□ curl 合同 Webhook：POST contracts→POST /webhooks/signing→contract.status==sent
□ go test -run TestConcurrent 全部 PASS
□ 并发选标：严格只有 1 人中标
□ 小程序 12+ 页面可交互
□ 管理后台 8 个功能页面可用
```

---

## 第三阶段：扩展业务（第 5-6 周）

### 第 5 周

#### A — 二手 + 用工 + 交易

| 序号 | 模块 | 具体接口/函数 |
|:--:|------|------|
| 1 | 二手交易 | `POST /listings` `GET /listings?status=&page=` `POST /listings/{id}/close` `POST /listings/{id}/favorites` — 商品 CRUD + 收藏 + 下架 |
| 2 | 用工派遣 | `POST /labour-orders` `GET /labour-orders` `POST /labour-orders/{id}/quote` `GET /labour-orders/quotes` `POST /assignments` — 订单+报价+分配 |
| 3 | 无人机交易 | `POST /products` `GET /products?prod_type=` `POST /repairs` `GET /repairs/mine` `POST /trade-orders` `PATCH /trade-orders/{id}/status` — 整机/配件/维修+交易订单 |
| 4 | 培训认证 | `POST /certificates` `GET /certificates/mine` `POST /admin/certificates/{id}/approve` `POST /training-courses` `GET /training-courses` `POST /training-courses/{id}/enroll` `POST /instructors` `POST /admin/instructors/{id}/approve` `POST /certified-pilots` `POST /admin/certified-pilots/{id}/approve` |
| 5 | 保险服务 | `POST /policies` `GET /policies/mine` `POST /inspections` `GET /inspections/mine` `GET /inspections/expiring?days=30` |

#### A — 第 6 周

| 序号 | 模块 | 具体接口/函数 |
|:--:|------|------|
| 6 | 金融服务 | `POST /loans` `GET /loans/mine` |
| 7 | 信用评价 | `POST /reviews` `GET /reviews?target_type=&target_id=` `GET /admin/reviews` `POST /admin/reviews/{id}/approve` `POST /admin/reviews/{id}/reject` `DELETE /admin/reviews/{id}` |
| 8 | 场地预约 | `POST /venues` `GET /venues` `POST /venues/{id}/book` |
| 9 | 行业资讯 | `POST /articles` `GET /articles?category=` `POST /articles/{id}/publish` |
| 10 | 管理工具 | `GET /admin/dashboard` — 统计数据聚合、`GET /admin/export/demands` `GET /admin/export/enterprises` — CSV 导出(UTF-8 BOM)、`POST /admin/demands/batch-approve`、`GET /admin/config` `POST /admin/config` — 首页配置、`POST /admin/users` `GET /admin/users` `POST /admin/users/{id}/role` |
| 11 | 全端点分页 | 所有 list 端点统一使用 `paginatedRespond`，补全 Employment/Contract/Labour 分页 |
| 12 | 全 Service 日志 | 所有 Service 文件（services.go/enterprise.go/jobs.go/community.go/listings_labour.go/training.go）的关键操作加 `slog.Info/Error` |

#### B — 全面测试

| 序号 | 测试用例 | 验证点 |
|:--:|------|------|
| 1 | `TestStress_2000MixedOperations` | 10 workers×3秒持续负载、统计 P50/P95/P99 |
| 2 | `TestStress_MemoryLeakDetection` | 10000 次 CRUD→HeapObjects 正常 |
| 3 | `TestStress_GoroutineBurst` | 500 goroutine 同时竞标+选标→0 goroutine 泄漏 |
| 4 | `TestStress_DeadlockDetection` | approve/search/list 三种操作×500 次→3 秒内无死锁 |
| 5 | `TestStress_AdminDashboardLoad` | 200 企业+200 需求→20 管理员×100 次刷新→P99<100ms |
| 6 | `TestExtreme_1000GoroutineSustained` | 100 workers×5秒→统计吞吐量+P50/P95/P99 |
| 7 | `TestExtreme_2000GoroutineBidFlood` | 2000 并发竞标→100% 成功+零泄漏 |
| 8 | `TestExtreme_SelectBidAtomicity` | 500 并发选标→严格只有 1 个胜出 |
| 9 | PG 集成测试 | `TestDemandRepo_CreateAndFind` `TestDemandRepo_CompareAndSetStatus` `TestBidRepo_CreateAndList` `TestContractRepo_CreateAndUpdateStatus` `TestEmploymentRepo_ListWithPagination` |

#### C — 小程序扩展

| 序号 | 页面 | 文件 |
|:--:|------|------|
| 1 | 招聘求职 | `pages/jobs/list.*` `pages/jobs/detail.*` `pages/jobs/resume.*` |
| 2 | 二手交易 | `pages/trading/listings.*` |
| 3 | 用工派遣 | `pages/labour/list.*` `pages/labour/detail.*` |
| 4 | 培训认证 | `pages/training/courses.*` `pages/training/certificates.*` `pages/training/pilots.*` |
| 5 | 保险/金融 | `pages/finance/insurance.*` `pages/finance/loans.*` |
| 6 | 评价/场地/资讯 | `pages/reviews/list.*` `pages/venues/list.*` `pages/news/list.*` |
| 7 | 消息中心 | `pages/messages/list.*` — 未读红点+未读计数+左滑删除 |
| 8 | 全局状态 | 所有页面加 `van-loading` `van-empty` 和错误提示 |

#### D — 后台扩展

| 序号 | 任务 | 具体产出 |
|:--:|------|------|
| 1 | 服务配置完善 | Banner 上传预览排序、公告增删改保存→调用 `/admin/config` |
| 2 | 数据导出完善 | 点导出卡片→`window.open('/admin/export/...')` →下载 CSV |
| 3 | 响应式布局 | 移动端（<768px）侧边栏 off-screen 滑入、指标卡片 2 列、图表全宽 |
| 4 | 动画细化 | 页面切换 fadeIn、弹窗 scaleIn、Toast slideDown、卡片 hover lift |
| 5 | 空状态 | 所有表格的空数据 `.empty` 居中图标+文案 |

### 第三阶段验收清单

```
□ 全部 16 个模块 130 条接口可用
□ 全部 33 个小程序页面可操作
□ 管理后台 10 个功能页面可用
□ go test ./internal/... 全部 PASS（目标 100+ 测试用例）
□ 极限压力测试通过（零错误/无泄漏/无死锁）
□ Service 覆盖率 >60%
```

---

## 第四阶段：上线收尾（第 7 周）

### A — 上线准备

| 序号 | 任务 | 具体产出 |
|:--:|------|------|
| 1 | 迁移 down 文件 | `migrations/000003~000011` 的 `.down.sql` — 每组的 DROP TABLE/ALTER COLUMN |
| 2 | `.env.example` | 10 个环境变量+说明+占位符 |
| 3 | `.dockerignore` | 排除 `.git` `.claude` `node_modules` `*.exe` `logs` `uploads` 等 |
| 4 | `.gitignore` | 排除 `.env` `*.exe` `uploads/` `.cache/` 等 |
| 5 | 密钥审查 | 确认 `start.bat` `docker-compose.yml` `.claude/settings.json` 中无真实密钥 |
| 6 | 错误消息审查 | `grep "error" internal/httpapi/*.go` 确认不泄露内部路径/用户名/密码 |
| 7 | 部署文档 | README.md 更新 — 环境变量表+启动步骤+架构图 |

### B — 安全审计

| 序号 | 任务 | 具体产出 |
|:--:|------|------|
| 1 | 权限矩阵全验证 | 4 种角色(platform_admin/association_admin/enterprise/individual)×130 条端点→越权测试通过 |
| 2 | 代码审计报告 | `docs/audit-report-2026-07-21.md` — API对照+数据模型对照+架构分层合规+命名规范 |
| 3 | 压力测试报告 | P50/P95/P99 延迟+吞吐量(ops/sec)+内存/goroutine 泄漏检测 |
| 4 | gosec 扫描 | `gosec ./...` — 零 CRITICAL |

### C — 小程序收尾

| 序号 | 任务 | 具体产出 |
|:--:|------|------|
| 1 | 真机测试 | 微信开发者工具→真机预览→扫码测试核心流程 |
| 2 | 分包加载确认 | 7 个分包加载无报错 |
| 3 | 占位图替换 | `miniprogram/images/default-avatar.png` `placeholder.png` 为有效 128×128 PNG |
| 4 | console 清理 | `grep -rn "console\." miniprogram/pages/` →全部移除 |

### D — 后台收尾

| 序号 | 任务 | 具体产出 |
|:--:|------|------|
| 1 | ECharts SRI | `<script integrity="sha384-...">` 计算并填入 |
| 2 | 安全头验证 | `curl -I` 确认 X-Content-Type-Options/X-Frame-Options/HSTS 响应 |
| 3 | 视觉走查 | 逐页对比色板/字体/间距/圆角/阴影与 `design-system.md` |
| 4 | 设计规范终稿 | 更新 `docs/design-system.md` — 加上后台暗色侧边栏规范 |

### 第四阶段验收清单

```
□ go test ./internal/... 全部 PASS（目标 100+ 用例）
□ 安全审计零 CRITICAL + 零 HIGH
□ 文档(api-contract/data-model/architecture)与代码一致
□ 小程序真机运行正常、分包正常
□ 管理后台生产模式(ADMIN_DEV_MODE=false)→/admin 返回 403
□ docker-compose up -d → curl http://localhost:8080/healthz → 200
□ CI/CD 全绿(build + integration)
```
