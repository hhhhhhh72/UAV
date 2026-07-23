const app = getApp()

Page({
  data: {
    city: '重庆',
    gauges: [
      { key: 'members', value: '180+', label: '会员单位' },
      { key: 'demands', value: '--', label: '活跃需求' },
      { key: 'pilots', value: '--', label: '认证飞手' },
      { key: 'completed', value: '--', label: '完成项目' }
    ],
    banners: [],
    entries: [
      { key: 'demands', name: '巡检作业', icon: '/images/entry-inspect.png' },
      { key: 'demands', name: '农林植保', icon: '/images/entry-agri.png' },
      { key: 'demands', name: '低空物流', icon: '/images/entry-logistics.png' },
      { key: 'demands', name: '应急测绘', icon: '/images/entry-emergency.png' },
      { key: 'demands', name: '设备采购', icon: '/images/entry-equip.png' },
      { key: 'training', name: '培训认证', icon: '/images/entry-training.png' },
      { key: 'events', name: '赛事活动', icon: '/images/entry-competition.png' },
      { key: 'association', name: '协会服务', icon: '/images/entry-association.png' },
      { key: 'experts', name: '专家智库', icon: '/images/entry-expert.png' },
      { key: 'more', name: '更多', icon: '/images/entry-more.png' }
    ],
    demands: [],
    notices: [
      { id: 'n1', title: '关于无人机实名登记管理的最新通知' },
      { id: 'n2', title: '2026年度重庆市无人机竞技大赛报名开启' },
      { id: 'n3', title: '协会关于征集低空经济团体标准的公告' }
    ]
  },

  onLoad() {
    this.loadHome()
    this.loadDemands()
    this.loadPilots()
  },

  onPullDownRefresh() {
    Promise.all([this.loadHome(), this.loadDemands()]).then(() => {
      wx.stopPullDownRefresh()
    })
  },

  /* ═══ API 请求 ═══ */

  loadHome() {
    return app.get('/api/v1/home', {
      city: app.globalData.city,
      lat: app.globalData.latitude,
      lng: app.globalData.longitude
    }).then(res => {
      this.setData({
        banners: res.data?.banners || [],
        notices: res.data?.notices?.length ? res.data.notices : this.data.notices
      })
    }).catch(() => {})
  },

  loadDemands() {
    return app.get('/api/v1/demands', { page: 1, page_size: 5 }).then(res => {
      const demands = (res.data || []).map(d => ({
        ...d,
        biz_type_text: this.bizTypeText(d.biz_type),
        budget_fen: d.budget_fen || 0,
        created_at: this.timeAgo(d.created_at)
      }))
      const counts = demands.length
      this.setData({
        demands,
        'gauges[1].value': counts > 0 ? counts + '+' : '--'
      })
    }).catch(() => {})
  },

  loadPilots() {
    return app.get('/api/v1/certified-pilots').then(res => {
      const count = (res.data || []).length
      this.setData({ 'gauges[2].value': count > 0 ? count + '+' : '--' })
    }).catch(() => {})
  },

  /* ═══ 工具函数 ═══ */

  bizTypeText(type) {
    const map = {
      cable_inspection: '线路巡检', plant_transport: '农林植保',
      spray_pesticide: '喷洒农药', clean_paint: '清洗喷涂',
      trade_lease: '设备采购', other: '其他'
    }
    return map[type] || type
  },

  timeAgo(iso) {
    if (!iso) return ''
    const diff = Date.now() - new Date(iso).getTime()
    const m = Math.floor(diff / 60000)
    if (m < 60) return m + '分钟前'
    const h = Math.floor(m / 60)
    if (h < 24) return h + '小时前'
    return Math.floor(h / 24) + '天前'
  },

  /* ═══ 交互 ═══ */

  onCityTap() {
    wx.showToast({ title: '定位功能开发中', icon: 'none' })
  },

  onSearchTap() {
    wx.navigateTo({ url: '/pages/search/search' })
  },

  onEntryTap(e) {
    const key = e.currentTarget.dataset.key
    const routes = {
      demands: '/pages/demands/list/list',
      training: '/pages/training/list/list',
      events: '/pages/events/list/list',
      experts: '/pages/experts/list/list',
      more: '/pages/services/services'
    }
    const path = routes[key]
    if (path) {
      wx.navigateTo({ url: path })
    } else if (key === 'more') {
      wx.switchTab({ url: '/pages/services/services' })
    } else {
      wx.navigateTo({ url: '/pages/demands/list/list?biz_type=' + key })
    }
  },

  onMoreDemands() {
    wx.navigateTo({ url: '/pages/demands/list/list' })
  },

  onDemandTap(e) {
    wx.navigateTo({ url: '/pages/demands/detail/detail?id=' + e.currentTarget.dataset.id })
  }
})
