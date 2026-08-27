package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)

// pgStudyTourEnrollRepo 低空研学报名 PG 实现。
type pgStudyTourEnrollRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewStudyTourEnrollmentRepository() repository.StudyTourEnrollmentRepository {
	return &pgStudyTourEnrollRepo{pool: s.Pool()}
}

func (r *pgStudyTourEnrollRepo) Create(ctx context.Context, e domain.StudyTourEnrollment) (domain.StudyTourEnrollment, error) {
	e.CreatedAt = time.Now()
	e.UpdatedAt = e.CreatedAt
	_, err := r.pool.Exec(ctx,
		`INSERT INTO study_tour_enrollments (id,tour_id,user_id,name,phone,adult_count,child_count,remark,status,created_at,updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		e.ID, e.TourID, e.UserID, e.Name, e.Phone, e.AdultCount, e.ChildCount, e.Remark, e.Status, e.CreatedAt, e.UpdatedAt)
	if err != nil {
		return domain.StudyTourEnrollment{}, fmt.Errorf("insert study enroll: %w", err)
	}
	return e, nil
}

func (r *pgStudyTourEnrollRepo) FindByID(ctx context.Context, id string) (domain.StudyTourEnrollment, error) {
	var e domain.StudyTourEnrollment
	err := r.pool.QueryRow(ctx,
		`SELECT id,tour_id,user_id,COALESCE(name,''),COALESCE(phone,''),adult_count,child_count,COALESCE(remark,''),status,created_at,updated_at FROM study_tour_enrollments WHERE id=$1`, id).
		Scan(&e.ID, &e.TourID, &e.UserID, &e.Name, &e.Phone, &e.AdultCount, &e.ChildCount, &e.Remark, &e.Status, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return domain.StudyTourEnrollment{}, fmt.Errorf("find study enroll %s: %w", id, err)
	}
	return e, nil
}

func (r *pgStudyTourEnrollRepo) ListByUser(ctx context.Context, userID string) ([]domain.StudyTourEnrollment, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id,tour_id,user_id,COALESCE(name,''),COALESCE(phone,''),adult_count,child_count,COALESCE(remark,''),status,created_at,updated_at FROM study_tour_enrollments WHERE user_id=$1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("list study enrolls by user: %w", err)
	}
	defer rows.Close()
	return r.scanRows(rows)
}

func (r *pgStudyTourEnrollRepo) ListByTour(ctx context.Context, tourID string) ([]domain.StudyTourEnrollment, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id,tour_id,user_id,COALESCE(name,''),COALESCE(phone,''),adult_count,child_count,COALESCE(remark,''),status,created_at,updated_at FROM study_tour_enrollments WHERE tour_id=$1 ORDER BY created_at DESC`, tourID)
	if err != nil {
		return nil, fmt.Errorf("list study enrolls by tour: %w", err)
	}
	defer rows.Close()
	return r.scanRows(rows)
}

func (r *pgStudyTourEnrollRepo) UpdateStatus(ctx context.Context, id, status string) (domain.StudyTourEnrollment, error) {
	var e domain.StudyTourEnrollment
	err := r.pool.QueryRow(ctx,
		`UPDATE study_tour_enrollments SET status=$1, updated_at=now() WHERE id=$2 RETURNING id,tour_id,user_id,COALESCE(name,''),COALESCE(phone,''),adult_count,child_count,COALESCE(remark,''),status,created_at,updated_at`,
		status, id).
		Scan(&e.ID, &e.TourID, &e.UserID, &e.Name, &e.Phone, &e.AdultCount, &e.ChildCount, &e.Remark, &e.Status, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return domain.StudyTourEnrollment{}, fmt.Errorf("update study enroll status: %w", err)
	}
	return e, nil
}

func (r *pgStudyTourEnrollRepo) scanRows(rows pgx.Rows) ([]domain.StudyTourEnrollment, error) {
	out := make([]domain.StudyTourEnrollment, 0)
	for rows.Next() {
		var e domain.StudyTourEnrollment
		if err := rows.Scan(&e.ID, &e.TourID, &e.UserID, &e.Name, &e.Phone, &e.AdultCount, &e.ChildCount, &e.Remark, &e.Status, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan study enroll: %w", err)
		}
		out = append(out, e)
	}
	return out, rows.Err()
}
