package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"drone-platform/internal/crypto"
	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)

// jsonbSlice 保证 JSONB 数组列写入非 NULL：pgx v5 将 nil slice 编码为 SQL NULL，
// 会违反 NOT NULL DEFAULT '[]' 约束（此前 training_courses.tags 等列踩过 23502）。
func jsonbSlice[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}

// ---- Connection ----

type Store struct {
	pool   *pgxpool.Pool
	cipher *crypto.Cipher
}

func NewStore(ctx context.Context, databaseURL string, cipher *crypto.Cipher) (*Store, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse db url: %w", err)
	}
	cfg.MaxConns = 50
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("connect db: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}
	return &Store{pool: pool, cipher: cipher}, nil
}

func (s *Store) Pool() *pgxpool.Pool { return s.pool }
func (s *Store) Close()              { s.pool.Close() }

// ---- Demand Repository ----

type demandRepo struct {
	pool   *pgxpool.Pool
	cipher *crypto.Cipher
}

func (s *Store) NewDemandRepository() repository.DemandRepository {
	return &demandRepo{pool: s.pool, cipher: s.cipher}
}

func (r *demandRepo) Create(d domain.Demand) (domain.Demand, error) {
	images, err := json.Marshal(d.Images)
	if err != nil {
		return domain.Demand{}, fmt.Errorf("marshal images: %w", err)
	}
	bizFields, err := json.Marshal(d.BizFields)
	if err != nil {
		return domain.Demand{}, fmt.Errorf("marshal bizFields: %w", err)
	}
	if r.cipher != nil && d.Contact != "" {
		enc, err := r.cipher.Encrypt(d.Contact)
		if err != nil {
			return domain.Demand{}, fmt.Errorf("encrypt contact: %w", err)
		}
		d.Contact = enc
	}
	now := time.Now()
	d.Version = 1
	d.CreatedAt = now
	d.UpdatedAt = now
	_, err = r.pool.Exec(context.Background(), `
		INSERT INTO demands (id, publisher_id, publisher_name, contact, district, city_code,
			biz_type, title, description, images, latitude, longitude, budget_fen, offline_amount_fen, biz_fields,
			status, version, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)`,
		d.ID, d.PublisherID, d.PublisherName, d.Contact, d.District, d.CityCode,
		string(d.BizType), d.Title, d.Description, images, d.Latitude, d.Longitude,
		d.BudgetFen, d.OfflineAmountFen, bizFields, string(d.Status), d.Version, d.CreatedAt, d.UpdatedAt)
	if err != nil {
		return domain.Demand{}, fmt.Errorf("create demand: %w", err)
	}
	return d, nil
}

func (r *demandRepo) FindByID(id string) (domain.Demand, error) {
	q := `SELECT id, publisher_id, publisher_name, contact, district, city_code,
		biz_type, title, description, images, latitude, longitude, budget_fen, offline_amount_fen, biz_fields,
		status, version, created_at, updated_at
		FROM demands WHERE id = $1`
	demands, err := scanDemands(r.pool, r.cipher, q, []any{id})
	if err != nil {
		return domain.Demand{}, err
	}
	if len(demands) == 0 {
		return domain.Demand{}, fmt.Errorf("demand %s not found", id)
	}
	return demands[0], nil
}

func (r *demandRepo) Update(d domain.Demand) (domain.Demand, error) {
	images, err := json.Marshal(d.Images)
	if err != nil {
		return domain.Demand{}, fmt.Errorf("marshal images: %w", err)
	}
	bizFields, err := json.Marshal(d.BizFields)
	if err != nil {
		return domain.Demand{}, fmt.Errorf("marshal bizFields: %w", err)
	}
	encContact := d.Contact
	if r.cipher != nil && d.Contact != "" {
		enc, err := r.cipher.Encrypt(d.Contact)
		if err != nil {
			return domain.Demand{}, fmt.Errorf("encrypt contact: %w", err)
		}
		encContact = enc
	}
	d.UpdatedAt = time.Now()
	d.Version++
	_, err = r.pool.Exec(context.Background(),
		`UPDATE demands SET publisher_name=$1, contact=$2, district=$3, city_code=$4,
		biz_type=$5, title=$6, description=$7, images=$8, latitude=$9, longitude=$10,
		budget_fen=$11, offline_amount_fen=$12, biz_fields=$13, status=$14, version=$15, updated_at=$16 WHERE id=$17`,
		d.PublisherName, encContact, d.District, d.CityCode,
		string(d.BizType), d.Title, d.Description, images, d.Latitude, d.Longitude,
		d.BudgetFen, d.OfflineAmountFen, bizFields, string(d.Status), d.Version, d.UpdatedAt, d.ID)
	return d, err
}

func (r *demandRepo) List(f repository.DemandFilter) ([]domain.Demand, error) {
	q := `SELECT id, publisher_id, publisher_name, contact, district, city_code,
		biz_type, title, description, images, latitude, longitude, budget_fen, offline_amount_fen, biz_fields,
		status, version, created_at, updated_at
		FROM demands WHERE status = 'published'`
	args := []any{}
	argIdx := 1

	if f.District != "" {
		q += fmt.Sprintf(" AND district = $%d", argIdx)
		args = append(args, f.District)
		argIdx++
	}
	if f.BizType != "" {
		q += fmt.Sprintf(" AND biz_type = $%d", argIdx)
		args = append(args, f.BizType)
		argIdx++
	}
	q += " ORDER BY created_at DESC"
	return scanDemands(r.pool, r.cipher, q, args)
}

// ListAll 管理端全量（含待审核等全部状态），status 过滤由 f.Status 控制。
func (r *demandRepo) ListAll(f repository.DemandFilter) ([]domain.Demand, error) {
	q := `SELECT id, publisher_id, publisher_name, contact, district, city_code,
		biz_type, title, description, images, latitude, longitude, budget_fen, offline_amount_fen, biz_fields,
		status, version, created_at, updated_at
		FROM demands`
	args := []any{}
	argIdx := 1
	conds := []string{}
	if f.Status != "" && f.Status != "all" {
		conds = append(conds, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, f.Status)
		argIdx++
	}
	if f.District != "" {
		conds = append(conds, fmt.Sprintf("district = $%d", argIdx))
		args = append(args, f.District)
		argIdx++
	}
	if f.BizType != "" {
		conds = append(conds, fmt.Sprintf("biz_type = $%d", argIdx))
		args = append(args, f.BizType)
		argIdx++
	}
	if len(conds) > 0 {
		q += " WHERE " + strings.Join(conds, " AND ")
	}
	q += " ORDER BY created_at DESC"
	return scanDemands(r.pool, r.cipher, q, args)
}

func (r *demandRepo) Search(q string) ([]domain.Demand, error) {
	sql := `SELECT id, publisher_id, publisher_name, contact, district, city_code,
		biz_type, title, description, images, latitude, longitude, budget_fen, offline_amount_fen, biz_fields,
		status, version, created_at, updated_at
		FROM demands WHERE status = 'published'
		AND (title ILIKE $1 OR publisher_name ILIKE $1 OR description ILIKE $1)
		ORDER BY created_at DESC LIMIT 50`
	return scanDemands(r.pool, r.cipher, sql, []any{"%" + q + "%"})
}

// ListByPublisher 返回某发布者的全部需求（全状态），供"我的"页统计/查询。
func (r *demandRepo) ListByPublisher(publisherID string) ([]domain.Demand, error) {
	q := `SELECT id, publisher_id, publisher_name, contact, district, city_code,
		biz_type, title, description, images, latitude, longitude, budget_fen, offline_amount_fen, biz_fields,
		status, version, created_at, updated_at
		FROM demands WHERE publisher_id = $1
		ORDER BY created_at DESC`
	return scanDemands(r.pool, r.cipher, q, []any{publisherID})
}

