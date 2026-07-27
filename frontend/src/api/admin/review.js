import axios from '@/utils/http'

export function getReviewList(params) {
  return axios.get('/api/admin/reviews', { params }).then(res => res.data)
}

export function updateReviewStatus(id, status) {
  return axios.post(`/api/admin/reviews/${id}`, { status }).then(res => res.data)
}

export function deleteReview(id) {
  return axios.delete(`/api/admin/reviews/${id}`).then(res => res.data)
}
