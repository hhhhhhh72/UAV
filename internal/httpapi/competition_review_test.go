package httpapi_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"drone-platform/internal/domain"
)

// 赛事审核流回归：企业发布 → pending（不公开）→ 管理端审核（enrolling）→ 公开；
// 非企业角色发布 403；关闭（closed）后不再公开。
func TestCompetitionEnterpriseReviewFlow(t *testing.T) {
	app := newBizServer(t)

	// 个人角色发布 → 403
	w := request(t, app, http.MethodPost, "/api/v1/competitions",
		[]byte(`{"title":"个人赛事"}`), domain.RoleIndividual)
	if w.Code != http.StatusForbidden {
		t.Fatalf("individual create competition: want 403, got %d", w.Code)
	}

	// 企业发布 → 201 + pending
	body := []byte(`{"title":"2026企业杯无人机大赛","category":"竞技","location":"重庆国博","start_date":"2026-10-01","end_date":"2026-10-03","fee":38000,"max_teams":50}`)
	w = requestAs(t, app, http.MethodPost, "/api/v1/competitions", body, "ent-1", domain.RoleEnterprise)
	if w.Code != http.StatusCreated {
		t.Fatalf("enterprise create competition: %d %s", w.Code, w.Body.String())
	}
	var created struct {
		Data struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if created.Data.Status != "pending" {
		t.Fatalf("status = %q, want pending", created.Data.Status)
	}

	// 公开列表不含 pending
	w = request(t, app, http.MethodGet, "/api/v1/competitions", nil, "")
	if strings.Contains(w.Body.String(), created.Data.ID) {
		t.Fatal("pending competition must not be public")
	}

	// 详情接口：非管理端访问 pending → 404（防绕过列表过滤直接看详情）
	dw := request(t, app, http.MethodGet, "/api/v1/competitions/"+created.Data.ID, nil, "")
	if dw.Code != http.StatusNotFound {
		t.Fatalf("pending detail as anonymous: want 404, got %d", dw.Code)
	}
	dw = request(t, app, http.MethodGet, "/api/v1/competitions/"+created.Data.ID, nil, domain.RoleIndividual)
	if dw.Code != http.StatusNotFound {
		t.Fatalf("pending detail as individual: want 404, got %d", dw.Code)
	}
	// 管理端可查看详情
	dwa := request(t, app, http.MethodGet, "/api/v1/competitions/"+created.Data.ID, nil, domain.RolePlatformAdmin)
	if dwa.Code != http.StatusOK {
		t.Fatalf("pending detail as admin: want 200, got %d", dwa.Code)
	}

	// 报名接口：pending 赛事不可报名（404）
	rw := request(t, app, http.MethodPost, "/api/v1/competitions/"+created.Data.ID+"/register",
		[]byte(`{"team_name":"测试队","name":"张三","phone":"13800000000"}`), domain.RoleIndividual)
	if rw.Code != http.StatusNotFound {
		t.Fatalf("register pending competition: want 404, got %d", rw.Code)
	}

	// 管理端审核（enrolling）→ 公开
	aw := request(t, app, http.MethodPut, "/api/v1/admin/competitions/"+created.Data.ID,
		[]byte(`{"title":"2026企业杯无人机大赛","status":"enrolling"}`), domain.RolePlatformAdmin)
	if aw.Code != http.StatusOK {
		t.Fatalf("admin review: %d %s", aw.Code, aw.Body.String())
	}
	w = request(t, app, http.MethodGet, "/api/v1/competitions", nil, "")
	if !strings.Contains(w.Body.String(), created.Data.ID) {
		t.Fatal("approved competition should be public")
	}

	// 管理端关闭（closed）→ 不再公开
	cl := request(t, app, http.MethodPut, "/api/v1/admin/competitions/"+created.Data.ID,
		[]byte(`{"title":"2026企业杯无人机大赛","status":"closed"}`), domain.RolePlatformAdmin)
	if cl.Code != http.StatusOK {
		t.Fatalf("admin close: %d %s", cl.Code, cl.Body.String())
	}
	w = request(t, app, http.MethodGet, "/api/v1/competitions", nil, "")
	if strings.Contains(w.Body.String(), created.Data.ID) {
		t.Fatal("closed competition must not be public")
	}
}
