package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)

// ---- Competition ----

type compRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewCompetitionRepository() repository.CompetitionRepository { return &compRepo{pool: s.Pool()} }

func (r *compRepo) Create(c domain.Competition) (domain.Competition, error) {
	c.CreatedAt = time.Now(); c.UpdatedAt = c.CreatedAt
	_, err := r.pool.Exec(context.Background(),
		`INSERT INTO competitions (id,title,category,description,location,start_date,end_date,max_teams,reg_count,sponsor,status,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		c.ID, c.Title, c.Category, c.Description, c.Location, c.StartDate, c.EndDate, c.MaxTeams, c.RegCount, c.Sponsor, c.Status, c.CreatedAt, c.UpdatedAt)
	return c, err
}
func (r *compRepo) FindByID(id string) (domain.Competition, error) {
	var c domain.Competition
	err := r.pool.QueryRow(context.Background(),
		`SELECT id,title,category,description,location,start_date,end_date,max_teams,reg_count,sponsor,status,created_at,updated_at FROM competitions WHERE id=$1`, id).
		Scan(&c.ID, &c.Title, &c.Category, &c.Description, &c.Location, &c.StartDate, &c.EndDate, &c.MaxTeams, &c.RegCount, &c.Sponsor, &c.Status, &c.CreatedAt, &c.UpdatedAt)
	return c, err
}
func (r *compRepo) List(offset, limit int) ([]domain.Competition, int, error) {
	var total int
	r.pool.QueryRow(context.Background(), `SELECT COUNT(*) FROM competitions`).Scan(&total)
	rows, err := r.pool.Query(context.Background(),
		`SELECT id,title,category,description,location,start_date,end_date,max_teams,reg_count,sponsor,status,created_at,updated_at FROM competitions ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil { return nil, 0, fmt.Errorf("list competitions: %w", err) }
	defer rows.Close()
	var out []domain.Competition
	for rows.Next() {
		var c domain.Competition
		rows.Scan(&c.ID, &c.Title, &c.Category, &c.Description, &c.Location, &c.StartDate, &c.EndDate, &c.MaxTeams, &c.RegCount, &c.Sponsor, &c.Status, &c.CreatedAt, &c.UpdatedAt)
		out = append(out, c)
	}
	return out, total, rows.Err()
}
func (r *compRepo) Update(c domain.Competition) (domain.Competition, error) {
	c.UpdatedAt = time.Now()
	_, err := r.pool.Exec(context.Background(),
		`UPDATE competitions SET title=$1,category=$2,description=$3,location=$4,start_date=$5,end_date=$6,max_teams=$7,reg_count=$8,sponsor=$9,status=$10,updated_at=$11 WHERE id=$12`,
		c.Title, c.Category, c.Description, c.Location, c.StartDate, c.EndDate, c.MaxTeams, c.RegCount, c.Sponsor, c.Status, c.UpdatedAt, c.ID)
	return c, err
}

func (r *compRepo) Delete(id string) error {
	_, err := r.pool.Exec(context.Background(), `DELETE FROM competitions WHERE id=$1`, id)
	return err
}

func (r *compRepo) CreateReg(reg domain.CompetitionReg) (domain.CompetitionReg, error) {
	reg.CreatedAt = time.Now()
	_, err := r.pool.Exec(context.Background(),
		`INSERT INTO competition_registrations (id,competition_id,user_id,team_name,member_count,contact_info,status,created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		reg.ID, reg.CompetitionID, reg.UserID, reg.TeamName, reg.MemberCount, reg.ContactInfo, reg.Status, reg.CreatedAt)
	return reg, err
}
func (r *compRepo) ListRegs(competitionID string) ([]domain.CompetitionReg, error) {
	rows, err := r.pool.Query(context.Background(),
		`SELECT id,competition_id,user_id,team_name,member_count,contact_info,status,created_at FROM competition_registrations WHERE competition_id=$1 ORDER BY created_at DESC`, competitionID)
	if err != nil { return nil, fmt.Errorf("list regs: %w", err) }
	defer rows.Close()
	var out []domain.CompetitionReg
	for rows.Next() {
		var reg domain.CompetitionReg
		rows.Scan(&reg.ID, &reg.CompetitionID, &reg.UserID, &reg.TeamName, &reg.MemberCount, &reg.ContactInfo, &reg.Status, &reg.CreatedAt)
		out = append(out, reg)
	}
	return out, rows.Err()
}

// ---- Event ----

type eventRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewEventRepository() repository.EventRepository { return &eventRepo{pool: s.Pool()} }

func (r *eventRepo) Create(e domain.AssociationEvent) (domain.AssociationEvent, error) {
	e.CreatedAt = time.Now(); e.UpdatedAt = e.CreatedAt
	_, err := r.pool.Exec(context.Background(),
		`INSERT INTO association_events (id,title,event_type,description,location,start_time,end_time,max_attendees,reg_count,cover_url,status,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		e.ID, e.Title, e.EventType, e.Description, e.Location, e.StartTime, e.EndTime, e.MaxAttendees, e.RegCount, e.CoverURL, e.Status, e.CreatedAt, e.UpdatedAt)
	return e, err
}
func (r *eventRepo) FindByID(id string) (domain.AssociationEvent, error) {
	var e domain.AssociationEvent
	err := r.pool.QueryRow(context.Background(),
		`SELECT id,title,event_type,description,location,start_time,end_time,max_attendees,reg_count,cover_url,status,created_at,updated_at FROM association_events WHERE id=$1`, id).
		Scan(&e.ID, &e.Title, &e.EventType, &e.Description, &e.Location, &e.StartTime, &e.EndTime, &e.MaxAttendees, &e.RegCount, &e.CoverURL, &e.Status, &e.CreatedAt, &e.UpdatedAt)
	return e, err
}
func (r *eventRepo) List(offset, limit int) ([]domain.AssociationEvent, int, error) {
	var total int
	r.pool.QueryRow(context.Background(), `SELECT COUNT(*) FROM association_events`).Scan(&total)
	rows, err := r.pool.Query(context.Background(),
		`SELECT id,title,event_type,description,location,start_time,end_time,max_attendees,reg_count,cover_url,status,created_at,updated_at FROM association_events ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil { return nil, 0, fmt.Errorf("list events: %w", err) }
	defer rows.Close()
	var out []domain.AssociationEvent
	for rows.Next() {
		var e domain.AssociationEvent
		rows.Scan(&e.ID, &e.Title, &e.EventType, &e.Description, &e.Location, &e.StartTime, &e.EndTime, &e.MaxAttendees, &e.RegCount, &e.CoverURL, &e.Status, &e.CreatedAt, &e.UpdatedAt)
		out = append(out, e)
	}
	return out, total, rows.Err()
}
func (r *eventRepo) Update(e domain.AssociationEvent) (domain.AssociationEvent, error) {
	e.UpdatedAt = time.Now()
	_, err := r.pool.Exec(context.Background(),
		`UPDATE association_events SET title=$1,event_type=$2,description=$3,location=$4,start_time=$5,end_time=$6,max_attendees=$7,reg_count=$8,cover_url=$9,status=$10,updated_at=$11 WHERE id=$12`,
		e.Title, e.EventType, e.Description, e.Location, e.StartTime, e.EndTime, e.MaxAttendees, e.RegCount, e.CoverURL, e.Status, e.UpdatedAt, e.ID)
	return e, err
}

func (r *eventRepo) Delete(id string) error {
	_, err := r.pool.Exec(context.Background(), "DELETE FROM association_events WHERE id=$1", id)
	return err
}

func (r *eventRepo) CreateReg(reg domain.EventRegistration) (domain.EventRegistration, error) {
	reg.CreatedAt = time.Now()
	_, err := r.pool.Exec(context.Background(),
		`INSERT INTO event_registrations (id,event_id,user_id,name,phone,org,status,created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		reg.ID, reg.EventID, reg.UserID, reg.Name, reg.Phone, reg.Org, reg.Status, reg.CreatedAt)
	return reg, err
}
func (r *eventRepo) ListRegs(eventID string) ([]domain.EventRegistration, error) {
	rows, err := r.pool.Query(context.Background(),
		`SELECT id,event_id,user_id,name,phone,org,status,created_at FROM event_registrations WHERE event_id=$1 ORDER BY created_at DESC`, eventID)
	if err != nil { return nil, fmt.Errorf("list event regs: %w", err) }
	defer rows.Close()
	var out []domain.EventRegistration
	for rows.Next() {
		var reg domain.EventRegistration
		rows.Scan(&reg.ID, &reg.EventID, &reg.UserID, &reg.Name, &reg.Phone, &reg.Org, &reg.Status, &reg.CreatedAt)
		out = append(out, reg)
	}
	return out, rows.Err()
}

// ---- Emergency ----

type emergRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewEmergencyRepository() repository.EmergencyRepository { return &emergRepo{pool: s.Pool()} }

func (r *emergRepo) CreateResource(res domain.EmergencyResource) (domain.EmergencyResource, error) {
	res.CreatedAt = time.Now(); res.UpdatedAt = res.CreatedAt
	_, err := r.pool.Exec(context.Background(),
		`INSERT INTO emergency_resources (id,owner_id,name,res_type,specs,quantity,location,contact_info,status,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		res.ID, res.OwnerID, res.Name, res.ResType, res.Specs, res.Quantity, res.Location, res.ContactInfo, res.Status, res.CreatedAt, res.UpdatedAt)
	return res, err
}
func (r *emergRepo) FindResourceByID(id string) (domain.EmergencyResource, error) {
	var res domain.EmergencyResource
	err := r.pool.QueryRow(context.Background(),
		`SELECT id,owner_id,name,res_type,specs,quantity,location,contact_info,status,created_at,updated_at FROM emergency_resources WHERE id=$1`, id).
		Scan(&res.ID, &res.OwnerID, &res.Name, &res.ResType, &res.Specs, &res.Quantity, &res.Location, &res.ContactInfo, &res.Status, &res.CreatedAt, &res.UpdatedAt)
	return res, err
}
func (r *emergRepo) ListResources(offset, limit int) ([]domain.EmergencyResource, int, error) {
	var total int
	r.pool.QueryRow(context.Background(), `SELECT COUNT(*) FROM emergency_resources`).Scan(&total)
	rows, err := r.pool.Query(context.Background(),
		`SELECT id,owner_id,name,res_type,specs,quantity,location,contact_info,status,created_at,updated_at FROM emergency_resources ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil { return nil, 0, fmt.Errorf("list emergency resources: %w", err) }
	defer rows.Close()
	var out []domain.EmergencyResource
	for rows.Next() {
		var res domain.EmergencyResource
		rows.Scan(&res.ID, &res.OwnerID, &res.Name, &res.ResType, &res.Specs, &res.Quantity, &res.Location, &res.ContactInfo, &res.Status, &res.CreatedAt, &res.UpdatedAt)
		out = append(out, res)
	}
	return out, total, rows.Err()
}
func (r *emergRepo) UpdateResource(res domain.EmergencyResource) (domain.EmergencyResource, error) {
	res.UpdatedAt = time.Now()
	_, err := r.pool.Exec(context.Background(),
		`UPDATE emergency_resources SET name=$1,res_type=$2,specs=$3,quantity=$4,location=$5,contact_info=$6,status=$7,updated_at=$8 WHERE id=$9`,
		res.Name, res.ResType, res.Specs, res.Quantity, res.Location, res.ContactInfo, res.Status, res.UpdatedAt, res.ID)
	return res, err
}
func (r *emergRepo) CreateDispatch(d domain.EmergencyDispatch) (domain.EmergencyDispatch, error) {
	d.CreatedAt = time.Now()
	_, err := r.pool.Exec(context.Background(),
		`INSERT INTO emergency_dispatches (id,resource_id,event_desc,location,start_time,end_time,commander,result,status,created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		d.ID, d.ResourceID, d.EventDesc, d.Location, d.StartTime, d.EndTime, d.Commander, d.Result, d.Status, d.CreatedAt)
	return d, err
}
func (r *emergRepo) ListDispatches(offset, limit int) ([]domain.EmergencyDispatch, int, error) {
	var total int
	r.pool.QueryRow(context.Background(), `SELECT COUNT(*) FROM emergency_dispatches`).Scan(&total)
	rows, err := r.pool.Query(context.Background(),
		`SELECT id,resource_id,event_desc,location,start_time,end_time,commander,result,status,created_at FROM emergency_dispatches ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil { return nil, 0, fmt.Errorf("list dispatches: %w", err) }
	defer rows.Close()
	var out []domain.EmergencyDispatch
	for rows.Next() {
		var d domain.EmergencyDispatch
		rows.Scan(&d.ID, &d.ResourceID, &d.EventDesc, &d.Location, &d.StartTime, &d.EndTime, &d.Commander, &d.Result, &d.Status, &d.CreatedAt)
		out = append(out, d)
	}
	return out, total, rows.Err()
}

func (r *emergRepo) DeleteResource(id string) error {
	_, err := r.pool.Exec(context.Background(), "DELETE FROM emergency_resources WHERE id=$1", id)
	return err
}

func (r *emergRepo) FindDispatchByID(id string) (domain.EmergencyDispatch, error) {
	var d domain.EmergencyDispatch
	err := r.pool.QueryRow(context.Background(), "SELECT id,resource_id,event_desc,location,start_time,end_time,commander,result,status,created_at FROM emergency_dispatches WHERE id=$1", id).
		Scan(&d.ID, &d.ResourceID, &d.EventDesc, &d.Location, &d.StartTime, &d.EndTime, &d.Commander, &d.Result, &d.Status, &d.CreatedAt)
	return d, err
}

func (r *emergRepo) UpdateDispatch(d domain.EmergencyDispatch) (domain.EmergencyDispatch, error) {
	_, err := r.pool.Exec(context.Background(),
		"UPDATE emergency_dispatches SET resource_id=$1,event_desc=$2,location=$3,start_time=$4,end_time=$5,commander=$6,result=$7,status=$8 WHERE id=$9",
		d.ResourceID, d.EventDesc, d.Location, d.StartTime, d.EndTime, d.Commander, d.Result, d.Status, d.ID)
	return d, err
}

func (r *emergRepo) DeleteDispatch(id string) error {
	_, err := r.pool.Exec(context.Background(), "DELETE FROM emergency_dispatches WHERE id=$1", id)
	return err
}
