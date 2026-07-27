import axios from '@/utils/http'

/**
 * 需求管理 - API 模块
 */

export function getDemandList(params) {
  return axios.get('/api/v1/admin/demands', { params }).then(res => res.data)
}

export function approveDemand(id) {
  return axios.post(`/api/v1/admin/demands/${id}/approve`).then(res => res.data)
}

export function rejectDemand(id, reason) {
  return axios.post(`/api/v1/admin/demands/${id}/review`, { action: 'reject', reason }).then(res => res.data)
}
