package memory

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"drone-platform/internal/crypto"
	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)

// ---- Demand ----

type demandRepo struct {
	mu     sync.RWMutex
	items  []domain.Demand
	cipher *crypto.Cipher
}

func NewDemandRepository(cipher *crypto.Cipher) repository.DemandRepository {
	now := time.Now()
	d := domain.Demand{
		ID:            "demand-001",
		PublisherID:   "enterprise-001",
		PublisherName: "重庆巡航科技",
		Contact:       "138****8888",
		District:      "南岸区",
		BizType:       domain.BizCableInspection,
		Title:         "南岸区线路巡检作业",
		Description:   "需要具备行业经验的飞手团队",
		Status:        domain.DemandPending,
		Version:       1,
		CreatedAt:     now.Add(-time.Hour),
		UpdatedAt:     now,
	}
	if cipher != nil && d.Contact != "" {
		enc, err := cipher.Encrypt(d.Contact)
		if err == nil {
			d.Contact = enc
		}
	}
	return &demandRepo{items: []domain.Demand{d}, cipher: cipher}
}

func (r *demandRepo) decrypt(d *domain.Demand) {
	if r.cipher != nil && d.Contact != "" {
		if dec, err := r.cipher.Decrypt(d.Contact); err == nil {
			d.Contact = dec
		}
	}
}

func (r *demandRepo) Create(d domain.Demand) (domain.Demand, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.cipher != nil && d.Contact != "" {
		enc, err := r.cipher.Encrypt(d.Contact)
		if err != nil {
			return domain.Demand{}, fmt.Errorf("encrypt contact: %w", err)
		}
		d.Contact = enc
	}
	r.items = append(r.items, d)
	return d, nil
}

func (r *demandRepo) Update(d domain.Demand) (domain.Demand, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.items {
		if r.items[i].ID == d.ID {
			if r.cipher != nil && d.Contact != "" {
				enc, err := r.cipher.Encrypt(d.Contact)
				if err != nil {
					return domain.Demand{}, fmt.Errorf("encrypt contact: %w", err)
				}
				d.Contact = enc
			}
			r.items[i] = d
			return d, nil
		}
	}
	return domain.Demand{}, fmt.Errorf("demand %s not found", d.ID)
}

func (r *demandRepo) FindByID(id string) (domain.Demand, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, d := range r.items {
		if d.ID == id {
			r.decrypt(&d)
			return d, nil
		}
	}
	return domain.Demand{}, fmt.Errorf("demand %s not found", id)
}

func (r *demandRepo) List(f repository.DemandFilter) ([]domain.Demand, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.Demand, 0)
	for _, d := range r.items {
		if f.Status == "all" {
			// 管理员查看全部，不过滤
		} else if f.Status != "" {
			if string(d.Status) != f.Status { continue }
		} else if d.Status != domain.DemandPublished {
			continue
		}
		if f.District != "" && d.District != f.District {
			continue
		}
		if f.BizType != "" && string(d.BizType) != f.BizType {
			continue
		}
		r.decrypt(&d)
		out = append(out, d)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}

func (r *demandRepo) Search(q string) ([]domain.Demand, error) {
	all, _ := r.List(repository.DemandFilter{})
	q = strings.ToLower(q)
	out := []domain.Demand{}
	for _, d := range all {
		if strings.Contains(strings.ToLower(d.Title+d.PublisherName+d.Description), q) {
			out = append(out, d)
		}
	}
	return out, nil
}

func (r *demandRepo) SetStatus(id string, status domain.DemandStatus) (domain.Demand, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.items {
		if r.items[i].ID == id {
			r.items[i].Status = status
			item := r.items[i] // copy
			r.decrypt(&item)   // decrypt on copy, do not mutate storage
			return item, nil
		}
	}
	return domain.Demand{}, fmt.Errorf("demand %s not found", id)
}

func (r *demandRepo) CompareAndSetStatus(id string, oldStatus, newStatus domain.DemandStatus) (bool, domain.Demand, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.items {
		if r.items[i].ID == id {
			if r.items[i].Status != oldStatus {
				return false, r.items[i], nil
			}
			r.items[i].Status = newStatus
			r.items[i].UpdatedAt = time.Now()
			r.items[i].Version++
			item := r.items[i]
			r.decrypt(&item)
			return true, item, nil
		}
	}
	return false, domain.Demand{}, fmt.Errorf("demand %s not found", id)
}

// ---- Enterprise ----

type enterpriseRepo struct {
	mu     sync.RWMutex
	items  []domain.Enterprise
	cipher *crypto.Cipher
}

func NewEnterpriseRepository(cipher *crypto.Cipher) repository.EnterpriseRepository {
	return &enterpriseRepo{items: []domain.Enterprise{}, cipher: cipher}
}

func (r *enterpriseRepo) encrypt(e *domain.Enterprise) error {
	if r.cipher == nil {
		return nil
	}
	if e.LicenseURL != "" {
		enc, err := r.cipher.Encrypt(e.LicenseURL)
		if err != nil {
			return fmt.Errorf("encrypt license_url: %w", err)
		}
		e.LicenseURL = enc
	}
	if e.AccountName != "" {
		enc, err := r.cipher.Encrypt(e.AccountName)
		if err != nil {
			return fmt.Errorf("encrypt account_name: %w", err)
		}
		e.AccountName = enc
	}
	return nil
}

