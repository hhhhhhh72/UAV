# 删除 H5 前台 + 两端脱离 Vant 实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** frontend 瘦身为纯 Admin 后台（删 H5 前台 17 路由 + 6 stores），两端移除 Vant 依赖（Admin toast → Element Plus、小程序自研 u- 组件库替换 Vant Weapp），视觉统一"扁平简约蓝"品牌化。

**Architecture:** 前端：新增 `utils/feedback.js` 封装 ElMessage/ElMessageBox，Admin 文件按函数名替换 import；登录页基于 `views/auth/Login.vue` 改造，`/login` 路由保留。小程序：`components/ui/` 下 15 个 u- 组件 + easycom 自动注册，页面 `van-x` → `u-x` 机械替换后品牌化微调；分 4 批独立提交。

**Tech Stack:** Vue 3 + Element Plus（frontend）、uni-app + Vue3 script setup（miniprogram）、Go 后端不动（登录接口已生产注册）

**Spec:** `docs/superpowers/specs/2026-08-03-decompose-vant-delete-h5-design.md`

## Global Constraints

- 品牌色 `#0A66C2`（深空蓝）、辅色 `#1DD4A8`（青绿），交互色仅限这两个色系
- 小程序不使用 emoji 图标，图标用 CSS 绘制或文字标签
- 组件属性名兼容 van- 常用项（type/size/label/value/placeholder/disabled 等）
- 重构不改变功能行为，仅替换实现层与统一视觉
- 每批独立 commit，可单独回滚
- 验收：`grep -r "from 'vant'" frontend/src` 零命中；`grep -r "van-" miniprogram/pages miniprogram/pages.json` 零命中；wxcomponents/ 删除

---

# 批 1：frontend — H5 删除 + Admin 去 Vant

### Task 1: 创建 feedback.js 统一反馈封装

**Files:**
- Create: `frontend/src/utils/feedback.js`

**Interfaces:**
- Produces: `showToast(msg)`, `showFailToast(msg)`, `showSuccessToast(msg)`, `showLoadingToast()`, `closeToast()`, `showConfirmDialog({title, message, ...})` — 与 vant 同名同参，调用方仅改 import

- [ ] **Step 1: 创建 feedback.js**

```js
import { ElMessage, ElMessageBox } from 'element-plus'

let loadingInstance = null

export function showToast(message) {
  ElMessage({ message, type: 'info' })
}

export function showFailToast(message) {
  ElMessage({ message, type: 'error' })
}

export function showSuccessToast(message) {
  ElMessage({ message, type: 'success' })
}

export function showLoadingToast(message = '加载中...') {
  loadingInstance = ElMessage({ message, type: 'info', duration: 0 })
}

export function closeToast() {
  if (loadingInstance) {
    loadingInstance.close()
    loadingInstance = null
  }
}

export function showConfirmDialog({ title = '提示', message = '', confirmButtonText = '确定', cancelButtonText = '取消' } = {}) {
  return ElMessageBox.confirm(message, title, {
    confirmButtonText,
    cancelButtonText,
    type: 'warning'
  })
}
```

- [ ] **Step 2: 验证 vite dev 无报错**

Run: `cd frontend && npx vite build`
Expected: 构建成功（此时尚无引用，仅验证文件可编译进模块图——若 tree-shake 未引用的模块不打包，跳过构建直接进下一步）

- [ ] **Step 3: Commit**

```bash
git add frontend/src/utils/feedback.js
git commit -m "feat(frontend): 新增 feedback.js 统一反馈封装（ElMessage/ElMessageBox）"
```

### Task 2: Admin 文件 vant import 替换为 feedback.js

**Files:**
- Modify: `frontend/src/views/admin/cases/CaseList.vue`
- Modify: `frontend/src/views/admin/competition/CompetitionList.vue`
- Modify: `frontend/src/views/admin/composables/useMedia.js`
- Modify: `frontend/src/views/admin/config/ImageCropper.vue`
- Modify: `frontend/src/views/admin/config/ServiceConfigList.vue`
- Modify: `frontend/src/views/admin/Dashboard.vue`
- Modify: `frontend/src/views/admin/demands/DemandList.vue`
- Modify: `frontend/src/views/admin/enterprises/EnterpriseList.vue`
- Modify: `frontend/src/views/admin/orders/OrderList.vue`
- Modify: `frontend/src/views/admin/reviews/ReviewList.vue`
- Modify: `frontend/src/router/index.js`（第 2 行 `import { showFailToast } from 'vant'`）

**Interfaces:**
- Consumes: Task 1 的 feedback.js 函数（同名）

- [ ] **Step 1: 替换全部 vant import 行**

对上述 11 个文件逐一执行（import 行替换为 feedback.js）：

```bash
cd frontend/src
# 每个文件：把形如
#   import { showFailToast, showSuccessToast, showLoadingToast, closeToast, showConfirmDialog } from 'vant'
#   import { showToast } from 'vant'
# 统一替换为
#   import { showToast, showFailToast, showSuccessToast, showLoadingToast, closeToast, showConfirmDialog } from '@/utils/feedback'
```

用 grep 确认全部命中后统一替换（一次性列出所有文件）：

```bash
grep -rl "from 'vant'" views/admin/ router/ | while read f; do
  sed -i "s|from 'vant'|from '@/utils/feedback'|" "$f"
done
grep -rn "from 'vant'" views/admin/ router/ || echo "OK: 无残留"
```

注意：若某文件 import 的 vant 函数 feedback.js 未导出（检查每个文件 import 的函数名），补全 feedback.js 导出。

- [ ] **Step 2: 验证 import 的函数名都在 feedback.js 中**

```bash
cd frontend
grep -rhoE "import \{[^}]+\} from '@/utils/feedback'" src/ | grep -oE "\b(show|close)\w+" | sort -u
```
Expected: 输出 ⊆ {showToast, showFailToast, showSuccessToast, showLoadingToast, closeToast, showConfirmDialog}

- [ ] **Step 3: 构建验证**

Run: `npx vite build`
Expected: 成功。若报 ElMessage 未导入等错误，修正 feedback.js。

- [ ] **Step 4: Commit**

```bash
git add -A frontend/src
git commit -m "refactor(frontend): Admin 内部 vant toast 替换为 Element Plus（feedback.js）"
```

### Task 3: 登录页改造为 Admin 风格

**Files:**
- Read: `frontend/src/views/auth/Login.vue`（作为风格基底）
- Replace: `frontend/src/views/login/Index.vue`（内容替换）
- Modify: `frontend/src/router/index.js`（/login 路由保持不动，确认指向 views/login/Index.vue）

**Interfaces:**
- Consumes: `@/utils/http` 的 axios、`@/utils/feedback` 的 showFailToast/showSuccessToast；后端 `POST /api/auth/login`、`POST /api/v1/admin/token`

