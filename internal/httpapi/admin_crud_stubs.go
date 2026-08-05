package httpapi

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"drone-platform/internal/domain"
)

func parsePagination(r *http.Request) (int, int) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return page, pageSize
}

// ============================================================
// Admin CRUD stub handlers — fill gaps for frontend useAdminApi()
// Handlers that already exist in other files are NOT redeclared.
// ============================================================

// adminListFilter 管理端列表通用过滤：keyword 匹配 kwField，status 精确匹配。
// 管理端数据量小，列表先全量拉取再做内存过滤，保证搜索/筛选真正生效。
func adminListFilter[T any](items []T, kw, status string, kwField func(T) string, statusField func(T) string) ([]T, int) {
	kw = strings.ToLower(strings.TrimSpace(kw))
	out := make([]T, 0, len(items))
	for _, it := range items {
		if status != "" && statusField(it) != status {
			continue
		}
		if kw != "" && !strings.Contains(strings.ToLower(kwField(it)), kw) {
			continue
		}
		out = append(out, it)
	}
	return out, len(out)
}

// adminSlicePage 对过滤后的全量列表按页码切片。
func adminSlicePage[T any](items []T, page, pageSize int) []T {
	offset := (page - 1) * pageSize
	if offset >= len(items) {
		return []T{}
	}
	end := offset + pageSize
	if end > len(items) {
		end = len(items)
	}
	return items[offset:end]
}

// ----- Enrollments (报名记录) -----
func (s *Server) listAdminEnrollments(w http.ResponseWriter, r *http.Request) {
	page, pageSize := parsePagination(r)
	all, _, err := s.enrollSvc.All(0, 10000)
	if err != nil {
		fail(w, r, 500, fmt.Errorf("list enrollments: %w", err))
		return
	}
	filtered, total := adminListFilter(all, r.URL.Query().Get("keyword"), r.URL.Query().Get("status"),
		func(e domain.Enrollment) string { return e.Name + e.Phone },
		func(e domain.Enrollment) string { return e.Status })
	paginatedRespond(w, r, adminSlicePage(filtered, page, pageSize), total)
}

// ----- Test site bookings (场地预约记录) -----
func (s *Server) listAdminBookings(w http.ResponseWriter, r *http.Request) {
	page, pageSize := parsePagination(r)
	all, _, err := s.testSiteSvc.ListAllBookings(0, 10000)
	if err != nil {
		fail(w, r, 500, fmt.Errorf("list bookings: %w", err))
		return
	}
	filtered, total := adminListFilter(all, r.URL.Query().Get("keyword"), r.URL.Query().Get("status"),
		func(b domain.TestSiteBooking) string { return b.ContactName + b.Purpose },
		func(b domain.TestSiteBooking) string { return b.Status })
	paginatedRespond(w, r, adminSlicePage(filtered, page, pageSize), total)
}

// ----- Orders (trade_orders) -----
func (s *Server) listAdminOrders(w http.ResponseWriter, r *http.Request) {
	page, pageSize := parsePagination(r)
	all, _, err := s.tradeSvc.ListAll(0, 10000)
	if err != nil {
		fail(w, r, 500, err)
		return
	}
	// 状态 + 关键词过滤（复用管理端通用过滤）
	filtered, total := adminListFilter(all, r.URL.Query().Get("keyword"), r.URL.Query().Get("status"),
		func(o domain.TradeOrder) string { return o.ID },
		func(o domain.TradeOrder) string { return o.Status })
	// 日期范围过滤（created_at）
	if sd := r.URL.Query().Get("start_date"); sd != "" {
		if t, err := time.Parse("2006-01-02", sd); err == nil {
			out := filtered[:0]
			for _, o := range filtered {
				if !o.CreatedAt.Before(t) {
					out = append(out, o)
				}
			}
			filtered = out
		}
	}
	if ed := r.URL.Query().Get("end_date"); ed != "" {
		if t, err := time.Parse("2006-01-02", ed); err == nil {
			end := t.Add(24 * time.Hour) // 含当日
			out := filtered[:0]
			for _, o := range filtered {
				if o.CreatedAt.Before(end) {
					out = append(out, o)
				}
			}
			filtered = out
		}
	}
	total = len(filtered)
	paginatedRespond(w, r, adminSlicePage(filtered, page, pageSize), total)
}

// ----- Reviews -----
func (s *Server) listAdminReviews(w http.ResponseWriter, r *http.Request) {
	page, pageSize := parsePagination(r)
	items, total, err := s.reviewSvc.ListAll("", (page-1)*pageSize, pageSize)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	paginatedRespond(w, r, items, total)
}

// ----- Case entries -----
func (s *Server) listAdminCaseEntries(w http.ResponseWriter, r *http.Request) {
	page, pageSize := parsePagination(r)
	category := r.URL.Query().Get("category")
	items, total, err := s.caseSvc.List(category, page, pageSize)
	if err != nil {
		fail(w, r, 500, err)
		return
	}
	paginatedRespond(w, r, items, total)
}

