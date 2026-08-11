// 发布页数据与业务配置
// 唯一设计基准：designs/publish-prototype-rebuild/publish-prototype-v2.html
// 后端接口未接入前，发布内容以本地 storage 持久化，状态转换在本模块内可运行；
// 接入后端后仅需替换 getPosts/savePosts 与提交动作，页面结构与视觉不变。

const STORAGE_KEY = 'publish_posts'

// ==================== 类型配置 ====================
// sections[].fields: [id, label, placeholder, kind(input|select|textarea), required, options?]
export const TYPES = {
  demand: {
    name: '发布需求', short: '需求', color: 'demand',
    desc: '发布具体项目，获得飞手与服务商报价',
    steps: ['需求信息', '作业要求'],
    sections: [
      {
        title: '项目基本信息', note: '用一句话说清要做什么，便于快速匹配',
        fields: [
          ['title', '需求标题', '例如：光伏电站红外巡检', 'input', true],
          ['biz', '业务类型', '选择巡检、植保、测绘等', 'select', true, ['巡检', '植保', '测绘', '航拍', '吊运', '其他']],
        ],
      },
      {
        title: '作业地点与工期',
        fields: [
          ['district', '所在地区', '选择重庆区县', 'select', true, ['南岸区', '渝北区', '沙坪坝区', '两江新区', '江北区']],
          ['date', '期望作业时间', '选择时间范围', 'select', true, ['一周内', '两周内', '本月内', '可协商']],
        ],
      },
      {
        title: '预算与对接',
        fields: [
          ['budget', '项目预算', '例如：20000', 'input', false],
          ['contact', '联系人电话', '用于审核通过后的对接', 'input', true],
          ['description', '需求说明', '作业面积、交付成果、现场限制等', 'textarea', false],
        ],
      },
      { title: '现场资料', upload: true },
    ],
  },
  service: {
    name: '发布服务能力', short: '服务能力', color: 'service',
    desc: '展示团队、设备和可承接服务，让需求方主动联系',
    steps: ['能力概览', '服务与资质'],
    sections: [
      {
        title: '服务名片', note: '用清晰的服务名和品类建立第一印象',
        fields: [
          ['title', '服务名称', '例如：低空电力巡检服务', 'input', true],
          ['category', '服务分类', '选择一项主营服务', 'select', true, ['巡检', '测绘', '航拍', '植保', '吊运', '检测标定']],
        ],
      },
      {
        title: '覆盖与报价',
        fields: [
          ['range', '服务范围', '例如：重庆市内及周边', 'input', true],
          ['quote', '报价方式', '选择报价方式', 'select', true, ['按项目报价', '按天报价', '按架次报价', '面议']],
        ],
      },
      {
        title: '能力与资质', note: '把设备、人员和合规情况说具体，提升信任',
        fields: [
          ['equipment', '设备与载荷', '例如：M350 RTK、热成像相机', 'textarea', true],
          ['cert', '飞手/企业资质', '例如：持证飞手 3 人，已投保', 'textarea', true],
          ['contact', '企业对接人', '姓名或部门 + 电话', 'input', true],
        ],
      },
    ],
  },
  product: {
    name: '发布商品设备', short: '商品设备', color: 'product',
    desc: '按商品逻辑补齐型号、成色、价格和交付方式',
    steps: ['商品信息', '价格与交付'],
    stepped: true,
    sections: [
      {
        title: '商品信息', note: '商品标题、类型和实拍图决定买家是否继续查看',
        fields: [
          ['title', '商品标题', '例如：DJI M350 RTK 行业套装', 'input', true],
          ['productType', '商品类型', '选择整机、零件或服务', 'select', true, ['整机', '零部件', '载荷设备', '租赁设备', '维修服务']],
          ['condition', '成色', '选择商品状态', 'select', true, ['全新未拆封', '全新', '二手 95 新', '二手 90 新']],
        ],
        upload: true,
      },
      {
        title: '价格与库存',
        fields: [
          ['price', '售价', '例如：68000', 'input', true],
          ['stock', '可售数量', '例如：1', 'input', true],
          ['brand', '品牌/型号', '例如：DJI / M350 RTK', 'input', true],
        ],
      },
      {
        title: '交付方式',
        fields: [
          ['delivery', '交付方式', '选择交付方式', 'select', true, ['自提', '同城配送', '物流发货', '可协商']],
          ['description', '商品说明', '配置清单、使用情况、售后承诺等', 'textarea', true],
        ],
      },
    ],
  },
  course: {
    name: '发布培训课程', short: '培训课程', color: 'course',
    desc: '用课程、证书、日期与招生信息回答学员的核心问题',
    steps: ['课程与机构', '开班与招生'],
    stepped: true,
    sections: [
      {
        title: '课程与机构', note: '先建立课程价值与办学主体的可信度',
        fields: [
          ['title', '课程标题', '例如：CAAC 多旋翼驾驶员执照班', 'input', true],
          ['certType', '证书类型', '选择对应证书', 'select', true, ['CAAC 民航局执照', 'AOPA 执照', '大疆 UTC 证书', '职业技能等级']],
          ['org', '培训机构', '填写机构全称', 'input', true],
        ],
      },
      {
        title: '开班与地点',
        fields: [
          ['district', '所属区县', '选择重庆区县', 'select', true, ['南岸区', '渝北区', '江北区', '两江新区', '沙坪坝区']],
          ['location', '培训地点', '例如：渝北区金开大道 68 号', 'input', true],
          ['schedule', '开班时间', '选择最近开班安排', 'select', true, ['8 月 25 日开班', '9 月 8 日开班', '滚动开班', '可预约']],
        ],
      },
      {
        title: '招生信息',
        fields: [
          ['price', '课程价格', '例如：9800', 'input', true],
          ['duration', '培训天数', '例如：25', 'input', true],
          ['quota', '招生名额', '例如：20', 'input', true],
          ['description', '课程介绍', '培训内容、适合人群、颁发证书等', 'textarea', true],
        ],
      },
    ],
  },
}

