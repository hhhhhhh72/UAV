const api = require('../../utils/api')
const auth = require('../../utils/auth')
const { formatPrice, bizTypeMap } = require('../../utils/constants')

Page({
  data: {
    statusBarHeight: 0,
    city: '重庆',
    keyword: '',
    notice: '全国招商正式启动 · 新飞手入驻享专属补贴 · 夏季植保节火热进行中',
    notices: ['全国招商正式启动', '新飞手入驻享专属补贴', '夏季植保节火热进行中'],
    banners: [
      { id: 1, title: '大疆T200农业无人飞机', subtitle: '智稳双全实力派', image_url: 'https://images.unsplash.com/photo-1506947411487-a56738267384?w=600&h=300&fit=crop' },
      { id: 2, title: '夏季植保节', subtitle: '全场服务8折起', image_url: 'https://images.unsplash.com/photo-1473968512647-3e447244af8f?w=600&h=300&fit=crop' }
    ],
    quickEntries: [
      { key: 'demand', name: '任务发布', icon: '📋', bg: 'linear-gradient(135deg,#e3f2fd,#bbdefb)', shadow: '0 4px 12px rgba(66,165,245,0.2)' },
      { key: 'enterprise', name: '会员积分', icon: '💎', bg: 'linear-gradient(135deg,#fce4ec,#f8bbd9)', shadow: '0 4px 12px rgba(236,64,122,0.2)' },
      { key: 'training', name: '销量排行', icon: '🏆', bg: 'linear-gradient(135deg,#fff3e0,#ffe0b2)', shadow: '0 4px 12px rgba(255,152,0,0.2)' },
      { key: 'trading', name: '二手评估', icon: '🔍', bg: 'linear-gradient(135deg,#e8f5e9,#c8e6c9)', shadow: '0 4px 12px rgba(76,175,80,0.2)' },
      { key: 'jobs', name: '培训考试', icon: '🎓', bg: 'linear-gradient(135deg,#f3e5f5,#e1bee7)', shadow: '0 4px 12px rgba(156,39,176,0.2)' },
      { key: 'community', name: '招商加盟', icon: '🤝', bg: 'linear-gradient(135deg,#e0f7fa,#b2ebf2)', shadow: '0 4px 12px rgba(0,188,212,0.2)' },
      { key: 'insurance', name: '分享赚', icon: '💰', bg: 'linear-gradient(135deg,#fff8e1,#ffecb3)', shadow: '0 4px 12px rgba(255,193,7,0.2)' },
      { key: 'finance', name: '维修服务', icon: '🔧', bg: 'linear-gradient(135deg,#fbe9e7,#ffccbc)', shadow: '0 4px 12px rgba(255,87,34,0.2)' },
      { key: 'lease', name: '租赁服务', icon: '🔄', bg: 'linear-gradient(135deg,#e8eaf6,#c5cae9)', shadow: '0 4px 12px rgba(63,81,181,0.2)' },
      { key: 'news', name: '行业动态', icon: '📰', bg: 'linear-gradient(135deg,#f1f8e9,#dcedc8)', shadow: '0 4px 12px rgba(139,195,74,0.2)' }
    ],
    bizTypeTabs: [
      { label: '全部', value: '' },
      { label: '电缆巡检', value: 'cable_inspection' },
      { label: '植保运输', value: 'plant_transport' },
      { label: '喷洒农药', value: 'spray_pesticide' },
      { label: '清洗喷绘', value: 'clean_paint' },
      { label: '买卖租赁', value: 'trade_lease' },
      { label: '其他', value: 'other' }
    ],
    currentBizType: '',
    demands: [],
    page: 1,
    hasMore: true,
    loading: true,
    loadingMore: false
  },

  onLoad() {
    const sys = wx.getSystemInfoSync()
    this.setData({ statusBarHeight: sys.statusBarHeight })
    this.loadHome()
    this.loadDemands()
  },

  onShow() {
    this.loadDemands(true)
  },

  async loadHome() {
    try {
      const res = await api.get('/api/v1/home', { city: this.data.city, lat: 0, lng: 0 })
      const d = res.data || {}
      this.setData({
        banners: (d.banners || this.data.banners).map(b => ({
          ...b,
          image_url: b.image_url?.startsWith('/') ? 'http://localhost:8080' + b.image_url : b.image_url
        })),
        notices: d.notices || this.data.notices,
        notice: (d.notices || this.data.notices).join(' · ') || '',
        loading: false
      })
    } catch (e) {
      this.setData({ loading: false })
    }
  },

  async loadDemands(reset = false) {
    if (reset) this.setData({ page: 1, demands: [], hasMore: true })
    if (!this.data.hasMore && !reset) return

    const { page, currentBizType } = this.data
    try {
      this.setData({ loadingMore: true })
      const params = { page, page_size: 20 }
      if (currentBizType) params.biz_type = currentBizType

      const res = await api.get('/api/v1/demands', params)
      const list = (res.data || []).map(d => ({
        ...d,
        images: (d.images || []).map(img => img.startsWith('/') ? 'http://localhost:8080' + img : img),
        bizTypeLabel: (bizTypeMap[d.biz_type] || {}).label || d.biz_type,
        priceText: d.budget_fen ? formatPrice(d.budget_fen) : '面议'
      }))

      this.setData({
        demands: reset ? list : [...this.data.demands, ...list],
        hasMore: list.length >= 20,
        loadingMore: false,
        loading: false,
        page: reset ? 2 : page + 1
      })
    } catch (e) {
      this.setData({ loadingMore: false, loading: false })
    }
  },

  onReachBottom() {
    if (this.data.loadingMore) return
    this.loadDemands()
  },

  onPullDownRefresh() {
    this.setData({ page: 1, demands: [], hasMore: true, loading: true })
    Promise.all([this.loadHome(), this.loadDemands(true)]).then(() => {
      wx.stopPullDownRefresh()
    })
  },

  onBizTypeChange(e) {
    const val = e.currentTarget.dataset.value
    if (val === this.data.currentBizType) return
    this.setData({ currentBizType: val, loading: true })
    this.loadDemands(true)
  },

  onSearchFocus() {
    wx.navigateTo({ url: '/pages/search/search' })
  },

  onCityTap() {
    wx.showToast({ title: '城市选择开发中', icon: 'none' })
  },

  onBannerTap(e) {
    const item = e.currentTarget.dataset.item
    if (item.link_url) wx.navigateTo({ url: item.link_url })
  },

  onQuickEntry(e) {
    const item = e.currentTarget.dataset.item
    wx.showToast({ title: item.name + ' 开发中', icon: 'none' })
  },

  onMoreDemand() {
    wx.switchTab({ url: '/pages/tasks/list' })
  },

  onDemandTap(e) {
    const item = e.currentTarget.dataset.item
    wx.navigateTo({ url: `/pages/demand/detail/detail?id=${item.id}` })
  },

  showToast(e) {
    const msg = e.currentTarget?.dataset?.msg || e
    wx.showToast({ title: msg, icon: 'none' })
  }
})