// ----- Experts admin wrapper -----
func (s *Server) listAdminExperts(w http.ResponseWriter, r *http.Request) {
	s.listExperts(w, r)
}

// ----- Competitions -----
func (s *Server) listAdminCompetitions(w http.ResponseWriter, r *http.Request) {
	page, pageSize := parsePagination(r)
	items, total, err := s.competitionSvc.List(page, pageSize)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	paginatedRespond(w, r, items, total)
}
func (s *Server) adminCreateCompetition(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Title       string `json:"title"`
		Category    string `json:"category"`
		Description string `json:"description"`
		Location    string `json:"location"`
		Sponsor     string `json:"sponsor"`
		StartDate   string `json:"start_date"`
		EndDate     string `json:"end_date"`
		MaxTeams    int    `json:"max_teams"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	c, err := s.competitionSvc.Create(in.Title, in.Category, in.Description, in.Location, in.Sponsor, domain.ParseTime(in.StartDate), domain.ParseTime(in.EndDate), in.MaxTeams)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusCreated, c)
}
func (s *Server) updateCompetition(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var in struct {
		Title       string `json:"title"`
		Category    string `json:"category"`
		Description string `json:"description"`
		Location    string `json:"location"`
		Sponsor     string `json:"sponsor"`
		StartDate   string `json:"start_date"`
		EndDate     string `json:"end_date"`
		MaxTeams    int    `json:"max_teams"`
		Status      string `json:"status"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	c, err := s.competitionSvc.Update(id, in.Title, in.Category, in.Description, in.Location, in.Sponsor, in.Status, domain.ParseTime(in.StartDate), domain.ParseTime(in.EndDate), in.MaxTeams)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, c)
}
func (s *Server) deleteCompetition(w http.ResponseWriter, r *http.Request) {
	if err := s.competitionSvc.Delete(r.PathValue("id")); err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, map[string]string{"deleted": "ok"})
}

// --- Training Courses (missing update/delete) ---
func (s *Server) adminCreateCourse(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Title       string `json:"title"`
		CertType    string `json:"cert_type"`
		Category    string `json:"category"`
		Description string `json:"description"`
		Location    string `json:"location"`
		StartDate   string `json:"start_date"`
		EndDate     string `json:"end_date"`
		MaxStudents int    `json:"max_students"`
		Capacity    int    `json:"capacity"`
		PriceFen    int64  `json:"price_fen"`
		Status      string `json:"status"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	ct := in.CertType
	if ct == "" {
		ct = in.Category
	}
	ms := in.MaxStudents
	if ms == 0 {
		ms = in.Capacity
	}
	sd := domain.ParseTime(in.StartDate)
	ed := domain.ParseTime(in.EndDate)
	c, err := s.trainingSvc.CreateCourse(domain.Actor{Role: domain.RolePlatformAdmin}, in.Title, domain.CertType(ct), in.Description, in.Location, sd, ed, ms, in.PriceFen)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusCreated, c)
}
func (s *Server) updateCourse(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var in struct {
		Title       string `json:"title"`
		CertType    string `json:"cert_type"`
		Category    string `json:"category"`
		Description string `json:"description"`
		Location    string `json:"location"`
		StartDate   string `json:"start_date"`
		EndDate     string `json:"end_date"`
		MaxStudents int    `json:"max_students"`
		Capacity    int    `json:"capacity"`
		PriceFen    int64  `json:"price_fen"`
		Status      string `json:"status"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	ct := in.CertType
	if ct == "" {
		ct = in.Category
	}
	ms := in.MaxStudents
	if ms == 0 {
		ms = in.Capacity
	}
	c, err := s.trainingSvc.UpdateCourse(id, in.Title, ct, in.Description, in.Location, domain.ParseTime(in.StartDate), domain.ParseTime(in.EndDate), ms, in.PriceFen, in.Status)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, c)
}
func (s *Server) deleteCourse(w http.ResponseWriter, r *http.Request) {
	if err := s.trainingSvc.DeleteCourse(r.PathValue("id")); err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, map[string]string{"deleted": "ok"})
}