func (r *enterpriseRepo) decrypt(e *domain.Enterprise) {
	if r.cipher == nil {
		return
	}
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

func (r *enterpriseRepo) Pending() ([]domain.Enterprise, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := []domain.Enterprise{}
	for _, e := range r.items {
		if e.Status == domain.EnterpriseSubmitted {
			r.decrypt(&e)
			out = append(out, e)
		}
	}
	return out, nil
}

func (r *enterpriseRepo) Create(e domain.Enterprise) (domain.Enterprise, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := r.encrypt(&e); err != nil {
		return domain.Enterprise{}, err
	}
	r.items = append(r.items, e)
	return e, nil
}

func (r *enterpriseRepo) Update(id string, e domain.Enterprise) (domain.Enterprise, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.items {
		if r.items[i].ID == id {
			if err := r.encrypt(&e); err != nil {
				return domain.Enterprise{}, err
			}
			r.items[i] = e
			return e, nil
		}
	}
	return domain.Enterprise{}, fmt.Errorf("enterprise %s not found", id)
}

func (r *enterpriseRepo) FindByID(id string) (domain.Enterprise, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, e := range r.items {
		if e.ID == id {
			r.decrypt(&e)
			return e, nil
		}
	}
	return domain.Enterprise{}, fmt.Errorf("enterprise %s not found", id)
}

func (r *enterpriseRepo) FindByOwner(userID string) ([]domain.Enterprise, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := []domain.Enterprise{}
	for _, e := range r.items {
		if e.OwnerUserID == userID {
			r.decrypt(&e)
			out = append(out, e)
		}
	}
	return out, nil
}

func (r *enterpriseRepo) ListByStatus(status string, offset, limit int) ([]domain.Enterprise, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	filtered := []domain.Enterprise{}
	for _, e := range r.items {
		if status == "" || string(e.Status) == status {
			r.decrypt(&e)
			filtered = append(filtered, e)
		}
	}
	total := len(filtered)
	if offset > len(filtered) {
		return nil, total, nil
	}
	end := offset + limit
	if end > len(filtered) {
		end = len(filtered)
	}
	return filtered[offset:end], total, nil
}

func (r *enterpriseRepo) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, e := range r.items {
		if e.ID == id {
			r.items = append(r.items[:i], r.items[i+1:]...)
			return nil
		}
	}
	return nil
}

func (r *enterpriseRepo) Search(q string) ([]domain.Enterprise, error) {
	q = strings.ToLower(q)
	out := []domain.Enterprise{}
	for _, e := range r.items {
		if strings.Contains(strings.ToLower(e.Name), q) {
			r.decrypt(&e)
			out = append(out, e)
		}
	}
	return out, nil
}

// paginateSlice returns a subslice of items for the given page, along with the total count.
func paginateSlice[T any](items []T, offset, limit int) ([]T, int, error) {
	total := len(items)
	start := offset
	if start > total {
		start = total
	}
	end := start + limit
	if end > total {
		end = total
	}
	return items[start:end], total, nil
}

// ---- Employment ----

type employmentRepo struct {
	mu    sync.RWMutex
	items []domain.EmploymentRequest
}

func NewEmploymentRepository() repository.EmploymentRepository {
	return &employmentRepo{}
}

func (r *employmentRepo) Create(v domain.EmploymentRequest) (domain.EmploymentRequest, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items = append(r.items, v)
	return v, nil
}

func (r *employmentRepo) ListByEnterprise(eid string, offset, limit int) ([]domain.EmploymentRequest, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	filtered := make([]domain.EmploymentRequest, 0)
	for _, v := range r.items {
		if v.EnterpriseID == eid {
			filtered = append(filtered, v)
		}
	}
	page, total, _ := paginateSlice(filtered, offset, limit)
	return page, total, nil
}

func (r *employmentRepo) ListAll(offset, limit int) ([]domain.EmploymentRequest, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	page, total, _ := paginateSlice(r.items, offset, limit)
	return append([]domain.EmploymentRequest(nil), page...), total, nil
}

// ---- Jobs ----

type jobRepo struct {
	mu    sync.RWMutex
	items []domain.Job
}

func NewJobRepository() repository.JobRepository { return &jobRepo{} }

func (r *jobRepo) Create(j domain.Job) (domain.Job, error) {
	r.mu.Lock(); defer r.mu.Unlock()
	r.items = append(r.items, j)
	return j, nil
}
func (r *jobRepo) Update(id string, j domain.Job) (domain.Job, error) {
	r.mu.Lock(); defer r.mu.Unlock()
	for i := range r.items {
		if r.items[i].ID == id { r.items[i] = j; return j, nil }
	}
	return domain.Job{}, fmt.Errorf("job %s not found", id)
}
func (r *jobRepo) FindByID(id string) (domain.Job, error) {
	r.mu.RLock(); defer r.mu.RUnlock()
	for _, j := range r.items {
		if j.ID == id { return j, nil }
	}
	return domain.Job{}, fmt.Errorf("job %s not found", id)
}
func (r *jobRepo) ListByEnterprise(eid string) ([]domain.Job, error) {
	r.mu.RLock(); defer r.mu.RUnlock()
	out := []domain.Job{}
	for _, j := range r.items {
		if j.EnterpriseID == eid { out = append(out, j) }
	}
	return out, nil
}
func (r *jobRepo) ListPublished(offset, limit int) ([]domain.Job, int, error) {
	r.mu.RLock(); defer r.mu.RUnlock()
	filtered := []domain.Job{}
	for _, j := range r.items {
		if j.Status == domain.JobPublished { filtered = append(filtered, j) }
	}
	total := len(filtered)
	if offset > total { return nil, total, nil }
	end := offset + limit; if end > total { end = total }
	return filtered[offset:end], total, nil
}

