package httpapi

import (
	"bytes"
	"context"
	"encoding/base64"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

// 1x1 透明 PNG（魔数可被 http.DetectContentType 识别为 image/png）。
const oneByOnePNG = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChwGA60e6kgAAAABJRU5ErkJggg=="

// quotaUploadServer 构造带每日配额的 FileService 的最小 Server。
func quotaUploadServer(t *testing.T, quota int64) (*Server, repository.UploadRepository) {
	t.Helper()
	tokens, err := NewTokenManager("01234567890123456789012345678901")
	if err != nil {
		t.Fatal(err)
	}
	upRepo := memory.NewUploadRepository()
	fs := service.NewFileService(t.TempDir(), service.WithUploadQuota(upRepo, quota))
	srv := NewServer(
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		fs, nil, nil, nil, nil, nil, nil, nil, nil,
		memory.NewUserRepository(nil), memory.NewRefreshTokenRepository(), tokens,
	)
	return srv, upRepo
}

func multipartUploadRequest(t *testing.T, actor domain.Actor) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, err := mw.CreateFormFile("file", "x.png")
	if err != nil {
		t.Fatal(err)
	}
	png, _ := base64.StdEncoding.DecodeString(oneByOnePNG)
	if _, err := fw.Write(png); err != nil {
		t.Fatal(err)
	}
	mw.Close()
	r := httptest.NewRequest(http.MethodPost, "/api/v1/files/upload", &buf)
	r.Header.Set("Content-Type", mw.FormDataContentType())
	return r.WithContext(contextWithActor(r, actor))
}

// TestUploadFileQuota413 验证每日配额超限时返回 413 且不落台账。
func TestUploadFileQuota413(t *testing.T) {
	srv, upRepo := quotaUploadServer(t, 10) // 配额 10 字节 < PNG 大小
	r := multipartUploadRequest(t, domain.Actor{ID: "u1", Role: domain.RoleEnterprise})
	w := httptest.NewRecorder()
	srv.uploadFile(w, r)
	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("over-quota upload should be 413, got %d (%s)", w.Code, w.Body.String())
	}
	used, err := upRepo.SumBytesSince(context.Background(), "u1", time.Time{})
	if err != nil {
		t.Fatalf("sum bytes: %v", err)
	}
	if used != 0 {
		t.Fatalf("over-quota upload must not be recorded, used=%d", used)
	}
}

// TestUploadFileQuotaOK 验证配额内上传成功（201 + 台账记录）。
func TestUploadFileQuotaOK(t *testing.T) {
	srv, upRepo := quotaUploadServer(t, 10<<20) // 10MB 配额
	r := multipartUploadRequest(t, domain.Actor{ID: "u1", Role: domain.RoleEnterprise})
	w := httptest.NewRecorder()
	srv.uploadFile(w, r)
	if w.Code != http.StatusCreated {
		t.Fatalf("within-quota upload should be 201, got %d (%s)", w.Code, w.Body.String())
	}
	used, err := upRepo.SumBytesSince(context.Background(), "u1", time.Time{})
	if err != nil {
		t.Fatalf("sum bytes: %v", err)
	}
	if used <= 0 {
		t.Fatalf("upload should be recorded, used=%d", used)
	}
}
