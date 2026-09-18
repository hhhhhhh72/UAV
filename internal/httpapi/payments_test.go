package httpapi_test

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"drone-platform/internal/domain"
	"drone-platform/internal/httpapi"
	"drone-platform/internal/repository"
	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

// 线上充值（微信支付）的 HTTP 层端到端用例。
//
// 这里刻意用**桩网关**而不是真实微信：本层要证的是「路由通了、鉴权对了、
// 状态码对了、钱按规则动了」，真实微信的协议细节由 internal/wechatpay 自己守。

const testPayerID = "payer-1"
const testPayerOpenID = "openid-payer-1"

// payStubGateway 桩网关（与 service 包用例里的同类，但这里要跨包，故各写一份）。
type payStubGateway struct {
	prepayID   string
	prepayErr  error
	lastOpenID string
	queried    []string
	result     service.PaymentResult
	queryErr   error
	notifyRef  string
	notifyErr  error
}

func (g *payStubGateway) Prepay(_ context.Context, in service.PrepayInput) (string, error) {
	g.lastOpenID = in.PayerOpenID
	if g.prepayErr != nil {
		return "", g.prepayErr
	}
	return g.prepayID, nil
}

func (g *payStubGateway) Query(_ context.Context, outTradeNo string) (service.PaymentResult, error) {
	g.queried = append(g.queried, outTradeNo)
	if g.queryErr != nil {
		return service.PaymentResult{}, g.queryErr
	}
	return g.result, nil
}

func (g *payStubGateway) PayParams(prepayID string) (map[string]string, error) {
	return map[string]string{
		"timeStamp": "1700000000", "nonceStr": "nonce-1",
		"package": "prepay_id=" + prepayID, "signType": "RSA", "paySign": "sig-1",
	}, nil
}

func (g *payStubGateway) DecodeNotify([]byte) (string, error) {
	if g.notifyErr != nil {
		return "", g.notifyErr
	}
	return g.notifyRef, nil
}

// newPayServer 装一台带线上充值的内存后端：
// 建一个**有微信 openid** 的付款人（seedCommonUsers 建的用户都没有 openid，
// 而小程序支付必须知道谁付钱），并注入支付服务。
func newPayServer(t *testing.T, gw *payStubGateway) (http.Handler, *service.EscrowService, repository.PaymentOrderRepository) {
	t.Helper()
	var esc *service.EscrowService
	var orders repository.PaymentOrderRepository
	app := newBizServerWith(t, func(srv *httpapi.Server, e *service.EscrowService, users repository.UserRepository) {
		if _, err := users.Create(context.Background(), domain.User{
			ID: testPayerID, Name: testPayerID, Role: domain.RoleIndividual,
			Status: "active", Version: 1, WechatOpenID: testPayerOpenID,
		}); err != nil {
			t.Fatalf("seed 付款人: %v", err)
		}
		orders = memory.NewPaymentOrderRepository()
		esc = e
		srv.SetPaymentService(service.NewPaymentService(orders, e, gw))
	})
	return app, esc, orders
}

func postJSON(t *testing.T, app http.Handler, path, body, userID string) *httpResult {
	t.Helper()
	w := requestAs(t, app, http.MethodPost, path, []byte(body), userID, domain.RoleIndividual)
	return &httpResult{code: w.Code, body: w.Body.String()}
}

type httpResult struct {
	code int
	body string
}

