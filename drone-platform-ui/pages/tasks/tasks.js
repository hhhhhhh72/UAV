Page({
  data: {
    tasks: [
      {
        id: 1, title: '吊运坟石', type: '吊运', location: '贵州省·遵义市',
        address: '播州区人民政府 贵...', people: 1,
        startDate: '2026-07-19', endDate: '2026-07-31',
        price: '1800.00',
        image: 'https://images.unsplash.com/photo-1579829366248-204fe8413f31?w=240&h=180&fit=crop',
        tags: ['飞手险'],
        description: '吊运坟石距离80米 1800需要三百斤飞机 初八上吊 大大小小50坨左右 有需要电话联系'
      },
      {
        id: 2, title: '砍工', type: '吊运', location: '浙江省·衢州市',
        address: '外山邊 龍遊縣', people: 3,
        startDate: '2026-07-20', endDate: '2026-12-25',
        price: '10000.00',
        image: 'https://images.unsplash.com/photo-1508614589041-895b8c9d7ef5?w=240&h=180&fit=crop',
        tags: [],
        description: ''
      },
      {
        id: 3, title: '水稻田植保喷洒', type: '植保', location: '重庆市·江津区',
        address: '江津区双福街道', people: 2,
        startDate: '2026-07-22', endDate: '2026-08-15',
        price: '面议',
        image: 'https://images.unsplash.com/photo-1473968512647-3e447244af8f?w=240&h=180&fit=crop',
        tags: ['飞手险', '长期合作'],
        description: '500亩水稻田需要植保作业，要求T50以上机型，有作业保险'
      }
    ]
  },

  showToast(e) {
    wx.showToast({ title: e.currentTarget.dataset.msg, icon: 'none' })
  }
})
