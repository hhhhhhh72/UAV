package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)

// ---- Expert ----

type expertRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewExpertRepository() repository.ExpertRepository { return &expertRepo{pool: s.Pool()} }

func (r *expertRepo) Create(ctx context.Context, e domain.Expert) (domain.Expert, error) {
	e.CreatedAt = time.Now()
	e.UpdatedAt = e.CreatedAt
	tags, err := json.Marshal(e.Tags)
	if err != nil {
		return domain.Expert{}, fmt.Errorf("marshal expert tags: %w", err)
	}
	_, err = r.pool.Exec(ctx,
		`INSERT INTO experts (id,name,title,org,field,tags,bio,avatar_url,status,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		e.ID, e.Name, e.Title, e.Org, e.Field, tags, e.Bio, e.AvatarURL, e.Status, e.CreatedAt, e.UpdatedAt)
	return e, err
}
func (r *expertRepo) FindByID(ctx context.Context, id string) (domain.Expert, error) {
	var e domain.Expert
	var tags []byte
	err := r.pool.QueryRow(ctx,
		`SELECT id,name,title,org,field,tags,bio,avatar_url,status,created_at,updated_at FROM experts WHERE id=$1`, id).
		Scan(&e.ID, &e.Name, &e.Title, &e.Org, &e.Field, &tags, &e.Bio, &e.AvatarURL, &e.Status, &e.CreatedAt, &e.UpdatedAt)
	json.Unmarshal(tags, &e.Tags)
	return e, err
}
func (r *expertRepo) List(ctx context.Context, field string) ([]domain.Expert, error) {
	q := `SELECT id,name,title,org,field,tags,bio,avatar_url,status,created_at,updated_at FROM experts`
	args := []any{}
	if field != "" {
		q += ` WHERE field=$1`
		args = append(args, field)
	}
	q += ` ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("list experts: %w", err)
	}
	defer rows.Close()
	var out []domain.Expert
	for rows.Next() {
		var e domain.Expert
		var tags []byte
		if err := rows.Scan(&e.ID, &e.Name, &e.Title, &e.Org, &e.Field, &tags, &e.Bio, &e.AvatarURL, &e.Status, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan expert: %w", err)
		}
		json.Unmarshal(tags, &e.Tags)
		out = append(out, e)
	}
	return out, rows.Err()
}
func (r *expertRepo) Update(ctx context.Context, e domain.Expert) (domain.Expert, error) {
	e.UpdatedAt = time.Now()
	tags, err := json.Marshal(e.Tags)
	if err != nil {
		return domain.Expert{}, fmt.Errorf("marshal expert tags: %w", err)
	}
	_, err = r.pool.Exec(ctx,
		`UPDATE experts SET name=$1,title=$2,org=$3,field=$4,tags=$5,bio=$6,avatar_url=$7,status=$8,updated_at=$9 WHERE id=$10`,
		e.Name, e.Title, e.Org, e.Field, tags, e.Bio, e.AvatarURL, e.Status, e.UpdatedAt, e.ID)
	return e, err
}
func (r *expertRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM experts WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("delete expert %s: %w", id, err)
	}
	return nil
}

// ---- Case ----

type caseRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewCaseRepository() repository.CaseRepository { return &caseRepo{pool: s.Pool()} }

