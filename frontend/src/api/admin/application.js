import axios from '@/utils/http'

export function getApplicationList(params) {
  return axios.get('/api/list', { params }).then(res => res.data)
}

export function updateApplicationStatus(id, status) {
  return axios.post('/api/update', { id, status }).then(res => res.data)
}

export function exportApplications(params) {
  const qs = Object.entries(params).filter(([,v]) => v).map(([k,v]) => `${k}=${encodeURIComponent(v)}`).join('&')
  return `/api/export?${qs}`
}
