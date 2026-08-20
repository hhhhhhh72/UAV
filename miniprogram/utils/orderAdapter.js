// 商城订单适配层（第一期·前端交互态）
//
// 背景：原型需要统一展示「商品 / 培训课程 / 无人机服务」三类订单，但后端目前
// 只有一个商品交易订单接口（GET /api/v1/trade-orders/mine），且 TradeOrder
// 字段极少（无类型、无快照标题/图片、无地址/物流/售后/评价标记）。
//
// 策略（对齐 docs 交付说明「第一阶段」）：
//   1. 商品订单：读取真实接口 /trade-orders/mine，并用 /products 建映射补齐标题/图片/规格；
//      缺失字段一律降级展示，不臆造。
//   2. 课程 / 服务订单：后端无统一订单契约，当前不返回数据，订单中心显示空态；
//      真实接入统一订单接口后以 adapter 替换，不动页面。
//   3. 金额一律以 amount_fen（分）存储，展示时 /100 格式化，禁止直接用浮点数存金额。
//
// 所有从这里输出的订单对象都是「展示层归一化」后的形态，页面不直接解析后端字段。

import { request, BASE_URL, getStoredUser } from './request'

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
  cancelled: '已取消',
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
  cancelled: '订单已取消，如有疑问请联系客服',
}

/* ================= 真实商品订单 ================= */

async function fetchProductMap() {
  const map = {}
  try {
    // 翻页拉全量（后端 page_size 上限 100，商品可能超过一页；映射缺失时订单标题/图会回落兜底）
    for (let page = 1; page <= 10; page++) {
      const res = await request({ url: '/api/v1/products', data: { page, page_size: 100 } })
      const list = Array.isArray(res) ? res : (res?.data || [])
      if (!Array.isArray(list) || list.length === 0) break
      list.forEach((p) => { if (p && p.id) map[p.id] = p })
      if (list.length < 100) break
    }
  } catch (e) {
    /* 映射失败时订单回落兜底展示 */
  }
  return map
}

async function fetchRealOrders() {
  try {
    const res = await request({ url: '/api/v1/trade-orders/mine' })
    const list = Array.isArray(res) ? res : (res?.data || [])
    return Array.isArray(list) ? list : []
  } catch (e) {
    // 网络/接口失败必须向上抛：订单中心（index/list）与订单详情等调用方已处理
    // error 分支（显示加载失败 + 重试），不能吞成空数组导致误显示「暂无订单」。
    throw e
  }
}

// 售后单 → 展示区块（aftersale 状态字段来自后端售后契约：
// aftersale_status = pending 待审核 / approved 已同意退款 / rejected 已驳回）
function aftersaleInfo(t) {
  if (!t.aftersale_status) return null
  const type = t.aftersale_type === 'return' ? '退货退款' : '仅退款'
  const time = fmtDate(t.aftersale_time)
  const statusMap = { pending: '待审核', approved: '售后已完成', rejected: '已驳回' }
  // 进度按申请状态推导（不臆造节点：时间缺省用申请时间/占位）
  const progress = []
  // 第一步：申请提交。待审核单才提示「等待平台审核」；已结案单只陈述已提交，
  // 避免退款完成后仍显示「等待审核」的误导。
  progress.push({ time: time || '-', text: t.aftersale_status === 'pending' ? '已提交售后申请，等待平台审核' : '已提交售后申请' })
  // 模拟支付体系下无真实资金动作：不承诺「款项原路退回」，只描述平台流程处理
  if (t.aftersale_status === 'approved') {
    progress.push({ time: '-', text: '已同意退款，退款按平台流程处理' })
    progress.push({ time: '-', text: '售后已完成，退款处理结束' })
  }
  if (t.aftersale_status === 'rejected') progress.push({ time: '-', text: '平台已驳回申请，订单已结案' })
  return {
    type,
    status: statusMap[t.aftersale_status] || t.aftersale_status,
    amount_fen: t.aftersale_amount_fen || 0,
    created_at: time,
    description: t.aftersale_desc || t.aftersale_reason || '',
    progress,
  }
}

