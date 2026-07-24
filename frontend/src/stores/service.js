/**
 * 服务配置状态管理
 */
import { defineStore } from 'pinia'
import axios from '@/utils/http'

export const useServiceStore = defineStore('service', {
  state: () => ({
    services: {},
    loading: false,
    error: null
  }),

  getters: {
    // 获取特定服务配置
    getServiceConfig: (state) => (serviceId) => {
      return state.services[serviceId] || null
    },

    // 获取所有服务列表
    serviceList: (state) => {
      return Object.entries(state.services).map(([id, config]) => ({
        id,
        ...config
      }))
    },

    // 获取研学展示数据
    studyShowcase: (state) => {
      return state.services['9']?.studyShowcase || []
    }
  },

  actions: {
    /**
     * 获取服务配置
     */
    async fetchServices() {
      this.loading = true
      this.error = null

      try {
        const res = await axios.get('/api/admin/services/config')

        if (res.data?.success) {
          this.services = res.data.data
        } else {
          this.error = res.data?.message || '获取配置失败'
        }
      } catch (error) {
        this.error = error.response?.data?.message || '网络错误'
        console.error('Failed to fetch services:', error)
      } finally {
        this.loading = false
      }
    },

    /**
     * 更新服务配置
     */
    async updateServices(newConfig) {
      this.loading = true
      this.error = null

      try {
        const res = await axios.post('/api/admin/services/config', newConfig)

        if (res.data?.success) {
          // 更新本地状态
          this.services = {
            ...this.services,
            ...newConfig
          }
          return { success: true }
        }

        return {
          success: false,
          message: res.data?.message || '更新失败'
        }
      } catch (error) {
        this.error = error.response?.data?.message || '网络错误'
        return {
          success: false,
          message: this.error
        }
      } finally {
        this.loading = false
      }
    },

    /**
     * 获取研学展示数据
     */
    async fetchStudyShowcase() {
      this.loading = true

      try {
        const res = await axios.get('/api/admin/study/showcase')

        if (res.data?.success) {
          // 更新服务9的配置
          if (this.services['9']) {
            this.services['9'].studyShowcase = res.data.data
          } else {
            this.services['9'] = { studyShowcase: res.data.data }
          }

          return res.data.data
        }
      } catch (error) {
        console.error('Failed to fetch study showcase:', error)
        throw error
      } finally {
        this.loading = false
      }
    },

    /**
     * 更新研学展示数据
     */
    async updateStudyShowcase(showcase) {
      this.loading = true

      try {
        const res = await axios.post('/api/admin/study/showcase', { showcase })

        if (res.data?.success) {
          // 更新本地状态
          if (this.services['9']) {
            this.services['9'].studyShowcase = showcase
          } else {
            this.services['9'] = { studyShowcase: showcase }
          }

          return { success: true }
        }

        return {
          success: false,
          message: res.data?.message || '更新失败'
        }
      } catch (error) {
        console.error('Failed to update study showcase:', error)
        return {
          success: false,
          message: error.response?.data?.message || '网络错误'
        }
      } finally {
        this.loading = false
      }
    },

    /**
     * 清空状态
     */
    clear() {
      this.services = {}
      this.loading = false
      this.error = null
    }
  }
})
