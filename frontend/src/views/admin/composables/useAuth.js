import { ref, computed } from 'vue'
import axios, { authStorage } from '@/utils/http'

export function useAuth() {
  const userStr = localStorage.getItem('user')
  const user = userStr ? JSON.parse(userStr) : null
  const userRole = ref(user ? user.role : 'user')
  // 当前登录账号 ID：列表里对自己不给操作入口（防自锁——把自己降级或删除，
  // 平台只有一名管理员时就再没人能进后台）。
  const currentUserId = ref(user && user.id ? user.id : '')

  // 「超级管理员」不再用前端猜：后端按 SUPER_ADMIN_PHONE 在用户列表里给出
  // is_super_admin 标记。这里只判断"当前登录人是不是平台管理员"。
  const isPlatformAdmin = computed(() => userRole.value === 'platform_admin')
  const isSuperAdmin = isPlatformAdmin // 兼容旧名（此前与 isPlatformAdmin 重复定义）
  const isAssociationAdmin = computed(() => userRole.value === 'association_admin')
  const canManage = computed(() => isPlatformAdmin.value || isAssociationAdmin.value)

  const refreshCurrentUser = async () => {
    const accessToken = authStorage.getAccessToken()
    if (!accessToken) return
    try {
      const res = await axios.get('/api/v1/me')
      if (res.data?.id) {
        const current = res.data
        localStorage.setItem('user', JSON.stringify(current))
        userRole.value = current.role || 'user'
        if (current.id) currentUserId.value = current.id
      }
    } catch (error) {
      // 仅在明确的认证失败(401)时清除登录状态
      // 网络错误、500等非认证问题不应导致登出
      if (error?.response?.status === 401) {
        authStorage.clearTokens()
        localStorage.removeItem('user')
      }
    }
  }

  return {
    userRole,
    currentUserId,
    isSuperAdmin,
    isPlatformAdmin,
    isAssociationAdmin,
    canManage,
    refreshCurrentUser
  }
}
