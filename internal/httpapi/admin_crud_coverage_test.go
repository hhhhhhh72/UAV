package httpapi_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"drone-platform/internal/domain"
)

// nonexistentID 是测试里统一使用的"不存在"资源 id，用于覆盖 get/update/delete
// handler 的 not-found 错误分支。
const nonexistentID = "zzz-nonexistent-0"

// contains 报告 code 是否落在允许状态码集合内（避免引入 golang.org/x/exp）。
func contains(allowed []int, code int) bool {
	for _, c := range allowed {
		if c == code {
			return true
		}
	}
	return false
}

// truncate 截断字符串前 n 个字符（按 rune，避免切开多字节 UTF-8 序列）。
func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) > n {
		return string(r[:n])
	}
	return s
}

// adminResource 描述一个管理端资源及其已注册的 CRUD 路由。
type adminResource struct {
	name      string
	listPath  string
	hasDetail bool // GET    /api/v1/admin/<name>/{id}
	hasUpdate bool // PUT    /api/v1/admin/<name>/{id}
	hasCreate bool // POST   /api/v1/admin/<name>
	hasDelete bool // DELETE /api/v1/admin/<name>/{id}
}

// adminResources 是有完整 CRUD 路由的管理端资源。
// service-listings 没有管理端 GET 详情路由（hasDetail=false）。
var adminResources = []adminResource{
	{name: "training-courses", listPath: "/api/v1/admin/training-courses", hasDetail: true, hasUpdate: true, hasCreate: true, hasDelete: true},
	{name: "certificates", listPath: "/api/v1/admin/certificates", hasDetail: true, hasUpdate: true, hasCreate: true, hasDelete: true},
	{name: "jobs", listPath: "/api/v1/admin/jobs", hasDetail: true, hasUpdate: true, hasCreate: true, hasDelete: true},
	{name: "colleges", listPath: "/api/v1/admin/colleges", hasDetail: true, hasUpdate: true, hasCreate: true, hasDelete: true},
	{name: "study-tours", listPath: "/api/v1/admin/study-tours", hasDetail: true, hasUpdate: true, hasCreate: true, hasDelete: true},
	{name: "achievements", listPath: "/api/v1/admin/achievements", hasDetail: true, hasUpdate: true, hasCreate: true, hasDelete: true},
	{name: "rd-challenges", listPath: "/api/v1/admin/rd-challenges", hasDetail: true, hasUpdate: true, hasCreate: true, hasDelete: true},
	{name: "research-projects", listPath: "/api/v1/admin/research-projects", hasDetail: true, hasUpdate: true, hasCreate: true, hasDelete: true},
	{name: "test-sites", listPath: "/api/v1/admin/test-sites", hasDetail: true, hasUpdate: true, hasCreate: true, hasDelete: true},
	{name: "transformations", listPath: "/api/v1/admin/transformations", hasDetail: true, hasUpdate: true, hasCreate: true, hasDelete: true},
	{name: "events", listPath: "/api/v1/admin/events", hasDetail: true, hasUpdate: true, hasCreate: true, hasDelete: true},
	{name: "portfolios", listPath: "/api/v1/admin/portfolios", hasDetail: true, hasUpdate: true, hasCreate: true, hasDelete: true},
	{name: "exhibitions", listPath: "/api/v1/admin/exhibitions", hasDetail: true, hasUpdate: true, hasCreate: true, hasDelete: true},
	{name: "industry-reports", listPath: "/api/v1/admin/industry-reports", hasDetail: true, hasUpdate: true, hasCreate: true, hasDelete: true},
	{name: "emergency-resources", listPath: "/api/v1/admin/emergency-resources", hasDetail: true, hasUpdate: true, hasCreate: true, hasDelete: true},
	{name: "emergency-dispatches", listPath: "/api/v1/admin/emergency-dispatches", hasDetail: true, hasUpdate: true, hasCreate: true, hasDelete: true},
	{name: "messages", listPath: "/api/v1/admin/messages", hasDetail: true, hasUpdate: true, hasCreate: true, hasDelete: true},
	{name: "compliance-docs", listPath: "/api/v1/admin/compliance-docs", hasDetail: true, hasUpdate: true, hasCreate: true, hasDelete: true},
	{name: "compliance-standards", listPath: "/api/v1/admin/compliance-standards", hasDetail: true, hasUpdate: true, hasCreate: true, hasDelete: true},
	{name: "industry-resources", listPath: "/api/v1/admin/industry-resources", hasDetail: true, hasUpdate: true, hasCreate: true, hasDelete: true},
	{name: "competitions", listPath: "/api/v1/admin/competitions", hasDetail: true, hasUpdate: true, hasCreate: true, hasDelete: true},
	{name: "cases", listPath: "/api/v1/admin/cases", hasDetail: true, hasUpdate: true, hasCreate: true, hasDelete: true},
	{name: "experts", listPath: "/api/v1/admin/experts", hasDetail: true, hasUpdate: true, hasCreate: true, hasDelete: true},
	{name: "orders", listPath: "/api/v1/admin/orders", hasDetail: true, hasUpdate: true, hasCreate: true, hasDelete: true},
	{name: "products", listPath: "/api/v1/admin/products", hasDetail: true, hasUpdate: true, hasCreate: true, hasDelete: true},
	{name: "service-listings", listPath: "/api/v1/admin/service-listings", hasDetail: false, hasUpdate: true, hasCreate: true, hasDelete: true},
}

