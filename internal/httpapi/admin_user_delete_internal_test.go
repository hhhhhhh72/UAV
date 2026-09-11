package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
	"drone-platform/internal/service"
)

// stubUsers 只实现本组测试触发的方法（其余方法嵌入接口，调用即 panic）。
type stubUsers struct {
	repository.UserRepository
	softErr error
}

func (s stubUsers) SoftDelete(ctx context.Context, id string) error { return s.softErr }

func deleteUserReq(srv *Server, id string, a domain.Actor) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/admin/users/"+id, nil)
	req.SetPathValue("id", id) // 直调 Handler 不经过 ServeMux，路径变量需手动注入
	req = req.WithContext(contextWithActor(req, a))
	w := httptest.NewRecorder()
	srv.deleteUser(w, req)
	return w
}

// 账号不存在 → 404；非平台管理员 → 403（Handler 的状态码映射）。
func TestDeleteUserNotFoundAndForbidden(t *testing.T) {
	srv := &Server{userSvc: service.NewUserService(stubUsers{softErr: repository.ErrUserNotFound})}
	if w := deleteUserReq(srv, "ghost", domain.Actor{ID: "admin-1", Role: domain.RolePlatformAdmin}); w.Code != http.StatusNotFound {
		t.Fatalf("不存在应 404，实际 %d %s", w.Code, w.Body.String())
	}
	srv2 := &Server{userSvc: service.NewUserService(stubUsers{})}
	if w := deleteUserReq(srv2, "user-1", domain.Actor{ID: "ent-1", Role: domain.RoleEnterprise}); w.Code != http.StatusForbidden {
		t.Fatalf("非管理员应 403，实际 %d %s", w.Code, w.Body.String())
	}
	if w := deleteUserReq(srv2, "admin", domain.Actor{ID: "admin-1", Role: domain.RolePlatformAdmin}); w.Code != http.StatusForbidden {
		t.Fatalf("内置超管应 403，实际 %d %s", w.Code, w.Body.String())
	}
}
