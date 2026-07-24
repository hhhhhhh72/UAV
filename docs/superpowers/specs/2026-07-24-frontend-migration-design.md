# 前端迁移设计 — microAPP-main → drone-platform

> 日期: 2026-07-24 | 方案: A (渐进替换) | 目标: 完整适配无人机产业平台

---

## 1. 背景与约束

### 1.1 起点
- 源项目: `microAPP-main/frontend/h5/` — Vue 3 + Vite 5 + Vant 4 + Pinia + Vue Router 4
- 源内容: 重庆无人机产业综合服务平台（医疗配送 + 14 项低空服务）
- Go 后端已有 `compat_routes.go` + `h5_compat.go` 覆盖 90% `/api/*` 兼容路径

### 1.2 目标
- 前端迁移到项目根 `frontend/`，完整适配重庆无人机产业协会平台
- 4 Tab + 管理后台布局，与小程序保持一致
- 对接 Go 后端 `:8080`，优先使用 `/api/v1/*` 新端点，回退兼容 `/api/*`

### 1.3 约束
- 不引入新框架依赖，复用 Vant 4 组件库
- 不修改 Go 后端代码
- 保留微APP框架能力（路由守卫/Token刷新/Axios拦截器）

---

## 2. 迁移步骤

### Step 1: 拷贝框架
| 操作 | 详情 |
|------|------|
| 拷贝 | `microAPP-main/frontend/h5/` → `frontend/` |
| 修改 | `.env.development`: `VITE_API_TARGET=http://localhost:8080` |
| 修改 | `.env.production`: 生产地址 |
| 修改 | `vite.config.js`: proxy `/api` + `/uploads` → `:8080` |
| 验证 | `npm install && npm run dev` → 首页可访问 |

### Step 2: 剥离医疗模块
| 操作 | 文件 |
|------|------|
| 删除 | `src/views/medical/` (8 文件) |
| 删除 | `src/stores/medical.js` |
| 删除 | `src/views/admin/medical/` |
| 删除 | `src/views/admin/orders/` |
| 删除 | 路由中 8 条 `/medical/*` + admin medical 路由 |

### Step 3: 品牌替换
| 项 | 原值 | 新值 |
|------|------|------|
| 标题 | 无人机产业综合服务平台 | 无人机产业综合服务平台 |
| 副标题 | 重庆无人机产业 | 重庆无人机产业协会 |
| 主色 | `#1d1d1f` | `#1565C0`（与小程序一致） |
| Logo | microAPP logo | 无人机图标 |
| SSO | 产业协会 | 移除 |

### Step 4: 首页改造
- 保留: 轮播/公告/搜索栏/城市定位
- 替换: 服务图标网格 → 7 大业务系统入口
- 新增: 需求大厅信息流（列表 + 分类筛选）
- API: `GET /api/v1/home` + `GET /api/v1/demands`

### Step 5: 服务页改造
- 业务分类: 产业供需 | 培训认证 | 无人机交易 | 合同签约 | 保险金融 | 应急协同
- 新建子页面: 需求详情/竞标报价、培训认证列表、合同管理、救援案例
- API: `/api/v1/demands`, `/api/v1/training-*`, `/api/v1/contracts`, `/api/v1/rescue-cases`

### Step 6: 消息中心
- 新建页面: 站内信列表 + 未读红点
- API: `GET /api/v1/messages`, `POST /api/v1/messages/{id}/read`, `GET /api/v1/messages/unread-count`

### Step 7: 个人中心
- 保留: 用户头像/昵称/角色标签
- 新增: 企业入驻入口、我的需求、我的证书、钱包余额、合同列表
- API: `/api/v1/enterprises`, `/api/v1/me`, `/api/v1/certificates/mine`

### Step 8: 管理后台
- 数据看板: 需求趋势/企业统计/培训认证数/成交率
- 审核管理: 企业审核(通过/驳回/补件)、需求审核、评价审核、举报处理
- 平台配置: 轮播图/公告/快捷入口
- API: `/api/v1/admin/*`

---

## 3. 路由设计

```
/                         → redirect /home
/home                     → 产业首页 (4 Tab)
/services                 → 业务大厅 (4 Tab)
/messages                 → 消息中心 (4 Tab)
/mine                     → 个人中心 (4 Tab)
/login                    → 登录
/register                 → 注册
/demand/:id               → 需求详情 + 竞标
/demand/publish           → 发布需求
/enterprise/apply         → 企业入驻
/training                 → 培训认证
/training/:id             → 课程详情
/contracts                → 合同列表
/contracts/:id            → 合同详情
/cases                    → 救援案例库
/cases/:id                → 案例详情
/reviews                  → 信用评价
/admin                    → 管理后台
/admin/enterprises        → 企业审核
/admin/demands            → 需求审核
/admin/reviews            → 评价管理
/admin/reports            → 举报管理
/admin/users              → 用户管理
/admin/config             → 平台配置
```

