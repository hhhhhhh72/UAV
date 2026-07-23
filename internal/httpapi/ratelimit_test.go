package httpapi

import (
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