type resumeRepo struct {
	mu    sync.RWMutex
	items []domain.Resume
}

func NewResumeRepository() repository.ResumeRepository { return &resumeRepo{} }

func (r *resumeRepo) Create(v domain.Resume) (domain.Resume, error) {
	r.mu.Lock(); defer r.mu.Unlock()
	r.items = append(r.items, v)
	return v, nil
}
func (r *resumeRepo) Update(id string, v domain.Resume) (domain.Resume, error) {
	r.mu.Lock(); defer r.mu.Unlock()
	for i := range r.items {
		if r.items[i].ID == id { r.items[i] = v; return v, nil }
	}
	return domain.Resume{}, fmt.Errorf("resume %s not found", id)
}
func (r *resumeRepo) FindByID(id string) (domain.Resume, error) {
	r.mu.RLock(); defer r.mu.RUnlock()
	for _, v := range r.items { if v.ID == id { return v, nil } }
	return domain.Resume{}, fmt.Errorf("resume %s not found", id)
}
func (r *resumeRepo) ListByUser(userID string) ([]domain.Resume, error) {
	r.mu.RLock(); defer r.mu.RUnlock()
	out := []domain.Resume{}
	for _, v := range r.items { if v.UserID == userID { out = append(out, v) } }
	return out, nil
}

type applicationRepo struct {
	mu    sync.RWMutex
	items []domain.JobApplication
}

func NewJobApplicationRepository() repository.JobApplicationRepository { return &applicationRepo{} }

func (r *applicationRepo) Create(a domain.JobApplication) (domain.JobApplication, error) {
	r.mu.Lock(); defer r.mu.Unlock()
	r.items = append(r.items, a)
	return a, nil
}
func (r *applicationRepo) UpdateStatus(id string, status domain.AppStatus) (domain.JobApplication, error) {
	r.mu.Lock(); defer r.mu.Unlock()
	for i := range r.items {
		if r.items[i].ID == id { r.items[i].Status = status; return r.items[i], nil }
	}
	return domain.JobApplication{}, fmt.Errorf("application %s not found", id)
}
func (r *applicationRepo) ListByJob(jobID string) ([]domain.JobApplication, error) {
	r.mu.RLock(); defer r.mu.RUnlock()
	out := []domain.JobApplication{}
	for _, a := range r.items { if a.JobID == jobID { out = append(out, a) } }
	return out, nil
}
func (r *applicationRepo) ListByApplicant(userID string) ([]domain.JobApplication, error) {
	r.mu.RLock(); defer r.mu.RUnlock()
	out := []domain.JobApplication{}
	for _, a := range r.items { if a.ApplicantID == userID { out = append(out, a) } }
	return out, nil
}

// ---- Post ----

type postRepo struct {
	mu    sync.RWMutex
	items []domain.Post
}

func NewPostRepository() repository.PostRepository { return &postRepo{} }

func (r *postRepo) Create(p domain.Post) (domain.Post, error) {
	r.mu.Lock(); defer r.mu.Unlock(); r.items = append(r.items, p); return p, nil
}
func (r *postRepo) Update(id string, p domain.Post) (domain.Post, error) {
	r.mu.Lock(); defer r.mu.Unlock()
	for i := range r.items { if r.items[i].ID == id { r.items[i] = p; return p, nil } }
	return domain.Post{}, fmt.Errorf("post %s not found", id)
}
func (r *postRepo) FindByID(id string) (domain.Post, error) {
	r.mu.RLock(); defer r.mu.RUnlock()
	for _, p := range r.items { if p.ID == id { return p, nil } }
	return domain.Post{}, fmt.Errorf("post %s not found", id)
}
func (r *postRepo) ListPublished(offset, limit int) ([]domain.Post, int, error) {
	r.mu.RLock(); defer r.mu.RUnlock()
	f := []domain.Post{}
	for _, p := range r.items { if p.Status == "published" { f = append(f, p) } }
	t := len(f); if offset > t { return nil, t, nil }; e := offset + limit; if e > t { e = t }
	return f[offset:e], t, nil
}
func (r *postRepo) ListByAuthor(uid string) ([]domain.Post, error) {
	r.mu.RLock(); defer r.mu.RUnlock()
	out := []domain.Post{}
	for _, p := range r.items { if p.AuthorID == uid { out = append(out, p) } }
	return out, nil
}

type commentRepo struct {
	mu    sync.RWMutex
	items []domain.Comment
}

func NewCommentRepository() repository.CommentRepository { return &commentRepo{} }

func (r *commentRepo) Create(c domain.Comment) (domain.Comment, error) {
	r.mu.Lock(); defer r.mu.Unlock(); r.items = append(r.items, c); return c, nil
}
func (r *commentRepo) ListByPost(postID string) ([]domain.Comment, error) {
	r.mu.RLock(); defer r.mu.RUnlock()
	out := []domain.Comment{}
	for _, c := range r.items { if c.PostID == postID { out = append(out, c) } }
	return out, nil
}

type reportRepo struct {
	mu    sync.RWMutex
	items []domain.Report
}

func NewReportRepository() repository.ReportRepository { return &reportRepo{} }

