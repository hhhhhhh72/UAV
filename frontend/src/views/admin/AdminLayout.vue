<template>
  <a-layout class="layout-container">
    <a-layout-header class="layout-header">
      <div class="logo" @click="goHome">
        <span class="logo-mark">
          <span class="logo-dot"></span>
        </span>
        <span class="logo-text">低空运营后台</span>
      </div>
      <div class="right-actions">
        <a-tooltip content="搜索">
          <a-button type="text" shape="circle" class="nav-btn" @click="searchVisible = true"><template #icon><icon-search size="18"/></template></a-button>
        </a-tooltip>
        <a-tooltip content="通知">
          <a-popover trigger="click" position="br" :content-style="{ padding: 0, minWidth: '320px' }" @popup-visible-change="(v) => v && loadMessages()">
            <a-button type="text" shape="circle" class="nav-btn">
              <template #icon>
                <a-badge :count="notifyCount" :max-count="99">
                  <icon-notification size="18"/>
                </a-badge>
              </template>
            </a-button>
            <template #content>
              <div class="notify-panel">
                <a-tabs default-active-key="1" size="medium" :style="{ padding: '0 16px' }">
                  <a-tab-pane key="1">
                    <template #title>消息 ({{ messages.length }})</template>
                    <a-list :bordered="false" size="small">
                      <a-list-item v-for="(msg, i) in messages" :key="msg.id" @click="readMessage(i)">
                        <a-list-item-meta :title="msg.title" :description="formatTime(msg.created_at)">
                          <template #avatar>
                            <a-avatar :style="{ backgroundColor: msg.is_read ? '#C9CDD4' : '#165DFF' }" :size="32">
                              <icon-message />
                            </a-avatar>
                          </template>
                        </a-list-item-meta>
                      </a-list-item>
                      <a-empty v-if="!messages.length" description="暂无消息" style="padding: 40px 0;" />
                    </a-list>
                  </a-tab-pane>
                  <a-tab-pane key="2">
                    <template #title>通知 ({{ notices.length }})</template>
                    <a-list v-if="notices.length" :bordered="false" size="small">
                      <a-list-item v-for="(n, i) in notices" :key="i">
                        <a-list-item-meta :title="n.title" :description="n.time" />
                      </a-list-item>
                    </a-list>
                    <a-empty v-else description="暂无通知" style="padding: 40px 0;" />
                  </a-tab-pane>
                  <a-tab-pane key="3">
                    <template #title>待办 ({{ todos.length }})</template>
                    <a-list v-if="todos.length" :bordered="false" size="small">
                      <a-list-item v-for="(t, i) in todos" :key="i">
                        <a-list-item-meta :title="t.title" :description="t.time" />
                      </a-list-item>
                    </a-list>
                    <a-empty v-else description="暂无待办" style="padding: 40px 0;" />
                  </a-tab-pane>
                </a-tabs>
                <div class="notify-footer">
                  <a-button type="text" long @click="markAllRead">全部已读</a-button>
                </div>
              </div>
            </template>
          </a-popover>
        </a-tooltip>
        <a-tooltip content="全屏">
          <a-button type="text" shape="circle" class="nav-btn" @click="toggleFullscreen"><template #icon><icon-fullscreen size="18" v-if="!isFullscreen" /><icon-fullscreen-exit size="18" v-else /></template></a-button>
        </a-tooltip>
        <a-tooltip content="设置">
          <a-button type="text" shape="circle" class="nav-btn" @click="settingsVisible = true"><template #icon><icon-settings size="18"/></template></a-button>
        </a-tooltip>
        <a-dropdown trigger="click">
          <a-avatar :size="32" :style="{ backgroundColor: '#165DFF', cursor: 'pointer', marginLeft: '8px' }">{{ userInitial }}</a-avatar>
          <template #content>
            <a-doption @click="goHome"><icon-home /> 返回首页</a-doption>
            <a-divider :margin="4" />
            <a-doption @click="handleLogout"><icon-export /> 退出登录</a-doption>
          </template>
        </a-dropdown>
      </div>
    </a-layout-header>
    <a-layout>
      <a-layout-sider
        class="layout-sider"
        :width="menuWidth"
        :collapsed-width="48"
        collapsible
        :collapsed="collapsed"
        hide-trigger
        breakpoint="xl"
        @collapse="onCollapse"
      >
        <div class="menu-wrapper">
          <a-menu
            :selected-keys="selectedKeys"
            :open-keys="openKeys"
            @menu-item-click="onMenuItemClick"
            @sub-menu-click="onSubMenuClick"
            :style="{ width: '100%' }"
            :collapsed="collapsed"
            auto-open-selected
          >
            <a-menu-item v-for="item in visibleMenus" :key="item.path">
              <template #icon><component :is="item.icon" /></template>
              {{ item.label }}
            </a-menu-item>
          </a-menu>
        </div>
        <div class="collapse-btn" @click="collapsed = !collapsed">
          <icon-menu-unfold v-if="collapsed" />
          <icon-menu-fold v-else />
        </div>
      </a-layout-sider>
      <a-layout-content class="layout-content">
        <div class="page-content-wrapper">
          <router-view v-slot="{ Component }">
            <transition name="fade" mode="out-in">
              <component :is="Component" />
            </transition>
          </router-view>
        </div>
      </a-layout-content>
    </a-layout>
  </a-layout>

  <!-- 消息详情弹窗 -->
  <a-modal v-model:visible="detailVisible" :footer="false" :title="currentMsg?.title || '消息详情'" :width="'min(520px, 94vw)'" :unmount-on-close="true">
    <template v-if="currentMsg">
      <div class="msg-detail-meta">
        <a-space :size="12">
          <a-tag :color="currentMsg.is_read ? 'gray' : 'blue'" size="small">{{ currentMsg.is_read ? '已读' : '未读' }}</a-tag>
          <span class="msg-detail-time">{{ formatTime(currentMsg.created_at) }}</span>
        </a-space>
      </div>
      <div class="msg-detail-content">{{ currentMsg.content || '暂无内容' }}</div>
    </template>
  </a-modal>

  <!-- 搜索菜单弹窗 -->
  <a-modal v-model:visible="searchVisible" :footer="false" :closable="false" simple unmount-on-close>
    <a-input-search v-model="searchKeyword" placeholder="搜索菜单..." size="large" allow-clear @search="onSearch" @press-enter="onSearch" />
    <div v-if="searchResults.length" class="search-results">
      <div v-for="item in searchResults" :key="item.path" class="search-result-item" @click="goToSearchResult(item.path)">
        {{ item.label }}
      </div>
    </div>
  </a-modal>

  <!-- 页面配置抽屉 -->
  <a-drawer v-model:visible="settingsVisible" title="页面配置" :width="320" unmount-on-close>
    <div class="settings-section">
      <div class="settings-label">主题模式</div>
      <a-radio-group v-model="theme" type="button" @change="changeTheme">
        <a-radio value="light">浅色</a-radio>
        <a-radio value="dark">暗色</a-radio>
      </a-radio-group>
    </div>
    <div class="settings-section">
      <div class="settings-label">导航栏</div>
      <a-space direction="vertical" :size="12">
        <a-checkbox v-model="settings.showBreadcrumb">显示面包屑</a-checkbox>
      </a-space>
    </div>
  </a-drawer>
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import axios from '@/utils/http'
import { useAuth } from './composables/useAuth'

