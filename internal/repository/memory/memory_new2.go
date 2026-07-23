package memory

import (
	"fmt"
	"sync"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)

// ---- Competition ----

type compRepo struct {
	mu    sync.RWMutex
	items []domain.Competition
	regs  []domain.CompetitionReg
}

func NewCompetitionRepository() repository.CompetitionRepository { return &compRepo{} }

func (r *compRepo) Create(c domain.Competition) (domain.Competition, error) {
	r.mu.Lock(); defer r.mu.Unlock()
	c.CreatedAt = time.Now(); c.UpdatedAt = c.CreatedAt
	r.items = append(r.items, c); return c, nil
}
func (r *compRepo) FindByID(id string) (domain.Competition, error) {
	r.mu.RLock(); defer r.mu.RUnlock()
	for _, c := range r.items { if c.ID == id { return c, nil } }
	return domain.Competition{}, fmt.Errorf("competition %s not found", id)
}
func (r *compRepo) List(offset, limit int) ([]domain.Competition, int, error) {
	r.mu.RLock(); defer r.mu.RUnlock()
	return paginateSlice(r.items, offset, limit)
}
func (r *compRepo) Update(c domain.Competition) (domain.Competition, error) {
	r.mu.Lock(); defer r.mu.Unlock()
	for i, v := range r.items { if v.ID == c.ID { c.UpdatedAt = time.Now(); r.items[i] = c; return c, nil } }
	return domain.Competition{}, fmt.Errorf("competition %s not found", c.ID)
}
func (r *compRepo) CreateReg(reg domain.CompetitionReg) (domain.CompetitionReg, error) {
	r.mu.Lock(); defer r.mu.Unlock()
	reg.CreatedAt = time.Now()
	r.regs = append(r.regs, reg)
	return reg, nil
}
func (r *compRepo) ListRegs(competitionID string) ([]domain.CompetitionReg, error) {
	r.mu.RLock(); defer r.mu.RUnlock()
	out := make([]domain.CompetitionReg, 0)
	for _, reg := range r.regs { if reg.CompetitionID == competitionID { out = append(out, reg) } }
	return out, nil
}

// ---- Event ----

type eventRepo struct {
	mu    sync.RWMutex
	items []domain.AssociationEvent
	regs  []domain.EventRegistration
}

func NewEventRepository() repository.EventRepository { return &eventRepo{} }

func (r *eventRepo) Create(e domain.AssociationEvent) (domain.AssociationEvent, error) {
	r.mu.Lock(); defer r.mu.Unlock()
	e.CreatedAt = time.Now(); e.UpdatedAt = e.CreatedAt
	r.items = append(r.items, e); return e, nil
}
func (r *eventRepo) FindByID(id string) (domain.AssociationEvent, error) {
	r.mu.RLock(); defer r.mu.RUnlock()
	for _, e := range r.items { if e.ID == id { return e, nil } }
	return domain.AssociationEvent{}, fmt.Errorf("event %s not found", id)
}
func (r *eventRepo) List(offset, limit int) ([]domain.AssociationEvent, int, error) {
	r.mu.RLock(); defer r.mu.RUnlock()
	return paginateSlice(r.items, offset, limit)
}
func (r *eventRepo) Update(e domain.AssociationEvent) (domain.AssociationEvent, error) {
	r.mu.Lock(); defer r.mu.Unlock()
	for i, v := range r.items { if v.ID == e.ID { e.UpdatedAt = time.Now(); r.items[i] = e; return e, nil } }
	return domain.AssociationEvent{}, fmt.Errorf("event %s not found", e.ID)
}
func (r *eventRepo) CreateReg(reg domain.EventRegistration) (domain.EventRegistration, error) {
	r.mu.Lock(); defer r.mu.Unlock()
	reg.CreatedAt = time.Now()
	r.regs = append(r.regs, reg)
	return reg, nil
}
func (r *eventRepo) ListRegs(eventID string) ([]domain.EventRegistration, error) {
	r.mu.RLock(); defer r.mu.RUnlock()
	out := make([]domain.EventRegistration, 0)
	for _, reg := range r.regs { if reg.EventID == eventID { out = append(out, reg) } }
	return out, nil
}

