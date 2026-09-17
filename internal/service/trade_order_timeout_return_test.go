package service_test

import (
	"testing"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/service"
)

// ── 超时维护：自动确认收货 / 支付超时自动取消 ──
//
// 用"未来 cutoff"把目标订单一次性纳入扫描窗口，避免 sleep 或注入时钟：
// 扫描语义是 updated_at/created_at < cutoff，未来时间戳即"全部命中"。

// 自动确认收货：发货后买家长期不确认 → 自动完成并放款给卖家。
func TestTradeOrderAutoConfirmShipped(t *testing.T) {
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

	// 买家不点确认收货：钱一直冻结，卖家一分没有（这正是修复前的死局）
	if b, f := balanceOf(t, ctx, escrowSvc, "buyer-1"); b != 0 || f != 10000 {
		t.Fatalf("发货后应仍冻结买家 10000，实际 balance=%d frozen=%d", b, f)
	}
	if b, _ := balanceOf(t, ctx, escrowSvc, "seller-1"); b != 0 {
		t.Fatalf("卖家此时不应有钱，实际 %d", b)
	}

	n, err := tradeSvc.AutoConfirmShipped(ctx, time.Now().Add(time.Hour), 100)
	if err != nil || n != 1 {
		t.Fatalf("应自动确认 1 单，实际 n=%d err=%v", n, err)
	}
	got, err := tradeSvc.FindByID(ctx, o.ID)
	if err != nil || got.Status != "completed" {
		t.Fatalf("订单应 completed，实际 %s err=%v", got.Status, err)
	}
	if b, _ := balanceOf(t, ctx, escrowSvc, "seller-1"); b != 10000 {
		t.Fatalf("卖家应到账 10000，实际 %d", b)
	}
	if b, f := balanceOf(t, ctx, escrowSvc, "buyer-1"); b != 0 || f != 0 {
		t.Fatalf("买家冻结应清零，实际 balance=%d frozen=%d", b, f)
	}

	// 幂等：已完成的订单不再被扫描到，也不会重复放款
	if n2, _ := tradeSvc.AutoConfirmShipped(ctx, time.Now().Add(time.Hour), 100); n2 != 0 {
		t.Fatalf("重复扫描不应再处理，实际 n=%d", n2)
	}
	if b, _ := balanceOf(t, ctx, escrowSvc, "seller-1"); b != 10000 {
		t.Fatalf("重复扫描不得重复放款，实际 %d", b)
	}
}

// 支付超时自动取消：未付款订单被关闭，且商品恢复为可售（否则一直锁货）。
func TestTradeOrderAutoCancelUnpaid(t *testing.T) {
	tradeSvc, prodRepo, escrowSvc, ctx := newMoneyFlow(t)
	productID := seedProduct(t, ctx, prodRepo, "seller-1", 10000)
	o, err := tradeSvc.Create(ctx, "buyer-1", productID, "seller-1", 10000, service.OrderReceiver{}, false)
	if err != nil {
		t.Fatalf("下单: %v", err)
	}
	// 复现 handler 的下单占位（service.Create 本身不动商品状态）
	if err := prodRepo.MarkSold(ctx, productID); err != nil {
		t.Fatalf("占位: %v", err)
	}
	if p, _ := prodRepo.FindByID(ctx, productID); p.Status != "sold" {
		t.Fatalf("下单后商品应为 sold，实际 %s", p.Status)
	}

	n, err := tradeSvc.AutoCancelUnpaid(ctx, time.Now().Add(time.Hour), 100)
	if err != nil || n != 1 {
		t.Fatalf("应自动取消 1 单，实际 n=%d err=%v", n, err)
	}
	got, err := tradeSvc.FindByID(ctx, o.ID)
	if err != nil || got.Status != "cancelled" {
		t.Fatalf("订单应 cancelled，实际 %s err=%v", got.Status, err)
	}
	if p, _ := prodRepo.FindByID(ctx, productID); p.Status != "listed" {
		t.Fatalf("超时取消后商品应恢复 listed（重新可售），实际 %s", p.Status)
	}
	// 未付款订单没有冻结资金，不应凭空产生退款流水
	if b, f := balanceOf(t, ctx, escrowSvc, "buyer-1"); b != 0 || f != 0 {
		t.Fatalf("未付款订单不应产生资金变动，实际 balance=%d frozen=%d", b, f)
	}
}

