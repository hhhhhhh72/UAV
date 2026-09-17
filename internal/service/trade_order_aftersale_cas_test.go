package service_test

import (
	"context"
	"testing"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

// 售后 B 分支的幂等：货款已放给卖家后，售后退款走 EscrowService.Transfer
// （卖家余额 → 买家余额）。仓储层 Transfer 是"余额条件 UPDATE + 写流水"，
// 没有任何 (refType, refID) 去重——重放会把卖家余额扣两次。
// 这组用例把"同一笔业务转账只能扣一次"钉死。
func TestEscrowTransferIsIdempotent(t *testing.T) {
	ctx := context.Background()
	escrowSvc := service.NewEscrowService(memory.NewEscrowRepository())
	if _, err := escrowSvc.Deposit(ctx, "seller-1", 1000); err != nil {
		t.Fatalf("给卖家充值: %v", err)
	}

	// 第一次：真实扣款
	if _, err := escrowSvc.Transfer(ctx, "seller-1", "buyer-1", 400, "trade_order", "torder-1"); err != nil {
		t.Fatalf("第一次转账: %v", err)
	}
	if b, _ := balanceOf(t, ctx, escrowSvc, "seller-1"); b != 600 {
		t.Fatalf("卖家应为 600，实际 %d", b)
	}
	if b, _ := balanceOf(t, ctx, escrowSvc, "buyer-1"); b != 400 {
		t.Fatalf("买家应为 400，实际 %d", b)
	}

	// 第二次：同一 (from, refType, refID) 重放——必须幂等（不再扣款）
	if _, err := escrowSvc.Transfer(ctx, "seller-1", "buyer-1", 400, "trade_order", "torder-1"); err != nil {
		t.Fatalf("重放应幂等成功，实际 %v", err)
	}
	if b, _ := balanceOf(t, ctx, escrowSvc, "seller-1"); b != 600 {
		t.Fatalf("重放不得二次扣款：卖家应仍为 600，实际 %d", b)
	}
	if b, _ := balanceOf(t, ctx, escrowSvc, "buyer-1"); b != 400 {
		t.Fatalf("重放不得重复入账：买家应仍为 400，实际 %d", b)
	}

	// 另一笔业务（不同 refID）仍应正常扣款——幂等键不能过度收窄
	if _, err := escrowSvc.Transfer(ctx, "seller-1", "buyer-1", 100, "trade_order", "torder-2"); err != nil {
		t.Fatalf("不同订单的转账不应被幂等拦截: %v", err)
	}
	if b, _ := balanceOf(t, ctx, escrowSvc, "seller-1"); b != 500 {
		t.Fatalf("卖家应为 500，实际 %d", b)
	}
}

// 售后重复审批：第一次同意后订单已离开 aftersale/pending，第二次必须被拒，
// 且卖家余额只能被扣一次（B 分支：钱已放给卖家后的退款）。
func TestTradeOrderAftersaleRepeatApproveRejected(t *testing.T) {
	tradeSvc, prodRepo, escrowSvc, ctx := newMoneyFlow(t)
	productID := seedProduct(t, ctx, prodRepo, "seller-1", 10000)
	if _, err := escrowSvc.Deposit(ctx, "buyer-1", 10000); err != nil {
		t.Fatalf("充值: %v", err)
	}
	o, err := tradeSvc.Create(ctx, "buyer-1", productID, "seller-1", 10000, service.OrderReceiver{}, false)
	if err != nil {
		t.Fatalf("下单: %v", err)
	}
	if _, err := tradeSvc.PayOrder(ctx, "buyer-1", o.ID); err != nil {
		t.Fatalf("付款: %v", err)
	}
	if _, err := tradeSvc.ShipOrder(ctx, domain.Actor{ID: "seller-1", Role: domain.RoleEnterprise}, o.ID, "顺丰速运", "SF-TEST-1"); err != nil {
		t.Fatalf("发货: %v", err)
	}
	if _, err := tradeSvc.UpdateStatus(ctx, o.ID, "buyer-1", "completed"); err != nil {
		t.Fatalf("确认收货: %v", err)
	}
	// 此刻货款已放给卖家（B 分支前置条件）
	if b, _ := balanceOf(t, ctx, escrowSvc, "seller-1"); b != 10000 {
		t.Fatalf("确认收货后卖家应为 10000，实际 %d", b)
	}

	if _, err := tradeSvc.ApplyAftersale(ctx, "buyer-1", o.ID, "refund", "质量问题", "", 10000); err != nil {
		t.Fatalf("申请售后: %v", err)
	}
	if _, err := tradeSvc.ReviewAftersale(ctx, o.ID, true); err != nil {
		t.Fatalf("首次同意售后: %v", err)
	}
	if b, _ := balanceOf(t, ctx, escrowSvc, "seller-1"); b != 0 {
		t.Fatalf("售后退款后卖家应为 0，实际 %d", b)
	}

	// 第二次审批：必须被拒（订单已不在 aftersale/pending）
	if _, err := tradeSvc.ReviewAftersale(ctx, o.ID, true); err == nil {
		t.Fatal("重复审批必须被拒绝，否则会把卖家余额再扣一次")
	}
	if b, _ := balanceOf(t, ctx, escrowSvc, "seller-1"); b != 0 {
		t.Fatalf("重复审批不得二次扣款：卖家应仍为 0，实际 %d", b)
	}
	if b, _ := balanceOf(t, ctx, escrowSvc, "buyer-1"); b != 10000 {
		t.Fatalf("买家应只收到一次退款（10000），实际 %d", b)
	}
}
