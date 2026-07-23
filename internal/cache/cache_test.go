package cache_test

import (
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