// ---- Portfolio ----

type portfolioRepo struct {
	mu    sync.RWMutex
	items []domain.MemberPortfolio
}

func NewPortfolioRepository() repository.PortfolioRepository { return &portfolioRepo{} }

func (r *portfolioRepo) Create(p domain.MemberPortfolio) (domain.MemberPortfolio, error) {
	r.mu.Lock(); defer r.mu.Unlock()
	p.CreatedAt = time.Now(); p.UpdatedAt = p.CreatedAt
	r.items = append(r.items, p); return p, nil
}
func (r *portfolioRepo) FindByID(id string) (domain.MemberPortfolio, error) {
	r.mu.RLock(); defer r.mu.RUnlock()
	for _, p := range r.items { if p.ID == id { return p, nil } }
	return domain.MemberPortfolio{}, fmt.Errorf("portfolio %s not found", id)
}
func (r *portfolioRepo) ListByEnterprise(eid string) ([]domain.MemberPortfolio, error) {
	r.mu.RLock(); defer r.mu.RUnlock()
	out := make([]domain.MemberPortfolio, 0)
	for _, p := range r.items { if p.EnterpriseID == eid { out = append(out, p) } }
	return out, nil
}
func (r *portfolioRepo) ListPublished(offset, limit int) ([]domain.MemberPortfolio, int, error) {
	r.mu.RLock(); defer r.mu.RUnlock()
	filtered := make([]domain.MemberPortfolio, 0)
	for _, p := range r.items { if p.Status == "published" { filtered = append(filtered, p) } }
	return paginateSlice(filtered, offset, limit)
}
func (r *portfolioRepo) Update(p domain.MemberPortfolio) (domain.MemberPortfolio, error) {
	r.mu.Lock(); defer r.mu.Unlock()
	for i, v := range r.items { if v.ID == p.ID { p.UpdatedAt = time.Now(); r.items[i] = p; return p, nil } }
	return domain.MemberPortfolio{}, fmt.Errorf("portfolio %s not found", p.ID)
}

// ---- IndustryReport ----

type industryReportRepo struct {
	mu    sync.RWMutex
	items []domain.IndustryReport
}

func NewIndustryReportRepository() repository.IndustryReportRepository { return &industryReportRepo{} }

func (r *industryReportRepo) Create(rp domain.IndustryReport) (domain.IndustryReport, error) {
	r.mu.Lock(); defer r.mu.Unlock()
	rp.CreatedAt = time.Now(); rp.UpdatedAt = rp.CreatedAt
	r.items = append(r.items, rp); return rp, nil
}
func (r *industryReportRepo) FindByID(id string) (domain.IndustryReport, error) {
	r.mu.RLock(); defer r.mu.RUnlock()
	for _, rp := range r.items { if rp.ID == id { return rp, nil } }
	return domain.IndustryReport{}, fmt.Errorf("report %s not found", id)
}
func (r *industryReportRepo) List(offset, limit int) ([]domain.IndustryReport, int, error) {
	r.mu.RLock(); defer r.mu.RUnlock()
	return paginateSlice(r.items, offset, limit)
}
func (r *industryReportRepo) Update(rp domain.IndustryReport) (domain.IndustryReport, error) {
	r.mu.Lock(); defer r.mu.Unlock()
	for i, v := range r.items { if v.ID == rp.ID { rp.UpdatedAt = time.Now(); r.items[i] = rp; return rp, nil } }
	return domain.IndustryReport{}, fmt.Errorf("report %s not found", rp.ID)
}
func (r *industryReportRepo) Delete(id string) error {
	r.mu.Lock(); defer r.mu.Unlock()
	for i, v := range r.items { if v.ID == id { r.items = append(r.items[:i], r.items[i+1:]...); return nil } }
	return fmt.Errorf("report %s not found", id)
}

