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
		`INSERT INTO uploads (id, owner_id, size_bytes, visibility, created_at)
		 VALUES ($1,$2,$3,$4,$5)`,
		rec.ID, rec.OwnerID, rec.SizeBytes, rec.Visibility, rec.CreatedAt)
	if err != nil {
		return fmt.Errorf("record upload %s: %w", rec.ID, err)
	}
	return nil
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
