package memory

import (
	"context"
	"testing"
	"time"

	"drone-platform/internal/domain"
)

// 搜索地区匹配回归：需求按 district（区县名）可被搜索命中，
// 企业按 address/industry_category 可被搜索命中（此前只匹配标题/名称，
// 搜"永川"等区县名无结果）。
func TestSearchMatchesDistrictAndAddress(t *testing.T) {
	ctx := context.Background()

	// 需求：标题不含区县，但 district=永川区 → 搜"永川"应命中
	dRepo := NewDemandRepository(nil)
	now := time.Now()
	if _, err := dRepo.Create(ctx, domain.Demand{
		ID: "d-yc", PublisherID: "u-1", Title: "低空巡检服务", Contact: "138",
		District: "永川区", Status: domain.DemandPublished, Version: 1,
		CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatalf("create demand: %v", err)
	}
	// 未发布的不应出现在搜索结果（List 只返回 published）
	_, _ = dRepo.Create(ctx, domain.Demand{
		ID: "d-yc-pending", PublisherID: "u-1", Title: "永川待审需求", Contact: "138",
		District: "永川区", Status: domain.DemandPending, Version: 1,
		CreatedAt: now, UpdatedAt: now,
	})

	hits, err := dRepo.Search(ctx, "永川")
	if err != nil {
		t.Fatalf("search demands: %v", err)
	}
	found := false
	for _, d := range hits {
		if d.ID == "d-yc" {
			found = true
		}
		if d.ID == "d-yc-pending" {
			t.Fatal("pending demand must not appear in search")
		}
	}
	if !found {
		t.Fatal("search by district must find published demand")
	}

	// 企业：名称不含区县，但 address 含永川 → 搜"永川"应命中
	eRepo := NewEnterpriseRepository(nil)
	if _, err := eRepo.Create(ctx, domain.Enterprise{
		ID: "e-yc", OwnerUserID: "u-1", Name: "某某科技公司", Address: "重庆市永川区人民大道1号",
		Status: domain.EnterpriseApproved, Version: 1, CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatalf("create enterprise: %v", err)
	}
	ehits, err := eRepo.Search(ctx, "永川")
	if err != nil {
		t.Fatalf("search enterprises: %v", err)
	}
	foundE := false
	for _, e := range ehits {
		if e.ID == "e-yc" {
			foundE = true
		}
	}
	if !foundE {
		t.Fatal("search by address must find enterprise")
	}

	// 按行业分类搜（如"测绘"）也应命中
	_, _ = eRepo.Create(ctx, domain.Enterprise{
		ID: "e-cate", OwnerUserID: "u-1", Name: "测绘服务公司", IndustryCategory: "测绘",
		Status: domain.EnterpriseApproved, Version: 1, CreatedAt: now, UpdatedAt: now,
	})
	chits, _ := eRepo.Search(ctx, "测绘")
	foundC := false
	for _, e := range chits {
		if e.ID == "e-cate" {
			foundC = true
		}
	}
	if !foundC {
		t.Fatal("search by industry_category must find enterprise")
	}
}
