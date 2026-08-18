package httpapi_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"drone-platform/internal/domain"
)

// 订单定价回归（P0）：下单金额与卖家必须来自服务端商品，忽略客户端传入值。
// 修复前 amount_fen/seller_id 客户端自报——1 分钱下单 / 伪造卖家。
func TestTradeOrderServerSidePricing(t *testing.T) {
	app := newBizServer(t)

	// 卖家发布商品（pending）→ 管理员审核上架（listed）
	pw := requestAs(t, app, http.MethodPost, "/api/v1/products",
		[]byte(`{"title":"定价回归无人机","prod_type":"drone","price_fen":888888}`),
		"seller-1", domain.RoleEnterprise)
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
	pid := created.Data.ID

	// 1) 未上架（pending）不可下单 → 409
	w := requestAs(t, app, http.MethodPost, "/api/v1/trade-orders",
		[]byte(`{"product_id":"`+pid+`","amount_fen":1}`), "buyer-1", domain.RoleIndividual)
	if w.Code != http.StatusConflict {
		t.Fatalf("order on pending product: want 409, got %d %s", w.Code, w.Body.String())
	}

	// 管理员上架
	appr := requestAs(t, app, http.MethodPut, "/api/v1/admin/products/"+pid,
		[]byte(`{"status":"listed"}`), "admin-1", domain.RolePlatformAdmin)
	if appr.Code != http.StatusOK {
		t.Fatalf("approve product: %d %s", appr.Code, appr.Body.String())
	}

	// 2) 客户端传 1 分钱 + 假卖家 → 订单金额/卖家为服务端商品值
	w = requestAs(t, app, http.MethodPost, "/api/v1/trade-orders",
		[]byte(`{"product_id":"`+pid+`","seller_id":"hacker-x","amount_fen":1}`),
		"buyer-1", domain.RoleIndividual)
	if w.Code != http.StatusCreated {
		t.Fatalf("create order: %d %s", w.Code, w.Body.String())
	}
	if !bytes.Contains(w.Body.Bytes(), []byte(`"amount_fen":888888`)) {
		t.Fatalf("order must use product price (888888), got: %s", w.Body.String())
	}
	if !bytes.Contains(w.Body.Bytes(), []byte(`"seller_id":"seller-1"`)) {
		t.Fatalf("order must use product seller, got: %s", w.Body.String())
	}

	// 3) 不存在的商品 → 404
	w = requestAs(t, app, http.MethodPost, "/api/v1/trade-orders",
		[]byte(`{"product_id":"product-nope"}`), "buyer-1", domain.RoleIndividual)
	if w.Code != http.StatusNotFound {
		t.Fatalf("order on missing product: want 404, got %d %s", w.Code, w.Body.String())
	}

	// 4) 下架商品（removed）→ 409
	rm := requestAs(t, app, http.MethodPut, "/api/v1/admin/products/"+pid,
		[]byte(`{"status":"removed"}`), "admin-1", domain.RolePlatformAdmin)
	if rm.Code != http.StatusOK {
		t.Fatalf("remove product: %d %s", rm.Code, rm.Body.String())
	}
	w = requestAs(t, app, http.MethodPost, "/api/v1/trade-orders",
		[]byte(`{"product_id":"`+pid+`"}`), "buyer-2", domain.RoleIndividual)
	if w.Code != http.StatusConflict {
		t.Fatalf("order on removed product: want 409, got %d %s", w.Code, w.Body.String())
	}
}
