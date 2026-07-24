# Pinia 状态管理使用指南

## 简介

项目已集成 Pinia 状态管理,提供了以下 Store:

- `useUserStore` - 用户认证和资料管理
- `useApplicationStore` - 申请/订单管理
- `useServiceStore` - 服务配置管理

## 快速开始

### 1. 在组件中使用 Store

```vue
<script setup>
import { useUserStore } from '@/stores/user'
import { useApplicationStore } from '@/stores/application'

const userStore = useUserStore()
const appStore = useApplicationStore()

// 访问状态
console.log(userStore.isLoggedIn)
console.log(appStore.applications)

// 调用 action
await userStore.login('13800138000', 'password')
await appStore.fetchApplications()
</script>
```

### 2. 用户认证示例

```vue
<script setup>
import { useUserStore } from '@/stores/user'
import { useRouter } from 'vue-router'

const userStore = useUserStore()
const router = useRouter()

const handleLogin = async () => {
  const result = await userStore.login(phone.value, password.value)

  if (result.success) {
    // 登录成功,跳转到首页
    router.push('/home')
  } else {
    // 显示错误消息
    alert(result.message)
  }
}

const handleLogout = async () => {
  await userStore.logout()
  router.push('/login')
}
</script>

<template>
  <div>
    <!-- 根据登录状态显示不同内容 -->
    <template v-if="userStore.isLoggedIn">
      <p>欢迎, {{ userStore.displayName }}</p>
      <button v-if="userStore.isAdmin" @click="goToAdmin">
        进入后台
      </button>
      <button @click="handleLogout">退出登录</button>
    </template>

    <template v-else>
      <button @click="router.push('/login')">登录</button>
    </template>
  </div>
</template>
```

### 3. 申请列表示例

```vue
<script setup>
import { ref, onMounted } from 'vue'
import { useApplicationStore } from '@/stores/application'

const appStore = useApplicationStore()
const currentPage = ref(1)

onMounted(async () => {
  // 加载第一页数据
  await appStore.fetchApplications({ page: 1, limit: 20 })
})

const loadMore = async () => {
  currentPage.value++
  await appStore.fetchApplications({
    page: currentPage.value,
    limit: 20
  })
}

const updateStatus = async (id, status) => {
  const result = await appStore.updateApplicationStatus(id, status)

  if (result.success) {
    alert('状态更新成功')
  } else {
    alert(result.message)
  }
}

const handleExport = async () => {
  const result = await appStore.exportApplications()

  if (result.success) {
    alert('导出成功')
  } else {
    alert(result.message)
  }
}
</script>

<template>
  <div>
    <!-- 加载状态 -->
    <div v-if="appStore.loading">加载中...</div>

    <!-- 申请列表 -->
    <div v-else-if="appStore.hasData">
      <div v-for="app in appStore.applications" :key="app.id">
        <h3>{{ app.serviceName }}</h3>
        <p>状态: {{ app.status }}</p>
        <button @click="updateStatus(app.id, '处理中')">
          标记为处理中
        </button>
      </div>

      <!-- 分页 -->
      <button
        v-if="currentPage < appStore.pagination.totalPages"
        @click="loadMore"
      >
        加载更多
      </button>
    </div>

    <!-- 空状态 -->
    <div v-else>暂无数据</div>

    <!-- 导出按钮 -->
    <button @click="handleExport">导出 Excel</button>
  </div>
</template>
```

### 4. 服务配置示例

```vue
<script setup>
import { ref, onMounted } from 'vue'
import { useServiceStore } from '@/stores/service'

const serviceStore = useServiceStore()

onMounted(async () => {
  await serviceStore.fetchServices()
})

const updateServiceConfig = async (serviceId, config) => {
  const result = await serviceStore.updateServices({
    [serviceId]: config
  })

  if (result.success) {
    alert('配置更新成功')
  }
}
</script>

<template>
  <div>
    <div v-if="serviceStore.loading">加载中...</div>

    <div v-else>
      <div v-for="service in serviceStore.serviceList" :key="service.id">
        <h3>{{ service.name }}</h3>
        <p>{{ service.intro }}</p>
      </div>
    </div>
  </div>
</template>
```

## 状态说明

### useUserStore

**State:**
- `user` - 当前用户信息
- `accessToken` - 访问令牌
- `refreshToken` - 刷新令牌

**Getters:**
- `isLoggedIn` - 是否已登录
- `isSuperAdmin` - 是否超级管理员
- `isAdmin` - 是否管理员
- `isDslAdmin` - 是否DSL管理员
- `isStudyAdmin` - 是否研学管理员
- `canManage` - 是否有管理权限
- `displayName` - 用户显示名称

**Actions:**
- `login(phone, password)` - 登录
- `register(phone, password, name)` - 注册
- `logout()` - 登出
- `fetchUser()` - 刷新用户信息
- `setUser(user, token, refreshToken)` - 设置用户信息
- `clearUser()` - 清除用户信息
- `updateProfile(profile)` - 更新资料

### useApplicationStore

**State:**
- `applications` - 申请列表
- `currentApplication` - 当前申请详情
- `loading` - 加载状态
- `pagination` - 分页信息
- `filters` - 筛选条件

**Getters:**
- `pendingApplications` - 待处理的申请
- `processingApplications` - 处理中的申请
- `completedApplications` - 已完成的申请
- `hasData` - 是否有数据

**Actions:**
- `fetchApplications(params)` - 获取列表
- `fetchApplication(id)` - 获取详情
- `updateApplicationStatus(id, status, remark)` - 更新状态
- `exportApplications()` - 导出Excel
- `clear()` - 清空状态

### useServiceStore

**State:**
- `services` - 服务配置
- `loading` - 加载状态
- `error` - 错误信息

**Getters:**
- `getServiceConfig(serviceId)` - 获取特定服务配置
- `serviceList` - 服务列表
- `studyShowcase` - 研学展示数据

**Actions:**
- `fetchServices()` - 获取配置
- `updateServices(config)` - 更新配置
- `fetchStudyShowcase()` - 获取研学展示
- `updateStudyShowcase(showcase)` - 更新研学展示
- `clear()` - 清空状态

## 最佳实践

### 1. 组件卸载时清空状态

```javascript
import { onUnmounted } from 'vue'

onUnmounted(() => {
  appStore.clear()
})
```

### 2. 使用 watch 监听状态变化

```javascript
import { watch } from 'vue'

watch(() => userStore.isLoggedIn, (newValue) => {
  if (!newValue) {
    router.push('/login')
  }
})
```

### 3. 使用 computed 优化性能

```javascript
import { computed } from 'vue'

const pendingCount = computed(() =>
  appStore.pendingApplications.length
)
```

### 4. 错误处理

```javascript
try {
  await appStore.fetchApplications()
} catch (error) {
  console.error('Failed to load:', error)
  // 显示错误提示
}
```

## 与 localStorage 的兼容

Store 会自动同步状态到 localStorage,所以:

1. 页面刷新后状态会自动恢复
2. 可以直接使用 `localStorage.getItem('user')` 获取用户信息
3. Store 状态变化会自动保存到 localStorage

## 迁移指南

如果你之前使用 `localStorage` 直接管理状态,可以逐步迁移:

**旧方式:**
```javascript
const user = JSON.parse(localStorage.getItem('user'))
const isAdmin = user?.role === 'admin'
```

**新方式:**
```javascript
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()
const isAdmin = userStore.isAdmin
```

优势:
- 类型安全
- 响应式更新
- 集中管理
- 易于测试
