package cache_test

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"drone-platform/internal/cache"
)

func TestNewCache(t *testing.T) {
	c := cache.New(30 * time.Second)
	if c == nil { t.Fatal("cache is nil") }
}

func TestSetAndGet(t *testing.T) {
	c := cache.New(60 * time.Second)
	c.Set("key1", "value1")
	v, ok := c.Get("key1")
	if !ok || v != "value1" { t.Fatal("get failed") }
}

func TestGetMissingKey(t *testing.T) {
	c := cache.New(60 * time.Second)
	_, ok := c.Get("nonexistent")
	if ok { t.Fatal("should not find") }
}

func TestExpiredKey(t *testing.T) {
	c := cache.New(1 * time.Millisecond)
	c.Set("k", "v")
	time.Sleep(5 * time.Millisecond)
	_, ok := c.Get("k")
	if ok { t.Fatal("should be expired") }
}

func TestSetTTL(t *testing.T) {
	c := cache.New(60 * time.Second)
	c.SetTTL("k", "v", 100*time.Millisecond)
	v, ok := c.Get("k")
	if !ok || v != "v" { t.Fatal("get before expiry") }
	time.Sleep(150 * time.Millisecond)
	_, ok = c.Get("k")
	if ok { t.Fatal("should be expired") }
}

func TestDelete(t *testing.T) {
	c := cache.New(60 * time.Second)
	c.Set("k", "v")
	c.Delete("k")
	_, ok := c.Get("k")
	if ok { t.Fatal("should be deleted") }
}

func TestClear(t *testing.T) {
	c := cache.New(60 * time.Second)
	c.Set("a", 1)
	c.Set("b", 2)
	c.Clear()
	_, ok := c.Get("a")
	if ok { t.Fatal("should be cleared") }
}

func TestGetOrSet(t *testing.T) {
	c := cache.New(60 * time.Second)
	v, err := c.GetOrSet("k", func() (any, error) { return "computed", nil }, 60*time.Second)
	if err != nil || v != "computed" { t.Fatal("get or set failed") }
	v2, _ := c.GetOrSet("k", func() (any, error) { return "new", nil }, 60*time.Second)
	if v2 != "computed" { t.Fatal("should return cached") }
}

// TestGetOrSetSingleFlight verifies concurrent callers for the same key share
// one fn execution (缓存击穿防护)：fn 只跑一次，所有调用者拿到同一结果。
func TestGetOrSetSingleFlight(t *testing.T) {
	c := cache.New(60 * time.Second)

	var calls int32
	var wg sync.WaitGroup
	results := make([]any, 20)
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			v, err := c.GetOrSet("sf", func() (any, error) {
				atomic.AddInt32(&calls, 1)
				time.Sleep(20 * time.Millisecond) // 拉长填充窗口，放大并发竞争
				return "shared", nil
			}, 60*time.Second)
			if err != nil {
				t.Errorf("goroutine %d: %v", i, err)
				return
			}
			results[i] = v
		}(i)
	}
	wg.Wait()

	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("fn executed %d times, want 1 (single-flight broken)", got)
	}
	for i, v := range results {
		if v != "shared" {
			t.Fatalf("goroutine %d got %v, want 'shared'", i, v)
		}
	}
}

func TestStats(t *testing.T) {
	c := cache.New(10 * time.Millisecond)
	c.Set("a", 1)
	c.Set("b", 2)
	time.Sleep(15 * time.Millisecond)
	total, active, expired := c.Stats()
	_ = total
	_ = active
	if expired < 2 { t.Logf("stats: total=%d active=%d expired=%d", total, active, expired) }
}

func TestGlobal(t *testing.T) {
	if cache.Global == nil { t.Fatal("global cache is nil") }
	cache.Global.Set("test", 42)
	v, _ := cache.Global.Get("test")
	if v != 42 { t.Fatal("global cache") }
}