func (r *caseRepo) Create(ctx context.Context, c domain.CaseEntry) (domain.CaseEntry, error) {
	c.CreatedAt = time.Now()
	c.UpdatedAt = c.CreatedAt
	imgs, err := json.Marshal(c.Images)
	if err != nil {
		return domain.CaseEntry{}, fmt.Errorf("marshal case images: %w", err)
	}
	_, err = r.pool.Exec(ctx,
		`INSERT INTO case_entries (id,title,category,description,images,client_name,result,status,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		c.ID, c.Title, c.Category, c.Description, imgs, c.ClientName, c.Result, c.Status, c.CreatedAt, c.UpdatedAt)
	return c, err
}
func (r *caseRepo) FindByID(ctx context.Context, id string) (domain.CaseEntry, error) {
	var c domain.CaseEntry
	var imgs []byte
	err := r.pool.QueryRow(ctx,
		`SELECT id,title,category,description,images,client_name,result,status,created_at,updated_at FROM case_entries WHERE id=$1`, id).
		Scan(&c.ID, &c.Title, &c.Category, &c.Description, &imgs, &c.ClientName, &c.Result, &c.Status, &c.CreatedAt, &c.UpdatedAt)
	json.Unmarshal(imgs, &c.Images)
	return c, err
}
func (r *caseRepo) List(ctx context.Context, category string, offset, limit int) ([]domain.CaseEntry, int, error) {
	where := ""
	args := []any{}
	if category != "" {
		where = `WHERE category=$1`
		args = append(args, category)
	}
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM case_entries `+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count cases: %w", err)
	}
	q := fmt.Sprintf(`SELECT id,title,category,description,images,client_name,result,status,created_at,updated_at FROM case_entries %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, where, len(args)+1, len(args)+2)
	rows, err := r.pool.Query(ctx, q, append(args, limit, offset)...)
	if err != nil {
		return nil, 0, fmt.Errorf("list cases: %w", err)
	}
	defer rows.Close()
	var out []domain.CaseEntry
	for rows.Next() {
		var c domain.CaseEntry
		var imgs []byte
		if err := rows.Scan(&c.ID, &c.Title, &c.Category, &c.Description, &imgs, &c.ClientName, &c.Result, &c.Status, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan case: %w", err)
		}
		json.Unmarshal(imgs, &c.Images)
		out = append(out, c)
	}
	return out, total, rows.Err()
}
func (r *caseRepo) Update(ctx context.Context, c domain.CaseEntry) (domain.CaseEntry, error) {
	c.UpdatedAt = time.Now()
	imgs, err := json.Marshal(c.Images)
	if err != nil {
		return domain.CaseEntry{}, fmt.Errorf("marshal case images: %w", err)
	}
	_, err = r.pool.Exec(ctx,
		`UPDATE case_entries SET title=$1,category=$2,description=$3,images=$4,client_name=$5,result=$6,status=$7,updated_at=$8 WHERE id=$9`,
		c.Title, c.Category, c.Description, imgs, c.ClientName, c.Result, c.Status, c.UpdatedAt, c.ID)
	return c, err
}
func (r *caseRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM case_entries WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("delete case %s: %w", id, err)
	}
	return nil
}

// ---- Compliance ----

type complianceRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewComplianceRepository() repository.ComplianceRepository {
	return &complianceRepo{pool: s.Pool()}
}

func (r *complianceRepo) CreateDoc(ctx context.Context, d domain.ComplianceDoc) (domain.ComplianceDoc, error) {
	d.CreatedAt = time.Now()
	d.UpdatedAt = d.CreatedAt
	tags, err := json.Marshal(d.Tags)
	if err != nil {
		return domain.ComplianceDoc{}, fmt.Errorf("marshal doc tags: %w", err)
	}
	_, err = r.pool.Exec(ctx,
		`INSERT INTO compliance_docs (id,title,category,publisher,publish_date,status,summary,file_url,tags,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		d.ID, d.Title, d.Category, d.Publisher, d.PublishDate, d.Status, d.Summary, d.FileURL, tags, d.CreatedAt, d.UpdatedAt)
	return d, err
}
func (r *complianceRepo) FindDocByID(ctx context.Context, id string) (domain.ComplianceDoc, error) {
	var d domain.ComplianceDoc
	var tags []byte
	err := r.pool.QueryRow(ctx,
		`SELECT id,title,category,publisher,COALESCE(publish_date,'1970-01-01 00:00:00+00'::timestamptz),status,summary,file_url,tags,created_at,updated_at FROM compliance_docs WHERE id=$1`, id).
		Scan(&d.ID, &d.Title, &d.Category, &d.Publisher, &d.PublishDate, &d.Status, &d.Summary, &d.FileURL, &tags, &d.CreatedAt, &d.UpdatedAt)
	json.Unmarshal(tags, &d.Tags)
	return d, err
}
func (r *complianceRepo) ListDocs(ctx context.Context, category string, offset, limit int) ([]domain.ComplianceDoc, int, error) {
	where := ""
	args := []any{}
	if category != "" {
		where = `WHERE category=$1`
		args = append(args, category)
	}
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM compliance_docs `+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count docs: %w", err)
	}
	q := fmt.Sprintf(`SELECT id,title,category,publisher,COALESCE(publish_date,'1970-01-01 00:00:00+00'::timestamptz),status,summary,file_url,tags,created_at,updated_at FROM compliance_docs %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, where, len(args)+1, len(args)+2)
	rows, err := r.pool.Query(ctx, q, append(args, limit, offset)...)
	if err != nil {
		return nil, 0, fmt.Errorf("list docs: %w", err)
	}
	defer rows.Close()
	var out []domain.ComplianceDoc
	for rows.Next() {
		var d domain.ComplianceDoc
		var tags []byte
		if err := rows.Scan(&d.ID, &d.Title, &d.Category, &d.Publisher, &d.PublishDate, &d.Status, &d.Summary, &d.FileURL, &tags, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan doc: %w", err)
		}
		json.Unmarshal(tags, &d.Tags)
		out = append(out, d)
	}
	return out, total, rows.Err()
}
func (r *complianceRepo) UpdateDoc(ctx context.Context, d domain.ComplianceDoc) (domain.ComplianceDoc, error) {
	d.UpdatedAt = time.Now()
	tags, err := json.Marshal(d.Tags)
	if err != nil {
		return domain.ComplianceDoc{}, fmt.Errorf("marshal doc tags: %w", err)
	}
	_, err = r.pool.Exec(ctx,
		`UPDATE compliance_docs SET title=$1,category=$2,publisher=$3,publish_date=$4,status=$5,summary=$6,file_url=$7,tags=$8,updated_at=$9 WHERE id=$10`,
		d.Title, d.Category, d.Publisher, d.PublishDate, d.Status, d.Summary, d.FileURL, tags, d.UpdatedAt, d.ID)
	return d, err
}
func (r *complianceRepo) DeleteDoc(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM compliance_docs WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("delete doc %s: %w", id, err)
	}
	return nil
}

