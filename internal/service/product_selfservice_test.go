package service_test

import (
	"context"
	"errors"
	"testing"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

// 卖家自助编辑的归属校验：改别人的商品必须按"不存在"处理。
func TestUpdateMyProductRequiresOwnership(t *testing.T) {
	ctx := context.Background()
	svc := service.NewTradingService(memory.NewProductRepository(), memory.NewRepairRepository(), nil, nil)

	p, err := svc.CreateProduct(ctx, domain.Actor{ID: "seller-1", Role: domain.RoleEnterprise},
		domain.ProductDrone, "大疆M350", "", "DJI", "M350", "new", "", "", 8800000, nil, nil)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	// 另一个卖家改它 → ErrNotFound（不泄露"存在但不是你的"）
	_, err = svc.UpdateMyProduct(ctx, domain.Actor{ID: "seller-2", Role: domain.RoleEnterprise}, p.ID,
		service.ProductEditInput{ProdType: domain.ProductDrone, Title: "被篡改", Condition: "new", PriceFen: 1})
	if !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("改别人的商品应返回 ErrNotFound，实际 %v", err)
	}
	got, err := svc.GetProduct(ctx, p.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Title != "大疆M350" {
		t.Fatalf("越权编辑不得改动商品，实际标题 %q", got.Title)
	}
}

// 内容编辑必须退回待审核：审核的对象是内容，内容变了原结论就失效。
// 否则卖家可以先发一件合规商品过审、再改成违规内容，审核形同虚设。
func TestUpdateMyProductReturnsToPendingReview(t *testing.T) {
	ctx := context.Background()
	svc := service.NewTradingService(memory.NewProductRepository(), memory.NewRepairRepository(), nil, nil)
	seller := domain.Actor{ID: "seller-1", Role: domain.RoleEnterprise}
	admin := domain.Actor{ID: "admin-1", Role: domain.RolePlatformAdmin}

	p, err := svc.CreateProduct(ctx, seller, domain.ProductDrone, "大疆M350", "", "DJI", "M350", "new", "", "", 8800000, nil, nil)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := svc.ReviewProduct(ctx, admin, p.ID, domain.ProductCheckPassed, ""); err != nil {
		t.Fatalf("review: %v", err)
	}
	listed, _ := svc.GetProduct(ctx, p.ID)
	if !service.ProductVisibleInHall(listed) {
		t.Fatal("审核通过后应公开可见")
	}

	// 卖家改价 → 退回待审，公开列表不再可见
	edited, err := svc.UpdateMyProduct(ctx, seller, p.ID, service.ProductEditInput{
		ProdType: domain.ProductDrone, Title: "大疆M350（降价）", Brand: "DJI", Model: "M350",
		Condition: "new", PriceMode: domain.PriceModeFixed, PriceFen: 8000000,
	})
	if err != nil {
		t.Fatalf("edit: %v", err)
	}
	if edited.CheckStatus != domain.ProductCheckPending {
		t.Fatalf("编辑后应退回待审核，实际 check_status=%q", edited.CheckStatus)
	}
	if edited.Status == "listed" {
		t.Fatalf("编辑后不得留在上架状态，实际 status=%q", edited.Status)
	}
	if service.ProductVisibleInHall(edited) {
		t.Fatal("退回待审的商品不得公开可见")
	}
	if edited.ReviewedAt != nil || edited.ReviewedBy != "" {
		t.Fatal("退回待审应清空上一次的审核留痕")
	}
	// 重新审核通过后再次可见
	if _, err := svc.ReviewProduct(ctx, admin, p.ID, domain.ProductCheckPassed, ""); err != nil {
		t.Fatalf("re-review: %v", err)
	}
	again, _ := svc.GetProduct(ctx, p.ID)
	if !service.ProductVisibleInHall(again) {
		t.Fatal("重新审核通过后应恢复可见")
	}
	if again.Title != "大疆M350（降价）" {
		t.Fatalf("编辑内容未落库：%q", again.Title)
	}
}

// 已售商品不可编辑（订单已占位）。
func TestUpdateMyProductRejectsSoldProduct(t *testing.T) {
	ctx := context.Background()
	svc := service.NewTradingService(memory.NewProductRepository(), memory.NewRepairRepository(), nil, nil)
	seller := domain.Actor{ID: "seller-1", Role: domain.RoleEnterprise}
	admin := domain.Actor{ID: "admin-1", Role: domain.RolePlatformAdmin}

	p, err := svc.CreateProduct(ctx, seller, domain.ProductDrone, "大疆M350", "", "", "", "new", "", "", 100, nil, nil)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := svc.ReviewProduct(ctx, admin, p.ID, domain.ProductCheckPassed, ""); err != nil {
		t.Fatalf("review: %v", err)
	}
	if err := svc.MarkProductSold(ctx, p.ID); err != nil {
		t.Fatalf("mark sold: %v", err)
	}
	_, err = svc.UpdateMyProduct(ctx, seller, p.ID, service.ProductEditInput{
		ProdType: domain.ProductDrone, Title: "改已售商品", Condition: "new", PriceFen: 1,
	})
	if !errors.Is(err, service.ErrProductInvalid) {
		t.Fatalf("已售商品不可编辑，应返回 ErrProductInvalid，实际 %v", err)
	}
}

// 卖家自助上下架的约束：只能 listed/removed；未过审不可上架；已售不可改；改别人的按不存在。
func TestSetMyProductStatusRules(t *testing.T) {
	ctx := context.Background()
	svc := service.NewTradingService(memory.NewProductRepository(), memory.NewRepairRepository(), nil, nil)
	seller := domain.Actor{ID: "seller-1", Role: domain.RoleEnterprise}
	other := domain.Actor{ID: "seller-2", Role: domain.RoleEnterprise}
	admin := domain.Actor{ID: "admin-1", Role: domain.RolePlatformAdmin}

	p, err := svc.CreateProduct(ctx, seller, domain.ProductDrone, "大疆M350", "", "", "", "new", "", "", 100, nil, nil)
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	// 未过审就想上架 → 拒绝
	if _, err := svc.SetMyProductStatus(ctx, seller, p.ID, "listed"); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("未过审商品不得自助上架，实际 %v", err)
	}
	// 非法状态值 → 400 语义
	if _, err := svc.SetMyProductStatus(ctx, seller, p.ID, "sold"); !errors.Is(err, service.ErrProductInvalid) {
		t.Fatalf("卖家不得把商品直接置为 sold，实际 %v", err)
	}
	if _, err := svc.SetMyProductStatus(ctx, seller, p.ID, "pending"); !errors.Is(err, service.ErrProductInvalid) {
		t.Fatalf("非法状态值应被拒，实际 %v", err)
	}
	// 改别人的 → 不存在
	if _, err := svc.SetMyProductStatus(ctx, other, p.ID, "removed"); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("改别人的商品应返回 ErrNotFound，实际 %v", err)
	}

	// 过审 → 卖家可下架、可重新上架
	if _, err := svc.ReviewProduct(ctx, admin, p.ID, domain.ProductCheckPassed, ""); err != nil {
		t.Fatalf("review: %v", err)
	}
	off, err := svc.SetMyProductStatus(ctx, seller, p.ID, "removed")
	if err != nil {
		t.Fatalf("下架: %v", err)
	}
	if off.Status != "removed" || off.CheckStatus != domain.ProductCheckPassed {
		t.Fatalf("下架只应改上架维度，实际 status=%q check_status=%q", off.Status, off.CheckStatus)
	}
	on, err := svc.SetMyProductStatus(ctx, seller, p.ID, "listed")
	if err != nil {
		t.Fatalf("重新上架: %v", err)
	}
	if !service.ProductVisibleInHall(on) {
		t.Fatal("重新上架后应公开可见")
	}

	// 已售不可自助改状态
	if err := svc.MarkProductSold(ctx, p.ID); err != nil {
		t.Fatalf("mark sold: %v", err)
	}
	if _, err := svc.SetMyProductStatus(ctx, seller, p.ID, "removed"); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("已售商品不得自助改状态，实际 %v", err)
	}
}

