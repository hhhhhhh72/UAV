package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)

type uploadRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewUploadRepository() repository.UploadRepository {
	return &uploadRepo{pool: s.Pool()}
}

func (r *uploadRepo) Create(ctx context.Context, rec domain.FileRecord) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO uploads (id, owner_id, storage_key, sha256, content_type, size_bytes, visibility, created_at, width, height)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		rec.ID, rec.OwnerID, rec.StorageKey, rec.SHA256, rec.ContentType, rec.SizeBytes, rec.Visibility, rec.CreatedAt, rec.Width, rec.Height)
	if err != nil {
		return fmt.Errorf("record upload %s: %w", rec.ID, err)
	}
	return nil
}

// uploadColumns 列顺序与 scanUpload 严格对齐（新增列时两处必须同步，否则静默串列）。
const uploadColumns = `id, owner_id, COALESCE(storage_key,''), COALESCE(sha256,''), COALESCE(content_type,''), size_bytes, visibility, created_at, width, height`

func scanUpload(row interface {
	Scan(dest ...any) error
}) (domain.FileRecord, error) {
	var rec domain.FileRecord
	err := row.Scan(&rec.ID, &rec.OwnerID, &rec.StorageKey, &rec.SHA256, &rec.ContentType,
		&rec.SizeBytes, &rec.Visibility, &rec.CreatedAt, &rec.Width, &rec.Height)
	return rec, err
}

func (r *uploadRepo) FindByID(ctx context.Context, id string) (domain.FileRecord, error) {
	rec, err := scanUpload(r.pool.QueryRow(ctx,
		`SELECT `+uploadColumns+` FROM uploads WHERE id=$1`, id))
	if err != nil {
		return domain.FileRecord{}, fmt.Errorf("find upload %s: %w", id, err)
	}
	return rec, nil
}

// ListByOwner 某用户的上传台账（注销账号时用于删除磁盘上的物理文件）。
func (r *uploadRepo) ListByOwner(ctx context.Context, ownerID string) ([]domain.FileRecord, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+uploadColumns+` FROM uploads WHERE owner_id=$1 ORDER BY created_at`,
		ownerID)
	if err != nil {
		return nil, fmt.Errorf("list uploads for %s: %w", ownerID, err)
	}
	defer rows.Close()
	out := []domain.FileRecord{}
	for rows.Next() {
		rec, err := scanUpload(rows)
		if err != nil {
			return nil, fmt.Errorf("scan upload: %w", err)
		}
		out = append(out, rec)
	}
	return out, rows.Err()
}

// FindByIDs 按 ID 批量查（商品卡片拿封面图宽高用）。空入参直接返回，避免拼出 `IN ()` 这种非法 SQL。
func (r *uploadRepo) FindByIDs(ctx context.Context, ids []string) ([]domain.FileRecord, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	rows, err := r.pool.Query(ctx, `SELECT `+uploadColumns+` FROM uploads WHERE id = ANY($1)`, ids)
	if err != nil {
		return nil, fmt.Errorf("find uploads by ids: %w", err)
	}
	defer rows.Close()
	var out []domain.FileRecord
	for rows.Next() {
		rec, err := scanUpload(rows)
		if err != nil {
			return nil, fmt.Errorf("scan upload: %w", err)
		}
		out = append(out, rec)
	}
	return out, rows.Err()
}
func (r *uploadRepo) SumBytesSince(ctx context.Context, ownerID string, since time.Time) (int64, error) {
	var sum int64
	if err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(size_bytes), 0) FROM uploads WHERE owner_id=$1 AND created_at >= $2`,
		ownerID, since).Scan(&sum); err != nil {
		return 0, fmt.Errorf("sum upload bytes for %s: %w", ownerID, err)
	}
	return sum, nil
}
