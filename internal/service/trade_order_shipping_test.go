package service_test

import (
	"context"
	"testing"
	"time"
	"unicode/utf8"
	"strings"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
	"drone-platform/internal/service"
)

// ── 交付链路：收货地址 + 发货单号 ──
//
// 补齐前的缺口：发布表单里有「物流发货」这个交付方式，但订单表既没有收货地址也没有
// 发货单号——卖家卖出一台整机后不知道寄给谁，也没地方填快递单号，
// 而订单详情页的"确认发货"按钮是存在的。

func rcvr() service.OrderReceiver {
	return service.OrderReceiver{
		Name: "张三", Phone: "13800000000",
		Region: "重庆市渝北区", Address: "龙兴镇某某路 1 号",
	}
}

// seedProductType 落一个指定类型、状态在售的商品（下单要求 listed）。
func seedProductType(t *testing.T, ctx context.Context, prodRepo repository.ProductRepository, sellerID string, priceFen int64, prodType domain.ProductType) string {
	t.Helper()
	p, err := prodRepo.Create(ctx, domain.DroneProduct{
		ID: "prod-" + string(prodType) + "-" + time.Now().Format("150405.000000000"), SellerID: sellerID,
		ProdType: prodType, Title: "测试商品", PriceFen: priceFen, Status: "listed",
	})
	if err != nil {
		t.Fatalf("建商品: %v", err)
	}
	return p.ID
}

// 实物商品下单必须带收货信息；服务类商品（不需要寄送）允许留空。
// 这是 Handler 用 product.ProdType 判断 needsShipping 的依据，服务层据此强制。
func TestCreateOrderReceiverRequirement(t *testing.T) {
	cases := []struct {
		name        string
		prodType    domain.ProductType
		wantBlocked bool
	}{
		{"整机（实物，需寄送）", domain.ProductDrone, true},
		{"配件（实物，需寄送）", domain.ProductPart, true},
		{"维修服务", domain.ProductRepair, false},
		{"试飞预约", domain.ProductTestFly, false},
		{"航拍服务", domain.ProductAerial, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			tradeSvc, prodRepo, _, ctx := newMoneyFlow(t)
			pid := seedProductType(t, ctx, prodRepo, "seller-1", 10000, c.prodType)

			// 空收货信息 + requireReceiver=true（Handler 对实物商品的判定）
			_, err := tradeSvc.Create(ctx, "buyer-1", pid, "seller-1", 10000, service.OrderReceiver{}, true)
			if c.wantBlocked {
				if err == nil {
					t.Fatal("实物商品缺收货信息应被拒——没有地址卖家发不出去")
				}
				if !strings.Contains(err.Error(), "收货信息") {
					t.Fatalf("错误信息应说明是收货信息问题，实际 %v", err)
				}
				return
			}
			// 服务类商品在 Handler 侧传 requireReceiver=false，允许空地址
			if _, err := tradeSvc.Create(ctx, "buyer-1", pid, "seller-1", 10000, service.OrderReceiver{}, false); err != nil {
				t.Fatalf("服务类商品不应强制收货信息，实际 %v", err)
			}
		})
	}
}

// 收货信息三项必填（收货人 / 手机号 / 详细地址），省市区选填。
func TestCreateOrderReceiverCompleteness(t *testing.T) {
	cases := []struct {
		name string
		r    service.OrderReceiver
		ok   bool
	}{
		{"完整", rcvr(), true},
		{"缺省市区（选填）", service.OrderReceiver{Name: "张三", Phone: "138", Address: "某路 1 号"}, true},
		{"缺收货人", service.OrderReceiver{Phone: "138", Address: "某路 1 号"}, false},
		{"缺手机号", service.OrderReceiver{Name: "张三", Address: "某路 1 号"}, false},
		{"缺详细地址", service.OrderReceiver{Name: "张三", Phone: "138"}, false},
		{"全空白", service.OrderReceiver{Name: "  ", Phone: " ", Address: "\t"}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if utf8.RuneCountInString(c.name) == 0 {
				t.Fatal("unreachable")
			}
			tradeSvc, prodRepo, _, ctx := newMoneyFlow(t)
			pid := seedProduct(t, ctx, prodRepo, "seller-1", 10000)
			_, err := tradeSvc.Create(ctx, "buyer-1", pid, "seller-1", 10000, c.r, true)
			if c.ok && err != nil {
				t.Fatalf("期望通过，实际 %v", err)
			}
			if !c.ok && err == nil {
				t.Fatal("期望拒绝，实际通过")
			}
		})
	}
}

