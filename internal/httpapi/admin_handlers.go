package httpapi

import (
	"fmt"
	"net/http"
	"strconv"
)

// ---- Admin List Handlers (批量补全) ----

func parsePagination(r *http.Request) (int, int) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if page < 1 { page = 1 }
	if pageSize < 1 || pageSize > 100 { pageSize = 20 }
	return page, pageSize
}

// listAdminResources GET /api/v1/admin/resources
func (s *Server) listAdminResources(w http.ResponseWriter, r *http.Request) {
	page, pageSize := parsePagination(r)
	all, total, err := s.resourceSvc.List("", page, pageSize)
	if err != nil { fail(w, r, 500, fmt.Errorf("list resources: %w", err)); return }
	paginatedRespond(w, r, all, total)
}

// listAdminJobs GET /api/v1/admin/jobs
func (s *Server) listAdminJobs(w http.ResponseWriter, r *http.Request) {
	page, pageSize := parsePagination(r)
	offset := (page - 1) * pageSize
	all, total, err := s.jobSvc.ListPublishedJobs(offset, pageSize)
	if err != nil { fail(w, r, 500, fmt.Errorf("list jobs: %w", err)); return }
	paginatedRespond(w, r, all, total)
}

// listAdminColleges GET /api/v1/admin/colleges
func (s *Server) listAdminColleges(w http.ResponseWriter, r *http.Request) {
	all, err := s.collegeSvc.List("")
	if err != nil { fail(w, r, 500, fmt.Errorf("list colleges: %w", err)); return }
	paginatedRespond(w, r, all, len(all))
}

// listAdminStudy GET /api/v1/admin/study
func (s *Server) listAdminStudy(w http.ResponseWriter, r *http.Request) {
	paginatedRespond(w, r, []struct{}{}, 0)
}

// listAdminTestSites GET /api/v1/admin/testsites
func (s *Server) listAdminTestSites(w http.ResponseWriter, r *http.Request) {
	all, err := s.testSiteSvc.List("")
	if err != nil { fail(w, r, 500, fmt.Errorf("list test sites: %w", err)); return }
	paginatedRespond(w, r, all, len(all))
}

// listAdminTransformations GET /api/v1/admin/transformations
func (s *Server) listAdminTransformations(w http.ResponseWriter, r *http.Request) {
	all, err := s.transSvc.List("")
	if err != nil { fail(w, r, 500, fmt.Errorf("list transformations: %w", err)); return }
	paginatedRespond(w, r, all, len(all))
}

// listAdminEvents GET /api/v1/admin/events
func (s *Server) listAdminEvents(w http.ResponseWriter, r *http.Request) {
	page, pageSize := parsePagination(r)
	all, total, err := s.eventSvc.List(page, pageSize)
	if err != nil { fail(w, r, 500, fmt.Errorf("list events: %w", err)); return }
	paginatedRespond(w, r, all, total)
}

// listAdminExhibitions GET /api/v1/admin/exhibitions
func (s *Server) listAdminExhibitions(w http.ResponseWriter, r *http.Request) {
	page, pageSize := parsePagination(r)
	all, total, err := s.exhibitionSvc.List(page, pageSize)
	if err != nil { fail(w, r, 500, fmt.Errorf("list exhibitions: %w", err)); return }
	paginatedRespond(w, r, all, total)
}

// listAdminEmergencyResources GET /api/v1/admin/emergency-resources
func (s *Server) listAdminEmergencyResources(w http.ResponseWriter, r *http.Request) {
	page, pageSize := parsePagination(r)
	all, total, err := s.emergencySvc.ListResources(page, pageSize)
	if err != nil { fail(w, r, 500, fmt.Errorf("list emergency resources: %w", err)); return }
	paginatedRespond(w, r, all, total)
}

// listAdminEmergencyDispatches GET /api/v1/admin/emergency-dispatches
func (s *Server) listAdminEmergencyDispatches(w http.ResponseWriter, r *http.Request) {
	page, pageSize := parsePagination(r)
	all, total, err := s.emergencySvc.ListDispatches(page, pageSize)
	if err != nil { fail(w, r, 500, fmt.Errorf("list dispatches: %w", err)); return }
	paginatedRespond(w, r, all, total)
}

// listAdminCerts GET /api/v1/admin/certs
func (s *Server) listAdminCerts(w http.ResponseWriter, r *http.Request) {
	certs, err := s.trainingSvc.ListAllCertificates()
	if err != nil { fail(w, r, 500, fmt.Errorf("list certs: %w", err)); return }
	paginatedRespond(w, r, certs, len(certs))
}

// listAdminMessages GET /api/v1/admin/messages
func (s *Server) listAdminMessages(w http.ResponseWriter, r *http.Request) {
	paginatedRespond(w, r, []struct{}{}, 0)
}
