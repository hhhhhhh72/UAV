Page({
  data: {
    userInfo: {
      name: 'wx_obVhC11w',
      role: '飞手',
      avatar: ''
    },
    orders: [
      { key: 'unpaid', label: '待付款', icon: 'icon-wallet' },
      { key: 'unship', label: '待发货', icon: 'icon-package' },
      { key: 'unrecv', label: '待收货', icon: 'icon-truck' },
      { key: 'done', label: '已完成', icon: 'icon-check-circle' },
      { key: 'refund', label: '退款/售后', icon: 'icon-refresh' }
    ],
    services: [
      { key: 'coupon', label: '优惠券', icon: 'icon-ticket', bg: '#fff3e0', color: '#ff9800' },
      { key: 'ship', label: '商家发货', icon: 'icon-truck', bg: '#fff3e0', color: '#ff9800' },
      { key: 'dist', label: '分销绑定', icon: 'icon-share-2', bg: '#fff3e0', color: '#ff9800' },
      { key: 'addr', label: '收货地址', icon: 'icon-map-pin', bg: '#fff3e0', color: '#ff9800' }
    ],
    tools: [
      { key: 'realname', label: '实名认证', icon: 'icon-user-check', bg: '#ff9800', color: '#fff' },
      { key: 'pilot', label: '飞手认证', icon: 'icon-award', bg: '#ff9800', color: '#fff' },
      { key: 'merchant', label: '商家认证', icon: 'icon-briefcase', bg: '#ff9800', color: '#fff' },
      { key: 'publish', label: '我的发布', icon: 'icon-edit', bg: '#ff9800', color: '#fff' },
      { key: 'points', label: '我的积分', icon: 'icon-star', bg: '#ff9800', color: '#fff' },
      { key: 'device', label: '设备绑定', icon: 'icon-cpu', bg: '#ff9800', color: '#fff' },
      { key: 'service', label: '官方客服', icon: 'icon-headphones', bg: '#ff9800', color: '#fff' },
      { key: 'about', label: '公司简介', icon: 'icon-file-text', bg: '#ff9800', color: '#fff' }
    ]
  },

  onLoad() {
    // 可在此调用 auth.isLoggedIn() 检查登录状态
  },

  showToast(e) {
    wx.showToast({ title: e.currentTarget.dataset.msg, icon: 'none' })
  }
})
