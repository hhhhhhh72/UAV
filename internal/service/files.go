package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)

// ErrUploadQuotaExceeded 当日上传配额已用尽（handler 映射为 413）。
var ErrUploadQuotaExceeded = errors.New("今日上传额度已用尽")

type FileService struct {
	uploadDir  string
	uploads    repository.UploadRepository
	dailyLimit int64 // 每用户每日字节上限；0 = 不限
}

type FileServiceOption func(*FileService)

// WithUploadQuota 启用按用户每日上传配额（写入 uploads 台账，跨实例持久）。
func WithUploadQuota(up repository.UploadRepository, dailyBytes int64) FileServiceOption {
	return func(s *FileService) {
		s.uploads = up
		s.dailyLimit = dailyBytes
	}
}

func NewFileService(uploadDir string, opts ...FileServiceOption) *FileService {
	os.MkdirAll(uploadDir, 0755)
	s := &FileService{uploadDir: uploadDir}
	for _, o := range opts {
		o(s)
	}
	return s
}

func (s *FileService) Upload(ctx context.Context, ownerID string, filename, contentType string, reader io.Reader) (domain.FileRecord, error) {
	return s.uploadTo(ctx, ownerID, filename, contentType, reader, s.uploadDir, "public")
}

// UploadPrivate 存入 uploads/private/ 子目录（身份证影像等敏感文件，仅鉴权后可读）。
func (s *FileService) UploadPrivate(ctx context.Context, ownerID string, filename, contentType string, reader io.Reader) (domain.FileRecord, error) {
	return s.uploadTo(ctx, ownerID, filename, contentType, reader, filepath.Join(s.uploadDir, "private"), "private")
}

func (s *FileService) uploadTo(ctx context.Context, ownerID string, filename, contentType string, reader io.Reader, dir, visibility string) (domain.FileRecord, error) {
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

	tee := io.TeeReader(reader, hasher)
	size, err := io.Copy(f, tee)
	if err != nil {
		f.Close()
		os.Remove(destPath)
		return domain.FileRecord{}, fmt.Errorf("write file: %w", err)
	}
	// 写盘完成即关闭句柄：Windows 下句柄未关时 Remove 会失败（Access denied），
	// 导致超限/失败文件残留磁盘。
	if err := f.Close(); err != nil {
		os.Remove(destPath)
		return domain.FileRecord{}, fmt.Errorf("close file: %w", err)
	}

	rec := domain.FileRecord{
		ID:          id,
		StorageKey:  destPath,
		SHA256:      fmt.Sprintf("%x", hasher.Sum(nil)),
		ContentType: contentType,
		SizeBytes:   size,
		Visibility:  visibility,
		OwnerID:     ownerID,
		CreatedAt:   now,
	}

	// 收尾批次：按用户每日配额记账（uploads 台账）。
	// 先写盘后校验：超限即删文件、不落台账，保证已记录数据与磁盘一致；
	// 并发同用户上传可能瞬时略超一档（单实例可接受，无事务必要）。
	if s.dailyLimit > 0 && s.uploads != nil {
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		used, err := s.uploads.SumBytesSince(ctx, ownerID, start)
		if err != nil {
			os.Remove(destPath)
			return domain.FileRecord{}, fmt.Errorf("query upload quota: %w", err)
		}
		if used+size > s.dailyLimit {
			os.Remove(destPath)
			return domain.FileRecord{}, ErrUploadQuotaExceeded
		}
		if err := s.uploads.Create(ctx, rec); err != nil {
			os.Remove(destPath)
			return domain.FileRecord{}, fmt.Errorf("record upload: %w", err)
		}
	}

	return rec, nil
}
