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

export function closeDemand(id, reason) {
  return axios.post(`/api/v1/admin/demands/${id}/close`, { reason }).then(res => res.data)
}

export function setOfflineAmount(id, offlineAmountFen) {
  return axios.post(`/api/v1/admin/demands/${id}/amount`, { offline_amount_fen: offlineAmountFen }).then(res => res.data)
}

export function deleteDemand(id) {
  return axios.delete(`/api/v1/admin/demands/${id}`).then(res => res.data)
}

export function rejectDemand(id, reason) {
  return axios.post(`/api/v1/admin/demands/${id}/review`, { action: 'reject', reason }).then(res => res.data)
}
