package httpapi_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/httpapi"
)

// authAs 签发指定用户/角色的 token（与 server_test.go 的 auth 相比支持自定义 ID）。
func authAs(t *testing.T, id string, role domain.Role) string {
	t.Helper()
	tokens, err := httpapi.NewTokenManager(testSecret)
	if err != nil {
		t.Fatal(err)
	}
	token, err := tokens.Issue(domain.Actor{ID: id, Role: role}, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	return "Bearer " + token
}

// doRaw 发送请求；token 为空表示匿名（无 Authorization 头）。
func doRaw(app http.Handler, method, path, body, token string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, path, bytes.NewReader([]byte(body)))
	if token != "" {
		r.Header.Set("Authorization", token)
	}
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	return w
}

// TestBatch2WritesRequireAuth: 回归 C2——非 /admin/ 前缀的四个写操作曾零鉴权：
//   - POST /api/v1/transformations/{id}/advance      匿名 401，非负责人 403，负责人/管理员 200
//   - POST /api/v1/transformations/{id}/milestones   同上
//   - POST /api/v1/cooperation-programs              仅管理员
//   - POST /api/v1/cooperation-programs/{id}/status  仅管理员
func TestBatch2WritesRequireAuth(t *testing.T) {
	app := newServer(t)
	adminTok := authAs(t, "admin-1", domain.RolePlatformAdmin)
	// 创建者即转化负责人（OwnerID=admin-1）：管理员建记录，负责人（同 ID）与管理员可推进，
	// 非负责人（user-2）403——authenticate 以库中角色为准，同一 ID 只能有一种角色。
	ownerTok := authAs(t, "admin-1", domain.RolePlatformAdmin)
	indivTok := authAs(t, "user-1", domain.RoleIndividual)
	strangerTok := authAs(t, "user-2", domain.RoleIndividual)

	// Arrange: 管理员建一条转化记录（OwnerID=admin-1）
	w := doRaw(app, http.MethodPost, "/api/v1/admin/transformations",
		`{"title":"成果A中试","achievement_id":"ach-1","partner_id":"ent-1"}`, adminTok)
	if w.Code != http.StatusCreated {
		t.Fatalf("create transformation: %d %s", w.Code, w.Body.String())
	}
	var created struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatalf("parse transformation: %v", err)
	}
	if created.Data.ID == "" {
		t.Fatalf("created transformation missing id: %s", w.Body.String())
	}

	cases := []struct {
		name     string
		method   string
		path     string
		body     string
		token    string
		wantCode int
	}{
		{"advance anonymous", http.MethodPost, "/api/v1/transformations/" + created.Data.ID + "/advance", `{"stage":"pilot"}`, "", http.StatusUnauthorized},
		{"advance non-owner", http.MethodPost, "/api/v1/transformations/" + created.Data.ID + "/advance", `{"stage":"pilot"}`, strangerTok, http.StatusForbidden},
		{"advance owner", http.MethodPost, "/api/v1/transformations/" + created.Data.ID + "/advance", `{"stage":"pilot","progress":"50%"}`, ownerTok, http.StatusOK},
		{"milestone anonymous", http.MethodPost, "/api/v1/transformations/" + created.Data.ID + "/milestones", `{"name":"样机"}`, "", http.StatusUnauthorized},
		{"milestone non-owner", http.MethodPost, "/api/v1/transformations/" + created.Data.ID + "/milestones", `{"name":"样机"}`, strangerTok, http.StatusForbidden},
		{"milestone admin", http.MethodPost, "/api/v1/transformations/" + created.Data.ID + "/milestones", `{"name":"样机"}`, adminTok, http.StatusOK},
		// 匿名请求在 authenticate 中间件即被 401 拦截（adminGate 同款行为），
		// 已登录非管理员在 handler 层被 403
		{"cooperation create anonymous", http.MethodPost, "/api/v1/cooperation-programs", `{"title":"共建"}`, "", http.StatusUnauthorized},
		{"cooperation create individual", http.MethodPost, "/api/v1/cooperation-programs", `{"title":"共建"}`, indivTok, http.StatusForbidden},
		{"cooperation create admin", http.MethodPost, "/api/v1/cooperation-programs", `{"title":"共建","college_id":"c1"}`, adminTok, http.StatusCreated},
		{"cooperation status anonymous", http.MethodPost, "/api/v1/cooperation-programs/x/status", `{"status":"active"}`, "", http.StatusUnauthorized},
		{"cooperation status individual", http.MethodPost, "/api/v1/cooperation-programs/x/status", `{"status":"active"}`, indivTok, http.StatusForbidden},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := doRaw(app, tc.method, tc.path, tc.body, tc.token)
			if w.Code != tc.wantCode {
				t.Fatalf("code: expected %d, got %d (%s)", tc.wantCode, w.Code, w.Body.String())
			}
		})
	}
}

// TestBatch2TransformationListMasksContact 安全回归：
// 公开列表匿名访问必须脱敏 contact_info（防爬取 PII），已登录/负责人可见完整值。
func TestBatch2TransformationListMasksContact(t *testing.T) {
	app := newServer(t)
	adminTok := authAs(t, "admin-1", domain.RolePlatformAdmin)
	ownerTok := authAs(t, "admin-1", domain.RolePlatformAdmin)

	w := doRaw(app, http.MethodPost, "/api/v1/admin/transformations",
		`{"title":"成果B中试","achievement_id":"ach-9","partner_id":"ent-9","contact_info":"13812345678"}`, adminTok)
	if w.Code != http.StatusCreated {
		t.Fatalf("create transformation: %d %s", w.Code, w.Body.String())
	}

	// 匿名：脱敏（保留前3后4），绝不出现完整号码
	an := doRaw(app, http.MethodGet, "/api/v1/transformations", "", "")
	if an.Code != http.StatusOK {
		t.Fatalf("anonymous list: %d %s", an.Code, an.Body.String())
	}
	if bytes.Contains(an.Body.Bytes(), []byte("13812345678")) {
		t.Fatalf("anonymous list leaked full contact: %s", an.Body.String())
	}
	if !bytes.Contains(an.Body.Bytes(), []byte("138****5678")) {
		t.Fatalf("anonymous list should contain masked contact: %s", an.Body.String())
	}

	// 已登录负责人：完整值可见
	au := doRaw(app, http.MethodGet, "/api/v1/transformations", "", ownerTok)
	if au.Code != http.StatusOK {
		t.Fatalf("owner list: %d", au.Code)
	}
	if !bytes.Contains(au.Body.Bytes(), []byte("13812345678")) {
		t.Fatalf("owner list should contain full contact: %s", au.Body.String())
	}
}
