package service

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"drone-platform/internal/domain"
)

type FileService struct {
	uploadDir string
}

func NewFileService(uploadDir string) *FileService {
	os.MkdirAll(uploadDir, 0755)
	return &FileService{uploadDir: uploadDir}
}

func (s *FileService) Upload(ownerID string, filename, contentType string, reader io.Reader) (domain.FileRecord, error) {
	now := time.Now()
	id := fmt.Sprintf("file-%d", now.UnixNano())

	hasher := sha256.New()
	destPath := filepath.Join(s.uploadDir, id)
	f, err := os.Create(destPath)
	if err != nil {
		return domain.FileRecord{}, fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	tee := io.TeeReader(reader, hasher)
	size, err := io.Copy(f, tee)
	if err != nil {
		os.Remove(destPath)
		return domain.FileRecord{}, fmt.Errorf("write file: %w", err)
	}

	return domain.FileRecord{
		ID:          id,
		StorageKey:  destPath,
		SHA256:      fmt.Sprintf("%x", hasher.Sum(nil)),
		ContentType: contentType,
		SizeBytes:   size,
		Visibility:  "private",
		OwnerID:     ownerID,
		CreatedAt:   now,
	}, nil
}
