package httpapi

import (
	"crypto/sha256"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"drone-platform/internal/middleware"
)

// GET /api/v1/image — resize & convert images with disk caching.
// Query params: url (path), width (default 800), quality (default 75), format (jpeg|png|webp)
func (s *Server) serveImage(w http.ResponseWriter, r *http.Request) {
	urlPath := r.URL.Query().Get("url")
	if urlPath == "" {
		fail(w, r, http.StatusBadRequest, fmt.Errorf("url parameter required"))
		return
	}
	urlPath = strings.TrimPrefix(urlPath, "/")
	baseDir := "uploads"
	imagePath := filepath.Join(baseDir, middleware.SanitizeString(urlPath))
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		fail(w, r, http.StatusNotFound, fmt.Errorf("image not found"))
		return
	}

	width, _ := strconv.Atoi(r.URL.Query().Get("width"))
	if width <= 0 {
		width = 800
	}
	quality, _ := strconv.Atoi(r.URL.Query().Get("quality"))
	if quality <= 0 || quality > 100 {
		quality = 75
	}
	outputFormat := r.URL.Query().Get("format")
	if outputFormat == "" {
		outputFormat = "jpeg"
	}

	// Check disk cache
	cacheDir := filepath.Join(baseDir, ".image-cache")
	os.MkdirAll(cacheDir, 0755)
	srcStat, _ := os.Stat(imagePath)
	cacheKey := fmt.Sprintf("%x_%d_%d_%s", sha256.Sum256([]byte(fmt.Sprintf("%s%d%d%s%d", urlPath, width, quality, outputFormat, srcStat.ModTime().Unix()))), width, quality, outputFormat)[:32]
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

	// Encode to cache
	var buf strings.Builder
	writer := io.StringWriter(&buf)
	_ = writer // unused in practice for jpeg

	// Write to temp buffer
	var encoded []byte
	cacheFile, _ := os.Create(cachePath)
	if cacheFile != nil {
		defer cacheFile.Close()
		switch outputFormat {
		case "png":
			png.Encode(cacheFile, resized)
		default:
			jpeg.Encode(cacheFile, resized, &jpeg.Options{Quality: quality})
		}
		cacheFile.Close()
		encoded, _ = os.ReadFile(cachePath)
	}

	if encoded == nil {
		jpeg.Encode(io.Discard, resized, &jpeg.Options{Quality: quality})
	}

	w.Header().Set("Content-Type", "image/"+outputFormat)
	w.Header().Set("Cache-Control", "public, max-age=2592000, immutable")
	w.Header().Set("Vary", "Accept")
	if encoded != nil {
		w.Write(encoded)
	}
}
