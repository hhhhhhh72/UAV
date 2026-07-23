const app = getApp()

Page({
  data: {
    keyword: '',
    activeType: '',
    sort: 'newest',
    page: 1,
    total: 0,
    demands: [],
    loading: false,
    bizTypes: [
      { key: 'cable_inspection', name: '线路巡检' },
      { key: 'plant_transport', name: '农林植保' },
      { key: 'spray_pesticide', name: '喷洒农药' },
      { key: 'clean_paint', name: '清洗喷涂' },
      { key: 'trade_lease', name: '设备采购' },
      { key: 'other', name: '其他' }
    ]
  },

  onLoad(options) {
    if (options.biz_type) this.setData({ activeType: options.biz_type })
    this.loadDemands()
  },

  onPullDownRefresh() {
    this.setData({ page: 1, demands: [] })
    this.loadDemands().then(() => wx.stopPullDownRefresh())
  },

  loadDemands() {
    if (this.data.loading) return
    this.setData({ loading: true })

    const { activeType, sort, page } = this.data
    return app.get('/api/v1/demands', {
      biz_type: activeType,
      sort,
      page,
      page_size: 20
    }).then(res => {
      const items = (res.data || []).map(d => ({
        ...d,
        bizTypeText: this.typeText(d.biz_type),
        budgetText: (d.budget_fen / 100).toFixed(0),
        timeAgo: this.timeAgo(d.created_at)
      }))
      this.setData({
        demands: page === 1 ? items : [...this.data.demands, ...items],
        total: res.total || items.length,
        loading: false
      })
    }).catch(() => {
      this.setData({ loading: false })
      wx.showToast({ title: '加载失败', icon: 'none' })
    })
  },

  onLoadMore() {
    if (this.data.loading || this.data.demands.length >= this.data.total) return
    this.setData({ page: this.data.page + 1 }, () => this.loadDemands())
  },

  onTypeTap(e) {
    const type = e.currentTarget.dataset.type
    this.setData({ activeType: type, page: 1, demands: [] })
    this.loadDemands()
  },

  onSortTap(e) {
    this.setData({ sort: e.currentTarget.dataset.sort, page: 1, demands: [] })
    this.loadDemands()
  },

  onSearchTap() {
    wx.navigateTo({ url: '/pages/search/search' })
  },

  onTap(e) {
    wx.navigateTo({ url: `/pages/demands/detail/detail?id=${e.currentTarget.dataset.id}` })
  },

  typeText(t) {
    const m = { cable_inspection:'线路巡检', plant_transport:'农林植保', spray_pesticide:'喷洒农药', clean_paint:'清洗喷涂', trade_lease:'设备采购', other:'其他' }
    return m[t] || t
  },

  timeAgo(iso) {
    if (!iso) return ''
    const m = Math.floor((Date.now() - new Date(iso).getTime()) / 60000)
    if (m < 60) return m + '分钟前'
    const h = Math.floor(m / 60)
    if (h < 24) return h + '小时前'
    return Math.floor(h / 24) + '天前'
  }
})