const { isPlatformAdmin, isAssociationAdmin, refreshCurrentUser } = useAuth()

const router = useRouter()
const route = useRoute()

/* 菜单（照抄 Arco Pro 结构：图标 + 名称；权限过滤） */
const allMenus = [
  { path: '/admin/dashboard', label: '数据看板', icon: 'icon-dashboard', roles: ['platform_admin', 'association_admin'] },
  { path: '/admin/members', label: '会员管理', icon: 'icon-user-group', roles: ['platform_admin', 'association_admin'] },
  { path: '/admin/trading', label: '交易管理', icon: 'icon-list', roles: ['platform_admin', 'association_admin'] },
  { path: '/admin/content', label: '内容管理', icon: 'icon-file', roles: ['platform_admin', 'association_admin'] },
  { path: '/admin/articles', label: '资讯管理', icon: 'icon-file', roles: ['platform_admin', 'association_admin'] },
  { path: '/admin/talent', label: '人才教育', icon: 'icon-book', roles: ['platform_admin', 'association_admin'] },
  { path: '/admin/innovation', label: '产学研', icon: 'icon-experiment', roles: ['platform_admin', 'association_admin'] },
  { path: '/admin/promotion', label: '运营推广', icon: 'icon-bulb', roles: ['platform_admin', 'association_admin'] },
  { path: '/admin/emergency', label: '应急协同', icon: 'icon-fire', roles: ['platform_admin', 'association_admin'] },
  { path: '/admin/settings', label: '系统设置', icon: 'icon-settings', roles: ['platform_admin', 'association_admin'] }
]

