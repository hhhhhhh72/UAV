package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRateLimiterAllowsAndBlocks(t *testing.T) {
	rl := newRateLimiter(10, 20)

	for i := 0; i < 20; i++ {
		if !rl.allow("test-ip") {
			t.Fatalf("request %d should be allowed", i)
		}
	}
	if rl.allow("test-ip") {
		t.Fatal("request 21 should be rate-limited")
	}
	t.Log("rate limiter: burst 20 OK, 21st blocked")
}

func TestRateLimiterSeparateKeys(t *testing.T) {
	rl := newRateLimiter(5, 5)
	// Exhaust IP1.
	for i := 0; i < 5; i++ {
		rl.allow("ip-1")
	}
	// IP2 should still be allowed.
	if !rl.allow("ip-2") {
		t.Fatal("separate IP should have independent bucket")
	}
	t.Log("rate limiter: separate keys independent OK")
}

func TestClientIP(t *testing.T) {
	cases := []struct {
		name string
		xff  string
		addr string
		want string
	}{
		{"xff single hop", "1.2.3.4", "127.0.0.1:8080", "1.2.3.4"},
		{"xff last hop of chain", "1.2.3.4, 10.0.0.1", "127.0.0.1:8080", "10.0.0.1"},
		{"xff trimmed", "  198.51.100.7  ", "127.0.0.1:8080", "198.51.100.7"},
		{"no xff falls back to peer", "", "203.0.113.9:5555", "203.0.113.9"},
		{"blank xff falls back to peer", "   ", "203.0.113.9:5555", "203.0.113.9"},
		{"invalid xff falls back to peer", "not-an-ip", "203.0.113.9:5555", "203.0.113.9"},
		{"peer without port", "", "203.0.113.9", "203.0.113.9"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/", nil)
			r.RemoteAddr = c.addr
			if c.xff != "" {
				r.Header.Set("X-Forwarded-For", c.xff)
			}
			if got := clientIP(r); got != c.want {
				t.Fatalf("clientIP = %q, want %q", got, c.want)
			}
		})
	}
}

// TestRateLimitUsesClientIP 验证反代场景下按真实客户端 IP 分桶：
// 相同 RemoteAddr（127.0.0.1）的不同 XFF IP 各自独立，不再全局共享单桶。
func TestRateLimitUsesClientIP(t *testing.T) {
	rl := newRateLimiter(1, 2)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !rl.allow(clientIP(r)) {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	req := func(xff string) int {
		r := httptest.NewRequest("GET", "/", nil)
		r.RemoteAddr = "127.0.0.1:8080"
		if xff != "" {
			r.Header.Set("X-Forwarded-For", xff)
		}
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		return w.Code
	}
	// A：burst=2，前两次放行、第三次被限。
	if code := req("198.51.100.1"); code != http.StatusOK {
		t.Fatalf("A req1 should pass, got %d", code)
	}
	if code := req("198.51.100.1"); code != http.StatusOK {
		t.Fatalf("A req2 should pass, got %d", code)
	}
	if code := req("198.51.100.1"); code != http.StatusTooManyRequests {
		t.Fatalf("A req3 should be rate-limited, got %d", code)
	}
	// B：不同 XFF IP 独立桶，仍可放行。
	if code := req("198.51.100.2"); code != http.StatusOK {
		t.Fatalf("B should have independent bucket, got %d", code)
	}
	// 无 XFF（直连/容器健康检查）：按 RemoteAddr 独立分桶。
	if code := req(""); code != http.StatusOK {
		t.Fatalf("direct peer should have independent bucket, got %d", code)
	}
}
