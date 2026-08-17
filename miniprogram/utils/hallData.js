// 供需大厅本地数据与辅助函数
// 本期为前端交互态：需求走真实 API（失败降级到模拟数据），供给/洽谈用模拟数据。
// 请求统一仍走 utils/request.js，这里仅提供本地兜底数据与展示层归一化。
import { authStorage, getStoredUser, BASE_URL } from './request'
import { getPosts as getPublishedPosts } from './publishData'

/* ================= 图片资源（复用首页静态图，保持稳定降级） ================= */
export const IMG_SOLAR = '/static/home/demand-solar.jpg'
export const IMG_LIFT = '/static/home/demand-lift.jpg'
export const IMG_HERO = '/static/home/hero-inspection.jpg'

/* ================= 模拟数据 ================= */
export const MOCK_DEMANDS = [
  {
    id: 'd1', type: '需求', cat: '巡检', status: '进行中',
    title: '光伏电站红外巡检服务', region: '重庆 · 江津区', time: '今天 08:45',
    price: '¥18,000', unit: '项目预算', deadline: '08 月 12 日',
    company: '重庆绿能新能源有限公司',
    desc: '50MW 光伏阵列月度巡检，要求具备热成像载荷和缺陷报告交付能力。计划两日内进场，报告需标注组件热斑、遮挡和破损位置。',
    image: IMG_SOLAR,
    fields: [['作业区域', '江津区白沙镇'], ['预计工期', '2 个工作日'], ['交付成果', '缺陷报告 + 原始影像'], ['资质要求', '保险与行业飞行资质']],
  },
  {
    id: 'd2', type: '需求', cat: '吊运', status: '进行中',
    title: '沙坪坝工地钢管无人机吊运', region: '重庆 · 沙坪坝区', time: '今天 09:20',
    price: '¥45,000', unit: '项目预算', deadline: '08 月 10 日',
    company: '中建三局重庆分公司',
    desc: '需将轻型钢管运送至高层施工平台，现场可勘查。请提供设备载重、团队配置和作业方案。',
    image: IMG_LIFT,
    fields: [['作业区域', '沙坪坝区西永'], ['预计工期', '3 天'], ['作业规模', '约 300 吨'], ['资质要求', '吊运作业经验']],
  },
  {
    id: 'd3', type: '需求', cat: '测绘', status: '进行中',
    title: '涪陵桥梁病害激光扫描', region: '重庆 · 涪陵区', time: '2 天前',
    price: '¥32,000', unit: '项目预算', deadline: '08 月 18 日',
    company: '重庆市交通规划设计院',
    desc: '对跨江大桥桥墩及附属设施开展点云采集，输出三维模型及病害标注成果。',
    image: IMG_HERO,
    fields: [['作业区域', '涪陵区'], ['预计工期', '5 个工作日'], ['交付成果', '三维点云'], ['资质要求', '测绘资质']],
  },
  {
    id: 'd4', type: '需求', cat: '植保', status: '已结束',
    title: '奉节柑橘园植保飞防', region: '重庆 · 奉节县', time: '3 天前',
    price: '¥12 /亩', unit: '作业单价', deadline: '已结束',
    company: '奉节县果业合作社',
    desc: '1200 亩柑橘园飞防作业，已由本地团队承接完成。',
    image: IMG_SOLAR,
    fields: [['作业区域', '奉节县'], ['作业规模', '1200 亩'], ['交付成果', '飞防记录'], ['状态', '已结束']],
  },
]