// ---- Resource ----

type resourceRepo struct {
	mu    sync.RWMutex
	items []domain.IndustryResource
}

func NewResourceRepository() repository.ResourceRepository { return &resourceRepo{} }

func (r *resourceRepo) Create(res domain.IndustryResource) (domain.IndustryResource, error) {
	r.mu.Lock(); defer r.mu.Unlock()
	res.CreatedAt = time.Now(); res.UpdatedAt = res.CreatedAt
	r.items = append(r.items, res); return res, nil
}
func (r *resourceRepo) FindByID(id string) (domain.IndustryResource, error) {
	r.mu.RLock(); defer r.mu.RUnlock()
	for _, res := range r.items { if res.ID == id { return res, nil } }
	return domain.IndustryResource{}, fmt.Errorf("resource %s not found", id)
}
func (r *resourceRepo) List(resType string, offset, limit int) ([]domain.IndustryResource, int, error) {
	r.mu.RLock(); defer r.mu.RUnlock()
	filtered := make([]domain.IndustryResource, 0)
	for _, res := range r.items { if resType == "" || res.ResType == resType { filtered = append(filtered, res) } }
	return paginateSlice(filtered, offset, limit)
}
func (r *resourceRepo) Update(res domain.IndustryResource) (domain.IndustryResource, error) {
	r.mu.Lock(); defer r.mu.Unlock()
	for i, v := range r.items { if v.ID == res.ID { res.UpdatedAt = time.Now(); r.items[i] = res; return res, nil } }
	return domain.IndustryResource{}, fmt.Errorf("resource %s not found", res.ID)
}

// ---- Emergency ----

type emergencyRepo struct {
	mu         sync.RWMutex
	resources  []domain.EmergencyResource
	dispatches []domain.EmergencyDispatch
}

func NewEmergencyRepository() repository.EmergencyRepository { return &emergencyRepo{} }

func (r *emergencyRepo) CreateResource(res domain.EmergencyResource) (domain.EmergencyResource, error) {
	r.mu.Lock(); defer r.mu.Unlock()
	res.CreatedAt = time.Now(); res.UpdatedAt = res.CreatedAt
	r.resources = append(r.resources, res); return res, nil
}
func (r *emergencyRepo) FindResourceByID(id string) (domain.EmergencyResource, error) {
	r.mu.RLock(); defer r.mu.RUnlock()
	for _, res := range r.resources { if res.ID == id { return res, nil } }
	return domain.EmergencyResource{}, fmt.Errorf("emergency resource %s not found", id)
}
func (r *emergencyRepo) ListResources(offset, limit int) ([]domain.EmergencyResource, int, error) {
	r.mu.RLock(); defer r.mu.RUnlock()
	return paginateSlice(r.resources, offset, limit)
}
func (r *emergencyRepo) UpdateResource(res domain.EmergencyResource) (domain.EmergencyResource, error) {
	r.mu.Lock(); defer r.mu.Unlock()
	for i, v := range r.resources {
		if v.ID == res.ID { res.UpdatedAt = time.Now(); r.resources[i] = res; return res, nil }
	}
	return domain.EmergencyResource{}, fmt.Errorf("emergency resource %s not found", res.ID)
}
func (r *emergencyRepo) CreateDispatch(d domain.EmergencyDispatch) (domain.EmergencyDispatch, error) {
	r.mu.Lock(); defer r.mu.Unlock()
	d.CreatedAt = time.Now()
	r.dispatches = append(r.dispatches, d); return d, nil
}
func (r *emergencyRepo) ListDispatches(offset, limit int) ([]domain.EmergencyDispatch, int, error) {
	r.mu.RLock(); defer r.mu.RUnlock()
	return paginateSlice(r.dispatches, offset, limit)
}
