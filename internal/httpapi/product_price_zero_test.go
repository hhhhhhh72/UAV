package httpapi_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"drone-platform/internal/domain"
)

// 商品价格更新回归。两个方向都要钉住：
//
//  1. 显式声明面议后，price_fen 必须能改成 0——此前 `if in.PriceFen > 0` 让价格改不了 0，
//     管理后台静默失败。
//  2. 但"标价 0 元"必须被拒。修复前 price_fen=0 一个值同时表示"卖家选了面议"和
//     "卖家填了 0 元"，后端无法区分；现在 0 只能配合 price_mode=negotiable 出现。
func TestAdminUpdateProductPriceToZero(t *testing.T) {
	app := newBizServer(t)

	// 卖家发布（明码标价 5000 元）→ 协会审核通过上架
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
	appr := requestAs(t, app, http.MethodPost, "/api/v1/admin/products/"+pid+"/review",
		[]byte(`{"check_status":"passed"}`), "admin-1", domain.RolePlatformAdmin)
	if appr.Code != http.StatusOK {
		t.Fatalf("approve product: %d %s", appr.Code, appr.Body.String())
	}

	// (1) 只把价格改成 0、不声明面议 → 拒绝（商品当前是明码标价，0 不是合法标价）
	bad := requestAs(t, app, http.MethodPut, "/api/v1/admin/products/"+pid,
		[]byte(`{"price_fen":0}`), "admin-1", domain.RolePlatformAdmin)
	if bad.Code != http.StatusBadRequest {
		t.Fatalf("price_fen=0 且未声明面议应返回 400，实际 %d %s", bad.Code, bad.Body.String())
	}
	// 被拒的更新不得改动商品
	still := requestAs(t, app, http.MethodGet, "/api/v1/admin/products/"+pid, nil, "admin-1", domain.RolePlatformAdmin)
	if !jsonContains(still.Body.String(), `"price_fen":500000`) {
		t.Fatalf("被拒的更新不应改动价格，got: %s", still.Body.String())
	}

	// (2) 显式声明面议 + 价格 0 → 生效
	up := requestAs(t, app, http.MethodPut, "/api/v1/admin/products/"+pid,
		[]byte(`{"price_fen":0,"price_mode":"negotiable"}`), "admin-1", domain.RolePlatformAdmin)
	if up.Code != http.StatusOK {
		t.Fatalf("update price: %d %s", up.Code, up.Body.String())
	}
	if !jsonContains(up.Body.String(), `"price_fen":0`) {
		t.Fatalf("price should be updated to 0, got: %s", up.Body.String())
	}
	if !jsonContains(up.Body.String(), `"price_mode":"negotiable"`) {
		t.Fatalf("price_mode should be negotiable, got: %s", up.Body.String())
	}

	// 详情确认已落库
	g := requestAs(t, app, http.MethodGet, "/api/v1/admin/products/"+pid, nil, "admin-1", domain.RolePlatformAdmin)
	if !jsonContains(g.Body.String(), `"price_fen":0`) || !jsonContains(g.Body.String(), `"price_mode":"negotiable"`) {
		t.Fatalf("stored price/mode wrong, got: %s", g.Body.String())
	}

	// (3) 面议商品再改回明码标价：必须给出大于 0 的价格
	back := requestAs(t, app, http.MethodPut, "/api/v1/admin/products/"+pid,
		[]byte(`{"price_mode":"fixed","price_fen":0}`), "admin-1", domain.RolePlatformAdmin)
	if back.Code != http.StatusBadRequest {
		t.Fatalf("明码标价填 0 应返回 400，实际 %d %s", back.Code, back.Body.String())
	}
	ok := requestAs(t, app, http.MethodPut, "/api/v1/admin/products/"+pid,
		[]byte(`{"price_mode":"fixed","price_fen":600000}`), "admin-1", domain.RolePlatformAdmin)
	if ok.Code != http.StatusOK {
		t.Fatalf("改回明码标价应成功: %d %s", ok.Code, ok.Body.String())
	}
}

func jsonContains(body, needle string) bool {
	return strings.Contains(body, needle)
}
