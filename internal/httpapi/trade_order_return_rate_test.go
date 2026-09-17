package httpapi_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"drone-platform/internal/domain"
)

// seedListedProduct 造一个"在售"商品（卖家发布 → 管理员上架），返回 productID。
func seedListedProduct(t *testing.T, app http.Handler, seller, title string, priceFen int64) string {
	t.Helper()
	pw := requestAs(t, app, http.MethodPost, "/api/v1/products",
		[]byte(fmt.Sprintf("{\"title\":%q,\"prod_type\":\"drone\",\"price_fen\":%d}", title, priceFen)),
		seller, domain.RoleEnterprise)
	if pw.Code != http.StatusCreated {
		t.Fatalf("create product: %d %s", pw.Code, pw.Body.String())
	}
	var created struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(pw.Body.Bytes(), &created); err != nil {
		t.Fatalf("parse product: %v", err)
	}
	appr := requestAs(t, app, http.MethodPut, "/api/v1/admin/products/"+created.Data.ID,
		[]byte(`{"status":"listed"}`), "admin-1", domain.RolePlatformAdmin)
	if appr.Code != http.StatusOK {
		t.Fatalf("list product: %d %s", appr.Code, appr.Body.String())
	}
	return created.Data.ID
}

// orderField 从订单响应里取一个字符串字段（响应被 respond 包在 data 里）。
func orderField(t *testing.T, body []byte, field string) string {
	t.Helper()
	var parsed struct {
		Data map[string]any `json:"data"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		t.Fatalf("parse order response: %v", err)
	}
	s, _ := parsed.Data[field].(string)
	return s
}

// 下单限频（20 分钟 10 单）：第 11 单必须 429。
// 下单即把商品置 sold，无频次约束时可用批量下单锁死他人商品（拒绝供给）。
func TestTradeOrderCreateRateLimited(t *testing.T) {
	app := newBizServer(t)
	const attempts = 11
	products := make([]string, 0, attempts)
	for i := 0; i < attempts; i++ {
		products = append(products, seedListedProduct(t, app, "seller-1", fmt.Sprintf("限频商品%d", i), 100))
	}

	created, lastCode := 0, 0
	for i, pid := range products {
		w := requestAs(t, app, http.MethodPost, "/api/v1/trade-orders",
			orderBody(pid), "buyer-4", domain.RoleIndividual)
		if w.Code == http.StatusCreated {
			created++
			continue
		}
		lastCode = w.Code
		if i != attempts-1 {
			t.Fatalf("第 %d 单不该被拒：%d %s", i+1, w.Code, w.Body.String())
		}
	}
	if created != 10 {
		t.Fatalf("20 分钟窗口内应允许 10 单，实际 %d", created)
	}
	if lastCode != http.StatusTooManyRequests {
		t.Fatalf("第 %d 单应为 429，实际 %d", attempts, lastCode)
	}
	// 另一个买家不受影响（限频按人，不是全局/按 IP）
	pid := seedListedProduct(t, app, "seller-1", "限频商品-他人", 100)
	if w := requestAs(t, app, http.MethodPost, "/api/v1/trade-orders",
		orderBody(pid), "user-5", domain.RoleIndividual); w.Code != http.StatusCreated {
		t.Fatalf("其他买家不应被限频影响：%d %s", w.Code, w.Body.String())
	}
}

// 退货退款 HTTP 全链路：申请退货 → 卖家同意（不退款）→ 买家寄回 → 卖家确认收货（才退款）。
func TestTradeOrderReturnFlowEndpoints(t *testing.T) {
	app := newBizServer(t)
	pid := seedListedProduct(t, app, "seller-1", "退货流程商品", 10000)

	// 买家充值，保证付款能冻结成功
	if w := requestAs(t, app, http.MethodPost, "/api/v1/escrow/deposit",
		[]byte(`{"amount_fen":20000}`), "buyer-2", domain.RoleIndividual); w.Code != http.StatusCreated {
		t.Fatalf("充值: %d %s", w.Code, w.Body.String())
	}

	ow := requestAs(t, app, http.MethodPost, "/api/v1/trade-orders",
		orderBody(pid), "buyer-2", domain.RoleIndividual)
	if ow.Code != http.StatusCreated {
		t.Fatalf("下单: %d %s", ow.Code, ow.Body.String())
	}
	orderID := orderField(t, ow.Body.Bytes(), "id")

	if w := requestAs(t, app, http.MethodPost, "/api/v1/trade-orders/"+orderID+"/pay", nil, "buyer-2", domain.RoleIndividual); w.Code != http.StatusOK {
		t.Fatalf("付款: %d %s", w.Code, w.Body.String())
	}
	if w := requestAs(t, app, http.MethodPost, "/api/v1/trade-orders/"+orderID+"/ship",
		[]byte(`{"shipping_company":"顺丰速运","shipping_tracking":"SF9876543210"}`), "seller-1", domain.RoleEnterprise); w.Code != http.StatusOK {
		t.Fatalf("发货: %d %s", w.Code, w.Body.String())
	}

	// 买家申请"退货退款"
	aw := requestAs(t, app, http.MethodPost, "/api/v1/trade-orders/"+orderID+"/aftersale",
		[]byte(`{"aftersale_type":"return","aftersale_reason":"货不对板","aftersale_desc":"","aftersale_amount_fen":10000}`),
		"buyer-2", domain.RoleIndividual)
	if aw.Code != http.StatusOK {
		t.Fatalf("申请退货退款: %d %s", aw.Code, aw.Body.String())
	}

	// 卖家同意退货 → returning，且钱没动（余额未回买家）
	rv := requestAs(t, app, http.MethodPost, "/api/v1/trade-orders/"+orderID+"/aftersale/review",
		[]byte(`{"action":"approve"}`), "seller-1", domain.RoleEnterprise)
	if rv.Code != http.StatusOK {
		t.Fatalf("同意退货: %d %s", rv.Code, rv.Body.String())
	}
	if got := orderField(t, rv.Body.Bytes(), "aftersale_status"); got != "returning" {
		t.Fatalf("同意退货后应为 returning（待买家寄回），实际 %q", got)
	}

	// 未寄回就确认收货 → 必须失败
	if w := requestAs(t, app, http.MethodPost, "/api/v1/trade-orders/"+orderID+"/aftersale/confirm-return",
		nil, "seller-1", domain.RoleEnterprise); w.Code == http.StatusOK {
		t.Fatalf("买家未寄回，确认收到退货不应成功：%d %s", w.Code, w.Body.String())
	}

	// 买家提交退货物流
	sw := requestAs(t, app, http.MethodPost, "/api/v1/trade-orders/"+orderID+"/aftersale/return",
		[]byte(`{"tracking_no":"SF123456789","note":"已寄出"}`), "buyer-2", domain.RoleIndividual)
	if sw.Code != http.StatusOK {
		t.Fatalf("提交退货物流: %d %s", sw.Code, sw.Body.String())
	}
	if got := orderField(t, sw.Body.Bytes(), "aftersale_status"); got != "returned" {
		t.Fatalf("提交物流后应为 returned，实际 %q", got)
	}
	if got := orderField(t, sw.Body.Bytes(), "return_tracking"); got != "SF123456789" {
		t.Fatalf("退货单号未回显，实际 %q", got)
	}

	// 卖家确认收到 → 才退款并结案
	cw := requestAs(t, app, http.MethodPost, "/api/v1/trade-orders/"+orderID+"/aftersale/confirm-return",
		nil, "seller-1", domain.RoleEnterprise)
	if cw.Code != http.StatusOK {
		t.Fatalf("确认收到退货: %d %s", cw.Code, cw.Body.String())
	}
	if got := orderField(t, cw.Body.Bytes(), "aftersale_status"); got != "approved" {
		t.Fatalf("确认收货后应 approved，实际 %q", got)
	}
	if got := orderField(t, cw.Body.Bytes(), "status"); got != "completed" {
		t.Fatalf("确认收货后订单应 completed，实际 %q", got)
	}

	// 退款已回买家：20000 充值 - 10000 冻结 + 10000 退回 = 20000 可用
	mw := requestAs(t, app, http.MethodGet, "/api/v1/escrow/mine", nil, "buyer-2", domain.RoleIndividual)
	if mw.Code != http.StatusOK {
		t.Fatalf("查托管金: %d %s", mw.Code, mw.Body.String())
	}
	var acc struct {
		Data struct {
			Account struct {
				BalanceFen int64 `json:"balance_fen"`
			} `json:"account"`
		} `json:"data"`
	}
	if err := json.Unmarshal(mw.Body.Bytes(), &acc); err != nil {
		t.Fatalf("解析托管金: %v", err)
	}
	if acc.Data.Account.BalanceFen != 20000 {
		t.Fatalf("退款应回到买家余额（20000），实际 %d", acc.Data.Account.BalanceFen)
	}
}
