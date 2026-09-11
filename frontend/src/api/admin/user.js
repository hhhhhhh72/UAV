import axios from '@/utils/http'

export function getUserList(params) {
  return axios.get('/api/v1/admin/users', { params }).then(res => res.data)
}

export function updateUserRole(id, role) {
  return axios.post(`/api/v1/admin/users/${id}/role`, { role }).then(res => res.data)
}

// 删除账号（唯一动作，不可恢复）：账号立即失效并从平台消失，账号行保留 7 天缓冲期，
// 到期由后台任务自动物理清除；其发布的内容一律保留（后端口径见 service.UserService.DeleteUser）。
// 平台管理员重置他人密码（无需旧密码；重置后该账号所有登录失效）
export function resetUserPassword(id, password) {
  return axios.post(`/api/v1/admin/users/${encodeURIComponent(id)}/password`, { password }).then(res => res.data)
}

export function deleteUser(id) {
  return axios.delete(`/api/v1/admin/users/${encodeURIComponent(id)}`).then(res => res.data)
}
