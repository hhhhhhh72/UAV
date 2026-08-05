package httpapi

import "net/http"

// ── Core route registration methods ──────────────────────────────────────
//
// Each method registers routes for one business module.
// This keeps server.go focused on the Server struct and middleware,
// while routes live alongside their handlers.

func (s *Server) registerCoreRoutes(mux *http.ServeMux) {
	s.registerMetaRoutes(mux)
	s.registerDemandRoutes(mux)
	s.registerEnterpriseRoutes(mux)
	s.registerEmploymentContractRoutes(mux)
	s.registerJobRoutes(mux)
	s.registerCommunityRoutes(mux)
	s.registerListingRoutes(mux)
	s.registerLabourRoutes(mux)
	s.registerTrainingRoutes(mux)
	s.registerTradingRoutes(mux)
	s.registerInsuranceRoutes(mux)
	s.registerFinanceRoutes(mux)
	s.registerMessageRoutes(mux)
	s.registerMiscRoutes(mux)
	s.registerFileRoutes(mux)
	s.registerAdminRoutes(mux)
	s.registerAuthRoutes(mux)
}

// ── Meta / Health ────────────────────────────────────────────────────────

func (s *Server) registerMetaRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /{$}", s.index)
	mux.HandleFunc("GET /uploads/", s.serveUploads)
	mux.HandleFunc("GET /favicon.ico", s.favicon)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		respond(w, r, http.StatusOK, map[string]any{
			"status": "ok",
			"checks": map[string]string{
				"server":  "up",
				"storage": s.storage,
			},
		})
	})
	mux.HandleFunc("GET /api/v1/home", s.home)
	mux.HandleFunc("GET /api/v1/search", s.search)
}

// ── Demands ──────────────────────────────────────────────────────────────

func (s *Server) registerDemandRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/demands", s.listDemands)
	mux.HandleFunc("GET /api/v1/demands/{id}", s.demandDetail)
	mux.HandleFunc("POST /api/v1/demands", s.createDemand)
	mux.HandleFunc("PATCH /api/v1/demands/{id}", s.updateDemand)
	mux.HandleFunc("POST /api/v1/demands/{id}/submit", s.submitDemand)
	mux.HandleFunc("POST /api/v1/demands/{id}/complete", s.completeDemand)
	mux.HandleFunc("POST /api/v1/demands/{id}/cancel", s.cancelDemand)
	mux.HandleFunc("GET /api/v1/admin/demands", s.listAdminDemands)
	mux.HandleFunc("POST /api/v1/admin/demands/{id}/review", s.reviewDemand)
	mux.HandleFunc("POST /api/v1/admin/demands/{id}/approve", s.approveDemand)
	mux.HandleFunc("POST /api/v1/admin/demands/{id}/close", s.closeDemand)
	mux.HandleFunc("POST /api/v1/admin/demands/{id}/amount", s.setDemandOfflineAmount)
	mux.HandleFunc("PATCH /api/v1/admin/demands/{id}", s.adminUpdateDemand)
}

// ── Enterprises ──────────────────────────────────────────────────────────

func (s *Server) registerEnterpriseRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/enterprises/public", s.listPublicEnterprises)
	mux.HandleFunc("GET /api/v1/enterprises", s.listMyEnterprises)
	mux.HandleFunc("POST /api/v1/enterprises", s.createEnterprise)
	mux.HandleFunc("PATCH /api/v1/enterprises/{id}", s.updateEnterprise)
	mux.HandleFunc("POST /api/v1/enterprises/{id}/submit", s.submitEnterprise)
	mux.HandleFunc("GET /api/v1/admin/enterprises", s.listEnterprises)
	mux.HandleFunc("GET /api/v1/admin/enterprises/pending", s.pendingEnterprises)
	mux.HandleFunc("POST /api/v1/admin/enterprises/{id}/review", s.reviewEnterprise)
	mux.HandleFunc("POST /api/v1/admin/enterprises/batch-review", s.batchReviewEnterprises)
	mux.HandleFunc("GET /api/v1/admin/enterprises/search", s.searchEnterprises)
}

