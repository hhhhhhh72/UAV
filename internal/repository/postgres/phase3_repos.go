package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"drone-platform/internal/crypto"
	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)

// ---- Certificate ----

type certRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewCertificateRepository() repository.CertificateRepository {
	return &certRepo{pool: s.Pool()}
}

// certTimeOrNull 把"没有有效期"（Go 零值时间）写成 SQL NULL。
//
// 直接写零值时间会落成 0001-01-01 08:05:43+08:05（+08:05:43 是当年的 LMT 偏移），
// 它**不是 NULL**，所以读侧的 COALESCE 兜不住；前端拿到这个假日期就会判"已过期"
// 并在卡片上显示"至 0001-01-01"。历史遗留的 1970/0001 哨兵值一并按"无有效期"处理。
func certTimeOrNull(t time.Time) *time.Time {
	if t.IsZero() || t.Before(certTimeFloor) {
		return nil
	}
	return &t
}

// certTimeFromNull 读侧映射：NULL 与历史哨兵值一律还原成 Go 零值时间，
// 语义 = "长期有效"（与 service 侧 validExpireDate 的下界判断同口径）。
func certTimeFromNull(t *time.Time) time.Time {
	if t == nil || t.Before(certTimeFloor) {
		return time.Time{}
	}
	return *t
}

// certTimeFloor 有效期下限：早于 2000-01-01 的值一律视为"没有填"。
var certTimeFloor = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

func (r *certRepo) Create(ctx context.Context, c domain.Certificate) (domain.Certificate, error) {
	c.Version = 1
	c.CreatedAt = time.Now()
	c.UpdatedAt = c.CreatedAt
	_, err := r.pool.Exec(ctx,
		`INSERT INTO certificates (id,user_id,cert_type,cert_number,level,issue_date,expire_date,issuer_org,image_url,status,version,created_at,updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		c.ID, c.UserID, string(c.CertType), c.CertNumber, c.Level, certTimeOrNull(c.IssueDate), certTimeOrNull(c.ExpireDate), c.IssuerOrg, c.ImageURL, c.Status, c.Version, c.CreatedAt, c.UpdatedAt)
	if err != nil {
		// 唯一索引 certificates_cert_number_unique 兜底：撞号映射为哨兵错误。
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.Certificate{}, repository.ErrCertNumberTaken
		}
		return domain.Certificate{}, fmt.Errorf("create certificate: %w", err)
	}
	return c, nil
}
func (r *certRepo) FindByID(ctx context.Context, id string) (domain.Certificate, error) {
	var c domain.Certificate
	var ct string
	var issueAt, expireAt *time.Time
	err := r.pool.QueryRow(ctx,
		`SELECT id,user_id,cert_type,COALESCE(cert_number,''),COALESCE(level,''),issue_date,expire_date,COALESCE(issuer_org,''),COALESCE(image_url,''),status,version,created_at,updated_at FROM certificates WHERE id=$1`, id).
		Scan(&c.ID, &c.UserID, &ct, &c.CertNumber, &c.Level, &issueAt, &expireAt, &c.IssuerOrg, &c.ImageURL, &c.Status, &c.Version, &c.CreatedAt, &c.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Certificate{}, fmt.Errorf("certificate %s: %w", id, repository.ErrNotFound)
	}
	c.CertType = domain.CertType(ct)
	c.IssueDate, c.ExpireDate = certTimeFromNull(issueAt), certTimeFromNull(expireAt)
	return c, err
}
func (r *certRepo) FindByNumber(ctx context.Context, certNumber string) (domain.Certificate, error) {
	var c domain.Certificate
	var ct string
	var issueAt, expireAt *time.Time
	err := r.pool.QueryRow(ctx,
		`SELECT id,user_id,cert_type,COALESCE(cert_number,''),COALESCE(level,''),issue_date,expire_date,COALESCE(issuer_org,''),COALESCE(image_url,''),status,version,created_at,updated_at FROM certificates WHERE cert_number=$1 LIMIT 1`, certNumber).
		Scan(&c.ID, &c.UserID, &ct, &c.CertNumber, &c.Level, &issueAt, &expireAt, &c.IssuerOrg, &c.ImageURL, &c.Status, &c.Version, &c.CreatedAt, &c.UpdatedAt)
	c.CertType = domain.CertType(ct)
	c.IssueDate, c.ExpireDate = certTimeFromNull(issueAt), certTimeFromNull(expireAt)
	return c, err
}
func (r *certRepo) ListByUser(ctx context.Context, userID string) ([]domain.Certificate, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id,user_id,cert_type,COALESCE(cert_number,''),COALESCE(level,''),issue_date,expire_date,COALESCE(issuer_org,''),COALESCE(image_url,''),status,version,created_at,updated_at FROM certificates WHERE user_id=$1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("list certificates: %w", err)
	}
	defer rows.Close()
	var out []domain.Certificate
	for rows.Next() {
		var c domain.Certificate
		var ct string
		var issueAt, expireAt *time.Time
		if err := rows.Scan(&c.ID, &c.UserID, &ct, &c.CertNumber, &c.Level, &issueAt, &expireAt, &c.IssuerOrg, &c.ImageURL, &c.Status, &c.Version, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan certificate: %w", err)
		}
		c.CertType = domain.CertType(ct)
		c.IssueDate, c.ExpireDate = certTimeFromNull(issueAt), certTimeFromNull(expireAt)
		out = append(out, c)
	}
	return out, rows.Err()
}
func (r *certRepo) UpdateStatus(ctx context.Context, id, status string) (domain.Certificate, error) {
	_, err := r.pool.Exec(ctx, `UPDATE certificates SET status=$1,updated_at=$2 WHERE id=$3`, status, time.Now(), id)
	if err != nil {
		return domain.Certificate{}, fmt.Errorf("update certificate status: %w", err)
	}
	return r.FindByID(ctx, id)
}
func (r *certRepo) ListAll(ctx context.Context) ([]domain.Certificate, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,user_id,cert_type,COALESCE(cert_number,''),COALESCE(level,''),issue_date,expire_date,COALESCE(issuer_org,''),COALESCE(image_url,''),status,version,created_at,updated_at FROM certificates ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list all certificates: %w", err)
	}
	defer rows.Close()
	var out []domain.Certificate
	for rows.Next() {
		var c domain.Certificate
		var ct string
		var issueAt, expireAt *time.Time
		if err := rows.Scan(&c.ID, &c.UserID, &ct, &c.CertNumber, &c.Level, &issueAt, &expireAt, &c.IssuerOrg, &c.ImageURL, &c.Status, &c.Version, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan certificate: %w", err)
		}
		c.CertType = domain.CertType(ct)
		c.IssueDate, c.ExpireDate = certTimeFromNull(issueAt), certTimeFromNull(expireAt)
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *certRepo) Update(ctx context.Context, c domain.Certificate) (domain.Certificate, error) {
	c.Version++
	c.UpdatedAt = time.Now()
	_, err := r.pool.Exec(ctx,
		`UPDATE certificates SET cert_type=$1,cert_number=$2,level=$3,issue_date=$4,expire_date=$5,issuer_org=$6,image_url=$7,status=$8,version=$9,updated_at=$10 WHERE id=$11`,
		string(c.CertType), c.CertNumber, c.Level, certTimeOrNull(c.IssueDate), certTimeOrNull(c.ExpireDate), c.IssuerOrg, c.ImageURL, c.Status, c.Version, c.UpdatedAt, c.ID)
	if err != nil {
		// 改号撞号同样映射哨兵（唯一索引覆盖 UPDATE 路径）。
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.Certificate{}, repository.ErrCertNumberTaken
		}
		return domain.Certificate{}, fmt.Errorf("update certificate: %w", err)
	}
	return c, nil
}

func (r *certRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM certificates WHERE id=$1`, id)
	return err
}