func (r *reportRepo) Create(rp domain.Report) (domain.Report, error) {
	r.mu.Lock(); defer r.mu.Unlock(); r.items = append(r.items, rp); return rp, nil
}
func (r *reportRepo) ListPending(offset, limit int) ([]domain.Report, int, error) {
	r.mu.RLock(); defer r.mu.RUnlock()
	f := []domain.Report{}
	for _, rp := range r.items { if rp.Status == "pending" { f = append(f, rp) } }
	t := len(f); if offset > t { return nil, t, nil }; e := offset + limit; if e > t { e = t }
	return f[offset:e], t, nil
}

// ---- Listing ----

type listingRepo struct {
	mu        sync.RWMutex
	items     []domain.Listing
	favorites map[string][]string // listingID -> []userID
}

func NewListingRepository() repository.ListingRepository {
	return &listingRepo{favorites: make(map[string][]string)}
}

func (r *listingRepo) Create(l domain.Listing) (domain.Listing, error) {
	r.mu.Lock(); defer r.mu.Unlock(); r.items = append(r.items, l); return l, nil
}
func (r *listingRepo) Update(id string, l domain.Listing) (domain.Listing, error) {
	r.mu.Lock(); defer r.mu.Unlock()
	for i := range r.items { if r.items[i].ID == id { r.items[i] = l; return l, nil } }
	return domain.Listing{}, fmt.Errorf("listing %s not found", id)
}
func (r *listingRepo) FindByID(id string) (domain.Listing, error) {
	r.mu.RLock(); defer r.mu.RUnlock()
	for _, l := range r.items { if l.ID == id { return l, nil } }
	return domain.Listing{}, fmt.Errorf("listing %s not found", id)
}
func (r *listingRepo) ListByStatus(status string, offset, limit int) ([]domain.Listing, int, error) {
	r.mu.RLock(); defer r.mu.RUnlock()
	f := []domain.Listing{}
	for _, l := range r.items { if l.Status == status || status == "" { f = append(f, l) } }
	t := len(f); if offset > t { return nil, t, nil }; e := offset + limit; if e > t { e = t }
	return f[offset:e], t, nil
}
func (r *listingRepo) ListBySeller(uid string) ([]domain.Listing, error) {
	r.mu.RLock(); defer r.mu.RUnlock()
	out := []domain.Listing{}
	for _, l := range r.items { if l.SellerID == uid { out = append(out, l) } }
	return out, nil
}
func (r *listingRepo) AddFavorite(listingID, userID string) error {
	r.mu.Lock(); defer r.mu.Unlock()
	r.favorites[listingID] = append(r.favorites[listingID], userID)
	return nil
}
func (r *listingRepo) RemoveFavorite(listingID, userID string) error {
	r.mu.Lock(); defer r.mu.Unlock()
	list := r.favorites[listingID]
	for i, u := range list { if u == userID { r.favorites[listingID] = append(list[:i], list[i+1:]...); break } }
	return nil
}

// ---- Labour ----

type labourOrderRepo struct {
	mu      sync.RWMutex
	orders  []domain.LabourOrder
	quotes  []domain.LabourQuote
	assigns []domain.Assignment
}

func NewLabourOrderRepository() repository.LabourOrderRepository { return &labourOrderRepo{} }

func (r *labourOrderRepo) Create(o domain.LabourOrder) (domain.LabourOrder, error) {
	r.mu.Lock(); defer r.mu.Unlock(); r.orders = append(r.orders, o); return o, nil
}
func (r *labourOrderRepo) FindByID(id string) (domain.LabourOrder, error) {
	r.mu.RLock(); defer r.mu.RUnlock()
	for _, o := range r.orders { if o.ID == id { return o, nil } }
	return domain.LabourOrder{}, fmt.Errorf("order %s not found", id)
}
func (r *labourOrderRepo) ListByEmployer(uid string) ([]domain.LabourOrder, error) {
	r.mu.RLock(); defer r.mu.RUnlock()
	out := []domain.LabourOrder{}
	for _, o := range r.orders { if o.EmployerID == uid { out = append(out, o) } }
	return out, nil
}
func (r *labourOrderRepo) ListAll(offset, limit int) ([]domain.LabourOrder, int, error) {
	r.mu.RLock(); defer r.mu.RUnlock()
	t := len(r.orders)
	if offset > t { return nil, t, nil }; e := offset + limit; if e > t { e = t }
	return r.orders[offset:e], t, nil
}
func (r *labourOrderRepo) CreateQuote(q domain.LabourQuote) (domain.LabourQuote, error) {
	r.mu.Lock(); defer r.mu.Unlock(); r.quotes = append(r.quotes, q); return q, nil
}
func (r *labourOrderRepo) ListQuotes(orderID string) ([]domain.LabourQuote, error) {
	r.mu.RLock(); defer r.mu.RUnlock()
	out := []domain.LabourQuote{}
	for _, q := range r.quotes { if q.OrderID == orderID { out = append(out, q) } }
	return out, nil
}
func (r *labourOrderRepo) CreateAssignment(a domain.Assignment) (domain.Assignment, error) {
	r.mu.Lock(); defer r.mu.Unlock(); r.assigns = append(r.assigns, a); return a, nil
}

// ---- User ----

type memUserRepo struct {
	mu     sync.RWMutex
	items  []domain.User
	cipher *crypto.Cipher
}

func NewUserRepository(cipher *crypto.Cipher) repository.UserRepository {
	return &memUserRepo{cipher: cipher}
}