// --- Certificates (missing admin list/update/delete) ---
func (s *Server) listAdminCerts(w http.ResponseWriter, r *http.Request) {
	page, pageSize := parsePagination(r)
	certs, err := s.trainingSvc.ListAllCertificates()
	if err != nil {
		fail(w, r, 500, fmt.Errorf("list certs: %w", err))
		return
	}
	filtered, total := adminListFilter(certs, r.URL.Query().Get("keyword"), r.URL.Query().Get("status"),
		func(c domain.Certificate) string { return c.CertNumber },
		func(c domain.Certificate) string { return c.Status })
	paginatedRespond(w, r, adminSlicePage(filtered, page, pageSize), total)
}
func (s *Server) updateCertificate(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var in struct {
		CertType   string    `json:"cert_type"`
		CertNumber string    `json:"cert_number"`
		Level      string    `json:"level"`
		IssuerOrg  string    `json:"issuer_org"`
		IssueDate  time.Time `json:"issue_date"`
		ExpireDate time.Time `json:"expire_date"`
		Status     string    `json:"status"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	c, err := s.trainingSvc.UpdateCertificate(id, in.CertType, in.CertNumber, in.Level, in.IssuerOrg, in.Status, in.IssueDate, in.ExpireDate)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, c)
}
func (s *Server) deleteCertificate(w http.ResponseWriter, r *http.Request) {
	if err := s.trainingSvc.DeleteCertificate(r.PathValue("id")); err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, map[string]string{"deleted": "ok"})
}

func (s *Server) adminCreateCertificate(w http.ResponseWriter, r *http.Request) {
	var in struct {
		CertType   string    `json:"cert_type"`
		CertNumber string    `json:"cert_number"`
		Level      string    `json:"level"`
		IssuerOrg  string    `json:"issuer_org"`
		IssueDate  time.Time `json:"issue_date"`
		ExpireDate time.Time `json:"expire_date"`
		Status     string    `json:"status"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, fmt.Errorf("auth required"))
		return
	}
	c, err := s.trainingSvc.AddCertificate(domain.Actor{ID: a.ID, Role: domain.RolePlatformAdmin}, domain.CertType(in.CertType), in.CertNumber, in.Level, in.IssuerOrg, in.IssueDate, in.ExpireDate)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusCreated, c)
}

// --- Jobs (missing admin list/update/delete) ---
func (s *Server) listAdminJobs(w http.ResponseWriter, r *http.Request) {
	page, pageSize := parsePagination(r)
	all, _, err := s.jobSvc.ListAllJobs(0, 10000)
	if err != nil {
		fail(w, r, 500, fmt.Errorf("list jobs: %w", err))
		return
	}
	filtered, total := adminListFilter(all, r.URL.Query().Get("keyword"), r.URL.Query().Get("status"),
		func(j domain.Job) string { return j.Title },
		func(j domain.Job) string { return string(j.Status) })
	paginatedRespond(w, r, adminSlicePage(filtered, page, pageSize), total)
}
func (s *Server) updateJob(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var in struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Location    string `json:"location"`
		JobType     string `json:"job_type"`
		SalaryFen   int64  `json:"salary_fen"`
		Status      string `json:"status"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	j, err := s.jobSvc.UpdateJob(id, in.Title, in.Description, in.Location, in.JobType, in.SalaryFen, in.Status)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, j)
}
func (s *Server) deleteJob(w http.ResponseWriter, r *http.Request) {
	if err := s.jobSvc.DeleteJob(r.PathValue("id")); err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, map[string]string{"deleted": "ok"})
}

func (s *Server) adminCreateJob(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Location    string `json:"location"`
		SalaryFen   int64  `json:"salary_fen"`
		JobType     string `json:"job_type"`
		Status      string `json:"status"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	j, err := s.jobSvc.CreateJob(domain.Actor{ID: "admin", Role: domain.RolePlatformAdmin}, in.Title, in.Description, in.Location, in.SalaryFen)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	if in.Status == "published" {
		var err error
		// 与 CreateJob 使用同一 actor（ID:"admin"），否则 owner 校验失败
		j, err = s.jobSvc.PublishJob(domain.Actor{ID: "admin", Role: domain.RolePlatformAdmin}, j.ID)
		if err != nil {
			fail(w, r, http.StatusInternalServerError, fmt.Errorf("publish job: %w", err))
			return
		}
	}
	respond(w, r, http.StatusCreated, j)
}

// --- Colleges (missing admin list/update/delete) ---
func (s *Server) listAdminColleges(w http.ResponseWriter, r *http.Request) {
	page, pageSize := parsePagination(r)
	all, err := s.collegeSvc.List("")
	if err != nil {
		fail(w, r, 500, fmt.Errorf("list colleges: %w", err))
		return
	}
	filtered, total := adminListFilter(all, r.URL.Query().Get("keyword"), r.URL.Query().Get("status"),
		func(c domain.College) string { return c.Name },
		func(c domain.College) string { return c.Status })
	paginatedRespond(w, r, adminSlicePage(filtered, page, pageSize), total)
}
func (s *Server) updateCollege(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var in struct {
		Name        string   `json:"name"`
		Region      string   `json:"region"`
		Description string   `json:"description"`
		LogoURL     string   `json:"logo_url"`
		Status      string   `json:"status"`
		CoopType    string   `json:"coop_type"`
		Majors      []string `json:"majors"`
		Facilities  []string `json:"facilities"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	c, err := s.collegeSvc.Update(id, in.Name, in.Region, in.Description, in.LogoURL, in.Status, in.CoopType, in.Majors, in.Facilities)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, c)
}
func (s *Server) deleteCollege(w http.ResponseWriter, r *http.Request) {
	if err := s.collegeSvc.Delete(r.PathValue("id")); err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, map[string]string{"deleted": "ok"})
}

func (s *Server) adminCreateCollege(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Name        string   `json:"name"`
		Region      string   `json:"region"`
		Description string   `json:"description"`
		LogoURL     string   `json:"logo_url"`
		CoopType    string   `json:"coop_type"`
		Majors      []string `json:"majors"`
		Facilities  []string `json:"facilities"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	c, err := s.collegeSvc.Create(in.Name, in.Region, in.Description, in.LogoURL, in.CoopType, in.Majors, in.Facilities)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusCreated, c)
}