// ── Employment & Contracts ───────────────────────────────────────────────

func (s *Server) registerEmploymentContractRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/employment-requests", s.listEmployment)
	mux.HandleFunc("POST /api/v1/employment-requests", s.createEmployment)
	mux.HandleFunc("GET /api/v1/contracts", s.listContracts)
	mux.HandleFunc("POST /api/v1/contracts", s.createContract)
	mux.HandleFunc("POST /api/v1/contracts/{id}/void", s.voidContract)
	mux.HandleFunc("POST /api/v1/webhooks/signing", s.signingWebhook)
}

// ── Jobs / Resumes / Applications ────────────────────────────────────────

func (s *Server) registerJobRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/jobs", s.createJob)
	mux.HandleFunc("GET /api/v1/jobs", s.listJobs)
	mux.HandleFunc("GET /api/v1/jobs/mine", s.listMyJobs)
	mux.HandleFunc("GET /api/v1/jobs/{id}", s.getJob)
	mux.HandleFunc("POST /api/v1/jobs/{id}/publish", s.publishJob)
	mux.HandleFunc("POST /api/v1/jobs/{id}/close", s.closeJob)
	mux.HandleFunc("POST /api/v1/resumes", s.createResume)
	mux.HandleFunc("PATCH /api/v1/resumes/{id}", s.updateResume)
	mux.HandleFunc("GET /api/v1/resumes/mine", s.listMyResumes)
	mux.HandleFunc("POST /api/v1/applications", s.createApplication)
	mux.HandleFunc("PATCH /api/v1/applications/{id}/status", s.updateApplicationStatus)
	mux.HandleFunc("GET /api/v1/applications", s.listApplications)
}

// ── Community (Posts / Comments / Reports) ───────────────────────────────

func (s *Server) registerCommunityRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/posts", s.createPost)
	mux.HandleFunc("GET /api/v1/posts", s.listPosts)
	mux.HandleFunc("POST /api/v1/posts/{id}/publish", s.publishPost)
	mux.HandleFunc("POST /api/v1/posts/{id}/remove", s.removePost)
	mux.HandleFunc("POST /api/v1/posts/{id}/comments", s.createComment)
	mux.HandleFunc("GET /api/v1/comments", s.listComments)
	mux.HandleFunc("POST /api/v1/reports", s.createReport)
	mux.HandleFunc("GET /api/v1/admin/reports", s.listReports)
}

// ── Listings ─────────────────────────────────────────────────────────────

func (s *Server) registerListingRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/listings", s.createListing)
	mux.HandleFunc("GET /api/v1/listings", s.listListings)
	mux.HandleFunc("POST /api/v1/listings/{id}/close", s.closeListing)
	mux.HandleFunc("POST /api/v1/listings/{id}/favorites", s.favoriteListing)
}

// ── Labour Orders ────────────────────────────────────────────────────────

func (s *Server) registerLabourRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/labour-orders", s.createLabourOrder)
	mux.HandleFunc("GET /api/v1/labour-orders", s.listLabourOrders)
	mux.HandleFunc("POST /api/v1/labour-orders/{id}/quote", s.createLabourQuote)
	mux.HandleFunc("GET /api/v1/labour-orders/quotes", s.listLabourQuotes)
}

// ── Training ─────────────────────────────────────────────────────────────