- [ ] **Step 1: 阅读现有 auth/Login.vue 并抽取其风格**

Run: `cat frontend/src/views/auth/Login.vue`
Expected: 了解现有 Admin 登录页的结构与样式（品牌色、表单布局）

- [ ] **Step 2: 重写 views/login/Index.vue 为 Admin 登录页**

内容要求（基于 auth/Login.vue 风格，两个登录方式）：
1. 手机号 + 密码表单（品牌蓝主题、居中卡片布局）
2. 提交调 `POST /api/auth/login`（body: `{ phone, password }`），成功存 `accessToken/refreshToken/user`（authStorage.setTokens + localStorage.setItem('user', ...)），跳 `/admin`
3. dev 模式快捷入口：调 `POST /api/v1/admin/token`（body `{ role: 'platform_admin' }`），失败静默（生产禁用时忽略）
4. 错误密码提示 `showFailToast('账号或密码错误')`
5. 移除所有 vant import；如用到 toast 用 feedback.js

- [ ] **Step 3: 构建 + 手动验证**

Run: `cd frontend && npx vite build`
Expected: 成功。

手动验证（dev 后端在 8080 运行）：浏览器访问 `http://localhost:5173/login`，正确密码登录 → 跳 /admin；错误密码 → 提示"账号或密码错误"。

- [ ] **Step 4: Commit**

```bash
git add frontend/src/views/login/Index.vue
git commit -m "feat(frontend): 登录页改造为 Admin 风格（保留 /login 路由）"
```

### Task 4: 删除 H5 前台页面/stores/路由

**Files:**
- Delete: `frontend/src/views/home/`、`views/services/`、`views/messages/`、`views/applications/`、`views/mine/`、`views/register/`、`views/study/`、`views/cases/`、`views/reviews/`、`views/demand/`、`views/layout/`
- Delete: `frontend/src/stores/demand.js`、`enterprise.js`、`home.js`、`message.js`、`service.js`、`user.js`
- Modify: `frontend/src/router/index.js`（删除 H5 路由、`/` 重定向 `/admin`）
- Modify: `frontend/src/main.js`（移除 `import Vant from 'vant'` 及 `app.use(Vant)`）
- Modify: `frontend/package.json`（移除 `vant` 依赖）

**注意：`views/auth/` 保留；`views/admin/` 保留。**

- [ ] **Step 1: 删除页面与 stores 目录**

```bash
cd frontend/src
rm -rf views/home views/services views/messages views/applications views/mine views/register views/study views/cases views/reviews views/demand views/layout
rm -f stores/demand.js stores/enterprise.js stores/home.js stores/message.js stores/service.js stores/user.js
```

- [ ] **Step 2: 重写 router/index.js**

保留：`/login` 路由、`/admin` 全部子路由、守卫逻辑（beforeEach 中 H5 相关分支删除：微信 OAuth 回调、SSO authcode 分支保留与否视引用——SSO 分支依赖 /api/sso/login（dev-only），H5 删除后保留无意义，一并删除）。`/` 重定向 `/admin`。

```js
import { createRouter, createWebHistory } from 'vue-router'
import axios, { authStorage } from '@/utils/http'

const routes = [
  {
    path: '/',
    redirect: '/admin'
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/login/Index.vue'),
    meta: { title: '登录' }
  },
  // ... /admin 及其 35 个子路由原样保留（从原文件复制）
]
```

守卫逻辑中删除：wechat_auth 回调分支（162-179）、authcode/SSO 分支（181-205）。`requiresAuth` 分支与 `/admin` 分支保留。

- [ ] **Step 3: 清理 main.js 与 package.json**

main.js：删除 `import Vant from 'vant'` 与 `app.use(Vant)`。
package.json：`npm uninstall vant`（或手动从 dependencies 移除后 `npm install`）。

- [ ] **Step 4: 构建 + 残留检查**

```bash
cd frontend
npx vite build
grep -rn "from 'vant'" src/ || echo "OK: vant 引用清零"
grep -rn "stores/user\|stores/home\|stores/service\|stores/demand\|stores/enterprise\|stores/message" src/ || echo "OK: H5 stores 引用清零"
```
Expected: 构建成功；两个 OK。

- [ ] **Step 5: 手动验证 Admin 全流程**

浏览器：`/login` 登录 → `/admin` Dashboard 及各聚合页（members/trading/content/talent/innovation/promotion/emergency/settings）浏览无报错。

- [ ] **Step 6: Commit**

```bash
git add -A frontend
git commit -m "refactor(frontend): 删除 H5 前台（17 路由 + 6 stores），frontend 瘦身为纯 Admin，vant 依赖清零"
```

---

# 批 2：小程序组件库

### Task 5: 设计令牌扩展 + easycom 配置

**Files:**
- Modify: `miniprogram/App.vue`（page 选择器内 CSS 变量）
- Modify: `miniprogram/pages.json`（easycom）

- [ ] **Step 1: App.vue 补充令牌**

在 `page { }` 内补充（已有基础令牌不动）：

```css
/* 组件库扩展令牌 */
--ui-radius-card: 24rpx;      /* 卡片大圆角 */
--ui-radius-btn: 50rpx;       /* 按钮圆角 */
--ui-shadow-card: 0 4rpx 16rpx rgba(0,0,0,0.06);  /* 卡片轻阴影 */
--ui-color-accent-light: #E6FAF5;  /* 青绿浅底 */
--ui-color-disabled: #c8c9cc;
--ui-color-text-secondary: #969799;
--ui-font-size-lg: 34rpx;
--ui-font-size-md: 30rpx;
--ui-font-size-sm: 26rpx;
--ui-space-card: 24rpx;      /* 卡片内边距 */
```

- [ ] **Step 2: pages.json 配置 easycom**

在 pages.json 顶层加入：

```json
"easycom": {
  "autoscan": true,
  "custom": {
    "^u-(.*)": "@/components/ui/u-$1.vue"
  }
}
```

- [ ] **Step 3: Commit**

```bash
git add miniprogram/App.vue miniprogram/pages.json
git commit -m "feat(miniprogram): 设计令牌扩展 + easycom 组件注册配置"
```

### Task 6: u-icon（CSS 绘制图标）

**Files:**
- Create: `miniprogram/components/ui/u-icon.vue`

**Interfaces:**
- Produces: `<u-icon name="arrow|search|close|plus|check|success|back|location|star" size="16px" color="#969799" />`

- [ ] **Step 1: 创建组件**