func (r *memUserRepo) decrypt(u *domain.User) {
	if r.cipher != nil && u.PhoneCipher != "" {
		if dec, err := r.cipher.Decrypt(u.PhoneCipher); err == nil {
			u.PhoneCipher = dec
		}
	}
}

func (r *memUserRepo) FindByOpenID(openid string) (domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, u := range r.items {
		if u.WechatOpenID == openid {
			r.decrypt(&u)
			return u, nil
		}
	}
	return domain.User{}, fmt.Errorf("user not found")
}
func (r *memUserRepo) Create(u domain.User) (domain.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items = append(r.items, u)
	return u, nil
}
func (r *memUserRepo) FindByID(id string) (domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, u := range r.items {
		if u.ID == id {
			r.decrypt(&u)
			return u, nil
		}
	}
	return domain.User{}, fmt.Errorf("user not found")
}
func (r *memUserRepo) All() ([]domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.User, len(r.items))
	copy(out, r.items)
	for i := range out {
		r.decrypt(&out[i])
	}
	return out, nil
}
func (r *memUserRepo) UpdateRole(id string, role domain.Role) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.items {
		if r.items[i].ID == id {
			r.items[i].Role = role
			return nil
		}
	}
	return fmt.Errorf("user not found")
}

type memRefreshRepo struct {
	mu    sync.RWMutex
	items []memRefreshEntry
}
type memRefreshEntry struct {
	userID, tokenHash string
	expiresAt         time.Time
	revoked           bool
}

func NewRefreshTokenRepository() repository.RefreshTokenRepository { return &memRefreshRepo{} }

func (r *memRefreshRepo) Store(userID, tokenHash string, expiresAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items = append(r.items, memRefreshEntry{userID, tokenHash, expiresAt, false})
	return nil
}
func (r *memRefreshRepo) Find(tokenHash string) (string, time.Time, bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, e := range r.items {
		if e.tokenHash == tokenHash {
			return e.userID, e.expiresAt, e.revoked, nil
		}
	}
	return "", time.Time{}, false, fmt.Errorf("token not found")
}
func (r *memRefreshRepo) Revoke(tokenHash string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.items {
		if r.items[i].tokenHash == tokenHash {
			r.items[i].revoked = true
			return nil
		}
	}
	return fmt.Errorf("token not found")
}

// ---- Contract ----

type contractRepo struct {
	mu    sync.RWMutex
	items []domain.Contract
}

func NewContractRepository() repository.ContractRepository {
	return &contractRepo{}
}

func (r *contractRepo) Create(v domain.Contract) (domain.Contract, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items = append(r.items, v)
	return v, nil
}

func (r *contractRepo) ListByEnterprise(eid string, offset, limit int) ([]domain.Contract, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	filtered := make([]domain.Contract, 0)
	for _, v := range r.items {
		if v.EnterpriseID == eid {
			filtered = append(filtered, v)
		}
	}
	page, total, _ := paginateSlice(filtered, offset, limit)
	return page, total, nil
}

func (r *contractRepo) ListAll(offset, limit int) ([]domain.Contract, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	page, total, _ := paginateSlice(r.items, offset, limit)
	return append([]domain.Contract(nil), page...), total, nil
}

func (r *contractRepo) FindByID(id string) (domain.Contract, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, v := range r.items {
		if v.ID == id {
			return v, nil
		}
	}
	return domain.Contract{}, fmt.Errorf("contract %s not found", id)
}

func (r *contractRepo) UpdateStatus(id string, status domain.ContractStatus) (domain.Contract, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, v := range r.items {
		if v.ID == id {
			r.items[i].Status = status
			r.items[i].UpdatedAt = time.Now()
			r.items[i].Version++
			return r.items[i], nil
		}
	}
	return domain.Contract{}, fmt.Errorf("contract %s not found", id)
}

// ---- DemandBid ----

type bidRepo struct {
	mu    sync.RWMutex
	items []domain.DemandBid
}

func NewBidRepository() repository.BidRepository {
	return &bidRepo{}
}

func (r *bidRepo) Create(b domain.DemandBid) (domain.DemandBid, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items = append(r.items, b)
	return b, nil
}

func (r *bidRepo) FindByID(id string) (domain.DemandBid, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, b := range r.items {
		if b.ID == id {
			return b, nil
		}
	}
	return domain.DemandBid{}, fmt.Errorf("bid %s not found", id)
}

func (r *bidRepo) ListByDemand(demandID string) ([]domain.DemandBid, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.DemandBid, 0)
	for _, b := range r.items {
		if b.DemandID == demandID {
			out = append(out, b)
		}
	}
	return out, nil
}

func (r *bidRepo) ListByBidder(bidderID string) ([]domain.DemandBid, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.DemandBid, 0)
	for _, b := range r.items {
		if b.BidderID == bidderID {
			out = append(out, b)
		}
	}
	return out, nil
}

func (r *bidRepo) UpdateStatus(id string, status string) (domain.DemandBid, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, b := range r.items {
		if b.ID == id {
			r.items[i].Status = status
			r.items[i].UpdatedAt = time.Now()
			r.items[i].Version++
			return r.items[i], nil
		}
	}
	return domain.DemandBid{}, fmt.Errorf("bid %s not found", id)
}

// ---- Certificate ----

type certRepo struct {
	mu    sync.RWMutex
	items []domain.Certificate
}

