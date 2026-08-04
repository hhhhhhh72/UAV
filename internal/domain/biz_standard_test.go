package domain

import "testing"

// 术语标准红线：所有业务域都必须有中文标签，禁止出现空标签。
func TestAllBizDomainsHaveLabels(t *testing.T) {
	for _, d := range []BizDomain{
		DomainInspection, DomainSurvey, DomainSpray, DomainLogistics, DomainLifting,
		DomainAerial, DomainPurchase, DomainMaintain, DomainCalibrate, DomainTestFly,
		DomainAirspace, DomainTraining, DomainTrade, DomainOther,
	} {
		if label := BizDomainLabel(d); label == "" || label == "其他" && d != DomainOther {
			t.Errorf("domain %q missing label (got %q)", d, label)
		}
	}
	if BizDomainLabel("unknown-domain") != "其他" {
		t.Errorf("unknown domain should fall back to 其他")
	}
}

// 需求类型 ↔ 业务域 映射
func TestBizTypeDomainMapping(t *testing.T) {
	cases := map[BizType]BizDomain{
		BizCableInspection: DomainInspection,
		BizPlantTransport:  DomainLogistics,
		BizSprayPesticide:  DomainSpray,
		BizTradeLease:      DomainTrade,
		BizCleanPaint:      DomainOther,
		BizOther:           DomainOther,
	}
	for in, want := range cases {
		if got := BizTypeDomain(in); got != want {
			t.Errorf("BizTypeDomain(%q) = %q, want %q", in, got, want)
		}
	}
}

// 供给类型 ↔ 业务域 映射
func TestProductTypeDomainMapping(t *testing.T) {
	cases := map[ProductType]BizDomain{
		ProductDrone:       DomainPurchase,
		ProductPart:        DomainPurchase,
		ProductRepair:      DomainMaintain,
		ProductAerial:      DomainAerial,
		ProductTestFly:     DomainTestFly,
		ProductCalibration: DomainCalibrate,
		ProductAirspace:    DomainAirspace,
	}
	for in, want := range cases {
		if got := ProductTypeDomain(in); got != want {
			t.Errorf("ProductTypeDomain(%q) = %q, want %q", in, got, want)
		}
	}
}

// 业务域 → 供给类型：反向映射必须落回原域（round-trip 语义）
func TestDomainProductTypesRoundTrip(t *testing.T) {
	for in, want := range map[BizDomain]ProductType{
		DomainMaintain:  ProductRepair,
		DomainCalibrate: ProductCalibration,
		DomainTestFly:   ProductTestFly,
		DomainAirspace:  ProductAirspace,
	} {
		types := DomainProductTypes(in)
		found := false
		for _, p := range types {
			if p == want {
				found = true
			}
		}
		if !found {
			t.Errorf("DomainProductTypes(%q) missing %q, got %v", in, want, types)
		}
	}
}

// 产业分类标签与归一化
func TestEnterpriseCategory(t *testing.T) {
	if got := EnterpriseCategoryLabel(CategoryDrone); got != "整机" {
		t.Errorf("CategoryDrone label = %q, want 整机", got)
	}
	if got := EnterpriseCategoryLabel("whatever"); got != "whatever" {
		t.Errorf("unknown category label should pass through, got %q", got)
	}
	if got := NormalizeCategory("整机"); got != CategoryDrone {
		// 历史自由文本"整机"不在标准枚举内，应归一为 other
		if got != EnterpriseCategory("other") {
			t.Errorf("NormalizeCategory(整机) = %q, want other", got)
		}
	}
	if got := NormalizeCategory("drone"); got != CategoryDrone {
		t.Errorf("NormalizeCategory(drone) = %q, want %q", got, CategoryDrone)
	}
}
