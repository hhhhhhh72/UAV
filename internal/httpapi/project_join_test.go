package httpapi_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"drone-platform/internal/domain"
)

// unwrapData 拆掉统一响应包装 {data: ...} 后再解析。
func unwrapData(t *testing.T, body []byte, v any) {
	t.Helper()
	var env struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		t.Fatalf("unwrap envelope: %v body=%s", err, body)
	}
	if err := json.Unmarshal(env.Data, v); err != nil {
		t.Fatalf("unwrap data: %v body=%s", err, body)
	}
}

// createTestProject 用管理端接口创建一条课题，返回课题 id。
func createTestProject(t *testing.T, app http.Handler) string {
	t.Helper()
	body := []byte(`{"title":"无人机适航验证课题","field":"适航验证","lead_org":"协会","status":"active","budget_fen":1000000}`)
	w := request(t, app, http.MethodPost, "/api/v1/admin/research-projects", body, domain.RolePlatformAdmin)
	if w.Code != http.StatusCreated {
		t.Fatalf("create project: %d %s", w.Code, w.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	unwrapData(t, w.Body.Bytes(), &p)
	if p.ID == "" {
		t.Fatalf("empty project id: %s", w.Body.String())
	}
	return p.ID
}

// TestProjectJoinFlow 课题攻关参与申请全链路：
// 登录申请 → 幂等重复 → 我的申请 → 非本人未申请 → 课题不存在 404 →
// 后台列表/越权 403 → 状态流转/非法状态 400。
func TestProjectJoinFlow(t *testing.T) {
	app := newBizServer(t)
	pid := createTestProject(t, app)

	// 未登录 → 401
	anon := httptest.NewRecorder()
	app.ServeHTTP(anon, httptest.NewRequest(http.MethodPost, "/api/v1/projects/"+pid+"/join", bytes.NewReader([]byte(`{"org_name":"天航科技"}`))))
	if anon.Code != http.StatusUnauthorized {
		t.Fatalf("join without auth: expected 401, got %d", anon.Code)
	}

	// 个人用户申请 → 201
	w := request(t, app, http.MethodPost, "/api/v1/projects/"+pid+"/join",
		[]byte(`{"org_name":"天航科技","message":"想参与样机测试"}`), domain.RoleIndividual)
	if w.Code != http.StatusCreated {
		t.Fatalf("join: expected 201, got %d %s", w.Code, w.Body.String())
	}
	var join struct {
		ID      string `json:"id"`
		Status  string `json:"status"`
		OrgName string `json:"org_name"`
	}
	unwrapData(t, w.Body.Bytes(), &join)
	if join.ID == "" || join.Status != "pending" || join.OrgName != "天航科技" {
		t.Fatalf("unexpected join: %+v", join)
	}

	// 重复申请 → 200 幂等（同一条记录，不重复创建）
	w2 := request(t, app, http.MethodPost, "/api/v1/projects/"+pid+"/join",
		[]byte(`{"org_name":"天航科技","message":"再次申请"}`), domain.RoleIndividual)
	if w2.Code != http.StatusOK {
		t.Fatalf("duplicate join: expected 200, got %d %s", w2.Code, w2.Body.String())
	}
	var join2 struct {
		ID string `json:"id"`
	}
	unwrapData(t, w2.Body.Bytes(), &join2)
	if join2.ID != join.ID {
		t.Fatalf("duplicate should return same id: %s vs %s", join2.ID, join.ID)
	}

	// 我的申请 → applied=true
	w3 := request(t, app, http.MethodGet, "/api/v1/projects/"+pid+"/join/mine", nil, domain.RoleIndividual)
	if w3.Code != http.StatusOK {
		t.Fatalf("mine: %d %s", w3.Code, w3.Body.String())
	}
	var mine struct {
		Applied bool `json:"applied"`
	}
	unwrapData(t, w3.Body.Bytes(), &mine)
	if !mine.Applied {
		t.Fatalf("mine applied should be true: %s", w3.Body.String())
	}

	// 其他用户（enterprise-1）未申请 → applied=false
	w4 := request(t, app, http.MethodGet, "/api/v1/projects/"+pid+"/join/mine", nil, domain.RoleEnterprise)
	if w4.Code != http.StatusOK {
		t.Fatalf("mine other: %d", w4.Code)
	}
	unwrapData(t, w4.Body.Bytes(), &mine)
	if mine.Applied {
		t.Fatalf("mine applied should be false: %s", w4.Body.String())
	}

	// 课题不存在 → 404
	w5 := request(t, app, http.MethodPost, "/api/v1/projects/proj-not-exist/join",
		[]byte(`{"org_name":"x"}`), domain.RoleIndividual)
	if w5.Code != http.StatusNotFound {
		t.Fatalf("join missing project: expected 404, got %d", w5.Code)
	}

	// 后台列表：管理员可见 1 条，个人越权 403
	w6 := request(t, app, http.MethodGet, "/api/v1/admin/projects/"+pid+"/joins", nil, domain.RolePlatformAdmin)
	if w6.Code != http.StatusOK {
		t.Fatalf("admin joins: %d %s", w6.Code, w6.Body.String())
	}
	var list struct {
		Total int `json:"total"`
	}
	unwrapData(t, w6.Body.Bytes(), &list)
	if list.Total != 1 {
		t.Fatalf("admin joins total should be 1: %s", w6.Body.String())
	}
	w7 := request(t, app, http.MethodGet, "/api/v1/admin/projects/"+pid+"/joins", nil, domain.RoleIndividual)
	if w7.Code != http.StatusForbidden {
		t.Fatalf("individual admin joins: expected 403, got %d", w7.Code)
	}

	// 状态流转：已对接 200；非法状态 400
	w8 := request(t, app, http.MethodPost, "/api/v1/admin/projects/"+pid+"/joins/"+join.ID+"/status",
		[]byte(`{"status":"contacted"}`), domain.RolePlatformAdmin)
	if w8.Code != http.StatusOK {
		t.Fatalf("update status: %d %s", w8.Code, w8.Body.String())
	}
	var updated struct {
		Status string `json:"status"`
	}
	unwrapData(t, w8.Body.Bytes(), &updated)
	if updated.Status != "contacted" {
		t.Fatalf("status should be contacted: %s", w8.Body.String())
	}
	w9 := request(t, app, http.MethodPost, "/api/v1/admin/projects/"+pid+"/joins/"+join.ID+"/status",
		[]byte(`{"status":"bogus"}`), domain.RolePlatformAdmin)
	if w9.Code != http.StatusBadRequest {
		t.Fatalf("invalid status: expected 400, got %d", w9.Code)
	}
}

// TestProjectJoinClosedRenevable closed 状态的申请可重新提交（同一记录重置 pending）。
func TestProjectJoinClosedRenevable(t *testing.T) {
	app := newBizServer(t)
	pid := createTestProject(t, app)

	w := request(t, app, http.MethodPost, "/api/v1/projects/"+pid+"/join",
		[]byte(`{"org_name":"天航科技","message":"参与"}`), domain.RoleIndividual)
	if w.Code != http.StatusCreated {
		t.Fatalf("join: %d %s", w.Code, w.Body.String())
	}
	var join struct {
		ID string `json:"id"`
	}
	unwrapData(t, w.Body.Bytes(), &join)

	if wc := request(t, app, http.MethodPost, "/api/v1/admin/projects/"+pid+"/joins/"+join.ID+"/status",
		[]byte(`{"status":"closed"}`), domain.RolePlatformAdmin); wc.Code != http.StatusOK {
		t.Fatalf("close: %d %s", wc.Code, wc.Body.String())
	}

	w2 := request(t, app, http.MethodPost, "/api/v1/projects/"+pid+"/join",
		[]byte(`{"org_name":"天航科技2","message":"重新参与"}`), domain.RoleIndividual)
	if w2.Code != http.StatusOK {
		t.Fatalf("renew join: expected 200, got %d %s", w2.Code, w2.Body.String())
	}
	var renewed struct {
		ID      string `json:"id"`
		Status  string `json:"status"`
		OrgName string `json:"org_name"`
	}
	unwrapData(t, w2.Body.Bytes(), &renewed)
	if renewed.ID != join.ID || renewed.Status != "pending" || renewed.OrgName != "天航科技2" {
		t.Fatalf("renewed should reuse record with pending status: %+v", renewed)
	}
}
