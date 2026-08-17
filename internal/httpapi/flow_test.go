package httpapi_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/httpapi"
	"net/http/httptest"
)

func TestHomeEndpoint(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodGet, "/api/v1/home?city=重庆", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK {
		t.Fatalf("home: %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "banners") {
		t.Fatal("missing banners")
	}
}

func TestSearchEndpoint(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodGet, "/api/v1/search?q=巡检", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK {
		t.Fatalf("search: %d %s", w.Code, w.Body.String())
	}
}

func TestMeEndpoint(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodGet, "/api/v1/me", nil, domain.RoleEnterprise)
	if w.Code != http.StatusOK {
		t.Fatalf("me: %d %s", w.Code, w.Body.String())
	}
}

// TestMeDemandCountIsPersonal: demand_count 必须只统计当前用户发布的需求。
// 此前实现误用全平台计数（List 空 filter），用户 A 看到的是全站需求数。
func TestMeDemandCountIsPersonal(t *testing.T) {
	app := newServer(t)

	// user-1（enterprise）发布一条需求
	w := request(t, app, http.MethodPost, "/api/v1/demands",
		[]byte(`{"publisher_name":"我司","contact":"13800138000","title":"个人计数需求","biz_type":"other"}`),
		domain.RoleEnterprise)
	if w.Code != http.StatusCreated {
		t.Fatalf("create demand: %d %s", w.Code, w.Body.String())
	}

	var resp struct {
		Data struct {
			DemandCount int `json:"demand_count"`
		} `json:"data"`
	}
	// user-1 本人：demand_count >= 1
	me := request(t, app, http.MethodGet, "/api/v1/me", nil, domain.RoleEnterprise)
	if me.Code != http.StatusOK {
		t.Fatalf("me: %d %s", me.Code, me.Body.String())
	}
	if err := json.Unmarshal(me.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode me: %v", err)
	}
	if resp.Data.DemandCount < 1 {
		t.Fatalf("user-1 demand_count should be >= 1, got %d (%s)", resp.Data.DemandCount, me.Body.String())
	}

	// 其他用户：demand_count 必须为 0（内存种子需求发布者是 enterprise-001）
	tokens, err := httpapi.NewTokenManager(testSecret)
	if err != nil {
		t.Fatal(err)
	}
	tok, err := tokens.Issue(domain.Actor{ID: "other-user", Role: domain.RoleIndividual}, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	r.Header.Set("Authorization", "Bearer "+tok)
	w2 := httptest.NewRecorder()
	app.ServeHTTP(w2, r)
	if w2.Code != http.StatusOK {
		t.Fatalf("other me: %d %s", w2.Code, w2.Body.String())
	}
	if err := json.Unmarshal(w2.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode other me: %v", err)
	}
	if resp.Data.DemandCount != 0 {
		t.Fatalf("other-user demand_count should be 0, got %d (%s)", resp.Data.DemandCount, w2.Body.String())
	}
}

// PATCH /api/v1/me 全字段保存后，GET /api/v1/me 必须能读回手机号/性别/生日/地区/简介；
// 非法手机号必须 400 拒绝。
func TestUpdateMeProfileRoundtrip(t *testing.T) {
	app := newBizServer(t)
	body := []byte(`{"name":"张三","gender":"male","birthday":"1995-06-18","region":"重庆市江北区","bio":"无人机测绘飞手","phone":"13800138000"}`)
	w := request(t, app, http.MethodPatch, "/api/v1/me", body, domain.RoleEnterprise)
	if w.Code != http.StatusOK {
		t.Fatalf("patch me: %d %s", w.Code, w.Body.String())
	}

	w = request(t, app, http.MethodGet, "/api/v1/me", nil, domain.RoleEnterprise)
	if w.Code != http.StatusOK {
		t.Fatalf("get me: %d %s", w.Code, w.Body.String())
	}
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
	if w.Code != http.StatusOK {
		t.Fatalf("dashboard: %d %s", w.Code, w.Body.String())
	}
}

func TestAdminDashboardForbidden(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodGet, "/api/v1/admin/dashboard", nil, domain.RoleIndividual)
	if w.Code != http.StatusForbidden {
		t.Fatalf("should be forbidden: %d", w.Code)
	}
}

func TestMessagesFlow(t *testing.T) {
	app := newBizServer(t)
	// Unread count
	w := request(t, app, http.MethodGet, "/api/v1/messages/unread-count", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK {
		t.Fatalf("unread: %d", w.Code)
	}
	// List
	w = request(t, app, http.MethodGet, "/api/v1/messages", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK {
		t.Fatalf("messages: %d", w.Code)
	}
}

func TestNewsFlow(t *testing.T) {
	app := newBizServer(t)
	// Public list
	w := request(t, app, http.MethodGet, "/api/v1/articles", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK {
		t.Fatalf("articles: %d", w.Code)
	}
	// Admin create
	body := []byte(`{"title":"测试新闻","content":"内容","category":"policy","source":"协会"}`)
	w = request(t, app, http.MethodPost, "/api/v1/articles", body, domain.RoleAssociationAdmin)
	if w.Code != http.StatusCreated {
		t.Fatalf("create article: %d %s", w.Code, w.Body.String())
	}
}

func TestReviewsFlow(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodGet, "/api/v1/reviews?target_type=enterprise&target_id=ent-1", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK {
		t.Fatalf("reviews: %d", w.Code)
	}
}

func TestVenuesFlow(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodGet, "/api/v1/venues", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK {
		t.Fatalf("venues: %d", w.Code)
	}
}

func TestTrainingCoursesFlow(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodGet, "/api/v1/training-courses", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK {
		t.Fatalf("courses: %d", w.Code)
	}
}

func TestCertificatesFlow(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodGet, "/api/v1/certificates/mine", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK {
		t.Fatalf("certs: %d", w.Code)
	}
}

func TestProductsFlow(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodGet, "/api/v1/products", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK {
		t.Fatalf("products: %d", w.Code)
	}
}

func TestLoansFlow(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodGet, "/api/v1/loans/mine", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK {
		t.Fatalf("loans: %d", w.Code)
	}
}

func TestProjectAppsFlow(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodGet, "/api/v1/project-applications/mine", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK {
		t.Fatalf("project-apps: %d", w.Code)
	}
}

func TestUnauthorizedAccess(t *testing.T) {
	app := newBizServer(t)
	// No auth for admin endpoint
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
	app.ServeHTTP(w, r)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("should be 401: %d", w.Code)
	}
}
