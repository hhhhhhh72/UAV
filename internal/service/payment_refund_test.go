package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

// stubRefundGateway 退款网关桩：把微信侧的状态完全交给用例摆布。
type stubRefundGateway struct {
	createResult service.RefundResult
	createErr    error
	queryResult  service.RefundResult
	queryErr     error
	lastCreate   service.RefundRequest
	createCalls  int
	queryCalls   int
	notifyRef    string
}

func (g *stubRefundGateway) CreateRefund(_ context.Context, in service.RefundRequest) (service.RefundResult, error) {
	g.createCalls++
	g.lastCreate = in
	if g.createErr != nil {
		return service.RefundResult{}, g.createErr
	}
	return g.createResult, nil
}

func (g *stubRefundGateway) QueryRefund(_ context.Context, _ string) (service.RefundResult, error) {
	g.queryCalls++
	if g.queryErr != nil {
		return service.RefundResult{}, g.queryErr
	}
	return g.queryResult, nil
}

func (g *stubRefundGateway) DecodeRefundNotify([]byte) (string, error) { return g.notifyRef, nil }

// newRefundFixture 装一套内存后端：充值单已支付 + 用户余额已到账，可直接发起退款。
func newRefundFixture(t *testing.T, gw service.RefundGateway, paidFen int64) (*service.PaymentRefundService, *service.EscrowService, string) {
	t.Helper()
	ctx := context.Background()
	orderRepo := memory.NewPaymentOrderRepository()
	refundRepo := memory.NewPaymentRefundRepository()
	escrow := service.NewEscrowService(memory.NewEscrowRepository())

	const userID = "payer-1"
	order, err := orderRepo.Create(ctx, domain.PaymentOrder{
		ID: "pay-1", OutTradeNo: "RC-TEST-1", UserID: userID, AmountFen: paidFen,
		Channel: domain.ChannelWeChat, Status: domain.PaymentCreated,
	})
	if err != nil {
		t.Fatalf("seed order: %v", err)
	}
	if _, err := orderRepo.MarkPaid(ctx, order.OutTradeNo, "wx-tx-1", time.Now()); err != nil {
		t.Fatalf("mark paid: %v", err)
	}
	// 用户余额：即真实支付回调入账后的状态。
	if _, _, err := escrow.DepositFromChannel(ctx, userID, paidFen, domain.ChannelWeChat, "wx-tx-1"); err != nil {
		t.Fatalf("deposit: %v", err)
	}
	return service.NewPaymentRefundService(orderRepo, refundRepo, escrow, gw), escrow, order.OutTradeNo
}

func balances(t *testing.T, esc *service.EscrowService, userID string) (int64, int64) {
	t.Helper()
	acc, err := esc.Balance(context.Background(), userID)
	if err != nil {
		t.Fatalf("查余额: %v", err)
	}
	return acc.BalanceFen, acc.FrozenFen
}

// TestRefundInitiateFreezesAndReserves 发起退款：钱先冻结、额度先占用，状态 processing。
func TestRefundInitiateFreezesAndReserves(t *testing.T) {
	gw := &stubRefundGateway{createResult: service.RefundResult{RefundID: "wx-rf-1", Status: service.RefundStateProcessing}}
	svc, esc, outTradeNo := newRefundFixture(t, gw, 10000)

	rf, err := svc.Initiate(context.Background(), outTradeNo, 3000, "用户申请", "admin-1")
	if err != nil {
		t.Fatalf("Initiate: %v", err)
	}
	if rf.Status != domain.PaymentRefundProcessing || rf.RefundID != "wx-rf-1" {
		t.Errorf("退款单状态不对：%+v", rf)
	}
	if b, f := balances(t, esc, "payer-1"); b != 7000 || f != 3000 {
		t.Fatalf("发起后应余额 7000 / 冻结 3000，实际 %d / %d", b, f)
	}
	if gw.lastCreate.RefundFen != 3000 || gw.lastCreate.TotalFen != 10000 {
		t.Errorf("发给微信的请求不对：%+v", gw.lastCreate)
	}
}

// TestRefundInitiateRejectsOverOrderAmount 累计退款不得超过原单金额（库级额度）。
func TestRefundInitiateRejectsOverOrderAmount(t *testing.T) {
	gw := &stubRefundGateway{createResult: service.RefundResult{RefundID: "wx-rf-1", Status: service.RefundStateProcessing}}
	svc, esc, outTradeNo := newRefundFixture(t, gw, 10000)
	ctx := context.Background()

	if _, err := svc.Initiate(ctx, outTradeNo, 6000, "第一次", "admin-1"); err != nil {
		t.Fatalf("第一笔: %v", err)
	}
	if _, err := svc.Initiate(ctx, outTradeNo, 6000, "超额", "admin-1"); err == nil {
		t.Fatal("累计超原单金额应当被拒")
	}
	// 第二笔必须完全没发生：不能有第二次冻结、也不能有第二次调用微信。
	if b, f := balances(t, esc, "payer-1"); b != 4000 || f != 6000 {
		t.Fatalf("被拒的退款不该动钱，实际余额 %d / 冻结 %d", b, f)
	}
	if gw.createCalls != 1 {
		t.Fatalf("被拒的退款不该调用微信，实际调用 %d 次", gw.createCalls)
	}
}

