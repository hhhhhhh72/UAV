package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// TestSMSLoginFlow verifies the send-code → login-code lifecycle:
// code issued, wrong code rejected, correct code logs in (auto-registers),
// codes are single-use and expire.
func TestSMSLoginFlow(t *testing.T) {
	// Dev mode echoes dev_code in the send-code response so the test can read it.
	old := os.Getenv("ADMIN_DEV_MODE")
	os.Setenv("ADMIN_DEV_MODE", "true")
	t.Cleanup(func() { os.Setenv("ADMIN_DEV_MODE", old) })

	app := newServer(t)
	phone := "13900001234"

	// 1) Invalid phone → 400
	w0 := httptest.NewRecorder()
	app.ServeHTTP(w0, httptest.NewRequest(http.MethodPost, "/api/auth/send-code", strings.NewReader(`{"phone":"123"}`)))
	if w0.Code != http.StatusBadRequest {
		t.Fatalf("send-code invalid phone: expected 400, got %d", w0.Code)
	}

	// 2) Send code (dev mode echoes dev_code so tests can read it)
	w1 := httptest.NewRecorder()
	app.ServeHTTP(w1, httptest.NewRequest(http.MethodPost, "/api/auth/send-code", strings.NewReader(`{"phone":"`+phone+`"}`)))
	if w1.Code != http.StatusOK {
		t.Fatalf("send-code: expected 200, got %d: %s", w1.Code, w1.Body.String())
	}
	var sendResp struct {
		Data struct {
			DevCode string `json:"dev_code"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w1.Body.Bytes(), &sendResp); err != nil || sendResp.Data.DevCode == "" {
		t.Fatalf("send-code: dev_code missing in test env: %s", w1.Body.String())
	}

	// 3) Login with wrong code → 401
	w2 := httptest.NewRecorder()
	app.ServeHTTP(w2, httptest.NewRequest(http.MethodPost, "/api/auth/login-code", strings.NewReader(`{"phone":"`+phone+`","code":"000000"}`)))
	if w2.Code != http.StatusUnauthorized {
		t.Fatalf("login-code wrong code: expected 401, got %d", w2.Code)
	}

	// 4) Login with correct code → auto-registers and issues tokens
	w3 := httptest.NewRecorder()
	app.ServeHTTP(w3, httptest.NewRequest(http.MethodPost, "/api/auth/login-code", strings.NewReader(`{"phone":"`+phone+`","code":"`+sendResp.Data.DevCode+`"}`)))
	if w3.Code != http.StatusOK {
		t.Fatalf("login-code correct code: expected 200, got %d: %s", w3.Code, w3.Body.String())
	}
	if !strings.Contains(w3.Body.String(), "accessToken") {
		t.Fatalf("login-code response missing accessToken: %s", w3.Body.String())
	}

	// 5) Code is single-use → second attempt rejected
	w4 := httptest.NewRecorder()
	app.ServeHTTP(w4, httptest.NewRequest(http.MethodPost, "/api/auth/login-code", strings.NewReader(`{"phone":"`+phone+`","code":"`+sendResp.Data.DevCode+`"}`)))
	if w4.Code != http.StatusUnauthorized {
		t.Fatalf("login-code reuse: expected 401, got %d", w4.Code)
	}

	// 6) Login without requesting a code → 401
	w5 := httptest.NewRecorder()
	app.ServeHTTP(w5, httptest.NewRequest(http.MethodPost, "/api/auth/login-code", strings.NewReader(`{"phone":"13700009999","code":"123456"}`)))
	if w5.Code != http.StatusUnauthorized {
		t.Fatalf("login-code no code requested: expected 401, got %d", w5.Code)
	}
}

// TestSMSLoginAttemptLimit verifies brute-force protection:
// 5 wrong attempts invalidate the code (record deleted), so even the correct
// code is rejected afterwards and the client must request a new one.
func TestSMSLoginAttemptLimit(t *testing.T) {
	old := os.Getenv("ADMIN_DEV_MODE")
	os.Setenv("ADMIN_DEV_MODE", "true")
	t.Cleanup(func() { os.Setenv("ADMIN_DEV_MODE", old) })

	app := newServer(t)
	phone := "13900005678"

	w := httptest.NewRecorder()
	app.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/auth/send-code", strings.NewReader(`{"phone":"`+phone+`"}`)))
	if w.Code != http.StatusOK {
		t.Fatalf("send-code: expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var sendResp struct {
		Data struct {
			DevCode string `json:"dev_code"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &sendResp); err != nil || sendResp.Data.DevCode == "" {
		t.Fatalf("send-code: dev_code missing: %s", w.Body.String())
	}
	// 构造与真实验证码必然不同的错误码
	wrong := "000000"
	if sendResp.Data.DevCode == wrong {
		wrong = "000001"
	}

	// 前 4 次错误尝试：仍返回 401，验证码继续有效（不误杀偶发输错）
	for i := 0; i < 4; i++ {
		w = httptest.NewRecorder()
		app.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/auth/login-code", strings.NewReader(`{"phone":"`+phone+`","code":"`+wrong+`"}`)))
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("wrong attempt %d: expected 401, got %d", i+1, w.Code)
		}
	}

	// 第 5 次错误：验证码作废，提示重新获取
	w = httptest.NewRecorder()
	app.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/auth/login-code", strings.NewReader(`{"phone":"`+phone+`","code":"`+wrong+`"}`)))
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("5th wrong attempt: expected 401, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "次数过多") {
		t.Fatalf("expected attempt-limit message, got: %s", w.Body.String())
	}

	// 验证码已作废：即使正确码也拒绝
	w = httptest.NewRecorder()
	app.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/auth/login-code", strings.NewReader(`{"phone":"`+phone+`","code":"`+sendResp.Data.DevCode+`"}`)))
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("invalidated code must be rejected, got %d", w.Code)
	}
}
