/**
 * 医疗配送模块状态管理
 */
import { defineStore } from 'pinia'
import axios from '@/utils/http'

export const useMedicalStore = defineStore('medical', {
  state: () => ({
    // 认证状态
    certification: null,
    certificationLoading: false,

    // 起降场
    pads: [],
    padsLoading: false,

    // 常用联系人
    contacts: [],
    contactsLoading: false,

    // 订单
    orders: [],
    ordersLoading: false,
    ordersPagination: { page: 1, limit: 20, total: 0 },

    // 当前订单详情
    currentOrder: null,

    // 管理端数据
    adminOrders: [],
    adminOrdersPagination: { page: 1, limit: 20, total: 0, pendingCount: 0 },
    adminCertifications: [],
    adminCertsPagination: { page: 1, limit: 20, total: 0 },
    adminPads: []
  }),

  getters: {
    isApproved: (state) => state.certification?.status === 'approved',
    isPending: (state) => state.certification?.status === 'pending',
    isRejected: (state) => state.certification?.status === 'rejected',
    enabledPads: (state) => state.pads.filter(p => p.enabled)
  },

  actions: {
    // ==================== 认证相关 ====================
    async fetchCertificationStatus() {
      this.certificationLoading = true
      try {
        const res = await axios.get('/api/medical/certification/status')
        if (res.data?.success) {
          this.certification = res.data.data
        }
      } catch (error) {
        console.error('获取认证状态失败:', error)
      } finally {
        this.certificationLoading = false
      }
    },

    async submitCertification(data) {
      const res = await axios.post('/api/medical/certification/apply', data)
      if (res.data?.success) {
        await this.fetchCertificationStatus()
      }
      return res.data
    },

    async resubmitCertification(data) {
      const res = await axios.post('/api/medical/certification/resubmit', data)
      if (res.data?.success) {
        await this.fetchCertificationStatus()
      }
      return res.data
    },

    // ==================== 起降场相关 ====================
    async fetchPads() {
      this.padsLoading = true
      try {
        const res = await axios.get('/api/medical/pads')
        if (res.data?.success) {
          this.pads = res.data.data
        }
      } catch (error) {
        console.error('获取起降场失败:', error)
      } finally {
        this.padsLoading = false
      }
    },

    // ==================== 常用联系人 ====================
    async fetchContacts() {
      this.contactsLoading = true
      try {
        const res = await axios.get('/api/medical/contacts')
        if (res.data?.success) {
          this.contacts = res.data.data
        }
      } catch (error) {
        console.error('获取联系人失败:', error)
      } finally {
        this.contactsLoading = false
      }
    },

    async addContact(data) {
      const res = await axios.post('/api/medical/contacts', data)
      if (res.data?.success) {
        this.contacts.push(res.data.data)
      }
      return res.data
    },

    async updateContact(id, data) {
      const res = await axios.put(`/api/medical/contacts/${id}`, data)
      if (res.data?.success) {
        const index = this.contacts.findIndex(c => c.id === id)
        if (index !== -1) this.contacts[index] = res.data.data
      }
      return res.data
    },

    async deleteContact(id) {
      const res = await axios.delete(`/api/medical/contacts/${id}`)
      if (res.data?.success) {
        this.contacts = this.contacts.filter(c => c.id !== id)
      }
      return res.data
    },

    // ==================== 订单相关 ====================
    async fetchMyOrders(params = {}) {
      this.ordersLoading = true
      try {
        const res = await axios.get('/api/medical/orders/my', { params })
        if (res.data?.success) {
          this.orders = res.data.data
          this.ordersPagination = { page: res.data.page, limit: res.data.limit, total: res.data.total }
        }
      } catch (error) {
        console.error('获取订单列表失败:', error)
      } finally {
        this.ordersLoading = false
      }
    },

    async fetchOrderDetail(id) {
      const res = await axios.get(`/api/medical/orders/${id}`)
      if (res.data?.success) {
        this.currentOrder = res.data.data
      }
      return res.data
    },

    async createOrder(data) {
      const res = await axios.post('/api/medical/orders', data)
      return res.data
    },

    async cancelOrder(id, data) {
      const res = await axios.post(`/api/medical/orders/${id}/cancel`, data)
      return res.data
    },

    // 收件人订单列表
    async fetchReceivedOrders(params = {}) {
      const res = await axios.get('/api/medical/orders/received', { params })
      return res.data
    },

    // 收件人签收确认
    async confirmReceipt(id) {
      const res = await axios.post(`/api/medical/orders/${id}/confirm-receipt`)
      return res.data
    },

    async getReorderData(id) {
      const res = await axios.post(`/api/medical/orders/${id}/reorder`)
      return res.data
    },

    async getEstimate(params) {
      const res = await axios.get('/api/medical/orders/estimate', { params })
      return res.data
    },

    async rateOrder(id, data) {
      const res = await axios.post(`/api/medical/orders/${id}/rate`, data)
      return res.data
    },

    async getOrderRating(id) {
      const res = await axios.get(`/api/medical/orders/${id}/rating`)
      return res.data
    },

    // ==================== 管理端 ====================
    async fetchAdminOrders(params = {}) {
      this.ordersLoading = true
      try {
        const res = await axios.get('/api/medical/orders', { params })
        if (res.data?.success) {
          this.adminOrders = res.data.data
          this.adminOrdersPagination = {
            page: res.data.page, limit: res.data.limit,
            total: res.data.total, pendingCount: res.data.pendingCount
          }
        }
      } catch (error) {
        console.error('获取管理端订单失败:', error)
      } finally {
        this.ordersLoading = false
      }
    },

    async acceptOrder(id) {
      const res = await axios.post(`/api/medical/orders/${id}/accept`)
      return res.data
    },

    async pickupOrder(id) {
      const res = await axios.post(`/api/medical/orders/${id}/pickup`)
      return res.data
    },

    async deliverOrder(id) {
      const res = await axios.post(`/api/medical/orders/${id}/deliver`)
      return res.data
    },

    async deliveredOrder(id) {
      const res = await axios.post(`/api/medical/orders/${id}/delivered`)
      return res.data
    },

    async completeOrder(id) {
      const res = await axios.post(`/api/medical/orders/${id}/complete`)
      return res.data
    },

    async markException(id, data) {
      const res = await axios.post(`/api/medical/orders/${id}/exception`, data)
      return res.data
    },

    // 认证审核
    async fetchAdminCertifications(params = {}) {
      try {
        const res = await axios.get('/api/medical/certifications', { params })
        if (res.data?.success) {
          this.adminCertifications = res.data.data
          this.adminCertsPagination = { page: res.data.page, limit: res.data.limit, total: res.data.total }
        }
      } catch (error) {
        console.error('获取认证列表失败:', error)
      }
    },

    async approveCertification(id) {
      const res = await axios.post(`/api/medical/certifications/${id}/approve`)
      return res.data
    },

    async rejectCertification(id, reason) {
      const res = await axios.post(`/api/medical/certifications/${id}/reject`, { reason })
      return res.data
    },

    // 起降场管理
    async fetchAdminPads() {
      try {
        const res = await axios.get('/api/medical/pads/all')
        if (res.data?.success) {
          this.adminPads = res.data.data
        }
      } catch (error) {
        console.error('获取全部起降场失败:', error)
      }
    },

    async createPad(data) {
      const res = await axios.post('/api/medical/pads', data)
      if (res.data?.success) {
        this.adminPads.push(res.data.data)
      }
      return res.data
    },

    async updatePad(id, data) {
      const res = await axios.put(`/api/medical/pads/${id}`, data)
      if (res.data?.success) {
        const index = this.adminPads.findIndex(p => p.id === id)
        if (index !== -1) this.adminPads[index] = res.data.data
      }
      return res.data
    },

    async deletePad(id) {
      const res = await axios.delete(`/api/medical/pads/${id}`)
      if (res.data?.success) {
        this.adminPads = this.adminPads.filter(p => p.id !== id)
      }
      return res.data
    },

    // 评价统计
    async fetchRatingStats() {
      const res = await axios.get('/api/medical/ratings/stats')
      return res.data
    }
  }
})
