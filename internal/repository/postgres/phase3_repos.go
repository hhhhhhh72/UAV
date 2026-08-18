package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
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

func (r *certRepo) Create(ctx context.Context, c domain.Certificate) (domain.Certificate, error) {
	c.Version = 1
	c.CreatedAt = time.Now()
	c.UpdatedAt = c.CreatedAt
	_, err := r.pool.Exec(ctx,
		`INSERT INTO certificates (id,user_id,cert_type,cert_number,level,issue_date,expire_date,issuer_org,image_url,status,version,created_at,updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		c.ID, c.UserID, string(c.CertType), c.CertNumber, c.Level, c.IssueDate, c.ExpireDate, c.IssuerOrg, c.ImageURL, c.Status, c.Version, c.CreatedAt, c.UpdatedAt)
	return c, err
}
func (r *certRepo) FindByID(ctx context.Context, id string) (domain.Certificate, error) {
	var c domain.Certificate
	var ct string
	err := r.pool.QueryRow(ctx,
		`SELECT id,user_id,cert_type,cert_number,level,issue_date,expire_date,issuer_org,image_url,status,version,created_at,updated_at FROM certificates WHERE id=$1`, id).
		Scan(&c.ID, &c.UserID, &ct, &c.CertNumber, &c.Level, &c.IssueDate, &c.ExpireDate, &c.IssuerOrg, &c.ImageURL, &c.Status, &c.Version, &c.CreatedAt, &c.UpdatedAt)
	c.CertType = domain.CertType(ct)
	return c, err
}
func (r *certRepo) ListByUser(ctx context.Context, userID string) ([]domain.Certificate, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id,user_id,cert_type,cert_number,level,issue_date,expire_date,issuer_org,image_url,status,version,created_at,updated_at FROM certificates WHERE user_id=$1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("list certificates: %w", err)
	}
	defer rows.Close()
	var out []domain.Certificate
	for rows.Next() {
		var c domain.Certificate
		var ct string
		if err := rows.Scan(&c.ID, &c.UserID, &ct, &c.CertNumber, &c.Level, &c.IssueDate, &c.ExpireDate, &c.IssuerOrg, &c.ImageURL, &c.Status, &c.Version, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan certificate: %w", err)
		}
		c.CertType = domain.CertType(ct)
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
	rows, err := r.pool.Query(ctx, `SELECT id,user_id,cert_type,cert_number,level,COALESCE(issue_date,'1970-01-01'::timestamptz),COALESCE(expire_date,'1970-01-01'::timestamptz),issuer_org,COALESCE(image_url,''),status,version,created_at,updated_at FROM certificates ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list all certificates: %w", err)
	}
	defer rows.Close()
	var out []domain.Certificate
	for rows.Next() {
		var c domain.Certificate
		var ct string
		if err := rows.Scan(&c.ID, &c.UserID, &ct, &c.CertNumber, &c.Level, &c.IssueDate, &c.ExpireDate, &c.IssuerOrg, &c.ImageURL, &c.Status, &c.Version, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan certificate: %w", err)
		}
		c.CertType = domain.CertType(ct)
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *certRepo) Update(ctx context.Context, c domain.Certificate) (domain.Certificate, error) {
	c.Version++
	c.UpdatedAt = time.Now()
	_, err := r.pool.Exec(ctx,
		`UPDATE certificates SET cert_type=$1,cert_number=$2,level=$3,issue_date=$4,expire_date=$5,issuer_org=$6,image_url=$7,status=$8,version=$9,updated_at=$10 WHERE id=$11`,
		string(c.CertType), c.CertNumber, c.Level, c.IssueDate, c.ExpireDate, c.IssuerOrg, c.ImageURL, c.Status, c.Version, c.UpdatedAt, c.ID)
	return c, err
}

func (r *certRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM certificates WHERE id=$1`, id)
	return err
}

// ---- Course ----

type courseRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewCourseRepository() repository.CourseRepository { return &courseRepo{pool: s.Pool()} }

// courseCols 与 training_courses 表列一一对应（迁移 000044/000045 补齐小程序页面字段）
const courseCols = `id,org_id,org_name,title,cert_type,description,start_date,end_date,max_students,enrolled_count,location,district,price_fen,rating,review_count,duration_days,image,tags,certificate,courses,prices,business_hours,phone,remain,environment,course_types,status,version,created_at,updated_at`

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
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27,$28,$29,$30)`,
		c.ID, c.OrgID, c.OrgName, c.Title, string(c.CertType), c.Description, c.StartDate, c.EndDate,
		c.MaxStudents, c.EnrolledCount, c.Location, c.District, c.PriceFen, c.Rating, c.ReviewCount,
		c.DurationDays, c.Image, c.Tags, c.Certificate, c.Courses, c.Prices, c.BusinessHours, c.Phone,
		c.Remain, c.Environment, c.CourseTypes,
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
			&c.Remain, &c.Environment, &c.CourseTypes,
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
			&c.Remain, &c.Environment, &c.CourseTypes,
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
		`UPDATE training_courses SET title=$1,cert_type=$2,description=$3,start_date=$4,end_date=$5,max_students=$6,location=$7,district=$8,price_fen=$9,rating=$10,review_count=$11,duration_days=$12,image=$13,tags=$14,certificate=$15,courses=$16,prices=$17,business_hours=$18,phone=$19,org_name=$20,remain=$21,environment=$22,course_types=$23,status=$24,version=$25,updated_at=$26 WHERE id=$27`,
		c.Title, string(c.CertType), c.Description, c.StartDate, c.EndDate, c.MaxStudents, c.Location,
		c.District, c.PriceFen, c.Rating, c.ReviewCount, c.DurationDays, c.Image, c.Tags, c.Certificate,
		c.Courses, c.Prices, c.BusinessHours, c.Phone, c.OrgName, c.Remain, c.Environment, c.CourseTypes,
		c.Status, c.Version, c.UpdatedAt, c.ID)
	return c, err
}

