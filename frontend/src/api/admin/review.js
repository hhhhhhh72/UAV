import axios from '@/utils/http'

export function getReviewList(params) {
  return axios.get('/api/v1/admin/reviews', { params }).then(res => res.data)
}

export function updateReviewStatus(id, status) {
  const action = status === 'approved' ? 'approve' : 'reject'
  return axios.post(`/api/v1/admin/reviews/${id}/${action}`, {}).then(res => res.data)
}

export function deleteReview(id) {
  return axios.delete(`/api/v1/admin/reviews/${id}`).then(res => res.data)
}