export const MOCK_SERVICES = [
  {
    id: 's1', type: '服务', cat: '巡检', status: '可对接',
    title: '桥梁与光伏设施精细化巡检', region: '服务重庆及周边', time: '已服务 36 个项目',
    price: '¥8,000 起', unit: '按项目报价',
    company: '重庆翼航科技有限公司',
    desc: '提供可见光、热成像、倾斜摄影和缺陷报告一体化巡检服务。团队配置多旋翼行业机和持证飞手，可支持紧急响应。',
    image: IMG_HERO,
    fields: [['服务范围', '桥梁 / 光伏 / 水务'], ['服务半径', '重庆市内及周边'], ['报价方式', '按项目报价'], ['设备资质', 'M350 RTK / 行业保险']],
  },
  {
    id: 's2', type: '服务', cat: '航拍', status: '可对接',
    title: '企业宣传航拍与 4K 成片制作', region: '重庆 · 两江新区', time: '最快 48 小时交付',
    price: '¥6,000 起', unit: '按项目报价',
    company: '山城视界文化传媒',
    desc: '企业园区、活动、文旅场景的航拍和后期制作。提供合规飞行方案、4K 成片和基础调色。',
    image: IMG_LIFT,
    fields: [['服务范围', '园区 / 活动 / 文旅'], ['服务半径', '主城九区'], ['报价方式', '套餐或定制'], ['设备资质', '4K 影像载荷']],
  },
  {
    id: 's3', type: '服务', cat: '测绘', status: '可对接',
    title: '工程测绘与实景三维建模', region: '重庆 · 渝北区', time: '可预约档期',
    price: '面议', unit: '按方案报价',
    company: '重庆空测数智有限公司',
    desc: '面向工程建设、矿山、道路与园区的航测建模服务，交付 DOM、DSM、点云及实景三维成果。',
    image: IMG_SOLAR,
    fields: [['服务范围', '测绘 / 建模'], ['服务半径', '全重庆'], ['报价方式', '按面积或项目'], ['设备资质', 'RTK / 激光雷达']],
  },
]

export const MOCK_PRODUCTS = [
  {
    id: 'p1', type: '商品', cat: '整机租赁', status: '可对接',
    title: '大疆 M350 RTK 行业套装租赁', region: '重庆 · 南岸区', time: '可当天取机',
    price: '¥1,200 /天', unit: '含基础保险',
    company: '重庆空域装备有限公司',
    desc: '行业级多旋翼平台，含遥控器、三电池和基础载荷。可配热成像、喊话器等任务载荷，支持企业短租。',
    image: IMG_HERO,
    fields: [['设备规格', 'M350 RTK'], ['设备成色', '九成新'], ['租赁方式', '按天 / 按周'], ['售后保障', '交付培训与保险']],
  },
  {
    id: 'p2', type: '商品', cat: '行业载荷', status: '可对接',
    title: '双光热成像云台载荷', region: '重庆 · 九龙坡区', time: '现货 3 套',
    price: '¥18,600', unit: '含税参考价',
    company: '渝航智造有限公司',
    desc: '适用于行业巡检与应急搜索的可见光 + 热成像双光云台，可提供安装适配和现场调试。',
    image: IMG_SOLAR,
    fields: [['设备规格', '640 热成像'], ['设备成色', '全新'], ['供货方式', '现货'], ['售后保障', '一年质保']],
  },
  {
    id: 'p3', type: '商品', cat: '零部件', status: '可对接',
    title: '植保机高效喷洒系统套件', region: '重庆 · 巴南区', time: '支持样机试用',
    price: '¥3,800 起', unit: '按配置报价',
    company: '巴南飞农装备',
    desc: '适配多型号植保机的喷洒系统升级套件，支持配件采购和技术指导。',
    image: IMG_LIFT,
    fields: [['适配机型', '主流植保机'], ['设备成色', '全新'], ['供货方式', '现货 / 定制'], ['售后保障', '安装指导']],
  },
]

/* ================= 分类 ================= */
export const HALL_CATEGORIES = {
  demand: ['全部', '巡检', '吊运', '测绘', '植保', '航拍'],
  service: ['全部', '巡检', '航拍', '测绘', '应急'],
  product: ['全部', '整机租赁', '行业载荷', '零部件'],
}

export function getKindItems(primary, supplyKind) {
  // 生产构建禁 mock：供给/商品在真实接口未就绪前显示空态，不展示演示数据
  if (!import.meta.env.DEV) return []
  if (primary === 'demand') return MOCK_DEMANDS
  return supplyKind === 'service' ? MOCK_SERVICES : MOCK_PRODUCTS
}

/* ================= 发布页本地内容 → 大厅卡片 ================= */
// 后端未接入期间，发布页内容存于本地 publish_posts（utils/publishData.js）。
// 已上架（live）内容由这里转换成大厅卡片结构，并入大厅列表展示；课程无大厅分类，不在此展示。
const PRODUCT_CAT_MAP = { 整机: '整机', 零部件: '配件', 载荷设备: '载荷', 租赁设备: '配件', 维修服务: '配件' }

