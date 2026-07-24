/**
 * 用户状态管理
 */
import { defineStore } from 'pinia'
import axios from '@/utils/http'

export const useUserStore = defineStore('user', {
  state: () => ({
    user: JSON.parse(localStorage.getItem('user') || 'null'),
    accessToken: localStorage.getItem('accessToken') || null,
    refreshToken: localStorage.getItem('refreshToken') || null
  }),

  getters: {
    // 是否已登录
    isLoggedIn: (state) => !!state.accessToken,

    // 是否为超级管理员
    isSuperAdmin: (state) => state.user?.phone === 'wzdkjjfzyxgs',

    // 是否为管理员
    isAdmin: (state) => state.user?.role === 'admin',

    // 是否为DSL管理员
    isDslAdmin: (state) => state.user?.role === 'dsl_admin',

    // 是否为研学管理员
    isStudyAdmin: (state) => state.user?.role === 'study_admin',

    // 是否有管理权限
    canManage: (state) => ['admin', 'dsl_admin', 'study_admin'].includes(state.user?.role),

    // 用户显示名称
    displayName: (state) => state.user?.name || state.user?.phone || '用户'
  },

  actions: {
    /**
     * 登录
     */
    async login(phone, password) {
      try {
        const res = await axios.post('/api/auth/login', { phone, password })

        if (res.data?.success) {
          this.setUser(res.data.user, res.data.accessToken, res.data.refreshToken)
          return { success: true }
        }

        return {
          success: false,
          message: res.data?.message || '登录失败'
        }
      } catch (error) {
        return {
          success: false,
          message: error.response?.data?.message || '网络错误'
        }
      }
    },

    /**
     * 注册
     */
    async register(phone, password, name) {
      try {
        const res = await axios.post('/api/auth/register', { phone, password, name })

        if (res.data?.success) {
          return { success: true }
        }

        return {
          success: false,
          message: res.data?.message || '注册失败'
        }
      } catch (error) {
        return {
          success: false,
          message: error.response?.data?.message || '网络错误'
        }
      }
    },

    /**
     * 登出
     */
    async logout() {
      try {
        await axios.post('/api/auth/logout')
      } catch (error) {
        // 即使服务器错误,也要清除本地状态
        console.warn('Logout error:', error)
      } finally {
        this.clearUser()
      }
    },

    /**
     * 刷新用户信息
     */
    async fetchUser() {
      if (!this.accessToken) {
        return
      }

      try {
        const res = await axios.get('/api/auth/me')

        if (res.data?.success) {
          this.user = res.data.user
          localStorage.setItem('user', JSON.stringify(this.user))
        }
      } catch (error) {
        if (error.response?.status === 401) {
          this.clearUser()
        }
        throw error
      }
    },

    /**
     * 设置用户信息
     */
    setUser(user, accessToken, refreshToken) {
      this.user = user
      this.accessToken = accessToken
      this.refreshToken = refreshToken

      localStorage.setItem('user', JSON.stringify(user))
      localStorage.setItem('accessToken', accessToken)
      localStorage.setItem('refreshToken', refreshToken)
    },

    /**
     * 清除用户信息
     */
    clearUser() {
      this.user = null
      this.accessToken = null
      this.refreshToken = null

      localStorage.removeItem('user')
      localStorage.removeItem('accessToken')
      localStorage.removeItem('refreshToken')
    },

    /**
     * 更新用户资料
     */
    async updateProfile(profile) {
      try {
        const res = await axios.post('/api/user/update', profile)

        if (res.data?.success) {
          // 更新本地状态
          if (this.user) {
            this.user = { ...this.user, ...profile }
            localStorage.setItem('user', JSON.stringify(this.user))
          }
          return { success: true }
        }

        return {
          success: false,
          message: res.data?.message || '更新失败'
        }
      } catch (error) {
        return {
          success: false,
          message: error.response?.data?.message || '网络错误'
        }
      }
    }
  }
})
