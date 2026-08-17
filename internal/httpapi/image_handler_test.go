package httpapi_test

import (
	"bytes"
	"image"
	"image/png"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"drone-platform/internal/domain"
)

// TestServeImageRejectsUnsupportedFormat: 回归 C7——webp 曾按 jpeg 编码
// 却声明 image/webp，现应显式 400；合法格式仍走正常流程（文件不存在 → 404）。
func TestServeImageRejectsUnsupportedFormat(t *testing.T) {
	app := newServer(t)

	w := doRaw(app, http.MethodGet, "/api/v1/image?url=nonexistent.png&format=webp", "", "")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("webp format: expected 400, got %d (%s)", w.Code, w.Body.String())
	}

	// 合法格式 + 不存在的文件 → 404（证明格式校验先于文件检查，且 400 不是误报）
	w = doRaw(app, http.MethodGet, "/api/v1/image?url=nonexistent.png&format=jpeg", "", "")
	if w.Code != http.StatusNotFound {
		t.Fatalf("jpeg format: expected 404, got %d (%s)", w.Code, w.Body.String())
	}
}

// TestServeImageRejectsTraversal: SanitizeString 只去 HTML 标签，
// ../ 逃逸到 uploads 目录外必须 400（旧实现会穿透做文件存在性探测）。
func TestServeImageRejectsTraversal(t *testing.T) {
	app := newServer(t)

	for _, u := range []string{"..%2F..%2Fgo.mod", "..%2F..%2Fetc%2Fpasswd", "%2e%2e%2fcmd%2fapi%2fmain.go"} {
		w := doRaw(app, http.MethodGet, "/api/v1/image?url="+u+"&format=jpeg", "", "")
		if w.Code != http.StatusBadRequest {
			t.Fatalf("url=%s: expected 400, got %d (%s)", u, w.Code, w.Body.String())
		}
	}
}

// TestServeImageServesAndCaches: 合法图片走完整流程——200 + JPEG 内容 +
// 磁盘缓存落盘 + 第二次请求命中缓存返回一致内容。
func TestServeImageServesAndCaches(t *testing.T) {
	app := newServer(t)

	// Arrange: 在 uploads 下放一张 64x48 PNG。
	const name = "c7-valid-test.png"
	if err := os.MkdirAll("uploads", 0755); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(filepath.Join("uploads", name))
	var pngBuf bytes.Buffer
	if err := png.Encode(&pngBuf, image.NewRGBA(image.Rect(0, 0, 64, 48))); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join("uploads", name), pngBuf.Bytes(), 0644); err != nil {
		t.Fatal(err)
	}

	// Act: 缩小到 32 宽 + jpeg 输出。
	w := doRaw(app, http.MethodGet, "/api/v1/image?url="+name+"&width=32&quality=70&format=jpeg", "", "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", w.Code, w.Body.String())
	}
	if ct := w.Header().Get("Content-Type"); ct != "image/jpeg" {
		t.Fatalf("content type = %q", ct)
	}
	body := w.Body.Bytes()
	if len(body) < 3 || body[0] != 0xFF || body[1] != 0xD8 || body[2] != 0xFF {
		t.Fatalf("response is not a JPEG (first bytes % x)", body[:min(3, len(body))])
	}

	// Assert: 缓存文件已落盘；第二次请求命中缓存返回一致内容。
	entries, err := filepath.Glob(filepath.Join("uploads", ".image-cache", "*.jpeg"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) == 0 {
		t.Fatal("expected cache file written")
	}
	defer func() {
		for _, e := range entries {
			os.Remove(e)
		}
		os.Remove(filepath.Join("uploads", ".image-cache"))
	}()

	w2 := doRaw(app, http.MethodGet, "/api/v1/image?url="+name+"&width=32&quality=70&format=jpeg", "", "")
	if w2.Code != http.StatusOK {
		t.Fatalf("cache hit: expected 200, got %d", w2.Code)
	}
	if !bytes.Equal(body, w2.Body.Bytes()) {
		t.Fatal("cache hit returned different bytes")
	}
}

// TestServeImagePrivateRequiresAuth: P0 回归——/api/v1/image 在公开白名单，
// 此前匿名可凭可预测的 file-<UnixNano> ID 重编码读取 uploads/private/ 下的
// 身份证影像。修复后 private 路径必须登录。
func TestServeImagePrivateRequiresAuth(t *testing.T) {
	app := newServer(t)

	// Arrange: uploads/private/ 下放一张测试 PNG
	const rel = "private/idcard-test.png"
	if err := os.MkdirAll(filepath.Join("uploads", "private"), 0755); err != nil {
		t.Fatal(err)
	}
	var pngBuf bytes.Buffer
	if err := png.Encode(&pngBuf, image.NewRGBA(image.Rect(0, 0, 32, 32))); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join("uploads", rel), pngBuf.Bytes(), 0644); err != nil {
		t.Fatal(err)
	}
	defer func() {
		os.Remove(filepath.Join("uploads", rel))
		entries, _ := filepath.Glob(filepath.Join("uploads", ".image-cache", "*.jpeg"))
		for _, e := range entries {
			os.Remove(e)
		}
	}()

	// 匿名访问 private 路径 → 401（修复前是 200 且能读到内容）
	w := doRaw(app, http.MethodGet, "/api/v1/image?url="+rel+"&width=16&format=jpeg", "", "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("anonymous private image: expected 401, got %d (%s)", w.Code, w.Body.String())
	}

	// 带 token → 200
	tok := authAs(t, "user-1", domain.RoleIndividual)
	w = doRaw(app, http.MethodGet, "/api/v1/image?url="+rel+"&width=16&format=jpeg", "", tok)
	if w.Code != http.StatusOK {
		t.Fatalf("authorized private image: expected 200, got %d (%s)", w.Code, w.Body.String())
	}
}