// ==================== 示例发布数据（原型真实样例） ====================
const SEED_POSTS = [
  { id: 'demand-1', type: 'demand', label: '需求', title: '光伏电站红外巡检', status: '审核中', statusKey: 'pending', date: '今日 10:24 提交', meta: ['巡检', '南岸区', '预算 2 万'], note: '平台正在核验项目描述与联系人信息。', values: { title: '光伏电站红外巡检', biz: '巡检', district: '南岸区', date: '两周内', contact: '', budget: '', description: '' }, photoCount: 0 },
  { id: 'service-1', type: 'service', label: '服务能力', title: '低空电力巡检服务', status: '已发布', statusKey: 'live', date: '发布于 8 月 08 日', meta: ['巡检', '重庆及周边', '按项目报价'], note: '已收到 3 条对接意向。', leads: '3 条意向', values: { title: '低空电力巡检服务', category: '巡检', range: '重庆及周边', quote: '按项目报价', equipment: '', cert: '', contact: '' }, photoCount: 0 },
  { id: 'product-1', type: 'product', label: '商品设备', title: 'DJI M350 RTK 行业套装', status: '已发布', statusKey: 'live', date: '发布于 8 月 06 日', meta: ['整机', '二手 90 新', '¥ 68,000'], note: '商品已在设备大厅展示。', leads: '8 次咨询', values: { title: 'DJI M350 RTK 行业套装', productType: '整机', condition: '二手 90 新', price: '68000', stock: '1', brand: 'DJI / M350 RTK', delivery: '物流发货', description: '' }, photoCount: 0 },
  { id: 'course-1', type: 'course', label: '培训课程', title: 'CAAC 多旋翼驾驶员执照班', status: '未通过', statusKey: 'rejected', date: '审核于 8 月 05 日', meta: ['CAAC 民航局执照', '渝北区', '¥ 9,800'], note: '缺少培训机构相关证明，请补充后重新提交。', values: { title: 'CAAC 多旋翼驾驶员执照班', certType: 'CAAC 民航局执照', org: '', district: '渝北区', location: '', schedule: '', price: '9800', duration: '', quota: '', description: '' }, photoCount: 0 },
  { id: 'draft-1', type: 'demand', label: '需求', title: '园区三维建模测绘项目', status: '草稿', statusKey: 'draft', date: '保存于 昨日 17:40', meta: ['测绘', '两江新区', '待完善'], note: '还有作业时间与联系人未填写。', values: { title: '园区三维建模测绘项目', biz: '测绘', district: '两江新区', date: '', contact: '', budget: '', description: '' }, photoCount: 0 },
]

// ==================== 状态/类型文案 ====================
export const STATUS_META = {
  pending: { label: '审核中', cls: 'status-pending', color: '#D56A00' },
  live: { label: '已发布', cls: 'status-live', color: '#219653' },
  rejected: { label: '未通过', cls: 'status-rejected', color: '#FF3B30' },
  draft: { label: '草稿', cls: 'status-draft', color: '#98A2B3' },
}