const visibleMenus = computed(() => {
  return allMenus.filter(m => {
    if (isPlatformAdmin.value) return true
    if (isAssociationAdmin.value) return m.roles.includes('association_admin')
    return false
  })
})

/* 顶栏状态 */
const searchVisible = ref(false)
const searchKeyword = ref('')
const settingsVisible = ref(false)
const theme = ref('light')
const isFullscreen = ref(false)
const settings = ref({ showBreadcrumb: true })

const userStr = localStorage.getItem('user')
const user = userStr ? JSON.parse(userStr) : null
const userName = user?.name || user?.phone || '管理员'
const userInitial = (userName || '管').charAt(0)

/* 消息通知（真实 API：/api/v1/messages*，按当前登录用户） */
const messages = ref([])
const notices = ref([])
const todos = ref([])
const notifyCount = ref(0)
let pollTimer = null

const formatTime = (d) => {
  if (!d) return ''
  const t = new Date(d).getTime()
  if (isNaN(t)) return ''
  const diff = (Date.now() - t) / 1000
  if (diff < 60) return '刚刚'
  if (diff < 3600) return `${Math.floor(diff / 60)} 分钟前`
  if (diff < 86400) return `${Math.floor(diff / 3600)} 小时前`
  const dt = new Date(t)
  return `${dt.getMonth() + 1}月${dt.getDate()}日`
}

const loadMessages = async () => {
  try {
    const list = await axios.get('/api/v1/messages').then(r => r.data)
    messages.value = Array.isArray(list) ? list : []
    const res = await axios.get('/api/v1/messages/unread-count').then(r => r.data)
    notifyCount.value = res?.count || 0
  } catch (e) {
    // 网络异常时保留当前数据，不打断页面
  }
}

const readMessage = async (i) => {
  const msg = messages.value[i]
  if (!msg) return
  currentMsg.value = msg
  detailVisible.value = true
  if (msg.is_read) return
  try {
    await axios.post(`/api/v1/messages/${msg.id}/read`)
    messages.value[i].is_read = true
    if (notifyCount.value > 0) notifyCount.value--
  } catch (e) { /* 忽略，角标以 unread-count 为准 */ }
}

const markAllRead = async () => {
  const unread = messages.value.filter(m => !m.is_read)
  await Promise.all(unread.map(m => axios.post(`/api/v1/messages/${m.id}/read`).catch(() => {})))
  await loadMessages()
}

/* 消息详情弹窗 */
const detailVisible = ref(false)
const currentMsg = ref(null)

/* 折叠 */
const menuWidth = ref(220)
const collapsed = ref(false)

const onCollapse = (v) => {
  collapsed.value = v
  menuWidth.value = v ? 48 : 220
}

/* 菜单高亮 + 展开 */
const selectedKeys = computed(() => [route.path])
const openKeys = ref([])

