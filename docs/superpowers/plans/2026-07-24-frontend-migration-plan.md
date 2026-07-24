# 前端迁移实施计划 — microAPP-main → drone-platform

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将 microAPP-main H5 前端完整迁移适配为无人机产业综合服务平台前端

**Architecture:** Vue 3 + Vite 5 + Vant 4 + Pinia + Vue Router 4 H5 应用，4 Tab + 管理后台布局，对接 Go 后端 :8080

**Tech Stack:** Vue 3.4, Vite 5.2, Vant 4.8, Pinia 2.1, Vue Router 4.3, Axios 1.6, ECharts 6

## Global Constraints

- 不新增 npm 依赖，复用现有 package.json
- 不修改 Go 后端代码
- 主色 `#1565C0`，与小程序设计令牌一致
- 所有 API 优先用 `/api/v1/*`，`/api/*` 仅作回退
- 4 Tab: 首页 / 业务大厅 / 消息 / 我的
- 删除所有医疗模块代码（零残留）

---

### Task 1: 拷贝前端项目到根目录

**Files:**
- Create: `frontend/` (全量拷贝)

- [ ] **Step 1: 拷贝整个 H5 目录**

```bash
cp -r d:/w-yao/microAPP-main/frontend/h5 d:/w-yao/frontend
```

- [ ] **Step 2: 修改 package.json 项目名**

File: `frontend/package.json`

将 `"name": "lowaltitude-service-platform"` 改为 `"name": "drone-platform-frontend"`

- [ ] **Step 3: 安装依赖并验证启动**

```bash
cd d:/w-yao/frontend && D:/Node/npm.cmd install
```

- [ ] **Step 4: 提交**

```bash
git add frontend/ && git commit -m "chore: copy microAPP-main H5 frontend to frontend/"
```

---

### Task 2: 修改 API 指向 Go 后端

**Files:**
- Modify: `frontend/.env.development`
- Modify: `frontend/.env.production`
- Modify: `frontend/vite.config.js`

- [ ] **Step 1: 修改开发环境 API 地址**

File: `frontend/.env.development`

```
VITE_API_TARGET=http://localhost:8080
```

- [ ] **Step 2: 修改生产环境 API 地址**

File: `frontend/.env.production`

```
VITE_API_TARGET=
```

- [ ] **Step 3: 修改 Vite proxy 配置**

File: `frontend/vite.config.js`

将 proxy target 从 `env.VITE_API_TARGET || 'http://localhost:3000'` 改为：
```js
server: {
  port: 5173,
  proxy: {
    '/api': { target: env.VITE_API_TARGET || 'http://localhost:8080', changeOrigin: true },
    '/uploads': { target: env.VITE_API_TARGET || 'http://localhost:8080', changeOrigin: true },
  }
}
```

- [ ] **Step 4: 修改 http.js baseURL**

File: `frontend/src/utils/http.js`

将 `baseURL: 'http://localhost:3000'` 改为 `baseURL: ''`（空字符串，走 Vite proxy）

- [ ] **Step 5: 提交**

```bash
git add frontend/.env.development frontend/.env.production frontend/vite.config.js frontend/src/utils/http.js
git commit -m "chore: point frontend API to Go backend :8080"
```

---

### Task 3: 剥离医疗模块

**Files:**
- Delete: `frontend/src/views/medical/` (9 files)
- Delete: `frontend/src/views/admin/medical/` (3 files)
- Delete: `frontend/src/stores/medical.js`
- Modify: `frontend/src/router/index.js`

- [ ] **Step 1: 删除医疗页面目录**

```bash
rm -rf d:/w-yao/frontend/src/views/medical
```

- [ ] **Step 2: 删除医疗管理后台页面**

```bash
rm -rf d:/w-yao/frontend/src/views/admin/medical
```

- [ ] **Step 3: 删除医疗 Store**

```bash
rm d:/w-yao/frontend/src/stores/medical.js
```

- [ ] **Step 4: 删除游戏页面（非业务功能）**

```bash
rm -rf d:/w-yao/frontend/src/views/games d:/w-yao/frontend/src/fames
```

- [ ] **Step 5: 从路由中移除医疗和游戏路由**

File: `frontend/src/router/index.js`

移除以下路由条目：
```js
// REMOVE these:
{ path: '/medical/...', ... }  // 所有 /medical/* 路由
{ path: '/games', ... }
{ path: '/games/play', ... }
// And their imports:
import Certification from '@/views/medical/Certification.vue'
// ... all medical imports
```

