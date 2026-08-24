package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"drone-platform/internal/crypto"
	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)

// ---- Competition ----

type compRepo struct {
	pool   *pgxpool.Pool
	cipher *crypto.Cipher // 报名实名信息（id_card/phone）静态加密（仿 pilotRepo）
}

func (s *Store) NewCompetitionRepository(cipher *crypto.Cipher) repository.CompetitionRepository {
	return &compRepo{pool: s.Pool(), cipher: cipher}
}

// compCols 与 competitions 表列一一对应（迁移 000044 补齐小程序页面字段，000069 补 original_fee）
const compCols = `id,title,category,description,location,start_date,end_date,deadline,max_teams,reg_count,sponsor,organizer_sub,fee,min_fee,original_fee,tags,poster,requirements,events,prizes,registration_status,status,created_at,updated_at`

func (r *compRepo) Create(ctx context.Context, c domain.Competition) (domain.Competition, error) {
	c.CreatedAt = time.Now()
	c.UpdatedAt = c.CreatedAt
	c.Tags = jsonbSlice(c.Tags)
	c.Requirements = jsonbSlice(c.Requirements)
	c.Events = jsonbSlice(c.Events)
	c.Prizes = jsonbSlice(c.Prizes)
	_, err := r.pool.Exec(ctx,
		`INSERT INTO competitions (`+compCols+`) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24)`,
		c.ID, c.Title, c.Category, c.Description, c.Location, c.StartDate, c.EndDate, c.Deadline, c.MaxTeams, c.RegCount,
		c.Sponsor, c.OrganizerSub, c.Fee, c.MinFee, c.OriginalFee, c.Tags, c.Poster, c.Requirements, c.Events, c.Prizes,
		c.RegistrationStatus, c.Status, c.CreatedAt, c.UpdatedAt)
	return c, err
}
func (r *compRepo) FindByID(ctx context.Context, id string) (domain.Competition, error) {
	var c domain.Competition
	err := r.pool.QueryRow(ctx,
		`SELECT `+compCols+` FROM competitions WHERE id=$1`, id).
		Scan(&c.ID, &c.Title, &c.Category, &c.Description, &c.Location, &c.StartDate, &c.EndDate, &c.Deadline, &c.MaxTeams, &c.RegCount,
			&c.Sponsor, &c.OrganizerSub, &c.Fee, &c.MinFee, &c.OriginalFee, &c.Tags, &c.Poster, &c.Requirements, &c.Events, &c.Prizes,
			&c.RegistrationStatus, &c.Status, &c.CreatedAt, &c.UpdatedAt)
	return c, err
}
func (r *compRepo) List(ctx context.Context, offset, limit int) ([]domain.Competition, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM competitions`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count competitions: %w", err)
	}
	rows, err := r.pool.Query(ctx,
		`SELECT `+compCols+` FROM competitions ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list competitions: %w", err)
	}
	defer rows.Close()
	var out []domain.Competition
	for rows.Next() {
		var c domain.Competition
		if err := rows.Scan(&c.ID, &c.Title, &c.Category, &c.Description, &c.Location, &c.StartDate, &c.EndDate, &c.Deadline, &c.MaxTeams, &c.RegCount,
			&c.Sponsor, &c.OrganizerSub, &c.Fee, &c.MinFee, &c.OriginalFee, &c.Tags, &c.Poster, &c.Requirements, &c.Events, &c.Prizes,
			&c.RegistrationStatus, &c.Status, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan competition: %w", err)
		}
		out = append(out, c)
	}
	return out, total, rows.Err()
}
func (r *compRepo) Update(ctx context.Context, c domain.Competition) (domain.Competition, error) {
	c.UpdatedAt = time.Now()
	c.Tags = jsonbSlice(c.Tags)
	c.Requirements = jsonbSlice(c.Requirements)
	c.Events = jsonbSlice(c.Events)
	c.Prizes = jsonbSlice(c.Prizes)
	// reg_count 不参与更新：报名数由报名系统维护，管理端编辑赛事不得清零
	_, err := r.pool.Exec(ctx,
		`UPDATE competitions SET title=$1,category=$2,description=$3,location=$4,start_date=$5,end_date=$6,deadline=$7,max_teams=$8,sponsor=$9,organizer_sub=$10,fee=$11,min_fee=$12,original_fee=$13,tags=$14,poster=$15,requirements=$16,events=$17,prizes=$18,registration_status=$19,status=$20,updated_at=$21 WHERE id=$22`,
		c.Title, c.Category, c.Description, c.Location, c.StartDate, c.EndDate, c.Deadline, c.MaxTeams,
		c.Sponsor, c.OrganizerSub, c.Fee, c.MinFee, c.OriginalFee, c.Tags, c.Poster, c.Requirements, c.Events, c.Prizes,
		c.RegistrationStatus, c.Status, c.UpdatedAt, c.ID)
	return c, err
}

