<template>
  <div class="user-list-page">
    <DataToolbar>
      <template #filters>
        <span class="toolbar-label">用户列表</span>
      </template>
      <template #actions>
        <van-button type="default" size="small" icon="replay" @click="fetchUsers">刷新</van-button>
      </template>
    </DataToolbar>

    <van-empty v-if="users.length === 0" description="暂无用户数据" />

    <van-cell-group v-else inset style="border-radius: var(--card-radius);">
      <van-cell v-for="u in users" :key="u.id">
        <template #title>
          <div style="display:flex; flex-direction: column; gap: 4px;">
            <div style="font-weight: 600; color: var(--text-color);">{{ u.name || '-' }}</div>
            <div style="font-size: 12px; color: var(--text-secondary);">{{ u.phone || '-' }}</div>
          </div>
        </template>
        <template #value>
          <div style="display:flex; flex-direction: column; align-items: flex-end; gap: 6px;">
            <van-tag
              :type="u.role === 'admin' ? 'success' : (u.role === 'dsl_admin' ? 'primary' : 'default')"
              size="medium"
            >
              {{ roleLabel(u.role) }}
            </van-tag>
            <van-button
              v-if="isSuperAdmin && u.role !== 'dsl_admin' && u.phone !== SUPER_ADMIN_PHONE"
              size="mini"
              type="primary"
              plain
              @click="toggleUserRole(u)"
            >
              切换权限
            </van-button>
          </div>
        </template>
      </van-cell>
    </van-cell-group>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from '@/utils/http'
import { showFailToast, showSuccessToast } from 'vant'
import DataToolbar from '../components/DataToolbar.vue'
import { useAuth } from '../composables/useAuth'

const { isSuperAdmin, SUPER_ADMIN_PHONE } = useAuth()

const users = ref([])

const roleLabel = (role) => {
  const map = { admin: '管理员', dsl_admin: 'DSL管理', user: '用户' }
  return map[role] || role || '-'
}

const fetchUsers = async () => {
  try {
    const res = await axios.get('/api/users')
    const raw = Array.isArray(res.data) ? res.data : (res.data?.data || res.data?.users || [])
    users.value = Array.isArray(raw) ? raw : []
  } catch (error) {
    showFailToast('获取用户数据失败')
    console.error(error)
  }
}

const toggleUserRole = async (user) => {
  if (!isSuperAdmin.value) {
    showFailToast('仅超级管理员可调整权限')
    return
  }
  if (user.phone === SUPER_ADMIN_PHONE) {
    showFailToast('超级管理员权限不可修改')
    return
  }
  const newRole = user.role === 'admin' ? 'user' : 'admin'
  try {
    await axios.post('/api/user/role', { id: user.id, role: newRole })
    user.role = newRole
    showSuccessToast('权限更新成功')
  } catch (error) {
    showFailToast('权限更新失败')
    console.error(error)
  }
}

onMounted(fetchUsers)
</script>

<style scoped>
.toolbar-label { font-size: 14px; font-weight: 500; color: var(--text-color); }
</style>
