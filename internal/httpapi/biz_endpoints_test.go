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

func newBizServer(t *testing.T) http.Handler {
	t.Helper()
	tokens, err := httpapi.NewTokenManager(testSecret)
	if err != nil {
		t.Fatal(err)
	}
	demandRepo := memory.NewDemandRepository(nil)
	intentRepo := memory.NewIntentRepository()
	// 商品仓库必须与 TradeOrderService 共享同一实例（与 main.go 装配一致）：
	// 订单取消/删除要恢复商品（sold→listed），分实例则 Restore 找不到商品。
	productRepo := memory.NewProductRepository()
	// auth() issues tokens for user-1; pre-seed common test users so authenticate
	// (存在性/状态/令牌版本校验) resolves.
	userRepo := memory.NewUserRepository(nil)
	seedCommonUsers(userRepo)
	srv := httpapi.NewServer(
		service.NewDemandService(demandRepo),
		service.NewEnterpriseService(memory.NewEnterpriseRepository(nil)),
		service.NewEnterpriseSvc(memory.NewEnterpriseRepository(nil), userRepo),
		service.NewEmploymentService(memory.NewEmploymentRepository()),
		service.NewContractService(memory.NewContractRepository()),
		service.NewJobService(memory.NewJobRepository(), memory.NewResumeRepository(), memory.NewJobApplicationRepository()),
		service.NewCommunityService(memory.NewPostRepository(), memory.NewCommentRepository(), memory.NewReportRepository()),
		service.NewListingService(memory.NewListingRepository()),
		service.NewLabourService(memory.NewLabourOrderRepository()),
		service.NewTrainingService(memory.NewCertificateRepository(), memory.NewCourseRepository(), memory.NewInstructorRepository(), memory.NewPilotRepository(nil)),
		service.NewTradingService(productRepo, memory.NewRepairRepository()),
		service.NewInsuranceService(memory.NewPolicyRepository(), memory.NewInspectionRepository()),
		service.NewFinanceService(memory.NewLoanRepository()),
		service.NewHomeService(memory.NewDemandRepository(nil), memory.NewEnterpriseRepository(nil)),
		service.NewFileService("test_uploads/", service.WithUploadQuota(memory.NewUploadRepository(), 1<<40)),
		service.NewMessageService(memory.NewMessageRepository()),
		service.NewEnrollmentService(memory.NewEnrollmentRepository(), memory.NewCourseRepository()),
		service.NewExpiryService(),
		service.NewTradeOrderService(memory.NewTradeOrderRepository(), productRepo),
		service.NewEscrowService(memory.NewEscrowRepository()),
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
	srv.SetAssociationMemberService(service.NewAssociationMemberService(memory.NewAssociationMemberRepository()))
	srv.SetContractTemplateService(service.NewContractTemplateService(memory.NewContractTemplateRepository()))
	srv.SetStudyTourRepo(memory.NewStudyTourRepository())
	srv.SetEmergencyService(service.NewEmergencyService(memory.NewEmergencyRepository()))
	srv.SetMatchingService(service.NewMatchingService(demandRepo))
	srv.SetIntentService(service.NewIntentService(intentRepo, demandRepo, httpIntentEntRepo(t, "worker-1"), memory.NewPilotRepository(nil)))
	srv.SetWorkOrderService(service.NewWorkOrderService(memory.NewWorkOrderRepository(), demandRepo, intentRepo))
	srv.SetServiceListingService(service.NewServiceListingService(memory.NewServiceListingRepository()))
	srv.SetStorage("memory")
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
