package httpapi_test

import (
	"net/http"
	"testing"

	"drone-platform/internal/domain"
)

// 角色矩阵回归：平台级操作（全局配置/全量导出/批量审批需求）仅平台管理员；
// 协会管理员（association_admin）此前可操作——垂直越权收紧。
func TestPlatformOnlyEndpointsForbidAssociationAdmin(t *testing.T) {
	app := newBizServer(t)

	platformOnly := []struct{ method, path, body string }{
		{http.MethodGet, "/api/v1/admin/config", ""},
		{http.MethodPost, "/api/v1/admin/config", `{}`},
		{http.MethodGet, "/api/v1/admin/export/demands", ""},
		{http.MethodGet, "/api/v1/admin/export/enterprises", ""},
		{http.MethodPost, "/api/v1/admin/demands/batch-approve", `{"ids":["d-1"]}`},
		{http.MethodPost, "/api/services/config", `{"config":{}}`},
	}
	for _, tc := range platformOnly {
		// 协会管理员 → 403
		w := requestAs(t, app, tc.method, tc.path, []byte(tc.body), "assoc-1", domain.RoleAssociationAdmin)
		if w.Code != http.StatusForbidden {
			t.Fatalf("association_admin %s %s: want 403, got %d %s", tc.method, tc.path, w.Code, w.Body.String())
		}
		// 平台管理员 → 放行（config GET 200 / export 200 / batch-approve 200 / services POST 200）
		w = requestAs(t, app, tc.method, tc.path, []byte(tc.body), "admin-1", domain.RolePlatformAdmin)
		if w.Code != http.StatusOK && w.Code != http.StatusCreated {
			t.Fatalf("platform_admin %s %s: want 200/201, got %d %s", tc.method, tc.path, w.Code, w.Body.String())
		}
	}

	// 业务内容管理仍对协会管理员开放（对照组）：创建专家应 201
	w := requestAs(t, app, http.MethodPost, "/api/v1/admin/experts",
		[]byte(`{"name":"协会可管专家","status":"published"}`), "assoc-1", domain.RoleAssociationAdmin)
	if w.Code != http.StatusCreated {
		t.Fatalf("association_admin create expert: want 201, got %d %s", w.Code, w.Body.String())
	}
}
