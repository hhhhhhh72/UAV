package service_test

import (
	"context"
	"testing"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
	"drone-platform/internal/service"
)

// productStatus 读商品当前状态（售后回架断言用）。
func productStatus(t *testing.T, ctx context.Context, repo repository.ProductRepository, id string) string {
	t.Helper()
	p, err := repo.FindByID(ctx, id)
	if err != nil {
		t.Fatalf("查商品 %s: %v", id, err)
	}
	return p.Status
}

// markSold 复刻下单 handler 的抢占动作（httpapi/phase3.go 里 MarkProductSold 在
// 调 tradeSvc.Create **之前**执行）：下单即把商品置 sold 占位，防一物多卖。
// 放在这里是因为占位发生在 handler 层，service.Create 本身不碰商品。
func markSold(t *testing.T, ctx context.Context, repo repository.ProductRepository, id string) {
	t.Helper()
	p, err := repo.FindByID(ctx, id)
	if err != nil {
		t.Fatalf("查商品 %s: %v", id, err)
	}
	p.Status = "sold"
	if _, err := repo.Update(ctx, p); err != nil {
		t.Fatalf("商品占位为 sold: %v", err)
	}
}

// 仅退款 + 从未发货：货一直在卖家手里，售后结案后商品必须回货架。
//
// 回归背景（生产实测暴露）：下单即把商品置 sold 占位，而 reviewAftersale 同意分支
// 此前只退款、不碰商品，商品就永久停在 sold——孤儿回收任务也救不了它，因为
// HasLiveOrderForProduct 把 status<>'cancelled' 一律当成"仍在交易中"，
// 而结案订单是 completed。表现为：钱退了、货没发、商品再也买不到，且毫无提示。
func TestAftersaleRefundBeforeShipRelistsProduct(t *testing.T) {
	tradeSvc, prodRepo, escrowSvc, ctx := newMoneyFlow(t)
	productID := seedProduct(t, ctx, prodRepo, "seller-1", 10000)
	if _, err := escrowSvc.Deposit(ctx, "buyer-1", 10000); err != nil {
		t.Fatalf("充值: %v", err)
	}

	markSold(t, ctx, prodRepo, productID)
	o, err := tradeSvc.Create(ctx, "buyer-1", productID, "seller-1", 10000, service.OrderReceiver{}, false)
	if err != nil {
		t.Fatalf("下单: %v", err)
	}
	if got := productStatus(t, ctx, prodRepo, productID); got != "sold" {
		t.Fatalf("下单后商品应被占位为 sold，实际 %q", got)
	}
	if _, err := tradeSvc.PayOrder(ctx, "buyer-1", o.ID); err != nil {
		t.Fatalf("付款: %v", err)
	}
	// 从未发货，直接申请仅退款
	if _, err := tradeSvc.ApplyAftersale(ctx, "buyer-1", o.ID, "refund", "不想要了", "", 10000); err != nil {
		t.Fatalf("申请售后: %v", err)
	}
	if _, err := tradeSvc.ReviewAftersale(ctx, o.ID, true); err != nil {
		t.Fatalf("同意售后: %v", err)
	}
	if b, f := balanceOf(t, ctx, escrowSvc, "buyer-1"); b != 10000 || f != 0 {
		t.Fatalf("退款后买家余额=%d 冻结=%d，want 10000/0", b, f)
	}
	// 关键断言：商品必须回货架，否则永久锁死
	if got := productStatus(t, ctx, prodRepo, productID); got != "listed" {
		t.Fatalf("仅退款(从未发货)结案后商品应恢复 listed，实际 %q——商品被永久锁死", got)
	}
}

