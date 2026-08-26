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

// ---- Achievement ----

type achieveRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewAchievementRepository() repository.AchievementRepository {
	return &achieveRepo{pool: s.Pool()}
}

func (r *achieveRepo) Create(ctx context.Context, a domain.Achievement) (domain.Achievement, error) {
	a.CreatedAt = time.Now()
	a.UpdatedAt = a.CreatedAt
	imgs, err := json.Marshal(a.Images)
	if err != nil {
		return domain.Achievement{}, fmt.Errorf("marshal achievement images: %w", err)
	}
	atts, err := json.Marshal(a.Attachments)
	if err != nil {
		return domain.Achievement{}, fmt.Errorf("marshal attachments: %w", err)
	}
	_, err = r.pool.Exec(ctx,
		`INSERT INTO achievements (id,owner_id,title,achieve_type,description,field,stage,images,attachments,contact_info,status,views,favs,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		a.ID, a.OwnerID, a.Title, a.AchieveType, a.Description, a.Field, a.Stage, imgs, atts, a.ContactInfo, a.Status, a.Views, a.Favs, a.CreatedAt, a.UpdatedAt)
	return a, err
}
func (r *achieveRepo) FindByID(ctx context.Context, id string) (domain.Achievement, error) {
	var a domain.Achievement
	var imgs, atts []byte
	err := r.pool.QueryRow(ctx,
		`SELECT id,owner_id,title,achieve_type,description,field,stage,images,attachments,contact_info,status,views,favs,created_at,updated_at FROM achievements WHERE id=$1`, id).
		Scan(&a.ID, &a.OwnerID, &a.Title, &a.AchieveType, &a.Description, &a.Field, &a.Stage, &imgs, &atts, &a.ContactInfo, &a.Status, &a.Views, &a.Favs, &a.CreatedAt, &a.UpdatedAt)
	json.Unmarshal(imgs, &a.Images)
	json.Unmarshal(atts, &a.Attachments)
	return a, err
}
func (r *achieveRepo) List(ctx context.Context, field string, offset, limit int) ([]domain.Achievement, int, error) {
	where := ""
	args := []any{}
	if field != "" {
		where = `WHERE field=$1`
		args = append(args, field)
	}
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM achievements `+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count achievements: %w", err)
	}
	q := fmt.Sprintf(`SELECT id,owner_id,title,achieve_type,description,field,stage,images,attachments,contact_info,status,views,favs,created_at,updated_at FROM achievements %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, where, len(args)+1, len(args)+2)
	rows, err := r.pool.Query(ctx, q, append(args, limit, offset)...)
	if err != nil {
		return nil, 0, fmt.Errorf("list achievements: %w", err)
	}
	defer rows.Close()
	var out []domain.Achievement
	for rows.Next() {
		var a domain.Achievement
		var imgs, atts []byte
		if err := rows.Scan(&a.ID, &a.OwnerID, &a.Title, &a.AchieveType, &a.Description, &a.Field, &a.Stage, &imgs, &atts, &a.ContactInfo, &a.Status, &a.Views, &a.Favs, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan achievement: %w", err)
		}
		json.Unmarshal(imgs, &a.Images)
		json.Unmarshal(atts, &a.Attachments)
		out = append(out, a)
	}
	return out, total, rows.Err()
}
func (r *achieveRepo) Update(ctx context.Context, a domain.Achievement) (domain.Achievement, error) {
	a.UpdatedAt = time.Now()
	imgs, err := json.Marshal(a.Images)
	if err != nil {
		return domain.Achievement{}, fmt.Errorf("marshal achievement images: %w", err)
	}
	atts, err := json.Marshal(a.Attachments)
	if err != nil {
		return domain.Achievement{}, fmt.Errorf("marshal achievement attachments: %w", err)
	}
	_, err = r.pool.Exec(ctx,
		`UPDATE achievements SET title=$1,achieve_type=$2,description=$3,field=$4,stage=$5,images=$6,attachments=$7,contact_info=$8,status=$9,updated_at=$10 WHERE id=$11`,
		a.Title, a.AchieveType, a.Description, a.Field, a.Stage, imgs, atts, a.ContactInfo, a.Status, a.UpdatedAt, a.ID)
	return a, err
}
func (r *achieveRepo) AdjustStats(ctx context.Context, id string, viewsDelta, favsDelta int) error {
	if viewsDelta == 0 && favsDelta == 0 {
		return nil
	}
	tag, err := r.pool.Exec(ctx,
		`UPDATE achievements SET views=GREATEST(views+$2,0), favs=GREATEST(favs+$3,0), updated_at=now() WHERE id=$1`,
		id, viewsDelta, favsDelta)
	if err != nil {
		return fmt.Errorf("adjust achievement stats %s: %w", id, err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("achievement %s not found", id)
	}
	return nil
}
func (r *achieveRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM achievements WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("delete achievement %s: %w", id, err)
	}
	return nil
}

// ---- RDChallenge ----

type rdChallengeRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewRDChallengeRepository() repository.RDChallengeRepository {
	return &rdChallengeRepo{pool: s.Pool()}
}

func (r *rdChallengeRepo) Create(ctx context.Context, c domain.RDChallenge) (domain.RDChallenge, error) {
	c.CreatedAt = time.Now()
	c.UpdatedAt = c.CreatedAt
	_, err := r.pool.Exec(ctx,
		`INSERT INTO rd_challenges (id,poster_id,title,field,description,requirements,budget_fen,deadline,status,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		c.ID, c.PosterID, c.Title, c.Field, c.Description, c.Requirements, c.BudgetFen, c.Deadline, c.Status, c.CreatedAt, c.UpdatedAt)
	return c, err
}
func (r *rdChallengeRepo) FindByID(ctx context.Context, id string) (domain.RDChallenge, error) {
	var c domain.RDChallenge
	err := r.pool.QueryRow(ctx,
		`SELECT id,poster_id,title,field,description,requirements,budget_fen,COALESCE(deadline,'1970-01-01 00:00:00+00'::timestamptz),status,created_at,updated_at FROM rd_challenges WHERE id=$1`, id).
		Scan(&c.ID, &c.PosterID, &c.Title, &c.Field, &c.Description, &c.Requirements, &c.BudgetFen, &c.Deadline, &c.Status, &c.CreatedAt, &c.UpdatedAt)
	return c, err
}
func (r *rdChallengeRepo) List(ctx context.Context, field string, offset, limit int) ([]domain.RDChallenge, int, error) {
	where := ""
	args := []any{}
	if field != "" {
		where = `WHERE field=$1`
		args = append(args, field)
	}
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM rd_challenges `+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count challenges: %w", err)
	}
	q := fmt.Sprintf(`SELECT id,poster_id,title,field,description,requirements,budget_fen,COALESCE(deadline,'1970-01-01 00:00:00+00'::timestamptz),status,created_at,updated_at FROM rd_challenges %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, where, len(args)+1, len(args)+2)
	rows, err := r.pool.Query(ctx, q, append(args, limit, offset)...)
	if err != nil {
		return nil, 0, fmt.Errorf("list challenges: %w", err)
	}
	defer rows.Close()
	var out []domain.RDChallenge
	for rows.Next() {
		var c domain.RDChallenge
		if err := rows.Scan(&c.ID, &c.PosterID, &c.Title, &c.Field, &c.Description, &c.Requirements, &c.BudgetFen, &c.Deadline, &c.Status, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan challenge: %w", err)
		}
		out = append(out, c)
	}
	return out, total, rows.Err()
}
func (r *rdChallengeRepo) Update(ctx context.Context, c domain.RDChallenge) (domain.RDChallenge, error) {
	c.UpdatedAt = time.Now()
	_, err := r.pool.Exec(ctx,
		`UPDATE rd_challenges SET title=$1,field=$2,description=$3,requirements=$4,budget_fen=$5,deadline=$6,status=$7,updated_at=$8 WHERE id=$9`,
		c.Title, c.Field, c.Description, c.Requirements, c.BudgetFen, c.Deadline, c.Status, c.UpdatedAt, c.ID)
	return c, err
}

