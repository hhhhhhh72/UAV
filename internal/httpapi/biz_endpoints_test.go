package httpapi_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/httpapi"
	"drone-platform/internal/repository"
	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

// httpIntentEntRepo 新建内存企业仓库并给 userID 预置一条 approved 企业认证：
// 登记对接的认证门槛（企业认证或飞手认证任一）——httpapi 工单/意向测试登记者白名单。
func httpIntentEntRepo(t *testing.T, userID string) repository.EnterpriseRepository {
	t.Helper()
	entRepo := memory.NewEnterpriseRepository(nil)
	if _, err := entRepo.Create(context.Background(), domain.Enterprise{
		ID: "ent-cert-" + userID, OwnerUserID: userID, Name: "认证企业-" + userID,
		Status: domain.EnterpriseApproved, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}); err != nil {
		t.Fatalf("seed enterprise cert: %v", err)
	}
	return entRepo
}

// productSellerEntRepo 商品测试卖家的企业认证仓库。
//
// 商品发布现在要求**企业认证（approved）**——商品详情页对买家承诺「平台认证商家」，
// 发布必须同等门槛（见 createProduct）。这批用例关注的是交易/售后/收藏，不关心认证
// 流程，所以统一在这里补齐前置条件，而不是让每个用例各自造企业。
func productSellerEntRepo(t *testing.T) repository.EnterpriseRepository {
	t.Helper()
	repo := memory.NewEnterpriseRepository(nil)
	// 只补 seller-1：它是 requestAs/authAs 直接指定的商品卖家。
	// **不要**给 enterprise-1 补——那是 request(..., RoleEnterprise) 的角色映射用户，
	// 企业入驻流程测试（biz_flow / p0_flow_regression / round3）要靠它从零建企业档案，
	// 预置一条 approved 会让它们撞上「每用户仅可维护一家企业档案」。
	for _, uid := range []string{"seller-1"} {
		if _, err := repo.Create(context.Background(), domain.Enterprise{
			ID: "ent-cert-" + uid, OwnerUserID: uid, Name: "认证商家-" + uid,
			Status: domain.EnterpriseApproved, CreatedAt: time.Now(), UpdatedAt: time.Now(),
		}); err != nil {
			t.Fatalf("seed enterprise cert for %s: %v", uid, err)
		}
	}
	return repo
}

// newBizServer 装配一个覆盖全部业务服务的内存后端，供 httpapi 层端到端用例使用。
func newBizServer(t *testing.T) http.Handler { return newBizServerWith(t, nil) }

