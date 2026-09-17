import axios from '@/utils/http'

// Trade orders — production endpoints (/api/v1/admin/orders).
export function getOrderList(params) {
  return axios.get('/api/v1/admin/orders', { params }).then(res => res.data)
}

export function updateOrderStatus(id, status) {
  return axios.put(`/api/v1/admin/orders/${id}`, { status }).then(res => res.data)
}

// 售后单审核：approve=同意退款 / reject=驳回（仅 aftersale_status=pending 可审）
// 注意：退货退款(return)单的 approve 只是"同意退货"——买家还需寄回、卖家确认收到后才退款。
export function reviewAftersale(id, action) {
  return axios.put(`/api/v1/admin/orders/${id}/aftersale`, { action }).then(res => res.data)
}

// 退货退款流程收尾：确认收到买家寄回的商品 → 此刻才发起退款（仅 aftersale_status=returned 可确认）
export function confirmReturnReceived(id) {
  return axios.put(`/api/v1/admin/orders/${id}/aftersale/confirm-return`).then(res => res.data)
}
