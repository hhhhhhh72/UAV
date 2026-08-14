// ============================================================
// 展会排期 · 演示数据（mock）
// ------------------------------------------------------------
// 【接口替换点】后端就绪后：
//   列表：GET /api/v1/exhibitions          (params: page / page_size)
//   详情：GET /api/v1/exhibitions/{id}
//   展位：GET /api/v1/exhibitions/{id}/booths
//   申请：POST /api/v1/exhibitions/{id}/booths  (body: booth_number / exhibit_name / exhibit_desc)
// 页面已在 list.vue / detail.vue / booth.vue 中标注替换位置（搜索“接口替换点”）。
// 字段约定遵循后端 snake_case（exhibitions / exhibition_booths），前端做优雅降级映射。
// ⚠️ 分类/状态枚举必须与后端一致，禁止另立枚举。
// ============================================================

// 分类枚举（后端 Exhibition.category）
export const EXPO_CATEGORY_LABEL = {
  drone_show: '无人机展',
  equipment_expo: '装备展',
  innovation_week: '创新周',
}

// 状态枚举（后端 Exhibition.status；draft 不对外展示）
export const EXPO_STATUS_LABEL = {
  recruiting: '报名中',
  underway: '进行中',
  ended: '已结束',
}

// 展位申请状态（后端 ExhibitionBooth.status）
export const BOOTH_STATUS_LABEL = {
  applied: '待审核',
  approved: '已通过',
  paid: '已支付',
  rejected: '未通过',
}

export const MOCK_EXHIBITIONS = [
  {
    id: 'demo-expo-1',
    title: '2026 重庆低空经济产业博览会',
    category: 'drone_show',
    description: '汇聚低空经济全产业链头部企业与创新力量，设整机、零部件、飞控、载荷、运营服务等主题展区，同期举办产业论坛、供需对接会与新品发布活动，预计参展企业 200 家、专业观众 3 万人。',
    location: '重庆国际博览中心 · N1-N3 馆',
    start_date: '2026-10-15T09:00:00+08:00',
    end_date: '2026-10-17T17:00:00+08:00',
    booth_count: 200,
    booth_price_fen: 800000,
    organizer: '重庆市无人机产业协会',
    cover_url: '',
    status: 'recruiting',
  },
  {
    id: 'demo-expo-2',
    title: '国际无人机应用及防控大会',
    category: 'equipment_expo',
    description: '面向无人机行业应用与公共安全防控的国际性专业展会，聚焦警用、消防、应急、巡检等场景装备与解决方案，同期举办应用案例发布与供需对接。',
    location: '悦来国际会展城',
    start_date: '2026-08-22T09:00:00+08:00',
    end_date: '2026-08-24T17:00:00+08:00',
    booth_count: 120,
    booth_price_fen: 1200000,
    organizer: '重庆市无人机产业协会',
    cover_url: '',
    status: 'underway',
  },
  {
    id: 'demo-expo-3',
    title: '低空飞行创新周',
    category: 'innovation_week',
    description: '以低空飞行器新技术、新场景为主题的活动周，包含飞行演示、创新路演、试飞体验与产业沙龙，推动低空经济创新成果落地。',
    location: '龙兴通用机场',
    start_date: '2026-11-05T09:00:00+08:00',
    end_date: '2026-11-08T17:00:00+08:00',
    booth_count: 80,
    booth_price_fen: 600000,
    organizer: '重庆市无人机产业协会',
    cover_url: '',
    status: 'recruiting',
  },
  {
    id: 'demo-expo-4',
    title: '西部航模公开赛暨航空嘉年华',
    category: 'drone_show',
    description: '面向航模与无人机爱好者的公开赛事与嘉年华活动，设置多组别竞技与飞行表演，配套航空科普体验区。',
    location: '北碚体育中心',
    start_date: '2026-06-18T09:00:00+08:00',
    end_date: '2026-06-20T17:00:00+08:00',
    booth_count: 60,
    booth_price_fen: 350000,
    organizer: '重庆市无人机产业协会',
    cover_url: '',
    status: 'ended',
  },
  {
    id: 'demo-expo-5',
    title: '智慧城市低空应用展',
    category: 'equipment_expo',
    description: '聚焦低空经济在城市治理、物流配送、文旅消费等场景的应用展示，呈现低空应用整体解决方案。',
    location: '两江新区数字经济产业园',
    start_date: '2026-04-11T09:00:00+08:00',
    end_date: '2026-04-13T17:00:00+08:00',
    booth_count: 90,
    booth_price_fen: 500000,
    organizer: '重庆市无人机产业协会',
    cover_url: '',
    status: 'ended',
  },
]