export function publishPostToCard(post) {
  const v = post.values || {}
  if (post.type === 'demand') {
    const cat = v.biz || '其他'
    return {
      id: post.id,
      type: '需求',
      cat,
      status: '进行中',
      title: post.title,
      region: v.district ? `重庆 · ${v.district}` : '重庆',
      time: post.date,
      price: v.budget ? '¥' + v.budget : '面议',
      unit: '项目预算',
      deadline: '近期',
      company: '平台用户',
      desc: v.description || '暂无详细描述',
      image: cat === '吊运' ? IMG_LIFT : IMG_SOLAR,
      fields: [['作业区域', v.district || '重庆'], ['发布时间', post.date], ['预算', v.budget ? v.budget + ' 元' : '面议'], ['状态', '进行中']],
    }
  }
  if (post.type === 'service') {
    return {
      id: post.id,
      type: '服务',
      cat: v.category || '其他',
      status: '可对接',
      title: post.title,
      region: v.range || '重庆',
      time: post.date,
      price: '面议',
      unit: v.quote || '按项目报价',
      company: '平台服务商',
      desc: [v.equipment, v.cert].filter(Boolean).join('；') || '暂无详细描述',
      image: IMG_HERO,
      fields: [['服务范围', v.range || '—'], ['设备载荷', v.equipment || '—'], ['资质证书', v.cert || '—'], ['报价方式', v.quote || '—']],
    }
  }
  if (post.type === 'product') {
    const condition = v.condition || ''
    return {
      id: post.id,
      // 已发布到后端的商品（发布页商品提交走 POST /api/v1/products），后端 id 供列表跳转商品详情
      backendId: post.backendId || '',
      type: '商品',
      title: post.title,
      brand: v.brand || '',
      model: v.stock ? '可售 ' + v.stock + ' 件' : '',
      spec: v.brand || '平台商品',
      condition,
      isUsed: condition.indexOf('二手') === 0,
      price: v.price || '面议',
      image: IMG_HERO,
      images: [IMG_HERO],
      seller: '平台商家',
      company: '平台商家',
      desc: v.description || '',
      cat: PRODUCT_CAT_MAP[v.productType] || '配件',
      views: 0,
      status: '在售',
      fields: [['品牌型号', v.brand || '—'], ['成色', condition || '—'], ['交付方式', v.delivery || '—'], ['可售数量', v.stock ? v.stock + ' 件' : '—']],
    }
  }
  return null
}

// 指定类型下已上架的本地发布（课程除外）。
// 商品类型：已发布到后端的（backendId 非空）由后端 /products 接口返回，这里不再
// 并入，避免同一商品在列表中同时来自本地与后端；仅需求/服务等无后端通道的类型全量并入。
export function getLocalLiveCards(kind) {
  return getPublishedPosts()
    .filter((p) => p.statusKey === 'live' && p.type === kind)
    .filter((p) => !(p.type === 'product' && p.backendId))
    .map(publishPostToCard)
    .filter(Boolean)
}

export function kindTypeLabel(primary, supplyKind) {
  if (primary === 'demand') return '需求'
  return supplyKind === 'service' ? '服务能力' : '商品设备'
}

export function isEnded(item) {
  return item.status === '已结束' || item.status === '已成交' || item.status === '已下架'
}

/* ================= 需求分类 / 归一化 ================= */
export function classifyDemand(d) {
  const biz = String(d.biz_type || '').toLowerCase()
  const text = `${d.biz_type || ''} ${d.title || ''} ${d.description || ''}`
  if (biz === 'cable_inspection' || /巡检/.test(text)) return '巡检'
  if (biz === 'plant_transport' || biz === 'spray_pesticide' || /植保|农药|喷洒/.test(text)) return '植保'
  if (/吊运|运输|物资/.test(text)) return '吊运'
  if (/航拍|摄影|影像/.test(text)) return '航拍'
  if (/测绘|建模|测量/.test(text)) return '测绘'
  return ''
}

