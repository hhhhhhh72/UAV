// 发布页数据与业务配置
// 唯一设计基准：designs/publish-prototype-rebuild/publish-prototype-v2.html
// 后端接口未接入前，发布内容以本地 storage 持久化，状态转换在本模块内可运行；
// 接入后端后仅需替换 getPosts/savePosts 与提交动作，页面结构与视觉不变。

import { SERVICE_PROD_TYPES } from './enums'

const STORAGE_KEY = 'publish_posts'

// ==================== 商品或服务类型映射 ====================
// 发布页选项（中文，见 TYPES.product 的 productType 字段）→ 后端 domain.ProductType 枚举（7 类）。
// 这是**唯一**一份中文 → 枚举映射：pages/publish/preview.vue 曾自带 PROD_TYPE_MAP，
// utils/hallData.js 又自带一份中文 → 大厅分类表，三份各改各的，
// 于是「维修服务」在一处算服务、在另一处落进「配件」页签。新增类型只改这里。
export const PROD_TYPE_MAP = {
  整机: 'drone',
  零部件: 'part',
  载荷设备: 'part',
  租赁设备: 'part',
  维修服务: 'repair',
  航拍服务: 'aerial',
  试飞测试: 'test_fly',
  检测标定: 'calibration',
  空域协调: 'airspace',
}

// 当前选中的「商品或服务类型」选项是否属于服务类。
// 服务没有"成色"概念：后端 internal/service/trading.go:159-167 对空成色直接归一为 'new'，
// 本就不要求填写；前端却把它标成必填（f[4]=true），等于逼着卖航拍服务的人选一个"二手 95 新"。
export function isServiceTypeOption(label) {
  const prodType = PROD_TYPE_MAP[String(label || '').trim()] || ''
  return SERVICE_PROD_TYPES.indexOf(prodType) >= 0
}

// prod_type → 发布页选项（PROD_TYPE_MAP 的反查）。编辑已发布商品回填时用。
// 注意：part 反查只会得到「零部件」——后端把 零部件/载荷设备/租赁设备 都存成 part，
// 三者在库里不可区分（信息丢失），回填只能取其一。
export function prodTypeToOption(prodType) {
  const t = String(prodType || '').trim()
  for (const k of Object.keys(PROD_TYPE_MAP)) {
    if (PROD_TYPE_MAP[k] === t) return k
  }
  return ''
}

// ==================== 类型配置 ====================

