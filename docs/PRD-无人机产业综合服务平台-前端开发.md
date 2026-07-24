# PRD — 无人机产业综合服务平台 前端开发

> 日期: 2026-07-24 | 团队: 4人 | 后端: 已就绪

---

## 一、项目背景

重庆市无人机产业协会委托，面向 **180+ 会员单位**，覆盖无人机全产业链的数字化服务平台。

### 当前状态

| 层 | 状态 | 说明 |
|------|:--:|------|
| **Go 后端** | ✅ 100% | 212 条 API、66 张表、7 大业务系统全部完成 |
| **H5 前端** | 🟡 框架 | 首页/业务大厅/消息/我的/登录 — 6大分类入口可用 |
| **小程序** | 🟡 框架 | 同 H5，Vant Weapp 组件就绪 |
| **管理后台** | ❌ 已删除 | 待重建 |
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

### 6.1 页面模板

每个页面对应一个 Vue SFC（小程序 uni-app .vue，H5 Vue3 .vue），结构统一：

```
pages/[module]/[page].vue
  ├── <template>  — Vant 组件（van-nav-bar + van-cell + van-loading 等）
  ├── <script>    — API 调用 / 数据状态 / 交互逻辑
  └── <style>     — 全局设计令牌引用
```

### 6.2 必须处理的 4 种状态

| 状态 | UI |
|------|------|
| 加载中 | `van-loading` 或骨架屏 |
| 空数据 | `van-empty` + 引导文案 |
| 错误 | `van-toast` 错误信息 |
| 正常 | 数据渲染 |

### 6.3 API 调用规范

```js
// 统一使用 http.js (H5) 或 utils/request.js (小程序)
const res = await http.get('/api/v1/demands', { params: { biz_type, page, page_size } })
// res.data → 已解包，直接使用
// res.data.items → 列表
// res.data.total → 总数
```

### 6.4 关键 API 分页参数

```
GET /api/v1/demands?biz_type=巡检&district=南岸区&sort=newest&page=1&page_size=20
GET /api/v1/experts?field=低空管控&page=1&page_size=20
GET /api/v1/achievements?field=无人机&page=1&page_size=20
```

### 6.5 设计令牌

| Token | 值 | 用途 |
|------|------|------|
| `--primary-color` | `#1565C0` | 主色 |
| `--bg-color` | `#FAFAFA` | 页面背景 |
| `--card-bg` | `#FFFFFF` | 卡片背景 |
| 标题 | 18px | 页面标题 |
| 正文 | 15px | 列表项标题 |
| 辅助 | 13px | 时间/标签 |
| 间距 | 8/12/16/24/32 | 四级统一 |

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

## 八、当前文件结构

```
d:/w-yao/
├── cmd/api/main.go           ← Go 后端 (212 API) ✅
├── internal/                 ← 业务逻辑 ✅
├── frontend/                 ← H5 前端 (Vue3+Vant4) 🟡 框架
│   └── src/views/
│       ├── home/             ✅ 首页
│       ├── services/         ✅ 6大分类
│       ├── demand/           ✅ 详情+发布
│       ├── messages/         ✅ 消息中心
│       ├── mine/             ✅ 个人中心
│       └── login/            ✅ 登录
├── miniprogram/              ← 小程序 (uni-app+Vant Weapp) 🟡 框架
│   └── pages/
│       ├── home/             ✅
│       ├── services/         ✅
│       ├── applications/     ✅
│       ├── messages/         ✅
│       └── mine/             ✅
├── docs/
│   ├── 需求文档/小程序开发规格.md   ← 42页规格
│   ├── 业务系统/ (7份)              ← 子系统详情
│   └── 接口文档/API契约.md          ← 212端点清单
└── migrations/               ← 66张表结构 ✅
```
