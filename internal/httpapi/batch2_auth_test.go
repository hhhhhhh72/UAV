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
	adminTok := authAs(t, "user-1", domain.RolePlatformAdmin)
	ownerTok := authAs(t, "user-1", domain.RoleIndividual)
	strangerTok := authAs(t, "user-2", domain.RoleIndividual)

	// Arrange: 管理员建一条转化记录（OwnerID=user-1）
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
		{"cooperation create individual", http.MethodPost, "/api/v1/cooperation-programs", `{"title":"共建"}`, ownerTok, http.StatusForbidden},
		{"cooperation create admin", http.MethodPost, "/api/v1/cooperation-programs", `{"title":"共建","college_id":"c1"}`, adminTok, http.StatusCreated},
		{"cooperation status anonymous", http.MethodPost, "/api/v1/cooperation-programs/x/status", `{"status":"active"}`, "", http.StatusUnauthorized},
		{"cooperation status individual", http.MethodPost, "/api/v1/cooperation-programs/x/status", `{"status":"active"}`, ownerTok, http.StatusForbidden},
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
