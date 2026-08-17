package httpapi

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"drone-platform/internal/middleware"
)

// imageCacheKey 生成磁盘缓存文件名：完整 hash 拼接宽/质量/格式。
// 回归 C7：旧实现 [:32] 截断把 width/quality/format 全部切掉，
// 同图不同尺寸命中同一缓存文件。
func imageCacheKey(urlPath string, width, quality int, outputFormat string, modTime int64) string {
	h := sha256.Sum256([]byte(fmt.Sprintf("%s%d%d%s%d", urlPath, width, quality, outputFormat, modTime)))
	return fmt.Sprintf("%x_%d_%d_%s", h, width, quality, outputFormat)
}

// GET /api/v1/image — resize & convert images with disk caching.
// Query params: url (path), width (default 800), quality (default 75), format (jpeg|png)
func (s *Server) serveImage(w http.ResponseWriter, r *http.Request) {
	urlPath := r.URL.Query().Get("url")
	if urlPath == "" {
		fail(w, r, http.StatusBadRequest, fmt.Errorf("url parameter required"))
		return
	}
	urlPath = strings.TrimPrefix(urlPath, "/")
	baseDir := "uploads"

	width, _ := strconv.Atoi(r.URL.Query().Get("width"))
	if width <= 0 {
		width = 800
	}
	// B 批加固：width 无上界可被恶意参数触发超大位图内存分配（OOM）。
	// 限制最大 2000px，超出按上限处理。
	if width > 2000 {
		width = 2000
	}
	quality, _ := strconv.Atoi(r.URL.Query().Get("quality"))
	if quality <= 0 || quality > 100 {
		quality = 75
	}
	outputFormat := r.URL.Query().Get("format")
	if outputFormat == "" {
		outputFormat = "jpeg"
	}
	// C7 修复：仅支持 jpeg/png。webp 曾按 jpeg 编码却声明 image/webp + .webp 后缀，
	// 与其返回错乱内容不如显式拒绝。
	if outputFormat != "jpeg" && outputFormat != "png" {
		fail(w, r, http.StatusBadRequest, fmt.Errorf("unsupported format %q (supported: jpeg, png)", outputFormat))
		return
	}

	imagePath := filepath.Join(baseDir, middleware.SanitizeString(urlPath))
	// 路径包含校验：SanitizeString 只去 HTML 标签，挡不住 ../ 逃逸。
	// 解析后的绝对路径必须仍在 uploads 目录内，否则拒绝。
	absBase, err := filepath.Abs(baseDir)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	absPath, err := filepath.Abs(imagePath)
	if err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	if !strings.HasPrefix(absPath, absBase+string(filepath.Separator)) {
		fail(w, r, http.StatusBadRequest, fmt.Errorf("invalid image path"))
		return
	}
	// P0 修复：uploads/private/ 下是身份证影像等敏感文件。
	// /api/v1/image 在公开白名单里，此前匿名凭文件 ID（file-<UnixNano>，可预测）
	// 即可重编码读取私有影像。B 批加固：private 路径一律拒绝经本代理访问——
	// 私有影像只允许登录用户直接读 /uploads/private/{id}（servePrivateUploads），
	// 不提供缩放代理，彻底消除该代理面的 IDOR/大小写绕过风险。
	if absPrivate, err := filepath.Abs(filepath.Join(baseDir, "private")); err == nil {
		if absPath == absPrivate || strings.HasPrefix(absPath, absPrivate+string(filepath.Separator)) {
			fail(w, r, http.StatusForbidden, errors.New("private images are not served through this endpoint"))
			return
		}
	}
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		fail(w, r, http.StatusNotFound, fmt.Errorf("image not found"))
		return
	}

	// Check disk cache
	cacheDir := filepath.Join(baseDir, ".image-cache")
	os.MkdirAll(cacheDir, 0755)
	srcStat, err := os.Stat(imagePath)
	if err != nil {
		fail(w, r, http.StatusNotFound, err)
		return
	}
	cacheKey := imageCacheKey(urlPath, width, quality, outputFormat, srcStat.ModTime().Unix())
	cachePath := filepath.Join(cacheDir, cacheKey+"."+outputFormat)

	if cached, err := os.ReadFile(cachePath); err == nil {
		w.Header().Set("Content-Type", "image/"+outputFormat)
		w.Header().Set("Cache-Control", "public, max-age=2592000, immutable")
		w.Header().Set("Vary", "Accept")
		w.Write(cached)
		return
	}

	// Decode source
	f, err := os.Open(imagePath)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	defer f.Close()
	srcImg, _, err := image.Decode(f)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, fmt.Errorf("decode failed: %w", err))
		return
	}

	// Resize (simple nearest-neighbor scaling; for production, use a proper library)
	bounds := srcImg.Bounds()
	origW := bounds.Dx()
	origH := bounds.Dy()
	newW, newH := width, origH*width/origW
	if origW <= width {
		newW, newH = origW, origH
	}
	resized := image.NewRGBA(image.Rect(0, 0, newW, newH))
	for y := 0; y < newH; y++ {
		for x := 0; x < newW; x++ {
			srcX := x * origW / newW
			srcY := y * origH / newH
			resized.Set(x, y, srcImg.At(srcX, srcY))
		}
	}

	// Encode once into a buffer, then persist to the disk cache and respond
	// from the same buffer. 旧实现写缓存文件后再 ReadFile 回读且吞错——
	// 回读失败时响应体为空。编码失败现在显式 500。
	var buf bytes.Buffer
	switch outputFormat {
	case "png":
		if err := png.Encode(&buf, resized); err != nil {
			fail(w, r, http.StatusInternalServerError, fmt.Errorf("encode png: %w", err))
			return
		}
	default:
		if err := jpeg.Encode(&buf, resized, &jpeg.Options{Quality: quality}); err != nil {
			fail(w, r, http.StatusInternalServerError, fmt.Errorf("encode jpeg: %w", err))
			return
		}
	}
	if err := os.WriteFile(cachePath, buf.Bytes(), 0644); err != nil {
		slog.Warn("image: write cache file", "path", cachePath, "err", err)
	}

	w.Header().Set("Content-Type", "image/"+outputFormat)
	w.Header().Set("Cache-Control", "public, max-age=2592000, immutable")
	w.Header().Set("Vary", "Accept")
	w.Write(buf.Bytes())
}
