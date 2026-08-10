// 商城订单适配层（第一期·前端交互态）
//
// 背景：原型需要统一展示「商品 / 培训课程 / 无人机服务」三类订单，但后端目前
// 只有一个商品交易订单接口（GET /api/v1/trade-orders/mine），且 TradeOrder
// 字段极少（无类型、无快照标题/图片、无地址/物流/售后/评价标记）。
//
// 策略（对齐 docs 交付说明「第一阶段」）：
//   1. 商品订单：读取真实接口 /trade-orders/mine，并用 /products 建映射补齐标题/图片/规格；
//      缺失字段一律降级展示，不臆造。
//   2. 课程 / 服务订单：后端无统一订单契约，使用明确标注 source='demo' 的本地演示数据，
//      让五条状态流程可达可演示；真实接入统一订单接口后以 adapter 替换，不动页面。
//   3. 金额一律以 amount_fen（分）存储，展示时 /100 格式化，禁止直接用浮点数存金额。
//
// 所有从这里输出的订单对象都是「展示层归一化」后的形态，页面不直接解析后端字段。

import { request, BASE_URL } from './request'

/* ================= 常量（唯一状态/类型定义，页面不得另立） ================= */

// 订单中心的状态展示映射。completed → 待评价 仅是订单中心的展示层映射：
// 后端 completed 不记录「是否已评价」，因此「待评价/已评价」由前端临时标注，
// 属于 mock 边界，不称为持久化真相。
export const ORDER_STATUS = {
  pending: '待付款',
  paid: '待发货',
  shipped: '待收货',
  completed: '待评价',
  aftersale: '退款/售后',
}

// 状态入口固定顺序与文案，不得改动
export const STATUS_KEYS = ['pending', 'paid', 'shipped', 'completed', 'aftersale']

export const ORDER_TYPES = ['all', 'product', 'course', 'service']

export const ORDER_TYPE_LABEL = {
  all: '全部类型',
  product: '商品',
  course: '培训课程',
  service: '无人机服务',
}

/* ================= 工具 ================= */