// --- Study Tours (missing list/create/update/delete) ---
func (s *Server) listAdminStudy(w http.ResponseWriter, r *http.Request) {
	page, pageSize := parsePagination(r)
	items, err := s.studyTourRepo.List()
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	filtered, total := adminListFilter(items, r.URL.Query().Get("keyword"), r.URL.Query().Get("status"),
		func(t domain.StudyTour) string { return t.Title },
		func(t domain.StudyTour) string { return t.Status })
	paginatedRespond(w, r, convStudy(adminSlicePage(filtered, page, pageSize)), total)
}

func convStudy(items []domain.StudyTour) []map[string]any {
	out := make([]map[string]any, len(items))
	for i, it := range items {
		out[i] = map[string]any{"id": it.ID, "title": it.Title, "destination": it.Destination, "duration": it.Duration, "capacity": it.Capacity, "status": it.Status, "created_at": it.CreatedAt, "updated_at": it.UpdatedAt}
	}
	return out
}

func (s *Server) createStudyTour(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Title       string `json:"title"`
		Destination string `json:"destination"`
		Duration    string `json:"duration"`
		Capacity    int    `json:"capacity"`
		Status      string `json:"status"`
		Description string `json:"description"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	st := domain.StudyTour{ID: fmt.Sprintf("study-%d", time.Now().UnixNano()), Title: in.Title, Destination: in.Destination, Duration: in.Duration, Capacity: in.Capacity, Status: in.Status, Description: in.Description}
	sr, err := s.studyTourRepo.Create(st)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusCreated, sr)
}
func (s *Server) updateStudyTour(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var in struct {
		Title       string `json:"title"`
		Destination string `json:"destination"`
		Duration    string `json:"duration"`
		Capacity    int    `json:"capacity"`
		Status      string `json:"status"`
		Description string `json:"description"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	st, err := s.studyTourRepo.FindByID(id)
	if err != nil {
		fail(w, r, http.StatusNotFound, err)
		return
	}
	st.Title = in.Title
	st.Destination = in.Destination
	st.Duration = in.Duration
	st.Capacity = in.Capacity
	st.Status = in.Status
	st.Description = in.Description
	sr, err := s.studyTourRepo.Update(st)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, sr)
}
func (s *Server) deleteStudyTour(w http.ResponseWriter, r *http.Request) {
	if err := s.studyTourRepo.Delete(r.PathValue("id")); err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, map[string]string{"deleted": "ok"})
}

// --- Test Sites (missing admin list/update/delete) ---
func (s *Server) listAdminTestSites(w http.ResponseWriter, r *http.Request) {
	page, pageSize := parsePagination(r)
	all, err := s.testSiteSvc.List(r.URL.Query().Get("site_type"))
	if err != nil {
		fail(w, r, 500, fmt.Errorf("list test sites: %w", err))
		return
	}
	filtered, total := adminListFilter(all, r.URL.Query().Get("keyword"), r.URL.Query().Get("status"),
		func(t domain.TestSite) string { return t.Name },
		func(t domain.TestSite) string { return t.Status })
	paginatedRespond(w, r, adminSlicePage(filtered, page, pageSize), total)
}
func (s *Server) updateTestSite(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var in struct {
		Name        string   `json:"name"`
		SiteType    string   `json:"site_type"`
		Location    string   `json:"location"`
		BookingRule string   `json:"booking_rule"`
		Status      string   `json:"status"`
		PriceFen    int64    `json:"price_fen"`
		Facilities  []string `json:"facilities"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, 400, err)
		return
	}
	site, err := s.testSiteSvc.UpdateSite(id, in.Name, in.SiteType, in.Location, in.BookingRule, in.Status, in.PriceFen, in.Facilities)
	if err != nil {
		fail(w, r, 500, err)
		return
	}
	respond(w, r, 200, site)
}
func (s *Server) deleteTestSite(w http.ResponseWriter, r *http.Request) {
	if err := s.testSiteSvc.DeleteSite(r.PathValue("id")); err != nil {
		fail(w, r, 500, err)
		return
	}
	respond(w, r, 200, map[string]string{"deleted": "ok"})
}

// --- Transformations (missing admin list/update/delete) ---
func (s *Server) listAdminTransformations(w http.ResponseWriter, r *http.Request) {
	page, pageSize := parsePagination(r)
	all, err := s.transSvc.List("")
	if err != nil {
		fail(w, r, 500, fmt.Errorf("list transformations: %w", err))
		return
	}
	filtered, total := adminListFilter(all, r.URL.Query().Get("keyword"), r.URL.Query().Get("status"),
		func(t domain.Transformation) string { return t.Title },
		func(t domain.Transformation) string { return t.Status })
	paginatedRespond(w, r, adminSlicePage(filtered, page, pageSize), total)
}

func (s *Server) updateTransformation(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var in struct {
		Title     string `json:"title"`
		Stage     string `json:"stage"` // lab/pilot/industrialized/listed
		Progress  string `json:"progress"`
		PartnerID string `json:"partner_id"`
		Status    string `json:"status"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, 400, err)
		return
	}
	t, err := s.transSvc.UpdateTrans(id, in.Title, in.Stage, in.Progress, in.PartnerID, in.Status)
	if err != nil {
		fail(w, r, 500, err)
		return
	}
	respond(w, r, 200, t)
}
func (s *Server) deleteTransformation(w http.ResponseWriter, r *http.Request) {
	if err := s.transSvc.DeleteTrans(r.PathValue("id")); err != nil {
		fail(w, r, 500, err)
		return
	}
	respond(w, r, 200, map[string]string{"deleted": "ok"})
}

