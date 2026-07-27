# PRD — 四人并行开发方案（零冲突 · 小程序 + 后台管理）

> 日期: 2026-07-27 (更新) | 团队: 4人 | GitHub 协作 | AI 辅助开发 | 后端 API 100% 就绪  
> 代码质量基线: P0/P1 已修复 — 零 JSON 错误忽略 + 零裸 error 返回 + 100% 测试通过  
> 页面精简: 59 → **45 页**（砍 11 子页，核心功能不减）

---

## 一、零冲突核心原则

### 🚨 铁律三条

|   #   | 规则                                                       | 违反后果                 |
| :---: | -------------------------------------------------------- | -------------------- |
| **1** | **每人只能编辑自己分配的文件目录**，绝不动别人的文件                             | Merge Conflict，阻塞所有人 |
| **2** | **共享组件/布局冻结** — 如需改共享层，必须在群内同步后由 A 统一操作                  | 全局崩塌，所有页面受影响         |
| **3** | **先 Pull 再 Push** — 每次开始工作前 `git pull --rebase`，结束后立即 PR | 分叉累积，合并地狱            |
| **4** | **🚨 禁止擅自推送 GitHub** — 任何人 push 前必须经 A 书面确认。A 本人亦然，AI 辅助代码必须人工审核后再 push | 不可逆的远程代码污染 |

### ✅ 为什么可以零冲突？

```
✅ 后端 API 全部完成（212 条），小程序和后台只需调用，不需要改任何 Go 代码
✅ 小程序页面按业务系统拆分，每个系统内部文件相互独立
✅ 后台管理 100% 归 A 一人，BCD 只做小程序，零交叉
✅ CSS 变量系统（40+ token）已就绪，写页面只需引用变量
✅ Element Plus 主题已配置，后台页面统一组件库
✅ 共享组件已稳定，不需要修改
✅ API 契约文档完整，前后端早已解耦
```

---

## 二、项目现状盘点

### 2.1 全局状态

| 层          |    状态    | 说明                                                 |
| ---------- | :------: | -------------------------------------------------- |
| **Go 后端**  |  ✅ 100%  | 212 API、66 表、7 大业务系统、中间件链完整                        |
| **代码质量**   |   ✅ 达标   | P0/P1 已清零: 26 处 JSON 错误忽略 → 全修复, 17 处裸 error → 全包装 |
| **测试覆盖**   | 🟡 45.8% | 100% 测试通过(92 HTTP case), httpapi 24.9% 需 Sprint 加强 |
| **微信小程序**  |  🟡 60%  | **45 页**（从 59 精简，砍 11 子页+合并 3 页），核心框架页可用 |
| **后台管理系统** |  🟡 35%  | Vue 3 + Element Plus 框架就绪，8/27 模块已完成               |
| **原型设计**   |  🟡 30%  | 首页(贴吧式)、商家页已完成，其余待设计                               |

### 2.2 后台管理系统现状

```
技术栈: Vue 3 + Vite + Element Plus + Pinia + Vue Router + ECharts + Axios
路由:   /admin/*  → AdminLayout（侧边栏 9 项 + 顶栏 + 内容区）
架构:   27 独立模块 → 合并为 9 大模块 (el-tabs 子模块切换)

✅ 9 模块全部就绪:
  📊 数据看板 — Dashboard
  🏢 会员管理 — 用户 + 企业 + 商家 + 专家 (4 tab)
  📦 交易管理 — 需求 + 订单 + 评价 (3 tab)
  📋 内容审核 — 案例 + 合规 (2 tab)
  🎓 人才教育 — 培训 + 证书 + 赛事 + 职位 + 院校 + 研学 (6 tab)
  🔬 产学研   — 成果 + 难题 + 课题 + 测试 + 转化 (5 tab)
  📣 运营推广 — 活动 + 品牌 + 展会 + 报告 (4 tab)
  🚨 应急协同 — 应急资源 + 调度 (2 tab)
  ⚙️ 系统设置 — 配置 + 消息 (2 tab)
```


### 2.3 页面精简策略

```
59 页 → 45 页（砍 11 子页 + 合并 3 页，核心功能不减）

砍掉的子页（功能合并到父页或弹窗）:
  match/recommend      → 合并到首页推荐流
  cases/detail         → 案例列表弹窗展示
  achievements/detail  → 成果列表弹窗展示
  applications/submit  → 合并到企业注册(一步完成)
  training/certificates → 合并到"我的"(子tab)
  study/detail         → H5 WebView 内嵌
  transformations/track → 合并到成果详情
  emergency/depts      → 合并到应急资源页(sub-tab)
  exhibitions/booth    → 活动详情内嵌表单
  portfolios/list      → 合并到企业详情
  demands/mine         → 合并到"我的"→我的需求
```

### 2.4 后台共享组件（已就绪）

| 组件                 | 路径                         | 用途                |
| ------------------ | -------------------------- | ----------------- |
| `AdminLayout.vue`  | `views/admin/`             | 侧边栏 + 顶栏 + 内容区布局  |
| `AdminSidebar.vue` | `views/admin/components/`  | 可折叠侧边栏菜单          |
| `AdminHeader.vue`  | `views/admin/components/`  | 顶栏（标题 + 角色 + 用户）  |
| `DataToolbar.vue`  | `views/admin/components/`  | 搜索 + 筛选 + 批量操作工具栏 |
| `MetricCard.vue`   | `views/admin/components/`  | 统计卡片（值 + 标签 + 趋势） |
| `useAuth.js`       | `views/admin/composables/` | 认证/角色 composable  |
| `useMedia.js`      | `views/admin/composables/` | 媒体上传 composable   |

