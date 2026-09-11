package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

// 列表置顶规则：超管第 1 行 → 管理员按「成为管理员的时间」升序（新设的排在上一个置顶的
// 下面）→ 普通账号保持仓储顺序。
//
// 「成为管理员的时间」由 UpdatedAt 代理：改角色会刷新它，所以越晚设为管理员、排序越靠下。
func TestListUsersPinsSuperAdminThenAdminsInOrder(t *testing.T) {
	repo := memory.NewUserRepository(nil)
	now := time.Now()
	seed := []domain.User{
		// 普通账号：注册顺序 01 → 02
		{ID: "user-13800000001", Name: "个人一", Role: domain.RoleIndividual, Status: "active", CreatedAt: now.Add(-100 * time.Hour), UpdatedAt: now.Add(-100 * time.Hour)},
		{ID: "user-13800000002", Name: "个人二", Role: domain.RoleIndividual, Status: "active", CreatedAt: now.Add(-90 * time.Hour), UpdatedAt: now.Add(-90 * time.Hour)},
		// 管理员 A：先设的（UpdatedAt 更早）→ 应排在管理员 B 上面
		{ID: "user-13900000001", Name: "管理员A", Role: domain.RoleAssociationAdmin, Status: "active", CreatedAt: now.Add(-80 * time.Hour), UpdatedAt: now.Add(-30 * time.Hour)},
		// 管理员 B：后设的 → 应紧跟在 A 下面
		{ID: "user-13900000002", Name: "管理员B", Role: domain.RolePlatformAdmin, Status: "active", CreatedAt: now.Add(-70 * time.Hour), UpdatedAt: now.Add(-10 * time.Hour)},
		// 超级管理员：UpdatedAt 最新（刚被改过资料），仍必须第 1 行
		{ID: "user-19800000000", Name: "超级管理员", Role: domain.RolePlatformAdmin, Status: "active", CreatedAt: now.Add(-60 * time.Hour), UpdatedAt: now},
	}
	for i, u := range seed {
		if _, err := repo.Create(context.Background(), u); err != nil {
			t.Fatalf("seed %d: %v", i, err)
		}
	}

	srv := &Server{userRepo: repo, userSvc: service.NewUserService(repo), superAdminPhone: "19800000000"}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
	req = req.WithContext(contextWithActor(req, domain.Actor{ID: "admin-1", Role: domain.RolePlatformAdmin}))
	w := httptest.NewRecorder()
	srv.listUsers(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("listUsers: %d %s", w.Code, w.Body.String())
	}
	var resp struct {
		Data []map[string]any `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse: %v", err)
	}
	got := make([]string, 0, len(resp.Data))
	for _, row := range resp.Data {
		id, _ := row["id"].(string)
		got = append(got, id)
	}
	// 期望顺序 = 超管 + 管理员（按 UpdatedAt 升序）+ 普通账号（保持仓储原始顺序）
	raw, err := repo.AllWithDeleted(context.Background())
	if err != nil {
		t.Fatalf("AllWithDeleted: %v", err)
	}
	want := []string{"user-19800000000", "user-13900000001", "user-13900000002"}
	for _, u := range raw {
		if u.ID == "user-19800000000" || isAdminRoleTest(u.Role) {
			continue
		}
		want = append(want, u.ID)
	}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("置顶顺序不对\n  want %v\n  got  %v", want, got)
	}
	if v, _ := resp.Data[0]["is_super_admin"].(bool); !v {
		t.Fatalf("第一行应带 is_super_admin 标记：%v", resp.Data[0])
	}
}

func isAdminRoleTest(r domain.Role) bool {
	return r == domain.RolePlatformAdmin || r == domain.RoleAssociationAdmin
}