// --- Events (missing admin list/update/delete) ---
func (s *Server) listAdminEvents(w http.ResponseWriter, r *http.Request) {
	page, pageSize := parsePagination(r)
	all, total, err := s.eventSvc.List(page, pageSize)
	if err != nil {
		fail(w, r, 500, fmt.Errorf("list events: %w", err))
		return
	}
	paginatedRespond(w, r, all, total)
}
func (s *Server) updateEvent(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var in struct {
		Title        string `json:"title"`
		EventType    string `json:"event_type"`
		Description  string `json:"description"`
		Location     string `json:"location"`
		CoverURL     string `json:"cover_url"`
		Status       string `json:"status"`
		StartTime    string `json:"start_time"`
		EndTime      string `json:"end_time"`
		MaxAttendees int    `json:"max_attendees"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, 400, err)
		return
	}
	ev, err := s.eventSvc.Update(id, in.Title, in.EventType, in.Description, in.Location, in.CoverURL, in.Status, domain.ParseTime(in.StartTime), domain.ParseTime(in.EndTime), in.MaxAttendees)
	if err != nil {
		fail(w, r, 500, err)
		return
	}
	respond(w, r, 200, ev)
}
func (s *Server) deleteEvent(w http.ResponseWriter, r *http.Request) {
	if err := s.eventSvc.Delete(r.PathValue("id")); err != nil {
		fail(w, r, 500, err)
		return
	}
	respond(w, r, 200, map[string]string{"deleted": "ok"})
}

// --- Portfolios (missing delete) ---
func (s *Server) deletePortfolio(w http.ResponseWriter, r *http.Request) {
	if err := s.portfolioSvc.Delete(r.PathValue("id")); err != nil {
		fail(w, r, 500, err)
		return
	}
	respond(w, r, 200, map[string]string{"deleted": "ok"})
}

// --- Exhibitions (missing admin list/update/delete) ---
func (s *Server) listAdminExhibitions(w http.ResponseWriter, r *http.Request) {
	page, pageSize := parsePagination(r)
	all, total, err := s.exhibitionSvc.List(page, pageSize)
	if err != nil {
		fail(w, r, 500, fmt.Errorf("list exhibitions: %w", err))
		return
	}
	paginatedRespond(w, r, all, total)
}
func (s *Server) updateExhibition(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var in struct {
		Title       string `json:"title"`
		Category    string `json:"category"`
		Description string `json:"description"`
		Location    string `json:"location"`
		Organizer   string `json:"organizer"`
		Status      string `json:"status"`
		StartDate   string `json:"start_date"`
		EndDate     string `json:"end_date"`
		BoothCount  int    `json:"booth_count"`
		BoothPrice  int64  `json:"booth_price_fen"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, 400, err)
		return
	}
	e, err := s.exhibitionSvc.Update(id, in.Title, in.Category, in.Description, in.Location, in.Organizer, domain.ParseTime(in.StartDate), domain.ParseTime(in.EndDate), in.BoothCount, in.BoothPrice, in.Status)
	if err != nil {
		fail(w, r, 500, err)
		return
	}
	respond(w, r, 200, e)
}
func (s *Server) deleteExhibition(w http.ResponseWriter, r *http.Request) {
	if err := s.exhibitionSvc.Delete(r.PathValue("id")); err != nil {
		fail(w, r, 500, err)
		return
	}
	respond(w, r, 200, map[string]string{"deleted": "ok"})
}

