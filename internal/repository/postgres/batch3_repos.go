package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)

// PG implementations for the batch-2/3 modules that previously fell back to
// in-memory storage even in PG mode (resource pools, cooperation programs,
// rescue cases, emergency departments, association members).

// ---- ResourcePool ----

type pgPoolRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewResourcePoolRepository() repository.ResourcePoolRepository {
	return &pgPoolRepo{pool: s.Pool()}
}

func (r *pgPoolRepo) Create(ctx context.Context, p domain.ResourcePool) (domain.ResourcePool, error) {
	p.CreatedAt = time.Now()
	p.UpdatedAt = p.CreatedAt
	resources, err := json.Marshal(p.Resources)
	if err != nil {
		return domain.ResourcePool{}, fmt.Errorf("marshal resources: %w", err)
	}
	_, err = r.pool.Exec(ctx,
		`INSERT INTO resource_pools (id,name,pool_type,description,owner_id,resources,status,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		p.ID, p.Name, p.PoolType, p.Description, p.OwnerID, resources, p.Status, p.CreatedAt, p.UpdatedAt)
	return p, err
}

func (r *pgPoolRepo) FindByID(ctx context.Context, id string) (domain.ResourcePool, error) {
	var p domain.ResourcePool
	var resources []byte
	err := r.pool.QueryRow(ctx,
		`SELECT id,name,pool_type,description,owner_id,resources,status,created_at,updated_at FROM resource_pools WHERE id=$1`, id).
		Scan(&p.ID, &p.Name, &p.PoolType, &p.Description, &p.OwnerID, &resources, &p.Status, &p.CreatedAt, &p.UpdatedAt)
	json.Unmarshal(resources, &p.Resources)
	return p, err
}

func (r *pgPoolRepo) List(ctx context.Context, poolType string) ([]domain.ResourcePool, error) {
	q := `SELECT id,name,pool_type,description,owner_id,resources,status,created_at,updated_at FROM resource_pools`
	args := []any{}
	if poolType != "" {
		q += ` WHERE pool_type=$1`
		args = append(args, poolType)
	}
	q += ` ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.ResourcePool
	for rows.Next() {
		var p domain.ResourcePool
		var resources []byte
		if err := rows.Scan(&p.ID, &p.Name, &p.PoolType, &p.Description, &p.OwnerID, &resources, &p.Status, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		json.Unmarshal(resources, &p.Resources)
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *pgPoolRepo) AddMember(ctx context.Context, m domain.ResourcePoolMember) (domain.ResourcePoolMember, error) {
	m.JoinedAt = time.Now()
	_, err := r.pool.Exec(ctx,
		`INSERT INTO resource_pool_members (id,pool_id,res_id,res_type,quantity,status,joined_at) VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		m.ID, m.PoolID, m.ResID, m.ResType, m.Quantity, m.Status, m.JoinedAt)
	return m, err
}

func (r *pgPoolRepo) ListMembers(ctx context.Context, poolID string) ([]domain.ResourcePoolMember, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id,pool_id,res_id,res_type,quantity,status,joined_at FROM resource_pool_members WHERE pool_id=$1 ORDER BY joined_at DESC`, poolID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.ResourcePoolMember
	for rows.Next() {
		var m domain.ResourcePoolMember
		if err := rows.Scan(&m.ID, &m.PoolID, &m.ResID, &m.ResType, &m.Quantity, &m.Status, &m.JoinedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// ---- CooperationProgram ----

type pgCoopRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewCooperationRepository() repository.CooperationRepository {
	return &pgCoopRepo{pool: s.Pool()}
}

func (r *pgCoopRepo) Create(ctx context.Context, c domain.CooperationProgram) (domain.CooperationProgram, error) {
	c.CreatedAt = time.Now()
	c.UpdatedAt = c.CreatedAt
	_, err := r.pool.Exec(ctx,
		`INSERT INTO cooperation_programs (id,title,college_id,enterprise_id,coop_type,description,start_date,end_date,student_quota,status,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		c.ID, c.Title, c.CollegeID, c.EnterpriseID, c.CoopType, c.Description, c.StartDate, c.EndDate, c.StudentQuota, c.Status, c.CreatedAt, c.UpdatedAt)
	return c, err
}

func (r *pgCoopRepo) FindByID(ctx context.Context, id string) (domain.CooperationProgram, error) {
	var c domain.CooperationProgram
	err := r.pool.QueryRow(ctx,
		`SELECT id,title,college_id,enterprise_id,coop_type,description,COALESCE(start_date,'1970-01-01 00:00:00+00'::timestamptz),COALESCE(end_date,'1970-01-01 00:00:00+00'::timestamptz),student_quota,status,created_at,updated_at FROM cooperation_programs WHERE id=$1`, id).
		Scan(&c.ID, &c.Title, &c.CollegeID, &c.EnterpriseID, &c.CoopType, &c.Description, &c.StartDate, &c.EndDate, &c.StudentQuota, &c.Status, &c.CreatedAt, &c.UpdatedAt)
	return c, err
}

func (r *pgCoopRepo) List(ctx context.Context, enterpriseID string) ([]domain.CooperationProgram, error) {
	q := `SELECT id,title,college_id,enterprise_id,coop_type,description,COALESCE(start_date,'1970-01-01 00:00:00+00'::timestamptz),COALESCE(end_date,'1970-01-01 00:00:00+00'::timestamptz),student_quota,status,created_at,updated_at FROM cooperation_programs`
	args := []any{}
	if enterpriseID != "" {
		q += ` WHERE enterprise_id=$1`
		args = append(args, enterpriseID)
	}
	q += ` ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.CooperationProgram
	for rows.Next() {
		var c domain.CooperationProgram
		if err := rows.Scan(&c.ID, &c.Title, &c.CollegeID, &c.EnterpriseID, &c.CoopType, &c.Description, &c.StartDate, &c.EndDate, &c.StudentQuota, &c.Status, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *pgCoopRepo) UpdateStatus(ctx context.Context, id, status string) (domain.CooperationProgram, error) {
	var c domain.CooperationProgram
	err := r.pool.QueryRow(ctx,
		`UPDATE cooperation_programs SET status=$1,updated_at=$2 WHERE id=$3 RETURNING id,title,college_id,enterprise_id,coop_type,description,COALESCE(start_date,'1970-01-01 00:00:00+00'::timestamptz),COALESCE(end_date,'1970-01-01 00:00:00+00'::timestamptz),student_quota,status,created_at,updated_at`,
		status, time.Now(), id).
		Scan(&c.ID, &c.Title, &c.CollegeID, &c.EnterpriseID, &c.CoopType, &c.Description, &c.StartDate, &c.EndDate, &c.StudentQuota, &c.Status, &c.CreatedAt, &c.UpdatedAt)
	return c, err
}

// ---- RescueCase ----

type pgRescueRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewRescueCaseRepository() repository.RescueCaseRepository {
	return &pgRescueRepo{pool: s.Pool()}
}

func (r *pgRescueRepo) Create(ctx context.Context, c domain.RescueCase) (domain.RescueCase, error) {
	c.CreatedAt = time.Now()
	c.UpdatedAt = c.CreatedAt
	media, err := json.Marshal(c.MediaURLs)
	if err != nil {
		return domain.RescueCase{}, fmt.Errorf("marshal media_urls: %w", err)
	}
	_, err = r.pool.Exec(ctx,
		`INSERT INTO rescue_cases (id,title,event_type,location,date,drone_model,team_name,summary,result,lessons,media_urls,source,status,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		c.ID, c.Title, c.EventType, c.Location, c.Date, c.DroneModel, c.TeamName, c.Summary, c.Result, c.Lessons, media, c.Source, c.Status, c.CreatedAt, c.UpdatedAt)
	return c, err
}

func (r *pgRescueRepo) FindByID(ctx context.Context, id string) (domain.RescueCase, error) {
	var c domain.RescueCase
	var media []byte
	err := r.pool.QueryRow(ctx,
		`SELECT id,title,event_type,location,COALESCE(date,'1970-01-01 00:00:00+00'::timestamptz),drone_model,team_name,summary,result,lessons,media_urls,source,status,created_at,updated_at FROM rescue_cases WHERE id=$1`, id).
		Scan(&c.ID, &c.Title, &c.EventType, &c.Location, &c.Date, &c.DroneModel, &c.TeamName, &c.Summary, &c.Result, &c.Lessons, &media, &c.Source, &c.Status, &c.CreatedAt, &c.UpdatedAt)
	json.Unmarshal(media, &c.MediaURLs)
	return c, err
}

func (r *pgRescueRepo) List(ctx context.Context, eventType, q string, offset, limit int) ([]domain.RescueCase, int, error) {
	where := "WHERE status='published'"
	args := []any{}
	if eventType != "" {
		where += " AND event_type=$1"
		args = append(args, eventType)
	}
	if q = strings.TrimSpace(q); q != "" {
		if len(q) > 100 {
			q = q[:100]
		}
		args = append(args, "%"+escapeLike(q)+"%")
		if where == "" {
			where = "WHERE "
		} else {
			where += " AND "
		}
		where += fmt.Sprintf(`(title ILIKE $%d ESCAPE '\' OR location ILIKE $%d ESCAPE '\' OR summary ILIKE $%d ESCAPE '\' OR team_name ILIKE $%d ESCAPE '\' OR drone_model ILIKE $%d ESCAPE '\')`,
			len(args), len(args), len(args), len(args), len(args))
	}
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM rescue_cases `+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count rescue cases: %w", err)
	}
	qStr := fmt.Sprintf(`SELECT id,title,event_type,location,COALESCE(date,'1970-01-01 00:00:00+00'::timestamptz),drone_model,team_name,summary,result,lessons,media_urls,source,status,created_at,updated_at FROM rescue_cases %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, where, len(args)+1, len(args)+2)
	rows, err := r.pool.Query(ctx, qStr, append(args, limit, offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []domain.RescueCase
	for rows.Next() {
		var c domain.RescueCase
		var media []byte
		if err := rows.Scan(&c.ID, &c.Title, &c.EventType, &c.Location, &c.Date, &c.DroneModel, &c.TeamName, &c.Summary, &c.Result, &c.Lessons, &media, &c.Source, &c.Status, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, 0, err
		}
		json.Unmarshal(media, &c.MediaURLs)
		out = append(out, c)
	}
	return out, total, rows.Err()
}

// ---- EmergencyDept ----

type pgEmergDeptRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewEmergencyDeptRepository() repository.EmergencyDeptRepository {
	return &pgEmergDeptRepo{pool: s.Pool()}
}

func (r *pgEmergDeptRepo) CreateDept(ctx context.Context, d domain.EmergencyDept) (domain.EmergencyDept, error) {
	d.CreatedAt = time.Now()
	d.UpdatedAt = d.CreatedAt
	_, err := r.pool.Exec(ctx,
		`INSERT INTO emergency_depts (id,name,dept_type,region,contact_name,contact_phone,protocol_url,status,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		d.ID, d.Name, d.DeptType, d.Region, d.ContactName, d.ContactPhone, d.ProtocolURL, d.Status, d.CreatedAt, d.UpdatedAt)
	return d, err
}

func (r *pgEmergDeptRepo) ListDepts(ctx context.Context) ([]domain.EmergencyDept, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id,name,dept_type,region,contact_name,contact_phone,protocol_url,status,created_at,updated_at FROM emergency_depts ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.EmergencyDept
	for rows.Next() {
		var d domain.EmergencyDept
		if err := rows.Scan(&d.ID, &d.Name, &d.DeptType, &d.Region, &d.ContactName, &d.ContactPhone, &d.ProtocolURL, &d.Status, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (r *pgEmergDeptRepo) CreateDrill(ctx context.Context, d domain.EmergencyDrill) (domain.EmergencyDrill, error) {
	d.CreatedAt = time.Now()
	_, err := r.pool.Exec(ctx,
		`INSERT INTO emergency_drills (id,dept_id,title,scenario,date,participants,drone_count,result,created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		d.ID, d.DeptID, d.Title, d.Scenario, d.Date, d.Participants, d.DroneCount, d.Result, d.CreatedAt)
	return d, err
}

func (r *pgEmergDeptRepo) ListDrills(ctx context.Context, deptID string) ([]domain.EmergencyDrill, error) {
	q := `SELECT id,dept_id,title,scenario,COALESCE(date,'1970-01-01 00:00:00+00'::timestamptz),participants,drone_count,result,created_at FROM emergency_drills`
	args := []any{}
	if deptID != "" {
		q += ` WHERE dept_id=$1`
		args = append(args, deptID)
	}
	q += ` ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.EmergencyDrill
	for rows.Next() {
		var d domain.EmergencyDrill
		if err := rows.Scan(&d.ID, &d.DeptID, &d.Title, &d.Scenario, &d.Date, &d.Participants, &d.DroneCount, &d.Result, &d.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

// ---- AssociationMember ----

type pgAssocMemberRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewAssociationMemberRepository() repository.AssociationMemberRepository {
	return &pgAssocMemberRepo{pool: s.Pool()}
}

func (r *pgAssocMemberRepo) Create(ctx context.Context, m domain.AssociationMember) (domain.AssociationMember, error) {
	m.CreatedAt = time.Now()
	m.UpdatedAt = m.CreatedAt
	if m.JoinDate.IsZero() {
		m.JoinDate = m.CreatedAt
	}
	_, err := r.pool.Exec(ctx,
		`INSERT INTO association_members (id,user_id,enterprise_id,role,join_date,expire_date,status,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		m.ID, m.UserID, m.EnterpriseID, string(m.Role), m.JoinDate, m.ExpireDate, m.Status, m.CreatedAt, m.UpdatedAt)
	return m, err
}

func (r *pgAssocMemberRepo) FindByUserID(ctx context.Context, userID string) (domain.AssociationMember, error) {
	var m domain.AssociationMember
	err := r.pool.QueryRow(ctx,
		`SELECT id,user_id,enterprise_id,role,join_date,COALESCE(expire_date,'1970-01-01 00:00:00+00'::timestamptz),status,created_at,updated_at FROM association_members WHERE user_id=$1 ORDER BY created_at DESC LIMIT 1`, userID).
		Scan(&m.ID, &m.UserID, &m.EnterpriseID, &m.Role, &m.JoinDate, &m.ExpireDate, &m.Status, &m.CreatedAt, &m.UpdatedAt)
	return m, err
}

func (r *pgAssocMemberRepo) List(ctx context.Context, role string, offset, limit int) ([]domain.AssociationMember, int, error) {
	where := ""
	args := []any{}
	if role != "" {
		where = `WHERE role=$1`
		args = append(args, role)
	}
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM association_members `+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count association members: %w", err)
	}
	q := fmt.Sprintf(`SELECT id,user_id,enterprise_id,role,join_date,COALESCE(expire_date,'1970-01-01 00:00:00+00'::timestamptz),status,created_at,updated_at FROM association_members %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, where, len(args)+1, len(args)+2)
	rows, err := r.pool.Query(ctx, q, append(args, limit, offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []domain.AssociationMember
	for rows.Next() {
		var m domain.AssociationMember
		if err := rows.Scan(&m.ID, &m.UserID, &m.EnterpriseID, &m.Role, &m.JoinDate, &m.ExpireDate, &m.Status, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, m)
	}
	return out, total, rows.Err()
}

func (r *pgAssocMemberRepo) UpdateRole(ctx context.Context, id string, role domain.AssociationRole) (domain.AssociationMember, error) {
	var m domain.AssociationMember
	err := r.pool.QueryRow(ctx,
		`UPDATE association_members SET role=$1,updated_at=$2 WHERE id=$3 RETURNING id,user_id,enterprise_id,role,join_date,COALESCE(expire_date,'1970-01-01 00:00:00+00'::timestamptz),status,created_at,updated_at`,
		string(role), time.Now(), id).
		Scan(&m.ID, &m.UserID, &m.EnterpriseID, &m.Role, &m.JoinDate, &m.ExpireDate, &m.Status, &m.CreatedAt, &m.UpdatedAt)
	return m, err
}