// 分 → 元字符串（千分位）。金额一律分存，这里统一格式化。
export function fmtFen(fen) {
  if (fen === null || fen === undefined || fen === '') return '面议'
  const n = Number(fen)
  if (Number.isNaN(n)) return '面议'
  return (n / 100).toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

export function fmtDate(iso) {
  if (!iso) return ''
  try {
    const d = new Date(iso)
    if (Number.isNaN(d.getTime())) return String(iso).slice(0, 10)
    const m = d.getMonth() + 1
    const day = d.getDate()
    return `${d.getFullYear()}-${m < 10 ? '0' : ''}${m}-${day < 10 ? '0' : ''}${day}`
  } catch (e) {
    return ''
  }
}

// 订单号：真实单号缺省时用 ID 前缀模拟展示
export function orderNo(order) {
  if (order && order.order_no) return order.order_no
  if (order && order.id) return 'UAV' + String(order.id).replace(/\D/g, '').slice(-12)
  return '--'
}

function fullUrl(u) {
  if (!u) return ''
  return u.startsWith('http') ? u : BASE_URL + u
}

// 商品图片：product.images 是 JSON 数组字符串或数组；首图兜底到需求图
function productImage(product) {
  try {
    const arr = typeof product?.images === 'string' ? JSON.parse(product.images) : product?.images
    if (Array.isArray(arr) && arr[0]) return fullUrl(arr[0])
  } catch (e) { /* ignore */ }
  return ''
}

function productSubtitle(product) {
  const parts = []
  if (product?.brand) parts.push(product.brand)
  if (product?.model) parts.push(product.model)
  if (parts.length) return parts.join(' · ')
  if (product?.spec) return product.spec
  return '商品规格待补充'
}

const STATUS_DUE_TEXT = {
  pending: '请尽快完成付款',
  paid: '商家承诺 48 小时内发货',
  shipped: '快递运输中，请留意物流节点',
  completed: '交易已完成，欢迎评价',
  aftersale: '售后处理中，请查看进度',
}

/* ================= 真实商品订单 ================= */

async function fetchProductMap() {
  try {
    const res = await request({ url: '/api/v1/products', data: { page: 1, page_size: 100 } })
    const list = Array.isArray(res) ? res : (res?.data || [])
    const map = {}
    list.forEach((p) => { if (p && p.id) map[p.id] = p })
    return map
  } catch (e) {
    return {}
  }
}

async function fetchRealOrders() {
  try {
    const res = await request({ url: '/api/v1/trade-orders/mine' })
    const list = Array.isArray(res) ? res : (res?.data || [])
    return Array.isArray(list) ? list : []
  } catch (e) {
    return []
  }
}

function normalizeRealOrder(t, product) {
  const status = t.status || 'pending'
  return {
    id: t.id,
    order_no: '',
    type: 'product',
    status,
    source: 'real',
    origin: product?.seller_name || '低空商城',
    kind_label: '商品',
    title: product?.title || '商品订单',
    subtitle: productSubtitle(product),
    amount_fen: t.amount_fen ?? 0,
    quantity_label: '共 1 件',
    due_text: STATUS_DUE_TEXT[status] || '待处理',
    action: statusAction(status),
    image: productImage(product) || '/static/home/demand-solar.jpg',
    created_at: t.created_at,
    detail: {
      sections: [
        {
          title: '订单信息',
          rows: [
            { label: '订单状态', value: ORDER_STATUS[status] || status, status: 'wait' },
            { label: '订单编号', value: orderNo({ id: t.id }) },
            { label: '创建时间', value: fmtDate(t.created_at) },
            { label: '商品来源', value: product?.seller_name || '平台商城' },
          ],
        },
      ],
    },
  }
}

function statusAction(status) {
  if (status === 'pending') return '去支付'
  if (status === 'paid') return '提醒发货'
  if (status === 'shipped') return '确认收货'
  if (status === 'completed') return '去评价'
  return '查看售后'
}

/* ================= 演示数据（课程 / 服务 + 商品兜底，source='demo'） ================= */

const DEMO_ORDERS = [
  {
    id: 'demo-pay-01',
    order_no: 'UAV202608100321',
    type: 'product',
    status: 'pending',
    source: 'demo',
    origin: '低空商城',
    kind_label: '商品',
    title: 'Mavic 3E 智能飞行电池',
    subtitle: 'DJI 原厂 · 5000mAh · 单块装',
    amount_fen: 199900,
    quantity_label: '共 1 件',
    due_text: '请在 18:24 内完成付款',
    action: '去支付',
    image: '/static/home/demand-solar.jpg',
    created_at: '2026-08-10T10:21:00+08:00',
    detail: {
      hero: { title: 'Mavic 3E 智能飞行电池', sub: '低空商城 · 待付款' },
      sections: [
        {
          title: '商品参数',
          rows: [
            { label: '型号规格', value: 'Mavic 3E 系列专用' },
            { label: '容量', value: '5000mAh' },
            { label: '质保', value: '原厂 12 个月质保' },
            { label: '配送方式', value: '顺丰寄送' },
          ],
        },
        {
          title: '订单信息',
          rows: [
            { label: '订单状态', value: '待付款', status: 'wait' },
            { label: '订单编号', value: 'UAV202608100321' },
            { label: '收货信息', value: '张航 · 重庆市两江新区' },
          ],
        },
      ],
    },
  },
  {
    id: 'demo-ship-01',
    order_no: 'UAV202608080211',
    type: 'product',
    status: 'paid',
    source: 'demo',
    origin: '低空商城',
    kind_label: '商品',
    title: '无人机测绘套件（RTK）',
    subtitle: '含 RTK 模块、三脚架与收纳箱',
    amount_fen: 680000,
    quantity_label: '共 1 件',
    due_text: '商家承诺 48 小时内发货',
    action: '提醒发货',
    image: '/static/home/demand-lift.jpg',
    created_at: '2026-08-08T14:05:00+08:00',
    detail: {
      hero: { title: '无人机测绘套件（RTK）', sub: '低空商城 · 待发货' },
      sections: [
        {
          title: '商品与发货',
          rows: [
            { label: '套件内容', value: 'RTK 模块、三脚架与收纳箱' },
            { label: '发货时效', value: '48 小时内发货' },
            { label: '收货信息', value: '张航 · 重庆市两江新区' },
            { label: '订单状态', value: '待发货', status: 'wait' },
          ],
        },
      ],
    },
  },
  {
    id: 'demo-recv-01',
    order_no: 'UAV202608060910',
    type: 'product',
    status: 'shipped',
    source: 'demo',
    origin: '低空商城',
    kind_label: '商品',
    title: 'DJI M350 RTK 智能飞行电池',
    subtitle: '原厂配件 · 已由顺丰揽收',
    amount_fen: 199900,
    quantity_label: '共 1 件',
    due_text: '快递运输中 · 预计明日送达',
    action: '确认收货',
    image: '/static/home/hero-inspection.jpg',
    created_at: '2026-08-06T09:10:00+08:00',
    detail: {
      hero: { title: 'DJI M350 RTK 智能飞行电池', sub: '低空商城 · 待收货' },
      logistics: {
        carrier: '顺丰速运',
        tracking_no: 'SF15800329145',
        latest: '已到达重庆两江新区营业点',
        nodes: [
          { time: '2026-08-10 08:20', text: '到达两江新区营业点，配送员将尽快派送' },
          { time: '2026-08-09 21:42', text: '重庆转运中心发出' },
          { time: '2026-08-09 16:13', text: '商家已交付顺丰揽收' },
        ],
      },
      sections: [
        {
          title: '物流信息',
          rows: [
            { label: '承运公司', value: '顺丰速运' },
            { label: '物流单号', value: 'SF15800329145' },
            { label: '最新状态', value: '已到达重庆两江新区营业点', status: 'good' },
            { label: '订单状态', value: '待收货', status: 'wait' },
          ],
        },
      ],
    },
  },
  {
    id: 'demo-review-01',
    order_no: 'UAV202607301133',
    type: 'course',
    status: 'completed',
    source: 'demo',
    origin: '协会培训中心',
    kind_label: '培训课程',
    title: '无人机运行安全与任务规划实训',
    subtitle: '线上直播 + 线下实操 · 16 课时',
    amount_fen: 68000,
    quantity_label: '1 名学员',
    due_text: '课程已结束，评价后可领取结课凭证',
    action: '去评价',
    image: '/static/images/study/science-class.svg',
    created_at: '2026-07-30T11:33:00+08:00',
    detail: {
      hero: { title: '无人机运行安全与任务规划实训', sub: '协会培训中心 · 待评价' },
      review: {
        rating: 5,
        hint: '请评价本次课程内容、讲师讲解与实操安排',
        default_text: '课程内容符合预期，实操安排清晰，对实际任务规划很有帮助。',
      },
      sections: [
        {
          title: '课程信息',
          rows: [
            { label: '授课方式', value: '线上直播 + 线下实操' },
            { label: '课程时长', value: '16 课时' },
            { label: '培训地点', value: '重庆两江新区低空实训基地' },
            { label: '学员人数', value: '1 名学员' },
            { label: '订单状态', value: '待评价', status: 'wait' },
          ],
        },
      ],
    },
  },
  {
    id: 'demo-after-01',
    order_no: 'UAV202608020846',
    type: 'service',
    status: 'aftersale',
    source: 'demo',
    origin: '渝航智能装备有限公司',
    kind_label: '无人机服务',
    title: '电网巡检无人机服务（含热成像）',
    subtitle: '重庆两江新区 · 50km 线路巡检',
    amount_fen: 880000,
    quantity_label: '1 个服务项目',
    due_text: '售后申请处理中 · 等待服务商响应',
    action: '查看售后',
    image: '/static/home/demand-lift.jpg',
    created_at: '2026-08-02T08:46:00+08:00',
    detail: {
      hero: { title: '电网巡检无人机服务（含热成像）', sub: '渝航智能装备有限公司 · 退款/售后' },
      aftersale: {
        type: '服务未按约执行',
        status: '服务商处理中',
        amount_fen: 880000,
        created_at: '2026-08-09 14:36',
        description: '原定巡检时间未能按预约执行，申请全额退款。可在正式页面补充图片或沟通记录。',
        progress: [
          { time: '2026-08-09 14:36', text: '已提交售后申请' },
          { time: '2026-08-09 15:10', text: '平台已受理申请' },
          { time: '处理中', text: '等待服务商响应' },
        ],
      },
      sections: [
        {
          title: '售后信息',
          rows: [
            { label: '售后类型', value: '服务未按约执行' },
            { label: '申请状态', value: '服务商处理中', status: 'wait' },
            { label: '申请时间', value: '2026-08-09 14:36' },
            { label: '退款金额', value: '¥' + fmtFen(880000) },
          ],
        },
      ],
    },
  },
]

/* ================= 对外主接口 ================= */

// 加载订单列表（按状态 + 类型筛选）。返回 Promise<Order[]>
// status 传 STATUS_KEYS 中的 key 或 'all'；order_type 传 ORDER_TYPES 中的值。
export async function loadOrders({ status = 'all', order_type = 'all' } = {}) {
  const [real, demo] = await Promise.all([fetchRealOrders(), Promise.resolve(DEMO_ORDERS)])

  let orders = []
  if (real.length) {
    const pMap = await fetchProductMap()
    orders = real.map((t) => normalizeRealOrder(t, pMap[t.product_id] || null))
  }
  orders = orders.concat(demo)

  const statusMatch = (o) => status === 'all' || o.status === status
  const typeMatch = (o) => order_type === 'all' || o.type === order_type
  return orders.filter(statusMatch).filter(typeMatch)
}

// 按状态统计各入口数量（含类型筛选），用于订单中心五状态角标
export async function loadStatusCounts(order_type = 'all') {
  const real = await fetchRealOrders()
  const demo = DEMO_ORDERS
  const all = real
    .map((t) => ({ status: t.status || 'pending', type: 'product' }))
    .concat(demo.map((d) => ({ status: d.status, type: d.type })))

  const counts = { pending: 0, paid: 0, shipped: 0, completed: 0, aftersale: 0 }
  all.forEach((o) => {
    if (order_type !== 'all' && o.type !== order_type) return
    if (counts[o.status] !== undefined) counts[o.status] += 1
  })
  return counts
}

// 按 ID 查单条订单（真实 + 演示）。返回 Promise<Order|null>
export async function loadOrder(orderId) {
  if (!orderId) return null
  const demo = DEMO_ORDERS.find((d) => d.id === orderId)
  if (demo) return demo

  const [real, pMap] = await Promise.all([fetchRealOrders(), fetchProductMap()])
  const t = real.find((r) => r.id === orderId || r.order_no === orderId)
  return t ? normalizeRealOrder(t, pMap[t.product_id] || null) : null
}

// 演示数据的「客服」占位：未接入真实客服会话前只提示，不制造假会话
export function toastCustomerService() {
  uni.showToast({ title: '客服接入中，请联系协会秘书处', icon: 'none' })
}