watch(() => route.path, () => {
  const path = route.path
  const menu = allMenus.find(m => m.path === path)
  if (menu) openKeys.value = [menu.path]
}, { immediate: true })

const onMenuItemClick = (key) => {
  if (key !== route.path) router.push(key)
}

const onSubMenuClick = (key) => {
  const idx = openKeys.value.indexOf(key)
  if (idx >= 0) openKeys.value.splice(idx, 1)
  else openKeys.value.push(key)
}

/* 搜索菜单 */
const searchResults = computed(() => {
  if (!searchKeyword.value) return []
  const kw = searchKeyword.value.toLowerCase()
  return visibleMenus.value.filter(m => m.label.toLowerCase().includes(kw))
})

const onSearch = () => {
  if (searchResults.value.length) goToSearchResult(searchResults.value[0].path)
}

const goToSearchResult = (path) => {
  router.push(path)
  searchVisible.value = false
  searchKeyword.value = ''
}

/* 全屏 */
const toggleFullscreen = () => {
  if (!document.fullscreenElement) {
    document.documentElement.requestFullscreen()
    isFullscreen.value = true
  } else {
    document.exitFullscreen()
    isFullscreen.value = false
  }
}

/* 主题：Arco 暗色变量 + 内联兜底（内联样式优先级最高，确保暗色生效） */
const DARK_VARS = {
  '--color-bg-page': '#17171a',
  '--color-bg-1': '#17171a',
  '--color-bg-2': '#232324',
  '--color-bg-3': '#2a2a2b',
  '--color-bg-4': '#2e2e30',
  '--color-bg-5': '#333335',
  '--color-fill-1': 'rgba(255,255,255,0.04)',
  '--color-fill-2': 'rgba(255,255,255,0.08)',
  '--color-fill-3': 'rgba(255,255,255,0.12)',
  '--color-fill-4': 'rgba(255,255,255,0.16)',
  '--color-text-1': 'rgba(255,255,255,0.9)',
  '--color-text-2': 'rgba(255,255,255,0.7)',
  '--color-text-3': 'rgba(255,255,255,0.5)',
  '--color-text-4': 'rgba(255,255,255,0.3)',
  '--color-border': '#333335',
  '--color-border-2': '#333335',
  '--color-neutral-2': 'rgba(255,255,255,0.08)',
  '--color-neutral-3': 'rgba(255,255,255,0.12)',
  '--color-neutral-4': 'rgba(255,255,255,0.16)',
  '--color-neutral-5': 'rgba(255,255,255,0.24)',
  '--color-neutral-6': 'rgba(255,255,255,0.34)',
  '--color-neutral-7': 'rgba(255,255,255,0.48)',
  '--color-neutral-8': 'rgba(255,255,255,0.68)',
  '--color-neutral-9': 'rgba(255,255,255,0.82)',
  '--color-neutral-10': 'rgba(255,255,255,0.9)'
}

const changeTheme = (value) => {
  if (value === 'dark') {
    document.body.setAttribute('arco-theme', 'dark')
    Object.entries(DARK_VARS).forEach(([k, v]) => document.body.style.setProperty(k, v))
    localStorage.setItem('arco-theme', 'dark')
  } else {
    document.body.removeAttribute('arco-theme')
    Object.keys(DARK_VARS).forEach(k => document.body.style.removeProperty(k))
    localStorage.setItem('arco-theme', 'light')
  }
}

// 挂载时恢复主题（刷新后保留暗色）
const restoreTheme = () => {
  const saved = localStorage.getItem('arco-theme')
  if (saved === 'dark') {
    theme.value = 'dark'
    document.body.setAttribute('arco-theme', 'dark')
    Object.entries(DARK_VARS).forEach(([k, v]) => document.body.style.setProperty(k, v))
  }
}

/* 导航 */
const goHome = () => router.push('/')