// TestRefundInitiateRollsBackWhenBalanceShort 余额不足时发起就失败，且额度要还回去。
func TestRefundInitiateRollsBackWhenBalanceShort(t *testing.T) {
	gw := &stubRefundGateway{createResult: service.RefundResult{RefundID: "wx-rf-1", Status: service.RefundStateProcessing}}
	svc, esc, outTradeNo := newRefundFixture(t, gw, 10000)
	ctx := context.Background()

	// 先把余额花掉（冻结到别的业务单上），模拟钱已不在账上
	if _, err := esc.Freeze(ctx, "payer-1", 9000, "trade_order", "torder-x"); err != nil {
		t.Fatalf("占住余额: %v", err)
	}
	if _, err := svc.Initiate(ctx, outTradeNo, 5000, "余额不够", "admin-1"); err == nil {
		t.Fatal("余额不足应当发起失败")
	}
	if _, f := balances(t, esc, "payer-1"); f != 9000 {
		t.Fatalf("被拒的退款不该新增冻结，实际冻结 %d", f)
	}
	if gw.createCalls != 0 {
		t.Fatalf("余额不足时不该调用微信，实际 %d 次", gw.createCalls)
	}
	// 额度已归还：此时再全额退款应当能通过额度校验（虽然余额仍不足，但错误不同）
	if _, err := svc.Initiate(ctx, outTradeNo, 10000, "再试", "admin-1"); err == nil {
		t.Fatal("余额仍不足，应当失败")
	}
}

// TestRefundConfirmSuccessWithdraws 退款成功：冻结被扣掉（钱离开平台），状态置 success。
func TestRefundConfirmSuccessWithdraws(t *testing.T) {
	gw := &stubRefundGateway{createResult: service.RefundResult{RefundID: "wx-rf-1", Status: service.RefundStateProcessing}}
	svc, esc, outTradeNo := newRefundFixture(t, gw, 10000)
	ctx := context.Background()

	rf, err := svc.Initiate(ctx, outTradeNo, 3000, "退款", "admin-1")
	if err != nil {
		t.Fatalf("Initiate: %v", err)
	}
	gw.queryResult = service.RefundResult{RefundID: "wx-rf-1", Status: service.RefundStateSuccess, RefundFen: 3000}

	got, settled, err := svc.Confirm(ctx, rf.OutRefundNo)
	if err != nil {
		t.Fatalf("Confirm: %v", err)
	}
	if !settled || got.Status != domain.PaymentRefundSuccess {
		t.Fatalf("应当结清为 success，实际 settled=%v status=%s", settled, got.Status)
	}
	// 钱真的离开平台：余额与冻结**都**少了 3000
	if b, f := balances(t, esc, "payer-1"); b != 7000 || f != 0 {
		t.Fatalf("成功后应余额 7000 / 冻结 0，实际 %d / %d", b, f)
	}
}

// TestRefundConfirmIsIdempotent 回调会重试：重复 Confirm 不能再扣一次冻结。
func TestRefundConfirmIsIdempotent(t *testing.T) {
	gw := &stubRefundGateway{createResult: service.RefundResult{RefundID: "wx-rf-1", Status: service.RefundStateProcessing}}
	svc, esc, outTradeNo := newRefundFixture(t, gw, 10000)
	ctx := context.Background()

	rf, _ := svc.Initiate(ctx, outTradeNo, 3000, "退款", "admin-1")
	gw.queryResult = service.RefundResult{RefundID: "wx-rf-1", Status: service.RefundStateSuccess, RefundFen: 3000}

	for i := 0; i < 3; i++ {
		if _, _, err := svc.Confirm(ctx, rf.OutRefundNo); err != nil {
			t.Fatalf("第 %d 次 Confirm: %v", i+1, err)
		}
	}
	if b, f := balances(t, esc, "payer-1"); b != 7000 || f != 0 {
		t.Fatalf("重复确认不得重复扣减，实际余额 %d / 冻结 %d", b, f)
	}
}

