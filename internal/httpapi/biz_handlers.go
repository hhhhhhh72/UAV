package httpapi

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"drone-platform/internal/domain"
)

// registerBizRoutes registers all business-module routes.
func (s *Server) registerBizRoutes(mux *http.ServeMux) {
	// ---- Experts ----
	mux.HandleFunc("GET /api/v1/experts", s.listExperts)
	mux.HandleFunc("POST /api/v1/admin/experts", s.createExpert)
	mux.HandleFunc("PUT /api/v1/admin/experts/{id}", s.updateExpert)
	mux.HandleFunc("DELETE /api/v1/admin/experts/{id}", s.deleteExpert)

	// ---- Cases ----
	mux.HandleFunc("GET /api/v1/cases", s.listCases)
	mux.HandleFunc("GET /api/v1/cases/{id}", s.getCase)
	mux.HandleFunc("POST /api/v1/admin/cases", s.createCase)
	mux.HandleFunc("GET /api/v1/admin/cases/{id}", s.getCase)
	mux.HandleFunc("PUT /api/v1/admin/cases/{id}", s.updateCase)
	mux.HandleFunc("DELETE /api/v1/admin/cases/{id}", s.deleteCase)

	// ---- Compliance ----
	mux.HandleFunc("GET /api/v1/compliance-docs", s.listComplianceDocs)
	mux.HandleFunc("POST /api/v1/admin/compliance-docs", s.createComplianceDoc)
	mux.HandleFunc("GET /api/v1/compliance-standards", s.listComplianceStandards)
	mux.HandleFunc("POST /api/v1/admin/compliance-standards", s.createComplianceStandard)

	// ---- Industry Reports ----
	mux.HandleFunc("GET /api/v1/industry-reports", s.listIndustryReports)
	mux.HandleFunc("POST /api/v1/admin/industry-reports", s.createIndustryReport)
	mux.HandleFunc("DELETE /api/v1/admin/industry-reports/{id}", s.deleteIndustryReport)

	// ---- Portfolio ----
	mux.HandleFunc("GET /api/v1/portfolios", s.listPortfolios)
	mux.HandleFunc("GET /api/v1/portfolios/mine", s.listMyPortfolios)
	mux.HandleFunc("POST /api/v1/portfolios", s.createPortfolio)
	mux.HandleFunc("PUT /api/v1/portfolios/{id}", s.updatePortfolio)

	// ---- Achievements ----
	mux.HandleFunc("GET /api/v1/achievements", s.listAchievements)
	mux.HandleFunc("GET /api/v1/achievements/{id}", s.getAchievement)
	mux.HandleFunc("POST /api/v1/achievements", s.createAchievement)
	mux.HandleFunc("PUT /api/v1/achievements/{id}", s.updateAchievement)
	mux.HandleFunc("DELETE /api/v1/achievements/{id}", s.deleteAchievement)

	// ---- RD Challenges (研发难题) ----
	mux.HandleFunc("GET /api/v1/challenges", s.listRDChallenges)
	mux.HandleFunc("GET /api/v1/challenges/{id}", s.getRDChallenge)

	// ---- Research Projects (课题攻关) ----
	mux.HandleFunc("GET /api/v1/projects", s.listResearchProjects)
	mux.HandleFunc("GET /api/v1/projects/search", s.listResearchProjects)
	mux.HandleFunc("GET /api/v1/projects/{id}", s.getResearchProject)

	// ---- R&D Challenges ----
	mux.HandleFunc("GET /api/v1/rd-challenges", s.listRDChallenges)
	mux.HandleFunc("POST /api/v1/rd-challenges", s.createRDChallenge)
	mux.HandleFunc("PUT /api/v1/rd-challenges/{id}", s.updateRDChallenge)

	// ---- Research Projects ----
	mux.HandleFunc("GET /api/v1/research-projects", s.listResearchProjects)
	mux.HandleFunc("POST /api/v1/admin/research-projects", s.createResearchProject)
	mux.HandleFunc("PUT /api/v1/admin/research-projects/{id}", s.updateResearchProject)

	// ---- Project Applications ----
	mux.HandleFunc("GET /api/v1/project-applications/mine", s.listMyProjectApps)
	mux.HandleFunc("GET /api/v1/admin/project-applications", s.listAllProjectApps)
	mux.HandleFunc("POST /api/v1/project-applications", s.createProjectApp)
	mux.HandleFunc("POST /api/v1/admin/project-applications/{id}/review", s.reviewProjectApp)

	// ---- Competitions ----
	mux.HandleFunc("GET /api/v1/competitions", s.listCompetitions)
	mux.HandleFunc("POST /api/v1/admin/competitions", s.createCompetition)
	mux.HandleFunc("GET /api/v1/competitions/{id}/registrations", s.listCompetitionRegs)
	mux.HandleFunc("POST /api/v1/competitions/{id}/register", s.registerCompetition)

	// ---- Events ----
	mux.HandleFunc("GET /api/v1/events", s.listEvents)
	mux.HandleFunc("POST /api/v1/admin/events", s.createEvent)
	mux.HandleFunc("GET /api/v1/events/{id}/registrations", s.listEventRegistrations)
	mux.HandleFunc("POST /api/v1/events/{id}/register", s.registerEvent)

	// ---- Industry Resources ----
	mux.HandleFunc("GET /api/v1/industry-resources", s.listIndustryResources)
	mux.HandleFunc("GET /api/v1/industry-resources/{id}", s.getIndustryResourcePublic)
	mux.HandleFunc("POST /api/v1/admin/industry-resources", s.createIndustryResource)
	mux.HandleFunc("PUT /api/v1/admin/industry-resources/{id}", s.updateIndustryResource)

	// ---- 公开详情（补齐小程序页面接口） ----
	mux.HandleFunc("GET /api/v1/experts/{id}", s.getExpert)
	mux.HandleFunc("GET /api/v1/test-sites/{id}", s.getTestSite)
	mux.HandleFunc("GET /api/v1/exhibitions/{id}", s.getExhibition)
	mux.HandleFunc("GET /api/v1/competitions/{id}", s.getCompetition)
	mux.HandleFunc("GET /api/v1/colleges/{id}", s.getCollege)

	// ---- Emergency ----
	mux.HandleFunc("GET /api/v1/emergency-resources", s.listEmergencyResources)
	mux.HandleFunc("POST /api/v1/admin/emergency-resources", s.createEmergencyResource)
	mux.HandleFunc("GET /api/v1/emergency-dispatches", s.listEmergencyDispatches)
	mux.HandleFunc("POST /api/v1/admin/emergency-dispatches", s.createEmergencyDispatch)

	// ---- Smart Matching ----
	mux.HandleFunc("GET /api/v1/recommendations", s.recommendDemands)
	mux.HandleFunc("GET /api/v1/match", s.searchAndMatch)
}

