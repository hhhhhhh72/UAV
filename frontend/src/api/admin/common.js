/**
 * 后台管理通用 CRUD API 工具
 * 用法: import { useAdminApi } from '@/api/admin/common'
 *       const api = useAdminApi('training-courses')
 *       api.list(params)  // GET  /api/v1/admin/training-courses
 *       api.get(id)       // GET  /api/v1/admin/training-courses/{id}
 *       api.create(data)  // POST /api/v1/admin/training-courses
 *       api.update(id, data) // PUT /api/v1/admin/training-courses/{id}
 *       api.delete(id)    // DELETE /api/v1/admin/training-courses/{id}
 */
import axios from '@/utils/http'

export function useAdminApi(resource) {
  const base = `/api/v1/admin/${resource}`

  return {
    list(params) {
      return axios.get(base, { params }).then(r => r.data)
    },
    get(id) {
      return axios.get(`${base}/${id}`).then(r => r.data)
    },
    create(data) {
      return axios.post(base, data).then(r => r.data)
    },
    update(id, data) {
      return axios.put(`${base}/${id}`, data).then(r => r.data)
    },
    delete(id) {
      return axios.delete(`${base}/${id}`).then(r => r.data)
    }
  }
}
