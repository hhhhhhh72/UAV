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

// PG implementations for the batch-2/3 modules that previously fell back to
// in-memory storage even in PG mode (resource pools, cooperation programs,
// rescue cases, emergency departments, association members).

// ---- ResourcePool ----

type pgPoolRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewResourcePoolRepository() repository.ResourcePoolRepository {
	return &pgPoolRepo{pool: s.Pool()}
}

func (r *pgPoolRepo) Create(p domain.ResourcePool) (domain.ResourcePool, error) {
	p.CreatedAt = time.Now(); p.UpdatedAt = p.CreatedAt
	resources, err := json.Marshal(p.Resources)
	if err != nil { return domain.ResourcePool{}, fmt.Errorf("marshal resources: %w", err) }
	_, err = r.pool.Exec(context.Background(),
		`INSERT INTO resource_pools (id,name,pool_type,description,owner_id,resources,status,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		p.ID, p.Name, p.PoolType, p.Description, p.OwnerID, resources, p.Status, p.CreatedAt, p.UpdatedAt)
	return p, err
}

func (r *pgPoolRepo) FindByID(id string) (domain.ResourcePool, error) {
	var p domain.ResourcePool; var resources []byte
	err := r.pool.QueryRow(context.Background(),
		`SELECT id,name,pool_type,description,owner_id,resources,status,created_at,updated_at FROM resource_pools WHERE id=$1`, id).
		Scan(&p.ID, &p.Name, &p.PoolType, &p.Description, &p.OwnerID, &resources, &p.Status, &p.CreatedAt, &p.UpdatedAt)
	json.Unmarshal(resources, &p.Resources)
	return p, err
}

func (r *pgPoolRepo) List(poolType string) ([]domain.ResourcePool, error) {
	q := `SELECT id,name,pool_type,description,owner_id,resources,status,created_at,updated_at FROM resource_pools`
	args := []any{}
	if poolType != "" { q += ` WHERE pool_type=$1`; args = append(args, poolType) }
	q += ` ORDER BY created_at DESC`
	rows, err := r.pool.Query(context.Background(), q, args...)
	if err != nil { return nil, err }
	defer rows.Close()
	var out []domain.ResourcePool
	for rows.Next() {
		var p domain.ResourcePool; var resources []byte
		if err := rows.Scan(&p.ID, &p.Name, &p.PoolType, &p.Description, &p.OwnerID, &resources, &p.Status, &p.CreatedAt, &p.UpdatedAt); err != nil { return nil, err }
		json.Unmarshal(resources, &p.Resources)
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *pgPoolRepo) AddMember(m domain.ResourcePoolMember) (domain.ResourcePoolMember, error) {
	m.JoinedAt = time.Now()
	_, err := r.pool.Exec(context.Background(),
		`INSERT INTO resource_pool_members (id,pool_id,res_id,res_type,quantity,status,joined_at) VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		m.ID, m.PoolID, m.ResID, m.ResType, m.Quantity, m.Status, m.JoinedAt)
	return m, err
}

func (r *pgPoolRepo) ListMembers(poolID string) ([]domain.ResourcePoolMember, error) {
	rows, err := r.pool.Query(context.Background(),
		`SELECT id,pool_id,res_id,res_type,quantity,status,joined_at FROM resource_pool_members WHERE pool_id=$1 ORDER BY joined_at DESC`, poolID)
	if err != nil { return nil, err }
	defer rows.Close()
	var out []domain.ResourcePoolMember
	for rows.Next() {
		var m domain.ResourcePoolMember
		if err := rows.Scan(&m.ID, &m.PoolID, &m.ResID, &m.ResType, &m.Quantity, &m.Status, &m.JoinedAt); err != nil { return nil, err }
		out = append(out, m)
	}
	return out, rows.Err()
}

// ---- CooperationProgram ----

type pgCoopRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewCooperationRepository() repository.CooperationRepository {
	return &pgCoopRepo{pool: s.Pool()}
}