// 价格模式与金额必须自洽——"面议"和"标价 0 元"不能再是同一个值。
func TestPriceModeValidation(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		name     string
		mode     string
		priceFen int64
		wantErr  bool
		wantMode string
	}{
		{"明码标价有价", domain.PriceModeFixed, 8800000, false, domain.PriceModeFixed},
		{"明码标价 0 元", domain.PriceModeFixed, 0, true, ""},
		{"明码标价负数", domain.PriceModeFixed, -1, true, ""},
		{"面议价格为 0", domain.PriceModeNegotiable, 0, false, domain.PriceModeNegotiable},
		{"面议却给了价格", domain.PriceModeNegotiable, 100, true, ""},
		{"非法价格模式", "free", 100, true, ""},
		// 兼容路径：未传 price_mode 时按金额推断（price_fen=0 的旧语义就是面议）
		{"未指定+0 价", "", 0, false, domain.PriceModeNegotiable},
		{"未指定+有价", "", 100, false, domain.PriceModeFixed},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			svc := service.NewTradingService(memory.NewProductRepository(), memory.NewRepairRepository(), nil, nil)
			p, err := svc.CreateProductByAdmin(ctx, domain.DroneProduct{
				SellerID: "platform", ProdType: domain.ProductDrone, Title: "大疆M350",
				Condition: "new", PriceMode: c.mode, PriceFen: c.priceFen, Status: "listed",
			})
			if c.wantErr {
				if err == nil {
					t.Fatalf("期望拒绝，实际入库：price_mode=%q price_fen=%d", p.PriceMode, p.PriceFen)
				}
				if !errors.Is(err, service.ErrProductInvalid) {
					t.Fatalf("错误必须包装 ErrProductInvalid，实际 %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("期望通过，实际 %v", err)
			}
			if p.PriceMode != c.wantMode {
				t.Fatalf("price_mode 应为 %q，实际 %q", c.wantMode, p.PriceMode)
			}
			// 面议商品的价格必须归零，避免前端再把它显示成 ¥0
			if p.PriceMode == domain.PriceModeNegotiable && p.PriceFen != 0 {
				t.Fatalf("面议商品价格必须为 0，实际 %d", p.PriceFen)
			}
		})
	}
}