// 收货信息必须落库：买家填的地址要能被卖家看到，否则整条链路白做。
func TestCreateOrderPersistsReceiver(t *testing.T) {
	tradeSvc, prodRepo, _, ctx := newMoneyFlow(t)
	pid := seedProduct(t, ctx, prodRepo, "seller-1", 10000)
	r := rcvr()
	o, err := tradeSvc.Create(ctx, "buyer-1", pid, "seller-1", 10000, r, true)
	if err != nil {
		t.Fatalf("下单: %v", err)
	}
	got, err := tradeSvc.FindByID(ctx, o.ID)
	if err != nil {
		t.Fatalf("查单: %v", err)
	}
	if got.ReceiverName != r.Name || got.ReceiverPhone != r.Phone ||
		got.ReceiverRegion != r.Region || got.ReceiverAddress != r.Address {
		t.Fatalf("收货信息未落库: name=%q phone=%q region=%q addr=%q",
			got.ReceiverName, got.ReceiverPhone, got.ReceiverRegion, got.ReceiverAddress)
	}
}

// 发货约束：单号必填、仅卖家、仅 paid 状态、不可重复发货、发货时间落库。
func TestShipOrderRules(t *testing.T) {
	tradeSvc, prodRepo, escrowSvc, ctx := newMoneyFlow(t)
	if _, err := escrowSvc.Deposit(ctx, "buyer-1", 100000); err != nil {
		t.Fatalf("充值: %v", err)
	}
	pid := seedProduct(t, ctx, prodRepo, "seller-1", 10000)
	seller := domain.Actor{ID: "seller-1", Role: domain.RoleEnterprise}
	other := domain.Actor{ID: "seller-2", Role: domain.RoleEnterprise}
	buyer := domain.Actor{ID: "buyer-1", Role: domain.RoleIndividual}

	o, err := tradeSvc.Create(ctx, "buyer-1", pid, "seller-1", 10000, rcvr(), true)
	if err != nil {
		t.Fatalf("下单: %v", err)
	}
	if _, err := tradeSvc.ShipOrder(ctx, seller, o.ID, "顺丰速运", "SF1"); err == nil {
		t.Fatal("pending（未付款）订单不应能发货")
	}
	if _, err := tradeSvc.PayOrder(ctx, "buyer-1", o.ID); err != nil {
		t.Fatalf("支付: %v", err)
	}

	if _, err := tradeSvc.ShipOrder(ctx, seller, o.ID, "顺丰速运", "   "); err == nil {
		t.Fatal("空单号必须被拒——没有单号的发货买家查不到物流、出问题也无从举证")
	}
	if _, err := tradeSvc.ShipOrder(ctx, other, o.ID, "顺丰速运", "SF1"); err == nil {
		t.Fatal("非卖家不应能发货")
	}
	if _, err := tradeSvc.ShipOrder(ctx, buyer, o.ID, "顺丰速运", "SF1"); err == nil {
		t.Fatal("买家不应能发货")
	}

	shipped, err := tradeSvc.ShipOrder(ctx, seller, o.ID, " 顺丰速运 ", " SF123456 ")
	if err != nil {
		t.Fatalf("发货: %v", err)
	}
	if shipped.Status != "shipped" || shipped.ShippingTracking != "SF123456" || shipped.ShippingCompany != "顺丰速运" {
		t.Fatalf("发货结果不对: status=%q company=%q tracking=%q", shipped.Status, shipped.ShippingCompany, shipped.ShippingTracking)
	}
	if shipped.ShippedAt == nil {
		t.Fatal("shipped_at 未写入")
	}
	if _, err := tradeSvc.ShipOrder(ctx, seller, o.ID, "顺丰速运", "SF999"); err == nil {
		t.Fatal("重复发货应被拒")
	}
}

