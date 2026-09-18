package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)

// ---- Inspection ----

type inspectRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewInspectionRepository() repository.InspectionRepository {
	return &inspectRepo{pool: s.Pool()}
}

func (r *inspectRepo) Create(ctx context.Context, i domain.AnnualInspection) (domain.AnnualInspection, error) {
	i.Version = 1
	i.CreatedAt = time.Now()
	i.UpdatedAt = i.CreatedAt
	_, err := r.pool.Exec(ctx,
		`INSERT INTO annual_inspections (id,user_id,drone_model,drone_sn,inspect_date,expire_date,result,report_url,status,version,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		i.ID, i.UserID, i.DroneModel, i.DroneSN, i.InspectDate, i.ExpireDate, i.Result, i.ReportURL, i.Status, i.Version, i.CreatedAt, i.UpdatedAt)
	return i, err
}
func (r *inspectRepo) ListByUser(ctx context.Context, userID string) ([]domain.AnnualInspection, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id,user_id,COALESCE(drone_model,''),COALESCE(drone_sn,''),COALESCE(inspect_date,'1970-01-01 00:00:00+00'::timestamptz),COALESCE(expire_date,'1970-01-01 00:00:00+00'::timestamptz),COALESCE(result,''),COALESCE(report_url,''),status,version,created_at,updated_at FROM annual_inspections WHERE user_id=$1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("list inspections: %w", err)
	}
	defer rows.Close()
	var out []domain.AnnualInspection
	for rows.Next() {
		var i domain.AnnualInspection
		if err := rows.Scan(&i.ID, &i.UserID, &i.DroneModel, &i.DroneSN, &i.InspectDate, &i.ExpireDate, &i.Result, &i.ReportURL, &i.Status, &i.Version, &i.CreatedAt, &i.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan inspection: %w", err)
		}
		out = append(out, i)
	}
	return out, rows.Err()
}
func (r *inspectRepo) ListAll(ctx context.Context) ([]domain.AnnualInspection, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id,user_id,COALESCE(drone_model,''),COALESCE(drone_sn,''),COALESCE(inspect_date,'1970-01-01 00:00:00+00'::timestamptz),COALESCE(expire_date,'1970-01-01 00:00:00+00'::timestamptz),COALESCE(result,''),COALESCE(report_url,''),status,version,created_at,updated_at FROM annual_inspections ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list all inspections: %w", err)
	}
	defer rows.Close()
	var out []domain.AnnualInspection
	for rows.Next() {
		var i domain.AnnualInspection
		if err := rows.Scan(&i.ID, &i.UserID, &i.DroneModel, &i.DroneSN, &i.InspectDate, &i.ExpireDate, &i.Result, &i.ReportURL, &i.Status, &i.Version, &i.CreatedAt, &i.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan inspection: %w", err)
		}
		out = append(out, i)
	}
	return out, rows.Err()
}

// ---- Loan ----

type loanRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewLoanRepository() repository.LoanRepository { return &loanRepo{pool: s.Pool()} }

func (r *loanRepo) Create(ctx context.Context, l domain.LoanApplication) (domain.LoanApplication, error) {
	l.Version = 1
	l.CreatedAt = time.Now()
	l.UpdatedAt = l.CreatedAt
	_, err := r.pool.Exec(ctx,
		`INSERT INTO loan_applications (id,user_id,amount_fen,term_months,purpose,status,approved_fen,monthly_pay_fen,version,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		l.ID, l.UserID, l.AmountFen, l.TermMonths, l.Purpose, l.Status, l.ApprovedFen, l.MonthlyPayFen, l.Version, l.CreatedAt, l.UpdatedAt)
	return l, err
}
func (r *loanRepo) ListByUser(ctx context.Context, userID string) ([]domain.LoanApplication, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id,user_id,amount_fen,term_months,COALESCE(purpose,''),status,approved_fen,monthly_pay_fen,version,created_at,updated_at FROM loan_applications WHERE user_id=$1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("list loans: %w", err)
	}
	defer rows.Close()
	var out []domain.LoanApplication
	for rows.Next() {
		var l domain.LoanApplication
		if err := rows.Scan(&l.ID, &l.UserID, &l.AmountFen, &l.TermMonths, &l.Purpose, &l.Status, &l.ApprovedFen, &l.MonthlyPayFen, &l.Version, &l.CreatedAt, &l.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}
func (r *loanRepo) ListAll(ctx context.Context, offset, limit int) ([]domain.LoanApplication, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT count(*) FROM loan_applications`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.pool.Query(ctx,
		`SELECT id,user_id,amount_fen,term_months,COALESCE(purpose,''),status,approved_fen,monthly_pay_fen,version,created_at,updated_at FROM loan_applications ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []domain.LoanApplication
	for rows.Next() {
		var l domain.LoanApplication
		if err := rows.Scan(&l.ID, &l.UserID, &l.AmountFen, &l.TermMonths, &l.Purpose, &l.Status, &l.ApprovedFen, &l.MonthlyPayFen, &l.Version, &l.CreatedAt, &l.UpdatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, l)
	}
	return out, total, rows.Err()
}

// ---- Message ----

type msgRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewMessageRepository() repository.MessageRepository { return &msgRepo{pool: s.Pool()} }

func (r *msgRepo) Create(ctx context.Context, m domain.Message) (domain.Message, error) {
	m.CreatedAt = time.Now()
	_, err := r.pool.Exec(ctx,
		`INSERT INTO messages (id,sender_id,receiver_id,title,content,resource_type,resource_id,is_read,created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		m.ID, m.SenderID, m.ReceiverID, m.Title, m.Content, m.ResourceType, m.ResourceID, m.IsRead, m.CreatedAt)
	return m, err
}
func (r *msgRepo) FindByID(ctx context.Context, id string) (domain.Message, error) {
	var m domain.Message
	err := r.pool.QueryRow(ctx,
		`SELECT id,sender_id,receiver_id,title,COALESCE(content,''),COALESCE(resource_type,''),COALESCE(resource_id,''),is_read,created_at FROM messages WHERE id=$1`, id).
		Scan(&m.ID, &m.SenderID, &m.ReceiverID, &m.Title, &m.Content, &m.ResourceType, &m.ResourceID, &m.IsRead, &m.CreatedAt)
	if err != nil {
		return domain.Message{}, fmt.Errorf("find message %s: %w", id, err)
	}
	return m, nil
}
func (r *msgRepo) ListByUser(ctx context.Context, userID string, unreadOnly bool) ([]domain.Message, error) {
	q := `SELECT id,sender_id,receiver_id,title,COALESCE(content,''),COALESCE(resource_type,''),COALESCE(resource_id,''),is_read,created_at FROM messages WHERE receiver_id=$1`
	if unreadOnly {
		q += ` AND is_read=false`
	}
	q += ` ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("list messages: %w", err)
	}
	defer rows.Close()
	var out []domain.Message
	for rows.Next() {
		var m domain.Message
		if err := rows.Scan(&m.ID, &m.SenderID, &m.ReceiverID, &m.Title, &m.Content, &m.ResourceType, &m.ResourceID, &m.IsRead, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}
		out = append(out, m)
	}
	return out, rows.Err()
}
func (r *msgRepo) MarkRead(ctx context.Context, id string) (domain.Message, error) {
	_, err := r.pool.Exec(ctx, `UPDATE messages SET is_read=true WHERE id=$1`, id)
	if err != nil {
		return domain.Message{}, fmt.Errorf("mark message read: %w", err)
	}
	var m domain.Message
	err = r.pool.QueryRow(ctx,
		`SELECT id,sender_id,receiver_id,title,COALESCE(content,''),COALESCE(resource_type,''),COALESCE(resource_id,''),is_read,created_at FROM messages WHERE id=$1`, id).
		Scan(&m.ID, &m.SenderID, &m.ReceiverID, &m.Title, &m.Content, &m.ResourceType, &m.ResourceID, &m.IsRead, &m.CreatedAt)
	return m, err
}
func (r *msgRepo) UnreadCount(ctx context.Context, userID string) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM messages WHERE receiver_id=$1 AND is_read=false`, userID).Scan(&n)
	return n, err
}

func (r *msgRepo) ListAll(ctx context.Context, offset, limit int) ([]domain.Message, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, "SELECT count(*) FROM messages").Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count messages: %w", err)
	}
	rows, err := r.pool.Query(ctx, "SELECT id,sender_id,receiver_id,title,COALESCE(content,''),COALESCE(resource_type,''),COALESCE(resource_id,''),is_read,created_at FROM messages ORDER BY created_at DESC LIMIT $1 OFFSET $2", limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list all messages: %w", err)
	}
	defer rows.Close()
	var out []domain.Message
	for rows.Next() {
		var m domain.Message
		if err := rows.Scan(&m.ID, &m.SenderID, &m.ReceiverID, &m.Title, &m.Content, &m.ResourceType, &m.ResourceID, &m.IsRead, &m.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan message: %w", err)
		}
		out = append(out, m)
	}
	return out, total, rows.Err()
}

func (r *msgRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM messages WHERE id=$1", id)
	return err
}

// DeleteByReference 按业务对象批量删消息。resource_type = ANY($2) 与 resource_id
// 同时限定：只按 resource_id 删，万一别的类型复用了同一个 id 会误删。
func (r *msgRepo) DeleteByReference(ctx context.Context, resourceID string, resourceTypes []string) (int, error) {
	if resourceID == "" || len(resourceTypes) == 0 {
		return 0, nil // 不给空条件：那会删掉管理端广播（resource_type='' 且 resource_id=''）
	}
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM messages WHERE resource_id=$1 AND resource_type = ANY($2)`, resourceID, resourceTypes)
	if err != nil {
		return 0, fmt.Errorf("delete messages by reference %s: %w", resourceID, err)
	}
	return int(tag.RowsAffected()), nil
}

