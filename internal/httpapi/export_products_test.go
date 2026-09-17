package httpapi_test

import (
	"net/http"
	"strings"
	"testing"

	"drone-platform/internal/domain"
)

// 商品导出接入回归。
//
// 补齐前 export_handler.go 的 switch 里没有 products → 404 → 前端静默回退"导出当前页"，
// 最多只能拿到 100 行，也拿不到全量导出的平台管理员门槛与限频保护。
func TestExportProducts(t *testing.T) {
	app := newBizServer(t)

	// 建两个商品：一个明码标价、一个面议，验证价格列与审核/上架两列的呈现
	for _, body := range []string{
		`{"title":"导出用整机","prod_type":"drone","condition":"new","price_mode":"fixed","price_fen":8800000,"status":"listed","delivery":"logistics"}`,
		`{"title":"导出用面议件","prod_type":"part","condition":"used","price_mode":"negotiable","price_fen":0,"status":"listed"}`,
	} {
		if w := requestAs(t, app, http.MethodPost, "/api/v1/admin/products", []byte(body), "admin-1", domain.RolePlatformAdmin); w.Code != http.StatusCreated {
			t.Fatalf("建商品: %d %s", w.Code, w.Body.String())
		}
	}

	w := requestAs(t, app, http.MethodGet, "/api/v1/admin/export/products", nil, "admin-1", domain.RolePlatformAdmin)
	if w.Code != http.StatusOK {
		t.Fatalf("导出商品: %d %s", w.Code, w.Body.String())
	}
	body := w.Body.String()
	// 审核维度与上架维度必须是**两列**，不能合并成一个"状态"
	for _, want := range []string{"审核状态", "上架状态", "价格方式", "交付方式", "导出用整机", "导出用面议件", "面议", "物流发货"} {
		if !strings.Contains(body, want) {
			t.Fatalf("导出内容缺少 %q：%s", want, firstN(body, 400))
		}
	}
	// 非平台管理员不得导出全量（与其它资源同一门槛）
	w = requestAs(t, app, http.MethodGet, "/api/v1/admin/export/products", nil, "assoc-1", domain.RoleAssociationAdmin)
	if w.Code != http.StatusForbidden {
		t.Fatalf("协会管理员不应能全量导出商品，实际 %d", w.Code)
	}
}

func firstN(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
