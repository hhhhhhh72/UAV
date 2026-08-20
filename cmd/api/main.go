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
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"sync"
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

// rotateWriter 大小轮转的文件 writer：日志同时落盘时使用。
// 单文件超过 maxBytes 即滚动新文件（文件名带日期+序号），目录内最多保留 keep 份。
// 项目此前 logger 包的 writeFile 只被 logger.Info/Warn/Error 调用，
// 而全代码一律用 slog.*（直出 stdout），文件日志实为死代码——这里把文件输出
// 接进 slog handler（MultiWriter: stdout + 文件），文件不再只写不读。
type rotateWriter struct {
	mu       sync.Mutex
	dir      string
	maxBytes int64
	keep     int
	f        *os.File
	seq      int
	size     int64
}

func newRotateWriter(dir string) *rotateWriter {
	return &rotateWriter{dir: dir, maxBytes: 50 << 20, keep: 5}
}

func (rw *rotateWriter) Write(p []byte) (int, error) {
	rw.mu.Lock()
	defer rw.mu.Unlock()
	if rw.f == nil {
		if err := rw.open(); err != nil {
			return 0, err
		}
	}
	if rw.size+int64(len(p)) > rw.maxBytes {
		_ = rw.f.Close()
		rw.f = nil
		if err := rw.open(); err != nil {
			return 0, err
		}
	}
	n, err := rw.f.Write(p)
	rw.size += int64(n)
	return n, err
}

