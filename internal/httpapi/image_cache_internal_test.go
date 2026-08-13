package httpapi

import (
	"testing"
)

// TestImageCacheKeyIncludesParams: 回归 C7——缓存 key 必须完整区分
// url/width/quality/format/mtime，且不得再被截断到 32 字节前缀。
func TestImageCacheKeyIncludesParams(t *testing.T) {
	base := imageCacheKey("banner.png", 800, 75, "jpeg", 100)

	variants := []struct {
		name string
		key  string
	}{
		{"width", imageCacheKey("banner.png", 400, 75, "jpeg", 100)},
		{"quality", imageCacheKey("banner.png", 800, 90, "jpeg", 100)},
		{"format", imageCacheKey("banner.png", 800, 75, "png", 100)},
		{"mtime", imageCacheKey("banner.png", 800, 75, "jpeg", 200)},
		{"url", imageCacheKey("other.png", 800, 75, "jpeg", 100)},
	}
	for _, v := range variants {
		if v.key == base {
			t.Fatalf("cache key must differ by %s: %s", v.name, v.key)
		}
	}

	// 旧实现 [:32] 截断：全 key 恰好等于 32 字节 hash 前缀（64 hex + 参数被切）
	if len(base) <= 32 {
		t.Fatalf("cache key truncated to %d bytes: %q", len(base), base)
	}

	// 相同参数必须产出相同 key（缓存可命中）
	if imageCacheKey("banner.png", 800, 75, "jpeg", 100) != base {
		t.Fatal("same params must produce the same key")
	}
}

