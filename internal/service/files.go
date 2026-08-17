package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
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
	return s.uploadTo(ownerID, filename, contentType, reader, s.uploadDir)
}

// UploadPrivate 存入 uploads/private/ 子目录（身份证影像等敏感文件，仅鉴权后可读）。
func (s *FileService) UploadPrivate(ownerID string, filename, contentType string, reader io.Reader) (domain.FileRecord, error) {
	return s.uploadTo(ownerID, filename, contentType, reader, filepath.Join(s.uploadDir, "private"))
}

func (s *FileService) uploadTo(ownerID string, filename, contentType string, reader io.Reader, dir string) (domain.FileRecord, error) {
	now := time.Now()
	// B 批加固：ID 由可预测的时间戳改为 128 位随机（防枚举——
	// 私有影像 ID 此前为 file-<UnixNano>，可被暴力遍历）。
	randBytes := make([]byte, 16)
	if _, err := rand.Read(randBytes); err != nil {
		return domain.FileRecord{}, fmt.Errorf("generate file id: %w", err)
	}
	id := "file-" + hex.EncodeToString(randBytes)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return domain.FileRecord{}, fmt.Errorf("create upload dir: %w", err)
	}

	hasher := sha256.New()
	destPath := filepath.Join(dir, id)
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
