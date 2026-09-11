package httpapi_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"drone-platform/internal/domain"
)

// 后台删除用户的语义（生产口径）：
//  1. 唯一动作"删除"，不可恢复：账号立即失效（已签发 token 401）、从列表/平台消失；
//  2. 账号行保留 7 天缓冲期，返回 purge_after；到期由后台任务自动清除；
//  3. 其发布的内容一律不删（46 张业务表以文本列记用户 ID，无外键级联）；
//  4. 内置超管不可删；非平台管理员 403；未登录 401；不存在 404。
func TestAdminDeleteUserSemantics(t *testing.T) {
	app := newBizServer(t)

	w := requestAs(t, app, http.MethodDelete, "/api/v1/admin/users/worker-2", nil, "admin-1", domain.RolePlatformAdmin)
	if w.Code != http.StatusOK {
		t.Fatalf("delete: %d %s", w.Code, w.Body.String())
	}
	var res struct {
		Data struct {
			ID           string `json:"id"`
			Mode         string `json:"mode"`
			Status       string `json:"status"`
			PurgeAfter   string `json:"purge_after"`
			KeptContents bool   `json:"kept_contents"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("parse: %v (body=%s)", err, w.Body.String())
	}
	if res.Data.Mode != "deleted" || res.Data.Status != domain.UserDeleted || !res.Data.KeptContents || res.Data.PurgeAfter == "" {
		t.Fatalf("删除响应不符：%+v", res.Data)
	}

	// 删除后令牌立即失效
	after := requestAs(t, app, http.MethodGet, "/api/v1/me", nil, "worker-2", domain.RoleIndividual)
	if after.Code != http.StatusUnauthorized {
		t.Fatalf("删除后应 401，实际 %d %s", after.Code, after.Body.String())
	}

	// 缓冲期内账号行仍在（内容作者可解析），列表可见并带自动清除日期
	rec := userRecord(t, app, "worker-2")
	if rec.Status != domain.UserDeleted {
		t.Fatalf("缓冲期内状态应为 deleted，实际 %q", rec.Status)
	}
	if rec.PurgeAfter == "" {
		t.Fatal("列表应给出 purge_after（自动清除日期）")
	}

	// 没有恢复入口：该路由不存在
	if rw := requestAs(t, app, http.MethodPost, "/api/v1/admin/users/worker-2/restore", nil, "admin-1", domain.RolePlatformAdmin); rw.Code != http.StatusNotFound {
		t.Fatalf("不应存在恢复路由，实际 %d", rw.Code)
	}

	// 保护与权限
	if w := requestAs(t, app, http.MethodDelete, "/api/v1/admin/users/admin", nil, "admin-1", domain.RolePlatformAdmin); w.Code != http.StatusForbidden {
		t.Fatalf("内置超管应 403，实际 %d %s", w.Code, w.Body.String())
	}
	if w := requestAs(t, app, http.MethodDelete, "/api/v1/admin/users/worker-4", nil, "admin-2", domain.RoleAssociationAdmin); w.Code != http.StatusForbidden {
		t.Fatalf("非平台管理员应 403，实际 %d %s", w.Code, w.Body.String())
	}
	if w := doRaw(app, http.MethodDelete, "/api/v1/admin/users/worker-4", "", ""); w.Code != http.StatusUnauthorized {
		t.Fatalf("未登录应 401，实际 %d", w.Code)
	}
	if w := requestAs(t, app, http.MethodDelete, "/api/v1/admin/users/ghost-user", nil, "admin-1", domain.RolePlatformAdmin); w.Code != http.StatusNotFound {
		t.Fatalf("不存在应 404，实际 %d", w.Code)
	}
	if rec := userRecord(t, app, "worker-4"); rec.Status != domain.UserActive {
		t.Fatalf("被拒路径改动了账号状态：%q", rec.Status)
	}
}

type userRec struct {
	ID         string `json:"id"`
	Status     string `json:"status"`
	PurgeAfter string `json:"purge_after"`
	DeletedAt  string `json:"deleted_at"`
}

// userRecord 从管理端用户列表里取指定账号（不存在时返回零值）。
func userRecord(t *testing.T, app http.Handler, id string) userRec {
	t.Helper()
	w := requestAs(t, app, http.MethodGet, "/api/v1/admin/users?page_size=200", nil, "admin-1", domain.RolePlatformAdmin)
	if w.Code != http.StatusOK {
		t.Fatalf("list users: %d %s", w.Code, w.Body.String())
	}
	var resp struct {
		Data []userRec `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse users: %v", err)
	}
	for _, u := range resp.Data {
		if u.ID == id {
			return u
		}
	}
	return userRec{}
}
