# PRD — 无人机产业综合服务平台 前端开发

> 日期: 2026-07-26 | 团队: 4人 | 后端: 已就绪 | 管理后台: 已重建

---

## 一、项目背景

重庆市无人机产业协会委托，面向 **180+ 会员单位**，覆盖无人机全产业链的数字化服务平台。

### 当前状态

| 层 | 状态 | 说明 |
|------|:--:|------|
| **Go 后端** | ✅ 100% | 212 条 API、66 张表、7 大业务系统全部完成 |
| **H5 前端** | 🟡 框架 | 首页/业务大厅/消息/我的/登录 — 6大分类入口可用 |
| **H5 管理后台** | 🟢 已重建 | Element Plus 桌面化 — 7 个列表页(企业/需求/用户/案例/评价/订单/赛事) + 数据看板 |
| **小程序** | 🟡 框架 | 同 H5，Vant Weapp 组件就绪，CSS 变量系统 + StateView 组件已引入 |
| **42 业务页面** | ❌ 0/42 | 全部待开发（依小程序开发规格文档） |

---

## 二、产品目标

### 2.1 交付物

| 端 | 目标 | 页面数 |
|------|------|:--:|
| 微信小程序 | 完整 7 大业务系统 + 通用页面 | 42 页 |
| H5 Web 端 | 与小程序功能对齐，共享 API | 42 页 |
| 管理后台 | 数据看板 + 审核中台 | ~15 页 |

### 2.2 设计原则

- **后端不动**：所有 API 已就绪，前端只做调用和展示
- **小程序优先**：先完成小程序 42 页，H5 复用同一批 API 同步跟进
- **组件复用**：H5(Vant4) ↔ 小程序(Vant Weapp) 组件名一致，减少心智负担
- **4 Tab 布局**：首页/业务大厅/消息/我的 + 独立管理后台入口

---

## 三、7 大业务系统页面清单

### ① 会员生态资源管控（6 页）

| # | 页面 | 路径 | API |
|:--:|------|------|------|
| 1 | 企业注册 | `pages/enterprise/register` | `POST /api/v1/enterprises` |
| 2 | 企业审核状态 | `pages/enterprise/status` | `GET /api/v1/enterprises` |
| 3 | 专家智库列表 | `pages/experts/list` | `GET /api/v1/experts` |
| 4 | 专家详情 | `pages/experts/detail` | `GET /api/v1/experts/{id}` |
| 5 | 产业资源台账 | `pages/resources/list` | `GET /api/v1/industry-resources` |
| 6 | 资源详情+预约 | `pages/resources/detail` | `GET /api/v1/industry-resources/{id}` |

### ② 产业供需智能对接（6 页）

| # | 页面 | 路径 | API |
|:--:|------|------|------|
| 7 | 需求大厅 | `pages/demands/list` | `GET /api/v1/demands` |
| 8 | 需求详情 | `pages/demands/detail` | `GET /api/v1/demands/{id}` |
| 9 | 需求发布 | `pages/demands/publish` | `POST /api/v1/demands` |
| 10 | 竞标报价 | `pages/demands/bid` | `POST /api/v1/demands/{id}/applications` |
| 11 | 我的需求 | `pages/demands/mine` | `GET /api/v1/demands?mine=1` |
| 12 | 智能推荐 | `pages/match/recommend` | `GET /api/v1/recommendations` |

### ③ 产学研协同创新（6 页）

| # | 页面 | 路径 | API |
|:--:|------|------|------|
| 13 | 成果库列表 | `pages/achievements/list` | `GET /api/v1/achievements` |
| 14 | 成果详情 | `pages/achievements/detail` | `GET /api/v1/achievements/{id}` |
| 15 | 研发难题广场 | `pages/challenges/list` | `GET /api/v1/rd-challenges` |
| 16 | 课题攻关列表 | `pages/projects/list` | `GET /api/v1/research-projects` |
| 17 | 测试场地预约 | `pages/testsites/book` | `GET /api/v1/test-sites`, `POST /api/v1/test-sites/{id}/book` |
| 18 | 成果转化追踪 | `pages/transformations/track` | `GET /api/v1/transformations` |

### ④ 合规政策服务（5 页）

| # | 页面 | 路径 | API |
|:--:|------|------|------|
| 19 | 政策资讯列表 | `pages/compliance/news` | `GET /api/v1/articles` |
| 20 | 合规知识库 | `pages/compliance/knowledge` | `GET /api/v1/compliance-docs` |
| 21 | 团体标准库 | `pages/compliance/standards` | `GET /api/v1/compliance-standards` |
| 22 | 项目申报 | `pages/applications/submit` | `POST /api/v1/project-applications` |
| 23 | 案例库列表 | `pages/cases/list` | `GET /api/v1/cases` |