func (r *complianceRepo) DeleteStandard(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM standard_docs WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("delete standard %s: %w", id, err)
	}
	return nil
}

func (r *complianceRepo) FindStandardByID(ctx context.Context, id string) (domain.StandardDoc, error) {
	var s domain.StandardDoc
	err := r.pool.QueryRow(ctx, "SELECT id,title,category,standard_no,publisher,COALESCE(effective_date,'1970-01-01 00:00:00+00'::timestamptz),status,scope,summary,file_url FROM standard_docs WHERE id=$1", id).
		Scan(&s.ID, &s.Title, &s.Category, &s.StandardNo, &s.Publisher, &s.EffectiveDate, &s.Status, &s.Scope, &s.Summary, &s.FileURL)
	return s, err
}

func (r *complianceRepo) UpdateStandard(ctx context.Context, s domain.StandardDoc) (domain.StandardDoc, error) {
	s.UpdatedAt = time.Now()
	_, err := r.pool.Exec(ctx,
		`UPDATE standard_docs SET title=$1,category=$2,standard_no=$3,publisher=$4,effective_date=$5,status=$6,scope=$7,file_url=$8,updated_at=$9 WHERE id=$10`,
		s.Title, s.Category, s.StandardNo, s.Publisher, s.EffectiveDate, s.Status, s.Scope, s.FileURL, s.UpdatedAt, s.ID)
	return s, err
}

