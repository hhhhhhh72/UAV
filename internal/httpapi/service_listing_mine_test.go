package httpapi_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"drone-platform/internal/domain"
)

// 用户自助发布服务能力：待审核（pending）、不进公开列表、mine=1 可见、匿名 mine 为空。
func TestCreateServiceListingPendingAndMine(t *testing.T) {
	app := newBizServer(t)

	// 1) 用户发布服务能力 → 201 + status=pending
	body := []byte(`{"provider_name":"测试机构","title":"低空巡检服务","category":"巡检","description":"电力巡检作业","region":"南岸区","price_fen":10000,"unit":"按次"}`)
	w := request(t, app, http.MethodPost, "/api/v1/service-listings", body, domain.RoleIndividual)
	if w.Code != http.StatusCreated {
		t.Fatalf("create service listing: %d %s", w.Code, w.Body.String())
	}
	var envelope struct {
		Data struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	created := envelope.Data
	if created.ID == "" {
		t.Fatal("created listing id is empty")
	}
	if created.Status != "pending" {
		t.Fatalf("status = %q, want pending", created.Status)
	}

	// 2) 公开列表不包含（待审核不公开），且匿名可访问（白名单回归：此前 401）
	w = request(t, app, http.MethodGet, "/api/v1/service-listings", nil, "")
	if w.Code != http.StatusOK {
		t.Fatalf("public list: %d (was 401 before whitelist fix)", w.Code)
	}
	if strings.Contains(w.Body.String(), created.ID) {
		t.Fatal("pending listing must not appear in public list")
	}

	// 3) mine=1 未登录 → 空列表
	w = request(t, app, http.MethodGet, "/api/v1/service-listings?mine=1", nil, "")
	if w.Code != http.StatusOK {
		t.Fatalf("anon mine: %d", w.Code)
	}
	if strings.Contains(w.Body.String(), created.ID) {
		t.Fatal("anonymous mine must not leak listing")
	}

	// 4) mine=1 登录 → 包含自己的待审核记录
	w = request(t, app, http.MethodGet, "/api/v1/service-listings?mine=1", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK {
		t.Fatalf("mine list: %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), created.ID) {
		t.Fatal("mine=1 should include own pending listing")
	}
}

// 用户发布课程待审核（pending），不进公开列表；管理端审核通过（published）后公开；
// mine=1 只看自己（机构）的课程，匿名为空。
func TestCreateCoursePendingAndMine(t *testing.T) {
	app := newBizServer(t)

	body := []byte(`{"title":"CAAC 多旋翼执照班","cert_type":"caac","description":"执照培训","org_name":"测试航校","district":"南岸区","location":"金开大道68号","price_fen":980000,"duration_days":25,"max_students":20}`)
	w := request(t, app, http.MethodPost, "/api/v1/training-courses", body, domain.RoleIndividual)
	if w.Code != http.StatusCreated {
		t.Fatalf("create course: %d %s", w.Code, w.Body.String())
	}
	var envelope struct {
		Data struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	created := envelope.Data
	if created.Status != "pending" {
		t.Fatalf("course status = %q, want pending", created.Status)
	}

	// 公开列表不含待审核课程
	w = request(t, app, http.MethodGet, "/api/v1/training-courses", nil, "")
	if strings.Contains(w.Body.String(), created.ID) {
		t.Fatal("pending course must not appear in public list")
	}

	// mine=1 未登录 → 空
	w = request(t, app, http.MethodGet, "/api/v1/training-courses?mine=1", nil, "")
	if strings.Contains(w.Body.String(), created.ID) {
		t.Fatal("anonymous course mine must not leak")
	}

	// mine=1 登录 → 包含（含待审核）
	w = request(t, app, http.MethodGet, "/api/v1/training-courses?mine=1", nil, domain.RoleIndividual)
	if !strings.Contains(w.Body.String(), created.ID) {
		t.Fatal("mine=1 should include own course")
	}

	// 管理端审核通过（status → published）后进入公开列表
	aw := request(t, app, http.MethodPut, "/api/v1/admin/training-courses/"+created.ID,
		[]byte(`{"status":"published"}`), domain.RolePlatformAdmin)
	if aw.Code != http.StatusOK {
		t.Fatalf("admin approve course: %d %s", aw.Code, aw.Body.String())
	}
	w = request(t, app, http.MethodGet, "/api/v1/training-courses", nil, "")
	if !strings.Contains(w.Body.String(), created.ID) {
		t.Fatal("approved course should appear in public list")
	}

	// 管理端关闭（closed）后不再公开
	cw := request(t, app, http.MethodPut, "/api/v1/admin/training-courses/"+created.ID,
		[]byte(`{"status":"closed"}`), domain.RolePlatformAdmin)
	if cw.Code != http.StatusOK {
		t.Fatalf("admin close course: %d %s", cw.Code, cw.Body.String())
	}
	w = request(t, app, http.MethodGet, "/api/v1/training-courses", nil, "")
	if strings.Contains(w.Body.String(), created.ID) {
		t.Fatal("closed course must not appear in public list")
	}
}
