package httpapi

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
	"drone-platform/internal/service"
)

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

// adminFail 管理端写操作错误映射：资源不存在 → 404，越权 → 403，其余 → 500。
// 修复前 update/delete 一律 500，前端无法区分"资源不存在"与真实服务故障。
func adminFail(w http.ResponseWriter, r *http.Request, err error) {
	code := http.StatusInternalServerError
	switch {
	case errors.Is(err, service.ErrNotOwner):
		code = http.StatusForbidden
	case strings.Contains(strings.ToLower(err.Error()), "not found"):
		code = http.StatusNotFound
	}
	fail(w, r, code, err)
}

// ----- Enrollments (报名记录) -----
func (s *Server) listAdminEnrollments(w http.ResponseWriter, r *http.Request) {
	all, _, err := s.enrollSvc.All(r.Context(), 0, 10000)
	if err != nil {
		fail(w, r, 500, fmt.Errorf("list enrollments: %w", err))
		return
	}
	filtered, total := adminListFilter(all, r.URL.Query().Get("keyword"), r.URL.Query().Get("status"),
		func(e domain.Enrollment) string { return e.Name + e.Phone },
		func(e domain.Enrollment) string { return e.Status })
	paginatedRespond(w, r, filtered, total)
}

// ----- Test site bookings (场地预约记录) -----
func (s *Server) listAdminBookings(w http.ResponseWriter, r *http.Request) {
	all, _, err := s.testSiteSvc.ListAllBookings(r.Context(), 0, 10000)
	if err != nil {
		fail(w, r, 500, fmt.Errorf("list bookings: %w", err))
		return
	}
	filtered, total := adminListFilter(all, r.URL.Query().Get("keyword"), r.URL.Query().Get("status"),
		func(b domain.TestSiteBooking) string { return b.ContactName + b.Purpose },
		func(b domain.TestSiteBooking) string { return b.Status })
	paginatedRespond(w, r, filtered, total)
}

// ----- Orders (trade_orders) -----
// P3 修复：过滤 + 分页下沉 SQL（此前全量拉 10000 条再内存过滤，更早数据永远不可见）。
func (s *Server) listAdminOrders(w http.ResponseWriter, r *http.Request) {
	page, pageSize := paginationFromQuery(r)
	f := repository.TradeOrderFilter{
		Status: strings.TrimSpace(r.URL.Query().Get("status")),
		Offset: (page - 1) * pageSize,
		Limit:  pageSize,
	}
	if kw := strings.TrimSpace(r.URL.Query().Get("keyword")); kw != "" {
		f.Keyword = kw
	}
	if sd := r.URL.Query().Get("start_date"); sd != "" {
		if t, err := time.Parse("2006-01-02", sd); err == nil {
			f.StartDate = &t
		}
	}
	if ed := r.URL.Query().Get("end_date"); ed != "" {
		if t, err := time.Parse("2006-01-02", ed); err == nil {
			end := t.Add(24 * time.Hour) // 含当日
			f.EndDate = &end
		}
	}
	items, total, err := s.tradeSvc.ListAllFiltered(r.Context(), f)
	if err != nil {
		adminFail(w, r, err)
		return
	}
	respondPage(w, r, items, total, page, pageSize)
}

// ----- Reviews -----
// GET /api/v1/admin/reviews 实际绑定 routes_phase3.go 的 listAllReviews。

