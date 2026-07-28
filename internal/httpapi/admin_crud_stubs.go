package httpapi

import (
	"fmt"
	"net/http"
	"strconv"
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
	paginatedRespond(w, r, []any{}, 0)
}

// ----- Reviews -----
func (s *Server) listAdminReviews(w http.ResponseWriter, r *http.Request) {
	paginatedRespond(w, r, []any{}, 0)
}

// ----- Case entries -----
func (s *Server) listAdminCaseEntries(w http.ResponseWriter, r *http.Request) {
	paginatedRespond(w, r, []any{}, 0)
}

// ----- Experts admin wrapper -----
func (s *Server) listAdminExperts(w http.ResponseWriter, r *http.Request) {
	s.listExperts(w, r)
}

// ----- Competitions -----
func (s *Server) listAdminCompetitions(w http.ResponseWriter, r *http.Request) {
	paginatedRespond(w, r, []any{}, 0)
}

// --- Training Courses (missing update/delete) ---
func (s *Server) updateCourse(w http.ResponseWriter, r *http.Request) {
	respond(w, r, 200, map[string]string{"status": "ok", "note": "updateCourse stub"})
}
func (s *Server) deleteCourse(w http.ResponseWriter, r *http.Request) {
	respond(w, r, 200, map[string]string{"status": "deleted", "note": "deleteCourse stub"})
}

// --- Certificates (missing admin list/update/delete) ---
func (s *Server) listAdminCerts(w http.ResponseWriter, r *http.Request) {
	certs, err := s.trainingSvc.ListAllCertificates()
	if err != nil { fail(w, r, 500, fmt.Errorf("list certs: %w", err)); return }
	paginatedRespond(w, r, certs, len(certs))
}
func (s *Server) updateCertificate(w http.ResponseWriter, r *http.Request) {
	respond(w, r, 200, map[string]string{"status": "ok", "note": "updateCertificate stub"})
}
func (s *Server) deleteCertificate(w http.ResponseWriter, r *http.Request) {
	respond(w, r, 200, map[string]string{"status": "deleted", "note": "deleteCertificate stub"})
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
	respond(w, r, 200, map[string]string{"status": "ok", "note": "updateJob stub"})
}
func (s *Server) deleteJob(w http.ResponseWriter, r *http.Request) {
	respond(w, r, 200, map[string]string{"status": "deleted", "note": "deleteJob stub"})
}

// --- Colleges (missing admin list/update/delete) ---
func (s *Server) listAdminColleges(w http.ResponseWriter, r *http.Request) {
	all, err := s.collegeSvc.List("")
	if err != nil { fail(w, r, 500, fmt.Errorf("list colleges: %w", err)); return }
	paginatedRespond(w, r, all, len(all))
}
func (s *Server) updateCollege(w http.ResponseWriter, r *http.Request) {
	respond(w, r, 200, map[string]string{"status": "ok", "note": "updateCollege stub"})
}
func (s *Server) deleteCollege(w http.ResponseWriter, r *http.Request) {
	respond(w, r, 200, map[string]string{"status": "deleted", "note": "deleteCollege stub"})
}

// --- Study Tours (missing list/create/update/delete) ---
func (s *Server) listAdminStudy(w http.ResponseWriter, r *http.Request) {
	paginatedRespond(w, r, []struct{}{}, 0)
}
func (s *Server) createStudyTour(w http.ResponseWriter, r *http.Request) {
	respond(w, r, 201, map[string]string{"status": "created", "note": "createStudyTour stub"})
}
func (s *Server) updateStudyTour(w http.ResponseWriter, r *http.Request) {
	respond(w, r, 200, map[string]string{"status": "ok", "note": "updateStudyTour stub"})
}
func (s *Server) deleteStudyTour(w http.ResponseWriter, r *http.Request) {
	respond(w, r, 200, map[string]string{"status": "deleted", "note": "deleteStudyTour stub"})
}

