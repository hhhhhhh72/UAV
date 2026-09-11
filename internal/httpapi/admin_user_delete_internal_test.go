package httpapi

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
	"drone-platform/internal/service"
)

// stubUsers 只实现本组测试触发的方法（其余方法嵌入接口，调用即 panic）。
// FindByID 必须实现：账号处置会先查目标是否超级管理员（按 SUPER_ADMIN_PHONE 判定）。
type stubUsers struct {
	repository.UserRepository
	softErr error
	found   *domain.User
}

func (s stubUsers) SoftDelete(ctx context.Context, id string) error { return s.softErr }

func (s stubUsers) FindByID(ctx context.Context, id string) (domain.User, error) {
	if s.found != nil {
		return *s.found, nil
	}
	return domain.User{}, fmt.Errorf("user %s: %w", id, repository.ErrUserNotFound)
}

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
	// 超级管理员（SUPER_ADMIN_PHONE 指定的账号）受保护 → 403
	superStub := stubUsers{found: &domain.User{ID: "user-19800000000", Role: domain.RolePlatformAdmin}}
	srv3 := &Server{userSvc: service.NewUserService(superStub, service.WithSuperAdminPhone("19800000000"))}
	if w := deleteUserReq(srv3, "user-19800000000", domain.Actor{ID: "admin-1", Role: domain.RolePlatformAdmin}); w.Code != http.StatusForbidden {
		t.Fatalf("超级管理员应 403，实际 %d %s", w.Code, w.Body.String())
	}
	// 不能删自己（自锁保护）→ 403
	actor := domain.Actor{ID: "admin-1", Role: domain.RolePlatformAdmin}
	if w := deleteUserReq(srv2, "admin-1", actor); w.Code != http.StatusForbidden {
		t.Fatalf("删自己应 403，实际 %d %s", w.Code, w.Body.String())
	}
}