```vue
<template>
  <view class="u-icon" :class="'u-icon--' + name" :style="{ width: size, height: size, borderColor: color }">
    <view v-if="name === 'arrow'" class="u-icon-arrow" :style="{ borderColor: color }" />
    <view v-else-if="name === 'close'" class="u-icon-close" :style="{ background: color }" />
    <view v-else-if="name === 'plus'" class="u-icon-plus" :style="{ background: color }" />
    <view v-else-if="name === 'check'" class="u-icon-check" :style="{ borderColor: color }" />
    <view v-else-if="name === 'search'" class="u-icon-search" :style="{ borderColor: color }" />
    <view v-else-if="name === 'back'" class="u-icon-arrow u-icon-arrow--left" :style="{ borderColor: color }" />
    <view v-else-if="name === 'location'" class="u-icon-location" :style="{ borderColor: color, background: color }" />
    <view v-else-if="name === 'success'" class="u-icon-success" :style="{ borderColor: color }"><text class="u-icon-success-mark">✓</text></view>
    <text v-else class="u-icon-char" :style="{ color, fontSize: size }">{{ name }}</text>
  </view>
</template>

<script setup>
defineProps({
  name: { type: String, default: '' },
  size: { type: String, default: '32rpx' },
  color: { type: String, default: '#969799' }
})
</script>

<style scoped>
.u-icon { position: relative; display: inline-flex; align-items: center; justify-content: center; flex-shrink: 0; }
.u-icon-arrow { width: 14rpx; height: 14rpx; border-right: 4rpx solid; border-top: 4rpx solid; transform: rotate(45deg); border-radius: 2rpx; }
.u-icon-arrow--left { transform: rotate(-135deg); }
.u-icon-close { width: 28rpx; height: 4rpx; border-radius: 4rpx; transform: rotate(45deg); }
.u-icon-close::after { content: ''; position: absolute; width: 28rpx; height: 4rpx; border-radius: 4rpx; background: inherit; transform: rotate(90deg); }
.u-icon-plus { width: 28rpx; height: 4rpx; border-radius: 4rpx; }
.u-icon-plus::after { content: ''; position: absolute; width: 28rpx; height: 4rpx; border-radius: 4rpx; background: inherit; transform: rotate(90deg); }
.u-icon-check { width: 24rpx; height: 14rpx; border-left: 4rpx solid; border-bottom: 4rpx solid; transform: rotate(-45deg); border-radius: 2rpx; }
.u-icon-search { width: 20rpx; height: 20rpx; border: 4rpx solid; border-radius: 50%; }
.u-icon-search::after { content: ''; position: absolute; right: -10rpx; bottom: -6rpx; width: 14rpx; height: 4rpx; border-radius: 4rpx; background: inherit; transform: rotate(45deg); }
.u-icon-location { width: 20rpx; height: 20rpx; border: 4rpx solid; border-radius: 50% 50% 50% 0; transform: rotate(-45deg); }
.u-icon-success { width: 28rpx; height: 28rpx; border: 4rpx solid; border-radius: 50%; display: flex; align-items: center; justify-content: center; }
.u-icon-success-mark { font-size: 18rpx; line-height: 1; }
.u-icon-char { line-height: 1; }
</style>
```

- [ ] **Step 2: 页面冒烟验证**

在任意已迁移页面中使用 `<u-icon name="arrow" />` 编译小程序，确认渲染为箭头。

- [ ] **Step 3: Commit**

```bash
git add miniprogram/components/ui/u-icon.vue
git commit -m "feat(miniprogram): u-icon CSS 绘制图标组件"
```

### Task 7: u-loading

**Files:**
- Create: `miniprogram/components/ui/u-loading.vue`

**Interfaces:**
- Produces: `<u-loading size="32rpx" color="#0A66C2" />`

- [ ] **Step 1: 创建组件**

```vue
<template>
  <view class="u-loading" :style="{ width: size, height: size, borderColor: color + '33', borderTopColor: color }" />
</template>

<script setup>
defineProps({
  size: { type: String, default: '32rpx' },
  color: { type: String, default: '#0A66C2' }
})
</script>

<style scoped>
.u-loading { border: 4rpx solid; border-radius: 50%; animation: u-loading-spin 0.8s linear infinite; }
@keyframes u-loading-spin { to { transform: rotate(360deg); } }
</style>
```

- [ ] **Step 2: Commit**

```bash
git add miniprogram/components/ui/u-loading.vue
git commit -m "feat(miniprogram): u-loading 加载指示组件"
```

### Task 8: u-button

**Files:**
- Create: `miniprogram/components/ui/u-button.vue`

**Interfaces:**
- Produces: `<u-button type="primary|default|danger|success" size="normal|small|large" block round disabled loading @click>`；`@click` 事件

- [ ] **Step 1: 创建组件**

```vue
<template>
  <button
    class="u-button"
    :class="[`u-button--${type}`, `u-button--${size}`, { 'u-button--block': block, 'u-button--round': round, 'u-button--disabled': disabled || loading }]"
    :disabled="disabled || loading"
    @click="onClick"
  >
    <u-loading v-if="loading" size="28rpx" color="#fff" />
    <text class="u-button-text" v-else><slot /></text>
  </button>
</template>

<script setup>
import { defineEmits } from 'vue'

defineProps({
  type: { type: String, default: 'primary' },
  size: { type: String, default: 'normal' },
  block: { type: Boolean, default: false },
  round: { type: Boolean, default: true },
  disabled: { type: Boolean, default: false },
  loading: { type: Boolean, default: false }
})
const emit = defineEmits(['click'])
function onClick(e) {
  emit('click', e)
}
</script>

<style scoped>
.u-button { display: inline-flex; align-items: center; justify-content: center; gap: 8rpx; padding: 0 32rpx; height: 72rpx; font-size: 30rpx; border: none; border-radius: 16rpx; box-sizing: border-box; }
.u-button--block { width: 100%; }
.u-button--round { border-radius: 50rpx; }
.u-button--large { height: 88rpx; font-size: 32rpx; }
.u-button--small { height: 56rpx; padding: 0 24rpx; font-size: 26rpx; }
.u-button--primary { background: var(--color-primary, #0A66C2); color: #fff; }
.u-button--default { background: var(--color-primary-light, #E8F2FC); color: var(--color-primary, #0A66C2); }
.u-button--danger { background: var(--color-danger, #ff3b30); color: #fff; }
.u-button--success { background: var(--color-success, #34c759); color: #fff; }
.u-button--disabled { opacity: 0.5; }
.u-button-text { line-height: 1; }
</style>
```

- [ ] **Step 2: Commit**

```bash
git add miniprogram/components/ui/u-button.vue
git commit -m "feat(miniprogram): u-button 按钮组件"
```

### Task 9: u-cell + u-cell-group

**Files:**
- Create: `miniprogram/components/ui/u-cell.vue`
- Create: `miniprogram/components/ui/u-cell-group.vue`

