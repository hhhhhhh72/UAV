const app = getApp()
Page({
  data: {
    title:'', district:'', budget:'', description:'', contact:'',
    selectedType:'', bizTypes:[
      {key:'cable_inspection',name:'线路巡检'},{key:'plant_transport',name:'农林植保'},
      {key:'spray_pesticide',name:'喷洒农药'},{key:'clean_paint',name:'清洗喷涂'},
      {key:'trade_lease',name:'设备采购'},{key:'other',name:'其他'}
    ]
  },
  onField(e) { const f = e.currentTarget.dataset.field; this.setData({ [f]: e.detail.value }) },
  onTypeChange(e) { this.setData({ selectedType: this.data.bizTypes[e.detail.value].name }) },
  onSubmit() {
    const { title, district, budget, description, contact, selectedType } = this.data
    if (!title || !selectedType || !district || !contact) {
      return wx.showToast({ title: '请填写必填项', icon: 'none' })
    }
    const type = this.data.bizTypes.find(b => b.name === selectedType)?.key || 'other'
    app.post('/api/v1/demands', {
      publisher_name: app.globalData.userInfo?.id || '匿名',
      title, biz_type: type, district, contact,
      budget_fen: (parseFloat(budget) || 0) * 100,
      description
    }).then(() => {
      wx.showToast({ title: '发布成功', icon: 'success' })
      setTimeout(() => wx.navigateBack(), 1500)
    }).catch(err => wx.showToast({ title: err.message, icon: 'none' }))
  }
})
