import axios from '@/utils/http'

// Trade orders — production endpoints (/api/v1/admin/orders).
export function getOrderList(params) {
  return axios.get('/api/v1/admin/orders', { params }).then(res => res.data)
}

export function updateOrderStatus(id, status) {
  return axios.put(`/api/v1/admin/orders/${id}`, { status }).then(res => res.data)
}

// 售后单审核：approve=同意退款 / reject=驳回（仅 aftersale+pending 可审）
export function reviewAftersale(id, action) {
  return axios.put(`/api/v1/admin/orders/${id}/aftersale`, { action }).then(res => res.data)
}
