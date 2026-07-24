<template>
  <aside class="admin-sidebar" :class="{ open: modelValue }">
    <div class="sidebar-brand">
      <span class="brand-text">后台管理</span>
    </div>

    <nav class="sidebar-nav">
      <router-link
        v-for="item in visibleMenus"
        :key="item.path"
        :to="item.path"
        class="nav-item"
        :class="{ active: isActive(item.path) }"
        @click="$emit('update:modelValue', false)"
      >
        <span class="nav-label">{{ item.label }}</span>
      </router-link>
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
  isAdmin: Boolean,
  isDslAdmin: Boolean,
  isStudyAdmin: Boolean
})

defineEmits(['update:modelValue'])

const route = useRoute()

const allMenus = [
  { path: '/admin/dashboard', label: '数据概览', icon: '◻', roles: ['admin', 'study_admin'] },
  { path: '/admin/orders', label: '订单管理', icon: '◎', roles: ['admin', 'study_admin'] },
  { path: '/admin/cases', label: '案例管理', icon: '◈', roles: ['admin'] },
  { path: '/admin/users', label: '用户管理', icon: '◉', roles: ['admin'] },
  { path: '/admin/competition', label: '赛事管理', icon: '◆', roles: ['admin', 'dsl_admin'] },
  { path: '/admin/config', label: '服务配置', icon: '◇', roles: ['admin', 'study_admin'] },
  { path: '/admin/reviews', label: '评价管理', icon: '★', roles: ['admin', 'study_admin'] },
  // 医疗配送管理
  { path: '/admin/medical/orders', label: '配送订单管理', icon: '✚', roles: ['admin'], group: '医疗配送' },
  { path: '/admin/medical/certifications', label: '认证审核', icon: '✔', roles: ['admin'], group: '医疗配送' },
  { path: '/admin/medical/pads', label: '起降场管理', icon: '▣', roles: ['admin'], group: '医疗配送' }
]

// 服务6(飞手培训)相关配置 - 研学管理员不可见
const trainingServiceIds = ['6']

const visibleMenus = computed(() => {
  return allMenus.filter(m => {
    if (props.isAdmin) return true
    if (props.isDslAdmin) return m.roles.includes('dsl_admin')
    if (props.isStudyAdmin) return m.roles.includes('study_admin')
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