func (r *complianceRepo) CreateStandard(ctx context.Context, s domain.StandardDoc) (domain.StandardDoc, error) {
	s.CreatedAt = time.Now()
	s.UpdatedAt = s.CreatedAt
	_, err := r.pool.Exec(ctx,
		`INSERT INTO standard_docs (id,title,category,standard_no,publisher,effective_date,status,scope,file_url,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		s.ID, s.Title, s.Category, s.StandardNo, s.Publisher, s.EffectiveDate, s.Status, s.Scope, s.FileURL, s.CreatedAt, s.UpdatedAt)
	return s, err
}
func (r *complianceRepo) ListStandards(ctx context.Context, category string, offset, limit int) ([]domain.StandardDoc, int, error) {
	where := ""
	args := []any{}
	if category != "" {
		where = `WHERE category=$1`
		args = append(args, category)
	}
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM standard_docs `+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count standards: %w", err)
	}
	q := fmt.Sprintf(`SELECT id,title,category,standard_no,publisher,COALESCE(effective_date,'1970-01-01 00:00:00+00'::timestamptz),status,scope,file_url,created_at,updated_at FROM standard_docs %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, where, len(args)+1, len(args)+2)
	rows, err := r.pool.Query(ctx, q, append(args, limit, offset)...)
	if err != nil {
		return nil, 0, fmt.Errorf("list standards: %w", err)
	}
	defer rows.Close()
	var out []domain.StandardDoc
	for rows.Next() {
		var s domain.StandardDoc
		if err := rows.Scan(&s.ID, &s.Title, &s.Category, &s.StandardNo, &s.Publisher, &s.EffectiveDate, &s.Status, &s.Scope, &s.FileURL, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan standard: %w", err)
		}
		out = append(out, s)
	}
	return out, total, rows.Err()
}

// ---- IndustryReport ----

type indReportRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewIndustryReportRepository() repository.IndustryReportRepository {
	return &indReportRepo{pool: s.Pool()}
}

func (r *indReportRepo) Create(ctx context.Context, rp domain.IndustryReport) (domain.IndustryReport, error) {
	rp.CreatedAt = time.Now()
	rp.UpdatedAt = rp.CreatedAt
	_, err := r.pool.Exec(ctx,
		`INSERT INTO industry_reports (id,title,period,category,summary,content,file_url,author,status,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		rp.ID, rp.Title, rp.Period, rp.Category, rp.Summary, rp.Content, rp.FileURL, rp.Author, rp.Status, rp.CreatedAt, rp.UpdatedAt)
	return rp, err
}
func (r *indReportRepo) FindByID(ctx context.Context, id string) (domain.IndustryReport, error) {
	var rp domain.IndustryReport
	err := r.pool.QueryRow(ctx,
		`SELECT id,title,period,category,summary,content,file_url,author,status,created_at,updated_at FROM industry_reports WHERE id=$1`, id).
		Scan(&rp.ID, &rp.Title, &rp.Period, &rp.Category, &rp.Summary, &rp.Content, &rp.FileURL, &rp.Author, &rp.Status, &rp.CreatedAt, &rp.UpdatedAt)
	return rp, err
}
func (r *indReportRepo) List(ctx context.Context, offset, limit int) ([]domain.IndustryReport, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM industry_reports`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count reports: %w", err)
	}
	rows, err := r.pool.Query(ctx,
		`SELECT id,title,period,category,summary,content,file_url,author,status,created_at,updated_at FROM industry_reports ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list reports: %w", err)
	}
	defer rows.Close()
	var out []domain.IndustryReport
	for rows.Next() {
		var rp domain.IndustryReport
		if err := rows.Scan(&rp.ID, &rp.Title, &rp.Period, &rp.Category, &rp.Summary, &rp.Content, &rp.FileURL, &rp.Author, &rp.Status, &rp.CreatedAt, &rp.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan report: %w", err)
		}
		out = append(out, rp)
	}
	return out, total, rows.Err()
}
func (r *indReportRepo) Update(ctx context.Context, rp domain.IndustryReport) (domain.IndustryReport, error) {
	rp.UpdatedAt = time.Now()
	_, err := r.pool.Exec(ctx,
		`UPDATE industry_reports SET title=$1,period=$2,category=$3,summary=$4,content=$5,file_url=$6,author=$7,status=$8,updated_at=$9 WHERE id=$10`,
		rp.Title, rp.Period, rp.Category, rp.Summary, rp.Content, rp.FileURL, rp.Author, rp.Status, rp.UpdatedAt, rp.ID)
	return rp, err
}
func (r *indReportRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM industry_reports WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("delete report %s: %w", id, err)
	}
	return nil
}