**Interfaces:**
- Produces: `<u-cell-group inset><u-cell title value label is-link is-clickable @click /></u-cell-group>`

- [ ] **Step 1: u-cell.vue**

```vue
<template>
  <view class="u-cell" :class="{ 'u-cell--clickable': isClickable || isLink }" @click="onClick">
    <view v-if="$slots.icon || icon" class="u-cell-icon"><slot name="icon"><u-icon v-if="icon" :name="icon" :size="iconSize" /></slot></view>
    <view class="u-cell-body">
      <view class="u-cell-title-row">
        <text class="u-cell-title" :style="{ color: titleColor }"><slot name="title">{{ title }}</slot></text>
        <u-tag v-if="tag" size="mini" :type="tagType" plain>{{ tag }}</u-tag>
      </view>
      <text v-if="label" class="u-cell-label">{{ label }}</text>
    </view>
    <view v-if="value || $slots.value" class="u-cell-value"><slot name="value">{{ value }}</slot></view>
    <u-icon v-if="isLink" name="arrow" size="26rpx" color="#c8c9cc" />
  </view>
</template>

<script setup>
import { defineEmits } from 'vue'

const props = defineProps({
  title: { type: String, default: '' },
  value: { type: String, default: '' },
  label: { type: String, default: '' },
  icon: { type: String, default: '' },
  iconSize: { type: String, default: '36rpx' },
  isLink: { type: Boolean, default: false },
  isClickable: { type: Boolean, default: false },
  titleColor: { type: String, default: '' },
  tag: { type: String, default: '' },
  tagType: { type: String, default: 'primary' }
})
const emit = defineEmits(['click'])
function onClick() {
  if (props.isClickable || props.isLink) emit('click')
}
</script>

<style scoped>
.u-cell { display: flex; align-items: center; padding: 24rpx var(--ui-space-card, 24rpx); background: #fff; gap: 16rpx; }
.u-cell--clickable { cursor: pointer; }
.u-cell-icon { flex-shrink: 0; }
.u-cell-body { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 6rpx; }
.u-cell-title-row { display: flex; align-items: center; gap: 12rpx; }
.u-cell-title { font-size: var(--ui-font-size-md, 30rpx); color: var(--color-text, #1a1a1a); }
.u-cell-label { font-size: var(--ui-font-size-sm, 26rpx); color: var(--ui-color-text-secondary, #969799); }
.u-cell-value { font-size: var(--ui-font-size-md, 30rpx); color: var(--ui-color-text-secondary, #969799); flex-shrink: 0; }
</style>
```

- [ ] **Step 2: u-cell-group.vue**

```vue
<template>
  <view class="u-cell-group" :class="{ 'u-cell-group--inset': inset }">
    <slot />
  </view>
</template>

<script setup>
defineProps({
  inset: { type: Boolean, default: false },
  border: { type: Boolean, default: true }
})
</script>

<style scoped>
.u-cell-group { background: #fff; }
.u-cell-group--inset { margin: var(--ui-space-card, 24rpx); border-radius: var(--ui-radius-card, 24rpx); overflow: hidden; box-shadow: var(--ui-shadow-card, 0 4rpx 16rpx rgba(0,0,0,0.06)); }
.u-cell-group .u-cell + .u-cell { border-top: 1rpx solid var(--color-divider, #ebedf0); }
</style>
```

- [ ] **Step 3: Commit**

```bash
git add miniprogram/components/ui/u-cell.vue miniprogram/components/ui/u-cell-group.vue
git commit -m "feat(miniprogram): u-cell/u-cell-group 列表组件"
```

### Task 10: u-field

**Files:**
- Create: `miniprogram/components/ui/u-field.vue`

**Interfaces:**
- Produces: `<u-field label="姓名" v-model="val" placeholder="请输入" type="text|number|textarea" />`；emit `update:modelValue`

- [ ] **Step 1: 创建组件**

```vue
<template>
  <view class="u-field" :class="{ 'u-field--textarea': type === 'textarea' }">
    <text v-if="label" class="u-field-label">{{ label }}</text>
    <input
      v-else-if="type !== 'textarea'"
      class="u-field-input"
      :type="type"
      :value="modelValue"
      :placeholder="placeholder"
      :placeholder-class="'u-field-ph'"
      :disabled="disabled"
      @input="onInput"
    />
    <textarea
      v-else
      class="u-field-input u-field-textarea"
      :value="modelValue"
      :placeholder="placeholder"
      :placeholder-class="'u-field-ph'"
      :disabled="disabled"
      :auto-height="autoHeight"
      @input="onInput"
    />
  </view>
</template>

<script setup>
import { defineEmits } from 'vue'

const props = defineProps({
  label: { type: String, default: '' },
  modelValue: { type: [String, Number], default: '' },
  type: { type: String, default: 'text' },
  placeholder: { type: String, default: '' },
  disabled: { type: Boolean, default: false },
  autoHeight: { type: Boolean, default: false }
})
const emit = defineEmits(['update:modelValue'])
function onInput(e) {
  emit('update:modelValue', e.detail.value)
}
</script>

<style scoped>
.u-field { display: flex; align-items: center; background: #fafafa; border-radius: 24rpx; padding: 20rpx 24rpx; gap: 16rpx; }
.u-field-label { font-size: var(--ui-font-size-md, 30rpx); color: var(--color-text, #1a1a1a); flex-shrink: 0; }
.u-field-input { flex: 1; font-size: var(--ui-font-size-md, 30rpx); color: var(--color-text, #1a1a1a); }
.u-field-textarea { min-height: 120rpx; }
.u-field-ph { color: var(--ui-color-text-placeholder, #c8c9cc); }
.u-field--textarea { align-items: flex-start; }
</style>
```

- [ ] **Step 2: Commit**

```bash
git add miniprogram/components/ui/u-field.vue
git commit -m "feat(miniprogram): u-field 输入组件"
```

### Task 11: u-empty

**Files:**
- Create: `miniprogram/components/ui/u-empty.vue`

**Interfaces:**
- Produces: `<u-empty description="暂无数据" />`（CSS 绘制插画 + 文案）

- [ ] **Step 1: 创建组件**

