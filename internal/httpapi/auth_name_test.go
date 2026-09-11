package httpapi_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"drone-platform/internal/domain"
)

// 昵称必须来自数据库：/api/auth/me 与 /api/auth/login 都要返回 name。
//
// 回归：h5AuthMe 此前把 name 取自 legacy users.json 文件，后台建号 / 改名 / 提权的账号
// 不在那个文件里 → name 恒为空 → 小程序「我的」页只能退化成显示手机号。
func TestAuthMeAndLoginReturnName(t *testing.T) {
	app := newBizServer(t)
	adminTok := authAs(t, "admin-1", domain.RolePlatformAdmin)

	// 用后台建号（昵称只写库，不写 users.json），再重置密码拿到可登录口令
	if w := doRaw(app, http.MethodPost, "/api/v1/admin/users",
		`{"phone":"13700008888","role":"individual","password":"InitPass123","name":"张三"}`, adminTok); w.Code != http.StatusCreated {
		t.Fatalf("建号: %d %s", w.Code, w.Body.String())
	}

	// 1) 密码登录的响应里必须带 name（客户端首屏直接用这个）
	login := doRaw(app, http.MethodPost, "/api/auth/login", `{"phone":"13700008888","password":"InitPass123"}`, "")
	if login.Code != http.StatusOK {
		t.Fatalf("登录: %d %s", login.Code, login.Body.String())
	}
	var lr struct {
		Data struct {
			AccessToken string `json:"accessToken"`
			User        struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"user"`
		} `json:"data"`
	}
	if err := json.Unmarshal(login.Body.Bytes(), &lr); err != nil {
		t.Fatalf("parse login: %v", err)
	}
	if lr.Data.User.Name != "张三" {
		t.Fatalf("登录响应应带昵称，实际 name=%q（响应 %s）", lr.Data.User.Name, login.Body.String())
	}

	// 2) /api/auth/me 也必须带 name（小程序每次进「我的」都会拉它）
	me := doRaw(app, http.MethodGet, "/api/auth/me", "", "Bearer "+lr.Data.AccessToken)
	if me.Code != http.StatusOK {
		t.Fatalf("me: %d %s", me.Code, me.Body.String())
	}
	var mr struct {
		Data struct {
			User struct {
				Name  string `json:"name"`
				Phone string `json:"phone"`
			} `json:"user"`
		} `json:"data"`
	}
	if err := json.Unmarshal(me.Body.Bytes(), &mr); err != nil {
		t.Fatalf("parse me: %v", err)
	}
	if mr.Data.User.Name != "张三" {
		t.Fatalf("me 应返回库里的昵称，实际 name=%q（响应 %s）", mr.Data.User.Name, me.Body.String())
	}
	if mr.Data.User.Phone != "13700008888" {
		t.Fatalf("me 应返回手机号，实际 %q", mr.Data.User.Phone)
	}
}
