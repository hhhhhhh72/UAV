import { BASE_URL } from './config'
export { BASE_URL }

const ACCESS_TOKEN_KEY = 'accessToken'
const REFRESH_TOKEN_KEY = 'refreshToken'

// Unwrap the Go backend envelope. Paginated responses ({ data: [...], total, ... })
// keep their total attached to the array so list pages can read res.total.
function unwrap(body) {
  if (body && typeof body === 'object' && Array.isArray(body.data) && typeof body.total === 'number') {
    body.data.total = body.total
    return body.data
  }
  return body?.data || body
}

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

export const authStorage = {
  getAccessToken() {
    return uni.getStorageSync(ACCESS_TOKEN_KEY) || ''
  },
  getRefreshToken() {
    return uni.getStorageSync(REFRESH_TOKEN_KEY) || ''
  },
  setTokens(accessToken, refreshToken) {
    if (accessToken) uni.setStorageSync(ACCESS_TOKEN_KEY, accessToken)
    if (refreshToken) uni.setStorageSync(REFRESH_TOKEN_KEY, refreshToken)
  },
  clearTokens() {
    uni.removeStorageSync(ACCESS_TOKEN_KEY)
    uni.removeStorageSync(REFRESH_TOKEN_KEY)
  }
}

async function refreshAccessToken() {
  const refreshToken = authStorage.getRefreshToken()
  if (!refreshToken) throw new Error('No refresh token')

  const res = await new Promise((resolve, reject) => {
    uni.request({
      url: BASE_URL + '/api/v1/auth/refresh',
      method: 'POST',
      data: { refresh_token: refreshToken },
      success: (r) => {
        if (r.statusCode >= 200 && r.statusCode < 300) resolve(r.data)
        else reject(r)
      },
      fail: reject
    })
  })

  const body = res?.data ?? res
  const newAccessToken = body?.access_token || body?.accessToken
  if (!newAccessToken) throw new Error('No access token returned')
  // Refresh tokens rotate on the server: the old one is revoked immediately,
  // so the new refresh token must be persisted, or the next refresh 401s.
  const newRefreshToken = body?.refresh_token || body?.refreshToken
  authStorage.setTokens(newAccessToken, newRefreshToken)
  return newAccessToken
}

export function request(options) {
  return new Promise((resolve, reject) => {
    const token = authStorage.getAccessToken()
    const header = { ...(options.header || {}) }
    if (token) {
      header['Authorization'] = `Bearer ${token}`
    }

    uni.request({
      url: BASE_URL + options.url,
      method: options.method || 'GET',
      data: options.data || {},
      header,
      success: async (res) => {
        if (res.statusCode >= 200 && res.statusCode < 300) {
          resolve(unwrap(res.data))
        } else if (res.statusCode === 401) {
          const refreshToken = authStorage.getRefreshToken()
          if (!refreshToken) {
            authStorage.clearTokens()
            uni.removeStorageSync('user')
            uni.navigateTo({ url: '/pages/login/index' })
            return reject(res)
          }

          if (isRefreshing) {
            return new Promise((r, rj) => {
              pendingQueue.push({
                resolve: (newToken) => {
                  header['Authorization'] = `Bearer ${newToken}`
                  uni.request({
                    url: BASE_URL + options.url,
                    method: options.method || 'GET',
                    data: options.data || {},
                    header,
                    success: (retryRes) => {
                      if (retryRes.statusCode >= 200 && retryRes.statusCode < 300) r(unwrap(retryRes.data))
                      else rj(retryRes)
                    },
                    fail: rj
                  })
                },
                reject: rj
              })
            }).then(resolve).catch(reject)
          }

          isRefreshing = true
          try {
            const newToken = await refreshAccessToken()
            resolveQueue(newToken)
            header['Authorization'] = `Bearer ${newToken}`
            uni.request({
              url: BASE_URL + options.url,
              method: options.method || 'GET',
              data: options.data || {},
              header,
              success: (retryRes) => {
                if (retryRes.statusCode >= 200 && retryRes.statusCode < 300) resolve(unwrap(retryRes.data))
                else reject(retryRes)
              },
              fail: reject
            })
          } catch (refreshError) {
            rejectQueue(refreshError)
            authStorage.clearTokens()
            uni.removeStorageSync('user')
            uni.navigateTo({ url: '/pages/login/index' })
            reject(refreshError)
          } finally {
            isRefreshing = false
          }
        } else {
          reject(res)
        }
      },
      fail: (err) => {
        reject(err)
      }
    })
  })
}

export function getStoredUser() {
  try {
    const userStr = uni.getStorageSync('user')
    return userStr ? JSON.parse(userStr) : null
  } catch (e) {
    return null
  }
}

// 提取后端统一错误包 { error: { code, message } } 中的中文提示
export function getErrorMessage(e) {
  try {
    if (e && e.data && e.data.error && e.data.error.message) return e.data.error.message
    if (e && e.message) return e.message
  } catch { /* ignore */ }
  return ''
}
