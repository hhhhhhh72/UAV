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

func (s *Store) NewCertificateRepository() repository.CertificateRepository { return &certRepo{pool: s.Pool()} }

func (r *certRepo) Create(c domain.Certificate) (domain.Certificate, error) {
	c.Version = 1; c.CreatedAt = time.Now(); c.UpdatedAt = c.CreatedAt
	_, err := r.pool.Exec(context.Background(),
		`INSERT INTO certificates (id,user_id,cert_type,cert_number,level,issue_date,expire_date,issuer_org,image_url,status,version,created_at,updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		c.ID, c.UserID, string(c.CertType), c.CertNumber, c.Level, c.IssueDate, c.ExpireDate, c.IssuerOrg, c.ImageURL, c.Status, c.Version, c.CreatedAt, c.UpdatedAt)
	return c, err
}
func (r *certRepo) FindByID(id string) (domain.Certificate, error) {
	var c domain.Certificate; var ct string
	err := r.pool.QueryRow(context.Background(),
		`SELECT id,user_id,cert_type,cert_number,level,issue_date,expire_date,issuer_org,image_url,status,version,created_at,updated_at FROM certificates WHERE id=$1`, id).
		Scan(&c.ID, &c.UserID, &ct, &c.CertNumber, &c.Level, &c.IssueDate, &c.ExpireDate, &c.IssuerOrg, &c.ImageURL, &c.Status, &c.Version, &c.CreatedAt, &c.UpdatedAt)
	c.CertType = domain.CertType(ct)
	return c, err
}
func (r *certRepo) ListByUser(userID string) ([]domain.Certificate, error) {
	rows, err := r.pool.Query(context.Background(),
		`SELECT id,user_id,cert_type,cert_number,level,issue_date,expire_date,issuer_org,image_url,status,version,created_at,updated_at FROM certificates WHERE user_id=$1 ORDER BY created_at DESC`, userID)
	if err != nil { return nil, fmt.Errorf("list certificates: %w", err) }
	defer rows.Close()
	var out []domain.Certificate
	for rows.Next() {
		var c domain.Certificate; var ct string
		rows.Scan(&c.ID, &c.UserID, &ct, &c.CertNumber, &c.Level, &c.IssueDate, &c.ExpireDate, &c.IssuerOrg, &c.ImageURL, &c.Status, &c.Version, &c.CreatedAt, &c.UpdatedAt)
		c.CertType = domain.CertType(ct)
		out = append(out, c)
	}
	return out, rows.Err()
}
func (r *certRepo) UpdateStatus(id, status string) (domain.Certificate, error) {
	_, err := r.pool.Exec(context.Background(), `UPDATE certificates SET status=$1,updated_at=$2 WHERE id=$3`, status, time.Now(), id)
	if err != nil { return domain.Certificate{}, fmt.Errorf("update certificate status: %w", err) }
	return r.FindByID(id)
}
func (r *certRepo) ListAll() ([]domain.Certificate, error) {
	rows, err := r.pool.Query(context.Background(), `SELECT id,user_id,cert_type,cert_number,level,issue_date,expire_date,issuer_org,image_url,status,version,created_at,updated_at FROM certificates ORDER BY created_at DESC`)
	if err != nil { return nil, fmt.Errorf("list all certificates: %w", err) }
	defer rows.Close()
	var out []domain.Certificate
	for rows.Next() {
		var c domain.Certificate; var ct string
		rows.Scan(&c.ID, &c.UserID, &ct, &c.CertNumber, &c.Level, &c.IssueDate, &c.ExpireDate, &c.IssuerOrg, &c.ImageURL, &c.Status, &c.Version, &c.CreatedAt, &c.UpdatedAt)
		c.CertType = domain.CertType(ct)
		out = append(out, c)
	}
	return out, rows.Err()
}

// ---- Course ----

type courseRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewCourseRepository() repository.CourseRepository { return &courseRepo{pool: s.Pool()} }

func (r *courseRepo) Create(c domain.TrainingCourse) (domain.TrainingCourse, error) {
	c.Version = 1; c.CreatedAt = time.Now(); c.UpdatedAt = c.CreatedAt
	_, err := r.pool.Exec(context.Background(),
		`INSERT INTO training_courses (id,org_id,title,cert_type,description,start_date,end_date,max_students,enrolled_count,location,price_fen,status,version,created_at,updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		c.ID, c.OrgID, c.Title, string(c.CertType), c.Description, c.StartDate, c.EndDate, c.MaxStudents, c.EnrolledCount, c.Location, c.PriceFen, c.Status, c.Version, c.CreatedAt, c.UpdatedAt)
	return c, err
}
func (r *courseRepo) List() ([]domain.TrainingCourse, error) {
	rows, err := r.pool.Query(context.Background(), `SELECT id,org_id,title,cert_type,description,start_date,end_date,max_students,enrolled_count,location,price_fen,status,version,created_at,updated_at FROM training_courses ORDER BY created_at DESC`)
	if err != nil { return nil, fmt.Errorf("list courses: %w", err) }
	defer rows.Close()
	var out []domain.TrainingCourse
	for rows.Next() {
		var c domain.TrainingCourse; var ct string
		rows.Scan(&c.ID, &c.OrgID, &c.Title, &ct, &c.Description, &c.StartDate, &c.EndDate, &c.MaxStudents, &c.EnrolledCount, &c.Location, &c.PriceFen, &c.Status, &c.Version, &c.CreatedAt, &c.UpdatedAt)
		c.CertType = domain.CertType(ct)
		out = append(out, c)
	}
	return out, rows.Err()
}

// ---- Instructor ----

type instructorRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewInstructorRepository() repository.InstructorRepository { return &instructorRepo{pool: s.Pool()} }

func (r *instructorRepo) Create(i domain.Instructor) (domain.Instructor, error) {
	i.Version = 1; i.CreatedAt = time.Now(); i.UpdatedAt = i.CreatedAt
	certTypes, _ := json.Marshal(i.CertTypes)
	_, err := r.pool.Exec(context.Background(),
		`INSERT INTO instructors (id,user_id,name,cert_types,bio,org_id,status,version,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		i.ID, i.UserID, i.Name, certTypes, i.Bio, i.OrgID, i.Status, i.Version, i.CreatedAt, i.UpdatedAt)
	return i, err
}
func (r *instructorRepo) FindByID(id string) (domain.Instructor, error) {
	var i domain.Instructor; var ct []byte
	err := r.pool.QueryRow(context.Background(), `SELECT id,user_id,name,cert_types,bio,org_id,status,version,created_at,updated_at FROM instructors WHERE id=$1`, id).
		Scan(&i.ID, &i.UserID, &i.Name, &ct, &i.Bio, &i.OrgID, &i.Status, &i.Version, &i.CreatedAt, &i.UpdatedAt)
	json.Unmarshal(ct, &i.CertTypes)
	return i, err
}
func (r *instructorRepo) List() ([]domain.Instructor, error) {
	rows, err := r.pool.Query(context.Background(), `SELECT id,user_id,name,cert_types,bio,org_id,status,version,created_at,updated_at FROM instructors ORDER BY created_at DESC`)
	if err != nil { return nil, fmt.Errorf("list instructors: %w", err) }
	defer rows.Close()
	var out []domain.Instructor
	for rows.Next() {
		var i domain.Instructor; var ct []byte
		rows.Scan(&i.ID, &i.UserID, &i.Name, &ct, &i.Bio, &i.OrgID, &i.Status, &i.Version, &i.CreatedAt, &i.UpdatedAt)
		json.Unmarshal(ct, &i.CertTypes)
		out = append(out, i)
	}
	return out, rows.Err()
}
func (r *instructorRepo) UpdateStatus(id, status string) (domain.Instructor, error) {
	_, err := r.pool.Exec(context.Background(), `UPDATE instructors SET status=$1,updated_at=$2 WHERE id=$3`, status, time.Now(), id)
	if err != nil { return domain.Instructor{}, fmt.Errorf("update instructor status: %w", err) }
	return r.FindByID(id)
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
	if r.cipher != nil && v != "" { if e, err := r.cipher.Encrypt(v); err == nil { return e } }
	return v
}
func (r *pilotRepo) dec(v string) string {
	if r.cipher != nil && v != "" { if d, err := r.cipher.Decrypt(v); err == nil { return d } }
	return v
}
func (r *pilotRepo) Create(p domain.CertifiedPilot) (domain.CertifiedPilot, error) {
	p.Version = 1; p.CreatedAt = time.Now(); p.UpdatedAt = p.CreatedAt
	p.IDCard = r.enc(p.IDCard)
	certIDs, _ := json.Marshal(p.CertIDs)
	_, err := r.pool.Exec(context.Background(),
		`INSERT INTO certified_pilots (id,user_id,real_name,id_card,cert_ids,flight_hours,rating,completed_jobs,status,version,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		p.ID, p.UserID, p.RealName, p.IDCard, certIDs, p.FlightHours, p.Rating, p.CompletedJobs, p.Status, p.Version, p.CreatedAt, p.UpdatedAt)
	p.IDCard = r.dec(p.IDCard)
	return p, err
}
func (r *pilotRepo) FindByID(id string) (domain.CertifiedPilot, error) {
	var p domain.CertifiedPilot; var certIDs []byte
	err := r.pool.QueryRow(context.Background(), `SELECT id,user_id,real_name,id_card,cert_ids,flight_hours,rating,completed_jobs,status,version,created_at,updated_at FROM certified_pilots WHERE id=$1`, id).
		Scan(&p.ID, &p.UserID, &p.RealName, &p.IDCard, &certIDs, &p.FlightHours, &p.Rating, &p.CompletedJobs, &p.Status, &p.Version, &p.CreatedAt, &p.UpdatedAt)
	json.Unmarshal(certIDs, &p.CertIDs)
	p.IDCard = r.dec(p.IDCard)
	return p, err
}
func (r *pilotRepo) List() ([]domain.CertifiedPilot, error) {
	rows, err := r.pool.Query(context.Background(), `SELECT id,user_id,real_name,id_card,cert_ids,flight_hours,rating,completed_jobs,status,version,created_at,updated_at FROM certified_pilots ORDER BY created_at DESC`)
	if err != nil { return nil, fmt.Errorf("list pilots: %w", err) }
	defer rows.Close()
	var out []domain.CertifiedPilot
	for rows.Next() {
		var p domain.CertifiedPilot; var certIDs []byte
		rows.Scan(&p.ID, &p.UserID, &p.RealName, &p.IDCard, &certIDs, &p.FlightHours, &p.Rating, &p.CompletedJobs, &p.Status, &p.Version, &p.CreatedAt, &p.UpdatedAt)
		json.Unmarshal(certIDs, &p.CertIDs)
		p.IDCard = r.dec(p.IDCard)
		out = append(out, p)
	}
	return out, rows.Err()
}
func (r *pilotRepo) UpdateStatus(id, status string) (domain.CertifiedPilot, error) {
	_, err := r.pool.Exec(context.Background(), `UPDATE certified_pilots SET status=$1,updated_at=$2 WHERE id=$3`, status, time.Now(), id)
	if err != nil { return domain.CertifiedPilot{}, fmt.Errorf("update pilot status: %w", err) }
	return r.FindByID(id)
}

// ---- Product ----

type prodRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewProductRepository() repository.ProductRepository { return &prodRepo{pool: s.Pool()} }

func (r *prodRepo) Create(p domain.DroneProduct) (domain.DroneProduct, error) {
	p.Version = 1; p.CreatedAt = time.Now(); p.UpdatedAt = p.CreatedAt
	images, _ := json.Marshal(p.Images)
	_, err := r.pool.Exec(context.Background(),
		`INSERT INTO drone_products (id,seller_id,seller_name,prod_type,title,description,price_fen,images,brand,model,condition,status,version,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		p.ID, p.SellerID, p.SellerName, string(p.ProdType), p.Title, p.Description, p.PriceFen, images, p.Brand, p.Model, p.Condition, p.Status, p.Version, p.CreatedAt, p.UpdatedAt)
	return p, err
}
func (r *prodRepo) List(prodType string) ([]domain.DroneProduct, error) {
	var rows pgx.Rows; var err error
	if prodType == "" {
		rows, err = r.pool.Query(context.Background(), `SELECT id,seller_id,seller_name,prod_type,title,description,price_fen,images,brand,model,condition,status,version,created_at,updated_at FROM drone_products ORDER BY created_at DESC`)
	} else {
		rows, err = r.pool.Query(context.Background(), `SELECT id,seller_id,seller_name,prod_type,title,description,price_fen,images,brand,model,condition,status,version,created_at,updated_at FROM drone_products WHERE prod_type=$1 ORDER BY created_at DESC`, prodType)
	}
	if err != nil { return nil, fmt.Errorf("list products: %w", err) }
	defer rows.Close()
	var out []domain.DroneProduct
	for rows.Next() {
		var p domain.DroneProduct; var pt string; var imgs []byte
		rows.Scan(&p.ID, &p.SellerID, &p.SellerName, &pt, &p.Title, &p.Description, &p.PriceFen, &imgs, &p.Brand, &p.Model, &p.Condition, &p.Status, &p.Version, &p.CreatedAt, &p.UpdatedAt)
		p.ProdType = domain.ProductType(pt)
		json.Unmarshal(imgs, &p.Images)
		out = append(out, p)
	}
	return out, rows.Err()
}

// ---- Repair ----

type repairRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewRepairRepository() repository.RepairRepository { return &repairRepo{pool: s.Pool()} }

func (r *repairRepo) Create(ro domain.RepairOrder) (domain.RepairOrder, error) {
	ro.Version = 1; ro.CreatedAt = time.Now(); ro.UpdatedAt = ro.CreatedAt
	_, err := r.pool.Exec(context.Background(),
		`INSERT INTO repair_orders (id,customer_id,product_desc,fault_desc,quote_fen,status,technician,version,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		ro.ID, ro.CustomerID, ro.ProductDesc, ro.FaultDesc, ro.QuoteFen, ro.Status, ro.Technician, ro.Version, ro.CreatedAt, ro.UpdatedAt)
	return ro, err
}
func (r *repairRepo) ListByUser(userID string) ([]domain.RepairOrder, error) {
	rows, err := r.pool.Query(context.Background(), `SELECT id,customer_id,product_desc,fault_desc,quote_fen,status,technician,version,created_at,updated_at FROM repair_orders WHERE customer_id=$1 ORDER BY created_at DESC`, userID)
	if err != nil { return nil, fmt.Errorf("list repairs: %w", err) }
	defer rows.Close()
	var out []domain.RepairOrder
	for rows.Next() {
		var ro domain.RepairOrder
		rows.Scan(&ro.ID, &ro.CustomerID, &ro.ProductDesc, &ro.FaultDesc, &ro.QuoteFen, &ro.Status, &ro.Technician, &ro.Version, &ro.CreatedAt, &ro.UpdatedAt)
		out = append(out, ro)
	}
	return out, rows.Err()
}

// ---- Policy ----

type policyRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewPolicyRepository() repository.PolicyRepository { return &policyRepo{pool: s.Pool()} }

func (r *policyRepo) Create(p domain.InsurancePolicy) (domain.InsurancePolicy, error) {
	p.Version = 1; p.CreatedAt = time.Now(); p.UpdatedAt = p.CreatedAt
	_, err := r.pool.Exec(context.Background(),
		`INSERT INTO insurance_policies (id,user_id,drone_model,drone_sn,policy_type,premium_fen,coverage_fen,start_date,end_date,insurer,status,version,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		p.ID, p.UserID, p.DroneModel, p.DroneSN, p.PolicyType, p.PremiumFen, p.CoverageFen, p.StartDate, p.EndDate, p.Insurer, p.Status, p.Version, p.CreatedAt, p.UpdatedAt)
	return p, err
}
func (r *policyRepo) ListByUser(userID string) ([]domain.InsurancePolicy, error) {
	rows, err := r.pool.Query(context.Background(), `SELECT id,user_id,drone_model,drone_sn,policy_type,premium_fen,coverage_fen,start_date,end_date,insurer,status,version,created_at,updated_at FROM insurance_policies WHERE user_id=$1 ORDER BY created_at DESC`, userID)
	if err != nil { return nil, fmt.Errorf("list policies: %w", err) }
	defer rows.Close()
	var out []domain.InsurancePolicy
	for rows.Next() {
		var p domain.InsurancePolicy
		rows.Scan(&p.ID, &p.UserID, &p.DroneModel, &p.DroneSN, &p.PolicyType, &p.PremiumFen, &p.CoverageFen, &p.StartDate, &p.EndDate, &p.Insurer, &p.Status, &p.Version, &p.CreatedAt, &p.UpdatedAt)
		out = append(out, p)
	}
	return out, rows.Err()
}