// newBizServerWith 同上，并在返回前把 srv / 托管金服务 / 用户仓储交给 tune，
// 让用例能追加「main.go 里才有」的装配（如线上充值服务），而不必复制这 70 行装配代码。
func newBizServerWith(t *testing.T, tune func(srv *httpapi.Server, escrow *service.EscrowService, users repository.UserRepository)) http.Handler {
	t.Helper()
	tokens, err := httpapi.NewTokenManager(testSecret)
	if err != nil {
		t.Fatal(err)
	}
	demandRepo := memory.NewDemandRepository(nil)
	intentRepo := memory.NewIntentRepository(demandRepo)
	// 商品仓库必须与 TradeOrderService 共享同一实例（与 main.go 装配一致）：
	// 订单取消/删除要恢复商品（sold→listed），分实例则 Restore 找不到商品。
	productRepo := memory.NewProductRepository()
	// auth() issues tokens for user-1; pre-seed common test users so authenticate
	// (存在性/状态/令牌版本校验) resolves.
	userRepo := memory.NewUserRepository(nil)
	seedCommonUsers(userRepo)
	// 课程仓库共享：训练服务与报名服务必须读同一存储（与生产 PG 一致）
	courseRepo := memory.NewCourseRepository()
	// 托管金服务单独持有：线上充值用例要用同一个实例断言「钱真的进了托管账户」。
	escrowSvc := service.NewEscrowService(memory.NewEscrowRepository())
	srv := httpapi.NewServer(
		service.NewDemandService(demandRepo),
		service.NewEnterpriseService(memory.NewEnterpriseRepository(nil)),
		service.NewEnterpriseSvc(productSellerEntRepo(t), userRepo),
		service.NewEmploymentService(memory.NewEmploymentRepository()),
		service.NewContractService(memory.NewContractRepository()),
		service.NewJobService(memory.NewJobRepository(), memory.NewResumeRepository(), memory.NewJobApplicationRepository()),
		service.NewCommunityService(memory.NewPostRepository(), memory.NewCommentRepository(), memory.NewReportRepository()),
		service.NewListingService(memory.NewListingRepository()),
		service.NewLabourService(memory.NewLabourOrderRepository()),
		service.NewTrainingService(memory.NewCertificateRepository(), courseRepo, memory.NewInstructorRepository(), memory.NewPilotRepository(nil)),
		service.NewTradingService(productRepo, memory.NewRepairRepository(), nil, nil),
		service.NewInsuranceService(memory.NewPolicyRepository(), memory.NewInspectionRepository()),
		service.NewFinanceService(memory.NewLoanRepository()),
		service.NewHomeService(memory.NewDemandRepository(nil), memory.NewEnterpriseRepository(nil)),
		service.NewFileService("test_uploads/", service.WithUploadQuota(memory.NewUploadRepository(), 1<<40)),
		service.NewMessageService(memory.NewMessageRepository()),
		service.NewEnrollmentService(memory.NewEnrollmentRepository(), courseRepo),
		service.NewExpiryService(),
		service.NewTradeOrderService(memory.NewTradeOrderRepository(), productRepo),
		escrowSvc,
		service.NewNewsService(memory.NewArticleRepository()),
		service.NewReviewService(memory.NewReviewRepository(), memory.NewWorkOrderRepository()),
		service.NewVenueService(memory.NewVenueRepository()),
		userRepo, memory.NewRefreshTokenRepository(), tokens,
	)
	srv.SetExpertService(service.NewExpertService(memory.NewExpertRepository()))
	srv.SetCaseService(service.NewCaseService(memory.NewCaseRepository()))
	srv.SetComplianceService(service.NewComplianceService(memory.NewComplianceRepository()))
	srv.SetReportService(service.NewReportService(memory.NewIndustryReportRepository()))
	srv.SetPortfolioService(service.NewPortfolioService(memory.NewPortfolioRepository()))
	srv.SetAchievementService(service.NewAchievementService(memory.NewAchievementRepository()))
	rdRepo := memory.NewRDChallengeRepository()
	srv.SetRDChallengeService(service.NewRDChallengeService(rdRepo))
	srv.SetChallengeClaimService(service.NewChallengeClaimService(memory.NewChallengeClaimRepository(), rdRepo))
	srv.SetResearchProjectService(service.NewResearchProjectService(memory.NewResearchProjectRepository()))
	srv.SetProjectAppService(service.NewProjectAppService(memory.NewProjectAppRepository()))
	srv.SetCompetitionService(service.NewCompetitionService(memory.NewCompetitionRepository(nil)))
	srv.SetEventService(service.NewEventService(memory.NewEventRepository()))
	srv.SetResourceService(service.NewResourceService(memory.NewResourceRepository()))
	srv.SetEmergencyService(service.NewEmergencyService(memory.NewEmergencyRepository()))
	srv.SetRescueCaseService(service.NewRescueCaseService(memory.NewRescueCaseRepository()))
	srv.SetCooperationService(service.NewCooperationService(memory.NewCooperationRepository()))
	srv.SetEmergencyDeptService(service.NewEmergencyDeptService(memory.NewEmergencyDeptRepository()))
	srv.SetCollegeService(service.NewCollegeService(memory.NewCollegeRepository()))
	srv.SetExhibitionService(service.NewExhibitionService(memory.NewExhibitionRepository()))
	srv.SetTestSiteService(service.NewTestSiteService(memory.NewTestSiteRepository()))
	srv.SetPoolService(service.NewResourcePoolService(memory.NewResourcePoolRepository()))
	srv.SetTransformationService(service.NewTransformationService(memory.NewTransformationRepository()))
	srv.SetContractTemplateService(service.NewContractTemplateService(memory.NewContractTemplateRepository()))
	studyTourTestRepo := memory.NewStudyTourRepository()
	srv.SetStudyTourRepo(studyTourTestRepo)
	srv.SetStudyTourEnrollmentService(service.NewStudyTourEnrollmentService(memory.NewStudyTourEnrollmentRepository(), studyTourTestRepo))
	srv.SetEmergencyService(service.NewEmergencyService(memory.NewEmergencyRepository()))
	srv.SetMatchingService(service.NewMatchingService(demandRepo))
	srv.SetIntentService(service.NewIntentService(intentRepo, demandRepo, httpIntentEntRepo(t, "worker-1"), memory.NewPilotRepository(nil)))
	srv.SetWorkOrderService(service.NewWorkOrderService(memory.NewWorkOrderRepository(), demandRepo, intentRepo))
	srv.SetServiceListingService(service.NewServiceListingService(memory.NewProductRepository()))
	srv.SetStorage("memory")
	if tune != nil {
		tune(srv, escrowSvc, userRepo)
	}
	return srv.Router()
}

func TestBizExpertPublicList(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodGet, "/api/v1/experts", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK {
		t.Fatalf("GET experts: %d %s", w.Code, w.Body.String())
	}
	t.Log("expert public list OK")
}

func TestBizCasePublicList(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodGet, "/api/v1/cases", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK {
		t.Fatalf("GET cases: %d", w.Code)
	}
}

func TestBizCompliancePublicList(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodGet, "/api/v1/compliance-docs", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK {
		t.Fatalf("GET docs: %d", w.Code)
	}
	w = request(t, app, http.MethodGet, "/api/v1/compliance-standards", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK {
		t.Fatalf("GET standards: %d", w.Code)
	}
}

func TestBizAchievementPublicList(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodGet, "/api/v1/achievements", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK {
		t.Fatalf("GET achievements: %d", w.Code)
	}
}

func TestBizCompetitionPublicList(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodGet, "/api/v1/competitions", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK {
		t.Fatalf("GET competitions: %d", w.Code)
	}
}

func TestBizEventPublicList(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodGet, "/api/v1/events", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK {
		t.Fatalf("GET events: %d", w.Code)
	}
}

func TestBizReportPublicList(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodGet, "/api/v1/industry-reports", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK {
		t.Fatalf("GET reports: %d", w.Code)
	}
}

func TestBizEmergencyPublicList(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodGet, "/api/v1/emergency-resources", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK {
		t.Fatalf("GET emergency: %d", w.Code)
	}
}

func TestBizResourcePublicList(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodGet, "/api/v1/industry-resources", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK {
		t.Fatalf("GET resources: %d", w.Code)
	}
}

func TestBizRecommendEndpoint(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodGet, "/api/v1/recommendations?biz_type=cable_inspection&limit=5", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK {
		t.Fatalf("GET recommend: %d %s", w.Code, w.Body.String())
	}
}
