import axios from 'axios'

const ACCESS_TOKEN_KEY = 'accessToken'
const REFRESH_TOKEN_KEY = 'refreshToken'

let isRefreshing = false
let pendingQueue = []

const resolveQueue = (token) => {
  pendingQueue.forEach(({ resolve }) => resolve(token))
  pendingQueue = []
}

const rejectQueue = (error) => {
  pendingQueue.forEach(({ reject }) => reject(error))
  pendingQueue = []
}

axios.interceptors.request.use((config) => {
  const token = localStorage.getItem(ACCESS_TOKEN_KEY)
  if (token) {
    config.headers = config.headers || {}
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// Unwrap Go backend { data: {...} } envelope transparently.
// Paginated responses ({ data, total, page, page_size }) are kept intact
// so the frontend can access pagination metadata.
axios.interceptors.response.use(
  (response) => {
    const body = response.data
    if (body && typeof body === 'object' && 'data' in body) {
      if ('total' in body) {
        // Paginated response — keep full structure for { data, total, page, page_size }
        response.data = body
      } else {
        response.data = body.data
      }
    }
    return response
  },
  async (error) => {
    const originalRequest = error.config
    if (!originalRequest || error.response?.status !== 401) {
      return Promise.reject(error)
    }

    // refresh 请求本身返回 401 时，不再重试，直接清除登录态
    if (originalRequest.url === '/api/auth/refresh') {
      localStorage.removeItem(ACCESS_TOKEN_KEY)
      localStorage.removeItem(REFRESH_TOKEN_KEY)
      localStorage.removeItem('user')
      if (window.location.pathname.startsWith('/admin')) {
        window.location.href = '/login'
      }
      return Promise.reject(error)
    }

    if (originalRequest._retry) {
      return Promise.reject(error)
    }
    originalRequest._retry = true

    const refreshToken = localStorage.getItem(REFRESH_TOKEN_KEY)
    if (!refreshToken) {
      localStorage.removeItem(ACCESS_TOKEN_KEY)
      localStorage.removeItem('user')
      return Promise.reject(error)
    }

    if (isRefreshing) {
      return new Promise((resolve, reject) => {
        pendingQueue.push({
          resolve: (token) => {
            originalRequest.headers.Authorization = `Bearer ${token}`
            resolve(axios(originalRequest))
          },
          reject
        })
      })
    }

    isRefreshing = true
    try {
      // P0 修复：后端 /api/auth/refresh 契约是 snake_case——
      // 请求体 refresh_token、响应 access_token / refresh_token。轮转后必须持久化
      // 新 refresh_token（旧令牌已被服务端 Revoke，不存则第二次刷新必 401）。
      const refreshRes = await axios.post('/api/auth/refresh', { refresh_token: refreshToken }, { timeout: 10000 })
      const data = refreshRes.data || {}
      const newAccessToken = data.access_token
      if (!newAccessToken) {
        throw new Error('No access token returned')
      }
      localStorage.setItem(ACCESS_TOKEN_KEY, newAccessToken)
      const newRefreshToken = data.refresh_token
      if (newRefreshToken) {
        localStorage.setItem(REFRESH_TOKEN_KEY, newRefreshToken)
      }
      resolveQueue(newAccessToken)
      originalRequest.headers.Authorization = `Bearer ${newAccessToken}`
      return axios(originalRequest)
    } catch (refreshError) {
      // 关键：pendingQueue 必须先 reject，否则等待中的请求会永久挂起
      rejectQueue(refreshError)

      // 仅当 refresh 明确返回 401（后端对无效/过期 refresh_token 返回 401）才清除登录态；
      // 网络错误 / 5xx 保留 token，避免误登出。
      if (refreshError?.response?.status === 401) {
        localStorage.removeItem(ACCESS_TOKEN_KEY)
        localStorage.removeItem(REFRESH_TOKEN_KEY)
        localStorage.removeItem('user')
        if (window.location.pathname.startsWith('/admin')) {
          window.location.href = '/login'
        }
      } else {
        console.error('[http] refresh token 失败（保留登录态）:', refreshError?.message || refreshError)
      }
      return Promise.reject(refreshError)
    } finally {
      isRefreshing = false
    }
  }
)

export const authStorage = {
  getAccessToken() {
    return localStorage.getItem(ACCESS_TOKEN_KEY)
  },
  getRefreshToken() {
    return localStorage.getItem(REFRESH_TOKEN_KEY)
  },
  setTokens(accessToken, refreshToken) {
    if (accessToken) localStorage.setItem(ACCESS_TOKEN_KEY, accessToken)
    if (refreshToken) localStorage.setItem(REFRESH_TOKEN_KEY, refreshToken)
  },
  clearTokens() {
    localStorage.removeItem(ACCESS_TOKEN_KEY)
    localStorage.removeItem(REFRESH_TOKEN_KEY)
  }
}

/**
 * 上传等非 axios 拦截器场景下，动态构造带最新 accessToken 的 Authorization 头。
 * 在调用时刻读取 localStorage，避免组件创建时快照导致 token 轮转/过期后仍用旧值。
 */
export function getAuthHeader() {
  const token = localStorage.getItem(ACCESS_TOKEN_KEY)
  return token ? { Authorization: `Bearer ${token}` } : {}
}

export default axios