// 自提订单没有快递单号可言：允许不带单号发货（判据取商品上卖家选定的交付方式）。
func TestShipPickupOrderNeedsNoTracking(t *testing.T) {
	tradeSvc, prodRepo, escrowSvc, ctx := newMoneyFlow(t)
	if _, err := escrowSvc.Deposit(ctx, "buyer-1", 100000); err != nil {
		t.Fatalf("充值: %v", err)
	}
	// 商品交付方式 = 自提
	p, err := prodRepo.Create(ctx, domain.DroneProduct{
		ID: "prod-pickup-1", SellerID: "seller-1", ProdType: domain.ProductDrone,
		Title: "自提整机", PriceFen: 10000, Status: "listed", Delivery: domain.DeliveryPickup,
	})
	if err != nil {
		t.Fatalf("建商品: %v", err)
	}
	seller := domain.Actor{ID: "seller-1", Role: domain.RoleEnterprise}

	o, err := tradeSvc.Create(ctx, "buyer-1", p.ID, "seller-1", 10000, service.OrderReceiver{}, false)
	if err != nil {
		t.Fatalf("下单: %v", err)
	}
	if _, err := tradeSvc.PayOrder(ctx, "buyer-1", o.ID); err != nil {
		t.Fatalf("支付: %v", err)
	}
	shipped, err := tradeSvc.ShipOrder(ctx, seller, o.ID, "", "")
	if err != nil {
		t.Fatalf("自提订单应允许不带快递单号发货，实际 %v", err)
	}
	if shipped.Status != "shipped" || shipped.ShippingTracking != "" {
		t.Fatalf("自提发货结果不对: status=%q tracking=%q", shipped.Status, shipped.ShippingTracking)
	}
	// 重复发货仍要被拒（单号为空也不能靠它绕过"只能发一次"）
	if _, err := tradeSvc.ShipOrder(ctx, seller, o.ID, "", ""); err == nil {
		t.Fatal("自提订单也不应能重复发货")
	}
}

// 回归：卖家不得经 PATCH status=shipped 绕过发货端点。
// 否则会造出"已发货但没有单号"的订单，本次补齐的单号字段形同虚设。
func TestPatchStatusCannotShip(t *testing.T) {
	tradeSvc, prodRepo, escrowSvc, ctx := newMoneyFlow(t)
	if _, err := escrowSvc.Deposit(ctx, "buyer-1", 100000); err != nil {
		t.Fatalf("充值: %v", err)
	}
	pid := seedProduct(t, ctx, prodRepo, "seller-1", 10000)
	o, err := tradeSvc.Create(ctx, "buyer-1", pid, "seller-1", 10000, rcvr(), true)
	if err != nil {
		t.Fatalf("下单: %v", err)
	}
	if _, err := tradeSvc.PayOrder(ctx, "buyer-1", o.ID); err != nil {
		t.Fatalf("支付: %v", err)
	}
	if _, err := tradeSvc.UpdateStatus(ctx, o.ID, "seller-1", "shipped"); err == nil {
		t.Fatal("经 PATCH 状态直接发货必须被拒（必须走发货端点带单号）")
	}
	got, _ := tradeSvc.FindByID(ctx, o.ID)
	if got.Status != "paid" {
		t.Fatalf("被拒的发货不得改动订单状态，实际 %q", got.Status)
	}
	if got.ShippingTracking != "" {
		t.Fatalf("被拒的发货不得写入单号，实际 %q", got.ShippingTracking)
	}
}
