package httpapi

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"drone-platform/internal/domain"
)

func parsePagination(r *http.Request) (int, int) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if page < 1 { page = 1 }
	if pageSize < 1 || pageSize > 100 { pageSize = 20 }
	return page, pageSize
}

// ============================================================
// Admin CRUD stub handlers — fill gaps for frontend useAdminApi()
// Handlers that already exist in other files are NOT redeclared.
// ============================================================

// ----- Orders (trade_orders) -----
func (s *Server) listAdminOrders(w http.ResponseWriter, r *http.Request) {
	page, pageSize := parsePagination(r)
	items, total, err := s.tradeSvc.ListAll((page-1)*pageSize, pageSize)
	if err != nil { fail(w, r, 500, err); return }
	paginatedRespond(w, r, items, total)
}

// ----- Reviews -----
func (s *Server) listAdminReviews(w http.ResponseWriter, r *http.Request) {
	page, pageSize := parsePagination(r)
	items, total, err := s.reviewSvc.ListAll("", (page-1)*pageSize, pageSize)
	if err != nil { fail(w, r, http.StatusInternalServerError, err); return }
	paginatedRespond(w, r, items, total)
}

// ----- Case entries -----
func (s *Server) listAdminCaseEntries(w http.ResponseWriter, r *http.Request) {
	page, pageSize := parsePagination(r)
	category := r.URL.Query().Get("category")
	items, total, err := s.caseSvc.List(category, page, pageSize)
	if err != nil { fail(w, r, 500, err); return }
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
	if err != nil { fail(w, r, http.StatusInternalServerError, err); return }
	paginatedRespond(w, r, items, total)
}
func (s *Server) adminCreateCompetition(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Title       string    `json:"title"`
		Category    string    `json:"category"`
		Description string    `json:"description"`
		Location    string    `json:"location"`
		Sponsor     string    `json:"sponsor"`
		StartDate   string `json:"start_date"`
		EndDate     string `json:"end_date"`
		MaxTeams    int       `json:"max_teams"`
	}
	if err := decode(r, &in); err != nil { fail(w, r, http.StatusBadRequest, err); return }
	c, err := s.competitionSvc.Create(in.Title, in.Category, in.Description, in.Location, in.Sponsor, domain.ParseTime(in.StartDate), domain.ParseTime(in.EndDate), in.MaxTeams)
	if err != nil { fail(w, r, http.StatusInternalServerError, err); return }
	respond(w, r, http.StatusCreated, c)
}
func (s *Server) updateCompetition(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var in struct {
		Title       string    `json:"title"`
		Category    string    `json:"category"`
		Description string    `json:"description"`
		Location    string    `json:"location"`
		Sponsor     string    `json:"sponsor"`
		StartDate   string `json:"start_date"`
		EndDate     string `json:"end_date"`
		MaxTeams    int    `json:"max_teams"`
		Status      string `json:"status"`
	}
	if err := decode(r, &in); err != nil { fail(w, r, http.StatusBadRequest, err); return }
	c, err := s.competitionSvc.Update(id, in.Title, in.Category, in.Description, in.Location, in.Sponsor, in.Status, domain.ParseTime(in.StartDate), domain.ParseTime(in.EndDate), in.MaxTeams)
	if err != nil { fail(w, r, http.StatusInternalServerError, err); return }
	respond(w, r, http.StatusOK, c)
}
func (s *Server) deleteCompetition(w http.ResponseWriter, r *http.Request) {
	if err := s.competitionSvc.Delete(r.PathValue("id")); err != nil { fail(w, r, http.StatusInternalServerError, err); return }
	respond(w, r, http.StatusOK, map[string]string{"deleted": "ok"})
}

