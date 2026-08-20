package httpapi

import (
	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
	"net/http"
	"strings"
)

// ── public handlers for mini-program ──

func (s *Server) publicListCourses(w http.ResponseWriter, r *http.Request) {
	all, err := s.trainingSvc.ListCourses(r.Context())
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, 200, all)
}

func (s *Server) publicListCerts(w http.ResponseWriter, r *http.Request) {
	all, err := s.trainingSvc.ListAllCertificates(r.Context())
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	// P2 修复：匿名接口此前返回全部用户的证书台账（cert_number/user_id，含 pending）。
	// 与飞手公开名录（listPilots）一致：仅公开已审核通过（approved）的证书，
	// 并脱敏 user_id（不暴露持有者身份）。
	out := make([]domain.Certificate, 0, len(all))
	for _, c := range all {
		if c.Status != "approved" {
			continue
		}
		c.UserID = ""
		out = append(out, c)
	}
	respond(w, r, http.StatusOK, out)
}

func (s *Server) publicListStudyTours(w http.ResponseWriter, r *http.Request) {
	all, err := s.studyTourRepo.List(r.Context())
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, 200, all)
}

func (s *Server) publicListRD(w http.ResponseWriter, r *http.Request) {
	page, size := paginationFromQuery(r)
	all, _, err := s.rdService.List(r.Context(), "", page, size)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, 200, all)
}

func (s *Server) publicListResearch(w http.ResponseWriter, r *http.Request) {
	page, size := paginationFromQuery(r)
	all, _, err := s.researchSvc.List(r.Context(), page, size)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, 200, all)
}

func (s *Server) publicListTestSites(w http.ResponseWriter, r *http.Request) {
	all, err := s.testSiteSvc.List(r.Context(), "")
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, 200, all)
}

func (s *Server) publicListEmergency(w http.ResponseWriter, r *http.Request) {
	page, size := paginationFromQuery(r)
	all, _, err := s.emergencySvc.ListResources(r.Context(), "", "", page, size)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, 200, all)
}

func (s *Server) publicListReports(w http.ResponseWriter, r *http.Request) {
	page, size := paginationFromQuery(r)
	all, _, err := s.reportSvc.List(r.Context(), page, size)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, 200, all)
}

func (s *Server) publicListIndustryResources(w http.ResponseWriter, r *http.Request) {
	// P2 修复：匿名接口此前未做 VisibilityLevel 过滤，绕过主列表/详情的分级控制
	// （任意未登录用户可见 member/partner/admin 级资源台账）。与 listIndustryResources
	// 一致按访问者级别过滤——匿名访客级别为 public(0)，只能看到 public 级资源。
	all, _, err := s.resourceSvc.List(r.Context(), "", 1, 100000)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	visitor := s.visitorResourceLevel(r)
	visible := make([]domain.IndustryResource, 0, len(all))
	for _, it := range all {
		if resourceLevelRank(it.VisibilityLevel) <= visitor {
			visible = append(visible, it)
		}
	}
	paginatedRespond(w, r, visible, len(visible))
}

func (s *Server) publicListServices(w http.ResponseWriter, r *http.Request) {
	// services = jobs + demands summary
	jobs, _, err := s.jobSvc.ListPublishedJobs(r.Context(), 0, 100)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
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
	if dem, err := s.demands.List(r.Context(), repository.DemandFilter{}); err == nil {
		for _, d := range dem {
			if q == "" || contains(d.Title, q) {
				results = append(results, map[string]any{"type": "demand", "id": d.ID, "title": d.Title})
			}
		}
	}
	if jobs, _, err := s.jobSvc.ListPublishedJobs(r.Context(), 0, 100); err == nil {
		for _, j := range jobs {
			if q == "" || contains(j.Title, q) {
				results = append(results, map[string]any{"type": "job", "id": j.ID, "title": j.Title})
			}
		}
	}
	// paginate
	if offset < len(results) {
		end := offset + size
		if end > len(results) {
			end = len(results)
		}
		results = results[offset:end]
	} else {
		results = nil
	}
	respond(w, r, 200, results)
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