// ---- Experts ----

// GET /api/v1/experts?field=农业
func (s *Server) listExperts(w http.ResponseWriter, r *http.Request) {
	items, err := s.expertSvc.List(r.URL.Query().Get("field"))
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, items)
}

// POST /api/v1/admin/experts
func (s *Server) createExpert(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		fail(w, r, http.StatusForbidden, errors.New("admin permission required"))
		return
	}
	var in struct {
		Name      string   `json:"name"`
		Title     string   `json:"title"`
		Org       string   `json:"org"`
		Field     string   `json:"field"`
		Bio       string   `json:"bio"`
		AvatarURL string   `json:"avatar_url"`
		Tags      []string `json:"tags"`
		Status    string   `json:"status"`
		CreatedAt string   `json:"created_at"`
		UpdatedAt string   `json:"updated_at"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	e, err := s.expertSvc.Create(in.Name, in.Title, in.Org, in.Field, in.Bio, in.AvatarURL, in.Status, in.Tags)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	s.audit(r.Context(), a.ID, "create_expert", "expert", e.ID, "created")
	respond(w, r, http.StatusCreated, e)
}

// PUT /api/v1/admin/experts/{id}
func (s *Server) updateExpert(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		fail(w, r, http.StatusForbidden, errors.New("admin permission required"))
		return
	}
	var in struct {
		ID        string   `json:"id"`
		Name      string   `json:"name"`
		Title     string   `json:"title"`
		Org       string   `json:"org"`
		Field     string   `json:"field"`
		Bio       string   `json:"bio"`
		AvatarURL string   `json:"avatar_url"`
		Tags      []string `json:"tags"`
		Status    string   `json:"status"`
		CreatedAt string   `json:"created_at"`
		UpdatedAt string   `json:"updated_at"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	e, err := s.expertSvc.Update(r.PathValue("id"), in.Name, in.Title, in.Org, in.Field, in.Bio, in.AvatarURL, in.Status, in.Tags)
	if err != nil {
		fail(w, r, http.StatusNotFound, err)
		return
	}
	respond(w, r, http.StatusOK, e)
}

// DELETE /api/v1/admin/experts/{id}
func (s *Server) deleteExpert(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		fail(w, r, http.StatusForbidden, errors.New("admin permission required"))
		return
	}
	if err := s.expertSvc.Delete(r.PathValue("id")); err != nil {
		fail(w, r, http.StatusNotFound, err)
		return
	}
	s.audit(r.Context(), a.ID, "delete_expert", "expert", r.PathValue("id"), "deleted")
	respond(w, r, http.StatusOK, map[string]string{"status": "deleted"})
}

// ---- Cases ----

// GET /api/v1/cases?category=农业&page=1&page_size=10
func (s *Server) listCases(w http.ResponseWriter, r *http.Request) {
	items, total, err := s.caseSvc.List(r.URL.Query().Get("category"), 1, 100000)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	paginatedRespond(w, r, items, total)
}

// POST /api/v1/admin/cases
func (s *Server) createCase(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		fail(w, r, http.StatusForbidden, errors.New("admin permission required"))
		return
	}
	var in struct {
		Title       string   `json:"title"`
		Category    string   `json:"category"`
		Description string   `json:"description"`
		ClientName  string   `json:"client_name"`
		Result      string   `json:"result"`
		Images      []string `json:"images"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	c, err := s.caseSvc.Create(in.Title, in.Category, in.Description, in.Images, in.ClientName, in.Result)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	s.audit(r.Context(), a.ID, "create_case", "case", c.ID, "created")
	respond(w, r, http.StatusCreated, c)
}

// GET /api/v1/admin/cases/{id}
func (s *Server) getCase(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		fail(w, r, http.StatusForbidden, errors.New("admin permission required"))
		return
	}
	c, err := s.caseSvc.Get(r.PathValue("id"))
	if err != nil {
		fail(w, r, http.StatusNotFound, err)
		return
	}
	respond(w, r, http.StatusOK, c)
}

// PUT /api/v1/admin/cases/{id}
func (s *Server) updateCase(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		fail(w, r, http.StatusForbidden, errors.New("admin permission required"))
		return
	}
	var in struct {
		Title       string   `json:"title"`
		Category    string   `json:"category"`
		Description string   `json:"description"`
		ClientName  string   `json:"client_name"`
		Result      string   `json:"result"`
		Status      string   `json:"status"`
		Images      []string `json:"images"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	c, err := s.caseSvc.Update(r.PathValue("id"), in.Title, in.Category, in.Description, in.Status, in.Images, in.ClientName, in.Result)
	if err != nil {
		fail(w, r, http.StatusNotFound, err)
		return
	}
	respond(w, r, http.StatusOK, c)
}

