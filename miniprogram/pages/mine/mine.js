const app = getApp()
Page({
  data: {
    isLogin: false, user: {}, balance: '0.00', frozen: '0.00',
    roleText: '',
    menus: [
      { key:'enterprise', icon:'🏢', name:'我的企业' },
      { key:'demands', icon:'📋', name:'我的需求' },
      { key:'contracts', icon:'📝', name:'合同管理' },
      { key:'certificates', icon:'🎖️', name:'我的证书' },
      { key:'orders', icon:'📦', name:'交易订单' },
      { key:'policies', icon:'🛡️', name:'保险保单' },
      { key:'portfolio', icon:'🏛️', name:'品牌展示' },
      { key:'settings', icon:'⚙️', name:'设置' }
    ]
  },
  onShow() {
    const token = wx.getStorageSync('token')
    if (token) {
      this.setData({ isLogin: true, user: app.globalData.userInfo || {} })
      this.loadBalance()
    }
  },
  loadBalance() {
    app.get('/api/v1/escrow/balance').then(res => {
      const a = res.data || {}
      this.setData({
        balance: ((a.balance_fen || 0) / 100).toFixed(2),
        frozen: ((a.frozen_fen || 0) / 100).toFixed(2)
      })
    }).catch(() => {})
  },
  onLogin() {
    app.login().then(() => {
      this.setData({ isLogin: true, user: app.globalData.userInfo || {} })
      this.loadBalance()
    }).catch(() => wx.showToast({ title: '登录失败', icon: 'none' }))
  },
  onMenuTap(e) {
    wx.showToast({ title: '即将开放', icon: 'none' })
  },
  onLogout() {
    wx.removeStorageSync('token')
    app.globalData.token = ''
    app.globalData.userInfo = null
    this.setData({ isLogin: false, balance: '0.00', frozen: '0.00' })
  }
})
