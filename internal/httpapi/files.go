package httpapi

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"drone-platform/internal/domain"
)

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

	// Validate content type: allow images and PDFs.
	ct := header.Header.Get("Content-Type")
	if ct != "image/jpeg" && ct != "image/png" && ct != "image/webp" && ct != "application/pdf" {
		fail(w, r, http.StatusBadRequest, errors.New("unsupported file type: only jpeg/png/webp/pdf allowed"))
		return
	}

	private := r.FormValue("private") == "true"
	var rec domain.FileRecord
	if private {
		rec, err = s.fileSvc.UploadPrivate(a.ID, header.Filename, ct, io.LimitReader(file, 10<<20))
	} else {
		rec, err = s.fileSvc.Upload(a.ID, header.Filename, ct, io.LimitReader(file, 10<<20))
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