// ----- Case entries -----
func (s *Server) listAdminCaseEntries(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	items, total, err := s.caseSvc.List(r.Context(), category, 1, 100000)
	if err != nil {
		adminFail(w, r, err)
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
	items, _, err := s.competitionSvc.List(r.Context(), 1, 100000)
	if err != nil {
		adminFail(w, r, err)
		return
	}
	filtered, ftotal := adminListFilter(items, r.URL.Query().Get("keyword"), r.URL.Query().Get("status"),
		func(c domain.Competition) string { return c.Title },
		func(c domain.Competition) string { return c.Status })
	paginatedRespond(w, r, filtered, ftotal)
}

// POST /api/v1/admin/competitions 实际绑定 biz_handlers.go 的 createCompetition。
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
		// 小程序赛事页扩展字段
		Deadline           string                          `json:"deadline"`
		OrganizerSub       string                          `json:"organizer_sub"`
		Fee                int                             `json:"fee"`
		MinFee             int                             `json:"min_fee"`
		OriginalFee        int                             `json:"original_fee"`
		Tags               []string                        `json:"tags"`
		Poster             string                          `json:"poster"`
		Requirements       []domain.CompetitionRequirement `json:"requirements"`
		Events             []domain.CompetitionEvent       `json:"events"`
		Prizes             []domain.CompetitionPrize       `json:"prizes"`
		RegistrationStatus string                          `json:"registration_status"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	var deadline *time.Time
	if d, err := parseDateInput(in.Deadline); err == nil && !d.IsZero() {
		deadline = &d
	}
	// P2 修复：严格解析日期——非法日期此前被 ParseTime 静默写成当前时间落库。
	startDate, ok := strictDate(w, r, in.StartDate)
	if !ok {
		return
	}
	endDate, ok := strictDate(w, r, in.EndDate)
	if !ok {
		return
	}
	c, err := s.competitionSvc.Update(r.Context(), domain.Competition{
		ID: id, Title: in.Title, Category: in.Category, Description: in.Description,
		Location: in.Location, Sponsor: in.Sponsor, Status: in.Status,
		StartDate: startDate, EndDate: endDate,
		MaxTeams: in.MaxTeams, Deadline: deadline, OrganizerSub: in.OrganizerSub,
		Fee: in.Fee, MinFee: in.MinFee, OriginalFee: in.OriginalFee, Tags: in.Tags, Poster: in.Poster,
		Requirements: in.Requirements, Events: in.Events, Prizes: in.Prizes,
		RegistrationStatus: in.RegistrationStatus,
	})
	if err != nil {
		adminFail(w, r, err)
		return
	}
	respond(w, r, http.StatusOK, c)
}
func (s *Server) deleteCompetition(w http.ResponseWriter, r *http.Request) {
	if err := s.competitionSvc.Delete(r.Context(), r.PathValue("id")); err != nil {
		adminFail(w, r, err)
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
		// 小程序培训页扩展字段
		OrgName       string               `json:"org_name"`
		Rating        string               `json:"rating"`
		ReviewCount   int                  `json:"review_count"`
		District      string               `json:"district"`
		DurationDays  int                  `json:"duration_days"`
		Image         string               `json:"image"`
		Tags          []string             `json:"tags"`
		Certificate   string               `json:"certificate"`
		Courses       []domain.CoursePrice `json:"courses"`
		Prices        []domain.CoursePrice `json:"prices"`
		BusinessHours string               `json:"business_hours"`
		Phone         string               `json:"phone"`
		Remain        int                  `json:"remain"`
		Environment   []string             `json:"environment"`
		CourseTypes   []string             `json:"course_types"`
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
	// P2 修复：严格解析日期——非法日期此前被 ParseTime 静默写成当前时间落库。
	startDate, ok := strictDate(w, r, in.StartDate)
	if !ok {
		return
	}
	endDate, ok := strictDate(w, r, in.EndDate)
	if !ok {
		return
	}
	c, err := s.trainingSvc.CreateCourse(r.Context(), domain.Actor{Role: domain.RolePlatformAdmin}, domain.TrainingCourse{
		Title: in.Title, CertType: domain.CertType(ct), Description: in.Description,
		Location: in.Location, StartDate: startDate, EndDate: endDate,
		MaxStudents: ms, PriceFen: in.PriceFen, Status: in.Status,
		OrgName: in.OrgName, Rating: in.Rating, ReviewCount: in.ReviewCount,
		District: in.District, DurationDays: in.DurationDays, Image: in.Image,
		Tags: in.Tags, Certificate: in.Certificate, Courses: in.Courses,
		Prices: in.Prices, BusinessHours: in.BusinessHours, Phone: in.Phone,
		Remain: in.Remain, Environment: in.Environment, CourseTypes: in.CourseTypes,
	})
	if err != nil {
		adminFail(w, r, err)
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
		// 小程序培训页扩展字段
		OrgName       string               `json:"org_name"`
		Rating        string               `json:"rating"`
		ReviewCount   int                  `json:"review_count"`
		District      string               `json:"district"`
		DurationDays  int                  `json:"duration_days"`
		Image         string               `json:"image"`
		Tags          []string             `json:"tags"`
		Certificate   string               `json:"certificate"`
		Courses       []domain.CoursePrice `json:"courses"`
		Prices        []domain.CoursePrice `json:"prices"`
		BusinessHours string               `json:"business_hours"`
		Phone         string               `json:"phone"`
		Remain        int                  `json:"remain"`
		Environment   []string             `json:"environment"`
		CourseTypes   []string             `json:"course_types"`
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
	// P2 修复：严格解析日期——非法日期此前被 ParseTime 静默写成当前时间落库。
	startDate, ok := strictDate(w, r, in.StartDate)
	if !ok {
		return
	}
	endDate, ok := strictDate(w, r, in.EndDate)
	if !ok {
		return
	}
	c, err := s.trainingSvc.UpdateCourse(r.Context(), domain.TrainingCourse{
		ID: id, Title: in.Title, CertType: domain.CertType(ct), Description: in.Description,
		Location: in.Location, StartDate: startDate, EndDate: endDate,
		MaxStudents: ms, PriceFen: in.PriceFen, Status: in.Status,
		OrgName: in.OrgName, Rating: in.Rating, ReviewCount: in.ReviewCount,
		District: in.District, DurationDays: in.DurationDays, Image: in.Image,
		Tags: in.Tags, Certificate: in.Certificate, Courses: in.Courses,
		Prices: in.Prices, BusinessHours: in.BusinessHours, Phone: in.Phone,
		Remain: in.Remain, Environment: in.Environment, CourseTypes: in.CourseTypes,
	})
	if err != nil {
		adminFail(w, r, err)
		return
	}
	respond(w, r, http.StatusOK, c)
}
func (s *Server) deleteCourse(w http.ResponseWriter, r *http.Request) {
	// 删除前校验关联报名：已有学员报名的课程拒绝删除（防孤儿报名数据）
	enrolls, err := s.enrollSvc.ListByCourse(r.Context(), r.PathValue("id"))
	if err != nil {
		adminFail(w, r, err)
		return
	}
	if len(enrolls) > 0 {
		fail(w, r, http.StatusConflict, fmt.Errorf("course has %d enrollment(s), cannot delete", len(enrolls)))
		return
	}
	if err := s.trainingSvc.DeleteCourse(r.Context(), r.PathValue("id")); err != nil {
		adminFail(w, r, err)
		return
	}
	respond(w, r, http.StatusOK, map[string]string{"deleted": "ok"})
}

// --- Certificates (missing admin list/update/delete) ---
func (s *Server) listAdminCerts(w http.ResponseWriter, r *http.Request) {
	certs, err := s.trainingSvc.ListAllCertificates(r.Context())
	if err != nil {
		fail(w, r, 500, fmt.Errorf("list certs: %w", err))
		return
	}
	filtered, total := adminListFilter(certs, r.URL.Query().Get("keyword"), r.URL.Query().Get("status"),
		func(c domain.Certificate) string { return c.CertNumber },
		func(c domain.Certificate) string { return c.Status })
	paginatedRespond(w, r, filtered, total)
}
func (s *Server) updateCertificate(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var in struct {
		CertType   string `json:"cert_type"`
		CertNumber string `json:"cert_number"`
		Level      string `json:"level"`
		IssuerOrg  string `json:"issuer_org"`
		// 前端 a-date-picker value-format="YYYY-MM-DD" 提交日期字符串（可空），
		// 用 string 承接再解析，避免 time.Time 只认 RFC3339 导致 'YYYY-MM-DD' 400。
		IssueDate  string `json:"issue_date"`
		ExpireDate string `json:"expire_date"`
		Status     string `json:"status"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	issueDate, err := parseDateInput(in.IssueDate)
	if err != nil {
		fail(w, r, http.StatusBadRequest, fmt.Errorf("无效的签发日期格式: %w", err))
		return
	}
	expireDate, err := parseDateInput(in.ExpireDate)
	if err != nil {
		fail(w, r, http.StatusBadRequest, fmt.Errorf("无效的到期日期格式: %w", err))
		return
	}
	c, err := s.trainingSvc.UpdateCertificate(r.Context(), id, in.CertType, in.CertNumber, in.Level, in.IssuerOrg, in.Status, issueDate, expireDate)
	if err != nil {
		adminFail(w, r, err)
		return
	}
	respond(w, r, http.StatusOK, c)
}
func (s *Server) deleteCertificate(w http.ResponseWriter, r *http.Request) {
	if err := s.trainingSvc.DeleteCertificate(r.Context(), r.PathValue("id")); err != nil {
		adminFail(w, r, err)
		return
	}
	respond(w, r, http.StatusOK, map[string]string{"deleted": "ok"})
}