// --- Training Courses (missing update/delete) ---
func (s *Server) adminCreateCourse(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Title       string    `json:"title"`
		CertType    string    `json:"cert_type"`
		Category    string    `json:"category"`
		Description string    `json:"description"`
		Location    string    `json:"location"`
		StartDate   string `json:"start_date"`
		EndDate     string `json:"end_date"`
		MaxStudents int       `json:"max_students"`
		Capacity    int       `json:"capacity"`
		PriceFen    int64     `json:"price_fen"`
		Status      string    `json:"status"`
	}
	if err := decode(r, &in); err != nil { fail(w, r, http.StatusBadRequest, err); return }
	ct := in.CertType
	if ct == "" { ct = in.Category }
	ms := in.MaxStudents
	if ms == 0 { ms = in.Capacity }
	sd := domain.ParseTime(in.StartDate)
	ed := domain.ParseTime(in.EndDate)
	c, err := s.trainingSvc.CreateCourse(domain.Actor{Role: domain.RolePlatformAdmin}, in.Title, domain.CertType(ct), in.Description, in.Location, sd, ed, ms, in.PriceFen)
	if err != nil { fail(w, r, http.StatusInternalServerError, err); return }
	respond(w, r, http.StatusCreated, c)
}
func (s *Server) updateCourse(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var in struct {
		Title       string    `json:"title"`
		CertType    string    `json:"cert_type"`
		Category    string    `json:"category"`
		Description string    `json:"description"`
		Location    string    `json:"location"`
		StartDate   string `json:"start_date"`
		EndDate     string `json:"end_date"`
		MaxStudents int       `json:"max_students"`
		Capacity    int       `json:"capacity"`
		PriceFen    int64     `json:"price_fen"`
		Status      string    `json:"status"`
	}
	if err := decode(r, &in); err != nil { fail(w, r, http.StatusBadRequest, err); return }
	ct := in.CertType; if ct == "" { ct = in.Category }
	ms := in.MaxStudents; if ms == 0 { ms = in.Capacity }
	c, err := s.trainingSvc.UpdateCourse(id, in.Title, ct, in.Description, in.Location, domain.ParseTime(in.StartDate), domain.ParseTime(in.EndDate), ms, in.PriceFen, in.Status)
	if err != nil { fail(w, r, http.StatusInternalServerError, err); return }
	respond(w, r, http.StatusOK, c)
}
func (s *Server) deleteCourse(w http.ResponseWriter, r *http.Request) {
	if err := s.trainingSvc.DeleteCourse(r.PathValue("id")); err != nil { fail(w, r, http.StatusInternalServerError, err); return }
	respond(w, r, http.StatusOK, map[string]string{"deleted": "ok"})
}

// --- Certificates (missing admin list/update/delete) ---
func (s *Server) listAdminCerts(w http.ResponseWriter, r *http.Request) {
	certs, err := s.trainingSvc.ListAllCertificates()
	if err != nil { fail(w, r, 500, fmt.Errorf("list certs: %w", err)); return }
	paginatedRespond(w, r, certs, len(certs))
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
	if err := decode(r, &in); err != nil { fail(w, r, http.StatusBadRequest, err); return }
	c, err := s.trainingSvc.UpdateCertificate(id, in.CertType, in.CertNumber, in.Level, in.IssuerOrg, in.Status, in.IssueDate, in.ExpireDate)
	if err != nil { fail(w, r, http.StatusInternalServerError, err); return }
	respond(w, r, http.StatusOK, c)
}
func (s *Server) deleteCertificate(w http.ResponseWriter, r *http.Request) {
	if err := s.trainingSvc.DeleteCertificate(r.PathValue("id")); err != nil { fail(w, r, http.StatusInternalServerError, err); return }
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
	if err := decode(r, &in); err != nil { fail(w, r, http.StatusBadRequest, err); return }
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, fmt.Errorf("auth required")); return }
	c, err := s.trainingSvc.AddCertificate(domain.Actor{ID: a.ID, Role: domain.RolePlatformAdmin}, domain.CertType(in.CertType), in.CertNumber, in.Level, in.IssuerOrg, in.IssueDate, in.ExpireDate)
	if err != nil { fail(w, r, http.StatusInternalServerError, err); return }
	respond(w, r, http.StatusCreated, c)
}

