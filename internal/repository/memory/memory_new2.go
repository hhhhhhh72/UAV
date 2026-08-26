package memory

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"drone-platform/internal/crypto"
	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)

// ---- Competition ----

type compRepo struct {
	mu     sync.RWMutex
	items  []domain.Competition
	regs   []domain.CompetitionReg
	cipher *crypto.Cipher // 报名实名信息（id_card/phone）静态加密（仿 pilotRepo）
}

func NewCompetitionRepository(cipher *crypto.Cipher) repository.CompetitionRepository {
	return &compRepo{cipher: cipher}
}

func (r *compRepo) Create(ctx context.Context, c domain.Competition) (domain.Competition, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	c.CreatedAt = time.Now()
	c.UpdatedAt = c.CreatedAt
	r.items = append(r.items, c)
	return c, nil
}
func (r *compRepo) FindByID(ctx context.Context, id string) (domain.Competition, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, c := range r.items {
		if c.ID == id {
			return c, nil
		}
	}
	return domain.Competition{}, fmt.Errorf("competition %s not found", id)
}
func (r *compRepo) List(ctx context.Context, offset, limit int) ([]domain.Competition, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := append([]domain.Competition(nil), r.items...)
	sort.SliceStable(items, func(i, j int) bool { return items[i].CreatedAt.After(items[j].CreatedAt) })
	return paginateSlice(items, offset, limit)
}
func (r *compRepo) Update(ctx context.Context, c domain.Competition) (domain.Competition, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, v := range r.items {
		if v.ID == c.ID {
			c.UpdatedAt = time.Now()
			c.RegCount = v.RegCount // 报名数由报名系统维护，编辑赛事不得清零（与 PG 一致）
			r.items[i] = c
			return c, nil
		}
	}
	return domain.Competition{}, fmt.Errorf("competition %s not found", c.ID)
}
func (r *compRepo) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.items {
		if r.items[i].ID == id {
			r.items = append(r.items[:i], r.items[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("competition %s not found", id)
}

// encryptRegInPlace 对报名实名字段静态加密（cipher 为空或加密失败时保留明文，兼容无 ENCRYPTION_KEY 的开发环境）。
func (r *compRepo) encryptRegInPlace(reg *domain.CompetitionReg) {
	if r.cipher == nil {
		return
	}
	for _, v := range []struct {
		src  string
		dest *string
	}{{reg.IDCard, &reg.IDCard}, {reg.Phone, &reg.Phone}} {
		if v.src == "" {
			continue
		}
		if enc, err := r.cipher.Encrypt(v.src); err == nil {
			*v.dest = enc
		}
	}
}
func (r *compRepo) decryptRegInPlace(reg *domain.CompetitionReg) {
	if r.cipher == nil {
		return
	}
	for _, v := range []struct {
		src  string
		dest *string
	}{{reg.IDCard, &reg.IDCard}, {reg.Phone, &reg.Phone}} {
		if v.src == "" {
			continue
		}
		if dec, err := r.cipher.Decrypt(v.src); err == nil {
			*v.dest = dec
		}
	}
}
func (r *compRepo) CreateReg(ctx context.Context, reg domain.CompetitionReg) (domain.CompetitionReg, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	reg.CreatedAt = time.Now()
	// 与 PG 唯一索引（uniq_competition_regs_user_comp）对齐：同用户同赛事只能报名一次。
	for _, x := range r.regs {
		if x.CompetitionID == reg.CompetitionID && x.UserID == reg.UserID {
			return domain.CompetitionReg{}, fmt.Errorf("已报名过该赛事，请勿重复报名")
		}
	}
	r.encryptRegInPlace(&reg)
	r.regs = append(r.regs, reg)
	// 与 PG 事务内 reg_count+1 对齐（含容量上限：max_teams<=0 不限制）。
	for i := range r.items {
		if r.items[i].ID == reg.CompetitionID {
			if r.items[i].MaxTeams > 0 && r.items[i].RegCount >= r.items[i].MaxTeams {
				r.decryptRegInPlace(&reg)
				return domain.CompetitionReg{}, fmt.Errorf("competition is full")
			}
			r.items[i].RegCount++
			break
		}
	}
	r.decryptRegInPlace(&reg)
	return reg, nil
}
func (r *compRepo) ListRegs(ctx context.Context, competitionID string) ([]domain.CompetitionReg, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.CompetitionReg, 0)
	for _, reg := range r.regs {
		if reg.CompetitionID == competitionID {
			r.decryptRegInPlace(&reg)
			out = append(out, reg)
		}
	}
	// 与 PG 对齐：ORDER BY created_at DESC。
	sort.SliceStable(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}

// ---- Event ----

type eventRepo struct {
	mu    sync.RWMutex
	items []domain.AssociationEvent
	regs  []domain.EventRegistration
}

func NewEventRepository() repository.EventRepository { return &eventRepo{} }

func (r *eventRepo) Create(ctx context.Context, e domain.AssociationEvent) (domain.AssociationEvent, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	e.CreatedAt = time.Now()
	e.UpdatedAt = e.CreatedAt
	r.items = append(r.items, e)
	return e, nil
}
func (r *eventRepo) FindByID(ctx context.Context, id string) (domain.AssociationEvent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, e := range r.items {
		if e.ID == id {
			return e, nil
		}
	}
	return domain.AssociationEvent{}, fmt.Errorf("event %s not found", id)
}
func (r *eventRepo) List(ctx context.Context, offset, limit int) ([]domain.AssociationEvent, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := append([]domain.AssociationEvent(nil), r.items...)
	sort.SliceStable(items, func(i, j int) bool { return items[i].CreatedAt.After(items[j].CreatedAt) })
	return paginateSlice(items, offset, limit)
}
func (r *eventRepo) Update(ctx context.Context, e domain.AssociationEvent) (domain.AssociationEvent, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, v := range r.items {
		if v.ID == e.ID {
			e.UpdatedAt = time.Now()
			r.items[i] = e
			return e, nil
		}
	}
	return domain.AssociationEvent{}, fmt.Errorf("event %s not found", e.ID)
}
func (r *eventRepo) CreateReg(ctx context.Context, reg domain.EventRegistration) (domain.EventRegistration, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	reg.CreatedAt = time.Now()
	// 与 PG 唯一索引（uniq_event_regs_user_event）对齐：同用户同活动只能报名一次。
	for _, x := range r.regs {
		if x.EventID == reg.EventID && x.UserID == reg.UserID {
			return domain.EventRegistration{}, fmt.Errorf("已报名过该活动，请勿重复报名")
		}
	}
	r.regs = append(r.regs, reg)
	// 与 PG 事务内 reg_count+1 对齐（含容量上限：max_attendees<=0 不限制）。
	for i := range r.items {
		if r.items[i].ID == reg.EventID {
			if r.items[i].MaxAttendees > 0 && r.items[i].RegCount >= r.items[i].MaxAttendees {
				return domain.EventRegistration{}, fmt.Errorf("event is full")
			}
			r.items[i].RegCount++
			break
		}
	}
	return reg, nil
}
func (r *eventRepo) ListRegs(ctx context.Context, eventID string) ([]domain.EventRegistration, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.EventRegistration, 0)
	for _, reg := range r.regs {
		if reg.EventID == eventID {
			out = append(out, reg)
		}
	}
	// 与 PG 对齐：ORDER BY created_at DESC。
	sort.SliceStable(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}

// ---- Portfolio ----

type portfolioRepo struct {
	mu    sync.RWMutex
	items []domain.MemberPortfolio
}

func NewPortfolioRepository() repository.PortfolioRepository { return &portfolioRepo{} }

func (r *portfolioRepo) Create(ctx context.Context, p domain.MemberPortfolio) (domain.MemberPortfolio, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	p.CreatedAt = time.Now()
	p.UpdatedAt = p.CreatedAt
	r.items = append(r.items, p)
	return p, nil
}
func (r *portfolioRepo) FindByID(ctx context.Context, id string) (domain.MemberPortfolio, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, p := range r.items {
		if p.ID == id {
			return p, nil
		}
	}
	return domain.MemberPortfolio{}, fmt.Errorf("portfolio %s not found", id)
}
func (r *portfolioRepo) ListByEnterprise(ctx context.Context, eid string) ([]domain.MemberPortfolio, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.MemberPortfolio, 0)
	for _, p := range r.items {
		if p.EnterpriseID == eid {
			out = append(out, p)
		}
	}
	// 与 PG 对齐：ORDER BY created_at DESC。
	sort.SliceStable(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}
func (r *portfolioRepo) ListPublished(ctx context.Context, offset, limit int) ([]domain.MemberPortfolio, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	filtered := make([]domain.MemberPortfolio, 0)
	for _, p := range r.items {
		if p.Status == "published" {
			filtered = append(filtered, p)
		}
	}
	sort.SliceStable(filtered, func(i, j int) bool { return filtered[i].CreatedAt.After(filtered[j].CreatedAt) })
	return paginateSlice(filtered, offset, limit)
}
func (r *portfolioRepo) List(ctx context.Context, offset, limit int) ([]domain.MemberPortfolio, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := append([]domain.MemberPortfolio(nil), r.items...)
	sort.SliceStable(items, func(i, j int) bool { return items[i].CreatedAt.After(items[j].CreatedAt) })
	return paginateSlice(items, offset, limit)
}
func (r *portfolioRepo) Update(ctx context.Context, p domain.MemberPortfolio) (domain.MemberPortfolio, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, v := range r.items {
		if v.ID == p.ID {
			p.UpdatedAt = time.Now()
			r.items[i] = p
			return p, nil
		}
	}
	return domain.MemberPortfolio{}, fmt.Errorf("portfolio %s not found", p.ID)
}

func (r *portfolioRepo) IncrementViews(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.items {
		if r.items[i].ID == id {
			r.items[i].Views++
			r.items[i].UpdatedAt = time.Now()
			return nil
		}
	}
	return fmt.Errorf("portfolio %s not found", id)
}

func (r *portfolioRepo) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.items {
		if r.items[i].ID == id {
			r.items = append(r.items[:i], r.items[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("portfolio %s not found", id)
}

func (r *eventRepo) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.items {
		if r.items[i].ID == id {
			r.items = append(r.items[:i], r.items[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("event %s not found", id)
}

// ---- IndustryReport ----

type industryReportRepo struct {
	mu    sync.RWMutex
	items []domain.IndustryReport
}

func NewIndustryReportRepository() repository.IndustryReportRepository { return &industryReportRepo{} }

func (r *industryReportRepo) Create(ctx context.Context, rp domain.IndustryReport) (domain.IndustryReport, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	rp.CreatedAt = time.Now()
	rp.UpdatedAt = rp.CreatedAt
	r.items = append(r.items, rp)
	return rp, nil
}
func (r *industryReportRepo) FindByID(ctx context.Context, id string) (domain.IndustryReport, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, rp := range r.items {
		if rp.ID == id {
			return rp, nil
		}
	}
	return domain.IndustryReport{}, fmt.Errorf("report %s not found", id)
}
func (r *industryReportRepo) List(ctx context.Context, offset, limit int) ([]domain.IndustryReport, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := append([]domain.IndustryReport(nil), r.items...)
	sort.SliceStable(items, func(i, j int) bool { return items[i].CreatedAt.After(items[j].CreatedAt) })
	return paginateSlice(items, offset, limit)
}
func (r *industryReportRepo) Update(ctx context.Context, rp domain.IndustryReport) (domain.IndustryReport, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, v := range r.items {
		if v.ID == rp.ID {
			rp.UpdatedAt = time.Now()
			r.items[i] = rp
			return rp, nil
		}
	}
	return domain.IndustryReport{}, fmt.Errorf("report %s not found", rp.ID)
}
func (r *industryReportRepo) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, v := range r.items {
		if v.ID == id {
			r.items = append(r.items[:i], r.items[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("report %s not found", id)
}

// ---- Resource ----

type resourceRepo struct {
	mu       sync.RWMutex
	items    []domain.IndustryResource
	bookings []domain.IndustryResourceBooking
}

func NewResourceRepository() repository.ResourceRepository { return &resourceRepo{} }

func (r *resourceRepo) Create(ctx context.Context, res domain.IndustryResource) (domain.IndustryResource, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	res.CreatedAt = time.Now()
	res.UpdatedAt = res.CreatedAt
	r.items = append(r.items, res)
	return res, nil
}
func (r *resourceRepo) FindByID(ctx context.Context, id string) (domain.IndustryResource, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, res := range r.items {
		if res.ID == id {
			return res, nil
		}
	}
	return domain.IndustryResource{}, fmt.Errorf("resource %s not found", id)
}
func (r *resourceRepo) List(ctx context.Context, resType string, offset, limit int) ([]domain.IndustryResource, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	filtered := make([]domain.IndustryResource, 0)
	for _, res := range r.items {
		if resType == "" || res.ResType == resType {
			filtered = append(filtered, res)
		}
	}
	sort.SliceStable(filtered, func(i, j int) bool { return filtered[i].CreatedAt.After(filtered[j].CreatedAt) })
	return paginateSlice(filtered, offset, limit)
}
func (r *resourceRepo) Update(ctx context.Context, res domain.IndustryResource) (domain.IndustryResource, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, v := range r.items {
		if v.ID == res.ID {
			res.UpdatedAt = time.Now()
			r.items[i] = res
			return res, nil
		}
	}
	return domain.IndustryResource{}, fmt.Errorf("resource %s not found", res.ID)
}

func (r *resourceRepo) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.items {
		if r.items[i].ID == id {
			r.items = append(r.items[:i], r.items[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("resource %s not found", id)
}

// ---- Resource bookings (C11) ----

func (r *resourceRepo) CreateBooking(ctx context.Context, b domain.IndustryResourceBooking) (domain.IndustryResourceBooking, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	b.CreatedAt = time.Now()
	b.UpdatedAt = b.CreatedAt
	r.bookings = append(r.bookings, b)
	return b, nil
}
func (r *resourceRepo) ListBookingsByResource(ctx context.Context, resourceID string) ([]domain.IndustryResourceBooking, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.IndustryResourceBooking, 0)
	for _, b := range r.bookings {
		if b.ResourceID == resourceID {
			out = append(out, b)
		}
	}
	// 与 PG 对齐：ORDER BY created_at DESC。
	sort.SliceStable(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}
func (r *resourceRepo) ListBookingsByUser(ctx context.Context, userID string) ([]domain.IndustryResourceBooking, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.IndustryResourceBooking, 0)
	for _, b := range r.bookings {
		if b.UserID == userID {
			out = append(out, b)
		}
	}
	// 与 PG 对齐：ORDER BY created_at DESC。
	sort.SliceStable(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}

// ---- Emergency ----

type emergencyRepo struct {
	mu         sync.RWMutex
	resources  []domain.EmergencyResource
	dispatches []domain.EmergencyDispatch
}

func NewEmergencyRepository() repository.EmergencyRepository { return &emergencyRepo{} }

func (r *emergencyRepo) CreateResource(ctx context.Context, res domain.EmergencyResource) (domain.EmergencyResource, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	res.CreatedAt = time.Now()
	res.UpdatedAt = res.CreatedAt
	r.resources = append(r.resources, res)
	return res, nil
}
func (r *emergencyRepo) FindResourceByID(ctx context.Context, id string) (domain.EmergencyResource, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, res := range r.resources {
		if res.ID == id {
			return res, nil
		}
	}
	return domain.EmergencyResource{}, fmt.Errorf("emergency resource %s not found", id)
}
func (r *emergencyRepo) ListResources(ctx context.Context, resType, q string, offset, limit int) ([]domain.EmergencyResource, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	query := strings.ToLower(strings.TrimSpace(q))
	filtered := make([]domain.EmergencyResource, 0)
	for _, res := range r.resources {
		if resType != "" && res.ResType != resType {
			continue
		}
		if query != "" && !matchAnyFold(query, res.Name, res.Specs, res.Location, res.ContactInfo) {
			continue
		}
		filtered = append(filtered, res)
	}
	sort.SliceStable(filtered, func(i, j int) bool { return filtered[i].CreatedAt.After(filtered[j].CreatedAt) })
	return paginateSlice(filtered, offset, limit)
}
func (r *emergencyRepo) UpdateResource(ctx context.Context, res domain.EmergencyResource) (domain.EmergencyResource, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, v := range r.resources {
		if v.ID == res.ID {
			res.UpdatedAt = time.Now()
			r.resources[i] = res
			return res, nil
		}
	}
	return domain.EmergencyResource{}, fmt.Errorf("emergency resource %s not found", res.ID)
}
func (r *emergencyRepo) AdjustResourceQuantity(ctx context.Context, id string, delta int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, v := range r.resources {
		if v.ID == id {
			q := v.Quantity + delta
			if q < 0 {
				q = 0
			}
			r.resources[i].Quantity = q
			r.resources[i].UpdatedAt = time.Now()
			return nil
		}
	}
	return fmt.Errorf("emergency resource %s not found", id)
}
func (r *emergencyRepo) CreateDispatch(ctx context.Context, d domain.EmergencyDispatch) (domain.EmergencyDispatch, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	d.CreatedAt = time.Now()
	r.dispatches = append(r.dispatches, d)
	return d, nil
}
func (r *emergencyRepo) ListDispatches(ctx context.Context, resourceID string, offset, limit int) ([]domain.EmergencyDispatch, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	// 按资源过滤 + 内嵌 related 资源摘要（与 PG 实现语义一致；锁内直接查 resources，避免读锁重入）
	var filtered []domain.EmergencyDispatch
	for _, d := range r.dispatches {
		if resourceID != "" && d.ResourceID != resourceID {
			continue
		}
		for _, res := range r.resources {
			if res.ID == d.ResourceID {
				d.Related = &domain.EmergencyResourceBrief{ID: res.ID, Name: res.Name, ResType: res.ResType, Status: res.Status}
				break
			}
		}
		filtered = append(filtered, d)
	}
	sort.SliceStable(filtered, func(i, j int) bool { return filtered[i].CreatedAt.After(filtered[j].CreatedAt) })
	page, total, err := paginateSlice(filtered, offset, limit)
	return page, total, err
}

func (r *emergencyRepo) DeleteResource(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.resources {
		if r.resources[i].ID == id {
			r.resources = append(r.resources[:i], r.resources[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("resource %s not found", id)
}

func (r *emergencyRepo) FindDispatchByID(ctx context.Context, id string) (domain.EmergencyDispatch, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, d := range r.dispatches {
		if d.ID == id {
			return d, nil
		}
	}
	return domain.EmergencyDispatch{}, fmt.Errorf("dispatch %s not found", id)
}

func (r *emergencyRepo) UpdateDispatch(ctx context.Context, d domain.EmergencyDispatch) (domain.EmergencyDispatch, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, v := range r.dispatches {
		if v.ID == d.ID {
			r.dispatches[i] = d
			return d, nil
		}
	}
	return domain.EmergencyDispatch{}, fmt.Errorf("dispatch %s not found", d.ID)
}

func (r *emergencyRepo) DeleteDispatch(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.dispatches {
		if r.dispatches[i].ID == id {
			r.dispatches = append(r.dispatches[:i], r.dispatches[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("dispatch %s not found", id)
}
