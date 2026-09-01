package memory

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)

// ---- Expert ----

type expertRepo struct {
	mu    sync.RWMutex
	items []domain.Expert
}

func NewExpertRepository() repository.ExpertRepository { return &expertRepo{} }

func (r *expertRepo) Create(ctx context.Context, e domain.Expert) (domain.Expert, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	e.CreatedAt = time.Now()
	e.UpdatedAt = e.CreatedAt
	r.items = append(r.items, e)
	return e, nil
}
func (r *expertRepo) FindByID(ctx context.Context, id string) (domain.Expert, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, e := range r.items {
		if e.ID == id {
			return e, nil
		}
	}
	return domain.Expert{}, fmt.Errorf("expert %s not found", id)
}
func (r *expertRepo) List(ctx context.Context, field string) ([]domain.Expert, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.Expert, 0)
	for _, e := range r.items {
		if field == "" || e.Field == field {
			out = append(out, e)
		}
	}
	// 与 PG 对齐：ORDER BY created_at DESC。
	sort.SliceStable(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}
func (r *expertRepo) Update(ctx context.Context, e domain.Expert) (domain.Expert, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, v := range r.items {
		if v.ID == e.ID {
			e.UpdatedAt = time.Now()
			r.items[i] = e
			return e, nil
		}
	}
	return domain.Expert{}, fmt.Errorf("expert %s not found", e.ID)
}
func (r *expertRepo) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, v := range r.items {
		if v.ID == id {
			r.items = append(r.items[:i], r.items[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("expert %s not found", id)
}

// ---- Case ----

type caseRepo struct {
	mu    sync.RWMutex
	items []domain.CaseEntry
}

func NewCaseRepository() repository.CaseRepository { return &caseRepo{} }

func (r *caseRepo) Create(ctx context.Context, c domain.CaseEntry) (domain.CaseEntry, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	c.CreatedAt = time.Now()
	c.UpdatedAt = c.CreatedAt
	r.items = append(r.items, c)
	return c, nil
}
func (r *caseRepo) FindByID(ctx context.Context, id string) (domain.CaseEntry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, c := range r.items {
		if c.ID == id {
			return c, nil
		}
	}
	return domain.CaseEntry{}, fmt.Errorf("case %s not found", id)
}
func (r *caseRepo) List(ctx context.Context, category string, offset, limit int) ([]domain.CaseEntry, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	filtered := make([]domain.CaseEntry, 0)
	for _, c := range r.items {
		if category == "" || c.Category == category {
			filtered = append(filtered, c)
		}
	}
	// 与 PG 对齐：ORDER BY created_at DESC。
	sort.SliceStable(filtered, func(i, j int) bool { return filtered[i].CreatedAt.After(filtered[j].CreatedAt) })
	page, total, _ := paginateSlice(filtered, offset, limit)
	return page, total, nil
}
func (r *caseRepo) Update(ctx context.Context, c domain.CaseEntry) (domain.CaseEntry, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, v := range r.items {
		if v.ID == c.ID {
			c.UpdatedAt = time.Now()
			r.items[i] = c
			return c, nil
		}
	}
	return domain.CaseEntry{}, fmt.Errorf("case %s not found", c.ID)
}
func (r *caseRepo) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, v := range r.items {
		if v.ID == id {
			r.items = append(r.items[:i], r.items[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("case %s not found", id)
}

// ---- Compliance ----

type complianceRepo struct {
	mu   sync.RWMutex
	docs []domain.ComplianceDoc
	stds []domain.StandardDoc
}

func NewComplianceRepository() repository.ComplianceRepository { return &complianceRepo{} }

func (r *complianceRepo) CreateDoc(ctx context.Context, d domain.ComplianceDoc) (domain.ComplianceDoc, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	d.CreatedAt = time.Now()
	d.UpdatedAt = d.CreatedAt
	r.docs = append(r.docs, d)
	return d, nil
}
func (r *complianceRepo) FindDocByID(ctx context.Context, id string) (domain.ComplianceDoc, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, d := range r.docs {
		if d.ID == id {
			return d, nil
		}
	}
	return domain.ComplianceDoc{}, fmt.Errorf("compliance doc %s not found", id)
}
func (r *complianceRepo) ListDocs(ctx context.Context, category string, offset, limit int) ([]domain.ComplianceDoc, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	filtered := make([]domain.ComplianceDoc, 0)
	for _, d := range r.docs {
		if category == "" || d.Category == category {
			filtered = append(filtered, d)
		}
	}
	// 与 PG 对齐：ORDER BY created_at DESC。
	sort.SliceStable(filtered, func(i, j int) bool { return filtered[i].CreatedAt.After(filtered[j].CreatedAt) })
	page, total, _ := paginateSlice(filtered, offset, limit)
	return page, total, nil
}
func (r *complianceRepo) UpdateDoc(ctx context.Context, d domain.ComplianceDoc) (domain.ComplianceDoc, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, v := range r.docs {
		if v.ID == d.ID {
			d.UpdatedAt = time.Now()
			r.docs[i] = d
			return d, nil
		}
	}
	return domain.ComplianceDoc{}, fmt.Errorf("compliance doc %s not found", d.ID)
}
func (r *complianceRepo) DeleteDoc(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, v := range r.docs {
		if v.ID == id {
			r.docs = append(r.docs[:i], r.docs[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("compliance doc %s not found", id)
}

func (r *complianceRepo) DeleteStandard(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, v := range r.stds {
		if v.ID == id {
			r.stds = append(r.stds[:i], r.stds[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("standard %s not found", id)
}

func (r *complianceRepo) FindStandardByID(ctx context.Context, id string) (domain.StandardDoc, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, s := range r.stds {
		if s.ID == id {
			return s, nil
		}
	}
	return domain.StandardDoc{}, fmt.Errorf("standard %s not found", id)
}

func (r *complianceRepo) CreateStandard(ctx context.Context, s domain.StandardDoc) (domain.StandardDoc, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	s.CreatedAt = time.Now()
	s.UpdatedAt = s.CreatedAt
	r.stds = append(r.stds, s)
	return s, nil
}
func (r *complianceRepo) ListStandards(ctx context.Context, category string, offset, limit int) ([]domain.StandardDoc, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	filtered := make([]domain.StandardDoc, 0)
	for _, s := range r.stds {
		if category == "" || s.Category == category {
			filtered = append(filtered, s)
		}
	}
	// 与 PG 对齐：ORDER BY created_at DESC。
	sort.SliceStable(filtered, func(i, j int) bool { return filtered[i].CreatedAt.After(filtered[j].CreatedAt) })
	page, total, _ := paginateSlice(filtered, offset, limit)
	return page, total, nil
}

func (r *complianceRepo) UpdateStandard(ctx context.Context, s domain.StandardDoc) (domain.StandardDoc, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, v := range r.stds {
		if v.ID == s.ID {
			r.stds[i] = s
			return s, nil
		}
	}
	return domain.StandardDoc{}, fmt.Errorf("standard %s not found", s.ID)
}

// ---- Achievement ----

type achieveRepo struct {
	mu    sync.RWMutex
	items []domain.Achievement
}

func NewAchievementRepository() repository.AchievementRepository { return &achieveRepo{} }

func (r *achieveRepo) Create(ctx context.Context, a domain.Achievement) (domain.Achievement, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	a.CreatedAt = time.Now()
	a.UpdatedAt = a.CreatedAt
	r.items = append(r.items, a)
	return a, nil
}
func (r *achieveRepo) FindByID(ctx context.Context, id string) (domain.Achievement, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, a := range r.items {
		if a.ID == id {
			return a, nil
		}
	}
	return domain.Achievement{}, fmt.Errorf("achievement %s not found", id)
}
func (r *achieveRepo) List(ctx context.Context, field string, offset, limit int) ([]domain.Achievement, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	filtered := make([]domain.Achievement, 0)
	for _, a := range r.items {
		if field == "" || a.Field == field {
			filtered = append(filtered, a)
		}
	}
	// 与 PG 对齐：ORDER BY created_at DESC。
	sort.SliceStable(filtered, func(i, j int) bool { return filtered[i].CreatedAt.After(filtered[j].CreatedAt) })
	page, total, _ := paginateSlice(filtered, offset, limit)
	return page, total, nil
}
func (r *achieveRepo) Update(ctx context.Context, a domain.Achievement) (domain.Achievement, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, v := range r.items {
		if v.ID == a.ID {
			a.UpdatedAt = time.Now()
			r.items[i] = a
			return a, nil
		}
	}
	return domain.Achievement{}, fmt.Errorf("achievement %s not found", a.ID)
}
func (r *achieveRepo) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, v := range r.items {
		if v.ID == id {
			r.items = append(r.items[:i], r.items[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("achievement %s not found", id)
}
func (r *achieveRepo) AdjustStats(ctx context.Context, id string, viewsDelta, favsDelta int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, v := range r.items {
		if v.ID == id {
			nv := v.Views + viewsDelta
			if nv < 0 {
				nv = 0
			}
			nf := v.Favs + favsDelta
			if nf < 0 {
				nf = 0
			}
			r.items[i].Views = nv
			r.items[i].Favs = nf
			r.items[i].UpdatedAt = time.Now()
			return nil
		}
	}
	return fmt.Errorf("achievement %s not found", id)
}

// ---- ChallengeClaim (揭榜意向) ----

type claimRepo struct {
	mu    sync.RWMutex
	items []domain.ChallengeClaim
}

func NewChallengeClaimRepository() repository.ChallengeClaimRepository { return &claimRepo{} }

func (r *claimRepo) Create(ctx context.Context, c domain.ChallengeClaim) (domain.ChallengeClaim, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, v := range r.items {
		if v.ChallengeID == c.ChallengeID && v.UserID == c.UserID {
			return v, nil // 幂等：同人同难题已有记录，原样返回
		}
	}
	c.CreatedAt = time.Now()
	r.items = append(r.items, c)
	return c, nil
}

func (r *claimRepo) FindByChallengeAndUser(ctx context.Context, challengeID, userID string) (domain.ChallengeClaim, bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, v := range r.items {
		if v.ChallengeID == challengeID && v.UserID == userID {
			return v, true, nil
		}
	}
	return domain.ChallengeClaim{}, false, nil
}

func (r *claimRepo) ListByChallenge(ctx context.Context, challengeID string) ([]domain.ChallengeClaim, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.ChallengeClaim, 0)
	for _, v := range r.items {
		if v.ChallengeID == challengeID {
			out = append(out, v)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}

// ---- RDChallenge ----

type rdRepo struct {
	mu    sync.RWMutex
	items []domain.RDChallenge
}

func NewRDChallengeRepository() repository.RDChallengeRepository { return &rdRepo{} }

func (r *rdRepo) Create(ctx context.Context, c domain.RDChallenge) (domain.RDChallenge, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	c.CreatedAt = time.Now()
	c.UpdatedAt = c.CreatedAt
	r.items = append(r.items, c)
	return c, nil
}
func (r *rdRepo) FindByID(ctx context.Context, id string) (domain.RDChallenge, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, c := range r.items {
		if c.ID == id {
			return c, nil
		}
	}
	return domain.RDChallenge{}, fmt.Errorf("challenge %s not found", id)
}
func (r *rdRepo) List(ctx context.Context, field string, offset, limit int) ([]domain.RDChallenge, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	filtered := make([]domain.RDChallenge, 0)
	for _, c := range r.items {
		if field == "" || c.Field == field {
			filtered = append(filtered, c)
		}
	}
	// 与 PG 对齐：ORDER BY created_at DESC。
	sort.SliceStable(filtered, func(i, j int) bool { return filtered[i].CreatedAt.After(filtered[j].CreatedAt) })
	page, total, _ := paginateSlice(filtered, offset, limit)
	return page, total, nil
}
func (r *rdRepo) Update(ctx context.Context, c domain.RDChallenge) (domain.RDChallenge, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, v := range r.items {
		if v.ID == c.ID {
			c.UpdatedAt = time.Now()
			r.items[i] = c
			return c, nil
		}
	}
	return domain.RDChallenge{}, fmt.Errorf("challenge %s not found", c.ID)
}

func (r *rdRepo) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.items {
		if r.items[i].ID == id {
			r.items = append(r.items[:i], r.items[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("challenge %s not found", id)
}

// ---- ResearchProject ----

type projRepo struct {
	mu    sync.RWMutex
	items []domain.ResearchProject
	joins []domain.ProjectJoinRequest
}

func NewResearchProjectRepository() repository.ResearchProjectRepository { return &projRepo{} }

func (r *projRepo) Create(ctx context.Context, p domain.ResearchProject) (domain.ResearchProject, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	p.CreatedAt = time.Now()
	p.UpdatedAt = p.CreatedAt
	r.items = append(r.items, p)
	return p, nil
}
func (r *projRepo) FindByID(ctx context.Context, id string) (domain.ResearchProject, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, p := range r.items {
		if p.ID == id {
			return p, nil
		}
	}
	return domain.ResearchProject{}, fmt.Errorf("project %s not found", id)
}
func (r *projRepo) List(ctx context.Context, offset, limit int) ([]domain.ResearchProject, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := append([]domain.ResearchProject(nil), r.items...)
	sort.SliceStable(items, func(i, j int) bool { return items[i].CreatedAt.After(items[j].CreatedAt) })
	return paginateSlice(items, offset, limit)
}
func (r *projRepo) Update(ctx context.Context, p domain.ResearchProject) (domain.ResearchProject, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, v := range r.items {
		if v.ID == p.ID {
			p.UpdatedAt = time.Now()
			r.items[i] = p
			return p, nil
		}
	}
	return domain.ResearchProject{}, fmt.Errorf("project %s not found", p.ID)
}

func (r *projRepo) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.items {
		if r.items[i].ID == id {
			r.items = append(r.items[:i], r.items[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("project %s not found", id)
}

// ---- 参与申请（课题攻关） ----

func (r *projRepo) CreateJoinRequest(ctx context.Context, v domain.ProjectJoinRequest) (domain.ProjectJoinRequest, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	v.CreatedAt = time.Now()
	v.UpdatedAt = v.CreatedAt
	for i := range r.joins {
		if r.joins[i].ProjectID == v.ProjectID && r.joins[i].UserID == v.UserID {
			return domain.ProjectJoinRequest{}, fmt.Errorf("join request already exists for project %s user %s", v.ProjectID, v.UserID)
		}
	}
	r.joins = append(r.joins, v)
	return v, nil
}

func (r *projRepo) FindJoinByID(ctx context.Context, id string) (domain.ProjectJoinRequest, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, v := range r.joins {
		if v.ID == id {
			return v, nil
		}
	}
	return domain.ProjectJoinRequest{}, fmt.Errorf("project join request %s not found", id)
}

func (r *projRepo) FindJoinByProjectUser(ctx context.Context, projectID, userID string) (domain.ProjectJoinRequest, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, v := range r.joins {
		if v.ProjectID == projectID && v.UserID == userID {
			return v, nil
		}
	}
	return domain.ProjectJoinRequest{}, fmt.Errorf("project join request for project %s user %s not found", projectID, userID)
}

func (r *projRepo) ListJoinRequests(ctx context.Context, projectID string) ([]domain.ProjectJoinRequest, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []domain.ProjectJoinRequest
	for _, v := range r.joins {
		if v.ProjectID == projectID {
			out = append(out, v)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}

func (r *projRepo) UpdateJoinRequest(ctx context.Context, v domain.ProjectJoinRequest) (domain.ProjectJoinRequest, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.joins {
		if r.joins[i].ID == v.ID {
			v.UpdatedAt = time.Now()
			r.joins[i] = v
			return v, nil
		}
	}
	return domain.ProjectJoinRequest{}, fmt.Errorf("project join request %s not found", v.ID)
}

// ---- ProjectApp ----

type projAppRepo struct {
	mu    sync.RWMutex
	items []domain.ProjectApplication
}

func NewProjectAppRepository() repository.ProjectAppRepository { return &projAppRepo{} }

func (r *projAppRepo) Create(ctx context.Context, a domain.ProjectApplication) (domain.ProjectApplication, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	a.CreatedAt = time.Now()
	a.UpdatedAt = a.CreatedAt
	r.items = append(r.items, a)
	return a, nil
}
func (r *projAppRepo) FindByID(ctx context.Context, id string) (domain.ProjectApplication, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, a := range r.items {
		if a.ID == id {
			return a, nil
		}
	}
	return domain.ProjectApplication{}, fmt.Errorf("application %s not found", id)
}
func (r *projAppRepo) ListByUser(ctx context.Context, userID string) ([]domain.ProjectApplication, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.ProjectApplication, 0)
	for _, a := range r.items {
		if a.ApplicantID == userID {
			out = append(out, a)
		}
	}
	// 与 PG 对齐：ORDER BY created_at DESC。
	sort.SliceStable(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}
func (r *projAppRepo) ListAll(ctx context.Context, status string, offset, limit int) ([]domain.ProjectApplication, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	filtered := make([]domain.ProjectApplication, 0)
	for _, a := range r.items {
		if status == "" || a.Status == status {
			filtered = append(filtered, a)
		}
	}
	sort.SliceStable(filtered, func(i, j int) bool { return filtered[i].CreatedAt.After(filtered[j].CreatedAt) })
	return paginateSlice(filtered, offset, limit)
}
func (r *projAppRepo) Update(ctx context.Context, a domain.ProjectApplication) (domain.ProjectApplication, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, v := range r.items {
		if v.ID == a.ID {
			a.UpdatedAt = time.Now()
			r.items[i] = a
			return a, nil
		}
	}
	return domain.ProjectApplication{}, fmt.Errorf("application %s not found", a.ID)
}

// ---- Application (service_applications) ----

type appRepo struct {
	mu    sync.RWMutex
	items []domain.Application
}

func NewApplicationRepository() repository.ApplicationRepository { return &appRepo{} }

func (r *appRepo) Create(ctx context.Context, a domain.Application) (domain.Application, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	a.CreatedAt = time.Now()
	r.items = append(r.items, a)
	return a, nil
}
func (r *appRepo) FindByID(ctx context.Context, id string) (domain.Application, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, a := range r.items {
		if a.ID == id {
			return a, nil
		}
	}
	return domain.Application{}, fmt.Errorf("application %s not found", id)
}
func (r *appRepo) ListByUser(ctx context.Context, userID string, offset, limit int) ([]domain.Application, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	all := make([]domain.Application, 0)
	for _, a := range r.items {
		if a.UserID == userID {
			all = append(all, a)
		}
	}
	sort.SliceStable(all, func(i, j int) bool { return all[i].CreatedAt.After(all[j].CreatedAt) })
	return slicePage(all, offset, limit)
}
func (r *appRepo) ListAll(ctx context.Context, offset, limit int) ([]domain.Application, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := append([]domain.Application(nil), r.items...)
	sort.SliceStable(items, func(i, j int) bool { return items[i].CreatedAt.After(items[j].CreatedAt) })
	return slicePage(items, offset, limit)
}

// slicePage returns the offset/limit page of a sorted slice plus its total count.
func slicePage[T any](items []T, offset, limit int) ([]T, int, error) {
	total := len(items)
	if offset >= total {
		return []T{}, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return items[offset:end], total, nil
}