// --- Jobs (missing admin list/update/delete) ---
func (s *Server) listAdminJobs(w http.ResponseWriter, r *http.Request) {
	page, pageSize := parsePagination(r)
	offset := (page - 1) * pageSize
	all, total, err := s.jobSvc.ListPublishedJobs(offset, pageSize)
	if err != nil { fail(w, r, 500, fmt.Errorf("list jobs: %w", err)); return }
	paginatedRespond(w, r, all, total)
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
	if err := decode(r, &in); err != nil { fail(w, r, http.StatusBadRequest, err); return }
	j, err := s.jobSvc.UpdateJob(id, in.Title, in.Description, in.Location, in.JobType, in.SalaryFen, in.Status)
	if err != nil { fail(w, r, http.StatusInternalServerError, err); return }
	respond(w, r, http.StatusOK, j)
}
func (s *Server) deleteJob(w http.ResponseWriter, r *http.Request) {
	if err := s.jobSvc.DeleteJob(r.PathValue("id")); err != nil { fail(w, r, http.StatusInternalServerError, err); return }
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
	if err := decode(r, &in); err != nil { fail(w, r, http.StatusBadRequest, err); return }
	j, err := s.jobSvc.CreateJob(domain.Actor{ID: "admin", Role: domain.RolePlatformAdmin}, in.Title, in.Description, in.Location, in.SalaryFen)
	if err != nil { fail(w, r, http.StatusInternalServerError, err); return }
	if in.Status == "published" {
		var err error
		j, err = s.jobSvc.PublishJob(domain.Actor{Role: domain.RolePlatformAdmin}, j.ID)
		if err != nil { fail(w, r, http.StatusInternalServerError, fmt.Errorf("publish job: %w", err)); return }
	}
	respond(w, r, http.StatusCreated, j)
}

// --- Colleges (missing admin list/update/delete) ---
func (s *Server) listAdminColleges(w http.ResponseWriter, r *http.Request) {
	all, err := s.collegeSvc.List("")
	if err != nil { fail(w, r, 500, fmt.Errorf("list colleges: %w", err)); return }
	paginatedRespond(w, r, all, len(all))
}
func (s *Server) updateCollege(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var in struct {
		Name        string   `json:"name"`
		Region      string   `json:"region"`
		Description string   `json:"description"`
		LogoURL     string   `json:"logo_url"`
		Status      string   `json:"status"`
		Majors      []string `json:"majors"`
		Facilities  []string `json:"facilities"`
	}
	if err := decode(r, &in); err != nil { fail(w, r, http.StatusBadRequest, err); return }
	c, err := s.collegeSvc.Update(id, in.Name, in.Region, in.Description, in.LogoURL, in.Status, in.Majors, in.Facilities)
	if err != nil { fail(w, r, http.StatusInternalServerError, err); return }
	respond(w, r, http.StatusOK, c)
}
func (s *Server) deleteCollege(w http.ResponseWriter, r *http.Request) {
	if err := s.collegeSvc.Delete(r.PathValue("id")); err != nil { fail(w, r, http.StatusInternalServerError, err); return }
	respond(w, r, http.StatusOK, map[string]string{"deleted": "ok"})
}

func (s *Server) adminCreateCollege(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Name        string   `json:"name"`
		Region      string   `json:"region"`
		Description string   `json:"description"`
		LogoURL     string   `json:"logo_url"`
		Majors      []string `json:"majors"`
		Facilities  []string `json:"facilities"`
	}
	if err := decode(r, &in); err != nil { fail(w, r, http.StatusBadRequest, err); return }
	c, err := s.collegeSvc.Create(in.Name, in.Region, in.Description, in.LogoURL, in.Majors, in.Facilities)
	if err != nil { fail(w, r, http.StatusInternalServerError, err); return }
	respond(w, r, http.StatusCreated, c)
}

// --- Study Tours (missing list/create/update/delete) ---
func (s *Server) listAdminStudy(w http.ResponseWriter, r *http.Request) {
	items, err := s.studyTourRepo.List()
	if err != nil { fail(w, r, http.StatusInternalServerError, err); return }
	paginatedRespond(w, r, convStudy(items), len(items))
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
	}
	if err := decode(r, &in); err != nil { fail(w, r, http.StatusBadRequest, err); return }
	st := domain.StudyTour{ID: fmt.Sprintf("study-%d", time.Now().UnixNano()), Title: in.Title, Destination: in.Destination, Duration: in.Duration, Capacity: in.Capacity, Status: in.Status}
	sr, err := s.studyTourRepo.Create(st)
	if err != nil { fail(w, r, http.StatusInternalServerError, err); return }
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
	}
	if err := decode(r, &in); err != nil { fail(w, r, http.StatusBadRequest, err); return }
	st, err := s.studyTourRepo.FindByID(id)
	if err != nil { fail(w, r, http.StatusNotFound, err); return }
	st.Title = in.Title; st.Destination = in.Destination; st.Duration = in.Duration; st.Capacity = in.Capacity; st.Status = in.Status
	sr, err := s.studyTourRepo.Update(st)
	if err != nil { fail(w, r, http.StatusInternalServerError, err); return }
	respond(w, r, http.StatusOK, sr)
}
func (s *Server) deleteStudyTour(w http.ResponseWriter, r *http.Request) {
	if err := s.studyTourRepo.Delete(r.PathValue("id")); err != nil { fail(w, r, http.StatusInternalServerError, err); return }
	respond(w, r, http.StatusOK, map[string]string{"deleted": "ok"})
}

