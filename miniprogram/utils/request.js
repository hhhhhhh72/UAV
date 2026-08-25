import { BASE_URL } from './config'
export { BASE_URL }

const ACCESS_TOKEN_KEY = 'accessToken'
const REFRESH_TOKEN_KEY = 'refreshToken'

// 确定性幂等键：POST/PATCH 自动附带，同 URL+body 的重试复用同 key，
// 服务端 24h 去重（防双击/网络重试重复创建）。用户修改内容 → 新 key。
function idempotencyKey(url, data) {
  const s = url + '|' + JSON.stringify(data || {})
  let h = 5381
  for (let i = 0; i < s.length; i++) h = ((h << 5) + h + s.charCodeAt(i)) | 0
  const hash = Math.abs(h).toString(36)
  return 'idem-' + hash + '-' + String(s.length).slice(0, 60)
}

// 补全后端返回的图片相对路径（/uploads/xxx → 完整域名）：培训/赛事/服务等
// 列表接口存的 image/poster 都是相对路径，小程序 <image> 直接渲染相对路径会
// 当本地资源 → 白图。仅在响应侧统一补全，提交给后端的数据保持相对路径不变。
// /uploads/private/（身份证影像等）不在业务卡片展示，跳过。
function resolveUploadsUrl(value) {
  if (typeof value === 'string') {
    if (value.indexOf('/uploads/') === 0 && value.indexOf('/uploads/private/') !== 0) {
      return BASE_URL + value
    }
    return value
  }
  if (Array.isArray(value)) {
    return value.map((v) => resolveUploadsUrl(v))
  }
  if (value && typeof value === 'object') {
    const out = {}
    for (const k of Object.keys(value)) {
      out[k] = resolveUploadsUrl(value[k])
    }
    return out
  }
  return value
}

// Unwrap the Go backend envelope. Paginated responses ({ data: [...], total, ... })
// keep their total attached to the array so list pages can read res.total.
function unwrap(body) {
  if (body && typeof body === 'object' && Array.isArray(body.data) && typeof body.total === 'number') {
    const data = resolveUploadsUrl(body.data)
    data.total = body.total
    return data
  }
  return resolveUploadsUrl(body?.data || body)
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
    // 写操作自动幂等键（不覆盖调用方显式指定的 key）
    const method = (options.method || 'GET').toUpperCase()
    if ((method === 'POST' || method === 'PATCH') && !header['Idempotency-Key']) {
      header['Idempotency-Key'] = idempotencyKey(options.url, options.data)
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
                      else if (retryRes.statusCode === 401) {
                        // 换了新 token 仍 401（账号被停用/权限变更/接口额外校验）：
                        // 登录态必须清理并跳登录，否则"假活"——页面显示已登录但所有接口不可用。
                        authStorage.clearTokens()
                        uni.removeStorageSync('user')
                        uni.navigateTo({ url: '/pages/login/index' })
                        rj(retryRes)
                      }
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
                else if (retryRes.statusCode === 401) {
                  // 同上：换新 token 仍 401 → 清理登录态并跳登录（防"假活"）。
                  authStorage.clearTokens()
                  uni.removeStorageSync('user')
                  uni.navigateTo({ url: '/pages/login/index' })
                  reject(retryRes)
                }
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