// ---- Article ----

type articleRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewArticleRepository() repository.ArticleRepository {
	return &articleRepo{pool: s.Pool()}
}

func (r *articleRepo) Create(ctx context.Context, a domain.Article) (domain.Article, error) {
	a.Version = 1
	a.CreatedAt = time.Now()
	a.UpdatedAt = a.CreatedAt
	_, err := r.pool.Exec(ctx,
		`INSERT INTO articles (id,title,content,summary,category,source,author,is_pinned,status,version,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		a.ID, a.Title, a.Content, a.Summary, a.Category, a.Source, a.Author, a.IsPinned, a.Status, a.Version, a.CreatedAt, a.UpdatedAt)
	return a, err
}
func (r *articleRepo) FindByID(ctx context.Context, id string) (domain.Article, error) {
	var a domain.Article
	err := r.pool.QueryRow(ctx,
		`SELECT id,title,COALESCE(content,''),COALESCE(summary,''),category,COALESCE(source,''),COALESCE(author,''),COALESCE(is_pinned,false),status,version,created_at,updated_at FROM articles WHERE id=$1`, id).
		Scan(&a.ID, &a.Title, &a.Content, &a.Summary, &a.Category, &a.Source, &a.Author, &a.IsPinned, &a.Status, &a.Version, &a.CreatedAt, &a.UpdatedAt)
	return a, err
}
func (r *articleRepo) Update(ctx context.Context, a domain.Article) (domain.Article, error) {
	a.UpdatedAt = time.Now()
	a.Version++
	_, err := r.pool.Exec(ctx,
		`UPDATE articles SET title=$1,content=$2,summary=$3,category=$4,source=$5,author=$6,is_pinned=$7,status=$8,version=$9,updated_at=$10 WHERE id=$11`,
		a.Title, a.Content, a.Summary, a.Category, a.Source, a.Author, a.IsPinned, a.Status, a.Version, a.UpdatedAt, a.ID)
	return a, err
}
func (r *articleRepo) ListByCategory(ctx context.Context, category string, offset, limit int) ([]domain.Article, int, error) {
	where := ""
	args := []any{}
	if category != "" {
		where = `WHERE category=$1`
		args = append(args, category)
	}
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM articles `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	q := `SELECT id,title,COALESCE(content,''),COALESCE(summary,''),category,COALESCE(source,''),COALESCE(author,''),COALESCE(is_pinned,false),status,version,created_at,updated_at FROM articles ` + where + ` ORDER BY is_pinned DESC, created_at DESC LIMIT $` + fmt.Sprintf("%d", len(args)+1) + ` OFFSET $` + fmt.Sprintf("%d", len(args)+2)
	allArgs := append(args, limit, offset)
	rows, err := r.pool.Query(ctx, q, allArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("list articles: %w", err)
	}
	defer rows.Close()
	var out []domain.Article
	for rows.Next() {
		var a domain.Article
		if err := rows.Scan(&a.ID, &a.Title, &a.Content, &a.Summary, &a.Category, &a.Source, &a.Author, &a.IsPinned, &a.Status, &a.Version, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan article: %w", err)
		}
		out = append(out, a)
	}
	return out, total, rows.Err()
}

func (r *articleRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM articles WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("delete article %s: %w", id, err)
	}
	return nil
}

// ---- Review ----

type reviewRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewReviewRepository() repository.ReviewRepository { return &reviewRepo{pool: s.Pool()} }

func (r *reviewRepo) Create(ctx context.Context, rv domain.Review) (domain.Review, error) {
	rv.CreatedAt = time.Now()
	_, err := r.pool.Exec(ctx,
		`INSERT INTO reviews (id,reviewer_id,target_type,target_id,rating,content,status,created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		rv.ID, rv.ReviewerID, rv.TargetType, rv.TargetID, rv.Rating, rv.Content, rv.Status, rv.CreatedAt)
	return rv, err
}
func (r *reviewRepo) ListByTarget(ctx context.Context, targetType, targetID string) ([]domain.Review, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id,reviewer_id,target_type,target_id,rating,COALESCE(content,''),status,created_at FROM reviews WHERE target_type=$1 AND target_id=$2 AND status='approved' ORDER BY created_at DESC`, targetType, targetID)
	if err != nil {
		return nil, fmt.Errorf("list reviews: %w", err)
	}
	defer rows.Close()
	var out []domain.Review
	for rows.Next() {
		var rv domain.Review
		if err := rows.Scan(&rv.ID, &rv.ReviewerID, &rv.TargetType, &rv.TargetID, &rv.Rating, &rv.Content, &rv.Status, &rv.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan review: %w", err)
		}
		out = append(out, rv)
	}
	return out, rows.Err()
}