export const TAB_ORDER = ['all', 'pending', 'live', 'rejected', 'draft']
export const TAB_LABEL = { all: '全部', pending: '审核中', live: '已发布', rejected: '未通过', draft: '草稿' }

export const KIND_ORDER = ['all', 'demand', 'service', 'product', 'course']
export const KIND_LABEL = { all: '全部类型', demand: '需求', service: '服务能力', product: '商品设备', course: '培训课程' }

// ==================== 存储 ====================
function readPosts() {
  try {
    const raw = uni.getStorageSync(STORAGE_KEY)
    if (Array.isArray(raw)) return raw
  } catch (e) { /* ignore */ }
  return null
}

export function getPosts() {
  let list = readPosts()
  if (!list) {
    list = SEED_POSTS.slice()
    try { uni.setStorageSync(STORAGE_KEY, list) } catch (e) { /* ignore */ }
  }
  return list
}

export function savePosts(list) {
  try { uni.setStorageSync(STORAGE_KEY, list) } catch (e) { /* ignore */ }
}

export function getPost(id) {
  return getPosts().find((p) => p.id === id) || null
}

export function upsertPost(post) {
  const list = getPosts()
  const idx = list.findIndex((p) => p.id === post.id)
  if (idx >= 0) list[idx] = post
  else list.unshift(post)
  savePosts(list)
  return post
}

export function removePost(id) {
  const list = getPosts().filter((p) => p.id !== id)
  savePosts(list)
}

export function countByStatus(key) {
  return getPosts().filter((p) => p.statusKey === key).length
}

export function draftPosts() {
  return getPosts().filter((p) => p.statusKey === 'draft')
}

// ==================== 元数据/文案 ====================
// 列表/详情 meta：三个关键信息标签
export function computeMeta(type, values) {
  const v = values || {}
  switch (type) {
    case 'demand':
      return [v.biz || '未选业务', v.district || '未选地区', v.budget ? '预算 ' + v.budget + ' 元' : '预算可协商']
    case 'service':
      return [v.category || '未选服务', v.range || '服务范围待定', v.quote || '报价待定']
    case 'product':
      return [v.productType || '商品类型待定', v.condition || '商品状态待定', v.price ? v.price + ' 元' : '价格待定']
    case 'course':
      return [v.certType || '课程类型待定', v.district || '所属区县待定', v.price ? v.price + ' 元' : '价格待定']
  }
  return []
}

// 预览页 meta：三个关键信息标签（与原型 preview() 一致）
export function computePreviewMeta(type, values) {
  const v = values || {}
  switch (type) {
    case 'demand':
      return [v.biz || '未选业务', v.district || '未选地区', v.budget ? v.budget + ' 元' : '预算可协商']
    case 'service':
      return [v.category || '未选服务', v.range || '服务范围待定', v.quote || '报价待定']
    case 'product':
      return [v.productType || '商品类型待定', v.condition || '商品状态待定', v.price ? v.price + ' 元' : '价格待定']
    case 'course':
      return [v.certType || '课程类型待定', v.district || '所属区县待定', v.price ? v.price + ' 元' : '价格待定']
  }
  return []
}

// 提交后的默认说明文案
export function submitNote(type) {
  return '平台正在核验发布内容与联系人信息。'
}

// 创建/更新一条发布
export function makePost({ id, type, values, photoCount, statusKey, status, date, note }) {
  const t = TYPES[type]
  const post = {
    id: id || 'post-' + Date.now(),
    type,
    label: t ? t.short : type,
    title: (values && values.title) || '未命名发布内容',
    status: status || '审核中',
    statusKey: statusKey || 'pending',
    date: date || '刚刚提交',
    meta: computeMeta(type, values),
    note: note || submitNote(type),
    values: values || {},
    photoCount: photoCount || 0,
  }
  if (post.leads === undefined) delete post.leads
  return post
}

// ==================== 当前编辑表单的跨页传递 ====================
// form -> preview 需要把当前表单值带到预览页（页面间不共享内存实例）
const FORM_KEY = 'publish_form_state'
export function saveFormState(state) {
  try { uni.setStorageSync(FORM_KEY, state) } catch (e) { /* ignore */ }
}
export function loadFormState() {
  try {
    const s = uni.getStorageSync(FORM_KEY)
    return s && typeof s === 'object' ? s : null
  } catch (e) { /* ignore */ }
  return null
}
export function clearFormState() {
  try { uni.removeStorageSync(FORM_KEY) } catch (e) { /* ignore */ }
}
