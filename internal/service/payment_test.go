package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

// stubGateway 支付网关桩：把「微信侧」的状态完全交给用例摆布，
// 于是到账规则（金额比对 / 交易状态判定 / 幂等）可以脱离网络精确验证。
type stubGateway struct {
	prepayID   string
	prepayErr  error
	lastPrepay service.PrepayInput

	result   service.PaymentResult
	queryErr error
	queried  []string

	params    map[string]string
	notifyRef string
	notifyErr error
}

func (g *stubGateway) Prepay(_ context.Context, in service.PrepayInput) (string, error) {
	g.lastPrepay = in
	if g.prepayErr != nil {
		return "", g.prepayErr
	}
	return g.prepayID, nil
}

func (g *stubGateway) Query(_ context.Context, outTradeNo string) (service.PaymentResult, error) {
	g.queried = append(g.queried, outTradeNo)
	if g.queryErr != nil {
		return service.PaymentResult{}, g.queryErr
	}
	return g.result, nil
}

func (g *stubGateway) PayParams(prepayID string) (map[string]string, error) {
	out := map[string]string{"package": "prepay_id=" + prepayID}
	for k, v := range g.params {
		out[k] = v
	}
	return out, nil
}

func (g *stubGateway) DecodeNotify([]byte) (string, error) {
	if g.notifyErr != nil {
		return "", g.notifyErr
	}
	return g.notifyRef, nil
}

func newPaymentFixture(gw service.PaymentGateway) (*service.PaymentService, *service.EscrowService) {
	esc := service.NewEscrowService(memory.NewEscrowRepository())
	return service.NewPaymentService(memory.NewPaymentOrderRepository(), esc, gw), esc
}

// escrowBalanceOf 读托管余额。命名带前缀是为了避开 trade_order_money_test.go 里
// 已有的同名助手（同包内重名会整包编译失败）。
func escrowBalanceOf(t *testing.T, esc *service.EscrowService, userID string) int64 {
	t.Helper()
	acc, err := esc.Balance(context.Background(), userID)
	if err != nil {
		t.Fatalf("查托管余额: %v", err)
	}
	return acc.BalanceFen
}

// TestPrepayRejectsOutOfRangeAmount 金额边界是业务规则，必须在 Service 层拦住：
// 下限挡掉 1 分钱刷单噪声，上限挡掉误操作与单笔风控敞口。
func TestPrepayRejectsOutOfRangeAmount(t *testing.T) {
	gw := &stubGateway{prepayID: "prepay-1"}
	svc, _ := newPaymentFixture(gw)
	ctx := context.Background()

	for _, amount := range []int64{0, -1, service.MinRechargeFen - 1, service.MaxRechargeFen + 1} {
		if _, _, err := svc.Prepay(ctx, "payer-1", "openid-1", amount); err == nil {
			t.Errorf("金额 %d 分应当被拒，却通过了", amount)
		}
	}
	for _, amount := range []int64{service.MinRechargeFen, service.MaxRechargeFen} {
		if _, _, err := svc.Prepay(ctx, "payer-1", "openid-1", amount); err != nil {
			t.Errorf("边界金额 %d 分应当放行，却被拒：%v", amount, err)
		}
	}
}

// TestPrepayRequiresOpenID 没有 openid 就没有付款人，微信侧会直接拒单，
// 不如在本地尽早报错并给出可读原因。
func TestPrepayRequiresOpenID(t *testing.T) {
	svc, _ := newPaymentFixture(&stubGateway{prepayID: "prepay-1"})
	if _, _, err := svc.Prepay(context.Background(), "payer-1", "  ", 10000); err == nil {
		t.Fatal("缺少 openid 应当报错")
	}
}

// TestPrepayCreatesCreatedOrderWithParams 下单成功后本地落一行 created 订单，
// 并回传 wx.requestPayment 所需参数。
func TestPrepayCreatesCreatedOrderWithParams(t *testing.T) {
	gw := &stubGateway{prepayID: "prepay-abc", params: map[string]string{"signType": "RSA"}}
	svc, _ := newPaymentFixture(gw)

	order, params, err := svc.Prepay(context.Background(), "payer-1", "openid-1", 10000)
	if err != nil {
		t.Fatalf("Prepay: %v", err)
	}
	if order.Status != domain.PaymentCreated {
		t.Errorf("下单后状态应为 %s，实际 %s", domain.PaymentCreated, order.Status)
	}
	if order.Channel != domain.ChannelWeChat {
		t.Errorf("渠道应为 %s，实际 %s", domain.ChannelWeChat, order.Channel)
	}
	if order.AmountFen != 10000 {
		t.Errorf("金额应为 10000，实际 %d", order.AmountFen)
	}
	if !strings.HasPrefix(order.OutTradeNo, "RC") || len(order.OutTradeNo) != 24 {
		t.Errorf("商户订单号格式不符：%q（应为 RC + 14 位时间 + 8 位 hex = 24 字符）", order.OutTradeNo)
	}
	if params["package"] != "prepay_id=prepay-abc" {
		t.Errorf("pay_params 未带上 prepay_id：%v", params)
	}
	// 下单时必须把商户订单号与付款人交给网关，且金额原样透传。
	if gw.lastPrepay.OutTradeNo != order.OutTradeNo {
		t.Errorf("网关收到的订单号 %q 与本地订单 %q 不一致", gw.lastPrepay.OutTradeNo, order.OutTradeNo)
	}
	if gw.lastPrepay.PayerOpenID != "openid-1" {
		t.Errorf("网关收到的 openid 不对：%q", gw.lastPrepay.PayerOpenID)
	}
}

