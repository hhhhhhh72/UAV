/**
 * 应用/订单状态管理
 */
import { defineStore } from 'pinia'
import axios from '@/utils/http'

export const useApplicationStore = defineStore('application', {
  state: () => ({
    applications: [],
    currentApplication: null,
    loading: false,
    pagination: {
      page: 1,
      limit: 20,
      total: 0,
      totalPages: 0
    },
    filters: {
      status: null,
      serviceId: null
    }
  }),

  getters: {
    // 待处理的申请
    pendingApplications: (state) =>
      state.applications.filter(app => app.status === '待处理'),

    // 处理中的申请
    processingApplications: (state) =>
      state.applications.filter(app => app.status === '处理中'),

    // 已完成的申请
    completedApplications: (state) =>
      state.applications.filter(app => app.status === '已完成'),

    // 是否有数据
    hasData: (state) => state.applications.length > 0
  },

  actions: {
    /**
     * 获取申请列表
     */
    async fetchApplications(params = {}) {
      this.loading = true

      try {
        const { page = 1, limit = 20, status, serviceId } = params

        const res = await axios.get('/api/admin/applications', {
          params: { page, limit, status, serviceId }
        })

        if (res.data?.success) {
          this.applications = res.data.data
          this.pagination = res.data.pagination
          this.filters = { status, serviceId }
        }
      } catch (error) {
        console.error('Failed to fetch applications:', error)
        throw error
      } finally {
        this.loading = false
      }
    },

    /**
     * 获取单个申请详情
     */
    async fetchApplication(id) {
      this.loading = true

      try {
        // 注意: 需要后端支持 GET /api/admin/applications/:id
        const res = await axios.get(`/api/admin/applications/${id}`)

        if (res.data?.success) {
          this.currentApplication = res.data.data
        }
      } catch (error) {
        console.error('Failed to fetch application:', error)
        throw error
      } finally {
        this.loading = false
      }
    },

    /**
     * 更新申请状态
     */
    async updateApplicationStatus(id, status, remark) {
      try {
        const res = await axios.post(`/api/admin/applications/${id}`, {
          status,
          remark
        })

        if (res.data?.success) {
          // 更新列表中的申请
          const index = this.applications.findIndex(app => app.id === id)
          if (index !== -1) {
            this.applications[index] = {
              ...this.applications[index],
              ...res.data.data
            }
          }

          // 更新当前申请
          if (this.currentApplication?.id === id) {
            this.currentApplication = res.data.data
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
    },

    /**
     * 导出申请列表
     */
    async exportApplications() {
      try {
        const response = await axios.get('/api/admin/applications/export', {
          responseType: 'blob'
        })

        // 创建下载链接
        const url = window.URL.createObjectURL(new Blob([response.data]))
        const link = document.createElement('a')
        link.href = url
        link.setAttribute('download', `applications_${Date.now()}.xlsx`)
        document.body.appendChild(link)
        link.click()
        link.remove()
        window.URL.revokeObjectURL(url)

        return { success: true }
      } catch (error) {
        console.error('Failed to export applications:', error)
        return {
          success: false,
          message: '导出失败'
        }
      }
    },

    /**
     * 清空状态
     */
    clear() {
      this.applications = []
      this.currentApplication = null
      this.pagination = {
        page: 1,
        limit: 20,
        total: 0,
        totalPages: 0
      }
      this.filters = {
        status: null,
        serviceId: null
      }
    }
  }
})
