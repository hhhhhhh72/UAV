package wechatpay

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// newTestClient 起一个假微信服务器，返回客户端与「收到的请求」记录器。
func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c, err := New(Config{
		AppID: "wx-test", MchID: "1900000109",
		APIv3Key: "0123456789012345678901234567890a",
		CertSerial: "ABCDEF1234567890", PrivateKeyPath: writeTestKey(t),
		NotifyURL: "https://example.com/api/v1/payments/wechat/notify",
		APIBase:   srv.URL,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return c
}

// TestCreateRefundSendsExpectedRequest 退款请求必须带原单号、退款单号、金额，
// 且回调地址由支付回调地址推导出来（只配一个域名就能接上两条链路）。
func TestCreateRefundSendsExpectedRequest(t *testing.T) {
	var gotPath, gotBody, gotAuth string
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"refund_id": "wx-refund-1", "out_refund_no": "RF1", "out_trade_no": "RC1",
			"status": "PROCESSING",
			"amount":  map[string]any{"total": 10000, "refund": 3000},
		})
	})

	ref, err := c.CreateRefund(context.Background(), RefundRequest{
		OutTradeNo: "RC1", OutRefundNo: "RF1", RefundFen: 3000, TotalFen: 10000, Reason: "订单取消",
	})
	if err != nil {
		t.Fatalf("CreateRefund: %v", err)
	}
	if gotPath != "/v3/refund/domestic/refunds" {
		t.Errorf("路径不对：%s", gotPath)
	}
	if !strings.HasPrefix(gotAuth, "WECHATPAY2-SHA256-RSA2048 ") {
		t.Errorf("Authorization 头格式不对：%q", gotAuth)
	}
	for _, want := range []string{"\"out_trade_no\":\"RC1\"", "\"out_refund_no\":\"RF1\"", "\"refund\":3000", "\"total\":10000", "\"currency\":\"CNY\""} {
		if !strings.Contains(gotBody, want) {
			t.Errorf("请求体缺少 %s：%s", want, gotBody)
		}
	}
	if !strings.Contains(gotBody, "refund-notify") {
		t.Errorf("退款回调地址应由支付回调推导：%s", gotBody)
	}
	if ref.RefundID != "wx-refund-1" || ref.Status != RefundStatusProcessing {
		t.Errorf("响应解析不对：%+v", ref)
	}
	if ref.Amount.Refund != 3000 || ref.Amount.Total != 10000 {
		t.Errorf("金额解析不对：%+v", ref.Amount)
	}
}

// TestCreateRefundRejectsOverTotal 退款额超过原单额必须在**发请求之前**就拒掉，
// 不能指望微信替我们校验。
func TestCreateRefundRejectsOverTotal(t *testing.T) {
	called := false
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) { called = true })
	if _, err := c.CreateRefund(context.Background(), RefundRequest{
		OutTradeNo: "RC1", OutRefundNo: "RF1", RefundFen: 10001, TotalFen: 10000,
	}); err == nil {
		t.Fatal("超额退款应当报错")
	}
	if called {
		t.Fatal("超额退款不该发出请求")
	}
	if _, err := c.CreateRefund(context.Background(), RefundRequest{OutTradeNo: "RC1", OutRefundNo: "RF1", RefundFen: 0, TotalFen: 10000}); err == nil {
		t.Fatal("零额退款应当报错")
	}
	if _, err := c.CreateRefund(context.Background(), RefundRequest{OutRefundNo: "RF1", RefundFen: 100, TotalFen: 100}); err == nil {
		t.Fatal("缺原单号应当报错")
	}
}

// TestQueryRefund 查退款单。
func TestQueryRefund(t *testing.T) {
	var gotPath string
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_ = json.NewEncoder(w).Encode(map[string]any{
			"refund_id": "wx-refund-9", "out_refund_no": "RF9", "status": "SUCCESS",
			"amount": map[string]any{"total": 5000, "refund": 5000},
		})
	})
	ref, err := c.QueryRefund(context.Background(), "RF9")
	if err != nil {
		t.Fatalf("QueryRefund: %v", err)
	}
	if gotPath != "/v3/refund/domestic/refunds/RF9" {
		t.Errorf("路径不对：%s", gotPath)
	}
	if ref.Status != RefundStatusSuccess || ref.RefundID != "wx-refund-9" {
		t.Errorf("解析不对：%+v", ref)
	}
}

// TestEffectiveRefundNotifyURL 回调地址推导：显式配置优先，否则由支付回调推导。
func TestEffectiveRefundNotifyURL(t *testing.T) {
	cases := []struct{ notify, refund, want string }{
		{"https://a.cn/api/v1/payments/wechat/notify", "", "https://a.cn/api/v1/payments/wechat/refund-notify"},
		{"https://a.cn/api/v1/payments/wechat/notify", "https://b.cn/rn", "https://b.cn/rn"},
		{"https://a.cn/whatever", "", "https://a.cn/whatever"},
		{"", "", ""},
	}
	for _, c := range cases {
		got := Config{NotifyURL: c.notify, RefundNotifyURL: c.refund}.effectiveRefundNotifyURL()
		if got != c.want {
			t.Errorf("notify=%q refund=%q: want %q got %q", c.notify, c.refund, c.want, got)
		}
	}
}