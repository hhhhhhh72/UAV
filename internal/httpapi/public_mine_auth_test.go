package httpapi_test

import (
	"net/http"
	"testing"

	"drone-platform/internal/domain"
)

// TestMineEndpointsRequireAuth 回归：isPublicPath 不得放行 me/mine 类端点——
// association-members/me、portfolios/mine、jobs/mine 与 certified-pilots/mine 同款显式排除。
// 匿名必须 401；登录后端点仍可达（证明不是把端点封死）。
func TestMineEndpointsRequireAuth(t *testing.T) {
	app := newBizServer(t)

	anonCases := []struct {
		path string
	}{
		{"/api/v1/association-members/me"},
		{"/api/v1/portfolios/mine"},
		{"/api/v1/jobs/mine"},
		{"/api/v1/certified-pilots/mine"},
	}
	for _, c := range anonCases {
		w := doRaw(app, http.MethodGet, c.path, "", "")
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("GET %s anonymous: expected 401, got %d (%s)", c.path, w.Code, w.Body.String())
		}
	}

	// 登录后可达（证明不是把端点封死）：企业用户查询我的品牌展示 → 200 空列表
	entTok := authAs(t, "enterprise-1", domain.RoleEnterprise)
	w := doRaw(app, http.MethodGet, "/api/v1/portfolios/mine", "", entTok)
	if w.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/portfolios/mine authenticated: expected 200, got %d (%s)", w.Code, w.Body.String())
	}
}
