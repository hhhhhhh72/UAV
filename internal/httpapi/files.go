package httpapi

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"drone-platform/internal/domain"
	"drone-platform/internal/service"
)

// maxUploadBytes 单文件上限：40 MiB。
//
// 2026-09-11 从 10 MiB 放宽到 40 MiB，为的是案例视频（10 MiB 连 10 秒 1080p 都放不下）。
// 与 nginx 的 client_max_body_size 50m 留出余量；再大请走对象存储而不是本机磁盘。
const maxUploadBytes = 40 << 20

// sniffAllowedType 读取前 512 字节做魔数检测，返回真实内容类型与已读字节。
// P1 修复：客户端声明的 multipart Content-Type 可伪造，类型判定必须以文件
// 内容为准；仅允许 jpeg/png/webp/pdf/mp4，其余一律拒绝。
//
// mp4 的识别依据是 ISO BMFF 的 ftyp box（http.DetectContentType 对 data[4:8]=="ftyp"
// 返回 video/mp4）——.mov 也带 ftyp，会被一并接受（同为 ISO BMFF，浏览器/小程序可播）。
func sniffAllowedType(r io.Reader) (detected string, head []byte, err error) {
	head = make([]byte, 512)
	n, err := io.ReadFull(r, head)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		return "", nil, fmt.Errorf("read file header: %w", err)
	}
	head = head[:n]
	detected = http.DetectContentType(head)
	// mp4/mov 单独兜底：Go 的 DetectContentType 只认少数几个 brand（mp41/isom/avc1…），
	// 国产设备/剪辑工具导出的 mp4 常带自定义 brand，会被判成 application/octet-stream。
	// 这里按 ISO BMFF 的结构判定：前 4 字节是 box 大小、紧随其后是 "ftyp"。
	if looksLikeISOBMFF(head) {
		return "video/mp4", head, nil
	}
	switch detected {
	case "image/jpeg", "image/png", "image/webp", "application/pdf", "video/mp4":
		return detected, head, nil
	default:
		return "", nil, fmt.Errorf("unsupported file type %q: only jpeg/png/webp/pdf/mp4 allowed", detected)
	}
}

// looksLikeISOBMFF 判定 ISO BMFF 容器（mp4/mov/m4v）：首 4 字节 box 长度 + "ftyp" 魔数。
func looksLikeISOBMFF(head []byte) bool {
	return len(head) >= 12 && string(head[4:8]) == "ftyp"
}

// POST /api/v1/files/upload
// Accepts multipart/form-data with field "file"; 表单字段 private=true 时存到
// uploads/private/（身份证影像等敏感文件），仅鉴权后可读（见 servePrivateUploads）。
// Returns { file_id, url, sha256, size_bytes, content_type }.
// Max 40 MiB per file（见 maxUploadBytes）。
func (s *Server) uploadFile(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes)
	if err := r.ParseMultipartForm(maxUploadBytes); err != nil {
		fail(w, r, http.StatusBadRequest, fmt.Errorf("file too large (max %dMB): %w", maxUploadBytes>>20, err))
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		fail(w, r, http.StatusBadRequest, errors.New("'file' field is required"))
		return
	}
	defer file.Close()

	// P1 修复：按魔数检测真实类型（客户端 Content-Type 可伪造），
	// 落库 ContentType 使用检测结果而非客户端声明。
	detected, head, err := sniffAllowedType(file)
	if err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	reader := io.MultiReader(bytes.NewReader(head), file)

	private := r.FormValue("private") == "true"
	var rec domain.FileRecord
	if private {
		rec, err = s.fileSvc.UploadPrivate(r.Context(), a.ID, header.Filename, detected, io.LimitReader(reader, maxUploadBytes))
	} else {
		rec, err = s.fileSvc.Upload(r.Context(), a.ID, header.Filename, detected, io.LimitReader(reader, maxUploadBytes))
	}
	if err != nil {
		// 收尾批次：每日上传配额超限 → 413（文件已由 service 侧删除，不落盘）。
		if errors.Is(err, service.ErrUploadQuotaExceeded) {
			fail(w, r, http.StatusRequestEntityTooLarge, err)
			return
		}
		fail(w, r, http.StatusInternalServerError, err)
		return
	}

	url := "/uploads/" + rec.ID
	if private {
		url = "/uploads/private/" + rec.ID
	}
	// Persist file metadata to DB (best-effort via audit).
	s.audit(r.Context(), a.ID, "upload_file", "file", rec.ID, "uploaded")
	respond(w, r, http.StatusCreated, map[string]any{
		"file_id":      rec.ID,
		"url":          url,
		"sha256":       rec.SHA256,
		"size_bytes":   rec.SizeBytes,
		"content_type": rec.ContentType,
	})
}

// POST /api/v1/enterprises/{id}/documents
// Link a previously uploaded file to an enterprise as a document.
func (s *Server) attachEnterpriseDocument(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}

	entID := r.PathValue("id")
	var in struct {
		FileID       string `json:"file_id"`
		DocumentType string `json:"document_type"` // business_license, id_card, etc.
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	if in.FileID == "" || in.DocumentType == "" {
		fail(w, r, http.StatusBadRequest, errors.New("file_id and document_type are required"))
		return
	}

	doc, err := s.enterpriseSvc.AttachDocument(r.Context(), a, entID, in.FileID, in.DocumentType)
	if err != nil {
		code := http.StatusForbidden
		if strings.Contains(err.Error(), "not found") {
			code = http.StatusNotFound
		}
		fail(w, r, code, err)
		return
	}
	s.audit(r.Context(), a.ID, "attach_enterprise_doc", "enterprise", entID, doc.ID)
	respond(w, r, http.StatusCreated, doc)
}

// GET /api/v1/enterprises/{id}/documents
func (s *Server) listEnterpriseDocuments(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	docs, err := s.enterpriseSvc.ListDocuments(r.Context(), a, r.PathValue("id"))
	if err != nil {
		code := http.StatusForbidden
		if strings.Contains(err.Error(), "not found") {
			code = http.StatusNotFound
		}
		fail(w, r, code, err)
		return
	}
	respond(w, r, http.StatusOK, docs)
}