// ListByReviewerTarget 查重专用：**不按 status 过滤**（含 pending/rejected）。
func (r *reviewRepo) ListByReviewerTarget(ctx context.Context, reviewerID, targetType, targetID string) ([]domain.Review, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id,reviewer_id,target_type,target_id,rating,COALESCE(content,''),status,created_at FROM reviews WHERE reviewer_id=$1 AND target_type=$2 AND target_id=$3 ORDER BY created_at DESC`, reviewerID, targetType, targetID)
	if err != nil {
		return nil, fmt.Errorf("list reviews by reviewer: %w", err)
	}
	defer rows.Close()
	var out []domain.Review
	for rows.Next() {
		var rv domain.Review
		if err := rows.Scan(&rv.ID, &rv.ReviewerID, &rv.TargetType, &rv.TargetID, &rv.Rating, &rv.Content, &rv.Status, &rv.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan review: %w", err)
		}
		out = append(out, rv)
	}
	return out, rows.Err()
}

func (r *reviewRepo) ListAll(ctx context.Context, status string, offset, limit int) ([]domain.Review, int, error) {
	where := ""
	args := []any{}
	if status != "" {
		where = `WHERE status=$1`
		args = append(args, status)
	}
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM reviews `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	q := `SELECT id,reviewer_id,target_type,target_id,rating,COALESCE(content,''),status,created_at FROM reviews ` + where + ` ORDER BY created_at DESC LIMIT $` + fmt.Sprintf("%d", len(args)+1) + ` OFFSET $` + fmt.Sprintf("%d", len(args)+2)
	allArgs := append(args, limit, offset)
	rows, err := r.pool.Query(ctx, q, allArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("list all reviews: %w", err)
	}
	defer rows.Close()
	var out []domain.Review
	for rows.Next() {
		var rv domain.Review
		if err := rows.Scan(&rv.ID, &rv.ReviewerID, &rv.TargetType, &rv.TargetID, &rv.Rating, &rv.Content, &rv.Status, &rv.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan review: %w", err)
		}
		out = append(out, rv)
	}
	return out, total, rows.Err()
}
func (r *reviewRepo) FindByID(ctx context.Context, id string) (domain.Review, error) {
	var rv domain.Review
	err := r.pool.QueryRow(ctx,
		`SELECT id,reviewer_id,target_type,target_id,rating,COALESCE(content,''),status,created_at FROM reviews WHERE id=$1`, id).
		Scan(&rv.ID, &rv.ReviewerID, &rv.TargetType, &rv.TargetID, &rv.Rating, &rv.Content, &rv.Status, &rv.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Review{}, fmt.Errorf("review %s: %w", id, repository.ErrNotFound)
	}
	if err != nil {
		return domain.Review{}, fmt.Errorf("review %s: %w", id, err)
	}
	return rv, nil
}
func (r *reviewRepo) UpdateStatus(ctx context.Context, id, status string) (domain.Review, error) {
	_, err := r.pool.Exec(ctx, `UPDATE reviews SET status=$1 WHERE id=$2`, status, id)
	if err != nil {
		return domain.Review{}, fmt.Errorf("update review status: %w", err)
	}
	var rv domain.Review
	err = r.pool.QueryRow(ctx,
		`SELECT id,reviewer_id,target_type,target_id,rating,COALESCE(content,''),status,created_at FROM reviews WHERE id=$1`, id).
		Scan(&rv.ID, &rv.ReviewerID, &rv.TargetType, &rv.TargetID, &rv.Rating, &rv.Content, &rv.Status, &rv.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		// UPDATE 影响 0 行（记录不存在）后复查也会落到这里
		return domain.Review{}, fmt.Errorf("review %s: %w", id, repository.ErrNotFound)
	}
	return rv, err
}
func (r *reviewRepo) Delete(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM reviews WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("delete review %s: %w", id, err)
	}
	// 删了 0 行 = 记录不存在：返回哨兵让上层回 404，而不是"删了个不存在的 id 却回 200"
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("review %s: %w", id, repository.ErrNotFound)
	}
	return nil
}

// ---- Venue ----

type venueRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewVenueRepository() repository.VenueRepository { return &venueRepo{pool: s.Pool()} }