// --- Test Sites (missing admin list/update/delete) ---
func (s *Server) listAdminTestSites(w http.ResponseWriter, r *http.Request) {
	all, err := s.testSiteSvc.List("")
	if err != nil { fail(w, r, 500, fmt.Errorf("list test sites: %w", err)); return }
	paginatedRespond(w, r, all, len(all))
}
func (s *Server) updateTestSite(w http.ResponseWriter, r *http.Request) {
	respond(w, r, 200, map[string]string{"status": "ok", "note": "updateTestSite stub"})
}
func (s *Server) deleteTestSite(w http.ResponseWriter, r *http.Request) {
	respond(w, r, 200, map[string]string{"status": "deleted", "note": "deleteTestSite stub"})
}

// --- Transformations (missing admin list/update/delete) ---
func (s *Server) listAdminTransformations(w http.ResponseWriter, r *http.Request) {
	all, err := s.transSvc.List("")
	if err != nil { fail(w, r, 500, fmt.Errorf("list transformations: %w", err)); return }
	paginatedRespond(w, r, all, len(all))
}
func (s *Server) updateTransformation(w http.ResponseWriter, r *http.Request) {
	respond(w, r, 200, map[string]string{"status": "ok", "note": "updateTransformation stub"})
}
func (s *Server) deleteTransformation(w http.ResponseWriter, r *http.Request) {
	respond(w, r, 200, map[string]string{"status": "deleted", "note": "deleteTransformation stub"})
}

// --- Events (missing admin list/update/delete) ---
func (s *Server) listAdminEvents(w http.ResponseWriter, r *http.Request) {
	page, pageSize := parsePagination(r)
	all, total, err := s.eventSvc.List(page, pageSize)
	if err != nil { fail(w, r, 500, fmt.Errorf("list events: %w", err)); return }
	paginatedRespond(w, r, all, total)
}
func (s *Server) updateEvent(w http.ResponseWriter, r *http.Request) {
	respond(w, r, 200, map[string]string{"status": "ok", "note": "updateEvent stub"})
}
func (s *Server) deleteEvent(w http.ResponseWriter, r *http.Request) {
	respond(w, r, 200, map[string]string{"status": "deleted", "note": "deleteEvent stub"})
}

// --- Portfolios (missing delete) ---
func (s *Server) deletePortfolio(w http.ResponseWriter, r *http.Request) {
	respond(w, r, 200, map[string]string{"status": "deleted", "note": "deletePortfolio stub"})
}

// --- Exhibitions (missing admin list/update/delete) ---
func (s *Server) listAdminExhibitions(w http.ResponseWriter, r *http.Request) {
	page, pageSize := parsePagination(r)
	all, total, err := s.exhibitionSvc.List(page, pageSize)
	if err != nil { fail(w, r, 500, fmt.Errorf("list exhibitions: %w", err)); return }
	paginatedRespond(w, r, all, total)
}
func (s *Server) updateExhibition(w http.ResponseWriter, r *http.Request) {
	respond(w, r, 200, map[string]string{"status": "ok", "note": "updateExhibition stub"})
}
func (s *Server) deleteExhibition(w http.ResponseWriter, r *http.Request) {
	respond(w, r, 200, map[string]string{"status": "deleted", "note": "deleteExhibition stub"})
}

// --- Industry Reports (missing update) ---
func (s *Server) updateIndustryReport(w http.ResponseWriter, r *http.Request) {
	respond(w, r, 200, map[string]string{"status": "ok", "note": "updateIndustryReport stub"})
}

// --- Emergency Resources (missing admin list/update/delete) ---
func (s *Server) listAdminEmergencyResources(w http.ResponseWriter, r *http.Request) {
	page, pageSize := parsePagination(r)
	all, total, err := s.emergencySvc.ListResources(page, pageSize)
	if err != nil { fail(w, r, 500, fmt.Errorf("list emergency resources: %w", err)); return }
	paginatedRespond(w, r, all, total)
}
func (s *Server) updateEmergencyResource(w http.ResponseWriter, r *http.Request) {
	respond(w, r, 200, map[string]string{"status": "ok", "note": "updateEmergencyResource stub"})
}
func (s *Server) deleteEmergencyResource(w http.ResponseWriter, r *http.Request) {
	respond(w, r, 200, map[string]string{"status": "deleted", "note": "deleteEmergencyResource stub"})
}

