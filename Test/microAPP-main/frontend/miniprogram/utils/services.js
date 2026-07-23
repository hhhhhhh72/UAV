// 基础服务数据池
export const serviceList = [
  { id: 'flight', name: '飞行服务', description: '空域查询、飞行申报', icon: '/static/icons/flight.svg', color: 'linear-gradient(135deg, #6366f1 0%, #a855f7 100%)' },
  { id: '2', name: '政务服务', description: '环保监测、安全巡查', icon: '/static/icons/government.svg', color: 'linear-gradient(135deg, #6366f1 0%, #a855f7 100%)' },
  { id: '8', name: '无人机外卖', description: '即时配送、在线下单', icon: '/static/icons/delivery.svg', color: 'linear-gradient(135deg, #38bdf8 0%, #3b82f6 100%)' },
  { id: '1', name: '无人机物流', description: '城市配送、物资运输', icon: '/static/icons/logistics-drone.svg', color: 'linear-gradient(135deg, #06b6d4 0%, #2563eb 100%)' },
  { id: '4', name: '无人机吊运', description: '高空吊运、设备安装', icon: '/static/icons/lifting.svg', color: 'linear-gradient(135deg, #fbbf24 0%, #ea580c 100%)' },
  { id: '5', name: '无人机表演', description: '活动表演、编队飞行', icon: '/static/icons/drone-show-v2.svg', color: 'linear-gradient(135deg, #6366f1 0%, #a855f7 100%)' },
  { id: '3', name: '无人机托管', description: '专业托管、保养维护', icon: '/static/icons/maintenance.svg', color: 'linear-gradient(135deg, #4ade80 0%, #16a34a 100%)' },
  { id: '7', name: '无人机租赁', description: '设备租赁、配件租赁', icon: '/static/icons/rent.svg', color: 'linear-gradient(135deg, #38bdf8 0%, #3b82f6 100%)' },
  { id: '6', name: '飞手培训', description: 'CAAC执照、技能培训', icon: '/static/icons/training-v2.svg', color: 'linear-gradient(135deg, #fbbf24 0%, #ea580c 100%)' },
  { id: '9', name: '低空研学', description: '科普教育、实践体验', icon: '/static/icons/study.svg', color: 'linear-gradient(135deg, #06b6d4 0%, #2563eb 100%)' },
  { id: '10', name: '无人机销售', description: '设备买卖、以旧换新', icon: '/static/icons/shop.svg', color: 'linear-gradient(135deg, #fbbf24 0%, #ea580c 100%)' },
  { id: '11', name: '金融服务', description: '设备保险、飞行护航', icon: '/static/icons/finance.svg', color: 'linear-gradient(135deg, #6366f1 0%, #a855f7 100%)' },
  { id: '12', name: '维修服务', description: '故障维修、定期保养', icon: '/static/icons/wrench.svg', color: 'linear-gradient(135deg, #38bdf8 0%, #3b82f6 100%)' },
  { id: '13', name: '无人机赛事', description: '竞技比赛、赛事组织', icon: '/static/icons/competition.svg', color: 'linear-gradient(135deg, #f43f5e 0%, #e11d48 100%)' }
]

// 定义分组结构
export const serviceGroupsConfig = [
  {
    title: '核心服务',
    subtitle: 'Core Services',
    ids: ['flight', '2', '8', '1']
  },
  {
    title: '商业应用',
    subtitle: 'Business Application',
    ids: ['4', '5', '3', '7', '13']
  },
  {
    title: '教育培训',
    subtitle: 'Education & Training',
    ids: ['6', '9']
  },
  {
    title: '增值服务',
    subtitle: 'Value-added Services',
    ids: ['10', '11', '12']
  }
]

export function getServiceById(id) {
  return serviceList.find(s => String(s.id) === String(id))
}

export const trainingShowcase = [
  {
    title: '实操场地',
    desc: '拥有 3000 平米标准室内外飞行场地，配套先进',
    image: '/static/images/training/practice-field.svg'
  },
  {
    title: '模拟实验室',
    desc: '高精度模拟飞行系统，真实还原各种气象与环境',
    image: '/static/images/training/simulator-lab.svg'
  },
  {
    title: '考证支持',
    desc: 'CAAC 官方授权考点，一站式报名与拿证支持',
    image: '/static/images/training/certification-support.svg'
  }
]

