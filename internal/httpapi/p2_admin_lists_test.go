package httpapi_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"drone-platform/internal/domain"
)

// P2-1 回归：七模块管理端列表路由必须存在、受 adminGate 保护、且能查到数据。
// 记录经由用户侧既有创建端点写入（newBizServer 已接线全部服务）。

func TestAdminListsP2Modules(t *testing.T) {
	app := newBizServer(t)

	// ── 造数：用户侧创建各模块记录 ──
	w := request(t, app, http.MethodPost, "/api/v1/policies",
		[]byte(`{"drone_model":"M30","drone_sn":"SN-1","policy_type":"第三者责任险","premium_fen":100,"coverage_fen":100000,"start_date":"2026-08-01T00:00:00Z","end_date":"2027-08-01T00:00:00Z"}`),
		domain.RoleIndividual)
	if w.Code != http.StatusCreated {
		t.Fatalf("create policy: %d %s", w.Code, w.Body.String())
	}
	w = request(t, app, http.MethodPost, "/api/v1/inspections",
		[]byte(`{"drone_model":"M30","drone_sn":"SN-1","inspect_date":"2026-08-01T00:00:00Z","expire_date":"2027-08-01T00:00:00Z"}`),
		domain.RoleIndividual)
	if w.Code != http.StatusCreated {
		t.Fatalf("create inspection: %d %s", w.Code, w.Body.String())
	}
	w = request(t, app, http.MethodPost, "/api/v1/repairs",
		[]byte(`{"product_desc":"M30 云台","fault_desc":"云台抖动"}`), domain.RoleIndividual)
	if w.Code != http.StatusCreated {
		t.Fatalf("create repair: %d %s", w.Code, w.Body.String())
	}
	w = request(t, app, http.MethodPost, "/api/v1/loans",
		[]byte(`{"amount_fen":500000,"term_months":12,"purpose":"购置无人机"}`), domain.RoleIndividual)
	if w.Code != http.StatusCreated {
		t.Fatalf("create loan: %d %s", w.Code, w.Body.String())
	}
	w = request(t, app, http.MethodPost, "/api/v1/resumes",
		[]byte(`{"title":"飞手简历","name":"张三","phone":"13800000000","email":"z@test.cn"}`), domain.RoleIndividual)
	if w.Code != http.StatusCreated {
		t.Fatalf("create resume: %d %s", w.Code, w.Body.String())
	}

	// ── 管理端列表：五条新分页接口必须返回数据 ──
	assertAdminPage(t, app, "/api/v1/admin/policies", 1)
	assertAdminPage(t, app, "/api/v1/admin/repairs", 1)
	assertAdminPage(t, app, "/api/v1/admin/loans", 1)
	assertAdminPage(t, app, "/api/v1/admin/resumes", 1)

	// inspections 为全量列表（非分页）
	w = request(t, app, http.MethodGet, "/api/v1/admin/inspections", nil, domain.RolePlatformAdmin)
	if w.Code != http.StatusOK {
		t.Fatalf("admin inspections: %d %s", w.Code, w.Body.String())
	}
	var insp struct {
		Data []domain.AnnualInspection `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &insp); err != nil || len(insp.Data) != 1 {
		t.Fatalf("admin inspections data: %+v err=%v", insp.Data, err)
	}

	// ── 复用公开 handler 的五条路由必须可用（空数据也 200）──
	for _, path := range []string{
		"/api/v1/admin/venues",
		"/api/v1/admin/instructors",
		"/api/v1/admin/cooperations",
		"/api/v1/admin/emergency-depts",
		"/api/v1/admin/rescue-cases",
	} {
		w = request(t, app, http.MethodGet, path, nil, domain.RolePlatformAdmin)
		if w.Code != http.StatusOK {
			t.Fatalf("%s: %d %s", path, w.Code, w.Body.String())
		}
	}

	// ── adminGate：非管理员访问新路由 → 403 ──
	for _, path := range []string{
		"/api/v1/admin/policies",
		"/api/v1/admin/inspections",
		"/api/v1/admin/repairs",
		"/api/v1/admin/loans",
		"/api/v1/admin/resumes",
		"/api/v1/admin/venues",
		"/api/v1/admin/instructors",
		"/api/v1/admin/cooperations",
		"/api/v1/admin/emergency-depts",
		"/api/v1/admin/rescue-cases",
	} {
		w = request(t, app, http.MethodGet, path, nil, domain.RoleIndividual)
		if w.Code != http.StatusForbidden {
			t.Fatalf("%s as individual: %d, want 403", path, w.Code)
		}
	}
}

// assertAdminPage 断言分页管理列表返回 200 且 total>=want。
func assertAdminPage(t *testing.T, app http.Handler, path string, wantTotal int) {
	t.Helper()
	w := request(t, app, http.MethodGet, path, nil, domain.RolePlatformAdmin)
	if w.Code != http.StatusOK {
		t.Fatalf("%s: %d %s", path, w.Code, w.Body.String())
	}
	_, total := listEnvelope(t, w)
	if total < wantTotal {
		t.Fatalf("%s total: %d, want >= %d", path, total, wantTotal)
	}
}