func (r *compRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM competitions WHERE id=$1`, id)
	return err
}

// encRegPII 加密报名实名字段（cipher 为空或失败时保留明文，兼容无 ENCRYPTION_KEY 环境）。
func (r *compRepo) encRegPII(reg *domain.CompetitionReg) {
	if r.cipher == nil {
		return
	}
	if reg.IDCard != "" {
		if enc, err := r.cipher.Encrypt(reg.IDCard); err == nil {
			reg.IDCard = enc
		}
	}
	if reg.Phone != "" {
		if enc, err := r.cipher.Encrypt(reg.Phone); err == nil {
			reg.Phone = enc
		}
	}
}
func (r *compRepo) decRegPII(reg *domain.CompetitionReg) {
	if r.cipher == nil {
		return
	}
	if reg.IDCard != "" {
		if dec, err := r.cipher.Decrypt(reg.IDCard); err == nil {
			reg.IDCard = dec
		}
	}
	if reg.Phone != "" {
		if dec, err := r.cipher.Decrypt(reg.Phone); err == nil {
			reg.Phone = dec
		}
	}
}
func (r *compRepo) CreateReg(ctx context.Context, reg domain.CompetitionReg) (domain.CompetitionReg, error) {
	reg.CreatedAt = time.Now()
	r.encRegPII(&reg)
	// 事务：报名 + 参赛计数自增（reg_count 与 registrations 行数保持一致）。
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		r.decRegPII(&reg)
		return domain.CompetitionReg{}, fmt.Errorf("begin reg tx: %w", err)
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx,
		`INSERT INTO competition_registrations (id,competition_id,user_id,team_name,member_count,contact_info,name,phone,id_card,photo_url,id_card_image,status,created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		reg.ID, reg.CompetitionID, reg.UserID, reg.TeamName, reg.MemberCount, reg.ContactInfo, reg.Name, reg.Phone, reg.IDCard, reg.PhotoURL, reg.IDCardImage, reg.Status, reg.CreatedAt)
	if err != nil {
		// P2 修复：并发重复报名由唯一索引（uniq_competition_regs_user_comp，迁移 000071）
		// 兜底——service 层 ListRegs 预检存在 TOCTOU 窗口，23505 映射为友好错误。
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			r.decRegPII(&reg)
			return domain.CompetitionReg{}, fmt.Errorf("已报名过该赛事，请勿重复报名")
		}
		r.decRegPII(&reg)
		return domain.CompetitionReg{}, fmt.Errorf("insert reg: %w", err)
	}
	if _, err := tx.Exec(ctx, `UPDATE competitions SET reg_count = reg_count + 1 WHERE id=$1`, reg.CompetitionID); err != nil {
		return domain.CompetitionReg{}, fmt.Errorf("bump competition reg_count: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.CompetitionReg{}, fmt.Errorf("commit reg tx: %w", err)
	}
	r.decRegPII(&reg)
	return reg, nil
}
func (r *compRepo) ListRegs(ctx context.Context, competitionID string) ([]domain.CompetitionReg, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id,competition_id,user_id,team_name,member_count,contact_info,name,phone,id_card,photo_url,id_card_image,status,created_at FROM competition_registrations WHERE competition_id=$1 ORDER BY created_at DESC`, competitionID)
	if err != nil {
		return nil, fmt.Errorf("list regs: %w", err)
	}
	defer rows.Close()
	var out []domain.CompetitionReg
	for rows.Next() {
		var reg domain.CompetitionReg
		if err := rows.Scan(&reg.ID, &reg.CompetitionID, &reg.UserID, &reg.TeamName, &reg.MemberCount, &reg.ContactInfo, &reg.Name, &reg.Phone, &reg.IDCard, &reg.PhotoURL, &reg.IDCardImage, &reg.Status, &reg.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan reg: %w", err)
		}
		r.decRegPII(&reg)
		out = append(out, reg)
	}
	return out, rows.Err()
}

// ---- Event ----

type eventRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewEventRepository() repository.EventRepository { return &eventRepo{pool: s.Pool()} }

func (r *eventRepo) Create(ctx context.Context, e domain.AssociationEvent) (domain.AssociationEvent, error) {
	e.CreatedAt = time.Now()
	e.UpdatedAt = e.CreatedAt
	_, err := r.pool.Exec(ctx,
		`INSERT INTO association_events (id,title,event_type,description,location,start_time,end_time,max_attendees,reg_count,cover_url,status,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		e.ID, e.Title, e.EventType, e.Description, e.Location, e.StartTime, e.EndTime, e.MaxAttendees, e.RegCount, e.CoverURL, e.Status, e.CreatedAt, e.UpdatedAt)
	return e, err
}
func (r *eventRepo) FindByID(ctx context.Context, id string) (domain.AssociationEvent, error) {
	var e domain.AssociationEvent
	err := r.pool.QueryRow(ctx,
		`SELECT id,title,event_type,description,location,start_time,end_time,max_attendees,reg_count,cover_url,status,created_at,updated_at FROM association_events WHERE id=$1`, id).
		Scan(&e.ID, &e.Title, &e.EventType, &e.Description, &e.Location, &e.StartTime, &e.EndTime, &e.MaxAttendees, &e.RegCount, &e.CoverURL, &e.Status, &e.CreatedAt, &e.UpdatedAt)
	return e, err
}
func (r *eventRepo) List(ctx context.Context, offset, limit int) ([]domain.AssociationEvent, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM association_events`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count events: %w", err)
	}
	rows, err := r.pool.Query(ctx,
		`SELECT id,title,event_type,description,location,start_time,end_time,max_attendees,reg_count,cover_url,status,created_at,updated_at FROM association_events ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list events: %w", err)
	}
	defer rows.Close()
	var out []domain.AssociationEvent
	for rows.Next() {
		var e domain.AssociationEvent
		if err := rows.Scan(&e.ID, &e.Title, &e.EventType, &e.Description, &e.Location, &e.StartTime, &e.EndTime, &e.MaxAttendees, &e.RegCount, &e.CoverURL, &e.Status, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan event: %w", err)
		}
		out = append(out, e)
	}
	return out, total, rows.Err()
}
func (r *eventRepo) Update(ctx context.Context, e domain.AssociationEvent) (domain.AssociationEvent, error) {
	e.UpdatedAt = time.Now()
	_, err := r.pool.Exec(ctx,
		`UPDATE association_events SET title=$1,event_type=$2,description=$3,location=$4,start_time=$5,end_time=$6,max_attendees=$7,reg_count=$8,cover_url=$9,status=$10,updated_at=$11 WHERE id=$12`,
		e.Title, e.EventType, e.Description, e.Location, e.StartTime, e.EndTime, e.MaxAttendees, e.RegCount, e.CoverURL, e.Status, e.UpdatedAt, e.ID)
	return e, err
}

