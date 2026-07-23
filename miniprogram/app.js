App({
  globalData: {
    baseURL: 'http://localhost:8080',
    token: '',
    userInfo: null,
    city: '重庆',
    cityCode: '500000',
    latitude: 29.5,
    longitude: 106.5
  },

  onLaunch() {
    const token = wx.getStorageSync('token')
    if (token) {
      this.globalData.token = token
      this.fetchMe()
    }
  },

  fetchMe() {
    this.request({ url: '/api/v1/me', method: 'GET' }).then(res => {
      this.globalData.userInfo = res.data
    }).catch(() => {
      this.globalData.token = ''
      wx.removeStorageSync('token')
    })
  },

  login() {
    return new Promise((resolve, reject) => {
      wx.login({
        success: (res) => {
          this.request({
            url: '/api/v1/auth/wechat/login',
            method: 'POST',
            data: { code: res.code },
            noAuth: true
          }).then(authRes => {
            const { access_token, user } = authRes.data
            this.globalData.token = access_token
            this.globalData.userInfo = user
            wx.setStorageSync('token', access_token)
            resolve(user)
          }).catch(reject)
        },
        fail: reject
      })
    })
  },

  request({ url, method = 'GET', data = {}, noAuth = false }) {
    const header = { 'Content-Type': 'application/json' }
    if (!noAuth && this.globalData.token) {
      header['Authorization'] = `Bearer ${this.globalData.token}`
    }
    return new Promise((resolve, reject) => {
      wx.request({
        url: this.globalData.baseURL + url,
        method,
        data,
        header,
        success(res) {
          if (res.statusCode >= 200 && res.statusCode < 300) {
            resolve(res.data)
          } else if (res.statusCode === 401 && !noAuth) {
            wx.removeStorageSync('token')
            reject(new Error('登录已过期'))
          } else {
            const err = res.data?.error?.message || '请求失败'
            reject(new Error(err))
          }
        },
        fail(err) {
          reject(new Error('网络连接失败'))
        }
      })
    })
  },

  get(path, params = {}) {
    const query = Object.keys(params).map(k => `${k}=${params[k]}`).join('&')
    const url = query ? `${path}?${query}` : path
    return this.request({ url, method: 'GET' })
  },

  post(path, data = {}) {
    return this.request({ url: path, method: 'POST', data })
  }
})