```vue
<template>
  <view class="u-empty">
    <view class="u-empty-art">
      <view class="u-empty-circle" />
      <view class="u-empty-line u-empty-line--1" />
      <view class="u-empty-line u-empty-line--2" />
    </view>
    <text v-if="description" class="u-empty-desc">{{ description }}</text>
    <view v-if="$slots.default" class="u-empty-action"><slot /></view>
  </view>
</template>

<script setup>
defineProps({
  description: { type: String, default: '暂无数据' },
  image: { type: String, default: '' }
})
</script>

<style scoped>
.u-empty { display: flex; flex-direction: column; align-items: center; justify-content: center; padding: 80rpx 40rpx; gap: 24rpx; }
.u-empty-art { position: relative; width: 160rpx; height: 120rpx; }
.u-empty-circle { position: absolute; top: 0; left: 40rpx; width: 80rpx; height: 80rpx; border-radius: 50%; background: var(--color-primary-light, #E8F2FC); }
.u-empty-line { position: absolute; height: 12rpx; border-radius: 6rpx; background: var(--color-primary-light, #E8F2FC); }
.u-empty-line--1 { bottom: 16rpx; left: 20rpx; width: 120rpx; }
.u-empty-line--2 { bottom: 0; left: 50rpx; width: 60rpx; }
.u-empty-desc { font-size: var(--ui-font-size-md, 30rpx); color: var(--ui-color-text-secondary, #969799); }
.u-empty-action { margin-top: 8rpx; }
</style>
```

- [ ] **Step 2: Commit**

```bash
git add miniprogram/components/ui/u-empty.vue
git commit -m "feat(miniprogram): u-empty 空状态组件"
```

### Task 12: u-tag

**Files:**
- Create: `miniprogram/components/ui/u-tag.vue`

**Interfaces:**
- Produces: `<u-tag type="primary|success|danger|warning|default" plain size="normal|mini" round>`

- [ ] **Step 1: 创建组件**

```vue
<template>
  <text class="u-tag" :class="[`u-tag--${type}`, { 'u-tag--plain': plain, 'u-tag--mini': size === 'mini', 'u-tag--round': round }]">
    <slot />
  </text>
</template>

<script setup>
defineProps({
  type: { type: String, default: 'default' },
  plain: { type: Boolean, default: false },
  size: { type: String, default: 'normal' },
  round: { type: Boolean, default: true }
})
</script>

<style scoped>
.u-tag { display: inline-flex; align-items: center; padding: 4rpx 16rpx; font-size: var(--ui-font-size-sm, 26rpx); border-radius: 8rpx; line-height: 1.6; }
.u-tag--round { border-radius: 999rpx; }
.u-tag--mini { padding: 2rpx 12rpx; font-size: 22rpx; }
.u-tag--primary { background: var(--color-primary, #0A66C2); color: #fff; }
.u-tag--primary.u-tag--plain { background: var(--color-primary-light, #E8F2FC); color: var(--color-primary, #0A66C2); }
.u-tag--success { background: var(--color-success, #34c759); color: #fff; }
.u-tag--success.u-tag--plain { background: #E6FAF5; color: var(--color-success, #34c759); }
.u-tag--danger { background: var(--color-danger, #ff3b30); color: #fff; }
.u-tag--danger.u-tag--plain { background: #FFECEB; color: var(--color-danger, #ff3b30); }
.u-tag--warning { background: var(--color-warning, #ff9f0a); color: #fff; }
.u-tag--warning.u-tag--plain { background: #FFF4E6; color: var(--color-warning, #ff9f0a); }
.u-tag--default { background: var(--color-primary-light, #E8F2FC); color: var(--color-primary, #0A66C2); }
</style>
```

- [ ] **Step 2: Commit**

```bash
git add miniprogram/components/ui/u-tag.vue
git commit -m "feat(miniprogram): u-tag 标签组件"
```

### Task 13: u-nav-bar

**Files:**
- Create: `miniprogram/components/ui/u-nav-bar.vue`

**Interfaces:**
- Produces: `<u-nav-bar title="标题" left-text="返回" right-text="保存" fixed @back @right />`

- [ ] **Step 1: 创建组件**

```vue
<template>
  <view class="u-nav-bar" :class="{ 'u-nav-bar--fixed': fixed }" :style="{ paddingTop: statusBarHeight + 'px' }">
    <view class="u-nav-bar-inner">
      <view class="u-nav-bar-side u-nav-bar-left" @click="onBack">
        <u-icon v-if="leftText === '返回' || showBack" name="back" size="36rpx" color="#1a1a1a" />
        <text v-if="leftText && leftText !== '返回'" class="u-nav-bar-text">{{ leftText }}</text>
      </view>
      <text class="u-nav-bar-title">{{ title }}</text>
      <view class="u-nav-bar-side u-nav-bar-right" @click="onRight">
        <text v-if="rightText" class="u-nav-bar-text">{{ rightText }}</text>
        <slot name="right" />
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { defineEmits } from 'vue'

defineProps({
  title: { type: String, default: '' },
  leftText: { type: String, default: '返回' },
  rightText: { type: String, default: '' },
  fixed: { type: Boolean, default: false },
  showBack: { type: Boolean, default: false }
})
const emit = defineEmits(['back', 'right'])
const statusBarHeight = ref(20)
onMounted(() => {
  // #ifdef MP-WEIXIN
  const sys = uni.getSystemInfoSync()
  statusBarHeight.value = sys.statusBarHeight || 20
  // #endif
})
function onBack() {
  emit('back')
}
function onRight() {
  emit('right')
}
</script>

<style scoped>
.u-nav-bar { background: #fff; }
.u-nav-bar--fixed { position: fixed; top: 0; left: 0; right: 0; z-index: 100; }
.u-nav-bar-inner { position: relative; display: flex; align-items: center; justify-content: center; height: 44px; }
.u-nav-bar-title { font-size: 32rpx; font-weight: 600; color: var(--color-text, #1a1a1a); }
.u-nav-bar-side { position: absolute; display: flex; align-items: center; padding: 0 24rpx; height: 100%; top: 0; }
.u-nav-bar-left { left: 0; gap: 8rpx; }
.u-nav-bar-right { right: 0; }
.u-nav-bar-text { font-size: 28rpx; color: var(--color-text, #1a1a1a); }
</style>
```

- [ ] **Step 2: Commit**

```bash
git add miniprogram/components/ui/u-nav-bar.vue
git commit -m "feat(miniprogram): u-nav-bar 顶部导航组件"
```

### Task 14: u-search

**Files:**
- Create: `miniprogram/components/ui/u-search.vue`

**Interfaces:**
- Produces: `<u-search v-model="kw" placeholder="搜索" @search @change />`

- [ ] **Step 1: 创建组件**