### ⑤ 人才教育融合（8 页）

| # | 页面 | 路径 | API |
|:--:|------|------|------|
| 24 | 培训课程列表 | `pages/training/courses` | `GET /api/v1/training-courses` |
| 25 | 课程详情+报名 | `pages/training/enroll` | `POST /api/v1/training-courses/{id}/pay-and-enroll` |
| 26 | 我的证书 | `pages/training/certificates` | `GET /api/v1/certificates/mine` |
| 27 | 赛事列表 | `pages/competitions/list` | `GET /api/v1/competitions` |
| 28 | 赛事报名 | `pages/competitions/register` | `POST /api/v1/competitions/{id}/register` |
| 29 | 职位列表 | `pages/jobs/list` | `GET /api/v1/jobs` |
| 30 | 简历管理 | `pages/jobs/resume` | `POST /api/v1/resumes` |
| 31 | 院校展示 | `pages/colleges/list` | `GET /api/v1/colleges` |

### ⑥ 活动与品牌服务（6 页）

| # | 页面 | 路径 | API |
|:--:|------|------|------|
| 32 | 活动列表 | `pages/events/list` | `GET /api/v1/events` |
| 33 | 活动详情+报名 | `pages/events/detail` | `POST /api/v1/events/{id}/register` |
| 34 | 品牌展示 | `pages/portfolios/list` | `GET /api/v1/portfolios` |
| 35 | 展会列表 | `pages/exhibitions/list` | `GET /api/v1/exhibitions` |
| 36 | 展位申请 | `pages/exhibitions/booth` | `POST /api/v1/exhibitions/{id}/booths` |
| 37 | 行业报告 | `pages/reports/list` | `GET /api/v1/industry-reports` |

### ⑦ 应急资源协同（4 页）

| # | 页面 | 路径 | API |
|:--:|------|------|------|
| 38 | 应急资源列表 | `pages/emergency/resources` | `GET /api/v1/emergency-resources` |
| 39 | 调度记录 | `pages/emergency/dispatches` | `GET /api/v1/emergency-dispatches` |
| 40 | 救援案例库 | `pages/emergency/cases` | `GET /api/v1/rescue-cases` |
| 41 | 应急部门对接 | `pages/emergency/depts` | `GET /api/v1/emergency-depts` |

### 通用页面（已部分完成）

| # | 页面 | 状态 |
|:--:|------|:--:|
| 42 | 全局搜索 | ❌ 待开发 |
| — | 首页 | ✅ 框架可用 |
| — | 消息中心 | ✅ 已开发 |
| — | 个人中心 | ✅ 框架可用 |
| — | 微信登录 | ✅ 已打通 |
| — | 管理后台 | ❌ 待重建 |

---

## 四、4 人团队分工

### 方案：按业务系统拆分

| 成员 | 系统 | 页数 | 优先级 |
|------|------|:--:|:--:|
| **A**（组长） | ② 供需对接 + 基础架构 | 6+6=12 | P0 |
| **B** | ① 会员资源 + ④ 合规政策 | 6+5=11 | P0/P1 |
| **C** | ⑤ 人才教育 + ⑦ 应急协同 | 8+4=12 | P1 |
| **D** | ③ 产学研 + ⑥ 活动品牌 | 6+6=12 | P1 |

### A 额外负责
- 管理后台重建（~15 页）
- 通用组件库维护（卡片/表单/筛选栏/空状态/骨架屏）
- API 对接规范 + 全局搜索页

### 工作量估算

| 指标 | 数据 |
|------|:--:|
| 总页面数 | 42 页 + 管理后台 ~15 页 = **57 页** |
| 人均页面 | ~14 页 |
| 每页工时 | 0.5-1 天（列表页简单/表单页复杂） |
| 预估工期 | **3-4 周**（4 人并行） |

---

## 五、Sprint 计划

### Sprint 1（第 1 周）：基础 + 展示型页面

| 成员 | 交付 |
|------|------|
| A | 需求大厅列表 + 详情 + 全局搜索 |
| B | 企业注册 + 专家智库列表 |
| C | 培训课程列表 + 赛事列表 |
| D | 成果库列表 + 活动列表 |

### Sprint 2（第 2 周）：交互型页面

| 成员 | 交付 |
|------|------|
| A | 需求发布 + 竞标报价 + 智能推荐 |
| B | 产业资源台账 + 合规知识库 + 政策资讯 |
| C | 课程报名 + 职位列表 + 简历管理 |
| D | 研发难题广场 + 测试场地预约 + 展位申请 |