// ---- Portfolio ----

type portfolioRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewPortfolioRepository() repository.PortfolioRepository {
	return &portfolioRepo{pool: s.Pool()}
}

func (r *portfolioRepo) Create(ctx context.Context, p domain.MemberPortfolio) (domain.MemberPortfolio, error) {
	p.CreatedAt = time.Now()
	p.UpdatedAt = p.CreatedAt
	products, err := json.Marshal(p.Products)
	if err != nil {
		return domain.MemberPortfolio{}, fmt.Errorf("marshal products: %w", err)
	}
	honors, err := json.Marshal(p.Honors)
	if err != nil {
		return domain.MemberPortfolio{}, fmt.Errorf("marshal honors: %w", err)
	}
	_, err = r.pool.Exec(ctx,
		`INSERT INTO member_portfolios (id,enterprise_id,name,logo_url,cover_url,description,products,honors,contact_info,status,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		p.ID, p.EnterpriseID, p.Name, p.LogoURL, p.CoverURL, p.Description, products, honors, p.ContactInfo, p.Status, p.CreatedAt, p.UpdatedAt)
	return p, err
}
func (r *portfolioRepo) FindByID(ctx context.Context, id string) (domain.MemberPortfolio, error) {
	var p domain.MemberPortfolio
	var prod, hon []byte
	err := r.pool.QueryRow(ctx,
		`SELECT id,enterprise_id,name,logo_url,cover_url,description,products,honors,contact_info,status,created_at,updated_at FROM member_portfolios WHERE id=$1`, id).
		Scan(&p.ID, &p.EnterpriseID, &p.Name, &p.LogoURL, &p.CoverURL, &p.Description, &prod, &hon, &p.ContactInfo, &p.Status, &p.CreatedAt, &p.UpdatedAt)
	json.Unmarshal(prod, &p.Products)
	json.Unmarshal(hon, &p.Honors)
	return p, err
}
func (r *portfolioRepo) ListByEnterprise(ctx context.Context, eid string) ([]domain.MemberPortfolio, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id,enterprise_id,name,logo_url,cover_url,description,products,honors,contact_info,status,created_at,updated_at FROM member_portfolios WHERE enterprise_id=$1 ORDER BY created_at DESC`, eid)
	if err != nil {
		return nil, fmt.Errorf("list portfolios: %w", err)
	}
	defer rows.Close()
	var out []domain.MemberPortfolio
	for rows.Next() {
		var p domain.MemberPortfolio
		var prod, hon []byte
		if err := rows.Scan(&p.ID, &p.EnterpriseID, &p.Name, &p.LogoURL, &p.CoverURL, &p.Description, &prod, &hon, &p.ContactInfo, &p.Status, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan portfolio: %w", err)
		}
		json.Unmarshal(prod, &p.Products)
		json.Unmarshal(hon, &p.Honors)
		out = append(out, p)
	}
	return out, rows.Err()
}
func (r *portfolioRepo) ListPublished(ctx context.Context, offset, limit int) ([]domain.MemberPortfolio, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM member_portfolios WHERE status='published'`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count published portfolios: %w", err)
	}
	rows, err := r.pool.Query(ctx,
		`SELECT id,enterprise_id,name,logo_url,cover_url,description,products,honors,contact_info,status,created_at,updated_at FROM member_portfolios WHERE status='published' ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list published portfolios: %w", err)
	}
	defer rows.Close()
	var out []domain.MemberPortfolio
	for rows.Next() {
		var p domain.MemberPortfolio
		var prod, hon []byte
		if err := rows.Scan(&p.ID, &p.EnterpriseID, &p.Name, &p.LogoURL, &p.CoverURL, &p.Description, &prod, &hon, &p.ContactInfo, &p.Status, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan portfolio: %w", err)
		}
		json.Unmarshal(prod, &p.Products)
		json.Unmarshal(hon, &p.Honors)
		out = append(out, p)
	}
	return out, total, rows.Err()
}
func (r *portfolioRepo) List(ctx context.Context, offset, limit int) ([]domain.MemberPortfolio, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM member_portfolios`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count portfolios: %w", err)
	}
	rows, err := r.pool.Query(ctx,
		`SELECT id,enterprise_id,name,logo_url,cover_url,description,products,honors,contact_info,status,created_at,updated_at FROM member_portfolios ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list portfolios: %w", err)
	}
	defer rows.Close()
	var out []domain.MemberPortfolio
	for rows.Next() {
		var p domain.MemberPortfolio
		var prod, hon []byte
		if err := rows.Scan(&p.ID, &p.EnterpriseID, &p.Name, &p.LogoURL, &p.CoverURL, &p.Description, &prod, &hon, &p.ContactInfo, &p.Status, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan portfolio: %w", err)
		}
		json.Unmarshal(prod, &p.Products)
		json.Unmarshal(hon, &p.Honors)
		out = append(out, p)
	}
	return out, total, rows.Err()
}
func (r *portfolioRepo) Update(ctx context.Context, p domain.MemberPortfolio) (domain.MemberPortfolio, error) {
	p.UpdatedAt = time.Now()
	prod, err := json.Marshal(p.Products)
	if err != nil {
		return domain.MemberPortfolio{}, fmt.Errorf("marshal products: %w", err)
	}
	hon, err := json.Marshal(p.Honors)
	if err != nil {
		return domain.MemberPortfolio{}, fmt.Errorf("marshal honors: %w", err)
	}
	_, err = r.pool.Exec(ctx,
		`UPDATE member_portfolios SET name=$1,logo_url=$2,cover_url=$3,description=$4,products=$5,honors=$6,contact_info=$7,status=$8,updated_at=$9 WHERE id=$10`,
		p.Name, p.LogoURL, p.CoverURL, p.Description, prod, hon, p.ContactInfo, p.Status, p.UpdatedAt, p.ID)
	return p, err
}