同时移除路由守卫中的 `requiresMedicalAuth` 相关逻辑。

- [ ] **Step 6: 删除 admin/orders（合并到需求管理）**

```bash
rm -rf d:/w-yao/frontend/src/views/admin/orders
```

- [ ] **Step 7: 验证编译**

```bash
cd d:/w-yao/frontend && D:/Node/npm.cmd run build 2>&1 | tail -5
```
期望: `Build complete` 无报错

- [ ] **Step 8: 提交**

```bash
git add -A frontend/ && git commit -m "chore: strip medical + games modules from frontend"
```

---

### Task 4: 品牌和主题替换

**Files:**
- Modify: `frontend/index.html`
- Modify: `frontend/src/styles/global.css`
- Modify: `frontend/src/views/layout/Index.vue`
- Modify: `frontend/src/views/admin/AdminLayout.vue`

- [ ] **Step 1: 修改 HTML 标题和 meta**

File: `frontend/index.html`

```html
<title>无人机产业综合服务平台</title>
<meta name="description" content="重庆无人机产业协会 — 产业综合服务平台">
```

- [ ] **Step 2: 修改全局主题色**

File: `frontend/src/styles/global.css`

将 `--primary-color: #1d1d1f` 改为：
```css
:root {
  --primary-color: #1565C0;
  --primary-light: #1E88E5;
  --primary-dark: #0D47A1;
  --accent-color: #0071e3;
  /* keep other vars */
}
```

- [ ] **Step 3: 修改 Layout TabBar 文字**

File: `frontend/src/views/layout/Index.vue`

Tab 标签改为：
```js
const tabs = [
  { path: '/home', label: '首页', icon: 'home-o' },
  { path: '/services', label: '业务大厅', icon: 'apps-o' },
  { path: '/messages', label: '消息', icon: 'chat-o' },
  { path: '/mine', label: '我的', icon: 'user-o' },
]
```

- [ ] **Step 4: 修改管理后台标题**

File: `frontend/src/views/admin/AdminLayout.vue`

将标题/logo 改为 `无人机产业协会 · 管理后台`

- [ ] **Step 5: 提交**

```bash
git add frontend/index.html frontend/src/styles/global.css frontend/src/views/layout/Index.vue frontend/src/views/admin/AdminLayout.vue
git commit -m "style: rebrand — 温州低空→重庆无人机产业协会, theme #1565C0"
```

---

### Task 5: 改造首页为产业首页

**Files:**
- Modify: `frontend/src/views/home/Index.vue`
- Create: `frontend/src/stores/home.js`

- [ ] **Step 1: 创建首页 Store**

File: `frontend/src/stores/home.js`

```js
import { defineStore } from 'pinia'
import { ref } from 'vue'
import http from '@/utils/http'

export const useHomeStore = defineStore('home', () => {
  const banners = ref([])
  const notices = ref([])
  const quickEntries = ref([])
  const demandFeed = ref([])
  const loading = ref(false)

  const systemEntries = [
    { id: 1, name: '会员资源', icon: 'user-o', color: '#1565C0' },
    { id: 2, name: '供需对接', icon: 'exchange-o', color: '#2E7D32' },
    { id: 3, name: '产学研', icon: 'certificate-o', color: '#E65100' },
    { id: 4, name: '合规政策', icon: 'shield-o', color: '#6A1B9A' },
    { id: 5, name: '人才教育', icon: 'bookmark-o', color: '#C62828' },
    { id: 6, name: '活动品牌', icon: 'star-o', color: '#00838F' },
    { id: 7, name: '应急协同', icon: 'warning-o', color: '#D84315' },
  ]

  async function fetchHome() {
    loading.value = true
    try {
      const { data } = await http.get('/api/v1/home')
      if (data.success !== false) {
        banners.value = data.banners || []
        notices.value = data.notices || []
        quickEntries.value = data.quickEntries || []
      }
    } finally {
      loading.value = false
    }
  }

  async function fetchDemands(params = {}) {
    const { data } = await http.get('/api/v1/demands', { params })
    demandFeed.value = data.items || data.data || []
    return data
  }

  return { banners, notices, quickEntries, demandFeed, loading, systemEntries, fetchHome, fetchDemands }
})
```

- [ ] **Step 2: 重写首页模板（服务图标区）**

File: `frontend/src/views/home/Index.vue`