// TestConfirmIgnoresUnpaidTradeState 微信说「还没付」就一分钱都不能加。
func TestConfirmIgnoresUnpaidTradeState(t *testing.T) {
	for _, state := range []string{"NOTPAY", "CLOSED", "REVOKED", "USERPAYING", "PAYERROR", ""} {
		gw := &stubGateway{prepayID: "p1"}
		svc, esc := newPaymentFixture(gw)
		ctx := context.Background()
		order, _, err := svc.Prepay(ctx, "payer-1", "openid-1", 10000)
		if err != nil {
			t.Fatalf("Prepay: %v", err)
		}
		gw.result = service.PaymentResult{TradeState: state, TransactionID: "wx-tx-1", AmountFen: 10000}
		_, credited, err := svc.Confirm(ctx, order.OutTradeNo)
		if err != nil {
			t.Fatalf("trade_state=%q 不应报错：%v", state, err)
		}
		if credited {
			t.Errorf("trade_state=%q 不得入账", state)
		}
		if got := escrowBalanceOf(t, esc, "payer-1"); got != 0 {
			t.Errorf("trade_state=%q 后余额应为 0，实际 %d", state, got)
		}
	}
}

// TestConfirmRejectsAmountMismatch 核心红线：金额只信查单结果，且必须与本地订单逐分相等。
// 少付、多付、串单都不得入账——否则「付 1 分充 100 元」成立。
func TestConfirmRejectsAmountMismatch(t *testing.T) {
	gw := &stubGateway{prepayID: "p1"}
	svc, esc := newPaymentFixture(gw)
	ctx := context.Background()
	order, _, err := svc.Prepay(ctx, "payer-1", "openid-1", 10000)
	if err != nil {
		t.Fatalf("Prepay: %v", err)
	}

	for _, paid := range []int64{1, 9999, 10001, 0} {
		gw.result = service.PaymentResult{TradeState: "SUCCESS", TransactionID: "wx-tx", AmountFen: paid}
		_, credited, err := svc.Confirm(ctx, order.OutTradeNo)
		if err == nil {
			t.Errorf("微信侧 %d 分 vs 本地 10000 分：应当报错", paid)
		}
		if credited {
			t.Errorf("金额不符时不得入账（微信侧 %d 分）", paid)
		}
	}
	if got := escrowBalanceOf(t, esc, "payer-1"); got != 0 {
		t.Fatalf("金额不符后余额必须为 0，实际 %d", got)
	}
	// 订单也不能被标成已支付，否则将永远无法再对账。
	again, err := svc.ListMine(ctx, "payer-1", 10)
	if err != nil {
		t.Fatalf("ListMine: %v", err)
	}
	if len(again) != 1 || again[0].Status != domain.PaymentCreated {
		t.Fatalf("金额不符后订单状态应仍为 created，实际 %+v", again)
	}
}

// TestConfirmCreditsEscrowExactlyOnce 到账 + 幂等：微信回调会重试，
// 第二次必须「不报错、不再加钱」。
func TestConfirmCreditsEscrowExactlyOnce(t *testing.T) {
	gw := &stubGateway{prepayID: "p1"}
	svc, esc := newPaymentFixture(gw)
	ctx := context.Background()
	order, _, err := svc.Prepay(ctx, "payer-1", "openid-1", 25000)
	if err != nil {
		t.Fatalf("Prepay: %v", err)
	}
	gw.result = service.PaymentResult{
		TradeState: "SUCCESS", TransactionID: "wx-tx-9", AmountFen: 25000,
		SuccessTime: time.Now(),
	}

	got, credited, err := svc.Confirm(ctx, order.OutTradeNo)
	if err != nil {
		t.Fatalf("首次确认: %v", err)
	}
	if !credited {
		t.Fatal("首次确认应当入账")
	}
	if got.Status != domain.PaymentPaid || got.TransactionID != "wx-tx-9" {
		t.Errorf("订单应置为已支付并记录微信单号，实际 %+v", got)
	}
	if b := escrowBalanceOf(t, esc, "payer-1"); b != 25000 {
		t.Fatalf("首次入账后余额应为 25000，实际 %d", b)
	}

	// 重放（微信回调重试）——必须不报错、不再加钱。
	_, credited2, err := svc.Confirm(ctx, order.OutTradeNo)
	if err != nil {
		t.Fatalf("重放不应报错：%v", err)
	}
	if credited2 {
		t.Error("重放不得再次入账")
	}
	if b := escrowBalanceOf(t, esc, "payer-1"); b != 25000 {
		t.Fatalf("重放后余额仍应为 25000，实际 %d", b)
	}
}

