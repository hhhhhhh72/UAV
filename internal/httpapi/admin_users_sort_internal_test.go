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

// 超级管理员必须排在列表最前面（其余账号仍按仓储给出的注册时间顺序）。
// 场景：超管账号往往建得最晚，按时间排序会落在中间，运营每次都要往下翻。
func TestListUsersPutsSuperAdminFirst(t *testing.T) {
	repo := memory.NewUserRepository(nil)
	now := time.Now()
	seed := []domain.User{
		{ID: "user-13800000001", Name: "个人一", Role: domain.RoleIndividual, Status: "active", CreatedAt: now.Add(-72 * time.Hour)},
		{ID: "user-13800000002", Name: "个人二", Role: domain.RoleIndividual, Status: "active", CreatedAt: now.Add(-48 * time.Hour)},
		{ID: "user-19800000000", Name: "超级管理员", Role: domain.RolePlatformAdmin, Status: "active", CreatedAt: now.Add(-24 * time.Hour)},
	}
	for i, u := range seed {
		if _, err := repo.Create(context.Background(), u); err != nil {
			t.Fatalf("seed %d: %v", i, err)
		}
	}

	srv := &Server{
		userRepo:        repo,
		userSvc:         service.NewUserService(repo),
		superAdminPhone: "19800000000",
	}
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
	if len(resp.Data) != 3 {
		t.Fatalf("应返回 3 行，实际 %d", len(resp.Data))
	}
	if id, _ := resp.Data[0]["id"].(string); id != "user-19800000000" {
		t.Fatalf("第一行应是超级管理员，实际 %v", resp.Data[0]["id"])
	}
	if v, _ := resp.Data[0]["is_super_admin"].(bool); !v {
		t.Fatalf("第一行应带 is_super_admin 标记：%v", resp.Data[0])
	}
	// 其余行的相对顺序必须与仓储返回的一致（稳定排序，只把超管提到最前）
	raw, err := repo.AllWithDeleted(context.Background())
	if err != nil {
		t.Fatalf("AllWithDeleted: %v", err)
	}
	wantRest := make([]string, 0, len(raw))
	for _, u := range raw {
		if u.ID == "user-19800000000" {
			continue
		}
		wantRest = append(wantRest, u.ID)
	}
	gotRest := make([]string, 0, len(resp.Data)-1)
	for _, row := range resp.Data[1:] {
		id, _ := row["id"].(string)
		gotRest = append(gotRest, id)
	}
	if strings.Join(gotRest, ",") != strings.Join(wantRest, ",") {
		t.Fatalf("其余行顺序应保持仓储顺序：want %v got %v", wantRest, gotRest)
	}
}