// --- Industry Reports (missing update) ---
func (s *Server) updateIndustryReport(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var in struct {
		Title    string `json:"title"`
		Period   string `json:"period"`
		Category string `json:"category"`
		Summary  string `json:"summary"`
		Content  string `json:"content"`
		FileURL  string `json:"file_url"`
		Author   string `json:"author"`
		Status   string `json:"status"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, 400, err)
		return
	}
	rpt, err := s.reportSvc.Update(id, in.Title, in.Period, in.Category, in.Summary, in.Content, in.FileURL, in.Author, in.Status)
	if err != nil {
		fail(w, r, 500, err)
		return
	}
	respond(w, r, 200, rpt)
}

// --- Emergency Resources (missing admin list/update/delete) ---
func (s *Server) listAdminEmergencyResources(w http.ResponseWriter, r *http.Request) {
	page, pageSize := parsePagination(r)
	all, total, err := s.emergencySvc.ListResources(page, pageSize)
	if err != nil {
		fail(w, r, 500, fmt.Errorf("list emergency resources: %w", err))
		return
	}
	paginatedRespond(w, r, all, total)
}
func (s *Server) updateEmergencyResource(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var in struct {
		Name        string `json:"name"`
		ResType     string `json:"res_type"`
		Specs       string `json:"specs"`
		Location    string `json:"location"`
		ContactInfo string `json:"contact_info"`
		Status      string `json:"status"`
		Quantity    int    `json:"quantity"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, 400, err)
		return
	}
	r2, err := s.emergencySvc.UpdateResource(id, in.Name, in.ResType, in.Specs, in.Location, in.ContactInfo, in.Status, in.Quantity)
	if err != nil {
		fail(w, r, 500, err)
		return
	}
	respond(w, r, 200, r2)
}
func (s *Server) deleteEmergencyResource(w http.ResponseWriter, r *http.Request) {
	if err := s.emergencySvc.DeleteResource(r.PathValue("id")); err != nil {
		fail(w, r, 500, err)
		return
	}
	respond(w, r, 200, map[string]string{"deleted": "ok"})
}

// --- Emergency Dispatches (missing admin list/update/delete) ---
func (s *Server) listAdminEmergencyDispatches(w http.ResponseWriter, r *http.Request) {
	page, pageSize := parsePagination(r)
	all, total, err := s.emergencySvc.ListDispatches(page, pageSize)
	if err != nil {
		fail(w, r, 500, fmt.Errorf("list dispatches: %w", err))
		return
	}
	paginatedRespond(w, r, all, total)
}
func (s *Server) updateEmergencyDispatch(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var in struct {
		ResourceID string `json:"resource_id"`
		EventDesc  string `json:"event_desc"`
		Location   string `json:"location"`
		Commander  string `json:"commander"`
		Result     string `json:"result"`
		Status     string `json:"status"`
		StartTime  string `json:"start_time"`
		EndTime    string `json:"end_time"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, 400, err)
		return
	}
	d, err := s.emergencySvc.UpdateDispatch(id, in.ResourceID, in.EventDesc, in.Location, in.Commander, in.Result, in.Status, domain.ParseTime(in.StartTime), domain.ParseTime(in.EndTime))
	if err != nil {
		fail(w, r, 500, err)
		return
	}
	respond(w, r, 200, d)
}
func (s *Server) deleteEmergencyDispatch(w http.ResponseWriter, r *http.Request) {
	if err := s.emergencySvc.DeleteDispatch(r.PathValue("id")); err != nil {
		fail(w, r, 500, err)
		return
	}
	respond(w, r, 200, map[string]string{"deleted": "ok"})
}

// --- Messages (missing admin list/create/update/delete) ---
func (s *Server) listAdminMessages(w http.ResponseWriter, r *http.Request) {
	page, pageSize := parsePagination(r)
	offset := (page - 1) * pageSize
	all, total, err := s.msgSvc.ListAll(offset, pageSize)
	if err != nil {
		fail(w, r, 500, fmt.Errorf("list messages: %w", err))
		return
	}
	paginatedRespond(w, r, all, total)
}
func (s *Server) createMessage(w http.ResponseWriter, r *http.Request) {
	var in struct {
		SenderID     string `json:"sender_id"`
		ReceiverID   string `json:"receiver_id"`
		Title        string `json:"title"`
		Content      string `json:"content"`
		ResourceType string `json:"resource_type"`
		ResourceID   string `json:"resource_id"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, 400, err)
		return
	}
	msg, err := s.msgSvc.Send(in.SenderID, in.ReceiverID, in.Title, in.Content, in.ResourceType, in.ResourceID)
	if err != nil {
		fail(w, r, 500, err)
		return
	}
	respond(w, r, 201, msg)
}
func (s *Server) updateMessage(w http.ResponseWriter, r *http.Request) {
	msg, err := s.msgSvc.MarkRead(r.PathValue("id"))
	if err != nil {
		fail(w, r, 500, err)
		return
	}
	respond(w, r, 200, msg)
}
func (s *Server) deleteMessage(w http.ResponseWriter, r *http.Request) {
	if err := s.msgSvc.Delete(r.PathValue("id")); err != nil {
		fail(w, r, 500, err)
		return
	}
	respond(w, r, 200, map[string]string{"deleted": "ok"})
}

