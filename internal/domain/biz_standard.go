package domain

// ============================================================================
// 业务类型统一标准（功能方案修订版 v2 附录 · 2026-08-04 评审落地）
// ----------------------------------------------------------------------------
// 全平台唯一术语标准：把历史上三套互不兼容的枚举
//   —— 需求类型 BizType（models.go）、供给类型 ProductType（models.go）、
//      企业注册分类 Category（free string）
// 映射到统一的「业务域」概念。
// 约定：
//   1. 线上枚举值保持不变（防破坏既有数据与 API 契约），本文件只做映射与标签；
//   2. 新增任何类型必须先在本文件登记并补映射，禁止各模块另立标签表；
//   3. 前端展示标签统一从这里取（小程序 utils/enums.js 引用本表语义）。
// ============================================================================

// BizDomain 是平台统一业务域（功能方案术语标准的机读值）。
type BizDomain string

const (
	DomainInspection BizDomain = "inspection" // 巡检（电力/管网/交通设施巡检）
	DomainSurvey     BizDomain = "survey"     // 测绘（地形/工程测量）
	DomainSpray      BizDomain = "spray"      // 植保（农林植保作业）
	DomainLogistics  BizDomain = "logistics"  // 物流（低空物流/配送）
	DomainLifting    BizDomain = "lifting"    // 吊运（大载重吊运）
	DomainAerial     BizDomain = "aerial"     // 航拍（航拍服务/影像采集）
	DomainPurchase   BizDomain = "purchase"   // 采购（设备采购需求）
	DomainMaintain   BizDomain = "maintain"   // 维修（维修服务/保养）
	DomainCalibrate  BizDomain = "calibrate"  // 检测标定（适航检测/标定）
	DomainTestFly    BizDomain = "test_fly"   // 试飞测试（场地预约）
	DomainAirspace   BizDomain = "airspace"   // 空域协调（飞行报备/申请代办）
	DomainTraining   BizDomain = "training"   // 培训（课程超市承载，五系统唯一归属）
	DomainTrade      BizDomain = "trade"      // 买卖租赁（二手/租售）
	DomainOther      BizDomain = "other"      // 其他
)

// 业务域中文标签（与功能方案术语一致，供前端展示）。
var bizDomainLabels = map[BizDomain]string{
	DomainInspection: "巡检",
	DomainSurvey:     "测绘",
	DomainSpray:      "植保",
	DomainLogistics:  "物流",
	DomainLifting:    "吊运",
	DomainAerial:     "航拍",
	DomainPurchase:   "采购",
	DomainMaintain:   "维修",
	DomainCalibrate:  "检测标定",
	DomainTestFly:    "试飞测试",
	DomainAirspace:   "空域协调",
	DomainTraining:   "培训",
	DomainTrade:      "买卖租赁",
	DomainOther:      "其他",
}

// BizDomainLabel 返回业务域中文名；未知域回退为 "其他"。
func BizDomainLabel(d BizDomain) string {
	if l, ok := bizDomainLabels[d]; ok {
		return l
	}
	return bizDomainLabels[DomainOther]
}

// --- 映射：需求类型 BizType → 业务域 ---

// BizTypeDomain 将需求类型归一到业务域。
func BizTypeDomain(t BizType) BizDomain {
	switch t {
	case BizCableInspection:
		return DomainInspection
	case BizPlantTransport:
		return DomainLogistics
	case BizSprayPesticide:
		return DomainSpray
	case BizCleanPaint:
		return DomainOther // 清洗粉刷暂归其他，未来如需独立域在本文件登记
	case BizTradeLease:
		return DomainTrade
	default:
		return DomainOther
	}
}

// --- 映射：供给类型 ProductType → 业务域 ---

// ProductTypeDomain 将供给类型归一到业务域。
func ProductTypeDomain(t ProductType) BizDomain {
	switch t {
	case ProductDrone, ProductPart:
		return DomainPurchase // 设备供给对应"采购/设备"业务域
	case ProductRepair:
		return DomainMaintain
	case ProductAerial:
		return DomainAerial
	case ProductTestFly:
		return DomainTestFly
	case ProductCalibration:
		return DomainCalibrate
	case ProductAirspace:
		return DomainAirspace
	default:
		return DomainOther
	}
}

// DomainProductTypes 返回某业务域下的供给类型（撮合/推荐侧可用）。
func DomainProductTypes(d BizDomain) []ProductType {
	switch d {
	case DomainInspection:
		return []ProductType{ProductDrone, ProductAerial}
	case DomainSurvey:
		return []ProductType{ProductDrone, ProductAerial}
	case DomainSpray:
		return []ProductType{ProductDrone}
	case DomainLogistics:
		return []ProductType{ProductDrone}
	case DomainLifting:
		return []ProductType{ProductDrone}
	case DomainAerial:
		return []ProductType{ProductAerial, ProductDrone}
	case DomainPurchase:
		return []ProductType{ProductDrone, ProductPart}
	case DomainMaintain:
		return []ProductType{ProductRepair}
	case DomainCalibrate:
		return []ProductType{ProductCalibration}
	case DomainTestFly:
		return []ProductType{ProductTestFly}
	case DomainAirspace:
		return []ProductType{ProductAirspace}
	case DomainTraining:
		return []ProductType{}
	default:
		return nil
	}
}

// ============================================================================
// 产业分类（供给侧主体分类，与业务域并列的第二条标准轴）
// 功能方案系统一：整机 / 零部件 / 飞控 / 载荷 / 运营服务 / 实训院校 / 通航机场 / 检测机构
// ============================================================================

// EnterpriseCategory 会员注册时的产业分类。
type EnterpriseCategory string

const (
	CategoryDrone      EnterpriseCategory = "drone"       // 整机
	CategoryPart       EnterpriseCategory = "part"        // 零部件
	CategoryFlightCtrl EnterpriseCategory = "flight_ctrl" // 飞控
	CategoryPayload    EnterpriseCategory = "payload"     // 载荷
	CategoryOperator   EnterpriseCategory = "operator"    // 运营服务
	CategoryCollege    EnterpriseCategory = "college"     // 实训院校
	CategoryAirport    EnterpriseCategory = "airport"     // 通航机场
	CategoryInspector  EnterpriseCategory = "inspector"   // 检测机构
)

var enterpriseCategoryLabels = map[EnterpriseCategory]string{
	CategoryDrone:      "整机",
	CategoryPart:       "零部件",
	CategoryFlightCtrl: "飞控",
	CategoryPayload:    "载荷",
	CategoryOperator:   "运营服务",
	CategoryCollege:    "实训院校",
	CategoryAirport:    "通航机场",
	CategoryInspector:  "检测机构",
}

// EnterpriseCategoryLabel 返回产业分类中文名；未知值原样返回。
func EnterpriseCategoryLabel(c EnterpriseCategory) string {
	if l, ok := enterpriseCategoryLabels[c]; ok {
		return l
	}
	return string(c)
}

// NormalizeCategory 将企业注册分类字符串归一为合法分类（未知 → "other"）。
// 历史数据可能为自由文本，落库前用本函数约束。
func NormalizeCategory(c string) EnterpriseCategory {
	cat := EnterpriseCategory(c)
	if _, ok := enterpriseCategoryLabels[cat]; ok {
		return cat
	}
	return EnterpriseCategory("other")
}