// ---- Course ----

type courseRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewCourseRepository() repository.CourseRepository { return &courseRepo{pool: s.Pool()} }

// courseCols 与 training_courses 表列一一对应（迁移 000044/000045 补齐小程序页面字段）
const courseCols = `id,org_id,org_name,title,cert_type,description,start_date,end_date,max_students,enrolled_count,location,district,price_fen,rating,review_count,duration_days,image,tags,certificate,courses,prices,business_hours,phone,remain,environment,course_types,pass_rate,years,status,version,created_at,updated_at`

// courseColsCount SELECT 目标数（与扫描目标一一对应；改表结构时必须同步更新）
const courseColsCount = 32

// VerifyCourseCols 启动自检：查询列定义并校验数量——新增列（如 pass_rate/years）漏补
// Scan 目标时，Select 列数与 Scan 目标不匹配会在该处暴露（LIMIT 0 空表同样校验）。
func (s *Store) VerifyCourseCols(ctx context.Context) error {
	rows, err := s.Pool().Query(ctx, `SELECT `+courseCols+` FROM training_courses LIMIT 0`)
	if err != nil {
		return fmt.Errorf("verify course cols: %w", err)
	}
	defer rows.Close()
	n := len(rows.FieldDescriptions())
	if n != courseColsCount {
		return fmt.Errorf("training_courses columns mismatch: got %d, want %d (courseCols/Scan 未同步?)", n, courseColsCount)
	}
	return nil
}

func (r *courseRepo) Create(ctx context.Context, c domain.TrainingCourse) (domain.TrainingCourse, error) {
	c.Version = 1
	c.CreatedAt = time.Now()
	c.UpdatedAt = c.CreatedAt
	c.Tags = jsonbSlice(c.Tags)
	c.Courses = jsonbSlice(c.Courses)
	c.Prices = jsonbSlice(c.Prices)
	c.Environment = jsonbSlice(c.Environment)
	c.CourseTypes = jsonbSlice(c.CourseTypes)
	_, err := r.pool.Exec(ctx,
		`INSERT INTO training_courses (`+courseCols+`)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27,$28,$29,$30,$31,$32)`,
		c.ID, c.OrgID, c.OrgName, c.Title, string(c.CertType), c.Description, c.StartDate, c.EndDate,
		c.MaxStudents, c.EnrolledCount, c.Location, c.District, c.PriceFen, c.Rating, c.ReviewCount,
		c.DurationDays, c.Image, c.Tags, c.Certificate, c.Courses, c.Prices, c.BusinessHours, c.Phone,
		c.Remain, c.Environment, c.CourseTypes, c.PassRate, c.Years,
		c.Status, c.Version, c.CreatedAt, c.UpdatedAt)
	return c, err
}
func (r *courseRepo) List(ctx context.Context) ([]domain.TrainingCourse, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+courseCols+` FROM training_courses ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list courses: %w", err)
	}
	defer rows.Close()
	var out []domain.TrainingCourse
	for rows.Next() {
		var c domain.TrainingCourse
		var ct string
		if err := rows.Scan(&c.ID, &c.OrgID, &c.OrgName, &c.Title, &ct, &c.Description, &c.StartDate, &c.EndDate,
			&c.MaxStudents, &c.EnrolledCount, &c.Location, &c.District, &c.PriceFen, &c.Rating, &c.ReviewCount,
			&c.DurationDays, &c.Image, &c.Tags, &c.Certificate, &c.Courses, &c.Prices, &c.BusinessHours, &c.Phone,
			&c.Remain, &c.Environment, &c.CourseTypes, &c.PassRate, &c.Years,
			&c.Status, &c.Version, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan course: %w", err)
		}
		c.CertType = domain.CertType(ct)
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *courseRepo) FindByID(ctx context.Context, id string) (domain.TrainingCourse, error) {
	var c domain.TrainingCourse
	var ct string
	err := r.pool.QueryRow(ctx,
		`SELECT `+courseCols+` FROM training_courses WHERE id=$1`, id).
		Scan(&c.ID, &c.OrgID, &c.OrgName, &c.Title, &ct, &c.Description, &c.StartDate, &c.EndDate,
			&c.MaxStudents, &c.EnrolledCount, &c.Location, &c.District, &c.PriceFen, &c.Rating, &c.ReviewCount,
			&c.DurationDays, &c.Image, &c.Tags, &c.Certificate, &c.Courses, &c.Prices, &c.BusinessHours, &c.Phone,
			&c.Remain, &c.Environment, &c.CourseTypes, &c.PassRate, &c.Years,
			&c.Status, &c.Version, &c.CreatedAt, &c.UpdatedAt)
	c.CertType = domain.CertType(ct)
	return c, err
}

func (r *courseRepo) Update(ctx context.Context, c domain.TrainingCourse) (domain.TrainingCourse, error) {
	c.Version++
	c.UpdatedAt = time.Now()
	c.Tags = jsonbSlice(c.Tags)
	c.Courses = jsonbSlice(c.Courses)
	c.Prices = jsonbSlice(c.Prices)
	c.Environment = jsonbSlice(c.Environment)
	c.CourseTypes = jsonbSlice(c.CourseTypes)
	_, err := r.pool.Exec(ctx,
		`UPDATE training_courses SET title=$1,cert_type=$2,description=$3,start_date=$4,end_date=$5,max_students=$6,location=$7,district=$8,price_fen=$9,rating=$10,review_count=$11,duration_days=$12,image=$13,tags=$14,certificate=$15,courses=$16,prices=$17,business_hours=$18,phone=$19,org_name=$20,remain=$21,environment=$22,course_types=$23,pass_rate=$24,years=$25,status=$26,version=$27,updated_at=$28 WHERE id=$29`,
		c.Title, string(c.CertType), c.Description, c.StartDate, c.EndDate, c.MaxStudents, c.Location,
		c.District, c.PriceFen, c.Rating, c.ReviewCount, c.DurationDays, c.Image, c.Tags, c.Certificate,
		c.Courses, c.Prices, c.BusinessHours, c.Phone, c.OrgName, c.Remain, c.Environment, c.CourseTypes,
		c.PassRate, c.Years, c.Status, c.Version, c.UpdatedAt, c.ID)
	return c, err
}

func (r *courseRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM training_courses WHERE id=$1`, id)
	return err
}
func (r *courseRepo) BumpEnrolled(ctx context.Context, id string, delta int) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE training_courses SET enrolled_count = GREATEST(0, enrolled_count + $2),
		 remain = GREATEST(0, max_students - (enrolled_count + $2)), updated_at = NOW() WHERE id=$1`,
		id, delta)
	if err != nil {
		return fmt.Errorf("bump enrolled %s: %w", id, err)
	}
	return nil
}

// ---- Course Favorites ----

func (r *courseRepo) FavoriteCourse(ctx context.Context, userID, courseID string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO training_course_favorites (id, user_id, course_id) VALUES ($1,$2,$3)
		 ON CONFLICT (user_id, course_id) DO NOTHING`,
		"cfav-"+userID+"-"+courseID, userID, courseID)
	if err != nil {
		return fmt.Errorf("favorite course %s: %w", courseID, err)
	}
	return nil
}