func (r *portfolioRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM member_portfolios WHERE id=$1", id)
	return err
}

// ---- Resource ----

type resourceRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewResourceRepository() repository.ResourceRepository {
	return &resourceRepo{pool: s.Pool()}
}
func (r *resourceRepo) Create(ctx context.Context, res domain.IndustryResource) (domain.IndustryResource, error) {
	res.CreatedAt = time.Now()
	res.UpdatedAt = res.CreatedAt
	if res.VisibilityLevel == "" {
		res.VisibilityLevel = "public"
	}
	_, err := r.pool.Exec(ctx,
		`INSERT INTO industry_resources (id,owner_id,name,res_type,model,specs,location,price_fen,booking_info,visibility_level,status,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		res.ID, res.OwnerID, res.Name, res.ResType, res.Model, res.Specs, res.Location, res.PriceFen, res.BookingInfo, res.VisibilityLevel, res.Status, res.CreatedAt, res.UpdatedAt)
	return res, err
}
func (r *resourceRepo) FindByID(ctx context.Context, id string) (domain.IndustryResource, error) {
	var res domain.IndustryResource
	err := r.pool.QueryRow(ctx,
		`SELECT id,owner_id,name,res_type,model,specs,location,price_fen,booking_info,visibility_level,status,created_at,updated_at FROM industry_resources WHERE id=$1`, id).
		Scan(&res.ID, &res.OwnerID, &res.Name, &res.ResType, &res.Model, &res.Specs, &res.Location, &res.PriceFen, &res.BookingInfo, &res.VisibilityLevel, &res.Status, &res.CreatedAt, &res.UpdatedAt)
	return res, err
}
func (r *resourceRepo) List(ctx context.Context, resType string, offset, limit int) ([]domain.IndustryResource, int, error) {
	where := ""
	args := []any{}
	if resType != "" {
		where = `WHERE res_type=$1`
		args = append(args, resType)
	}
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM industry_resources `+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count resources: %w", err)
	}
	q := fmt.Sprintf(`SELECT id,owner_id,name,res_type,model,specs,location,price_fen,booking_info,visibility_level,status,created_at,updated_at FROM industry_resources %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, where, len(args)+1, len(args)+2)
	rows, err := r.pool.Query(ctx, q, append(args, limit, offset)...)
	if err != nil {
		return nil, 0, fmt.Errorf("list resources: %w", err)
	}
	defer rows.Close()
	var out []domain.IndustryResource
	for rows.Next() {
		var res domain.IndustryResource
		if err := rows.Scan(&res.ID, &res.OwnerID, &res.Name, &res.ResType, &res.Model, &res.Specs, &res.Location, &res.PriceFen, &res.BookingInfo, &res.VisibilityLevel, &res.Status, &res.CreatedAt, &res.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan resource: %w", err)
		}
		out = append(out, res)
	}
	return out, total, rows.Err()
}
func (r *resourceRepo) Update(ctx context.Context, res domain.IndustryResource) (domain.IndustryResource, error) {
	res.UpdatedAt = time.Now()
	_, err := r.pool.Exec(ctx,
		`UPDATE industry_resources SET name=$1,res_type=$2,model=$3,specs=$4,location=$5,price_fen=$6,booking_info=$7,visibility_level=$8,status=$9,updated_at=$10 WHERE id=$11`,
		res.Name, res.ResType, res.Model, res.Specs, res.Location, res.PriceFen, res.BookingInfo, res.VisibilityLevel, res.Status, res.UpdatedAt, res.ID)
	return res, err
}

func (r *resourceRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM industry_resources WHERE id=$1`, id)
	return err
}