---

## 三、四人分工 — A 全栈（小程序+后台），BCD 纯小程序

### 📁 文件所有权地图

```
              ┌─────────────────────────────────┐
              │          A（组长 · 全栈）         │
              │  小程序 8 核心页 + 全部后台(27模块) │
              │  基础设施 + Git 管理              │
              └─────────────────────────────────┘
                              │
        ┌─────────────────────┼─────────────────────┐
        │                     │                     │
┌───────┴───────┐   ┌─────────┴────────┐   ┌───────┴───────┐
│  B            │   │  C              │   │  D            │
│  纯小程序     │   │  纯小程序       │   │  纯小程序     │
│  会员+合规    │   │  人才+应急      │   │  产学研+活动  │
│  12 页        │   │  12 页          │   │  10 页        │
└───────────────┘   └─────────────────┘   └───────────────┘
```

> **小程序总计: 45 页** (8 + 12 + 12 + 10 + 3 公共页). 公共页(login/register/webview)由 A 维护。

---

### 成员 A（组长）— 小程序核心页 + 全部后台管理 + 基础设施

#### 小程序页面（8 页）

| 页面      | 路径                          |   状态   |
| ------- | --------------------------- | :----: |
| 首页（贴吧式） | `pages/home/index.vue`      | 🆕 重设计 |
| 商家列表    | `pages/shops/index.vue`     |  🆕 新建 |
| 业务大厅    | `pages/services/index.vue`  |  设计升级  |
| 服务详情    | `pages/services/detail.vue` |  设计升级  |
| 服务申请    | `pages/services/apply.vue`  |  设计升级  |
| 个人中心    | `pages/mine/index.vue`      |  设计升级  |
| 搜索页     | `pages/search/index.vue`    |  功能补全  |
| 消息中心    | `pages/messages/index.vue`  |   维护   |

#### 后台管理 — A 全权负责（27 模块，8 已完成 + 19 待建）

| 分类      | 模块             |   状态  |
| ------- | -------------- | :---: |
| **核心**  | 仪表盘(Dashboard) | ✅ 已完成 |
| **核心**  | 服务配置           | ✅ 已完成 |
| **核心**  | 消息通知           | 🆕 新建 |
| **用户**  | 用户管理           | ✅ 已完成 |
| **会员**  | 企业管理           | ✅ 已完成 |
| **会员**  | 专家管理           | 🆕 新建 |
| **会员**  | 资源管理           | 🆕 新建 |
| **会员**  | 商家管理           | 🆕 新建 |
| **交易**  | 需求管理           | ✅ 已完成 |
| **交易**  | 订单管理           | ✅ 已完成 |
| **交易**  | 评价管理           | ✅ 已完成 |
| **内容**  | 案例管理           | ✅ 已完成 |
| **内容**  | 合规管理           | 🆕 新建 |
| **人才**  | 赛事管理           | ✅ 已完成 |
| **人才**  | 培训管理           | 🆕 新建 |
| **人才**  | 证书管理           | 🆕 新建 |
| **人才**  | 职位管理           | 🆕 新建 |
| **人才**  | 院校管理           | 🆕 新建 |
| **人才**  | 研学管理           | 🆕 新建 |
| **应急**  | 应急资源管理         | 🆕 新建 |
| **应急**  | 应急调度管理         | 🆕 新建 |
| **产学研** | 成果管理           | 🆕 新建 |
| **产学研** | 难题管理           | 🆕 新建 |
| **产学研** | 课题管理           | 🆕 新建 |
| **产学研** | 测试场地管理         | 🆕 新建 |
| **产学研** | 成果转化管理         | 🆕 新建 |
| **活动**  | 活动管理           | 🆕 新建 |
| **活动**  | 品牌管理           | 🆕 新建 |
| **活动**  | 展会/报告管理        | 🆕 新建 |

#### 共享层（🔒 仅 A 可改）

```
小程序共享层:
  components/          ← StateView / Layout / TabBar / HomeFloatButton
  utils/request.js     ← API 请求封装
  App.vue              ← CSS 变量系统 + 全局样式
  pages.json           ← uni-app 路由注册
  main.js              ← 入口文件

后台共享层:
  views/admin/AdminLayout.vue     ← 后台布局
  views/admin/components/         ← AdminSidebar / AdminHeader / DataToolbar / MetricCard
  views/admin/composables/        ← useAuth / useMedia
  router/index.js                 ← 路由注册（含 /admin/* 子路由）
  stores/                         ← Pinia stores
  styles/                         ← Element Plus 主题覆盖
  utils/http.js                   ← Axios 封装
```

---

### 成员 B — 纯小程序：会员资源 + 合规政策（12 页，砍 2 子页）

