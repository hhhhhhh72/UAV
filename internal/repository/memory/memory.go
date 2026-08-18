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
		} else {
			// 解密失败（密钥变更/数据损坏）绝不回传密文——置空而非泄露加密串。
			d.Contact = ""
		}
	}
}

func (r *demandRepo) Create(ctx context.Context, d domain.Demand) (domain.Demand, error) {
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

func (r *demandRepo) Update(ctx context.Context, d domain.Demand) (domain.Demand, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.items {
		if r.items[i].ID == d.ID {
			// 乐观锁：与 PG 实现对齐——旧版本号不匹配视为并发修改
			if r.items[i].Version != d.Version {
				return domain.Demand{}, fmt.Errorf("demand %s 已被他人修改，请刷新后重试", d.ID)
			}
			if r.cipher != nil && d.Contact != "" {
				enc, err := r.cipher.Encrypt(d.Contact)
				if err != nil {
					return domain.Demand{}, fmt.Errorf("encrypt contact: %w", err)
				}
				d.Contact = enc
			}
			d.Version++
			r.items[i] = d
			return d, nil
		}
	}
	return domain.Demand{}, fmt.Errorf("demand %s not found", d.ID)
}

func (r *demandRepo) FindByID(ctx context.Context, id string) (domain.Demand, error) {
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

func (r *demandRepo) List(ctx context.Context, f repository.DemandFilter) ([]domain.Demand, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.Demand, 0)
	for _, d := range r.items {
		// 公开语义：仅已发布
		if d.Status != domain.DemandPublished {
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

// ListAll 管理端全量（含待审核/已驳回等），按 filter 过滤状态/地区/类型。
func (r *demandRepo) ListAll(ctx context.Context, f repository.DemandFilter) ([]domain.Demand, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.Demand, 0)
	for _, d := range r.items {
		if f.Status != "" && f.Status != "all" && string(d.Status) != f.Status {
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

func (r *demandRepo) Search(ctx context.Context, q string) ([]domain.Demand, error) {
	all, _ := r.List(ctx, repository.DemandFilter{})
	q = strings.ToLower(q)
	out := []domain.Demand{}
	for _, d := range all {
		if strings.Contains(strings.ToLower(d.Title+d.PublisherName+d.Description), q) {
			out = append(out, d)
		}
	}
	return out, nil
}

// ListByPublisher 返回某发布者的全部需求（全状态），供"我的"页统计/查询。
func (r *demandRepo) ListByPublisher(ctx context.Context, publisherID string) ([]domain.Demand, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.Demand, 0)
	for _, d := range r.items {
		if d.PublisherID != publisherID {
			continue
		}
		r.decrypt(&d)
		out = append(out, d)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}

func (r *demandRepo) SetStatus(ctx context.Context, id string, status domain.DemandStatus) (domain.Demand, error) {
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

func (r *demandRepo) CompareAndSetStatus(ctx context.Context, id string, oldStatus, newStatus domain.DemandStatus) (bool, domain.Demand, error) {
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

func (r *demandRepo) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.items {
		if r.items[i].ID == id {
			r.items = append(r.items[:i], r.items[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("demand %s not found", id)
}

// ---- Enterprise ----

type enterpriseRepo struct {
	mu     sync.RWMutex
	items  []domain.Enterprise
	docs   []domain.EnterpriseDocument
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

func (r *enterpriseRepo) Pending(ctx context.Context) ([]domain.Enterprise, error) {
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

func (r *enterpriseRepo) Create(ctx context.Context, e domain.Enterprise) (domain.Enterprise, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := r.encrypt(&e); err != nil {
		return domain.Enterprise{}, err
	}
	r.items = append(r.items, e)
	return e, nil
}

func (r *enterpriseRepo) Update(ctx context.Context, id string, e domain.Enterprise) (domain.Enterprise, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.items {
		if r.items[i].ID == id {
			// 乐观锁：与 PG 实现对齐——旧版本号不匹配视为并发修改
			if r.items[i].Version != e.Version {
				return domain.Enterprise{}, fmt.Errorf("enterprise %s 已被他人修改，请刷新后重试", id)
			}
			if err := r.encrypt(&e); err != nil {
				return domain.Enterprise{}, err
			}
			e.Version++
			r.items[i] = e
			return e, nil
		}
	}
	return domain.Enterprise{}, fmt.Errorf("enterprise %s not found", id)
}

func (r *enterpriseRepo) FindByID(ctx context.Context, id string) (domain.Enterprise, error) {
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

func (r *enterpriseRepo) FindByOwner(ctx context.Context, userID string) ([]domain.Enterprise, error) {
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

func (r *enterpriseRepo) ListByStatus(ctx context.Context, status string, offset, limit int) ([]domain.Enterprise, int, error) {
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

func (r *enterpriseRepo) Delete(ctx context.Context, id string) error {
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

func (r *enterpriseRepo) Search(ctx context.Context, q string) ([]domain.Enterprise, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
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

func (r *enterpriseRepo) AddDocument(ctx context.Context, d domain.EnterpriseDocument) (domain.EnterpriseDocument, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.docs = append(r.docs, d)
	return d, nil
}

func (r *enterpriseRepo) ListDocuments(ctx context.Context, enterpriseID string) ([]domain.EnterpriseDocument, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := []domain.EnterpriseDocument{}
	for _, d := range r.docs {
		if d.EnterpriseID == enterpriseID {
			out = append(out, d)
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

func (r *employmentRepo) Create(ctx context.Context, v domain.EmploymentRequest) (domain.EmploymentRequest, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items = append(r.items, v)
	return v, nil
}

func (r *employmentRepo) ListByEnterprise(ctx context.Context, eid string, offset, limit int) ([]domain.EmploymentRequest, int, error) {
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

func (r *employmentRepo) ListAll(ctx context.Context, offset, limit int) ([]domain.EmploymentRequest, int, error) {
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

func (r *jobRepo) Create(ctx context.Context, j domain.Job) (domain.Job, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items = append(r.items, j)
	return j, nil
}
func (r *jobRepo) Update(ctx context.Context, id string, j domain.Job) (domain.Job, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.items {
		if r.items[i].ID == id {
			// 乐观锁：与 PG 实现对齐——旧版本号不匹配视为并发修改
			if r.items[i].Version != j.Version {
				return domain.Job{}, fmt.Errorf("job %s 已被他人修改，请刷新后重试", id)
			}
			j.Version++
			r.items[i] = j
			return j, nil
		}
	}
	return domain.Job{}, fmt.Errorf("job %s not found", id)
}
func (r *jobRepo) FindByID(ctx context.Context, id string) (domain.Job, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, j := range r.items {
		if j.ID == id {
			return j, nil
		}
	}
	return domain.Job{}, fmt.Errorf("job %s not found", id)
}
func (r *jobRepo) ListByEnterprise(ctx context.Context, eid string) ([]domain.Job, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := []domain.Job{}
	for _, j := range r.items {
		if j.EnterpriseID == eid {
			out = append(out, j)
		}
	}
	return out, nil
}
func (r *jobRepo) ListPublished(ctx context.Context, offset, limit int) ([]domain.Job, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	filtered := []domain.Job{}
	for _, j := range r.items {
		if j.Status == domain.JobPublished {
			filtered = append(filtered, j)
		}
	}
	total := len(filtered)
	if offset > total {
		return nil, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return filtered[offset:end], total, nil
}
func (r *jobRepo) ListAll(ctx context.Context, offset, limit int) ([]domain.Job, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := append([]domain.Job(nil), r.items...)
	total := len(items)
	if offset > total {
		return nil, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return items[offset:end], total, nil
}

func (r *jobRepo) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.items {
		if r.items[i].ID == id {
			r.items = append(r.items[:i], r.items[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("job %s not found", id)
}

type resumeRepo struct {
	mu    sync.RWMutex
	items []domain.Resume
}

func NewResumeRepository() repository.ResumeRepository { return &resumeRepo{} }

func (r *resumeRepo) Create(ctx context.Context, v domain.Resume) (domain.Resume, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items = append(r.items, v)
	return v, nil
}
func (r *resumeRepo) Update(ctx context.Context, id string, v domain.Resume) (domain.Resume, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.items {
		if r.items[i].ID == id {
			// 乐观锁：与 PG 实现对齐——旧版本号不匹配视为并发修改
			if r.items[i].Version != v.Version {
				return domain.Resume{}, fmt.Errorf("resume %s 已被他人修改，请刷新后重试", id)
			}
			v.Version++
			r.items[i] = v
			return v, nil
		}
	}
	return domain.Resume{}, fmt.Errorf("resume %s not found", id)
}
func (r *resumeRepo) FindByID(ctx context.Context, id string) (domain.Resume, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, v := range r.items {
		if v.ID == id {
			return v, nil
		}
	}
	return domain.Resume{}, fmt.Errorf("resume %s not found", id)
}
func (r *resumeRepo) ListByUser(ctx context.Context, userID string) ([]domain.Resume, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := []domain.Resume{}
	for _, v := range r.items {
		if v.UserID == userID {
			out = append(out, v)
		}
	}
	return out, nil
}
func (r *resumeRepo) ListAll(ctx context.Context, offset, limit int) ([]domain.Resume, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return paginateSlice(r.items, offset, limit)
}

type applicationRepo struct {
	mu    sync.RWMutex
	items []domain.JobApplication
}

func NewJobApplicationRepository() repository.JobApplicationRepository { return &applicationRepo{} }

func (r *applicationRepo) Create(ctx context.Context, a domain.JobApplication) (domain.JobApplication, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items = append(r.items, a)
	return a, nil
}
func (r *applicationRepo) FindByID(ctx context.Context, id string) (domain.JobApplication, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, a := range r.items {
		if a.ID == id {
			return a, nil
		}
	}
	return domain.JobApplication{}, fmt.Errorf("application %s not found", id)
}
func (r *applicationRepo) UpdateStatus(ctx context.Context, id string, status domain.AppStatus) (domain.JobApplication, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.items {
		if r.items[i].ID == id {
			r.items[i].Status = status
			return r.items[i], nil
		}
	}
	return domain.JobApplication{}, fmt.Errorf("application %s not found", id)
}
func (r *applicationRepo) ListByJob(ctx context.Context, jobID string) ([]domain.JobApplication, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := []domain.JobApplication{}
	for _, a := range r.items {
		if a.JobID == jobID {
			out = append(out, a)
		}
	}
	return out, nil
}
func (r *applicationRepo) ListByApplicant(ctx context.Context, userID string) ([]domain.JobApplication, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := []domain.JobApplication{}
	for _, a := range r.items {
		if a.ApplicantID == userID {
			out = append(out, a)
		}
	}
	return out, nil
}

// ---- Post ----

type postRepo struct {
	mu    sync.RWMutex
	items []domain.Post
}

func NewPostRepository() repository.PostRepository { return &postRepo{} }

func (r *postRepo) Create(ctx context.Context, p domain.Post) (domain.Post, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items = append(r.items, p)
	return p, nil
}
func (r *postRepo) Update(ctx context.Context, id string, p domain.Post) (domain.Post, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.items {
		if r.items[i].ID == id {
			r.items[i] = p
			return p, nil
		}
	}
	return domain.Post{}, fmt.Errorf("post %s not found", id)
}
func (r *postRepo) FindByID(ctx context.Context, id string) (domain.Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, p := range r.items {
		if p.ID == id {
			return p, nil
		}
	}
	return domain.Post{}, fmt.Errorf("post %s not found", id)
}
func (r *postRepo) ListPublished(ctx context.Context, offset, limit int) ([]domain.Post, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	f := []domain.Post{}
	for _, p := range r.items {
		if p.Status == "published" {
			f = append(f, p)
		}
	}
	t := len(f)
	if offset > t {
		return nil, t, nil
	}
	e := offset + limit
	if e > t {
		e = t
	}
	return f[offset:e], t, nil
}
func (r *postRepo) ListByAuthor(ctx context.Context, uid string) ([]domain.Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := []domain.Post{}
	for _, p := range r.items {
		if p.AuthorID == uid {
			out = append(out, p)
		}
	}
	return out, nil
}

type commentRepo struct {
	mu    sync.RWMutex
	items []domain.Comment
}

func NewCommentRepository() repository.CommentRepository { return &commentRepo{} }

func (r *commentRepo) Create(ctx context.Context, c domain.Comment) (domain.Comment, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items = append(r.items, c)
	return c, nil
}
func (r *commentRepo) ListByPost(ctx context.Context, postID string) ([]domain.Comment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := []domain.Comment{}
	for _, c := range r.items {
		if c.PostID == postID {
			out = append(out, c)
		}
	}
	return out, nil
}

type reportRepo struct {
	mu    sync.RWMutex
	items []domain.Report
}

func NewReportRepository() repository.ReportRepository { return &reportRepo{} }

func (r *reportRepo) Create(ctx context.Context, rp domain.Report) (domain.Report, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items = append(r.items, rp)
	return rp, nil
}
func (r *reportRepo) ListPending(ctx context.Context, offset, limit int) ([]domain.Report, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	f := []domain.Report{}
	for _, rp := range r.items {
		if rp.Status == "pending" {
			f = append(f, rp)
		}
	}
	t := len(f)
	if offset > t {
		return nil, t, nil
	}
	e := offset + limit
	if e > t {
		e = t
	}
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

func (r *listingRepo) Create(ctx context.Context, l domain.Listing) (domain.Listing, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items = append(r.items, l)
	return l, nil
}
func (r *listingRepo) Update(ctx context.Context, id string, l domain.Listing) (domain.Listing, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.items {
		if r.items[i].ID == id {
			r.items[i] = l
			return l, nil
		}
	}
	return domain.Listing{}, fmt.Errorf("listing %s not found", id)
}
func (r *listingRepo) FindByID(ctx context.Context, id string) (domain.Listing, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, l := range r.items {
		if l.ID == id {
			return l, nil
		}
	}
	return domain.Listing{}, fmt.Errorf("listing %s not found", id)
}
func (r *listingRepo) ListByStatus(ctx context.Context, status string, offset, limit int) ([]domain.Listing, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	f := []domain.Listing{}
	for _, l := range r.items {
		if l.Status == status || status == "" {
			f = append(f, l)
		}
	}
	t := len(f)
	if offset > t {
		return nil, t, nil
	}
	e := offset + limit
	if e > t {
		e = t
	}
	return f[offset:e], t, nil
}
func (r *listingRepo) ListBySeller(ctx context.Context, uid string) ([]domain.Listing, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := []domain.Listing{}
	for _, l := range r.items {
		if l.SellerID == uid {
			out = append(out, l)
		}
	}
	return out, nil
}
func (r *listingRepo) AddFavorite(ctx context.Context, listingID, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.favorites[listingID] = append(r.favorites[listingID], userID)
	return nil
}
func (r *listingRepo) RemoveFavorite(ctx context.Context, listingID, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	list := r.favorites[listingID]
	for i, u := range list {
		if u == userID {
			r.favorites[listingID] = append(list[:i], list[i+1:]...)
			break
		}
	}
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

func (r *labourOrderRepo) Create(ctx context.Context, o domain.LabourOrder) (domain.LabourOrder, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.orders = append(r.orders, o)
	return o, nil
}
func (r *labourOrderRepo) FindByID(ctx context.Context, id string) (domain.LabourOrder, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, o := range r.orders {
		if o.ID == id {
			return o, nil
		}
	}
	return domain.LabourOrder{}, fmt.Errorf("order %s not found", id)
}
func (r *labourOrderRepo) ListByEmployer(ctx context.Context, uid string) ([]domain.LabourOrder, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := []domain.LabourOrder{}
	for _, o := range r.orders {
		if o.EmployerID == uid {
			out = append(out, o)
		}
	}
	return out, nil
}
func (r *labourOrderRepo) ListAll(ctx context.Context, offset, limit int) ([]domain.LabourOrder, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t := len(r.orders)
	if offset > t {
		return nil, t, nil
	}
	e := offset + limit
	if e > t {
		e = t
	}
	return r.orders[offset:e], t, nil
}
func (r *labourOrderRepo) CreateQuote(ctx context.Context, q domain.LabourQuote) (domain.LabourQuote, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.quotes = append(r.quotes, q)
	return q, nil
}
func (r *labourOrderRepo) ListQuotes(ctx context.Context, orderID string) ([]domain.LabourQuote, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := []domain.LabourQuote{}
	for _, q := range r.quotes {
		if q.OrderID == orderID {
			out = append(out, q)
		}
	}
	return out, nil
}
func (r *labourOrderRepo) CreateAssignment(ctx context.Context, a domain.Assignment) (domain.Assignment, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.assigns = append(r.assigns, a)
	return a, nil
}

func (r *labourOrderRepo) ListAssignmentsByOrder(ctx context.Context, orderID string) ([]domain.Assignment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := []domain.Assignment{}
	for _, a := range r.assigns {
		if a.OrderID == orderID {
			out = append(out, a)
		}
	}
	return out, nil
}

func (r *labourOrderRepo) ListAssignmentsByWorker(ctx context.Context, workerID string) ([]domain.Assignment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := []domain.Assignment{}
	for _, a := range r.assigns {
		if a.WorkerID == workerID {
			out = append(out, a)
		}
	}
	return out, nil
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

func (r *memUserRepo) FindByOpenID(ctx context.Context, openid string) (domain.User, error) {
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
func (r *memUserRepo) Create(ctx context.Context, u domain.User) (domain.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items = append(r.items, u)
	return u, nil
}
func (r *memUserRepo) FindByID(ctx context.Context, id string) (domain.User, error) {
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
func (r *memUserRepo) All(ctx context.Context) ([]domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.User, len(r.items))
	copy(out, r.items)
	for i := range out {
		r.decrypt(&out[i])
	}
	return out, nil
}
func (r *memUserRepo) UpdateRole(ctx context.Context, id string, role domain.Role) error {
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

func (r *memUserRepo) UpdateAvatar(ctx context.Context, userID, avatarURL string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.items {
		if r.items[i].ID == userID {
			r.items[i].AvatarURL = avatarURL
			return nil
		}
	}
	return fmt.Errorf("user not found")
}

func (r *memUserRepo) UpdateName(ctx context.Context, userID, name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.items {
		if r.items[i].ID == userID {
			r.items[i].Name = name
			return nil
		}
	}
	return fmt.Errorf("user not found")
}

func (r *memUserRepo) UpdateProfile(ctx context.Context, id string, p domain.UserProfile) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.items {
		if r.items[i].ID == id {
			r.items[i].Gender = p.Gender
			r.items[i].Birthday = p.Birthday
			r.items[i].Region = p.Region
			r.items[i].Bio = p.Bio
			if p.Phone != "" {
				enc := p.Phone
				if r.cipher != nil {
					if c, err := r.cipher.Encrypt(p.Phone); err == nil {
						enc = c
					}
				}
				r.items[i].PhoneCipher = enc
			}
			return nil
		}
	}
	return fmt.Errorf("user not found")
}

func (r *memUserRepo) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.items {
		if r.items[i].ID == id {
			r.items = append(r.items[:i], r.items[i+1:]...)
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

func (r *memRefreshRepo) Store(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items = append(r.items, memRefreshEntry{userID, tokenHash, expiresAt, false})
	return nil
}
func (r *memRefreshRepo) Find(ctx context.Context, tokenHash string) (string, time.Time, bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, e := range r.items {
		if e.tokenHash == tokenHash {
			return e.userID, e.expiresAt, e.revoked, nil
		}
	}
	return "", time.Time{}, false, fmt.Errorf("token not found")
}
func (r *memRefreshRepo) Revoke(ctx context.Context, tokenHash string) error {
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

// ---- Contract Template ----

type contractTplRepo struct {
	mu    sync.RWMutex
	items []domain.ContractTemplate
}

// NewContractTemplateRepository 初始化时即内置默认模板（对齐 PG 种子迁移 000062）。
func NewContractTemplateRepository() repository.ContractTemplateRepository {
	return &contractTplRepo{items: append([]domain.ContractTemplate(nil), domain.DefaultContractTemplates...)}
}

func (r *contractTplRepo) List(ctx context.Context) ([]domain.ContractTemplate, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]domain.ContractTemplate(nil), r.items...), nil
}

func (r *contractTplRepo) Create(ctx context.Context, t domain.ContractTemplate) (domain.ContractTemplate, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items = append(r.items, t)
	return t, nil
}

func (r *contractRepo) Create(ctx context.Context, v domain.Contract) (domain.Contract, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items = append(r.items, v)
	return v, nil
}

func (r *contractRepo) ListByEnterprise(ctx context.Context, eid string, offset, limit int) ([]domain.Contract, int, error) {
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

func (r *contractRepo) ListAll(ctx context.Context, offset, limit int) ([]domain.Contract, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	page, total, _ := paginateSlice(r.items, offset, limit)
	return append([]domain.Contract(nil), page...), total, nil
}

func (r *contractRepo) FindByID(ctx context.Context, id string) (domain.Contract, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, v := range r.items {
		if v.ID == id {
			return v, nil
		}
	}
	return domain.Contract{}, fmt.Errorf("contract %s not found", id)
}

func (r *contractRepo) UpdateStatus(ctx context.Context, id string, status domain.ContractStatus) (domain.Contract, error) {
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

func (r *bidRepo) Create(ctx context.Context, b domain.DemandBid) (domain.DemandBid, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items = append(r.items, b)
	return b, nil
}

func (r *bidRepo) FindByID(ctx context.Context, id string) (domain.DemandBid, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, b := range r.items {
		if b.ID == id {
			return b, nil
		}
	}
	return domain.DemandBid{}, fmt.Errorf("bid %s not found", id)
}

func (r *bidRepo) ListByDemand(ctx context.Context, demandID string) ([]domain.DemandBid, error) {
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

func (r *bidRepo) ListByBidder(ctx context.Context, bidderID string) ([]domain.DemandBid, error) {
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

func (r *bidRepo) UpdateStatus(ctx context.Context, id string, status string) (domain.DemandBid, error) {
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

// ---- DemandIntent ----

type intentRepo struct {
	mu    sync.RWMutex
	items []domain.DemandIntent
}

func NewIntentRepository() repository.IntentRepository {
	return &intentRepo{}
}

func (r *intentRepo) Create(ctx context.Context, it domain.DemandIntent) (domain.DemandIntent, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	// P1 修复：锁内去重——同一 (demand, intentor) 只允许一条 pending 意向，
	// 与 PG 部分唯一索引（WHERE status='pending'）语义对齐；已确认的旧意向不阻塞再登记。
	for _, e := range r.items {
		if e.DemandID == it.DemandID && e.IntentorID == it.IntentorID && e.Status == "pending" {
			return domain.DemandIntent{}, fmt.Errorf("duplicate pending intent for demand %s by %s", it.DemandID, it.IntentorID)
		}
	}
	r.items = append(r.items, it)
	return it, nil
}

func (r *intentRepo) ListByDemand(ctx context.Context, demandID string) ([]domain.DemandIntent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.DemandIntent, 0)
	for _, it := range r.items {
		if it.DemandID == demandID {
			out = append(out, it)
		}
	}
	return out, nil
}

func (r *intentRepo) ListByIntentor(ctx context.Context, intentorID string) ([]domain.DemandIntent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.DemandIntent, 0)
	for _, it := range r.items {
		if it.IntentorID == intentorID {
			out = append(out, it)
		}
	}
	return out, nil
}

func (r *intentRepo) UpdateStatus(ctx context.Context, id string, status string) (domain.DemandIntent, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, it := range r.items {
		if it.ID == id {
			// B 批加固：CAS 语义——仅 pending 可流转，与 PG 实现对齐
			if it.Status != "pending" {
				return domain.DemandIntent{}, fmt.Errorf("意向不存在或已处理")
			}
			r.items[i].Status = status
			r.items[i].UpdatedAt = time.Now()
			r.items[i].Version++
			return r.items[i], nil
		}
	}
	return domain.DemandIntent{}, fmt.Errorf("intent %s not found", id)
}

// ---- WorkOrder (接单派单闭环) ----

type workOrderRepo struct {
	mu    sync.RWMutex
	items []domain.WorkOrder
}

func NewWorkOrderRepository() repository.WorkOrderRepository {
	return &workOrderRepo{}
}

func (r *workOrderRepo) Create(ctx context.Context, wo domain.WorkOrder) (domain.WorkOrder, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items = append(r.items, wo)
	return wo, nil
}

func (r *workOrderRepo) FindByID(ctx context.Context, id string) (domain.WorkOrder, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, wo := range r.items {
		if wo.ID == id {
			return wo, nil
		}
	}
	return domain.WorkOrder{}, fmt.Errorf("work order %s not found", id)
}

func (r *workOrderRepo) ListByPublisher(ctx context.Context, publisherID string) ([]domain.WorkOrder, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.WorkOrder, 0)
	for _, wo := range r.items {
		if wo.PublisherID == publisherID {
			out = append(out, wo)
		}
	}
	return out, nil
}

func (r *workOrderRepo) ListByWorker(ctx context.Context, workerID string) ([]domain.WorkOrder, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.WorkOrder, 0)
	for _, wo := range r.items {
		if wo.WorkerID == workerID {
			out = append(out, wo)
		}
	}
	return out, nil
}

func (r *workOrderRepo) UpdateStatus(ctx context.Context, id string, status domain.WorkOrderStatus) (domain.WorkOrder, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, wo := range r.items {
		if wo.ID == id {
			r.items[i].Status = status
			r.items[i].UpdatedAt = time.Now()
			return r.items[i], nil
		}
	}
	return domain.WorkOrder{}, fmt.Errorf("work order %s not found", id)
}

func (r *workOrderRepo) UpdatePhotos(ctx context.Context, id string, photos []string) (domain.WorkOrder, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, wo := range r.items {
		if wo.ID == id {
			r.items[i].ResultPhotos = photos
			r.items[i].UpdatedAt = time.Now()
			return r.items[i], nil
		}
	}
	return domain.WorkOrder{}, fmt.Errorf("work order %s not found", id)
}

func (r *workOrderRepo) UpdateRework(ctx context.Context, id string, note string) (domain.WorkOrder, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, wo := range r.items {
		if wo.ID == id {
			r.items[i].ReworkNote = note
			r.items[i].UpdatedAt = time.Now()
			return r.items[i], nil
		}
	}
	return domain.WorkOrder{}, fmt.Errorf("work order %s not found", id)
}

func (r *workOrderRepo) UpdateCancel(ctx context.Context, id string, reason string) (domain.WorkOrder, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, wo := range r.items {
		if wo.ID == id {
			r.items[i].CancelReason = reason
			r.items[i].UpdatedAt = time.Now()
			return r.items[i], nil
		}
	}
	return domain.WorkOrder{}, fmt.Errorf("work order %s not found", id)
}

// ---- Certificate ----

type certRepo struct {
	mu    sync.RWMutex
	items []domain.Certificate
}

func NewCertificateRepository() repository.CertificateRepository { return &certRepo{} }

func (r *certRepo) Create(ctx context.Context, c domain.Certificate) (domain.Certificate, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items = append(r.items, c)
	return c, nil
}
func (r *certRepo) FindByID(ctx context.Context, id string) (domain.Certificate, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, c := range r.items {
		if c.ID == id {
			return c, nil
		}
	}
	return domain.Certificate{}, fmt.Errorf("certificate %s not found", id)
}
func (r *certRepo) ListByUser(ctx context.Context, userID string) ([]domain.Certificate, error) {
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
func (r *certRepo) UpdateStatus(ctx context.Context, id string, status string) (domain.Certificate, error) {
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
func (r *certRepo) ListAll(ctx context.Context) ([]domain.Certificate, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]domain.Certificate(nil), r.items...), nil
}

func (r *certRepo) Update(ctx context.Context, c domain.Certificate) (domain.Certificate, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.items {
		if r.items[i].ID == c.ID {
			r.items[i] = c
			return c, nil
		}
	}
	return domain.Certificate{}, fmt.Errorf("cert %s not found", c.ID)
}

func (r *certRepo) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.items {
		if r.items[i].ID == id {
			r.items = append(r.items[:i], r.items[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("cert %s not found", id)
}

// ---- Course ----

type courseRepo struct {
	mu    sync.RWMutex
	items []domain.TrainingCourse
}

func NewCourseRepository() repository.CourseRepository { return &courseRepo{} }

func (r *courseRepo) Create(ctx context.Context, c domain.TrainingCourse) (domain.TrainingCourse, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items = append(r.items, c)
	return c, nil
}
func (r *courseRepo) List(ctx context.Context) ([]domain.TrainingCourse, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]domain.TrainingCourse(nil), r.items...), nil
}
func (r *courseRepo) FindByID(ctx context.Context, id string) (domain.TrainingCourse, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, c := range r.items {
		if c.ID == id {
			return c, nil
		}
	}
	return domain.TrainingCourse{}, fmt.Errorf("course %s not found", id)
}
func (r *courseRepo) Update(ctx context.Context, c domain.TrainingCourse) (domain.TrainingCourse, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.items {
		if r.items[i].ID == c.ID {
			r.items[i] = c
			return c, nil
		}
	}
	return domain.TrainingCourse{}, fmt.Errorf("course %s not found", c.ID)
}
func (r *courseRepo) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.items {
		if r.items[i].ID == id {
			r.items = append(r.items[:i], r.items[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("course %s not found", id)
}

// ---- Instructor ----

type instructorRepo struct {
	mu    sync.RWMutex
	items []domain.Instructor
}

func NewInstructorRepository() repository.InstructorRepository { return &instructorRepo{} }

func (r *instructorRepo) Create(ctx context.Context, i domain.Instructor) (domain.Instructor, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items = append(r.items, i)
	return i, nil
}
func (r *instructorRepo) FindByID(ctx context.Context, id string) (domain.Instructor, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, i := range r.items {
		if i.ID == id {
			return i, nil
		}
	}
	return domain.Instructor{}, fmt.Errorf("instructor %s not found", id)
}
func (r *instructorRepo) List(ctx context.Context) ([]domain.Instructor, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]domain.Instructor(nil), r.items...), nil
}
func (r *instructorRepo) UpdateStatus(ctx context.Context, id string, status string) (domain.Instructor, error) {
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

func NewPilotRepository(cipher *crypto.Cipher) repository.PilotRepository {
	return &pilotRepo{cipher: cipher}
}

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
func (r *pilotRepo) Create(ctx context.Context, p domain.CertifiedPilot) (domain.CertifiedPilot, error) {
	r.encryptInPlace(&p)
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items = append(r.items, p)
	r.decryptInPlace(&p)
	return p, nil
}
func (r *pilotRepo) FindByID(ctx context.Context, id string) (domain.CertifiedPilot, error) {
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
func (r *pilotRepo) List(ctx context.Context) ([]domain.CertifiedPilot, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.CertifiedPilot, len(r.items))
	copy(out, r.items)
	for i := range out {
		r.decryptInPlace(&out[i])
	}
	return out, nil
}
func (r *pilotRepo) Update(ctx context.Context, p domain.CertifiedPilot) (domain.CertifiedPilot, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, v := range r.items {
		if v.ID == p.ID {
			p.UpdatedAt = time.Now()
			r.items[i] = p
			result := p
			r.decryptInPlace(&result)
			return result, nil
		}
	}
	return domain.CertifiedPilot{}, fmt.Errorf("pilot %s not found", p.ID)
}

func (r *pilotRepo) UpdateStatus(ctx context.Context, id string, status string) (domain.CertifiedPilot, error) {
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

func (r *prodRepo) Create(ctx context.Context, p domain.DroneProduct) (domain.DroneProduct, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items = append(r.items, p)
	return p, nil
}
func (r *prodRepo) FindByID(ctx context.Context, id string) (domain.DroneProduct, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, p := range r.items {
		if p.ID == id {
			return p, nil
		}
	}
	return domain.DroneProduct{}, fmt.Errorf("product %s not found", id)
}

func (r *prodRepo) Update(ctx context.Context, p domain.DroneProduct) (domain.DroneProduct, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.items {
		if r.items[i].ID == p.ID {
			r.items[i] = p
			return p, nil
		}
	}
	return domain.DroneProduct{}, fmt.Errorf("product %s not found", p.ID)
}

func (r *prodRepo) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.items {
		if r.items[i].ID == id {
			r.items = append(r.items[:i], r.items[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("product %s not found", id)
}

func (r *prodRepo) IncrementViews(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.items {
		if r.items[i].ID == id {
			r.items[i].Views++
			return nil
		}
	}
	return fmt.Errorf("product %s not found", id)
}

func (r *prodRepo) List(ctx context.Context, prodType string) ([]domain.DroneProduct, error) {
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

// ---- Service Listings ----

type slrRepo struct {
	mu    sync.RWMutex
	items []domain.ServiceListing
}

func NewServiceListingRepository() repository.ServiceListingRepository { return &slrRepo{} }

func (r *slrRepo) Create(ctx context.Context, sl domain.ServiceListing) (domain.ServiceListing, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items = append(r.items, sl)
	return sl, nil
}

func (r *slrRepo) FindByID(ctx context.Context, id string) (domain.ServiceListing, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, sl := range r.items {
		if sl.ID == id {
			return sl, nil
		}
	}
	return domain.ServiceListing{}, fmt.Errorf("service listing %s not found", id)
}

func (r *slrRepo) List(ctx context.Context) ([]domain.ServiceListing, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]domain.ServiceListing(nil), r.items...), nil
}

func (r *slrRepo) Update(ctx context.Context, sl domain.ServiceListing) (domain.ServiceListing, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.items {
		if r.items[i].ID == sl.ID {
			r.items[i] = sl
			return sl, nil
		}
	}
	return domain.ServiceListing{}, fmt.Errorf("service listing %s not found", sl.ID)
}

func (r *slrRepo) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.items {
		if r.items[i].ID == id {
			r.items = append(r.items[:i], r.items[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("service listing %s not found", id)
}

// ---- Repair ----

type repairRepo struct {
	mu    sync.RWMutex
	items []domain.RepairOrder
}

func NewRepairRepository() repository.RepairRepository { return &repairRepo{} }

func (r *repairRepo) Create(ctx context.Context, ro domain.RepairOrder) (domain.RepairOrder, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items = append(r.items, ro)
	return ro, nil
}
func (r *repairRepo) ListByUser(ctx context.Context, userID string) ([]domain.RepairOrder, error) {
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
func (r *repairRepo) ListAll(ctx context.Context, offset, limit int) ([]domain.RepairOrder, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return paginateSlice(r.items, offset, limit)
}

// ---- Insurance Policy ----

type policyRepo struct {
	mu    sync.RWMutex
	items []domain.InsurancePolicy
}

func NewPolicyRepository() repository.PolicyRepository { return &policyRepo{} }

func (r *policyRepo) Create(ctx context.Context, p domain.InsurancePolicy) (domain.InsurancePolicy, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items = append(r.items, p)
	return p, nil
}
func (r *policyRepo) ListByUser(ctx context.Context, userID string) ([]domain.InsurancePolicy, error) {
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
func (r *policyRepo) ListAll(ctx context.Context, offset, limit int) ([]domain.InsurancePolicy, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return paginateSlice(r.items, offset, limit)
}

// ---- Inspection ----

type inspectRepo struct {
	mu    sync.RWMutex
	items []domain.AnnualInspection
}

func NewInspectionRepository() repository.InspectionRepository { return &inspectRepo{} }

func (r *inspectRepo) Create(ctx context.Context, i domain.AnnualInspection) (domain.AnnualInspection, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items = append(r.items, i)
	return i, nil
}
func (r *inspectRepo) ListByUser(ctx context.Context, userID string) ([]domain.AnnualInspection, error) {
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
func (r *inspectRepo) ListAll(ctx context.Context) ([]domain.AnnualInspection, error) {
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

func (r *loanRepo) Create(ctx context.Context, l domain.LoanApplication) (domain.LoanApplication, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items = append(r.items, l)
	return l, nil
}
func (r *loanRepo) ListByUser(ctx context.Context, userID string) ([]domain.LoanApplication, error) {
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
func (r *loanRepo) ListAll(ctx context.Context, offset, limit int) ([]domain.LoanApplication, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return paginateSlice(r.items, offset, limit)
}

// ---- Message ----

type msgRepo struct {
	mu    sync.RWMutex
	items []domain.Message
}

func NewMessageRepository() repository.MessageRepository { return &msgRepo{} }

func (r *msgRepo) Create(ctx context.Context, m domain.Message) (domain.Message, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items = append(r.items, m)
	return m, nil
}
func (r *msgRepo) FindByID(ctx context.Context, id string) (domain.Message, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, m := range r.items {
		if m.ID == id {
			return m, nil
		}
	}
	return domain.Message{}, fmt.Errorf("message %s not found", id)
}
func (r *msgRepo) ListByUser(ctx context.Context, userID string, unreadOnly bool) ([]domain.Message, error) {
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
func (r *msgRepo) MarkRead(ctx context.Context, id string) (domain.Message, error) {
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
func (r *msgRepo) UnreadCount(ctx context.Context, userID string) (int, error) {
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

func (r *msgRepo) ListAll(ctx context.Context, offset, limit int) ([]domain.Message, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	total := len(r.items)
	if offset > total {
		return nil, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return append([]domain.Message(nil), r.items[offset:end]...), total, nil
}

func (r *msgRepo) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.items {
		if r.items[i].ID == id {
			r.items = append(r.items[:i], r.items[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("message %s not found", id)
}

// ---- Article ----

type articleRepo struct {
	mu    sync.RWMutex
	items []domain.Article
}

func NewArticleRepository() repository.ArticleRepository { return &articleRepo{} }

func (r *articleRepo) Create(ctx context.Context, a domain.Article) (domain.Article, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items = append(r.items, a)
	return a, nil
}
func (r *articleRepo) FindByID(ctx context.Context, id string) (domain.Article, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, a := range r.items {
		if a.ID == id {
			return a, nil
		}
	}
	return domain.Article{}, fmt.Errorf("article %s not found", id)
}
func (r *articleRepo) Update(ctx context.Context, a domain.Article) (domain.Article, error) {
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
func (r *articleRepo) ListByCategory(ctx context.Context, category string, offset, limit int) ([]domain.Article, int, error) {
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

func (r *reviewRepo) Create(ctx context.Context, rv domain.Review) (domain.Review, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items = append(r.items, rv)
	return rv, nil
}
func (r *reviewRepo) ListByTarget(ctx context.Context, targetType, targetID string) ([]domain.Review, error) {
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
func (r *reviewRepo) ListAll(ctx context.Context, status string, offset, limit int) ([]domain.Review, int, error) {
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
func (r *reviewRepo) UpdateStatus(ctx context.Context, id string, status string) (domain.Review, error) {
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
func (r *reviewRepo) Delete(ctx context.Context, id string) error {
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

func (r *venueRepo) Create(ctx context.Context, v domain.Venue) (domain.Venue, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.venues = append(r.venues, v)
	return v, nil
}
func (r *venueRepo) List(ctx context.Context, venueType string) ([]domain.Venue, error) {
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
func (r *venueRepo) FindByID(ctx context.Context, id string) (domain.Venue, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, v := range r.venues {
		if v.ID == id {
			return v, nil
		}
	}
	return domain.Venue{}, fmt.Errorf("venue %s not found", id)
}
func (r *venueRepo) CreateBooking(ctx context.Context, b domain.VenueBooking) (domain.VenueBooking, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.bookings = append(r.bookings, b)
	return b, nil
}
func (r *venueRepo) ListBookings(ctx context.Context, venueID string) ([]domain.VenueBooking, error) {
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

func (r *enrollRepo) Create(ctx context.Context, e domain.Enrollment) (domain.Enrollment, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items = append(r.items, e)
	return e, nil
}

func (r *enrollRepo) Update(ctx context.Context, e domain.Enrollment) (domain.Enrollment, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.items {
		if r.items[i].ID == e.ID {
			r.items[i] = e
			return e, nil
		}
	}
	return domain.Enrollment{}, fmt.Errorf("enrollment %s not found", e.ID)
}
func (r *enrollRepo) ListByCourse(ctx context.Context, courseID string) ([]domain.Enrollment, error) {
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
func (r *enrollRepo) ListAll(ctx context.Context, offset, limit int) ([]domain.Enrollment, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := append([]domain.Enrollment(nil), r.items...)
	total := len(items)
	if offset > total {
		return []domain.Enrollment{}, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return items[offset:end], total, nil
}
func (r *enrollRepo) FindByUserAndCourse(ctx context.Context, userID, courseID string) (domain.Enrollment, bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, e := range r.items {
		if e.UserID == userID && e.CourseID == courseID {
			return e, true, nil
		}
	}
	return domain.Enrollment{}, false, nil
}
func (r *enrollRepo) FindByID(ctx context.Context, id string) (domain.Enrollment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, e := range r.items {
		if e.ID == id {
			return e, nil
		}
	}
	return domain.Enrollment{}, fmt.Errorf("enrollment %s not found", id)
}

// ---- TradeOrder ----

type tradeOrderRepo struct {
	mu    sync.RWMutex
	items []domain.TradeOrder
}

func NewTradeOrderRepository() repository.TradeOrderRepository { return &tradeOrderRepo{} }

func (r *tradeOrderRepo) Create(ctx context.Context, o domain.TradeOrder) (domain.TradeOrder, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items = append(r.items, o)
	return o, nil
}
func (r *tradeOrderRepo) FindByID(ctx context.Context, id string) (domain.TradeOrder, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, o := range r.items {
		if o.ID == id {
			return o, nil
		}
	}
	return domain.TradeOrder{}, fmt.Errorf("order %s not found", id)
}
func (r *tradeOrderRepo) UpdateStatus(ctx context.Context, id string, status string) (domain.TradeOrder, error) {
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
func (r *tradeOrderRepo) UpdateAftersale(ctx context.Context, o domain.TradeOrder) (domain.TradeOrder, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.items {
		if r.items[i].ID == o.ID {
			r.items[i].Status = o.Status
			r.items[i].AftersaleType = o.AftersaleType
			r.items[i].AftersaleReason = o.AftersaleReason
			r.items[i].AftersaleDesc = o.AftersaleDesc
			r.items[i].AftersaleAmountFen = o.AftersaleAmountFen
			r.items[i].AftersaleStatus = o.AftersaleStatus
			r.items[i].AftersaleTime = o.AftersaleTime
			r.items[i].UpdatedAt = time.Now()
			r.items[i].Version++
			return r.items[i], nil
		}
	}
	return domain.TradeOrder{}, fmt.Errorf("order %s not found", o.ID)
}
func (r *tradeOrderRepo) ListByUser(ctx context.Context, userID string) ([]domain.TradeOrder, error) {
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
func (r *tradeOrderRepo) ListAll(ctx context.Context, offset, limit int) ([]domain.TradeOrder, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	total := len(r.items)
	if offset > total {
		return nil, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return append([]domain.TradeOrder(nil), r.items[offset:end]...), total, nil
}
func (r *tradeOrderRepo) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.items {
		if r.items[i].ID == id {
			r.items = append(r.items[:i], r.items[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("order %s not found", id)
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

func (r *escrowRepo) GetAccount(ctx context.Context, userID string) (domain.EscrowAccount, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	acct, ok := r.accts[userID]
	if !ok {
		acct = &domain.EscrowAccount{UserID: userID}
		r.accts[userID] = acct
	}
	return *acct, nil
}

// 原子资金操作（C6 修复）：校验、改账、记流水在同一临界区内完成，
// 消除旧接口读-改-写的并发丢更新窗口。
func (r *escrowRepo) Deposit(ctx context.Context, userID string, amountFen int64, tx domain.EscrowTransaction) (domain.EscrowTransaction, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	acct, ok := r.accts[userID]
	if !ok {
		acct = &domain.EscrowAccount{UserID: userID}
		r.accts[userID] = acct
	}
	acct.BalanceFen += amountFen
	acct.UpdatedAt = time.Now()
	r.txs = append(r.txs, tx)
	return tx, nil
}
func (r *escrowRepo) Freeze(ctx context.Context, userID string, amountFen int64, tx domain.EscrowTransaction) (domain.EscrowTransaction, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	acct, ok := r.accts[userID]
	if !ok {
		acct = &domain.EscrowAccount{UserID: userID}
		r.accts[userID] = acct
	}
	if acct.BalanceFen < amountFen {
		return domain.EscrowTransaction{}, repository.ErrInsufficientBalance
	}
	acct.BalanceFen -= amountFen
	acct.FrozenFen += amountFen
	acct.UpdatedAt = time.Now()
	r.txs = append(r.txs, tx)
	return tx, nil
}
func (r *escrowRepo) Release(ctx context.Context, fromUser, toUser string, amountFen int64, tx domain.EscrowTransaction) (domain.EscrowTransaction, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	from, ok := r.accts[fromUser]
	if !ok {
		from = &domain.EscrowAccount{UserID: fromUser}
		r.accts[fromUser] = from
	}
	if from.FrozenFen < amountFen {
		return domain.EscrowTransaction{}, repository.ErrInsufficientFrozenBalance
	}
	to, ok := r.accts[toUser]
	if !ok {
		to = &domain.EscrowAccount{UserID: toUser}
		r.accts[toUser] = to
	}
	from.FrozenFen -= amountFen
	from.UpdatedAt = time.Now()
	to.BalanceFen += amountFen
	to.UpdatedAt = time.Now()
	r.txs = append(r.txs, tx)
	return tx, nil
}
func (r *escrowRepo) Refund(ctx context.Context, userID string, amountFen int64, tx domain.EscrowTransaction) (domain.EscrowTransaction, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	acct, ok := r.accts[userID]
	if !ok {
		acct = &domain.EscrowAccount{UserID: userID}
		r.accts[userID] = acct
	}
	if acct.FrozenFen < amountFen {
		return domain.EscrowTransaction{}, repository.ErrInsufficientFrozenBalance
	}
	acct.FrozenFen -= amountFen
	acct.BalanceFen += amountFen
	acct.UpdatedAt = time.Now()
	r.txs = append(r.txs, tx)
	return tx, nil
}
func (r *escrowRepo) ListTransactions(ctx context.Context, userID string) ([]domain.EscrowTransaction, error) {
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

// ---- Upload（文件上传台账，配额统计用） ----

type uploadRepo struct {
	mu      sync.RWMutex
	records []domain.FileRecord
}

func NewUploadRepository() repository.UploadRepository {
	return &uploadRepo{}
}

func (r *uploadRepo) Create(ctx context.Context, rec domain.FileRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.records = append(r.records, rec)
	return nil
}

func (r *uploadRepo) FindByID(ctx context.Context, id string) (domain.FileRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, rec := range r.records {
		if rec.ID == id {
			return rec, nil
		}
	}
	return domain.FileRecord{}, fmt.Errorf("upload %s not found", id)
}

func (r *uploadRepo) SumBytesSince(ctx context.Context, ownerID string, since time.Time) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var sum int64
	for _, rec := range r.records {
		if rec.OwnerID == ownerID && !rec.CreatedAt.Before(since) {
			sum += rec.SizeBytes
		}
	}
	return sum, nil
}
