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
	"strings"
	"sync"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)

// ErrUploadQuotaExceeded 当日上传配额已用尽（handler 映射为 413）。
var ErrUploadQuotaExceeded = errors.New("今日上传额度已用尽")

// uploadQuotaMu 串行化配额路径的"检查配额→写盘→记账"（P2 修复）。
// 此前先写盘→查账→记账非原子：并发同用户上传会各自读到旧用量再各自记账，
// 放大当日配额（如 2 并发各写 6MB、配额 10MB 双双通过）。上传是低频操作，
// 全局互斥足够且实现最简单；多实例部署时进程内锁不跨实例（注释说明局限，
// 当前为单实例部署；跨实例需依赖数据库层原子记账）。
var uploadQuotaMu sync.Mutex

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

// FindUpload 查询上传台账（私有文件归属校验用）。
// 未启用台账（uploads == nil）或查无记录时返回 not found——调用方须 fail closed。
func (s *FileService) FindUpload(ctx context.Context, id string) (domain.FileRecord, error) {
	if s.uploads == nil {
		return domain.FileRecord{}, fmt.Errorf("upload ledger disabled")
	}
	return s.uploads.FindByID(ctx, id)
}

// RemoveByStorageKey 删除一个上传文件（按台账里的 StorageKey）。
// 幂等：文件已不存在视为成功；路径越界（不在 uploadDir 内）拒绝执行——
// 台账里的 key 是历史数据，不能被用来删上传目录以外的文件。
func (s *FileService) RemoveByStorageKey(key string) error {
	if key == "" {
		return nil
	}
	clean := filepath.Clean(key)
	base := filepath.Clean(s.uploadDir)
	if clean != base && !strings.HasPrefix(clean, base+string(os.PathSeparator)) {
		return fmt.Errorf("refuse to remove file outside upload dir: %s", key)
	}
	if err := os.Remove(clean); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove file %s: %w", clean, err)
	}
	return nil
}

// RemoveFilesForOwner 删除某用户上传的全部物理文件（磁盘），返回删除个数。
// 注销账号时调用：台账行由内容处置计划删除，这里负责把磁盘上的副本也真删掉
//（个人信息注销后不应留副本）。单个文件删除失败即返回，由调用方记日志、缓冲期任务兜底重试。
func (s *FileService) RemoveFilesForOwner(ctx context.Context, ownerID string) (int, error) {
	if s.uploads == nil || ownerID == "" {
		return 0, nil
	}
	recs, err := s.uploads.ListByOwner(ctx, ownerID)
	if err != nil {
		return 0, fmt.Errorf("list uploads of %s: %w", ownerID, err)
	}
	removed := 0
	for _, rec := range recs {
		if err := s.RemoveByStorageKey(rec.StorageKey); err != nil {
			return removed, err
		}
		removed++
	}
	return removed, nil
}
func (s *FileService) uploadTo(ctx context.Context, ownerID string, filename, contentType string, reader io.Reader, dir, visibility string) (domain.FileRecord, error) {
	now := time.Now()
	// P2 修复：启用配额时整个"检查配额→写盘→记账"串行化（全局互斥），
	// 防并发上传各自读到旧用量放大当日配额；未启用配额不取锁，正常上传零阻塞。
	quotaEnabled := s.dailyLimit > 0 && s.uploads != nil
	if quotaEnabled {
		uploadQuotaMu.Lock()
		defer uploadQuotaMu.Unlock()
	}
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
	// 配额路径整体由 uploadQuotaMu 串行化，并发上传不再放大当日配额。
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
