package httpapi_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"drone-platform/internal/domain"
)

// TestAdminUsersPagination: 管理端用户列表必须支持分页（page/page_size + total）。
// 回归：listUsers 曾忽略分页参数全量返回，导致前端翻页失效。
func TestAdminUsersPagination(t *testing.T) {
	app := newServer(t)

	// Arrange: 创建 5 个用户（加上内置超级管理员共 6 条）
	for i := 0; i < 5; i++ {
		body := []byte(fmt.Sprintf(`{"id":"user-p%d","role":"individual"}`, i))
		w := request(t, app, http.MethodPost, "/api/v1/admin/users", body, domain.RolePlatformAdmin)
		if w.Code != http.StatusCreated {
			t.Fatalf("create user-p%d: %d %s", i, w.Code, w.Body.String())
		}
	}

	// Act: 第 1 页，每页 2 条
	w := request(t, app, http.MethodGet, "/api/v1/admin/users?page=1&page_size=2", nil, domain.RolePlatformAdmin)
	if w.Code != http.StatusOK {
		t.Fatalf("list users: %d %s", w.Code, w.Body.String())
	}
	var page1 struct {
		Data  []map[string]any `json:"data"`
		Total int              `json:"total"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &page1); err != nil {
		t.Fatalf("parse page1: %v", err)
	}

	// Assert: 总数 6（5 个新建 + 内置 admin），第 1 页 2 条
	if page1.Total != 6 {
		t.Fatalf("total: expected 6, got %d", page1.Total)
	}
	if len(page1.Data) != 2 {
		t.Fatalf("page1 items: expected 2, got %d", len(page1.Data))
	}

	// Act: 第 2 页 2 条 → 第 3 页应是剩余 2 条
	w = request(t, app, http.MethodGet, "/api/v1/admin/users?page=3&page_size=2", nil, domain.RolePlatformAdmin)
	if w.Code != http.StatusOK {
		t.Fatalf("list users page3: %d", w.Code)
	}
	var page3 struct {
		Data []map[string]any `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &page3); err != nil {
		t.Fatalf("parse page3: %v", err)
	}
	if len(page3.Data) != 2 {
		t.Fatalf("page3 items: expected 2, got %d", len(page3.Data))
	}
}

// TestCreateUserRoleRestrictions: createUser 防提权回归（C1）。
//   - association_admin 不得创建 platform_admin / association_admin 账号
//   - 任意管理员可创建 individual / enterprise
//   - 非法角色字符串一律 400
func TestCreateUserRoleRestrictions(t *testing.T) {
	app := newServer(t)

	cases := []struct {
		name     string
		actor    domain.Role
		target   string
		wantCode int
	}{
		{"association admin cannot create platform_admin", domain.RoleAssociationAdmin, "platform_admin", http.StatusForbidden},
		{"association admin cannot create association_admin", domain.RoleAssociationAdmin, "association_admin", http.StatusForbidden},
		{"association admin can create individual", domain.RoleAssociationAdmin, "individual", http.StatusCreated},
		{"association admin can create enterprise", domain.RoleAssociationAdmin, "enterprise", http.StatusCreated},
		{"platform admin can create platform_admin", domain.RolePlatformAdmin, "platform_admin", http.StatusCreated},
		{"invalid role rejected", domain.RolePlatformAdmin, "superuser", http.StatusBadRequest},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body := []byte(fmt.Sprintf(`{"id":"u-%s","role":"%s"}`, tc.target, tc.target))
			w := request(t, app, http.MethodPost, "/api/v1/admin/users", body, tc.actor)
			if w.Code != tc.wantCode {
				t.Fatalf("code: expected %d, got %d (%s)", tc.wantCode, w.Code, w.Body.String())
			}
		})
	}
}