func (r *pgCoopRepo) Create(c domain.CooperationProgram) (domain.CooperationProgram, error) {
	c.CreatedAt = time.Now(); c.UpdatedAt = c.CreatedAt
	_, err := r.pool.Exec(context.Background(),
		`INSERT INTO cooperation_programs (id,title,college_id,enterprise_id,coop_type,description,start_date,end_date,student_quota,status,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		c.ID, c.Title, c.CollegeID, c.EnterpriseID, c.CoopType, c.Description, c.StartDate, c.EndDate, c.StudentQuota, c.Status, c.CreatedAt, c.UpdatedAt)
	return c, err
}

func (r *pgCoopRepo) FindByID(id string) (domain.CooperationProgram, error) {
	var c domain.CooperationProgram
	err := r.pool.QueryRow(context.Background(),
		`SELECT id,title,college_id,enterprise_id,coop_type,description,start_date,end_date,student_quota,status,created_at,updated_at FROM cooperation_programs WHERE id=$1`, id).
		Scan(&c.ID, &c.Title, &c.CollegeID, &c.EnterpriseID, &c.CoopType, &c.Description, &c.StartDate, &c.EndDate, &c.StudentQuota, &c.Status, &c.CreatedAt, &c.UpdatedAt)
	return c, err
}

func (r *pgCoopRepo) List(enterpriseID string) ([]domain.CooperationProgram, error) {
	q := `SELECT id,title,college_id,enterprise_id,coop_type,description,start_date,end_date,student_quota,status,created_at,updated_at FROM cooperation_programs`
	args := []any{}
	if enterpriseID != "" { q += ` WHERE enterprise_id=$1`; args = append(args, enterpriseID) }
	q += ` ORDER BY created_at DESC`
	rows, err := r.pool.Query(context.Background(), q, args...)
	if err != nil { return nil, err }
	defer rows.Close()
	var out []domain.CooperationProgram
	for rows.Next() {
		var c domain.CooperationProgram
		if err := rows.Scan(&c.ID, &c.Title, &c.CollegeID, &c.EnterpriseID, &c.CoopType, &c.Description, &c.StartDate, &c.EndDate, &c.StudentQuota, &c.Status, &c.CreatedAt, &c.UpdatedAt); err != nil { return nil, err }
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *pgCoopRepo) UpdateStatus(id, status string) (domain.CooperationProgram, error) {
	var c domain.CooperationProgram
	err := r.pool.QueryRow(context.Background(),
		`UPDATE cooperation_programs SET status=$1,updated_at=$2 WHERE id=$3 RETURNING id,title,college_id,enterprise_id,coop_type,description,start_date,end_date,student_quota,status,created_at,updated_at`,
		status, time.Now(), id).
		Scan(&c.ID, &c.Title, &c.CollegeID, &c.EnterpriseID, &c.CoopType, &c.Description, &c.StartDate, &c.EndDate, &c.StudentQuota, &c.Status, &c.CreatedAt, &c.UpdatedAt)
	return c, err
}

// ---- RescueCase ----

type pgRescueRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewRescueCaseRepository() repository.RescueCaseRepository {
	return &pgRescueRepo{pool: s.Pool()}
}

func (r *pgRescueRepo) Create(c domain.RescueCase) (domain.RescueCase, error) {
	c.CreatedAt = time.Now(); c.UpdatedAt = c.CreatedAt
	media, err := json.Marshal(c.MediaURLs)
	if err != nil { return domain.RescueCase{}, fmt.Errorf("marshal media_urls: %w", err) }
	_, err = r.pool.Exec(context.Background(),
		`INSERT INTO rescue_cases (id,title,event_type,location,date,drone_model,team_name,summary,result,lessons,media_urls,source,status,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		c.ID, c.Title, c.EventType, c.Location, c.Date, c.DroneModel, c.TeamName, c.Summary, c.Result, c.Lessons, media, c.Source, c.Status, c.CreatedAt, c.UpdatedAt)
	return c, err
}

func (r *pgRescueRepo) FindByID(id string) (domain.RescueCase, error) {
	var c domain.RescueCase; var media []byte
	err := r.pool.QueryRow(context.Background(),
		`SELECT id,title,event_type,location,date,drone_model,team_name,summary,result,lessons,media_urls,source,status,created_at,updated_at FROM rescue_cases WHERE id=$1`, id).
		Scan(&c.ID, &c.Title, &c.EventType, &c.Location, &c.Date, &c.DroneModel, &c.TeamName, &c.Summary, &c.Result, &c.Lessons, &media, &c.Source, &c.Status, &c.CreatedAt, &c.UpdatedAt)
	json.Unmarshal(media, &c.MediaURLs)
	return c, err
}

func (r *pgRescueRepo) List(eventType string, offset, limit int) ([]domain.RescueCase, int, error) {
	where := ""; args := []any{}
	if eventType != "" { where = `WHERE event_type=$1`; args = append(args, eventType) }
	var total int
	r.pool.QueryRow(context.Background(), `SELECT COUNT(*) FROM rescue_cases `+where, args...).Scan(&total)
	q := fmt.Sprintf(`SELECT id,title,event_type,location,date,drone_model,team_name,summary,result,lessons,media_urls,source,status,created_at,updated_at FROM rescue_cases %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, where, len(args)+1, len(args)+2)
	rows, err := r.pool.Query(context.Background(), q, append(args, limit, offset)...)
	if err != nil { return nil, 0, err }
	defer rows.Close()
	var out []domain.RescueCase
	for rows.Next() {
		var c domain.RescueCase; var media []byte
		if err := rows.Scan(&c.ID, &c.Title, &c.EventType, &c.Location, &c.Date, &c.DroneModel, &c.TeamName, &c.Summary, &c.Result, &c.Lessons, &media, &c.Source, &c.Status, &c.CreatedAt, &c.UpdatedAt); err != nil { return nil, 0, err }
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

func (r *pgEmergDeptRepo) CreateDept(d domain.EmergencyDept) (domain.EmergencyDept, error) {
	d.CreatedAt = time.Now(); d.UpdatedAt = d.CreatedAt
	_, err := r.pool.Exec(context.Background(),
		`INSERT INTO emergency_depts (id,name,dept_type,region,contact_name,contact_phone,protocol_url,status,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		d.ID, d.Name, d.DeptType, d.Region, d.ContactName, d.ContactPhone, d.ProtocolURL, d.Status, d.CreatedAt, d.UpdatedAt)
	return d, err
}

func (r *pgEmergDeptRepo) ListDepts() ([]domain.EmergencyDept, error) {
	rows, err := r.pool.Query(context.Background(),
		`SELECT id,name,dept_type,region,contact_name,contact_phone,protocol_url,status,created_at,updated_at FROM emergency_depts ORDER BY created_at DESC`)
	if err != nil { return nil, err }
	defer rows.Close()
	var out []domain.EmergencyDept
	for rows.Next() {
		var d domain.EmergencyDept
		if err := rows.Scan(&d.ID, &d.Name, &d.DeptType, &d.Region, &d.ContactName, &d.ContactPhone, &d.ProtocolURL, &d.Status, &d.CreatedAt, &d.UpdatedAt); err != nil { return nil, err }
		out = append(out, d)
	}
	return out, rows.Err()
}

func (r *pgEmergDeptRepo) CreateDrill(d domain.EmergencyDrill) (domain.EmergencyDrill, error) {
	d.CreatedAt = time.Now()
	_, err := r.pool.Exec(context.Background(),
		`INSERT INTO emergency_drills (id,dept_id,title,scenario,date,participants,drone_count,result,created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		d.ID, d.DeptID, d.Title, d.Scenario, d.Date, d.Participants, d.DroneCount, d.Result, d.CreatedAt)
	return d, err
}

func (r *pgEmergDeptRepo) ListDrills(deptID string) ([]domain.EmergencyDrill, error) {
	q := `SELECT id,dept_id,title,scenario,date,participants,drone_count,result,created_at FROM emergency_drills`
	args := []any{}
	if deptID != "" { q += ` WHERE dept_id=$1`; args = append(args, deptID) }
	q += ` ORDER BY created_at DESC`
	rows, err := r.pool.Query(context.Background(), q, args...)
	if err != nil { return nil, err }
	defer rows.Close()
	var out []domain.EmergencyDrill
	for rows.Next() {
		var d domain.EmergencyDrill
		if err := rows.Scan(&d.ID, &d.DeptID, &d.Title, &d.Scenario, &d.Date, &d.Participants, &d.DroneCount, &d.Result, &d.CreatedAt); err != nil { return nil, err }
		out = append(out, d)
	}
	return out, rows.Err()
}

// ---- AssociationMember ----

type pgAssocMemberRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewAssociationMemberRepository() repository.AssociationMemberRepository {
	return &pgAssocMemberRepo{pool: s.Pool()}
}

func (r *pgAssocMemberRepo) Create(m domain.AssociationMember) (domain.AssociationMember, error) {
	m.CreatedAt = time.Now(); m.UpdatedAt = m.CreatedAt
	if m.JoinDate.IsZero() { m.JoinDate = m.CreatedAt }
	_, err := r.pool.Exec(context.Background(),
		`INSERT INTO association_members (id,user_id,enterprise_id,role,join_date,expire_date,status,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		m.ID, m.UserID, m.EnterpriseID, string(m.Role), m.JoinDate, m.ExpireDate, m.Status, m.CreatedAt, m.UpdatedAt)
	return m, err
}

func (r *pgAssocMemberRepo) FindByUserID(userID string) (domain.AssociationMember, error) {
	var m domain.AssociationMember
	err := r.pool.QueryRow(context.Background(),
		`SELECT id,user_id,enterprise_id,role,join_date,expire_date,status,created_at,updated_at FROM association_members WHERE user_id=$1 ORDER BY created_at DESC LIMIT 1`, userID).
		Scan(&m.ID, &m.UserID, &m.EnterpriseID, &m.Role, &m.JoinDate, &m.ExpireDate, &m.Status, &m.CreatedAt, &m.UpdatedAt)
	return m, err
}

func (r *pgAssocMemberRepo) List(role string, offset, limit int) ([]domain.AssociationMember, int, error) {
	where := ""; args := []any{}
	if role != "" { where = `WHERE role=$1`; args = append(args, role) }
	var total int
	r.pool.QueryRow(context.Background(), `SELECT COUNT(*) FROM association_members `+where, args...).Scan(&total)
	q := fmt.Sprintf(`SELECT id,user_id,enterprise_id,role,join_date,expire_date,status,created_at,updated_at FROM association_members %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, where, len(args)+1, len(args)+2)
	rows, err := r.pool.Query(context.Background(), q, append(args, limit, offset)...)
	if err != nil { return nil, 0, err }
	defer rows.Close()
	var out []domain.AssociationMember
	for rows.Next() {
		var m domain.AssociationMember
		if err := rows.Scan(&m.ID, &m.UserID, &m.EnterpriseID, &m.Role, &m.JoinDate, &m.ExpireDate, &m.Status, &m.CreatedAt, &m.UpdatedAt); err != nil { return nil, 0, err }
		out = append(out, m)
	}
	return out, total, rows.Err()
}

func (r *pgAssocMemberRepo) UpdateRole(id string, role domain.AssociationRole) (domain.AssociationMember, error) {
	var m domain.AssociationMember
	err := r.pool.QueryRow(context.Background(),
		`UPDATE association_members SET role=$1,updated_at=$2 WHERE id=$3 RETURNING id,user_id,enterprise_id,role,join_date,expire_date,status,created_at,updated_at`,
		string(role), time.Now(), id).
		Scan(&m.ID, &m.UserID, &m.EnterpriseID, &m.Role, &m.JoinDate, &m.ExpireDate, &m.Status, &m.CreatedAt, &m.UpdatedAt)
	return m, err
}