### Sprint 3（第 3 周）：剩余页面 + 管理后台

| 成员 | 交付 |
|------|------|
| A | 管理后台（看板+企业审核+需求审核） |
| B | 企业案例库 + 团体标准库 + 行业报告 |
| C | 应急资源 + 救援案例 + 调度记录 |
| D | 品牌展示 + 成果转化追踪 + 院校展示 |

### Sprint 4（第 4 周）：联调 + 测试 + 上线

| 成员 | 交付 |
|------|------|
| 全员 | H5 ↔ 小程序对齐、E2E 测试、Bug 修复、性能优化 |

---

## 六、技术规范

### 6.0 技术栈

| 端 | 框架 | UI 库 | 状态管理 | 构建工具 |
|------|------|------|------|------|
| H5 C端 | Vue 3 + Composition API | Vant 4 | Pinia | Vite 5 |
| H5 管理后台 | Vue 3 + Composition API | Element Plus 2 | Pinia | Vite 5 |
| 小程序 | uni-app (Vue 3) | Vant Weapp 1.11 | Storage | uni-app CLI |

### 6.1 页面模板

每个页面对应一个 Vue SFC（小程序 uni-app .vue，H5 Vue3 .vue），结构统一：

```
pages/[module]/[page].vue
  ├── <template>  — 组件渲染（C端 Vant / 管理后台 Element Plus）
  ├── <script>    — API 调用 / 数据状态 / 交互逻辑
  └── <style>     — 全局 CSS 变量引用
```

### 6.2 管理后台组件模式（2026-07-26 更新）

管理后台列表页统一使用 Element Plus 桌面组件，遵循以下模式：

```
src/views/admin/[module]/[Module]List.vue
  ├── 搜索过滤区  — el-input + el-select + el-date-picker + 搜索/重置按钮
  ├── 批量操作栏  — 勾选后显示批量通过/驳回（el-table selection）
  ├── 数据表格    — el-table（stripe/border/sortable） + v-loading 加载态
  ├── 分页        — el-pagination（total/sizes/pager/jumper）
  └── 详情弹窗    — el-dialog + el-descriptions（替代 van-popup 底部弹窗）
```

数据获取统一使用 `useListRequest` Hook：

```js
// src/hooks/useListRequest.js — 封装分页/搜索/排序/选中/批量操作/loading
const { listData, loading, total, selectedIds, filterParams,
        loadData, onSearchSubmit, onSortChange, onSelectChange,
        onBatchAction, resetParams } = useListRequest({
  apiFunction: getEnterpriseList,  // API 函数
  idKey: 'id',                     // 行唯一标识
  defaultParams: { status: 'submitted' }  // 默认查询参数
})
```

API 调用统一模块化：

```
src/api/admin/
  ├── enterprise.js   — 企业审核 CRUD
  ├── demand.js       — 需求管理
  ├── user.js         — 用户管理
  ├── review.js       — 评价管理
  └── application.js  — 申请单管理
```

### 6.3 必须处理的 4 种状态

| 状态 | UI（C端） | UI（管理后台） |
|------|----------|--------------|
| 加载中 | `van-loading` 或 StateView | `el-table v-loading` |
| 空数据 | StateView（`:empty="true"`） | `el-empty` |
| 错误 | StateView（`:error="true"` @retry） | `showFailToast` |
| 正常 | 数据渲染 | 数据渲染 |

小程序统一 StateView 组件：

```vue
<!-- components/StateView.vue — loading / error / empty 三态统一 -->
<StateView :loading :error :empty empty-text="暂无数据" @retry="fetchData">
  <view v-for="item in list">...</view>
</StateView>
```

### 6.4 API 调用规范

```js
// H5 端：使用 axios 实例（自动携带 Token + 自动刷新 + 分页响应保留元数据）
import axios from '@/utils/http'
const res = await axios.get('/api/v1/admin/enterprises', { params: { page: 1, page_size: 20, status: 'submitted' } })
// res.data = { data: [...], total: 100, page: 1, page_size: 20 }
// — 含 total 的分页响应不解包，保留全部元数据

// 小程序端：使用 utils/request
const { request } = require('@/utils/request')
const res = await request({ url: '/api/v1/demands', data: { biz_type, page: 1, page_size: 20 } })
```

### 6.5 关键 API 分页参数

```
GET /api/v1/demands?biz_type=巡检&district=南岸区&sort=newest&page=1&page_size=20
GET /api/v1/experts?field=低空管控&page=1&page_size=20
GET /api/v1/achievements?field=无人机&page=1&page_size=20
```

### 6.6 设计令牌（2026-07-26 更新）

小程序全局 CSS 变量系统（`App.vue`）：

