package httpapi

import "net/http"

// registerPublicAPIRoutes registers public read-only API routes for the mini-program.
func (s *Server) registerPublicAPIRoutes(mux *http.ServeMux) {
	// 人才教育
	mux.HandleFunc("GET /api/v1/training/courses", s.publicListCourses)
	mux.HandleFunc("GET /api/v1/training/certificates", s.publicListCerts)
	mux.HandleFunc("GET /api/v1/study/tours", s.publicListStudyTours)
	// 创新
	mux.HandleFunc("GET /api/v1/rd/challenges", s.publicListRD)
	mux.HandleFunc("GET /api/v1/research/projects", s.publicListResearch)
	mux.HandleFunc("GET /api/v1/test/sites", s.publicListTestSites)
	// 研学详情：冷启动/分享直达自取（此前仅列表入 storage 传递，直达即误判不存在）
	mux.HandleFunc("GET /api/v1/study/tours/{id}", s.publicGetStudyTour)
	// 应急
	mux.HandleFunc("GET /api/v1/emergency/resources", s.publicListEmergency)
	// 行业
	mux.HandleFunc("GET /api/v1/industry/reports", s.publicListReports)
	mux.HandleFunc("GET /api/v1/industry/resources", s.publicListIndustryResources)
	// 服务
	mux.HandleFunc("GET /api/v1/services", s.publicListServices)
}
