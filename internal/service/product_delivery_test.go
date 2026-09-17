package service_test

import (
	"context"
	"testing"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

// OrderNeedsReceiver：下单是否需要收货地址，**优先看卖家选定的交付方式**。
//
// 这条规则的存在意义就是修掉"采集即丢"——此前交付方式在提交时被丢弃，
// 只能按商品类型猜，于是卖家选了「自提」，买家仍被要求填收货地址。
func TestOrderNeedsReceiverByDelivery(t *testing.T) {
	cases := []struct {
		name     string
		prodType domain.ProductType
		delivery string
		want     bool
	}{
		// 交付方式优先，压过商品类型
		{"自提（整机）→ 不需要", domain.ProductDrone, domain.DeliveryPickup, false},
		{"物流发货（整机）→ 需要", domain.ProductDrone, domain.DeliveryLogistics, true},
		{"同城配送（整机）→ 需要", domain.ProductDrone, domain.DeliveryCity, true},
		{"物流发货（服务类）→ 需要", domain.ProductAerial, domain.DeliveryLogistics, true},
		{"自提（服务类）→ 不需要", domain.ProductAerial, domain.DeliveryPickup, false},

		// 未选/可协商 → 按商品类型兜底（与补齐前的行为一致）
		{"可协商（整机）→ 兜底：需要", domain.ProductDrone, domain.DeliveryNegotiable, true},
		{"未指定（整机）→ 兜底：需要", domain.ProductDrone, "", true},
		{"未指定（配件）→ 兜底：需要", domain.ProductPart, "", true},
		{"可协商（试飞）→ 兜底：不需要", domain.ProductTestFly, domain.DeliveryNegotiable, false},
		{"未指定（维修）→ 兜底：不需要", domain.ProductRepair, "", false},
		{"未指定（航拍）→ 兜底：不需要", domain.ProductAerial, "", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := service.OrderNeedsReceiver(domain.DroneProduct{ProdType: c.prodType, Delivery: c.delivery})
			if got != c.want {
				t.Fatalf("prod_type=%q delivery=%q: 期望 %v 实际 %v", c.prodType, c.delivery, c.want, got)
			}
		})
	}
}

// 交付方式必须落库并能读回——否则上面那条规则就是空转。
func TestProductDeliveryPersists(t *testing.T) {
	ctx := context.Background()
	prodRepo := memory.NewProductRepository()
	svc := service.NewTradingService(prodRepo, memory.NewRepairRepository(), nil, nil)
	seller := domain.Actor{ID: "seller-1", Role: domain.RoleEnterprise}

	p, err := svc.CreateProduct(ctx, seller, domain.ProductDrone, "大疆M350", "", "DJI", "M350", "new",
		domain.DeliveryPickup, domain.PriceModeFixed, 8800000, nil, nil)
	if err != nil {
		t.Fatalf("发布: %v", err)
	}
	got, err := prodRepo.FindByID(ctx, p.ID)
	if err != nil {
		t.Fatalf("读回: %v", err)
	}
	if got.Delivery != domain.DeliveryPickup {
		t.Fatalf("交付方式未落库，实际 %q", got.Delivery)
	}

	// 非法交付方式必须被拒（白名单）
	if _, err := svc.CreateProduct(ctx, seller, domain.ProductDrone, "非法交付", "", "", "", "new",
		"teleport", domain.PriceModeFixed, 100, nil, nil); err == nil {
		t.Fatal("白名单外的交付方式应被拒")
	}
}