| Token | 值 | 用途 |
|------|------|------|
| `--color-primary` | `#1989fa` | 主色 |
| `--color-success` | `#34c759` | 成功 |
| `--color-warning` | `#ff9f0a` | 警告 |
| `--color-danger` | `#ff3b30` | 危险 |
| `--color-bg` | `#f5f6f8` | 页面背景 |
| `--color-bg-card` | `#ffffff` | 卡片背景 |
| `--color-text` | `#1a1a1a` | 主文字 |
| `--color-text-secondary` | `#969799` | 辅助文字 |
| `--radius-sm/md/lg` | `8/16/24rpx` | 三级圆角 |
| `--shadow-sm/md/lg` | — | 三级阴影 |
| `--font-xs~xxl` | `20~40rpx` | 六级字号 |
| `--tabbar-height` | `50px` | TabBar 高度 |
| `--safe-bottom` | `env(safe-area-inset-bottom)` | 底部安全区 |

工具类：`flex-center` / `flex-between` / `text-ellipsis` / `card` / `bg-white` / `shadow-sm` / `px-md` / `gap-sm` 等 40+ 类。

---

## 七、验收标准

### 7.1 每个页面
- [ ] 4 种状态全部覆盖（加载/空/错误/正常）
- [ ] API 调用返回 200，数据正确渲染
- [ ] Token 过期自动刷新
- [ ] 下拉刷新（列表页）
- [ ] 小程序和 H5 功能对齐

### 7.2 Sprint 交付
- [ ] `npm run build`（H5） 无报错
- [ ] 微信开发者工具编译通过（小程序）
- [ ] `go test ./...` 后端测试不退化

---

## 八、当前文件结构（2026-07-26 更新）

```
d:/w-yao/
├── cmd/api/main.go           ← Go 后端 (212 API) ✅
├── internal/                 ← 业务逻辑 ✅
├── frontend/                 ← H5 前端 (Vue3+Vant4+ElementPlus) 🟡 框架
│   └── src/
│       ├── main.js           ← Element Plus + Vant 全局注册
│       ├── hooks/
│       │   └── useListRequest.js    🆕 通用列表 Hook
│       ├── api/admin/
│       │   ├── enterprise.js        🆕 企业审核 API
│       │   ├── demand.js            🆕 需求管理 API
│       │   ├── user.js              🆕 用户管理 API
│       │   ├── review.js            🆕 评价管理 API
│       │   └── application.js       🆕 申请单 API
│       ├── utils/
│       │   └── http.js         ✅ Token拦截器 + 分页不解包
│       ├── views/
│       │   ├── home/           ✅ 首页
│       │   ├── services/       ✅ 6大分类
│       │   ├── demand/         ✅ 详情+发布
│       │   ├── messages/       ✅ 消息中心
│       │   ├── mine/           ✅ 个人中心
│       │   ├── login/          ✅ 登录
│       │   └── admin/          🟢 管理后台 (Element Plus)
│       │       ├── Dashboard.vue          ✅ 数据看板
│       │       ├── AdminLayout.vue        ✅ 布局
│       │       ├── enterprises/           🆕 el-table 企业审核
│       │       ├── demands/               🆕 el-table 需求管理
│       │       ├── users/                 🆕 el-table 用户管理
│       │       ├── cases/                 🆕 el-table 案例管理
│       │       ├── reviews/               🆕 el-table 评价管理
│       │       ├── orders/                🆕 el-table 订单管理
│       │       └── competition/           🆕 el-table 赛事管理
│       └── stores/            🟡 7个 Pinia store
├── miniprogram/              ← 小程序 (uni-app+Vant Weapp) 🟡 框架
│   ├── App.vue               🆕 CSS 变量 + 40+ 工具类
│   ├── components/
│   │   ├── Layout.vue        🆕 布局(使用CSS变量)
│   │   ├── StateView.vue     🆕 统一状态组件(loading/error/empty)
│   │   ├── TabBar.vue        ✅ 自定义 TabBar
│   │   └── HomeFloatButton.vue ✅ 首页浮动按钮
│   └── pages/
│       ├── home/             ✅
│       ├── services/         ✅
│       ├── applications/     ✅
│       ├── messages/         ✅
│       └── mine/             ✅
├── docs/
│   ├── PRD-无人机产业综合服务平台-前端开发.md  ← 本文档
│   ├── PRD-详细分工-页面级.md
│   ├── 需求文档/小程序开发规格.md   ← 42页规格
│   ├── 业务系统/ (7份)              ← 子系统详情
│   └── 接口文档/API契约.md          ← 212端点清单
└── migrations/               ← 66张表结构 ✅
```