---

## 4. 状态管理 (Pinia Stores)

| Store | 模块 | 职责 |
|------|------|------|
| `user.js` | 用户 | 微信登录/Token/角色/刷新 |
| `home.js` | 首页 | 轮播/公告/快捷入口配置 |
| `demand.js` | 需求 | 需求列表/详情/竞标/发布 |
| `enterprise.js` | 企业 | 入驻表单/审核状态 |
| `training.js` | 培训 | 证书/课程/教练/飞手 |
| `message.js` | 消息 | 站内信列表/未读数 |
| `service.js` | 服务 | 服务分类配置（保留复用） |

删除: `medical.js`, `application.js`（功能合并到 demand.js）

---

## 5. API 对接策略

### 5.1 优先级
1. 首选 `/api/v1/*` — Go 后端原生端点（212 条，完整业务覆盖）
2. 回退 `/api/*` — compat_routes + h5_compat（开发/过渡用）

### 5.2 HTTP 客户端
- 复用 `src/utils/http.js` — Axios 实例 + Bearer Token 拦截器 + 401 自动刷新
- 修改 `baseURL` 指向 Go 后端（不再走 Vite proxy 到 Node.js）

### 5.3 Token 格式
- Go 后端 JWT 格式: `header.payload.signature`（HS256）
- 兼容旧格式: `payload.signature`（Verify 双格式支持）
- 存储: localStorage `accessToken` + `refreshToken`

---

## 6. 文件变更清单

### 6.1 删除
```
src/views/medical/                    (8 文件)
src/views/admin/medical/              (3 文件)
src/views/admin/orders/               (2 文件)
src/views/games/                      (3 文件)
src/stores/medical.js
src/stores/application.js
```

### 6.2 新建
```
src/views/messages/Index.vue          消息中心
src/views/demand/Detail.vue           需求详情
src/views/demand/Publish.vue          发布需求
src/views/enterprise/Apply.vue        企业入驻
src/views/training/Index.vue          培训认证
src/views/contracts/Index.vue         合同列表
src/stores/demand.js                  需求状态
src/stores/enterprise.js              企业状态
src/stores/training.js                培训状态
src/stores/message.js                 消息状态
src/views/admin/enterprises/          企业审核
src/views/admin/demands/              需求审核
```

### 6.3 修改
```
.env.development                      API 地址
.env.production                       API 地址
vite.config.js                        proxy 配置
src/router/index.js                   路由表
src/App.vue                           全局样式变量
src/styles/global.css                 主题色
src/views/home/Index.vue              首页内容
src/views/services/Index.vue          服务分类
src/views/mine/Index.vue              个人中心
src/views/admin/Dashboard.vue         数据看板
src/views/admin/Index.vue             审核Tabs
src/views/admin/AdminLayout.vue       Logo/标题
src/stores/user.js                    适配JWT
src/utils/http.js                     baseURL
index.html                            title/meta
```

---

## 7. 验收标准

### 7.1 功能验收
- [ ] 4 Tab 首页/服务/消息/我的 可正常切换
- [ ] 微信登录流程打通（code2Session → JWT → 首页）
- [ ] 需求大厅列表加载 + 详情页 + 竞标报价
- [ ] 企业入驻表单提交 + 审核流程
- [ ] 消息列表 + 未读红点
- [ ] 个人中心展示用户信息/证书/需求
- [ ] 管理后台数据看板 + 审核流

### 7.2 技术验收
- [ ] `npm run build` 无报错
- [ ] Go 后端 `go build && go vet && go test` 全部通过
- [ ] 所有页面 API 调用返回 200（无 404/500）
- [ ] Token 过期自动刷新，刷新失败跳登录页
- [ ] 医疗模块代码零残留

---

## 8. 风险与缓解

| 风险 | 缓解 |
|------|------|
| Go 兼容路由覆盖不全 | 优先用 `/api/v1/*` 端点，compat 仅作回退 |
| Vant 4 组件版本兼容 | 锁定现有版本不升级 |
| 微信 JS-SDK 依赖 | 微信登录先走 Go 后端 code2Session，不依赖 JSSDK |
| 小程序与 H5 体验不一致 | 共用一个设计令牌文件（30 CSS 变量） |