// TestWeChatPayPrepayReturnsPayParams 下单成功：201 + 商户订单号 + wx.requestPayment 参数。
func TestWeChatPayPrepayReturnsPayParams(t *testing.T) {
	gw := &payStubGateway{prepayID: "prepay-xyz"}
	app, _, _ := newPayServer(t, gw)

	res := postJSON(t, app, "/api/v1/payments/wechat/prepay", `{"amount_fen":10000}`, testPayerID)
	if res.code != http.StatusCreated {
		t.Fatalf("prepay 应返回 201，实际 %d %s", res.code, res.body)
	}
	var out struct {
		Data struct {
			OutTradeNo string            `json:"out_trade_no"`
			AmountFen  int64             `json:"amount_fen"`
			PayParams  map[string]string `json:"pay_params"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(res.body), &out); err != nil {
		t.Fatalf("解析响应: %v (%s)", err, res.body)
	}
	if !strings.HasPrefix(out.Data.OutTradeNo, "RC") {
		t.Errorf("商户订单号格式不对：%q", out.Data.OutTradeNo)
	}
	if out.Data.AmountFen != 10000 {
		t.Errorf("金额应为 10000，实际 %d", out.Data.AmountFen)
	}
	if out.Data.PayParams["package"] != "prepay_id=prepay-xyz" {
		t.Errorf("pay_params 不对：%v", out.Data.PayParams)
	}
	// 前端调起支付所需的五个字段一个都不能少。
	for _, k := range []string{"timeStamp", "nonceStr", "package", "signType", "paySign"} {
		if out.Data.PayParams[k] == "" {
			t.Errorf("pay_params 缺少 %s：%v", k, out.Data.PayParams)
		}
	}
}

// TestWeChatPayPrepayTakesOpenIDFromDB 安全回归：payer.openid 必须取自库里的当前用户，
// **不接受前端传入**——否则调用方可以指定「谁付钱」，把别人的 openid 塞进来。
func TestWeChatPayPrepayTakesOpenIDFromDB(t *testing.T) {
	gw := &payStubGateway{prepayID: "prepay-xyz"}
	app, _, _ := newPayServer(t, gw)

	res := postJSON(t, app, "/api/v1/payments/wechat/prepay",
		`{"amount_fen":10000,"openid":"openid-of-someone-else","payer":{"openid":"openid-of-someone-else"}}`, testPayerID)
	if res.code != http.StatusCreated {
		t.Fatalf("prepay 应返回 201，实际 %d %s", res.code, res.body)
	}
	if gw.lastOpenID != testPayerOpenID {
		t.Fatalf("网关收到的 openid 应来自数据库（%q），实际 %q——前端传入的 openid 被采信了", testPayerOpenID, gw.lastOpenID)
	}
}

// TestWeChatPayPrepayRejectsBadAmount 金额越界回 400（可读错误，不是 500）。
func TestWeChatPayPrepayRejectsBadAmount(t *testing.T) {
	gw := &payStubGateway{prepayID: "prepay-xyz"}
	app, _, _ := newPayServer(t, gw)
	for _, body := range []string{`{"amount_fen":1}`, `{"amount_fen":0}`, `{"amount_fen":-500}`, `{"amount_fen":99999999}`} {
		if res := postJSON(t, app, "/api/v1/payments/wechat/prepay", body, testPayerID); res.code != http.StatusBadRequest {
			t.Errorf("%s 应回 400，实际 %d %s", body, res.code, res.body)
		}
	}
}

// TestWeChatPayPrepayRequiresAuth 下单是写操作，匿名必须 401。
func TestWeChatPayPrepayRequiresAuth(t *testing.T) {
	gw := &payStubGateway{prepayID: "prepay-xyz"}
	app, _, _ := newPayServer(t, gw)
	w := doRaw(app, http.MethodPost, "/api/v1/payments/wechat/prepay", `{"amount_fen":10000}`, "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("匿名 prepay 应回 401，实际 %d %s", w.Code, w.Body.String())
	}
}

// TestWeChatPayDisabledIsNotServerError 未开通微信支付时**不得回 5xx**。
//
// 回归的是一个真实踩过的坑：先前 prepay 与 payments/mine 都回 503，而 fail() 对
// status>=500 会把 message 统一脱敏成 "internal server error" 并记一条 ERROR 日志——
// 于是「没开通」这个完全预期的状态，在客户端读起来像服务故障，在日志/监控里像真故障。
// 生产实测印证过：GET /api/v1/payments/mine 返回 503 + {"code":"INTERNAL"}。
//
// 现在的约定：
//   - POST prepay  → 409（4xx，消息原样透传）
//   - GET  mine    → 200 + enabled:false（探测语义，不该是错误）
func TestWeChatPayDisabledIsNotServerError(t *testing.T) {
	// 用 newBizServer（没有 SetPaymentService）。注意这里的用户必须是 seedCommonUsers
	// 里已入库的 user-1——用未入库的 ID 会先被 authenticate 挡成 401，测不到本条规则。
	app := newBizServer(t)

	w := requestAs(t, app, http.MethodPost, "/api/v1/payments/wechat/prepay",
		[]byte(`{"amount_fen":10000}`), "user-1", domain.RoleIndividual)
	if w.Code >= 500 {
		t.Fatalf("未开通时 prepay 不得回 5xx（会被脱敏 + 记 ERROR），实际 %d %s", w.Code, w.Body.String())
	}
	if w.Code != http.StatusConflict {
		t.Fatalf("未开通时 prepay 应回 409，实际 %d %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "微信支付未开通") {
		t.Fatalf("409 的消息应当原样透传，实际 %s", w.Body.String())
	}

	mw := requestAs(t, app, http.MethodGet, "/api/v1/payments/mine", nil, "user-1", domain.RoleIndividual)
	if mw.Code != http.StatusOK {
		t.Fatalf("未开通时 payments/mine 应回 200（探测语义），实际 %d %s", mw.Code, mw.Body.String())
	}
	var probe struct {
		Data struct {
			Enabled bool  `json:"enabled"`
			Items   []any `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(mw.Body.Bytes(), &probe); err != nil {
		t.Fatalf("解析探测响应: %v (%s)", err, mw.Body.String())
	}
	if probe.Data.Enabled {
		t.Fatalf("未开通时 enabled 应为 false，实际 %s", mw.Body.String())
	}
}

// TestPrepayRateLimited 充值下单限频：每成功一次都会在微信侧真实建一个支付单，
// 不设限时一个账号能在几秒内造出成千上万个待支付订单（还会被微信风控标记商户号）。
func TestPrepayRateLimited(t *testing.T) {
	gw := &payStubGateway{prepayID: "prepay-xyz"}
	app, _, _ := newPayServer(t, gw)

	var last int
	for i := 0; i < prepayMaxForTest+1; i++ {
		last = postJSON(t, app, "/api/v1/payments/wechat/prepay", `{"amount_fen":10000}`, testPayerID).code
	}
	if last != http.StatusTooManyRequests {
		t.Fatalf("第 %d 次下单应被限频回 429，实际 %d", prepayMaxForTest+1, last)
	}
}

// prepayMaxForTest 与 internal/httpapi 里的 prepayMax 对齐（外部测试包读不到常量，
// 改限频上限时这里要同步——不一致会让用例失败而不是静默放过）。
const prepayMaxForTest = 10

// TestSimulatedDepositClosedOncePaymentEnabled 真实支付开通后，普通用户的
// 「模拟托管通道自助充值」必须自动关闭；否则真实支付旁边就摆着一个免费充值按钮
// （单笔 ¥20 万、不限次数，且无外部单号时完全不做查重）。
func TestSimulatedDepositClosedOncePaymentEnabled(t *testing.T) {
	app, esc, _ := newPayServer(t, &payStubGateway{prepayID: "p"})

	w := requestAs(t, app, http.MethodPost, "/api/v1/escrow/deposit",
		[]byte(`{"amount_fen":100}`), testPayerID, domain.RoleIndividual)
	if w.Code != http.StatusForbidden {
		t.Fatalf("真实支付已开通时普通用户自助充值应 403，实际 %d %s", w.Code, w.Body.String())
	}
	if acc, err := esc.Balance(context.Background(), testPayerID); err == nil && acc.BalanceFen != 0 {
		t.Fatalf("被拒绝的充值不得改动余额，实际 %d", acc.BalanceFen)
	}

	// 管理员代充（线下来款补记）不受影响。
	aw := requestAs(t, app, http.MethodPost, "/api/v1/escrow/deposit",
		[]byte(`{"amount_fen":100,"to_user":"`+testPayerID+`"}`), "admin-1", domain.RolePlatformAdmin)
	if aw.Code != http.StatusCreated {
		t.Fatalf("管理员代充应仍然可用，实际 %d %s", aw.Code, aw.Body.String())
	}
}

// TestPaymentMineReportsEnabledWhenConfigured 已开通时探测点回 enabled:true。
func TestPaymentMineReportsEnabledWhenConfigured(t *testing.T) {
	app, _, _ := newPayServer(t, &payStubGateway{prepayID: "p"})
	w := requestAs(t, app, http.MethodGet, "/api/v1/payments/mine", nil, testPayerID, domain.RoleIndividual)
	if w.Code != http.StatusOK {
		t.Fatalf("payments/mine 应回 200，实际 %d %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), `"enabled":true`) {
		t.Fatalf("已开通时 enabled 应为 true，实际 %s", w.Body.String())
	}
}

// TestWeChatPayNotifyIsPublicAndCredits 回调必须**匿名可达**（微信不带我们的令牌），
// 且应答体是微信认的 {"code":"SUCCESS"}；钱要真的进托管账户。
func TestWeChatPayNotifyIsPublicAndCredits(t *testing.T) {
	gw := &payStubGateway{prepayID: "prepay-xyz"}
	app, esc, _ := newPayServer(t, gw)

	res := postJSON(t, app, "/api/v1/payments/wechat/prepay", `{"amount_fen":30000}`, testPayerID)
	if res.code != http.StatusCreated {
		t.Fatalf("prepay: %d %s", res.code, res.body)
	}
	var out struct {
		Data struct {
			OutTradeNo string `json:"out_trade_no"`
		} `json:"data"`
	}
	_ = json.Unmarshal([]byte(res.body), &out)

	gw.notifyRef = out.Data.OutTradeNo
	gw.result = service.PaymentResult{TradeState: "SUCCESS", TransactionID: "wx-tx-777", AmountFen: 30000}

	// 匿名（无 Authorization 头）打回调。
	w := doRaw(app, http.MethodPost, "/api/v1/payments/wechat/notify", `{"id":"evt-1"}`, "")
	if w.Code != http.StatusOK {
		t.Fatalf("回调应回 200，实际 %d %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "SUCCESS") {
		t.Fatalf("回调应答必须是微信认的成功格式，实际 %s", w.Body.String())
	}
	acc, err := esc.Balance(context.Background(), testPayerID)
	if err != nil {
		t.Fatalf("查托管余额: %v", err)
	}
	if acc.BalanceFen != 30000 {
		t.Fatalf("回调后托管余额应为 30000，实际 %d", acc.BalanceFen)
	}

	// 微信会重试同一个回调——不得重复加钱。
	w2 := doRaw(app, http.MethodPost, "/api/v1/payments/wechat/notify", `{"id":"evt-1"}`, "")
	if w2.Code != http.StatusOK {
		t.Fatalf("重放回调应回 200，实际 %d %s", w2.Code, w2.Body.String())
	}
	acc2, _ := esc.Balance(context.Background(), testPayerID)
	if acc2.BalanceFen != 30000 {
		t.Fatalf("重放后余额仍应为 30000，实际 %d（重复入账）", acc2.BalanceFen)
	}
}

// TestWeChatPayNotifyRejectsUndecryptableBody 解不开的报文回 400（可能是伪造，也可能是密钥配错）。
func TestWeChatPayNotifyRejectsUndecryptableBody(t *testing.T) {
	gw := &payStubGateway{prepayID: "p", notifyErr: context.DeadlineExceeded}
	app, esc, _ := newPayServer(t, gw)
	w := doRaw(app, http.MethodPost, "/api/v1/payments/wechat/notify", `{"id":"evt-x"}`, "")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("解不开的回调应回 400，实际 %d %s", w.Code, w.Body.String())
	}
	if acc, _ := esc.Balance(context.Background(), testPayerID); acc.BalanceFen != 0 {
		t.Fatalf("解不开的回调不得入账，余额 %d", acc.BalanceFen)
	}
}

// TestWeChatPayNotifyReturns500OnQueryFailure 查单失败必须回非 2xx：让微信重试。
// 若在这里回 200「成功」，这笔已收的钱就永远不到账了。
func TestWeChatPayNotifyReturns500OnQueryFailure(t *testing.T) {
	gw := &payStubGateway{prepayID: "prepay-xyz"}
	app, esc, _ := newPayServer(t, gw)
	res := postJSON(t, app, "/api/v1/payments/wechat/prepay", `{"amount_fen":10000}`, testPayerID)
	var out struct {
		Data struct {
			OutTradeNo string `json:"out_trade_no"`
		} `json:"data"`
	}
	_ = json.Unmarshal([]byte(res.body), &out)
	gw.notifyRef = out.Data.OutTradeNo
	gw.queryErr = context.DeadlineExceeded

	w := doRaw(app, http.MethodPost, "/api/v1/payments/wechat/notify", `{"id":"evt-y"}`, "")
	if w.Code < 500 {
		t.Fatalf("查单失败应回 5xx 让微信重试，实际 %d %s", w.Code, w.Body.String())
	}
	if acc, _ := esc.Balance(context.Background(), testPayerID); acc.BalanceFen != 0 {
		t.Fatalf("查单失败不得入账，余额 %d", acc.BalanceFen)
	}
}

// TestWeChatPayNotifyPathIsExactMatch 公开路径必须是**精确匹配**：
// 前缀放行会让 /api/v1/payments/wechat/notify/任意后缀 也变成匿名可达。
func TestWeChatPayNotifyPathIsExactMatch(t *testing.T) {
	gw := &payStubGateway{prepayID: "p"}
	app, _, _ := newPayServer(t, gw)
	w := doRaw(app, http.MethodPost, "/api/v1/payments/wechat/notify/extra", `{}`, "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("notify 的任意子路径不得匿名可达，实际 %d %s", w.Code, w.Body.String())
	}
}

// TestPaymentMineListsMyOrders 我的充值记录只列自己的单。
func TestPaymentMineListsMyOrders(t *testing.T) {
	gw := &payStubGateway{prepayID: "prepay-xyz"}
	app, _, _ := newPayServer(t, gw)
	if res := postJSON(t, app, "/api/v1/payments/wechat/prepay", `{"amount_fen":10000}`, testPayerID); res.code != http.StatusCreated {
		t.Fatalf("prepay: %d %s", res.code, res.body)
	}
	w := requestAs(t, app, http.MethodGet, "/api/v1/payments/mine", nil, testPayerID, domain.RoleIndividual)
	if w.Code != http.StatusOK {
		t.Fatalf("payments/mine 应回 200，实际 %d %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "\"amount_fen\":10000") {
		t.Fatalf("记录里应含刚下的单，实际 %s", w.Body.String())
	}
	// 别人的记录里不该出现这笔。
	other := requestAs(t, app, http.MethodGet, "/api/v1/payments/mine", nil, "user-1", domain.RoleIndividual)
	if strings.Contains(other.Body.String(), "\"amount_fen\":10000") {
		t.Fatalf("他人不应看到这笔充值，实际 %s", other.Body.String())
	}
}