在 template 中替换原服务图标网格为 7 大业务系统入口：

```html
<!-- 7 大业务系统入口 -->
<div class="system-grid">
  <div v-for="sys in homeStore.systemEntries" :key="sys.id"
       class="system-card" :style="{ borderTopColor: sys.color }"
       @click="onSystemTap(sys.id)">
    <van-icon :name="sys.icon" :color="sys.color" size="24" />
    <span>{{ sys.name }}</span>
  </div>
</div>

<!-- 需求大厅信息流 -->
<van-list v-model:loading="homeStore.loading"
          :finished="finished" @load="onLoadDemands">
  <van-card v-for="d in homeStore.demandFeed" :key="d.id"
            :title="d.title" :desc="d.description"
            :tag="d.biz_type" :price="formatPrice(d.budget_fen)"
            @click="goDemand(d.id)" />
</van-list>
```

- [ ] **Step 3: 添加 script setup 逻辑**

```js
import { useHomeStore } from '@/stores/home'
const homeStore = useHomeStore()

onMounted(() => { homeStore.fetchHome() })

function onSystemTap(id) {
  // Navigate to business system sub-page
  const routes = { 1:'/members', 2:'/demands', 3:'/innovation', 4:'/compliance',
                   5:'/training', 6:'/events', 7:'/emergency' }
  router.push(routes[id] || '/services')
}
```

- [ ] **Step 4: 提交**

```bash
git add frontend/src/stores/home.js frontend/src/views/home/Index.vue
git commit -m "feat: homepage — 7 business system entries + demand feed"
```

---

### Task 6: 改造业务大厅页

**Files:**
- Modify: `frontend/src/views/services/Index.vue`
- Create: `frontend/src/stores/demand.js`

- [ ] **Step 1: 创建需求 Store**

File: `frontend/src/stores/demand.js`

```js
import { defineStore } from 'pinia'
import { ref } from 'vue'
import http from '@/utils/http'

export const useDemandStore = defineStore('demand', () => {
  const list = ref([])
  const total = ref(0)
  const loading = ref(false)

  async function fetchDemands(params = {}) {
    loading.value = true
    try {
      const { data } = await http.get('/api/v1/demands', { params })
      list.value = data.items || data.data || []
      total.value = data.total || list.value.length
    } finally { loading.value = false }
  }

  async function fetchDetail(id) {
    const { data } = await http.get(`/api/v1/demands/${id}`)
    return data
  }

  async function createBid(demandId, payload) {
    const { data } = await http.post(`/api/v1/demands/${demandId}/applications`, payload)
    return data
  }

  return { list, total, loading, fetchDemands, fetchDetail, createBid }
})
```

- [ ] **Step 2: 重写服务页为业务分类**

File: `frontend/src/views/services/Index.vue`

替换原 14 项低空服务为 6 大业务分类：
- 产业供需对接（需求大厅/供应展示/竞标报价）
- 培训认证（CAAC/UTC/人社/飞手）
- 无人机交易（整机/维修/配件）
- 合同签约（模板/签章/作废）
- 保险金融（保单/年审/贷款）
- 应急资源协同（救援案例/资源调度）

- [ ] **Step 3: 提交**

```bash
git add frontend/src/stores/demand.js frontend/src/views/services/Index.vue
git commit -m "feat: services — business categories replacing low-altitude services"
```

---

### Task 7: 新建需求详情和发布页面

**Files:**
- Create: `frontend/src/views/demand/Detail.vue`
- Create: `frontend/src/views/demand/Publish.vue`
- Modify: `frontend/src/router/index.js`

- [ ] **Step 1: 创建需求详情页**

File: `frontend/src/views/demand/Detail.vue`

```vue
<template>
  <div class="page">
    <van-nav-bar title="需求详情" left-arrow @click-left="$router.back()" />
    <van-loading v-if="!detail" />
    <template v-else>
      <van-cell-group>
        <van-cell :title="detail.title" :label="detail.description" />
        <van-cell title="类型" :value="detail.biz_type" />
        <van-cell title="预算" :value="'¥' + (detail.budget_fen / 100).toFixed(2)" />
        <van-cell title="状态" :value="detail.status" />
      </van-cell-group>
      <div class="bids">
        <h3>竞标列表</h3>
        <van-cell v-for="bid in bids" :key="bid.id"
          :title="bid.bidder_name" :label="bid.proposal"
          :value="'¥' + (bid.amount_fen / 100).toFixed(2)" />
      </div>
      <van-button type="primary" block @click="showBidForm = true">参与竞标</van-button>
      <!-- bid form popup -->
    </template>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useDemandStore } from '@/stores/demand'
import http from '@/utils/http'

const route = useRoute()
const detail = ref(null)
const bids = ref([])

onMounted(async () => {
  const { data } = await http.get(`/api/v1/demands/${route.params.id}`)
  detail.value = data
  const bres = await http.get(`/api/v1/demands/${route.params.id}/applications`)
  bids.value = bres.data || []
})
</script>
```