function normalizeRealOrder(t, product) {
  const status = t.status || 'pending'
  // 交易角色：seller_id 是自己 → 我是卖家（发货方），否则买家。
  // demo 用户无 id 时按买家渲染（正常登录用户必有 id）。
  const me = getStoredUser()
  const myId = me && me.id
  const role = myId && t.seller_id === myId ? 'seller' : 'buyer'
  const af = aftersaleInfo(t)
  const sections = [
    {
      title: '订单信息',
      rows: [
        // 有售后记录时显示售后状态（结案单状态已回 completed，不再显示「待评价」）
        { label: '订单状态', value: af ? af.status : (ORDER_STATUS[status] || status), status: 'wait' },
        { label: '订单编号', value: orderNo({ id: t.id }) },
        { label: '创建时间', value: fmtDate(t.created_at) },
        { label: '交易角色', value: role === 'seller' ? '我是卖家（卖出）' : '我是买家（买入）' },
        { label: '商品来源', value: product?.seller_name || '平台商城' },
      ],
    },
  ]
  if (af) {
    sections.push({
      title: '售后信息',
      rows: [
        { label: '售后类型', value: af.type },
        { label: '申请状态', value: af.status, status: af.status === '待审核' ? 'wait' : '' },
        { label: '退款金额', value: '¥' + fmtFen(af.amount_fen) },
        { label: '申请时间', value: af.created_at },
      ],
    })
  }
  return {
    id: t.id,
    order_no: '',
    type: 'product',
    status,
    role,
    source: 'real',
    origin: product?.seller_name || '低空商城',
    kind_label: '商品',
    title: product?.title || '商品订单',
    subtitle: productSubtitle(product),
    amount_fen: t.amount_fen ?? 0,
    quantity_label: '共 1 件',
    due_text: STATUS_DUE_TEXT[status] || '待处理',
    // 列表状态标签：售后单显示售后状态（结案单不回「待评价」）
    status_text: af ? af.status : undefined,
    action: statusAction(status, role, af),
    image: productImage(product) || '/static/home/demand-solar.jpg',
    created_at: t.created_at,
    detail: { sections },
    // 售后区块：aftersale.vue 直接消费（type/status/amount_fen/created_at/progress/description）
    aftersale: af,
  }
}

// 主操作按钮文案：按交易角色区分（买家/卖家视图提示，交易功能上线前为中性文案）。
// 有售后记录（af 非空）的订单无论当前状态一律「查看售后」——结案单（approved/rejected）状态已回 completed。
function statusAction(status, role, af) {
  if (af) return '查看售后'
  if (status === 'cancelled') return '已取消'
  if (role === 'seller') {
    if (status === 'paid') return '已付款'
    if (status === 'pending') return '等待付款'
    if (status === 'shipped') return '待确认'
    if (status === 'completed') return '已完成'
    return '查看售后'
  }
  if (status === 'pending') return '待支付'
  if (status === 'paid') return '待发货'
  if (status === 'shipped') return '待收货'
  if (status === 'completed') return '去评价'
  return '查看售后'
}

/* ================= 对外主接口 ================= */

// 加载订单列表（按状态 + 类型筛选）。返回 Promise<Order[]>
// status 传 STATUS_KEYS 中的 key 或 'all'；order_type 传 ORDER_TYPES 中的值。
export async function loadOrders({ status = 'all', order_type = 'all' } = {}) {
  const real = await fetchRealOrders()

  let orders = []
  if (real.length) {
    const pMap = await fetchProductMap()
    orders = real.map((t) => normalizeRealOrder(t, pMap[t.product_id] || null))
  }
  orders = orders.map(applyReviewed)

  // 售后归类规则：有售后记录（aftersale 非空）的订单归「退款/售后」——
  // 审核结案后 status 回 completed，但仍是售后单，不应出现在「待评价」。
  const statusMatch = (o) => {
    if (status === 'all') return true
    if (status === 'aftersale') return !!o.aftersale
    if (status === 'completed') return o.status === 'completed' && !o.aftersale
    return o.status === status
  }
  const typeMatch = (o) => order_type === 'all' || o.type === order_type
  return orders.filter(statusMatch).filter(typeMatch)
}