// TestConfirmAlwaysQueriesGateway 回调只做触发、以主动查询为准：
// 每次确认都必须真的去问一次微信，而不是相信任何本地状态。
func TestConfirmAlwaysQueriesGateway(t *testing.T) {
	gw := &stubGateway{prepayID: "p1", result: service.PaymentResult{TradeState: "NOTPAY"}}
	svc, _ := newPaymentFixture(gw)
	ctx := context.Background()
	order, _, err := svc.Prepay(ctx, "payer-1", "openid-1", 10000)
	if err != nil {
		t.Fatalf("Prepay: %v", err)
	}
	for i := 0; i < 3; i++ {
		if _, _, err := svc.Confirm(ctx, order.OutTradeNo); err != nil {
			t.Fatalf("第 %d 次确认: %v", i+1, err)
		}
	}
	if len(gw.queried) != 3 {
		t.Fatalf("应当查单 3 次，实际 %d 次：%v", len(gw.queried), gw.queried)
	}
	for _, q := range gw.queried {
		if q != order.OutTradeNo {
			t.Errorf("查单用的订单号不对：%q", q)
		}
	}
}

// TestConfirmRejectsUnknownOrder 未知订单号（伪造回调）必须被拒，且不得去打扰微信。
func TestConfirmRejectsUnknownOrder(t *testing.T) {
	gw := &stubGateway{result: service.PaymentResult{TradeState: "SUCCESS", AmountFen: 10000}}
	svc, esc := newPaymentFixture(gw)
	if _, credited, err := svc.Confirm(context.Background(), "RC99999999999999deadbeef"); err == nil || credited {
		t.Fatalf("未知订单号应当报错且不入账，got credited=%v err=%v", credited, err)
	}
	if len(gw.queried) != 0 {
		t.Errorf("未知订单号不应发起查单，实际查了 %v", gw.queried)
	}
	if b := escrowBalanceOf(t, esc, "payer-1"); b != 0 {
		t.Errorf("余额应保持 0，实际 %d", b)
	}
}

// TestConfirmSurfacesQueryFailure 查单失败必须向上报错（handler 据此让微信重试），
// 而不是当作「未支付」静默吞掉——后者会让已收的钱永远不到账。
func TestConfirmSurfacesQueryFailure(t *testing.T) {
	gw := &stubGateway{prepayID: "p1", queryErr: errors.New("dial tcp: i/o timeout")}
	svc, _ := newPaymentFixture(gw)
	ctx := context.Background()
	order, _, err := svc.Prepay(ctx, "payer-1", "openid-1", 10000)
	if err != nil {
		t.Fatalf("Prepay: %v", err)
	}
	if _, credited, err := svc.Confirm(ctx, order.OutTradeNo); err == nil || credited {
		t.Fatalf("查单失败应当报错且不入账，got credited=%v err=%v", credited, err)
	}
}

// TestPaymentDisabledWhenNoGateway 未装配网关＝真实支付未开通，
// 所有入口都应返回 ErrPaymentDisabled（handler 据此回 503 而不是 500）。
func TestPaymentDisabledWhenNoGateway(t *testing.T) {
	svc, _ := newPaymentFixture(nil)
	ctx := context.Background()
	if svc.Enabled() {
		t.Fatal("网关为 nil 时 Enabled() 应为 false")
	}
	if _, _, err := svc.Prepay(ctx, "payer-1", "openid-1", 10000); !errors.Is(err, service.ErrPaymentDisabled) {
		t.Errorf("Prepay 应返回 ErrPaymentDisabled，实际 %v", err)
	}
	if _, _, err := svc.Confirm(ctx, "RC1"); !errors.Is(err, service.ErrPaymentDisabled) {
		t.Errorf("Confirm 应返回 ErrPaymentDisabled，实际 %v", err)
	}
	if _, err := svc.ListMine(ctx, "payer-1", 10); !errors.Is(err, service.ErrPaymentDisabled) {
		t.Errorf("ListMine 应返回 ErrPaymentDisabled，实际 %v", err)
	}
}

// TestNewOutTradeNoIsUniqueAndWellFormed 订单号是回调对账的唯一主键，
// 同一秒内也必须互不相同（8 位随机 hex），且长度落在微信要求的 6~32 内。
func TestNewOutTradeNoIsUniqueAndWellFormed(t *testing.T) {
	seen := map[string]bool{}
	now := time.Now()
	for i := 0; i < 200; i++ {
		no, err := service.NewOutTradeNo(now)
		if err != nil {
			t.Fatalf("NewOutTradeNo: %v", err)
		}
		if len(no) < 6 || len(no) > 32 {
			t.Fatalf("订单号长度 %d 超出微信要求 6~32：%q", len(no), no)
		}
		for _, c := range no {
			ok := (c >= '0' && c <= '9') || (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || c == '_' || c == '-' || c == '|' || c == '*' || c == '@'
			if !ok {
				t.Fatalf("订单号含微信不允许的字符 %q：%s", c, no)
			}
		}
		if seen[no] {
			t.Fatalf("同一秒内出现重复订单号：%s", no)
		}
		seen[no] = true
	}
}
