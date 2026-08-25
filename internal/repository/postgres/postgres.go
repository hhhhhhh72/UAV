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
	cfg.MinConns = 5 // C 批：常驻最小连接，避免突发流量下冷启动建连延迟
	// P1 修复（批3）：查询级超时 + 连接生命周期管理——
	// 此前无 statement_timeout，慢查询/锁等待会无限挂起并耗尽连接池；
	// 连接老化无回收导致空闲连接被 PG 回收后复用报错。
	cfg.ConnConfig.RuntimeParams["statement_timeout"] = "10000" // 单条 SQL 上限 10s
	cfg.MaxConnLifetime = 30 * time.Minute
	cfg.MaxConnIdleTime = 5 * time.Minute
	cfg.HealthCheckPeriod = 30 * time.Second
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

func (r *demandRepo) Create(ctx context.Context, d domain.Demand) (domain.Demand, error) {
	images, err := json.Marshal(d.Images)
	if err != nil {
		return domain.Demand{}, fmt.Errorf("marshal images: %w", err)
	}
	bizFields, err := json.Marshal(d.BizFields)
	if err != nil {
		return domain.Demand{}, fmt.Errorf("marshal bizFields: %w", err)
	}
	// 只对落库值加密：响应/业务流转保留明文（与 Update 的 encContact 一致），
	// 此前就地加密后 return d 会把密文 base64 回传给创建接口前端。
	encContact := d.Contact
	if r.cipher != nil && d.Contact != "" {
		enc, err := r.cipher.Encrypt(d.Contact)
		if err != nil {
			return domain.Demand{}, fmt.Errorf("encrypt contact: %w", err)
		}
		encContact = enc
	}
	now := time.Now()
	d.Version = 1
	d.CreatedAt = now
	d.UpdatedAt = now
	_, err = r.pool.Exec(ctx, `
		INSERT INTO demands (id, publisher_id, publisher_name, contact, district, city_code,
			biz_type, title, description, images, latitude, longitude, budget_fen, offline_amount_fen, biz_fields,
			status, version, created_at, updated_at, deadline)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20)`,
		d.ID, d.PublisherID, d.PublisherName, encContact, d.District, d.CityCode,
		string(d.BizType), d.Title, d.Description, images, d.Latitude, d.Longitude,
		d.BudgetFen, d.OfflineAmountFen, bizFields, string(d.Status), d.Version, d.CreatedAt, d.UpdatedAt, d.Deadline)
	if err != nil {
		return domain.Demand{}, fmt.Errorf("create demand: %w", err)
	}
	return d, nil
}

func (r *demandRepo) FindByID(ctx context.Context, id string) (domain.Demand, error) {
	q := `SELECT id, publisher_id, publisher_name, contact, district, city_code,
		biz_type, title, description, images, latitude, longitude, budget_fen, offline_amount_fen, biz_fields,
		status, version, created_at, updated_at, deadline
		FROM demands WHERE id = $1`
	demands, err := scanDemands(ctx, r.pool, r.cipher, q, []any{id})
	if err != nil {
		return domain.Demand{}, err
	}
	if len(demands) == 0 {
		return domain.Demand{}, fmt.Errorf("demand %s not found", id)
	}
	return demands[0], nil
}

func (r *demandRepo) Update(ctx context.Context, d domain.Demand) (domain.Demand, error) {
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
	oldVersion := d.Version // 乐观锁：WHERE version=$旧值，并发编辑时后写方失败
	d.Version++
	tag, err := r.pool.Exec(ctx,
		`UPDATE demands SET publisher_name=$1, contact=$2, district=$3, city_code=$4,
		biz_type=$5, title=$6, description=$7, images=$8, latitude=$9, longitude=$10,
		budget_fen=$11, offline_amount_fen=$12, biz_fields=$13, status=$14, deadline=$15,
		version=$16, updated_at=$17
		WHERE id=$18 AND version=$19`,
		d.PublisherName, encContact, d.District, d.CityCode,
		string(d.BizType), d.Title, d.Description, images, d.Latitude, d.Longitude,
		d.BudgetFen, d.OfflineAmountFen, bizFields, string(d.Status), d.Deadline, d.Version, d.UpdatedAt, d.ID, oldVersion)
	if err != nil {
		return domain.Demand{}, fmt.Errorf("update demand %s: %w", d.ID, err)
	}
	if tag.RowsAffected() == 0 {
		return domain.Demand{}, fmt.Errorf("demand %s 已被他人修改，请刷新后重试", d.ID)
	}
	return d, nil
}

func (r *demandRepo) List(ctx context.Context, f repository.DemandFilter) ([]domain.Demand, error) {
	q := `SELECT id, publisher_id, publisher_name, contact, district, city_code,
		biz_type, title, description, images, latitude, longitude, budget_fen, offline_amount_fen, biz_fields,
		status, version, created_at, updated_at, deadline
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
	return scanDemands(ctx, r.pool, r.cipher, q, args)
}

// ListAll 管理端全量（含待审核等全部状态），status 过滤由 f.Status 控制。
func (r *demandRepo) ListAll(ctx context.Context, f repository.DemandFilter) ([]domain.Demand, error) {
	q := `SELECT id, publisher_id, publisher_name, contact, district, city_code,
		biz_type, title, description, images, latitude, longitude, budget_fen, offline_amount_fen, biz_fields,
		status, version, created_at, updated_at, deadline
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
	return scanDemands(ctx, r.pool, r.cipher, q, args)
}