func (r *venueRepo) Create(ctx context.Context, v domain.Venue) (domain.Venue, error) {
	v.CreatedAt = time.Now()
	_, err := r.pool.Exec(ctx,
		`INSERT INTO venues (id,owner_id,name,venue_type,location,price_fen,status,created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		v.ID, v.OwnerID, v.Name, v.VenueType, v.Location, v.PriceFen, v.Status, v.CreatedAt)
	return v, err
}
func (r *venueRepo) List(ctx context.Context, venueType string) ([]domain.Venue, error) {
	var rows pgx.Rows
	var err error
	if venueType == "" {
		rows, err = r.pool.Query(ctx, `SELECT id,owner_id,name,venue_type,COALESCE(location,''),COALESCE(price_fen,0),status,created_at FROM venues ORDER BY created_at DESC`)
	} else {
		rows, err = r.pool.Query(ctx, `SELECT id,owner_id,name,venue_type,COALESCE(location,''),COALESCE(price_fen,0),status,created_at FROM venues WHERE venue_type=$1 ORDER BY created_at DESC`, venueType)
	}
	if err != nil {
		return nil, fmt.Errorf("list venues: %w", err)
	}
	defer rows.Close()
	var out []domain.Venue
	for rows.Next() {
		var v domain.Venue
		if err := rows.Scan(&v.ID, &v.OwnerID, &v.Name, &v.VenueType, &v.Location, &v.PriceFen, &v.Status, &v.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan venue: %w", err)
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
func (r *venueRepo) FindByID(ctx context.Context, id string) (domain.Venue, error) {
	var v domain.Venue
	err := r.pool.QueryRow(ctx,
		`SELECT id,owner_id,name,venue_type,COALESCE(location,''),COALESCE(price_fen,0),status,created_at FROM venues WHERE id=$1`, id).
		Scan(&v.ID, &v.OwnerID, &v.Name, &v.VenueType, &v.Location, &v.PriceFen, &v.Status, &v.CreatedAt)
	return v, err
}
func (r *venueRepo) CreateBooking(ctx context.Context, b domain.VenueBooking) (domain.VenueBooking, error) {
	b.CreatedAt = time.Now()
	_, err := r.pool.Exec(ctx,
		`INSERT INTO venue_bookings (id,venue_id,user_id,start_time,end_time,status,created_at) VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		b.ID, b.VenueID, b.UserID, b.StartTime, b.EndTime, b.Status, b.CreatedAt)
	return b, err
}
func (r *venueRepo) ListBookings(ctx context.Context, venueID string) ([]domain.VenueBooking, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id,venue_id,user_id,start_time,end_time,status,created_at FROM venue_bookings WHERE venue_id=$1 ORDER BY created_at DESC`, venueID)
	if err != nil {
		return nil, fmt.Errorf("list bookings: %w", err)
	}
	defer rows.Close()
	var out []domain.VenueBooking
	for rows.Next() {
		var b domain.VenueBooking
		if err := rows.Scan(&b.ID, &b.VenueID, &b.UserID, &b.StartTime, &b.EndTime, &b.Status, &b.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan venue booking: %w", err)
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

// ---- Enrollment ----

type enrollRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewEnrollmentRepository() repository.EnrollmentRepository {
	return &enrollRepo{pool: s.Pool()}
}

func (r *enrollRepo) Create(ctx context.Context, e domain.Enrollment) (domain.Enrollment, error) {
	e.CreatedAt = time.Now()
	_, err := r.pool.Exec(ctx,
		`INSERT INTO training_enrollments (id,course_id,user_id,name,phone,id_card,gender,birthday,email,education,experience,photo_url,id_card_image,id_card_back,no_crime,status,paid_amount_fen,review_note,created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)`,
		e.ID, e.CourseID, e.UserID, e.Name, e.Phone, e.IDCard, e.Gender, e.Birthday, e.Email, e.Education, e.Experience, e.PhotoURL, e.IDCardImage, e.IDCardBack, e.NoCrime, e.Status, e.PaidAmountFen, e.ReviewNote, e.CreatedAt)
	if err != nil {
		// P1 修复：唯一索引 (user_id, course_id) 并发兜底——重复报名映射为友好错误。
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.Enrollment{}, fmt.Errorf("已报名过该课程，请勿重复报名")
		}
		return domain.Enrollment{}, fmt.Errorf("create enrollment: %w", err)
	}
	return e, nil
}

func (r *enrollRepo) Update(ctx context.Context, e domain.Enrollment) (domain.Enrollment, error) {
	_, err := r.pool.Exec(ctx,
		`UPDATE training_enrollments SET name=$1,phone=$2,id_card=$3,gender=$4,birthday=$5,email=$6,education=$7,experience=$8,photo_url=$9,id_card_image=$10,id_card_back=$11,no_crime=$12,status=$13,paid_amount_fen=$14,review_note=$15 WHERE id=$16`,
		e.Name, e.Phone, e.IDCard, e.Gender, e.Birthday, e.Email, e.Education, e.Experience, e.PhotoURL, e.IDCardImage, e.IDCardBack, e.NoCrime, e.Status, e.PaidAmountFen, e.ReviewNote, e.ID)
	if err != nil {
		return domain.Enrollment{}, fmt.Errorf("update enrollment %s: %w", e.ID, err)
	}
	return e, nil
}

// UpdateStatusCas 原子状态迁移：仅当前状态 == from 时改为 to（completed 终态 CAS，
// 防并发完成报名的双释放学费——此前盲写 WHERE id= 会让并发双请求都执行 Release）。
func (r *enrollRepo) UpdateStatusCas(ctx context.Context, id, from, to string) (bool, error) {
	tag, err := r.pool.Exec(ctx,
		`UPDATE training_enrollments SET status=$3 WHERE id=$1 AND status=$2`, id, from, to)
	if err != nil {
		return false, fmt.Errorf("update enrollment %s status: %w", id, err)
	}
	return tag.RowsAffected() > 0, nil
}

// UpdateReviewNote 只写审核备注列。审核必须先 CAS 改状态、再用本方法补备注，
// 绝不能用全列 Update 回写审核前的整行快照——那会把并发完成报名写入的
// status=completed 与 paid_amount_fen 覆盖回 approved/rejected，绕过 completed 终态保护。
func (r *enrollRepo) UpdateReviewNote(ctx context.Context, id, note string) (domain.Enrollment, error) {
	if _, err := r.pool.Exec(ctx,
		`UPDATE training_enrollments SET review_note=$2 WHERE id=$1`, id, note); err != nil {
		return domain.Enrollment{}, fmt.Errorf("update enrollment %s review note: %w", id, err)
	}
	return r.FindByID(ctx, id)
}
func (r *enrollRepo) ListAll(ctx context.Context, offset, limit int) ([]domain.Enrollment, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT count(*) FROM training_enrollments`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count enrollments: %w", err)
	}
	rows, err := r.pool.Query(ctx,
		`SELECT id,course_id,user_id,name,phone,id_card,gender,birthday,email,education,experience,photo_url,id_card_image,id_card_back,no_crime,status,paid_amount_fen,review_note,created_at FROM training_enrollments ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list all enrollments: %w", err)
	}
	defer rows.Close()
	var out []domain.Enrollment
	for rows.Next() {
		var e domain.Enrollment
		if err := rows.Scan(&e.ID, &e.CourseID, &e.UserID, &e.Name, &e.Phone, &e.IDCard, &e.Gender, &e.Birthday, &e.Email, &e.Education, &e.Experience, &e.PhotoURL, &e.IDCardImage, &e.IDCardBack, &e.NoCrime, &e.Status, &e.PaidAmountFen, &e.ReviewNote, &e.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan enrollment: %w", err)
		}
		out = append(out, e)
	}
	return out, total, rows.Err()
}
func (r *enrollRepo) ListByCourse(ctx context.Context, courseID string) ([]domain.Enrollment, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id,course_id,user_id,name,phone,id_card,gender,birthday,email,education,experience,photo_url,id_card_image,id_card_back,no_crime,status,paid_amount_fen,review_note,created_at FROM training_enrollments WHERE course_id=$1 ORDER BY created_at DESC`, courseID)
	if err != nil {
		return nil, fmt.Errorf("list enrollments: %w", err)
	}
	defer rows.Close()
	var out []domain.Enrollment
	for rows.Next() {
		var e domain.Enrollment
		if err := rows.Scan(&e.ID, &e.CourseID, &e.UserID, &e.Name, &e.Phone, &e.IDCard, &e.Gender, &e.Birthday, &e.Email, &e.Education, &e.Experience, &e.PhotoURL, &e.IDCardImage, &e.IDCardBack, &e.NoCrime, &e.Status, &e.PaidAmountFen, &e.ReviewNote, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan enrollment: %w", err)
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// ListByUser 某用户全部报名（"我的报名"一次查询，避免按课程 N+1）。
func (r *enrollRepo) ListByUser(ctx context.Context, userID string) ([]domain.Enrollment, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id,course_id,user_id,name,phone,id_card,gender,birthday,email,education,experience,photo_url,id_card_image,id_card_back,no_crime,status,paid_amount_fen,review_note,created_at FROM training_enrollments WHERE user_id=$1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("list enrollments by user: %w", err)
	}
	defer rows.Close()
	var out []domain.Enrollment
	for rows.Next() {
		var e domain.Enrollment
		if err := rows.Scan(&e.ID, &e.CourseID, &e.UserID, &e.Name, &e.Phone, &e.IDCard, &e.Gender, &e.Birthday, &e.Email, &e.Education, &e.Experience, &e.PhotoURL, &e.IDCardImage, &e.IDCardBack, &e.NoCrime, &e.Status, &e.PaidAmountFen, &e.ReviewNote, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan enrollment: %w", err)
		}
		out = append(out, e)
	}
	return out, rows.Err()
}
func (r *enrollRepo) FindByUserAndCourse(ctx context.Context, userID, courseID string) (domain.Enrollment, bool, error) {
	var e domain.Enrollment
	err := r.pool.QueryRow(ctx,
		`SELECT id,course_id,user_id,name,phone,id_card,gender,birthday,email,education,experience,photo_url,id_card_image,id_card_back,no_crime,status,paid_amount_fen,review_note,created_at FROM training_enrollments WHERE user_id=$1 AND course_id=$2`, userID, courseID).
		Scan(&e.ID, &e.CourseID, &e.UserID, &e.Name, &e.Phone, &e.IDCard, &e.Gender, &e.Birthday, &e.Email, &e.Education, &e.Experience, &e.PhotoURL, &e.IDCardImage, &e.IDCardBack, &e.NoCrime, &e.Status, &e.PaidAmountFen, &e.ReviewNote, &e.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Enrollment{}, false, nil
		}
		return domain.Enrollment{}, false, fmt.Errorf("find enrollment by user and course: %w", err)
	}
	return e, true, nil
}
func (r *enrollRepo) FindByID(ctx context.Context, id string) (domain.Enrollment, error) {
	var e domain.Enrollment
	err := r.pool.QueryRow(ctx,
		`SELECT id,course_id,user_id,name,phone,id_card,gender,birthday,email,education,experience,photo_url,id_card_image,id_card_back,no_crime,status,paid_amount_fen,review_note,created_at FROM training_enrollments WHERE id=$1`, id).
		Scan(&e.ID, &e.CourseID, &e.UserID, &e.Name, &e.Phone, &e.IDCard, &e.Gender, &e.Birthday, &e.Email, &e.Education, &e.Experience, &e.PhotoURL, &e.IDCardImage, &e.IDCardBack, &e.NoCrime, &e.Status, &e.PaidAmountFen, &e.ReviewNote, &e.CreatedAt)
	if err != nil {
		return domain.Enrollment{}, fmt.Errorf("enrollment %s not found: %w", id, err)
	}
	return e, nil
}

// ---- TradeOrder ----

type tradeOrderRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewTradeOrderRepository() repository.TradeOrderRepository {
	return &tradeOrderRepo{pool: s.Pool()}
}

