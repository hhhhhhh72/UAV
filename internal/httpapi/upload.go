package httpapi

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"

	"drone-platform/internal/service"
)

// POST /api/v1/upload — 上传文件，返回文件 URL
// Requires authentication and restricts content types (jpeg/png/webp/pdf),
// mirroring the rules of /api/v1/files/upload.
// 统一走 fileSvc：魔数检测 + 随机 ID（不可枚举）+ 每日上传配额记账
// （此前独立落盘：可预测 UnixNano 文件名、绕过配额）。
func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
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
	reader := io.MultiReader(bytes.NewReader(head), file)

	rec, err := s.fileSvc.Upload(r.Context(), a.ID, "upload", detected, io.LimitReader(reader, 10<<20))
	if err != nil {
		if errors.Is(err, service.ErrUploadQuotaExceeded) {
			fail(w, r, http.StatusRequestEntityTooLarge, err)
			return
		}
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	s.audit(r.Context(), a.ID, "upload_file", "file", rec.ID, "uploaded")
	respond(w, r, http.StatusOK, map[string]string{
		"url": "/uploads/" + rec.ID,
	})
}