// ListTop 公开语义（仅已发布）按 created_at 倒序取前 limit 条——SQL 端 LIMIT，
// 首页 Top-N 不再整表拉取。
func (r *demandRepo) ListTop(ctx context.Context, f repository.DemandFilter, limit int) ([]domain.Demand, error) {
	q := `SELECT id, publisher_id, publisher_name, contact, district, city_code,
		biz_type, title, description, images, latitude, longitude, budget_fen, offline_amount_fen, biz_fields,
		status, version, created_at, updated_at, deadline
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
	q += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d", argIdx)
	args = append(args, limit)
	return scanDemands(ctx, r.pool, r.cipher, q, args)
}

// Count 按 filter 统计条数（ListAll 语义：status 为空统计全部；首页公开计数传
// Status=published）。聚合查询不物化行。
func (r *demandRepo) Count(ctx context.Context, f repository.DemandFilter) (int, error) {
	q := `SELECT count(*) FROM demands`
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
	var n int
	if err := r.pool.QueryRow(ctx, q, args...).Scan(&n); err != nil {
		return 0, fmt.Errorf("count demands: %w", err)
	}
	return n, nil
}

// escapeLike 转义 LIKE 通配符（% _ \），配合 ESCAPE '\' 使用，防止用户输入被当作通配符。
func escapeLike(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}

func (r *demandRepo) Search(ctx context.Context, q string) ([]domain.Demand, error) {
	if len(q) > 100 {
		q = q[:100]
	}
	kw := "%" + escapeLike(q) + "%"
	sql := `SELECT id, publisher_id, publisher_name, contact, district, city_code,
		biz_type, title, description, images, latitude, longitude, budget_fen, offline_amount_fen, biz_fields,
		status, version, created_at, updated_at, deadline
		FROM demands WHERE status = 'published'
		AND (title ILIKE $1 ESCAPE '\' OR publisher_name ILIKE $1 ESCAPE '\' OR description ILIKE $1 ESCAPE '\' OR district ILIKE $1 ESCAPE '\' OR city_code ILIKE $1 ESCAPE '\')
		ORDER BY created_at DESC LIMIT 50`
	return scanDemands(ctx, r.pool, r.cipher, sql, []any{kw})
}

// ListByPublisher 返回某发布者的全部需求（全状态），供"我的"页统计/查询。
func (r *demandRepo) ListByPublisher(ctx context.Context, publisherID string) ([]domain.Demand, error) {
	q := `SELECT id, publisher_id, publisher_name, contact, district, city_code,
		biz_type, title, description, images, latitude, longitude, budget_fen, offline_amount_fen, biz_fields,
		status, version, created_at, updated_at, deadline
		FROM demands WHERE publisher_id = $1
		ORDER BY created_at DESC`
	return scanDemands(ctx, r.pool, r.cipher, q, []any{publisherID})
}

func (r *demandRepo) SetStatus(ctx context.Context, id string, status domain.DemandStatus) (domain.Demand, error) {
	tag, err := r.pool.Exec(ctx,
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
	err = r.pool.QueryRow(ctx,
		`SELECT id, publisher_id, publisher_name, contact, district, city_code,
		biz_type, title, description, images, latitude, longitude, budget_fen, offline_amount_fen, biz_fields,
		status, version, created_at, updated_at, deadline
		FROM demands WHERE id = $1`, id).Scan(
		&d.ID, &d.PublisherID, &d.PublisherName, &d.Contact, &d.District, &d.CityCode,
		&bizType, &d.Title, &d.Description, &images, &d.Latitude, &d.Longitude,
		&d.BudgetFen, &d.OfflineAmountFen, &bizFields, &status, &d.Version, &d.CreatedAt, &d.UpdatedAt, &d.Deadline)
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
		} else {
			// 解密失败（密钥变更/数据损坏）绝不回传密文——置空而非泄露加密串。
			d.Contact = ""
		}
	}
	return d, nil
}

func scanDemands(ctx context.Context, pool *pgxpool.Pool, cipher *crypto.Cipher, q string, args []any) ([]domain.Demand, error) {
	rows, err := pool.Query(ctx, q, args...)
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
			&d.BudgetFen, &d.OfflineAmountFen, &bizFields, &status, &d.Version, &d.CreatedAt, &d.UpdatedAt, &d.Deadline); err != nil {
			return nil, fmt.Errorf("scan demand: %w", err)
		}
		json.Unmarshal(images, &d.Images)
		json.Unmarshal(bizFields, &d.BizFields)
		d.BizType = domain.BizType(bizType)
		d.Status = domain.DemandStatus(status)
		if cipher != nil && d.Contact != "" {
			if dec, err := cipher.Decrypt(d.Contact); err == nil {
				d.Contact = dec
			} else {
				// 解密失败（密钥变更/数据损坏）绝不回传密文——置空而非泄露加密串。
				d.Contact = ""
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

func (r *enterpriseRepo) Pending(ctx context.Context) ([]domain.Enterprise, error) {
	rows, err := r.pool.Query(ctx, `
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

func (r *enterpriseRepo) Create(ctx context.Context, e domain.Enterprise) (domain.Enterprise, error) {
	now := time.Now()
	if e.CreatedAt.IsZero() {
		e.CreatedAt = now
	}
	if e.UpdatedAt.IsZero() {
		e.UpdatedAt = now
	}
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
	_, err := r.pool.Exec(ctx, `
		INSERT INTO enterprises (id, owner_user_id, name, credit_code, legal_person, contact_phone, industry_category, scale, address, description, business_hours, logo, cover_image, license_url, account_name, contact_person, email, founded_at, capability_tags, status, review_comment, is_member, version, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25)`,
		e.ID, e.OwnerUserID, e.Name, e.CreditCode, e.LegalPerson, e.ContactPhone, e.IndustryCategory, e.Scale, e.Address, e.Description, e.BusinessHours, e.Logo, e.CoverImage, e.LicenseURL, e.AccountName, e.ContactPerson, e.Email, e.FoundedAt, e.CapabilityTags, string(e.Status), e.ReviewComment, e.IsMember, e.Version, e.CreatedAt, e.UpdatedAt)
	if err != nil {
		return domain.Enterprise{}, fmt.Errorf("create enterprise: %w", err)
	}
	return e, nil
}

func (r *enterpriseRepo) Update(ctx context.Context, id string, e domain.Enterprise) (domain.Enterprise, error) {
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
	oldVersion := e.Version // 乐观锁：WHERE version=$旧值
	e.Version++
	e.UpdatedAt = time.Now()
	tag, err := r.pool.Exec(ctx, `
		UPDATE enterprises SET name=$1, credit_code=$2, legal_person=$3, contact_phone=$4, industry_category=$5, scale=$6, address=$7, description=$8, business_hours=$9, logo=$10, cover_image=$11, license_url=$12, account_name=$13, contact_person=$14, email=$15, founded_at=$16, capability_tags=$17, status=$18, review_comment=$19, version=$20, updated_at=$21
		WHERE id=$22 AND version=$23`,
		e.Name, e.CreditCode, e.LegalPerson, e.ContactPhone, e.IndustryCategory, e.Scale, e.Address, e.Description, e.BusinessHours, e.Logo, e.CoverImage, e.LicenseURL, e.AccountName, e.ContactPerson, e.Email, e.FoundedAt, e.CapabilityTags, string(e.Status), e.ReviewComment, e.Version, e.UpdatedAt, id, oldVersion)
	if err != nil {
		return domain.Enterprise{}, fmt.Errorf("update enterprise: %w", err)
	}
	if tag.RowsAffected() == 0 {
		// 区分"不存在"与"并发版本冲突"
		var one int
		if err := r.pool.QueryRow(ctx, `SELECT 1 FROM enterprises WHERE id=$1`, id).Scan(&one); err != nil {
			return domain.Enterprise{}, fmt.Errorf("enterprise %s not found", id)
		}
		return domain.Enterprise{}, fmt.Errorf("enterprise %s 已被他人修改，请刷新后重试", id)
	}
	return e, nil
}

func (r *enterpriseRepo) FindByID(ctx context.Context, id string) (domain.Enterprise, error) {
	var e domain.Enterprise
	var status string
	err := r.pool.QueryRow(ctx, `
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

func (r *enterpriseRepo) FindByOwner(ctx context.Context, userID string) ([]domain.Enterprise, error) {
	return scanEnterprises(ctx, r.pool, r.cipher, "WHERE owner_user_id = $1", userID)
}

func (r *enterpriseRepo) ListByStatus(ctx context.Context, status string, offset, limit int) ([]domain.Enterprise, int, error) {
	// 空 status = 全部状态（与 memory 实现语义一致）
	if status == "" {
		var total int
		if err := r.pool.QueryRow(ctx, `SELECT count(*) FROM enterprises`).Scan(&total); err != nil {
			return nil, 0, fmt.Errorf("count enterprises: %w", err)
		}
		items, err := scanEnterprises(ctx, r.pool, r.cipher, "ORDER BY created_at DESC LIMIT $1 OFFSET $2", limit, offset)
		return items, total, err
	}
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT count(*) FROM enterprises WHERE status=$1`, status).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count enterprises: %w", err)
	}
	items, err := scanEnterprises(ctx, r.pool, r.cipher, "WHERE status=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3", status, limit, offset)
	return items, total, err
}

func scanEnterprises(ctx context.Context, pool *pgxpool.Pool, cipher *crypto.Cipher, where string, args ...any) ([]domain.Enterprise, error) {
	q := `SELECT id, owner_user_id, name, credit_code, legal_person, contact_phone, industry_category, scale, address, description, business_hours, logo, cover_image, license_url, account_name, contact_person, email, founded_at, capability_tags, status, review_comment, is_member, version, created_at, updated_at FROM enterprises ` + where
	rows, err := pool.Query(ctx, q, args...)
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

func (r *enterpriseRepo) AddDocument(ctx context.Context, d domain.EnterpriseDocument) (domain.EnterpriseDocument, error) {
	_, err := r.pool.Exec(ctx, `INSERT INTO enterprise_documents(id,enterprise_id,file_id,document_type,review_status,created_at) VALUES($1,$2,$3,$4,$5,$6)`,
		d.ID, d.EnterpriseID, d.FileID, d.DocumentType, d.ReviewStatus, d.CreatedAt)
	if err != nil {
		return domain.EnterpriseDocument{}, fmt.Errorf("insert enterprise document: %w", err)
	}
	return d, nil
}

func (r *enterpriseRepo) ListDocuments(ctx context.Context, enterpriseID string) ([]domain.EnterpriseDocument, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,enterprise_id,file_id,document_type,review_status,created_at FROM enterprise_documents WHERE enterprise_id=$1 ORDER BY created_at DESC`, enterpriseID)
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

func (r *enterpriseRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM enterprises WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("delete enterprise %s: %w", id, err)
	}
	return nil
}

func (r *enterpriseRepo) Search(ctx context.Context, q string) ([]domain.Enterprise, error) {
	if len(q) > 100 {
		q = q[:100]
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id, owner_user_id, name, license_url, account_name, status, is_member, version, created_at, updated_at
		FROM enterprises WHERE name ILIKE $1 ESCAPE '\' OR address ILIKE $1 ESCAPE '\' OR industry_category ILIKE $1 ESCAPE '\' ORDER BY created_at DESC LIMIT 50`, "%"+escapeLike(q)+"%")
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

func (r *employmentRepo) Create(ctx context.Context, v domain.EmploymentRequest) (domain.EmploymentRequest, error) {
	now := time.Now()
	v.Version = 1
	v.CreatedAt = now
	v.UpdatedAt = now
	_, err := r.pool.Exec(ctx, `
		INSERT INTO employment_requests (id, enterprise_id, position, headcount, start_date, end_date, status, version, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		v.ID, v.EnterpriseID, v.Position, v.Headcount, v.StartDate, v.EndDate,
		string(v.Status), v.Version, v.CreatedAt, v.UpdatedAt)
	if err != nil {
		return domain.EmploymentRequest{}, fmt.Errorf("create employment: %w", err)
	}
	return v, nil
}

func (r *employmentRepo) ListByEnterprise(ctx context.Context, eid string, offset, limit int) ([]domain.EmploymentRequest, int, error) {
	return scanEmploymentPaged(ctx, r.pool, "WHERE enterprise_id = $1", offset, limit, eid)
}

func (r *employmentRepo) ListAll(ctx context.Context, offset, limit int) ([]domain.EmploymentRequest, int, error) {
	return scanEmploymentPaged(ctx, r.pool, "", offset, limit)
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

func (r *contractTplRepo) List(ctx context.Context) ([]domain.ContractTemplate, error) {
	rows, err := r.pool.Query(ctx,
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

func (r *contractTplRepo) Create(ctx context.Context, t domain.ContractTemplate) (domain.ContractTemplate, error) {
	now := time.Now()
	t.CreatedAt = now
	t.UpdatedAt = now
	_, err := r.pool.Exec(ctx,
		`INSERT INTO contract_templates (id,name,version,content,status,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		t.ID, t.Name, t.Version, t.Content, t.Status, t.CreatedAt, t.UpdatedAt)
	if err != nil {
		return domain.ContractTemplate{}, fmt.Errorf("create contract template: %w", err)
	}
	return t, nil
}

func (r *contractRepo) Create(ctx context.Context, v domain.Contract) (domain.Contract, error) {
	now := time.Now()
	v.Version = 1
	v.CreatedAt = now
	v.UpdatedAt = now
	_, err := r.pool.Exec(ctx, `
		INSERT INTO contracts (id, enterprise_id, template_id, sign_url, status, version, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		v.ID, v.EnterpriseID, v.TemplateID, v.SignURL, string(v.Status), v.Version, v.CreatedAt, v.UpdatedAt)
	if err != nil {
		return domain.Contract{}, fmt.Errorf("create contract: %w", err)
	}
	return v, nil
}

func (r *contractRepo) ListByEnterprise(ctx context.Context, eid string, offset, limit int) ([]domain.Contract, int, error) {
	return scanContractsPaged(ctx, r.pool, "WHERE enterprise_id = $1", offset, limit, eid)
}

func (r *contractRepo) ListAll(ctx context.Context, offset, limit int) ([]domain.Contract, int, error) {
	return scanContractsPaged(ctx, r.pool, "", offset, limit)
}

// ---- Job Repository ----

type jobRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewJobRepository() repository.JobRepository { return &jobRepo{pool: s.pool} }

func (r *jobRepo) Create(ctx context.Context, j domain.Job) (domain.Job, error) {
	now := time.Now()
	j.Version = 1
	j.CreatedAt = now
	j.UpdatedAt = now
	_, err := r.pool.Exec(ctx,
		`INSERT INTO jobs (id, enterprise_id, title, description, location, salary_fen, job_type, status, version, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		j.ID, j.EnterpriseID, j.Title, j.Description, j.Location, j.SalaryFen, j.JobType, string(j.Status), j.Version, j.CreatedAt, j.UpdatedAt)
	return j, err
}
func (r *jobRepo) Update(ctx context.Context, id string, j domain.Job) (domain.Job, error) {
	oldVersion := j.Version // 乐观锁：WHERE version=$旧值
	j.Version++
	j.UpdatedAt = time.Now()
	tag, err := r.pool.Exec(ctx,
		`UPDATE jobs SET title=$1,description=$2,location=$3,salary_fen=$4,job_type=$5,status=$6,version=$7,updated_at=$8 WHERE id=$9 AND version=$10`,
		j.Title, j.Description, j.Location, j.SalaryFen, j.JobType, string(j.Status), j.Version, j.UpdatedAt, id, oldVersion)
	if err != nil {
		return domain.Job{}, err
	}
	if tag.RowsAffected() == 0 {
		var one int
		if err := r.pool.QueryRow(ctx, `SELECT 1 FROM jobs WHERE id=$1`, id).Scan(&one); err != nil {
			return domain.Job{}, fmt.Errorf("job %s not found", id)
		}
		return domain.Job{}, fmt.Errorf("job %s 已被他人修改，请刷新后重试", id)
	}
	return j, nil
}
func (r *jobRepo) FindByID(ctx context.Context, id string) (domain.Job, error) {
	var j domain.Job
	var status string
	err := r.pool.QueryRow(ctx,
		`SELECT id, enterprise_id, title, description, location, salary_fen, job_type, status, version, created_at, updated_at FROM jobs WHERE id=$1`, id).
		Scan(&j.ID, &j.EnterpriseID, &j.Title, &j.Description, &j.Location, &j.SalaryFen, &j.JobType, &status, &j.Version, &j.CreatedAt, &j.UpdatedAt)
	j.Status = domain.JobStatus(status)
	return j, err
}
func (r *jobRepo) ListByEnterprise(ctx context.Context, eid string) ([]domain.Job, error) {
	return scanJobs(ctx, r.pool, "WHERE enterprise_id=$1 ORDER BY created_at DESC", eid)
}
func (r *jobRepo) ListPublished(ctx context.Context, offset, limit int) ([]domain.Job, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT count(*) FROM jobs WHERE status='published'`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count published jobs: %w", err)
	}
	items, err := scanJobs(ctx, r.pool, "WHERE status='published' ORDER BY created_at DESC LIMIT $1 OFFSET $2", limit, offset)
	return items, total, err
}
func (r *jobRepo) ListAll(ctx context.Context, offset, limit int) ([]domain.Job, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT count(*) FROM jobs`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count jobs: %w", err)
	}
	items, err := scanJobs(ctx, r.pool, "ORDER BY created_at DESC LIMIT $1 OFFSET $2", limit, offset)
	return items, total, err
}
func scanJobs(ctx context.Context, pool *pgxpool.Pool, where string, args ...any) ([]domain.Job, error) {
	q := `SELECT id, enterprise_id, title, description, location, salary_fen, job_type, status, version, created_at, updated_at FROM jobs ` + where
	rows, err := pool.Query(ctx, q, args...)
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

func (r *jobRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM jobs WHERE id=$1`, id)
	return err
}

// ---- Resume Repository ----

type pgResumeRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewResumeRepository() repository.ResumeRepository { return &pgResumeRepo{pool: s.pool} }

func (r *pgResumeRepo) Create(ctx context.Context, v domain.Resume) (domain.Resume, error) {
	now := time.Now()
	v.Version = 1
	v.CreatedAt = now
	v.UpdatedAt = now
	skills, _ := json.Marshal(v.Skills) // text 列存 JSON，读取端 json.Unmarshal 对称解析
	_, err := r.pool.Exec(ctx,
		`INSERT INTO resumes (id, user_id, title, name, phone, email, education, work_experience, skills, certificate_url, content, visibility, version, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		v.ID, v.UserID, v.Title, v.Name, v.Phone, v.Email, v.Education, v.WorkExperience, skills, v.CertificateURL, v.Content, v.Visibility, v.Version, v.CreatedAt, v.UpdatedAt)
	return v, err
}
func (r *pgResumeRepo) Update(ctx context.Context, id string, v domain.Resume) (domain.Resume, error) {
	oldVersion := v.Version // 乐观锁：WHERE version=$旧值
	v.Version++
	v.UpdatedAt = time.Now()
	skills, _ := json.Marshal(v.Skills)
	tag, err := r.pool.Exec(ctx,
		`UPDATE resumes SET title=$1,name=$2,phone=$3,email=$4,education=$5,work_experience=$6,skills=$7,certificate_url=$8,content=$9,visibility=$10,version=$11,updated_at=$12 WHERE id=$13 AND version=$14`,
		v.Title, v.Name, v.Phone, v.Email, v.Education, v.WorkExperience, skills, v.CertificateURL, v.Content, v.Visibility, v.Version, v.UpdatedAt, id, oldVersion)
	if err != nil {
		return domain.Resume{}, err
	}
	if tag.RowsAffected() == 0 {
		var one int
		if err := r.pool.QueryRow(ctx, `SELECT 1 FROM resumes WHERE id=$1`, id).Scan(&one); err != nil {
			return domain.Resume{}, fmt.Errorf("resume %s not found", id)
		}
		return domain.Resume{}, fmt.Errorf("resume %s 已被他人修改，请刷新后重试", id)
	}
	return v, nil
}
func (r *pgResumeRepo) FindByID(ctx context.Context, id string) (domain.Resume, error) {
	var v domain.Resume
	var skills string
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, title, name, phone, email, education, work_experience, skills, certificate_url, content, visibility, version, created_at, updated_at FROM resumes WHERE id=$1`, id).
		Scan(&v.ID, &v.UserID, &v.Title, &v.Name, &v.Phone, &v.Email, &v.Education, &v.WorkExperience, &skills, &v.CertificateURL, &v.Content, &v.Visibility, &v.Version, &v.CreatedAt, &v.UpdatedAt)
	if err == nil {
		json.Unmarshal([]byte(skills), &v.Skills)
	}
	return v, err
}
func (r *pgResumeRepo) ListByUser(ctx context.Context, userID string) ([]domain.Resume, error) {
	return scanResumes(ctx, r.pool, "WHERE user_id=$1 ORDER BY created_at DESC", userID)
}
func (r *pgResumeRepo) ListAll(ctx context.Context, offset, limit int) ([]domain.Resume, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT count(*) FROM resumes`).Scan(&total); err != nil {
		return nil, 0, err
	}
	items, err := scanResumes(ctx, r.pool, "ORDER BY created_at DESC LIMIT $1 OFFSET $2", limit, offset)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// ListByIDs 批量按 ID 取简历（ListApplicantsForJob 防 N+1）。
func (r *pgResumeRepo) ListByIDs(ctx context.Context, ids []string) ([]domain.Resume, error) {
	if len(ids) == 0 {
		return []domain.Resume{}, nil
	}
	return scanResumes(ctx, r.pool, "WHERE id = ANY($1)", ids)
}
func scanResumes(ctx context.Context, pool *pgxpool.Pool, where string, args ...any) ([]domain.Resume, error) {
	q := `SELECT id, user_id, title, name, phone, email, education, work_experience, skills, certificate_url, content, visibility, version, created_at, updated_at FROM resumes ` + where
	rows, err := pool.Query(ctx, q, args...)
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

func (r *pgAppRepo) Create(ctx context.Context, a domain.JobApplication) (domain.JobApplication, error) {
	now := time.Now()
	a.Version = 1
	a.CreatedAt = now
	a.UpdatedAt = now
	_, err := r.pool.Exec(ctx,
		`INSERT INTO job_applications (id, job_id, resume_id, applicant_id, status, version, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		a.ID, a.JobID, a.ResumeID, a.ApplicantID, string(a.Status), a.Version, a.CreatedAt, a.UpdatedAt)
	return a, err
}
func (r *pgAppRepo) FindByID(ctx context.Context, id string) (domain.JobApplication, error) {
	var a domain.JobApplication
	var s string
	err := r.pool.QueryRow(ctx,
		`SELECT id, job_id, resume_id, applicant_id, status, version, created_at, updated_at FROM job_applications WHERE id=$1`, id).
		Scan(&a.ID, &a.JobID, &a.ResumeID, &a.ApplicantID, &s, &a.Version, &a.CreatedAt, &a.UpdatedAt)
	a.Status = domain.AppStatus(s)
	return a, err
}
func (r *pgAppRepo) UpdateStatus(ctx context.Context, id string, status domain.AppStatus) (domain.JobApplication, error) {
	tag, err := r.pool.Exec(ctx,
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
	err = r.pool.QueryRow(ctx,
		`SELECT id, job_id, resume_id, applicant_id, status, version, created_at, updated_at FROM job_applications WHERE id=$1`, id).
		Scan(&a.ID, &a.JobID, &a.ResumeID, &a.ApplicantID, &s, &a.Version, &a.CreatedAt, &a.UpdatedAt)
	a.Status = domain.AppStatus(s)
	return a, err
}
func (r *pgAppRepo) ListByJob(ctx context.Context, jobID string) ([]domain.JobApplication, error) {
	return scanApps(ctx, r.pool, "WHERE job_id=$1 ORDER BY created_at DESC", jobID)
}
func (r *pgAppRepo) ListByApplicant(ctx context.Context, userID string) ([]domain.JobApplication, error) {
	return scanApps(ctx, r.pool, "WHERE applicant_id=$1 ORDER BY created_at DESC", userID)
}
func scanApps(ctx context.Context, pool *pgxpool.Pool, where string, args ...any) ([]domain.JobApplication, error) {
	q := `SELECT id, job_id, resume_id, applicant_id, status, version, created_at, updated_at FROM job_applications ` + where
	rows, err := pool.Query(ctx, q, args...)
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
func (r *pgPostRepo) Create(ctx context.Context, p domain.Post) (domain.Post, error) {
	img, err := json.Marshal(p.Images)
	if err != nil {
		return domain.Post{}, fmt.Errorf("marshal post images: %w", err)
	}
	now := time.Now()
	p.Version = 1
	p.CreatedAt = now
	p.UpdatedAt = now
	_, err = r.pool.Exec(ctx,
		`INSERT INTO posts(id,author_id,title,content,images,city_code,status,version,created_at,updated_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		p.ID, p.AuthorID, p.Title, p.Content, img, p.CityCode, p.Status, p.Version, p.CreatedAt, p.UpdatedAt)
	return p, err
}
func (r *pgPostRepo) Update(ctx context.Context, id string, p domain.Post) (domain.Post, error) {
	img, err := json.Marshal(p.Images)
	if err != nil {
		return domain.Post{}, fmt.Errorf("marshal post images: %w", err)
	}
	p.Version++
	p.UpdatedAt = time.Now()
	_, err = r.pool.Exec(ctx, `UPDATE posts SET title=$1,content=$2,images=$3,status=$4,version=$5,updated_at=$6 WHERE id=$7`,
		p.Title, p.Content, img, p.Status, p.Version, p.UpdatedAt, id)
	return p, err
}
func (r *pgPostRepo) FindByID(ctx context.Context, id string) (domain.Post, error) {
	var p domain.Post
	var img []byte
	err := r.pool.QueryRow(ctx, `SELECT id,author_id,title,content,images,city_code,status,version,created_at,updated_at FROM posts WHERE id=$1`, id).
		Scan(&p.ID, &p.AuthorID, &p.Title, &p.Content, &img, &p.CityCode, &p.Status, &p.Version, &p.CreatedAt, &p.UpdatedAt)
	json.Unmarshal(img, &p.Images)
	return p, err
}
func (r *pgPostRepo) ListPublished(ctx context.Context, offset, limit int) ([]domain.Post, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT count(*) FROM posts WHERE status='published'`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count published posts: %w", err)
	}
	rows, err := r.pool.Query(ctx, `SELECT id,author_id,title,content,images,city_code,status,version,created_at,updated_at FROM posts WHERE status='published' ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
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
func (r *pgPostRepo) ListAll(ctx context.Context, offset, limit int) ([]domain.Post, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT count(*) FROM posts`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count all posts: %w", err)
	}
	rows, err := r.pool.Query(ctx, `SELECT id,author_id,title,content,images,city_code,status,version,created_at,updated_at FROM posts ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
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
func (r *pgPostRepo) ListByAuthor(ctx context.Context, uid string) ([]domain.Post, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,author_id,title,content,images,city_code,status,version,created_at,updated_at FROM posts WHERE author_id=$1 ORDER BY created_at DESC`, uid)
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
func (r *pgCommentRepo) Create(ctx context.Context, c domain.Comment) (domain.Comment, error) {
	_, err := r.pool.Exec(ctx, `INSERT INTO comments(id,post_id,author_id,content,status,created_at) VALUES($1,$2,$3,$4,$5,$6)`,
		c.ID, c.PostID, c.AuthorID, c.Content, c.Status, time.Now())
	return c, err
}
func (r *pgCommentRepo) ListByPost(ctx context.Context, postID string) ([]domain.Comment, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,post_id,author_id,content,status,created_at FROM comments WHERE post_id=$1 ORDER BY created_at`, postID)
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
func (r *pgReportRepo) Create(ctx context.Context, rp domain.Report) (domain.Report, error) {
	_, err := r.pool.Exec(ctx, `INSERT INTO reports(id,reporter_id,resource_type,resource_id,reason,status,created_at) VALUES($1,$2,$3,$4,$5,$6,$7)`,
		rp.ID, rp.ReporterID, rp.ResourceType, rp.ResourceID, rp.Reason, "pending", time.Now())
	return rp, err
}
func (r *pgReportRepo) ListPending(ctx context.Context, offset, limit int) ([]domain.Report, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT count(*) FROM reports WHERE status='pending'`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count pending reports: %w", err)
	}
	rows, err := r.pool.Query(ctx, `SELECT id,reporter_id,resource_type,resource_id,reason,status,created_at FROM reports WHERE status='pending' ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
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
func (r *pgListingRepo) Create(ctx context.Context, l domain.Listing) (domain.Listing, error) {
	img, err := json.Marshal(l.Images)
	if err != nil {
		return domain.Listing{}, fmt.Errorf("marshal listing images: %w", err)
	}
	now := time.Now()
	l.Version = 1
	l.CreatedAt = now
	l.UpdatedAt = now
	_, err = r.pool.Exec(ctx, `INSERT INTO listings(id,seller_id,title,description,category,price_fen,images,district,status,version,created_at,updated_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		l.ID, l.SellerID, l.Title, l.Description, l.Category, l.PriceFen, img, l.District, l.Status, l.Version, l.CreatedAt, l.UpdatedAt)
	return l, err
}
func (r *pgListingRepo) Update(ctx context.Context, id string, l domain.Listing) (domain.Listing, error) {
	img, err := json.Marshal(l.Images)
	if err != nil {
		return domain.Listing{}, fmt.Errorf("marshal listing images: %w", err)
	}
	l.Version++
	l.UpdatedAt = time.Now()
	_, err = r.pool.Exec(ctx, `UPDATE listings SET title=$1,description=$2,price_fen=$3,images=$4,status=$5,version=$6,updated_at=$7 WHERE id=$8`,
		l.Title, l.Description, l.PriceFen, img, l.Status, l.Version, l.UpdatedAt, id)
	return l, err
}
func (r *pgListingRepo) FindByID(ctx context.Context, id string) (domain.Listing, error) {
	var l domain.Listing
	var img []byte
	err := r.pool.QueryRow(ctx, `SELECT id,seller_id,title,description,category,price_fen,images,district,status,version,created_at,updated_at FROM listings WHERE id=$1`, id).
		Scan(&l.ID, &l.SellerID, &l.Title, &l.Description, &l.Category, &l.PriceFen, &img, &l.District, &l.Status, &l.Version, &l.CreatedAt, &l.UpdatedAt)
	json.Unmarshal(img, &l.Images)
	return l, err
}
func (r *pgListingRepo) ListByStatus(ctx context.Context, status string, offset, limit int) ([]domain.Listing, int, error) {
	var total int
	where := ""
	args := []any{}
	if status != "" {
		where = "WHERE status=$1"
		args = append(args, status)
	}
	if err := r.pool.QueryRow(ctx, `SELECT count(*) FROM listings `+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count listings: %w", err)
	}
	args = append(args, limit, offset)
	rows, err := r.pool.Query(ctx, `SELECT id,seller_id,title,description,category,price_fen,images,district,status,version,created_at,updated_at FROM listings `+where+` ORDER BY created_at DESC LIMIT $`+fmt.Sprintf("%d", len(args)-1)+` OFFSET $`+fmt.Sprintf("%d", len(args)), args...)
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
func (r *pgListingRepo) ListBySeller(ctx context.Context, uid string) ([]domain.Listing, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,seller_id,title,description,category,price_fen,images,district,status,version,created_at,updated_at FROM listings WHERE seller_id=$1 ORDER BY created_at DESC`, uid)
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
func (r *pgListingRepo) AddFavorite(ctx context.Context, lid, uid string) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO listing_favorites(listing_id,user_id,created_at) VALUES($1,$2,$3) ON CONFLICT DO NOTHING`, lid, uid, time.Now())
	if err != nil {
		return fmt.Errorf("add favorite: %w", err)
	}
	return nil
}
func (r *pgListingRepo) RemoveFavorite(ctx context.Context, lid, uid string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM listing_favorites WHERE listing_id=$1 AND user_id=$2`, lid, uid)
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
func (r *pgLabourRepo) Create(ctx context.Context, o domain.LabourOrder) (domain.LabourOrder, error) {
	now := time.Now()
	o.Version = 1
	o.CreatedAt = now
	o.UpdatedAt = now
	_, err := r.pool.Exec(ctx, `INSERT INTO labour_orders(id,employer_id,title,description,worker_count,start_date,end_date,budget_fen,status,version,created_at,updated_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		o.ID, o.EmployerID, o.Title, o.Description, o.WorkerCount, o.StartDate, o.EndDate, o.BudgetFen, o.Status, o.Version, o.CreatedAt, o.UpdatedAt)
	return o, err
}
func (r *pgLabourRepo) FindByID(ctx context.Context, id string) (domain.LabourOrder, error) {
	var o domain.LabourOrder
	err := r.pool.QueryRow(ctx, `SELECT id,employer_id,title,description,worker_count,start_date,end_date,budget_fen,status,version,created_at,updated_at FROM labour_orders WHERE id=$1`, id).
		Scan(&o.ID, &o.EmployerID, &o.Title, &o.Description, &o.WorkerCount, &o.StartDate, &o.EndDate, &o.BudgetFen, &o.Status, &o.Version, &o.CreatedAt, &o.UpdatedAt)
	return o, err
}
func (r *pgLabourRepo) ListByEmployer(ctx context.Context, uid string) ([]domain.LabourOrder, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,employer_id,title,description,worker_count,start_date,end_date,budget_fen,status,version,created_at,updated_at FROM labour_orders WHERE employer_id=$1 ORDER BY created_at DESC`, uid)
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
func (r *pgLabourRepo) ListAll(ctx context.Context, offset, limit int) ([]domain.LabourOrder, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT count(*) FROM labour_orders`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count labour orders: %w", err)
	}
	rows, err := r.pool.Query(ctx, `SELECT id,employer_id,title,description,worker_count,start_date,end_date,budget_fen,status,version,created_at,updated_at FROM labour_orders ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
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
func (r *pgLabourRepo) CreateQuote(ctx context.Context, q domain.LabourQuote) (domain.LabourQuote, error) {
	_, err := r.pool.Exec(ctx, `INSERT INTO labour_quotes(id,order_id,quoter_id,quoter_name,amount_fen,proposal,status,created_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8)`,
		q.ID, q.OrderID, q.QuoterID, q.QuoterName, q.AmountFen, q.Proposal, "pending", time.Now())
	return q, err
}
func (r *pgLabourRepo) ListQuotes(ctx context.Context, orderID string) ([]domain.LabourQuote, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,order_id,quoter_id,quoter_name,amount_fen,proposal,status,created_at FROM labour_quotes WHERE order_id=$1 ORDER BY created_at`, orderID)
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
func (r *pgLabourRepo) CreateAssignment(ctx context.Context, a domain.Assignment) (domain.Assignment, error) {
	_, err := r.pool.Exec(ctx, `INSERT INTO assignments(id,order_id,worker_id,status,created_at) VALUES($1,$2,$3,$4,$5)`,
		a.ID, a.OrderID, a.WorkerID, "assigned", time.Now())
	return a, err
}

func (r *pgLabourRepo) ListAssignmentsByOrder(ctx context.Context, orderID string) ([]domain.Assignment, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,order_id,worker_id,status,created_at FROM assignments WHERE order_id=$1 ORDER BY created_at DESC`, orderID)
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

func (r *pgLabourRepo) ListAssignmentsByWorker(ctx context.Context, workerID string) ([]domain.Assignment, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,order_id,worker_id,status,created_at FROM assignments WHERE worker_id=$1 ORDER BY created_at DESC`, workerID)
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

func (r *userRepo) FindByOpenID(ctx context.Context, openid string) (domain.User, error) {
	var u domain.User
	err := r.pool.QueryRow(ctx,
		`SELECT id, wechat_openid, phone_ciphertext, password_hash, name, avatar_url, gender, birthday, region, bio, role, status, token_version, version, created_at, updated_at FROM users WHERE wechat_openid=$1 AND deleted_at IS NULL`, openid).
		Scan(&u.ID, &u.WechatOpenID, &u.PhoneCipher, &u.PasswordHash, &u.Name, &u.AvatarURL, &u.Gender, &u.Birthday, &u.Region, &u.Bio, &u.Role, &u.Status, &u.TokenVersion, &u.Version, &u.CreatedAt, &u.UpdatedAt)
	if r.cipher != nil && u.PhoneCipher != "" {
		if dec, err := r.cipher.Decrypt(u.PhoneCipher); err == nil {
			u.PhoneCipher = dec
		}
	}
	return u, err
}

func (r *userRepo) Create(ctx context.Context, u domain.User) (domain.User, error) {
	now := time.Now()
	u.Version = 1
	u.CreatedAt = now
	u.UpdatedAt = now
	if u.Role == "" {
		u.Role = domain.RoleIndividual
	}
	_, err := r.pool.Exec(ctx,
		`INSERT INTO users (id, wechat_openid, phone_ciphertext, password_hash, name, avatar_url, role, status, token_version, version, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		u.ID, u.WechatOpenID, u.PhoneCipher, u.PasswordHash, u.Name, u.AvatarURL, string(u.Role), u.Status, u.TokenVersion, u.Version, u.CreatedAt, u.UpdatedAt)
	return u, err
}

func (r *userRepo) FindByID(ctx context.Context, id string) (domain.User, error) {
	var u domain.User
	err := r.pool.QueryRow(ctx,
		`SELECT id, wechat_openid, phone_ciphertext, password_hash, name, avatar_url, gender, birthday, region, bio, role, status, token_version, version, created_at, updated_at FROM users WHERE id=$1 AND deleted_at IS NULL`, id).
		Scan(&u.ID, &u.WechatOpenID, &u.PhoneCipher, &u.PasswordHash, &u.Name, &u.AvatarURL, &u.Gender, &u.Birthday, &u.Region, &u.Bio, &u.Role, &u.Status, &u.Version, &u.CreatedAt, &u.UpdatedAt)
	if r.cipher != nil && u.PhoneCipher != "" {
		if dec, err := r.cipher.Decrypt(u.PhoneCipher); err == nil {
			u.PhoneCipher = dec
		}
	}
	return u, err
}

func (r *userRepo) All(ctx context.Context) ([]domain.User, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, wechat_openid, phone_ciphertext, password_hash, name, avatar_url, gender, birthday, region, bio, role, status, token_version, version, created_at, updated_at FROM users WHERE deleted_at IS NULL ORDER BY created_at DESC LIMIT 200`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.WechatOpenID, &u.PhoneCipher, &u.PasswordHash, &u.Name, &u.AvatarURL, &u.Gender, &u.Birthday, &u.Region, &u.Bio, &u.Role, &u.Status, &u.TokenVersion, &u.Version, &u.CreatedAt, &u.UpdatedAt); err != nil {
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

// Count 统计未删除用户总数（首页 stats 计数，聚合查询不物化行）。
func (r *userRepo) Count(ctx context.Context) (int, error) {
	var n int
	if err := r.pool.QueryRow(ctx, `SELECT count(*) FROM users WHERE deleted_at IS NULL`).Scan(&n); err != nil {
		return 0, fmt.Errorf("count users: %w", err)
	}
	return n, nil
}

func (r *userRepo) UpdateRole(ctx context.Context, id string, role domain.Role) error {
	// 角色变更同时令牌版本+1：已签发 token 立即失效（重登后生效新角色，
	// 防"降权后旧 token 继续以高权限使用"）。
	_, err := r.pool.Exec(ctx, `UPDATE users SET role=$1, token_version=token_version+1, version=version+1, updated_at=NOW() WHERE id=$2`, string(role), id)
	if err != nil {
		return fmt.Errorf("update role for %s: %w", id, err)
	}
	return nil
}

func (r *userRepo) UpdateAvatar(ctx context.Context, userID, avatarURL string) error {
	_, err := r.pool.Exec(ctx, `UPDATE users SET avatar_url=$1, updated_at=NOW() WHERE id=$2`, avatarURL, userID)
	if err != nil {
		return fmt.Errorf("update avatar for %s: %w", userID, err)
	}
	return nil
}

func (r *userRepo) UpdateName(ctx context.Context, userID, name string) error {
	_, err := r.pool.Exec(ctx, `UPDATE users SET name=$1, updated_at=NOW() WHERE id=$2`, name, userID)
	if err != nil {
		return fmt.Errorf("update name for %s: %w", userID, err)
	}
	return nil
}

// UpdateProfile updates the editable profile fields. Phone is plaintext here;
// it is encrypted before persistence. An empty Phone leaves it unchanged.
func (r *userRepo) UpdateProfile(ctx context.Context, id string, p domain.UserProfile) error {
	enc := p.Phone
	if p.Phone != "" && r.cipher != nil {
		if c, err := r.cipher.Encrypt(p.Phone); err == nil {
			enc = c
		}
	}
	_, err := r.pool.Exec(ctx,
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
func (r *userRepo) Delete(ctx context.Context, id string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin delete user: %w", err)
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `DELETE FROM refresh_tokens WHERE user_id=$1`, id); err != nil {
		return fmt.Errorf("delete refresh tokens for %s: %w", id, err)
	}
	if _, err := tx.Exec(ctx, `DELETE FROM user_roles WHERE user_id=$1`, id); err != nil {
		return fmt.Errorf("delete user roles for %s: %w", id, err)
	}
	if _, err := tx.Exec(ctx, `DELETE FROM users WHERE id=$1`, id); err != nil {
		return fmt.Errorf("delete user %s: %w", id, err)
	}
	return tx.Commit(ctx)
}

// ---- RefreshToken ----

type refreshTokenRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewRefreshTokenRepository() repository.RefreshTokenRepository {
	return &refreshTokenRepo{pool: s.pool}
}

func (r *refreshTokenRepo) Store(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, created_at) VALUES ($1,$2,$3,$4,$5)`,
		fmt.Sprintf("rt-%d", time.Now().UnixNano()), userID, tokenHash, expiresAt, time.Now())
	if err != nil {
		return fmt.Errorf("store refresh token: %w", err)
	}
	return nil
}

func (r *refreshTokenRepo) Find(ctx context.Context, tokenHash string) (userID string, expiresAt time.Time, revoked bool, err error) {
	var revokedAt *time.Time
	err = r.pool.QueryRow(ctx,
		`SELECT user_id, expires_at, revoked_at FROM refresh_tokens WHERE token_hash=$1`, tokenHash).
		Scan(&userID, &expiresAt, &revokedAt)
	revoked = revokedAt != nil
	return
}

func (r *refreshTokenRepo) Revoke(ctx context.Context, tokenHash string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE refresh_tokens SET revoked_at=$1 WHERE token_hash=$2`, time.Now(), tokenHash)
	if err != nil {
		return fmt.Errorf("revoke refresh token: %w", err)
	}
	return nil
}

// Consume 原子消费：DELETE 成功（影响 1 行）才返回 found=true。并发第二次消费影响 0 行 → 拒绝，
// 保证同一刷新令牌不会签发两份新会话。
func (r *refreshTokenRepo) Consume(ctx context.Context, tokenHash string) (bool, string, time.Time, error) {
	var userID string
	var expiresAt time.Time
	err := r.pool.QueryRow(ctx,
		`DELETE FROM refresh_tokens WHERE token_hash=$1 AND revoked_at IS NULL AND expires_at > NOW()
		 RETURNING user_id, expires_at`, tokenHash).
		Scan(&userID, &expiresAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, "", time.Time{}, nil
		}
		return false, "", time.Time{}, fmt.Errorf("consume refresh token: %w", err)
	}
	return true, userID, expiresAt, nil
}

// scanContracts removed — replaced by scanContractsPaged

// ---- DemandIntent Repository ----

type pgIntentRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewIntentRepository() repository.IntentRepository {
	return &pgIntentRepo{pool: s.pool}
}

func (r *pgIntentRepo) Create(ctx context.Context, it domain.DemandIntent) (domain.DemandIntent, error) {
	now := time.Now()
	it.Version = 1
	it.CreatedAt = now
	it.UpdatedAt = now
	_, err := r.pool.Exec(ctx,
		`INSERT INTO demand_intents (id, demand_id, intentor_id, intentor_name, contact, remark, status, version, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		it.ID, it.DemandID, it.IntentorID, it.IntentorName, it.Contact, it.Remark, it.Status, it.Version, it.CreatedAt, it.UpdatedAt)
	if err != nil {
		// P1 修复：唯一索引 (demand_id, intentor_id) 并发兜底——重复登记映射为友好错误。
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.DemandIntent{}, fmt.Errorf("已登记过该需求的对接意向，请勿重复提交")
		}
		return domain.DemandIntent{}, fmt.Errorf("create intent: %w", err)
	}
	return it, nil
}

func (r *pgIntentRepo) ListByDemand(ctx context.Context, demandID string) ([]domain.DemandIntent, error) {
	rows, err := r.pool.Query(ctx,
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

func (r *pgIntentRepo) ListByIntentor(ctx context.Context, intentorID string) ([]domain.DemandIntent, error) {
	rows, err := r.pool.Query(ctx,
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

func (r *pgIntentRepo) UpdateStatus(ctx context.Context, id string, status string) (domain.DemandIntent, error) {
	// B 批加固：CAS 语义——仅 pending 状态可流转（确认/关闭/拒绝）。
	// 并发重复确认同一意向时，后到者 RowsAffected=0 → 明确报错，
	// 消除"同一意向生成多张工单"的竞态窗口（配合 work_orders.intent_id 唯一索引）。
	var it domain.DemandIntent
	err := r.pool.QueryRow(ctx,
		`UPDATE demand_intents SET status=$2, version=version+1, updated_at=now()
		WHERE id=$1 AND status='pending'
		RETURNING id, demand_id, intentor_id, intentor_name, contact, remark, status, version, created_at, updated_at`,
		id, status).
		Scan(&it.ID, &it.DemandID, &it.IntentorID, &it.IntentorName, &it.Contact, &it.Remark, &it.Status, &it.Version, &it.CreatedAt, &it.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.DemandIntent{}, fmt.Errorf("意向不存在或已处理")
		}
		return domain.DemandIntent{}, fmt.Errorf("update intent %s: %w", id, err)
	}
	return it, nil
}

// ---- WorkOrder Repository (接单派单闭环) ----

type pgWorkOrderRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewWorkOrderRepository() repository.WorkOrderRepository {
	return &pgWorkOrderRepo{pool: s.pool}
}

// workOrderInsertCols 用于 INSERT（原列名）；workOrderSelectCols 用于 SELECT/RETURNING——
// 存量工单 intent_id 可能为 NULL，直接 Scan 进 *string 会报错导致全部工单接口 500，
// SELECT 侧用 COALESCE 包裹成空串，两处列顺序一致（scanWorkOrder 的 Scan 顺序不变）。
const workOrderInsertCols = `id, order_no, demand_id, intent_id, publisher_id, publisher_name, worker_id, worker_name, amount_fen, status, result_photos, rework_note, cancel_reason, created_at, updated_at`

const workOrderSelectCols = `id, order_no, demand_id, COALESCE(intent_id,'') AS intent_id, publisher_id, publisher_name, worker_id, worker_name, amount_fen, status, result_photos, rework_note, cancel_reason, created_at, updated_at`

func scanWorkOrder(row interface{ Scan(...any) error }) (domain.WorkOrder, error) {
	var wo domain.WorkOrder
	var photos []byte
	err := row.Scan(&wo.ID, &wo.OrderNo, &wo.DemandID, &wo.IntentID, &wo.PublisherID, &wo.PublisherName,
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

func (r *pgWorkOrderRepo) Create(ctx context.Context, wo domain.WorkOrder) (domain.WorkOrder, error) {
	now := time.Now()
	wo.CreatedAt = now
	wo.UpdatedAt = now
	photos, _ := json.Marshal(wo.ResultPhotos)
	_, err := r.pool.Exec(ctx,
		`INSERT INTO work_orders (`+workOrderInsertCols+`)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		wo.ID, wo.OrderNo, wo.DemandID, wo.IntentID, wo.PublisherID, wo.PublisherName, wo.WorkerID, wo.WorkerName,
		wo.AmountFen, wo.Status, photos, wo.ReworkNote, wo.CancelReason, wo.CreatedAt, wo.UpdatedAt)
	if err != nil {
		// B 批加固：同一意向重复建单被唯一索引兜底（并发双建单场景）
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.WorkOrder{}, fmt.Errorf("该意向已生成工单，请勿重复确认")
		}
		return domain.WorkOrder{}, fmt.Errorf("create work order: %w", err)
	}
	return wo, nil
}

func (r *pgWorkOrderRepo) FindByID(ctx context.Context, id string) (domain.WorkOrder, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT `+workOrderSelectCols+` FROM work_orders WHERE id=$1`, id)
	wo, err := scanWorkOrder(row)
	if err != nil {
		return domain.WorkOrder{}, fmt.Errorf("find work order %s: %w", id, err)
	}
	return wo, nil
}

func (r *pgWorkOrderRepo) ListByPublisher(ctx context.Context, publisherID string) ([]domain.WorkOrder, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+workOrderSelectCols+` FROM work_orders WHERE publisher_id=$1 ORDER BY created_at DESC`, publisherID)
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

func (r *pgWorkOrderRepo) ListByWorker(ctx context.Context, workerID string) ([]domain.WorkOrder, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+workOrderSelectCols+` FROM work_orders WHERE worker_id=$1 ORDER BY created_at DESC`, workerID)
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

func (r *pgWorkOrderRepo) UpdateStatus(ctx context.Context, id string, oldStatus, status domain.WorkOrderStatus) (domain.WorkOrder, error) {
	// CAS 语义：WHERE 带旧状态，并发时后到者 RowsAffected=0 → 明确报错，
	// 消除"取消与开始作业并发导致已取消订单复活"的竞态窗口。
	row := r.pool.QueryRow(ctx,
		`UPDATE work_orders SET status=$2, updated_at=now() WHERE id=$1 AND status=$3
		RETURNING `+workOrderSelectCols, id, status, oldStatus)
	wo, err := scanWorkOrder(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.WorkOrder{}, fmt.Errorf("工单状态已变更，请刷新后重试")
		}
		return domain.WorkOrder{}, fmt.Errorf("update work order status %s: %w", id, err)
	}
	return wo, nil
}

func (r *pgWorkOrderRepo) UpdatePhotos(ctx context.Context, id string, photos []string) (domain.WorkOrder, error) {
	data, _ := json.Marshal(photos)
	row := r.pool.QueryRow(ctx,
		`UPDATE work_orders SET result_photos=$2, updated_at=now() WHERE id=$1
		RETURNING `+workOrderSelectCols, id, data)
	wo, err := scanWorkOrder(row)
	if err != nil {
		return domain.WorkOrder{}, fmt.Errorf("update work order photos %s: %w", id, err)
	}
	return wo, nil
}

func (r *pgWorkOrderRepo) UpdateRework(ctx context.Context, id string, note string) (domain.WorkOrder, error) {
	row := r.pool.QueryRow(ctx,
		`UPDATE work_orders SET rework_note=$2, updated_at=now() WHERE id=$1
		RETURNING `+workOrderSelectCols, id, note)
	wo, err := scanWorkOrder(row)
	if err != nil {
		return domain.WorkOrder{}, fmt.Errorf("update work order rework %s: %w", id, err)
	}
	return wo, nil
}

func (r *pgWorkOrderRepo) UpdateCancel(ctx context.Context, id string, reason string) (domain.WorkOrder, error) {
	row := r.pool.QueryRow(ctx,
		`UPDATE work_orders SET cancel_reason=$2, updated_at=now() WHERE id=$1
		RETURNING `+workOrderSelectCols, id, reason)
	wo, err := scanWorkOrder(row)
	if err != nil {
		return domain.WorkOrder{}, fmt.Errorf("update work order cancel %s: %w", id, err)
	}
	return wo, nil
}

func (r *contractRepo) FindByID(ctx context.Context, id string) (domain.Contract, error) {
	var v domain.Contract
	var status string
	err := r.pool.QueryRow(ctx,
		`SELECT id, enterprise_id, template_id, sign_url, status, version, created_at, updated_at
		FROM contracts WHERE id=$1`, id).
		Scan(&v.ID, &v.EnterpriseID, &v.TemplateID, &v.SignURL, &status, &v.Version, &v.CreatedAt, &v.UpdatedAt)
	if err != nil {
		return domain.Contract{}, fmt.Errorf("contract %s: %w", id, err)
	}
	v.Status = domain.ContractStatus(status)
	return v, nil
}

func (r *contractRepo) UpdateStatus(ctx context.Context, id string, status domain.ContractStatus) (domain.Contract, error) {
	tag, err := r.pool.Exec(ctx,
		`UPDATE contracts SET status=$1, version=version+1, updated_at=$2 WHERE id=$3`,
		string(status), time.Now(), id)
	if err != nil {
		return domain.Contract{}, fmt.Errorf("update contract status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.Contract{}, fmt.Errorf("contract %s not found", id)
	}
	return r.FindByID(ctx, id)
}

// scanEmploymentPaged queries employment_requests with pagination.
// The where clause must be a compile-time constant — never pass user input as where.
func scanEmploymentPaged(ctx context.Context, pool *pgxpool.Pool, where string, offset, limit int, args ...any) ([]domain.EmploymentRequest, int, error) {
	countQ := `SELECT COUNT(*) FROM employment_requests ` + where
	var total int
	if err := pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count employment: %w", err)
	}

	q := `SELECT id, enterprise_id, position, headcount, start_date, end_date, status, version, created_at, updated_at
		FROM employment_requests ` + where + ` ORDER BY created_at DESC LIMIT $` + fmt.Sprintf("%d", len(args)+1) + ` OFFSET $` + fmt.Sprintf("%d", len(args)+2)
	allArgs := append(args, limit, offset)
	rows, err := pool.Query(ctx, q, allArgs...)
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
func scanContractsPaged(ctx context.Context, pool *pgxpool.Pool, where string, offset, limit int, args ...any) ([]domain.Contract, int, error) {
	countQ := `SELECT COUNT(*) FROM contracts ` + where
	var total int
	if err := pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count contracts: %w", err)
	}

	q := `SELECT id, enterprise_id, template_id, sign_url, status, version, created_at, updated_at
		FROM contracts ` + where + ` ORDER BY created_at DESC LIMIT $` + fmt.Sprintf("%d", len(args)+1) + ` OFFSET $` + fmt.Sprintf("%d", len(args)+2)
	allArgs := append(args, limit, offset)
	rows, err := pool.Query(ctx, q, allArgs...)
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

func (r *demandRepo) CompareAndSetStatus(ctx context.Context, id string, oldStatus, newStatus domain.DemandStatus) (bool, domain.Demand, error) {
	tag, err := r.pool.Exec(ctx,
		`UPDATE demands SET status=$1, updated_at=$2, version=version+1 WHERE id=$3 AND status=$4`,
		string(newStatus), time.Now(), id, string(oldStatus))
	if err != nil {
		return false, domain.Demand{}, fmt.Errorf("compare-and-set demand status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		// Either demand not found, or status didn't match. Check which.
		d, findErr := r.FindByID(ctx, id)
		if findErr != nil {
			return false, domain.Demand{}, findErr
		}
		return false, d, nil
	}
	d, err := r.FindByID(ctx, id)
	return true, d, err
}

func (r *demandRepo) Delete(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM demands WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("delete demand %s: %w", id, err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("demand %s not found", id)
	}
	return nil
}

func (r *demandRepo) FavoriteDemand(ctx context.Context, userID, demandID string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO demand_favorites (id, user_id, demand_id) VALUES ($1,$2,$3)
		 ON CONFLICT (user_id, demand_id) DO NOTHING`,
		"dfav-"+userID+"-"+demandID, userID, demandID)
	if err != nil {
		return fmt.Errorf("favorite demand %s: %w", demandID, err)
	}
	return nil
}

func (r *demandRepo) UnfavoriteDemand(ctx context.Context, userID, demandID string) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM demand_favorites WHERE user_id=$1 AND demand_id=$2`, userID, demandID)
	if err != nil {
		return fmt.Errorf("unfavorite demand %s: %w", demandID, err)
	}
	return nil
}

func (r *demandRepo) ListFavoriteDemandIDs(ctx context.Context, userID string) ([]string, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT demand_id FROM demand_favorites WHERE user_id=$1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("list favorite demands: %w", err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan favorite demand: %w", err)
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

func (r *demandRepo) ListFavoriteDemands(ctx context.Context, userID string) ([]domain.Demand, error) {
	q := `SELECT d.id, d.publisher_id, d.publisher_name, d.contact, d.district, d.city_code,
		d.biz_type, d.title, d.description, d.images, d.latitude, d.longitude, d.budget_fen, d.offline_amount_fen, d.biz_fields,
		d.status, d.version, d.created_at, d.updated_at, d.deadline
		FROM demands d
		JOIN demand_favorites f ON f.demand_id = d.id
		WHERE f.user_id = $1
		ORDER BY f.created_at DESC`
	return scanDemands(ctx, r.pool, r.cipher, q, []any{userID})
}

// ---- College Repository ----

type pgCollegeRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewCollegeRepository() repository.CollegeRepository {
	return &pgCollegeRepo{pool: s.Pool()}
}

// collegeCols 与 colleges 表列一一对应（迁移 000044 补齐小程序页面字段）
const collegeCols = `id,name,region,city,description,logo_url,status,coop_type,majors,facilities,tags,short_name,level_tags,specialties,major_count,partner_count,teacher_count,student_count,graduate_rate,partners,cover,photos,phone,website,intro,majors_detail,created_at,updated_at`

func (r *pgCollegeRepo) Create(ctx context.Context, c domain.College) (domain.College, error) {
	c.CreatedAt = time.Now()
	c.UpdatedAt = c.CreatedAt
	c.Majors = jsonbSlice(c.Majors)
	c.Facilities = jsonbSlice(c.Facilities)
	c.Tags = jsonbSlice(c.Tags)
	c.Specialties = jsonbSlice(c.Specialties)
	c.Partners = jsonbSlice(c.Partners)
	c.Photos = jsonbSlice(c.Photos)
	c.MajorsDetail = jsonbSlice(c.MajorsDetail)
	_, err := r.pool.Exec(ctx,
		`INSERT INTO colleges (`+collegeCols+`) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27,$28)`,
		c.ID, c.Name, c.Region, c.City, c.Description, c.LogoURL, c.Status, c.CoopType, c.Majors, c.Facilities,
		c.Tags, c.ShortName, c.LevelTags, c.Specialties, c.MajorCount, c.PartnerCount, c.TeacherCount, c.StudentCount,
		c.GraduateRate, c.Partners, c.CoverURL, c.Photos, c.Phone, c.Website, c.Intro, c.MajorsDetail, c.CreatedAt, c.UpdatedAt)
	return c, err
}

func (r *pgCollegeRepo) FindByID(ctx context.Context, id string) (domain.College, error) {
	var c domain.College
	err := r.pool.QueryRow(ctx, `SELECT `+collegeCols+` FROM colleges WHERE id=$1`, id).
		Scan(&c.ID, &c.Name, &c.Region, &c.City, &c.Description, &c.LogoURL, &c.Status, &c.CoopType, &c.Majors, &c.Facilities,
			&c.Tags, &c.ShortName, &c.LevelTags, &c.Specialties, &c.MajorCount, &c.PartnerCount, &c.TeacherCount, &c.StudentCount,
			&c.GraduateRate, &c.Partners, &c.CoverURL, &c.Photos, &c.Phone, &c.Website, &c.Intro, &c.MajorsDetail, &c.CreatedAt, &c.UpdatedAt)
	return c, err
}

func (r *pgCollegeRepo) List(ctx context.Context, region string) ([]domain.College, error) {
	q := `SELECT ` + collegeCols + ` FROM colleges`
	args := []any{}
	if region != "" {
		q += ` WHERE region=$1 AND status='active'`
		args = append(args, region)
	} else {
		// P 批修复：公开列表仅展示 active——此前不过滤 status，下架/无效院校仍公开显示。
		q += ` WHERE status='active'`
	}
	q += ` ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, q, args...)
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

func (r *pgCollegeRepo) Update(ctx context.Context, c domain.College) (domain.College, error) {
	c.UpdatedAt = time.Now()
	c.Majors = jsonbSlice(c.Majors)
	c.Facilities = jsonbSlice(c.Facilities)
	c.Tags = jsonbSlice(c.Tags)
	c.Specialties = jsonbSlice(c.Specialties)
	c.Partners = jsonbSlice(c.Partners)
	c.Photos = jsonbSlice(c.Photos)
	c.MajorsDetail = jsonbSlice(c.MajorsDetail)
	_, err := r.pool.Exec(ctx,
		`UPDATE colleges SET name=$1,region=$2,city=$3,description=$4,logo_url=$5,status=$6,coop_type=$7,majors=$8,facilities=$9,tags=$10,short_name=$11,level_tags=$12,specialties=$13,major_count=$14,partner_count=$15,teacher_count=$16,student_count=$17,graduate_rate=$18,partners=$19,cover=$20,photos=$21,phone=$22,website=$23,intro=$24,majors_detail=$25,updated_at=$26 WHERE id=$27`,
		c.Name, c.Region, c.City, c.Description, c.LogoURL, c.Status, c.CoopType, c.Majors, c.Facilities,
		c.Tags, c.ShortName, c.LevelTags, c.Specialties, c.MajorCount, c.PartnerCount, c.TeacherCount, c.StudentCount,
		c.GraduateRate, c.Partners, c.CoverURL, c.Photos, c.Phone, c.Website, c.Intro, c.MajorsDetail, c.UpdatedAt, c.ID)
	return c, err
}

func (r *pgCollegeRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM colleges WHERE id=$1`, id)
	return err
}

// ---- StudyTour Repository ----

type pgStudyTourRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewStudyTourRepository() repository.StudyTourRepository {
	return &pgStudyTourRepo{pool: s.Pool()}
}

func (r *pgStudyTourRepo) Create(ctx context.Context, st domain.StudyTour) (domain.StudyTour, error) {
	st.CreatedAt = time.Now()
	st.UpdatedAt = st.CreatedAt
	scheduleJSON, err := json.Marshal(jsonbSlice(st.Schedule))
	if err != nil {
		return domain.StudyTour{}, fmt.Errorf("marshal schedule: %w", err)
	}
	_, err = r.pool.Exec(ctx,
		`INSERT INTO study_tours (id,title,destination,duration,capacity,status,description,location,organizer_id,start_date,end_date,cover_image,price_fen,schedule,created_at,updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`,
		st.ID, st.Title, st.Destination, st.Duration, st.Capacity, st.Status,
		st.Description, st.Location, st.OrganizerID, st.StartDate, st.EndDate,
		st.CoverImage, st.PriceFen, scheduleJSON, st.CreatedAt, st.UpdatedAt)
	return st, err
}

func (r *pgStudyTourRepo) FindByID(ctx context.Context, id string) (domain.StudyTour, error) {
	var s domain.StudyTour
	var schedule []byte
	err := r.pool.QueryRow(ctx, `SELECT id,title,destination,duration,capacity,status,description,location,organizer_id,start_date,end_date,cover_image,price_fen,schedule,created_at,updated_at FROM study_tours WHERE id=$1`, id).
		Scan(&s.ID, &s.Title, &s.Destination, &s.Duration, &s.Capacity, &s.Status,
			&s.Description, &s.Location, &s.OrganizerID, &s.StartDate, &s.EndDate,
			&s.CoverImage, &s.PriceFen, &schedule, &s.CreatedAt, &s.UpdatedAt)
	if schedule != nil {
		json.Unmarshal(schedule, &s.Schedule)
	}
	return s, err
}

func (r *pgStudyTourRepo) List(ctx context.Context) ([]domain.StudyTour, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,title,destination,duration,capacity,status,description,location,organizer_id,start_date,end_date,cover_image,price_fen,schedule,created_at,updated_at FROM study_tours ORDER BY created_at DESC`)
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

func (r *pgStudyTourRepo) Update(ctx context.Context, s domain.StudyTour) (domain.StudyTour, error) {
	s.UpdatedAt = time.Now()
	scheduleJSON, err := json.Marshal(jsonbSlice(s.Schedule))
	if err != nil {
		return domain.StudyTour{}, fmt.Errorf("marshal schedule: %w", err)
	}
	_, err = r.pool.Exec(ctx,
		`UPDATE study_tours SET title=$1,destination=$2,duration=$3,capacity=$4,status=$5,description=$6,location=$7,organizer_id=$8,start_date=$9,end_date=$10,cover_image=$11,price_fen=$12,schedule=$13,updated_at=$14 WHERE id=$15`,
		s.Title, s.Destination, s.Duration, s.Capacity, s.Status,
		s.Description, s.Location, s.OrganizerID, s.StartDate, s.EndDate,
		s.CoverImage, s.PriceFen, scheduleJSON, s.UpdatedAt, s.ID)
	return s, err
}

func (r *pgStudyTourRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM study_tours WHERE id=$1`, id)
	return err
}

// ---- Exhibition Repository ----

type pgExhibitionRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewExhibitionRepository() repository.ExhibitionRepository {
	return &pgExhibitionRepo{pool: s.Pool()}
}

func (r *pgExhibitionRepo) Create(ctx context.Context, e domain.Exhibition) (domain.Exhibition, error) {
	e.CreatedAt = time.Now()
	e.UpdatedAt = e.CreatedAt
	_, err := r.pool.Exec(ctx,
		"INSERT INTO exhibitions (id,title,category,description,location,start_date,end_date,booth_count,booth_price_fen,organizer,cover_url,status,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)",
		e.ID, e.Title, e.Category, e.Description, e.Location, e.StartDate, e.EndDate, e.BoothCount, e.BoothPrice, e.Organizer, e.CoverURL, e.Status, e.CreatedAt, e.UpdatedAt)
	return e, err
}
func (r *pgExhibitionRepo) FindByID(ctx context.Context, id string) (domain.Exhibition, error) {
	var e domain.Exhibition
	err := r.pool.QueryRow(ctx, "SELECT id,title,category,description,location,start_date,end_date,booth_count,booth_price_fen,organizer,cover_url,status,created_at,updated_at FROM exhibitions WHERE id=$1", id).
		Scan(&e.ID, &e.Title, &e.Category, &e.Description, &e.Location, &e.StartDate, &e.EndDate, &e.BoothCount, &e.BoothPrice, &e.Organizer, &e.CoverURL, &e.Status, &e.CreatedAt, &e.UpdatedAt)
	return e, err
}
func (r *pgExhibitionRepo) List(ctx context.Context, offset, limit int) ([]domain.Exhibition, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, "SELECT count(*) FROM exhibitions").Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count exhibitions: %w", err)
	}
	rows, err := r.pool.Query(ctx, "SELECT id,title,category,description,location,start_date,end_date,booth_count,booth_price_fen,organizer,cover_url,status,created_at,updated_at FROM exhibitions ORDER BY created_at DESC LIMIT $1 OFFSET $2", limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []domain.Exhibition
	for rows.Next() {
		var e domain.Exhibition
		if err := rows.Scan(&e.ID, &e.Title, &e.Category, &e.Description, &e.Location, &e.StartDate, &e.EndDate, &e.BoothCount, &e.BoothPrice, &e.Organizer, &e.CoverURL, &e.Status, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, e)
	}
	return out, total, rows.Err()
}
func (r *pgExhibitionRepo) Update(ctx context.Context, e domain.Exhibition) (domain.Exhibition, error) {
	e.UpdatedAt = time.Now()
	_, err := r.pool.Exec(ctx,
		"UPDATE exhibitions SET title=$1,category=$2,description=$3,location=$4,start_date=$5,end_date=$6,booth_count=$7,booth_price_fen=$8,organizer=$9,cover_url=$10,status=$11,updated_at=$12 WHERE id=$13",
		e.Title, e.Category, e.Description, e.Location, e.StartDate, e.EndDate, e.BoothCount, e.BoothPrice, e.Organizer, e.CoverURL, e.Status, e.UpdatedAt, e.ID)
	return e, err
}
func (r *pgExhibitionRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM exhibitions WHERE id=$1", id)
	return err
}
func (r *pgExhibitionRepo) CreateBooth(ctx context.Context, b domain.ExhibitionBooth) (domain.ExhibitionBooth, error) {
	b.CreatedAt = time.Now()
	_, err := r.pool.Exec(ctx,
		`INSERT INTO exhibition_booths (id,exhibition_id,exhibitor_id,booth_number,exhibit_name,exhibit_desc,status,created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		b.ID, b.ExhibitionID, b.ExhibitorID, b.BoothNumber, b.ExhibitName, b.ExhibitDesc, b.Status, b.CreatedAt)
	return b, err
}
func (r *pgExhibitionRepo) ListBooths(ctx context.Context, exhibitionID string) ([]domain.ExhibitionBooth, error) {
	rows, err := r.pool.Query(ctx,
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
func (r *pgExhibitionRepo) UpdateBoothStatus(ctx context.Context, id, status string) (domain.ExhibitionBooth, error) {
	var b domain.ExhibitionBooth
	err := r.pool.QueryRow(ctx,
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

func (r *pgTestSiteRepo) Create(ctx context.Context, t domain.TestSite) (domain.TestSite, error) {
	t.CreatedAt = time.Now()
	t.UpdatedAt = t.CreatedAt
	facJSON, _ := json.Marshal(t.Facilities)
	_, err := r.pool.Exec(ctx,
		`INSERT INTO test_sites (id,name,site_type,owner_id,location,booking_rule,status,price_fen,facilities,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		t.ID, t.Name, t.SiteType, t.OwnerID, t.Location, t.BookingRule, t.Status, t.PriceFen, facJSON, t.CreatedAt, t.UpdatedAt)
	return t, err
}

func (r *pgTestSiteRepo) FindByID(ctx context.Context, id string) (domain.TestSite, error) {
	var t domain.TestSite
	var fj []byte
	err := r.pool.QueryRow(ctx,
		`SELECT id,name,site_type,owner_id,location,booking_rule,status,price_fen,facilities,created_at,updated_at FROM test_sites WHERE id=$1`, id).
		Scan(&t.ID, &t.Name, &t.SiteType, &t.OwnerID, &t.Location, &t.BookingRule, &t.Status, &t.PriceFen, &fj, &t.CreatedAt, &t.UpdatedAt)
	if err == nil {
		json.Unmarshal(fj, &t.Facilities)
	}
	return t, err
}

func (r *pgTestSiteRepo) List(ctx context.Context, siteType string) ([]domain.TestSite, error) {
	q := `SELECT id,name,site_type,owner_id,location,booking_rule,status,price_fen,facilities,created_at,updated_at FROM test_sites`
	args := []any{}
	if siteType != "" {
		q += ` WHERE site_type=$1`
		args = append(args, siteType)
	}
	q += ` ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, q, args...)
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

func (r *pgTestSiteRepo) UpdateSite(ctx context.Context, t domain.TestSite) (domain.TestSite, error) {
	t.UpdatedAt = time.Now()
	facJSON, _ := json.Marshal(t.Facilities)
	_, err := r.pool.Exec(ctx,
		`UPDATE test_sites SET name=$1,site_type=$2,location=$3,booking_rule=$4,status=$5,price_fen=$6,facilities=$7,updated_at=$8 WHERE id=$9`,
		t.Name, t.SiteType, t.Location, t.BookingRule, t.Status, t.PriceFen, facJSON, t.UpdatedAt, t.ID)
	return t, err
}

func (r *pgTestSiteRepo) DeleteSite(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM test_sites WHERE id=$1`, id)
	return err
}

func (r *pgTestSiteRepo) CreateBooking(ctx context.Context, b domain.TestSiteBooking) (domain.TestSiteBooking, error) {
	b.CreatedAt = time.Now()
	_, err := r.pool.Exec(ctx,
		`INSERT INTO test_site_bookings (id,site_id,user_id,purpose,start_time,end_time,contact_name,contact_phone,status,review_note,created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		b.ID, b.SiteID, b.UserID, b.Purpose, b.StartTime, b.EndTime, b.ContactName, b.ContactPhone, b.Status, b.ReviewNote, b.CreatedAt)
	return b, err
}
func (r *pgTestSiteRepo) FindBookingByID(ctx context.Context, id string) (domain.TestSiteBooking, error) {
	var b domain.TestSiteBooking
	err := r.pool.QueryRow(ctx,
		`SELECT id,site_id,user_id,purpose,start_time,end_time,contact_name,contact_phone,status,review_note,created_at FROM test_site_bookings WHERE id=$1`, id).
		Scan(&b.ID, &b.SiteID, &b.UserID, &b.Purpose, &b.StartTime, &b.EndTime, &b.ContactName, &b.ContactPhone, &b.Status, &b.ReviewNote, &b.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.TestSiteBooking{}, fmt.Errorf("booking not found")
	}
	return b, err
}
func (r *pgTestSiteRepo) UpdateBookingStatus(ctx context.Context, id, status, note string) (domain.TestSiteBooking, error) {
	var b domain.TestSiteBooking
	err := r.pool.QueryRow(ctx,
		`UPDATE test_site_bookings SET status=$1,review_note=$2 WHERE id=$3 RETURNING id,site_id,user_id,purpose,start_time,end_time,contact_name,contact_phone,status,review_note,created_at`,
		status, note, id).
		Scan(&b.ID, &b.SiteID, &b.UserID, &b.Purpose, &b.StartTime, &b.EndTime, &b.ContactName, &b.ContactPhone, &b.Status, &b.ReviewNote, &b.CreatedAt)
	return b, err
}
func (r *pgTestSiteRepo) ListBookings(ctx context.Context, siteID string) ([]domain.TestSiteBooking, error) {
	rows, err := r.pool.Query(ctx,
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
func (r *pgTestSiteRepo) ListBookingsByUser(ctx context.Context, userID string) ([]domain.TestSiteBooking, error) {
	rows, err := r.pool.Query(ctx,
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
func (r *pgTestSiteRepo) ListAllBookings(ctx context.Context, offset, limit int) ([]domain.TestSiteBooking, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT count(*) FROM test_site_bookings`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count test site bookings: %w", err)
	}
	rows, err := r.pool.Query(ctx,
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

func (r *pgTransRepo) Create(ctx context.Context, t domain.Transformation) (domain.Transformation, error) {
	t.CreatedAt = time.Now()
	t.UpdatedAt = t.CreatedAt
	_, err := r.pool.Exec(ctx,
		`INSERT INTO transformations (id,title,achievement_id,owner_id,progress,partner_id,status,stage,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		t.ID, t.Title, t.AchievementID, t.OwnerID, t.Progress, t.PartnerID, t.Status, t.Stage, t.CreatedAt, t.UpdatedAt)
	return t, err
}

func (r *pgTransRepo) FindByID(ctx context.Context, id string) (domain.Transformation, error) {
	var t domain.Transformation
	err := r.pool.QueryRow(ctx,
		`SELECT id,title,achievement_id,owner_id,progress,partner_id,status,stage,created_at,updated_at FROM transformations WHERE id=$1`, id).
		Scan(&t.ID, &t.Title, &t.AchievementID, &t.OwnerID, &t.Progress, &t.PartnerID, &t.Status, &t.Stage, &t.CreatedAt, &t.UpdatedAt)
	return t, err
}

func (r *pgTransRepo) List(ctx context.Context, ownerID string) ([]domain.Transformation, error) {
	q := `SELECT id,title,achievement_id,owner_id,progress,partner_id,status,stage,created_at,updated_at FROM transformations`
	args := []any{}
	if ownerID != "" {
		q += ` WHERE owner_id=$1`
		args = append(args, ownerID)
	}
	q += ` ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, q, args...)
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

func (r *pgTransRepo) Update(ctx context.Context, t domain.Transformation) (domain.Transformation, error) {
	t.UpdatedAt = time.Now()
	_, err := r.pool.Exec(ctx,
		`UPDATE transformations SET title=$1,achievement_id=$2,progress=$3,partner_id=$4,status=$5,stage=$6,updated_at=$7 WHERE id=$8`,
		t.Title, t.AchievementID, t.Progress, t.PartnerID, t.Status, t.Stage, t.UpdatedAt, t.ID)
	return t, err
}

func (r *pgTransRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM transformations WHERE id=$1`, id)
	return err
}