// 仅退款 + 已发货：货在买家手上，商品不该自动回到货架（否则等于一件货卖两次）。
func TestAftersaleRefundAfterShipKeepsProductSold(t *testing.T) {
	tradeSvc, prodRepo, escrowSvc, ctx := newMoneyFlow(t)
	productID := seedProduct(t, ctx, prodRepo, "seller-1", 10000)
	if _, err := escrowSvc.Deposit(ctx, "buyer-1", 10000); err != nil {
		t.Fatalf("充值: %v", err)
	}
	markSold(t, ctx, prodRepo, productID)
	o, err := tradeSvc.Create(ctx, "buyer-1", productID, "seller-1", 10000, service.OrderReceiver{}, false)
	if err != nil {
		t.Fatalf("下单: %v", err)
	}
	if _, err := tradeSvc.PayOrder(ctx, "buyer-1", o.ID); err != nil {
		t.Fatalf("付款: %v", err)
	}
	seller := domain.Actor{ID: "seller-1", Role: domain.RoleEnterprise}
	if _, err := tradeSvc.ShipOrder(ctx, seller, o.ID, "顺丰速运", "SF-TEST-1"); err != nil {
		t.Fatalf("发货: %v", err)
	}
	if _, err := tradeSvc.ApplyAftersale(ctx, "buyer-1", o.ID, "refund", "质量问题", "", 10000); err != nil {
		t.Fatalf("申请仅退款: %v", err)
	}
	if _, err := tradeSvc.ReviewAftersale(ctx, o.ID, true); err != nil {
		t.Fatalf("同意售后: %v", err)
	}
	if b, _ := balanceOf(t, ctx, escrowSvc, "buyer-1"); b != 10000 {
		t.Fatalf("退款应回到买家，实际 %d", b)
	}
	if got := productStatus(t, ctx, prodRepo, productID); got != "sold" {
		t.Fatalf("已发货的仅退款买家留下货物，商品应保持 sold，实际 %q", got)
	}
}

// 退货退款走完全程：货已退回卖家，商品必须回货架；但买家还没寄回时不能提前回架。
func TestAftersaleReturnRelistsProduct(t *testing.T) {
	tradeSvc, prodRepo, escrowSvc, ctx := newMoneyFlow(t)
	productID := seedProduct(t, ctx, prodRepo, "seller-1", 10000)
	if _, err := escrowSvc.Deposit(ctx, "buyer-1", 10000); err != nil {
		t.Fatalf("充值: %v", err)
	}
	markSold(t, ctx, prodRepo, productID)
	o, err := tradeSvc.Create(ctx, "buyer-1", productID, "seller-1", 10000, service.OrderReceiver{}, false)
	if err != nil {
		t.Fatalf("下单: %v", err)
	}
	if _, err := tradeSvc.PayOrder(ctx, "buyer-1", o.ID); err != nil {
		t.Fatalf("付款: %v", err)
	}
	seller := domain.Actor{ID: "seller-1", Role: domain.RoleEnterprise}
	if _, err := tradeSvc.ShipOrder(ctx, seller, o.ID, "顺丰速运", "SF-TEST-1"); err != nil {
		t.Fatalf("发货: %v", err)
	}
	if _, err := tradeSvc.ApplyAftersale(ctx, "buyer-1", o.ID, "return", "货不对板", "", 10000); err != nil {
		t.Fatalf("申请退货退款: %v", err)
	}

	// ① 卖家同意退货：只到"待买家寄回"，此时货还在买家手上，商品不得回架
	if _, err := tradeSvc.ReviewAftersale(ctx, o.ID, true); err != nil {
		t.Fatalf("同意退货: %v", err)
	}
	if got := productStatus(t, ctx, prodRepo, productID); got != "sold" {
		t.Fatalf("买家尚未寄回，商品不该回货架，实际 %q", got)
	}

	// ② 买家寄回 + 卖家确认收到 → 退款结案，货已回到卖家
	if _, err := tradeSvc.SubmitReturnShipment(ctx, "buyer-1", o.ID, "SF-BACK-1", ""); err != nil {
		t.Fatalf("提交退货物流: %v", err)
	}
	if _, err := tradeSvc.ConfirmReturnReceived(ctx, seller, o.ID); err != nil {
		t.Fatalf("确认收到退货: %v", err)
	}
	if b, f := balanceOf(t, ctx, escrowSvc, "buyer-1"); b != 10000 || f != 0 {
		t.Fatalf("退货结案后买家余额=%d 冻结=%d，want 10000/0", b, f)
	}
	if got := productStatus(t, ctx, prodRepo, productID); got != "listed" {
		t.Fatalf("退货结案后商品应恢复 listed，实际 %q", got)
	}
}
