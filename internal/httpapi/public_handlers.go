package httpapi

import (
	"drone-platform/internal/repository"
	"net/http"
	"strings"
)

// ── public handlers for mini-program ──

func (s *Server) publicListCourses(w http.ResponseWriter, r *http.Request) {
	all, _ := s.trainingSvc.ListCourses()
	respond(w, r, 200, all)
}

func (s *Server) publicListCerts(w http.ResponseWriter, r *http.Request) {
	all, _ := s.trainingSvc.ListAllCertificates()
	respond(w, r, 200, all)
}

func (s *Server) publicListStudyTours(w http.ResponseWriter, r *http.Request) {
	all, _ := s.studyTourRepo.List()
	respond(w, r, 200, all)
}

func (s *Server) publicListRD(w http.ResponseWriter, r *http.Request) {
	page, size := paginationFromQuery(r)
	all, _, _ := s.rdService.List("", page, size)
	respond(w, r, 200, all)
}

func (s *Server) publicListResearch(w http.ResponseWriter, r *http.Request) {
	page, size := paginationFromQuery(r)
	all, _, _ := s.researchSvc.List(page, size)
	respond(w, r, 200, all)
}

func (s *Server) publicListTestSites(w http.ResponseWriter, r *http.Request) {
	all, _ := s.testSiteSvc.List("")
	respond(w, r, 200, all)
}

func (s *Server) publicListEmergency(w http.ResponseWriter, r *http.Request) {
	page, size := paginationFromQuery(r)
	all, _, _ := s.emergencySvc.ListResources(page, size)
	respond(w, r, 200, all)
}

func (s *Server) publicListReports(w http.ResponseWriter, r *http.Request) {
	page, size := paginationFromQuery(r)
	all, _, _ := s.reportSvc.List(page, size)
	respond(w, r, 200, all)
}

func (s *Server) publicListIndustryResources(w http.ResponseWriter, r *http.Request) {
	page, size := paginationFromQuery(r)
	all, _, _ := s.resourceSvc.List("", page, size)
	respond(w, r, 200, all)
}

func (s *Server) publicListServices(w http.ResponseWriter, r *http.Request) {
	// services = jobs + demands summary
	jobs, _, _ := s.jobSvc.ListPublishedJobs(0, 100)
	respond(w, r, 200, map[string]any{
		"jobs":    jobs,
		"demands": "see /api/v1/demands",
	})
}

func (s *Server) publicMatch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	page, size := paginationFromQuery(r)
	offset := (page - 1) * size
	// Search across multiple sources
	var results []map[string]any
	if dem, err := s.demands.List(repository.DemandFilter{}); err == nil {
		for _, d := range dem {
			if q == "" || contains(d.Title, q) {
				results = append(results, map[string]any{"type":"demand","id":d.ID,"title":d.Title})
			}
		}
	}
	if jobs, _, err := s.jobSvc.ListPublishedJobs(0, 100); err == nil {
		for _, j := range jobs {
			if q == "" || contains(j.Title, q) {
				results = append(results, map[string]any{"type":"job","id":j.ID,"title":j.Title})
			}
		}
	}
	// paginate
	if offset < len(results) {
		end := offset + size
		if end > len(results) { end = len(results) }
		results = results[offset:end]
	} else {
		results = nil
	}
	respond(w, r, 200, results)
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