// --- Emergency Dispatches (missing admin list/update/delete) ---
func (s *Server) listAdminEmergencyDispatches(w http.ResponseWriter, r *http.Request) {
	page, pageSize := parsePagination(r)
	all, total, err := s.emergencySvc.ListDispatches(page, pageSize)
	if err != nil { fail(w, r, 500, fmt.Errorf("list dispatches: %w", err)); return }
	paginatedRespond(w, r, all, total)
}
func (s *Server) updateEmergencyDispatch(w http.ResponseWriter, r *http.Request) {
	respond(w, r, 200, map[string]string{"status": "ok", "note": "updateEmergencyDispatch stub"})
}
func (s *Server) deleteEmergencyDispatch(w http.ResponseWriter, r *http.Request) {
	respond(w, r, 200, map[string]string{"status": "deleted", "note": "deleteEmergencyDispatch stub"})
}

// --- Messages (missing admin list/create/update/delete) ---
func (s *Server) listAdminMessages(w http.ResponseWriter, r *http.Request) {
	paginatedRespond(w, r, []struct{}{}, 0)
}
func (s *Server) createMessage(w http.ResponseWriter, r *http.Request) {
	respond(w, r, 201, map[string]string{"status": "sent", "note": "createMessage stub"})
}
func (s *Server) updateMessage(w http.ResponseWriter, r *http.Request) {
	respond(w, r, 200, map[string]string{"status": "ok", "note": "updateMessage stub"})
}
func (s *Server) deleteMessage(w http.ResponseWriter, r *http.Request) {
	respond(w, r, 200, map[string]string{"status": "deleted", "note": "deleteMessage stub"})
}

// --- Compliance Docs (missing update/delete) ---
func (s *Server) updateComplianceDoc(w http.ResponseWriter, r *http.Request) {
	respond(w, r, 200, map[string]string{"status": "ok", "note": "updateComplianceDoc stub"})
}
func (s *Server) deleteComplianceDoc(w http.ResponseWriter, r *http.Request) {
	respond(w, r, 200, map[string]string{"status": "deleted", "note": "deleteComplianceDoc stub"})
}

// --- Compliance Standards (missing update/delete) ---
func (s *Server) updateComplianceStandard(w http.ResponseWriter, r *http.Request) {
	respond(w, r, 200, map[string]string{"status": "ok", "note": "updateComplianceStandard stub"})
}
func (s *Server) deleteComplianceStandard(w http.ResponseWriter, r *http.Request) {
	respond(w, r, 200, map[string]string{"status": "deleted", "note": "deleteComplianceStandard stub"})
}

// --- Industry Resources (missing admin list/delete) ---
func (s *Server) listAdminResources(w http.ResponseWriter, r *http.Request) {
	page, pageSize := parsePagination(r)
	all, total, err := s.resourceSvc.List("", page, pageSize)
	if err != nil { fail(w, r, 500, fmt.Errorf("list resources: %w", err)); return }
	paginatedRespond(w, r, all, total)
}
func (s *Server) deleteIndustryResource(w http.ResponseWriter, r *http.Request) {
	respond(w, r, 200, map[string]string{"status": "deleted", "note": "deleteIndustryResource stub"})
}

// --- RD Challenges (missing delete) ---
func (s *Server) deleteRDChallenge(w http.ResponseWriter, r *http.Request) {
	respond(w, r, 200, map[string]string{"status": "deleted", "note": "deleteRDChallenge stub"})
}

// --- Research Projects (missing delete) ---
func (s *Server) deleteResearchProject(w http.ResponseWriter, r *http.Request) {
	respond(w, r, 200, map[string]string{"status": "deleted", "note": "deleteResearchProject stub"})
}
