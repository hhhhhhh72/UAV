package httpapi

import "net/http"

// registerAdminListRoutes adds GET/PUT/DELETE admin routes for all modules.
// POST (create) routes are registered in other files; we DO NOT duplicate them
// because Go 1.22+ mux panics on duplicate METHOD+PATH patterns.
func (s *Server) registerAdminListRoutes(mux *http.ServeMux) {
	// === 培训课程 === (POST: POST /api/v1/training-courses in training.go)
	mux.HandleFunc("GET /api/v1/admin/training-courses", s.listCourses)
	mux.HandleFunc("PUT /api/v1/admin/training-courses/{id}", s.updateCourse)
	mux.HandleFunc("DELETE /api/v1/admin/training-courses/{id}", s.deleteCourse)

	// === 证书 === (POST: POST /api/v1/certificates in training.go)
	mux.HandleFunc("GET /api/v1/admin/certificates", s.listAdminCerts)
	mux.HandleFunc("PUT /api/v1/admin/certificates/{id}", s.updateCertificate)
	mux.HandleFunc("DELETE /api/v1/admin/certificates/{id}", s.deleteCertificate)

	// === 职位 === (POST: POST /api/v1/jobs in jobs.go)
	mux.HandleFunc("GET /api/v1/admin/jobs", s.listAdminJobs)
	mux.HandleFunc("PUT /api/v1/admin/jobs/{id}", s.updateJob)
	mux.HandleFunc("DELETE /api/v1/admin/jobs/{id}", s.deleteJob)

	// === 院校 === (POST: POST /api/v1/admin/colleges in batch2_handlers.go)
	mux.HandleFunc("GET /api/v1/admin/colleges", s.listAdminColleges)
	mux.HandleFunc("PUT /api/v1/admin/colleges/{id}", s.updateCollege)
	mux.HandleFunc("DELETE /api/v1/admin/colleges/{id}", s.deleteCollege)

	// === 研学 === (no existing POST)
	mux.HandleFunc("GET /api/v1/admin/study-tours", s.listAdminStudy)
	mux.HandleFunc("POST /api/v1/admin/study-tours", s.createStudyTour)
	mux.HandleFunc("PUT /api/v1/admin/study-tours/{id}", s.updateStudyTour)
	mux.HandleFunc("DELETE /api/v1/admin/study-tours/{id}", s.deleteStudyTour)

	// === 成果 === (POST: POST /api/v1/achievements in biz_handlers.go — public path, so admin POST is new)
	mux.HandleFunc("GET /api/v1/admin/achievements", s.listAchievements)
	mux.HandleFunc("POST /api/v1/admin/achievements", s.createAchievement)
	mux.HandleFunc("PUT /api/v1/admin/achievements/{id}", s.updateAchievement)
	mux.HandleFunc("DELETE /api/v1/admin/achievements/{id}", s.deleteAchievement)

	// === 研发难题 === (POST: POST /api/v1/rd-challenges in biz_handlers.go — public path)
	mux.HandleFunc("GET /api/v1/admin/rd-challenges", s.listRDChallenges)
	mux.HandleFunc("POST /api/v1/admin/rd-challenges", s.createRDChallenge)
	mux.HandleFunc("PUT /api/v1/admin/rd-challenges/{id}", s.updateRDChallenge)
	mux.HandleFunc("DELETE /api/v1/admin/rd-challenges/{id}", s.deleteRDChallenge)

	// === 课题攻关 === (POST+PUT already in biz_handlers.go — skip duplicates)
	mux.HandleFunc("GET /api/v1/admin/research-projects", s.listResearchProjects)
	mux.HandleFunc("DELETE /api/v1/admin/research-projects/{id}", s.deleteResearchProject)

	// === 测试场地 === (POST: POST /api/v1/admin/test-sites in batch1_handlers.go — DUPLICATE, SKIP)
	mux.HandleFunc("GET /api/v1/admin/test-sites", s.listAdminTestSites)
	mux.HandleFunc("PUT /api/v1/admin/test-sites/{id}", s.updateTestSite)
	mux.HandleFunc("DELETE /api/v1/admin/test-sites/{id}", s.deleteTestSite)

	// === 成果转化 === (POST: POST /api/v1/admin/transformations in batch2_handlers.go — DUPLICATE, SKIP)
	mux.HandleFunc("GET /api/v1/admin/transformations", s.listAdminTransformations)
	mux.HandleFunc("PUT /api/v1/admin/transformations/{id}", s.updateTransformation)
	mux.HandleFunc("DELETE /api/v1/admin/transformations/{id}", s.deleteTransformation)

	// === 活动 === (POST: POST /api/v1/admin/events in biz_handlers.go — DUPLICATE, SKIP)
	mux.HandleFunc("GET /api/v1/admin/events", s.listAdminEvents)
	mux.HandleFunc("PUT /api/v1/admin/events/{id}", s.updateEvent)
	mux.HandleFunc("DELETE /api/v1/admin/events/{id}", s.deleteEvent)

	// === 品牌 === (POST: POST /api/v1/portfolios in biz_handlers.go — public path)
	mux.HandleFunc("GET /api/v1/admin/portfolios", s.listPortfolios)
	mux.HandleFunc("POST /api/v1/admin/portfolios", s.createPortfolio)
	mux.HandleFunc("PUT /api/v1/admin/portfolios/{id}", s.updatePortfolio)
	mux.HandleFunc("DELETE /api/v1/admin/portfolios/{id}", s.deletePortfolio)

	// === 展会 === (POST: POST /api/v1/admin/exhibitions in batch1_handlers.go — DUPLICATE, SKIP)
	mux.HandleFunc("GET /api/v1/admin/exhibitions", s.listAdminExhibitions)
	mux.HandleFunc("PUT /api/v1/admin/exhibitions/{id}", s.updateExhibition)
	mux.HandleFunc("DELETE /api/v1/admin/exhibitions/{id}", s.deleteExhibition)

	// === 行业报告 === (POST+DELETE already in biz_handlers.go — skip duplicates)
	mux.HandleFunc("GET /api/v1/admin/industry-reports", s.listIndustryReports)
	mux.HandleFunc("PUT /api/v1/admin/industry-reports/{id}", s.updateIndustryReport)

	// === 应急资源 === (POST: POST /api/v1/admin/emergency-resources in biz_handlers.go — DUPLICATE, SKIP)
	mux.HandleFunc("GET /api/v1/admin/emergency-resources", s.listAdminEmergencyResources)
	mux.HandleFunc("PUT /api/v1/admin/emergency-resources/{id}", s.updateEmergencyResource)
	mux.HandleFunc("DELETE /api/v1/admin/emergency-resources/{id}", s.deleteEmergencyResource)

	// === 应急调度 === (POST: POST /api/v1/admin/emergency-dispatches in biz_handlers.go — DUPLICATE, SKIP)
	mux.HandleFunc("GET /api/v1/admin/emergency-dispatches", s.listAdminEmergencyDispatches)
	mux.HandleFunc("PUT /api/v1/admin/emergency-dispatches/{id}", s.updateEmergencyDispatch)
	mux.HandleFunc("DELETE /api/v1/admin/emergency-dispatches/{id}", s.deleteEmergencyDispatch)

	// === 消息通知 === (no existing POST)
	mux.HandleFunc("GET /api/v1/admin/messages", s.listAdminMessages)
	mux.HandleFunc("POST /api/v1/admin/messages", s.createMessage)
	mux.HandleFunc("PUT /api/v1/admin/messages/{id}", s.updateMessage)
	mux.HandleFunc("DELETE /api/v1/admin/messages/{id}", s.deleteMessage)

	// === 合规文档 === (POST: POST /api/v1/admin/compliance-docs in biz_handlers.go — DUPLICATE, SKIP)
	mux.HandleFunc("GET /api/v1/admin/compliance-docs", s.listComplianceDocs)
	mux.HandleFunc("PUT /api/v1/admin/compliance-docs/{id}", s.updateComplianceDoc)
	mux.HandleFunc("DELETE /api/v1/admin/compliance-docs/{id}", s.deleteComplianceDoc)

	// === 团体标准 === (POST: POST /api/v1/admin/compliance-standards in biz_handlers.go — DUPLICATE, SKIP)
	mux.HandleFunc("GET /api/v1/admin/compliance-standards", s.listComplianceStandards)
	mux.HandleFunc("PUT /api/v1/admin/compliance-standards/{id}", s.updateComplianceStandard)
	mux.HandleFunc("DELETE /api/v1/admin/compliance-standards/{id}", s.deleteComplianceStandard)

	// === 产业资源 === (POST+PUT already in biz_handlers.go — skip duplicates)
	mux.HandleFunc("GET /api/v1/admin/industry-resources", s.listAdminResources)
	mux.HandleFunc("DELETE /api/v1/admin/industry-resources/{id}", s.deleteIndustryResource)

	// === 订单管理 === (fill gap)
	mux.HandleFunc("GET /api/v1/admin/orders", s.listAdminOrders)

	// === 案例管理 === (fill gap)
	mux.HandleFunc("GET /api/v1/admin/cases", s.listAdminCaseEntries)

	// === 专家管理 admin === (fill gap)
	mux.HandleFunc("GET /api/v1/admin/experts", s.listAdminExperts)

	// === 赛事管理 === (fill gap)
	mux.HandleFunc("GET /api/v1/admin/competitions", s.listAdminCompetitions)
}
