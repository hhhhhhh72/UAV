package memory

import (
	"fmt"
	"sync"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)

// ── ResourcePool ──
type poolRepo struct {
	mu      sync.RWMutex
	pools   []domain.ResourcePool
	members []domain.ResourcePoolMember
}

func NewResourcePoolRepository() repository.ResourcePoolRepository { return &poolRepo{} }

func (r *poolRepo) Create(p domain.ResourcePool) (domain.ResourcePool, error) {
	r.mu.Lock(); defer r.mu.Unlock()
	r.pools = append(r.pools, p); return p, nil
}
func (r *poolRepo) FindByID(id string) (domain.ResourcePool, error) {
	r.mu.RLock(); defer r.mu.RUnlock()
	for _, p := range r.pools { if p.ID == id { return p, nil } }
	return domain.ResourcePool{}, fmt.Errorf("pool %s not found", id)
}
func (r *poolRepo) List(poolType string) ([]domain.ResourcePool, error) {
	r.mu.RLock(); defer r.mu.RUnlock()
	out := make([]domain.ResourcePool, 0)
	for _, p := range r.pools { if poolType == "" || p.PoolType == poolType { out = append(out, p) } }
	return out, nil
}
func (r *poolRepo) AddMember(m domain.ResourcePoolMember) (domain.ResourcePoolMember, error) {
	r.mu.Lock(); defer r.mu.Unlock()
	r.members = append(r.members, m); return m, nil
}
func (r *poolRepo) ListMembers(poolID string) ([]domain.ResourcePoolMember, error) {
	r.mu.RLock(); defer r.mu.RUnlock()
	out := make([]domain.ResourcePoolMember, 0)
	for _, m := range r.members { if m.PoolID == poolID { out = append(out, m) } }
	return out, nil
}

// ── TestSite ──
type testSiteRepo struct {
	mu       sync.RWMutex
	sites    []domain.TestSite
	bookings []domain.TestSiteBooking
}

func NewTestSiteRepository() repository.TestSiteRepository { return &testSiteRepo{} }

func (r *testSiteRepo) Create(ts domain.TestSite) (domain.TestSite, error) {
	r.mu.Lock(); defer r.mu.Unlock()
	r.sites = append(r.sites, ts); return ts, nil
}
func (r *testSiteRepo) FindByID(id string) (domain.TestSite, error) {
	r.mu.RLock(); defer r.mu.RUnlock()
	for _, ts := range r.sites { if ts.ID == id { return ts, nil } }
	return domain.TestSite{}, fmt.Errorf("testsite %s not found", id)
}
func (r *testSiteRepo) List(siteType string) ([]domain.TestSite, error) {
	r.mu.RLock(); defer r.mu.RUnlock()
	out := make([]domain.TestSite, 0)
	for _, ts := range r.sites { if siteType == "" || ts.SiteType == siteType { out = append(out, ts) } }
	return out, nil
}
func (r *testSiteRepo) CreateBooking(b domain.TestSiteBooking) (domain.TestSiteBooking, error) {
	r.mu.Lock(); defer r.mu.Unlock()
	r.bookings = append(r.bookings, b); return b, nil
}
func (r *testSiteRepo) UpdateBookingStatus(id, status, note string) (domain.TestSiteBooking, error) {
	r.mu.Lock(); defer r.mu.Unlock()
	for i, b := range r.bookings {
		if b.ID == id { r.bookings[i].Status = status; r.bookings[i].ReviewNote = note; return r.bookings[i], nil }
	}
	return domain.TestSiteBooking{}, fmt.Errorf("booking %s not found", id)
}
func (r *testSiteRepo) ListBookings(siteID string) ([]domain.TestSiteBooking, error) {
	r.mu.RLock(); defer r.mu.RUnlock()
	out := make([]domain.TestSiteBooking, 0)
	for _, b := range r.bookings { if b.SiteID == siteID { out = append(out, b) } }
	return out, nil
}

// ── Exhibition ──
type exhibitionRepo struct {
	mu      sync.RWMutex
	expos   []domain.Exhibition
	booths  []domain.ExhibitionBooth
}

func NewExhibitionRepository() repository.ExhibitionRepository { return &exhibitionRepo{} }

func (r *exhibitionRepo) Create(e domain.Exhibition) (domain.Exhibition, error) {
	r.mu.Lock(); defer r.mu.Unlock()
	e.CreatedAt = time.Now(); e.UpdatedAt = e.CreatedAt
	r.expos = append(r.expos, e); return e, nil
}
func (r *exhibitionRepo) FindByID(id string) (domain.Exhibition, error) {
	r.mu.RLock(); defer r.mu.RUnlock()
	for _, e := range r.expos { if e.ID == id { return e, nil } }
	return domain.Exhibition{}, fmt.Errorf("exhibition %s not found", id)
}
func (r *exhibitionRepo) List(offset, limit int) ([]domain.Exhibition, int, error) {
	r.mu.RLock(); defer r.mu.RUnlock()
	return paginateSlice(r.expos, offset, limit)
}
func (r *exhibitionRepo) CreateBooth(b domain.ExhibitionBooth) (domain.ExhibitionBooth, error) {
	r.mu.Lock(); defer r.mu.Unlock()
	r.booths = append(r.booths, b); return b, nil
}
func (r *exhibitionRepo) ListBooths(exhibitionID string) ([]domain.ExhibitionBooth, error) {
	r.mu.RLock(); defer r.mu.RUnlock()
	out := make([]domain.ExhibitionBooth, 0)
	for _, b := range r.booths { if b.ExhibitionID == exhibitionID { out = append(out, b) } }
	return out, nil
}
func (r *exhibitionRepo) UpdateBoothStatus(id, status string) (domain.ExhibitionBooth, error) {
	r.mu.Lock(); defer r.mu.Unlock()
	for i, b := range r.booths {
		if b.ID == id { r.booths[i].Status = status; return r.booths[i], nil }
	}
	return domain.ExhibitionBooth{}, fmt.Errorf("booth %s not found", id)
}