func (r *eventRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM association_events WHERE id=$1", id)
	return err
}

func (r *eventRepo) CreateReg(ctx context.Context, reg domain.EventRegistration) (domain.EventRegistration, error) {
	reg.CreatedAt = time.Now()
	// 事务：报名 + 参与计数自增（reg_count 与 registrations 行数保持一致）。
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.EventRegistration{}, fmt.Errorf("begin reg tx: %w", err)
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx,
		`INSERT INTO event_registrations (id,event_id,user_id,name,phone,org,status,created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		reg.ID, reg.EventID, reg.UserID, reg.Name, reg.Phone, reg.Org, reg.Status, reg.CreatedAt)
	if err != nil {
		// 唯一索引 uniq_event_regs_user_event（迁移 000077）兜底并发重复报名。
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.EventRegistration{}, fmt.Errorf("已报名过该活动，请勿重复报名")
		}
		return domain.EventRegistration{}, fmt.Errorf("insert event reg: %w", err)
	}
	if _, err := tx.Exec(ctx, `UPDATE association_events SET reg_count = reg_count + 1 WHERE id=$1`, reg.EventID); err != nil {
		return domain.EventRegistration{}, fmt.Errorf("bump event reg_count: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.EventRegistration{}, fmt.Errorf("commit reg tx: %w", err)
	}
	return reg, nil
}
func (r *eventRepo) ListRegs(ctx context.Context, eventID string) ([]domain.EventRegistration, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id,event_id,user_id,name,phone,org,status,created_at FROM event_registrations WHERE event_id=$1 ORDER BY created_at DESC`, eventID)
	if err != nil {
		return nil, fmt.Errorf("list event regs: %w", err)
	}
	defer rows.Close()
	var out []domain.EventRegistration
	for rows.Next() {
		var reg domain.EventRegistration
		if err := rows.Scan(&reg.ID, &reg.EventID, &reg.UserID, &reg.Name, &reg.Phone, &reg.Org, &reg.Status, &reg.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan event reg: %w", err)
		}
		out = append(out, reg)
	}
	return out, rows.Err()
}