// --- Compliance Docs (missing update/delete) ---
func (s *Server) updateComplianceDoc(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var in struct {
		ID          string   `json:"id"`
		Title       string   `json:"title"`
		Category    string   `json:"category"`
		Publisher   string   `json:"publisher"`
		PublishDate string   `json:"publish_date"`
		Status      string   `json:"status"`
		Summary     string   `json:"summary"`
		FileURL     string   `json:"file_url"`
		Tags        []string `json:"tags"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	d, err := s.complianceSvc.UpdateDoc(id, in.Title, in.Category, in.Publisher, in.PublishDate, in.Status, in.Summary, in.FileURL, in.Tags)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, d)
}
func (s *Server) deleteComplianceDoc(w http.ResponseWriter, r *http.Request) {
	if err := s.complianceSvc.DeleteDoc(r.PathValue("id")); err != nil {
		fail(w, r, 500, err)
		return
	}
	respond(w, r, 200, map[string]string{"deleted": "ok"})
}

// --- Compliance Standards (missing update/delete) ---
func (s *Server) updateComplianceStandard(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var in struct {
		Title         string `json:"title"`
		StandardNo    string `json:"standard_no"`
		Publisher     string `json:"publisher"`
		EffectiveDate string `json:"effective_date"`
		Scope         string `json:"scope"`
		Status        string `json:"status"`
		FileURL       string `json:"file_url"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, 400, err)
		return
	}
	sd, err := s.complianceSvc.UpdateStandard(id, in.Title, in.StandardNo, in.Publisher, in.EffectiveDate, in.Scope, in.Status, in.FileURL)
	if err != nil {
		fail(w, r, 500, err)
		return
	}
	respond(w, r, 200, sd)
}
func (s *Server) deleteComplianceStandard(w http.ResponseWriter, r *http.Request) {
	if err := s.complianceSvc.DeleteStandard(r.PathValue("id")); err != nil {
		fail(w, r, 500, err)
		return
	}
	respond(w, r, 200, map[string]string{"deleted": "ok"})
}

// --- Industry Resources (missing admin list/delete) ---
func (s *Server) listAdminResources(w http.ResponseWriter, r *http.Request) {
	page, pageSize := parsePagination(r)
	all, total, err := s.resourceSvc.List("", page, pageSize)
	if err != nil {
		fail(w, r, 500, fmt.Errorf("list resources: %w", err))
		return
	}
	paginatedRespond(w, r, all, total)
}
func (s *Server) deleteIndustryResource(w http.ResponseWriter, r *http.Request) {
	if err := s.resourceSvc.Delete(r.PathValue("id")); err != nil {
		fail(w, r, 500, err)
		return
	}
	respond(w, r, 200, map[string]string{"deleted": "ok"})
}

// --- RD Challenges (missing delete) ---
func (s *Server) deleteRDChallenge(w http.ResponseWriter, r *http.Request) {
	if err := s.rdService.Delete(r.PathValue("id")); err != nil {
		fail(w, r, 500, err)
		return
	}
	respond(w, r, 200, map[string]string{"deleted": "ok"})
}

// --- Research Projects (missing delete) ---
func (s *Server) deleteResearchProject(w http.ResponseWriter, r *http.Request) {
	if err := s.researchSvc.Delete(r.PathValue("id")); err != nil {
		fail(w, r, 500, err)
		return
	}
	respond(w, r, 200, map[string]string{"deleted": "ok"})
}

func (s *Server) getCourse(w http.ResponseWriter, r *http.Request) {
	c, err := s.trainingSvc.GetCourse(r.PathValue("id"))
	if err != nil {
		fail(w, r, 404, err)
		return
	}
	respond(w, r, 200, c)
}

func (s *Server) getCert(w http.ResponseWriter, r *http.Request) {
	c, err := s.trainingSvc.GetCert(r.PathValue("id"))
	if err != nil {
		fail(w, r, 404, err)
		return
	}
	respond(w, r, 200, c)
}

func (s *Server) getJob(w http.ResponseWriter, r *http.Request) {
	j, err := s.jobSvc.GetJob(r.PathValue("id"))
	if err != nil {
		fail(w, r, 404, err)
		return
	}
	respond(w, r, 200, j)
}

func (s *Server) getCompetition(w http.ResponseWriter, r *http.Request) {
	c, err := s.competitionSvc.Get(r.PathValue("id"))
	if err != nil {
		fail(w, r, 404, err)
		return
	}
	respond(w, r, 200, c)
}

func (s *Server) getCollege(w http.ResponseWriter, r *http.Request) {
	c, err := s.collegeSvc.Get(r.PathValue("id"))
	if err != nil {
		fail(w, r, 404, err)
		return
	}
	respond(w, r, 200, c)
}

func (s *Server) getStudyTour(w http.ResponseWriter, r *http.Request) {
	s2, err := s.studyTourRepo.FindByID(r.PathValue("id"))
	if err != nil {
		fail(w, r, 404, err)
		return
	}
	respond(w, r, 200, s2)
}

func (s *Server) getEvent(w http.ResponseWriter, r *http.Request) {
	e, err := s.eventSvc.Get(r.PathValue("id"))
	if err != nil {
		fail(w, r, 404, err)
		return
	}
	respond(w, r, 200, e)
}

func (s *Server) getExhibition(w http.ResponseWriter, r *http.Request) {
	e, err := s.exhibitionSvc.Get(r.PathValue("id"))
	if err != nil {
		fail(w, r, 404, err)
		return
	}
	respond(w, r, 200, e)
}

func (s *Server) getEmergencyResource(w http.ResponseWriter, r *http.Request) {
	r2, err := s.emergencySvc.GetResource(r.PathValue("id"))
	if err != nil {
		fail(w, r, 404, err)
		return
	}
	respond(w, r, 200, r2)
}

