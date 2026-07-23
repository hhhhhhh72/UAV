Page({
  data: {
    currentCategory: 'farm',
    currentCategoryName: '农用机',
    categories: [
      { id: 'farm', name: '农用机' },
      { id: 'aerial', name: '航拍机' },
      { id: 'gimbal', name: '手持云台' },
      { id: 'camera', name: '运动相机' },
      { id: 'parts', name: '配件' },
      { id: 'cargo', name: '运载机' },
      { id: 'trainer', name: '训练机' }
    ],
    brands: [
      { id: 1, name: '大疆', logo: 'https://upload.wikimedia.org/wikipedia/commons/thumb/9/9c/DJI_Logo.svg/120px-DJI_Logo.svg.png' }
    ],
    products: [
      { id: 1, name: 'DJI T100S', slogan: '85L大容量 | 喷洒吊运全能旗舰农机', image: 'https://images.unsplash.com/photo-1506947411487-a56738267384?w=400&h=200&fit=crop' },
      { id: 2, name: 'DJI T100', slogan: '149.9kg 重载喷撒吊运全能植保机', image: 'https://images.unsplash.com/photo-1473968512647-3e447244af8f?w=400&h=200&fit=crop' },
      { id: 3, name: 'DJI T70', slogan: '50L 基础载重平地大田植保农机', image: 'https://images.unsplash.com/photo-1579829366248-204fe8413f31?w=400&h=200&fit=crop' },
      { id: 4, name: 'DJI T70P', slogan: '重载旗舰农业无人机', image: 'https://images.unsplash.com/photo-1508614589041-895b8c9d7ef5?w=400&h=200&fit=crop' }
    ],
    loadingMore: false
  },

  switchCategory(e) {
    const id = e.currentTarget.dataset.id
    const cat = this.data.categories.find(c => c.id === id)
    this.setData({ currentCategory: id, currentCategoryName: cat.name })
  },

  loadMore() {
    this.setData({ loadingMore: true })
    setTimeout(() => this.setData({ loadingMore: false }), 500)
  },

  showToast(e) {
    wx.showToast({ title: e.currentTarget.dataset.msg, icon: 'none' })
  }
})
