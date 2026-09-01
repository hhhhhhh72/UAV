package middleware_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"drone-platform/internal/middleware"
)

func TestSanitizeString(t *testing.T) {
	tests := []struct{ in, expected string }{
		{"<script>alert('xss')</script>", "alert('xss')"},
		{"普通文本", "普通文本"},
		{`<a href="evil">click</a>`, "click"},
		{"", ""},
	}
	for _, tc := range tests {
		got := middleware.SanitizeString(tc.in)
		if got != tc.expected {
			t.Errorf("SanitizeString(%q) = %q, want %q", tc.in, got, tc.expected)
		}
	}
}

func TestSanitizeStringLong(t *testing.T) {
	long := strings.Repeat("A", middleware.MaxSanitizeFieldBytes+1)
	// 超长不再静默截断：strict 版本报错（用户提交超长内容时数据被悄悄改掉的旧行为已修复）
	if _, err := middleware.SanitizeStringStrict(long); err == nil {
		t.Fatal("SanitizeStringStrict: expected error for overlong field")
	}
	// 合法长度正常通过
	if _, err := middleware.SanitizeStringStrict(strings.Repeat("A", middleware.MaxSanitizeFieldBytes)); err != nil {
		t.Fatalf("SanitizeStringStrict: max-length field should pass: %v", err)
	}
}

func TestSanitizeBodyFieldTooLong(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	sanitized := middleware.SanitizeBody(next)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/test",
		strings.NewReader(`{"title":"`+strings.Repeat("A", middleware.MaxSanitizeFieldBytes+1)+`"}`))
	r.Header.Set("Content-Type", "application/json")
	sanitized.ServeHTTP(w, r)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("overlong field should be rejected with 400, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "FIELD_TOO_LONG") {
		t.Fatalf("missing FIELD_TOO_LONG code: %s", w.Body.String())
	}
}

func TestSanitizeMap(t *testing.T) {
	input := map[string]any{
		"name":  "<b>张三</b>",
		"bio":   "正常文本",
		"count": 42,
	}
	got := middleware.SanitizeMap(input)
	if got["name"] != "张三" { t.Fatalf("name not sanitized: %v", got["name"]) }
	if got["bio"] != "正常文本" { t.Fatalf("bio changed: %v", got["bio"]) }
	if got["count"] != 42 { t.Fatalf("non-string changed: %v", got["count"]) }
}

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()
	middleware.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	if w.Code != http.StatusOK { t.Fatalf("status: %d", w.Code) }
	if w.Body.Len() == 0 { t.Fatal("empty body") }
}

func TestWriteError(t *testing.T) {
	w := httptest.NewRecorder()
	middleware.WriteError(w, http.StatusBadRequest, "BAD_INPUT", "参数错误")
	if w.Code != http.StatusBadRequest { t.Fatalf("status: %d", w.Code) }
	if w.Body.Len() == 0 { t.Fatal("empty body") }
	if !strings.Contains(w.Body.String(), "BAD_INPUT") { t.Fatal("missing code") }
}

func TestSanitizeBodyGETPassthrough(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	sanitized := middleware.SanitizeBody(next)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	sanitized.ServeHTTP(w, r)
	if w.Code != http.StatusOK { t.Fatalf("status: %d", w.Code) }
}

func TestSanitizeBodyNonJSONPassthrough(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	sanitized := middleware.SanitizeBody(next)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/test", nil)
	r.Header.Set("Content-Type", "text/plain")
	sanitized.ServeHTTP(w, r)
	if w.Code != http.StatusOK { t.Fatalf("status: %d", w.Code) }
}

// TestSanitizeBodyJSON verifies the middleware actually sanitizes JSON string
// values (previously it was a no-op): HTML tags stripped, password values
// preserved, and large integers kept verbatim (json.Number, no float64 loss).
func TestSanitizeBodyJSON(t *testing.T) {
	var gotBody string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read body: %v", err)
		}
		gotBody = string(b)
		w.WriteHeader(http.StatusOK)
	})
	sanitized := middleware.SanitizeBody(next)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(
		`{"name":"<b>张三</b>","password":"<keep>me</keep>","big":1234567890123456789,"nested":{"bio":"<i>飞手</i>"},"arr":["<a href='/x'>ok</a>"]}`))
	r.Header.Set("Content-Type", "application/json")
	sanitized.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status: %d", w.Code)
	}

	var m map[string]any
	if err := json.Unmarshal([]byte(gotBody), &m); err != nil {
		t.Fatalf("sanitized body is not valid JSON: %v (%s)", err, gotBody)
	}
	if m["name"] != "<b>张三</b>" {
		t.Fatalf("name should keep whitelisted tag: %v", m["name"])
	}
	if m["password"] != "<keep>me</keep>" {
		t.Fatalf("password must be preserved: %v", m["password"])
	}
	nested, _ := m["nested"].(map[string]any)
	if nested["bio"] != "<i>飞手</i>" {
		t.Fatalf("nested bio should keep whitelisted tag: %v", nested["bio"])
	}
	arr, _ := m["arr"].([]any)
	if arr[0] != "<a href=\"/x\">ok</a>" {
		t.Fatalf("array item should keep <a> with relative href: %v", arr[0])
	}
	// 白名单外：危险标签连内容删除、a 的 javascript: href 整标签剥除、事件属性剥除
	s, _ := middleware.SanitizeStringStrict(`<script>alert(1)</script><p onclick="x">你好 <a href="javascript:alert(1)">链接</a><img src="x" onerror="y"></p>`)
	if s != "<p>你好 链接</p>" {
		t.Fatalf("whitelist sanitize mismatch: %q", s)
	}
	// 大整数必须原样保留（budget_fen 等 int64 字段精度）
	if !strings.Contains(gotBody, "1234567890123456789") {
		t.Fatalf("large integer lost precision: %s", gotBody)
	}
}