// open 打开下一个日志文件：跳过已超限的同名文件（进程重启后继续滚动），并清理旧文件。
func (rw *rotateWriter) open() error {
	for {
		rw.seq++
		name := filepath.Join(rw.dir, fmt.Sprintf("app-%s-%03d.log", time.Now().Format("2006-01-02"), rw.seq))
		f, err := os.OpenFile(name, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		st, err := f.Stat()
		if err != nil {
			_ = f.Close()
			return err
		}
		if st.Size() >= rw.maxBytes {
			_ = f.Close()
			continue // 该文件已满，滚动到下一个序号
		}
		rw.f = f
		rw.size = st.Size()
		rw.cleanup()
		return nil
	}
}

// cleanup 只保留最近 keep 份日志文件（按修改时间倒序）。
func (rw *rotateWriter) cleanup() {
	entries, _ := filepath.Glob(filepath.Join(rw.dir, "app-*.log"))
	if len(entries) <= rw.keep {
		return
	}
	sort.Slice(entries, func(i, j int) bool {
		fi, _ := os.Stat(entries[i])
		fj, _ := os.Stat(entries[j])
		if fi == nil || fj == nil {
			return false
		}
		return fi.ModTime().After(fj.ModTime())
	})
	for _, e := range entries[rw.keep:] {
		_ = os.Remove(e)
	}
}

func main() {
	cfg := config.Load()
	logger.Init(cfg.Server.Env)
	// 文件日志接线：stdout + LOG_DIR（默认 ./logs，容器 cwd=/ 即 /logs 卷）双写，
	// 大小轮转由 rotateWriter 负责；级别/格式沿用 logger.Init 的 env 约定。
	logDir := os.Getenv("LOG_DIR")
	if logDir == "" {
		logDir = "./logs"
	}
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		slog.Warn("create log dir failed, file logging disabled", "dir", logDir, "error", err)
	} else {
		rw := newRotateWriter(logDir)
		level := slog.LevelInfo
		if cfg.Server.Env == "development" || cfg.Server.Env == "dev" {
			level = slog.LevelDebug
		}
		switch os.Getenv("LOG_LEVEL") {
		case "error":
			level = slog.LevelError
		case "warn":
			level = slog.LevelWarn
		case "debug":
			level = slog.LevelDebug
		}
		opts := &slog.HandlerOptions{Level: level}
		var handler slog.Handler
		if cfg.Server.Env == "production" {
			handler = slog.NewJSONHandler(io.MultiWriter(os.Stdout, rw), opts)
		} else {
			handler = slog.NewTextHandler(io.MultiWriter(os.Stdout, rw), opts)
		}
		slog.SetDefault(slog.New(handler))
		slog.Info("file logging enabled", "dir", logDir, "max_bytes_per_file", rw.maxBytes, "keep", rw.keep)
	}
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
		demandRepo         repository.DemandRepository
		intentRepo         repository.IntentRepository
		workOrderRepo      repository.WorkOrderRepository
		enterpriseRepo     repository.EnterpriseRepository
		employmentRepo     repository.EmploymentRepository
		contractRepo       repository.ContractRepository
		contractTplRepo    repository.ContractTemplateRepository
		jobRepo            repository.JobRepository
		resumeRepo         repository.ResumeRepository
		appRepo            repository.JobApplicationRepository
		postRepo           repository.PostRepository
		commentRepo        repository.CommentRepository
		reportRepo         repository.ReportRepository
		listingRepo        repository.ListingRepository
		labourRepo         repository.LabourOrderRepository
		userRepo           repository.UserRepository
		refreshTokenRepo   repository.RefreshTokenRepository
		certRepo           repository.CertificateRepository
		courseRepo         repository.CourseRepository
		instructorRepo     repository.InstructorRepository
		pilotRepo          repository.PilotRepository
		productRepo        repository.ProductRepository
		serviceListingRepo repository.ServiceListingRepository
		repairRepo         repository.RepairRepository
		policyRepo         repository.PolicyRepository
		inspectRepo        repository.InspectionRepository
		loanRepo           repository.LoanRepository
		msgRepo            repository.MessageRepository
		articleRepo        repository.ArticleRepository
		reviewRepo         repository.ReviewRepository
		venueRepo          repository.VenueRepository
		enrollRepo         repository.EnrollmentRepository
		tradeOrderRepo     repository.TradeOrderRepository
		escrowRepo         repository.EscrowRepository
		// 资源池/校企/救援案例/应急部门/协会成员：PG 实现位于 batch3_repos.go，DATABASE_URL 分支下会替换为 PG 实现。
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
		caseRepo        repository.CaseRepository
		svcAppRepo      repository.ApplicationRepository
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
		uploadRepo      repository.UploadRepository
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
		contractTplRepo = pgStore.NewContractTemplateRepository()
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
		serviceListingRepo = pgStore.NewServiceListingRepository()
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
		caseRepo = pgStore.NewCaseRepository()
		svcAppRepo = pgStore.NewApplicationRepository()
		complianceRepo = pgStore.NewComplianceRepository()
		achieveRepo = pgStore.NewAchievementRepository()
		rdChallengeRepo = pgStore.NewRDChallengeRepository()
		researchRepo = pgStore.NewResearchProjectRepository()
		projAppRepo = pgStore.NewProjectAppRepository()
		competitionRepo = pgStore.NewCompetitionRepository(cipher)
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
		intentRepo = pgStore.NewIntentRepository()
		workOrderRepo = pgStore.NewWorkOrderRepository()
		uploadRepo = pgStore.NewUploadRepository()
	} else {
		// A3 生产安全告警：脱离 compose 部署（ENV 默认 development）且未配 DATABASE_URL 时，
		// 系统静默退回内存存储，重启即丢数据。此处醒目告警，运维排障第一眼可见。
		slog.Warn("running with IN-MEMORY storage, data will be lost on restart (DATABASE_URL not set; NOT FOR PRODUCTION)")
		demandRepo = memory.NewDemandRepository(cipher)
		intentRepo = memory.NewIntentRepository()
		workOrderRepo = memory.NewWorkOrderRepository()
		enterpriseRepo = memory.NewEnterpriseRepository(cipher)
		employmentRepo = memory.NewEmploymentRepository()
		contractRepo = memory.NewContractRepository()
		contractTplRepo = memory.NewContractTemplateRepository()
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
		serviceListingRepo = memory.NewServiceListingRepository()
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
		caseRepo = memory.NewCaseRepository()
		svcAppRepo = memory.NewApplicationRepository()
		complianceRepo = memory.NewComplianceRepository()
		achieveRepo = memory.NewAchievementRepository()
		rdChallengeRepo = memory.NewRDChallengeRepository()
		researchRepo = memory.NewResearchProjectRepository()
		projAppRepo = memory.NewProjectAppRepository()
		competitionRepo = memory.NewCompetitionRepository(cipher)
		eventRepo = memory.NewEventRepository()
		portfolioRepo = memory.NewPortfolioRepository()
		industryRptRepo = memory.NewIndustryReportRepository()
		resourceRepo = memory.NewResourceRepository()
		emergencyRepo = memory.NewEmergencyRepository()
		uploadRepo = memory.NewUploadRepository()
	}

	// 托管金服务（独立变量：后台孤儿冻结补偿任务复用）
	escrowSvc := service.NewEscrowService(escrowRepo)

	app := httpapi.NewServer(
		service.NewDemandService(demandRepo),
		service.NewEnterpriseService(enterpriseRepo),
		service.NewEnterpriseSvc(enterpriseRepo, userRepo),
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
		service.NewHomeService(demandRepo, enterpriseRepo),
		service.NewFileService("uploads/", service.WithUploadQuota(uploadRepo, cfg.Server.UploadDailyQuotaBytes)),
		service.NewMessageService(msgRepo),
		service.NewEnrollmentService(enrollRepo, courseRepo),
		service.NewExpiryService(),
		service.NewTradeOrderService(tradeOrderRepo, productRepo),
		escrowSvc,
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
	app.SetIntentService(service.NewIntentService(intentRepo, demandRepo))
	app.SetWorkOrderService(service.NewWorkOrderService(workOrderRepo, demandRepo, intentRepo))
	app.SetServiceListingService(service.NewServiceListingService(serviceListingRepo))
	app.SetContractTemplateService(service.NewContractTemplateService(contractTplRepo))

	// Batch2/3 与扩展服务：PG 与内存双实现均已齐备，按 DATABASE_URL 分支注入。
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
	app.SetApplicationService(service.NewApplicationService(svcAppRepo))

	if pgStore != nil {
		app.SetAuditWriter(postgres.NewAuditAdapter(pgStore))
		app.SetDBPinger(pgStore.Pool())
		app.SetStorage("postgres")
	} else {
		app.SetStorage("memory")
	}

	// Seed super admin user from SUPER_ADMIN_PHONE env var.
	superPhone := os.Getenv("SUPER_ADMIN_PHONE")
	if superPhone != "" {
		seedCtx, seedCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer seedCancel()
		if _, err := userRepo.FindByID(seedCtx, superPhone); err != nil {
			now := time.Now()
			userRepo.Create(seedCtx, domain.User{
				ID: superPhone, WechatOpenID: superPhone, Role: domain.RolePlatformAdmin,
				Status: "active", Version: 1, CreatedAt: now, UpdatedAt: now,
			})
			slog.Info("seeded super admin", "phone", superPhone)
		}
	}

	server := &http.Server{Addr: addr, Handler: app.Router(),
		ReadHeaderTimeout: 10 * time.Second, // 慢速头攻击防护（批3 P1）
		ReadTimeout:       30 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second}

	// 孤儿冻结自动补偿：培训报名"先冻结后落库"的崩溃窗口可能导致资金滞留，
	// 每 10 分钟扫描一次 10 分钟前的冻结流水，业务记录不存在则自动退回余额。
	if escrowSvc != nil {
		go func() {
			defer func() { _ = recover() }() // 后台任务兜底，不拖垮主进程
			ticker := time.NewTicker(10 * time.Minute)
			defer ticker.Stop()
			for range ticker.C {
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				n, err := escrowSvc.RefundOrphanFreezes(ctx, "training_course", time.Now().Add(-10*time.Minute), 50)
				cancel()
				if err != nil {
					slog.Warn("orphan freeze scan failed", "error", err)
					continue
				}
				if n > 0 {
					slog.Info("orphan freezes refunded", "count", n)
				}
			}
		}()
	}

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
		// 批3 P2：Shutdown 超时后强制关闭，避免挂起连接拖住进程退出。
		_ = server.Close()
	}
	slog.Info("server exited")
}