```vue
<template>
  <view class="u-search">
    <view class="u-search-box">
      <u-icon name="search" size="28rpx" color="#969799" />
      <input
        class="u-search-input"
        :value="modelValue"
        :placeholder="placeholder"
        :placeholder-class="'u-search-ph'"
        confirm-type="search"
        @input="onInput"
        @confirm="onConfirm"
      />
      <view v-if="modelValue" class="u-search-clear" @click="onClear"><u-icon name="close" size="24rpx" color="#c8c9cc" /></view>
    </view>
  </view>
</template>

<script setup>
import { defineEmits } from 'vue'

const props = defineProps({
  modelValue: { type: String, default: '' },
  placeholder: { type: String, default: '搜索' },
  disabled: { type: Boolean, default: false }
})
const emit = defineEmits(['update:modelValue', 'search', 'change'])
function onInput(e) {
  emit('update:modelValue', e.detail.value)
  emit('change', e.detail.value)
}
function onConfirm() {
  emit('search', props.modelValue)
}
function onClear() {
  emit('update:modelValue', '')
  emit('change', '')
}
</script>

<style scoped>
.u-search { padding: 16rpx 24rpx; background: #f5f6f8; }
.u-search-box { display: flex; align-items: center; gap: 12rpx; background: #fff; border-radius: 50rpx; padding: 12rpx 24rpx; }
.u-search-input { flex: 1; font-size: 28rpx; }
.u-search-ph { color: #c8c9cc; }
.u-search-clear { padding: 4rpx; }
</style>
```

- [ ] **Step 2: Commit**

```bash
git add miniprogram/components/ui/u-search.vue
git commit -m "feat(miniprogram): u-search 搜索组件"
```

### Task 15: u-popup

**Files:**
- Create: `miniprogram/components/ui/u-popup.vue`

**Interfaces:**
- Produces: `<u-popup :show="visible" position="bottom|center|top" round close-on-click-overlay @close />`（内部 slot）

- [ ] **Step 1: 创建组件**

```vue
<template>
  <view v-if="show" class="u-popup" @click="onOverlayClick">
    <view class="u-popup-mask" />
    <view class="u-popup-panel" :class="[`u-popup--${position}`, { 'u-popup--round': round }]" @click.stop>
      <view v-if="showClose" class="u-popup-close" @click="onClose"><u-icon name="close" size="28rpx" color="#969799" /></view>
      <slot />
    </view>
  </view>
</template>

<script setup>
import { defineEmits } from 'vue'

defineProps({
  show: { type: Boolean, default: false },
  position: { type: String, default: 'bottom' },
  round: { type: Boolean, default: false },
  closeOnClickOverlay: { type: Boolean, default: true },
  showClose: { type: Boolean, default: false }
})
const emit = defineEmits(['close', 'update:show'])
function onClose() {
  emit('close')
  emit('update:show', false)
}
function onOverlayClick() {
  if (props.closeOnClickOverlay) onClose()
}
</script>

<style scoped>
.u-popup { position: fixed; inset: 0; z-index: 1000; }
.u-popup-mask { position: absolute; inset: 0; background: rgba(0,0,0,0.5); }
.u-popup-panel { position: absolute; background: #fff; }
.u-popup--bottom { left: 0; right: 0; bottom: 0; }
.u-popup--top { left: 0; right: 0; top: 0; }
.u-popup--center { left: 50%; top: 50%; transform: translate(-50%, -50%); border-radius: 24rpx; min-width: 600rpx; }
.u-popup--round { border-radius: 24rpx 24rpx 0 0; }
.u-popup-close { position: absolute; top: 20rpx; right: 24rpx; z-index: 2; }
</style>
```

- [ ] **Step 2: Commit**

```bash
git add miniprogram/components/ui/u-popup.vue
git commit -m "feat(miniprogram): u-popup 弹层组件"
```

### Task 16: u-tabs + u-tab

**Files:**
- Create: `miniprogram/components/ui/u-tabs.vue`
- Create: `miniprogram/components/ui/u-tab.vue`

**Interfaces:**
- Produces: `<u-tabs v-model:active="idx"><u-tab title="全部" /><u-tab title="进行中" /></u-tabs>`；emit `update:active`、`change`

- [ ] **Step 1: u-tab.vue（纯容器，声明 title）**

```vue
<template>
  <view class="u-tab" @click="onTap"><slot /></view>
</template>

<script setup>
import { defineEmits } from 'vue'
defineProps({ title: { type: String, default: '' } })
const emit = defineEmits(['tap'])
function onTap() { emit('tap') }
</script>
```

- [ ] **Step 2: u-tabs.vue（管理 active 与滑动条）**

```vue
<template>
  <view class="u-tabs">
    <scroll-view scroll-x :show-scrollbar="false" class="u-tabs-scroll">
      <view class="u-tabs-inner">
        <view
          v-for="(t, i) in tabs"
          :key="i"
          class="u-tabs-item"
          :class="{ 'u-tabs-item--active': i === active }"
          @click="onSelect(i)"
        >
          <text class="u-tabs-text">{{ t }}</text>
        </view>
      </view>
    </scroll-view>
    <view class="u-tabs-line" :style="{ left: lineLeft + '%', width: lineWidth + '%' }" />
  </view>
</template>

<script setup>
import { ref, watch, computed, useSlots, onMounted } from 'vue'
import { defineEmits } from 'vue'

const props = defineProps({
  active: { type: Number, default: 0 },
  type: { type: String, default: 'line' },
  swipeThreshold: { type: Number, default: 5 }
})
const emit = defineEmits(['update:active', 'change'])
const slots = useSlots()
const tabs = ref([])
const itemCount = computed(() => tabs.value.length || 1)
const lineLeft = computed(() => (props.active / itemCount.value) * 100)
const lineWidth = computed(() => (100 / itemCount.value) * 80)
function onSelect(i) {
  emit('update:active', i)
  emit('change', i)
}
onMounted(() => {
  const children = slots.default?.() || []
  tabs.value = children
    .filter(c => c.type?.name === 'Tab')
    .map(c => c.props?.title || '')
  if (!tabs.value.length) tabs.value = children.map(() => '')
})
</script>

<style scoped>
.u-tabs { position: relative; background: #fff; }
.u-tabs-scroll { white-space: nowrap; }
.u-tabs-inner { display: inline-flex; }
.u-tabs-item { padding: 24rpx 32rpx; font-size: 30rpx; color: var(--ui-color-text-secondary, #969799); position: relative; }
.u-tabs-item--active { color: var(--color-primary, #0A66C2); font-weight: 600; }
.u-tabs-line { position: absolute; bottom: 0; height: 6rpx; border-radius: 3rpx; background: var(--color-primary, #0A66C2); transition: left 0.2s; }
</style>
```

- [ ] **Step 3: Commit**

```bash
git add miniprogram/components/ui/u-tabs.vue miniprogram/components/ui/u-tab.vue
git commit -m "feat(miniprogram): u-tabs/u-tab 选项卡组件"
```

### Task 17: u-sticky

**Files:**
- Create: `miniprogram/components/ui/u-sticky.vue`

**Interfaces:**
- Produces: `<u-sticky offset-top="0"><slot /></u-sticky>`（基于 scroll-view 监听或 CSS position: sticky）

