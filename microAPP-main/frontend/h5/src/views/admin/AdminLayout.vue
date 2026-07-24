<template>
  <div class="admin-layout">
    <AdminSidebar
      v-model="sidebarOpen"
      :is-admin="isAdmin"
      :is-dsl-admin="isDslAdmin"
      :is-study-admin="isStudyAdmin"
    />

    <div class="admin-main">
      <AdminHeader
        :title="pageTitle"
        :user-role="userRole"
        @toggle-sidebar="sidebarOpen = !sidebarOpen"
      />
      <main class="admin-content">
        <router-view />
      </main>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import AdminSidebar from './components/AdminSidebar.vue'
import AdminHeader from './components/AdminHeader.vue'
import { useAuth } from './composables/useAuth'

const { userRole, isAdmin, isDslAdmin, isStudyAdmin, refreshCurrentUser } = useAuth()
const sidebarOpen = ref(false)
const route = useRoute()

const titleMap = {
  '/admin/dashboard': '数据概览',
  '/admin/orders': '订单管理',
  '/admin/cases': '案例管理',
  '/admin/users': '用户管理',
  '/admin/competition': '赛事管理',
  '/admin/config': '服务配置',
  '/admin/reviews': '评价管理',
  '/admin/medical/orders': '配送订单管理',
  '/admin/medical/certifications': '认证审核',
  '/admin/medical/pads': '起降场管理'
}

const pageTitle = computed(() => titleMap[route.path] || '后台管理')

onMounted(() => {
  refreshCurrentUser()
})
</script>

<style scoped>
.admin-layout {
  display: flex;
  min-height: 100vh;
  background: var(--background-color, #f5f5f7);
}

.admin-main {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
}

.admin-content {
  flex: 1;
  padding: 20px;
  overflow-y: auto;
}

/* PC: offset main by sidebar width */
@media (min-width: 768px) {
  .admin-main {
    margin-left: var(--sidebar-width, 220px);
  }
}

/* Mobile: full width */
@media (max-width: 767px) {
  .admin-content {
    padding: 12px;
  }
}
</style>
