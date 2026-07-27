package httpapi

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"drone-platform/internal/domain"
)

// registerBizRoutes registers all business-module routes.
func (s *Server) registerBizRoutes(mux *http.ServeMux) {
	// ---- Shops (Admin CRUD) ----
	mux.HandleFunc("GET /api/v1/admin/shops", s.listAdminShops)
	mux.HandleFunc("POST /api/v1/admin/shops", s.createAdminShop)
	mux.HandleFunc("PUT /api/v1/admin/shops/{id}", s.updateAdminShop)
	mux.HandleFunc("DELETE /api/v1/admin/shops/{id}", s.deleteAdminShop)

	// ---- Experts ----
	mux.HandleFunc("GET /api/v1/experts", s.listExperts)
	mux.HandleFunc("POST /api/v1/admin/experts", s.createExpert)
	mux.HandleFunc("PUT /api/v1/admin/experts/{id}", s.updateExpert)
	mux.HandleFunc("DELETE /api/v1/admin/experts/{id}", s.deleteExpert)

	// ---- Cases ----
	mux.HandleFunc("GET /api/v1/cases", s.listCases)
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
	mux.HandleFunc("POST /api/v1/achievements", s.createAchievement)
	mux.HandleFunc("PUT /api/v1/achievements/{id}", s.updateAchievement)
	mux.HandleFunc("DELETE /api/v1/achievements/{id}", s.deleteAchievement)

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
	mux.HandleFunc("POST /api/v1/admin/industry-resources", s.createIndustryResource)
	mux.HandleFunc("PUT /api/v1/admin/industry-resources/{id}", s.updateIndustryResource)

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
		Name, Title, Org, Field, Bio string
		Tags                          []string `json:"tags"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	e, err := s.expertSvc.Create(in.Name, in.Title, in.Org, in.Field, in.Bio, in.Tags)
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
		Name, Title, Org, Field, Bio string
		Tags                          []string `json:"tags"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	e, err := s.expertSvc.Update(r.PathValue("id"), in.Name, in.Title, in.Org, in.Field, in.Bio, in.Tags)
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
	page, pageSize := paginationFromQuery(r)
	items, total, err := s.caseSvc.List(r.URL.Query().Get("category"), page, pageSize)
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
		Title, Category, Description, ClientName, Result string
		Images                                            []string `json:"images"`
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
		Title, Category, Description, ClientName, Result string
		Images                                            []string `json:"images"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	c, err := s.caseSvc.Update(r.PathValue("id"), in.Title, in.Category, in.Description, in.Images, in.ClientName, in.Result)
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
		Title, Category, Content, Summary, Source string
		Tags                                       []string `json:"tags"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	d, err := s.complianceSvc.CreateDoc(in.Title, in.Category, in.Content, in.Summary, in.Source, in.Tags)
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
		Title, StdNumber, Category, Version, Publisher, Content, FileURL string
		IssueDate                                                         string `json:"issue_date"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	issueDate, _ := time.Parse("2006-01-02", in.IssueDate)
	sd, err := s.complianceSvc.CreateStandard(in.Title, in.StdNumber, in.Category, in.Version, in.Publisher, in.Content, in.FileURL, issueDate)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	s.audit(r.Context(), a.ID, "create_compliance_standard", "standard", sd.ID, "created")
	respond(w, r, http.StatusCreated, sd)
}

// ---- Industry Reports ----

// GET /api/v1/industry-reports?page=1&page_size=10
func (s *Server) listIndustryReports(w http.ResponseWriter, r *http.Request) {
	page, pageSize := paginationFromQuery(r)
	items, total, err := s.reportSvc.List(page, pageSize)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	paginatedRespond(w, r, items, total)
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
		Title, Period, Category, Summary, Content, FileURL, Author string
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
		Name, LogoURL, CoverURL, Description, ContactInfo string
		Products []string `json:"products"`
		Honors   []string `json:"honors"`
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
		Name, LogoURL, CoverURL, Description, ContactInfo string
		Products []string `json:"products"`
		Honors   []string `json:"honors"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	p, err := s.portfolioSvc.Update(r.PathValue("id"), in.Name, in.LogoURL, in.CoverURL, in.Description, in.ContactInfo, in.Products, in.Honors)
	if err != nil {
		fail(w, r, http.StatusNotFound, err)
		return
	}
	respond(w, r, http.StatusOK, p)
}