// TestRefundConfirmClosedUnfreezes 微信侧关闭：解冻 + 归还额度，钱留在用户余额里。
func TestRefundConfirmClosedUnfreezes(t *testing.T) {
	gw := &stubRefundGateway{createResult: service.RefundResult{RefundID: "wx-rf-1", Status: service.RefundStateProcessing}}
	svc, esc, outTradeNo := newRefundFixture(t, gw, 10000)
	ctx := context.Background()

	rf, _ := svc.Initiate(ctx, outTradeNo, 3000, "退款", "admin-1")
	gw.queryResult = service.RefundResult{RefundID: "wx-rf-1", Status: service.RefundStateClosed}

	got, settled, err := svc.Confirm(ctx, rf.OutRefundNo)
	if err != nil {
		t.Fatalf("Confirm: %v", err)
	}
	if !settled || got.Status != domain.PaymentRefundClosed {
		t.Fatalf("应当结清为 closed，实际 settled=%v status=%s", settled, got.Status)
	}
	if b, f := balances(t, esc, "payer-1"); b != 10000 || f != 0 {
		t.Fatalf("关闭后钱应全额回到余额，实际余额 %d / 冻结 %d", b, f)
	}
}

// TestRefundInitiateRollsBackOnWechatRejection 微信**明确拒绝**：安全回滚。
func TestRefundInitiateRollsBackOnWechatRejection(t *testing.T) {
	gw := &stubRefundGateway{createErr: &service.RefundRejectedError{Reason: "NOT_ENOUGH"}}
	svc, esc, outTradeNo := newRefundFixture(t, gw, 10000)
	ctx := context.Background()

	rf, err := svc.Initiate(ctx, outTradeNo, 3000, "退款", "admin-1")
	if err == nil {
		t.Fatal("微信拒绝应当返回错误")
	}
	if rf.Status != domain.PaymentRefundClosed {
		t.Errorf("被拒的退款单应为 closed，实际 %s", rf.Status)
	}
	if b, f := balances(t, esc, "payer-1"); b != 10000 || f != 0 {
		t.Fatalf("被拒后钱应全额回到余额，实际余额 %d / 冻结 %d", b, f)
	}
}

// TestRefundInitiateKeepsHoldOnUnknownError **结果未知**时绝不能回滚：
// 微信可能已经受理，回滚等于钱退出去、平台又把余额还给了用户。
func TestRefundInitiateKeepsHoldOnUnknownError(t *testing.T) {
	gw := &stubRefundGateway{createErr: errors.New("dial tcp: i/o timeout")}
	svc, esc, outTradeNo := newRefundFixture(t, gw, 10000)
	ctx := context.Background()

	if _, err := svc.Initiate(ctx, outTradeNo, 3000, "退款", "admin-1"); err == nil {
		t.Fatal("未知错误应当返回错误")
	}
	// 冻结必须保留：钱不能既退出去又留在余额里
	if b, f := balances(t, esc, "payer-1"); b != 7000 || f != 3000 {
		t.Fatalf("结果未知时必须保留冻结，实际余额 %d / 冻结 %d", b, f)
	}
}

// TestRefundDisabled / 未装配网关时所有入口都回 ErrPaymentDisabled。
func TestRefundDisabled(t *testing.T) {
	svc := service.NewPaymentRefundService(memory.NewPaymentOrderRepository(), memory.NewPaymentRefundRepository(), service.NewEscrowService(memory.NewEscrowRepository()), nil)
	ctx := context.Background()
	if svc.Enabled() {
		t.Fatal("网关为 nil 时 Enabled() 应为 false")
	}
	if _, err := svc.Initiate(ctx, "RC1", 100, "", "a"); !errors.Is(err, service.ErrPaymentDisabled) {
		t.Errorf("Initiate 应返回 ErrPaymentDisabled，实际 %v", err)
	}
	if _, _, err := svc.Confirm(ctx, "RF1"); !errors.Is(err, service.ErrPaymentDisabled) {
		t.Errorf("Confirm 应返回 ErrPaymentDisabled，实际 %v", err)
	}
}

// TestNewOutRefundNoIsWellFormed 退款单号是微信回调的唯一对账主键，格式与唯一性都要守住。
func TestNewOutRefundNoIsWellFormed(t *testing.T) {
	seen := map[string]bool{}
	now := time.Now()
	for i := 0; i < 200; i++ {
		no, err := service.NewOutRefundNo(now)
		if err != nil {
			t.Fatalf("NewOutRefundNo: %v", err)
		}
		if len(no) < 6 || len(no) > 32 || no[:2] != "RF" {
			t.Fatalf("退款单号格式不符：%q", no)
		}
		if seen[no] {
			t.Fatalf("同一秒内出现重复退款单号：%s", no)
		}
		seen[no] = true
	}
}