// DELETE /api/v1/admin/cases/{id}
func (s *Server) deleteCase(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		fail(w, r, http.StatusForbidden, errors.New("admin permission required"))
		return
	}
	if err := s.caseSvc.Delete(r.PathValue("id")); err != nil {
		fail(w, r, http.StatusNotFound, err)
		return
	}
	s.audit(r.Context(), a.ID, "delete_case", "case", r.PathValue("id"), "deleted")
	respond(w, r, http.StatusOK, map[string]string{"status": "deleted"})
}

// ---- Compliance ----

// GET /api/v1/compliance-docs?category=政策法规&page=1&page_size=10
func (s *Server) listComplianceDocs(w http.ResponseWriter, r *http.Request) {
	page, pageSize := paginationFromQuery(r)
	items, total, err := s.complianceSvc.ListDocs(r.URL.Query().Get("category"), page, pageSize)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	paginatedRespond(w, r, items, total)
}

// POST /api/v1/admin/compliance-docs
func (s *Server) createComplianceDoc(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		fail(w, r, http.StatusForbidden, errors.New("admin permission required"))
		return
	}
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
	d, err := s.complianceSvc.CreateDoc(in.Title, in.Category, in.Publisher, in.PublishDate, in.Status, in.Summary, in.FileURL, in.Tags)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	s.audit(r.Context(), a.ID, "create_compliance_doc", "compliance_doc", d.ID, "created")
	respond(w, r, http.StatusCreated, d)
}

// GET /api/v1/compliance-standards?category=团体标准&page=1&page_size=10
func (s *Server) listComplianceStandards(w http.ResponseWriter, r *http.Request) {
	page, pageSize := paginationFromQuery(r)
	items, total, err := s.complianceSvc.ListStandards(r.URL.Query().Get("category"), page, pageSize)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	paginatedRespond(w, r, items, total)
}

// POST /api/v1/admin/compliance-standards
func (s *Server) createComplianceStandard(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		fail(w, r, http.StatusForbidden, errors.New("admin permission required"))
		return
	}
	var in struct {
		ID            string `json:"id"`
		Title         string `json:"title"`
		StandardNo    string `json:"standard_no"`
		Publisher     string `json:"publisher"`
		EffectiveDate string `json:"effective_date"`
		Status        string `json:"status"`
		Scope         string `json:"scope"`
		FileURL       string `json:"file_url"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	sd, err := s.complianceSvc.CreateStandard(in.Title, in.StandardNo, in.Publisher, in.EffectiveDate, in.Status, in.Scope, in.FileURL)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	s.audit(r.Context(), a.ID, "create_compliance_standard", "standard", sd.ID, "created")
	respond(w, r, http.StatusCreated, sd)
}

// ---- Industry Reports ----

// GET /api/v1/industry-reports?page=1&page_size=10
// 支持 keyword/status 过滤（管理端搜索/筛选；无参时与公开列表行为一致）
func (s *Server) listIndustryReports(w http.ResponseWriter, r *http.Request) {
	items, _, err := s.reportSvc.List(1, 10000)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	filtered, _ := adminListFilter(items, r.URL.Query().Get("keyword"), r.URL.Query().Get("status"),
		func(rep domain.IndustryReport) string { return rep.Title },
		func(rep domain.IndustryReport) string { return rep.Status })
	// 类型筛选（category：whitepaper/research/analysis/other）
	if cat := r.URL.Query().Get("category"); cat != "" {
		tmp := make([]domain.IndustryReport, 0, len(filtered))
		for _, rep := range filtered {
			if rep.Category == cat {
				tmp = append(tmp, rep)
			}
		}
		filtered = tmp
	}
	paginatedRespond(w, r, filtered, len(filtered))
}

// POST /api/v1/admin/industry-reports
func (s *Server) createIndustryReport(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		fail(w, r, http.StatusForbidden, errors.New("admin permission required"))
		return
	}
	var in struct {
		Title    string `json:"title"`
		Period   string `json:"period"`
		Category string `json:"category"`
		Summary  string `json:"summary"`
		Content  string `json:"content"`
		FileURL  string `json:"file_url"`
		Author   string `json:"author"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	rep, err := s.reportSvc.Create(in.Title, in.Period, in.Category, in.Summary, in.Content, in.FileURL, in.Author)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	s.audit(r.Context(), a.ID, "create_report", "report", rep.ID, "created")
	respond(w, r, http.StatusCreated, rep)
}

// DELETE /api/v1/admin/industry-reports/{id}
func (s *Server) deleteIndustryReport(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		fail(w, r, http.StatusForbidden, errors.New("admin permission required"))
		return
	}
	if err := s.reportSvc.Delete(r.PathValue("id")); err != nil {
		fail(w, r, http.StatusNotFound, err)
		return
	}
	s.audit(r.Context(), a.ID, "delete_report", "report", r.PathValue("id"), "deleted")
	respond(w, r, http.StatusOK, map[string]string{"status": "deleted"})
}

// ---- Portfolio ----

// GET /api/v1/portfolios?page=1&page_size=10
func (s *Server) listPortfolios(w http.ResponseWriter, r *http.Request) {
	page, pageSize := paginationFromQuery(r)
	items, total, err := s.portfolioSvc.ListPublished(page, pageSize)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	paginatedRespond(w, r, items, total)
}