func (s *Server) adminCreateCertificate(w http.ResponseWriter, r *http.Request) {
	var in struct {
		CertType   string `json:"cert_type"`
		CertNumber string `json:"cert_number"`
		Level      string `json:"level"`
		IssuerOrg  string `json:"issuer_org"`
		// 前端 a-date-picker value-format="YYYY-MM-DD" 提交日期字符串（可空），
		// 用 string 承接再解析，避免 time.Time 只认 RFC3339 导致 'YYYY-MM-DD' 400。
		IssueDate  string `json:"issue_date"`
		ExpireDate string `json:"expire_date"`
		Status     string `json:"status"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	issueDate, err := parseDateInput(in.IssueDate)
	if err != nil {
		fail(w, r, http.StatusBadRequest, fmt.Errorf("无效的签发日期格式: %w", err))
		return
	}
	expireDate, err := parseDateInput(in.ExpireDate)
	if err != nil {
		fail(w, r, http.StatusBadRequest, fmt.Errorf("无效的到期日期格式: %w", err))
		return
	}
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, fmt.Errorf("auth required"))
		return
	}
	c, err := s.trainingSvc.AddCertificate(r.Context(), domain.Actor{ID: a.ID, Role: domain.RolePlatformAdmin}, domain.CertType(in.CertType), in.CertNumber, in.Level, in.IssuerOrg, issueDate, expireDate)
	if err != nil {
		adminFail(w, r, err)
		return
	}
	respond(w, r, http.StatusCreated, c)
}

// --- Jobs (missing admin list/update/delete) ---
func (s *Server) listAdminJobs(w http.ResponseWriter, r *http.Request) {
	all, _, err := s.jobSvc.ListAllJobs(r.Context(), 0, 10000)
	if err != nil {
		fail(w, r, 500, fmt.Errorf("list jobs: %w", err))
		return
	}
	filtered, total := adminListFilter(all, r.URL.Query().Get("keyword"), r.URL.Query().Get("status"),
		func(j domain.Job) string { return j.Title },
		func(j domain.Job) string { return string(j.Status) })
	paginatedRespond(w, r, filtered, total)
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
	j, err := s.jobSvc.UpdateJob(r.Context(), id, in.Title, in.Description, in.Location, in.JobType, in.SalaryFen, in.Status)
	if err != nil {
		code := http.StatusInternalServerError
		if errors.Is(err, service.ErrInvalidJobStatus) {
			code = http.StatusBadRequest
		}
		fail(w, r, code, err)
		return
	}
	respond(w, r, http.StatusOK, j)
}
func (s *Server) deleteJob(w http.ResponseWriter, r *http.Request) {
	if err := s.jobSvc.DeleteJob(r.Context(), r.PathValue("id")); err != nil {
		adminFail(w, r, err)
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
	j, err := s.jobSvc.CreateJob(r.Context(), domain.Actor{ID: "admin", Role: domain.RolePlatformAdmin}, in.Title, in.Description, in.Location, in.SalaryFen, in.JobType)
	if err != nil {
		adminFail(w, r, err)
		return
	}
	if in.Status == "published" {
		var err error
		// 与 CreateJob 使用同一 actor（ID:"admin"），否则 owner 校验失败
		j, err = s.jobSvc.PublishJob(r.Context(), domain.Actor{ID: "admin", Role: domain.RolePlatformAdmin}, j.ID)
		if err != nil {
			fail(w, r, http.StatusInternalServerError, fmt.Errorf("publish job: %w", err))
			return
		}
	}
	respond(w, r, http.StatusCreated, j)
}

// --- Colleges (missing admin list/update/delete) ---
func (s *Server) listAdminColleges(w http.ResponseWriter, r *http.Request) {
	all, err := s.collegeSvc.List(r.Context(), "")
	if err != nil {
		fail(w, r, 500, fmt.Errorf("list colleges: %w", err))
		return
	}
	filtered, total := adminListFilter(all, r.URL.Query().Get("keyword"), r.URL.Query().Get("status"),
		func(c domain.College) string { return c.Name },
		func(c domain.College) string { return c.Status })
	paginatedRespond(w, r, filtered, total)
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
		// 小程序院校页扩展字段
		City         string                  `json:"city"`
		Tags         []string                `json:"tags"`
		ShortName    string                  `json:"short_name"`
		LevelTags    string                  `json:"level_tags"`
		Specialties  []string                `json:"specialties"`
		MajorCount   int                     `json:"major_count"`
		PartnerCount int                     `json:"partner_count"`
		TeacherCount int                     `json:"teacher_count"`
		StudentCount int                     `json:"student_count"`
		GraduateRate string                  `json:"graduate_rate"`
		Partners     []domain.CollegePartner `json:"partners"`
		Cover        string                  `json:"cover"`
		Photos       []string                `json:"photos"`
		Phone        string                  `json:"phone"`
		Website      string                  `json:"website"`
		Intro        string                  `json:"intro"`
		MajorsDetail []domain.CollegeMajor   `json:"majors_detail"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	c, err := s.collegeSvc.Update(r.Context(), domain.College{
		ID: id, Name: in.Name, Region: in.Region, City: in.City, Description: in.Description,
		LogoURL: in.LogoURL, Status: in.Status, CoopType: in.CoopType,
		Majors: in.Majors, Facilities: in.Facilities,
		Tags: in.Tags, ShortName: in.ShortName, LevelTags: in.LevelTags, Specialties: in.Specialties,
		MajorCount: in.MajorCount, PartnerCount: in.PartnerCount, TeacherCount: in.TeacherCount,
		StudentCount: in.StudentCount, GraduateRate: in.GraduateRate, Partners: in.Partners,
		CoverURL: in.Cover, Photos: in.Photos, Phone: in.Phone, Website: in.Website,
		Intro: in.Intro, MajorsDetail: in.MajorsDetail,
	})
	if err != nil {
		adminFail(w, r, err)
		return
	}
	respond(w, r, http.StatusOK, c)
}
func (s *Server) deleteCollege(w http.ResponseWriter, r *http.Request) {
	if err := s.collegeSvc.Delete(r.Context(), r.PathValue("id")); err != nil {
		adminFail(w, r, err)
		return
	}
	respond(w, r, http.StatusOK, map[string]string{"deleted": "ok"})
}

// POST /api/v1/admin/colleges 实际绑定 batch2_handlers.go 的 createCollege。

// --- Study Tours (missing list/create/update/delete) ---
func (s *Server) listAdminStudy(w http.ResponseWriter, r *http.Request) {
	items, err := s.studyTourRepo.List(r.Context())
	if err != nil {
		adminFail(w, r, err)
		return
	}
	filtered, total := adminListFilter(items, r.URL.Query().Get("keyword"), r.URL.Query().Get("status"),
		func(t domain.StudyTour) string { return t.Title },
		func(t domain.StudyTour) string { return t.Status })
	paginatedRespond(w, r, convStudy(filtered), total)
}

func convStudy(items []domain.StudyTour) []map[string]any {
	out := make([]map[string]any, len(items))
	for i, it := range items {
		out[i] = map[string]any{
			"id": it.ID, "title": it.Title, "destination": it.Destination,
			"duration": it.Duration, "capacity": it.Capacity, "status": it.Status,
			"description": it.Description, "location": it.Location, "organizer_id": it.OrganizerID,
			"cover_image": it.CoverImage, "price_fen": it.PriceFen, "schedule": it.Schedule,
			"start_date": it.StartDate, "end_date": it.EndDate,
			"created_at": it.CreatedAt, "updated_at": it.UpdatedAt,
		}
	}
	return out
}

func (s *Server) createStudyTour(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Title       string                 `json:"title"`
		Destination string                 `json:"destination"`
		Duration    string                 `json:"duration"`
		Capacity    int                    `json:"capacity"`
		Status      string                 `json:"status"`
		Description string                 `json:"description"`
		Location    string                 `json:"location"`
		StartDate   string                 `json:"start_date"`
		EndDate     string                 `json:"end_date"`
		CoverImage  string                 `json:"cover_image"`
		PriceFen    int64                  `json:"price_fen"`
		Schedule    []domain.StudySchedule `json:"schedule"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	// P2 修复：严格解析日期——非法日期此前被 ParseTime 静默写成当前时间落库。
	startDate, ok := strictDate(w, r, in.StartDate)
	if !ok {
		return
	}
	endDate, ok := strictDate(w, r, in.EndDate)
	if !ok {
		return
	}
	st := domain.StudyTour{ID: fmt.Sprintf("study-%d", time.Now().UnixNano()), Title: in.Title, Destination: in.Destination, Duration: in.Duration, Capacity: in.Capacity, Status: in.Status, Description: in.Description,
		Location: in.Location, StartDate: startDate, EndDate: endDate,
		CoverImage: in.CoverImage, PriceFen: in.PriceFen, Schedule: in.Schedule}
	sr, err := s.studyTourRepo.Create(r.Context(), st)
	if err != nil {
		adminFail(w, r, err)
		return
	}
	respond(w, r, http.StatusCreated, sr)
}
func (s *Server) updateStudyTour(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var in struct {
		Title       string                 `json:"title"`
		Destination string                 `json:"destination"`
		Duration    string                 `json:"duration"`
		Capacity    int                    `json:"capacity"`
		Status      string                 `json:"status"`
		Description string                 `json:"description"`
		Location    string                 `json:"location"`
		StartDate   string                 `json:"start_date"`
		EndDate     string                 `json:"end_date"`
		CoverImage  string                 `json:"cover_image"`
		PriceFen    int64                  `json:"price_fen"`
		Schedule    []domain.StudySchedule `json:"schedule"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	st, err := s.studyTourRepo.FindByID(r.Context(), id)
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
	st.Location = in.Location
	if in.StartDate != "" {
		// P2 修复：严格解析——非法日期此前被 ParseTime 静默写成当前时间落库。
		t, ok := strictDate(w, r, in.StartDate)
		if !ok {
			return
		}
		st.StartDate = t
	}
	if in.EndDate != "" {
		t, ok := strictDate(w, r, in.EndDate)
		if !ok {
			return
		}
		st.EndDate = t
	}
	st.CoverImage = in.CoverImage
	st.PriceFen = in.PriceFen
	st.Schedule = in.Schedule
	sr, err := s.studyTourRepo.Update(r.Context(), st)
	if err != nil {
		adminFail(w, r, err)
		return
	}
	respond(w, r, http.StatusOK, sr)
}
func (s *Server) deleteStudyTour(w http.ResponseWriter, r *http.Request) {
	if err := s.studyTourRepo.Delete(r.Context(), r.PathValue("id")); err != nil {
		adminFail(w, r, err)
		return
	}
	respond(w, r, http.StatusOK, map[string]string{"deleted": "ok"})
}

// --- Test Sites (missing admin list/update/delete) ---
func (s *Server) listAdminTestSites(w http.ResponseWriter, r *http.Request) {
	all, err := s.testSiteSvc.List(r.Context(), r.URL.Query().Get("site_type"))
	if err != nil {
		fail(w, r, 500, fmt.Errorf("list test sites: %w", err))
		return
	}
	filtered, total := adminListFilter(all, r.URL.Query().Get("keyword"), r.URL.Query().Get("status"),
		func(t domain.TestSite) string { return t.Name },
		func(t domain.TestSite) string { return t.Status })
	paginatedRespond(w, r, filtered, total)
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
	site, err := s.testSiteSvc.UpdateSite(r.Context(), id, in.Name, in.SiteType, in.Location, in.BookingRule, in.Status, in.PriceFen, in.Facilities)
	if err != nil {
		adminFail(w, r, err)
		return
	}
	respond(w, r, 200, site)
}
func (s *Server) deleteTestSite(w http.ResponseWriter, r *http.Request) {
	if err := s.testSiteSvc.DeleteSite(r.Context(), r.PathValue("id")); err != nil {
		adminFail(w, r, err)
		return
	}
	respond(w, r, 200, map[string]string{"deleted": "ok"})
}

// --- Transformations (missing admin list/update/delete) ---
func (s *Server) listAdminTransformations(w http.ResponseWriter, r *http.Request) {
	all, err := s.transSvc.List(r.Context(), "")
	if err != nil {
		fail(w, r, 500, fmt.Errorf("list transformations: %w", err))
		return
	}
	filtered, _ := adminListFilter(all, r.URL.Query().Get("keyword"), r.URL.Query().Get("status"),
		func(t domain.Transformation) string { return t.Title },
		func(t domain.Transformation) string { return t.Status })
	// 阶段筛选（stage）
	if st := r.URL.Query().Get("stage"); st != "" {
		tmp := make([]domain.Transformation, 0, len(filtered))
		for _, t := range filtered {
			if string(t.Stage) == st {
				tmp = append(tmp, t)
			}
		}
		filtered = tmp
	}
	paginatedRespond(w, r, filtered, len(filtered))
}

func (s *Server) updateTransformation(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var in struct {
		Title         string `json:"title"`
		AchievementID string `json:"achievement_id"`
		Stage         string `json:"stage"` // lab/pilot/industrialized/listed
		Progress      string `json:"progress"`
		PartnerID     string `json:"partner_id"`
		Status        string `json:"status"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, 400, err)
		return
	}
	t, err := s.transSvc.UpdateTrans(r.Context(), id, in.Title, in.AchievementID, in.Stage, in.Progress, in.PartnerID, in.Status)
	if err != nil {
		adminFail(w, r, err)
		return
	}
	respond(w, r, 200, t)
}
func (s *Server) deleteTransformation(w http.ResponseWriter, r *http.Request) {
	if err := s.transSvc.DeleteTrans(r.Context(), r.PathValue("id")); err != nil {
		adminFail(w, r, err)
		return
	}
	respond(w, r, 200, map[string]string{"deleted": "ok"})
}

// --- Events (missing admin list/update/delete) ---
func (s *Server) listAdminEvents(w http.ResponseWriter, r *http.Request) {
	all, _, err := s.eventSvc.List(r.Context(), 1, 100000)
	if err != nil {
		fail(w, r, 500, fmt.Errorf("list events: %w", err))
		return
	}
	filtered, _ := adminListFilter(all, r.URL.Query().Get("keyword"), r.URL.Query().Get("status"),
		func(e domain.AssociationEvent) string { return e.Title },
		func(e domain.AssociationEvent) string { return e.Status })
	paginatedRespond(w, r, filtered, len(filtered))
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
	// P2 修复：严格解析日期——非法日期此前被 ParseTime 静默写成当前时间落库。
	startTime, ok := strictDate(w, r, in.StartTime)
	if !ok {
		return
	}
	endTime, ok := strictDate(w, r, in.EndTime)
	if !ok {
		return
	}
	ev, err := s.eventSvc.Update(r.Context(), id, in.Title, in.EventType, in.Description, in.Location, in.CoverURL, in.Status, startTime, endTime, in.MaxAttendees)
	if err != nil {
		adminFail(w, r, err)
		return
	}
	respond(w, r, 200, ev)
}
func (s *Server) deleteEvent(w http.ResponseWriter, r *http.Request) {
	if err := s.eventSvc.Delete(r.Context(), r.PathValue("id")); err != nil {
		adminFail(w, r, err)
		return
	}
	respond(w, r, 200, map[string]string{"deleted": "ok"})
}

// --- Portfolios (missing delete) ---
func (s *Server) deletePortfolio(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	if err := s.portfolioSvc.Delete(r.Context(), a, r.PathValue("id")); err != nil {
		adminFail(w, r, err)
		return
	}
	respond(w, r, 200, map[string]string{"deleted": "ok"})
}

// --- Exhibitions (missing admin list/update/delete) ---
func (s *Server) listAdminExhibitions(w http.ResponseWriter, r *http.Request) {
	all, _, err := s.exhibitionSvc.List(r.Context(), 1, 100000)
	if err != nil {
		fail(w, r, 500, fmt.Errorf("list exhibitions: %w", err))
		return
	}
	filtered, _ := adminListFilter(all, r.URL.Query().Get("keyword"), r.URL.Query().Get("status"),
		func(e domain.Exhibition) string { return e.Title },
		func(e domain.Exhibition) string { return e.Status })
	paginatedRespond(w, r, filtered, len(filtered))
}
func (s *Server) updateExhibition(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var in struct {
		Title       string `json:"title"`
		Category    string `json:"category"`
		Description string `json:"description"`
		Location    string `json:"location"`
		Organizer   string `json:"organizer"`
		CoverURL    string `json:"cover_url"`
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
	// P2 修复：严格解析日期——非法日期此前被 ParseTime 静默写成当前时间落库。
	startDate, ok := strictDate(w, r, in.StartDate)
	if !ok {
		return
	}
	endDate, ok := strictDate(w, r, in.EndDate)
	if !ok {
		return
	}
	e, err := s.exhibitionSvc.Update(r.Context(), id, in.Title, in.Category, in.Description, in.Location, in.Organizer, in.CoverURL, startDate, endDate, in.BoothCount, in.BoothPrice, in.Status)
	if err != nil {
		adminFail(w, r, err)
		return
	}
	respond(w, r, 200, e)
}
func (s *Server) deleteExhibition(w http.ResponseWriter, r *http.Request) {
	if err := s.exhibitionSvc.Delete(r.Context(), r.PathValue("id")); err != nil {
		adminFail(w, r, err)
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
	rpt, err := s.reportSvc.Update(r.Context(), id, in.Title, in.Period, in.Category, in.Summary, in.Content, in.FileURL, in.Author, in.Status)
	if err != nil {
		adminFail(w, r, err)
		return
	}
	respond(w, r, 200, rpt)
}

// --- Emergency Resources (missing admin list/update/delete) ---
func (s *Server) listAdminEmergencyResources(w http.ResponseWriter, r *http.Request) {
	all, _, err := s.emergencySvc.ListResources(r.Context(), "", "", 1, 100000)
	if err != nil {
		fail(w, r, 500, fmt.Errorf("list emergency resources: %w", err))
		return
	}
	filtered := all
	// 关键词筛选（资源名称）
	if kw := strings.TrimSpace(r.URL.Query().Get("keyword")); kw != "" {
		tmp := make([]domain.EmergencyResource, 0, len(filtered))
		for _, res := range filtered {
			if strings.Contains(res.Name, kw) {
				tmp = append(tmp, res)
			}
		}
		filtered = tmp
	}
	// 资源类型筛选（res_type）
	if t := r.URL.Query().Get("res_type"); t != "" {
		tmp := make([]domain.EmergencyResource, 0, len(filtered))
		for _, res := range filtered {
			if res.ResType == t {
				tmp = append(tmp, res)
			}
		}
		filtered = tmp
	}
	paginatedRespond(w, r, filtered, len(filtered))
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
	r2, err := s.emergencySvc.UpdateResource(r.Context(), id, in.Name, in.ResType, in.Specs, in.Location, in.ContactInfo, in.Status, in.Quantity)
	if err != nil {
		adminFail(w, r, err)
		return
	}
	respond(w, r, 200, r2)
}
func (s *Server) deleteEmergencyResource(w http.ResponseWriter, r *http.Request) {
	if err := s.emergencySvc.DeleteResource(r.Context(), r.PathValue("id")); err != nil {
		adminFail(w, r, err)
		return
	}
	respond(w, r, 200, map[string]string{"deleted": "ok"})
}

// --- Emergency Dispatches (missing admin list/update/delete) ---
func (s *Server) listAdminEmergencyDispatches(w http.ResponseWriter, r *http.Request) {
	all, _, err := s.emergencySvc.ListDispatches(r.Context(), "", 1, 100000)
	if err != nil {
		fail(w, r, 500, fmt.Errorf("list dispatches: %w", err))
		return
	}
	filtered, _ := adminListFilter(all, r.URL.Query().Get("keyword"), r.URL.Query().Get("status"),
		func(d domain.EmergencyDispatch) string { return d.EventDesc },
		func(d domain.EmergencyDispatch) string { return d.Status })
	paginatedRespond(w, r, filtered, len(filtered))
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
	// P2 修复：严格解析日期——非法日期此前被 ParseTime 静默写成当前时间落库。
	startTime, ok := strictDate(w, r, in.StartTime)
	if !ok {
		return
	}
	endTime, ok := strictDate(w, r, in.EndTime)
	if !ok {
		return
	}
	d, err := s.emergencySvc.UpdateDispatch(r.Context(), id, in.ResourceID, in.EventDesc, in.Location, in.Commander, in.Result, in.Status, startTime, endTime)
	if err != nil {
		adminFail(w, r, err)
		return
	}
	respond(w, r, 200, d)
}
func (s *Server) deleteEmergencyDispatch(w http.ResponseWriter, r *http.Request) {
	if err := s.emergencySvc.DeleteDispatch(r.Context(), r.PathValue("id")); err != nil {
		adminFail(w, r, err)
		return
	}
	respond(w, r, 200, map[string]string{"deleted": "ok"})
}

// --- Messages (missing admin list/create/update/delete) ---
func (s *Server) listAdminMessages(w http.ResponseWriter, r *http.Request) {
	all, total, err := s.msgSvc.ListAll(r.Context(), 0, 100000)
	if err != nil {
		fail(w, r, 500, fmt.Errorf("list messages: %w", err))
		return
	}
	// 关键词过滤（标题/正文 contains，大小写不敏感）——通知量小，内存过滤即可。
	if kw := strings.TrimSpace(r.URL.Query().Get("keyword")); kw != "" {
		lower := strings.ToLower(kw)
		filtered := make([]domain.Message, 0, len(all))
		for _, m := range all {
			if strings.Contains(strings.ToLower(m.Title), lower) ||
				strings.Contains(strings.ToLower(m.Content), lower) {
				filtered = append(filtered, m)
			}
		}
		paginatedRespond(w, r, filtered, len(filtered))
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
	// 广播：receiver_id 留空 → 发给全部用户（管理员 + 企业 + 个人）
	// 此前仅发管理员——普通用户小程序消息中心永远收不到通知，流程未打通。
	if strings.TrimSpace(in.ReceiverID) == "" {
		sent, err := s.broadcastMessageToAll(r, in.SenderID, in.Title, in.Content, in.ResourceType, in.ResourceID)
		if err != nil {
			fail(w, r, 500, fmt.Errorf("broadcast messages: %w", err))
			return
		}
		respond(w, r, 201, map[string]any{"broadcast": len(sent), "messages": sent})
		return
	}
	msg, err := s.msgSvc.Send(r.Context(), in.SenderID, in.ReceiverID, in.Title, in.Content, in.ResourceType, in.ResourceID)
	if err != nil {
		adminFail(w, r, err)
		return
	}
	respond(w, r, 201, msg)
}

// broadcastMessageToAll sends one message to every platform user
// (admin + enterprise + individual), plus the current requester,
// and returns the messages created.
func (s *Server) broadcastMessageToAll(r *http.Request, senderID, title, content, resType, resID string) ([]domain.Message, error) {
	receivers := map[string]bool{}
	users, err := s.userRepo.All(r.Context())
	if err != nil {
		return nil, err
	}
	for _, u := range users {
		receivers[u.ID] = true
	}
	if a, ok := authenticatedActor(r); ok {
		receivers[a.ID] = true
	}
	sent := make([]domain.Message, 0, len(receivers))
	for rid := range receivers {
		m, err := s.msgSvc.Send(r.Context(), senderID, rid, title, content, resType, resID)
		if err != nil {
			return nil, err
		}
		sent = append(sent, m)
	}
	return sent, nil
}
func (s *Server) updateMessage(w http.ResponseWriter, r *http.Request) {
	// 该路由在 /api/v1/admin/ 前缀下（adminGate 已校验管理员角色）。
	// 先取消息再用收件人身份调用 MarkRead，与 C10 归属校验口径一致。
	m, err := s.msgSvc.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		fail(w, r, http.StatusNotFound, err)
		return
	}
	msg, err := s.msgSvc.MarkRead(r.Context(), m.ReceiverID, m.ID)
	if err != nil {
		adminFail(w, r, err)
		return
	}
	respond(w, r, http.StatusOK, msg)
}
func (s *Server) deleteMessage(w http.ResponseWriter, r *http.Request) {
	if err := s.msgSvc.Delete(r.Context(), r.PathValue("id")); err != nil {
		adminFail(w, r, err)
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
	d, err := s.complianceSvc.UpdateDoc(r.Context(), id, in.Title, in.Category, in.Publisher, in.PublishDate, in.Status, in.Summary, in.FileURL, in.Tags)
	if err != nil {
		adminFail(w, r, err)
		return
	}
	respond(w, r, http.StatusOK, d)
}
func (s *Server) deleteComplianceDoc(w http.ResponseWriter, r *http.Request) {
	if err := s.complianceSvc.DeleteDoc(r.Context(), r.PathValue("id")); err != nil {
		adminFail(w, r, err)
		return
	}
	respond(w, r, 200, map[string]string{"deleted": "ok"})
}

// --- Compliance Standards (missing update/delete) ---
func (s *Server) updateComplianceStandard(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var in struct {
		Title         string `json:"title"`
		Category      string `json:"category"`
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
	sd, err := s.complianceSvc.UpdateStandard(r.Context(), id, in.Title, in.Category, in.StandardNo, in.Publisher, in.EffectiveDate, in.Scope, in.Status, in.FileURL)
	if err != nil {
		adminFail(w, r, err)
		return
	}
	respond(w, r, 200, sd)
}
func (s *Server) deleteComplianceStandard(w http.ResponseWriter, r *http.Request) {
	if err := s.complianceSvc.DeleteStandard(r.Context(), r.PathValue("id")); err != nil {
		adminFail(w, r, err)
		return
	}
	respond(w, r, 200, map[string]string{"deleted": "ok"})
}

// --- Industry Resources (missing admin list/delete) ---
func (s *Server) listAdminResources(w http.ResponseWriter, r *http.Request) {
	all, _, err := s.resourceSvc.List(r.Context(), "", 1, 100000)
	if err != nil {
		fail(w, r, 500, fmt.Errorf("list resources: %w", err))
		return
	}
	filtered := all
	// 关键词筛选（资源名称）
	if kw := strings.TrimSpace(r.URL.Query().Get("keyword")); kw != "" {
		tmp := make([]domain.IndustryResource, 0, len(filtered))
		for _, res := range filtered {
			if strings.Contains(res.Name, kw) {
				tmp = append(tmp, res)
			}
		}
		filtered = tmp
	}
	// 资源类型筛选（res_type）
	if t := r.URL.Query().Get("res_type"); t != "" {
		tmp := make([]domain.IndustryResource, 0, len(filtered))
		for _, res := range filtered {
			if res.ResType == t {
				tmp = append(tmp, res)
			}
		}
		filtered = tmp
	}
	paginatedRespond(w, r, filtered, len(filtered))
}
func (s *Server) deleteIndustryResource(w http.ResponseWriter, r *http.Request) {
	if err := s.resourceSvc.Delete(r.Context(), r.PathValue("id")); err != nil {
		adminFail(w, r, err)
		return
	}
	respond(w, r, 200, map[string]string{"deleted": "ok"})
}

// --- RD Challenges (missing delete) ---
func (s *Server) deleteRDChallenge(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	if err := s.rdService.Delete(r.Context(), a, r.PathValue("id")); err != nil {
		adminFail(w, r, err)
		return
	}
	respond(w, r, 200, map[string]string{"deleted": "ok"})
}

// --- Research Projects (missing delete) ---
func (s *Server) deleteResearchProject(w http.ResponseWriter, r *http.Request) {
	if err := s.researchSvc.Delete(r.Context(), r.PathValue("id")); err != nil {
		adminFail(w, r, err)
		return
	}
	respond(w, r, 200, map[string]string{"deleted": "ok"})
}

func (s *Server) getCourse(w http.ResponseWriter, r *http.Request) {
	c, err := s.trainingSvc.GetCourse(r.Context(), r.PathValue("id"))
	if err != nil {
		fail(w, r, 404, err)
		return
	}
	// 非管理端不可通过详情接口查看未公开课程（待审核/草稿/已下架）——与列表过滤一致
	if !isAdminRequest(r) && isNonPublicStatus(c.Status) {
		fail(w, r, http.StatusNotFound, errors.New("course not found"))
		return
	}
	respond(w, r, 200, c)
}

func (s *Server) getCert(w http.ResponseWriter, r *http.Request) {
	c, err := s.trainingSvc.GetCert(r.Context(), r.PathValue("id"))
	if err != nil {
		fail(w, r, 404, err)
		return
	}
	respond(w, r, 200, c)
}

func (s *Server) getJob(w http.ResponseWriter, r *http.Request) {
	j, err := s.jobSvc.GetJob(r.Context(), r.PathValue("id"))
	if err != nil {
		fail(w, r, 404, err)
		return
	}
	// 非管理端不可通过公开详情接口查看未发布职位（draft/closed/archived）——
	// 与公开列表 ListPublished 的过滤一致，否则公开详情可绕过列表状态过滤读取草稿/已关闭职位。
	if !isAdminRequest(r) && j.Status != domain.JobPublished {
		fail(w, r, http.StatusNotFound, errors.New("job not found"))
		return
	}
	respond(w, r, 200, j)
}

func (s *Server) getCompetition(w http.ResponseWriter, r *http.Request) {
	c, err := s.competitionSvc.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		fail(w, r, 404, err)
		return
	}
	// 非管理端不可通过详情接口查看未公开赛事（待审核/草稿/已下架）——与列表过滤一致
	if !isAdminRequest(r) && isNonPublicStatus(c.Status) {
		fail(w, r, http.StatusNotFound, errors.New("competition not found"))
		return
	}
	respond(w, r, 200, c)
}

func (s *Server) getCollege(w http.ResponseWriter, r *http.Request) {
	c, err := s.collegeSvc.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		fail(w, r, 404, err)
		return
	}
	respond(w, r, 200, c)
}

func (s *Server) getStudyTour(w http.ResponseWriter, r *http.Request) {
	s2, err := s.studyTourRepo.FindByID(r.Context(), r.PathValue("id"))
	if err != nil {
		fail(w, r, 404, err)
		return
	}
	respond(w, r, 200, s2)
}

func (s *Server) getEvent(w http.ResponseWriter, r *http.Request) {
	e, err := s.eventSvc.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		fail(w, r, 404, err)
		return
	}
	respond(w, r, 200, e)
}

func (s *Server) getExhibition(w http.ResponseWriter, r *http.Request) {
	e, err := s.exhibitionSvc.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		fail(w, r, 404, err)
		return
	}
	respond(w, r, 200, e)
}

func (s *Server) getEmergencyResource(w http.ResponseWriter, r *http.Request) {
	r2, err := s.emergencySvc.GetResource(r.Context(), r.PathValue("id"))
	if err != nil {
		fail(w, r, 404, err)
		return
	}
	respond(w, r, 200, r2)
}

func (s *Server) getReport(w http.ResponseWriter, r *http.Request) {
	rpt, err := s.reportSvc.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		fail(w, r, 404, err)
		return
	}
	respond(w, r, 200, rpt)
}
func (s *Server) getIndustryResource(w http.ResponseWriter, r *http.Request) {
	r2, err := s.resourceSvc.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		fail(w, r, 404, err)
		return
	}
	respond(w, r, 200, r2)
}
func (s *Server) getPortfolio(w http.ResponseWriter, r *http.Request) {
	p, err := s.portfolioSvc.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		fail(w, r, 404, err)
		return
	}
	// 非管理端不可通过公开详情接口查看未发布名片（draft 等）——与公开列表
	// listPortfolios（ListPublished 仅返回 published）的过滤一致，防止草稿名片
	// （含 contact_info 联系方式）被公开详情接口绕过列表过滤读取。
	if !isAdminRequest(r) && p.Status != "published" {
		fail(w, r, http.StatusNotFound, errors.New("portfolio not found"))
		return
	}
	respond(w, r, 200, p)
}
func (s *Server) getExpert(w http.ResponseWriter, r *http.Request) {
	e, err := s.expertSvc.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		fail(w, r, 404, err)
		return
	}
	// 公开详情仅已发布专家；pending/archived 仅管理员可见（admin 路由复用本 handler）
	if e.Status != "published" {
		if a, ok := authenticatedActor(r); !ok ||
			(a.Role != domain.RolePlatformAdmin && a.Role != domain.RoleAssociationAdmin) {
			fail(w, r, http.StatusNotFound, errors.New("expert not found"))
			return
		}
	}
	respond(w, r, 200, e)
}
func (s *Server) getAchievement(w http.ResponseWriter, r *http.Request) {
	a, err := s.achievementSvc.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		fail(w, r, 404, err)
		return
	}
	respond(w, r, 200, a)
}
func (s *Server) getRDChallenge(w http.ResponseWriter, r *http.Request) {
	c2, err := s.rdService.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		fail(w, r, 404, err)
		return
	}
	respond(w, r, 200, c2)
}
func (s *Server) getResearchProject(w http.ResponseWriter, r *http.Request) {
	rp, err := s.researchSvc.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		fail(w, r, 404, err)
		return
	}
	respond(w, r, 200, rp)
}
func (s *Server) getTestSite(w http.ResponseWriter, r *http.Request) {
	ts, err := s.testSiteSvc.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		fail(w, r, 404, err)
		return
	}
	respond(w, r, 200, ts)
}
func (s *Server) getTransformation(w http.ResponseWriter, r *http.Request) {
	t, err := s.transSvc.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		fail(w, r, 404, err)
		return
	}
	respond(w, r, 200, t)
}
func (s *Server) getComplianceDoc(w http.ResponseWriter, r *http.Request) {
	cd, err := s.complianceSvc.FindDocByID(r.Context(), r.PathValue("id"))
	if err != nil {
		fail(w, r, 404, err)
		return
	}
	respond(w, r, 200, cd)
}
func (s *Server) getComplianceStandard(w http.ResponseWriter, r *http.Request) {
	cs, err := s.complianceSvc.FindStandardByID(r.Context(), r.PathValue("id"))
	if err != nil {
		fail(w, r, 404, err)
		return
	}
	respond(w, r, 200, cs)
}
func (s *Server) getMessage(w http.ResponseWriter, r *http.Request) {
	m, err := s.msgSvc.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		fail(w, r, 404, err)
		return
	}
	respond(w, r, 200, m)
}
func (s *Server) getEmergencyDispatch(w http.ResponseWriter, r *http.Request) {
	d, err := s.emergencySvc.FindDispatchByID(r.Context(), r.PathValue("id"))
	if err != nil {
		fail(w, r, 404, err)
		return
	}
	respond(w, r, 200, d)
}
func (s *Server) getOrder(w http.ResponseWriter, r *http.Request) {
	o, err := s.tradeSvc.FindByID(r.Context(), r.PathValue("id"))
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
	o, err := s.tradeSvc.Create(r.Context(), in.BuyerID, in.ProductID, in.SellerID, in.AmountFen)
	if err != nil {
		adminFail(w, r, err)
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
	o, err := s.tradeSvc.UpdateStatusAdmin(r.Context(), r.PathValue("id"), in.Status)
	if err != nil {
		adminFail(w, r, err)
		return
	}
	respond(w, r, 200, o)
}

// PUT /api/v1/admin/orders/{id}/aftersale — 售后单审核（同意退款 / 驳回），仅售后待审核单可审
func (s *Server) reviewAftersale(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Action string `json:"action"` // approve=同意退款 / reject=驳回
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, 400, err)
		return
	}
	switch in.Action {
	case "approve", "reject":
	default:
		fail(w, r, 400, fmt.Errorf("action 仅支持 approve / reject"))
		return
	}
	o, err := s.tradeSvc.ReviewAftersale(r.Context(), r.PathValue("id"), in.Action == "approve")
	if err != nil {
		fail(w, r, 400, err)
		return
	}
	adminID := ""
	if a, ok := authenticatedActor(r); ok {
		adminID = a.ID
	}
	s.audit(r.Context(), adminID, "review_aftersale", "trade_order", o.ID, o.AftersaleStatus)
	respond(w, r, 200, o)
}
func (s *Server) deleteOrder(w http.ResponseWriter, r *http.Request) {
	if err := s.tradeSvc.Delete(r.Context(), r.PathValue("id")); err != nil {
		adminFail(w, r, fmt.Errorf("delete order: %w", err))
		return
	}
	if a, ok := authenticatedActor(r); ok {
		s.audit(r.Context(), a.ID, "delete_order", "trade_order", r.PathValue("id"), "deleted")
	}
	respond(w, r, 200, map[string]string{"deleted": "ok"})
}
