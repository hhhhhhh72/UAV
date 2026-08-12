package httpapi_test

import (
	"net/http"
	"strings"
	"testing"

	"drone-platform/internal/domain"
	"net/http/httptest"
)

func TestHomeEndpoint(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodGet, "/api/v1/home?city=重庆", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK { t.Fatalf("home: %d", w.Code) }
	if !strings.Contains(w.Body.String(), "banners") { t.Fatal("missing banners") }
}

func TestSearchEndpoint(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodGet, "/api/v1/search?q=巡检", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK { t.Fatalf("search: %d %s", w.Code, w.Body.String()) }
}

func TestMeEndpoint(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodGet, "/api/v1/me", nil, domain.RoleEnterprise)
	if w.Code != http.StatusOK { t.Fatalf("me: %d %s", w.Code, w.Body.String()) }
}

// PATCH /api/v1/me 全字段保存后，GET /api/v1/me 必须能读回手机号/性别/生日/地区/简介；
// 非法手机号必须 400 拒绝。
func TestUpdateMeProfileRoundtrip(t *testing.T) {
	app := newBizServer(t)
	body := []byte(`{"name":"张三","gender":"male","birthday":"1995-06-18","region":"重庆市江北区","bio":"无人机测绘飞手","phone":"13800138000"}`)
	w := request(t, app, http.MethodPatch, "/api/v1/me", body, domain.RoleEnterprise)
	if w.Code != http.StatusOK { t.Fatalf("patch me: %d %s", w.Code, w.Body.String()) }

	w = request(t, app, http.MethodGet, "/api/v1/me", nil, domain.RoleEnterprise)
	if w.Code != http.StatusOK { t.Fatalf("get me: %d %s", w.Code, w.Body.String()) }
	s := w.Body.String()
	for _, kv := range []struct{ k, v string }{
		{`"gender":"male"`, "gender"},
		{`"birthday":"1995-06-18"`, "birthday"},
		{`"region":"重庆市江北区"`, "region"},
		{`"bio":"无人机测绘飞手"`, "bio"},
		{`"phone":"13800138000"`, "phone"},
	} {
		if !strings.Contains(s, kv.k) {
			t.Fatalf("me response missing %s: %s", kv.v, s)
		}
	}

	bad := request(t, app, http.MethodPatch, "/api/v1/me", []byte(`{"phone":"123"}`), domain.RoleEnterprise)
	if bad.Code != http.StatusBadRequest {
		t.Fatalf("invalid phone should be 400, got %d %s", bad.Code, bad.Body.String())
	}
}

func TestAdminDashboard(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodGet, "/api/v1/admin/dashboard", nil, domain.RolePlatformAdmin)
	if w.Code != http.StatusOK { t.Fatalf("dashboard: %d %s", w.Code, w.Body.String()) }
}

func TestAdminDashboardForbidden(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodGet, "/api/v1/admin/dashboard", nil, domain.RoleIndividual)
	if w.Code != http.StatusForbidden { t.Fatalf("should be forbidden: %d", w.Code) }
}

func TestMessagesFlow(t *testing.T) {
	app := newBizServer(t)
	// Unread count
	w := request(t, app, http.MethodGet, "/api/v1/messages/unread-count", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK { t.Fatalf("unread: %d", w.Code) }
	// List
	w = request(t, app, http.MethodGet, "/api/v1/messages", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK { t.Fatalf("messages: %d", w.Code) }
}

func TestNewsFlow(t *testing.T) {
	app := newBizServer(t)
	// Public list
	w := request(t, app, http.MethodGet, "/api/v1/articles", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK { t.Fatalf("articles: %d", w.Code) }
	// Admin create
	body := []byte(`{"title":"测试新闻","content":"内容","category":"policy","source":"协会"}`)
	w = request(t, app, http.MethodPost, "/api/v1/articles", body, domain.RoleAssociationAdmin)
	if w.Code != http.StatusCreated { t.Fatalf("create article: %d %s", w.Code, w.Body.String()) }
}

func TestReviewsFlow(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodGet, "/api/v1/reviews?target_type=enterprise&target_id=ent-1", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK { t.Fatalf("reviews: %d", w.Code) }
}

func TestVenuesFlow(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodGet, "/api/v1/venues", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK { t.Fatalf("venues: %d", w.Code) }
}

func TestTrainingCoursesFlow(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodGet, "/api/v1/training-courses", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK { t.Fatalf("courses: %d", w.Code) }
}

func TestCertificatesFlow(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodGet, "/api/v1/certificates/mine", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK { t.Fatalf("certs: %d", w.Code) }
}

func TestProductsFlow(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodGet, "/api/v1/products", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK { t.Fatalf("products: %d", w.Code) }
}

func TestLoansFlow(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodGet, "/api/v1/loans/mine", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK { t.Fatalf("loans: %d", w.Code) }
}

func TestProjectAppsFlow(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodGet, "/api/v1/project-applications/mine", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK { t.Fatalf("project-apps: %d", w.Code) }
}

func TestUnauthorizedAccess(t *testing.T) {
	app := newBizServer(t)
	// No auth for admin endpoint
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
	app.ServeHTTP(w, r)
	if w.Code != http.StatusUnauthorized { t.Fatalf("should be 401: %d", w.Code) }
}