// ==================== 类型配置 ====================
// sections[].fields: [id, label, placeholder, kind(input|select|textarea), required, options?, rule?]
//   rule: 'phone'（手机号/座机）| 'number'（非负数字，最多两位小数）——为空不校验（选填留空允许）
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
          ['district', '所在地区', '选择重庆区县', 'select', true, ['渝中区', '大渡口区', '江北区', '沙坪坝区', '九龙坡区', '南岸区', '北碚区', '渝北区', '巴南区', '两江新区', '长寿区', '江津区', '合川区', '永川区', '南川区', '綦江区', '大足区', '璧山区', '铜梁区', '潼南区', '荣昌区', '开州区', '梁平区', '武隆区', '万州区', '黔江区', '涪陵区', '奉节县', '云阳县', '忠县', '垫江县', '丰都县', '城口县', '巫山县', '巫溪县', '石柱县', '秀山县', '酉阳县', '彭水县']],
          ['date', '期望作业时间（选填）', '选择时间范围', 'select', false, ['一周内', '两周内', '本月内', '可协商']],
        ],
      },
      {
        title: '预算与要求',
        fields: [
          ['budget_min', '预算下限（元，选填）', '例如：5000；不填或与上限都空 = 面议', 'input', false, undefined, 'number'],
          ['budget_max', '预算上限（元，选填）', '例如：20000；0 或与下限同空 = 面议', 'input', false, undefined, 'number'],
          ['aircraft', '机型要求（选填）', '如：多旋翼、固定翼（逗号分隔）', 'input', false],
          ['pilot_count', '飞手数量（选填）', '如：3；0 或空 = 不限', 'input', false, undefined, 'number'],
          ['contact', '联系人电话', '用于审核通过后的对接', 'input', true, undefined, 'phone'],
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
        upload: true,
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
          ['contact', '企业对接人（选填）', '姓名或部门 + 电话', 'input', false],
        ],
      },
    ],
  },
  product: {
    // name 是 form.vue:6/22 的导航标题与表单大标题，short 是 preview.vue:11 的「… · 发布预览」。
    // 发布入口卡片（pages/publish/index.vue:82）早已叫「发布商品与服务」，这里不改就会出现
    // 「从『发布商品与服务』点进去、页面标题写着『发布商品设备』」的自我矛盾。
    name: '发布商品与服务', short: '商品与服务', color: 'product',
    desc: '先选发布类型，表单只显示该类型要填的字段',
    steps: ['基本信息', '价格与交付'],
    stepped: true,
    // 字段数组第 8 位（f[7]）是 scope：'goods' 只在「商品」下出现，
    // 'service' 只在「服务能力」下出现，留空则两分支共用。
    // 实物商品与服务能力是两种业务对象，字段集几乎不重叠（商品有成色/型号/物流，
    // 服务有类目/覆盖区域/计价单位）。此前硬塞进一个表单，只能靠一堆条件显隐打补丁，
    // 结果每个人都看到一半用不上的字段 —— 改为第一步显式分流。
    //
    // 注意：**后端仍是同一张 drone_products 表**（服务并入商品，migration 000110），
    // 分流只发生在表单层，两个分支都提交 POST /api/v1/products。
    sections: [
      {
        title: '基本信息', note: '发布类型决定下面出现哪些字段',
        fields: [
          ['bizKind', '发布类型', '', 'segment', true, ['商品', '服务能力']],
          ['title', '标题', '商品：DJI M350 RTK 行业套装 / 服务：电力线路巡检', 'input', true],
          // 商品：后端 prod_type ∈ {drone, part}。
          // 载荷设备 / 租赁设备同样落 part（见 preview.vue 的 PROD_TYPE_MAP）——
          // 后端只到这一粒度，三者在库里不可区分。
          ['productType', '商品类型', '选择商品类型', 'select', true,
            ['整机', '零部件', '载荷设备', '租赁设备'], undefined, 'goods'],
          // 服务：后端 prod_type ∈ {repair, aerial, test_fly, calibration, airspace}。
          // 此前小程序只能产出 drone/part/repair 三类，航拍/试飞/检测/空域**根本发不出来**。
          ['serviceType', '服务类型', '选择服务类型', 'select', true,
            ['维修服务', '航拍服务', '试飞测试', '检测标定', '空域协调'], undefined, 'service'],
          ['condition', '成色', '选择商品状态', 'select', true,
            ['全新未拆封', '全新', '二手 95 新', '二手 90 新'], undefined, 'goods'],
          ['brand', '品牌/型号', '例如：DJI / M350 RTK', 'input', true, undefined, undefined, 'goods'],
          ['category', '服务类目（选填）', '如：巡检 / 测绘 / 应急', 'input', false, undefined, undefined, 'service'],
        ],
        upload: true,
      },
      {
        title: '价格',
        fields: [
          // 价格方式必须显式选择：此前靠"售价留空 = 面议"隐式表达，
          // 后端无法区分"卖家选了面议"和"卖家填了 0 元"，设备区卡片还把它显示成 ¥0。
          ['priceMode', '价格方式', '选择报价方式', 'select', true, ['明码标价', '面议']],
          ['price', '售价 / 报价', '例如：68000；选择面议时留空', 'input', false, undefined, 'number'],
          // 报价单位只对服务有意义：没有它"¥800"不知道是每次还是每天。
          ['unit', '报价单位（选填）', '如：次 / 天 / 公里', 'input', false, undefined, undefined, 'service'],
        ],
      },
      {
        title: '交付方式',
        fields: [
          // 交付方式只对实物商品有意义：它决定下单时是否强制收货地址
          //（OrderNeedsReceiver，internal/service/trading.go:81）。
          // 服务类的兜底是 prod_type 不属于 drone/part → 本就不需要地址。
          ['delivery', '交付方式（选填）', '选择交付方式', 'select', false,
            ['自提', '同城配送', '物流发货', '可协商'], undefined, 'goods'],
          // 服务区域：语义是"服务能覆盖到哪"，不是精确区县。
          // 生产数据里覆盖范围式取值占多数（重庆及西南 / 川渝地区 / 重庆及周边 / 重庆全域），
          // 单选区县表达不了；需求/课程那两个区县选择器写的是**另一列**（district）——那边要规范
          // 区县是因为需求大厅按区县筛，而供给大厅没有地区筛选。
          // 所以做预置范围 +「其他」手填：主流取值规范化，同时不丢表达能力。
          ['region', '服务区域（选填）', '选择服务覆盖范围', 'select', false,
            ['重庆全域', '重庆主城', '渝东北', '渝东南', '川渝地区', '全国', '其他'], undefined, 'service'],
          // 选「其他」时才出现。required=true 只在字段可见时生效（见 form.vue fieldVisible）：
          // 没选「其他」时它根本不渲染，requiredMissing 也就不会拿它拦人。
          ['regionOther', '自定义服务区域', '如：渝东北 + 渝东南', 'input', true, undefined, undefined, 'service'],
          ['description', '说明', '商品填配置清单、使用情况；服务填服务内容、交付成果', 'textarea', true],
        ],
        // 详情图上传位：与「商品与服务」分区的顶部图集是两组独立的图
        uploadDetail: true,
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
        upload: true,
      },
      {
        title: '开班与地点',
        fields: [
          ['district', '所属区县', '选择重庆区县', 'select', true, ['渝中区', '大渡口区', '江北区', '沙坪坝区', '九龙坡区', '南岸区', '北碚区', '渝北区', '巴南区', '两江新区', '长寿区', '江津区', '合川区', '永川区', '南川区', '綦江区', '大足区', '璧山区', '铜梁区', '潼南区', '荣昌区', '开州区', '梁平区', '武隆区', '万州区', '黔江区', '涪陵区', '奉节县', '云阳县', '忠县', '垫江县', '丰都县', '城口县', '巫山县', '巫溪县', '石柱县', '秀山县', '酉阳县', '彭水县']],
          ['location', '培训地点', '例如：渝北区金开大道 68 号', 'input', true],
          ['schedule', '开班时间（选填）', '选择最近开班安排', 'select', false, ['8 月 25 日开班', '9 月 8 日开班', '滚动开班', '可预约']],
        ],
      },
      {
        title: '招生信息',
        fields: [
          ['price', '课程价格', '例如：9800', 'input', true, undefined, 'number'],
          ['duration', '培训天数', '例如：25', 'input', true, undefined, 'number'],
          ['quota', '招生名额', '例如：20', 'input', true, undefined, 'number'],
          ['description', '课程介绍', '培训内容、适合人群、颁发证书等', 'textarea', true],
        ],
      },
    ],
  },
}

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

