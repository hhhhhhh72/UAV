package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

// 用户发布商品的必填校验回归。
//
// 修复前 service.CreateProduct 只检查价格非负：空标题、空/非法 prod_type、任意 condition、
// 任意数量图片全部能入库。待审队列里会混进无标题商品，协会审核时也无从写驳回理由。
func TestCreateProductValidation(t *testing.T) {
	ctx := context.Background()
	actor := domain.Actor{ID: "seller-1", Role: domain.RoleEnterprise}

	cases := []struct {
		name      string
		prodType  domain.ProductType
		title     string
		condition string
		price     int64
		images    []string
		wantErr   bool
	}{
		{"空白标题", domain.ProductDrone, "   ", "new", 100, nil, true},
		{"空类型", "", "大疆M350", "new", 100, nil, true},
		{"白名单外类型", domain.ProductType("banana"), "大疆M350", "new", 100, nil, true},
		{"非法成色", domain.ProductDrone, "大疆M350", "brand_new", 100, nil, true},
		{"负数价格", domain.ProductDrone, "大疆M350", "new", -1, nil, true},
		{"图片超限", domain.ProductDrone, "大疆M350", "new", 100, make([]string, 10), true},
		{"标题过长", domain.ProductDrone, strings.Repeat("长", 101), "new", 100, nil, true},
		{"合法整机", domain.ProductDrone, "大疆M350", "new", 8800000, []string{"a.jpg"}, false},
		{"合法服务类", domain.ProductTestFly, "试飞测试", "new", 50000, nil, false},
		{"成色留空归一", domain.ProductPart, "M350 电池", "", 100, nil, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := memory.NewProductRepository()
			svc := service.NewTradingService(repo, memory.NewRepairRepository(), nil, nil)
			p, err := svc.CreateProduct(ctx, actor, tc.prodType, tc.title, "描述", "DJI", "M350", tc.condition, "", "", tc.price, tc.images, nil)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("期望拒绝，实际入库：%+v", p)
				}
				// Handler 靠这个哨兵把"填错了"映射成 400（而不是 403/500）。
				if !errors.Is(err, service.ErrProductInvalid) {
					t.Fatalf("错误必须包装 ErrProductInvalid，实际 %v", err)
				}
				// 校验失败不得留下任何行。
				if list, _ := repo.List(ctx, ""); len(list) != 0 {
					t.Fatalf("校验失败却写入了 %d 行", len(list))
				}
				return
			}
			if err != nil {
				t.Fatalf("期望通过，实际 %v", err)
			}
			if p.Status != "pending" {
				t.Fatalf("用户发布必须进待审核，实际 status=%q", p.Status)
			}
			if tc.condition == "" && p.Condition != "new" {
				t.Fatalf("空成色应归一为 new，实际 %q", p.Condition)
			}
		})
	}
}

// 卖家展示名必须是可读昵称，不能是用户 ID。
//
// 修复前 service/trading.go 写的是 SellerName: a.ID，公开接口为了不泄露手机号
// （user-<手机号> 形态的 ID）又不得不对它做哈希脱敏，最终买家在小程序商品详情
// 看到的"卖家"、店铺卡片、"该商品由 X 发布"全是一串哈希。
func TestCreateProductStoresSellerNickname(t *testing.T) {
	ctx := context.Background()
	userRepo := memory.NewUserRepository(nil)
	if _, err := userRepo.Create(ctx, domain.User{
		ID: "user-13800000001", Name: "重庆低空科技有限公司", Role: domain.RoleEnterprise,
	}); err != nil {
		t.Fatalf("seed user: %v", err)
	}

	svc := service.NewTradingService(memory.NewProductRepository(), memory.NewRepairRepository(), nil, userRepo)
	p, err := svc.CreateProduct(ctx, domain.Actor{ID: "user-13800000001", Role: domain.RoleEnterprise},
		domain.ProductDrone, "大疆M350", "", "DJI", "M350", "new", "", "", 100, nil, nil)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if p.SellerName != "重庆低空科技有限公司" {
		t.Fatalf("卖家展示名应为昵称，实际 %q", p.SellerName)
	}
	if p.SellerID != "user-13800000001" {
		t.Fatalf("seller_id 必须保留真实 ID（归属校验依赖它），实际 %q", p.SellerID)
	}
}

