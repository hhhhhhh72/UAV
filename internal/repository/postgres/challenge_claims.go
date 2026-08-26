package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)

// pgClaimRepo 揭榜意向（rd_challenge_claims）PG 实现。
type pgClaimRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewChallengeClaimRepository() repository.ChallengeClaimRepository {
	return &pgClaimRepo{pool: s.Pool()}
}

func (r *pgClaimRepo) Create(ctx context.Context, c domain.ChallengeClaim) (domain.ChallengeClaim, error) {
	c.CreatedAt = time.Now()
	_, err := r.pool.Exec(ctx,
		`INSERT INTO rd_challenge_claims (id,challenge_id,user_id,status,created_at) VALUES ($1,$2,$3,$4,$5)`,
		c.ID, c.ChallengeID, c.UserID, c.Status, c.CreatedAt)
	if err != nil {
		// 并发重复揭榜由唯一索引兜底：幂等返回已有记录（service 层再给 409 语义校验）。
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			existing, found, ferr := r.FindByChallengeAndUser(ctx, c.ChallengeID, c.UserID)
			if ferr == nil && found {
				return existing, nil
			}
			return domain.ChallengeClaim{}, ferr
		}
		return domain.ChallengeClaim{}, fmt.Errorf("insert claim: %w", err)
	}
	return c, nil
}

func (r *pgClaimRepo) FindByChallengeAndUser(ctx context.Context, challengeID, userID string) (domain.ChallengeClaim, bool, error) {
	var c domain.ChallengeClaim
	err := r.pool.QueryRow(ctx,
		`SELECT id,challenge_id,user_id,status,created_at FROM rd_challenge_claims WHERE challenge_id=$1 AND user_id=$2`,
		challengeID, userID).
		Scan(&c.ID, &c.ChallengeID, &c.UserID, &c.Status, &c.CreatedAt)
	if err != nil {
		return domain.ChallengeClaim{}, false, nil // 未找到：非错误
	}
	return c, true, nil
}

func (r *pgClaimRepo) ListByChallenge(ctx context.Context, challengeID string) ([]domain.ChallengeClaim, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id,challenge_id,user_id,status,created_at FROM rd_challenge_claims WHERE challenge_id=$1 ORDER BY created_at DESC`,
		challengeID)
	if err != nil {
		return nil, fmt.Errorf("list claims: %w", err)
	}
	defer rows.Close()
	out := make([]domain.ChallengeClaim, 0)
	for rows.Next() {
		var c domain.ChallengeClaim
		if err := rows.Scan(&c.ID, &c.ChallengeID, &c.UserID, &c.Status, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan claim: %w", err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
