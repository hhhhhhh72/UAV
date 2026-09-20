package postgres

import (
	"context"
	"errors"
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

// CreateWithCapacity 原子报名：同一事务里先锁住研学行（SELECT … FOR UPDATE），
// 再统计活跃人数（pending/approved 的成人+儿童），未超容量才插入。
//
// 为什么必须锁行：研学容量是**求和式**的（不像课程那样有 enrolled_count 计数列能做
// 条件更新），check-then-insert 在两个并发事务里都会读到同一个旧和；而 service 的
// lockByKey 是**进程内**锁，多一个 API 进程就各锁各的（同类问题实测：容量 10 被 200
// 并发收下 87 人）。FOR UPDATE 让同一个团的报名在**数据库层面**排队，多实例也成立。
func (r *pgStudyTourEnrollRepo) CreateWithCapacity(ctx context.Context, e domain.StudyTourEnrollment, headcount, capacityLimit int) (domain.StudyTourEnrollment, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.StudyTourEnrollment{}, fmt.Errorf("begin study enroll tx: %w", err)
	}
	defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()

	var locked string
	if err := tx.QueryRow(ctx, `SELECT id FROM study_tours WHERE id=$1 FOR UPDATE`, e.TourID).Scan(&locked); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.StudyTourEnrollment{}, fmt.Errorf("tour %s not found", e.TourID)
		}
		return domain.StudyTourEnrollment{}, fmt.Errorf("lock tour %s: %w", e.TourID, err)
	}
	if capacityLimit > 0 {
		var taken int
		if err := tx.QueryRow(ctx,
			`SELECT COALESCE(SUM(adult_count + child_count), 0) FROM study_tour_enrollments
			  WHERE tour_id=$1 AND status IN ('pending','approved')`, e.TourID).Scan(&taken); err != nil {
			return domain.StudyTourEnrollment{}, fmt.Errorf("sum study enroll headcount: %w", err)
		}
		if taken+headcount > capacityLimit {
			return domain.StudyTourEnrollment{}, repository.ErrCapacityFull
		}
	}
	e.CreatedAt = time.Now()
	e.UpdatedAt = e.CreatedAt
	if _, err := tx.Exec(ctx,
		`INSERT INTO study_tour_enrollments (id,tour_id,user_id,name,phone,adult_count,child_count,remark,status,created_at,updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		e.ID, e.TourID, e.UserID, e.Name, e.Phone, e.AdultCount, e.ChildCount, e.Remark, e.Status, e.CreatedAt, e.UpdatedAt); err != nil {
		return domain.StudyTourEnrollment{}, fmt.Errorf("insert study enroll: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.StudyTourEnrollment{}, fmt.Errorf("commit study enroll: %w", err)
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