- [ ] **Step 2: 创建需求发布页**

File: `frontend/src/views/demand/Publish.vue`

包含表单：标题/描述/业务类型(巡检/植保/农药/租赁/清洗/其他)/预算/城市。提交到 `POST /api/v1/demands`。

- [ ] **Step 3: 添加路由**

File: `frontend/src/router/index.js`

```js
{ path: '/demand/:id', component: () => import('@/views/demand/Detail.vue') },
{ path: '/demand/publish', component: () => import('@/views/demand/Publish.vue'), meta: { requiresAuth: true } },
```

- [ ] **Step 4: 提交**

```bash
git add frontend/src/views/demand/ frontend/src/router/index.js
git commit -m "feat: demand detail + publish pages"
```

---

### Task 8: 新建消息中心页

**Files:**
- Create: `frontend/src/views/messages/Index.vue`
- Create: `frontend/src/stores/message.js`
- Modify: `frontend/src/router/index.js`

- [ ] **Step 1: 创建消息 Store**

File: `frontend/src/stores/message.js`

```js
import { defineStore } from 'pinia'
import { ref } from 'vue'
import http from '@/utils/http'

export const useMessageStore = defineStore('message', () => {
  const list = ref([])
  const unread = ref(0)

  async function fetchMessages() {
    const { data } = await http.get('/api/v1/messages')
    list.value = data.items || data.data || []
  }

  async function fetchUnread() {
    const { data } = await http.get('/api/v1/messages/unread-count')
    unread.value = data.count || 0
  }

  async function markRead(id) {
    await http.post(`/api/v1/messages/${id}/read`)
    unread.value = Math.max(0, unread.value - 1)
  }

  return { list, unread, fetchMessages, fetchUnread, markRead }
})
```

- [ ] **Step 2: 创建消息中心页面**

File: `frontend/src/views/messages/Index.vue`

使用 `van-nav-bar` + `van-cell` 列表 + `van-badge` 未读红点。调用 `messageStore.fetchMessages()` onMounted。

- [ ] **Step 3: 提交**

```bash
git add frontend/src/views/messages/ frontend/src/stores/message.js frontend/src/router/index.js
git commit -m "feat: message center with unread badge"
```

---

### Task 9: 改造个人中心

**Files:**
- Modify: `frontend/src/views/mine/Index.vue`
- Create: `frontend/src/stores/enterprise.js`

- [ ] **Step 1: 创建企业 Store**

File: `frontend/src/stores/enterprise.js`

```js
import { defineStore } from 'pinia'
import { ref } from 'vue'
import http from '@/utils/http'

export const useEnterpriseStore = defineStore('enterprise', () => {
  const myEnterprises = ref([])
  const loading = ref(false)

  async function fetchMy() {
    loading.value = true
    try {
      const { data } = await http.get('/api/v1/enterprises')
      myEnterprises.value = data.items || data.data || data
    } finally { loading.value = false }
  }

  async function apply(form) {
    const { data } = await http.post('/api/v1/enterprises', form)
    return data
  }

  return { myEnterprises, loading, fetchMy, apply }
})
```

- [ ] **Step 2: 重写个人中心菜单**

File: `frontend/src/views/mine/Index.vue`

替换原有菜单为：
- 企业入驻（审核状态）
- 我的需求（已发布/竞标中/已完成）
- 我的证书（CAAC/UTC/人社）
- 我的合同
- 钱包余额
- 设置

- [ ] **Step 3: 提交**

```bash
git add frontend/src/views/mine/Index.vue frontend/src/stores/enterprise.js
git commit -m "feat: profile — enterprise apply, my demands, certificates, wallet"
```

---

### Task 10: 改造管理后台

**Files:**
- Modify: `frontend/src/views/admin/Dashboard.vue`
- Modify: `frontend/src/views/admin/Index.vue`
- Modify: `frontend/src/views/admin/AdminSidebar.vue`

