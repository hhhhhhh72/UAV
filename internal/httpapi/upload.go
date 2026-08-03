package httpapi

import (
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
	file, header, err := r.FormFile("file")
	if err != nil {
		fail(w, r, http.StatusBadRequest, fmt.Errorf("read file: %w", err))
		return
	}
	defer file.Close()

	// 校验文件类型：仅允许图片与 PDF
	ct := header.Header.Get("Content-Type")
	if ct != "image/jpeg" && ct != "image/png" && ct != "image/webp" && ct != "application/pdf" {
		fail(w, r, http.StatusBadRequest, errors.New("unsupported file type: only jpeg/png/webp/pdf allowed"))
		return
	}

	// 确保上传目录存在
	dir := "uploads"
	if err := os.MkdirAll(dir, 0755); err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}

	// 生成唯一文件名
	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = ".jpg"
	}
	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	path := filepath.Join(dir, filename)

	dst, err := os.Create(path)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}

	respond(w, r, http.StatusOK, map[string]string{
		"url": "/uploads/" + filename,
	})
}