func (r *tradeOrderRepo) Create(ctx context.Context, o domain.TradeOrder) (domain.TradeOrder, error) {
	o.Version = 1
	o.CreatedAt = time.Now()
	o.UpdatedAt = o.CreatedAt
	_, err := r.pool.Exec(ctx,
		`INSERT INTO trade_orders (id,product_id,buyer_id,seller_id,amount_fen,status,receiver_name,receiver_phone,receiver_region,receiver_address,version,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		o.ID, o.ProductID, o.BuyerID, o.SellerID, o.AmountFen, o.Status,
		o.ReceiverName, o.ReceiverPhone, o.ReceiverRegion, o.ReceiverAddress,
		o.Version, o.CreatedAt, o.UpdatedAt)
	return o, err
}

// tradeOrderColumns trade_orders 查询列（含收货/发货/售后字段），各查询统一复用。
// tradeOrderColumns 必须与 scanTradeOrder 的 Scan 目标严格同序同数（当前各 26 个）。
// 二者是手工维护的同一份契约——历史上 demands 的收藏查询就因少写 4 列导致生产必 500。
const tradeOrderColumns = `id,product_id,buyer_id,seller_id,amount_fen,status,receiver_name,receiver_phone,receiver_region,receiver_address,shipping_company,shipping_tracking,shipped_at,aftersale_type,aftersale_reason,aftersale_desc,aftersale_amount_fen,aftersale_status,aftersale_time,aftersale_from,return_tracking,return_note,returned_at,version,created_at,updated_at`

func scanTradeOrder(row interface{ Scan(...any) error }) (domain.TradeOrder, error) {
	var o domain.TradeOrder
	var returnedAt *time.Time // returned_at 可空（未寄回退货时为 NULL）
	var shippedAt *time.Time  // shipped_at 可空（未发货时为 NULL）
	err := row.Scan(&o.ID, &o.ProductID, &o.BuyerID, &o.SellerID, &o.AmountFen, &o.Status,
		&o.ReceiverName, &o.ReceiverPhone, &o.ReceiverRegion, &o.ReceiverAddress,
		&o.ShippingCompany, &o.ShippingTracking, &shippedAt,
		&o.AftersaleType, &o.AftersaleReason, &o.AftersaleDesc, &o.AftersaleAmountFen, &o.AftersaleStatus, &o.AftersaleTime,
		&o.AftersaleFrom, &o.ReturnTracking, &o.ReturnNote, &returnedAt, &o.Version, &o.CreatedAt, &o.UpdatedAt)
	o.ReturnedAt = returnedAt
	o.ShippedAt = shippedAt
	// 按 id 找不到一律返回 repository.ErrNotFound（接口契约 repositories.go），
	// 此前直接外抛 pgx.ErrNoRows，导致 Service 层 notFoundErr 无法翻译、调用方
	// 区分不出"订单不存在"与"DB 故障"（会误把故障当不存在）。
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.TradeOrder{}, repository.ErrNotFound
	}
	return o, err
}

func (r *tradeOrderRepo) FindByID(ctx context.Context, id string) (domain.TradeOrder, error) {
	return scanTradeOrder(r.pool.QueryRow(ctx,
		`SELECT `+tradeOrderColumns+` FROM trade_orders WHERE id=$1`, id))
}
func (r *tradeOrderRepo) UpdateStatus(ctx context.Context, id, status string) (domain.TradeOrder, error) {
	_, err := r.pool.Exec(ctx, `UPDATE trade_orders SET status=$1,updated_at=$2,version=version+1 WHERE id=$3`, status, time.Now(), id)
	if err != nil {
		return domain.TradeOrder{}, fmt.Errorf("update trade order status: %w", err)
	}
	return r.FindByID(ctx, id)
}
func (r *tradeOrderRepo) CompareAndSetStatus(ctx context.Context, id, oldStatus, newStatus string) (bool, domain.TradeOrder, error) {
	tag, err := r.pool.Exec(ctx,
		`UPDATE trade_orders SET status=$1,updated_at=$2,version=version+1 WHERE id=$3 AND status=$4`,
		newStatus, time.Now(), id, oldStatus)
	if err != nil {
		return false, domain.TradeOrder{}, fmt.Errorf("compare-and-set trade order status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return false, domain.TradeOrder{}, nil
	}
	o, err := r.FindByID(ctx, id)
	return true, o, err
}

// Ship 卖家发货：写入快递公司/单号并把状态从 paid 迁到 shipped（条件更新，防重复发货）。
// 两条约束放 WHERE：状态必须是 paid、且尚未写过单号（shipping_tracking=''）。
func (r *tradeOrderRepo) Ship(ctx context.Context, id, company, tracking string) (domain.TradeOrder, error) {
	tag, err := r.pool.Exec(ctx,
		`UPDATE trade_orders
		    SET status='shipped', shipping_company=$2, shipping_tracking=$3, shipped_at=NOW(),
		        updated_at=NOW(), version=version+1
		  WHERE id=$1 AND status='paid' AND shipping_tracking=''`,
		id, company, tracking)
	if err != nil {
		return domain.TradeOrder{}, fmt.Errorf("ship trade order %s: %w", id, err)
	}
	if tag.RowsAffected() == 0 {
		// 不存在 / 状态不是 paid / 已发过货 —— 统一按 ErrNotFound 上报，由 Service 决定提示。
		return domain.TradeOrder{}, fmt.Errorf("trade order %s: %w", id, repository.ErrNotFound)
	}
	return r.FindByID(ctx, id)
}

// ListPendingBefore 支付超时扫描：status='pending' 且 created_at 早于 cutoff。
// 下单即把商品标 sold（支付前就占位），无人取消就会一直锁货；后台任务据此自动关单并恢复商品。
func (r *tradeOrderRepo) ListPendingBefore(ctx context.Context, cutoff time.Time, limit int) ([]domain.TradeOrder, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := r.pool.Query(ctx,
		`SELECT `+tradeOrderColumns+` FROM trade_orders WHERE status='pending' AND created_at < $1 ORDER BY created_at ASC LIMIT $2`,
		cutoff, limit)
	if err != nil {
		return nil, fmt.Errorf("list pending trade orders: %w", err)
	}
	defer rows.Close()
	out := []domain.TradeOrder{}
	for rows.Next() {
		o, scanErr := scanTradeOrder(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("scan pending trade order: %w", scanErr)
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

// ListShippedBefore 自动确认收货扫描：status='shipped' 且 updated_at 早于 cutoff。
// 买家长期不点"确认收货"时，卖家永远收不到钱——后台任务据此自动完成并放款。
func (r *tradeOrderRepo) ListShippedBefore(ctx context.Context, cutoff time.Time, limit int) ([]domain.TradeOrder, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := r.pool.Query(ctx,
		`SELECT `+tradeOrderColumns+` FROM trade_orders WHERE status='shipped' AND updated_at < $1 ORDER BY updated_at ASC LIMIT $2`,
		cutoff, limit)
	if err != nil {
		return nil, fmt.Errorf("list shipped trade orders: %w", err)
	}
	defer rows.Close()
	out := []domain.TradeOrder{}
	for rows.Next() {
		o, scanErr := scanTradeOrder(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("scan shipped trade order: %w", scanErr)
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

// HasLiveOrderForProduct 报告该商品是否还有有效订单（未取消即视为有效）。
func (r *tradeOrderRepo) HasLiveOrderForProduct(ctx context.Context, productID string) (bool, error) {
	var exists bool
	if err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM trade_orders WHERE product_id=$1 AND status <> 'cancelled')`, productID).Scan(&exists); err != nil {
		return false, fmt.Errorf("check live order for product %s: %w", productID, err)
	}
	return exists, nil
}

