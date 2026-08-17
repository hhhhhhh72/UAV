package service_test

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

// TestFileServiceQuotaEnforced 验证每日配额：超限拒绝、文件不落盘、其他用户不受影响。
func TestFileServiceQuotaEnforced(t *testing.T) {
	repo := memory.NewUploadRepository()
	dir := t.TempDir()
	svc := service.NewFileService(dir, service.WithUploadQuota(repo, 100))

	rec, err := svc.Upload(context.Background(), "owner-1", "a.txt", "text/plain", strings.NewReader(strings.Repeat("a", 60)))
	if err != nil {
		t.Fatalf("first upload should pass: %v", err)
	}
	if rec.SizeBytes != 60 {
		t.Fatalf("size = %d, want 60", rec.SizeBytes)
	}
	if rec.Visibility != "public" {
		t.Fatalf("public upload visibility = %q, want public", rec.Visibility)
	}

	// 60 + 60 = 120 > 100 → 超限拒绝，且被拒文件未落盘（目录中只有已成功的文件）。
	_, err = svc.Upload(context.Background(), "owner-1", "b.txt", "text/plain", strings.NewReader(strings.Repeat("b", 60)))
	if !errors.Is(err, service.ErrUploadQuotaExceeded) {
		t.Fatalf("second upload should be quota-exceeded, got %v", err)
	}
	entries, listErr := os.ReadDir(dir)
	if listErr != nil {
		t.Fatalf("list upload dir: %v", listErr)
	}
	if len(entries) != 1 {
		t.Fatalf("upload dir has %d files, want 1 (over-quota file must not persist)", len(entries))
	}

	// 其他用户独立配额。
	rec2, err := svc.Upload(context.Background(), "owner-2", "c.txt", "text/plain", strings.NewReader(strings.Repeat("c", 60)))
	if err != nil {
		t.Fatalf("other owner should have independent quota: %v", err)
	}
	if rec2.SizeBytes != 60 {
		t.Fatalf("owner-2 size = %d, want 60", rec2.SizeBytes)
	}
}

// TestFileServiceQuotaPrivate 验证私有上传计入同一配额且 visibility=private。
func TestFileServiceQuotaPrivate(t *testing.T) {
	repo := memory.NewUploadRepository()
	svc := service.NewFileService(t.TempDir(), service.WithUploadQuota(repo, 100))

	rec, err := svc.UploadPrivate(context.Background(), "owner-1", "id.png", "image/png", strings.NewReader(strings.Repeat("x", 90)))
	if err != nil {
		t.Fatalf("private upload should pass: %v", err)
	}
	if rec.Visibility != "private" {
		t.Fatalf("visibility = %q, want private", rec.Visibility)
	}
	if _, err := svc.UploadPrivate(context.Background(), "owner-1", "id2.png", "image/png", strings.NewReader(strings.Repeat("y", 90))); !errors.Is(err, service.ErrUploadQuotaExceeded) {
		t.Fatalf("second private upload should be quota-exceeded, got %v", err)
	}
}

// TestFileServiceNoQuota 验证未配置配额时无限上传（默认/测试路径）。
func TestFileServiceNoQuota(t *testing.T) {
	svc := service.NewFileService(t.TempDir())
	big := strings.Repeat("z", 1<<20) // 1MB
	for i := 0; i < 3; i++ {
		if _, err := svc.Upload(context.Background(), "owner-1", "big.bin", "application/octet-stream", strings.NewReader(big)); err != nil {
			t.Fatalf("upload %d should pass without quota: %v", i, err)
		}
	}
}

// TestFileServiceQuotaByDay 验证配额按自然日重置：昨日用量不计入今日。
func TestFileServiceQuotaByDay(t *testing.T) {
	repo := memory.NewUploadRepository()
	svc := service.NewFileService(t.TempDir(), service.WithUploadQuota(repo, 100))

	if _, err := svc.Upload(context.Background(), "owner-1", "old.txt", "text/plain", strings.NewReader(strings.Repeat("o", 90))); err != nil {
		t.Fatalf("upload should pass: %v", err)
	}
	if _, err := svc.Upload(context.Background(), "owner-1", "new.txt", "text/plain", strings.NewReader(strings.Repeat("n", 90))); !errors.Is(err, service.ErrUploadQuotaExceeded) {
		t.Fatalf("same-day second upload should exceed quota, got %v", err)
	}
}