// GET /api/v1/admin/portfolios — 管理端全量（含草稿/待审），公开端仅 published
func (s *Server) listAdminPortfolios(w http.ResponseWriter, r *http.Request) {
	items, _, err := s.portfolioSvc.List(1, 100000)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	filtered, _ := adminListFilter(items, r.URL.Query().Get("keyword"), r.URL.Query().Get("status"),
		func(p domain.MemberPortfolio) string { return p.Name },
		func(p domain.MemberPortfolio) string { return p.Status })
	paginatedRespond(w, r, filtered, len(filtered))
}

// GET /api/v1/portfolios/mine
func (s *Server) listMyPortfolios(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	items, err := s.portfolioSvc.ListByEnterprise(a.ID)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, items)
}

// POST /api/v1/portfolios
func (s *Server) createPortfolio(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	var in struct {
		Name        string   `json:"name"`
		LogoURL     string   `json:"logo_url"`
		CoverURL    string   `json:"cover_url"`
		Description string   `json:"description"`
		ContactInfo string   `json:"contact_info"`
		Products    []string `json:"products"`
		Honors      []string `json:"honors"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	p, err := s.portfolioSvc.Create(a.ID, in.Name, in.LogoURL, in.CoverURL, in.Description, in.ContactInfo, in.Products, in.Honors)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	s.audit(r.Context(), a.ID, "create_portfolio", "portfolio", p.ID, "created")
	respond(w, r, http.StatusCreated, p)
}

// PUT /api/v1/portfolios/{id}
func (s *Server) updatePortfolio(w http.ResponseWriter, r *http.Request) {
	_, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	var in struct {
		Name        string   `json:"name"`
		LogoURL     string   `json:"logo_url"`
		CoverURL    string   `json:"cover_url"`
		Description string   `json:"description"`
		ContactInfo string   `json:"contact_info"`
		Status      string   `json:"status"`
		Products    []string `json:"products"`
		Honors      []string `json:"honors"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	p, err := s.portfolioSvc.Update(r.PathValue("id"), in.Name, in.LogoURL, in.CoverURL, in.Description, in.ContactInfo, in.Status, in.Products, in.Honors)
	if err != nil {
		fail(w, r, http.StatusNotFound, err)
		return
	}
	respond(w, r, http.StatusOK, p)
}

// ---- Achievements ----

// GET /api/v1/achievements?field=智能巡检&page=1&page_size=10
func (s *Server) listAchievements(w http.ResponseWriter, r *http.Request) {
	items, _, err := s.achievementSvc.List(r.URL.Query().Get("field"), 1, 100000)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	// 关键词筛选（成果名称/领域）
	if kw := strings.TrimSpace(r.URL.Query().Get("keyword")); kw != "" {
		tmp := make([]domain.Achievement, 0, len(items))
		for _, a := range items {
			if strings.Contains(a.Title, kw) || strings.Contains(a.Field, kw) {
				tmp = append(tmp, a)
			}
		}
		items = tmp
	}
	paginatedRespond(w, r, items, len(items))
}

// POST /api/v1/achievements
func (s *Server) createAchievement(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	var in struct {
		Title       string              `json:"title"`
		AchieveType string              `json:"achieve_type"`
		Description string              `json:"description"`
		Field       string              `json:"field"`
		Stage       string              `json:"stage"`
		ContactInfo string              `json:"contact_info"`
		Images      []string            `json:"images"`
		Attachments []domain.Attachment `json:"attachments"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	ach, err := s.achievementSvc.Create(a.ID, in.Title, in.AchieveType, in.Description, in.Field, in.Stage, in.ContactInfo, in.Images, in.Attachments)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	s.audit(r.Context(), a.ID, "create_achievement", "achievement", ach.ID, "created")
	respond(w, r, http.StatusCreated, ach)
}

// PUT /api/v1/achievements/{id}
func (s *Server) updateAchievement(w http.ResponseWriter, r *http.Request) {
	_, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	var in struct {
		Title       string              `json:"title"`
		AchieveType string              `json:"achieve_type"`
		Description string              `json:"description"`
		Field       string              `json:"field"`
		Stage       string              `json:"stage"`
		ContactInfo string              `json:"contact_info"`
		Images      []string            `json:"images"`
		Attachments []domain.Attachment `json:"attachments"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	ach, err := s.achievementSvc.Update(r.PathValue("id"), in.Title, in.AchieveType, in.Description, in.Field, in.Stage, in.ContactInfo, in.Images, in.Attachments)
	if err != nil {
		fail(w, r, http.StatusNotFound, err)
		return
	}
	respond(w, r, http.StatusOK, ach)
}

// DELETE /api/v1/achievements/{id}
func (s *Server) deleteAchievement(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	if err := s.achievementSvc.Delete(r.PathValue("id")); err != nil {
		fail(w, r, http.StatusNotFound, err)
		return
	}
	s.audit(r.Context(), a.ID, "delete_achievement", "achievement", r.PathValue("id"), "deleted")
	respond(w, r, http.StatusOK, map[string]string{"status": "deleted"})
}

// ---- R&D Challenges ----

// GET /api/v1/rd-challenges?field=飞控&page=1&page_size=10
func (s *Server) listRDChallenges(w http.ResponseWriter, r *http.Request) {
	items, _, err := s.rdService.List(r.URL.Query().Get("field"), 1, 100000)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	filtered := items
	// 关键词筛选（难题标题/领域）
	if kw := strings.TrimSpace(r.URL.Query().Get("keyword")); kw != "" {
		tmp := make([]domain.RDChallenge, 0, len(filtered))
		for _, c := range filtered {
			if strings.Contains(c.Title, kw) || strings.Contains(c.Field, kw) {
				tmp = append(tmp, c)
			}
		}
		filtered = tmp
	}
	// 状态筛选
	if st := r.URL.Query().Get("status"); st != "" {
		tmp := make([]domain.RDChallenge, 0, len(filtered))
		for _, c := range filtered {
			if c.Status == st {
				tmp = append(tmp, c)
			}
		}
		filtered = tmp
	}
	paginatedRespond(w, r, filtered, len(filtered))
}

// POST /api/v1/rd-challenges
func (s *Server) createRDChallenge(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	var in struct {
		Title, Field, Description, Status, Deadline string
		BudgetFen                                   int64 `json:"budget_fen"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	deadline, err := parseDateInput(in.Deadline)
	if err != nil {
		fail(w, r, http.StatusBadRequest, fmt.Errorf("无效的截止日期格式: %w", err))
		return
	}
	ch, err := s.rdService.Create(a.ID, in.Title, in.Field, in.Description, in.BudgetFen, deadline)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	s.audit(r.Context(), a.ID, "create_rd_challenge", "rd_challenge", ch.ID, "created")
	respond(w, r, http.StatusCreated, ch)
}

// PUT /api/v1/rd-challenges/{id}
func (s *Server) updateRDChallenge(w http.ResponseWriter, r *http.Request) {
	_, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	var in struct {
		Title, Field, Description, Status, Deadline string
		BudgetFen                                   int64 `json:"budget_fen"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	deadline, err := parseDateInput(in.Deadline)
	if err != nil {
		fail(w, r, http.StatusBadRequest, fmt.Errorf("无效的截止日期格式: %w", err))
		return
	}
	ch, err := s.rdService.Update(r.PathValue("id"), in.Title, in.Field, in.Description, in.Status, in.BudgetFen, deadline)
	if err != nil {
		fail(w, r, http.StatusNotFound, err)
		return
	}
	respond(w, r, http.StatusOK, ch)
}

// ---- Research Projects ----

// GET /api/v1/research-projects?page=1&page_size=10
// GET /api/v1/research-projects — 支持 keyword/status 过滤（管理端搜索/筛选；无参时与公开列表行为一致）
func (s *Server) listResearchProjects(w http.ResponseWriter, r *http.Request) {
	items, _, err := s.researchSvc.List(1, 10000)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	filtered, total := adminListFilter(items, r.URL.Query().Get("keyword"), r.URL.Query().Get("status"),
		func(p domain.ResearchProject) string { return p.Title },
		func(p domain.ResearchProject) string { return p.Status })
	paginatedRespond(w, r, filtered, total)
}

// POST /api/v1/admin/research-projects
func (s *Server) createResearchProject(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		fail(w, r, http.StatusForbidden, errors.New("admin permission required"))
		return
	}
	var in struct {
		Title       string   `json:"title"`
		Field       string   `json:"field"`
		Description string   `json:"description"`
		Status      string   `json:"status"`
		LeadOrg     string   `json:"lead_org"`
		Milestones  string   `json:"milestones"`
		StartDate   string   `json:"start_date"`
		EndDate     string   `json:"end_date"`
		Members     []string `json:"members"`
		BudgetFen   int64    `json:"budget_fen"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	startDate, err := parseDateInput(in.StartDate)
	if err != nil {
		fail(w, r, http.StatusBadRequest, fmt.Errorf("无效的开始日期格式: %w", err))
		return
	}
	endDate, err := parseDateInput(in.EndDate)
	if err != nil {
		fail(w, r, http.StatusBadRequest, fmt.Errorf("无效的结束日期格式: %w", err))
		return
	}
	p, err := s.researchSvc.Create(in.Title, in.Field, in.Description, in.LeadOrg, in.Milestones, in.Members, in.BudgetFen, startDate, endDate)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	s.audit(r.Context(), a.ID, "create_research_project", "research_project", p.ID, "created")
	respond(w, r, http.StatusCreated, p)
}

// PUT /api/v1/admin/research-projects/{id}
func (s *Server) updateResearchProject(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		fail(w, r, http.StatusForbidden, errors.New("admin permission required"))
		return
	}
	var in struct {
		Title       string   `json:"title"`
		Field       string   `json:"field"`
		Description string   `json:"description"`
		Status      string   `json:"status"`
		LeadOrg     string   `json:"lead_org"`
		Milestones  string   `json:"milestones"`
		StartDate   string   `json:"start_date"`
		EndDate     string   `json:"end_date"`
		Members     []string `json:"members"`
		BudgetFen   int64    `json:"budget_fen"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	startDate, err := parseDateInput(in.StartDate)
	if err != nil {
		fail(w, r, http.StatusBadRequest, fmt.Errorf("无效的开始日期格式: %w", err))
		return
	}
	endDate, err := parseDateInput(in.EndDate)
	if err != nil {
		fail(w, r, http.StatusBadRequest, fmt.Errorf("无效的结束日期格式: %w", err))
		return
	}
	p, err := s.researchSvc.Update(r.PathValue("id"), in.Title, in.Field, in.Description, in.Status, in.LeadOrg, in.Milestones, in.Members, in.BudgetFen, startDate, endDate)
	if err != nil {
		fail(w, r, http.StatusNotFound, err)
		return
	}
	respond(w, r, http.StatusOK, p)
}

// ---- Project Applications ----

// GET /api/v1/project-applications/mine
func (s *Server) listMyProjectApps(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	items, err := s.projectAppSvc.ListMy(a.ID)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, items)
}

// GET /api/v1/admin/project-applications?status=submitted&page=1&page_size=10
func (s *Server) listAllProjectApps(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		fail(w, r, http.StatusForbidden, errors.New("admin permission required"))
		return
	}
	items, total, err := s.projectAppSvc.ListAll(r.URL.Query().Get("status"), 1, 100000)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	paginatedRespond(w, r, items, total)
}

// POST /api/v1/project-applications
func (s *Server) createProjectApp(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	var in struct {
		ProjectName, Category, Description string
		BudgetFen                          int64    `json:"budget_fen"`
		Attachments                        []string `json:"attachments"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	app, err := s.projectAppSvc.Create(a.ID, in.ProjectName, in.Category, in.Description, in.BudgetFen, in.Attachments)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	s.audit(r.Context(), a.ID, "create_project_app", "project_app", app.ID, "created")
	respond(w, r, http.StatusCreated, app)
}

// POST /api/v1/admin/project-applications/{id}/review
func (s *Server) reviewProjectApp(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		fail(w, r, http.StatusForbidden, errors.New("admin permission required"))
		return
	}
	var in struct {
		Action     string `json:"action"` // "approve" or "reject"
		ReviewNote string `json:"review_note"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	app, err := s.projectAppSvc.Review(r.PathValue("id"), in.ReviewNote, in.Action)
	if err != nil {
		fail(w, r, http.StatusNotFound, err)
		return
	}
	s.audit(r.Context(), a.ID, "review_project_app", "project_app", app.ID, in.Action)
	respond(w, r, http.StatusOK, app)
}

// ---- Competitions ----

// GET /api/v1/competitions?page=1&page_size=10&status=enrolling&keyword=无人机
// status 筛选兼容页面值域（enrolling/open/ongoing/closed/full）与后端状态（published/...）。
func (s *Server) listCompetitions(w http.ResponseWriter, r *http.Request) {
	items, _, err := s.competitionSvc.List(1, 100000)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	status := r.URL.Query().Get("status")
	keyword := r.URL.Query().Get("keyword")
	var out []domain.Competition
	for _, c := range items {
		if status != "" && !matchCompetitionStatus(status, c.Status) {
			continue
		}
		if keyword != "" && !strings.Contains(c.Title, keyword) && !strings.Contains(c.Category, keyword) {
			continue
		}
		out = append(out, c)
	}
	paginatedRespond(w, r, out, len(out))
}

// matchCompetitionStatus 页面 tab（enrolling/ongoing/closed/full）映射到后端状态值。
func matchCompetitionStatus(query, s string) bool {
	switch query {
	case "enrolling", "open":
		return s == "published" || s == "enrolling" || s == "open" || s == "upcoming"
	case "ongoing":
		return s == "ongoing" || s == "active" || s == "in_progress"
	case "closed", "full":
		return s == "closed" || s == "full" || s == "ended" || s == "finished"
	default:
		return s == query
	}
}

// POST /api/v1/admin/competitions
func (s *Server) createCompetition(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		fail(w, r, http.StatusForbidden, errors.New("admin permission required"))
		return
	}
	var in struct {
		Title       string `json:"title"`
		Category    string `json:"category"`
		Description string `json:"description"`
		Location    string `json:"location"`
		Sponsor     string `json:"sponsor"`
		StartDate   string `json:"start_date"`
		EndDate     string `json:"end_date"`
		MaxTeams    int    `json:"max_teams"`
		// 小程序赛事页扩展字段（competitions/detail + register）
		Deadline           string                          `json:"deadline"`
		OrganizerSub       string                          `json:"organizer_sub"`
		Fee                int                             `json:"fee"`
		MinFee             int                             `json:"min_fee"`
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
	startDate, err := parseDateInput(in.StartDate)
	if err != nil {
		fail(w, r, http.StatusBadRequest, fmt.Errorf("无效的开始日期格式: %w", err))
		return
	}
	endDate, err := parseDateInput(in.EndDate)
	if err != nil {
		fail(w, r, http.StatusBadRequest, fmt.Errorf("无效的结束日期格式: %w", err))
		return
	}
	var deadline *time.Time
	if d, err := parseDateInput(in.Deadline); err == nil && !d.IsZero() {
		deadline = &d
	}
	c, err := s.competitionSvc.Create(domain.Competition{
		Title: in.Title, Category: in.Category, Description: in.Description,
		Location: in.Location, Sponsor: in.Sponsor, StartDate: startDate, EndDate: endDate,
		MaxTeams: in.MaxTeams, Deadline: deadline, OrganizerSub: in.OrganizerSub,
		Fee: in.Fee, MinFee: in.MinFee, Tags: in.Tags, Poster: in.Poster,
		Requirements: in.Requirements, Events: in.Events, Prizes: in.Prizes,
		RegistrationStatus: in.RegistrationStatus,
	})
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	s.audit(r.Context(), a.ID, "create_competition", "competition", c.ID, "created")
	respond(w, r, http.StatusCreated, c)
}

// GET /api/v1/competitions/{id}/registrations
func (s *Server) listCompetitionRegs(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		fail(w, r, http.StatusForbidden, errors.New("admin permission required"))
		return
	}
	regs, err := s.competitionSvc.ListRegs(r.PathValue("id"))
	if err != nil {
		fail(w, r, http.StatusNotFound, err)
		return
	}
	respond(w, r, http.StatusOK, regs)
}

// POST /api/v1/competitions/{id}/register
func (s *Server) registerCompetition(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	var in struct {
		TeamName    string `json:"team_name"`
		MemberCount int    `json:"member_count"`
		ContactInfo string `json:"contact_info"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	reg, err := s.competitionSvc.Register(r.PathValue("id"), a.ID, in.TeamName, in.MemberCount, in.ContactInfo)
	if err != nil {
		fail(w, r, http.StatusNotFound, err)
		return
	}
	respond(w, r, http.StatusCreated, reg)
}

// ---- Events ----

// GET /api/v1/events?page=1&page_size=10
func (s *Server) listEvents(w http.ResponseWriter, r *http.Request) {
	items, total, err := s.eventSvc.List(1, 100000)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	paginatedRespond(w, r, items, total)
}

// POST /api/v1/admin/events
func (s *Server) createEvent(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		fail(w, r, http.StatusForbidden, errors.New("admin permission required"))
		return
	}
	var in struct {
		Title        string `json:"title"`
		EventType    string `json:"event_type"`
		Description  string `json:"description"`
		Location     string `json:"location"`
		CoverURL     string `json:"cover_url"`
		StartTime    string `json:"start_time"`
		EndTime      string `json:"end_time"`
		MaxAttendees int    `json:"max_attendees"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	startTime, err := parseDateInput(in.StartTime)
	if err != nil {
		fail(w, r, http.StatusBadRequest, fmt.Errorf("无效的开始时间格式: %w", err))
		return
	}
	endTime, err := parseDateInput(in.EndTime)
	if err != nil {
		fail(w, r, http.StatusBadRequest, fmt.Errorf("无效的结束时间格式: %w", err))
		return
	}
	ev, err := s.eventSvc.Create(in.Title, in.EventType, in.Description, in.Location, in.CoverURL, startTime, endTime, in.MaxAttendees)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	s.audit(r.Context(), a.ID, "create_event", "event", ev.ID, "created")
	respond(w, r, http.StatusCreated, ev)
}

// GET /api/v1/events/{id}/registrations
func (s *Server) listEventRegistrations(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		fail(w, r, http.StatusForbidden, errors.New("admin permission required"))
		return
	}
	regs, err := s.eventSvc.ListRegs(r.PathValue("id"))
	if err != nil {
		fail(w, r, http.StatusNotFound, err)
		return
	}
	respond(w, r, http.StatusOK, regs)
}

// POST /api/v1/events/{id}/register
func (s *Server) registerEvent(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	var in struct {
		Name, Phone, Org string
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	reg, err := s.eventSvc.Register(r.PathValue("id"), a.ID, in.Name, in.Phone, in.Org)
	if err != nil {
		fail(w, r, http.StatusNotFound, err)
		return
	}
	respond(w, r, http.StatusCreated, reg)
}

// ---- Industry Resources ----

// GET /api/v1/industry-resources?res_type=drone&page=1&page_size=10
// resourceLevelRank 可见级别排序：public(0) < member(1) < partner(2) < admin(3)
func resourceLevelRank(lv string) int {
	switch lv {
	case "admin":
		return 3
	case "partner":
		return 2
	case "member":
		return 1
	default:
		return 0
	}
}

// visitorResourceLevel 判定访问者的资源可见级别
// 协会管理员 > 副会长单位(partner) > 合作院校/普通会员(member) > 政府访客(public)
// 注意：公开 GET 路径上 authenticate 不注入 actor，需手动解析 token
func (s *Server) visitorResourceLevel(r *http.Request) int {
	h := r.Header.Get("Authorization")
	if !strings.HasPrefix(h, "Bearer ") {
		return 0 // 政府访客（未登录）
	}
	a, err := s.tokens.Verify(strings.TrimPrefix(h, "Bearer "))
	if err != nil {
		return 0
	}
	if a.Role == domain.RolePlatformAdmin || a.Role == domain.RoleAssociationAdmin {
		return 3 // 协会管理员
	}
	// 查单位身份（association_members.role: partner 副会长单位 / college 合作院校）
	if s.assocMemberSvc != nil {
		if m, err := s.assocMemberSvc.GetByUserID(a.ID); err == nil {
			if m.Role == domain.AssocPartner {
				return 2
			}
			if m.Role == domain.AssocCollege {
				return 1
			}
		}
	}
	return 1 // 普通会员
}

// GET /api/v1/industry-resources/{id} — 公开详情（分级校验：资源级别 ≤ 访问者级别）
func (s *Server) getIndustryResourcePublic(w http.ResponseWriter, r *http.Request) {
	res, err := s.resourceSvc.Get(r.PathValue("id"))
	if err != nil {
		fail(w, r, http.StatusNotFound, err)
		return
	}
	if resourceLevelRank(res.VisibilityLevel) > s.visitorResourceLevel(r) {
		fail(w, r, http.StatusNotFound, errors.New("resource not found"))
		return
	}
	respond(w, r, http.StatusOK, res)
}

func (s *Server) listIndustryResources(w http.ResponseWriter, r *http.Request) {
	items, _, err := s.resourceSvc.List(r.URL.Query().Get("res_type"), 1, 100000)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	// 分级浏览过滤（.doc 原始需求：协会管理员/副会长单位/普通会员/合作院校/政府访客分级浏览）
	visitor := s.visitorResourceLevel(r)
	visible := make([]domain.IndustryResource, 0, len(items))
	for _, it := range items {
		if resourceLevelRank(it.VisibilityLevel) <= visitor {
			visible = append(visible, it)
		}
	}
	paginatedRespond(w, r, visible, len(visible))
}

// POST /api/v1/admin/industry-resources
func (s *Server) createIndustryResource(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		fail(w, r, http.StatusForbidden, errors.New("admin permission required"))
		return
	}
	var in struct {
		Name            string `json:"name"`
		ResType         string `json:"res_type"`
		Model           string `json:"model"`
		Specs           string `json:"specs"`
		Location        string `json:"location"`
		BookingInfo     string `json:"booking_info"`
		VisibilityLevel string `json:"visibility_level"`
		PriceFen        int64  `json:"price_fen"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	res, err := s.resourceSvc.Create(a.ID, in.Name, in.ResType, in.Model, in.Specs, in.Location, in.BookingInfo, in.PriceFen, in.VisibilityLevel)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	s.audit(r.Context(), a.ID, "create_resource", "resource", res.ID, "created")
	respond(w, r, http.StatusCreated, res)
}

// PUT /api/v1/admin/industry-resources/{id}
func (s *Server) updateIndustryResource(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		fail(w, r, http.StatusForbidden, errors.New("admin permission required"))
		return
	}
	var in struct {
		Name            string `json:"name"`
		ResType         string `json:"res_type"`
		Model           string `json:"model"`
		Specs           string `json:"specs"`
		Location        string `json:"location"`
		BookingInfo     string `json:"booking_info"`
		VisibilityLevel string `json:"visibility_level"`
		Status          string `json:"status"`
		PriceFen        int64  `json:"price_fen"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	res, err := s.resourceSvc.Update(r.PathValue("id"), in.Name, in.ResType, in.Model, in.Specs, in.Location, in.BookingInfo, in.PriceFen, in.VisibilityLevel, in.Status)
	if err != nil {
		fail(w, r, http.StatusNotFound, err)
		return
	}
	respond(w, r, http.StatusOK, res)
}

// ---- Emergency ----

// GET /api/v1/emergency-resources?page=1&page_size=10
func (s *Server) listEmergencyResources(w http.ResponseWriter, r *http.Request) {
	page, pageSize := paginationFromQuery(r)
	items, total, err := s.emergencySvc.ListResources(page, pageSize)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	paginatedRespond(w, r, items, total)
}

// POST /api/v1/admin/emergency-resources
func (s *Server) createEmergencyResource(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		fail(w, r, http.StatusForbidden, errors.New("admin permission required"))
		return
	}
	var in struct {
		Name        string `json:"name"`
		ResType     string `json:"res_type"`
		Specs       string `json:"specs"`
		Location    string `json:"location"`
		ContactInfo string `json:"contact_info"`
		Quantity    int    `json:"quantity"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	res, err := s.emergencySvc.CreateResource(a.ID, in.Name, in.ResType, in.Specs, in.Location, in.ContactInfo, in.Quantity)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	s.audit(r.Context(), a.ID, "create_emergency_resource", "emergency_resource", res.ID, "created")
	respond(w, r, http.StatusCreated, res)
}

// GET /api/v1/emergency-dispatches?page=1&page_size=10&status=pending
// 公开展示（与救援案例一致）：调度记录作为应急协同成果对会员公开展示
// status 筛选支持页面 dispatches.vue 值域：pending / dispatched / completed / ongoing / done / cancelled
func (s *Server) listEmergencyDispatches(w http.ResponseWriter, r *http.Request) {
	page, pageSize := paginationFromQuery(r)
	items, total, err := s.emergencySvc.ListDispatches(page, pageSize)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	if status := r.URL.Query().Get("status"); status != "" {
		var out []domain.EmergencyDispatch
		for _, d := range items {
			if d.Status == status {
				out = append(out, d)
			}
		}
		paginatedRespond(w, r, out, len(out))
		return
	}
	paginatedRespond(w, r, items, total)
}

// POST /api/v1/admin/emergency-dispatches
func (s *Server) createEmergencyDispatch(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		fail(w, r, http.StatusForbidden, errors.New("admin permission required"))
		return
	}
	var in struct {
		ResourceID string `json:"resource_id"`
		EventDesc  string `json:"event_desc"`
		Location   string `json:"location"`
		Commander  string `json:"commander"`
		Result     string `json:"result"`
		StartTime  string `json:"start_time"`
		EndTime    string `json:"end_time"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	startTime, err := parseDateInput(in.StartTime)
	if err != nil {
		fail(w, r, http.StatusBadRequest, fmt.Errorf("无效的开始时间格式: %w", err))
		return
	}
	endTime, err := parseDateInput(in.EndTime)
	if err != nil {
		fail(w, r, http.StatusBadRequest, fmt.Errorf("无效的结束时间格式: %w", err))
		return
	}
	d, err := s.emergencySvc.CreateDispatch(in.ResourceID, in.EventDesc, in.Location, in.Commander, in.Result, startTime, endTime)
	if err != nil {
		fail(w, r, http.StatusNotFound, err)
		return
	}
	s.audit(r.Context(), a.ID, "create_emergency_dispatch", "emergency_dispatch", d.ID, "created")
	respond(w, r, http.StatusCreated, d)
}

// ---- Smart Matching ----

// GET /api/v1/recommendations?biz_type=cable_inspection&district=南岸区&lat=29.5&lng=106.5
func (s *Server) recommendDemands(w http.ResponseWriter, r *http.Request) {
	bizType := r.URL.Query().Get("biz_type")
	district := r.URL.Query().Get("district")
	lat, _ := strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
	lng, _ := strconv.ParseFloat(r.URL.Query().Get("lng"), 64)
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 20
	}
	userID := ""
	if a, ok := authenticatedActor(r); ok {
		userID = a.ID
	}
	results, err := s.matchingSvc.Recommend(userID, lat, lng, bizType, district, limit)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, results)
}

// GET /api/v1/match?q=巡检&biz_type=cable_inspection&lat=29.5&lng=106.5
func (s *Server) searchAndMatch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		fail(w, r, http.StatusBadRequest, errors.New("q parameter required"))
		return
	}
	bizType := r.URL.Query().Get("biz_type")
	lat, _ := strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
	lng, _ := strconv.ParseFloat(r.URL.Query().Get("lng"), 64)
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 20
	}
	results, err := s.matchingSvc.SearchAndMatch(q, lat, lng, bizType, limit)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, results)
}
