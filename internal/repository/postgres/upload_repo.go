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
		`INSERT INTO uploads (id, owner_id, storage_key, sha256, content_type, size_bytes, visibility, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		rec.ID, rec.OwnerID, rec.StorageKey, rec.SHA256, rec.ContentType, rec.SizeBytes, rec.Visibility, rec.CreatedAt)
	if err != nil {
		return fmt.Errorf("record upload %s: %w", rec.ID, err)
	}
	return nil
}

func (r *uploadRepo) FindByID(ctx context.Context, id string) (domain.FileRecord, error) {
	var rec domain.FileRecord
	err := r.pool.QueryRow(ctx,
		`SELECT id, owner_id, COALESCE(storage_key,''), COALESCE(sha256,''), COALESCE(content_type,''), size_bytes, visibility, created_at FROM uploads WHERE id=$1`,
		id).Scan(&rec.ID, &rec.OwnerID, &rec.StorageKey, &rec.SHA256, &rec.ContentType, &rec.SizeBytes, &rec.Visibility, &rec.CreatedAt)
	if err != nil {
		return domain.FileRecord{}, fmt.Errorf("find upload %s: %w", id, err)
	}
	return rec, nil
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
