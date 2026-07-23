module.exports = {
  isLoggedIn() {
    return !!wx.getStorageSync('access_token')
  },
  wxLogin() {
    return new Promise((resolve, reject) => {
      wx.login({
        success: res => {
          if (res.code) resolve(res.code)
          else reject(new Error('登录失败'))
        },
        fail: reject
      })
    })
  }
}