func (r *tradeOrderRepo) UpdateAftersale(ctx context.Context, o domain.TradeOrder, expectAftersaleStatus string) (domain.TradeOrder, error) {
	// CAS：COALESCE(aftersale_status,'') 与期望值比较，0 行即"售后状态已被并发改动"。
	// 此前是 WHERE id=$10 盲写，两个并发 approve 都能通过 Service 层的内存判断，
	// 各自调用 Transfer 把卖家余额扣两次（Transfer 本身没有幂等键）。
	tag, err := r.pool.Exec(ctx,
		`UPDATE trade_orders SET status=$1,aftersale_type=$2,aftersale_reason=$3,aftersale_desc=$4,aftersale_amount_fen=$5,aftersale_status=$6,aftersale_time=$7,aftersale_from=$8,
			return_tracking=$9,return_note=$10,returned_at=$11,updated_at=$12,version=version+1
		 WHERE id=$13 AND COALESCE(aftersale_status,'')=$14`,
		o.Status, o.AftersaleType, o.AftersaleReason, o.AftersaleDesc, o.AftersaleAmountFen, o.AftersaleStatus, o.AftersaleTime, o.AftersaleFrom,
		o.ReturnTracking, o.ReturnNote, o.ReturnedAt, time.Now(), o.ID, expectAftersaleStatus)
	if err != nil {
		return domain.TradeOrder{}, fmt.Errorf("update trade order aftersale: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.TradeOrder{}, fmt.Errorf("订单 %s 售后状态已变更，请刷新后重试", o.ID)
	}
	return r.FindByID(ctx, o.ID)
}
func (r *tradeOrderRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM trade_orders WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("delete trade order: %w", err)
	}
	return nil
}
func (r *tradeOrderRepo) ListByUser(ctx context.Context, userID string) ([]domain.TradeOrder, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+tradeOrderColumns+` FROM trade_orders WHERE buyer_id=$1 OR seller_id=$1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("list trade orders: %w", err)
	}
	defer rows.Close()
	var out []domain.TradeOrder
	for rows.Next() {
		o, err := scanTradeOrder(rows)
		if err != nil {
			return nil, fmt.Errorf("scan trade order: %w", err)
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

func (r *tradeOrderRepo) ListAll(ctx context.Context, offset, limit int) ([]domain.TradeOrder, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, "SELECT count(*) FROM trade_orders").Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count trade orders: %w", err)
	}
	rows, err := r.pool.Query(ctx, "SELECT "+tradeOrderColumns+" FROM trade_orders ORDER BY created_at DESC LIMIT $1 OFFSET $2", limit, offset)
	if err != nil {
		return nil, total, fmt.Errorf("order query failed: %w", err)
	}
	defer rows.Close()
	var out []domain.TradeOrder
	for rows.Next() {
		o, err := scanTradeOrder(rows)
		if err != nil {
			return nil, total, fmt.Errorf("scan trade order: %w", err)
		}
		out = append(out, o)
	}
	return out, total, rows.Err()
}

// ListFiltered 管理端订单列表：过滤 + COUNT + LIMIT/OFFSET 全下沉 SQL。
func (r *tradeOrderRepo) ListFiltered(ctx context.Context, f repository.TradeOrderFilter) ([]domain.TradeOrder, int, error) {
	where := "WHERE TRUE"
	args := []any{}
	if f.Status != "" {
		args = append(args, f.Status)
		where += fmt.Sprintf(" AND status=$%d", len(args))
	}
	if k := strings.TrimSpace(f.Keyword); k != "" {
		args = append(args, "%"+escapeLike(k)+"%")
		where += fmt.Sprintf(" AND id ILIKE $%d ESCAPE '\\'", len(args))
	}
	if f.StartDate != nil {
		args = append(args, *f.StartDate)
		where += fmt.Sprintf(" AND created_at >= $%d", len(args))
	}
	if f.EndDate != nil {
		args = append(args, *f.EndDate)
		where += fmt.Sprintf(" AND created_at < $%d", len(args))
	}
	var total int
	if err := r.pool.QueryRow(ctx, "SELECT count(*) FROM trade_orders "+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count trade orders: %w", err)
	}
	args = append(args, f.Limit, f.Offset)
	rows, err := r.pool.Query(ctx, "SELECT "+tradeOrderColumns+" FROM trade_orders "+where+
		fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", len(args)-1, len(args)), args...)
	if err != nil {
		return nil, total, fmt.Errorf("list trade orders filtered: %w", err)
	}
	defer rows.Close()
	var out []domain.TradeOrder
	for rows.Next() {
		o, err := scanTradeOrder(rows)
		if err != nil {
			return nil, total, fmt.Errorf("scan trade order: %w", err)
		}
		out = append(out, o)
	}
	return out, total, rows.Err()
}

// ---- Escrow ----

type escrowRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewEscrowRepository() repository.EscrowRepository { return &escrowRepo{pool: s.Pool()} }

func (r *escrowRepo) GetAccount(ctx context.Context, userID string) (domain.EscrowAccount, error) {
	var a domain.EscrowAccount
	err := r.pool.QueryRow(ctx,
		`SELECT user_id,balance_fen,frozen_fen,updated_at FROM escrow_accounts WHERE user_id=$1`, userID).
		Scan(&a.UserID, &a.BalanceFen, &a.FrozenFen, &a.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		// 账户不存在：返回零值账户（与内存实现一致）。
		return domain.EscrowAccount{UserID: userID}, nil
	}
	if err != nil {
		// C6 修复：连接/查询错误必须上抛——旧实现把一切错误当"账户不存在"，
		// 数据库故障时会在幻影零余额账户上继续操作。
		return domain.EscrowAccount{}, fmt.Errorf("get escrow account %s: %w", userID, err)
	}
	return a, nil
}

// insertEscrowTx 在同一事务中写入流水。
// channel/external_txn_id 一并落库：channel 标明资金来自哪里（内部记账 / 微信支付），
// external_txn_id 存外部支付单号，是真实资金入账的幂等键（唯一索引 idx_escrow_external）。
// isUniqueViolation 报告错误是否为唯一约束冲突（SQLSTATE 23505）。
// 放款侧就是被 idx_escrow_once_per_ref（migration 000116）挡下时命中。
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

// findEscrowTxByRef 按 (from_user, tx_type, ref_type, ref_id) 取已存在的那条流水。
// 并发放款被库级唯一索引挡下时，用它把「已放款」的真实记录还给调用方，
// 而不是回一个凭空构造的占位对象。
func (r *escrowRepo) findEscrowTxByRef(ctx context.Context, fromUser, txType, refType, refID string) (domain.EscrowTransaction, bool, error) {
	var tx domain.EscrowTransaction
	err := r.pool.QueryRow(ctx,
		`SELECT id,from_user,to_user,amount_fen,tx_type,reference_type,reference_id,status,channel,external_txn_id,created_at
		 FROM escrow_transactions
		 WHERE from_user=$1 AND tx_type=$2 AND reference_type=$3 AND reference_id=$4
		 LIMIT 1`, fromUser, txType, refType, refID).
		Scan(&tx.ID, &tx.FromUser, &tx.ToUser, &tx.AmountFen, &tx.TxType, &tx.ReferenceType, &tx.ReferenceID,
			&tx.Status, &tx.Channel, &tx.ExternalTxnID, &tx.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.EscrowTransaction{}, false, nil
		}
		return domain.EscrowTransaction{}, false, err
	}
	return tx, true, nil
}

// commitFundMove 放款/转账的收尾：写流水 + 提交。
//
// 若被 idx_escrow_once_per_ref 挡下（说明另一个并发请求已经放过款），
// 本次事务整体回滚——**付款方的冻结额/余额不会被扣减**——并按幂等成功返回。
// 这是把「同一付款方对同一业务单只放款一次」从应用层的 check-then-act
// 升级为数据库保证的关键一步：应用层的 HasReleased 查询再快也只是查询，
// 两个并发请求可以同时通过它；唯一索引不可能被同时通过。
func (r *escrowRepo) commitFundMove(ctx context.Context, btx pgx.Tx, tx domain.EscrowTransaction, op string) (domain.EscrowTransaction, error) {
	if err := insertEscrowTx(ctx, btx, tx); err != nil {
		if isUniqueViolation(err) {
			_ = btx.Rollback(ctx)
			if existing, found, ferr := r.findEscrowTxByRef(ctx, tx.FromUser, tx.TxType, tx.ReferenceType, tx.ReferenceID); ferr == nil && found {
				return existing, nil
			}
			return tx, nil
		}
		return domain.EscrowTransaction{}, fmt.Errorf("insert %s tx: %w", op, err)
	}
	if err := btx.Commit(ctx); err != nil {
		return domain.EscrowTransaction{}, fmt.Errorf("commit escrow %s: %w", op, err)
	}
	return tx, nil
}

func insertEscrowTx(ctx context.Context, q pgx.Tx, tx domain.EscrowTransaction) error {
	channel := tx.Channel
	if channel == "" {
		channel = domain.ChannelInternal
	}
	_, err := q.Exec(ctx,
		`INSERT INTO escrow_transactions (id,from_user,to_user,amount_fen,tx_type,reference_type,reference_id,status,channel,external_txn_id,created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		tx.ID, tx.FromUser, tx.ToUser, tx.AmountFen, tx.TxType, tx.ReferenceType, tx.ReferenceID, tx.Status, channel, tx.ExternalTxnID, tx.CreatedAt)
	return err
}