// ── 退货退款全流程 ──

// 退货退款：同意退货只推进状态、不退款；买家寄回、卖家确认收到后**才**退款。
func TestTradeOrderReturnFlowRefundsOnlyAfterGoodsReceived(t *testing.T) {
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
	if _, err := tradeSvc.ApplyAftersale(ctx, "buyer-1", o.ID, "return", "货不对板", "", 10000); err != nil {
		t.Fatalf("申请退货退款: %v", err)
	}

	// ① 卖家同意退货：只到"待买家寄回"，钱一分不动
	cur, err := tradeSvc.ReviewAftersale(ctx, o.ID, true)
	if err != nil {
		t.Fatalf("同意退货: %v", err)
	}
	if cur.AftersaleStatus != "returning" {
		t.Fatalf("应进入 returning(待买家寄回)，实际 %s", cur.AftersaleStatus)
	}
	if cur.Status != "aftersale" {
		t.Fatalf("订单应仍在 aftersale，实际 %s", cur.Status)
	}
	if b, f := balanceOf(t, ctx, escrowSvc, "buyer-1"); b != 0 || f != 10000 {
		t.Fatalf("同意退货阶段不得动钱，实际 balance=%d frozen=%d", b, f)
	}

	// ② 未寄回就确认收货 → 拒绝；非买家提交物流 → 拒绝；空单号 → 拒绝
	seller := domain.Actor{ID: "seller-1", Role: domain.RoleEnterprise}
	if _, err := tradeSvc.ConfirmReturnReceived(ctx, seller, o.ID); err == nil {
		t.Fatal("买家尚未寄回，确认收到退货必须被拒绝")
	}
	if _, err := tradeSvc.SubmitReturnShipment(ctx, "buyer-2", o.ID, "SF123", ""); err == nil {
		t.Fatal("非买家不得提交退货物流")
	}
	if _, err := tradeSvc.SubmitReturnShipment(ctx, "buyer-1", o.ID, "   ", ""); err == nil {
		t.Fatal("空物流单号必须被拒绝")
	}

	// ③ 买家寄回：记录物流，仍不动钱
	ret, err := tradeSvc.SubmitReturnShipment(ctx, "buyer-1", o.ID, "SF123456789", "已寄出")
	if err != nil {
		t.Fatalf("提交退货物流: %v", err)
	}
	if ret.AftersaleStatus != "returned" || ret.ReturnTracking != "SF123456789" || ret.ReturnedAt == nil {
		t.Fatalf("退货物流未正确记录: %+v", ret)
	}
	if b, f := balanceOf(t, ctx, escrowSvc, "buyer-1"); b != 0 || f != 10000 {
		t.Fatalf("寄回阶段仍不得动钱，实际 balance=%d frozen=%d", b, f)
	}
	// 重复提交物流：CAS 期望值已是 returned，必须被拒
	if _, err := tradeSvc.SubmitReturnShipment(ctx, "buyer-1", o.ID, "SF999", ""); err == nil {
		t.Fatal("重复提交退货物流必须被拒绝")
	}

	// ④ 卖家确认收到 → 此刻才退款结案
	done, err := tradeSvc.ConfirmReturnReceived(ctx, seller, o.ID)
	if err != nil {
		t.Fatalf("确认收到退货: %v", err)
	}
	if done.AftersaleStatus != "approved" || done.Status != "completed" {
		t.Fatalf("应结案 approved/completed，实际 aftersale=%s status=%s", done.AftersaleStatus, done.Status)
	}
	if b, f := balanceOf(t, ctx, escrowSvc, "buyer-1"); b != 10000 || f != 0 {
		t.Fatalf("退款应回到买家，实际 balance=%d frozen=%d", b, f)
	}
	// 重复确认不得二次退款
	if _, err := tradeSvc.ConfirmReturnReceived(ctx, seller, o.ID); err == nil {
		t.Fatal("重复确认必须被拒绝")
	}
	if b, _ := balanceOf(t, ctx, escrowSvc, "buyer-1"); b != 10000 {
		t.Fatalf("重复确认不得重复退款，实际 %d", b)
	}
}

