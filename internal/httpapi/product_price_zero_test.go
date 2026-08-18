package httpapi_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"drone-platform/internal/domain"
)

// 商品价格更新回归：price_fen 传 0（改为面议）必须生效——
// 此前 `if in.PriceFen > 0` 导致价格无法改 0，管理后台静默失败。
func TestAdminUpdateProductPriceToZero(t *testing.T) {
	app := newBizServer(t)

	// 卖家发布商品 → 管理员上架
	pw := requestAs(t, app, http.MethodPost, "/api/v1/products",
		[]byte(`{"title":"改价商品","prod_type":"drone","price_fen":500000}`),
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
	appr := requestAs(t, app, http.MethodPut, "/api/v1/admin/products/"+pid,
		[]byte(`{"status":"listed"}`), "admin-1", domain.RolePlatformAdmin)
	if appr.Code != http.StatusOK {
		t.Fatalf("approve product: %d %s", appr.Code, appr.Body.String())
	}

	// 价格改为 0（面议）→ 生效
	up := requestAs(t, app, http.MethodPut, "/api/v1/admin/products/"+pid,
		[]byte(`{"price_fen":0}`), "admin-1", domain.RolePlatformAdmin)
	if up.Code != http.StatusOK {
		t.Fatalf("update price: %d %s", up.Code, up.Body.String())
	}
	if !jsonContains(up.Body.String(), `"price_fen":0`) {
		t.Fatalf("price should be updated to 0, got: %s", up.Body.String())
	}

	// 详情确认
	g := requestAs(t, app, http.MethodGet, "/api/v1/admin/products/"+pid, nil, "admin-1", domain.RolePlatformAdmin)
	if !jsonContains(g.Body.String(), `"price_fen":0`) {
		t.Fatalf("stored price should be 0, got: %s", g.Body.String())
	}
}

func jsonContains(body, needle string) bool {
	return strings.Contains(body, needle)
}
