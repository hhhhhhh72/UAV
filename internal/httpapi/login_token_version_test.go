package httpapi_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"drone-platform/internal/domain"
)

// 改过/重置过密码的账号（token_version > 0）用密码登录后，令牌必须立刻可用。
//
// 回归：h5 密码登录签发令牌时漏带 token_version（tv 恒为 0），而中间件每次请求都用库里的
// token_version 复验 → 这类账号表现为"登录成功、下一个请求就 401 账号已失效"，
// 管理后台会卡在"登录→被踢回登录页"的死循环（重置密码后尤其必现）。
func TestPasswordLoginTokenCarriesTokenVersion(t *testing.T) {
	app := newBizServer(t)
	adminTok := authAs(t, "admin-1", domain.RolePlatformAdmin)

	// 建号 → 重置一次密码（token_version 自增到 1）
	if w := doRaw(app, http.MethodPost, "/api/v1/admin/users",
		`{"phone":"13700009999","role":"individual","password":"InitPass123"}`, adminTok); w.Code != http.StatusCreated {
		t.Fatalf("建号: %d %s", w.Code, w.Body.String())
	}
	if w := doRaw(app, http.MethodPost, "/api/v1/admin/users/user-13700009999/password",
		`{"password":"ResetPass456"}`, adminTok); w.Code != http.StatusOK {
		t.Fatalf("重置密码: %d %s", w.Code, w.Body.String())
	}

	// 用新密码登录 → 拿令牌 → 打 /api/v1/me 必须 200
	w := doRaw(app, http.MethodPost, "/api/auth/login", `{"phone":"13700009999","password":"ResetPass456"}`, "")
	if w.Code != http.StatusOK {
		t.Fatalf("登录: %d %s", w.Code, w.Body.String())
	}
	var resp struct {
		Data struct {
			AccessToken string `json:"accessToken"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse login: %v", err)
	}
	if resp.Data.AccessToken == "" {
		t.Fatal("登录未返回 accessToken")
	}
	me := doRaw(app, http.MethodGet, "/api/v1/me", "", "Bearer "+resp.Data.AccessToken)
	if me.Code != http.StatusOK {
		t.Fatalf("改过密码的账号登录后 /me 应 200，实际 %d %s", me.Code, me.Body.String())
	}
}