- [ ] **Step 1: 创建组件（CSS sticky 实现，微信小程序基础库 2.4.0+ 支持）**

```vue
<template>
  <view class="u-sticky" :style="{ top: offsetTop + 'px', zIndex }">
    <slot />
  </view>
</template>

<script setup>
defineProps({
  offsetTop: { type: [String, Number], default: 0 },
  zIndex: { type: Number, default: 99 }
})
</script>

<style scoped>
.u-sticky { position: sticky; }
</style>
```

- [ ] **Step 2: Commit**

```bash
git add miniprogram/components/ui/u-sticky.vue
git commit -m "feat(miniprogram): u-sticky 吸顶组件"
```

### Task 18: u-picker（简易单选选择器）

**Files:**
- Create: `miniprogram/components/ui/u-picker.vue`

**Interfaces:**
- Produces: `<u-picker :columns="['a','b']" :model-value="v" @update:model-value @confirm>`（底部弹层 + 取消/确定）

- [ ] **Step 1: 创建组件**

```vue
<template>
  <u-popup :show="show" position="bottom" round @close="onCancel">
    <view class="u-picker">
      <view class="u-picker-bar">
        <text class="u-picker-btn" @click="onCancel">取消</text>
        <text class="u-picker-title">{{ title }}</text>
        <text class="u-picker-btn u-picker-btn--confirm" @click="onConfirm">确定</text>
      </view>
      <picker-view :value="[index]" class="u-picker-view" @change="onChange">
        <picker-view-column>
          <view v-for="(item, i) in columns" :key="i" class="u-picker-item">{{ item }}</view>
        </picker-view-column>
      </picker-view>
    </view>
  </u-popup>
</template>

<script setup>
import { ref, watch, computed, defineEmits } from 'vue'

const props = defineProps({
  show: { type: Boolean, default: false },
  columns: { type: Array, default: () => [] },
  modelValue: { type: [String, Number], default: '' },
  title: { type: String, default: '请选择' }
})
const emit = defineEmits(['update:modelValue', 'update:show', 'confirm', 'cancel'])
const index = ref(0)
watch(() => props.modelValue, v => {
  const i = props.columns.indexOf(v)
  if (i >= 0) index.value = i
}, { immediate: true })
function onChange(e) {
  index.value = e.detail.value[0]
}
function onConfirm() {
  const v = props.columns[index.value]
  emit('update:modelValue', v)
  emit('confirm', v)
  emit('update:show', false)
}
function onCancel() {
  emit('cancel')
  emit('update:show', false)
}
</script>

<style scoped>
.u-picker { padding-bottom: env(safe-area-inset-bottom); }
.u-picker-bar { display: flex; align-items: center; justify-content: space-between; padding: 24rpx 32rpx; border-bottom: 1rpx solid var(--color-divider, #ebedf0); }
.u-picker-title { font-size: 30rpx; font-weight: 600; }
.u-picker-btn { font-size: 28rpx; color: var(--ui-color-text-secondary, #969799); }
.u-picker-btn--confirm { color: var(--color-primary, #0A66C2); font-weight: 600; }
.u-picker-view { height: 400rpx; }
.u-picker-item { display: flex; align-items: center; justify-content: center; font-size: 30rpx; color: var(--color-text, #1a1a1a); }
</style>
```

- [ ] **Step 2: Commit**

```bash
git add miniprogram/components/ui/u-picker.vue
git commit -m "feat(miniprogram): u-picker 简易选择器"
```

---

# 批 3：小程序高频页面迁移（约 25 页）

### Task 19: 迁移模式与首页

**Files:**
- Modify: `miniprogram/pages/home/index.vue`（示例页，完整迁移）

**迁移模式（本批所有页面统一应用）：**
1. 模板中 `van-button` → `u-button`、`van-cell` → `u-cell`、`van-empty` → `u-empty`、`van-loading` → `u-loading`、`van-tag` → `u-tag`、`van-nav-bar` → `u-nav-bar`、`van-search` → `u-search`、`van-popup` → `u-popup`、`van-tabs/van-tab` → `u-tabs/u-tab`、`van-sticky` → `u-sticky`、`van-icon` → `u-icon`、`van-field` → `u-field`、`van-picker` → `u-picker`
2. 属性保持同名（type/size/label/value/placeholder/disabled/is-link）；`van-cell-group` → `u-cell-group`；`van-cell-group` 的 `:border` 属性删除
3. `van-tabs` 的 `v-model:active` 保持；`van-popup` 的 `:show` 保持
4. 事件名保持（@click/@confirm/@change/@close/@back）
5. 页面 `<style>` 中的品牌化：`#ff6b35`/`#ff3b30`/`#1a1a1a` 等非品牌色替换为 `var(--color-*)` 或品牌色；emoji 图标替换为文字标签或 u-icon
6. 页面删掉对 `wxcomponents` 组件的 usingComponents 依赖（easycom 自动注册 u- 组件）

- [ ] **Step 1: 迁移 pages/home/index.vue**

按上述模式逐段替换（该页含 van-cell/van-tag/van-loading/van-nav-bar/van-search/van-icon/van-popup 等）。替换完成后：

```bash
cd miniprogram
grep -n "van-" pages/home/index.vue || echo "OK: home 无 van 残留"
```

- [ ] **Step 2: 品牌化检查（示例）**

检查该页 style 中的硬编码色值（如 `background: #ff6b35` 的 TabBar 激活条等），替换为品牌色/令牌。emoji（📍 等）替换为文字或 u-icon。

- [ ] **Step 3: 编译验证**

用 HBuilderX 或 `uni build` 编译小程序，确认 home 页渲染正常（无组件报错、样式符合预期）。

- [ ] **Step 4: Commit**

```bash
git add miniprogram/pages/home/index.vue
git commit -m "refactor(miniprogram): home 首页迁移至 u- 组件（品牌化）"
```

### Task 20: 高频列表/详情页迁移（第一批 ~10 页）

**Files:**
- Modify: `miniprogram/pages/mine/index.vue`
- Modify: `miniprogram/pages/tasks/index.vue`
- Modify: `miniprogram/pages/tasks/detail.vue`
- Modify: `miniprogram/pages/demands/list.vue`
- Modify: `miniprogram/pages/demands/detail.vue`
- Modify: `miniprogram/pages/demands/publish.vue`
- Modify: `miniprogram/pages/shops/index.vue`
- Modify: `miniprogram/pages/mall/index.vue`
- Modify: `miniprogram/pages/publish/index.vue`
- Modify: `miniprogram/pages/login/index.vue`

**迁移模式：同 Task 19 Step 1（每页 van-x → u-x + 品牌化）**