func (r *courseRepo) UnfavoriteCourse(ctx context.Context, userID, courseID string) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM training_course_favorites WHERE user_id=$1 AND course_id=$2`, userID, courseID)
	if err != nil {
		return fmt.Errorf("unfavorite course %s: %w", courseID, err)
	}
	return nil
}

// ListFavoriteCourses 按收藏时间倒序返回完整课程（我的收藏列表）。
// JOIN 查询必须给列加 c. 前缀：favorites 表同样有 id/created_at，未限定会 42702 歧义。
func (r *courseRepo) ListFavoriteCourses(ctx context.Context, userID string) ([]domain.TrainingCourse, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT c.id,c.org_id,c.org_name,c.title,c.cert_type,c.description,c.start_date,c.end_date,c.max_students,c.enrolled_count,c.location,c.district,c.price_fen,c.rating,c.review_count,c.duration_days,c.image,c.tags,c.certificate,c.courses,c.prices,c.business_hours,c.phone,c.remain,c.environment,c.course_types,c.pass_rate,c.years,c.status,c.version,c.created_at,c.updated_at FROM training_courses c
		 JOIN training_course_favorites f ON f.course_id = c.id
		 WHERE f.user_id=$1
		 ORDER BY f.created_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("list favorite courses: %w", err)
	}
	defer rows.Close()
	var out []domain.TrainingCourse
	for rows.Next() {
		var c domain.TrainingCourse
		var ct string
		if err := rows.Scan(&c.ID, &c.OrgID, &c.OrgName, &c.Title, &ct, &c.Description, &c.StartDate, &c.EndDate,
			&c.MaxStudents, &c.EnrolledCount, &c.Location, &c.District, &c.PriceFen, &c.Rating, &c.ReviewCount,
			&c.DurationDays, &c.Image, &c.Tags, &c.Certificate, &c.Courses, &c.Prices, &c.BusinessHours, &c.Phone,
			&c.Remain, &c.Environment, &c.CourseTypes, &c.PassRate, &c.Years,
			&c.Status, &c.Version, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan favorite course: %w", err)
		}
		c.CertType = domain.CertType(ct)
		out = append(out, c)
	}
	return out, rows.Err()
}

// ---- Instructor ----

type instructorRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewInstructorRepository() repository.InstructorRepository {
	return &instructorRepo{pool: s.Pool()}
}

func (r *instructorRepo) Create(ctx context.Context, i domain.Instructor) (domain.Instructor, error) {
	i.Version = 1
	i.CreatedAt = time.Now()
	i.UpdatedAt = i.CreatedAt
	certTypes, err := json.Marshal(i.CertTypes)
	if err != nil {
		return domain.Instructor{}, fmt.Errorf("marshal cert types: %w", err)
	}
	_, err = r.pool.Exec(ctx,
		`INSERT INTO instructors (id,user_id,name,photo,cert_types,bio,org_id,status,version,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		i.ID, i.UserID, i.Name, i.Photo, certTypes, i.Bio, i.OrgID, i.Status, i.Version, i.CreatedAt, i.UpdatedAt)
	return i, err
}
func (r *instructorRepo) FindByID(ctx context.Context, id string) (domain.Instructor, error) {
	var i domain.Instructor
	var ct []byte
	err := r.pool.QueryRow(ctx, `SELECT id,user_id,name,photo,cert_types,bio,org_id,status,version,created_at,updated_at FROM instructors WHERE id=$1`, id).
		Scan(&i.ID, &i.UserID, &i.Name, &i.Photo, &ct, &i.Bio, &i.OrgID, &i.Status, &i.Version, &i.CreatedAt, &i.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Instructor{}, fmt.Errorf("instructor %s: %w", id, repository.ErrNotFound)
	}
	json.Unmarshal(ct, &i.CertTypes)
	return i, err
}
func (r *instructorRepo) List(ctx context.Context) ([]domain.Instructor, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,user_id,name,photo,cert_types,bio,org_id,status,version,created_at,updated_at FROM instructors ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list instructors: %w", err)
	}
	defer rows.Close()
	var out []domain.Instructor
	for rows.Next() {
		var i domain.Instructor
		var ct []byte
		if err := rows.Scan(&i.ID, &i.UserID, &i.Name, &i.Photo, &ct, &i.Bio, &i.OrgID, &i.Status, &i.Version, &i.CreatedAt, &i.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan instructor: %w", err)
		}
		json.Unmarshal(ct, &i.CertTypes)
		out = append(out, i)
	}
	return out, rows.Err()
}
func (r *instructorRepo) UpdateStatus(ctx context.Context, id, status string) (domain.Instructor, error) {
	_, err := r.pool.Exec(ctx, `UPDATE instructors SET status=$1,updated_at=$2 WHERE id=$3`, status, time.Now(), id)
	if err != nil {
		return domain.Instructor{}, fmt.Errorf("update instructor status: %w", err)
	}
	return r.FindByID(ctx, id)
}
func (r *instructorRepo) Update(ctx context.Context, i domain.Instructor) (domain.Instructor, error) {
	i.UpdatedAt = time.Now()
	ct, err := json.Marshal(i.CertTypes)
	if err != nil {
		return domain.Instructor{}, fmt.Errorf("marshal instructor cert types: %w", err)
	}
	_, err = r.pool.Exec(ctx,
		`UPDATE instructors SET name=$1,photo=$2,cert_types=$3,bio=$4,org_id=$5,status=$6,version=version+1,updated_at=$7 WHERE id=$8`,
		i.Name, i.Photo, ct, i.Bio, i.OrgID, i.Status, i.UpdatedAt, i.ID)
	if err != nil {
		return domain.Instructor{}, fmt.Errorf("update instructor %s: %w", i.ID, err)
	}
	return i, nil
}

