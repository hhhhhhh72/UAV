package httpapi

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"drone-platform/internal/domain"
)

// sniffAllowedType 读取前 512 字节做魔数检测，返回真实内容类型与已读字节。
// P1 修复：客户端声明的 multipart Content-Type 可伪造，类型判定必须以文件
// 内容为准；仅允许 jpeg/png/webp/pdf，其余一律拒绝。
func sniffAllowedType(r io.Reader) (detected string, head []byte, err error) {
	head = make([]byte, 512)
	n, err := io.ReadFull(r, head)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		return "", nil, fmt.Errorf("read file header: %w", err)
	}
	head = head[:n]
	detected = http.DetectContentType(head)
	switch detected {
	case "image/jpeg", "image/png", "image/webp", "application/pdf":
		return detected, head, nil
	default:
		return "", nil, fmt.Errorf("unsupported file type %q: only jpeg/png/webp/pdf allowed", detected)
	}
}

// POST /api/v1/files/upload
// Accepts multipart/form-data with field "file"; 表单字段 private=true 时存到
// uploads/private/（身份证影像等敏感文件），仅鉴权后可读（见 servePrivateUploads）。
// Returns { file_id, url, sha256, size_bytes, content_type }.
// Max 10 MiB per file.
func (s *Server) uploadFile(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}

	// Limit to 10 MiB.
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		fail(w, r, http.StatusBadRequest, fmt.Errorf("file too large (max 10MB): %w", err))
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
		rec, err = s.fileSvc.UploadPrivate(a.ID, header.Filename, detected, io.LimitReader(reader, 10<<20))
	} else {
		rec, err = s.fileSvc.Upload(a.ID, header.Filename, detected, io.LimitReader(reader, 10<<20))
	}
	if err != nil {
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

	doc, err := s.enterpriseSvc.AttachDocument(a, entID, in.FileID, in.DocumentType)
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
	docs, err := s.enterpriseSvc.ListDocuments(a, r.PathValue("id"))
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