func NewCertificateRepository() repository.CertificateRepository { return &certRepo{} }

func (r *certRepo) Create(c domain.Certificate) (domain.Certificate, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items = append(r.items, c)
	return c, nil
}
func (r *certRepo) FindByID(id string) (domain.Certificate, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, c := range r.items {
		if c.ID == id {
			return c, nil
		}
	}
	return domain.Certificate{}, fmt.Errorf("certificate %s not found", id)
}
func (r *certRepo) ListByUser(userID string) ([]domain.Certificate, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.Certificate, 0)
	for _, c := range r.items {
		if c.UserID == userID {
			out = append(out, c)
		}
	}
	return out, nil
}
func (r *certRepo) UpdateStatus(id string, status string) (domain.Certificate, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, c := range r.items {
		if c.ID == id {
			r.items[i].Status = status
			r.items[i].UpdatedAt = time.Now()
			return r.items[i], nil
		}
	}
	return domain.Certificate{}, fmt.Errorf("certificate %s not found", id)
}
func (r *certRepo) ListAll() ([]domain.Certificate, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]domain.Certificate(nil), r.items...), nil
}

// ---- Course ----

type courseRepo struct {
	mu    sync.RWMutex
	items []domain.TrainingCourse
}

func NewCourseRepository() repository.CourseRepository { return &courseRepo{} }

func (r *courseRepo) Create(c domain.TrainingCourse) (domain.TrainingCourse, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items = append(r.items, c)
	return c, nil
}
func (r *courseRepo) List() ([]domain.TrainingCourse, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]domain.TrainingCourse(nil), r.items...), nil
}

// ---- Instructor ----

type instructorRepo struct {
	mu    sync.RWMutex
	items []domain.Instructor
}

func NewInstructorRepository() repository.InstructorRepository { return &instructorRepo{} }

func (r *instructorRepo) Create(i domain.Instructor) (domain.Instructor, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items = append(r.items, i)
	return i, nil
}
func (r *instructorRepo) FindByID(id string) (domain.Instructor, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, i := range r.items {
		if i.ID == id {
			return i, nil
		}
	}
	return domain.Instructor{}, fmt.Errorf("instructor %s not found", id)
}
func (r *instructorRepo) List() ([]domain.Instructor, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]domain.Instructor(nil), r.items...), nil
}
func (r *instructorRepo) UpdateStatus(id string, status string) (domain.Instructor, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, ins := range r.items {
		if ins.ID == id {
			r.items[i].Status = status
			r.items[i].UpdatedAt = time.Now()
			return r.items[i], nil
		}
	}
	return domain.Instructor{}, fmt.Errorf("instructor %s not found", id)
}

// ---- Pilot ----

type pilotRepo struct {
	mu     sync.RWMutex
	items  []domain.CertifiedPilot
	cipher *crypto.Cipher
}

func NewPilotRepository(cipher *crypto.Cipher) repository.PilotRepository { return &pilotRepo{cipher: cipher} }

func (r *pilotRepo) encryptInPlace(p *domain.CertifiedPilot) {
	if r.cipher != nil && p.IDCard != "" {
		enc, err := r.cipher.Encrypt(p.IDCard)
		if err == nil {
			p.IDCard = enc
		}
	}
}
func (r *pilotRepo) decryptInPlace(p *domain.CertifiedPilot) {
	if r.cipher != nil && p.IDCard != "" {
		dec, err := r.cipher.Decrypt(p.IDCard)
		if err == nil {
			p.IDCard = dec
		}
	}
}
func (r *pilotRepo) Create(p domain.CertifiedPilot) (domain.CertifiedPilot, error) {
	r.encryptInPlace(&p)
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items = append(r.items, p)
	r.decryptInPlace(&p)
	return p, nil
}
func (r *pilotRepo) FindByID(id string) (domain.CertifiedPilot, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, p := range r.items {
		if p.ID == id {
			r.decryptInPlace(&p)
			return p, nil
		}
	}
	return domain.CertifiedPilot{}, fmt.Errorf("pilot %s not found", id)
}
func (r *pilotRepo) List() ([]domain.CertifiedPilot, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.CertifiedPilot, len(r.items))
	copy(out, r.items)
	for i := range out {
		r.decryptInPlace(&out[i])
	}
	return out, nil
}
func (r *pilotRepo) UpdateStatus(id string, status string) (domain.CertifiedPilot, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, p := range r.items {
		if p.ID == id {
			r.items[i].Status = status
			r.items[i].UpdatedAt = time.Now()
			result := r.items[i]
			r.decryptInPlace(&result)
			return result, nil
		}
	}
	return domain.CertifiedPilot{}, fmt.Errorf("pilot %s not found", id)
}

// ---- Product ----

type prodRepo struct {
	mu    sync.RWMutex
	items []domain.DroneProduct
}

func NewProductRepository() repository.ProductRepository { return &prodRepo{} }

func (r *prodRepo) Create(p domain.DroneProduct) (domain.DroneProduct, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items = append(r.items, p)
	return p, nil
}
func (r *prodRepo) List(prodType string) ([]domain.DroneProduct, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if prodType == "" {
		return append([]domain.DroneProduct(nil), r.items...), nil
	}
	out := make([]domain.DroneProduct, 0)
	for _, p := range r.items {
		if string(p.ProdType) == prodType {
			out = append(out, p)
		}
	}
	return out, nil
}

// ---- Repair ----