// ---- Emergency ----

type emergRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewEmergencyRepository() repository.EmergencyRepository {
	return &emergRepo{pool: s.Pool()}
}

func (r *emergRepo) CreateResource(ctx context.Context, res domain.EmergencyResource) (domain.EmergencyResource, error) {
	res.CreatedAt = time.Now()
	res.UpdatedAt = res.CreatedAt
	_, err := r.pool.Exec(ctx,
		`INSERT INTO emergency_resources (id,owner_id,name,res_type,specs,quantity,location,contact_info,status,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		res.ID, res.OwnerID, res.Name, res.ResType, res.Specs, res.Quantity, res.Location, res.ContactInfo, res.Status, res.CreatedAt, res.UpdatedAt)
	return res, err
}
func (r *emergRepo) FindResourceByID(ctx context.Context, id string) (domain.EmergencyResource, error) {
	var res domain.EmergencyResource
	err := r.pool.QueryRow(ctx,
		`SELECT id,owner_id,name,res_type,specs,quantity,location,contact_info,status,created_at,updated_at FROM emergency_resources WHERE id=$1`, id).
		Scan(&res.ID, &res.OwnerID, &res.Name, &res.ResType, &res.Specs, &res.Quantity, &res.Location, &res.ContactInfo, &res.Status, &res.CreatedAt, &res.UpdatedAt)
	return res, err
}
func (r *emergRepo) ListResources(ctx context.Context, resType, q string, offset, limit int) ([]domain.EmergencyResource, int, error) {
	where := ""
	args := []any{}
	if resType != "" {
		where = `WHERE res_type=$1`
		args = append(args, resType)
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
		where += fmt.Sprintf(`(name ILIKE $%d ESCAPE '\' OR specs ILIKE $%d ESCAPE '\' OR location ILIKE $%d ESCAPE '\' OR contact_info ILIKE $%d ESCAPE '\')`,
			len(args), len(args), len(args), len(args))
	}
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM emergency_resources `+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count emergency resources: %w", err)
	}
	rows, err := r.pool.Query(ctx,
		fmt.Sprintf(`SELECT id,owner_id,name,res_type,specs,quantity,location,contact_info,status,created_at,updated_at FROM emergency_resources %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, where, len(args)+1, len(args)+2), append(args, limit, offset)...)
	if err != nil {
		return nil, 0, fmt.Errorf("list emergency resources: %w", err)
	}
	defer rows.Close()
	var out []domain.EmergencyResource
	for rows.Next() {
		var res domain.EmergencyResource
		if err := rows.Scan(&res.ID, &res.OwnerID, &res.Name, &res.ResType, &res.Specs, &res.Quantity, &res.Location, &res.ContactInfo, &res.Status, &res.CreatedAt, &res.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan emergency resource: %w", err)
		}
		out = append(out, res)
	}
	return out, total, rows.Err()
}
func (r *emergRepo) UpdateResource(ctx context.Context, res domain.EmergencyResource) (domain.EmergencyResource, error) {
	res.UpdatedAt = time.Now()
	_, err := r.pool.Exec(ctx,
		`UPDATE emergency_resources SET name=$1,res_type=$2,specs=$3,quantity=$4,location=$5,contact_info=$6,status=$7,updated_at=$8 WHERE id=$9`,
		res.Name, res.ResType, res.Specs, res.Quantity, res.Location, res.ContactInfo, res.Status, res.UpdatedAt, res.ID)
	return res, err
}

// nullableEndTime 把零值时间转为 NULL：进行中/待响应的调度没有结束时间，
// end_time 列可空，零值应存 NULL 而非 0001-01-01
func nullableEndTime(t time.Time) any {
	if t.IsZero() {
		return nil
	}
	return t
}