// 昵称为空 / 用户查不到时的兜底：仍须是"名字"，不能退化成 ID。
func TestCreateProductSellerNameFallback(t *testing.T) {
	ctx := context.Background()

	// 昵称为空 → "平台用户"
	userRepo := memory.NewUserRepository(nil)
	if _, err := userRepo.Create(ctx, domain.User{ID: "u-blank", Role: domain.RoleIndividual}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	svc := service.NewTradingService(memory.NewProductRepository(), memory.NewRepairRepository(), nil, userRepo)
	p, err := svc.CreateProduct(ctx, domain.Actor{ID: "u-blank", Role: domain.RoleIndividual},
		domain.ProductDrone, "M350", "", "", "", "new", "", "", 100, nil, nil)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if p.SellerName != "平台用户" {
		t.Fatalf("空昵称应回退为「平台用户」，实际 %q", p.SellerName)
	}

	// 用户不存在 → 也不能把 ID 写进去
	svc2 := service.NewTradingService(memory.NewProductRepository(), memory.NewRepairRepository(), nil, userRepo)
	p2, err := svc2.CreateProduct(ctx, domain.Actor{ID: "user-13900000009", Role: domain.RoleIndividual},
		domain.ProductDrone, "M350", "", "", "", "new", "", "", 100, nil, nil)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if p2.SellerName == "user-13900000009" || p2.SellerName == "" {
		t.Fatalf("卖家名不能是 ID 也不能为空，实际 %q", p2.SellerName)
	}
}

// 删除保护：有进行中订单的商品不得被删；订单取消后才可以。
func TestDeleteProductBlockedByLiveOrder(t *testing.T) {
	ctx := context.Background()
	prodRepo := memory.NewProductRepository()
	orderRepo := memory.NewTradeOrderRepository()
	svc := service.NewTradingService(prodRepo, memory.NewRepairRepository(), orderRepo, nil)

	p, err := svc.CreateProductByAdmin(ctx, domain.DroneProduct{
		SellerID: "platform", ProdType: domain.ProductDrone, Title: "大疆M350",
		Condition: "new", PriceFen: 8800000, Status: "listed",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := orderRepo.Create(ctx, domain.TradeOrder{
		ID: "o-live", ProductID: p.ID, BuyerID: "buyer-1", SellerID: "platform",
		AmountFen: 8800000, Status: "paid",
	}); err != nil {
		t.Fatalf("create order: %v", err)
	}

	// 有未取消订单 → 409 语义的错误，且商品原封不动
	if err := svc.DeleteProduct(ctx, p.ID); !errors.Is(err, service.ErrProductInTrade) {
		t.Fatalf("有进行中订单应返回 ErrProductInTrade，实际 %v", err)
	}
	if _, err := prodRepo.FindByID(ctx, p.ID); err != nil {
		t.Fatalf("被拒绝的删除不应改动商品: %v", err)
	}

	// 订单取消后 → 允许删除
	if _, err := orderRepo.UpdateStatus(ctx, "o-live", "cancelled"); err != nil {
		t.Fatalf("cancel order: %v", err)
	}
	if err := svc.DeleteProduct(ctx, p.ID); err != nil {
		t.Fatalf("订单全部取消后应可删除，实际 %v", err)
	}
}

// 软删除语义：行保留（订单还要靠它显示商品名），但对读接口不可见，可还原。
func TestDeleteProductIsSoftDelete(t *testing.T) {
	ctx := context.Background()
	prodRepo := memory.NewProductRepository()
	svc := service.NewTradingService(prodRepo, memory.NewRepairRepository(), nil, nil)

	p, err := svc.CreateProductByAdmin(ctx, domain.DroneProduct{
		SellerID: "platform", ProdType: domain.ProductDrone, Title: "大疆M350",
		Condition: "new", PriceFen: 100, Status: "listed",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if err := svc.DeleteProduct(ctx, p.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	// 对公开读不可见
	if _, err := prodRepo.FindByID(ctx, p.ID); err == nil {
		t.Fatal("软删除后 FindByID 应查不到")
	}
	if list, _ := prodRepo.List(ctx, ""); len(list) != 0 {
		t.Fatalf("软删除后列表应为空，实际 %d 行", len(list))
	}
	if n, _ := prodRepo.SumViews(ctx, ""); n != 0 {
		t.Fatalf("软删除后浏览量统计应为 0")
	}
	// 订单侧仍能按 ID 取到（这是软删除存在的意义）
	if byIDs, _ := prodRepo.ListByIDs(ctx, []string{p.ID}); len(byIDs) != 1 {
		t.Fatalf("ListByIDs 必须仍能取到已删除商品（订单要显示商品名），实际 %d 行", len(byIDs))
	}
	// 回收站可见 + 可还原
	deleted, err := svc.ListDeletedProducts(ctx)
	if err != nil {
		t.Fatalf("list deleted: %v", err)
	}
	if len(deleted) != 1 || deleted[0].ID != p.ID {
		t.Fatalf("回收站应有 1 件商品，实际 %+v", deleted)
	}
	if err := svc.UndeleteProduct(ctx, p.ID); err != nil {
		t.Fatalf("undelete: %v", err)
	}
	if _, err := prodRepo.FindByID(ctx, p.ID); err != nil {
		t.Fatalf("还原后应可查到: %v", err)
	}
	if list, _ := svc.ListDeletedProducts(ctx); len(list) != 0 {
		t.Fatalf("还原后回收站应为空，实际 %d 行", len(list))
	}
	// 重复删除 / 还原不存在的行必须报错，而不是假装成功
	if err := svc.UndeleteProduct(ctx, p.ID); err == nil {
		t.Fatal("还原一件不在回收站的商品应报错")
	}
}

// 回收站里的商品不得被下单占用，也不得被孤儿回收任务重新上架。
func TestDeletedProductNotSellable(t *testing.T) {
	ctx := context.Background()
	prodRepo := memory.NewProductRepository()
	svc := service.NewTradingService(prodRepo, memory.NewRepairRepository(), nil, nil)

	p, err := svc.CreateProductByAdmin(ctx, domain.DroneProduct{
		SellerID: "platform", ProdType: domain.ProductDrone, Title: "大疆M350",
		Condition: "new", PriceFen: 100, Status: "listed",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := svc.DeleteProduct(ctx, p.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if err := svc.MarkProductSold(ctx, p.ID); err == nil {
		t.Fatal("回收站商品不应能被下单占用（MarkSold 必须拒绝）")
	}
}
