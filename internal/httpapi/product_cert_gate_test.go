package httpapi_test

import (
	"net/http"
	"strings"
	"testing"

	"drone-platform/internal/domain"
)

// 商品发布的认证门禁。
//
// 背景：商品详情页对买家白纸黑字承诺「平台认证商家」，而此前只要登录就能发商品
// ——标识对每一件在售商品都是假话（生产实测 17/17 卖家无企业认证）。
// 现在只有完成**企业认证（approved）**的账号能上架，标识随之变成真话。
//
// 门禁放在 handler 而不是 service.CreateProduct：这是**路径策略**，不是"创建商品"的
// 固有不变式——管理端代建（adminCreateProduct）要能替任意卖家建商品，不该被卡住。
// 所以这条用例走的是用户自助发布的 HTTP 路径。
func TestCreateProductRequiresEnterpriseCert(t *testing.T) {
	app := newBizServer(t)
	body := `{"title":"认证门禁测试机","prod_type":"drone","price_fen":100000}`

	// seller-1 在测试环境（productSellerEntRepo）已预置 approved 企业认证 → 可发布
	if w := requestAs(t, app, http.MethodPost, "/api/v1/products", []byte(body),
		"seller-1", domain.RoleEnterprise); w.Code != http.StatusCreated {
		t.Fatalf("已认证卖家应可发布商品，code=%d body=%s", w.Code, w.Body.String())
	}

	// 未认证账号 → 403。个人身份必然没有企业认证，用它代表"没认证的账号"。
	w := requestAs(t, app, http.MethodPost, "/api/v1/products", []byte(body),
		authRoleUserID(domain.RoleIndividual), domain.RoleIndividual)
	if w.Code != http.StatusForbidden {
		t.Fatalf("未认证账号不应能发布商品，code=%d body=%s", w.Code, w.Body.String())
	}
	// 403 必须说清原因，否则用户不知道要去认证
	if !strings.Contains(w.Body.String(), "企业认证") {
		t.Fatalf("403 文案应指明需要企业认证，实际 %s", w.Body.String())
	}
}
