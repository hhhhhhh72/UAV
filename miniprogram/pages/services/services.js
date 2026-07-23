Page({
  data: { services: [
    { id:'demand', icon:'🎯', name:'产业供需对接', desc:'需求大厅·竞标·匹配' },
    { id:'innovation', icon:'💡', name:'产学研协同创新', desc:'成果·难题·攻关·转化' },
    { id:'compliance', icon:'📋', name:'合规政策服务', desc:'法规·标准·申报·案例' },
    { id:'talent', icon:'🎓', name:'人才教育融合', desc:'培训·赛事·招聘·校企' },
    { id:'brand', icon:'🏛️', name:'活动与品牌', desc:'活动·品牌·展会·报告' },
    { id:'emergency', icon:'🚨', name:'应急资源协同', desc:'应急·调度·救援·对接' },
    { id:'members', icon:'👥', name:'会员资源管控', desc:'专家·台账·人才·权限' },
    { id:'escrow', icon:'💰', name:'资金托管', desc:'充值·冻结·释放·退款' }
  ]},
  onTap(e) {
    const routes = {
      demand: '/pages/demands/list/list',
      escrow: '/pages/mine/mine'
    }
    const path = routes[e.currentTarget.dataset.id]
    if (path) wx.navigateTo({ url: path })
    else wx.showToast({ title: '即将开放', icon: 'none' })
  }
})