// 按状态统计各入口角标数（含类型筛选），用于订单中心五状态角标。混合语义：
// ① pending/paid/shipped = 待办数量：状态流转（付款/发货/收货）后移出，查看不消失；
// ② completed（待评价）= 提醒：查看即消（已读时间），评价完也消；
// ③ aftersale（退款/售后）= 提醒：查看即消，仅统计待审核单，审核结案后消失；
// ④ 售后结案单（approved/rejected，状态回 completed）不进任何角标。
export async function loadStatusCounts(order_type = 'all') {
  const real = await fetchRealOrders()
  const all = real.map((t) => {
    const asPending = t.aftersale_status === 'pending'
    const reviewed = !t.aftersale_status && t.status === 'completed' && isReviewed(t.id)
    const closed = !!t.aftersale_status && !asPending
    return {
      status: asPending ? 'aftersale' : (reviewed ? 'reviewed' : (closed ? 'closed' : (t.status || 'pending'))),
      type: 'product',
      created_at: parseTs(t.created_at),
    }
  })

  const counts = { pending: 0, paid: 0, shipped: 0, completed: 0, aftersale: 0 }
  all.forEach((o) => {
    if (order_type !== 'all' && o.type !== order_type) return
    if (counts[o.status] === undefined) return
    // 待办型（pending/paid/shipped）恒计数；提醒型（completed/aftersale）只计未读
    const isReminder = o.status === 'completed' || o.status === 'aftersale'
    if (!isReminder || o.created_at > getSeenTime(o.status)) counts[o.status] += 1
  })
  return counts
}

// 按 ID 查单条订单。返回 Promise<Order|null>
export async function loadOrder(orderId) {
  if (!orderId) return null
  const [real, pMap] = await Promise.all([fetchRealOrders(), fetchProductMap()])
  const t = real.find((r) => r.id === orderId || r.order_no === orderId)
  return t ? applyReviewed(normalizeRealOrder(t, pMap[t.product_id] || null)) : null
}

// 演示数据的「客服」占位：未接入真实客服会话前只提示，不制造假会话
export function toastCustomerService() {
  uni.showToast({ title: '客服接入中，请联系协会秘书处', icon: 'none' })
}

/* ================= 提醒型状态已读存储 ================= */
// 仅「待评价」「退款/售后」使用：进入对应列表即标记已读，角标查看后消失。
// 本地存储；换设备/清缓存后历史单重新视为未读。
const SEEN_KEY = 'order_center_seen'

export function getSeenTime(status) {
  try {
    const raw = uni.getStorageSync(SEEN_KEY)
    return raw && raw[status] ? raw[status] : 0
  } catch (e) { return 0 }
}

export function markStatusSeen(status) {
  let raw = {}
  try {
    const r = uni.getStorageSync(SEEN_KEY)
    if (r && typeof r === 'object' && !Array.isArray(r)) raw = r
  } catch (e) { /* ignore */ }
  raw[status] = Date.now()
  try { uni.setStorageSync(SEEN_KEY, raw) } catch (e) { /* ignore */ }
}

// ISO 时间 → 毫秒时间戳；解析失败按旧单（0）处理，不阻塞角标逻辑
function parseTs(iso) {
  if (!iso) return 0
  const ts = Date.parse(iso)
  return Number.isNaN(ts) ? 0 : ts
}