// ---- Resource bookings (C11) ----

func (r *resourceRepo) CreateBooking(ctx context.Context, b domain.IndustryResourceBooking) (domain.IndustryResourceBooking, error) {
	b.CreatedAt = time.Now()
	b.UpdatedAt = b.CreatedAt
	_, err := r.pool.Exec(ctx,
		`INSERT INTO industry_resource_bookings (id,resource_id,user_id,booking_date,purpose,contact_name,contact_phone,status,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		b.ID, b.ResourceID, b.UserID, b.BookingDate, b.Purpose, b.ContactName, b.ContactPhone, b.Status, b.CreatedAt, b.UpdatedAt)
	if err != nil {
		return domain.IndustryResourceBooking{}, fmt.Errorf("insert resource booking: %w", err)
	}
	return b, nil
}

func scanResourceBookings(rows pgx.Rows) ([]domain.IndustryResourceBooking, error) {
	defer rows.Close()
	out := make([]domain.IndustryResourceBooking, 0)
	for rows.Next() {
		var b domain.IndustryResourceBooking
		if err := rows.Scan(&b.ID, &b.ResourceID, &b.UserID, &b.BookingDate, &b.Purpose, &b.ContactName, &b.ContactPhone, &b.Status, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan resource booking: %w", err)
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (r *resourceRepo) ListBookingsByResource(ctx context.Context, resourceID string) ([]domain.IndustryResourceBooking, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id,resource_id,user_id,booking_date,purpose,contact_name,contact_phone,status,created_at,updated_at FROM industry_resource_bookings WHERE resource_id=$1 ORDER BY created_at DESC`, resourceID)
	if err != nil {
		return nil, fmt.Errorf("list resource bookings: %w", err)
	}
	return scanResourceBookings(rows)
}

