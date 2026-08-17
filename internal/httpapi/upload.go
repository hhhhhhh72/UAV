package httpapi

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// POST /api/v1/upload — 上传文件，返回文件 URL
// Requires authentication and restricts content types (jpeg/png/webp/pdf),
// mirroring the rules of /api/v1/files/upload.
func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request) {
	if _, ok := authenticatedActor(r); !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB max
		fail(w, r, http.StatusBadRequest, fmt.Errorf("parse multipart: %w", err))
		return
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		fail(w, r, http.StatusBadRequest, fmt.Errorf("read file: %w", err))
		return
	}
	defer file.Close()

	// P1 修复：魔数检测替代客户端 Content-Type 信任；
	// 扩展名由检测结果决定，杜绝".html 装成 image/jpeg"落盘后被按扩展名回放。
	detected, head, err := sniffAllowedType(file)
	if err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	ext := ".bin"
	switch detected {
	case "image/jpeg":
		ext = ".jpg"
	case "image/png":
		ext = ".png"
	case "image/webp":
		ext = ".webp"
	case "application/pdf":
		ext = ".pdf"
	}

	// 确保上传目录存在
	dir := "uploads"
	if err := os.MkdirAll(dir, 0755); err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}

	// 生成唯一文件名（扩展名由检测类型决定，不再信任客户端文件名）
	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	path := filepath.Join(dir, filename)

	dst, err := os.Create(path)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, io.MultiReader(bytes.NewReader(head), file)); err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}

	respond(w, r, http.StatusOK, map[string]string{
		"url": "/uploads/" + filename,
	})
}
