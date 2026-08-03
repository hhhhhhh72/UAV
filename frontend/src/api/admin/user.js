import axios from '@/utils/http'

export function getUserList(params) {
  return axios.get('/api/v1/admin/users', { params }).then(res => res.data)
}

export function updateUserRole(id, role) {
  return axios.post(`/api/v1/admin/users/${id}/role`, { role }).then(res => res.data)
}
