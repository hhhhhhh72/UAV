package httpapi

import "net/http"

// registerAdminListRoutes adds GET/PUT/DELETE admin routes for all modules.
// POST (create) routes are registered in other files; we DO NOT duplicate them
// because Go 1.22+ mux panics on duplicate METHOD+PATH patterns.
func (s *Server) registerAdminListRoutes(mux *http.ServeMux) {
	// === 培训课程 === (POST: POST /api/v1/training-courses in training.go)
	mux.HandleFunc("GET /api/v1/admin/training-courses", s.listCourses)
	mux.HandleFunc("GET /api/v1/admin/training-courses/{id}", s.getCourse)
	mux.HandleFunc("POST /api/v1/admin/training-courses", s.adminCreateCourse)
	mux.HandleFunc("PUT /api/v1/admin/training-courses/{id}", s.updateCourse)
	mux.HandleFunc("DELETE /api/v1/admin/training-courses/{id}", s.deleteCourse)

	// === 证书 === (POST: POST /api/v1/certificates in training.go)
	mux.HandleFunc("GET /api/v1/admin/certificates", s.listAdminCerts)
	// === 报名记录 ===
	mux.HandleFunc("GET /api/v1/admin/enrollments", s.listAdminEnrollments)
	mux.HandleFunc("GET /api/v1/admin/certificates/{id}", s.getCert)
	mux.HandleFunc("POST /api/v1/admin/certificates", s.adminCreateCertificate)
	mux.HandleFunc("PUT /api/v1/admin/certificates/{id}", s.updateCertificate)
	mux.HandleFunc("DELETE /api/v1/admin/certificates/{id}", s.deleteCertificate)

	// === 职位 === (POST: POST /api/v1/jobs in jobs.go)
	mux.HandleFunc("GET /api/v1/admin/jobs", s.listAdminJobs)
	mux.HandleFunc("GET /api/v1/admin/jobs/{id}", s.getJob)
	mux.HandleFunc("POST /api/v1/admin/jobs", s.adminCreateJob)
	mux.HandleFunc("PUT /api/v1/admin/jobs/{id}", s.updateJob)
	mux.HandleFunc("DELETE /api/v1/admin/jobs/{id}", s.deleteJob)

	// === 院校 === (POST: POST /api/v1/admin/colleges in batch2_handlers.go)
	mux.HandleFunc("GET /api/v1/admin/colleges", s.listAdminColleges)
	mux.HandleFunc("GET /api/v1/admin/colleges/{id}", s.getCollege)
	mux.HandleFunc("PUT /api/v1/admin/colleges/{id}", s.updateCollege)
	mux.HandleFunc("DELETE /api/v1/admin/colleges/{id}", s.deleteCollege)

	// === 服务能力 === (PRD ②-2 供给能力展示)
	mux.HandleFunc("GET /api/v1/admin/service-listings", s.adminListServiceListings)
	mux.HandleFunc("POST /api/v1/admin/service-listings", s.adminCreateServiceListing)
	mux.HandleFunc("PUT /api/v1/admin/service-listings/{id}", s.adminUpdateServiceListing)
	mux.HandleFunc("DELETE /api/v1/admin/service-listings/{id}", s.adminDeleteServiceListing)

	// === 研学 === (no existing POST)
	mux.HandleFunc("GET /api/v1/admin/study-tours", s.listAdminStudy)
	mux.HandleFunc("GET /api/v1/admin/study-tours/{id}", s.getStudyTour)
	mux.HandleFunc("POST /api/v1/admin/study-tours", s.createStudyTour)
	mux.HandleFunc("PUT /api/v1/admin/study-tours/{id}", s.updateStudyTour)
	mux.HandleFunc("DELETE /api/v1/admin/study-tours/{id}", s.deleteStudyTour)

	// === 成果 === (POST: POST /api/v1/achievements in biz_handlers.go — public path, so admin POST is new)
	mux.HandleFunc("GET /api/v1/admin/achievements", s.listAchievements)
	mux.HandleFunc("GET /api/v1/admin/achievements/{id}", s.getAchievement)
	mux.HandleFunc("POST /api/v1/admin/achievements", s.createAchievement)
	mux.HandleFunc("PUT /api/v1/admin/achievements/{id}", s.updateAchievement)
	mux.HandleFunc("DELETE /api/v1/admin/achievements/{id}", s.deleteAchievement)

	// === 研发难题 === (POST: POST /api/v1/rd-challenges in biz_handlers.go — public path)
	mux.HandleFunc("GET /api/v1/admin/rd-challenges", s.listRDChallenges)
	mux.HandleFunc("GET /api/v1/admin/rd-challenges/{id}", s.getRDChallenge)
	mux.HandleFunc("POST /api/v1/admin/rd-challenges", s.createRDChallenge)
	mux.HandleFunc("PUT /api/v1/admin/rd-challenges/{id}", s.updateRDChallenge)
	mux.HandleFunc("DELETE /api/v1/admin/rd-challenges/{id}", s.deleteRDChallenge)

	// === 课题攻关 === (POST+PUT already in biz_handlers.go — skip duplicates)
	mux.HandleFunc("GET /api/v1/admin/research-projects", s.listResearchProjects)
	mux.HandleFunc("GET /api/v1/admin/research-projects/{id}", s.getResearchProject)
	mux.HandleFunc("DELETE /api/v1/admin/research-projects/{id}", s.deleteResearchProject)

	// === 测试场地 === (POST: POST /api/v1/admin/test-sites in batch1_handlers.go — DUPLICATE, SKIP)
	mux.HandleFunc("GET /api/v1/admin/test-sites", s.listAdminTestSites)
	mux.HandleFunc("GET /api/v1/admin/test-sites/bookings", s.listAdminBookings)
	mux.HandleFunc("GET /api/v1/admin/test-sites/{id}", s.getTestSite)
	mux.HandleFunc("PUT /api/v1/admin/test-sites/{id}", s.updateTestSite)
	mux.HandleFunc("DELETE /api/v1/admin/test-sites/{id}", s.deleteTestSite)

	// === 成果转化 === (POST: POST /api/v1/admin/transformations in batch2_handlers.go — DUPLICATE, SKIP)
	mux.HandleFunc("GET /api/v1/admin/transformations", s.listAdminTransformations)
	mux.HandleFunc("GET /api/v1/admin/transformations/{id}", s.getTransformation)
	mux.HandleFunc("PUT /api/v1/admin/transformations/{id}", s.updateTransformation)
	mux.HandleFunc("DELETE /api/v1/admin/transformations/{id}", s.deleteTransformation)

	// === 活动 === (POST: POST /api/v1/admin/events in biz_handlers.go — DUPLICATE, SKIP)
	mux.HandleFunc("GET /api/v1/admin/events", s.listAdminEvents)
	mux.HandleFunc("GET /api/v1/admin/events/{id}", s.getEvent)
	mux.HandleFunc("PUT /api/v1/admin/events/{id}", s.updateEvent)
	mux.HandleFunc("DELETE /api/v1/admin/events/{id}", s.deleteEvent)

	// === 品牌 === (POST: POST /api/v1/portfolios in biz_handlers.go — public path)
	mux.HandleFunc("GET /api/v1/admin/portfolios", s.listAdminPortfolios)
	mux.HandleFunc("GET /api/v1/admin/portfolios/{id}", s.getPortfolio)
	mux.HandleFunc("POST /api/v1/admin/portfolios", s.createPortfolio)
	mux.HandleFunc("PUT /api/v1/admin/portfolios/{id}", s.updatePortfolio)
	mux.HandleFunc("DELETE /api/v1/admin/portfolios/{id}", s.deletePortfolio)

	// === 展会 === (POST: POST /api/v1/admin/exhibitions in batch1_handlers.go — DUPLICATE, SKIP)
	mux.HandleFunc("GET /api/v1/admin/exhibitions", s.listAdminExhibitions)
	mux.HandleFunc("GET /api/v1/admin/exhibitions/{id}", s.getExhibition)
	mux.HandleFunc("PUT /api/v1/admin/exhibitions/{id}", s.updateExhibition)
	mux.HandleFunc("DELETE /api/v1/admin/exhibitions/{id}", s.deleteExhibition)

	// === 行业报告 === (POST+DELETE already in biz_handlers.go — skip duplicates)
	mux.HandleFunc("GET /api/v1/admin/industry-reports", s.listIndustryReports)
	mux.HandleFunc("GET /api/v1/admin/industry-reports/{id}", s.getReport)
	mux.HandleFunc("PUT /api/v1/admin/industry-reports/{id}", s.updateIndustryReport)

	// === 应急资源 === (POST: POST /api/v1/admin/emergency-resources in biz_handlers.go — DUPLICATE, SKIP)
	mux.HandleFunc("GET /api/v1/admin/emergency-resources", s.listAdminEmergencyResources)
	mux.HandleFunc("GET /api/v1/admin/emergency-resources/{id}", s.getEmergencyResource)
	mux.HandleFunc("PUT /api/v1/admin/emergency-resources/{id}", s.updateEmergencyResource)
	mux.HandleFunc("DELETE /api/v1/admin/emergency-resources/{id}", s.deleteEmergencyResource)

	// === 应急调度 === (POST: POST /api/v1/admin/emergency-dispatches in biz_handlers.go — DUPLICATE, SKIP)
	mux.HandleFunc("GET /api/v1/admin/emergency-dispatches", s.listAdminEmergencyDispatches)
	mux.HandleFunc("GET /api/v1/admin/emergency-dispatches/{id}", s.getEmergencyDispatch)
	mux.HandleFunc("PUT /api/v1/admin/emergency-dispatches/{id}", s.updateEmergencyDispatch)
	mux.HandleFunc("DELETE /api/v1/admin/emergency-dispatches/{id}", s.deleteEmergencyDispatch)

	// === 消息通知 === (no existing POST)
	mux.HandleFunc("GET /api/v1/admin/messages", s.listAdminMessages)
	mux.HandleFunc("GET /api/v1/admin/messages/{id}", s.getMessage)
	mux.HandleFunc("POST /api/v1/admin/messages", s.createMessage)
	mux.HandleFunc("PUT /api/v1/admin/messages/{id}", s.updateMessage)
	mux.HandleFunc("DELETE /api/v1/admin/messages/{id}", s.deleteMessage)

	// === 资讯 === (公开 GET /api/v1/articles 仅 published；管理端全量含草稿)
	mux.HandleFunc("GET /api/v1/admin/articles", s.listAdminArticles)

	// === 保险金融/维修/简历 管理端列表（P2-1 补齐，见 admin_lists_p2.go）===
	mux.HandleFunc("GET /api/v1/admin/policies", s.listAdminPolicies)
	mux.HandleFunc("GET /api/v1/admin/inspections", s.listAdminInspections)
	mux.HandleFunc("GET /api/v1/admin/repairs", s.listAdminRepairs)
	mux.HandleFunc("GET /api/v1/admin/loans", s.listAdminLoans)
	mux.HandleFunc("GET /api/v1/admin/resumes", s.listAdminResumes)

	// === 场地/讲师/校企/应急部门/救援案例：复用既有公开列表 handler ===
	mux.HandleFunc("GET /api/v1/admin/venues", s.listVenues)
	// 导师列表管理端专用（含 pending/rejected 待审，公开列表只出 approved）
	mux.HandleFunc("GET /api/v1/admin/instructors", s.listAdminInstructors)
	mux.HandleFunc("GET /api/v1/admin/cooperations", s.listCooperations)
	mux.HandleFunc("GET /api/v1/admin/emergency-depts", s.listEmergencyDepts)
	mux.HandleFunc("GET /api/v1/admin/rescue-cases", s.listRescueCases)

	// === 合规文档 === (POST: POST /api/v1/admin/compliance-docs in biz_handlers.go — DUPLICATE, SKIP)
	mux.HandleFunc("GET /api/v1/admin/compliance-docs", s.listComplianceDocs)
	mux.HandleFunc("GET /api/v1/admin/compliance-docs/{id}", s.getComplianceDoc)
	mux.HandleFunc("PUT /api/v1/admin/compliance-docs/{id}", s.updateComplianceDoc)
	mux.HandleFunc("DELETE /api/v1/admin/compliance-docs/{id}", s.deleteComplianceDoc)

	// === 团体标准 === (POST: POST /api/v1/admin/compliance-standards in biz_handlers.go — DUPLICATE, SKIP)
	mux.HandleFunc("GET /api/v1/admin/compliance-standards", s.listComplianceStandards)
	mux.HandleFunc("GET /api/v1/admin/compliance-standards/{id}", s.getComplianceStandard)
	mux.HandleFunc("PUT /api/v1/admin/compliance-standards/{id}", s.updateComplianceStandard)
	mux.HandleFunc("DELETE /api/v1/admin/compliance-standards/{id}", s.deleteComplianceStandard)

	// === 产业资源 === (POST+PUT already in biz_handlers.go — skip duplicates)
	mux.HandleFunc("GET /api/v1/admin/industry-resources", s.listAdminResources)
	mux.HandleFunc("GET /api/v1/admin/industry-resources/{id}", s.getIndustryResource)
	mux.HandleFunc("DELETE /api/v1/admin/industry-resources/{id}", s.deleteIndustryResource)

	// === 订单管理 === (fill gap)
	mux.HandleFunc("GET /api/v1/admin/orders", s.listAdminOrders)
	mux.HandleFunc("GET /api/v1/admin/orders/{id}", s.getOrder)
	mux.HandleFunc("POST /api/v1/admin/orders", s.createOrder)
	mux.HandleFunc("PUT /api/v1/admin/orders/{id}", s.updateOrder)
	mux.HandleFunc("PUT /api/v1/admin/orders/{id}/aftersale", s.reviewAftersale)
	mux.HandleFunc("DELETE /api/v1/admin/orders/{id}", s.deleteOrder)

	// === 商品管理 === (商城上架)
	mux.HandleFunc("GET /api/v1/admin/products", s.listAdminProducts)
	mux.HandleFunc("GET /api/v1/admin/products/{id}", s.getProductDetail)
	mux.HandleFunc("POST /api/v1/admin/products", s.adminCreateProduct)
	mux.HandleFunc("PUT /api/v1/admin/products/{id}", s.adminUpdateProduct)
	mux.HandleFunc("DELETE /api/v1/admin/products/{id}", s.adminDeleteProduct)

	// === 案例管理 === (fill gap)
	mux.HandleFunc("GET /api/v1/admin/cases", s.listAdminCaseEntries)

	// === 专家管理 admin === (fill gap)
	mux.HandleFunc("GET /api/v1/admin/experts", s.listAdminExperts)
	mux.HandleFunc("GET /api/v1/admin/experts/{id}", s.getExpert)

	// === 赛事管理 === (fill gap)
	mux.HandleFunc("GET /api/v1/admin/competitions", s.listAdminCompetitions)
	mux.HandleFunc("GET /api/v1/admin/competitions/{id}", s.getCompetition)
	mux.HandleFunc("PUT /api/v1/admin/competitions/{id}", s.updateCompetition)
	mux.HandleFunc("DELETE /api/v1/admin/competitions/{id}", s.deleteCompetition)

	// === 社区帖子 === (POST: POST /api/v1/posts in community_listings_labour.go — public path;
	// 审核闭环：CreatePost 默认 pending，管理端经此接口查看全量（含待审），publish 用 POST /api/v1/posts/{id}/publish)
	mux.HandleFunc("GET /api/v1/admin/posts", s.listAdminPosts)
}