func (s *Server) registerTrainingRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/certificates", s.addCertificate)
	mux.HandleFunc("POST /api/v1/admin/certificates/{id}/approve", s.approveCertificate)
	mux.HandleFunc("GET /api/v1/certificates/mine", s.listMyCertificates)
	mux.HandleFunc("POST /api/v1/training-courses", s.createCourse)
	mux.HandleFunc("GET /api/v1/training-courses", s.listCourses)
	mux.HandleFunc("POST /api/v1/instructors", s.registerInstructor)
	mux.HandleFunc("POST /api/v1/admin/instructors/{id}/approve", s.approveInstructor)
	mux.HandleFunc("GET /api/v1/instructors", s.listInstructors)
	mux.HandleFunc("POST /api/v1/certified-pilots", s.registerPilot)
	mux.HandleFunc("GET /api/v1/admin/certified-pilots", s.listAdminPilots)
	mux.HandleFunc("POST /api/v1/admin/certified-pilots/{id}/approve", s.approvePilot)
	mux.HandleFunc("POST /api/v1/admin/certified-pilots/{id}/reject", s.rejectPilot)
	mux.HandleFunc("GET /api/v1/certified-pilots", s.listPilots)
	mux.HandleFunc("GET /api/v1/certified-pilots/mine", s.getMyPilot)
}

// ── Trading ──────────────────────────────────────────────────────────────

func (s *Server) registerTradingRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/products", s.createProduct)
	mux.HandleFunc("GET /api/v1/products", s.listProducts)
	mux.HandleFunc("GET /api/v1/products/{id}", s.getProductDetail)
	mux.HandleFunc("POST /api/v1/repairs", s.createRepair)
	mux.HandleFunc("GET /api/v1/repairs/mine", s.listMyRepairs)
}

// ── Insurance ────────────────────────────────────────────────────────────

func (s *Server) registerInsuranceRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/policies", s.createPolicy)
	mux.HandleFunc("GET /api/v1/policies/mine", s.listMyPolicies)
	mux.HandleFunc("POST /api/v1/inspections", s.createInspection)
	mux.HandleFunc("GET /api/v1/inspections/mine", s.listMyInspections)
}

// ── Finance ──────────────────────────────────────────────────────────────

func (s *Server) registerFinanceRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/loans", s.applyLoan)
	mux.HandleFunc("GET /api/v1/loans/mine", s.listMyLoans)
}

// ── Messages ─────────────────────────────────────────────────────────────

func (s *Server) registerMessageRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/messages", s.listMessages)
	mux.HandleFunc("POST /api/v1/messages/{id}/read", s.markMessageRead)
	mux.HandleFunc("GET /api/v1/messages/unread-count", s.unreadCount)
}

// ── Misc (templates, import, assignments) ────────────────────────────────

func (s *Server) registerMiscRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/contract-templates", s.listContractTemplates)
	mux.HandleFunc("POST /api/v1/admin/members/import", s.importMembers)
	mux.HandleFunc("POST /api/v1/assignments", s.createAssignment)
}

// ── Files ────────────────────────────────────────────────────────────────

func (s *Server) registerFileRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/files/upload", s.uploadFile)
	mux.HandleFunc("POST /api/v1/enterprises/{id}/documents", s.attachEnterpriseDocument)
}

// ── Admin (config, export, image) ────────────────────────────────────────

func (s *Server) registerAdminRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/admin/token", s.adminDevLogin)
	mux.HandleFunc("GET /api/v1/admin/config", s.getConfig)
	mux.HandleFunc("POST /api/v1/admin/config", s.updateConfig)
	mux.HandleFunc("GET /api/v1/admin/export/demands", s.exportDemands)
	mux.HandleFunc("GET /api/v1/admin/export/enterprises", s.exportEnterprises)
	mux.HandleFunc("POST /api/v1/admin/demands/batch-approve", s.batchApproveDemands)
	mux.HandleFunc("GET /api/v1/image", s.serveImage)
}

// ── Auth ─────────────────────────────────────────────────────────────────

func (s *Server) registerAuthRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/auth/wechat/login", s.wechatLogin)
	mux.HandleFunc("POST /api/v1/auth/refresh", s.refreshToken)
	mux.HandleFunc("POST /api/v1/auth/logout", s.logout)
	mux.HandleFunc("GET /api/v1/me", s.me)
	mux.HandleFunc("PATCH /api/v1/me", s.updateMe)
}
