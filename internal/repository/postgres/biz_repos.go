package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)

// ---- Expert ----

type expertRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewExpertRepository() repository.ExpertRepository { return &expertRepo{pool: s.Pool()} }

func (r *expertRepo) Create(e domain.Expert) (domain.Expert, error) {
	e.CreatedAt = time.Now(); e.UpdatedAt = e.CreatedAt
	tags, _ := json.Marshal(e.Tags)
	_, err := r.pool.Exec(context.Background(),
		`INSERT INTO experts (id,name,title,org,field,tags,bio,avatar_url,status,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		e.ID, e.Name, e.Title, e.Org, e.Field, tags, e.Bio, e.AvatarURL, e.Status, e.CreatedAt, e.UpdatedAt)
	return e, err
}
func (r *expertRepo) FindByID(id string) (domain.Expert, error) {
	var e domain.Expert; var tags []byte
	err := r.pool.QueryRow(context.Background(),
		`SELECT id,name,title,org,field,tags,bio,avatar_url,status,created_at,updated_at FROM experts WHERE id=$1`, id).
		Scan(&e.ID, &e.Name, &e.Title, &e.Org, &e.Field, &tags, &e.Bio, &e.AvatarURL, &e.Status, &e.CreatedAt, &e.UpdatedAt)
	json.Unmarshal(tags, &e.Tags)
	return e, err
}
func (r *expertRepo) List(field string) ([]domain.Expert, error) {
	q := `SELECT id,name,title,org,field,tags,bio,avatar_url,status,created_at,updated_at FROM experts`
	args := []any{}
	if field != "" { q += ` WHERE field=$1`; args = append(args, field) }
	q += ` ORDER BY created_at DESC`
	rows, err := r.pool.Query(context.Background(), q, args...)
	if err != nil { return nil, fmt.Errorf("list experts: %w", err) }
	defer rows.Close()
	var out []domain.Expert
	for rows.Next() {
		var e domain.Expert; var tags []byte
		rows.Scan(&e.ID, &e.Name, &e.Title, &e.Org, &e.Field, &tags, &e.Bio, &e.AvatarURL, &e.Status, &e.CreatedAt, &e.UpdatedAt)
		json.Unmarshal(tags, &e.Tags)
		out = append(out, e)
	}
	return out, rows.Err()
}
func (r *expertRepo) Update(e domain.Expert) (domain.Expert, error) {
	e.UpdatedAt = time.Now()
	tags, _ := json.Marshal(e.Tags)
	_, err := r.pool.Exec(context.Background(),
		`UPDATE experts SET name=$1,title=$2,org=$3,field=$4,tags=$5,bio=$6,avatar_url=$7,status=$8,updated_at=$9 WHERE id=$10`,
		e.Name, e.Title, e.Org, e.Field, tags, e.Bio, e.AvatarURL, e.Status, e.UpdatedAt, e.ID)
	return e, err
}
func (r *expertRepo) Delete(id string) error {
	_, err := r.pool.Exec(context.Background(), `DELETE FROM experts WHERE id=$1`, id)
	return err
}

// ---- Case ----

type caseRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewCaseRepository() repository.CaseRepository { return &caseRepo{pool: s.Pool()} }

func (r *caseRepo) Create(c domain.CaseEntry) (domain.CaseEntry, error) {
	c.CreatedAt = time.Now(); c.UpdatedAt = c.CreatedAt
	imgs, _ := json.Marshal(c.Images)
	_, err := r.pool.Exec(context.Background(),
		`INSERT INTO case_entries (id,title,category,description,images,client_name,result,status,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		c.ID, c.Title, c.Category, c.Description, imgs, c.ClientName, c.Result, c.Status, c.CreatedAt, c.UpdatedAt)
	return c, err
}
func (r *caseRepo) FindByID(id string) (domain.CaseEntry, error) {
	var c domain.CaseEntry; var imgs []byte
	err := r.pool.QueryRow(context.Background(),
		`SELECT id,title,category,description,images,client_name,result,status,created_at,updated_at FROM case_entries WHERE id=$1`, id).
		Scan(&c.ID, &c.Title, &c.Category, &c.Description, &imgs, &c.ClientName, &c.Result, &c.Status, &c.CreatedAt, &c.UpdatedAt)
	json.Unmarshal(imgs, &c.Images)
	return c, err
}
func (r *caseRepo) List(category string, offset, limit int) ([]domain.CaseEntry, int, error) {
	where := ""; args := []any{}
	if category != "" { where = `WHERE category=$1`; args = append(args, category) }
	var total int
	r.pool.QueryRow(context.Background(), `SELECT COUNT(*) FROM case_entries `+where, args...).Scan(&total)
	q := fmt.Sprintf(`SELECT id,title,category,description,images,client_name,result,status,created_at,updated_at FROM case_entries %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, where, len(args)+1, len(args)+2)
	rows, err := r.pool.Query(context.Background(), q, append(args, limit, offset)...)
	if err != nil { return nil, 0, fmt.Errorf("list cases: %w", err) }
	defer rows.Close()
	var out []domain.CaseEntry
	for rows.Next() {
		var c domain.CaseEntry; var imgs []byte
		rows.Scan(&c.ID, &c.Title, &c.Category, &c.Description, &imgs, &c.ClientName, &c.Result, &c.Status, &c.CreatedAt, &c.UpdatedAt)
		json.Unmarshal(imgs, &c.Images)
		out = append(out, c)
	}
	return out, total, rows.Err()
}
func (r *caseRepo) Update(c domain.CaseEntry) (domain.CaseEntry, error) {
	c.UpdatedAt = time.Now()
	imgs, _ := json.Marshal(c.Images)
	_, err := r.pool.Exec(context.Background(),
		`UPDATE case_entries SET title=$1,category=$2,description=$3,images=$4,client_name=$5,result=$6,status=$7,updated_at=$8 WHERE id=$9`,
		c.Title, c.Category, c.Description, imgs, c.ClientName, c.Result, c.Status, c.UpdatedAt, c.ID)
	return c, err
}
func (r *caseRepo) Delete(id string) error {
	_, err := r.pool.Exec(context.Background(), `DELETE FROM case_entries WHERE id=$1`, id)
	return err
}

// ---- Compliance ----

type complianceRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewComplianceRepository() repository.ComplianceRepository { return &complianceRepo{pool: s.Pool()} }

func (r *complianceRepo) CreateDoc(d domain.ComplianceDoc) (domain.ComplianceDoc, error) {
	d.CreatedAt = time.Now(); d.UpdatedAt = d.CreatedAt
	tags, _ := json.Marshal(d.Tags)
	_, err := r.pool.Exec(context.Background(),
		`INSERT INTO compliance_docs (id,title,category,content,summary,source,tags,status,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		d.ID, d.Title, d.Category, d.Content, d.Summary, d.Source, tags, d.Status, d.CreatedAt, d.UpdatedAt)
	return d, err
}
func (r *complianceRepo) FindDocByID(id string) (domain.ComplianceDoc, error) {
	var d domain.ComplianceDoc; var tags []byte
	err := r.pool.QueryRow(context.Background(),
		`SELECT id,title,category,content,summary,source,tags,status,created_at,updated_at FROM compliance_docs WHERE id=$1`, id).
		Scan(&d.ID, &d.Title, &d.Category, &d.Content, &d.Summary, &d.Source, &tags, &d.Status, &d.CreatedAt, &d.UpdatedAt)
	json.Unmarshal(tags, &d.Tags)
	return d, err
}
func (r *complianceRepo) ListDocs(category string, offset, limit int) ([]domain.ComplianceDoc, int, error) {
	where := ""; args := []any{}
	if category != "" { where = `WHERE category=$1`; args = append(args, category) }
	var total int
	r.pool.QueryRow(context.Background(), `SELECT COUNT(*) FROM compliance_docs `+where, args...).Scan(&total)
	q := fmt.Sprintf(`SELECT id,title,category,content,summary,source,tags,status,created_at,updated_at FROM compliance_docs %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, where, len(args)+1, len(args)+2)
	rows, err := r.pool.Query(context.Background(), q, append(args, limit, offset)...)
	if err != nil { return nil, 0, fmt.Errorf("list docs: %w", err) }
	defer rows.Close()
	var out []domain.ComplianceDoc
	for rows.Next() {
		var d domain.ComplianceDoc; var tags []byte
		rows.Scan(&d.ID, &d.Title, &d.Category, &d.Content, &d.Summary, &d.Source, &tags, &d.Status, &d.CreatedAt, &d.UpdatedAt)
		json.Unmarshal(tags, &d.Tags)
		out = append(out, d)
	}
	return out, total, rows.Err()
}
func (r *complianceRepo) UpdateDoc(d domain.ComplianceDoc) (domain.ComplianceDoc, error) {
	d.UpdatedAt = time.Now()
	tags, _ := json.Marshal(d.Tags)
	_, err := r.pool.Exec(context.Background(),
		`UPDATE compliance_docs SET title=$1,category=$2,content=$3,summary=$4,source=$5,tags=$6,status=$7,updated_at=$8 WHERE id=$9`,
		d.Title, d.Category, d.Content, d.Summary, d.Source, tags, d.Status, d.UpdatedAt, d.ID)
	return d, err
}
func (r *complianceRepo) DeleteDoc(id string) error {
	_, err := r.pool.Exec(context.Background(), `DELETE FROM compliance_docs WHERE id=$1`, id)
	return err
}
func (r *complianceRepo) CreateStandard(s domain.StandardDoc) (domain.StandardDoc, error) {
	s.CreatedAt = time.Now(); s.UpdatedAt = s.CreatedAt
	_, err := r.pool.Exec(context.Background(),
		`INSERT INTO standard_docs (id,title,std_number,category,version,issue_date,publisher,content,file_url,status,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		s.ID, s.Title, s.StdNumber, s.Category, s.Version, s.IssueDate, s.Publisher, s.Content, s.FileURL, s.Status, s.CreatedAt, s.UpdatedAt)
	return s, err
}
func (r *complianceRepo) ListStandards(category string, offset, limit int) ([]domain.StandardDoc, int, error) {
	where := ""; args := []any{}
	if category != "" { where = `WHERE category=$1`; args = append(args, category) }
	var total int
	r.pool.QueryRow(context.Background(), `SELECT COUNT(*) FROM standard_docs `+where, args...).Scan(&total)
	q := fmt.Sprintf(`SELECT id,title,std_number,category,version,issue_date,publisher,content,file_url,status,created_at,updated_at FROM standard_docs %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, where, len(args)+1, len(args)+2)
	rows, err := r.pool.Query(context.Background(), q, append(args, limit, offset)...)
	if err != nil { return nil, 0, fmt.Errorf("list standards: %w", err) }
	defer rows.Close()
	var out []domain.StandardDoc
	for rows.Next() {
		var s domain.StandardDoc
		rows.Scan(&s.ID, &s.Title, &s.StdNumber, &s.Category, &s.Version, &s.IssueDate, &s.Publisher, &s.Content, &s.FileURL, &s.Status, &s.CreatedAt, &s.UpdatedAt)
		out = append(out, s)
	}
	return out, total, rows.Err()
}

// ---- IndustryReport ----

type indReportRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewIndustryReportRepository() repository.IndustryReportRepository {
	return &indReportRepo{pool: s.Pool()}
}

func (r *indReportRepo) Create(rp domain.IndustryReport) (domain.IndustryReport, error) {
	rp.CreatedAt = time.Now(); rp.UpdatedAt = rp.CreatedAt
	_, err := r.pool.Exec(context.Background(),
		`INSERT INTO industry_reports (id,title,period,category,summary,content,file_url,author,status,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		rp.ID, rp.Title, rp.Period, rp.Category, rp.Summary, rp.Content, rp.FileURL, rp.Author, rp.Status, rp.CreatedAt, rp.UpdatedAt)
	return rp, err
}
func (r *indReportRepo) FindByID(id string) (domain.IndustryReport, error) {
	var rp domain.IndustryReport
	err := r.pool.QueryRow(context.Background(),
		`SELECT id,title,period,category,summary,content,file_url,author,status,created_at,updated_at FROM industry_reports WHERE id=$1`, id).
		Scan(&rp.ID, &rp.Title, &rp.Period, &rp.Category, &rp.Summary, &rp.Content, &rp.FileURL, &rp.Author, &rp.Status, &rp.CreatedAt, &rp.UpdatedAt)
	return rp, err
}
func (r *indReportRepo) List(offset, limit int) ([]domain.IndustryReport, int, error) {
	var total int
	r.pool.QueryRow(context.Background(), `SELECT COUNT(*) FROM industry_reports`).Scan(&total)
	rows, err := r.pool.Query(context.Background(),
		`SELECT id,title,period,category,summary,content,file_url,author,status,created_at,updated_at FROM industry_reports ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil { return nil, 0, fmt.Errorf("list reports: %w", err) }
	defer rows.Close()
	var out []domain.IndustryReport
	for rows.Next() {
		var rp domain.IndustryReport
		rows.Scan(&rp.ID, &rp.Title, &rp.Period, &rp.Category, &rp.Summary, &rp.Content, &rp.FileURL, &rp.Author, &rp.Status, &rp.CreatedAt, &rp.UpdatedAt)
		out = append(out, rp)
	}
	return out, total, rows.Err()
}
func (r *indReportRepo) Update(rp domain.IndustryReport) (domain.IndustryReport, error) {
	rp.UpdatedAt = time.Now()
	_, err := r.pool.Exec(context.Background(),
		`UPDATE industry_reports SET title=$1,period=$2,category=$3,summary=$4,content=$5,file_url=$6,author=$7,status=$8,updated_at=$9 WHERE id=$10`,
		rp.Title, rp.Period, rp.Category, rp.Summary, rp.Content, rp.FileURL, rp.Author, rp.Status, rp.UpdatedAt, rp.ID)
	return rp, err
}
func (r *indReportRepo) Delete(id string) error {
	_, err := r.pool.Exec(context.Background(), `DELETE FROM industry_reports WHERE id=$1`, id)
	return err
}

// ---- Portfolio ----

type portfolioRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewPortfolioRepository() repository.PortfolioRepository { return &portfolioRepo{pool: s.Pool()} }

func (r *portfolioRepo) Create(p domain.MemberPortfolio) (domain.MemberPortfolio, error) {
	p.CreatedAt = time.Now(); p.UpdatedAt = p.CreatedAt
	products, _ := json.Marshal(p.Products); honors, _ := json.Marshal(p.Honors)
	_, err := r.pool.Exec(context.Background(),
		`INSERT INTO member_portfolios (id,enterprise_id,name,logo_url,cover_url,description,products,honors,contact_info,status,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		p.ID, p.EnterpriseID, p.Name, p.LogoURL, p.CoverURL, p.Description, products, honors, p.ContactInfo, p.Status, p.CreatedAt, p.UpdatedAt)
	return p, err
}
func (r *portfolioRepo) FindByID(id string) (domain.MemberPortfolio, error) {
	var p domain.MemberPortfolio; var prod, hon []byte
	err := r.pool.QueryRow(context.Background(),
		`SELECT id,enterprise_id,name,logo_url,cover_url,description,products,honors,contact_info,status,created_at,updated_at FROM member_portfolios WHERE id=$1`, id).
		Scan(&p.ID, &p.EnterpriseID, &p.Name, &p.LogoURL, &p.CoverURL, &p.Description, &prod, &hon, &p.ContactInfo, &p.Status, &p.CreatedAt, &p.UpdatedAt)
	json.Unmarshal(prod, &p.Products); json.Unmarshal(hon, &p.Honors)
	return p, err
}
func (r *portfolioRepo) ListByEnterprise(eid string) ([]domain.MemberPortfolio, error) {
	rows, err := r.pool.Query(context.Background(),
		`SELECT id,enterprise_id,name,logo_url,cover_url,description,products,honors,contact_info,status,created_at,updated_at FROM member_portfolios WHERE enterprise_id=$1 ORDER BY created_at DESC`, eid)
	if err != nil { return nil, fmt.Errorf("list portfolios: %w", err) }
	defer rows.Close()
	var out []domain.MemberPortfolio
	for rows.Next() {
		var p domain.MemberPortfolio; var prod, hon []byte
		rows.Scan(&p.ID, &p.EnterpriseID, &p.Name, &p.LogoURL, &p.CoverURL, &p.Description, &prod, &hon, &p.ContactInfo, &p.Status, &p.CreatedAt, &p.UpdatedAt)
		json.Unmarshal(prod, &p.Products); json.Unmarshal(hon, &p.Honors)
		out = append(out, p)
	}
	return out, rows.Err()
}
func (r *portfolioRepo) ListPublished(offset, limit int) ([]domain.MemberPortfolio, int, error) {
	var total int
	r.pool.QueryRow(context.Background(), `SELECT COUNT(*) FROM member_portfolios WHERE status='published'`).Scan(&total)
	rows, err := r.pool.Query(context.Background(),
		`SELECT id,enterprise_id,name,logo_url,cover_url,description,products,honors,contact_info,status,created_at,updated_at FROM member_portfolios WHERE status='published' ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil { return nil, 0, fmt.Errorf("list published portfolios: %w", err) }
	defer rows.Close()
	var out []domain.MemberPortfolio
	for rows.Next() {
		var p domain.MemberPortfolio; var prod, hon []byte
		rows.Scan(&p.ID, &p.EnterpriseID, &p.Name, &p.LogoURL, &p.CoverURL, &p.Description, &prod, &hon, &p.ContactInfo, &p.Status, &p.CreatedAt, &p.UpdatedAt)
		json.Unmarshal(prod, &p.Products); json.Unmarshal(hon, &p.Honors)
		out = append(out, p)
	}
	return out, total, rows.Err()
}
func (r *portfolioRepo) Update(p domain.MemberPortfolio) (domain.MemberPortfolio, error) {
	p.UpdatedAt = time.Now()
	prod, _ := json.Marshal(p.Products); hon, _ := json.Marshal(p.Honors)
	_, err := r.pool.Exec(context.Background(),
		`UPDATE member_portfolios SET name=$1,logo_url=$2,cover_url=$3,description=$4,products=$5,honors=$6,contact_info=$7,status=$8,updated_at=$9 WHERE id=$10`,
		p.Name, p.LogoURL, p.CoverURL, p.Description, prod, hon, p.ContactInfo, p.Status, p.UpdatedAt, p.ID)
	return p, err
}

// ---- Resource ----

type resourceRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewResourceRepository() repository.ResourceRepository { return &resourceRepo{pool: s.Pool()} }

func (r *resourceRepo) Create(res domain.IndustryResource) (domain.IndustryResource, error) {
	res.CreatedAt = time.Now(); res.UpdatedAt = res.CreatedAt
	_, err := r.pool.Exec(context.Background(),
		`INSERT INTO industry_resources (id,owner_id,name,res_type,model,specs,location,price_fen,booking_info,status,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		res.ID, res.OwnerID, res.Name, res.ResType, res.Model, res.Specs, res.Location, res.PriceFen, res.BookingInfo, res.Status, res.CreatedAt, res.UpdatedAt)
	return res, err
}
func (r *resourceRepo) FindByID(id string) (domain.IndustryResource, error) {
	var res domain.IndustryResource
	err := r.pool.QueryRow(context.Background(),
		`SELECT id,owner_id,name,res_type,model,specs,location,price_fen,booking_info,status,created_at,updated_at FROM industry_resources WHERE id=$1`, id).
		Scan(&res.ID, &res.OwnerID, &res.Name, &res.ResType, &res.Model, &res.Specs, &res.Location, &res.PriceFen, &res.BookingInfo, &res.Status, &res.CreatedAt, &res.UpdatedAt)
	return res, err
}
func (r *resourceRepo) List(resType string, offset, limit int) ([]domain.IndustryResource, int, error) {
	where := ""; args := []any{}
	if resType != "" { where = `WHERE res_type=$1`; args = append(args, resType) }
	var total int
	r.pool.QueryRow(context.Background(), `SELECT COUNT(*) FROM industry_resources `+where, args...).Scan(&total)
	q := fmt.Sprintf(`SELECT id,owner_id,name,res_type,model,specs,location,price_fen,booking_info,status,created_at,updated_at FROM industry_resources %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, where, len(args)+1, len(args)+2)
	rows, err := r.pool.Query(context.Background(), q, append(args, limit, offset)...)
	if err != nil { return nil, 0, fmt.Errorf("list resources: %w", err) }
	defer rows.Close()
	var out []domain.IndustryResource
	for rows.Next() {
		var res domain.IndustryResource
		rows.Scan(&res.ID, &res.OwnerID, &res.Name, &res.ResType, &res.Model, &res.Specs, &res.Location, &res.PriceFen, &res.BookingInfo, &res.Status, &res.CreatedAt, &res.UpdatedAt)
		out = append(out, res)
	}
	return out, total, rows.Err()
}
func (r *resourceRepo) Update(res domain.IndustryResource) (domain.IndustryResource, error) {
	res.UpdatedAt = time.Now()
	_, err := r.pool.Exec(context.Background(),
		`UPDATE industry_resources SET name=$1,res_type=$2,model=$3,specs=$4,location=$5,price_fen=$6,booking_info=$7,status=$8,updated_at=$9 WHERE id=$10`,
		res.Name, res.ResType, res.Model, res.Specs, res.Location, res.PriceFen, res.BookingInfo, res.Status, res.UpdatedAt, res.ID)
	return res, err
}
