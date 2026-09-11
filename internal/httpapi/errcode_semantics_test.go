package httpapi_test

import (
	"net/http"
	"testing"

	"drone-platform/internal/domain"
)

// 错误码语义回归：同一个「资源不存在」，此前在不同模块分别被报成 403（审批类）、
// 400（删需求）、500（内存删评价）、200（PG 删评价）——四套答案。
// 现在统一为 404，且不再把「没这个记录」说成「你没权限」（权限排查最怕这个）。
func TestAdminGhostResourceReturns404(t *testing.T) {
	app := newBizServer(t)
	adminTok := authAs(t, "admin-1", domain.RolePlatformAdmin)

	cases := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{"评价-通过", http.MethodPost, "/api/v1/admin/reviews/ghost-review/approve", "{}"},
		{"评价-驳回", http.MethodPost, "/api/v1/admin/reviews/ghost-review/reject", "{}"},
		{"评价-删除", http.MethodDelete, "/api/v1/admin/reviews/ghost-review", ""},
		{"证书-通过", http.MethodPost, "/api/v1/admin/certificates/ghost-cert/approve", "{}"},
		{"证书-驳回", http.MethodPost, "/api/v1/admin/certificates/ghost-cert/reject", "{}"},
		{"飞手-通过", http.MethodPost, "/api/v1/admin/certified-pilots/ghost-pilot/approve", "{}"},
		{"飞手-驳回", http.MethodPost, "/api/v1/admin/certified-pilots/ghost-pilot/reject", `{"reason":"测试驳回"}`},
		{"培训师-通过", http.MethodPost, "/api/v1/admin/instructors/ghost-ins/approve", "{}"},
		{"需求-管理端删除", http.MethodDelete, "/api/v1/admin/demands/ghost-demand", ""},
		{"需求-本人端删除", http.MethodDelete, "/api/v1/demands/ghost-demand", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := doRaw(app, tc.method, tc.path, tc.body, adminTok)
			if w.Code != http.StatusNotFound {
				t.Fatalf("%s %s：期望 404（记录不存在），实际 %d %s", tc.method, tc.path, w.Code, w.Body.String())
			}
		})
	}
}
