package httpapi_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// 密码登录失败锁定回归：连续 10 次密码错误后账号锁定（429），
// 正确密码也拒绝；此前无账号维度限制，分布式爆破可绕过 IP 限频。
func TestPasswordLoginLockout(t *testing.T) {
	old := os.Getenv("ADMIN_DEV_MODE")
	os.Setenv("ADMIN_DEV_MODE", "true")
	t.Cleanup(func() { os.Setenv("ADMIN_DEV_MODE", old) })

	app := newServer(t)

	// 注册一个密码账号（H5 兼容层注册 → 密码登录）
	reg := httptest.NewRecorder()
	app.ServeHTTP(reg, httptest.NewRequest(http.MethodPost, "/api/auth/register",
		strings.NewReader(`{"phone":"13800001234","password":"Secret123"}`)))
	if reg.Code != http.StatusCreated && reg.Code != http.StatusOK {
		t.Fatalf("register: %d %s", reg.Code, reg.Body.String())
	}

	// 前 9 次错误密码 → 401
	for i := 0; i < 9; i++ {
		w := httptest.NewRecorder()
		app.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/auth/login",
			strings.NewReader(`{"phone":"13800001234","password":"WrongPass"}`)))
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("wrong attempt %d: want 401, got %d", i+1, w.Code)
		}
	}

	// 第 10 次错误 → 401（触发锁定）
	w := httptest.NewRecorder()
	app.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/auth/login",
		strings.NewReader(`{"phone":"13800001234","password":"WrongPass"}`)))
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("10th wrong attempt: want 401, got %d", w.Code)
	}

	// 锁定后：即使正确密码也 429
	w = httptest.NewRecorder()
	app.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/auth/login",
		strings.NewReader(`{"phone":"13800001234","password":"Secret123"}`)))
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("locked account with correct password: want 429, got %d %s", w.Code, w.Body.String())
	}
}

// 账号不存在与密码错误的响应时间差防护：dummy bcrypt 使两者耗时接近。
// 此处只验证行为（均 401、统一文案），时序由实现保证。
func TestPasswordLoginUnknownUserUniform(t *testing.T) {
	app := newServer(t)
	for _, phone := range []string{"13900009998", "13900009997"} {
		w := httptest.NewRecorder()
		app.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/auth/login",
			strings.NewReader(`{"phone":"`+phone+`","password":"whatever"}`)))
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("login %s: want 401, got %d", phone, w.Code)
		}
		if !strings.Contains(w.Body.String(), "账号或密码错误") {
			t.Fatalf("login %s: message must not leak account existence, got: %s", phone, w.Body.String())
		}
	}
}
