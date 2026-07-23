<template>
  <header class="admin-header">
    <div class="header-left">
      <button class="menu-btn" @click="$emit('toggle-sidebar')">
        <span class="menu-icon">&#9776;</span>
      </button>
      <h1 class="page-title">{{ title }}</h1>
    </div>
    <div class="header-right">
      <span class="user-name">{{ userName }}</span>
      <span class="user-role">{{ roleLabel }}</span>
    </div>
  </header>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  title: { type: String, default: '' },
  userRole: { type: String, default: 'user' }
})

defineEmits(['toggle-sidebar'])

const userStr = localStorage.getItem('user')
const user = userStr ? JSON.parse(userStr) : null
const userName = user?.name || user?.phone || '管理员'

const roleLabel = computed(() => {
  const map = { admin: '管理员', dsl_admin: 'DSL管理', user: '用户' }
  return map[props.userRole] || props.userRole
})
</script>

<style scoped>
.admin-header {
  height: var(--admin-header-height, 56px);
  background: var(--bg-primary, #fff);
  border-bottom: 1px solid var(--border-color, #e5e5e7);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  position: sticky;
  top: 0;
  z-index: 100;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.menu-btn {
  display: none;
  background: none;
  border: none;
  cursor: pointer;
  padding: 6px;
  border-radius: 6px;
  color: var(--text-color, #1d1d1f);
  font-size: 18px;
  line-height: 1;
}

.menu-btn:hover {
  background: rgba(0, 0, 0, 0.05);
}

.page-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-color, #1d1d1f);
  margin: 0;
  letter-spacing: -0.2px;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.user-name {
  font-size: 13px;
  color: var(--text-color, #1d1d1f);
  font-weight: 500;
}

.user-role {
  font-size: 11px;
  color: var(--text-secondary, #86868b);
  background: var(--background-color, #f5f5f7);
  padding: 2px 8px;
  border-radius: 10px;
}

@media (max-width: 767px) {
  .menu-btn {
    display: block;
  }
  .admin-header {
    padding: 0 12px;
  }
}
</style>
