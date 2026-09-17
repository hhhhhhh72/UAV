package service_test

import (
	"context"
	"errors"
	"testing"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

// 管理端批量上下架：单条条件更新，跳过不该改的行，并如实返回实际改动数。
//
// 此前是前端逐行 PUT 整行——只改一个 status 却把所有列写回去。
func TestBulkSetProductStatusRules(t *testing.T) {
	ctx := context.Background()
	prodRepo := memory.NewProductRepository()
	svc := service.NewTradingService(prodRepo, memory.NewRepairRepository(), nil, nil)

	mk := func(id, check, status string) {
		t.Helper()
		if _, err := prodRepo.Create(ctx, domain.DroneProduct{
			ID: id, SellerID: "seller-1", ProdType: domain.ProductDrone, Title: id,
			Condition: "new", PriceMode: domain.PriceModeFixed, PriceFen: 100,
			CheckStatus: check, Status: status,
		}); err != nil {
			t.Fatalf("建商品 %s: %v", id, err)
		}
	}
	mk("p-listed", domain.ProductCheckPassed, "listed")    // 已过审在售
	mk("p-removed", domain.ProductCheckPassed, "removed")  // 已过审已下架
	mk("p-pending", domain.ProductCheckPending, "pending") // 未过审
	mk("p-sold", domain.ProductCheckPassed, "sold")        // 已售
	mk("p-rejected", domain.ProductCheckRejected, "pending")

	statusOf := func(id string) string {
		t.Helper()
		p, err := prodRepo.FindByID(ctx, id)
		if err != nil {
			t.Fatalf("读 %s: %v", id, err)
		}
		return p.Status
	}

	// 批量上架：只有已过审且非已售的行会被改动
	ids := []string{"p-removed", "p-pending", "p-sold", "p-rejected", "p-nope"}
	n, err := svc.BulkSetProductStatus(ctx, ids, "listed")
	if err != nil {
		t.Fatalf("批量上架: %v", err)
	}
	if n != 1 {
		t.Fatalf("只应改动 1 行（p-removed），实际 %d", n)
	}
	if got := statusOf("p-removed"); got != "listed" {
		t.Fatalf("p-removed 应被上架，实际 %q", got)
	}
	// 未过审的两件必须原封不动——批量入口不能成为绕过审核的通道
	if got := statusOf("p-pending"); got != "pending" {
		t.Fatalf("未过审商品不得被批量上架，实际 %q", got)
	}
	if got := statusOf("p-rejected"); got != "pending" {
		t.Fatalf("被驳回商品不得被批量上架，实际 %q", got)
	}
	if got := statusOf("p-sold"); got != "sold" {
		t.Fatalf("已售商品不得被批量改状态，实际 %q", got)
	}

	// 批量下架：已过审的两件（listed/sold 中的 listed）可下架，sold 仍被跳过
	n, err = svc.BulkSetProductStatus(ctx, []string{"p-listed", "p-sold"}, "removed")
	if err != nil {
		t.Fatalf("批量下架: %v", err)
	}
	if n != 1 {
		t.Fatalf("只应改动 1 行（p-listed），实际 %d", n)
	}
	if got := statusOf("p-sold"); got != "sold" {
		t.Fatalf("已售商品不得被批量下架，实际 %q", got)
	}

	// 参数校验
	if _, err := svc.BulkSetProductStatus(ctx, []string{"p-listed"}, "sold"); !errors.Is(err, service.ErrProductInvalid) {
		t.Fatalf("批量置为 sold 应被拒，实际 %v", err)
	}
	if _, err := svc.BulkSetProductStatus(ctx, nil, "listed"); !errors.Is(err, service.ErrProductInvalid) {
		t.Fatalf("空 ids 应被拒，实际 %v", err)
	}
}

// 批量上下架不得碰其他列——这正是"逐行 PUT 整行"的老毛病。
func TestBulkSetProductStatusTouchesOnlyStatus(t *testing.T) {
	ctx := context.Background()
	prodRepo := memory.NewProductRepository()
	svc := service.NewTradingService(prodRepo, memory.NewRepairRepository(), nil, nil)

	if _, err := prodRepo.Create(ctx, domain.DroneProduct{
		ID: "p-1", SellerID: "seller-1", SellerName: "重庆低空科技",
		ProdType: domain.ProductDrone, Title: "大疆M350", Description: "描述",
		Brand: "DJI", Model: "M350", Condition: "used",
		PriceMode: domain.PriceModeFixed, PriceFen: 8800000, Delivery: domain.DeliveryLogistics,
		CheckStatus: domain.ProductCheckPassed, Status: "listed", Views: 42,
	}); err != nil {
		t.Fatalf("建商品: %v", err)
	}
	before, _ := prodRepo.FindByID(ctx, "p-1")

	if _, err := svc.BulkSetProductStatus(ctx, []string{"p-1"}, "removed"); err != nil {
		t.Fatalf("批量下架: %v", err)
	}
	after, _ := prodRepo.FindByID(ctx, "p-1")

	if after.Status != "removed" {
		t.Fatalf("状态未改动: %q", after.Status)
	}
	// 除 status/version/updated_at 外，其余字段必须逐字节不变
	if after.Title != before.Title || after.Description != before.Description ||
		after.Brand != before.Brand || after.Model != before.Model ||
		after.Condition != before.Condition || after.PriceFen != before.PriceFen ||
		after.PriceMode != before.PriceMode || after.Delivery != before.Delivery ||
		after.SellerName != before.SellerName || after.CheckStatus != before.CheckStatus ||
		after.Views != before.Views {
		t.Fatalf("批量改状态动了其他列: before=%+v after=%+v", before, after)
	}
}