| 系统 | 页面    | 路径                               | API                            | 说明 |
| -- | ----- | -------------------------------- | ------------------------------ | -- |
| 会员 | 企业注册  | `pages/enterprise/register.vue`  | `POST /enterprises`            | 含项目申报(一步完成) |
| 会员 | 审核状态  | `pages/enterprise/status.vue`    | `GET /enterprises`             | |
| 会员 | 专家列表  | `pages/experts/list.vue`         | `GET /experts`                 | |
| 会员 | 专家详情  | `pages/experts/detail.vue`       | `GET /experts/{id}`            | |
| 会员 | 资源台账  | `pages/resources/list.vue`       | `GET /industry-resources`      | |
| 会员 | 资源详情  | `pages/resources/detail.vue`     | `GET /industry-resources/{id}` | |
| 会员 | 我的业务  | `pages/applications/index.vue`   | `GET /me`                      | |
| 合规 | 政策资讯  | `pages/compliance/news.vue`      | `GET /articles`                | |
| 合规 | 合规知识库 | `pages/compliance/knowledge.vue` | `GET /compliance-docs`         | |
| 合规 | 团体标准  | `pages/compliance/standards.vue` | `GET /compliance-standards`    | |
| 合规 | 案例列表  | `pages/cases/index.vue`          | `GET /cases`                   | 详情弹窗展示 |
| — | 供需推荐  | ~~`pages/match/recommend.vue`~~ | — | **砍: 合并到首页推荐流(A负责)** |

> ⚠️ B 不碰 `frontend/` 目录。后台管理接口已有 `/api/v1/admin/enterprises` 等，A 来写。

---

### 成员 C — 纯小程序：人才教育 + 应急协同（12 页，砍 2 子页）

| 系统 | 页面   | 路径                                | API                                          | 说明 |
| -- | ---- | --------------------------------- | -------------------------------------------- | -- |
| 人才 | 培训课程 | `pages/training/courses.vue`      | `GET /training-courses`                      | |
| 人才 | 课程报名 | `pages/training/enroll.vue`       | `POST /training-courses/{id}/pay-and-enroll` | |
| 人才 | 赛事列表 | `pages/competitions/list.vue`     | `GET /competitions`                          | |
| 人才 | 赛事报名 | `pages/competitions/register.vue` | `POST /competitions/{id}/register`           | |
| 人才 | 职位列表 | `pages/jobs/list.vue`             | `GET /jobs`                                  | |
| 人才 | 简历管理 | `pages/jobs/resume.vue`           | `POST /resumes`                              | |
| 人才 | 院校展示 | `pages/colleges/list.vue`         | `GET /colleges`                              | |
| 人才 | 研学列表 | `pages/study/index.vue`           | —                                            | 详情WebView内嵌 |
| 应急 | 应急资源 | `pages/emergency/resources.vue`   | `GET /emergency-resources`                   | 部门对接合并为sub-tab |
| 应急 | 调度记录 | `pages/emergency/dispatches.vue`  | `GET /emergency-dispatches`                  | |
| 应急 | 救援案例 | `pages/emergency/cases.vue`       | `GET /rescue-cases`                          | |
| — | 我的证书 | ~~`pages/training/certificates.vue`~~ | — | **砍: 合并到"我的"子tab(A负责)** |
| — | 部门对接 | ~~`pages/emergency/depts.vue`~~ | — | **砍: 合并到应急资源sub-tab** |

> ⚠️ C 不碰 `frontend/` 目录。后台管理接口已有 `/api/v1/admin/competition` 等，A 来写。

---

### 成员 D — 纯小程序：产学研 + 活动品牌（10 页，砍 3 子页）

| 系统  | 页面   | 路径                                | API                                               | 说明 |
| --- | ---- | --------------------------------- | ------------------------------------------------- | -- |
| 产学研 | 成果列表 | `pages/achievements/list.vue`     | `GET /achievements`                               | 详情弹窗 + 含转化追踪 |
| 产学研 | 研发难题 | `pages/challenges/list.vue`       | `GET /rd-challenges`                              | |
| 产学研 | 课题攻关 | `pages/projects/list.vue`         | `GET /research-projects`                          | |
| 产学研 | 测试预约 | `pages/testsites/book.vue`        | `GET /test-sites` + `POST /test-sites/{id}/book`  | |
| 活动  | 活动列表 | `pages/events/list.vue`           | `GET /events`                                     | |
| 活动  | 活动详情 | `pages/events/detail.vue`         | `GET /events/{id}` + `POST /events/{id}/register` | 含展位申请表单 |
| 活动  | 展会列表 | `pages/exhibitions/list.vue`      | `GET /exhibitions`                                | |
| 活动  | 行业报告 | `pages/reports/list.vue`          | `GET /industry-reports`                           | |
| 供需  | 需求大厅 | `pages/demands/list.vue`          | `GET /demands`                                    | |
| — | 品牌展示 | ~~`pages/portfolios/list.vue`~~ | — | **砍: 合并到企业详情(B负责)** |
| — | 展位申请 | ~~`pages/exhibitions/booth.vue`~~ | — | **砍: 合并在活动详情表单** |

> ⚠️ D 不碰 `frontend/` 目录。后台管理接口已有 `/api/v1/admin/achievements` 等，A 来写。

---

### 🔒 共享组件冻结清单（只有 A 可以修改）

#### 小程序共享层

| 文件                               | 说明                     | 修改规则               |
| -------------------------------- | ---------------------- | ------------------ |
| `App.vue`                        | CSS 变量系统 + 全局样式        | 🔒 冻结，必须群内同步后由 A 改 |
| `components/StateView.vue`       | loading/error/empty 三态 | 🔒 冻结              |
| `components/Layout.vue`          | 页面布局                   | 🔒 冻结              |
| `components/TabBar.vue`          | 底部导航栏                  | 🔒 冻结              |
| `components/HomeFloatButton.vue` | 首页浮动按钮                 | 🔒 冻结              |
| `utils/request.js`               | API 请求封装               | 🔒 冻结              |
| `pages.json`                     | uni-app 路由注册           | 🔒 需新页注册时群内@A      |
| `main.js`                        | 入口文件                   | 🔒 冻结              |