// adminListOnly 是只有管理端 GET 列表的路由（复用公开 handler 或 P2-1 补齐），
// 只打"匿名 401 + 管理员列表 200"两类请求。
var adminListOnly = []adminResource{
	{name: "enrollments", listPath: "/api/v1/admin/enrollments"},
	{name: "test-site-bookings", listPath: "/api/v1/admin/test-sites/bookings"},
	{name: "policies", listPath: "/api/v1/admin/policies"},
	{name: "inspections", listPath: "/api/v1/admin/inspections"},
	{name: "repairs", listPath: "/api/v1/admin/repairs"},
	{name: "loans", listPath: "/api/v1/admin/loans"},
	{name: "resumes", listPath: "/api/v1/admin/resumes"},
	{name: "venues", listPath: "/api/v1/admin/venues"},
	{name: "instructors", listPath: "/api/v1/admin/instructors"},
	{name: "cooperations", listPath: "/api/v1/admin/cooperations"},
	{name: "emergency-depts", listPath: "/api/v1/admin/emergency-depts"},
	{name: "rescue-cases", listPath: "/api/v1/admin/rescue-cases"},
}

// 错误分支状态码集合：adminFail 统一把 not-found 映射为 404；
// 其余（decode 失败等）为 400/409/422。500 已从 update/delete 集合移除——
// 若再出现 500 即回归（此前 update/delete 对 not-found 一律 500）。
var (
	okGetCodes    = []int{http.StatusBadRequest, http.StatusNotFound, http.StatusConflict}                              // 400/404/409
	okUpdateCodes = []int{http.StatusBadRequest, http.StatusNotFound, http.StatusConflict, http.StatusUnprocessableEntity} // 400/404/409/422
	okDeleteCodes = []int{http.StatusBadRequest, http.StatusNotFound, http.StatusConflict}                               // 400/404/409
)

// checkStatus 断言响应状态码等于 want，失败时输出 method/path/code/body 前 200 字符。
func checkStatus(t *testing.T, app http.Handler, method, path, body, token string, want int) {
	t.Helper()
	w := doRaw(app, method, path, body, token)
	if w.Code != want {
		t.Fatalf("%s %s: got %d, want %d; body: %s", method, path, w.Code, want, truncate(w.Body.String(), 200))
	}
}

// checkStatusIn 断言响应状态码落在 ok 集合内。
func checkStatusIn(t *testing.T, app http.Handler, method, path, body, token string, ok []int) {
	t.Helper()
	w := doRaw(app, method, path, body, token)
	if !contains(ok, w.Code) {
		t.Fatalf("%s %s: got %d, not in %v; body: %s", method, path, w.Code, ok, truncate(w.Body.String(), 200))
	}
}

// checkStatusNotOK 断言响应状态码不是 200（创建 handler 空对象可能走 service 返回
// 201/400/500，唯一不可能的是 200，因为所有创建成功均返回 201）。
func checkStatusNotOK(t *testing.T, app http.Handler, method, path, body, token string) {
	t.Helper()
	w := doRaw(app, method, path, body, token)
	if w.Code == http.StatusOK {
		t.Fatalf("%s %s: got 200 (unexpected success); body: %s", method, path, truncate(w.Body.String(), 200))
	}
}

