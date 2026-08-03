import axios from '@/utils/http'

// Competitions — production endpoints (/api/v1/admin/competitions).
export function getCompetitionList(params) {
  return axios.get('/api/v1/admin/competitions', { params }).then(res => res.data)
}

export function updateCompetition(id, data) {
  return axios.put(`/api/v1/admin/competitions/${id}`, data).then(res => res.data)
}

export function deleteCompetition(id) {
  return axios.delete(`/api/v1/admin/competitions/${id}`).then(res => res.data)
}