func (r *courseRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM training_courses WHERE id=$1`, id)
	return err
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
	err := r.pool.QueryRow(ctx, `SELECT id,user_id,real_name,id_card,avatar,region,cert_ids,flight_hours,bio,rating,completed_jobs,status,version,created_at,updated_at FROM certified_pilots WHERE id=$1`, id).
		Scan(&p.ID, &p.UserID, &p.RealName, &p.IDCard, &p.Avatar, &p.Region, &certIDs, &p.FlightHours, &p.Bio, &p.Rating, &p.CompletedJobs, &p.Status, &p.Version, &p.CreatedAt, &p.UpdatedAt)
	json.Unmarshal(certIDs, &p.CertIDs)
	p.IDCard = r.dec(p.IDCard)
	return p, err
}
func (r *pilotRepo) List(ctx context.Context) ([]domain.CertifiedPilot, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,user_id,real_name,id_card,avatar,region,cert_ids,flight_hours,bio,rating,completed_jobs,status,version,created_at,updated_at FROM certified_pilots ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list pilots: %w", err)
	}
	defer rows.Close()
	var out []domain.CertifiedPilot
	for rows.Next() {
		var p domain.CertifiedPilot
		var certIDs []byte
		if err := rows.Scan(&p.ID, &p.UserID, &p.RealName, &p.IDCard, &p.Avatar, &p.Region, &certIDs, &p.FlightHours, &p.Bio, &p.Rating, &p.CompletedJobs, &p.Status, &p.Version, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan pilot: %w", err)
		}
		json.Unmarshal(certIDs, &p.CertIDs)
		p.IDCard = r.dec(p.IDCard)
		out = append(out, p)
	}
	return out, rows.Err()
}
func (r *pilotRepo) UpdateStatus(ctx context.Context, id, status string) (domain.CertifiedPilot, error) {
	_, err := r.pool.Exec(ctx, `UPDATE certified_pilots SET status=$1,updated_at=$2 WHERE id=$3`, status, time.Now(), id)
	if err != nil {
		return domain.CertifiedPilot{}, fmt.Errorf("update pilot status: %w", err)
	}
	return r.FindByID(ctx, id)
}

// ---- Product ----

type prodRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewProductRepository() repository.ProductRepository { return &prodRepo{pool: s.Pool()} }

func (r *prodRepo) Create(ctx context.Context, p domain.DroneProduct) (domain.DroneProduct, error) {
	p.Version = 1
	p.CreatedAt = time.Now()
	p.UpdatedAt = p.CreatedAt
	images, err := json.Marshal(p.Images)
	if err != nil {
		return domain.DroneProduct{}, fmt.Errorf("marshal product images: %w", err)
	}
	_, err = r.pool.Exec(ctx,
		`INSERT INTO drone_products (id,seller_id,seller_name,prod_type,title,description,price_fen,images,brand,model,condition,views,status,version,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`,
		p.ID, p.SellerID, p.SellerName, string(p.ProdType), p.Title, p.Description, p.PriceFen, images, p.Brand, p.Model, p.Condition, p.Views, p.Status, p.Version, p.CreatedAt, p.UpdatedAt)
	return p, err
}
func (r *prodRepo) FindByID(ctx context.Context, id string) (domain.DroneProduct, error) {
	var p domain.DroneProduct
	var pt string
	var imgs []byte
	err := r.pool.QueryRow(ctx,
		`SELECT id,seller_id,seller_name,prod_type,title,description,price_fen,images,brand,model,condition,views,status,version,created_at,updated_at FROM drone_products WHERE id=$1`, id).
		Scan(&p.ID, &p.SellerID, &p.SellerName, &pt, &p.Title, &p.Description, &p.PriceFen, &imgs, &p.Brand, &p.Model, &p.Condition, &p.Views, &p.Status, &p.Version, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return domain.DroneProduct{}, fmt.Errorf("product %s not found: %w", id, err)
	}
	p.ProdType = domain.ProductType(pt)
	json.Unmarshal(imgs, &p.Images)
	return p, nil
}

func (r *prodRepo) IncrementViews(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `UPDATE drone_products SET views = views + 1 WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("increment views for %s: %w", id, err)
	}
	return nil
}

