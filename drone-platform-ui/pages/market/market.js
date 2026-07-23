Page({
  data: {
    shops: [
      { id: 1, name: '赤水市大疆...', logo: 'https://images.unsplash.com/photo-1558618666-fcd25c85f82e?w=100&h=100&fit=crop' },
      { id: 2, name: '遵义鼎疆科...', logo: 'https://images.unsplash.com/photo-1560472355-536de3962603?w=100&h=100&fit=crop' },
      { id: 3, name: '湄潭大疆农...', logo: 'https://images.unsplash.com/photo-1497366216548-37526070297c?w=100&h=100&fit=crop' },
      { id: 4, name: '贵州硒疆农...', logo: 'https://images.unsplash.com/photo-1486406146926-c627a92ad1ab?w=100&h=100&fit=crop' },
      { id: 5, name: '贵州瑞', logo: 'https://images.unsplash.com/photo-1504384308090-c54be3855833?w=100&h=100&fit=crop' }
    ],
    products: [
      {
        id: 1, name: '大疆70P农业无人机', price: '9500.00', stock: '库存 1  1人浏览',
        image: 'https://images.unsplash.com/photo-1506947411487-a56738267384?w=300&h=300&fit=crop',
        tags: [
          { text: '95新', bg: '#ff4d4f', color: '#fff' },
          { text: '商家发布', bg: '#ff9800', color: '#fff' },
          { text: '农用机', bg: '#fff', color: '#666' }
        ]
      },
      {
        id: 2, name: 'T100S播撒系统', price: '2299.00', stock: '库存 1  33人浏览',
        image: 'https://images.unsplash.com/photo-1581091226825-a6a2a5aee158?w=300&h=300&fit=crop',
        tags: [
          { text: '商家发布', bg: '#ff9800', color: '#fff' },
          { text: '配件', bg: '#fff', color: '#666' }
        ]
      },
      {
        id: 3, name: '大疆 T100S', soldOut: true,
        image: 'https://images.unsplash.com/photo-1473968512647-3e447244af8f?w=300&h=300&fit=crop',
        tags: []
      },
      {
        id: 4, name: '大疆 T70', soldOut: true,
        image: 'https://images.unsplash.com/photo-1508614589041-895b8c9d7ef5?w=300&h=300&fit=crop',
        tags: []
      }
    ]
  },

  showToast(e) {
    wx.showToast({ title: e.currentTarget.dataset.msg, icon: 'none' })
  }
})