#### 后台管理共享层

| 文件                                        | 说明                | 修改规则          |
| ----------------------------------------- | ----------------- | ------------- |
| `views/admin/AdminLayout.vue`             | 后台整体布局            | 🔒 冻结         |
| `views/admin/components/AdminSidebar.vue` | 侧边栏菜单             | 🔒 冻结（A 统一维护） |
| `views/admin/components/AdminHeader.vue`  | 顶栏                | 🔒 冻结         |
| `views/admin/components/DataToolbar.vue`  | 工具栏（搜索/筛选/批量）     | 🔒 冻结         |
| `views/admin/components/MetricCard.vue`   | 统计卡片              | 🔒 冻结         |
| `views/admin/composables/useAuth.js`      | 认证/角色 composable  | 🔒 冻结         |
| `router/index.js`                         | 路由注册（含 /admin/*）  | 🔒 需加路由时群内@A  |
| `stores/`                                 | Pinia 全局状态        | 🔒 冻结         |
| `styles/`                                 | Element Plus 主题覆盖 | 🔒 冻结         |
| `utils/http.js`                           | Axios 封装          | 🔒 冻结         |

**如果需要共享组件不支持某个场景怎么办？**  
→ 在自己的目录里写局部组件（小程序: `pages/xxx/components/`，后台: `views/admin/xxx/components/`），不要改全局组件。

---

## 四、后台管理技术规范

### 4.1 技术栈

```
Vue 3 (Composition API + <script setup>)
Element Plus 2.14        ← UI 组件库
@element-plus/icons-vue  ← 图标库
ECharts 6 + vue-echarts  ← 图表
Pinia 2                  ← 状态管理
Vue Router 4             ← 路由
Axios 1.6                ← HTTP 请求
Vite 5                   ← 构建工具
```

### 4.2 后台页面模板骨架

```vue
<template>
  <div class="admin-page">
    
    <div class="page-header">
      <h2>XX管理</h2>
      <el-button type="primary" @click="handleAdd">新增</el-button>
    </div>

    
    <DataToolbar
      v-model="searchForm"
      :filters="filterConfig"
      @search="handleSearch"
      @reset="handleReset"
    />

    
    <el-table
      v-loading="loading"
      :data="tableData"
      border stripe
      style="width: 100%"
    >
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="name" label="名称" />
      <el-table-column prop="status" label="状态">
        <template #default="{ row }">
          <el-tag :type="statusMap[row.status]">{{ row.status }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200">
        <template #default="{ row }">
          <el-button size="small" @click="handleEdit(row)">编辑</el-button>
          <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    
    <el-pagination
      v-model:current-page="pagination.page"
      v-model:page-size="pagination.pageSize"
      :total="pagination.total"
      layout="total, prev, pager, next, jumper"
      @change="loadData"
    />
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import DataToolbar from '../components/DataToolbar.vue'
import axios from '@/utils/http'

const loading = ref(false)
const tableData = ref([])
const pagination = reactive({ page: 1, pageSize: 20, total: 0 })

const loadData = async () => {
  loading.value = true
  try {
    const res = await axios.get('/api/v1/admin/xxx', {
      params: { page: pagination.page, page_size: pagination.pageSize }
    })
    tableData.value = res.data.data || []
    pagination.total = res.data.total || 0
  } catch (e) {
    ElMessage.error('加载失败')
  } finally {
    loading.value = false
  }
}

onMounted(loadData)
</script>

<style scoped>
.admin-page { padding: 20px; }
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}
</style>
```

### 4.3 后台 API 规范

```
基础路径: /api/v1/admin/
认证方式: Authorization: Bearer <token>（由 useAuth.js 自动注入）

标准 CRUD:
  GET    /api/v1/admin/{resource}          → 列表（分页）
  GET    /api/v1/admin/{resource}/{id}     → 详情
  POST   /api/v1/admin/{resource}          → 创建
  PUT    /api/v1/admin/{resource}/{id}     → 更新
  DELETE /api/v1/admin/{resource}/{id}     → 删除

响应格式:
  列表: { data: [...], total: 100, page: 1, page_size: 20 }
  详情: { data: {...} }
  操作: { success: true, data: {...} }
```

### 4.4 后台侧边栏菜单结构（A 统一维护）

```
📊 仪表盘           /admin/dashboard
📋 用户管理         /admin/users
🏢 企业管理         /admin/enterprises
👨‍🔧 专家管理         /admin/experts         ← 🆕
📦 资源管理         /admin/resources        ← 🆕
🏪 商家管理         /admin/shops            ← 🆕
📝 需求管理         /admin/demands
📦 订单管理         /admin/orders
⭐ 评价管理         /admin/reviews
📋 案例管理         /admin/cases
📜 合规管理         /admin/compliance       ← 🆕
---
🏆 赛事管理         /admin/competition
📚 培训管理         /admin/training         ← 🆕
💼 职位管理         /admin/jobs             ← 🆕
🎓 院校管理         /admin/colleges          ← 🆕
📖 研学管理         /admin/study            ← 🆕
---
🔬 成果管理         /admin/achievements     ← 🆕
❓ 难题管理         /admin/challenges       ← 🆕
🔧 课题管理         /admin/projects         ← 🆕
📍 测试场地         /admin/testsites        ← 🆕
🔄 成果转化         /admin/transformations  ← 🆕
---
🎉 活动管理         /admin/events           ← 🆕
🏷️ 品牌管理         /admin/portfolios       ← 🆕
🎪 展会管理         /admin/exhibitions      ← 🆕
📊 报告管���         /admin/reports          ← 🆕
---
🚨 应急资源         /admin/emergency/resources ← 🆕
📡 应急调度         /admin/emergency/dispatches ← 🆕
---
⚙️ 服务配置         /admin/config
📬 消息通知         /admin/messages         ← 🆕
```

---

## 五、Git 工作流 — GitHub 远程协作

### 5.1 分支策略

```
main                       ← 生产就绪（受保护，只有 A 可 merge）
  └── develop              ← 集成开发分支（A 负责 merge PR）
        ├── feat/a-home     ← A 小程序首页 + 商家 + 共享层
        ├── feat/a-admin-1  ← A 后台：会员/内容/交易模块
        ├── feat/a-admin-2  ← A 后台：人才/应急/产学研/活动模块
        ├── feat/b-pages    ← B 小程序页面（会员+合规）
        ├── feat/c-pages    ← C 小程序页面（人才+应急）
        └── feat/d-pages    ← D 小程序页面（产学研+活动）
```

### 5.2 每天工作四步走

```
上午 9:00
  git checkout develop
  git pull --rebase origin develop        ← ① 拉最新代码

  按当天任务开分支:
  A: git checkout -b feat/a-home  或  feat/a-admin-1  或  feat/a-admin-2
  B: git checkout -b feat/b-pages
  C: git checkout -b feat/c-pages
  D: git checkout -b feat/d-pages

全天
  写代码 → git add → git commit -m "feat(模块): xxx"
  ⚠️ commit 只包含自己目录的文件！
  ⚠️ B/C/D 绝对不要动 frontend/ 目录！

下午 17:00
  git checkout develop                     ← ③ 切回 develop
  git pull --rebase origin develop         ← 再拉一次最新
  git checkout feat/x-xxx                  ← 切回自己分支
  git rebase develop                       ← 变基到最新 develop
  
  解决冲突后（如有）→ git push origin feat/x-xxx
  
  开 Pull Request (develop ← feat/x-xxx)   ← ④ 提 PR
  ⚠️ PR 标题: [小程序|后台] 模块名 - 改动简述
  ⚠️ PR 描述: 改了哪些文件 + 小程序截图/后台截图
```

### 5.3 PR 审阅规则

| 规则        | 说明                                          |
| --------- | ------------------------------------------- |
| **审阅人**   | 必须至少 1 人审阅通过才能 merge                        |
| **A（组长）** | 所有 PR 最终由 A 点击 merge                        |
| **审阅时间**  | PR 提交后 4 小时内必须有人审阅                          |
| **自审**    | 小程序: 微信开发者工具编译通过；后台: `npm run dev` 无报错      |
| **冲突处理**  | 如有冲突，PR 作者自己 rebase develop 解决，再 force push |

### 5.4 Commit 规范

```
feat(模块): 新增XXX功能
fix(模块): 修复XXX问题
style(模块): 样式调整
refactor(模块): 重构XXX

示例:
feat(商家管理): 商家列表+搜索+审核功能
feat(小程序-首页): 贴吧式任务大厅信息流
fix(专家管理): 修复专家列表分页错误
style(后台): 统一表格操作栏按钮样式
```

---

## 六、API 契约 — 即插即用

### 6.1 基础规范

```
小程序 API:
  基础路径: /api/v1/
  认证方式: Authorization: Bearer <token>
  分页格式: ?page=1&page_size=20

后台管理 API:
  基础路径: /api/v1/admin/
  认证方式: Authorization: Bearer <token>（自动注入）
  分页格式: ?page=1&page_size=20

响应格式:
  成功: {"data": {...}, "request_id": "req_xxx"}
  分页: {"data": [...], "total": 100, "page": 1, "page_size": 20, "request_id": "req_xxx"}
  失败: {"error": {"code": "FORBIDDEN", "message": "..."}, "request_id": "req_xxx"}
```

### 6.2 各模块 API 速查

| 模块                         |   成员  | 小程序 API                                                                                                                                           |
| -------------------------- | :---: | ------------------------------------------------------------------------------------------------------------------------------------------------- |
| 企业/专家/资源/案例/合规             |   B   | `/enterprises` `/experts` `/industry-resources` `/cases` `/articles` `/compliance-docs` `/compliance-standards` `/recommendations`                |
| 培训/赛事/职位/院校/研学/应急          |   C   | `/training-courses` `/competitions` `/jobs` `/colleges` `/emergency-resources` `/rescue-cases` `/emergency-depts` `/emergency-dispatches`         |
| 成果/难题/课题/测试/转化/活动/品牌/展会/报告 |   D   | `/achievements` `/rd-challenges` `/research-projects` `/test-sites` `/transformations` `/events` `/portfolios` `/exhibitions` `/industry-reports` |
| 首页/商家/服务/我的/搜索/消息/需求       |   A   | `/demands` `/shops` `/messages` `/services-config`                                                                                                |
| **全部后台管理**                 | **A** | `/api/v1/admin/*`（27 个模块，A 一个人写）                                                                                                                  |

> 详细端点清单：`docs/接口文档/API契约.md`（212 条完整列表）

> 详细端点清单：`docs/接口文档/API契约.md`（212 条完整列表）

---

## 七、AI 开发规范 — 4 人都在用 AI

### 7.1 B/C/D — 小程序 AI Prompt 模板（统一使用）

```
## 任务
为无人机产业综合服务平台开发 [页面名称] 小程序页面

## 技术栈
uni-app (Vue 3) + Vant Weapp 1.11

## 设计参考
[贴原型截图或描述]

## API
端点: [API路径]
响应格式: { data: {...}, request_id: "..." }

## 约束
1. 只改 pages/[模块]/ 目录下的文件
2. 使用 App.vue 中定义的 CSS 变量（--color-primary 等）
3. 调用 utils/request.js 发请求
4. 使用 components/StateView.vue 处理加载/空/错误状态
5. 不要修改 components/、utils/、App.vue、pages.json
```

### 7.2 A 专用 — 后���管理 AI Prompt 模板

```
## 任务
为无人机产业综合服务平台开发 [XX管理] 后台页面

## 技术栈
Vue 3 (Composition API) + Element Plus 2.14 + ECharts 6

## 页面结构
- 搜索/筛选工具栏（使用 DataToolbar 组件）
- el-table 数据表格（分页、排序）
- el-dialog 新增/编辑弹窗
- el-pagination 分页器

## API
端点: GET/POST/PUT/DELETE /api/v1/admin/{resource}
响应格式: 分页 { data: [...], total: N }

## 约束
1. 只改 views/admin/[模块]/ 目录下的文件
2. 使用 Element Plus 组件（el-table / el-form / el-dialog / el-tag / el-button）
3. 使用 @/utils/http (Axios) 发请求
4. 搜索工具栏复用 views/admin/components/DataToolbar.vue
5. 删除操作必须弹 el-message-box 二次确认
6. 不要修改 router/index.js、AdminLayout.vue、AdminSidebar.vue、components/、composables/、stores/、styles/
7. 表格必须带 v-loading 加载态
```

### 7.3 B/C/D — AI 开发红线

| ❌ 不要让 AI 做的事                                    | ✅ 应该让 AI 做的事                             |
| ----------------------------------------------- | ---------------------------------------- |
| 修改 `components/` 下的共享组件                         | 在自己的 `pages/xxx/` 下写局部组件                 |
| 修改 `App.vue` 的 CSS 变量                           | 引用已有 CSS 变量                              |
| 修改 `pages.json` 注册新路由                           | 路由注册统一@A 处理                              |
| 修改 `frontend/` 目录下任何文件                          | BCD 完全不碰后台管理                             |
| 生成虚构的 API 路径                                    | 只调用 `docs/接口文档/API契约.md` 中存在的端点          |
| 引入新的 npm 包                                      | 只使用 Vant Weapp 已有组件                      |
| 使用硬编码颜色值                                        | 必须使用 CSS 变量                              |
| **Go 代码忽略 JSON 序列化错误** (`_, _ := json.Marshal`) | 必须检查 err: `if err != nil { return ... }` |
| **Go 代码裸返回 error** (`return err` 无包装)           | 必须用 `fmt.Errorf("context: %w", err)` 包装  |

### 7.4 AI 生成代码检查清单

**B/C/D — 小程序页面：**

- [ ] 文件路径是否在自己分配的目录下（`pages/xxx/`）
- [ ] API 路径是否与契约一致
- [ ] 是否使用了 CSS 变量（无硬编码颜色）
- [ ] 是否包含了 4 种状态（loading/empty/error/data）
- [ ] 是否有下拉刷新（列表页）
- [ ] 微信开发者工具是否能编译通过

**A — 小程序 + 后台管理（全部自检）：**

- [ ] 文件路径是否在自己分配的目录下
- [ ] API 路径是否与契约一致
- [ ] 是否使用了 CSS 变量（无硬编码颜色）
- [ ] 是否包含了 4 种状态（loading/empty/error/data）
- [ ] 是否有下拉刷新（列表页）
- [ ] 微信开发者工具是否能编译通过

**后台管理页面：**

- [ ] 文件路径是否在 `views/admin/[模块]/` 下
- [ ] API 路径是否正确（`/api/v1/admin/` 前缀）
- [ ] 表格是否带 v-loading
- [ ] 删除是否有 el-message-box 二次确认
- [ ] **Go 后端 PR 前: 跑 `go vet ./...` + `go test ./...` + 确认无 `_, _ := json.Marshal` 模式**
- [ ] 表单是否有 el-form 校验
- [ ] 分页是否正确（page/pageSize/total 三字段）
- [ ] `npm run dev` 是否无报错

---

## 八、设计令牌

### 8.1 小程序 CSS 变量（`App.vue` 已定义）

| Token                    | 值         | 用途    |
| ------------------------ | --------- | ----- |
| `--color-primary`        | `#1989fa` | 主色（蓝） |
| `--color-success`        | `#34c759` | 成功（绿） |
| `--color-warning`        | `#ff9f0a` | 警告（橙） |
| `--color-danger`         | `#ff3b30` | 危险（红） |
| `--color-bg`             | `#f5f6f8` | 页面背景  |
| `--color-bg-card`        | `#ffffff` | 卡片背景  |
| `--color-text`           | `#1a1a1a` | 主文字   |
| `--color-text-secondary` | `#969799` | 辅助文字  |
| `--radius-sm`            | `8rpx`    | 小圆角   |
| `--radius-md`            | `16rpx`   | 中圆角   |
| `--radius-lg`            | `24rpx`   | 大圆角   |

### 8.2 后台管理 Element Plus 主题

```
主色: #409EFF (Element Plus 默认蓝)
成功: #67C23A
警告: #E6A23C
危险: #F56C6C
信息: #909399

表格: border stripe, 操作栏宽度 200px
弹窗: width="600px", destroy-on-close
分页: layout="total, prev, pager, next, jumper"
```

---

## 九、Sprint 计划 — 6 周交付

### 总览

|   成员  | 小程序      | 后台管理          |   总工作量  |
| :---: | -------- | ------------- | :-----: |
| **A** | 8 页（核心页） | **27 模块（全部）** |  🔴 最重  |
| **B** | 12 页     | —             | 🟢 纯小程序 |
| **C** | 12 页     | —             | 🟢 纯小程序 |
| **D** | 10 页     | —             | 🟢 纯小程序 |
| 公共页 | login/register/webview | — | A 维护 |
| **合计** | **45 页** | **27 模块** | **72 页总交付** |

### Sprint 0（本周）— 环境搭建 + 原型设计

| 成员    | 小程序                       | 后台管理（仅 A）                                            |
| ----- | ------------------------- | ---------------------------------------------------- |
| **A** | 首页贴吧式→代码；商家页→代码           | Dashboard 增强 + AdminSidebar 菜单补全 + 全部新路由注册（一次性 19 个） |
| **B** | Clone → 开发者工具跑通 → 企业注册页原型 | —                                                    |
| **C** | Clone → 开发者工具跑通 → 培训课程页原型 | —                                                    |
| **D** | Clone → 开发者工具跑通 → 成果库页原型  | —                                                    |

### Sprint 1（第 1 周）— 核心框架页 + 展示型列表

| 成员    | 小程序                       | 后台管理（仅 A）                       |
| ----- | ------------------------- | ------------------------------- |
| **A** | 首页(贴吧式)、业务大厅(升级)、我的页面(升级) | 商家管理 + 专家管理 + 资源管理 + 合规管理（4 模块） |
| **B** | 企业注册、审核状态、专家列表、政策资讯       | —                               |
| **C** | 培训课程、赛事列表、院校展示、应急资源       | —                               |
| **D** | 成果列表、研发难题、活动列表、品牌展示       | —                               |

### Sprint 2（第 2 周）— 详情页 + 表单页 + 更多后台

| 成员    | 小程序                  | 后台管理（仅 A）                              |
| ----- | -------------------- | -------------------------------------- |
| **A** | 商家页(新建)、搜索页(补全)、消息中心 | 培训管理 + 证书管理 + 职位管理 + 院校管理 + 研学管理（5 模块） |
| **B** | 专家详情、资源台账、合规知识库、案例列表 | —                                      |
| **C** | 课程报名、赛事报名、职位列表、简历管理  | —                                      |
| **D** | 成果详情、测试预约、展会列表、行业报告  | —                                      |

### Sprint 3（第 3 周）— 剩余页面 + 后台补全

| 成员    | 小程序                  | 后台管理（仅 A）                              |
| ----- | -------------------- | -------------------------------------- |
| **A** | 需求大厅(升级)、需求详情、竞标报价   | 应急资源 + 应急调度 + 成果管理 + 难题管理 + 课题管理（5 模块） |
| **B** | 资源详情、团体标准、项目申报、案例详情  | —                                      |
| **C** | 我的证书、研学详情、救援案例、调度/部门 | —                                      |
| **D** | 课题攻关、成果转化、活动详情、展位/推荐 | —                                      |

### Sprint 4（第 4 周）— 小程序收尾 + 后台最后一波

| 成员    | 小程序                     | 后台管理（仅 A）                                  |
| ----- | ----------------------- | ------------------------------------------ |
| **A** | 功能补全 + Layout/TabBar 调整 | 测试场地 + 成果转化 + 活动 + 品牌 + 展会/报告 + 消息通知（6 模块） |
| **B** | 剩余页面 + 联调 bugfix        | —                                          |
| **C** | 剩余页面 + 联调 bugfix        | —                                          |
| **D** | 剩余页面 + 联调 bugfix        | —                                          |

### Sprint 5-6（第 5-6 周）— 联调 + 测试 + 上线

|    周   | 小程序（全员）                        | 后台管理（A）                     |
| :----: | ------------------------------ | --------------------------- |
| **W5** | 全员联调 bugfix + UI 走查 + 微信审核材料准备 | 后台权限细化 + 数据校验 + 批量操作 + 主题统一 |
| **W6** | 微信审核提交 + 灰度发布                  | 后台部署 + 最终联调 + 管理员文档         |

---

## 十、验收标准

### 10.1 小程序 — 每个页面

- [ ] 4 种状态全部覆盖（加载 / 空数据 / 错误 / 正常）
- [ ] API 调用返回 200，数据正确渲染
- [ ] Token 过期自动跳登录
- [ ] 下拉刷新（列表页）
- [ ] 使用 CSS 变量（无硬编码颜色）
- [ ] 微信开发者工具编译通过，无 console.error

### 10.2 后台管理 — 每个模块

- [ ] 列表页: el-table + 分页 + v-loading + 搜索/筛选
- [ ] 新增: el-dialog + el-form + 校验 + 提交
- [ ] 编辑: 复用新增弹窗 + 数据回填
- [ ] 删除: el-message-box 二次确认 + 刷新列表
- [ ] 表格操作栏: 查看/编辑/删除按钮
- [ ] `npm run dev` 编译通过，浏览器无 console.error
- [ ] 权限控制: 仅 `platform_admin` 可见敏感操作

### 10.3 每个 Sprint

- [ ] 微信开发者工具编译通过
- [ ] `npm run dev` 后台编译通过
- [ ] PR Review 通过（至少 1 人 Approve）
- [ ] 文件目录无越界修改
- [ ] `go test ./...` 后端测试不退化
- [ ] **代码质量硬门槛**: `go vet ./...` 零告警 / 禁止 `_, _ := json.Marshal(...)` / 禁止裸 `return err`（必须 `fmt.Errorf` 包装）

### 10.4 最终交付

- [ ] 7 大业务系统全部小程序页面可用（**45 页**，从 59 精简，核心功能不减）
- [ ] 后台管理 27 个模块全部可用
- [ ] GitHub Actions CI 绿灯（build + vet + test + integration）
- [ ] 微信小程序审核通过
- [ ] 后台管理系统部署上线
- [ ] 后端测试覆盖率 ≥ 60%（当前 45.8% → Sprint 3 达 60%）

---

## 十一、沟通协议

### 11.1 日常节奏

| 时间    | 事项                                         |
| ----- | ------------------------------------------ |
| 09:00 | 各自 `git pull --rebase develop`，开始工作        |
| 09:10 | 群内报今天计划（BCD: 小程序 X 页；A: 小程序 Y 页 + 后台 Z 模块） |
| 17:00 | 提 PR，群内发 PR 链接                             |
| 17:30 | 互相 Review 代码                               |
| 18:00 | A 合并当天所有 PR 到 develop                      |
| — | **push 规则: 任何人(含A) push 前必须经 A 确认，AI 写的代码也必须人工审核后 push** |

### 11.2 应急情况处理

| 情况                 | 处理方式                                               |
| ------------------ | -------------------------------------------------- |
| **发现别人文件的 Bug**    | 群内@对方，不要自己改                                        |
| **需要新增共享组件**       | 群内讨论 → A 统一添加                                      |
| **需要注册新路由**        | 小程序 pages.json 群内@A 处理；后台 router/index.js 仅 A 自己处理 |
| **需要新增侧边栏菜单**      | 仅 A 自己处理（AdminSidebar.vue）                         |
| **Merge Conflict** | 冲突方自己 rebase develop 解决                            |
| **API 返回数据不符合预期**  | 检查 API 契约文档 → 确认后群内讨论                              |
| **设计稿有歧义**         | @A 确认原型 → A 更新 `prototypes/`                       |
| **需要新建页面目录**       | @A 创建目录骨架 + 注册路由                                   |

### 11.3 GitHub 仓库配置（A 执行）

```
1. Settings → Branches → Add rule
   Branch: main
   ✅ Require pull request reviews before merging (1 approval)
   ✅ Require status checks to pass before merging
   ✅ Require branches to be up to date before merging

2. Settings → Branches → Add rule
   Branch: develop
   ✅ Require pull request reviews before merging (1 approval)

3. Settings → Collaborators → 添加 3 人 (Write 权限)
```

---

## 十二、附录 — 快速上手清单

### B/C/D 第一天 Checklist — 小程序

```
☐ 1. git clone <repo-url> && cd w-yao
☐ 2. 安装微信开发者工具，导入 miniprogram/ 目录
☐ 3. 编译运行 — 确认模拟器能出首页
☐ 4. 阅读 docs/系统架构/架构总览.md（5分钟）
☐ 5. 阅读 docs/接口文档/API契约.md 自己负责的模块
☐ 6. 阅读 docs/开发规范/Code-Review-Checklist.md
☐ 7. 熟悉 App.vue 中的 CSS 变量
☐ 8. 熟悉 components/StateView.vue 用法
☐ 9. 跑通一个已有页面 → 编译 → 查看效果
☐ 10. 确认自己的文件分配范围（第三节）
☐ 11. ⚠️ 记住：绝对不要动 frontend/ 目录！
```

### A 第一天 Checklist — 小程序 + 后台管理

```
小程序:
☐ 1. 微信开发者工具编译通过，确认首页正常
☐ 2. 完成 prototypes/ 下的最新原型设计

后台管理:
☐ 1. cd frontend && npm install
☐ 2. npm run dev → 浏览器打开 http://localhost:5173/admin
☐ 3. 确认 Dashboard 正常显示（dev auto-login 自动获取 token）
☐ 4. 阅读 views/admin/AdminLayout.vue 理解布局结构
☐ 5. 阅读 views/admin/components/DataToolbar.vue 了解工具栏用法
☐ 6. 在 router/index.js 一次性注册 19 个新路由
☐ 7. 在 AdminSidebar.vue 添加 19 个新菜单项
☐ 8. 创建 views/admin/experts/、resources/ 等 19 个新目录
```

---

> **最后更新**: 2026-07-27 | **文档维护**: A（组长）| **下次评审**: Sprint 1 结束时