type repairRepo struct {
	mu    sync.RWMutex
	items []domain.RepairOrder
}

func NewRepairRepository() repository.RepairRepository { return &repairRepo{} }

func (r *repairRepo) Create(ro domain.RepairOrder) (domain.RepairOrder, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items = append(r.items, ro)
	return ro, nil
}
func (r *repairRepo) ListByUser(userID string) ([]domain.RepairOrder, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.RepairOrder, 0)
	for _, ro := range r.items {
		if ro.CustomerID == userID {
			out = append(out, ro)
		}
	}
	return out, nil
}

// ---- Insurance Policy ----

type policyRepo struct {
	mu    sync.RWMutex
	items []domain.InsurancePolicy
}

func NewPolicyRepository() repository.PolicyRepository { return &policyRepo{} }

func (r *policyRepo) Create(p domain.InsurancePolicy) (domain.InsurancePolicy, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items = append(r.items, p)
	return p, nil
}
func (r *policyRepo) ListByUser(userID string) ([]domain.InsurancePolicy, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.InsurancePolicy, 0)
	for _, p := range r.items {
		if p.UserID == userID {
			out = append(out, p)
		}
	}
	return out, nil
}

// ---- Inspection ----

type inspectRepo struct {
	mu    sync.RWMutex
	items []domain.AnnualInspection
}

func NewInspectionRepository() repository.InspectionRepository { return &inspectRepo{} }

func (r *inspectRepo) Create(i domain.AnnualInspection) (domain.AnnualInspection, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items = append(r.items, i)
	return i, nil
}
func (r *inspectRepo) ListByUser(userID string) ([]domain.AnnualInspection, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.AnnualInspection, 0)
	for _, i := range r.items {
		if i.UserID == userID {
			out = append(out, i)
		}
	}
	return out, nil
}
func (r *inspectRepo) ListAll() ([]domain.AnnualInspection, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]domain.AnnualInspection(nil), r.items...), nil
}

// ---- Loan ----

type loanRepo struct {
	mu    sync.RWMutex
	items []domain.LoanApplication
}

func NewLoanRepository() repository.LoanRepository { return &loanRepo{} }

func (r *loanRepo) Create(l domain.LoanApplication) (domain.LoanApplication, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items = append(r.items, l)
	return l, nil
}
func (r *loanRepo) ListByUser(userID string) ([]domain.LoanApplication, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.LoanApplication, 0)
	for _, l := range r.items {
		if l.UserID == userID {
			out = append(out, l)
		}
	}
	return out, nil
}

// ---- Message ----

type msgRepo struct {
	mu    sync.RWMutex
	items []domain.Message
}

func NewMessageRepository() repository.MessageRepository { return &msgRepo{} }

func (r *msgRepo) Create(m domain.Message) (domain.Message, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items = append(r.items, m)
	return m, nil
}
func (r *msgRepo) ListByUser(userID string, unreadOnly bool) ([]domain.Message, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.Message, 0)
	for _, m := range r.items {
		if m.ReceiverID != userID {
			continue
		}
		if unreadOnly && m.IsRead {
			continue
		}
		out = append(out, m)
	}
	return out, nil
}
func (r *msgRepo) MarkRead(id string) (domain.Message, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, m := range r.items {
		if m.ID == id {
			r.items[i].IsRead = true
			return r.items[i], nil
		}
	}
	return domain.Message{}, fmt.Errorf("message %s not found", id)
}
func (r *msgRepo) UnreadCount(userID string) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	n := 0
	for _, m := range r.items {
		if m.ReceiverID == userID && !m.IsRead {
			n++
		}
	}
	return n, nil
}

// ---- Article ----

type articleRepo struct {
	mu    sync.RWMutex
	items []domain.Article
}

func NewArticleRepository() repository.ArticleRepository { return &articleRepo{} }

func (r *articleRepo) Create(a domain.Article) (domain.Article, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items = append(r.items, a)
	return a, nil
}
func (r *articleRepo) FindByID(id string) (domain.Article, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, a := range r.items {
		if a.ID == id {
			return a, nil
		}
	}
	return domain.Article{}, fmt.Errorf("article %s not found", id)
}
func (r *articleRepo) Update(a domain.Article) (domain.Article, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, v := range r.items {
		if v.ID == a.ID {
			r.items[i] = a
			return a, nil
		}
	}
	return domain.Article{}, fmt.Errorf("article %s not found", a.ID)
}
func (r *articleRepo) ListByCategory(category string, offset, limit int) ([]domain.Article, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	filtered := make([]domain.Article, 0)
	for _, a := range r.items {
		if category != "" && a.Category != category {
			continue
		}
		filtered = append(filtered, a)
	}
	page, total, _ := paginateSlice(filtered, offset, limit)
	return page, total, nil
}

// ---- Review ----

type reviewRepo struct {
	mu    sync.RWMutex
	items []domain.Review
}

func NewReviewRepository() repository.ReviewRepository { return &reviewRepo{} }

