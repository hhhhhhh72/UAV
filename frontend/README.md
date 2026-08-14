# 无人机产业综合服务平台 — Web 管理后台

面向协会管理员的 Web 管理后台（SPA），覆盖 7 大业务系统的数据管理、审核与统计。

> 历史说明：本目录曾包含 H5 移动端前台（Vant + Element Plus），已于 2026-08 随「瘦身为纯 Admin」重构删除；本 README 已同步为纯管理后台描述。

## 技术栈

- **Vue 3.4** + `<script setup>`
- **Arco Design Vue ^2.58**（含 ArcoVueIcon，2026-08 从 Element Plus 全面切换）
- **ECharts 6** + vue-echarts 8（看板图表）
- **Vue Router 4** / **Pinia 2**（未建 store）/ **Axios**
- **Vite 5**（`/api`、`/uploads` 代理到 `VITE_API_TARGET || http://localhost:8080`）

## 快速开始

```bash
npm install
npm run dev        # → http://localhost:5173
npm run build      # 生产构建
```

- 开发登录：`POST /api/v1/admin/token {role}`（需后端 `ADMIN_DEV_MODE=true`，生产自动剔除）
- 生产登录：`POST /api/auth/login`（账号密码 + bcrypt）

## 路由结构（`src/router/index.js`）

- `/login` 登录页
- `/admin` 挂 AdminLayout，**40 条子路由**（全部懒加载，守卫校验 `user.role ∈ {platform_admin, association_admin}`）：
  - 基础 11：dashboard / cases / users / competition / config / reviews / orders / products / service-listings / enterprises / demands
  - 业务 20：experts / resources / compliance / training / certs / jobs / colleges / admin-study / achievements / challenges / projects / testsites / transformations / events / portfolios / exhibitions / reports / emergency-resources / emergency-dispatches / messages
  - 聚合 9：members / trading / content / articles / talent / innovation / promotion / emergency / settings

## 核心抽象

- **`useAdminApi(resource)`**（`src/api/admin/common.js`）：一行生成 `/api/v1/admin/{resource}` 的 list/get/create/update/delete
- **`CrudList.vue`**：配置驱动列表——页面只声明 `columns + searchFields + batchActions`；批量动作传**完整行数据**（后端 PUT 全字段覆盖语义）
- **聚合页**（`views/admin/consolidated/`）：指标卡 + ECharts + a-tabs 嵌套子列表，按 7 大系统组织
- **`utils/http.js`**：`{data}` 信封透明解包（分页保留 total）、401 单飞刷新 + pendingQueue 排队重放（`/api/auth/refresh`）

## 对接约定（易踩坑）

1. **全字段覆盖更新**：PUT/PATCH 必须提交完整行对象，否则其余字段被清空
2. `/api/v1/admin/config` 为整包替换语义：先 GET 全量、改字段后整体 POST 写回
3. 企业编辑走**非 admin 前缀** `PATCH /api/v1/enterprises/{id}`
4. 资讯列表用公开读接口 `/api/v1/articles`
5. Arco 主题色必须 `rgb(var(--primary-6))` 分量格式，写十六进制会生成非法 CSS
6. 排序参数 `sort_field/sort_order`，分页参数 `page/page_size`
