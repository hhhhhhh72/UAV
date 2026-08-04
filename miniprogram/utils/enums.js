// 业务类型统一标签
// 对应后端 internal/domain/biz_standard.go 的术语标准（功能方案修订版 v2）。
// 页面禁止另立类型 map；新增类型先改后端标准，再同步本文件。

// --- 供给类型（ProductType） ---
export const PRODUCT_TYPE_LABEL = {
  drone: '整机',
  part: '零件',
  repair: '维修服务',
  aerial: '航拍服务',
  test_fly: '试飞测试',
  calibration: '检测标定',
  airspace: '空域协调',
}
export const PRODUCT_TYPE_SHORT = { drone: '整机', part: '零件', test_fly: '试飞' }
export const productTypeLabel = (t) => PRODUCT_TYPE_LABEL[t] || t || ''
export const productTypeShort = (t) => PRODUCT_TYPE_SHORT[t] || PRODUCT_TYPE_LABEL[t] || t || ''

// --- 需求类型（BizType）筛选 Tab + 标签 ---
export const BIZ_TYPE_TABS = [
  { label: '全部', value: '' },
  { label: '巡检', value: 'cable_inspection' },
  { label: '植保', value: 'plant_transport' },
  { label: '农药', value: 'spray_pesticide' },
  { label: '租赁', value: 'trade_lease' },
  { label: '清洗', value: 'clean_paint' },
  { label: '其他', value: 'other' },
]
export const BIZ_TYPE_LABEL = {
  cable_inspection: '巡检',
  plant_transport: '植保',
  spray_pesticide: '农药',
  trade_lease: '租赁',
  clean_paint: '清洗',
  other: '其他',
}
export const bizTypeLabel = (t) => BIZ_TYPE_LABEL[t] || t || '其他'

// --- 院校分域（College.coop_type，功能方案修订版 三·五 分域） ---
export const COOP_TYPE_LABEL = { research: '科研合作', talent: '人才培养', both: '综合' }
export const coopTypeLabel = (t) => COOP_TYPE_LABEL[t] || ''

// --- 产业分类（EnterpriseCategory，功能方案系统一） ---
export const ENTERPRISE_CATEGORY_LABEL = {
  drone: '整机',
  part: '零部件',
  flight_ctrl: '飞控',
  payload: '载荷',
  operator: '运营服务',
  college: '实训院校',
  airport: '通航机场',
  inspector: '检测机构',
}
export const enterpriseCategoryLabel = (c) => ENTERPRISE_CATEGORY_LABEL[c] || c || ''
