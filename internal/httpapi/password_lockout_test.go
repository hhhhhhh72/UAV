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

// 账号级跨 IP 锁定上限（回归修复）：(loginID, IP) 维度 10 次/15 分钟可被换 IP 绕过，
// 现另按 loginID 累计：15 分钟窗口内跨 IP 合计失败 ≥50 次锁定账号 15 分钟（无论 IP）。
// 每次失败用不同 X-Forwarded-For，确保只触发账号维度、不触发单 IP 维度。
func TestPasswordAccountLevelLockoutCrossIP(t *testing.T) {
	old := os.Getenv("ADMIN_DEV_MODE")
	os.Setenv("ADMIN_DEV_MODE", "true")
	t.Cleanup(func() { os.Setenv("ADMIN_DEV_MODE", old) })

	app := newServer(t)
	const phone = "13800005678"

	reg := httptest.NewRecorder()
	app.ServeHTTP(reg, httptest.NewRequest(http.MethodPost, "/api/auth/register",
		strings.NewReader(`{"phone":"`+phone+`","password":"Secret123"}`)))
	if reg.Code != http.StatusCreated && reg.Code != http.StatusOK {
		t.Fatalf("register: %d %s", reg.Code, reg.Body.String())
	}

	// 前 49 次错误密码，每次来自不同 IP → 401（未达账号上限 50）
	for i := 0; i < 49; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/auth/login",
			strings.NewReader(`{"phone":"`+phone+`","password":"WrongPass"}`))
		// clientIP 仅信任受信代理（回环）的 XFF：模拟 nginx 反代形态
		req.RemoteAddr = "127.0.0.1:12345"
		req.Header.Set("X-Forwarded-For", ipFor(i))
		app.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("attempt %d from ip %s: want 401, got %d", i+1, ipFor(i), w.Code)
		}
	}

	// 第 50 次错误（新 IP）→ 401（该次触发账号锁定）
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login",
		strings.NewReader(`{"phone":"`+phone+`","password":"WrongPass"}`))
	req.RemoteAddr = "127.0.0.1:12345"
	req.Header.Set("X-Forwarded-For", ipFor(49))
	app.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("50th wrong attempt: want 401, got %d", w.Code)
	}

	// 锁定后：换一个从未失败的 IP 用正确密码 → 仍 429（账号级锁定跨 IP 生效）
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/auth/login",
		strings.NewReader(`{"phone":"`+phone+`","password":"Secret123"}`))
	req.RemoteAddr = "127.0.0.1:12345"
	req.Header.Set("X-Forwarded-For", ipFor(50))
	app.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("account locked across IPs: correct password from fresh ip want 429, got %d %s", w.Code, w.Body.String())
	}
}

// ipFor 生成测试用不同客户端 IP（10.0.0.N / 10.0.1.N），避开单 IP 维度锁定。
func ipFor(i int) string {
	if i < 250 {
		return "10.0.0." + itoa(i+1)
	}
	return "10.0.1." + itoa(i-249)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [8]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
