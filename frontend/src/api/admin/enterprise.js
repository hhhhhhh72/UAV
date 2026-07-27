import axios from '@/utils/http'

/**
 * 企业审核 - API 模块
 * 对应的 Go 后端路由: /api/v1/admin/enterprises, /api/v1/enterprises
 */

/**
 * 获取企业列表（管理后台）
 * @param {Object} params - { page, page_size, status, keyword }
 * @returns {Promise<{ data: Array, total: number, page: number, page_size: number }>}
 */
export function getEnterpriseList(params) {
  return axios.get('/api/v1/admin/enterprises', { params }).then(res => res.data)
}

/**
 * 审核企业（通过/驳回）
 * @param {string} id - 企业 ID
 * @param {Object} data - { action: 'approved'|'rejected', reason: string }
 */
export function reviewEnterprise(id, data) {
  return axios.post(`/api/v1/admin/enterprises/${id}/review`, data).then(res => res.data)
}

/**
 * 批量审核企业
 * @param {string} action - 'approved' | 'rejected'
 * @param {Object} data - { ids: string[] }
 */
export function batchReviewEnterprise(action, data) {
  return axios.post('/api/v1/admin/enterprises/batch-review', { action, ...data }).then(res => res.data)
}