- [ ] **Step 1: 批量替换标签**

对每页执行模板替换（参照 Task 19 模式 1-6）。对纯机械替换，可用脚本辅助：

```bash
cd miniprogram/pages
for f in mine/index.vue tasks/index.vue tasks/detail.vue demands/list.vue demands/detail.vue demands/publish.vue shops/index.vue mall/index.vue publish/index.vue login/index.vue; do
  sed -i 's/van-button/u-button/g; s/van-cell-group/u-cell-group/g; s/van-cell/u-cell/g; s/van-field/u-field/g; s/van-empty/u-empty/g; s/van-loading/u-loading/g; s/van-tag/u-tag/g; s/van-nav-bar/u-nav-bar/g; s/van-search/u-search/g; s/van-popup/u-popup/g; s/van-tabs/u-tabs/g; s/van-tab/u-tab/g; s/van-sticky/u-sticky/g; s/van-icon/u-icon/g; s/van-picker/u-picker/g; s/van-image/u-image/g' "$f"
done
```

注意：sed 后必须**逐个页面人工检查**——属性差异（如 `van-tabs` 的 `:active` 属性、`van-cell-group` 的 border 属性、`van-popup` 的 `@close` 事件）和布局适配（u- 组件视觉与 van 略有差异）。**每页检查后**再编译。

- [ ] **Step 2: 品牌化清理（每页）**

非品牌色硬编码替换；emoji 替换。参考 Task 19 Step 2。

- [ ] **Step 3: 编译验证**

小程序编译通过；逐页浏览 10 个页面无报错、布局正常。

- [ ] **Step 4: Commit（每 2-3 页一个 commit，便于回滚）**

```bash
git add miniprogram/pages/mine miniprogram/pages/tasks
git commit -m "refactor(miniprogram): mine/tasks 迁移至 u- 组件"
git add miniprogram/pages/demands
git commit -m "refactor(miniprogram): demands 系列页迁移至 u- 组件"
# ... 其余页面按模块分 commit
```

### Task 21: 高频页面第二批（~15 页）

**Files:**
- Modify: `miniprogram/pages/training/courses.vue`、`training/enroll.vue`、`training/certificates.vue`、`training/register.vue`
- Modify: `miniprogram/pages/jobs/list.vue`、`jobs/resume.vue`
- Modify: `miniprogram/pages/experts/list.vue`、`experts/detail.vue`
- Modify: `miniprogram/pages/events/list.vue`、`events/detail.vue`、`events/register.vue`
- Modify: `miniprogram/pages/colleges/list.vue`、`colleges/detail.vue`
- Modify: `miniprogram/pages/emergency/resources.vue`
- Modify: `miniprogram/pages/search/index.vue`

**迁移模式：同 Task 19（逐页替换 + 品牌化 + 编译验证）**

- [ ] **Step 1: 逐页迁移（每页按 Task 19 模式）**

先 sed 机械替换，再逐页人工检查属性/布局。

- [ ] **Step 2: 品牌化清理**

- [ ] **Step 3: 编译验证**

- [ ] **Step 4: Commit（按模块分组）**

```bash
git add miniprogram/pages/training
git commit -m "refactor(miniprogram): training 页面迁移至 u- 组件"
# ... jobs/experts/events/colleges/emergency/search 各自或合并 commit
```

---

# 批 4：中低频页面迁移 + 清理

### Task 22: 中低频页面迁移（~43 页，按目录批量）

**Files:**
- Modify: `miniprogram/pages/` 下剩余全部含 van- 的页面（achievements、challenges、compliance、pilots、portfolios、reports、resources、projects、testsites、study、applications、messages、services、admin、cases、more、webview、register、index 等）

**迁移模式：同 Task 19**

- [ ] **Step 1: 找出所有残留 van- 的页面**

```bash
cd miniprogram
grep -rl "van-" pages/ | sort
```

- [ ] **Step 2: 逐目录迁移**

对每个目录：sed 机械替换 + 人工检查 + 品牌化。每完成一个目录 `grep -rn "van-" pages/<dir>/ || echo OK`。

- [ ] **Step 3: 编译验证**

全量编译小程序，逐页浏览全部 68 页。

- [ ] **Step 4: Commit（每 3-5 个目录一个 commit）**

### Task 23: 清理 Vant Weapp 与收尾

**Files:**
- Delete: `miniprogram/wxcomponents/`
- Modify: `miniprogram/pages.json`（删除 globalStyle.usingComponents 中全部 van- 声明）
- Modify: `miniprogram/App.vue`（如引用 van 样式则清理）

- [ ] **Step 1: 删除 usingComponents 中 van 声明**

pages.json 的 `globalStyle.usingComponents` 中删除所有 `"van-*": "/wxcomponents/@vant/weapp/*/index"` 行。

- [ ] **Step 2: 删除 wxcomponents 目录**

```bash
rm -rf miniprogram/wxcomponents
```

- [ ] **Step 3: 残留检查 + 编译**

```bash
cd miniprogram
grep -r "van-" pages/ pages.json || echo "OK: 无 van 残留"
grep -r "wxcomponents" pages.json || echo "OK: 无 wxcomponents 引用"
# 全量编译验证
```

- [ ] **Step 4: Commit**

```bash
git add -A miniprogram
git commit -m "refactor(miniprogram): 删除 wxcomponents（Vant Weapp），68 页全部迁移至 u- 组件库"
```

### Task 24: 全量验收

- [ ] **Step 1: frontend 验收**

```bash
cd frontend
npx vite build
grep -rn "from 'vant'" src/ || echo "OK: vant 清零"
grep -c "vant" package.json || echo "OK: package.json 无 vant"
```

- [ ] **Step 2: miniprogram 验收**

```bash
cd miniprogram
grep -r "van-" pages/ pages.json || echo "OK: van 清零"
ls wxcomponents 2>/dev/null || echo "OK: wxcomponents 已删除"
```

- [ ] **Step 3: 视觉抽查**

小程序 5 个 Tab 页 + Admin Dashboard/聚合页截图对比：主色为品牌蓝/青绿；无 emoji 图标；无橙色/红色等非品牌硬编码。

- [ ] **Step 4: 功能回归**

小程序核心流程：登录 → 需求大厅 → 详情 → 发布；Admin：登录 → 各聚合页数据浏览。确认行为与重构前一致。

- [ ] **Step 5: 更新 CLAUDE.md 前端项目表**

将"微信小程序（68 页，5 Tab）"技术栈从"Vant Weapp"改为"自研 u- 组件库"，Web 管理后台技术栈移除 Vant。

- [ ] **Step 6: Commit**

```bash
git add CLAUDE.md
git commit -m "docs: 前端技术栈更新（Vant → 自研 u- 组件库）"
```