func (r *rdChallengeRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM rd_challenges WHERE id=$1", id)
	return err
}

// ---- ResearchProject ----

type researchProjRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewResearchProjectRepository() repository.ResearchProjectRepository {
	return &researchProjRepo{pool: s.Pool()}
}

func (r *researchProjRepo) Create(ctx context.Context, p domain.ResearchProject) (domain.ResearchProject, error) {
	p.CreatedAt = time.Now()
	p.UpdatedAt = p.CreatedAt
	members, err := json.Marshal(p.Members)
	if err != nil {
		return domain.ResearchProject{}, fmt.Errorf("marshal members: %w", err)
	}
	_, err = r.pool.Exec(ctx,
		`INSERT INTO research_projects (id,title,field,description,lead_org,members,budget_fen,start_date,end_date,milestones,status,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		p.ID, p.Title, p.Field, p.Description, p.LeadOrg, members, p.BudgetFen, p.StartDate, p.EndDate, p.Milestones, p.Status, p.CreatedAt, p.UpdatedAt)
	return p, err
}
func (r *researchProjRepo) FindByID(ctx context.Context, id string) (domain.ResearchProject, error) {
	var p domain.ResearchProject
	var members []byte
	err := r.pool.QueryRow(ctx,
		`SELECT id,title,field,description,lead_org,members,budget_fen,COALESCE(start_date,'1970-01-01 00:00:00+00'::timestamptz),COALESCE(end_date,'1970-01-01 00:00:00+00'::timestamptz),milestones,status,created_at,updated_at FROM research_projects WHERE id=$1`, id).
		Scan(&p.ID, &p.Title, &p.Field, &p.Description, &p.LeadOrg, &members, &p.BudgetFen, &p.StartDate, &p.EndDate, &p.Milestones, &p.Status, &p.CreatedAt, &p.UpdatedAt)
	json.Unmarshal(members, &p.Members)
	return p, err
}
func (r *researchProjRepo) List(ctx context.Context, offset, limit int) ([]domain.ResearchProject, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM research_projects`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count projects: %w", err)
	}
	rows, err := r.pool.Query(ctx,
		`SELECT id,title,field,description,lead_org,members,budget_fen,COALESCE(start_date,'1970-01-01 00:00:00+00'::timestamptz),COALESCE(end_date,'1970-01-01 00:00:00+00'::timestamptz),milestones,status,created_at,updated_at FROM research_projects ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list projects: %w", err)
	}
	defer rows.Close()
	var out []domain.ResearchProject
	for rows.Next() {
		var p domain.ResearchProject
		var members []byte
		if err := rows.Scan(&p.ID, &p.Title, &p.Field, &p.Description, &p.LeadOrg, &members, &p.BudgetFen, &p.StartDate, &p.EndDate, &p.Milestones, &p.Status, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan project: %w", err)
		}
		json.Unmarshal(members, &p.Members)
		out = append(out, p)
	}
	return out, total, rows.Err()
}
func (r *researchProjRepo) Update(ctx context.Context, p domain.ResearchProject) (domain.ResearchProject, error) {
	p.UpdatedAt = time.Now()
	members, err := json.Marshal(p.Members)
	if err != nil {
		return domain.ResearchProject{}, fmt.Errorf("marshal members: %w", err)
	}
	_, err = r.pool.Exec(ctx,
		`UPDATE research_projects SET title=$1,field=$2,description=$3,lead_org=$4,members=$5,budget_fen=$6,start_date=$7,end_date=$8,milestones=$9,status=$10,updated_at=$11 WHERE id=$12`,
		p.Title, p.Field, p.Description, p.LeadOrg, members, p.BudgetFen, p.StartDate, p.EndDate, p.Milestones, p.Status, p.UpdatedAt, p.ID)
	return p, err
}

func (r *researchProjRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM research_projects WHERE id=$1", id)
	return err
}

// ---- ProjectApp ----

type projAppRepo struct{ pool *pgxpool.Pool }

func (s *Store) NewProjectAppRepository() repository.ProjectAppRepository {
	return &projAppRepo{pool: s.Pool()}
}

func (r *projAppRepo) Create(ctx context.Context, a domain.ProjectApplication) (domain.ProjectApplication, error) {
	a.CreatedAt = time.Now()
	a.UpdatedAt = a.CreatedAt
	att, err := json.Marshal(a.Attachments)
	if err != nil {
		return domain.ProjectApplication{}, fmt.Errorf("marshal attachments: %w", err)
	}
	_, err = r.pool.Exec(ctx,
		`INSERT INTO project_applications (id,applicant_id,project_name,category,budget_fen,description,attachments,status,review_note,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		a.ID, a.ApplicantID, a.ProjectName, a.Category, a.BudgetFen, a.Description, att, a.Status, a.ReviewNote, a.CreatedAt, a.UpdatedAt)
	return a, err
}
func (r *projAppRepo) FindByID(ctx context.Context, id string) (domain.ProjectApplication, error) {
	var a domain.ProjectApplication
	var att []byte
	err := r.pool.QueryRow(ctx,
		`SELECT id,applicant_id,project_name,category,budget_fen,description,attachments,status,review_note,created_at,updated_at FROM project_applications WHERE id=$1`, id).
		Scan(&a.ID, &a.ApplicantID, &a.ProjectName, &a.Category, &a.BudgetFen, &a.Description, &att, &a.Status, &a.ReviewNote, &a.CreatedAt, &a.UpdatedAt)
	json.Unmarshal(att, &a.Attachments)
	return a, err
}
func (r *projAppRepo) ListByUser(ctx context.Context, userID string) ([]domain.ProjectApplication, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id,applicant_id,project_name,category,budget_fen,description,attachments,status,review_note,created_at,updated_at FROM project_applications WHERE applicant_id=$1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("list apps by user: %w", err)
	}
	defer rows.Close()
	var out []domain.ProjectApplication
	for rows.Next() {
		var a domain.ProjectApplication
		var att []byte
		if err := rows.Scan(&a.ID, &a.ApplicantID, &a.ProjectName, &a.Category, &a.BudgetFen, &a.Description, &att, &a.Status, &a.ReviewNote, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan app: %w", err)
		}
		json.Unmarshal(att, &a.Attachments)
		out = append(out, a)
	}
	return out, rows.Err()
}
func (r *projAppRepo) ListAll(ctx context.Context, status string, offset, limit int) ([]domain.ProjectApplication, int, error) {
	where := ""
	args := []any{}
	if status != "" {
		where = `WHERE status=$1`
		args = append(args, status)
	}
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM project_applications `+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count apps: %w", err)
	}
	q := fmt.Sprintf(`SELECT id,applicant_id,project_name,category,budget_fen,description,attachments,status,review_note,created_at,updated_at FROM project_applications %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, where, len(args)+1, len(args)+2)
	rows, err := r.pool.Query(ctx, q, append(args, limit, offset)...)
	if err != nil {
		return nil, 0, fmt.Errorf("list all apps: %w", err)
	}
	defer rows.Close()
	var out []domain.ProjectApplication
	for rows.Next() {
		var a domain.ProjectApplication
		var att []byte
		if err := rows.Scan(&a.ID, &a.ApplicantID, &a.ProjectName, &a.Category, &a.BudgetFen, &a.Description, &att, &a.Status, &a.ReviewNote, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan app: %w", err)
		}
		json.Unmarshal(att, &a.Attachments)
		out = append(out, a)
	}
	return out, total, rows.Err()
}
func (r *projAppRepo) Update(ctx context.Context, a domain.ProjectApplication) (domain.ProjectApplication, error) {
	a.UpdatedAt = time.Now()
	att, err := json.Marshal(a.Attachments)
	if err != nil {
		return domain.ProjectApplication{}, fmt.Errorf("marshal attachments: %w", err)
	}
	_, err = r.pool.Exec(ctx,
		`UPDATE project_applications SET project_name=$1,category=$2,budget_fen=$3,description=$4,attachments=$5,status=$6,review_note=$7,updated_at=$8 WHERE id=$9`,
		a.ProjectName, a.Category, a.BudgetFen, a.Description, att, a.Status, a.ReviewNote, a.UpdatedAt, a.ID)
	return a, err
}
