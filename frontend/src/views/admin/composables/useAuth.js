import { ref, computed } from 'vue'
import axios, { authStorage } from '@/utils/http'

const SUPER_ADMIN_PHONE = 'drone-platform-admin'

export function useAuth() {
  const userStr = localStorage.getItem('user')
  const user = userStr ? JSON.parse(userStr) : null
  const userRole = ref(user ? user.role : 'user')
  // 登录只存 { id, role }，无 phone 字段：超级管理员以角色 platform_admin 判定
  // （此前用 phone === 'drone-platform-admin' 恒为 false，导致角色管理按钮不可见）
  const isSuperAdmin = ref(user ? user.role === 'platform_admin' : false)

  const isPlatformAdmin = computed(() => userRole.value === 'platform_admin')
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
        isSuperAdmin.value = current.role === 'platform_admin'
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
    isSuperAdmin,
    isPlatformAdmin,
    isAssociationAdmin,
    canManage,
    refreshCurrentUser,
    SUPER_ADMIN_PHONE
  }
}
