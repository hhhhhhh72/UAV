package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

// 商城订单资金闭环：付款冻结买家余额 → 确认收货放款给卖家 → 取消/售后退款。
//
// 背景：此前订单只有状态机、钱一分不动（付款不校验余额、完成不结算、取消/售后不退款），
// 等于"假交易"。这组用例把四个资金动作与幂等要求钉死。
func newMoneyFlow(t *testing.T) (*service.TradeOrderService, repository.ProductRepository, *service.EscrowService, context.Context) {
	t.Helper()
	ctx := context.Background()
	escrowSvc := service.NewEscrowService(memory.NewEscrowRepository())
	prodRepo := memory.NewProductRepository()
	tradeSvc := service.NewTradeOrderService(memory.NewTradeOrderRepository(), prodRepo)
	tradeSvc.SetEscrow(escrowSvc)
	return tradeSvc, prodRepo, escrowSvc, ctx
}

// seedProduct 直接落一个「在售」商品（用户发布默认待审核，下单要求 listed）。
func seedProduct(t *testing.T, ctx context.Context, prodRepo repository.ProductRepository, sellerID string, priceFen int64) string {
	t.Helper()
	p, err := prodRepo.Create(ctx, domain.DroneProduct{
		ID: "prod-" + sellerID + "-" + time.Now().Format("150405.000000000"), SellerID: sellerID,
		ProdType: domain.ProductType("drone"), Title: "测试商品", PriceFen: priceFen, Status: "listed",
	})
	if err != nil {
		t.Fatalf("建商品: %v", err)
	}
	return p.ID
}

func balanceOf(t *testing.T, ctx context.Context, e *service.EscrowService, userID string) (int64, int64) {
	t.Helper()
	acc, err := e.Balance(ctx, userID)
	if err != nil {
		t.Fatalf("查余额: %v", err)
	}
	return acc.BalanceFen, acc.FrozenFen
}

// 全流程：充值 → 下单 → 付款（冻结）→ 确认收货（放款给卖家）。
func TestTradeOrderMoneyHappyPath(t *testing.T) {
	tradeSvc, prodRepo, escrowSvc, ctx := newMoneyFlow(t)
	productID := seedProduct(t, ctx, prodRepo, "seller-1", 10000) // 100 元
	if _, err := escrowSvc.Deposit(ctx, "buyer-1", 10000); err != nil {
		t.Fatalf("充值: %v", err)
	}

	o, err := tradeSvc.Create(ctx, "buyer-1", productID, "seller-1", 10000, service.OrderReceiver{}, false)
	if err != nil {
		t.Fatalf("下单: %v", err)
	}
	if b, f := balanceOf(t, ctx, escrowSvc, "buyer-1"); b != 10000 || f != 0 {
		t.Fatalf("下单不应动钱，实际余额=%d 冻结=%d", b, f)
	}

	if _, err := tradeSvc.PayOrder(ctx, "buyer-1", o.ID); err != nil {
		t.Fatalf("付款: %v", err)
	}
	if b, f := balanceOf(t, ctx, escrowSvc, "buyer-1"); b != 0 || f != 10000 {
		t.Fatalf("付款后买家应余额0/冻结10000，实际余额=%d 冻结=%d", b, f)
	}
	if b, _ := balanceOf(t, ctx, escrowSvc, "seller-1"); b != 0 {
		t.Fatalf("付款阶段卖家不应到账，实际 %d", b)
	}

	if _, err := tradeSvc.ShipOrder(ctx, domain.Actor{ID: "seller-1", Role: domain.RoleEnterprise}, o.ID, "顺丰速运", "SF-TEST-1"); err != nil {
		t.Fatalf("发货: %v", err)
	}
	if _, err := tradeSvc.UpdateStatus(ctx, o.ID, "buyer-1", "completed"); err != nil {
		t.Fatalf("确认收货: %v", err)
	}
	if b, f := balanceOf(t, ctx, escrowSvc, "buyer-1"); b != 0 || f != 0 {
		t.Fatalf("放款后买家应余额0/冻结0，实际余额=%d 冻结=%d", b, f)
	}
	if b, _ := balanceOf(t, ctx, escrowSvc, "seller-1"); b != 10000 {
		t.Fatalf("放款后卖家应到账 10000，实际 %d", b)
	}

	// 幂等：重复放款不双倍入账（EscrowService.Release 按 refID 查重）
	if _, err := escrowSvc.Release(ctx, "buyer-1", "seller-1", 10000, "trade_order", o.ID); err != nil {
		t.Fatalf("重复放款应幂等成功: %v", err)
	}
	if b, _ := balanceOf(t, ctx, escrowSvc, "seller-1"); b != 10000 {
		t.Fatalf("重复放款导致了双倍入账，卖家余额=%d", b)
	}
}

