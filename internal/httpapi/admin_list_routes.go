package httpapi

import "net/http"

// registerAdminListRoutes adds GET list admin endpoints for all modules
// that currently only have POST/Create endpoints.
func (s *Server) registerAdminListRoutes(mux *http.ServeMux) {
	// --- 会员资源 ---
	mux.HandleFunc("GET /api/v1/admin/experts", s.listExperts)            // reuses public list handler
	mux.HandleFunc("GET /api/v1/admin/resources", s.listAdminResources)
	mux.HandleFunc("GET /api/v1/admin/compliance", s.listComplianceDocs)  // reuses public handler

	// --- 人才教育 ---
	mux.HandleFunc("GET /api/v1/admin/training", s.listCourses)
	mux.HandleFunc("GET /api/v1/admin/certs", s.listAdminCerts)
	mux.HandleFunc("GET /api/v1/admin/jobs", s.listAdminJobs)
	mux.HandleFunc("GET /api/v1/admin/colleges", s.listAdminColleges)
	mux.HandleFunc("GET /api/v1/admin/study", s.listAdminStudy)

	// --- 产学研 ---
	mux.HandleFunc("GET /api/v1/admin/achievements", s.listAchievements)      // reuses
	mux.HandleFunc("GET /api/v1/admin/challenges", s.listRDChallenges)        // reuses
	mux.HandleFunc("GET /api/v1/admin/projects", s.listResearchProjects)      // reuses
	mux.HandleFunc("GET /api/v1/admin/testsites", s.listAdminTestSites)
	mux.HandleFunc("GET /api/v1/admin/transformations", s.listAdminTransformations)

	// --- 活动品牌 ---
	mux.HandleFunc("GET /api/v1/admin/events", s.listAdminEvents)
	mux.HandleFunc("GET /api/v1/admin/portfolios", s.listPortfolios)          // reuses
	mux.HandleFunc("GET /api/v1/admin/exhibitions", s.listAdminExhibitions)

	// --- 应急协同 ---
	mux.HandleFunc("GET /api/v1/admin/emergency-resources", s.listAdminEmergencyResources)
	mux.HandleFunc("GET /api/v1/admin/emergency-dispatches", s.listAdminEmergencyDispatches)

	// --- 系统管理 ---
	mux.HandleFunc("GET /api/v1/admin/messages", s.listAdminMessages)
}