// 仅退款（refund）保持原语义：卖家同意即退款，不经过寄回环节。
func TestTradeOrderRefundTypeStillRefundsImmediately(t *testing.T) {
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
	if _, err := tradeSvc.ApplyAftersale(ctx, "buyer-1", o.ID, "refund", "不想要了", "", 10000); err != nil {
		t.Fatalf("申请仅退款: %v", err)
	}
	done, err := tradeSvc.ReviewAftersale(ctx, o.ID, true)
	if err != nil {
		t.Fatalf("同意仅退款: %v", err)
	}
	if done.AftersaleStatus != "approved" || done.Status != "completed" {
		t.Fatalf("仅退款应直接结案，实际 aftersale=%s status=%s", done.AftersaleStatus, done.Status)
	}
	if b, f := balanceOf(t, ctx, escrowSvc, "buyer-1"); b != 10000 || f != 0 {
		t.Fatalf("仅退款应立即退回买家，实际 balance=%d frozen=%d", b, f)
	}
}

// 售后类型白名单：脏值必须被拒——否则"退货退款"会被静默存成非法值而走成仅退款。
func TestTradeOrderAftersaleRejectsUnknownType(t *testing.T) {
	tradeSvc, prodRepo, _, ctx := newMoneyFlow(t)
	productID := seedProduct(t, ctx, prodRepo, "seller-1", 10000)
	o, err := tradeSvc.Create(ctx, "buyer-1", productID, "seller-1", 10000, service.OrderReceiver{}, false)
	if err != nil {
		t.Fatalf("下单: %v", err)
	}
	for _, bad := range []string{"", "RETURN", "exchange", "退货"} {
		if _, err := tradeSvc.ApplyAftersale(ctx, "buyer-1", o.ID, bad, "理由", "", 100); err == nil {
			t.Fatalf("非法售后类型 %q 必须被拒绝", bad)
		}
	}
}

// 退货单进入 returning/returned 后不得重复申请（与 pending/approved 同口径）。
func TestTradeOrderReturnBlocksReapply(t *testing.T) {
	tradeSvc, prodRepo, escrowSvc, ctx := newMoneyFlow(t)
	productID := seedProduct(t, ctx, prodRepo, "seller-1", 10000)
	if _, err := escrowSvc.Deposit(ctx, "buyer-1", 10000); err != nil {
		t.Fatalf("充值: %v", err)
	}
	o, err := tradeSvc.Create(ctx, "buyer-1", productID, "seller-1", 10000, service.OrderReceiver{}, false)
	if err != nil {
		t.Fatalf("下单: %v", err)
	}
	// 售后只在 paid/shipped/completed 可申请（pending→aftersale 非法），先付款
	if _, err := tradeSvc.PayOrder(ctx, "buyer-1", o.ID); err != nil {
		t.Fatalf("付款: %v", err)
	}
	if _, err := tradeSvc.ApplyAftersale(ctx, "buyer-1", o.ID, "return", "理由", "", 100); err != nil {
		t.Fatalf("申请: %v", err)
	}
	if _, err := tradeSvc.ReviewAftersale(ctx, o.ID, true); err != nil {
		t.Fatalf("同意退货: %v", err)
	}
	// 已进入 returning：再次申请应被拒
	if _, err := tradeSvc.ApplyAftersale(ctx, "buyer-1", o.ID, "refund", "再申请", "", 100); err == nil {
		t.Fatal("退货单在 returning 阶段不得重复申请售后")
	}
}