// 余额不足：付款被拒（ErrInsufficientBalance → Handler 402），订单保持 pending 且不动钱。
func TestTradeOrderPayInsufficientBalance(t *testing.T) {
	tradeSvc, prodRepo, escrowSvc, ctx := newMoneyFlow(t)
	productID := seedProduct(t, ctx, prodRepo, "seller-1", 50000)
	o, err := tradeSvc.Create(ctx, "buyer-1", productID, "seller-1", 50000, service.OrderReceiver{}, false)
	if err != nil {
		t.Fatalf("下单: %v", err)
	}
	_, err = tradeSvc.PayOrder(ctx, "buyer-1", o.ID)
	if !errors.Is(err, repository.ErrInsufficientBalance) {
		t.Fatalf("余额为 0 付款应 ErrInsufficientBalance，实际 %v", err)
	}
	cur, _ := tradeSvc.FindByID(ctx, o.ID)
	if cur.Status != "pending" {
		t.Fatalf("付款失败后订单应仍为 pending，实际 %s", cur.Status)
	}
	if b, f := balanceOf(t, ctx, escrowSvc, "buyer-1"); b != 0 || f != 0 {
		t.Fatalf("付款失败不应动钱，实际余额=%d 冻结=%d", b, f)
	}
}

// 取消订单：解冻退款给买家（幂等，重复取消不重复退款）。
func TestTradeOrderCancelRefunds(t *testing.T) {
	tradeSvc, prodRepo, escrowSvc, ctx := newMoneyFlow(t)
	productID := seedProduct(t, ctx, prodRepo, "seller-1", 20000)
	if _, err := escrowSvc.Deposit(ctx, "buyer-1", 20000); err != nil {
		t.Fatalf("充值: %v", err)
	}
	o, _ := tradeSvc.Create(ctx, "buyer-1", productID, "seller-1", 20000, service.OrderReceiver{}, false)
	if _, err := tradeSvc.PayOrder(ctx, "buyer-1", o.ID); err != nil {
		t.Fatalf("付款: %v", err)
	}
	if _, err := tradeSvc.UpdateStatusAdmin(ctx, o.ID, "cancelled"); err != nil {
		t.Fatalf("取消: %v", err)
	}
	if b, f := balanceOf(t, ctx, escrowSvc, "buyer-1"); b != 20000 || f != 0 {
		t.Fatalf("取消后应全额退回（余额20000/冻结0），实际余额=%d 冻结=%d", b, f)
	}
	// 重复退款幂等
	if _, err := escrowSvc.Refund(ctx, "buyer-1", 20000, "trade_order", o.ID); err == nil {
		t.Fatal("冻结已释放，再退款应失败（不能凭空加钱）")
	}
	if b, _ := balanceOf(t, ctx, escrowSvc, "buyer-1"); b != 20000 {
		t.Fatalf("重复退款导致余额异常：%d", b)
	}
}

// 售后（货款还在冻结里）：部分退款给买家，剩余放给卖家。
func TestTradeOrderAftersaleBeforeSettle(t *testing.T) {
	tradeSvc, prodRepo, escrowSvc, ctx := newMoneyFlow(t)
	productID := seedProduct(t, ctx, prodRepo, "seller-1", 10000)
	if _, err := escrowSvc.Deposit(ctx, "buyer-1", 10000); err != nil {
		t.Fatalf("充值: %v", err)
	}
	o, _ := tradeSvc.Create(ctx, "buyer-1", productID, "seller-1", 10000, service.OrderReceiver{}, false)
	if _, err := tradeSvc.PayOrder(ctx, "buyer-1", o.ID); err != nil {
		t.Fatalf("付款: %v", err)
	}
	if _, err := tradeSvc.ApplyAftersale(ctx, "buyer-1", o.ID, "refund", "有瑕疵", "协商部分退款", 3000); err != nil {
		t.Fatalf("申请售后: %v", err)
	}
	if _, err := tradeSvc.ReviewAftersale(ctx, o.ID, true); err != nil {
		t.Fatalf("同意售后: %v", err)
	}
	if b, f := balanceOf(t, ctx, escrowSvc, "buyer-1"); b != 3000 || f != 0 {
		t.Fatalf("售后应退买家 3000，实际余额=%d 冻结=%d", b, f)
	}
	if b, _ := balanceOf(t, ctx, escrowSvc, "seller-1"); b != 7000 {
		t.Fatalf("剩余 7000 应放给卖家，实际 %d", b)
	}
}