/* ================= 评价提交（真实接口） ================= */
// 评价走真实接口 POST /api/v1/reviews（target_type/target_id/rating/content 必填）。
// 本地 order_reviews 仅开发环境作为回退（接口不可用时的演示闭环），生产不再写本地。
const REVIEWS_KEY = 'order_reviews'

// 生产环境「已评价」标记：评价提交成功后写入（order_reviews 仅开发环境回退，
// 生产不落评价内容，只记 orderId 用于「待评价」角标/文案消失判定）。
const REVIEWED_KEY = 'order_reviewed_prod'

export function markReviewed(orderId) {
  if (!orderId) return
  let raw = []
  try {
    const r = uni.getStorageSync(REVIEWED_KEY)
    if (Array.isArray(r)) raw = r
  } catch (e) { /* ignore */ }
  const key = String(orderId)
  if (!raw.includes(key)) {
    raw.push(key)
    try { uni.setStorageSync(REVIEWED_KEY, raw) } catch (e) { /* ignore */ }
  }
}

export function isReviewed(orderId) {
  // 开发环境：order_reviews 本地评价数据即视为已评价（保持 DEV 逻辑不变）
  if (getReview(orderId)) return true
  // 生产/通用：读取评价提交成功时写入的标记
  try {
    const raw = uni.getStorageSync(REVIEWED_KEY)
    return Array.isArray(raw) && raw.includes(String(orderId))
  } catch (e) { return false }
}

export function getReview(orderId) {
  if (!import.meta.env.DEV) return null
  try {
    const raw = uni.getStorageSync(REVIEWS_KEY)
    if (raw && typeof raw === 'object' && !Array.isArray(raw)) return raw[orderId] || null
  } catch (e) { /* ignore */ }
  return null
}

function saveReviewLocal(orderId, { rating, content }) {
  let raw = {}
  try {
    const r = uni.getStorageSync(REVIEWS_KEY)
    if (r && typeof r === 'object' && !Array.isArray(r)) raw = r
  } catch (e) { /* ignore */ }
  raw[orderId] = { rating, content, created_at: fmtDate(new Date().toISOString()) }
  try { uni.setStorageSync(REVIEWS_KEY, raw) } catch (e) { /* ignore */ }
}

// 提交订单评价：优先真实接口；仅开发环境接口失败时回退本地存储（不阻断演示闭环）。
// 成功后统一写入「已评价」标记（生产环境的角标/文案消失依赖它）。
export async function submitReview(orderId, { rating, content }) {
  try {
    await request({
      url: '/api/v1/reviews',
      method: 'POST',
      data: { target_type: 'order', target_id: String(orderId), rating, content: content || '' },
    })
    if (import.meta.env.DEV) saveReviewLocal(orderId, { rating, content })
    markReviewed(orderId)
  } catch (e) {
    if (!import.meta.env.DEV) throw e
    saveReviewLocal(orderId, { rating, content })
    markReviewed(orderId)
  }
}

// 已完成订单若已有本地评价：展示层标记「已评价」。
// 不修改 status 字段（避免破坏状态筛选与角标统计），仅覆盖展示文案。
// 售后单（order.aftersale 非空）不适用——结案单状态也是 completed，但应显示售后状态而非「已评价」。
function applyReviewed(order) {
  if (!order || order.status !== 'completed' || order.aftersale || !isReviewed(order.id)) return order
  const detail = order.detail
    ? {
        ...order.detail,
        hero: order.detail.hero
          ? { ...order.detail.hero, sub: String(order.detail.hero.sub || '').replace('待评价', '已评价') }
          : order.detail.hero,
        sections: (order.detail.sections || []).map((s) => ({
          ...s,
          rows: (s.rows || []).map((r) =>
            r.label === '订单状态' ? { ...r, value: '已评价' } : r
          ),
        })),
      }
    : order.detail
  return { ...order, status_text: '已评价', action: '已评价', due_text: '已完成评价，感谢你的反馈', detail }
}