- [ ] **Step 1: 重写数据看板**

File: `frontend/src/views/admin/Dashboard.vue`

替换原医疗统计为无人机业务指标：
- MetricCard × 4: 需求总数 / 企业入驻数 / 培训认证数 / 成交率
- ECharts 折线图: 近30天需求发布趋势
- ECharts 饼图: 需求类型分布（巡检/植保/农药/租赁/清洗）
- ECharts 饼图: 企业审核状态分布

```js
// Fetch from Go backend
const { data: dashboard } = await http.get('/api/v1/admin/dashboard')
// dashboard = { totalDemands, totalEnterprises, totalCerts, completionRate, trends, typeDist, statusDist }
```

- [ ] **Step 2: 重写审核管理 Tabs**

File: `frontend/src/views/admin/Index.vue`

替换 Tab 为：
```html
<van-tabs>
  <van-tab title="企业审核"> <EnterpriseReview /> </van-tab>
  <van-tab title="需求审核"> <DemandReview /> </van-tab>
  <van-tab title="评价审核"> <ReviewManage /> </van-tab>
  <van-tab title="举报处理"> <ReportManage /> </van-tab>
  <van-tab title="用户管理"> <UserManage /> </van-tab>
  <van-tab title="平台配置"> <ConfigManage /> </van-tab>
</van-tabs>
```

- [ ] **Step 3: 更新侧边栏菜单**

File: `frontend/src/views/admin/AdminSidebar.vue`

```js
const menuItems = [
  { path: '/admin', label: '数据看板', icon: 'chart-trending-o' },
  { path: '/admin/enterprises', label: '企业审核', icon: 'shop-o' },
  { path: '/admin/demands', label: '需求管理', icon: 'orders-o' },
  { path: '/admin/reviews', label: '评价管理', icon: 'star-o' },
  { path: '/admin/reports', label: '举报处理', icon: 'warning-o' },
  { path: '/admin/users', label: '用户管理', icon: 'manager-o' },
  { path: '/admin/config', label: '平台配置', icon: 'setting-o' },
]
```

- [ ] **Step 4: 提交**

```bash
git add frontend/src/views/admin/
git commit -m "feat: admin dashboard — drone metrics + review tabs"
```

---

### Task 11: 适配用户 Store 到 Go JWT

**Files:**
- Modify: `frontend/src/stores/user.js`

- [ ] **Step 1: 更新 Token 存储逻辑**

File: `frontend/src/stores/user.js`

Go 后端的 JWT response 格式：
```json
{ "accessToken": "header.payload.sig", "refreshToken": "random-string", "expiresIn": 900,
  "user": { "id": "...", "role": "...", "status": "..." } }
```

确保 store 正确提取 `accessToken` / `refreshToken`。Go Token 不使用 `token_type: "Bearer"` 前缀（http.js 拦截器自动加）。

- [ ] **Step 2: 微信登录适配**

登录流程：`POST /api/v1/auth/wechat/login` body: `{ code: "wx_login_code" }`
响应包含 `accessToken` + `refreshToken` + `user`。

在 store 的 `wechatLogin(code)` action 中调用此端点。

- [ ] **Step 3: 提交**

```bash
git add frontend/src/stores/user.js
git commit -m "fix: user store adapt to Go JWT + WeChat code2Session flow"
```

---

### Task 12: 端到端验证

- [ ] **Step 1: 启动 Go 后端**

```bash
cd d:/w-yao && AUTH_SECRET="dev-secret-key-at-least-32-bytes-long!!" ADMIN_DEV_MODE=true ./drone-api.exe &
```

- [ ] **Step 2: 验证健康检查**

```bash
curl -s http://localhost:8080/healthz
```
期望: `{"status":"ok","checks":{"server":"up","storage":"memory"}}`

- [ ] **Step 3: 验证首页 API**

```bash
curl -s http://localhost:8080/api/v1/home
```
期望: 200 + JSON

- [ ] **Step 4: 构建前端**

```bash
cd d:/w-yao/frontend && D:/Node/npm.cmd run build
```
期望: `Build complete` 无报错

- [ ] **Step 5: 验证 Go 后端测试不受影响**

```bash
cd d:/w-yao && go build ./... && go vet ./... && go test ./internal/... -count=1
```
期望: 全部 PASS

- [ ] **Step 6: 提交最终状态**

```bash
git add -A && git commit -m "chore: finalize frontend migration — all 12 tasks complete"
```