func (r *resourceRepo) ListBookingsByUser(ctx context.Context, userID string) ([]domain.IndustryResourceBooking, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id,resource_id,user_id,booking_date,purpose,contact_name,contact_phone,status,created_at,updated_at FROM industry_resource_bookings WHERE user_id=$1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("list user bookings: %w", err)
	}
	return scanResourceBookings(rows)
}

// ---- Application (service_applications) ----

type appRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewApplicationRepository() repository.ApplicationRepository {
	return &appRepo{pool: s.Pool()}
}

func (r *appRepo) Create(ctx context.Context, a domain.Application) (domain.Application, error) {
	a.CreatedAt = time.Now()
	data, err := json.Marshal(a.FormData)
	if err != nil {
		return domain.Application{}, fmt.Errorf("marshal application form: %w", err)
	}
	_, err = r.pool.Exec(ctx,
		`INSERT INTO service_applications (id,user_id,service_id,service_name,order_no,status,apply_time,form_data,created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		a.ID, a.UserID, a.ServiceID, a.ServiceName, a.OrderNo, a.Status, a.ApplyTime, data, a.CreatedAt)
	if err != nil {
		return domain.Application{}, fmt.Errorf("insert application: %w", err)
	}
	return a, nil
}
func (r *appRepo) FindByID(ctx context.Context, id string) (domain.Application, error) {
	var a domain.Application
	var data []byte
	err := r.pool.QueryRow(ctx,
		`SELECT id,user_id,service_id,service_name,order_no,status,apply_time,form_data,created_at FROM service_applications WHERE id=$1`, id).
		Scan(&a.ID, &a.UserID, &a.ServiceID, &a.ServiceName, &a.OrderNo, &a.Status, &a.ApplyTime, &data, &a.CreatedAt)
	if err != nil {
		return domain.Application{}, fmt.Errorf("find application %s: %w", id, err)
	}
	json.Unmarshal(data, &a.FormData)
	return a, nil
}
func (r *appRepo) ListByUser(ctx context.Context, userID string, offset, limit int) ([]domain.Application, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM service_applications WHERE user_id=$1`, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count applications: %w", err)
	}
	rows, err := r.pool.Query(ctx,
		`SELECT id,user_id,service_id,service_name,order_no,status,apply_time,form_data,created_at FROM service_applications WHERE user_id=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset)
	if err != nil {
		return nil, total, fmt.Errorf("list applications: %w", err)
	}
	defer rows.Close()
	var out []domain.Application
	for rows.Next() {
		var a domain.Application
		var data []byte
		if err := rows.Scan(&a.ID, &a.UserID, &a.ServiceID, &a.ServiceName, &a.OrderNo, &a.Status, &a.ApplyTime, &data, &a.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan application: %w", err)
		}
		json.Unmarshal(data, &a.FormData)
		out = append(out, a)
	}
	return out, total, rows.Err()
}
func (r *appRepo) ListAll(ctx context.Context, offset, limit int) ([]domain.Application, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM service_applications`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count applications: %w", err)
	}
	rows, err := r.pool.Query(ctx,
		`SELECT id,user_id,service_id,service_name,order_no,status,apply_time,form_data,created_at FROM service_applications ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
		limit, offset)
	if err != nil {
		return nil, total, fmt.Errorf("list all applications: %w", err)
	}
	defer rows.Close()
	var out []domain.Application
	for rows.Next() {
		var a domain.Application
		var data []byte
		if err := rows.Scan(&a.ID, &a.UserID, &a.ServiceID, &a.ServiceName, &a.OrderNo, &a.Status, &a.ApplyTime, &data, &a.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan application: %w", err)
		}
		json.Unmarshal(data, &a.FormData)
		out = append(out, a)
	}
	return out, total, rows.Err()
}
