import axios from '@/utils/http'

// Trade orders — production endpoints (/api/v1/admin/orders).
export function getOrderList(params) {
  return axios.get('/api/v1/admin/orders', { params }).then(res => res.data)
}

export function updateOrderStatus(id, status) {
  return axios.put(`/api/v1/admin/orders/${id}`, { status }).then(res => res.data)
}