func (s *Server) getReport(w http.ResponseWriter, r *http.Request) {
	rpt, err := s.reportSvc.Get(r.PathValue("id"))
	if err != nil {
		fail(w, r, 404, err)
		return
	}
	respond(w, r, 200, rpt)
}
func (s *Server) getIndustryResource(w http.ResponseWriter, r *http.Request) {
	r2, err := s.resourceSvc.Get(r.PathValue("id"))
	if err != nil {
		fail(w, r, 404, err)
		return
	}
	respond(w, r, 200, r2)
}
func (s *Server) getPortfolio(w http.ResponseWriter, r *http.Request) {
	p, err := s.portfolioSvc.Get(r.PathValue("id"))
	if err != nil {
		fail(w, r, 404, err)
		return
	}
	respond(w, r, 200, p)
}
func (s *Server) getExpert(w http.ResponseWriter, r *http.Request) {
	e, err := s.expertSvc.Get(r.PathValue("id"))
	if err != nil {
		fail(w, r, 404, err)
		return
	}
	respond(w, r, 200, e)
}
func (s *Server) getAchievement(w http.ResponseWriter, r *http.Request) {
	a, err := s.achievementSvc.Get(r.PathValue("id"))
	if err != nil {
		fail(w, r, 404, err)
		return
	}
	respond(w, r, 200, a)
}
func (s *Server) getRDChallenge(w http.ResponseWriter, r *http.Request) {
	c2, err := s.rdService.Get(r.PathValue("id"))
	if err != nil {
		fail(w, r, 404, err)
		return
	}
	respond(w, r, 200, c2)
}
func (s *Server) getResearchProject(w http.ResponseWriter, r *http.Request) {
	rp, err := s.researchSvc.Get(r.PathValue("id"))
	if err != nil {
		fail(w, r, 404, err)
		return
	}
	respond(w, r, 200, rp)
}
func (s *Server) getTestSite(w http.ResponseWriter, r *http.Request) {
	ts, err := s.testSiteSvc.Get(r.PathValue("id"))
	if err != nil {
		fail(w, r, 404, err)
		return
	}
	respond(w, r, 200, ts)
}
func (s *Server) getTransformation(w http.ResponseWriter, r *http.Request) {
	t, err := s.transSvc.Get(r.PathValue("id"))
	if err != nil {
		fail(w, r, 404, err)
		return
	}
	respond(w, r, 200, t)
}
func (s *Server) getComplianceDoc(w http.ResponseWriter, r *http.Request) {
	cd, err := s.complianceSvc.FindDocByID(r.PathValue("id"))
	if err != nil {
		fail(w, r, 404, err)
		return
	}
	respond(w, r, 200, cd)
}
func (s *Server) getComplianceStandard(w http.ResponseWriter, r *http.Request) {
	cs, err := s.complianceSvc.FindStandardByID(r.PathValue("id"))
	if err != nil {
		fail(w, r, 404, err)
		return
	}
	respond(w, r, 200, cs)
}
func (s *Server) getMessage(w http.ResponseWriter, r *http.Request) {
	// messages don't have a direct Get, just mark as read
	respond(w, r, 200, map[string]string{"note": "detail not supported"})
}
func (s *Server) getEmergencyDispatch(w http.ResponseWriter, r *http.Request) {
	d, err := s.emergencySvc.FindDispatchByID(r.PathValue("id"))
	if err != nil {
		fail(w, r, 404, err)
		return
	}
	respond(w, r, 200, d)
}
func (s *Server) getOrder(w http.ResponseWriter, r *http.Request) {
	o, err := s.tradeSvc.FindByID(r.PathValue("id"))
	if err != nil {
		fail(w, r, 404, err)
		return
	}
	respond(w, r, 200, o)
}

func (s *Server) createOrder(w http.ResponseWriter, r *http.Request) {
	var in struct {
		ProductID string `json:"product_id"`
		BuyerID   string `json:"buyer_id"`
		SellerID  string `json:"seller_id"`
		AmountFen int64  `json:"amount_fen"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, 400, err)
		return
	}
	o, err := s.tradeSvc.Create(in.BuyerID, in.ProductID, in.SellerID, in.AmountFen)
	if err != nil {
		fail(w, r, 500, err)
		return
	}
	respond(w, r, 201, o)
}
func (s *Server) updateOrder(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Status string `json:"status"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, 400, err)
		return
	}
	o, err := s.tradeSvc.UpdateStatusAdmin(r.PathValue("id"), in.Status)
	if err != nil {
		fail(w, r, 500, err)
		return
	}
	respond(w, r, 200, o)
}
func (s *Server) deleteOrder(w http.ResponseWriter, r *http.Request) {
	if err := s.tradeSvc.Delete(r.PathValue("id")); err != nil {
		fail(w, r, 500, fmt.Errorf("delete order: %w", err))
		return
	}
	respond(w, r, 200, map[string]string{"deleted": "ok"})
}
