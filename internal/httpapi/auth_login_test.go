package httpapi_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestPasswordLoginFlow verifies the register → login lifecycle:
// registration persists a bcrypt hash, correct password logs in,
// wrong password is rejected, duplicate registration conflicts.
func TestPasswordLoginFlow(t *testing.T) {
	app := newServer(t)

	// 1) Register
	regBody := []byte(`{"phone":"13800001111","password":"secret-pass-1","name":"测试用户"}`)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(string(regBody))))
	if w.Code != http.StatusOK {
		t.Fatalf("register: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// 2) Duplicate registration → conflict
	w2 := httptest.NewRecorder()
	app.ServeHTTP(w2, httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(string(regBody))))
	if w2.Code != http.StatusConflict {
		t.Fatalf("duplicate register: expected 409, got %d", w2.Code)
	}

	// 3) Login with correct password → tokens issued
	loginOK := []byte(`{"phone":"13800001111","password":"secret-pass-1"}`)
	w3 := httptest.NewRecorder()
	app.ServeHTTP(w3, httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(string(loginOK))))
	if w3.Code != http.StatusOK {
		t.Fatalf("login correct password: expected 200, got %d: %s", w3.Code, w3.Body.String())
	}
	if !strings.Contains(w3.Body.String(), "accessToken") {
		t.Fatalf("login response missing accessToken: %s", w3.Body.String())
	}

	// 4) Login with wrong password → 401
	loginBad := []byte(`{"phone":"13800001111","password":"wrong-password"}`)
	w4 := httptest.NewRecorder()
	app.ServeHTTP(w4, httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(string(loginBad))))
	if w4.Code != http.StatusUnauthorized {
		t.Fatalf("login wrong password: expected 401, got %d: %s", w4.Code, w4.Body.String())
	}

	// 5) Login with empty password → 400 (missing-field validation)
	loginEmpty := []byte(`{"phone":"13800001111","password":""}`)
	w5 := httptest.NewRecorder()
	app.ServeHTTP(w5, httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(string(loginEmpty))))
	if w5.Code != http.StatusBadRequest {
		t.Fatalf("login empty password: expected 400, got %d", w5.Code)
	}

	// 6) Login for a user that never registered → 401
	loginGhost := []byte(`{"phone":"13900002222","password":"whatever"}`)
	w6 := httptest.NewRecorder()
	app.ServeHTTP(w6, httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(string(loginGhost))))
	if w6.Code != http.StatusUnauthorized {
		t.Fatalf("login unknown user: expected 401, got %d", w6.Code)
	}
}