// ---- Pilot ----

type pilotRepo struct {
	pool   *pgxpool.Pool
	cipher *crypto.Cipher
}

func (s *Store) NewPilotRepository(cipher *crypto.Cipher) repository.PilotRepository {
	return &pilotRepo{pool: s.Pool(), cipher: cipher}
}

func (r *pilotRepo) enc(v string) string {
	if r.cipher != nil && v != "" {
		if e, err := r.cipher.Encrypt(v); err == nil {
			return e
		}
	}
	return v
}
func (r *pilotRepo) dec(v string) string {
	if r.cipher != nil && v != "" {
		if d, err := r.cipher.Decrypt(v); err == nil {
			return d
		}
		// 解密失败（密钥变更/数据损坏）绝不回传密文——置空而非泄露加密串。
		return ""
	}
	return v
}
func (r *pilotRepo) Create(ctx context.Context, p domain.CertifiedPilot) (domain.CertifiedPilot, error) {
	p.Version = 1
	p.CreatedAt = time.Now()
	p.UpdatedAt = p.CreatedAt
	p.IDCard = r.enc(p.IDCard)
	certIDs, err := json.Marshal(p.CertIDs)
	if err != nil {
		return domain.CertifiedPilot{}, fmt.Errorf("marshal cert ids: %w", err)
	}
	_, err = r.pool.Exec(ctx,
		`INSERT INTO certified_pilots (id,user_id,real_name,id_card,avatar,region,cert_ids,flight_hours,bio,rating,completed_jobs,status,version,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		p.ID, p.UserID, p.RealName, p.IDCard, p.Avatar, p.Region, certIDs, p.FlightHours, p.Bio, p.Rating, p.CompletedJobs, p.Status, p.Version, p.CreatedAt, p.UpdatedAt)
	p.IDCard = r.dec(p.IDCard)
	return p, err
}
func (r *pilotRepo) Update(ctx context.Context, p domain.CertifiedPilot) (domain.CertifiedPilot, error) {
	p.UpdatedAt = time.Now()
	p.IDCard = r.enc(p.IDCard)
	certIDs, err := json.Marshal(p.CertIDs)
	if err != nil {
		return domain.CertifiedPilot{}, fmt.Errorf("marshal cert ids: %w", err)
	}
	_, err = r.pool.Exec(ctx,
		`UPDATE certified_pilots SET real_name=$1,id_card=$2,avatar=$3,region=$4,cert_ids=$5,flight_hours=$6,bio=$7,status=$8,updated_at=$9 WHERE id=$10`,
		p.RealName, p.IDCard, p.Avatar, p.Region, certIDs, p.FlightHours, p.Bio, p.Status, p.UpdatedAt, p.ID)
	if err != nil {
		return domain.CertifiedPilot{}, fmt.Errorf("update pilot: %w", err)
	}
	return r.FindByID(ctx, p.ID)
}

func (r *pilotRepo) FindByID(ctx context.Context, id string) (domain.CertifiedPilot, error) {
	var p domain.CertifiedPilot
	var certIDs []byte
	err := r.pool.QueryRow(ctx, `SELECT id,user_id,real_name,COALESCE(id_card,''),COALESCE(avatar,''),COALESCE(region,''),cert_ids,flight_hours,bio,rating,completed_jobs,status,reject_reason,version,created_at,updated_at FROM certified_pilots WHERE id=$1`, id).
		Scan(&p.ID, &p.UserID, &p.RealName, &p.IDCard, &p.Avatar, &p.Region, &certIDs, &p.FlightHours, &p.Bio, &p.Rating, &p.CompletedJobs, &p.Status, &p.RejectReason, &p.Version, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.CertifiedPilot{}, fmt.Errorf("pilot %s: %w", id, repository.ErrNotFound)
	}
	json.Unmarshal(certIDs, &p.CertIDs)
	p.IDCard = r.dec(p.IDCard)
	return p, err
}
func (r *pilotRepo) List(ctx context.Context) ([]domain.CertifiedPilot, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,user_id,real_name,COALESCE(id_card,''),COALESCE(avatar,''),COALESCE(region,''),cert_ids,flight_hours,bio,rating,completed_jobs,status,reject_reason,version,created_at,updated_at FROM certified_pilots ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list pilots: %w", err)
	}
	defer rows.Close()
	var out []domain.CertifiedPilot
	for rows.Next() {
		var p domain.CertifiedPilot
		var certIDs []byte
		if err := rows.Scan(&p.ID, &p.UserID, &p.RealName, &p.IDCard, &p.Avatar, &p.Region, &certIDs, &p.FlightHours, &p.Bio, &p.Rating, &p.CompletedJobs, &p.Status, &p.RejectReason, &p.Version, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan pilot: %w", err)
		}
		json.Unmarshal(certIDs, &p.CertIDs)
		p.IDCard = r.dec(p.IDCard)
		out = append(out, p)
	}
	return out, rows.Err()
}

// ListApproved 公开名录分页：仅 status='approved'，keyword 匹配姓名，COUNT + LIMIT/OFFSET。
func (r *pilotRepo) ListApproved(ctx context.Context, keyword string, offset, limit int) ([]domain.CertifiedPilot, int, error) {
	where := `WHERE status='approved'`
	args := []any{}
	if k := strings.TrimSpace(keyword); k != "" {
		args = append(args, "%"+escapeLike(k)+"%")
		where += fmt.Sprintf(` AND real_name ILIKE $%d ESCAPE '\'`, len(args))
	}
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT count(*) FROM certified_pilots `+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count approved pilots: %w", err)
	}
	args = append(args, limit, offset)
	rows, err := r.pool.Query(ctx, `SELECT id,user_id,real_name,COALESCE(id_card,''),COALESCE(avatar,''),COALESCE(region,''),cert_ids,flight_hours,bio,rating,completed_jobs,status,reject_reason,version,created_at,updated_at FROM certified_pilots `+where+
		fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, len(args)-1, len(args)), args...)
	if err != nil {
		return nil, total, fmt.Errorf("list approved pilots: %w", err)
	}
	defer rows.Close()
	var out []domain.CertifiedPilot
	for rows.Next() {
		var p domain.CertifiedPilot
		var certIDs []byte
		if err := rows.Scan(&p.ID, &p.UserID, &p.RealName, &p.IDCard, &p.Avatar, &p.Region, &certIDs, &p.FlightHours, &p.Bio, &p.Rating, &p.CompletedJobs, &p.Status, &p.RejectReason, &p.Version, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, total, fmt.Errorf("scan pilot: %w", err)
		}
		json.Unmarshal(certIDs, &p.CertIDs)
		p.IDCard = r.dec(p.IDCard)
		out = append(out, p)
	}
	return out, total, rows.Err()
}
func (r *pilotRepo) UpdateStatus(ctx context.Context, id, status string) (domain.CertifiedPilot, error) {
	// 一并清掉 reject_reason：本方法唯一调用方是 ApprovePilot（改为 approved）。
	// 此前这里不清、RegisterPilot 的"驳回后重提"也不清，于是 reject_reason 只写不删——
	// 一条 pending/approved 记录会一直挂着上一轮的驳回理由。
	_, err := r.pool.Exec(ctx, `UPDATE certified_pilots SET status=$1,reject_reason='',updated_at=$2 WHERE id=$3`, status, time.Now(), id)
	if err != nil {
		return domain.CertifiedPilot{}, fmt.Errorf("update pilot status: %w", err)
	}
	return r.FindByID(ctx, id)
}

func (r *pilotRepo) UpdateReject(ctx context.Context, id, reason string) (domain.CertifiedPilot, error) {
	_, err := r.pool.Exec(ctx, `UPDATE certified_pilots SET status='rejected',reject_reason=$2,updated_at=$3 WHERE id=$1`, id, reason, time.Now())
	if err != nil {
		return domain.CertifiedPilot{}, fmt.Errorf("reject pilot: %w", err)
	}
	return r.FindByID(ctx, id)
}

// ---- Product ----

type prodRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewProductRepository() repository.ProductRepository { return &prodRepo{pool: s.Pool()} }

// productColumns 商品查询列的**唯一来源**。
//
// 此前这段列清单在 7 个查询里各写一份，列数与扫描目标极易漂移——同一类事故已经发生过
// （ListFavoriteDemands 的 SELECT 是 20 列、scanDemands 要 24 个目标，只在生产 PG 下 500）。
// 新增列时只改这里和 scanProduct（以及 Create/Update 的列名列表），
// 两处对不上会立刻扫描报错，而不是静默错位。
const productColumns = `id,seller_id,seller_name,prod_type,title,COALESCE(description,''),price_fen,price_mode,delivery,images,COALESCE(detail_images,'[]'),COALESCE(brand,''),COALESCE(model,''),condition,COALESCE(category,''),COALESCE(region,''),COALESCE(unit,''),views,status,check_status,COALESCE(check_reason,''),reviewed_at,reviewed_by,version,created_at,updated_at`

// productColumnsAliased 与 productColumns **同序同义**，只给列名加 p. 前缀，
// 供需要 JOIN 的查询使用（ListFavoriteProducts 要按收藏时间排序，不能用子查询改写）。
// 两者必须一起改——所以紧挨着放。
const productColumnsAliased = `p.id,p.seller_id,p.seller_name,p.prod_type,p.title,COALESCE(p.description,''),p.price_fen,p.price_mode,p.delivery,p.images,COALESCE(p.detail_images,'[]'),COALESCE(p.brand,''),COALESCE(p.model,''),p.condition,COALESCE(p.category,''),COALESCE(p.region,''),COALESCE(p.unit,''),p.views,p.status,p.check_status,COALESCE(p.check_reason,''),p.reviewed_at,p.reviewed_by,p.version,p.created_at,p.updated_at`

// rowScanner 同时覆盖 pgx.Row 与 pgx.Rows。
type rowScanner interface{ Scan(dest ...any) error }

// scanProduct 按 productColumns 的顺序扫描一行商品。
func scanProduct(row rowScanner) (domain.DroneProduct, error) {
	var (
		p          domain.DroneProduct
		pt         string
		imgs       []byte
		detailImgs []byte
		reviewedAt *time.Time
		reviewedBy string
	)
	if err := row.Scan(&p.ID, &p.SellerID, &p.SellerName, &pt, &p.Title, &p.Description, &p.PriceFen, &p.PriceMode, &p.Delivery, &imgs,
		&detailImgs, &p.Brand, &p.Model, &p.Condition, &p.Category, &p.Region, &p.Unit,
		&p.Views, &p.Status, &p.CheckStatus, &p.CheckReason,
		&reviewedAt, &reviewedBy, &p.Version, &p.CreatedAt, &p.UpdatedAt); err != nil {
		return domain.DroneProduct{}, err
	}
	p.ProdType = domain.ProductType(pt)
	p.ReviewedAt = reviewedAt
	p.ReviewedBy = reviewedBy
	json.Unmarshal(imgs, &p.Images)
	json.Unmarshal(detailImgs, &p.DetailImages)
	return p, nil
}

func (r *prodRepo) Create(ctx context.Context, p domain.DroneProduct) (domain.DroneProduct, error) {
	p.Version = 1
	p.CreatedAt = time.Now()
	p.UpdatedAt = p.CreatedAt
	images, err := json.Marshal(p.Images)
	if err != nil {
		return domain.DroneProduct{}, fmt.Errorf("marshal product images: %w", err)
	}
	detailImages, err := json.Marshal(p.DetailImages)
	if err != nil {
		return domain.DroneProduct{}, fmt.Errorf("marshal product detail images: %w", err)
	}
	if p.CheckStatus == "" {
		p.CheckStatus = domain.ProductCheckPending
	}
	_, err = r.pool.Exec(ctx,
		`INSERT INTO drone_products (id,seller_id,seller_name,prod_type,title,description,price_fen,price_mode,delivery,images,detail_images,brand,model,condition,category,region,unit,views,status,check_status,check_reason,version,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24)`,
		p.ID, p.SellerID, p.SellerName, string(p.ProdType), p.Title, p.Description, p.PriceFen, p.PriceMode, p.Delivery, images, detailImages, p.Brand, p.Model, p.Condition, p.Category, p.Region, p.Unit, p.Views, p.Status, p.CheckStatus, p.CheckReason, p.Version, p.CreatedAt, p.UpdatedAt)
	return p, err
}
func (r *prodRepo) FindByID(ctx context.Context, id string) (domain.DroneProduct, error) {
	p, err := scanProduct(r.pool.QueryRow(ctx,
		`SELECT `+productColumns+` FROM drone_products WHERE id=$1 AND deleted_at IS NULL`, id))
	if err != nil {
		return domain.DroneProduct{}, fmt.Errorf("product %s not found: %w", id, err)
	}
	return p, nil
}

// SetStatusBulk 批量改上架状态（管理端）：单条 UPDATE，条件全部在 WHERE 里。
func (r *prodRepo) SetStatusBulk(ctx context.Context, ids []string, status string) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	tag, err := r.pool.Exec(ctx,
		`UPDATE drone_products SET status=$2, version=version+1, updated_at=NOW()
		  WHERE id = ANY($1) AND deleted_at IS NULL
		    AND status <> 'sold'
		    AND ($2 <> 'listed' OR check_status = $3)`,
		ids, status, domain.ProductCheckPassed)
	if err != nil {
		return 0, fmt.Errorf("bulk set product status: %w", err)
	}
	return int(tag.RowsAffected()), nil
}

// SetProductStatus 卖家自助上下架。条件全部写进 WHERE，避免"先查后改"的竞态：
// 归属（seller_id）+ 未删除 + 非已售 + 上架时必须已过审。
func (r *prodRepo) SetProductStatus(ctx context.Context, id, sellerID, status string) (domain.DroneProduct, error) {
	p, err := scanProduct(r.pool.QueryRow(ctx,
		`UPDATE drone_products SET status=$4, version=version+1, updated_at=NOW()
		  WHERE id=$1 AND seller_id=$2 AND deleted_at IS NULL
		    AND status <> 'sold'
		    AND ($4 <> 'listed' OR check_status = $3)
		 RETURNING `+productColumns,
		id, sellerID, domain.ProductCheckPassed, status))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.DroneProduct{}, fmt.Errorf("product %s: %w", id, repository.ErrNotFound)
		}
		return domain.DroneProduct{}, fmt.Errorf("set product %s status: %w", id, err)
	}
	return p, nil
}

// ReviewProduct 审核商品（条件更新：仅待审/已驳回的行可被审）。
//
// 审核通过同时把 status 置为 listed（与旧的"通过即上架"行为一致）；
// 驳回**不动 status**——驳回 ≠ 下架，这正是把审核维度拆出来的目的。
func (r *prodRepo) ReviewProduct(ctx context.Context, id, checkStatus, reason, reviewerID string) (domain.DroneProduct, error) {
	newStatus := "pending"
	if checkStatus == domain.ProductCheckPassed {
		newStatus = "listed"
	}
	tag, err := r.pool.Exec(ctx,
		`UPDATE drone_products
		    SET check_status=$2, check_reason=$3, reviewed_at=NOW(), reviewed_by=$4,
		        status=$5, version=version+1, updated_at=NOW()
		  WHERE id=$1 AND deleted_at IS NULL AND check_status IN ('pending','rejected')`,
		id, checkStatus, reason, reviewerID, newStatus)
	if err != nil {
		return domain.DroneProduct{}, fmt.Errorf("review product %s: %w", id, err)
	}
	if tag.RowsAffected() == 0 {
		// 不存在 / 已进回收站 / 已终审（含并发重复审核）
		return domain.DroneProduct{}, fmt.Errorf("product %s: %w", id, repository.ErrNotFound)
	}
	return r.FindByID(ctx, id)
}

func (r *prodRepo) IncrementViews(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `UPDATE drone_products SET views = views + 1 WHERE id=$1 AND deleted_at IS NULL`, id)
	if err != nil {
		return fmt.Errorf("increment views for %s: %w", id, err)
	}
	return nil
}

// MarkSold 下单抢占：仅 listed/空状态可标记 sold（条件更新防一物多卖/超卖）。
func (r *prodRepo) MarkSold(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE drone_products SET status='sold', version=version+1, updated_at=NOW() WHERE id=$1 AND status IN ('','listed') AND deleted_at IS NULL`, id)
	if err != nil {
		return fmt.Errorf("mark product %s sold: %w", id, err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("product %s not available", id)
	}
	return nil
}

