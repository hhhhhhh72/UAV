export const serviceList = [
  { id: 'demands', name: '需求大厅', description: '发布浏览无人机作业需求', icon: '/static/icons/exchange.svg', color: 'linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%)' },
  { id: 'enterprise', name: '企业入驻', description: '企业注册与资质审核', icon: '/static/icons/shop.svg', color: 'linear-gradient(135deg, #06b6d4 0%, #0891b2 100%)' },
  { id: 'experts', name: '专家智库', description: '行业专家在线咨询', icon: '/static/icons/government.svg', color: 'linear-gradient(135deg, #f59e0b 0%, #d97706 100%)' },
  { id: 'resources', name: '资源台账', description: '无人机/机场/试飞场地', icon: '/static/icons/logistics-drone.svg', color: 'linear-gradient(135deg, #06b6d4 0%, #2563eb 100%)' },
  { id: 'achievements', name: '成果库', description: '高校院所专利与产品', icon: '/static/icons/study.svg', color: 'linear-gradient(135deg, #7c3aed 0%, #4f46e5 100%)' },
  { id: 'challenges', name: '研发难题', description: '揭榜挂帅技术攻关', icon: '/static/icons/competition.svg', color: 'linear-gradient(135deg, #f97316 0%, #ea580c 100%)' },
  { id: 'training', name: '培训认证', description: 'CAAC/UTC/人社证书', icon: '/static/icons/training-v2.svg', color: 'linear-gradient(135deg, #f59e0b 0%, #ea580c 100%)' },
  { id: 'competitions', name: '赛事中心', description: '无人机竞技与创新大赛', icon: '/static/icons/drone-show-v2.svg', color: 'linear-gradient(135deg, #f43f5e 0%, #e11d48 100%)' },
  { id: 'jobs', name: '招聘求职', description: '职位发布与简历投递', icon: '/static/icons/rent.svg', color: 'linear-gradient(135deg, #38bdf8 0%, #3b82f6 100%)' },
  { id: 'colleges', name: '院校展示', description: '无人机专业院校信息', icon: '/static/icons/study.svg', color: 'linear-gradient(135deg, #06b6d4 0%, #2563eb 100%)' },
  { id: 'trade', name: '无人机交易', description: '整机/维修/配件商城', icon: '/static/icons/lifting.svg', color: 'linear-gradient(135deg, #fbbf24 0%, #ea580c 100%)' },
  { id: 'insurance', name: '保险金融', description: '保单/年审/贷款', icon: '/static/icons/finance.svg', color: 'linear-gradient(135deg, #ef4444 0%, #dc2626 100%)' },
  { id: 'events', name: '协会活动', description: '论坛/走访/沙龙/培训', icon: '/static/icons/flight.svg', color: 'linear-gradient(135deg, #10b981 0%, #059669 100%)' },
  { id: 'exhibitions', name: '展会排期', description: '产业展会与展位申请', icon: '/static/icons/delivery.svg', color: 'linear-gradient(135deg, #f59e0b 0%, #d97706 100%)' },
  { id: 'portfolios', name: '品牌展示', description: '企业名片与产品展示', icon: '/static/icons/service.svg', color: 'linear-gradient(135deg, #6366f1 0%, #a855f7 100%)' },
  { id: 'reports', name: '行业报告', description: '白皮书与调研报告', icon: '/static/icons/volume.svg', color: 'linear-gradient(135deg, #38bdf8 0%, #3b82f6 100%)' },
  { id: 'emergency', name: '应急资源', description: '应急装备与调度', icon: '/static/icons/wrench.svg', color: 'linear-gradient(135deg, #f97316 0%, #ea580c 100%)' },
  { id: 'rescue', name: '救援案例', description: '实战救援经验库', icon: '/static/icons/maintenance.svg', color: 'linear-gradient(135deg, #4ade80 0%, #16a34a 100%)' },
]

export const serviceGroupsConfig = [
  { title: '产业供需', subtitle: 'Supply & Demand', ids: ['demands', 'enterprise', 'experts', 'resources'] },
  { title: '产学研协同', subtitle: 'Innovation', ids: ['achievements', 'challenges', 'reports'] },
  { title: '人才教育', subtitle: 'Talent & Education', ids: ['training', 'competitions', 'jobs', 'colleges'] },
  { title: '交易与金融', subtitle: 'Trading & Finance', ids: ['trade', 'insurance'] },
  { title: '活动与品牌', subtitle: 'Events & Brand', ids: ['events', 'exhibitions', 'portfolios'] },
  { title: '应急协同', subtitle: 'Emergency', ids: ['emergency', 'rescue'] },
]

export function getServiceById(id) {
  return serviceList.find(s => String(s.id) === String(id))
}

export const trainingShowcase = []
