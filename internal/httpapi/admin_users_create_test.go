package httpapi_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"drone-platform/internal/domain"
)

// 管理员建号契约：登录名 = 手机号（必填）。
//  1. 手机号缺失/格式非法 → 400；弱口令 → 400；号码已存在 → 409；
//  2. 建号成功：id = user-<手机号>，响应只回脱敏手机号，不泄露明文/密码哈希；
//  3. 闭环：新账号可直接用「手机号 + 初始密码」走 /api/auth/login 登录；
//  4. 昵称留空时自动生成默认昵称（不产生空名账号）。
func TestCreateUserByPhone(t *testing.T) {
	app := newBizServer(t)

	// 手机号缺失 → 400
	missing := requestAs(t, app, http.MethodPost, "/api/v1/admin/users",
		[]byte(`{"role":"individual","password":"InitPass123"}`), "admin-1", domain.RolePlatformAdmin)
	if missing.Code != http.StatusBadRequest {
		t.Fatalf("缺少手机号应 400，实际 %d %s", missing.Code, missing.Body.String())
	}

	// 手机号格式非法 → 400
	for _, bad := range []string{"12345", "23800000001", "1380000000", "138000000001"} {
		w := requestAs(t, app, http.MethodPost, "/api/v1/admin/users",
			[]byte(`{"phone":"`+bad+`","role":"individual","password":"InitPass123"}`), "admin-1", domain.RolePlatformAdmin)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("手机号 %q 应 400，实际 %d %s", bad, w.Code, w.Body.String())
		}
	}

	// 弱口令 → 400
	weak := requestAs(t, app, http.MethodPost, "/api/v1/admin/users",
		[]byte(`{"phone":"13500000001","role":"individual","password":"short"}`), "admin-1", domain.RolePlatformAdmin)
	if weak.Code != http.StatusBadRequest {
		t.Fatalf("弱口令应 400，实际 %d %s", weak.Code, weak.Body.String())
	}

	// 正常建号（昵称留空 → 默认昵称）
	created := requestAs(t, app, http.MethodPost, "/api/v1/admin/users",
		[]byte(`{"phone":"13500000001","role":"individual","password":"InitPass123"}`), "admin-1", domain.RolePlatformAdmin)
	if created.Code != http.StatusCreated {
		t.Fatalf("建号应 201，实际 %d %s", created.Code, created.Body.String())
	}
	var resp struct {
		Data struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			PhoneMasked string `json:"phone_masked"`
			Password    string `json:"password"`
			PasswordHash string `json:"password_hash"`
		} `json:"data"`
	}
	if err := json.Unmarshal(created.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse create resp: %v", err)
	}
	if resp.Data.ID != "user-13500000001" {
		t.Fatalf("id 应为 user-13500000001，实际 %q", resp.Data.ID)
	}
	if resp.Data.PhoneMasked != "135****0001" {
		t.Fatalf("应只回脱敏手机号，实际 %q", resp.Data.PhoneMasked)
	}
	if resp.Data.Name == "" {
		t.Fatalf("昵称留空应生成默认昵称，实际为空")
	}
	if resp.Data.Password != "" || resp.Data.PasswordHash != "" {
		t.Fatalf("响应不得回传口令/哈希")
	}

	// 号码重复 → 409
	dup := requestAs(t, app, http.MethodPost, "/api/v1/admin/users",
		[]byte(`{"phone":"13500000001","role":"enterprise","password":"InitPass123"}`), "admin-1", domain.RolePlatformAdmin)
	if dup.Code != http.StatusConflict {
		t.Fatalf("重复号码应 409，实际 %d %s", dup.Code, dup.Body.String())
	}

	// 登录名 = 手机号：新账号可直接登录（端到端闭环）
	login := doRaw(app, http.MethodPost, "/api/auth/login", `{"phone":"13500000001","password":"InitPass123"}`, "")
	if login.Code != http.StatusOK {
		t.Fatalf("手机号+初始密码登录应 200，实际 %d %s", login.Code, login.Body.String())
	}
	var lr struct {
		Data struct {
			AccessToken string `json:"accessToken"`
			User        struct {
				ID   string `json:"id"`
				Role string `json:"role"`
			} `json:"user"`
		} `json:"data"`
	}
	if err := json.Unmarshal(login.Body.Bytes(), &lr); err != nil {
		t.Fatalf("parse login resp: %v", err)
	}
	if lr.Data.AccessToken == "" || lr.Data.User.ID != "user-13500000001" {
		t.Fatalf("登录应签发令牌且身份为新建账号，实际 %s %s", lr.Data.AccessToken, lr.Data.User.ID)
	}

	// 协会管理员仍不得建管理员账号（防提权回归）
	esc := requestAs(t, app, http.MethodPost, "/api/v1/admin/users",
		[]byte(`{"phone":"13500000002","role":"platform_admin","password":"InitPass123"}`), "admin-2", domain.RoleAssociationAdmin)
	if esc.Code != http.StatusForbidden {
		t.Fatalf("协会管理员建平台管理员应 403，实际 %d %s", esc.Code, esc.Body.String())
	}
}