// Restore 订单创建失败回滚：sold → listed（条件更新，仅 sold 可恢复）。
func (r *prodRepo) Restore(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE drone_products SET status='listed', version=version+1, updated_at=NOW() WHERE id=$1 AND status='sold' AND deleted_at IS NULL`, id)
	if err != nil {
		return fmt.Errorf("restore product %s: %w", id, err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("product %s not in sold state", id)
	}
	return nil
}

func (r *prodRepo) Update(ctx context.Context, p domain.DroneProduct) (domain.DroneProduct, error) {
	p.UpdatedAt = time.Now()
	images, err := json.Marshal(p.Images)
	if err != nil {
		return domain.DroneProduct{}, fmt.Errorf("marshal product images: %w", err)
	}
	detailImages, err := json.Marshal(p.DetailImages)
	if err != nil {
		return domain.DroneProduct{}, fmt.Errorf("marshal product detail images: %w", err)
	}
	tag, err := r.pool.Exec(ctx,
		`UPDATE drone_products SET seller_id=$1,seller_name=$2,prod_type=$3,title=$4,description=$5,price_fen=$6,price_mode=$7,delivery=$8,images=$9,detail_images=$10,brand=$11,model=$12,condition=$13,category=$14,region=$15,unit=$16,status=$17,check_status=$18,check_reason=$19,version=version+1,updated_at=$20 WHERE id=$21 AND deleted_at IS NULL`,
		p.SellerID, p.SellerName, string(p.ProdType), p.Title, p.Description, p.PriceFen, p.PriceMode, p.Delivery, images, detailImages, p.Brand, p.Model, p.Condition, p.Category, p.Region, p.Unit, p.Status, p.CheckStatus, p.CheckReason, p.UpdatedAt, p.ID)
	if err != nil {
		return domain.DroneProduct{}, fmt.Errorf("update product %s: %w", p.ID, err)
	}
	// 0 行 = 不存在或已在回收站。此前不检查 RowsAffected，管理端改回收站商品会静默"成功"。
	if tag.RowsAffected() == 0 {
		return domain.DroneProduct{}, fmt.Errorf("product %s: %w", p.ID, repository.ErrNotFound)
	}
	return p, nil
}

// SoftDelete 软删除（回收站）：只写 deleted_at，不物理删除。
// 物理删除会让 trade_orders.product_id 变成指向不存在行的孤儿（该列无外键），
// 历史订单的商品名与详情链接一并丢失——用户侧现象就是"商品消失了"。
// "AND deleted_at IS NULL" 让重复删除返回 ErrNotFound，而不是假装成功。
func (r *prodRepo) SoftDelete(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE drone_products SET deleted_at=NOW(), version=version+1, updated_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, id)
	if err != nil {
		return fmt.Errorf("soft delete product %s: %w", id, err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("product %s: %w", id, repository.ErrNotFound)
	}
	return nil
}

// Undelete 回收站还原：清空 deleted_at（仅对已在回收站的行生效）。
func (r *prodRepo) Undelete(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE drone_products SET deleted_at=NULL, version=version+1, updated_at=NOW() WHERE id=$1 AND deleted_at IS NOT NULL`, id)
	if err != nil {
		return fmt.Errorf("undelete product %s: %w", id, err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("product %s: %w", id, repository.ErrNotFound)
	}
	return nil
}