// 展位演示数据（仅 demo-expo-1 预置部分已订展位；其余展会视为全部可选）
export const MOCK_BOOTHS_BY_EXPO = {
  'demo-expo-1': [
    { id: 'mb1', exhibition_id: 'demo-expo-1', exhibitor_id: '', booth_number: 'A3', exhibit_name: '宗申航空动力', exhibit_desc: '', status: 'approved' },
    { id: 'mb2', exhibition_id: 'demo-expo-1', exhibitor_id: '', booth_number: 'A4', exhibit_name: '云之翼通航', exhibit_desc: '', status: 'paid' },
    { id: 'mb3', exhibition_id: 'demo-expo-1', exhibitor_id: '', booth_number: 'A6', exhibit_name: '重大飞行器研究院', exhibit_desc: '', status: 'approved' },
    { id: 'mb4', exhibition_id: 'demo-expo-1', exhibitor_id: '', booth_number: 'B1', exhibit_name: '精微无人机科技', exhibit_desc: '', status: 'approved' },
    { id: 'mb5', exhibition_id: 'demo-expo-1', exhibitor_id: '', booth_number: 'B4', exhibit_name: '两江云动科技', exhibit_desc: '', status: 'applied' },
    { id: 'mb6', exhibition_id: 'demo-expo-1', exhibitor_id: '', booth_number: 'C2', exhibit_name: '巴山通用航空', exhibit_desc: '', status: 'approved' },
    { id: 'mb7', exhibition_id: 'demo-expo-1', exhibitor_id: '', booth_number: 'C3', exhibit_name: '三峡无人机学院', exhibit_desc: '', status: 'approved' },
    { id: 'mb8', exhibition_id: 'demo-expo-1', exhibitor_id: '', booth_number: 'C5', exhibit_name: '飞宇智能装备', exhibit_desc: '', status: 'applied' },
    { id: 'mb9', exhibition_id: 'demo-expo-1', exhibitor_id: '', booth_number: 'D1', exhibit_name: '宗申航空动力', exhibit_desc: '', status: 'approved' },
    { id: 'mb10', exhibition_id: 'demo-expo-1', exhibitor_id: '', booth_number: 'D6', exhibit_name: '两江云动科技', exhibit_desc: '', status: 'paid' },
  ],
}

// ===== 公共格式化 / 展位格子工具（页面共用，禁止另立实现） =====

// 时间格式化：RFC3339 / 'YYYY-MM-DD' → 'MM.DD'（跨年补年份）
export const fmtDay = (v) => {
  if (!v) return ''
  const s = String(v).slice(0, 10)
  const p = s.split('-')
  if (p.length !== 3) return s
  const now = new Date()
  const y = Number(p[0]) !== now.getFullYear() ? p[0] + '.' : ''
  return y + p[1] + '.' + p[2]
}

// 展期文案：'10.15 - 10.17'
export const fmtRange = (start, end) => {
  const a = fmtDay(start)
  const b = fmtDay(end)
  if (a && b) return a + ' - ' + b
  return a || b || ''
}

// 分（fen）→ '¥8,000'（正则千分位，兼容小程序 JS 引擎）
export const fmtFen = (fen) => {
  const n = Number(fen || 0)
  if (!n) return ''
  const yuan = Math.round(n / 100)
  return '¥' + String(yuan).replace(/\B(?=(\d{3})+(?!\d))/g, ',')
}

// 分类 → 渐变 class（占位视觉，真实数据以 cover_url 替换）
export const gradOfCategory = (category) => {
  const map = { drone_show: 'gd-1', equipment_expo: 'gd-2', innovation_week: 'gd-4' }
  return map[category] || 'gd-5'
}

// 生成展位格子：booth_count 个，编号 A1..；超过上限时截断并提示
// occupiedSet：已订展位号集合（大小写不敏感）
export const buildBoothCells = (boothCount, occupiedSet = new Set(), cap = 60) => {
  const total = Math.max(0, Number(boothCount) || 0)
  if (!total) return { cells: [], capped: false, total: 0 }
  const list = []
  const rows = ['A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N']
  const shown = Math.min(total, cap)
  for (let i = 0; i < shown; i++) {
    const row = rows[Math.floor(i / 24)] || 'Z'
    const col = (i % 24) + 1
    const no = row + col
    const key = no.toLowerCase()
    list.push({ no, occupied: occupiedSet.has(no) || occupiedSet.has(key) })
  }
  return { cells: list, capped: total > cap, total }
}
