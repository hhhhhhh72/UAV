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

axios.interceptors.response.use(
  (response) => response,
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
      if (window.location.pathname.startsWith('/medical') || window.location.pathname.startsWith('/admin')) {
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
      if (window.location.pathname.startsWith('/medical')) {
        window.location.href = '/login'
      }
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
      const refreshRes = await axios.post('/api/auth/refresh', { refreshToken })
      const newAccessToken = refreshRes.data?.accessToken
      if (!newAccessToken) {
        throw new Error('No access token returned')
      }
      localStorage.setItem(ACCESS_TOKEN_KEY, newAccessToken)
      resolveQueue(newAccessToken)
      originalRequest.headers.Authorization = `Bearer ${newAccessToken}`
      return axios(originalRequest)
    } catch (refreshError) {
      rejectQueue(refreshError)
      localStorage.removeItem(ACCESS_TOKEN_KEY)
      localStorage.removeItem(REFRESH_TOKEN_KEY)
      localStorage.removeItem('user')
      if (window.location.pathname.startsWith('/medical') || window.location.pathname.startsWith('/admin')) {
        window.location.href = '/login'
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

export default axios