func TestAdminCRUDCoverage(t *testing.T) {
	app := newBizServer(t)
	adminTok := authAs(t, "admin-1", domain.RolePlatformAdmin)

	all := make([]adminResource, 0, len(adminResources)+len(adminListOnly))
	all = append(all, adminResources...)
	all = append(all, adminListOnly...)

	for _, res := range all {
		res := res
		t.Run(res.name, func(t *testing.T) {
			// 1. 匿名 → 401（覆盖鉴权门）
			checkStatus(t, app, http.MethodGet, res.listPath, "", "", http.StatusUnauthorized)

			// 2. 管理员 GET 列表 → 200（覆盖 list handler / adminListFilter / paginatedRespond）
			checkStatus(t, app, http.MethodGet, res.listPath, "", adminTok, http.StatusOK)

			// 3. 管理员 GET 详情（不存在 id）→ 覆盖 get handler 错误分支
			if res.hasDetail {
				checkStatusIn(t, app, http.MethodGet, res.listPath+"/"+nonexistentID, "", adminTok, okGetCodes)
			}
			// 4. 管理员 PUT 更新（不存在 id）→ 覆盖 update handler 错误分支
			if res.hasUpdate {
				checkStatusIn(t, app, http.MethodPut, res.listPath+"/"+nonexistentID, `{}`, adminTok, okUpdateCodes)
			}
			// 5. 管理员 POST 创建（空对象）→ 覆盖 create handler 的 decode/校验分支
			if res.hasCreate {
				checkStatusNotOK(t, app, http.MethodPost, res.listPath, `{}`, adminTok)
			}
			// 6. 管理员 DELETE（不存在 id）→ 覆盖 delete handler 错误分支
			if res.hasDelete {
				checkStatusIn(t, app, http.MethodDelete, res.listPath+"/"+nonexistentID, "", adminTok, okDeleteCodes)
			}
		})
	}
}

// TestAdminCRUDCoverageExpertPositive 正向用例：管理员创建专家 → 201，列表能查到该 id。
func TestAdminCRUDCoverageExpertPositive(t *testing.T) {
	app := newBizServer(t)
	adminTok := authAs(t, "admin-1", domain.RolePlatformAdmin)

	w := doRaw(app, http.MethodPost, "/api/v1/admin/experts",
		`{"name":"覆盖测试专家","title":"教授","org":"重庆大学","field":"无人机","bio":"覆盖率正向用例"}`, adminTok)
	if w.Code != http.StatusCreated {
		t.Fatalf("POST /api/v1/admin/experts: got %d, want 201; body: %s", w.Code, truncate(w.Body.String(), 200))
	}
	var created struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatalf("parse created expert: %v; body: %s", err, truncate(w.Body.String(), 200))
	}
	if created.Data.ID == "" {
		t.Fatalf("created expert missing id: %s", truncate(w.Body.String(), 200))
	}

	lw := doRaw(app, http.MethodGet, "/api/v1/admin/experts", "", adminTok)
	if lw.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/admin/experts: got %d, want 200; body: %s", lw.Code, truncate(lw.Body.String(), 200))
	}
	if !strings.Contains(lw.Body.String(), created.Data.ID) {
		t.Fatalf("expert list should contain %q; body: %s", created.Data.ID, truncate(lw.Body.String(), 200))
	}
}

// TestAdminCRUDCoverageCasePositive 正向用例：管理员创建案例 → 201，详情能查到。
func TestAdminCRUDCoverageCasePositive(t *testing.T) {
	app := newBizServer(t)
	adminTok := authAs(t, "admin-1", domain.RolePlatformAdmin)

	w := doRaw(app, http.MethodPost, "/api/v1/admin/cases",
		`{"title":"覆盖测试案例","category":"logistics","description":"无人机配送案例","client_name":"客户A","result":"成果"}`, adminTok)
	if w.Code != http.StatusCreated {
		t.Fatalf("POST /api/v1/admin/cases: got %d, want 201; body: %s", w.Code, truncate(w.Body.String(), 200))
	}
	var created struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatalf("parse created case: %v; body: %s", err, truncate(w.Body.String(), 200))
	}
	if created.Data.ID == "" {
		t.Fatalf("created case missing id: %s", truncate(w.Body.String(), 200))
	}

	gw := doRaw(app, http.MethodGet, "/api/v1/admin/cases/"+created.Data.ID, "", adminTok)
	if gw.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/admin/cases/%s: got %d, want 200; body: %s", created.Data.ID, gw.Code, truncate(gw.Body.String(), 200))
	}
}