// 售后（货款已放给卖家）：从卖家余额扣回给买家；卖家余额不足则审批失败且状态不变。
func TestTradeOrderAftersaleAfterSettle(t *testing.T) {
	tradeSvc, prodRepo, escrowSvc, ctx := newMoneyFlow(t)
	productID := seedProduct(t, ctx, prodRepo, "seller-1", 10000)
	if _, err := escrowSvc.Deposit(ctx, "buyer-1", 10000); err != nil {
		t.Fatalf("充值: %v", err)
	}
	o, _ := tradeSvc.Create(ctx, "buyer-1", productID, "seller-1", 10000, service.OrderReceiver{}, false)
	if _, err := tradeSvc.PayOrder(ctx, "buyer-1", o.ID); err != nil {
		t.Fatalf("付款: %v", err)
	}
	if _, err := tradeSvc.ShipOrder(ctx, domain.Actor{ID: "seller-1", Role: domain.RoleEnterprise}, o.ID, "顺丰速运", "SF-TEST-1"); err != nil {
		t.Fatalf("发货: %v", err)
	}
	if _, err := tradeSvc.UpdateStatus(ctx, o.ID, "buyer-1", "completed"); err != nil {
		t.Fatalf("确认收货: %v", err)
	}
	if _, err := tradeSvc.ApplyAftersale(ctx, "buyer-1", o.ID, "refund", "质量问题", "全额退", 10000); err != nil {
		t.Fatalf("申请售后: %v", err)
	}
	if _, err := tradeSvc.ReviewAftersale(ctx, o.ID, true); err != nil {
		t.Fatalf("同意售后: %v", err)
	}
	if b, _ := balanceOf(t, ctx, escrowSvc, "buyer-1"); b != 10000 {
		t.Fatalf("退款应回到买家，实际 %d", b)
	}
	if b, _ := balanceOf(t, ctx, escrowSvc, "seller-1"); b != 0 {
		t.Fatalf("卖家余额应被扣回，实际 %d", b)
	}
}

// 售后被驳回：订单状态回到售后前状态，且买家可以重新申请（此前被一票否决）。
func TestTradeOrderAftersaleRejectedCanReapply(t *testing.T) {
	tradeSvc, prodRepo, escrowSvc, ctx := newMoneyFlow(t)
	productID := seedProduct(t, ctx, prodRepo, "seller-1", 10000)
	if _, err := escrowSvc.Deposit(ctx, "buyer-1", 10000); err != nil {
		t.Fatalf("充值: %v", err)
	}
	o, _ := tradeSvc.Create(ctx, "buyer-1", productID, "seller-1", 10000, service.OrderReceiver{}, false)
	if _, err := tradeSvc.PayOrder(ctx, "buyer-1", o.ID); err != nil {
		t.Fatalf("付款: %v", err)
	}
	if _, err := tradeSvc.ApplyAftersale(ctx, "buyer-1", o.ID, "refund", "第一次", "", 5000); err != nil {
		t.Fatalf("第一次申请: %v", err)
	}
	cur, err := tradeSvc.ReviewAftersale(ctx, o.ID, false)
	if err != nil {
		t.Fatalf("驳回: %v", err)
	}
	if cur.Status != "paid" {
		t.Fatalf("驳回后应回到 paid（实际 %s）——否则卖家无法发货", cur.Status)
	}
	if _, err := tradeSvc.ApplyAftersale(ctx, "buyer-1", o.ID, "refund", "第二次", "", 5000); err != nil {
		t.Fatalf("驳回后应允许重新申请，实际 %v", err)
	}
	// 未退款的驳回不应动钱
	if b, f := balanceOf(t, ctx, escrowSvc, "buyer-1"); b != 0 || f != 10000 {
		t.Fatalf("驳回不应动钱，实际余额=%d 冻结=%d", b, f)
	}
}