const DEMAND_STATUS_MAP = {
  published: '进行中', pending: '待审核', matched: '已匹配',
  completed: '已结束', cancelled: '已结束', closed: '已结束', rejected: '已驳回',
}

export function fmtBudget(fen) {
  if (fen == null || fen <= 0) return '面议'
  const yuan = fen / 100
  const whole = Math.floor(yuan)
  const cents = Math.round((yuan - whole) * 100)
  const w = whole.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ',')
  return cents > 0 ? `¥${w}.${cents < 10 ? '0' : ''}${cents}` : `¥${w}`
}

export function fmtDate(iso) {
  if (!iso) return ''
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return ''
  const m = d.getMonth() + 1
  const day = d.getDate()
  return `${d.getFullYear()}-${m < 10 ? '0' : ''}${m}-${day < 10 ? '0' : ''}${day}`
}

export function fmtRelative(iso) {
  if (!iso) return ''
  const t = new Date(iso).getTime()
  if (Number.isNaN(t)) return ''
  const diff = Date.now() - t
  const min = Math.floor(diff / 60000)
  if (min < 1) return '刚刚'
  if (min < 60) return `${min} 分钟前`
  const hour = Math.floor(min / 60)
  if (hour < 24) return `${hour} 小时前`
  const d = new Date(t)
  return `${d.getMonth() + 1}-${d.getDate()}`
}

// 后端 Demand → 大厅卡片
export function normalizeDemand(d) {
  if (!d) return null
  const title = String(d.title || '').trim()
  if (!title) return null
  const cat = classifyDemand(d)
  // 存库为相对路径 /uploads/xxx，展示层补全域名（否则小程序按本地包内资源加载 → 白图）；
  // /static/ 是本地包资源保持原样（补域名会 404），http 开头原样
  const resolveUrl = (u) => (u.indexOf('http') === 0 || u.indexOf('/static/') === 0 ? u : BASE_URL + u)
  const imgs = Array.isArray(d.images) ? d.images.filter((u) => typeof u === 'string' && u.trim()).map(resolveUrl) : []
  const image = imgs[0] || (cat === '吊运' ? IMG_LIFT : IMG_SOLAR)
  const status = DEMAND_STATUS_MAP[d.status] || '进行中'
  const budget = fmtBudget(d.budget_fen)
  return {
    id: String(d.id || ''),
    type: '需求',
    cat: cat || '其他',
    status,
    title,
    region: d.district ? `重庆 · ${d.district}` : '重庆',
    time: fmtRelative(d.created_at),
    price: budget,
    unit: '项目预算',
    deadline: status === '已结束' ? '已结束' : '近期',
    company: d.publisher_name || '平台用户',
    desc: d.description || '暂无详细描述',
    image,
    images: imgs,
    fields: [['地区', d.district || '重庆'], ['发布时间', fmtDate(d.created_at)], ['预算', budget], ['状态', status]],
  }
}

/* ================= 会话状态（本期前端交互态） ================= */
export const isLoggedIn = () => !!authStorage.getAccessToken()
export const isCertified = () => uni.getStorageSync('hall_certified') === '1'
export const setCertified = (v) => uni.setStorageSync('hall_certified', v ? '1' : '0')
export const getSession = () => ({ loggedIn: isLoggedIn(), certified: isCertified() })

export function simulateLogin() {
  if (!import.meta.env.DEV) return // 生产环境 no-op：不写入任何 mock token
  authStorage.setTokens('mock-access-token', 'mock-refresh-token')
  uni.setStorageSync('user', JSON.stringify({ name: '重庆云翼低空科技有限公司', role: 'enterprise', phone: '138****2468' }))
}

export function currentUserName() {
  const u = getStoredUser()
  return (u && (u.name || u.phone)) || '微信用户'
}

/* ================= 本地发布 / 意向存储（前端交互态） ================= */
const POSTS_KEY = 'hall_posts'
const SENT_KEY = 'hall_intents_sent'

