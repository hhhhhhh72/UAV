package httpapi_test

import (
	"net/http"
	"testing"

	"drone-platform/internal/domain"
)

// 账号密码接口：
//  1. POST /api/v1/auth/password 本人改密：旧密码校验 + 成功后旧令牌立即失效；
//  2. POST /api/v1/admin/users/{id}/password 仅平台管理员可重置；
//  3. 弱口令一律 400。
func TestPasswordEndpoints(t *testing.T) {
	app := newBizServer(t)

	// 平台管理员建一个带密码的账号（此前只有建号时能设一次密码）
	w := requestAs(t, app, http.MethodPost, "/api/v1/admin/users",
		[]byte(`{"phone":"13600000001","role":"individual","password":"InitPass123"}`), "admin-1", domain.RolePlatformAdmin)
	if w.Code != http.StatusCreated {
		t.Fatalf("create user: %d %s", w.Code, w.Body.String())
	}

	// 旧密码错 → 400
	bad := requestAs(t, app, http.MethodPost, "/api/v1/auth/password",
		[]byte(`{"old_password":"wrong-old","new_password":"NewPass456"}`), "user-13600000001", domain.RoleIndividual)
	if bad.Code != http.StatusBadRequest {
		t.Fatalf("旧密码错误应 400，实际 %d %s", bad.Code, bad.Body.String())
	}
	// 弱口令 → 400
	weak := requestAs(t, app, http.MethodPost, "/api/v1/auth/password",
		[]byte(`{"old_password":"InitPass123","new_password":"short"}`), "user-13600000001", domain.RoleIndividual)
	if weak.Code != http.StatusBadRequest {
		t.Fatalf("弱口令应 400，实际 %d %s", weak.Code, weak.Body.String())
	}
	// 正常改密 → 200
	ok := requestAs(t, app, http.MethodPost, "/api/v1/auth/password",
		[]byte(`{"old_password":"InitPass123","new_password":"NewPass456"}`), "user-13600000001", domain.RoleIndividual)
	if ok.Code != http.StatusOK {
		t.Fatalf("改密应 200，实际 %d %s", ok.Code, ok.Body.String())
	}
	// 改密后旧令牌失效（token_version 自增）→ /me 401
	me := requestAs(t, app, http.MethodGet, "/api/v1/me", nil, "user-13600000001", domain.RoleIndividual)
	if me.Code != http.StatusUnauthorized {
		t.Fatalf("改密后旧令牌应失效，实际 %d %s", me.Code, me.Body.String())
	}

	// 未登录 → 401
	if anon := doRaw(app, http.MethodPost, "/api/v1/auth/password", `{"old_password":"x","new_password":"NewPass456"}`, ""); anon.Code != http.StatusUnauthorized {
		t.Fatalf("未登录应 401，实际 %d", anon.Code)
	}

	// 管理员重置：协会管理员 → 403；平台管理员 → 200；弱口令 → 400；不存在 → 404
	if forbidden := requestAs(t, app, http.MethodPost, "/api/v1/admin/users/user-13600000001/password",
		[]byte(`{"password":"ResetPass789"}`), "admin-2", domain.RoleAssociationAdmin); forbidden.Code != http.StatusForbidden {
		t.Fatalf("协会管理员重置应 403，实际 %d %s", forbidden.Code, forbidden.Body.String())
	}
	if rw := requestAs(t, app, http.MethodPost, "/api/v1/admin/users/user-13600000001/password",
		[]byte(`{"password":"ResetPass789"}`), "admin-1", domain.RolePlatformAdmin); rw.Code != http.StatusOK {
		t.Fatalf("平台管理员重置应 200，实际 %d %s", rw.Code, rw.Body.String())
	}
	if rweak := requestAs(t, app, http.MethodPost, "/api/v1/admin/users/user-13600000001/password",
		[]byte(`{"password":"123"}`), "admin-1", domain.RolePlatformAdmin); rweak.Code != http.StatusBadRequest {
		t.Fatalf("重置弱口令应 400，实际 %d", rweak.Code)
	}
	if rmiss := requestAs(t, app, http.MethodPost, "/api/v1/admin/users/ghost-user/password",
		[]byte(`{"password":"ResetPass789"}`), "admin-1", domain.RolePlatformAdmin); rmiss.Code != http.StatusNotFound {
		t.Fatalf("重置不存在账号应 404，实际 %d", rmiss.Code)
	}
}
