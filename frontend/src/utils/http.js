import axios from 'axios'

// 安全加固（P1）：会话令牌从 localStorage 降级为 sessionStorage + 内存态——
// localStorage 持久到浏览器整个生命周期，任何同源 XSS 可长期窃取（含 7 天 refreshToken）；
// sessionStorage 随标签页关闭即清，令牌窗口收窄为单次会话。生产可进一步将 refresh
// 移到 httpOnly+Secure+SameSite Cookie（后端改造），此处为前端纵深防御下限。

// 内存态：页面内最快读取，不依赖存储（登录后即写，刷新页面前从 sessionStorage 恢复）
let memAccessToken = ''
let memRefreshToken = ''

const ACCESS_TOKEN_KEY = 'accessToken'
const REFRESH_TOKEN_KEY = 'refreshToken'

const MEM_SESSION = '__uav_session__'

// 一键读存储：sessionStorage 优先，降级 localStorage（老版本已登录用户过渡）
function readToken(key) {
  try {
    const v = sessionStorage.getItem(key)
    if (v != null) return v
  } catch (e) { /* storage 不可用 */ }
  try {
    return localStorage.getItem(key)
  } catch (e) { return '' }
}

function writeToken(key, value) {
  if (value == null || value === '') return
  try {
    sessionStorage.setItem(key, value)
  } catch (e) { /* storage 不可用 */ }
}

function removeToken(key) {
  try {
    sessionStorage.removeItem(key)
  } catch (e) { /* storage 不可用 */ }
  try {
    localStorage.removeItem(key)
  } catch (e) { /* storage 不可用 */ }
}

// 确定性幂等键：POST/PATCH 自动附带，同 URL+body 的重试复用同 key，
// 服务端 24h 去重（防双击/网络重试重复创建）。用户修改内容 → 新 key。
function idempotencyKey(url, data) {
  const s = url + '|' + JSON.stringify(data || {})
  let h = 5381
  for (let i = 0; i < s.length; i++) h = ((h << 5) + h + s.charCodeAt(i)) | 0
  const hash = Math.abs(h).toString(36)
  // 服务端要求 8-128 字符：前缀 + hash + 长度（截断保持上限）
  return 'idem-' + hash + '-' + String(s.length).slice(0, 60)
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

axios.interceptors.request.use((config) => {
  const token = memAccessToken || readToken(ACCESS_TOKEN_KEY)
  if (token) {
    config.headers = config.headers || {}
    config.headers.Authorization = `Bearer ${token}`
  }
  // 写操作自动幂等键（不覆盖调用方显式指定的 key）。
  // FormData/multipart 上传一律不加：JSON.stringify(FormData) 恒为 '{}'，
  // 会导致同一用户所有上传共用一个 key——24h 内第 2 次上传被服务端回放第 1 次的结果。
  const method = (config.method || 'get').toLowerCase()
  const isMultipart = typeof FormData !== 'undefined' && config.data instanceof FormData
  if ((method === 'post' || method === 'patch') && !isMultipart && !config.headers['Idempotency-Key']) {
    config.headers['Idempotency-Key'] = idempotencyKey(config.url || '', config.data)
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
    const resp = error.response
    // 统一解包后端错误信封 {error:{code,message}}：调用方 catch 读 error.message
    // 即可拿到真实原因（此前多数页面读 error?.response?.data?.message 为 undefined，
    // 统一显示"请求失败"，后端原因丢失）。
    if (resp?.data && typeof resp.data === 'object') {
      const backendMsg = (resp.data.error && resp.data.error.message) || resp.data.message
      if (backendMsg) error.message = backendMsg
    }
    if (!originalRequest) {
      return Promise.reject(error)
    }
    if (resp?.status !== 401) {
      // 403（角色不足等）：路由守卫按页拦不到的由后端兜底；错误消息已在上面解包，
      // 页面 catch 展示 error.message（如"only platform admin…"/"无权限执行该操作"）。
      return Promise.reject(error)
    }

    // refresh 请求本身返回 401 时，不再重试，直接清除登录态
    if (originalRequest.url === '/api/auth/refresh') {
      clearTokens()
      if (window.location.pathname.startsWith('/admin')) {
        window.location.href = '/login'
      }
      return Promise.reject(error)
    }

    if (originalRequest._retry) {
      // 刷新换新 token 后重试仍 401：登录态已失效（与小程序端对齐），清理并跳登录。
      if (resp?.status === 401) {
        clearTokens()
        if (window.location.pathname.startsWith('/admin')) {
          window.location.href = '/login'
        }
      }
      return Promise.reject(error)
    }
    originalRequest._retry = true

    const refreshToken = memRefreshToken || readToken(REFRESH_TOKEN_KEY)
    if (!refreshToken) {
      clearTokens()
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
      memAccessToken = newAccessToken
      writeToken(ACCESS_TOKEN_KEY, newAccessToken)
      const newRefreshToken = data.refresh_token
      if (newRefreshToken) {
        memRefreshToken = newRefreshToken
        writeToken(REFRESH_TOKEN_KEY, newRefreshToken)
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
        clearTokens()
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

function clearTokens() {
  memAccessToken = ''
  memRefreshToken = ''
  removeToken(ACCESS_TOKEN_KEY)
  removeToken(REFRESH_TOKEN_KEY)
  try {
    localStorage.removeItem('user')
    sessionStorage.removeItem('user')
  } catch (e) { /* storage 不可用 */ }
}

export const authStorage = {
  getAccessToken() {
    return memAccessToken || readToken(ACCESS_TOKEN_KEY)
  },
  getRefreshToken() {
    return memRefreshToken || readToken(REFRESH_TOKEN_KEY)
  },
  setTokens(accessToken, refreshToken) {
    if (accessToken) {
      memAccessToken = accessToken
      writeToken(ACCESS_TOKEN_KEY, accessToken)
    }
    if (refreshToken) {
      memRefreshToken = refreshToken
      writeToken(REFRESH_TOKEN_KEY, refreshToken)
    }
  },
  clearTokens
}

// 页面加载时从 sessionStorage 恢复内存态（一次读，后续拦截器走内存）
try {
  memAccessToken = sessionStorage.getItem(ACCESS_TOKEN_KEY) || ''
  memRefreshToken = sessionStorage.getItem(REFRESH_TOKEN_KEY) || ''
  // 老版本 localStorage 过渡：检测到则迁移到 sessionStorage 后清除（只迁一次）
  if (!memAccessToken && localStorage.getItem(ACCESS_TOKEN_KEY)) {
    memAccessToken = localStorage.getItem(ACCESS_TOKEN_KEY) || ''
    if (localStorage.getItem(REFRESH_TOKEN_KEY)) memRefreshToken = localStorage.getItem(REFRESH_TOKEN_KEY) || ''
    localStorage.removeItem(ACCESS_TOKEN_KEY)
    localStorage.removeItem(REFRESH_TOKEN_KEY)
    localStorage.removeItem(MEM_SESSION)
    if (memAccessToken) writeToken(ACCESS_TOKEN_KEY, memAccessToken)
    if (memRefreshToken) writeToken(REFRESH_TOKEN_KEY, memRefreshToken)
  }
} catch (e) { /* storage 不可用 */ }

/**
 * 上传等非 axios 拦截器场景下，动态构造带最新 accessToken 的 Authorization 头。
 * 在调用时刻读取（内存优先），避免组件创建时快照导致 token 轮转/过期后仍用旧值。
 */
export function getAuthHeader() {
  const token = memAccessToken || readToken(ACCESS_TOKEN_KEY)
  return token ? { Authorization: `Bearer ${token}` } : {}
}

export default axios
