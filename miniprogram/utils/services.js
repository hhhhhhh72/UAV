// 基础服务数据池 - 重庆无人机产业平台
export const serviceList = [
  { id: 'supply-demand', name: '产业供需对接', description: '需求大厅、供应展示、竞标报价', icon: '/static/icons/flight.svg', color: 'linear-gradient(135deg, #6366f1 0%, #a855f7 100%)' },
  { id: 'training-cert', name: '培训认证', description: 'CAAC执照、UTC认证、人社认证、飞手培训', icon: '/static/icons/training-v2.svg', color: 'linear-gradient(135deg, #fbbf24 0%, #ea580c 100%)' },
  { id: 'drone-trade', name: '无人机交易', description: '整机购买、维修服务、配件商城', icon: '/static/icons/shop.svg', color: 'linear-gradient(135deg, #06b6d4 0%, #2563eb 100%)' },
  { id: 'contract-sign', name: '合同签约', description: '合同模板、在线签章、合同作废', icon: '/static/icons/maintenance.svg', color: 'linear-gradient(135deg, #4ade80 0%, #16a34a 100%)' },
  { id: 'insurance-finance', name: '保险金融', description: '无人机保单、年审服务、金融贷款', icon: '/static/icons/finance.svg', color: 'linear-gradient(135deg, #38bdf8 0%, #3b82f6 100%)' },
  { id: 'emergency-resource', name: '应急资源协同', description: '救援案例、资源调度', icon: '/static/icons/drone-show-v2.svg', color: 'linear-gradient(135deg, #f43f5e 0%, #e11d48 100%)' }
]

// 定义分组结构
export const serviceGroupsConfig = [
  {
    title: '核心产业',
    subtitle: 'Core Industry',
    ids: ['supply-demand', 'drone-trade', 'contract-sign']
  },
  {
    title: '配套服务',
    subtitle: 'Supporting Services',
    ids: ['training-cert', 'insurance-finance', 'emergency-resource']
  }
]

export function getServiceById(id) {
  return serviceList.find(s => String(s.id) === String(id))
}

export const trainingShowcase = [
  {
    title: '无人机实训基地',
    desc: '占地 5000 平米专业飞行训练场地，配备多型号训练无人机',
    image: '/static/images/training/practice-field.svg'
  },
  {
    title: '模拟飞行中心',
    desc: '高精度飞行模拟器，支持多机型、多场景实操训练',
    image: '/static/images/training/simulator-lab.svg'
  },
  {
    title: '考证服务中心',
    desc: 'CAAC/UTC 官方授权考点，一站式报名培训与认证服务',
    image: '/static/images/training/certification-support.svg'
  }
]