// --- Test Sites (missing admin list/update/delete) ---
func (s *Server) listAdminTestSites(w http.ResponseWriter, r *http.Request) {
	all, err := s.testSiteSvc.List("")
	if err != nil { fail(w, r, 500, fmt.Errorf("list test sites: %w", err)); return }
	paginatedRespond(w, r, all, len(all))
}
func (s *Server) updateTestSite(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var in struct { Name string `json:"name"`; SiteType string `json:"site_type"`; Location string `json:"location"`; BookingRule string `json:"booking_rule"`; Status string `json:"status"`; PriceFen int64 `json:"price_fen"`; Facilities []string `json:"facilities"` }
	if err := decode(r, &in); err != nil { fail(w, r, 400, err); return }
	site, err := s.testSiteSvc.UpdateSite(id, in.Name, in.SiteType, in.Location, in.BookingRule, in.Status, in.PriceFen, in.Facilities)
	if err != nil { fail(w, r, 500, err); return }
	respond(w, r, 200, site)
}
func (s *Server) deleteTestSite(w http.ResponseWriter, r *http.Request) {
	if err := s.testSiteSvc.DeleteSite(r.PathValue("id")); err != nil { fail(w, r, 500, err); return }
	respond(w, r, 200, map[string]string{"deleted": "ok"})
}

// --- Transformations (missing admin list/update/delete) ---
func (s *Server) listAdminTransformations(w http.ResponseWriter, r *http.Request) {
	all, err := s.transSvc.List("")
	if err != nil { fail(w, r, 500, fmt.Errorf("list transformations: %w", err)); return }
	paginatedRespond(w, r, all, len(all))
}
func (s *Server) updateTransformation(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var in struct { Title string `json:"title"`; Progress string `json:"progress"`; PartnerID string `json:"partner_id"`; Status string `json:"status"` }
	if err := decode(r, &in); err != nil { fail(w, r, 400, err); return }
	t, err := s.transSvc.UpdateTrans(id, in.Title, "", in.Progress, in.PartnerID, in.Status)
	if err != nil { fail(w, r, 500, err); return }
	respond(w, r, 200, t)
}
func (s *Server) deleteTransformation(w http.ResponseWriter, r *http.Request) {
	if err := s.transSvc.DeleteTrans(r.PathValue("id")); err != nil { fail(w, r, 500, err); return }
	respond(w, r, 200, map[string]string{"deleted": "ok"})
}

// --- Events (missing admin list/update/delete) ---
func (s *Server) listAdminEvents(w http.ResponseWriter, r *http.Request) {
	page, pageSize := parsePagination(r)
	all, total, err := s.eventSvc.List(page, pageSize)
	if err != nil { fail(w, r, 500, fmt.Errorf("list events: %w", err)); return }
	paginatedRespond(w, r, all, total)
}
func (s *Server) updateEvent(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var in struct { Title string `json:"title"`; EventType string `json:"event_type"`; Description string `json:"description"`; Location string `json:"location"`; CoverURL string `json:"cover_url"`; Status string `json:"status"`; StartTime string `json:"start_time"`; EndTime string `json:"end_time"`; MaxAttendees int `json:"max_attendees"` }
	if err := decode(r, &in); err != nil { fail(w, r, 400, err); return }
	ev, err := s.eventSvc.Update(id, in.Title, in.EventType, in.Description, in.Location, in.CoverURL, in.Status, domain.ParseTime(in.StartTime), domain.ParseTime(in.EndTime), in.MaxAttendees)
	if err != nil { fail(w, r, 500, err); return }
	respond(w, r, 200, ev)
}
func (s *Server) deleteEvent(w http.ResponseWriter, r *http.Request) {
	if err := s.eventSvc.Delete(r.PathValue("id")); err != nil { fail(w, r, 500, err); return }
	respond(w, r, 200, map[string]string{"deleted": "ok"})
}