// ---- Achievements ----

// GET /api/v1/achievements?field=智能巡检&page=1&page_size=10
func (s *Server) listAchievements(w http.ResponseWriter, r *http.Request) {
	page, pageSize := paginationFromQuery(r)
	items, total, err := s.achievementSvc.List(r.URL.Query().Get("field"), page, pageSize)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	paginatedRespond(w, r, items, total)
}

// POST /api/v1/achievements
func (s *Server) createAchievement(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	var in struct {
		Title, AchieveType, Description, Field, Stage, ContactInfo string
		Images                                                      []string `json:"images"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	ach, err := s.achievementSvc.Create(a.ID, in.Title, in.AchieveType, in.Description, in.Field, in.Stage, in.ContactInfo, in.Images)
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
		Title, AchieveType, Description, Field, Stage, ContactInfo string
		Images                                                      []string `json:"images"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	ach, err := s.achievementSvc.Update(r.PathValue("id"), in.Title, in.AchieveType, in.Description, in.Field, in.Stage, in.ContactInfo, in.Images)
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
	page, pageSize := paginationFromQuery(r)
	items, total, err := s.rdService.List(r.URL.Query().Get("field"), page, pageSize)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	paginatedRespond(w, r, items, total)
}

// POST /api/v1/rd-challenges
func (s *Server) createRDChallenge(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	var in struct {
		Title, Field, Description, Deadline string
		BudgetFen                           int64 `json:"budget_fen"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	deadline, _ := time.Parse("2006-01-02", in.Deadline)
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
		Title, Field, Description, Deadline string
		BudgetFen                           int64 `json:"budget_fen"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	deadline, _ := time.Parse("2006-01-02", in.Deadline)
	ch, err := s.rdService.Update(r.PathValue("id"), in.Title, in.Field, in.Description, in.BudgetFen, deadline)
	if err != nil {
		fail(w, r, http.StatusNotFound, err)
		return
	}
	respond(w, r, http.StatusOK, ch)
}

// ---- Research Projects ----

// GET /api/v1/research-projects?page=1&page_size=10
func (s *Server) listResearchProjects(w http.ResponseWriter, r *http.Request) {
	page, pageSize := paginationFromQuery(r)
	items, total, err := s.researchSvc.List(page, pageSize)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	paginatedRespond(w, r, items, total)
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
		Title, Field, Description, LeadOrg, Milestones, StartDate, EndDate string
		Members                                                              []string `json:"members"`
		BudgetFen                                                            int64    `json:"budget_fen"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	startDate, _ := time.Parse("2006-01-02", in.StartDate)
	endDate, _ := time.Parse("2006-01-02", in.EndDate)
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
		Title, Field, Description, LeadOrg, Milestones, StartDate, EndDate string
		Members                                                              []string `json:"members"`
		BudgetFen                                                            int64    `json:"budget_fen"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	startDate, _ := time.Parse("2006-01-02", in.StartDate)
	endDate, _ := time.Parse("2006-01-02", in.EndDate)
	p, err := s.researchSvc.Update(r.PathValue("id"), in.Title, in.Field, in.Description, in.LeadOrg, in.Milestones, in.Members, in.BudgetFen, startDate, endDate)
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
	page, pageSize := paginationFromQuery(r)
	items, total, err := s.projectAppSvc.ListAll(r.URL.Query().Get("status"), page, pageSize)
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
		BudgetFen                           int64    `json:"budget_fen"`
		Attachments                         []string `json:"attachments"`
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

// GET /api/v1/competitions?page=1&page_size=10
func (s *Server) listCompetitions(w http.ResponseWriter, r *http.Request) {
	page, pageSize := paginationFromQuery(r)
	items, total, err := s.competitionSvc.List(page, pageSize)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	paginatedRespond(w, r, items, total)
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
		Title, Category, Description, Location, Sponsor, StartDate, EndDate string
		MaxTeams                                                             int `json:"max_teams"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	startDate, _ := time.Parse("2006-01-02", in.StartDate)
	endDate, _ := time.Parse("2006-01-02", in.EndDate)
	c, err := s.competitionSvc.Create(in.Title, in.Category, in.Description, in.Location, in.Sponsor, startDate, endDate, in.MaxTeams)
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
	page, pageSize := paginationFromQuery(r)
	items, total, err := s.eventSvc.List(page, pageSize)
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
		Title, EventType, Description, Location, CoverURL, StartTime, EndTime string
		MaxAttendees                                                           int `json:"max_attendees"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	startTime, _ := time.Parse("2006-01-02T15:04", in.StartTime)
	endTime, _ := time.Parse("2006-01-02T15:04", in.EndTime)
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
func (s *Server) listIndustryResources(w http.ResponseWriter, r *http.Request) {
	page, pageSize := paginationFromQuery(r)
	items, total, err := s.resourceSvc.List(r.URL.Query().Get("res_type"), page, pageSize)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	paginatedRespond(w, r, items, total)
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
		Name, ResType, Model, Specs, Location, BookingInfo string
		PriceFen                                            int64 `json:"price_fen"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	res, err := s.resourceSvc.Create(a.ID, in.Name, in.ResType, in.Model, in.Specs, in.Location, in.BookingInfo, in.PriceFen)
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
		Name, ResType, Model, Specs, Location, BookingInfo string
		PriceFen                                            int64 `json:"price_fen"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	res, err := s.resourceSvc.Update(r.PathValue("id"), in.Name, in.ResType, in.Model, in.Specs, in.Location, in.BookingInfo, in.PriceFen)
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
		Name, ResType, Specs, Location, ContactInfo string
		Quantity                                     int `json:"quantity"`
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

// GET /api/v1/emergency-dispatches?page=1&page_size=10
func (s *Server) listEmergencyDispatches(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		fail(w, r, http.StatusForbidden, errors.New("admin permission required"))
		return
	}
	page, pageSize := paginationFromQuery(r)
	items, total, err := s.emergencySvc.ListDispatches(page, pageSize)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
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
		ResourceID, EventDesc, Location, Commander, Result, StartTime, EndTime string
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	startTime, _ := time.Parse("2006-01-02T15:04", in.StartTime)
	endTime, _ := time.Parse("2006-01-02T15:04", in.EndTime)
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
	if limit <= 0 { limit = 20 }
	userID := ""
	if a, ok := authenticatedActor(r); ok { userID = a.ID }
	results, err := s.matchingSvc.Recommend(userID, lat, lng, bizType, district, limit)
	if err != nil { fail(w, r, http.StatusInternalServerError, err); return }
	respond(w, r, http.StatusOK, results)
}

// GET /api/v1/match?q=巡检&biz_type=cable_inspection&lat=29.5&lng=106.5
func (s *Server) searchAndMatch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" { fail(w, r, http.StatusBadRequest, errors.New("q parameter required")); return }
	bizType := r.URL.Query().Get("biz_type")
	lat, _ := strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
	lng, _ := strconv.ParseFloat(r.URL.Query().Get("lng"), 64)
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 { limit = 20 }
	results, err := s.matchingSvc.SearchAndMatch(q, lat, lng, bizType, limit)
	if err != nil { fail(w, r, http.StatusInternalServerError, err); return }
	respond(w, r, http.StatusOK, results)
}
