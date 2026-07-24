export const caseList = [
  {
    id: 1,
    title: '重庆山区无人机物资配送',
    description: '利用大载重无人机在重庆山区实现高效物资运输，将传统车程的3小时缩短至40分钟，有效解决了山区物流"最后一公里"难题。',
    cover: 'https://images.unsplash.com/photo-1473968512647-3e447244af8f?auto=format&fit=crop&w=600&q=80',
    coverType: 'image',
    date: '2024-06-15',
    views: 1560,
    service: '无人机交易',
    location: '重庆武隆山区',
    media: [
      { type: 'image', url: 'https://images.unsplash.com/photo-1473968512647-3e447244af8f?auto=format&fit=crop&w=1000&q=80' },
      { type: 'image', url: 'https://images.unsplash.com/photo-1508614589041-895b88991e3e?auto=format&fit=crop&w=1000&q=80' }
    ],
    highlights: ['缩短运输时间78%', '适应复杂山地地形', '最大载重50kg']
  },
  {
    id: 2,
    title: '重庆两江新区无人机巡检',
    description: '在两江新区部署无人机自动巡检系统，对园区基础设施、在建工程进行全天候智能监测，实现安全隐患实时预警。',
    cover: 'https://www-cdn.djiits.com/dps/3e196dbfade1b1734dbbb335dde5de12.jpg?w=1184&h=592',
    coverType: 'image',
    date: '2024-07-20',
    views: 1120,
    service: '应急资源协同',
    location: '重庆两江新区',
    media: [
      { type: 'image', url: 'https://www-cdn.djiits.com/dps/3e196dbfade1b1734dbbb335dde5de12.jpg?w=1184&h=592' }
    ],
    highlights: ['AI自动识别隐患', '全自动航线规划', '实时数据回传']
  },
  {
    id: 3,
    title: '重庆无人机驾驶培训认证项目',
    description: '联合多家培训机构在重庆开展CAAC无人机驾驶员执照培训，已累计培养超过2000名持证飞手，为西南地区无人机产业输送专业人才。',
    cover: 'https://images.unsplash.com/photo-1506947411487-a56738267384?auto=format&fit=crop&w=600&q=80',
    coverType: 'image',
    date: '2024-08-10',
    views: 2800,
    service: '培训认证',
    location: '重庆渝北区',
    media: [
      { type: 'image', url: 'https://images.unsplash.com/photo-1506947411487-a56738267384?auto=format&fit=crop&w=1000&q=80' }
    ],
    highlights: ['CAAC官方认证考点', '2000+持证飞手', '87%通过率']
  }
]

export function getCaseById(id) {
  return caseList.find(c => String(c.id) === String(id))
}
