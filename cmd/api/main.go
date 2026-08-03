package main

// @title           无人机产业综合服务平台 API
// @version         1.0
// @description     面向微信小程序与 Web 管理后台的全栈服务平台，覆盖无人机产业链 7 大业务系统。
// @description     ## 认证
// @description     Bearer Token (HMAC-SHA256)，15 分钟过期。
// @description     在请求头中添加 `Authorization: Bearer <token>`。
// @description     ## 分页
// @description     分页接口统一使用 `?page=1&page_size=20`，返回 `{ data, total, page, page_size }`。
// @description     ## 角色
// @description     - `platform_admin` 平台管理员
// @description     - `association_admin` 协会管理员
// @description     - `enterprise` 企业用户
// @description     - `individual` 个人用户
// @contact.name    重庆市无人机产业协会
// @host            localhost:8080
// @BasePath        /
// @schemes         http
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"drone-platform/internal/config"
	"drone-platform/internal/crypto"
	"drone-platform/internal/domain"
	"drone-platform/internal/httpapi"
	"drone-platform/internal/logger"
	"drone-platform/internal/repository"
	"drone-platform/internal/repository/memory"
	"drone-platform/internal/repository/postgres"
	"drone-platform/internal/service"
)

func main() {
	cfg := config.Load()
	logger.Init(cfg.Server.Env)
	val := cfg.Validate()
	if len(val.Errors) > 0 {
		for _, e := range val.Errors {
			slog.Error(e)
		}
		os.Exit(1)
	}
	for _, w := range val.Warnings {
		slog.Warn(w)
	}
	cfg.Print()

	addr := cfg.Server.Port
	secret := cfg.JWT.Secret

	// 加密密钥（可选，未设置时密文不加密，仅开发环境使用）
	var cipher *crypto.Cipher
	if ek := os.Getenv("ENCRYPTION_KEY"); ek != "" {
		c, err := crypto.NewCipher(ek)
		if err != nil {
			slog.Error("invalid ENCRYPTION_KEY", "error", err)
			os.Exit(1)
		}
		cipher = c
		slog.Info("encryption enabled")
	} else {
		slog.Warn("ENCRYPTION_KEY not set, sensitive fields stored in plaintext (NOT FOR PRODUCTION)")
	}

	tokens, err := httpapi.NewTokenManager(secret)
	if err != nil {
		slog.Error("invalid AUTH_SECRET", "error", err)
		os.Exit(1)
	}

	var (
		demandRepo       repository.DemandRepository
		enterpriseRepo   repository.EnterpriseRepository
		employmentRepo   repository.EmploymentRepository
		contractRepo     repository.ContractRepository
		jobRepo          repository.JobRepository
		resumeRepo       repository.ResumeRepository
		appRepo          repository.JobApplicationRepository
		postRepo         repository.PostRepository
		commentRepo      repository.CommentRepository
		reportRepo       repository.ReportRepository
		listingRepo      repository.ListingRepository
		labourRepo       repository.LabourOrderRepository
		userRepo         repository.UserRepository
		refreshTokenRepo repository.RefreshTokenRepository
		certRepo         repository.CertificateRepository
		courseRepo       repository.CourseRepository
		instructorRepo   repository.InstructorRepository
		pilotRepo        repository.PilotRepository
		productRepo      repository.ProductRepository
		repairRepo       repository.RepairRepository
		policyRepo       repository.PolicyRepository
		inspectRepo      repository.InspectionRepository
		loanRepo         repository.LoanRepository
		msgRepo          repository.MessageRepository
		articleRepo      repository.ArticleRepository
		reviewRepo       repository.ReviewRepository
		venueRepo        repository.VenueRepository
		enrollRepo       repository.EnrollmentRepository
		tradeOrderRepo   repository.TradeOrderRepository
		escrowRepo       repository.EscrowRepository
		// Memory-only (coop/rescue/emergDept/assocMember — complex sub-methods).
		poolRepo        = memory.NewResourcePoolRepository()
		coopRepo        = memory.NewCooperationRepository()
		rescueCaseRepo  = memory.NewRescueCaseRepository()
		emergDeptRepo   = memory.NewEmergencyDeptRepository()
		testSiteRepo    repository.TestSiteRepository
		transRepo       repository.TransformationRepository
		exhibitionRepo  repository.ExhibitionRepository
		collegeRepo     repository.CollegeRepository
		studyTourRepo   repository.StudyTourRepository
		assocMemberRepo = memory.NewAssociationMemberRepository()
		expertRepo      repository.ExpertRepository
		shopRepo        repository.ShopRepository
		caseRepo        repository.CaseRepository
		complianceRepo  repository.ComplianceRepository
		achieveRepo     repository.AchievementRepository
		rdChallengeRepo repository.RDChallengeRepository
		researchRepo    repository.ResearchProjectRepository
		projAppRepo     repository.ProjectAppRepository
		competitionRepo repository.CompetitionRepository
		eventRepo       repository.EventRepository
		portfolioRepo   repository.PortfolioRepository
		industryRptRepo repository.IndustryReportRepository
		resourceRepo    repository.ResourceRepository
		emergencyRepo   repository.EmergencyRepository
		pgStore         *postgres.Store
	)

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		pgStore, err = postgres.NewStore(ctx, databaseURL, cipher)
		if err != nil {
			slog.Error("connect to postgres failed", "error", err)
			os.Exit(1)
		}
		defer pgStore.Close()

		if err := pgStore.RunMigrationsFromDir(ctx, postgres.MigrationsDir()); err != nil {
			slog.Error("migrations failed", "error", err)
			os.Exit(1)
		}
		slog.Info("postgres connected, migrations applied")

		demandRepo = pgStore.NewDemandRepository()
		enterpriseRepo = pgStore.NewEnterpriseRepository()
		employmentRepo = pgStore.NewEmploymentRepository()
		contractRepo = pgStore.NewContractRepository()
		jobRepo = pgStore.NewJobRepository()
		resumeRepo = pgStore.NewResumeRepository()
		appRepo = pgStore.NewJobApplicationRepository()
		postRepo = pgStore.NewPostRepository()
		testSiteRepo = pgStore.NewTestSiteRepository()
		transRepo = pgStore.NewTransformationRepository()
		commentRepo = pgStore.NewCommentRepository()
		reportRepo = pgStore.NewReportRepository()
		listingRepo = pgStore.NewListingRepository()
		labourRepo = pgStore.NewLabourOrderRepository()
		userRepo = pgStore.NewUserRepository()
		certRepo = pgStore.NewCertificateRepository()
		courseRepo = pgStore.NewCourseRepository()
		instructorRepo = pgStore.NewInstructorRepository()
		pilotRepo = pgStore.NewPilotRepository(cipher)
		productRepo = pgStore.NewProductRepository()
		repairRepo = pgStore.NewRepairRepository()
		policyRepo = pgStore.NewPolicyRepository()
		inspectRepo = pgStore.NewInspectionRepository()
		loanRepo = pgStore.NewLoanRepository()
		msgRepo = pgStore.NewMessageRepository()
		articleRepo = pgStore.NewArticleRepository()
		reviewRepo = pgStore.NewReviewRepository()
		venueRepo = pgStore.NewVenueRepository()
		enrollRepo = pgStore.NewEnrollmentRepository()
		tradeOrderRepo = pgStore.NewTradeOrderRepository()
		escrowRepo = pgStore.NewEscrowRepository()
		refreshTokenRepo = pgStore.NewRefreshTokenRepository()
		expertRepo = pgStore.NewExpertRepository()
		shopRepo = pgStore.NewShopRepository()
		caseRepo = pgStore.NewCaseRepository()
		complianceRepo = pgStore.NewComplianceRepository()
		achieveRepo = pgStore.NewAchievementRepository()
		rdChallengeRepo = pgStore.NewRDChallengeRepository()
		researchRepo = pgStore.NewResearchProjectRepository()
		projAppRepo = pgStore.NewProjectAppRepository()
		competitionRepo = pgStore.NewCompetitionRepository()
		eventRepo = pgStore.NewEventRepository()
		portfolioRepo = pgStore.NewPortfolioRepository()
		industryRptRepo = pgStore.NewIndustryReportRepository()
		resourceRepo = pgStore.NewResourceRepository()
		emergencyRepo = pgStore.NewEmergencyRepository()

		exhibitionRepo = pgStore.NewExhibitionRepository()
		collegeRepo = pgStore.NewCollegeRepository()
		studyTourRepo = pgStore.NewStudyTourRepository()
		poolRepo = pgStore.NewResourcePoolRepository()
		coopRepo = pgStore.NewCooperationRepository()
		rescueCaseRepo = pgStore.NewRescueCaseRepository()
		emergDeptRepo = pgStore.NewEmergencyDeptRepository()
		assocMemberRepo = pgStore.NewAssociationMemberRepository()
	} else {
		slog.Warn("DATABASE_URL not set, using in-memory storage (NOT FOR PRODUCTION)")
		demandRepo = memory.NewDemandRepository(cipher)
		enterpriseRepo = memory.NewEnterpriseRepository(cipher)
		employmentRepo = memory.NewEmploymentRepository()
		contractRepo = memory.NewContractRepository()
		jobRepo = memory.NewJobRepository()
		resumeRepo = memory.NewResumeRepository()
		appRepo = memory.NewJobApplicationRepository()
		postRepo = memory.NewPostRepository()
		commentRepo = memory.NewCommentRepository()
		reportRepo = memory.NewReportRepository()
		listingRepo = memory.NewListingRepository()
		labourRepo = memory.NewLabourOrderRepository()
		userRepo = memory.NewUserRepository(cipher)
		refreshTokenRepo = memory.NewRefreshTokenRepository()
		certRepo = memory.NewCertificateRepository()
		courseRepo = memory.NewCourseRepository()
		instructorRepo = memory.NewInstructorRepository()
		pilotRepo = memory.NewPilotRepository(cipher)
		productRepo = memory.NewProductRepository()
		repairRepo = memory.NewRepairRepository()
		policyRepo = memory.NewPolicyRepository()
		inspectRepo = memory.NewInspectionRepository()
		loanRepo = memory.NewLoanRepository()
		msgRepo = memory.NewMessageRepository()
		articleRepo = memory.NewArticleRepository()
		reviewRepo = memory.NewReviewRepository()
		venueRepo = memory.NewVenueRepository()
		enrollRepo = memory.NewEnrollmentRepository()
		tradeOrderRepo = memory.NewTradeOrderRepository()
		escrowRepo = memory.NewEscrowRepository()
		coopRepo = memory.NewCooperationRepository()
		rescueCaseRepo = memory.NewRescueCaseRepository()
		emergDeptRepo = memory.NewEmergencyDeptRepository()
		assocMemberRepo = memory.NewAssociationMemberRepository()
		testSiteRepo = memory.NewTestSiteRepository()
		transRepo = memory.NewTransformationRepository()
		exhibitionRepo = memory.NewExhibitionRepository()
		collegeRepo = memory.NewCollegeRepository()
		studyTourRepo = memory.NewStudyTourRepository()
		expertRepo = memory.NewExpertRepository()
		shopRepo = memory.NewShopRepository()
		caseRepo = memory.NewCaseRepository()
		complianceRepo = memory.NewComplianceRepository()
		achieveRepo = memory.NewAchievementRepository()
		rdChallengeRepo = memory.NewRDChallengeRepository()
		researchRepo = memory.NewResearchProjectRepository()
		projAppRepo = memory.NewProjectAppRepository()
		competitionRepo = memory.NewCompetitionRepository()
		eventRepo = memory.NewEventRepository()
		portfolioRepo = memory.NewPortfolioRepository()
		industryRptRepo = memory.NewIndustryReportRepository()
		resourceRepo = memory.NewResourceRepository()
		emergencyRepo = memory.NewEmergencyRepository()
	}

	app := httpapi.NewServer(
		service.NewDemandService(demandRepo),
		service.NewEnterpriseService(enterpriseRepo),
		service.NewEnterpriseSvc(enterpriseRepo),
		service.NewEmploymentService(employmentRepo),
		service.NewContractService(contractRepo),
		service.NewJobService(jobRepo, resumeRepo, appRepo),
		service.NewCommunityService(postRepo, commentRepo, reportRepo),
		service.NewListingService(listingRepo),
		service.NewLabourService(labourRepo),
		service.NewTrainingService(certRepo, courseRepo, instructorRepo, pilotRepo),
		service.NewTradingService(productRepo, repairRepo),
		service.NewInsuranceService(policyRepo, inspectRepo),
		service.NewFinanceService(loanRepo),
		service.NewHomeService(demandRepo),
		service.NewFileService("uploads/"),
		service.NewMessageService(msgRepo),
		service.NewEnrollmentService(enrollRepo),
		service.NewExpiryService(),
		service.NewTradeOrderService(tradeOrderRepo),
		service.NewEscrowService(escrowRepo),
		service.NewNewsService(articleRepo),
		service.NewReviewService(reviewRepo),
		service.NewVenueService(venueRepo),
		userRepo,
		refreshTokenRepo,
		tokens,
	)

	// ── Wire extended services ──────────────────────────────────────
	// PG-backed services: wired for both storage modes.
	app.SetExpertService(service.NewExpertService(expertRepo))
	app.SetShopService(service.NewShopService(shopRepo))
	app.SetCaseService(service.NewCaseService(caseRepo))
	app.SetComplianceService(service.NewComplianceService(complianceRepo))
	app.SetReportService(service.NewReportService(industryRptRepo))
	app.SetPortfolioService(service.NewPortfolioService(portfolioRepo))
	app.SetAchievementService(service.NewAchievementService(achieveRepo))
	app.SetRDChallengeService(service.NewRDChallengeService(rdChallengeRepo))
	app.SetResearchProjectService(service.NewResearchProjectService(researchRepo))
	app.SetProjectAppService(service.NewProjectAppService(projAppRepo))
	app.SetCompetitionService(service.NewCompetitionService(competitionRepo))
	app.SetEventService(service.NewEventService(eventRepo))
	app.SetResourceService(service.NewResourceService(resourceRepo))
	app.SetEmergencyService(service.NewEmergencyService(emergencyRepo))
	app.SetMatchingService(service.NewMatchingService(demandRepo))

	// Memory-only repos: PG implementations pending (see docs/项目管理/项目审计报告).
	app.SetRescueCaseService(service.NewRescueCaseService(rescueCaseRepo))
	app.SetEmergencyDeptService(service.NewEmergencyDeptService(emergDeptRepo))
	app.SetAssociationMemberService(service.NewAssociationMemberService(assocMemberRepo))
	app.SetTransformationService(service.NewTransformationService(transRepo))
	app.SetCollegeService(service.NewCollegeService(collegeRepo))
	app.SetStudyTourRepo(studyTourRepo)
	app.SetCooperationService(service.NewCooperationService(coopRepo))
	app.SetPoolService(service.NewResourcePoolService(poolRepo))
	app.SetTestSiteService(service.NewTestSiteService(testSiteRepo))
	app.SetExhibitionService(service.NewExhibitionService(exhibitionRepo))

	if pgStore != nil {
		app.SetAuditWriter(postgres.NewAuditAdapter(pgStore))
		app.SetStorage("postgres")
		slog.Warn("PG mode: 4 repos lack PG impl, using in-memory (coop/rescueCase/emergDept/assocMember)")
	} else {
		app.SetStorage("memory")
	}

	// Seed super admin user from SUPER_ADMIN_PHONE env var.
	superPhone := os.Getenv("SUPER_ADMIN_PHONE")
	if superPhone != "" {
		if _, err := userRepo.FindByID(superPhone); err != nil {
			now := time.Now()
			userRepo.Create(domain.User{
				ID: superPhone, WechatOpenID: superPhone, Role: domain.RolePlatformAdmin,
				Status: "active", Version: 1, CreatedAt: now, UpdatedAt: now,
			})
			slog.Info("seeded super admin", "phone", superPhone)
		}
	}

	server := &http.Server{Addr: addr, Handler: app.Router(), ReadTimeout: 30 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second}
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		slog.Error("server forced shutdown", "error", err)
	}
	slog.Info("server exited")
}
