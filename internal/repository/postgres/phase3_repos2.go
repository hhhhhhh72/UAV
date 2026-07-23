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

// ---- Inspection ----

type inspectRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewInspectionRepository() repository.InspectionRepository { return &inspectRepo{pool: s.Pool()} }

func (r *inspectRepo) Create(i domain.AnnualInspection) (domain.AnnualInspection, error) {
	i.Version = 1; i.CreatedAt = time.Now(); i.UpdatedAt = i.CreatedAt
	_, err := r.pool.Exec(context.Background(),
		`INSERT INTO annual_inspections (id,user_id,drone_model,drone_sn,inspect_date,expire_date,result,report_url,status,version,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		i.ID, i.UserID, i.DroneModel, i.DroneSN, i.InspectDate, i.ExpireDate, i.Result, i.ReportURL, i.Status, i.Version, i.CreatedAt, i.UpdatedAt)
	return i, err
}
func (r *inspectRepo) ListByUser(userID string) ([]domain.AnnualInspection, error) {
	rows, err := r.pool.Query(context.Background(),
		`SELECT id,user_id,drone_model,drone_sn,inspect_date,expire_date,result,report_url,status,version,created_at,updated_at FROM annual_inspections WHERE user_id=$1 ORDER BY created_at DESC`, userID)
	if err != nil { return nil, fmt.Errorf("list inspections: %w", err) }
	defer rows.Close()
	var out []domain.AnnualInspection
	for rows.Next() {
		var i domain.AnnualInspection
		rows.Scan(&i.ID, &i.UserID, &i.DroneModel, &i.DroneSN, &i.InspectDate, &i.ExpireDate, &i.Result, &i.ReportURL, &i.Status, &i.Version, &i.CreatedAt, &i.UpdatedAt)
		out = append(out, i)
	}
	return out, rows.Err()
}
func (r *inspectRepo) ListAll() ([]domain.AnnualInspection, error) {
	rows, err := r.pool.Query(context.Background(),
		`SELECT id,user_id,drone_model,drone_sn,inspect_date,expire_date,result,report_url,status,version,created_at,updated_at FROM annual_inspections ORDER BY created_at DESC`)
	if err != nil { return nil, fmt.Errorf("list all inspections: %w", err) }
	defer rows.Close()
	var out []domain.AnnualInspection
	for rows.Next() {
		var i domain.AnnualInspection
		rows.Scan(&i.ID, &i.UserID, &i.DroneModel, &i.DroneSN, &i.InspectDate, &i.ExpireDate, &i.Result, &i.ReportURL, &i.Status, &i.Version, &i.CreatedAt, &i.UpdatedAt)
		out = append(out, i)
	}
	return out, rows.Err()
}

// ---- Loan ----

type loanRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewLoanRepository() repository.LoanRepository { return &loanRepo{pool: s.Pool()} }

func (r *loanRepo) Create(l domain.LoanApplication) (domain.LoanApplication, error) {
	l.Version = 1; l.CreatedAt = time.Now(); l.UpdatedAt = l.CreatedAt
	_, err := r.pool.Exec(context.Background(),
		`INSERT INTO loan_applications (id,user_id,amount_fen,term_months,purpose,status,approved_fen,monthly_pay_fen,version,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		l.ID, l.UserID, l.AmountFen, l.TermMonths, l.Purpose, l.Status, l.ApprovedFen, l.MonthlyPayFen, l.Version, l.CreatedAt, l.UpdatedAt)
	return l, err
}
func (r *loanRepo) ListByUser(userID string) ([]domain.LoanApplication, error) {
	rows, err := r.pool.Query(context.Background(),
		`SELECT id,user_id,amount_fen,term_months,purpose,status,approved_fen,monthly_pay_fen,version,created_at,updated_at FROM loan_applications WHERE user_id=$1 ORDER BY created_at DESC`, userID)
	if err != nil { return nil, fmt.Errorf("list loans: %w", err) }
	defer rows.Close()
	var out []domain.LoanApplication
	for rows.Next() {
		var l domain.LoanApplication
		rows.Scan(&l.ID, &l.UserID, &l.AmountFen, &l.TermMonths, &l.Purpose, &l.Status, &l.ApprovedFen, &l.MonthlyPayFen, &l.Version, &l.CreatedAt, &l.UpdatedAt)
		out = append(out, l)
	}
	return out, rows.Err()
}

// ---- Message ----

type msgRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewMessageRepository() repository.MessageRepository { return &msgRepo{pool: s.Pool()} }

func (r *msgRepo) Create(m domain.Message) (domain.Message, error) {
	m.CreatedAt = time.Now()
	_, err := r.pool.Exec(context.Background(),
		`INSERT INTO messages (id,sender_id,receiver_id,title,content,resource_type,resource_id,is_read,created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		m.ID, m.SenderID, m.ReceiverID, m.Title, m.Content, m.ResourceType, m.ResourceID, m.IsRead, m.CreatedAt)
	return m, err
}
func (r *msgRepo) ListByUser(userID string, unreadOnly bool) ([]domain.Message, error) {
	q := `SELECT id,sender_id,receiver_id,title,content,resource_type,resource_id,is_read,created_at FROM messages WHERE receiver_id=$1`
	if unreadOnly { q += ` AND is_read=false` }
	q += ` ORDER BY created_at DESC`
	rows, err := r.pool.Query(context.Background(), q, userID)
	if err != nil { return nil, fmt.Errorf("list messages: %w", err) }
	defer rows.Close()
	var out []domain.Message
	for rows.Next() {
		var m domain.Message
		rows.Scan(&m.ID, &m.SenderID, &m.ReceiverID, &m.Title, &m.Content, &m.ResourceType, &m.ResourceID, &m.IsRead, &m.CreatedAt)
		out = append(out, m)
	}
	return out, rows.Err()
}
func (r *msgRepo) MarkRead(id string) (domain.Message, error) {
	_, err := r.pool.Exec(context.Background(), `UPDATE messages SET is_read=true WHERE id=$1`, id)
	if err != nil { return domain.Message{}, fmt.Errorf("mark message read: %w", err) }
	var m domain.Message
	err = r.pool.QueryRow(context.Background(),
		`SELECT id,sender_id,receiver_id,title,content,resource_type,resource_id,is_read,created_at FROM messages WHERE id=$1`, id).
		Scan(&m.ID, &m.SenderID, &m.ReceiverID, &m.Title, &m.Content, &m.ResourceType, &m.ResourceID, &m.IsRead, &m.CreatedAt)
	return m, err
}
func (r *msgRepo) UnreadCount(userID string) (int, error) {
	var n int
	err := r.pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM messages WHERE receiver_id=$1 AND is_read=false`, userID).Scan(&n)
	return n, err
}

// ---- Article ----

type articleRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewArticleRepository() repository.ArticleRepository { return &articleRepo{pool: s.Pool()} }

func (r *articleRepo) Create(a domain.Article) (domain.Article, error) {
	a.Version = 1; a.CreatedAt = time.Now(); a.UpdatedAt = a.CreatedAt
	_, err := r.pool.Exec(context.Background(),
		`INSERT INTO articles (id,title,content,summary,category,source,author,is_pinned,status,version,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		a.ID, a.Title, a.Content, a.Summary, a.Category, a.Source, a.Author, a.IsPinned, a.Status, a.Version, a.CreatedAt, a.UpdatedAt)
	return a, err
}
func (r *articleRepo) FindByID(id string) (domain.Article, error) {
	var a domain.Article
	err := r.pool.QueryRow(context.Background(),
		`SELECT id,title,content,summary,category,source,author,is_pinned,status,version,created_at,updated_at FROM articles WHERE id=$1`, id).
		Scan(&a.ID, &a.Title, &a.Content, &a.Summary, &a.Category, &a.Source, &a.Author, &a.IsPinned, &a.Status, &a.Version, &a.CreatedAt, &a.UpdatedAt)
	return a, err
}
func (r *articleRepo) Update(a domain.Article) (domain.Article, error) {
	a.UpdatedAt = time.Now(); a.Version++
	_, err := r.pool.Exec(context.Background(),
		`UPDATE articles SET title=$1,content=$2,summary=$3,category=$4,source=$5,author=$6,is_pinned=$7,status=$8,version=$9,updated_at=$10 WHERE id=$11`,
		a.Title, a.Content, a.Summary, a.Category, a.Source, a.Author, a.IsPinned, a.Status, a.Version, a.UpdatedAt, a.ID)
	return a, err
}
func (r *articleRepo) ListByCategory(category string, offset, limit int) ([]domain.Article, int, error) {
	where := ""; args := []any{}
	if category != "" { where = `WHERE category=$1`; args = append(args, category) }
	var total int
	if err := r.pool.QueryRow(context.Background(), `SELECT COUNT(*) FROM articles `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	q := `SELECT id,title,content,summary,category,source,author,is_pinned,status,version,created_at,updated_at FROM articles ` + where + ` ORDER BY created_at DESC LIMIT $` + fmt.Sprintf("%d", len(args)+1) + ` OFFSET $` + fmt.Sprintf("%d", len(args)+2)
	allArgs := append(args, limit, offset)
	rows, err := r.pool.Query(context.Background(), q, allArgs...)
	if err != nil { return nil, 0, fmt.Errorf("list articles: %w", err) }
	defer rows.Close()
	var out []domain.Article
	for rows.Next() {
		var a domain.Article
		rows.Scan(&a.ID, &a.Title, &a.Content, &a.Summary, &a.Category, &a.Source, &a.Author, &a.IsPinned, &a.Status, &a.Version, &a.CreatedAt, &a.UpdatedAt)
		out = append(out, a)
	}
	return out, total, rows.Err()
}

// ---- Review ----

type reviewRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewReviewRepository() repository.ReviewRepository { return &reviewRepo{pool: s.Pool()} }

func (r *reviewRepo) Create(rv domain.Review) (domain.Review, error) {
	rv.CreatedAt = time.Now()
	_, err := r.pool.Exec(context.Background(),
		`INSERT INTO reviews (id,reviewer_id,target_type,target_id,rating,content,status,created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		rv.ID, rv.ReviewerID, rv.TargetType, rv.TargetID, rv.Rating, rv.Content, rv.Status, rv.CreatedAt)
	return rv, err
}
func (r *reviewRepo) ListByTarget(targetType, targetID string) ([]domain.Review, error) {
	rows, err := r.pool.Query(context.Background(),
		`SELECT id,reviewer_id,target_type,target_id,rating,content,status,created_at FROM reviews WHERE target_type=$1 AND target_id=$2 AND status='approved' ORDER BY created_at DESC`, targetType, targetID)
	if err != nil { return nil, fmt.Errorf("list reviews: %w", err) }
	defer rows.Close()
	var out []domain.Review
	for rows.Next() {
		var rv domain.Review
		rows.Scan(&rv.ID, &rv.ReviewerID, &rv.TargetType, &rv.TargetID, &rv.Rating, &rv.Content, &rv.Status, &rv.CreatedAt)
		out = append(out, rv)
	}
	return out, rows.Err()
}
func (r *reviewRepo) ListAll(status string, offset, limit int) ([]domain.Review, int, error) {
	where := ""; args := []any{}
	if status != "" { where = `WHERE status=$1`; args = append(args, status) }
	var total int
	if err := r.pool.QueryRow(context.Background(), `SELECT COUNT(*) FROM reviews `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	q := `SELECT id,reviewer_id,target_type,target_id,rating,content,status,created_at FROM reviews ` + where + ` ORDER BY created_at DESC LIMIT $` + fmt.Sprintf("%d", len(args)+1) + ` OFFSET $` + fmt.Sprintf("%d", len(args)+2)
	allArgs := append(args, limit, offset)
	rows, err := r.pool.Query(context.Background(), q, allArgs...)
	if err != nil { return nil, 0, fmt.Errorf("list all reviews: %w", err) }
	defer rows.Close()
	var out []domain.Review
	for rows.Next() {
		var rv domain.Review
		rows.Scan(&rv.ID, &rv.ReviewerID, &rv.TargetType, &rv.TargetID, &rv.Rating, &rv.Content, &rv.Status, &rv.CreatedAt)
		out = append(out, rv)
	}
	return out, total, rows.Err()
}
func (r *reviewRepo) UpdateStatus(id, status string) (domain.Review, error) {
	_, err := r.pool.Exec(context.Background(), `UPDATE reviews SET status=$1 WHERE id=$2`, status, id)
	if err != nil { return domain.Review{}, fmt.Errorf("update review status: %w", err) }
	var rv domain.Review
	err = r.pool.QueryRow(context.Background(),
		`SELECT id,reviewer_id,target_type,target_id,rating,content,status,created_at FROM reviews WHERE id=$1`, id).
		Scan(&rv.ID, &rv.ReviewerID, &rv.TargetType, &rv.TargetID, &rv.Rating, &rv.Content, &rv.Status, &rv.CreatedAt)
	return rv, err
}
func (r *reviewRepo) Delete(id string) error {
	_, err := r.pool.Exec(context.Background(), `DELETE FROM reviews WHERE id=$1`, id)
	return err
}

// ---- Venue ----

type venueRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewVenueRepository() repository.VenueRepository { return &venueRepo{pool: s.Pool()} }

func (r *venueRepo) Create(v domain.Venue) (domain.Venue, error) {
	v.CreatedAt = time.Now()
	_, err := r.pool.Exec(context.Background(),
		`INSERT INTO venues (id,owner_id,name,venue_type,location,price_fen,status,created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		v.ID, v.OwnerID, v.Name, v.VenueType, v.Location, v.PriceFen, v.Status, v.CreatedAt)
	return v, err
}
func (r *venueRepo) List(venueType string) ([]domain.Venue, error) {
	var rows pgx.Rows; var err error
	if venueType == "" {
		rows, err = r.pool.Query(context.Background(), `SELECT id,owner_id,name,venue_type,location,price_fen,status,created_at FROM venues ORDER BY created_at DESC`)
	} else {
		rows, err = r.pool.Query(context.Background(), `SELECT id,owner_id,name,venue_type,location,price_fen,status,created_at FROM venues WHERE venue_type=$1 ORDER BY created_at DESC`, venueType)
	}
	if err != nil { return nil, fmt.Errorf("list venues: %w", err) }
	defer rows.Close()
	var out []domain.Venue
	for rows.Next() {
		var v domain.Venue
		rows.Scan(&v.ID, &v.OwnerID, &v.Name, &v.VenueType, &v.Location, &v.PriceFen, &v.Status, &v.CreatedAt)
		out = append(out, v)
	}
	return out, rows.Err()
}
func (r *venueRepo) FindByID(id string) (domain.Venue, error) {
	var v domain.Venue
	err := r.pool.QueryRow(context.Background(),
		`SELECT id,owner_id,name,venue_type,location,price_fen,status,created_at FROM venues WHERE id=$1`, id).
		Scan(&v.ID, &v.OwnerID, &v.Name, &v.VenueType, &v.Location, &v.PriceFen, &v.Status, &v.CreatedAt)
	return v, err
}
func (r *venueRepo) CreateBooking(b domain.VenueBooking) (domain.VenueBooking, error) {
	b.CreatedAt = time.Now()
	_, err := r.pool.Exec(context.Background(),
		`INSERT INTO venue_bookings (id,venue_id,user_id,start_time,end_time,status,created_at) VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		b.ID, b.VenueID, b.UserID, b.StartTime, b.EndTime, b.Status, b.CreatedAt)
	return b, err
}
func (r *venueRepo) ListBookings(venueID string) ([]domain.VenueBooking, error) {
	rows, err := r.pool.Query(context.Background(),
		`SELECT id,venue_id,user_id,start_time,end_time,status,created_at FROM venue_bookings WHERE venue_id=$1 ORDER BY created_at DESC`, venueID)
	if err != nil { return nil, fmt.Errorf("list bookings: %w", err) }
	defer rows.Close()
	var out []domain.VenueBooking
	for rows.Next() {
		var b domain.VenueBooking
		rows.Scan(&b.ID, &b.VenueID, &b.UserID, &b.StartTime, &b.EndTime, &b.Status, &b.CreatedAt)
		out = append(out, b)
	}
	return out, rows.Err()
}

// ---- Enrollment ----

type enrollRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewEnrollmentRepository() repository.EnrollmentRepository { return &enrollRepo{pool: s.Pool()} }

func (r *enrollRepo) Create(e domain.Enrollment) (domain.Enrollment, error) {
	e.CreatedAt = time.Now()
	_, err := r.pool.Exec(context.Background(),
		`INSERT INTO training_enrollments (id,course_id,user_id,status,created_at) VALUES ($1,$2,$3,$4,$5)`,
		e.ID, e.CourseID, e.UserID, e.Status, e.CreatedAt)
	return e, err
}
func (r *enrollRepo) ListByCourse(courseID string) ([]domain.Enrollment, error) {
	rows, err := r.pool.Query(context.Background(),
		`SELECT id,course_id,user_id,status,created_at FROM training_enrollments WHERE course_id=$1 ORDER BY created_at DESC`, courseID)
	if err != nil { return nil, fmt.Errorf("list enrollments: %w", err) }
	defer rows.Close()
	var out []domain.Enrollment
	for rows.Next() {
		var e domain.Enrollment
		rows.Scan(&e.ID, &e.CourseID, &e.UserID, &e.Status, &e.CreatedAt)
		out = append(out, e)
	}
	return out, rows.Err()
}
func (r *enrollRepo) FindByUserAndCourse(userID, courseID string) (domain.Enrollment, bool, error) {
	var e domain.Enrollment
	err := r.pool.QueryRow(context.Background(),
		`SELECT id,course_id,user_id,status,created_at FROM training_enrollments WHERE user_id=$1 AND course_id=$2`, userID, courseID).
		Scan(&e.ID, &e.CourseID, &e.UserID, &e.Status, &e.CreatedAt)
	if err != nil { return domain.Enrollment{}, false, nil }
	return e, true, nil
}

// ---- TradeOrder ----

type tradeOrderRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewTradeOrderRepository() repository.TradeOrderRepository { return &tradeOrderRepo{pool: s.Pool()} }

func (r *tradeOrderRepo) Create(o domain.TradeOrder) (domain.TradeOrder, error) {
	o.Version = 1; o.CreatedAt = time.Now(); o.UpdatedAt = o.CreatedAt
	_, err := r.pool.Exec(context.Background(),
		`INSERT INTO trade_orders (id,product_id,buyer_id,seller_id,amount_fen,status,version,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		o.ID, o.ProductID, o.BuyerID, o.SellerID, o.AmountFen, o.Status, o.Version, o.CreatedAt, o.UpdatedAt)
	return o, err
}
func (r *tradeOrderRepo) FindByID(id string) (domain.TradeOrder, error) {
	var o domain.TradeOrder
	err := r.pool.QueryRow(context.Background(),
		`SELECT id,product_id,buyer_id,seller_id,amount_fen,status,version,created_at,updated_at FROM trade_orders WHERE id=$1`, id).
		Scan(&o.ID, &o.ProductID, &o.BuyerID, &o.SellerID, &o.AmountFen, &o.Status, &o.Version, &o.CreatedAt, &o.UpdatedAt)
	return o, err
}
func (r *tradeOrderRepo) UpdateStatus(id, status string) (domain.TradeOrder, error) {
	_, err := r.pool.Exec(context.Background(), `UPDATE trade_orders SET status=$1,updated_at=$2,version=version+1 WHERE id=$3`, status, time.Now(), id)
	if err != nil { return domain.TradeOrder{}, fmt.Errorf("update trade order status: %w", err) }
	return r.FindByID(id)
}
func (r *tradeOrderRepo) ListByUser(userID string) ([]domain.TradeOrder, error) {
	rows, err := r.pool.Query(context.Background(),
		`SELECT id,product_id,buyer_id,seller_id,amount_fen,status,version,created_at,updated_at FROM trade_orders WHERE buyer_id=$1 OR seller_id=$1 ORDER BY created_at DESC`, userID)
	if err != nil { return nil, fmt.Errorf("list trade orders: %w", err) }
	defer rows.Close()
	var out []domain.TradeOrder
	for rows.Next() {
		var o domain.TradeOrder
		rows.Scan(&o.ID, &o.ProductID, &o.BuyerID, &o.SellerID, &o.AmountFen, &o.Status, &o.Version, &o.CreatedAt, &o.UpdatedAt)
		out = append(out, o)
	}
	return out, rows.Err()
}

// ---- Escrow ----

type escrowRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewEscrowRepository() repository.EscrowRepository { return &escrowRepo{pool: s.Pool()} }

func (r *escrowRepo) GetAccount(userID string) (domain.EscrowAccount, error) {
	var a domain.EscrowAccount
	err := r.pool.QueryRow(context.Background(),
		`SELECT user_id,balance_fen,frozen_fen,updated_at FROM escrow_accounts WHERE user_id=$1`, userID).
		Scan(&a.UserID, &a.BalanceFen, &a.FrozenFen, &a.UpdatedAt)
	if err != nil {
		// Return a zero account if not found.
		return domain.EscrowAccount{UserID: userID}, nil
	}
	return a, nil
}
func (r *escrowRepo) UpsertAccount(a domain.EscrowAccount) error {
	a.UpdatedAt = time.Now()
	_, err := r.pool.Exec(context.Background(),
		`INSERT INTO escrow_accounts (user_id,balance_fen,frozen_fen,updated_at) VALUES ($1,$2,$3,$4)
		 ON CONFLICT (user_id) DO UPDATE SET balance_fen=$2,frozen_fen=$3,updated_at=$4`,
		a.UserID, a.BalanceFen, a.FrozenFen, a.UpdatedAt)
	return err
}
func (r *escrowRepo) CreateTransaction(tx domain.EscrowTransaction) (domain.EscrowTransaction, error) {
	tx.CreatedAt = time.Now()
	_, err := r.pool.Exec(context.Background(),
		`INSERT INTO escrow_transactions (id,from_user,to_user,amount_fen,tx_type,reference_type,reference_id,status,created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		tx.ID, tx.FromUser, tx.ToUser, tx.AmountFen, tx.TxType, tx.ReferenceType, tx.ReferenceID, tx.Status, tx.CreatedAt)
	return tx, err
}
func (r *escrowRepo) ListTransactions(userID string) ([]domain.EscrowTransaction, error) {
	rows, err := r.pool.Query(context.Background(),
		`SELECT id,from_user,to_user,amount_fen,tx_type,reference_type,reference_id,status,created_at FROM escrow_transactions WHERE from_user=$1 OR to_user=$1 ORDER BY created_at DESC`, userID)
	if err != nil { return nil, fmt.Errorf("list escrow transactions: %w", err) }
	defer rows.Close()
	var out []domain.EscrowTransaction
	for rows.Next() {
		var tx domain.EscrowTransaction
		rows.Scan(&tx.ID, &tx.FromUser, &tx.ToUser, &tx.AmountFen, &tx.TxType, &tx.ReferenceType, &tx.ReferenceID, &tx.Status, &tx.CreatedAt)
		out = append(out, tx)
	}
	return out, rows.Err()
}