// 原子资金操作（C6 修复）：余额调整 + 流水写入在同一事务中，全成或全败；
// 条件 UPDATE（WHERE balance_fen>=$1）防并发丢更新。

func (r *escrowRepo) Deposit(ctx context.Context, userID string, amountFen int64, tx domain.EscrowTransaction) (domain.EscrowTransaction, error) {
	now := time.Now()
	btx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.EscrowTransaction{}, fmt.Errorf("begin escrow deposit: %w", err)
	}
	defer btx.Rollback(ctx)
	if _, err := btx.Exec(ctx,
		`INSERT INTO escrow_accounts (user_id,balance_fen,frozen_fen,updated_at) VALUES ($1,$2,0,$3)
		 ON CONFLICT (user_id) DO UPDATE SET balance_fen=escrow_accounts.balance_fen+$2, updated_at=$3`,
		userID, amountFen, now); err != nil {
		return domain.EscrowTransaction{}, fmt.Errorf("deposit %s: %w", userID, err)
	}
	if err := insertEscrowTx(ctx, btx, tx); err != nil {
		return domain.EscrowTransaction{}, fmt.Errorf("insert deposit tx: %w", err)
	}
	if err := btx.Commit(ctx); err != nil {
		return domain.EscrowTransaction{}, fmt.Errorf("commit escrow deposit: %w", err)
	}
	return tx, nil
}

func (r *escrowRepo) Freeze(ctx context.Context, userID string, amountFen int64, tx domain.EscrowTransaction) (domain.EscrowTransaction, error) {
	btx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.EscrowTransaction{}, fmt.Errorf("begin escrow freeze: %w", err)
	}
	defer btx.Rollback(ctx)
	tag, err := btx.Exec(ctx,
		`UPDATE escrow_accounts SET balance_fen=balance_fen-$1, frozen_fen=frozen_fen+$1, updated_at=$3
		 WHERE user_id=$2 AND balance_fen>=$1`,
		amountFen, userID, time.Now())
	if err != nil {
		return domain.EscrowTransaction{}, fmt.Errorf("freeze %s: %w", userID, err)
	}
	if tag.RowsAffected() == 0 {
		return domain.EscrowTransaction{}, repository.ErrInsufficientBalance
	}
	if err := insertEscrowTx(ctx, btx, tx); err != nil {
		return domain.EscrowTransaction{}, fmt.Errorf("insert freeze tx: %w", err)
	}
	if err := btx.Commit(ctx); err != nil {
		return domain.EscrowTransaction{}, fmt.Errorf("commit escrow freeze: %w", err)
	}
	return tx, nil
}

func (r *escrowRepo) Release(ctx context.Context, fromUser, toUser string, amountFen int64, tx domain.EscrowTransaction) (domain.EscrowTransaction, error) {
	now := time.Now()
	btx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.EscrowTransaction{}, fmt.Errorf("begin escrow release: %w", err)
	}
	defer btx.Rollback(ctx)
	tag, err := btx.Exec(ctx,
		`UPDATE escrow_accounts SET frozen_fen=frozen_fen-$1, updated_at=$3
		 WHERE user_id=$2 AND frozen_fen>=$1`,
		amountFen, fromUser, now)
	if err != nil {
		return domain.EscrowTransaction{}, fmt.Errorf("release from %s: %w", fromUser, err)
	}
	if tag.RowsAffected() == 0 {
		return domain.EscrowTransaction{}, repository.ErrInsufficientFrozenBalance
	}
	// 收款方可能尚无账户：upsert + 原子累加（C6 修复——旧实现两次独立
	// Upsert，付款方扣减成功而收款方失败时资金凭空消失）
	if _, err := btx.Exec(ctx,
		`INSERT INTO escrow_accounts (user_id,balance_fen,frozen_fen,updated_at) VALUES ($1,$2,0,$3)
		 ON CONFLICT (user_id) DO UPDATE SET balance_fen=escrow_accounts.balance_fen+$2, updated_at=$3`,
		toUser, amountFen, now); err != nil {
		return domain.EscrowTransaction{}, fmt.Errorf("release to %s: %w", toUser, err)
	}
	return r.commitFundMove(ctx, btx, tx, "release")
}

func (r *escrowRepo) Refund(ctx context.Context, userID string, amountFen int64, tx domain.EscrowTransaction) (domain.EscrowTransaction, error) {
	btx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.EscrowTransaction{}, fmt.Errorf("begin escrow refund: %w", err)
	}
	defer btx.Rollback(ctx)
	tag, err := btx.Exec(ctx,
		`UPDATE escrow_accounts SET frozen_fen=frozen_fen-$1, balance_fen=balance_fen+$1, updated_at=$3
		 WHERE user_id=$2 AND frozen_fen>=$1`,
		amountFen, userID, time.Now())
	if err != nil {
		return domain.EscrowTransaction{}, fmt.Errorf("refund %s: %w", userID, err)
	}
	if tag.RowsAffected() == 0 {
		return domain.EscrowTransaction{}, repository.ErrInsufficientFrozenBalance
	}
	// 走统一的收尾：被 idx_escrow_refund_once_per_order（migration 000117）挡下时
	// 整体回滚并按幂等成功返回——付款方不会被重复退回同一笔冻结款。
	return r.commitFundMove(ctx, btx, tx, "refund")
}
// Withdraw 出账：从冻结里扣掉、余额不动（钱离开平台）。
// 冻结额条件更新防并发超扣；流水与余额调整同事务，重复出账由
// idx_escrow_once_per_ref（000119 起覆盖 withdraw）挡下并整体回滚。
func (r *escrowRepo) Withdraw(ctx context.Context, userID string, amountFen int64, tx domain.EscrowTransaction) (domain.EscrowTransaction, error) {
	btx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.EscrowTransaction{}, fmt.Errorf("begin escrow withdraw: %w", err)
	}
	defer btx.Rollback(ctx)
	tag, err := btx.Exec(ctx,
		`UPDATE escrow_accounts SET frozen_fen=frozen_fen-$1, updated_at=$3 WHERE user_id=$2 AND frozen_fen>=$1`,
		amountFen, userID, time.Now())
	if err != nil {
		return domain.EscrowTransaction{}, fmt.Errorf("withdraw %s: %w", userID, err)
	}
	if tag.RowsAffected() == 0 {
		return domain.EscrowTransaction{}, repository.ErrInsufficientFrozenBalance
	}
	return r.commitFundMove(ctx, btx, tx, "withdraw")
}

// escrowTxColumns 流水查询列（各查询共用，避免漏列导致 channel 静默丢失）。
const escrowTxColumns = `id,from_user,to_user,amount_fen,tx_type,COALESCE(reference_type,''),COALESCE(reference_id,''),status,COALESCE(channel,''),COALESCE(external_txn_id,''),created_at`

// scanEscrowTx 按 escrowTxColumns 的顺序映射一行流水。
func scanEscrowTx(rows pgx.Rows) (domain.EscrowTransaction, error) {
	var tx domain.EscrowTransaction
	err := rows.Scan(&tx.ID, &tx.FromUser, &tx.ToUser, &tx.AmountFen, &tx.TxType, &tx.ReferenceType, &tx.ReferenceID, &tx.Status, &tx.Channel, &tx.ExternalTxnID, &tx.CreatedAt)
	return tx, err
}