// 旧版内置示例数据的 id（已移除），用于清理早期版本写入本地的演示数据
const LEGACY_SEED_IDS = ['demand-1', 'service-1', 'product-1', 'course-1', 'draft-1']

export function getPosts() {
  const list = readPosts()
  if (!list) return []
  const cleaned = list.filter((p) => !LEGACY_SEED_IDS.includes(p.id))
  if (cleaned.length !== list.length) savePosts(cleaned)
  return cleaned
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

// 「我的发布」三个接口统一下拉 100 条。分页是否被截断靠它判断（见 pruneStalePointers）。
export const MINE_PAGE_SIZE = 100

// pruneStalePointers 清掉"指向已不存在后端实体"的本地死指针，返回清理条数。
//
// 为什么需要：本地记录里带 backendId 的只是后端实体的**指针**，后端那份才是真值。
// 用户删掉需求/商品后指针就成了死链——所有"后端已有的就不重复展示/计数"的判据都是
// id 比对，id 不在后端集合里就会被当成"后端没有"而漏出来，变成幽灵条目。
//
// complete 必须由调用方明确传 true，且只在**拿到完整权威列表**时才传：
//   - 某个接口失败时会返回空数组，此时把"空"当成"实体已删"会把有效指针一起清掉；
//   - 分页被截断（返回条数 == page_size）时，第 101 条之后的实体只是没返回，
//     不是不存在，同样会误清。
// 拿不准就别清：展示层已经能把死指针挡在外面，物理清理只是锦上添花。
//
// 只清"带 backendId 且非草稿"的记录。不带 backendId 的是纯本地记录（后端看不到，
// 删了就真没了），草稿同理一律不碰。
export function pruneStalePointers(liveIds, options) {
  if (!options || options.complete !== true) return 0
  const live = new Set((liveIds || []).map((v) => String(v)))
  const list = getPosts()
  const kept = list.filter(
    (p) => !(p.backendId && p.statusKey !== 'draft' && !live.has(String(p.backendId)))
  )
  const removed = list.length - kept.length
  if (removed > 0) savePosts(kept)
  return removed
}

// ==================== 元数据/文案 ====================
// 商品/服务共用的 meta 分支：发布类型决定读哪个 key
//（商品 productType / 服务能力 serviceType）。
// computeMeta 与 computePreviewMeta 两个函数都要用，抽出来免得两份各改各的。
function productMeta(v) {
  const t = v.productType || v.serviceType || '类型待定'
  // 服务没有"成色"这个概念，第二个标签改用服务类目，
  // 否则卖航拍服务的人会看到"商品状态待定"。
  const mid = v.bizKind === '服务能力' ? (v.category || '类目待定') : (v.condition || '商品状态待定')
  return [t, mid, v.price ? v.price + ' 元' : '价格待定']
}

// 列表/详情 meta：三个关键信息标签
export function computeMeta(type, values) {
  const v = values || {}
  switch (type) {
    case 'demand':
      return [v.biz || '未选业务', v.district || '未选地区', v.budget ? '预算 ' + v.budget + ' 元' : '预算可协商']
    case 'service':
      return [v.category || '未选服务', v.range || '服务范围待定', v.quote || '报价待定']
    case 'product':
      return productMeta(v)
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
      return productMeta(v)
    case 'course':
      return [v.certType || '课程类型待定', v.district || '所属区县待定', v.price ? v.price + ' 元' : '价格待定']
  }
  return []
}

// 创建/更新一条发布
export function makePost({ id, type, values, photoCount, statusKey, status, date, note }) {
  const t = TYPES[type]
  const post = {
    id: id || 'post-' + Date.now(),
    type,
    label: t ? t.short : type,
    title: (values && values.title) || '未命名发布内容',
    status: status || '已发布',
    statusKey: statusKey || 'live',
    date: date || '刚刚发布',
    meta: computeMeta(type, values),
    note: note || '',
    values: values || {},
    photoCount: photoCount || 0,
  }
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
