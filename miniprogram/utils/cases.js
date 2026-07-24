export const caseList = [
  {
    id: 1,
    title: '海岛物资无人机配送',
    description: '成功实现温州海岛之间的高效配送，将传统海运的2小时航程缩短至20分钟，极大提升了急需物资的转运效率。',
    cover: 'https://images.unsplash.com/photo-1473968512647-3e447244af8f?auto=format&fit=crop&w=600&q=80',
    coverType: 'image',
    date: '2024-03-20',
    views: 1250,
    service: '无人机物流',
    location: '温州洞头海域',
    media: [
      { type: 'image', url: 'https://images.unsplash.com/photo-1473968512647-3e447244af8f?auto=format&fit=crop&w=1000&q=80' },
      { type: 'image', url: 'https://images.unsplash.com/photo-1508614589041-895b88991e3e?auto=format&fit=crop&w=1000&q=80' }
    ],
    highlights: ['缩短运输时间80%', '支持复杂海况飞行', '全流程自动驾驶']
  },
  {
    id: 2,
    title: '城市河道环保巡查',
    description: '利用无人机搭载高光谱相机，对城市内河进行全天候自动巡查，实时识别排污口及水质异常区域。',
    cover: 'https://www-cdn.djiits.com/dps/3e196dbfade1b1734dbbb335dde5de12.jpg?w=1184&h=592',
    coverType: 'image',
    date: '2024-02-15',
    views: 890,
    service: '政务服务',
    location: '温州温瑞塘河',
    media: [
      { type: 'image', url: 'https://www-cdn.djiits.com/dps/3e196dbfade1b1734dbbb335dde5de12.jpg?w=1184&h=592' }
    ],
    highlights: ['多源数据融合', 'AI智能识别污迹', '自动生成巡查报告']
  },
  {
    id: 3,
    title: '5.18 洞头灯光秀表演',
    description: '500架无人机在洞头海滨上演了一场精彩的灯光秀，为市民带来了震撼的视觉盛宴，展现了温州低空经济的魅力。',
    cover: 'https://images.unsplash.com/photo-1506947411487-a56738267384?auto=format&fit=crop&w=600&q=80',
    coverType: 'image',
    date: '2024-05-18',
    views: 3200,
    service: '无人机表演',
    location: '温州洞头新城广场',
    media: [
      { type: 'image', url: 'https://images.unsplash.com/photo-1506947411487-a56738267384?auto=format&fit=crop&w=1000&q=80' }
    ],
    highlights: ['500架大规模编队', 'RTK高精度定位', '全程地面站监控']
  }
]

export function getCaseById(id) {
  return caseList.find(c => String(c.id) === String(id))
}

