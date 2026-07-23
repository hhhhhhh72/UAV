const BASE_URL = 'http://localhost:8080'

function request(method) {
  return (url, data = {}, header = {}) => {
    return new Promise((resolve, reject) => {
      wx.request({
        url: BASE_URL + url,
        method,
        data,
        header: { 'Content-Type': 'application/json', ...header },
        success: res => resolve(res.data),
        fail: reject
      })
    })
  }
}

module.exports = {
  get: request('GET'),
  post: request('POST')
}