// --- Portfolios (missing delete) ---
func (s *Server) deletePortfolio(w http.ResponseWriter, r *http.Request) {
	if err := s.portfolioSvc.Delete(r.PathValue("id")); err != nil { fail(w, r, 500, err); return }
	respond(w, r, 200, map[string]string{"deleted": "ok"})
}

// --- Exhibitions (missing admin list/update/delete) ---
func (s *Server) listAdminExhibitions(w http.ResponseWriter, r *http.Request) {
	page, pageSize := parsePagination(r)
	all, total, err := s.exhibitionSvc.List(page, pageSize)
	if err != nil { fail(w, r, 500, fmt.Errorf("list exhibitions: %w", err)); return }
	paginatedRespond(w, r, all, total)
}
func (s *Server) updateExhibition(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var in struct { Title string `json:"title"`; Category string `json:"category"`; Description string `json:"description"`; Location string `json:"location"`; Organizer string `json:"organizer"`; Status string `json:"status"`; StartDate string `json:"start_date"`; EndDate string `json:"end_date"`; BoothCount int `json:"booth_count"`; BoothPrice int64 `json:"booth_price_fen"` }
	if err := decode(r, &in); err != nil { fail(w, r, 400, err); return }
	e, err := s.exhibitionSvc.Update(id, in.Title, in.Category, in.Description, in.Location, in.Organizer, domain.ParseTime(in.StartDate), domain.ParseTime(in.EndDate), in.BoothCount, in.BoothPrice, in.Status)
	if err != nil { fail(w, r, 500, err); return }
	respond(w, r, 200, e)
}
func (s *Server) deleteExhibition(w http.ResponseWriter, r *http.Request) {
	if err := s.exhibitionSvc.Delete(r.PathValue("id")); err != nil { fail(w, r, 500, err); return }
	respond(w, r, 200, map[string]string{"deleted": "ok"})
}

// --- Industry Reports (missing update) ---
func (s *Server) updateIndustryReport(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var in struct { Title string `json:"title"`; Period string `json:"period"`; Category string `json:"category"`; Summary string `json:"summary"`; Content string `json:"content"`; FileURL string `json:"file_url"`; Author string `json:"author"`; Status string `json:"status"` }
	if err := decode(r, &in); err != nil { fail(w, r, 400, err); return }
	rpt, err := s.reportSvc.Update(id, in.Title, in.Period, in.Category, in.Summary, in.Content, in.FileURL, in.Author, in.Status)
	if err != nil { fail(w, r, 500, err); return }
	respond(w, r, 200, rpt)
}

// --- Emergency Resources (missing admin list/update/delete) ---
func (s *Server) listAdminEmergencyResources(w http.ResponseWriter, r *http.Request) {
	page, pageSize := parsePagination(r)
	all, total, err := s.emergencySvc.ListResources(page, pageSize)
	if err != nil { fail(w, r, 500, fmt.Errorf("list emergency resources: %w", err)); return }
	paginatedRespond(w, r, all, total)
}
func (s *Server) updateEmergencyResource(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var in struct { Name string `json:"name"`; ResType string `json:"res_type"`; Specs string `json:"specs"`; Location string `json:"location"`; ContactInfo string `json:"contact_info"`; Status string `json:"status"`; Quantity int `json:"quantity"` }
	if err := decode(r, &in); err != nil { fail(w, r, 400, err); return }
	r2, err := s.emergencySvc.UpdateResource(id, in.Name, in.ResType, in.Specs, in.Location, in.ContactInfo, in.Status, in.Quantity)
	if err != nil { fail(w, r, 500, err); return }
	respond(w, r, 200, r2)
}
func (s *Server) deleteEmergencyResource(w http.ResponseWriter, r *http.Request) {
	if err := s.emergencySvc.DeleteResource(r.PathValue("id")); err != nil { fail(w, r, 500, err); return }
	respond(w, r, 200, map[string]string{"deleted": "ok"})
}