func (r *demandRepo) SetStatus(id string, status domain.DemandStatus) (domain.Demand, error) {
	tag, err := r.pool.Exec(context.Background(),
		`UPDATE demands SET status = $1, updated_at = $2, version = version + 1 WHERE id = $3`,
		string(status), time.Now(), id)
	if err != nil {
		return domain.Demand{}, fmt.Errorf("set demand status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.Demand{}, fmt.Errorf("demand %s not found", id)
	}
	// Fetch the updated row.
	var d domain.Demand
	var images, bizFields []byte
	var bizType string
	err = r.pool.QueryRow(context.Background(),
		`SELECT id, publisher_id, publisher_name, contact, district, city_code,
		biz_type, title, description, images, latitude, longitude, budget_fen, offline_amount_fen, biz_fields,
		status, version, created_at, updated_at
		FROM demands WHERE id = $1`, id).Scan(
		&d.ID, &d.PublisherID, &d.PublisherName, &d.Contact, &d.District, &d.CityCode,
		&bizType, &d.Title, &d.Description, &images, &d.Latitude, &d.Longitude,
		&d.BudgetFen, &d.OfflineAmountFen, &bizFields, &status, &d.Version, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return domain.Demand{}, fmt.Errorf("fetch demand after update: %w", err)
	}
	json.Unmarshal(images, &d.Images)
	json.Unmarshal(bizFields, &d.BizFields)
	d.BizType = domain.BizType(bizType)
	d.Status = domain.DemandStatus(status)
	if r.cipher != nil && d.Contact != "" {
		if dec, err := r.cipher.Decrypt(d.Contact); err == nil {
			d.Contact = dec
		}
	}
	return d, nil
}

func scanDemands(pool *pgxpool.Pool, cipher *crypto.Cipher, q string, args []any) ([]domain.Demand, error) {
	rows, err := pool.Query(context.Background(), q, args...)
	if err != nil {
		return nil, fmt.Errorf("query demands: %w", err)
	}
	defer rows.Close()
	out := []domain.Demand{}
	for rows.Next() {
		var d domain.Demand
		var images, bizFields []byte
		var status, bizType string
		if err := rows.Scan(&d.ID, &d.PublisherID, &d.PublisherName, &d.Contact, &d.District,
			&d.CityCode, &bizType, &d.Title, &d.Description, &images, &d.Latitude, &d.Longitude,
			&d.BudgetFen, &d.OfflineAmountFen, &bizFields, &status, &d.Version, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan demand: %w", err)
		}
		json.Unmarshal(images, &d.Images)
		json.Unmarshal(bizFields, &d.BizFields)
		d.BizType = domain.BizType(bizType)
		d.Status = domain.DemandStatus(status)
		if cipher != nil && d.Contact != "" {
			if dec, err := cipher.Decrypt(d.Contact); err == nil {
				d.Contact = dec
			}
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

// ---- Enterprise Repository ----

type enterpriseRepo struct {
	pool   *pgxpool.Pool
	cipher *crypto.Cipher
}

func (s *Store) NewEnterpriseRepository() repository.EnterpriseRepository {
	return &enterpriseRepo{pool: s.pool, cipher: s.cipher}
}

func (r *enterpriseRepo) Pending() ([]domain.Enterprise, error) {
	rows, err := r.pool.Query(context.Background(), `
		SELECT id, owner_user_id, name, credit_code, legal_person, contact_phone, industry_category, scale, address, description, business_hours, logo, cover_image, license_url, account_name, contact_person, email, founded_at, capability_tags, status, review_comment, is_member, version, created_at, updated_at
		FROM enterprises WHERE status = 'submitted' ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("pending enterprises: %w", err)
	}
	defer rows.Close()
	out := []domain.Enterprise{}
	for rows.Next() {
		var e domain.Enterprise
		var status string
		if err := rows.Scan(&e.ID, &e.OwnerUserID, &e.Name, &e.CreditCode, &e.LegalPerson, &e.ContactPhone, &e.IndustryCategory, &e.Scale, &e.Address, &e.Description, &e.BusinessHours, &e.Logo, &e.CoverImage,
			&e.LicenseURL, &e.AccountName, &e.ContactPerson, &e.Email, &e.FoundedAt, &e.CapabilityTags, &status, &e.ReviewComment, &e.IsMember, &e.Version, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan enterprise: %w", err)
		}
		e.Status = domain.EnterpriseStatus(status)
		if r.cipher != nil {
			if e.LicenseURL != "" {
				if dec, err := r.cipher.Decrypt(e.LicenseURL); err == nil {
					e.LicenseURL = dec
				}
			}
			if e.AccountName != "" {
				if dec, err := r.cipher.Decrypt(e.AccountName); err == nil {
					e.AccountName = dec
				}
			}
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *enterpriseRepo) Create(e domain.Enterprise) (domain.Enterprise, error) {
	if r.cipher != nil {
		if e.LicenseURL != "" {
			enc, err := r.cipher.Encrypt(e.LicenseURL)
			if err != nil {
				return domain.Enterprise{}, fmt.Errorf("encrypt license_url: %w", err)
			}
			e.LicenseURL = enc
		}
		if e.AccountName != "" {
			enc, err := r.cipher.Encrypt(e.AccountName)
			if err != nil {
				return domain.Enterprise{}, fmt.Errorf("encrypt account_name: %w", err)
			}
			e.AccountName = enc
		}
	}
	_, err := r.pool.Exec(context.Background(), `
		INSERT INTO enterprises (id, owner_user_id, name, credit_code, legal_person, contact_phone, industry_category, scale, address, description, business_hours, logo, cover_image, license_url, account_name, contact_person, email, founded_at, capability_tags, status, review_comment, is_member, version, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25)`,
		e.ID, e.OwnerUserID, e.Name, e.CreditCode, e.LegalPerson, e.ContactPhone, e.IndustryCategory, e.Scale, e.Address, e.Description, e.BusinessHours, e.Logo, e.CoverImage, e.LicenseURL, e.AccountName, e.ContactPerson, e.Email, e.FoundedAt, e.CapabilityTags, string(e.Status), e.ReviewComment, e.IsMember, e.Version, e.CreatedAt, e.UpdatedAt)
	if err != nil {
		return domain.Enterprise{}, fmt.Errorf("create enterprise: %w", err)
	}
	return e, nil
}

func (r *enterpriseRepo) Update(id string, e domain.Enterprise) (domain.Enterprise, error) {
	if r.cipher != nil {
		if e.LicenseURL != "" {
			enc, err := r.cipher.Encrypt(e.LicenseURL)
			if err != nil {
				return domain.Enterprise{}, fmt.Errorf("encrypt license_url: %w", err)
			}
			e.LicenseURL = enc
		}
		if e.AccountName != "" {
			enc, err := r.cipher.Encrypt(e.AccountName)
			if err != nil {
				return domain.Enterprise{}, fmt.Errorf("encrypt account_name: %w", err)
			}
			e.AccountName = enc
		}
	}
	e.Version++
	e.UpdatedAt = time.Now()
	tag, err := r.pool.Exec(context.Background(), `
		UPDATE enterprises SET name=$1, credit_code=$2, legal_person=$3, contact_phone=$4, industry_category=$5, scale=$6, address=$7, description=$8, business_hours=$9, logo=$10, cover_image=$11, license_url=$12, account_name=$13, contact_person=$14, email=$15, founded_at=$16, capability_tags=$17, status=$18, review_comment=$19, version=$20, updated_at=$21 WHERE id=$22`,
		e.Name, e.CreditCode, e.LegalPerson, e.ContactPhone, e.IndustryCategory, e.Scale, e.Address, e.Description, e.BusinessHours, e.Logo, e.CoverImage, e.LicenseURL, e.AccountName, e.ContactPerson, e.Email, e.FoundedAt, e.CapabilityTags, string(e.Status), e.ReviewComment, e.Version, e.UpdatedAt, id)
	if err != nil {
		return domain.Enterprise{}, fmt.Errorf("update enterprise: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.Enterprise{}, fmt.Errorf("enterprise %s not found", id)
	}
	return e, nil
}

func (r *enterpriseRepo) FindByID(id string) (domain.Enterprise, error) {
	var e domain.Enterprise
	var status string
	err := r.pool.QueryRow(context.Background(), `
		SELECT id, owner_user_id, name, credit_code, legal_person, contact_phone, industry_category, scale, address, description, business_hours, logo, cover_image, license_url, account_name, contact_person, email, founded_at, capability_tags, status, review_comment, is_member, version, created_at, updated_at
		FROM enterprises WHERE id=$1`, id).Scan(
		&e.ID, &e.OwnerUserID, &e.Name, &e.CreditCode, &e.LegalPerson, &e.ContactPhone, &e.IndustryCategory, &e.Scale, &e.Address, &e.Description, &e.BusinessHours, &e.Logo, &e.CoverImage, &e.LicenseURL, &e.AccountName, &e.ContactPerson, &e.Email, &e.FoundedAt, &e.CapabilityTags, &status, &e.ReviewComment, &e.IsMember, &e.Version, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return domain.Enterprise{}, fmt.Errorf("enterprise %s not found", id)
	}
	e.Status = domain.EnterpriseStatus(status)
	if r.cipher != nil {
		if e.LicenseURL != "" {
			if dec, err := r.cipher.Decrypt(e.LicenseURL); err == nil {
				e.LicenseURL = dec
			}
		}
		if e.AccountName != "" {
			if dec, err := r.cipher.Decrypt(e.AccountName); err == nil {
				e.AccountName = dec
			}
		}
	}
	return e, nil
}

func (r *enterpriseRepo) FindByOwner(userID string) ([]domain.Enterprise, error) {
	return scanEnterprises(r.pool, r.cipher, "WHERE owner_user_id = $1", userID)
}

func (r *enterpriseRepo) ListByStatus(status string, offset, limit int) ([]domain.Enterprise, int, error) {
	// 空 status = 全部状态（与 memory 实现语义一致）
	if status == "" {
		var total int
		if err := r.pool.QueryRow(context.Background(), `SELECT count(*) FROM enterprises`).Scan(&total); err != nil {
			return nil, 0, fmt.Errorf("count enterprises: %w", err)
		}
		items, err := scanEnterprises(r.pool, r.cipher, "ORDER BY created_at DESC LIMIT $1 OFFSET $2", limit, offset)
		return items, total, err
	}
	var total int
	if err := r.pool.QueryRow(context.Background(), `SELECT count(*) FROM enterprises WHERE status=$1`, status).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count enterprises: %w", err)
	}
	items, err := scanEnterprises(r.pool, r.cipher, "WHERE status=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3", status, limit, offset)
	return items, total, err
}

func scanEnterprises(pool *pgxpool.Pool, cipher *crypto.Cipher, where string, args ...any) ([]domain.Enterprise, error) {
	q := `SELECT id, owner_user_id, name, credit_code, legal_person, contact_phone, industry_category, scale, address, description, business_hours, logo, cover_image, license_url, account_name, contact_person, email, founded_at, capability_tags, status, review_comment, is_member, version, created_at, updated_at FROM enterprises ` + where
	rows, err := pool.Query(context.Background(), q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Enterprise{}
	for rows.Next() {
		var e domain.Enterprise
		var status string
		if err := rows.Scan(&e.ID, &e.OwnerUserID, &e.Name, &e.CreditCode, &e.LegalPerson, &e.ContactPhone, &e.IndustryCategory, &e.Scale, &e.Address, &e.Description, &e.BusinessHours, &e.Logo, &e.CoverImage, &e.LicenseURL, &e.AccountName, &e.ContactPerson, &e.Email, &e.FoundedAt, &e.CapabilityTags, &status, &e.ReviewComment, &e.IsMember, &e.Version, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		e.Status = domain.EnterpriseStatus(status)
		if cipher != nil {
			if e.LicenseURL != "" {
				if dec, err := cipher.Decrypt(e.LicenseURL); err == nil {
					e.LicenseURL = dec
				}
			}
			if e.AccountName != "" {
				if dec, err := cipher.Decrypt(e.AccountName); err == nil {
					e.AccountName = dec
				}
			}
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *enterpriseRepo) AddDocument(d domain.EnterpriseDocument) (domain.EnterpriseDocument, error) {
	_, err := r.pool.Exec(context.Background(), `INSERT INTO enterprise_documents(id,enterprise_id,file_id,document_type,review_status,created_at) VALUES($1,$2,$3,$4,$5,$6)`,
		d.ID, d.EnterpriseID, d.FileID, d.DocumentType, d.ReviewStatus, d.CreatedAt)
	if err != nil {
		return domain.EnterpriseDocument{}, fmt.Errorf("insert enterprise document: %w", err)
	}
	return d, nil
}

func (r *enterpriseRepo) ListDocuments(enterpriseID string) ([]domain.EnterpriseDocument, error) {
	rows, err := r.pool.Query(context.Background(), `SELECT id,enterprise_id,file_id,document_type,review_status,created_at FROM enterprise_documents WHERE enterprise_id=$1 ORDER BY created_at DESC`, enterpriseID)
	if err != nil {
		return nil, fmt.Errorf("list enterprise documents: %w", err)
	}
	defer rows.Close()
	out := []domain.EnterpriseDocument{}
	for rows.Next() {
		var d domain.EnterpriseDocument
		if err := rows.Scan(&d.ID, &d.EnterpriseID, &d.FileID, &d.DocumentType, &d.ReviewStatus, &d.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan enterprise document: %w", err)
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (r *enterpriseRepo) Delete(id string) error {
	_, err := r.pool.Exec(context.Background(), `DELETE FROM enterprises WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("delete enterprise %s: %w", id, err)
	}
	return nil
}

func (r *enterpriseRepo) Search(q string) ([]domain.Enterprise, error) {
	rows, err := r.pool.Query(context.Background(), `
		SELECT id, owner_user_id, name, license_url, account_name, status, is_member, version, created_at, updated_at
		FROM enterprises WHERE name ILIKE $1 ORDER BY created_at DESC LIMIT 50`, "%"+q+"%")
	if err != nil {
		return nil, fmt.Errorf("search enterprises: %w", err)
	}
	defer rows.Close()
	out := []domain.Enterprise{}
	for rows.Next() {
		var e domain.Enterprise
		var status string
		if err := rows.Scan(&e.ID, &e.OwnerUserID, &e.Name, &e.LicenseURL, &e.AccountName,
			&status, &e.IsMember, &e.Version, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan enterprise: %w", err)
		}
		e.Status = domain.EnterpriseStatus(status)
		if r.cipher != nil {
			if e.LicenseURL != "" {
				if dec, err := r.cipher.Decrypt(e.LicenseURL); err == nil {
					e.LicenseURL = dec
				}
			}
			if e.AccountName != "" {
				if dec, err := r.cipher.Decrypt(e.AccountName); err == nil {
					e.AccountName = dec
				}
			}
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// ---- Employment Repository ----

type employmentRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewEmploymentRepository() repository.EmploymentRepository {
	return &employmentRepo{pool: s.pool}
}

func (r *employmentRepo) Create(v domain.EmploymentRequest) (domain.EmploymentRequest, error) {
	now := time.Now()
	v.Version = 1
	v.CreatedAt = now
	v.UpdatedAt = now
	_, err := r.pool.Exec(context.Background(), `
		INSERT INTO employment_requests (id, enterprise_id, position, headcount, start_date, end_date, status, version, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		v.ID, v.EnterpriseID, v.Position, v.Headcount, v.StartDate, v.EndDate,
		string(v.Status), v.Version, v.CreatedAt, v.UpdatedAt)
	if err != nil {
		return domain.EmploymentRequest{}, fmt.Errorf("create employment: %w", err)
	}
	return v, nil
}

func (r *employmentRepo) ListByEnterprise(eid string, offset, limit int) ([]domain.EmploymentRequest, int, error) {
	return scanEmploymentPaged(r.pool, "WHERE enterprise_id = $1", offset, limit, eid)
}

func (r *employmentRepo) ListAll(offset, limit int) ([]domain.EmploymentRequest, int, error) {
	return scanEmploymentPaged(r.pool, "", offset, limit)
}

// ---- Contract Repository ----

type contractRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewContractRepository() repository.ContractRepository {
	return &contractRepo{pool: s.pool}
}

// ---- Contract Template Repository ----

type contractTplRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewContractTemplateRepository() repository.ContractTemplateRepository {
	return &contractTplRepo{pool: s.pool}
}

func (r *contractTplRepo) List() ([]domain.ContractTemplate, error) {
	rows, err := r.pool.Query(context.Background(),
		`SELECT id,name,version,COALESCE(content,''),COALESCE(status,'active'),created_at,updated_at FROM contract_templates WHERE COALESCE(status,'active')='active' ORDER BY created_at`)
	if err != nil {
		return nil, fmt.Errorf("list contract templates: %w", err)
	}
	defer rows.Close()
	out := []domain.ContractTemplate{}
	for rows.Next() {
		var t domain.ContractTemplate
		if err := rows.Scan(&t.ID, &t.Name, &t.Version, &t.Content, &t.Status, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *contractTplRepo) Create(t domain.ContractTemplate) (domain.ContractTemplate, error) {
	now := time.Now()
	t.CreatedAt = now
	t.UpdatedAt = now
	_, err := r.pool.Exec(context.Background(),
		`INSERT INTO contract_templates (id,name,version,content,status,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		t.ID, t.Name, t.Version, t.Content, t.Status, t.CreatedAt, t.UpdatedAt)
	if err != nil {
		return domain.ContractTemplate{}, fmt.Errorf("create contract template: %w", err)
	}
	return t, nil
}

func (r *contractRepo) Create(v domain.Contract) (domain.Contract, error) {
	now := time.Now()
	v.Version = 1
	v.CreatedAt = now
	v.UpdatedAt = now
	_, err := r.pool.Exec(context.Background(), `
		INSERT INTO contracts (id, enterprise_id, template_id, sign_url, status, version, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		v.ID, v.EnterpriseID, v.TemplateID, v.SignURL, string(v.Status), v.Version, v.CreatedAt, v.UpdatedAt)
	if err != nil {
		return domain.Contract{}, fmt.Errorf("create contract: %w", err)
	}
	return v, nil
}

func (r *contractRepo) ListByEnterprise(eid string, offset, limit int) ([]domain.Contract, int, error) {
	return scanContractsPaged(r.pool, "WHERE enterprise_id = $1", offset, limit, eid)
}

func (r *contractRepo) ListAll(offset, limit int) ([]domain.Contract, int, error) {
	return scanContractsPaged(r.pool, "", offset, limit)
}

// ---- Job Repository ----

type jobRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewJobRepository() repository.JobRepository { return &jobRepo{pool: s.pool} }

func (r *jobRepo) Create(j domain.Job) (domain.Job, error) {
	now := time.Now()
	j.Version = 1
	j.CreatedAt = now
	j.UpdatedAt = now
	_, err := r.pool.Exec(context.Background(),
		`INSERT INTO jobs (id, enterprise_id, title, description, location, salary_fen, job_type, status, version, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		j.ID, j.EnterpriseID, j.Title, j.Description, j.Location, j.SalaryFen, j.JobType, string(j.Status), j.Version, j.CreatedAt, j.UpdatedAt)
	return j, err
}
func (r *jobRepo) Update(id string, j domain.Job) (domain.Job, error) {
	j.Version++
	j.UpdatedAt = time.Now()
	tag, err := r.pool.Exec(context.Background(),
		`UPDATE jobs SET title=$1,description=$2,location=$3,salary_fen=$4,job_type=$5,status=$6,version=$7,updated_at=$8 WHERE id=$9`,
		j.Title, j.Description, j.Location, j.SalaryFen, j.JobType, string(j.Status), j.Version, j.UpdatedAt, id)
	if err != nil {
		return domain.Job{}, err
	}
	if tag.RowsAffected() == 0 {
		return domain.Job{}, fmt.Errorf("job %s not found", id)
	}
	return j, nil
}
func (r *jobRepo) FindByID(id string) (domain.Job, error) {
	var j domain.Job
	var status string
	err := r.pool.QueryRow(context.Background(),
		`SELECT id, enterprise_id, title, description, location, salary_fen, job_type, status, version, created_at, updated_at FROM jobs WHERE id=$1`, id).
		Scan(&j.ID, &j.EnterpriseID, &j.Title, &j.Description, &j.Location, &j.SalaryFen, &j.JobType, &status, &j.Version, &j.CreatedAt, &j.UpdatedAt)
	j.Status = domain.JobStatus(status)
	return j, err
}
func (r *jobRepo) ListByEnterprise(eid string) ([]domain.Job, error) {
	return scanJobs(r.pool, "WHERE enterprise_id=$1 ORDER BY created_at DESC", eid)
}
func (r *jobRepo) ListPublished(offset, limit int) ([]domain.Job, int, error) {
	var total int
	if err := r.pool.QueryRow(context.Background(), `SELECT count(*) FROM jobs WHERE status='published'`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count published jobs: %w", err)
	}
	items, err := scanJobs(r.pool, "WHERE status='published' ORDER BY created_at DESC LIMIT $1 OFFSET $2", limit, offset)
	return items, total, err
}
func (r *jobRepo) ListAll(offset, limit int) ([]domain.Job, int, error) {
	var total int
	if err := r.pool.QueryRow(context.Background(), `SELECT count(*) FROM jobs`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count jobs: %w", err)
	}
	items, err := scanJobs(r.pool, "ORDER BY created_at DESC LIMIT $1 OFFSET $2", limit, offset)
	return items, total, err
}
func scanJobs(pool *pgxpool.Pool, where string, args ...any) ([]domain.Job, error) {
	q := `SELECT id, enterprise_id, title, description, location, salary_fen, job_type, status, version, created_at, updated_at FROM jobs ` + where
	rows, err := pool.Query(context.Background(), q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Job{}
	for rows.Next() {
		var j domain.Job
		var status string
		if err := rows.Scan(&j.ID, &j.EnterpriseID, &j.Title, &j.Description, &j.Location, &j.SalaryFen, &j.JobType, &status, &j.Version, &j.CreatedAt, &j.UpdatedAt); err != nil {
			return nil, err
		}
		j.Status = domain.JobStatus(status)
		out = append(out, j)
	}
	return out, rows.Err()
}

func (r *jobRepo) Delete(id string) error {
	_, err := r.pool.Exec(context.Background(), `DELETE FROM jobs WHERE id=$1`, id)
	return err
}

// ---- Resume Repository ----

type pgResumeRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewResumeRepository() repository.ResumeRepository { return &pgResumeRepo{pool: s.pool} }

func (r *pgResumeRepo) Create(v domain.Resume) (domain.Resume, error) {
	now := time.Now()
	v.Version = 1
	v.CreatedAt = now
	v.UpdatedAt = now
	skills, _ := json.Marshal(v.Skills) // text 列存 JSON，读取端 json.Unmarshal 对称解析
	_, err := r.pool.Exec(context.Background(),
		`INSERT INTO resumes (id, user_id, title, name, phone, email, education, work_experience, skills, certificate_url, content, visibility, version, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		v.ID, v.UserID, v.Title, v.Name, v.Phone, v.Email, v.Education, v.WorkExperience, skills, v.CertificateURL, v.Content, v.Visibility, v.Version, v.CreatedAt, v.UpdatedAt)
	return v, err
}
func (r *pgResumeRepo) Update(id string, v domain.Resume) (domain.Resume, error) {
	v.Version++
	v.UpdatedAt = time.Now()
	skills, _ := json.Marshal(v.Skills)
	tag, err := r.pool.Exec(context.Background(),
		`UPDATE resumes SET title=$1,name=$2,phone=$3,email=$4,education=$5,work_experience=$6,skills=$7,certificate_url=$8,content=$9,visibility=$10,version=$11,updated_at=$12 WHERE id=$13`,
		v.Title, v.Name, v.Phone, v.Email, v.Education, v.WorkExperience, skills, v.CertificateURL, v.Content, v.Visibility, v.Version, v.UpdatedAt, id)
	if err != nil {
		return domain.Resume{}, err
	}
	if tag.RowsAffected() == 0 {
		return domain.Resume{}, fmt.Errorf("resume %s not found", id)
	}
	return v, nil
}
func (r *pgResumeRepo) FindByID(id string) (domain.Resume, error) {
	var v domain.Resume
	var skills string
	err := r.pool.QueryRow(context.Background(),
		`SELECT id, user_id, title, name, phone, email, education, work_experience, skills, certificate_url, content, visibility, version, created_at, updated_at FROM resumes WHERE id=$1`, id).
		Scan(&v.ID, &v.UserID, &v.Title, &v.Name, &v.Phone, &v.Email, &v.Education, &v.WorkExperience, &skills, &v.CertificateURL, &v.Content, &v.Visibility, &v.Version, &v.CreatedAt, &v.UpdatedAt)
	if err == nil {
		json.Unmarshal([]byte(skills), &v.Skills)
	}
	return v, err
}
func (r *pgResumeRepo) ListByUser(userID string) ([]domain.Resume, error) {
	return scanResumes(r.pool, "WHERE user_id=$1 ORDER BY created_at DESC", userID)
}
func (r *pgResumeRepo) ListAll(offset, limit int) ([]domain.Resume, int, error) {
	var total int
	if err := r.pool.QueryRow(context.Background(), `SELECT count(*) FROM resumes`).Scan(&total); err != nil {
		return nil, 0, err
	}
	items, err := scanResumes(r.pool, "ORDER BY created_at DESC LIMIT $1 OFFSET $2", limit, offset)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}
func scanResumes(pool *pgxpool.Pool, where string, args ...any) ([]domain.Resume, error) {
	q := `SELECT id, user_id, title, name, phone, email, education, work_experience, skills, certificate_url, content, visibility, version, created_at, updated_at FROM resumes ` + where
	rows, err := pool.Query(context.Background(), q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Resume{}
	for rows.Next() {
		var v domain.Resume
		var skills string
		if err := rows.Scan(&v.ID, &v.UserID, &v.Title, &v.Name, &v.Phone, &v.Email, &v.Education, &v.WorkExperience, &skills, &v.CertificateURL, &v.Content, &v.Visibility, &v.Version, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, err
		}
		json.Unmarshal([]byte(skills), &v.Skills)
		out = append(out, v)
	}
	return out, rows.Err()
}

// ---- JobApplication Repository ----

type pgAppRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewJobApplicationRepository() repository.JobApplicationRepository {
	return &pgAppRepo{pool: s.pool}
}

func (r *pgAppRepo) Create(a domain.JobApplication) (domain.JobApplication, error) {
	now := time.Now()
	a.Version = 1
	a.CreatedAt = now
	a.UpdatedAt = now
	_, err := r.pool.Exec(context.Background(),
		`INSERT INTO job_applications (id, job_id, resume_id, applicant_id, status, version, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		a.ID, a.JobID, a.ResumeID, a.ApplicantID, string(a.Status), a.Version, a.CreatedAt, a.UpdatedAt)
	return a, err
}
func (r *pgAppRepo) FindByID(id string) (domain.JobApplication, error) {
	var a domain.JobApplication
	var s string
	err := r.pool.QueryRow(context.Background(),
		`SELECT id, job_id, resume_id, applicant_id, status, version, created_at, updated_at FROM job_applications WHERE id=$1`, id).
		Scan(&a.ID, &a.JobID, &a.ResumeID, &a.ApplicantID, &s, &a.Version, &a.CreatedAt, &a.UpdatedAt)
	a.Status = domain.AppStatus(s)
	return a, err
}
func (r *pgAppRepo) UpdateStatus(id string, status domain.AppStatus) (domain.JobApplication, error) {
	tag, err := r.pool.Exec(context.Background(),
		`UPDATE job_applications SET status=$1, updated_at=$2, version=version+1 WHERE id=$3`,
		string(status), time.Now(), id)
	if err != nil {
		return domain.JobApplication{}, err
	}
	if tag.RowsAffected() == 0 {
		return domain.JobApplication{}, fmt.Errorf("application %s not found", id)
	}
	var a domain.JobApplication
	var s string
	err = r.pool.QueryRow(context.Background(),
		`SELECT id, job_id, resume_id, applicant_id, status, version, created_at, updated_at FROM job_applications WHERE id=$1`, id).
		Scan(&a.ID, &a.JobID, &a.ResumeID, &a.ApplicantID, &s, &a.Version, &a.CreatedAt, &a.UpdatedAt)
	a.Status = domain.AppStatus(s)
	return a, err
}
func (r *pgAppRepo) ListByJob(jobID string) ([]domain.JobApplication, error) {
	return scanApps(r.pool, "WHERE job_id=$1 ORDER BY created_at DESC", jobID)
}
func (r *pgAppRepo) ListByApplicant(userID string) ([]domain.JobApplication, error) {
	return scanApps(r.pool, "WHERE applicant_id=$1 ORDER BY created_at DESC", userID)
}
func scanApps(pool *pgxpool.Pool, where string, args ...any) ([]domain.JobApplication, error) {
	q := `SELECT id, job_id, resume_id, applicant_id, status, version, created_at, updated_at FROM job_applications ` + where
	rows, err := pool.Query(context.Background(), q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.JobApplication{}
	for rows.Next() {
		var a domain.JobApplication
		var status string
		if err := rows.Scan(&a.ID, &a.JobID, &a.ResumeID, &a.ApplicantID, &status, &a.Version, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		a.Status = domain.AppStatus(status)
		out = append(out, a)
	}
	return out, rows.Err()
}

// ---- Post ----

type pgPostRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewPostRepository() repository.PostRepository { return &pgPostRepo{pool: s.pool} }
func (r *pgPostRepo) Create(p domain.Post) (domain.Post, error) {
	img, err := json.Marshal(p.Images)
	if err != nil {
		return domain.Post{}, fmt.Errorf("marshal post images: %w", err)
	}
	now := time.Now()
	p.Version = 1
	p.CreatedAt = now
	p.UpdatedAt = now
	_, err = r.pool.Exec(context.Background(),
		`INSERT INTO posts(id,author_id,title,content,images,city_code,status,version,created_at,updated_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		p.ID, p.AuthorID, p.Title, p.Content, img, p.CityCode, p.Status, p.Version, p.CreatedAt, p.UpdatedAt)
	return p, err
}
func (r *pgPostRepo) Update(id string, p domain.Post) (domain.Post, error) {
	img, err := json.Marshal(p.Images)
	if err != nil {
		return domain.Post{}, fmt.Errorf("marshal post images: %w", err)
	}
	p.Version++
	p.UpdatedAt = time.Now()
	_, err = r.pool.Exec(context.Background(), `UPDATE posts SET title=$1,content=$2,images=$3,status=$4,version=$5,updated_at=$6 WHERE id=$7`,
		p.Title, p.Content, img, p.Status, p.Version, p.UpdatedAt, id)
	return p, err
}
func (r *pgPostRepo) FindByID(id string) (domain.Post, error) {
	var p domain.Post
	var img []byte
	err := r.pool.QueryRow(context.Background(), `SELECT id,author_id,title,content,images,city_code,status,version,created_at,updated_at FROM posts WHERE id=$1`, id).
		Scan(&p.ID, &p.AuthorID, &p.Title, &p.Content, &img, &p.CityCode, &p.Status, &p.Version, &p.CreatedAt, &p.UpdatedAt)
	json.Unmarshal(img, &p.Images)
	return p, err
}
func (r *pgPostRepo) ListPublished(offset, limit int) ([]domain.Post, int, error) {
	var total int
	if err := r.pool.QueryRow(context.Background(), `SELECT count(*) FROM posts WHERE status='published'`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count published posts: %w", err)
	}
	rows, err := r.pool.Query(context.Background(), `SELECT id,author_id,title,content,images,city_code,status,version,created_at,updated_at FROM posts WHERE status='published' ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []domain.Post{}
	for rows.Next() {
		var p domain.Post
		var img []byte
		if err := rows.Scan(&p.ID, &p.AuthorID, &p.Title, &p.Content, &img, &p.CityCode, &p.Status, &p.Version, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, 0, err
		}
		json.Unmarshal(img, &p.Images)
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}
func (r *pgPostRepo) ListByAuthor(uid string) ([]domain.Post, error) {
	rows, err := r.pool.Query(context.Background(), `SELECT id,author_id,title,content,images,city_code,status,version,created_at,updated_at FROM posts WHERE author_id=$1 ORDER BY created_at DESC`, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Post{}
	for rows.Next() {
		var p domain.Post
		var img []byte
		if err := rows.Scan(&p.ID, &p.AuthorID, &p.Title, &p.Content, &img, &p.CityCode, &p.Status, &p.Version, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		json.Unmarshal(img, &p.Images)
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// ---- Comment ----

type pgCommentRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewCommentRepository() repository.CommentRepository {
	return &pgCommentRepo{pool: s.pool}
}
func (r *pgCommentRepo) Create(c domain.Comment) (domain.Comment, error) {
	_, err := r.pool.Exec(context.Background(), `INSERT INTO comments(id,post_id,author_id,content,status,created_at) VALUES($1,$2,$3,$4,$5,$6)`,
		c.ID, c.PostID, c.AuthorID, c.Content, c.Status, time.Now())
	return c, err
}
func (r *pgCommentRepo) ListByPost(postID string) ([]domain.Comment, error) {
	rows, err := r.pool.Query(context.Background(), `SELECT id,post_id,author_id,content,status,created_at FROM comments WHERE post_id=$1 ORDER BY created_at`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Comment{}
	for rows.Next() {
		var c domain.Comment
		if err := rows.Scan(&c.ID, &c.PostID, &c.AuthorID, &c.Content, &c.Status, &c.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// ---- Report ----

type pgReportRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewReportRepository() repository.ReportRepository { return &pgReportRepo{pool: s.pool} }
func (r *pgReportRepo) Create(rp domain.Report) (domain.Report, error) {
	_, err := r.pool.Exec(context.Background(), `INSERT INTO reports(id,reporter_id,resource_type,resource_id,reason,status,created_at) VALUES($1,$2,$3,$4,$5,$6,$7)`,
		rp.ID, rp.ReporterID, rp.ResourceType, rp.ResourceID, rp.Reason, "pending", time.Now())
	return rp, err
}
func (r *pgReportRepo) ListPending(offset, limit int) ([]domain.Report, int, error) {
	var total int
	if err := r.pool.QueryRow(context.Background(), `SELECT count(*) FROM reports WHERE status='pending'`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count pending reports: %w", err)
	}
	rows, err := r.pool.Query(context.Background(), `SELECT id,reporter_id,resource_type,resource_id,reason,status,created_at FROM reports WHERE status='pending' ORDER BY created_at LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []domain.Report{}
	for rows.Next() {
		var rp domain.Report
		if err := rows.Scan(&rp.ID, &rp.ReporterID, &rp.ResourceType, &rp.ResourceID, &rp.Reason, &rp.Status, &rp.CreatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, rp)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

// ---- Listing ----

type pgListingRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewListingRepository() repository.ListingRepository {
	return &pgListingRepo{pool: s.pool}
}
func (r *pgListingRepo) Create(l domain.Listing) (domain.Listing, error) {
	img, err := json.Marshal(l.Images)
	if err != nil {
		return domain.Listing{}, fmt.Errorf("marshal listing images: %w", err)
	}
	now := time.Now()
	l.Version = 1
	l.CreatedAt = now
	l.UpdatedAt = now
	_, err = r.pool.Exec(context.Background(), `INSERT INTO listings(id,seller_id,title,description,category,price_fen,images,district,status,version,created_at,updated_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		l.ID, l.SellerID, l.Title, l.Description, l.Category, l.PriceFen, img, l.District, l.Status, l.Version, l.CreatedAt, l.UpdatedAt)
	return l, err
}
func (r *pgListingRepo) Update(id string, l domain.Listing) (domain.Listing, error) {
	img, err := json.Marshal(l.Images)
	if err != nil {
		return domain.Listing{}, fmt.Errorf("marshal listing images: %w", err)
	}
	l.Version++
	l.UpdatedAt = time.Now()
	_, err = r.pool.Exec(context.Background(), `UPDATE listings SET title=$1,description=$2,price_fen=$3,images=$4,status=$5,version=$6,updated_at=$7 WHERE id=$8`,
		l.Title, l.Description, l.PriceFen, img, l.Status, l.Version, l.UpdatedAt, id)
	return l, err
}
func (r *pgListingRepo) FindByID(id string) (domain.Listing, error) {
	var l domain.Listing
	var img []byte
	err := r.pool.QueryRow(context.Background(), `SELECT id,seller_id,title,description,category,price_fen,images,district,status,version,created_at,updated_at FROM listings WHERE id=$1`, id).
		Scan(&l.ID, &l.SellerID, &l.Title, &l.Description, &l.Category, &l.PriceFen, &img, &l.District, &l.Status, &l.Version, &l.CreatedAt, &l.UpdatedAt)
	json.Unmarshal(img, &l.Images)
	return l, err
}
func (r *pgListingRepo) ListByStatus(status string, offset, limit int) ([]domain.Listing, int, error) {
	var total int
	where := ""
	args := []any{}
	if status != "" {
		where = "WHERE status=$1"
		args = append(args, status)
	}
	if err := r.pool.QueryRow(context.Background(), `SELECT count(*) FROM listings `+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count listings: %w", err)
	}
	args = append(args, limit, offset)
	rows, err := r.pool.Query(context.Background(), `SELECT id,seller_id,title,description,category,price_fen,images,district,status,version,created_at,updated_at FROM listings `+where+` ORDER BY created_at DESC LIMIT $`+fmt.Sprintf("%d", len(args)-1)+` OFFSET $`+fmt.Sprintf("%d", len(args)), args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []domain.Listing{}
	for rows.Next() {
		var l domain.Listing
		var img []byte
		if err := rows.Scan(&l.ID, &l.SellerID, &l.Title, &l.Description, &l.Category, &l.PriceFen, &img, &l.District, &l.Status, &l.Version, &l.CreatedAt, &l.UpdatedAt); err != nil {
			return nil, 0, err
		}
		json.Unmarshal(img, &l.Images)
		out = append(out, l)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}
func (r *pgListingRepo) ListBySeller(uid string) ([]domain.Listing, error) {
	rows, err := r.pool.Query(context.Background(), `SELECT id,seller_id,title,description,category,price_fen,images,district,status,version,created_at,updated_at FROM listings WHERE seller_id=$1 ORDER BY created_at DESC`, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Listing{}
	for rows.Next() {
		var l domain.Listing
		var img []byte
		if err := rows.Scan(&l.ID, &l.SellerID, &l.Title, &l.Description, &l.Category, &l.PriceFen, &img, &l.District, &l.Status, &l.Version, &l.CreatedAt, &l.UpdatedAt); err != nil {
			return nil, err
		}
		json.Unmarshal(img, &l.Images)
		out = append(out, l)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
func (r *pgListingRepo) AddFavorite(lid, uid string) error {
	_, err := r.pool.Exec(context.Background(), `INSERT INTO listing_favorites(listing_id,user_id,created_at) VALUES($1,$2,$3) ON CONFLICT DO NOTHING`, lid, uid, time.Now())
	if err != nil {
		return fmt.Errorf("add favorite: %w", err)
	}
	return nil
}
func (r *pgListingRepo) RemoveFavorite(lid, uid string) error {
	_, err := r.pool.Exec(context.Background(), `DELETE FROM listing_favorites WHERE listing_id=$1 AND user_id=$2`, lid, uid)
	if err != nil {
		return fmt.Errorf("remove favorite: %w", err)
	}
	return nil
}

// ---- Labour ----

type pgLabourRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewLabourOrderRepository() repository.LabourOrderRepository {
	return &pgLabourRepo{pool: s.pool}
}
func (r *pgLabourRepo) Create(o domain.LabourOrder) (domain.LabourOrder, error) {
	now := time.Now()
	o.Version = 1
	o.CreatedAt = now
	o.UpdatedAt = now
	_, err := r.pool.Exec(context.Background(), `INSERT INTO labour_orders(id,employer_id,title,description,worker_count,start_date,end_date,budget_fen,status,version,created_at,updated_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		o.ID, o.EmployerID, o.Title, o.Description, o.WorkerCount, o.StartDate, o.EndDate, o.BudgetFen, o.Status, o.Version, o.CreatedAt, o.UpdatedAt)
	return o, err
}
func (r *pgLabourRepo) FindByID(id string) (domain.LabourOrder, error) {
	var o domain.LabourOrder
	err := r.pool.QueryRow(context.Background(), `SELECT id,employer_id,title,description,worker_count,start_date,end_date,budget_fen,status,version,created_at,updated_at FROM labour_orders WHERE id=$1`, id).
		Scan(&o.ID, &o.EmployerID, &o.Title, &o.Description, &o.WorkerCount, &o.StartDate, &o.EndDate, &o.BudgetFen, &o.Status, &o.Version, &o.CreatedAt, &o.UpdatedAt)
	return o, err
}
func (r *pgLabourRepo) ListByEmployer(uid string) ([]domain.LabourOrder, error) {
	rows, err := r.pool.Query(context.Background(), `SELECT id,employer_id,title,description,worker_count,start_date,end_date,budget_fen,status,version,created_at,updated_at FROM labour_orders WHERE employer_id=$1 ORDER BY created_at DESC`, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.LabourOrder{}
	for rows.Next() {
		var o domain.LabourOrder
		if err := rows.Scan(&o.ID, &o.EmployerID, &o.Title, &o.Description, &o.WorkerCount, &o.StartDate, &o.EndDate, &o.BudgetFen, &o.Status, &o.Version, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
func (r *pgLabourRepo) ListAll(offset, limit int) ([]domain.LabourOrder, int, error) {
	var total int
	if err := r.pool.QueryRow(context.Background(), `SELECT count(*) FROM labour_orders`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count labour orders: %w", err)
	}
	rows, err := r.pool.Query(context.Background(), `SELECT id,employer_id,title,description,worker_count,start_date,end_date,budget_fen,status,version,created_at,updated_at FROM labour_orders ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []domain.LabourOrder{}
	for rows.Next() {
		var o domain.LabourOrder
		if err := rows.Scan(&o.ID, &o.EmployerID, &o.Title, &o.Description, &o.WorkerCount, &o.StartDate, &o.EndDate, &o.BudgetFen, &o.Status, &o.Version, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, o)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}
func (r *pgLabourRepo) CreateQuote(q domain.LabourQuote) (domain.LabourQuote, error) {
	_, err := r.pool.Exec(context.Background(), `INSERT INTO labour_quotes(id,order_id,quoter_id,quoter_name,amount_fen,proposal,status,created_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8)`,
		q.ID, q.OrderID, q.QuoterID, q.QuoterName, q.AmountFen, q.Proposal, "pending", time.Now())
	return q, err
}
func (r *pgLabourRepo) ListQuotes(orderID string) ([]domain.LabourQuote, error) {
	rows, err := r.pool.Query(context.Background(), `SELECT id,order_id,quoter_id,quoter_name,amount_fen,proposal,status,created_at FROM labour_quotes WHERE order_id=$1 ORDER BY created_at`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.LabourQuote{}
	for rows.Next() {
		var q domain.LabourQuote
		if err := rows.Scan(&q.ID, &q.OrderID, &q.QuoterID, &q.QuoterName, &q.AmountFen, &q.Proposal, &q.Status, &q.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, q)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
func (r *pgLabourRepo) CreateAssignment(a domain.Assignment) (domain.Assignment, error) {
	_, err := r.pool.Exec(context.Background(), `INSERT INTO assignments(id,order_id,worker_id,status,created_at) VALUES($1,$2,$3,$4,$5)`,
		a.ID, a.OrderID, a.WorkerID, "assigned", time.Now())
	return a, err
}

func (r *pgLabourRepo) ListAssignmentsByOrder(orderID string) ([]domain.Assignment, error) {
	rows, err := r.pool.Query(context.Background(), `SELECT id,order_id,worker_id,status,created_at FROM assignments WHERE order_id=$1 ORDER BY created_at DESC`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Assignment{}
	for rows.Next() {
		var a domain.Assignment
		if err := rows.Scan(&a.ID, &a.OrderID, &a.WorkerID, &a.Status, &a.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *pgLabourRepo) ListAssignmentsByWorker(workerID string) ([]domain.Assignment, error) {
	rows, err := r.pool.Query(context.Background(), `SELECT id,order_id,worker_id,status,created_at FROM assignments WHERE worker_id=$1 ORDER BY created_at DESC`, workerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Assignment{}
	for rows.Next() {
		var a domain.Assignment
		if err := rows.Scan(&a.ID, &a.OrderID, &a.WorkerID, &a.Status, &a.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// ---- User ----

type userRepo struct {
	pool   *pgxpool.Pool
	cipher *crypto.Cipher
}

func (s *Store) NewUserRepository() repository.UserRepository {
	return &userRepo{pool: s.pool, cipher: s.cipher}
}

func (r *userRepo) FindByOpenID(openid string) (domain.User, error) {
	var u domain.User
	err := r.pool.QueryRow(context.Background(),
		`SELECT id, wechat_openid, phone_ciphertext, password_hash, name, avatar_url, gender, birthday, region, bio, role, status, version, created_at, updated_at FROM users WHERE wechat_openid=$1 AND deleted_at IS NULL`, openid).
		Scan(&u.ID, &u.WechatOpenID, &u.PhoneCipher, &u.PasswordHash, &u.Name, &u.AvatarURL, &u.Gender, &u.Birthday, &u.Region, &u.Bio, &u.Role, &u.Status, &u.Version, &u.CreatedAt, &u.UpdatedAt)
	if r.cipher != nil && u.PhoneCipher != "" {
		if dec, err := r.cipher.Decrypt(u.PhoneCipher); err == nil {
			u.PhoneCipher = dec
		}
	}
	return u, err
}

func (r *userRepo) Create(u domain.User) (domain.User, error) {
	now := time.Now()
	u.Version = 1
	u.CreatedAt = now
	u.UpdatedAt = now
	if u.Role == "" {
		u.Role = domain.RoleIndividual
	}
	_, err := r.pool.Exec(context.Background(),
		`INSERT INTO users (id, wechat_openid, phone_ciphertext, password_hash, name, avatar_url, role, status, version, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		u.ID, u.WechatOpenID, u.PhoneCipher, u.PasswordHash, u.Name, u.AvatarURL, string(u.Role), u.Status, u.Version, u.CreatedAt, u.UpdatedAt)
	return u, err
}

func (r *userRepo) FindByID(id string) (domain.User, error) {
	var u domain.User
	err := r.pool.QueryRow(context.Background(),
		`SELECT id, wechat_openid, phone_ciphertext, password_hash, name, avatar_url, gender, birthday, region, bio, role, status, version, created_at, updated_at FROM users WHERE id=$1 AND deleted_at IS NULL`, id).
		Scan(&u.ID, &u.WechatOpenID, &u.PhoneCipher, &u.PasswordHash, &u.Name, &u.AvatarURL, &u.Gender, &u.Birthday, &u.Region, &u.Bio, &u.Role, &u.Status, &u.Version, &u.CreatedAt, &u.UpdatedAt)
	if r.cipher != nil && u.PhoneCipher != "" {
		if dec, err := r.cipher.Decrypt(u.PhoneCipher); err == nil {
			u.PhoneCipher = dec
		}
	}
	return u, err
}

func (r *userRepo) All() ([]domain.User, error) {
	rows, err := r.pool.Query(context.Background(), `SELECT id, wechat_openid, phone_ciphertext, password_hash, name, avatar_url, gender, birthday, region, bio, role, status, version, created_at, updated_at FROM users WHERE deleted_at IS NULL ORDER BY created_at DESC LIMIT 200`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.WechatOpenID, &u.PhoneCipher, &u.PasswordHash, &u.Name, &u.AvatarURL, &u.Gender, &u.Birthday, &u.Region, &u.Bio, &u.Role, &u.Status, &u.Version, &u.CreatedAt, &u.UpdatedAt); err != nil {
			continue
		}
		if r.cipher != nil && u.PhoneCipher != "" {
			if dec, err := r.cipher.Decrypt(u.PhoneCipher); err == nil {
				u.PhoneCipher = dec
			}
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

func (r *userRepo) UpdateRole(id string, role domain.Role) error {
	_, err := r.pool.Exec(context.Background(), `UPDATE users SET role=$1, version=version+1, updated_at=NOW() WHERE id=$2`, string(role), id)
	if err != nil {
		return fmt.Errorf("update role for %s: %w", id, err)
	}
	return nil
}

func (r *userRepo) UpdateAvatar(userID, avatarURL string) error {
	_, err := r.pool.Exec(context.Background(), `UPDATE users SET avatar_url=$1, updated_at=NOW() WHERE id=$2`, avatarURL, userID)
	if err != nil {
		return fmt.Errorf("update avatar for %s: %w", userID, err)
	}
	return nil
}

func (r *userRepo) UpdateName(userID, name string) error {
	_, err := r.pool.Exec(context.Background(), `UPDATE users SET name=$1, updated_at=NOW() WHERE id=$2`, name, userID)
	if err != nil {
		return fmt.Errorf("update name for %s: %w", userID, err)
	}
	return nil
}

// UpdateProfile updates the editable profile fields. Phone is plaintext here;
// it is encrypted before persistence. An empty Phone leaves it unchanged.
func (r *userRepo) UpdateProfile(id string, p domain.UserProfile) error {
	enc := p.Phone
	if p.Phone != "" && r.cipher != nil {
		if c, err := r.cipher.Encrypt(p.Phone); err == nil {
			enc = c
		}
	}
	_, err := r.pool.Exec(context.Background(),
		`UPDATE users SET gender=$2, birthday=$3, region=$4, bio=$5,
		 phone_ciphertext=CASE WHEN $6='' THEN phone_ciphertext ELSE $6 END,
		 updated_at=NOW() WHERE id=$1`,
		id, p.Gender, p.Birthday, p.Region, p.Bio, enc)
	if err != nil {
		return fmt.Errorf("update profile for %s: %w", id, err)
	}
	return nil
}

// Delete removes a user together with its session and role rows in one
// transaction — the refresh_tokens/user_roles foreign keys otherwise block
// deletion of any user that has logged in.
func (r *userRepo) Delete(id string) error {
	tx, err := r.pool.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("begin delete user: %w", err)
	}
	defer tx.Rollback(context.Background())
	if _, err := tx.Exec(context.Background(), `DELETE FROM refresh_tokens WHERE user_id=$1`, id); err != nil {
		return fmt.Errorf("delete refresh tokens for %s: %w", id, err)
	}
	if _, err := tx.Exec(context.Background(), `DELETE FROM user_roles WHERE user_id=$1`, id); err != nil {
		return fmt.Errorf("delete user roles for %s: %w", id, err)
	}
	if _, err := tx.Exec(context.Background(), `DELETE FROM users WHERE id=$1`, id); err != nil {
		return fmt.Errorf("delete user %s: %w", id, err)
	}
	return tx.Commit(context.Background())
}

// ---- RefreshToken ----

type refreshTokenRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewRefreshTokenRepository() repository.RefreshTokenRepository {
	return &refreshTokenRepo{pool: s.pool}
}

func (r *refreshTokenRepo) Store(userID, tokenHash string, expiresAt time.Time) error {
	_, err := r.pool.Exec(context.Background(),
		`INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, created_at) VALUES ($1,$2,$3,$4,$5)`,
		fmt.Sprintf("rt-%d", time.Now().UnixNano()), userID, tokenHash, expiresAt, time.Now())
	if err != nil {
		return fmt.Errorf("store refresh token: %w", err)
	}
	return nil
}

func (r *refreshTokenRepo) Find(tokenHash string) (userID string, expiresAt time.Time, revoked bool, err error) {
	var revokedAt *time.Time
	err = r.pool.QueryRow(context.Background(),
		`SELECT user_id, expires_at, revoked_at FROM refresh_tokens WHERE token_hash=$1`, tokenHash).
		Scan(&userID, &expiresAt, &revokedAt)
	revoked = revokedAt != nil
	return
}

func (r *refreshTokenRepo) Revoke(tokenHash string) error {
	_, err := r.pool.Exec(context.Background(),
		`UPDATE refresh_tokens SET revoked_at=$1 WHERE token_hash=$2`, time.Now(), tokenHash)
	if err != nil {
		return fmt.Errorf("revoke refresh token: %w", err)
	}
	return nil
}

// scanContracts removed — replaced by scanContractsPaged

// ---- DemandIntent Repository ----

type pgIntentRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewIntentRepository() repository.IntentRepository {
	return &pgIntentRepo{pool: s.pool}
}

func (r *pgIntentRepo) Create(it domain.DemandIntent) (domain.DemandIntent, error) {
	now := time.Now()
	it.Version = 1
	it.CreatedAt = now
	it.UpdatedAt = now
	_, err := r.pool.Exec(context.Background(),
		`INSERT INTO demand_intents (id, demand_id, intentor_id, intentor_name, contact, remark, status, version, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		it.ID, it.DemandID, it.IntentorID, it.IntentorName, it.Contact, it.Remark, it.Status, it.Version, it.CreatedAt, it.UpdatedAt)
	if err != nil {
		return domain.DemandIntent{}, fmt.Errorf("create intent: %w", err)
	}
	return it, nil
}

func (r *pgIntentRepo) ListByDemand(demandID string) ([]domain.DemandIntent, error) {
	rows, err := r.pool.Query(context.Background(),
		`SELECT id, demand_id, intentor_id, intentor_name, contact, remark, status, version, created_at, updated_at
		FROM demand_intents WHERE demand_id=$1 ORDER BY created_at DESC`, demandID)
	if err != nil {
		return nil, fmt.Errorf("query intents: %w", err)
	}
	defer rows.Close()
	out := []domain.DemandIntent{}
	for rows.Next() {
		var it domain.DemandIntent
		if err := rows.Scan(&it.ID, &it.DemandID, &it.IntentorID, &it.IntentorName, &it.Contact, &it.Remark, &it.Status, &it.Version, &it.CreatedAt, &it.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan intent: %w", err)
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

func (r *pgIntentRepo) ListByIntentor(intentorID string) ([]domain.DemandIntent, error) {
	rows, err := r.pool.Query(context.Background(),
		`SELECT id, demand_id, intentor_id, intentor_name, contact, remark, status, version, created_at, updated_at
		FROM demand_intents WHERE intentor_id=$1 ORDER BY created_at DESC`, intentorID)
	if err != nil {
		return nil, fmt.Errorf("query intents by intentor: %w", err)
	}
	defer rows.Close()
	out := []domain.DemandIntent{}
	for rows.Next() {
		var it domain.DemandIntent
		if err := rows.Scan(&it.ID, &it.DemandID, &it.IntentorID, &it.IntentorName, &it.Contact, &it.Remark, &it.Status, &it.Version, &it.CreatedAt, &it.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan intent: %w", err)
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

func (r *pgIntentRepo) UpdateStatus(id string, status string) (domain.DemandIntent, error) {
	var it domain.DemandIntent
	err := r.pool.QueryRow(context.Background(),
		`UPDATE demand_intents SET status=$2, version=version+1, updated_at=now()
		WHERE id=$1
		RETURNING id, demand_id, intentor_id, intentor_name, contact, remark, status, version, created_at, updated_at`,
		id, status).
		Scan(&it.ID, &it.DemandID, &it.IntentorID, &it.IntentorName, &it.Contact, &it.Remark, &it.Status, &it.Version, &it.CreatedAt, &it.UpdatedAt)
	if err != nil {
		return domain.DemandIntent{}, fmt.Errorf("update intent %s: %w", id, err)
	}
	return it, nil
}

// ---- WorkOrder Repository (接单派单闭环) ----

type pgWorkOrderRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewWorkOrderRepository() repository.WorkOrderRepository {
	return &pgWorkOrderRepo{pool: s.pool}
}

const workOrderCols = `id, order_no, demand_id, publisher_id, publisher_name, worker_id, worker_name, amount_fen, status, result_photos, rework_note, cancel_reason, created_at, updated_at`

func scanWorkOrder(row interface{ Scan(...any) error }) (domain.WorkOrder, error) {
	var wo domain.WorkOrder
	var photos []byte
	err := row.Scan(&wo.ID, &wo.OrderNo, &wo.DemandID, &wo.PublisherID, &wo.PublisherName,
		&wo.WorkerID, &wo.WorkerName, &wo.AmountFen, &wo.Status, &photos, &wo.ReworkNote, &wo.CancelReason,
		&wo.CreatedAt, &wo.UpdatedAt)
	if err != nil {
		return domain.WorkOrder{}, err
	}
	if len(photos) > 0 {
		if err := json.Unmarshal(photos, &wo.ResultPhotos); err != nil {
			return domain.WorkOrder{}, fmt.Errorf("parse result_photos: %w", err)
		}
	}
	return wo, nil
}

func (r *pgWorkOrderRepo) Create(wo domain.WorkOrder) (domain.WorkOrder, error) {
	now := time.Now()
	wo.CreatedAt = now
	wo.UpdatedAt = now
	photos, _ := json.Marshal(wo.ResultPhotos)
	_, err := r.pool.Exec(context.Background(),
		`INSERT INTO work_orders (`+workOrderCols+`)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		wo.ID, wo.OrderNo, wo.DemandID, wo.PublisherID, wo.PublisherName, wo.WorkerID, wo.WorkerName,
		wo.AmountFen, wo.Status, photos, wo.ReworkNote, wo.CancelReason, wo.CreatedAt, wo.UpdatedAt)
	if err != nil {
		return domain.WorkOrder{}, fmt.Errorf("create work order: %w", err)
	}
	return wo, nil
}

func (r *pgWorkOrderRepo) FindByID(id string) (domain.WorkOrder, error) {
	row := r.pool.QueryRow(context.Background(),
		`SELECT `+workOrderCols+` FROM work_orders WHERE id=$1`, id)
	wo, err := scanWorkOrder(row)
	if err != nil {
		return domain.WorkOrder{}, fmt.Errorf("find work order %s: %w", id, err)
	}
	return wo, nil
}

func (r *pgWorkOrderRepo) ListByPublisher(publisherID string) ([]domain.WorkOrder, error) {
	rows, err := r.pool.Query(context.Background(),
		`SELECT `+workOrderCols+` FROM work_orders WHERE publisher_id=$1 ORDER BY created_at DESC`, publisherID)
	if err != nil {
		return nil, fmt.Errorf("query work orders by publisher: %w", err)
	}
	defer rows.Close()
	out := []domain.WorkOrder{}
	for rows.Next() {
		wo, err := scanWorkOrder(rows)
		if err != nil {
			return nil, fmt.Errorf("scan work order: %w", err)
		}
		out = append(out, wo)
	}
	return out, rows.Err()
}

func (r *pgWorkOrderRepo) ListByWorker(workerID string) ([]domain.WorkOrder, error) {
	rows, err := r.pool.Query(context.Background(),
		`SELECT `+workOrderCols+` FROM work_orders WHERE worker_id=$1 ORDER BY created_at DESC`, workerID)
	if err != nil {
		return nil, fmt.Errorf("query work orders by worker: %w", err)
	}
	defer rows.Close()
	out := []domain.WorkOrder{}
	for rows.Next() {
		wo, err := scanWorkOrder(rows)
		if err != nil {
			return nil, fmt.Errorf("scan work order: %w", err)
		}
		out = append(out, wo)
	}
	return out, rows.Err()
}

func (r *pgWorkOrderRepo) UpdateStatus(id string, status domain.WorkOrderStatus) (domain.WorkOrder, error) {
	var photos []byte
	var wo domain.WorkOrder
	err := r.pool.QueryRow(context.Background(),
		`UPDATE work_orders SET status=$2, updated_at=now() WHERE id=$1
		RETURNING `+workOrderCols, id, status).
		Scan(&wo.ID, &wo.OrderNo, &wo.DemandID, &wo.PublisherID, &wo.PublisherName,
			&wo.WorkerID, &wo.WorkerName, &wo.AmountFen, &wo.Status, &photos, &wo.ReworkNote, &wo.CancelReason,
			&wo.CreatedAt, &wo.UpdatedAt)
	if err != nil {
		return domain.WorkOrder{}, fmt.Errorf("update work order status %s: %w", id, err)
	}
	if len(photos) > 0 {
		if err := json.Unmarshal(photos, &wo.ResultPhotos); err != nil {
			return domain.WorkOrder{}, fmt.Errorf("parse result_photos: %w", err)
		}
	}
	return wo, nil
}

func (r *pgWorkOrderRepo) UpdatePhotos(id string, photos []string) (domain.WorkOrder, error) {
	data, _ := json.Marshal(photos)
	row := r.pool.QueryRow(context.Background(),
		`UPDATE work_orders SET result_photos=$2, updated_at=now() WHERE id=$1
		RETURNING `+workOrderCols, id, data)
	wo, err := scanWorkOrder(row)
	if err != nil {
		return domain.WorkOrder{}, fmt.Errorf("update work order photos %s: %w", id, err)
	}
	return wo, nil
}

func (r *pgWorkOrderRepo) UpdateRework(id string, note string) (domain.WorkOrder, error) {
	row := r.pool.QueryRow(context.Background(),
		`UPDATE work_orders SET rework_note=$2, updated_at=now() WHERE id=$1
		RETURNING `+workOrderCols, id, note)
	wo, err := scanWorkOrder(row)
	if err != nil {
		return domain.WorkOrder{}, fmt.Errorf("update work order rework %s: %w", id, err)
	}
	return wo, nil
}

func (r *pgWorkOrderRepo) UpdateCancel(id string, reason string) (domain.WorkOrder, error) {
	row := r.pool.QueryRow(context.Background(),
		`UPDATE work_orders SET cancel_reason=$2, updated_at=now() WHERE id=$1
		RETURNING `+workOrderCols, id, reason)
	wo, err := scanWorkOrder(row)
	if err != nil {
		return domain.WorkOrder{}, fmt.Errorf("update work order cancel %s: %w", id, err)
	}
	return wo, nil
}

// ---- DemandBid Repository ----

type pgBidRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewBidRepository() repository.BidRepository {
	return &pgBidRepo{pool: s.pool}
}

func (r *pgBidRepo) Create(b domain.DemandBid) (domain.DemandBid, error) {
	now := time.Now()
	b.Version = 1
	b.CreatedAt = now
	b.UpdatedAt = now
	_, err := r.pool.Exec(context.Background(),
		`INSERT INTO demand_bids (id, demand_id, bidder_id, bidder_name, amount_fen, proposal, status, version, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		b.ID, b.DemandID, b.BidderID, b.BidderName, b.AmountFen, b.Proposal, b.Status, b.Version, b.CreatedAt, b.UpdatedAt)
	if err != nil {
		return domain.DemandBid{}, fmt.Errorf("create bid: %w", err)
	}
	return b, nil
}

func (r *pgBidRepo) FindByID(id string) (domain.DemandBid, error) {
	var b domain.DemandBid
	err := r.pool.QueryRow(context.Background(),
		`SELECT id, demand_id, bidder_id, bidder_name, amount_fen, proposal, status, version, created_at, updated_at
		FROM demand_bids WHERE id=$1`, id).
		Scan(&b.ID, &b.DemandID, &b.BidderID, &b.BidderName, &b.AmountFen, &b.Proposal, &b.Status, &b.Version, &b.CreatedAt, &b.UpdatedAt)
	if err != nil {
		return domain.DemandBid{}, fmt.Errorf("bid %s: %w", id, err)
	}
	return b, nil
}

func (r *pgBidRepo) ListByDemand(demandID string) ([]domain.DemandBid, error) {
	rows, err := r.pool.Query(context.Background(),
		`SELECT id, demand_id, bidder_id, bidder_name, amount_fen, proposal, status, version, created_at, updated_at
		FROM demand_bids WHERE demand_id=$1 ORDER BY created_at DESC`, demandID)
	if err != nil {
		return nil, fmt.Errorf("query bids: %w", err)
	}
	defer rows.Close()
	out := []domain.DemandBid{}
	for rows.Next() {
		var b domain.DemandBid
		if err := rows.Scan(&b.ID, &b.DemandID, &b.BidderID, &b.BidderName, &b.AmountFen, &b.Proposal, &b.Status, &b.Version, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan bid: %w", err)
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (r *pgBidRepo) ListByBidder(bidderID string) ([]domain.DemandBid, error) {
	rows, err := r.pool.Query(context.Background(),
		`SELECT id, demand_id, bidder_id, bidder_name, amount_fen, proposal, status, version, created_at, updated_at
		FROM demand_bids WHERE bidder_id=$1 ORDER BY created_at DESC`, bidderID)
	if err != nil {
		return nil, fmt.Errorf("query bids by bidder: %w", err)
	}
	defer rows.Close()
	out := []domain.DemandBid{}
	for rows.Next() {
		var b domain.DemandBid
		if err := rows.Scan(&b.ID, &b.DemandID, &b.BidderID, &b.BidderName, &b.AmountFen, &b.Proposal, &b.Status, &b.Version, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan bid: %w", err)
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (r *pgBidRepo) UpdateStatus(id string, status string) (domain.DemandBid, error) {
	tag, err := r.pool.Exec(context.Background(),
		`UPDATE demand_bids SET status=$1, version=version+1, updated_at=$2 WHERE id=$3`,
		status, time.Now(), id)
	if err != nil {
		return domain.DemandBid{}, fmt.Errorf("update bid status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.DemandBid{}, fmt.Errorf("bid %s not found", id)
	}
	return r.FindByID(id)
}

func (r *contractRepo) FindByID(id string) (domain.Contract, error) {
	var v domain.Contract
	var status string
	err := r.pool.QueryRow(context.Background(),
		`SELECT id, enterprise_id, template_id, sign_url, status, version, created_at, updated_at
		FROM contracts WHERE id=$1`, id).
		Scan(&v.ID, &v.EnterpriseID, &v.TemplateID, &v.SignURL, &status, &v.Version, &v.CreatedAt, &v.UpdatedAt)
	if err != nil {
		return domain.Contract{}, fmt.Errorf("contract %s: %w", id, err)
	}
	v.Status = domain.ContractStatus(status)
	return v, nil
}

func (r *contractRepo) UpdateStatus(id string, status domain.ContractStatus) (domain.Contract, error) {
	tag, err := r.pool.Exec(context.Background(),
		`UPDATE contracts SET status=$1, version=version+1, updated_at=$2 WHERE id=$3`,
		string(status), time.Now(), id)
	if err != nil {
		return domain.Contract{}, fmt.Errorf("update contract status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.Contract{}, fmt.Errorf("contract %s not found", id)
	}
	return r.FindByID(id)
}

// scanEmploymentPaged queries employment_requests with pagination.
// The where clause must be a compile-time constant — never pass user input as where.
func scanEmploymentPaged(pool *pgxpool.Pool, where string, offset, limit int, args ...any) ([]domain.EmploymentRequest, int, error) {
	countQ := `SELECT COUNT(*) FROM employment_requests ` + where
	var total int
	if err := pool.QueryRow(context.Background(), countQ, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count employment: %w", err)
	}

	q := `SELECT id, enterprise_id, position, headcount, start_date, end_date, status, version, created_at, updated_at
		FROM employment_requests ` + where + ` ORDER BY created_at DESC LIMIT $` + fmt.Sprintf("%d", len(args)+1) + ` OFFSET $` + fmt.Sprintf("%d", len(args)+2)
	allArgs := append(args, limit, offset)
	rows, err := pool.Query(context.Background(), q, allArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("query employment: %w", err)
	}
	defer rows.Close()
	out := []domain.EmploymentRequest{}
	for rows.Next() {
		var v domain.EmploymentRequest
		var status string
		if err := rows.Scan(&v.ID, &v.EnterpriseID, &v.Position, &v.Headcount,
			&v.StartDate, &v.EndDate, &status, &v.Version, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan employment: %w", err)
		}
		v.Status = domain.EmploymentStatus(status)
		out = append(out, v)
	}
	return out, total, rows.Err()
}

// scanContractsPaged queries contracts with pagination.
// The where clause must be a compile-time constant — never pass user input as where.
func scanContractsPaged(pool *pgxpool.Pool, where string, offset, limit int, args ...any) ([]domain.Contract, int, error) {
	countQ := `SELECT COUNT(*) FROM contracts ` + where
	var total int
	if err := pool.QueryRow(context.Background(), countQ, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count contracts: %w", err)
	}

	q := `SELECT id, enterprise_id, template_id, sign_url, status, version, created_at, updated_at
		FROM contracts ` + where + ` ORDER BY created_at DESC LIMIT $` + fmt.Sprintf("%d", len(args)+1) + ` OFFSET $` + fmt.Sprintf("%d", len(args)+2)
	allArgs := append(args, limit, offset)
	rows, err := pool.Query(context.Background(), q, allArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("query contracts: %w", err)
	}
	defer rows.Close()
	out := []domain.Contract{}
	for rows.Next() {
		var v domain.Contract
		var status string
		if err := rows.Scan(&v.ID, &v.EnterpriseID, &v.TemplateID, &v.SignURL,
			&status, &v.Version, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan contract: %w", err)
		}
		v.Status = domain.ContractStatus(status)
		out = append(out, v)
	}
	return out, total, rows.Err()
}

func (r *demandRepo) CompareAndSetStatus(id string, oldStatus, newStatus domain.DemandStatus) (bool, domain.Demand, error) {
	tag, err := r.pool.Exec(context.Background(),
		`UPDATE demands SET status=$1, updated_at=$2, version=version+1 WHERE id=$3 AND status=$4`,
		string(newStatus), time.Now(), id, string(oldStatus))
	if err != nil {
		return false, domain.Demand{}, fmt.Errorf("compare-and-set demand status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		// Either demand not found, or status didn't match. Check which.
		d, findErr := r.FindByID(id)
		if findErr != nil {
			return false, domain.Demand{}, findErr
		}
		return false, d, nil
	}
	d, err := r.FindByID(id)
	return true, d, err
}

func (r *demandRepo) Delete(id string) error {
	tag, err := r.pool.Exec(context.Background(), `DELETE FROM demands WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("delete demand %s: %w", id, err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("demand %s not found", id)
	}
	return nil
}

// ---- College Repository ----

type pgCollegeRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewCollegeRepository() repository.CollegeRepository {
	return &pgCollegeRepo{pool: s.Pool()}
}

// collegeCols 与 colleges 表列一一对应（迁移 000044 补齐小程序页面字段）
const collegeCols = `id,name,region,city,description,logo_url,status,coop_type,majors,facilities,tags,short_name,level_tags,specialties,major_count,partner_count,teacher_count,student_count,graduate_rate,partners,cover,photos,phone,website,intro,majors_detail,created_at,updated_at`

func (r *pgCollegeRepo) Create(c domain.College) (domain.College, error) {
	c.CreatedAt = time.Now()
	c.UpdatedAt = c.CreatedAt
	c.Majors = jsonbSlice(c.Majors)
	c.Facilities = jsonbSlice(c.Facilities)
	c.Tags = jsonbSlice(c.Tags)
	c.Specialties = jsonbSlice(c.Specialties)
	c.Partners = jsonbSlice(c.Partners)
	c.Photos = jsonbSlice(c.Photos)
	c.MajorsDetail = jsonbSlice(c.MajorsDetail)
	_, err := r.pool.Exec(context.Background(),
		`INSERT INTO colleges (`+collegeCols+`) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27,$28)`,
		c.ID, c.Name, c.Region, c.City, c.Description, c.LogoURL, c.Status, c.CoopType, c.Majors, c.Facilities,
		c.Tags, c.ShortName, c.LevelTags, c.Specialties, c.MajorCount, c.PartnerCount, c.TeacherCount, c.StudentCount,
		c.GraduateRate, c.Partners, c.CoverURL, c.Photos, c.Phone, c.Website, c.Intro, c.MajorsDetail, c.CreatedAt, c.UpdatedAt)
	return c, err
}

func (r *pgCollegeRepo) FindByID(id string) (domain.College, error) {
	var c domain.College
	err := r.pool.QueryRow(context.Background(), `SELECT `+collegeCols+` FROM colleges WHERE id=$1`, id).
		Scan(&c.ID, &c.Name, &c.Region, &c.City, &c.Description, &c.LogoURL, &c.Status, &c.CoopType, &c.Majors, &c.Facilities,
			&c.Tags, &c.ShortName, &c.LevelTags, &c.Specialties, &c.MajorCount, &c.PartnerCount, &c.TeacherCount, &c.StudentCount,
			&c.GraduateRate, &c.Partners, &c.CoverURL, &c.Photos, &c.Phone, &c.Website, &c.Intro, &c.MajorsDetail, &c.CreatedAt, &c.UpdatedAt)
	return c, err
}

func (r *pgCollegeRepo) List(region string) ([]domain.College, error) {
	q := `SELECT ` + collegeCols + ` FROM colleges`
	args := []any{}
	if region != "" {
		q += ` WHERE region=$1`
		args = append(args, region)
	}
	q += ` ORDER BY created_at DESC`
	rows, err := r.pool.Query(context.Background(), q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.College
	for rows.Next() {
		var c domain.College
		if err := rows.Scan(&c.ID, &c.Name, &c.Region, &c.City, &c.Description, &c.LogoURL, &c.Status, &c.CoopType, &c.Majors, &c.Facilities,
			&c.Tags, &c.ShortName, &c.LevelTags, &c.Specialties, &c.MajorCount, &c.PartnerCount, &c.TeacherCount, &c.StudentCount,
			&c.GraduateRate, &c.Partners, &c.CoverURL, &c.Photos, &c.Phone, &c.Website, &c.Intro, &c.MajorsDetail, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *pgCollegeRepo) Update(c domain.College) (domain.College, error) {
	c.UpdatedAt = time.Now()
	c.Majors = jsonbSlice(c.Majors)
	c.Facilities = jsonbSlice(c.Facilities)
	c.Tags = jsonbSlice(c.Tags)
	c.Specialties = jsonbSlice(c.Specialties)
	c.Partners = jsonbSlice(c.Partners)
	c.Photos = jsonbSlice(c.Photos)
	c.MajorsDetail = jsonbSlice(c.MajorsDetail)
	_, err := r.pool.Exec(context.Background(),
		`UPDATE colleges SET name=$1,region=$2,city=$3,description=$4,logo_url=$5,status=$6,coop_type=$7,majors=$8,facilities=$9,tags=$10,short_name=$11,level_tags=$12,specialties=$13,major_count=$14,partner_count=$15,teacher_count=$16,student_count=$17,graduate_rate=$18,partners=$19,cover=$20,photos=$21,phone=$22,website=$23,intro=$24,majors_detail=$25,updated_at=$26 WHERE id=$27`,
		c.Name, c.Region, c.City, c.Description, c.LogoURL, c.Status, c.CoopType, c.Majors, c.Facilities,
		c.Tags, c.ShortName, c.LevelTags, c.Specialties, c.MajorCount, c.PartnerCount, c.TeacherCount, c.StudentCount,
		c.GraduateRate, c.Partners, c.CoverURL, c.Photos, c.Phone, c.Website, c.Intro, c.MajorsDetail, c.UpdatedAt, c.ID)
	return c, err
}

func (r *pgCollegeRepo) Delete(id string) error {
	_, err := r.pool.Exec(context.Background(), `DELETE FROM colleges WHERE id=$1`, id)
	return err
}

// ---- StudyTour Repository ----

type pgStudyTourRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewStudyTourRepository() repository.StudyTourRepository {
	return &pgStudyTourRepo{pool: s.Pool()}
}

func (r *pgStudyTourRepo) Create(st domain.StudyTour) (domain.StudyTour, error) {
	st.CreatedAt = time.Now()
	st.UpdatedAt = st.CreatedAt
	scheduleJSON, err := json.Marshal(jsonbSlice(st.Schedule))
	if err != nil {
		return domain.StudyTour{}, fmt.Errorf("marshal schedule: %w", err)
	}
	_, err = r.pool.Exec(context.Background(),
		`INSERT INTO study_tours (id,title,destination,duration,capacity,status,description,location,organizer_id,start_date,end_date,cover_image,price_fen,schedule,created_at,updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`,
		st.ID, st.Title, st.Destination, st.Duration, st.Capacity, st.Status,
		st.Description, st.Location, st.OrganizerID, st.StartDate, st.EndDate,
		st.CoverImage, st.PriceFen, scheduleJSON, st.CreatedAt, st.UpdatedAt)
	return st, err
}

func (r *pgStudyTourRepo) FindByID(id string) (domain.StudyTour, error) {
	var s domain.StudyTour
	var schedule []byte
	err := r.pool.QueryRow(context.Background(), `SELECT id,title,destination,duration,capacity,status,description,location,organizer_id,start_date,end_date,cover_image,price_fen,schedule,created_at,updated_at FROM study_tours WHERE id=$1`, id).
		Scan(&s.ID, &s.Title, &s.Destination, &s.Duration, &s.Capacity, &s.Status,
			&s.Description, &s.Location, &s.OrganizerID, &s.StartDate, &s.EndDate,
			&s.CoverImage, &s.PriceFen, &schedule, &s.CreatedAt, &s.UpdatedAt)
	if schedule != nil {
		json.Unmarshal(schedule, &s.Schedule)
	}
	return s, err
}

func (r *pgStudyTourRepo) List() ([]domain.StudyTour, error) {
	rows, err := r.pool.Query(context.Background(), `SELECT id,title,destination,duration,capacity,status,description,location,organizer_id,start_date,end_date,cover_image,price_fen,schedule,created_at,updated_at FROM study_tours ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.StudyTour
	for rows.Next() {
		var s domain.StudyTour
		var schedule []byte
		if err := rows.Scan(&s.ID, &s.Title, &s.Destination, &s.Duration, &s.Capacity, &s.Status,
			&s.Description, &s.Location, &s.OrganizerID, &s.StartDate, &s.EndDate,
			&s.CoverImage, &s.PriceFen, &schedule, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		if schedule != nil {
			json.Unmarshal(schedule, &s.Schedule)
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *pgStudyTourRepo) Update(s domain.StudyTour) (domain.StudyTour, error) {
	s.UpdatedAt = time.Now()
	scheduleJSON, err := json.Marshal(jsonbSlice(s.Schedule))
	if err != nil {
		return domain.StudyTour{}, fmt.Errorf("marshal schedule: %w", err)
	}
	_, err = r.pool.Exec(context.Background(),
		`UPDATE study_tours SET title=$1,destination=$2,duration=$3,capacity=$4,status=$5,description=$6,location=$7,organizer_id=$8,start_date=$9,end_date=$10,cover_image=$11,price_fen=$12,schedule=$13,updated_at=$14 WHERE id=$15`,
		s.Title, s.Destination, s.Duration, s.Capacity, s.Status,
		s.Description, s.Location, s.OrganizerID, s.StartDate, s.EndDate,
		s.CoverImage, s.PriceFen, scheduleJSON, s.UpdatedAt, s.ID)
	return s, err
}

func (r *pgStudyTourRepo) Delete(id string) error {
	_, err := r.pool.Exec(context.Background(), `DELETE FROM study_tours WHERE id=$1`, id)
	return err
}

// ---- Exhibition Repository ----

type pgExhibitionRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewExhibitionRepository() repository.ExhibitionRepository {
	return &pgExhibitionRepo{pool: s.Pool()}
}

func (r *pgExhibitionRepo) Create(e domain.Exhibition) (domain.Exhibition, error) {
	e.CreatedAt = time.Now()
	e.UpdatedAt = e.CreatedAt
	_, err := r.pool.Exec(context.Background(),
		"INSERT INTO exhibitions (id,title,category,description,location,start_date,end_date,booth_count,booth_price_fen,organizer,status,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)",
		e.ID, e.Title, e.Category, e.Description, e.Location, e.StartDate, e.EndDate, e.BoothCount, e.BoothPrice, e.Organizer, e.Status, e.CreatedAt, e.UpdatedAt)
	return e, err
}
func (r *pgExhibitionRepo) FindByID(id string) (domain.Exhibition, error) {
	var e domain.Exhibition
	err := r.pool.QueryRow(context.Background(), "SELECT id,title,category,description,location,start_date,end_date,booth_count,booth_price_fen,organizer,status,created_at,updated_at FROM exhibitions WHERE id=$1", id).
		Scan(&e.ID, &e.Title, &e.Category, &e.Description, &e.Location, &e.StartDate, &e.EndDate, &e.BoothCount, &e.BoothPrice, &e.Organizer, &e.Status, &e.CreatedAt, &e.UpdatedAt)
	return e, err
}
func (r *pgExhibitionRepo) List(offset, limit int) ([]domain.Exhibition, int, error) {
	var total int
	if err := r.pool.QueryRow(context.Background(), "SELECT count(*) FROM exhibitions").Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count exhibitions: %w", err)
	}
	rows, err := r.pool.Query(context.Background(), "SELECT id,title,category,description,location,start_date,end_date,booth_count,booth_price_fen,organizer,status,created_at,updated_at FROM exhibitions ORDER BY created_at DESC LIMIT $1 OFFSET $2", limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []domain.Exhibition
	for rows.Next() {
		var e domain.Exhibition
		if err := rows.Scan(&e.ID, &e.Title, &e.Category, &e.Description, &e.Location, &e.StartDate, &e.EndDate, &e.BoothCount, &e.BoothPrice, &e.Organizer, &e.Status, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, e)
	}
	return out, total, rows.Err()
}
func (r *pgExhibitionRepo) Update(e domain.Exhibition) (domain.Exhibition, error) {
	e.UpdatedAt = time.Now()
	_, err := r.pool.Exec(context.Background(),
		"UPDATE exhibitions SET title=$1,category=$2,description=$3,location=$4,start_date=$5,end_date=$6,booth_count=$7,booth_price_fen=$8,organizer=$9,status=$10,updated_at=$11 WHERE id=$12",
		e.Title, e.Category, e.Description, e.Location, e.StartDate, e.EndDate, e.BoothCount, e.BoothPrice, e.Organizer, e.Status, e.UpdatedAt, e.ID)
	return e, err
}
func (r *pgExhibitionRepo) Delete(id string) error {
	_, err := r.pool.Exec(context.Background(), "DELETE FROM exhibitions WHERE id=$1", id)
	return err
}
func (r *pgExhibitionRepo) CreateBooth(b domain.ExhibitionBooth) (domain.ExhibitionBooth, error) {
	b.CreatedAt = time.Now()
	_, err := r.pool.Exec(context.Background(),
		`INSERT INTO exhibition_booths (id,exhibition_id,exhibitor_id,booth_number,exhibit_name,exhibit_desc,status,created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		b.ID, b.ExhibitionID, b.ExhibitorID, b.BoothNumber, b.ExhibitName, b.ExhibitDesc, b.Status, b.CreatedAt)
	return b, err
}
func (r *pgExhibitionRepo) ListBooths(exhibitionID string) ([]domain.ExhibitionBooth, error) {
	rows, err := r.pool.Query(context.Background(),
		`SELECT id,exhibition_id,exhibitor_id,booth_number,exhibit_name,exhibit_desc,status,created_at FROM exhibition_booths WHERE exhibition_id=$1 ORDER BY created_at DESC`, exhibitionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.ExhibitionBooth
	for rows.Next() {
		var b domain.ExhibitionBooth
		if err := rows.Scan(&b.ID, &b.ExhibitionID, &b.ExhibitorID, &b.BoothNumber, &b.ExhibitName, &b.ExhibitDesc, &b.Status, &b.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}
func (r *pgExhibitionRepo) UpdateBoothStatus(id, status string) (domain.ExhibitionBooth, error) {
	var b domain.ExhibitionBooth
	err := r.pool.QueryRow(context.Background(),
		`UPDATE exhibition_booths SET status=$1 WHERE id=$2 RETURNING id,exhibition_id,exhibitor_id,booth_number,exhibit_name,exhibit_desc,status,created_at`,
		status, id).
		Scan(&b.ID, &b.ExhibitionID, &b.ExhibitorID, &b.BoothNumber, &b.ExhibitName, &b.ExhibitDesc, &b.Status, &b.CreatedAt)
	return b, err
}

// ── TestSite PG ──

type pgTestSiteRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewTestSiteRepository() repository.TestSiteRepository {
	return &pgTestSiteRepo{pool: s.Pool()}
}

func (r *pgTestSiteRepo) Create(t domain.TestSite) (domain.TestSite, error) {
	t.CreatedAt = time.Now()
	t.UpdatedAt = t.CreatedAt
	facJSON, _ := json.Marshal(t.Facilities)
	_, err := r.pool.Exec(context.Background(),
		`INSERT INTO test_sites (id,name,site_type,owner_id,location,booking_rule,status,price_fen,facilities,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		t.ID, t.Name, t.SiteType, t.OwnerID, t.Location, t.BookingRule, t.Status, t.PriceFen, facJSON, t.CreatedAt, t.UpdatedAt)
	return t, err
}

func (r *pgTestSiteRepo) FindByID(id string) (domain.TestSite, error) {
	var t domain.TestSite
	var fj []byte
	err := r.pool.QueryRow(context.Background(),
		`SELECT id,name,site_type,owner_id,location,booking_rule,status,price_fen,facilities,created_at,updated_at FROM test_sites WHERE id=$1`, id).
		Scan(&t.ID, &t.Name, &t.SiteType, &t.OwnerID, &t.Location, &t.BookingRule, &t.Status, &t.PriceFen, &fj, &t.CreatedAt, &t.UpdatedAt)
	if err == nil {
		json.Unmarshal(fj, &t.Facilities)
	}
	return t, err
}

func (r *pgTestSiteRepo) List(siteType string) ([]domain.TestSite, error) {
	q := `SELECT id,name,site_type,owner_id,location,booking_rule,status,price_fen,facilities,created_at,updated_at FROM test_sites`
	args := []any{}
	if siteType != "" {
		q += ` WHERE site_type=$1`
		args = append(args, siteType)
	}
	q += ` ORDER BY created_at DESC`
	rows, err := r.pool.Query(context.Background(), q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.TestSite
	for rows.Next() {
		var t domain.TestSite
		var fj []byte
		if err := rows.Scan(&t.ID, &t.Name, &t.SiteType, &t.OwnerID, &t.Location, &t.BookingRule, &t.Status, &t.PriceFen, &fj, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		json.Unmarshal(fj, &t.Facilities)
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *pgTestSiteRepo) UpdateSite(t domain.TestSite) (domain.TestSite, error) {
	t.UpdatedAt = time.Now()
	facJSON, _ := json.Marshal(t.Facilities)
	_, err := r.pool.Exec(context.Background(),
		`UPDATE test_sites SET name=$1,site_type=$2,location=$3,booking_rule=$4,status=$5,price_fen=$6,facilities=$7,updated_at=$8 WHERE id=$9`,
		t.Name, t.SiteType, t.Location, t.BookingRule, t.Status, t.PriceFen, facJSON, t.UpdatedAt, t.ID)
	return t, err
}

func (r *pgTestSiteRepo) DeleteSite(id string) error {
	_, err := r.pool.Exec(context.Background(), `DELETE FROM test_sites WHERE id=$1`, id)
	return err
}

func (r *pgTestSiteRepo) CreateBooking(b domain.TestSiteBooking) (domain.TestSiteBooking, error) {
	b.CreatedAt = time.Now()
	_, err := r.pool.Exec(context.Background(),
		`INSERT INTO test_site_bookings (id,site_id,user_id,purpose,start_time,end_time,contact_name,contact_phone,status,review_note,created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		b.ID, b.SiteID, b.UserID, b.Purpose, b.StartTime, b.EndTime, b.ContactName, b.ContactPhone, b.Status, b.ReviewNote, b.CreatedAt)
	return b, err
}
func (r *pgTestSiteRepo) UpdateBookingStatus(id, status, note string) (domain.TestSiteBooking, error) {
	var b domain.TestSiteBooking
	err := r.pool.QueryRow(context.Background(),
		`UPDATE test_site_bookings SET status=$1,review_note=$2 WHERE id=$3 RETURNING id,site_id,user_id,purpose,start_time,end_time,contact_name,contact_phone,status,review_note,created_at`,
		status, note, id).
		Scan(&b.ID, &b.SiteID, &b.UserID, &b.Purpose, &b.StartTime, &b.EndTime, &b.ContactName, &b.ContactPhone, &b.Status, &b.ReviewNote, &b.CreatedAt)
	return b, err
}
func (r *pgTestSiteRepo) ListBookings(siteID string) ([]domain.TestSiteBooking, error) {
	rows, err := r.pool.Query(context.Background(),
		`SELECT id,site_id,user_id,purpose,start_time,end_time,contact_name,contact_phone,status,review_note,created_at FROM test_site_bookings WHERE site_id=$1 ORDER BY created_at DESC`, siteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.TestSiteBooking
	for rows.Next() {
		var b domain.TestSiteBooking
		if err := rows.Scan(&b.ID, &b.SiteID, &b.UserID, &b.Purpose, &b.StartTime, &b.EndTime, &b.ContactName, &b.ContactPhone, &b.Status, &b.ReviewNote, &b.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

// ListBookingsByUser 我的预约：按用户返回全部预约（最新在前）
func (r *pgTestSiteRepo) ListBookingsByUser(userID string) ([]domain.TestSiteBooking, error) {
	rows, err := r.pool.Query(context.Background(),
		`SELECT id,site_id,user_id,purpose,start_time,end_time,contact_name,contact_phone,status,review_note,created_at FROM test_site_bookings WHERE user_id=$1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("list bookings by user: %w", err)
	}
	defer rows.Close()
	var out []domain.TestSiteBooking
	for rows.Next() {
		var b domain.TestSiteBooking
		if err := rows.Scan(&b.ID, &b.SiteID, &b.UserID, &b.Purpose, &b.StartTime, &b.EndTime, &b.ContactName, &b.ContactPhone, &b.Status, &b.ReviewNote, &b.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}
func (r *pgTestSiteRepo) ListAllBookings(offset, limit int) ([]domain.TestSiteBooking, int, error) {
	var total int
	if err := r.pool.QueryRow(context.Background(), `SELECT count(*) FROM test_site_bookings`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count test site bookings: %w", err)
	}
	rows, err := r.pool.Query(context.Background(),
		`SELECT id,site_id,user_id,purpose,start_time,end_time,contact_name,contact_phone,status,review_note,created_at FROM test_site_bookings ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list all bookings: %w", err)
	}
	defer rows.Close()
	var out []domain.TestSiteBooking
	for rows.Next() {
		var b domain.TestSiteBooking
		if err := rows.Scan(&b.ID, &b.SiteID, &b.UserID, &b.Purpose, &b.StartTime, &b.EndTime, &b.ContactName, &b.ContactPhone, &b.Status, &b.ReviewNote, &b.CreatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, b)
	}
	return out, total, rows.Err()
}

// ── Transformation PG ──

type pgTransRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewTransformationRepository() repository.TransformationRepository {
	return &pgTransRepo{pool: s.Pool()}
}

func (r *pgTransRepo) Create(t domain.Transformation) (domain.Transformation, error) {
	t.CreatedAt = time.Now()
	t.UpdatedAt = t.CreatedAt
	_, err := r.pool.Exec(context.Background(),
		`INSERT INTO transformations (id,title,achievement_id,owner_id,progress,partner_id,status,stage,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		t.ID, t.Title, t.AchievementID, t.OwnerID, t.Progress, t.PartnerID, t.Status, t.Stage, t.CreatedAt, t.UpdatedAt)
	return t, err
}

func (r *pgTransRepo) FindByID(id string) (domain.Transformation, error) {
	var t domain.Transformation
	err := r.pool.QueryRow(context.Background(),
		`SELECT id,title,achievement_id,owner_id,progress,partner_id,status,stage,created_at,updated_at FROM transformations WHERE id=$1`, id).
		Scan(&t.ID, &t.Title, &t.AchievementID, &t.OwnerID, &t.Progress, &t.PartnerID, &t.Status, &t.Stage, &t.CreatedAt, &t.UpdatedAt)
	return t, err
}

func (r *pgTransRepo) List(ownerID string) ([]domain.Transformation, error) {
	q := `SELECT id,title,achievement_id,owner_id,progress,partner_id,status,stage,created_at,updated_at FROM transformations`
	args := []any{}
	if ownerID != "" {
		q += ` WHERE owner_id=$1`
		args = append(args, ownerID)
	}
	q += ` ORDER BY created_at DESC`
	rows, err := r.pool.Query(context.Background(), q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Transformation
	for rows.Next() {
		var t domain.Transformation
		if err := rows.Scan(&t.ID, &t.Title, &t.AchievementID, &t.OwnerID, &t.Progress, &t.PartnerID, &t.Status, &t.Stage, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *pgTransRepo) Update(t domain.Transformation) (domain.Transformation, error) {
	t.UpdatedAt = time.Now()
	_, err := r.pool.Exec(context.Background(),
		`UPDATE transformations SET title=$1,achievement_id=$2,progress=$3,partner_id=$4,status=$5,stage=$6,updated_at=$7 WHERE id=$8`,
		t.Title, t.AchievementID, t.Progress, t.PartnerID, t.Status, t.Stage, t.UpdatedAt, t.ID)
	return t, err
}

func (r *pgTransRepo) Delete(id string) error {
	_, err := r.pool.Exec(context.Background(), `DELETE FROM transformations WHERE id=$1`, id)
	return err
}
