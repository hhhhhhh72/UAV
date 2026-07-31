package httpapi

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// POST /api/v1/upload — 上传文件，返回文件 URL
func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request) {
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