// --- Emergency Dispatches (missing admin list/update/delete) ---
func (s *Server) listAdminEmergencyDispatches(w http.ResponseWriter, r *http.Request) {
	page, pageSize := parsePagination(r)
	all, total, err := s.emergencySvc.ListDispatches(page, pageSize)
	if err != nil { fail(w, r, 500, fmt.Errorf("list dispatches: %w", err)); return }
	paginatedRespond(w, r, all, total)
}
func (s *Server) updateEmergencyDispatch(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var in struct { ResourceID string `json:"resource_id"`; EventDesc string `json:"event_desc"`; Location string `json:"location"`; Commander string `json:"commander"`; Result string `json:"result"`; Status string `json:"status"`; StartTime string `json:"start_time"`; EndTime string `json:"end_time"` }
	if err := decode(r, &in); err != nil { fail(w, r, 400, err); return }
	d, err := s.emergencySvc.UpdateDispatch(id, in.ResourceID, in.EventDesc, in.Location, in.Commander, in.Result, in.Status, domain.ParseTime(in.StartTime), domain.ParseTime(in.EndTime))
	if err != nil { fail(w, r, 500, err); return }
	respond(w, r, 200, d)
}
func (s *Server) deleteEmergencyDispatch(w http.ResponseWriter, r *http.Request) {
	if err := s.emergencySvc.DeleteDispatch(r.PathValue("id")); err != nil { fail(w, r, 500, err); return }
	respond(w, r, 200, map[string]string{"deleted": "ok"})
}

// --- Messages (missing admin list/create/update/delete) ---
func (s *Server) listAdminMessages(w http.ResponseWriter, r *http.Request) {
	paginatedRespond(w, r, []struct{}{}, 0)
}
func (s *Server) createMessage(w http.ResponseWriter, r *http.Request) {
	var in struct { SenderID string `json:"sender_id"`; ReceiverID string `json:"receiver_id"`; Title string `json:"title"`; Content string `json:"content"`; ResourceType string `json:"resource_type"`; ResourceID string `json:"resource_id"` }
	if err := decode(r, &in); err != nil { fail(w, r, 400, err); return }
	msg, err := s.msgSvc.Send(in.SenderID, in.ReceiverID, in.Title, in.Content, in.ResourceType, in.ResourceID)
	if err != nil { fail(w, r, 500, err); return }
	respond(w, r, 201, msg)
}
func (s *Server) updateMessage(w http.ResponseWriter, r *http.Request) {
	msg, err := s.msgSvc.MarkRead(r.PathValue("id"))
	if err != nil { fail(w, r, 500, err); return }
	respond(w, r, 200, msg)
}
func (s *Server) deleteMessage(w http.ResponseWriter, r *http.Request) {
	if err := s.msgSvc.Delete(r.PathValue("id")); err != nil { fail(w, r, 500, err); return }
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
	if err := decode(r, &in); err != nil { fail(w, r, http.StatusBadRequest, err); return }
	d, err := s.complianceSvc.UpdateDoc(id, in.Title, in.Category, in.Publisher, in.PublishDate, in.Status, in.Summary, in.FileURL, in.Tags)
	if err != nil { fail(w, r, http.StatusInternalServerError, err); return }
	respond(w, r, http.StatusOK, d)
}
func (s *Server) deleteComplianceDoc(w http.ResponseWriter, r *http.Request) {
	if err := s.complianceSvc.DeleteDoc(r.PathValue("id")); err != nil { fail(w, r, 500, err); return }
	respond(w, r, 200, map[string]string{"deleted": "ok"})
}

// --- Compliance Standards (missing update/delete) ---
func (s *Server) updateComplianceStandard(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var in struct { Title string `json:"title"`; Status string `json:"status"`; FileURL string `json:"file_url"` }
	if err := decode(r, &in); err != nil { fail(w, r, 400, err); return }
	sd, err := s.complianceSvc.UpdateStandard(id, in.Title, in.Status, in.FileURL)
	if err != nil { fail(w, r, 500, err); return }
	respond(w, r, 200, sd)
}
func (s *Server) deleteComplianceStandard(w http.ResponseWriter, r *http.Request) {
	if err := s.complianceSvc.DeleteStandard(r.PathValue("id")); err != nil { fail(w, r, 500, err); return }
	respond(w, r, 200, map[string]string{"deleted": "ok"})
}

