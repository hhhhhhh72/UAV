const app = getApp()
Page({
  data: { demand: {}, bids: [], bizTypeText: '', budgetText: '0', isOwner: false, canBid: false, canSelect: false },
  onLoad(options) {
    const id = options.id
    app.get('/api/v1/demands').then(res => {
      const list = res.data || []
      const d = list.find(x => x.id === id) || {}
      this.setData({
        demand: d,
        bizTypeText: this.tText(d.biz_type),
        budgetText: ((d.budget_fen || 0) / 100).toFixed(0),
        isOwner: d.publisher_id === app.globalData.userInfo?.id
      })
    }).catch(() => {})
    this.loadBids(id)
  },
  loadBids(id) {
    app.request({ url: `/api/v1/demands/${id}/applications`, method: 'GET' }).then(res => {
      const bids = (res.data || []).map(b => ({ ...b, bidText: ((b.amount_fen || 0) / 100).toFixed(0) }))
      this.setData({ bids })
    }).catch(() => {})
  },
  onBidTap() { wx.showToast({ title: '报价功能开发中', icon: 'none' }) },
  onCompleteTap() { wx.showToast({ title: '确认完成功能开发中', icon: 'none' }) },
  tText(t) { const m = {cable_inspection:'线路巡检',plant_transport:'农林植保',spray_pesticide:'喷洒农药',clean_paint:'清洗喷涂',trade_lease:'设备采购',other:'其他'}; return m[t]||t }
})
