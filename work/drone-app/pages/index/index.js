Page({
  data: {
    statusBarHeight: 44,
    currentBanner: 0,
    filterActive: 'distance',
    activeTab: 0,

    banners: [
      { title: '大疆T100S农业无人机', subtitle: '全能高手，安全无忧', bgFrom: '#1e293b', bgTo: '#0f172a' },
      { title: '极飞P150 Pro', subtitle: '智能喷洒，精准作业', bgFrom: '#1e3a5f', bgTo: '#0f172a' },
      { title: '大疆Mavic 3E', subtitle: '测绘利器，影像专家', bgFrom: '#1a2f4a', bgTo: '#0f172a' },
      { title: '大疆T100S农业无人机', subtitle: '全能高手，安全无忧', bgFrom: '#1e293b', bgTo: '#0f172a' }
    ],

    features: [
      { name: '二手评估', iconType: 'money' },
      { name: '无人机租赁', iconType: 'drone' },
      { name: '培训考试', iconType: 'book' },
      { name: '招商加盟', iconType: 'flag' },
      { name: '分享赚', iconType: 'share' }
    ],

    taskFeatures: ['海量飞手', '极速接单', '履约保障'],

    productEntries: [
      { name: '二手无人机' },
      { name: '无人机配件' }
    ],

    products: [
      { id: 1, name: '大疆T40 农业植保无人机', price: '28,800', distance: '1.2km', tag: '99新' },
      { id: 2, name: '极飞P150 2024款', price: '32,500', distance: '3.5km', tag: '95新' },
      { id: 3, name: '大疆Mavic 3E 测绘版', price: '19,999', distance: '5.8km', tag: '全新' },
      { id: 4, name: '大疆T50 农业无人机', price: '45,000', distance: '8.2km', tag: '90新' }
    ]
  },

  onLoad() {
    const systemInfo = wx.getSystemInfoSync();
    this.setData({ statusBarHeight: systemInfo.statusBarHeight });
  },

  onBannerChange(e) {
    this.setData({ currentBanner: e.detail.current });
  },

  setBanner(e) {
    this.setData({ currentBanner: parseInt(e.currentTarget.dataset.index) });
  },

  onLocationTap() { wx.showToast({ title: '选择位置', icon: 'none' }); },
  onMoreTap() { wx.showToast({ title: '更多', icon: 'none' }); },
  onScanTap() { wx.scanCode({ success: (res) => console.log(res) }); },
  onSearch() { wx.showToast({ title: '搜索', icon: 'none' }); },
  onNoticeTap() { wx.showToast({ title: '查看公告', icon: 'none' }); },

  onFeatureTap(e) {
    wx.showToast({ title: e.currentTarget.dataset.name, icon: 'none' });
  },

  onPublishTask() { wx.showToast({ title: '发布任务', icon: 'none' }); },

  onProductEntryTap(e) {
    wx.showToast({ title: e.currentTarget.dataset.name, icon: 'none' });
  },

  onFilterLocationTap() { wx.showToast({ title: '选择位置', icon: 'none' }); },

  onFilterTap(e) {
    const type = e.currentTarget.dataset.type;
    this.setData({ filterActive: type });
    const labels = { distance: '距离优先', price: '价格排序', filter: '筛选' };
    wx.showToast({ title: labels[type], icon: 'none' });
  },

  onProductTap(e) {
    wx.showToast({ title: '查看商品 ' + e.currentTarget.dataset.id, icon: 'none' });
  },

  switchTab(e) {
    const index = parseInt(e.currentTarget.dataset.index);
    this.setData({ activeTab: index });
    const labels = ['首页', '分类', '卖机', '任务', '我的'];
    wx.showToast({ title: labels[index], icon: 'none' });
  },

  onPullDownRefresh() {
    setTimeout(() => {
      wx.stopPullDownRefresh();
      wx.showToast({ title: '刷新成功', icon: 'success' });
    }, 1000);
  },

  onReachBottom() {
    wx.showLoading({ title: '加载中...' });
    setTimeout(() => wx.hideLoading(), 1000);
  }
});