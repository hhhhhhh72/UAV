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
        :title="'无人机产业协会 · 管理后台'"
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
import { ref, onMounted } from 'vue'
import AdminSidebar from './components/AdminSidebar.vue'
import AdminHeader from './components/AdminHeader.vue'
import { useAuth } from './composables/useAuth'

const { userRole, isAdmin, isDslAdmin, isStudyAdmin, refreshCurrentUser } = useAuth()
const sidebarOpen = ref(false)

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