export const SEED_POSTS = [
  { id: 'm1', type: '需求', title: '两江新区产业园三维建模项目', status: '已上架', date: '今天 09:40' },
  { id: 'm2', type: '服务', title: '桥梁与光伏设施精细化巡检', status: '已上架', date: '昨天 16:10' },
  { id: 'm3', type: '商品', title: '大疆 M350 RTK 行业套装租赁', status: '已驳回', date: '08-02', reason: '请补充设备实拍图和租赁押金说明。' },
]

export function getPosts() {
  const raw = uni.getStorageSync(POSTS_KEY)
  if (Array.isArray(raw)) return raw
  // 生产构建不播种演示帖子：无本地记录时返回空，不做假数据
  if (!import.meta.env.DEV) return []
  uni.setStorageSync(POSTS_KEY, SEED_POSTS)
  return SEED_POSTS
}
export function savePosts(posts) {
  uni.setStorageSync(POSTS_KEY, posts || [])
}

export function getSentIntents() {
  const raw = uni.getStorageSync(SENT_KEY)
  return Array.isArray(raw) ? raw : []
}
export function saveSentIntents(list) {
  uni.setStorageSync(SENT_KEY, list || [])
}

/* ================= 商品设备（电商化展示层） ================= */
// 图片相对路径处理：/static 为小程序本地资源原样使用，/uploads 等拼后端地址
export function fullImgUrl(u) {
  if (!u) return ''
  if (u.startsWith('http') || u.startsWith('/static/')) return u
  return BASE_URL + u
}

export const PRODUCT_CATEGORIES = ['全部', '整机', '配件', '载荷']

const PRODUCT_CONDITION_MAP = { new: '全新', used: '二手' }

// 电商分类：整机 / 配件 / 载荷（标题关键词兜底）
export function classifyProduct(p) {
  const text = `${p.title || ''} ${p.brand || ''}`
  if (p.prod_type === 'drone' || /无人机|植保机|行业机|套机/.test(text)) return '整机'
  if (/云台|载荷|热成像|喊话|探照/.test(text)) return '载荷'
  return '配件'
}

export function fmtPriceFen(fen) {
  if (fen == null || fen <= 0) return '面议'
  const yuan = fen / 100
  const whole = Math.floor(yuan)
  const cents = Math.round((yuan - whole) * 100)
  const w = whole.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ',')
  return cents > 0 ? `${w}.${cents < 10 ? '0' : ''}${cents}` : `${w}`
}

// 后端 DroneProduct → 电商卡片
export function normalizeProduct(p) {
  if (!p || !p.id) return null
  const title = String(p.title || '').trim()
  if (!title) return null
  const imgs = Array.isArray(p.images) ? p.images.filter((u) => typeof u === 'string' && u.trim()) : []
  const brand = p.brand || ''
  const model = p.model || ''
  return {
    id: String(p.id),
    title,
    brand,
    model,
    spec: brand && model ? `${brand} · ${model}` : brand || model || '平台商品',
    condition: PRODUCT_CONDITION_MAP[p.condition] || '商家发布',
    isUsed: p.condition === 'used',
    price: fmtPriceFen(p.price_fen),
    image: fullImgUrl(imgs[0] || ''),
    images: imgs.map(fullImgUrl),
    seller: p.seller_name || '平台商家',
    desc: p.description || '',
    cat: classifyProduct(p),
    views: p.views || 0,
    status: p.status || 'listed',
  }
}

// 后端 ServiceListing → 大厅服务卡片（trade-card 结构）
export function normalizeService(s) {
  if (!s || !s.id) return null
  const title = String(s.title || '').trim()
  if (!title) return null
  const price = fmtPriceFen(s.price_fen)
  return {
    id: String(s.id),
    type: '服务',
    cat: s.category || '其他',
    status: '可对接',
    title,
    region: s.region || '重庆',
    time: s.unit ? `按${s.unit}报价` : '',
    price: price === '面议' ? '面议' : '¥' + price + ' 起',
    unit: s.unit ? '按' + s.unit + '报价' : '按项目报价',
    company: s.provider_name || '平台服务商',
    desc: s.description || '暂无详细描述',
    image: IMG_HERO,
    fields: [['服务范围', s.category || '—'], ['服务区域', s.region || '—'], ['报价方式', s.unit ? '按' + s.unit + '报价' : '按项目报价'], ['服务商', s.provider_name || '—']],
  }
}
