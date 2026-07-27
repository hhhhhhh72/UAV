<template>
  <aside class="admin-sidebar" :class="{ open: modelValue }">
    <div class="sidebar-brand">
      <span class="brand-text">后台管理</span>
    </div>

    <nav class="sidebar-nav">
      <template v-for="item in visibleMenus" :key="item.path || item.label">
        <div v-if="item.type === 'divider'" class="nav-divider">{{ item.label }}</div>
        <router-link
          v-else
          :to="item.path"
          class="nav-item"
          :class="{ active: isActive(item.path) }"
          @click="$emit('update:modelValue', false)"
        >
          <span class="nav-label">{{ item.label }}</span>
        </router-link>
      </template>
    </nav>

    <div class="sidebar-footer">
      <router-link to="/" class="nav-item" @click="$emit('update:modelValue', false)">
        <span class="nav-label">返回首页</span>
      </router-link>
    </div>
  </aside>

  <!-- Mobile overlay -->
  <transition name="fade-overlay">
    <div v-if="modelValue" class="sidebar-overlay" @click="$emit('update:modelValue', false)" />
  </transition>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'

const props = defineProps({
  modelValue: Boolean,
  isPlatformAdmin: Boolean,
  isAssociationAdmin: Boolean
})

defineEmits(['update:modelValue'])

const route = useRoute()

const allMenus = [
  { path: '/admin/dashboard', label: '📊 数据看板', roles: ['platform_admin', 'association_admin'] },
  { path: '/admin/members', label: '🏢 会员管理', roles: ['platform_admin', 'association_admin'] },
  { path: '/admin/trading', label: '📦 交易管理', roles: ['platform_admin', 'association_admin'] },
  { path: '/admin/content', label: '📋 内容审核', roles: ['platform_admin', 'association_admin'] },
  { path: '/admin/talent', label: '🎓 人才教育', roles: ['platform_admin', 'association_admin'] },
  { path: '/admin/innovation', label: '🔬 产学研', roles: ['platform_admin', 'association_admin'] },
  { path: '/admin/promotion', label: '📣 运营推广', roles: ['platform_admin', 'association_admin'] },
  { path: '/admin/emergency', label: '🚨 应急协同', roles: ['platform_admin', 'association_admin'] },
  { path: '/admin/settings', label: '⚙️ 系统设置', roles: ['platform_admin', 'association_admin'] }
]

const visibleMenus = computed(() => {
  return allMenus.filter(m => {
    if (props.isPlatformAdmin) return true
    if (props.isAssociationAdmin) return m.roles.includes('association_admin')
    return false
  })
})

const isActive = (path) => route.path === path
</script>

<style scoped>
.admin-sidebar {
  position: fixed;
  top: 0;
  left: 0;
  bottom: 0;
  width: var(--sidebar-width, 220px);
  background: var(--bg-sidebar, #fbfbfd);
  border-right: 1px solid var(--border-color, #e5e5e7);
  display: flex;
  flex-direction: column;
  z-index: 200;
  transition: transform 0.28s cubic-bezier(0.4, 0, 0.2, 1);
}

.sidebar-brand {
  height: var(--admin-header-height, 56px);
  display: flex;
  align-items: center;
  padding: 0 20px;
  border-bottom: 1px solid var(--border-color, #e5e5e7);
}

.brand-text {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-color, #1d1d1f);
  letter-spacing: -0.2px;
}

.sidebar-nav {
  flex: 1;
  overflow-y: auto;
  padding: 8px 0;
}

.nav-item {
  display: flex;
  align-items: center;
  padding: 10px 20px;
  margin: 2px 8px;
  border-radius: 8px;
  color: var(--text-secondary, #86868b);
  text-decoration: none;
  font-size: 14px;
  transition: all 0.15s ease;
  cursor: pointer;
}

.nav-item:hover {
  background: rgba(0, 0, 0, 0.04);
  color: var(--text-color, #1d1d1f);
}

.nav-item.active {
  background: var(--accent-light, #e8f2fc);
  color: var(--accent-color, #0071e3);
  font-weight: 500;
}

.nav-label {
  white-space: nowrap;
}

.nav-divider {
  padding: 12px 20px 4px;
  font-size: 11px;
  font-weight: 600;
  color: #999;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.sidebar-footer {
  padding: 8px 0;
  border-top: 1px solid var(--border-color, #e5e5e7);
}

.sidebar-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.3);
  z-index: 199;
}

.fade-overlay-enter-active,
.fade-overlay-leave-active {
  transition: opacity 0.25s ease;
}
.fade-overlay-enter-from,
.fade-overlay-leave-to {
  opacity: 0;
}

/* Mobile: hidden by default, slide-in as drawer */
@media (max-width: 767px) {
  .admin-sidebar {
    transform: translateX(-100%);
  }
  .admin-sidebar.open {
    transform: translateX(0);
  }
}

/* PC: always visible */
@media (min-width: 768px) {
  .sidebar-overlay {
    display: none;
  }
}
</style>