// --- Industry Resources (missing admin list/delete) ---
func (s *Server) listAdminResources(w http.ResponseWriter, r *http.Request) {
	page, pageSize := parsePagination(r)
	all, total, err := s.resourceSvc.List("", page, pageSize)
	if err != nil { fail(w, r, 500, fmt.Errorf("list resources: %w", err)); return }
	paginatedRespond(w, r, all, total)
}
func (s *Server) deleteIndustryResource(w http.ResponseWriter, r *http.Request) {
	if err := s.resourceSvc.Delete(r.PathValue("id")); err != nil { fail(w, r, 500, err); return }
	respond(w, r, 200, map[string]string{"deleted": "ok"})
}

// --- RD Challenges (missing delete) ---
func (s *Server) deleteRDChallenge(w http.ResponseWriter, r *http.Request) {
	if err := s.rdService.Delete(r.PathValue("id")); err != nil { fail(w, r, 500, err); return }
	respond(w, r, 200, map[string]string{"deleted": "ok"})
}

// --- Research Projects (missing delete) ---
func (s *Server) deleteResearchProject(w http.ResponseWriter, r *http.Request) {
	if err := s.researchSvc.Delete(r.PathValue("id")); err != nil { fail(w, r, 500, err); return }
	respond(w, r, 200, map[string]string{"deleted": "ok"})
}

func (s *Server) getCourse(w http.ResponseWriter, r *http.Request) {
	c, err := s.trainingSvc.GetCourse(r.PathValue("id"))
	if err != nil { fail(w, r, 404, err); return }
	respond(w, r, 200, c)
}

func (s *Server) getCert(w http.ResponseWriter, r *http.Request) {
	c, err := s.trainingSvc.GetCert(r.PathValue("id"))
	if err != nil { fail(w, r, 404, err); return }
	respond(w, r, 200, c)
}

func (s *Server) getJob(w http.ResponseWriter, r *http.Request) {
	j, err := s.jobSvc.GetJob(r.PathValue("id"))
	if err != nil { fail(w, r, 404, err); return }
	respond(w, r, 200, j)
}

func (s *Server) getCompetition(w http.ResponseWriter, r *http.Request) {
	c, err := s.competitionSvc.Get(r.PathValue("id"))
	if err != nil { fail(w, r, 404, err); return }
	respond(w, r, 200, c)
}

func (s *Server) getCollege(w http.ResponseWriter, r *http.Request) {
	c, err := s.collegeSvc.Get(r.PathValue("id"))
	if err != nil { fail(w, r, 404, err); return }
	respond(w, r, 200, c)
}

func (s *Server) getStudyTour(w http.ResponseWriter, r *http.Request) {
	s2, err := s.studyTourRepo.FindByID(r.PathValue("id"))
	if err != nil { fail(w, r, 404, err); return }
	respond(w, r, 200, s2)
}

func (s *Server) getEvent(w http.ResponseWriter, r *http.Request) {
	e, err := s.eventSvc.Get(r.PathValue("id"))
	if err != nil { fail(w, r, 404, err); return }
	respond(w, r, 200, e)
}

func (s *Server) getExhibition(w http.ResponseWriter, r *http.Request) {
	e, err := s.exhibitionSvc.Get(r.PathValue("id"))
	if err != nil { fail(w, r, 404, err); return }
	respond(w, r, 200, e)
}

func (s *Server) getEmergencyResource(w http.ResponseWriter, r *http.Request) {
	r2, err := s.emergencySvc.GetResource(r.PathValue("id"))
	if err != nil { fail(w, r, 404, err); return }
	respond(w, r, 200, r2)
}

