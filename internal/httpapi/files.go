package httpapi

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"drone-platform/internal/domain"
)

// POST /api/v1/files/upload
// Accepts multipart/form-data with field "file".
// Returns { file_id, sha256, size_bytes, content_type }.
// Max 10 MiB per file.
func (s *Server) uploadFile(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }

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

	rec, err := s.fileSvc.Upload(a.ID, header.Filename, ct, io.LimitReader(file, 10<<20))
	if err != nil { fail(w, r, http.StatusInternalServerError, err); return }

	// Persist file metadata to DB (best-effort via audit).
	s.audit(r.Context(), a.ID, "upload_file", "file", rec.ID, "uploaded")
	respond(w, r, http.StatusCreated, map[string]any{
		"file_id":      rec.ID,
		"sha256":       rec.SHA256,
		"size_bytes":   rec.SizeBytes,
		"content_type": rec.ContentType,
	})
}

// POST /api/v1/enterprises/{id}/documents
// Link a previously uploaded file to an enterprise as a document.
func (s *Server) attachEnterpriseDocument(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }

	entID := r.PathValue("id")
	var in struct {
		FileID       string `json:"file_id"`
		DocumentType string `json:"document_type"` // business_license, id_card, etc.
	}
	if err := decode(r, &in); err != nil { fail(w, r, http.StatusBadRequest, err); return }

	now := time.Now()
	doc := domain.EnterpriseDocument{
		ID:           fmt.Sprintf("edoc-%d", now.UnixNano()),
		EnterpriseID: entID,
		FileID:       in.FileID,
		DocumentType: in.DocumentType,
		ReviewStatus: "pending",
		CreatedAt:    now,
	}
	s.audit(r.Context(), a.ID, "attach_enterprise_doc", "enterprise", entID, doc.ID)
	respond(w, r, http.StatusCreated, doc)
}