const handleLogout = () => {
  // 登出时先吊销 refresh token（后端 revoke 落库），再清本地登录态；
  // 吊销失败不阻塞登出（本地清理必须执行，避免残留登录态）。
  const refreshToken = localStorage.getItem('refreshToken')
  if (refreshToken) {
    axios.post('/api/auth/logout', { refresh_token: refreshToken }).catch(() => {})
  }
  localStorage.removeItem('user')
  localStorage.removeItem('accessToken')
  localStorage.removeItem('refreshToken')
  router.push('/login')
}

onMounted(() => {
  refreshCurrentUser()
  restoreTheme()
  loadMessages()
  pollTimer = setInterval(loadMessages, 60000)
})

onUnmounted(() => {
  if (pollTimer) clearInterval(pollTimer)
})
</script>

<style scoped>
/* 照抄 Arco Design Pro layout 全部样式 */
.layout-container {
  height: 100vh;
}

.layout-header {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  background-color: var(--color-bg-2);
  border-bottom: 1px solid var(--color-border);
  box-sizing: border-box;
  z-index: 100;
}

.logo {
  display: flex;
  align-items: center;
  cursor: pointer;
}

.logo-mark {
  width: 28px;
  height: 28px;
  border-radius: 6px;
  background: #165DFF;
  display: flex;
  align-items: center;
  justify-content: center;
}

.logo-dot {
  width: 10px;
  height: 10px;
  border-radius: 3px;
  background: #1DD4A8;
}

.logo-text {
  font-size: 18px;
  font-weight: 600;
  color: var(--color-text-1);
  margin-left: 10px;
}

.right-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.nav-btn {
  color: var(--color-text-1);
  font-size: 16px;
  width: 36px;
  height: 36px;
  background-color: var(--color-fill-2);
}

.nav-btn:hover {
  background-color: var(--color-fill-3);
}

.layout-sider {
  background-color: var(--color-bg-2);
  border-right: 1px solid var(--color-border);
  z-index: 99;
  display: flex;
  flex-direction: column;
}

.menu-wrapper {
  padding-top: 8px;
  flex: 1;
  overflow-y: auto;
}

.menu-wrapper::-webkit-scrollbar {
  width: 6px;
}

.menu-wrapper::-webkit-scrollbar-thumb {
  background-color: var(--color-fill-3);
  border-radius: 4px;
}

.collapse-btn {
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  cursor: pointer;
  color: var(--color-text-2);
  border-top: 1px solid var(--color-border);
  transition: all 0.2s;
}

.collapse-btn:hover {
  color: var(--color-text-1);
  background-color: var(--color-fill-2);
}

.layout-content {
  background-color: var(--color-bg-page);
  height: calc(100vh - 60px);
  overflow: auto;
}

.page-content-wrapper {
  padding: 20px;
}

/* 通知面板 */
.notify-panel {
  width: 320px;
}

.notify-footer {
  border-top: 1px solid var(--color-border);
  padding: 8px;
}

.notify-panel :deep(.arco-list-item) {
  cursor: pointer;
}

.notify-panel :deep(.arco-list-item:hover) {
  background-color: var(--color-fill-2);
}

/* 消息详情弹窗 */
.msg-detail-meta {
  margin-bottom: 16px;
}

.msg-detail-time {
  color: var(--color-text-3);
  font-size: 12px;
}

.msg-detail-content {
  color: var(--color-text-1);
  line-height: 1.8;
  white-space: pre-wrap;
  word-break: break-word;
}

/* 搜索弹窗 */
.search-results {
  margin-top: 12px;
}

.search-result-item {
  padding: 10px 12px;
  border-radius: 4px;
  cursor: pointer;
  color: var(--color-text-1);
  transition: background 0.15s;
}

.search-result-item:hover {
  background-color: var(--color-fill-2);
}

/* 设置抽屉 */
.settings-section {
  margin-bottom: 24px;
}

.settings-label {
  font-size: 14px;
  font-weight: 500;
  color: var(--color-text-1);
  margin-bottom: 12px;
}

/* 页面过渡 */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.15s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