func (r *emergRepo) CreateDispatch(ctx context.Context, d domain.EmergencyDispatch) (domain.EmergencyDispatch, error) {
	d.CreatedAt = time.Now()
	_, err := r.pool.Exec(ctx,
		`INSERT INTO emergency_dispatches (id,resource_id,event_desc,location,start_time,end_time,commander,result,status,created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		d.ID, d.ResourceID, d.EventDesc, d.Location, d.StartTime, nullableEndTime(d.EndTime), d.Commander, d.Result, d.Status, d.CreatedAt)
	return d, err
}
func (r *emergRepo) ListDispatches(ctx context.Context, resourceID string, offset, limit int) ([]domain.EmergencyDispatch, int, error) {
	// resourceID 非空时按资源过滤（防注入：参数化）
	where := ""
	args := []any{limit, offset}
	if resourceID != "" {
		where = " WHERE d.resource_id = $3"
		args = append(args, resourceID)
	}
	var total int
	if resourceID != "" {
		// COUNT 独立查询：占位符从 $1 起（避免 $3 单独出现类型无法推断）
		if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM emergency_dispatches d WHERE d.resource_id = $1`, resourceID).Scan(&total); err != nil {
			return nil, 0, fmt.Errorf("count dispatches: %w", err)
		}
	} else {
		if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM emergency_dispatches d`).Scan(&total); err != nil {
			return nil, 0, fmt.Errorf("count dispatches: %w", err)
		}
	}
	// LEFT JOIN 资源表：内嵌 related 摘要（资源可能已删除 → 保留调度记录，related 为空）
	rows, err := r.pool.Query(ctx,
		`SELECT d.id,d.resource_id,d.event_desc,d.location,d.start_time,d.end_time,d.commander,d.result,d.status,d.created_at,
		        res.id,res.name,res.res_type,res.status
		 FROM emergency_dispatches d
		 LEFT JOIN emergency_resources res ON res.id = d.resource_id`+where+
			` ORDER BY d.created_at DESC LIMIT $1 OFFSET $2`, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list dispatches: %w", err)
	}
	defer rows.Close()
	var out []domain.EmergencyDispatch
	for rows.Next() {
		var d domain.EmergencyDispatch
		var resID, resName, resType, resStatus pgtype.Text
		var endTime pgtype.Timestamptz
		if err := rows.Scan(&d.ID, &d.ResourceID, &d.EventDesc, &d.Location, &d.StartTime, &endTime, &d.Commander, &d.Result, &d.Status, &d.CreatedAt,
			&resID, &resName, &resType, &resStatus); err != nil {
			return nil, 0, fmt.Errorf("scan dispatch: %w", err)
		}
		d.EndTime = endTime.Time
		if resID.Valid {
			d.Related = &domain.EmergencyResourceBrief{ID: resID.String, Name: resName.String, ResType: resType.String, Status: resStatus.String}
		}
		out = append(out, d)
	}
	return out, total, rows.Err()
}

func (r *emergRepo) DeleteResource(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM emergency_resources WHERE id=$1", id)
	return err
}

func (r *emergRepo) FindDispatchByID(ctx context.Context, id string) (domain.EmergencyDispatch, error) {
	var d domain.EmergencyDispatch
	var endTime pgtype.Timestamptz
	err := r.pool.QueryRow(ctx, "SELECT id,resource_id,event_desc,location,start_time,end_time,commander,result,status,created_at FROM emergency_dispatches WHERE id=$1", id).
		Scan(&d.ID, &d.ResourceID, &d.EventDesc, &d.Location, &d.StartTime, &endTime, &d.Commander, &d.Result, &d.Status, &d.CreatedAt)
	d.EndTime = endTime.Time
	return d, err
}

func (r *emergRepo) UpdateDispatch(ctx context.Context, d domain.EmergencyDispatch) (domain.EmergencyDispatch, error) {
	_, err := r.pool.Exec(ctx,
		"UPDATE emergency_dispatches SET resource_id=$1,event_desc=$2,location=$3,start_time=$4,end_time=$5,commander=$6,result=$7,status=$8 WHERE id=$9",
		d.ResourceID, d.EventDesc, d.Location, d.StartTime, nullableEndTime(d.EndTime), d.Commander, d.Result, d.Status, d.ID)
	return d, err
}

func (r *emergRepo) DeleteDispatch(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM emergency_dispatches WHERE id=$1", id)
	return err
}