// ListDeleted 回收站列表：仅 deleted_at 非空的行，按删除时间倒序。
func (r *prodRepo) ListDeleted(ctx context.Context) ([]domain.DroneProduct, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+productColumns+` FROM drone_products WHERE deleted_at IS NOT NULL ORDER BY deleted_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list deleted products: %w", err)
	}
	defer rows.Close()
	var out []domain.DroneProduct
	for rows.Next() {
		p, err := scanProduct(rows)
		if err != nil {
			return nil, fmt.Errorf("scan deleted product: %w", err)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *prodRepo) List(ctx context.Context, prodType string) ([]domain.DroneProduct, error) {
	var rows pgx.Rows
	var err error
	if prodType == "" {
		rows, err = r.pool.Query(ctx, `SELECT `+productColumns+` FROM drone_products WHERE deleted_at IS NULL ORDER BY created_at DESC`)
	} else {
		rows, err = r.pool.Query(ctx, `SELECT `+productColumns+` FROM drone_products WHERE prod_type=$1 AND deleted_at IS NULL ORDER BY created_at DESC`, prodType)
	}
	if err != nil {
		return nil, fmt.Errorf("list products: %w", err)
	}
	defer rows.Close()
	var out []domain.DroneProduct
	for rows.Next() {
		p, err := scanProduct(rows)
		if err != nil {
			return nil, fmt.Errorf("scan product: %w", err)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// ListSoldBefore 列出 status='sold' 且 updated_at 早于 cutoff 的商品（孤儿已售商品回收用）。
func (r *prodRepo) ListSoldBefore(ctx context.Context, cutoff time.Time, limit int) ([]domain.DroneProduct, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := r.pool.Query(ctx, `SELECT `+productColumns+`
		FROM drone_products WHERE status='sold' AND check_status='passed' AND updated_at < $1 AND deleted_at IS NULL ORDER BY updated_at ASC LIMIT $2`, cutoff, limit)
	if err != nil {
		return nil, fmt.Errorf("list sold products before cutoff: %w", err)
	}
	defer rows.Close()
	out := []domain.DroneProduct{}
	for rows.Next() {
		p, err := scanProduct(rows)
		if err != nil {
			return nil, fmt.Errorf("scan sold product: %w", err)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// ListTop 按创建时间倒序取前 limit 条（首页 Top-N，SQL 端 LIMIT 不整表）。
// 仅取已上架商品：待审核/下架/已售商品不得出现在首页公开区（P0 半断修复）。
func (r *prodRepo) ListTop(ctx context.Context, prodType string, limit int) ([]domain.DroneProduct, error) {
	var rows pgx.Rows
	var err error
	if prodType == "" {
		rows, err = r.pool.Query(ctx, `SELECT `+productColumns+` FROM drone_products WHERE status='listed' AND deleted_at IS NULL ORDER BY created_at DESC LIMIT $1`, limit)
	} else {
		rows, err = r.pool.Query(ctx, `SELECT `+productColumns+` FROM drone_products WHERE prod_type=$1 AND status='listed' AND deleted_at IS NULL ORDER BY created_at DESC LIMIT $2`, prodType, limit)
	}
	if err != nil {
		return nil, fmt.Errorf("list top products: %w", err)
	}
	defer rows.Close()
	var out []domain.DroneProduct
	for rows.Next() {
		p, err := scanProduct(rows)
		if err != nil {
			return nil, fmt.Errorf("scan product: %w", err)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// ---- Product Favorites ----

func (r *prodRepo) FavoriteProduct(ctx context.Context, userID, productID string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO product_favorites (id, user_id, product_id) VALUES ($1,$2,$3)
		 ON CONFLICT (user_id, product_id) DO NOTHING`,
		"pfav-"+userID+"-"+productID, userID, productID)
	if err != nil {
		return fmt.Errorf("favorite product %s: %w", productID, err)
	}
	return nil
}