func (r *reviewRepo) Create(rv domain.Review) (domain.Review, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items = append(r.items, rv)
	return rv, nil
}
func (r *reviewRepo) ListByTarget(targetType, targetID string) ([]domain.Review, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.Review, 0)
	for _, rv := range r.items {
		if rv.TargetType == targetType && rv.TargetID == targetID {
			out = append(out, rv)
		}
	}
	return out, nil
}
func (r *reviewRepo) ListAll(status string, offset, limit int) ([]domain.Review, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	filtered := make([]domain.Review, 0)
	for _, rv := range r.items {
		if status == "" || rv.Status == status {
			filtered = append(filtered, rv)
		}
	}
	page, total, _ := paginateSlice(filtered, offset, limit)
	return page, total, nil
}
func (r *reviewRepo) UpdateStatus(id string, status string) (domain.Review, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, rv := range r.items {
		if rv.ID == id {
			r.items[i].Status = status
			return r.items[i], nil
		}
	}
	return domain.Review{}, fmt.Errorf("review %s not found", id)
}
func (r *reviewRepo) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, rv := range r.items {
		if rv.ID == id {
			r.items = append(r.items[:i], r.items[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("review %s not found", id)
}

// ---- Venue ----

type venueRepo struct {
	mu       sync.RWMutex
	venues   []domain.Venue
	bookings []domain.VenueBooking
}

func NewVenueRepository() repository.VenueRepository { return &venueRepo{} }

func (r *venueRepo) Create(v domain.Venue) (domain.Venue, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.venues = append(r.venues, v)
	return v, nil
}
func (r *venueRepo) List(venueType string) ([]domain.Venue, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.Venue, 0)
	for _, v := range r.venues {
		if venueType == "" || v.VenueType == venueType {
			out = append(out, v)
		}
	}
	return out, nil
}
func (r *venueRepo) FindByID(id string) (domain.Venue, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, v := range r.venues {
		if v.ID == id {
			return v, nil
		}
	}
	return domain.Venue{}, fmt.Errorf("venue %s not found", id)
}
func (r *venueRepo) CreateBooking(b domain.VenueBooking) (domain.VenueBooking, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.bookings = append(r.bookings, b)
	return b, nil
}
func (r *venueRepo) ListBookings(venueID string) ([]domain.VenueBooking, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.VenueBooking, 0)
	for _, b := range r.bookings {
		if b.VenueID == venueID {
			out = append(out, b)
		}
	}
	return out, nil
}

// ---- Enrollment ----

type enrollRepo struct {
	mu    sync.RWMutex
	items []domain.Enrollment
}

func NewEnrollmentRepository() repository.EnrollmentRepository { return &enrollRepo{} }

func (r *enrollRepo) Create(e domain.Enrollment) (domain.Enrollment, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items = append(r.items, e)
	return e, nil
}
func (r *enrollRepo) ListByCourse(courseID string) ([]domain.Enrollment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.Enrollment, 0)
	for _, e := range r.items {
		if e.CourseID == courseID {
			out = append(out, e)
		}
	}
	return out, nil
}
func (r *enrollRepo) FindByUserAndCourse(userID, courseID string) (domain.Enrollment, bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, e := range r.items {
		if e.UserID == userID && e.CourseID == courseID {
			return e, true, nil
		}
	}
	return domain.Enrollment{}, false, nil
}

// ---- TradeOrder ----

type tradeOrderRepo struct {
	mu    sync.RWMutex
	items []domain.TradeOrder
}

func NewTradeOrderRepository() repository.TradeOrderRepository { return &tradeOrderRepo{} }

func (r *tradeOrderRepo) Create(o domain.TradeOrder) (domain.TradeOrder, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items = append(r.items, o)
	return o, nil
}
func (r *tradeOrderRepo) FindByID(id string) (domain.TradeOrder, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, o := range r.items {
		if o.ID == id {
			return o, nil
		}
	}
	return domain.TradeOrder{}, fmt.Errorf("order %s not found", id)
}
func (r *tradeOrderRepo) UpdateStatus(id string, status string) (domain.TradeOrder, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, o := range r.items {
		if o.ID == id {
			r.items[i].Status = status
			r.items[i].UpdatedAt = time.Now()
			r.items[i].Version++
			return r.items[i], nil
		}
	}
	return domain.TradeOrder{}, fmt.Errorf("order %s not found", id)
}
func (r *tradeOrderRepo) ListByUser(userID string) ([]domain.TradeOrder, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.TradeOrder, 0)
	for _, o := range r.items {
		if o.BuyerID == userID || o.SellerID == userID {
			out = append(out, o)
		}
	}
	return out, nil
}

// ---- Escrow ----

type escrowRepo struct {
	mu    sync.RWMutex
	accts map[string]*domain.EscrowAccount
	txs   []domain.EscrowTransaction
}

func NewEscrowRepository() repository.EscrowRepository {
	return &escrowRepo{accts: make(map[string]*domain.EscrowAccount)}
}

func (r *escrowRepo) GetAccount(userID string) (domain.EscrowAccount, error) {
	r.mu.RLock()
	acct, ok := r.accts[userID]
	r.mu.RUnlock()
	if ok {
		return *acct, nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.accts[userID]; !ok {
		r.accts[userID] = &domain.EscrowAccount{UserID: userID}
	}
	return *r.accts[userID], nil
}
func (r *escrowRepo) UpsertAccount(a domain.EscrowAccount) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.accts[a.UserID] = &a
	return nil
}
func (r *escrowRepo) CreateTransaction(tx domain.EscrowTransaction) (domain.EscrowTransaction, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.txs = append(r.txs, tx)
	return tx, nil
}
func (r *escrowRepo) ListTransactions(userID string) ([]domain.EscrowTransaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.EscrowTransaction, 0)
	for _, tx := range r.txs {
		if tx.FromUser == userID || tx.ToUser == userID {
			out = append(out, tx)
		}
	}
	return out, nil
}