func (r *prodRepo) Update(ctx context.Context, p domain.DroneProduct) (domain.DroneProduct, error) {
	p.UpdatedAt = time.Now()
	images, err := json.Marshal(p.Images)
	if err != nil {
		return domain.DroneProduct{}, fmt.Errorf("marshal product images: %w", err)
	}
	_, err = r.pool.Exec(ctx,
		`UPDATE drone_products SET seller_id=$1,seller_name=$2,prod_type=$3,title=$4,description=$5,price_fen=$6,images=$7,brand=$8,model=$9,condition=$10,status=$11,version=version+1,updated_at=$12 WHERE id=$13`,
		p.SellerID, p.SellerName, string(p.ProdType), p.Title, p.Description, p.PriceFen, images, p.Brand, p.Model, p.Condition, p.Status, p.UpdatedAt, p.ID)
	return p, err
}

func (r *prodRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM drone_products WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("delete product %s: %w", id, err)
	}
	return nil
}

func (r *prodRepo) List(ctx context.Context, prodType string) ([]domain.DroneProduct, error) {
	var rows pgx.Rows
	var err error
	if prodType == "" {
		rows, err = r.pool.Query(ctx, `SELECT id,seller_id,seller_name,prod_type,title,description,price_fen,images,brand,model,condition,views,status,version,created_at,updated_at FROM drone_products ORDER BY created_at DESC`)
	} else {
		rows, err = r.pool.Query(ctx, `SELECT id,seller_id,seller_name,prod_type,title,description,price_fen,images,brand,model,condition,views,status,version,created_at,updated_at FROM drone_products WHERE prod_type=$1 ORDER BY created_at DESC`, prodType)
	}
	if err != nil {
		return nil, fmt.Errorf("list products: %w", err)
	}
	defer rows.Close()
	var out []domain.DroneProduct
	for rows.Next() {
		var p domain.DroneProduct
		var pt string
		var imgs []byte
		if err := rows.Scan(&p.ID, &p.SellerID, &p.SellerName, &pt, &p.Title, &p.Description, &p.PriceFen, &imgs, &p.Brand, &p.Model, &p.Condition, &p.Views, &p.Status, &p.Version, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan product: %w", err)
		}
		p.ProdType = domain.ProductType(pt)
		json.Unmarshal(imgs, &p.Images)
		out = append(out, p)
	}
	return out, rows.Err()
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
	rows, err := r.pool.Query(ctx, `SELECT id,customer_id,product_desc,fault_desc,quote_fen,status,technician,version,created_at,updated_at FROM repair_orders WHERE customer_id=$1 ORDER BY created_at DESC`, userID)
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
	rows, err := r.pool.Query(ctx, `SELECT id,customer_id,product_desc,fault_desc,quote_fen,status,technician,version,created_at,updated_at FROM repair_orders ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
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
	rows, err := r.pool.Query(ctx, `SELECT id,user_id,drone_model,drone_sn,policy_type,premium_fen,coverage_fen,start_date,end_date,insurer,status,version,created_at,updated_at FROM insurance_policies WHERE user_id=$1 ORDER BY created_at DESC`, userID)
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
	rows, err := r.pool.Query(ctx, `SELECT id,user_id,drone_model,drone_sn,policy_type,premium_fen,coverage_fen,start_date,end_date,insurer,status,version,created_at,updated_at FROM insurance_policies ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
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