func (r *prodRepo) UnfavoriteProduct(ctx context.Context, userID, productID string) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM product_favorites WHERE user_id=$1 AND product_id=$2`, userID, productID)
	if err != nil {
		return fmt.Errorf("unfavorite product %s: %w", productID, err)
	}
	return nil
}

// ListFavoriteProducts 按收藏时间倒序返回完整商品（我的收藏列表）。
func (r *prodRepo) ListFavoriteProducts(ctx context.Context, userID string) ([]domain.DroneProduct, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+productColumnsAliased+`
		 FROM drone_products p
		 JOIN product_favorites f ON f.product_id = p.id
		 WHERE f.user_id=$1 AND p.deleted_at IS NULL
		 ORDER BY f.created_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("list favorite products: %w", err)
	}
	defer rows.Close()
	var out []domain.DroneProduct
	for rows.Next() {
		p, err := scanProduct(rows)
		if err != nil {
			return nil, fmt.Errorf("scan favorite product: %w", err)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// ListByIDs 批量按 ID 取商品（订单列表补商品名防 N+1）。
//
// 这里**故意不过滤 deleted_at**：商品进回收站后，历史订单仍要显示出商品名。
// 软删除的全部意义就在于保住这条关联——过滤掉它等于退回物理删除的老毛病。
func (r *prodRepo) ListByIDs(ctx context.Context, ids []string) ([]domain.DroneProduct, error) {
	if len(ids) == 0 {
		return []domain.DroneProduct{}, nil
	}
	rows, err := r.pool.Query(ctx, `SELECT `+productColumns+` FROM drone_products WHERE id = ANY($1)`, ids)
	if err != nil {
		return nil, fmt.Errorf("list products by ids: %w", err)
	}
	defer rows.Close()
	var out []domain.DroneProduct
	for rows.Next() {
		p, err := scanProduct(rows)
		if err != nil {
			return nil, fmt.Errorf("scan product: %w", err)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// SumViews 商品浏览量总和（可选按类型；首页 stats.views 聚合查询）。
func (r *prodRepo) SumViews(ctx context.Context, prodType string) (int, error) {
	q := `SELECT COALESCE(SUM(views),0) FROM drone_products WHERE deleted_at IS NULL`
	args := []any{}
	if prodType != "" {
		q += ` AND prod_type=$1`
		args = append(args, prodType)
	}
	var n int
	if err := r.pool.QueryRow(ctx, q, args...).Scan(&n); err != nil {
		return 0, fmt.Errorf("sum product views: %w", err)
	}
	return n, nil
}

// ---- Repair ----

type repairRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewRepairRepository() repository.RepairRepository { return &repairRepo{pool: s.Pool()} }

func (r *repairRepo) Create(ctx context.Context, ro domain.RepairOrder) (domain.RepairOrder, error) {
	ro.Version = 1
	ro.CreatedAt = time.Now()
	ro.UpdatedAt = ro.CreatedAt
	_, err := r.pool.Exec(ctx,
		`INSERT INTO repair_orders (id,customer_id,product_desc,fault_desc,quote_fen,status,technician,version,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		ro.ID, ro.CustomerID, ro.ProductDesc, ro.FaultDesc, ro.QuoteFen, ro.Status, ro.Technician, ro.Version, ro.CreatedAt, ro.UpdatedAt)
	return ro, err
}
func (r *repairRepo) ListByUser(ctx context.Context, userID string) ([]domain.RepairOrder, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,customer_id,COALESCE(product_desc,''),COALESCE(fault_desc,''),quote_fen,status,COALESCE(technician,''),version,created_at,updated_at FROM repair_orders WHERE customer_id=$1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("list repairs: %w", err)
	}
	defer rows.Close()
	var out []domain.RepairOrder
	for rows.Next() {
		var ro domain.RepairOrder
		if err := rows.Scan(&ro.ID, &ro.CustomerID, &ro.ProductDesc, &ro.FaultDesc, &ro.QuoteFen, &ro.Status, &ro.Technician, &ro.Version, &ro.CreatedAt, &ro.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, ro)
	}
	return out, rows.Err()
}
func (r *repairRepo) ListAll(ctx context.Context, offset, limit int) ([]domain.RepairOrder, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT count(*) FROM repair_orders`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.pool.Query(ctx, `SELECT id,customer_id,COALESCE(product_desc,''),COALESCE(fault_desc,''),quote_fen,status,COALESCE(technician,''),version,created_at,updated_at FROM repair_orders ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []domain.RepairOrder
	for rows.Next() {
		var ro domain.RepairOrder
		if err := rows.Scan(&ro.ID, &ro.CustomerID, &ro.ProductDesc, &ro.FaultDesc, &ro.QuoteFen, &ro.Status, &ro.Technician, &ro.Version, &ro.CreatedAt, &ro.UpdatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, ro)
	}
	return out, total, rows.Err()
}

// ---- Policy ----

type policyRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewPolicyRepository() repository.PolicyRepository { return &policyRepo{pool: s.Pool()} }

func (r *policyRepo) Create(ctx context.Context, p domain.InsurancePolicy) (domain.InsurancePolicy, error) {
	p.Version = 1
	p.CreatedAt = time.Now()
	p.UpdatedAt = p.CreatedAt
	_, err := r.pool.Exec(ctx,
		`INSERT INTO insurance_policies (id,user_id,drone_model,drone_sn,policy_type,premium_fen,coverage_fen,start_date,end_date,insurer,status,version,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		p.ID, p.UserID, p.DroneModel, p.DroneSN, p.PolicyType, p.PremiumFen, p.CoverageFen, p.StartDate, p.EndDate, p.Insurer, p.Status, p.Version, p.CreatedAt, p.UpdatedAt)
	return p, err
}
func (r *policyRepo) ListByUser(ctx context.Context, userID string) ([]domain.InsurancePolicy, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,user_id,COALESCE(drone_model,''),COALESCE(drone_sn,''),COALESCE(policy_type,''),premium_fen,coverage_fen,COALESCE(start_date,'1970-01-01 00:00:00+00'::timestamptz),COALESCE(end_date,'1970-01-01 00:00:00+00'::timestamptz),COALESCE(insurer,''),status,version,created_at,updated_at FROM insurance_policies WHERE user_id=$1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("list policies: %w", err)
	}
	defer rows.Close()
	var out []domain.InsurancePolicy
	for rows.Next() {
		var p domain.InsurancePolicy
		if err := rows.Scan(&p.ID, &p.UserID, &p.DroneModel, &p.DroneSN, &p.PolicyType, &p.PremiumFen, &p.CoverageFen, &p.StartDate, &p.EndDate, &p.Insurer, &p.Status, &p.Version, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}
func (r *policyRepo) ListAll(ctx context.Context, offset, limit int) ([]domain.InsurancePolicy, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT count(*) FROM insurance_policies`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.pool.Query(ctx, `SELECT id,user_id,COALESCE(drone_model,''),COALESCE(drone_sn,''),COALESCE(policy_type,''),premium_fen,coverage_fen,COALESCE(start_date,'1970-01-01 00:00:00+00'::timestamptz),COALESCE(end_date,'1970-01-01 00:00:00+00'::timestamptz),COALESCE(insurer,''),status,version,created_at,updated_at FROM insurance_policies ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []domain.InsurancePolicy
	for rows.Next() {
		var p domain.InsurancePolicy
		if err := rows.Scan(&p.ID, &p.UserID, &p.DroneModel, &p.DroneSN, &p.PolicyType, &p.PremiumFen, &p.CoverageFen, &p.StartDate, &p.EndDate, &p.Insurer, &p.Status, &p.Version, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, p)
	}
	return out, total, rows.Err()
}