func (r *escrowRepo) ListTransactions(ctx context.Context, userID string) ([]domain.EscrowTransaction, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+escrowTxColumns+` FROM escrow_transactions WHERE from_user=$1 OR to_user=$1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("list escrow transactions: %w", err)
	}
	defer rows.Close()
	var out []domain.EscrowTransaction
	for rows.Next() {
		tx, err := scanEscrowTx(rows)
		if err != nil {
			return nil, fmt.Errorf("scan escrow transaction: %w", err)
		}
		out = append(out, tx)
	}
	return out, rows.Err()
}

// FindByExternalTxn 按 (渠道, 外部支付单号) 精确查流水——真实支付回调入账幂等用。
// 只认已完成（status=completed）的流水：未完成的入账不算"钱已到账"。
func (r *escrowRepo) FindByExternalTxn(ctx context.Context, channel, externalTxnID string) (domain.EscrowTransaction, bool, error) {
	if externalTxnID == "" {
		return domain.EscrowTransaction{}, false, nil
	}
	rows, err := r.pool.Query(ctx,
		`SELECT `+escrowTxColumns+` FROM escrow_transactions
		 WHERE channel=$1 AND external_txn_id=$2 AND status='completed' ORDER BY created_at LIMIT 1`, channel, externalTxnID)
	if err != nil {
		return domain.EscrowTransaction{}, false, fmt.Errorf("find escrow tx by external id: %w", err)
	}
	defer rows.Close()
	if !rows.Next() {
		return domain.EscrowTransaction{}, false, rows.Err()
	}
	tx, err := scanEscrowTx(rows)
	if err != nil {
		return domain.EscrowTransaction{}, false, fmt.Errorf("scan escrow tx by external id: %w", err)
	}
	return tx, true, rows.Err()
}

// ListByChannel 按渠道 + 时间区间列流水（对账明细，created_at 升序）。
// channel 为空串表示不限渠道；from/to 为零值时该端不限。
func (r *escrowRepo) ListByChannel(ctx context.Context, channel string, from, to time.Time, limit int) ([]domain.EscrowTransaction, error) {
	if limit <= 0 || limit > 5000 {
		limit = 500
	}
	rows, err := r.pool.Query(ctx,
		`SELECT `+escrowTxColumns+` FROM escrow_transactions
		 WHERE ($1 = '' OR channel = $1)
		   AND ($2::timestamptz IS NULL OR created_at >= $2)
		   AND ($3::timestamptz IS NULL OR created_at < $3)
		 ORDER BY created_at LIMIT $4`, channel, nullTime(from), nullTime(to), limit)
	if err != nil {
		return nil, fmt.Errorf("list escrow tx by channel: %w", err)
	}
	defer rows.Close()
	var out []domain.EscrowTransaction
	for rows.Next() {
		tx, err := scanEscrowTx(rows)
		if err != nil {
			return nil, fmt.Errorf("scan escrow tx by channel: %w", err)
		}
		out = append(out, tx)
	}
	return out, rows.Err()
}

// nullTime 把零值时间转成 SQL NULL（"该端不限"），非零值原样传给 pgx。
func nullTime(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

// hasEscrowTx 报告 (userID, refType, refID, txType) 是否已有完成流水。
// HasReleased / HasFrozen / HasRefunded 共用同一查询，只差 tx_type。
func (r *escrowRepo) hasEscrowTx(ctx context.Context, userID, refType, refID, txType string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM escrow_transactions
			WHERE from_user=$1 AND reference_type=$2 AND reference_id=$3
			  AND tx_type=$4 AND status='completed'
		)`, userID, refType, refID, txType).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check escrow %s: %w", txType, err)
	}
	return exists, nil
}

// HasFrozen 查 userID 对 (refType, refID) 是否冻结过资金（取消订单判断"是否有钱可退"）。
func (r *escrowRepo) HasFrozen(ctx context.Context, userID, refType, refID string) (bool, error) {
	return r.hasEscrowTx(ctx, userID, refType, refID, "freeze")
}

// HasRefunded 查 userID 对 (refType, refID) 是否已退款（取消/售后退款幂等）。
func (r *escrowRepo) HasRefunded(ctx context.Context, userID, refType, refID string) (bool, error) {
	return r.hasEscrowTx(ctx, userID, refType, refID, "refund")
}

// HasTransferred 查 fromUser 对 (refType, refID) 是否已有完成的 transfer 流水（售后退款幂等）。
func (r *escrowRepo) HasTransferred(ctx context.Context, fromUser, refType, refID string) (bool, error) {
	return r.hasEscrowTx(ctx, fromUser, refType, refID, "transfer")
}

// Transfer 双方余额内转账（不涉及冻结）：from 余额扣减、to 余额增加，同一事务写流水。
// 用于"货款已放给卖家后通过的售后退款"——钱已不在冻结里，只能从卖家余额扣回买家。
func (r *escrowRepo) Transfer(ctx context.Context, fromUser, toUser string, amountFen int64, tx domain.EscrowTransaction) (domain.EscrowTransaction, error) {
	if fromUser == toUser {
		return domain.EscrowTransaction{}, fmt.Errorf("transfer to self")
	}
	now := time.Now()
	btx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.EscrowTransaction{}, fmt.Errorf("begin escrow transfer: %w", err)
	}
	defer btx.Rollback(ctx)
	tag, err := btx.Exec(ctx,
		`UPDATE escrow_accounts SET balance_fen=balance_fen-$1, updated_at=$3
		 WHERE user_id=$2 AND balance_fen>=$1`,
		amountFen, fromUser, now)
	if err != nil {
		return domain.EscrowTransaction{}, fmt.Errorf("transfer from %s: %w", fromUser, err)
	}
	if tag.RowsAffected() == 0 {
		return domain.EscrowTransaction{}, repository.ErrInsufficientBalance
	}
	if _, err := btx.Exec(ctx,
		`INSERT INTO escrow_accounts (user_id,balance_fen,frozen_fen,updated_at) VALUES ($1,$2,0,$3)
		 ON CONFLICT (user_id) DO UPDATE SET balance_fen=escrow_accounts.balance_fen+$2, updated_at=$3`,
		toUser, amountFen, now); err != nil {
		return domain.EscrowTransaction{}, fmt.Errorf("transfer to %s: %w", toUser, err)
	}
	return r.commitFundMove(ctx, btx, tx, "transfer")
}

// HasReleased 查 fromUser 对 (refType, refID) 是否已有完成的 release 流水。
func (r *escrowRepo) HasReleased(ctx context.Context, fromUser, refType, refID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM escrow_transactions
			WHERE from_user=$1 AND reference_type=$2 AND reference_id=$3
			  AND tx_type='release' AND status='completed'
		)`, fromUser, refType, refID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check escrow release: %w", err)
	}
	return exists, nil
}

// ListOrphanFreezes 查"冻结但业务记录不存在"的孤儿冻结流水。
// refType 为业务引用类型（如 training_course）；业务表名与引用字段由 refType 映射。
func (r *escrowRepo) ListOrphanFreezes(ctx context.Context, refType string, olderThan time.Time, limit int) ([]domain.EscrowTransaction, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	// refType → 业务表 / 属主列 / 关联业务 ID 列（新增引用类型需同步扩展）
	table, userCol, keyCol := "", "", ""
	switch refType {
	case "training_course":
		table, userCol, keyCol = "training_enrollments", "user_id", "course_id"
	case "trade_order":
		// 商城订单：买家即冻结方；订单行存在即"业务已落库"，不存在才算孤儿冻结
		table, userCol, keyCol = "trade_orders", "buyer_id", "id"
	default:
		return nil, nil // 未知引用类型无孤儿判定规则，跳过
	}
	rows, err := r.pool.Query(ctx,
		`SELECT `+escrowTxColumns+`
		 FROM escrow_transactions t
		 WHERE t.tx_type='freeze' AND t.reference_type=$1 AND t.status='completed' AND t.created_at < $2
		   AND NOT EXISTS (
		     SELECT 1 FROM `+table+` e WHERE e.`+userCol+` = t.from_user AND e.`+keyCol+` = t.reference_id
		   )
		 ORDER BY t.created_at LIMIT $3`, refType, olderThan, limit)
	if err != nil {
		return nil, fmt.Errorf("list orphan freezes: %w", err)
	}
	defer rows.Close()
	var out []domain.EscrowTransaction
	for rows.Next() {
		tx, err := scanEscrowTx(rows)
		if err != nil {
			return nil, fmt.Errorf("scan orphan freeze: %w", err)
		}
		out = append(out, tx)
	}
	return out, rows.Err()
}
