const BASE_URL = 'http://localhost:3000'

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
      url: BASE_URL + '/api/auth/refresh',
      method: 'POST',
      data: { refreshToken },
      success: (r) => {
        if (r.statusCode >= 200 && r.statusCode < 300) resolve(r.data)
        else reject(r)
      },
      fail: reject
    })
  })

  const newAccessToken = res?.accessToken
  if (!newAccessToken) throw new Error('No access token returned')
  uni.setStorageSync(ACCESS_TOKEN_KEY, newAccessToken)
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
          resolve(res.data)
        } else if (res.statusCode === 401) {
          const refreshToken = authStorage.getRefreshToken()
          if (!refreshToken) return reject(res)

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
                      if (retryRes.statusCode >= 200 && retryRes.statusCode < 300) r(retryRes.data)
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
                if (retryRes.statusCode >= 200 && retryRes.statusCode < 300) resolve(retryRes.data)
                else reject(retryRes)
              },
              fail: reject
            })
          } catch (refreshError) {
            rejectQueue(refreshError)
            authStorage.clearTokens()
            reject(refreshError)
          } finally {
            isRefreshing = false
          }
        } else {
          reject(res)
        }
      },
      fail: (err) => {
        console.warn('API Request Fail:', options.url, err)
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