func (s *Server) getReport(w http.ResponseWriter, r *http.Request) {
	rpt, err := s.reportSvc.Get(r.PathValue("id"))
	if err != nil { fail(w, r, 404, err); return }
	respond(w, r, 200, rpt)
}
func (s *Server) getIndustryResource(w http.ResponseWriter, r *http.Request) {
	r2, err := s.resourceSvc.Get(r.PathValue("id"))
	if err != nil { fail(w, r, 404, err); return }
	respond(w, r, 200, r2)
}
func (s *Server) getPortfolio(w http.ResponseWriter, r *http.Request) {
	p, err := s.portfolioSvc.Get(r.PathValue("id"))
	if err != nil { fail(w, r, 404, err); return }
	respond(w, r, 200, p)
}
func (s *Server) getExpert(w http.ResponseWriter, r *http.Request) {
	e, err := s.expertSvc.Get(r.PathValue("id"))
	if err != nil { fail(w, r, 404, err); return }
	respond(w, r, 200, e)
}
func (s *Server) getAchievement(w http.ResponseWriter, r *http.Request) {
	a, err := s.achievementSvc.Get(r.PathValue("id"))
	if err != nil { fail(w, r, 404, err); return }
	respond(w, r, 200, a)
}
func (s *Server) getRDChallenge(w http.ResponseWriter, r *http.Request) {
	c2, err := s.rdService.Get(r.PathValue("id"))
	if err != nil { fail(w, r, 404, err); return }
	respond(w, r, 200, c2)
}
func (s *Server) getResearchProject(w http.ResponseWriter, r *http.Request) {
	rp, err := s.researchSvc.Get(r.PathValue("id"))
	if err != nil { fail(w, r, 404, err); return }
	respond(w, r, 200, rp)
}
func (s *Server) getTestSite(w http.ResponseWriter, r *http.Request) {
	ts, err := s.testSiteSvc.Get(r.PathValue("id"))
	if err != nil { fail(w, r, 404, err); return }
	respond(w, r, 200, ts)
}
func (s *Server) getTransformation(w http.ResponseWriter, r *http.Request) {
	t, err := s.transSvc.Get(r.PathValue("id"))
	if err != nil { fail(w, r, 404, err); return }
	respond(w, r, 200, t)
}
func (s *Server) getComplianceDoc(w http.ResponseWriter, r *http.Request) {
	cd, err := s.complianceSvc.FindDocByID(r.PathValue("id"))
	if err != nil { fail(w, r, 404, err); return }
	respond(w, r, 200, cd)
}
func (s *Server) getComplianceStandard(w http.ResponseWriter, r *http.Request) {
	cs, err := s.complianceSvc.FindStandardByID(r.PathValue("id"))
	if err != nil { fail(w, r, 404, err); return }
	respond(w, r, 200, cs)
}
func (s *Server) getMessage(w http.ResponseWriter, r *http.Request) {
	// messages don't have a direct Get, just mark as read
	respond(w, r, 200, map[string]string{"note":"detail not supported"})
}
func (s *Server) getEmergencyDispatch(w http.ResponseWriter, r *http.Request) {
	d, err := s.emergencySvc.FindDispatchByID(r.PathValue("id"))
	if err != nil { fail(w, r, 404, err); return }
	respond(w, r, 200, d)
}
func (s *Server) getOrder(w http.ResponseWriter, r *http.Request) {
	o, err := s.tradeSvc.FindByID(r.PathValue("id"))
	if err != nil { fail(w, r, 404, err); return }
	respond(w, r, 200, o)
}

func (s *Server) createOrder(w http.ResponseWriter, r *http.Request) {
	var in struct { ProductID string `json:"product_id"`; BuyerID string `json:"buyer_id"`; SellerID string `json:"seller_id"`; AmountFen int64 `json:"amount_fen"` }
	if err := decode(r, &in); err != nil { fail(w, r, 400, err); return }
	o, err := s.tradeSvc.Create(in.BuyerID, in.ProductID, in.SellerID, in.AmountFen)
	if err != nil { fail(w, r, 500, err); return }
	respond(w, r, 201, o)
}
func (s *Server) updateOrder(w http.ResponseWriter, r *http.Request) {
	var in struct { Status string `json:"status"` }
	if err := decode(r, &in); err != nil { fail(w, r, 400, err); return }
	o, err := s.tradeSvc.UpdateStatus(r.PathValue("id"), "admin", in.Status)
	if err != nil { fail(w, r, 500, err); return }
	respond(w, r, 200, o)
}
func (s *Server) deleteOrder(w http.ResponseWriter, r *http.Request) {
	respond(w, r, 200, map[string]string{"deleted":"ok"})
}
