package httpapi_test

import (
	"net/http"
	"testing"

	"drone-platform/internal/domain"
	"drone-platform/internal/httpapi"
	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

func newBizServer(t *testing.T) http.Handler {
	t.Helper()
	tokens, err := httpapi.NewTokenManager(testSecret)
	if err != nil { t.Fatal(err) }
	srv := httpapi.NewServer(
		service.NewDemandService(memory.NewDemandRepository(nil), memory.NewBidRepository()),
		service.NewEnterpriseService(memory.NewEnterpriseRepository(nil)),
		service.NewEnterpriseSvc(memory.NewEnterpriseRepository(nil)),
		service.NewEmploymentService(memory.NewEmploymentRepository()),
		service.NewContractService(memory.NewContractRepository()),
		service.NewJobService(memory.NewJobRepository(), memory.NewResumeRepository(), memory.NewJobApplicationRepository()),
		service.NewCommunityService(memory.NewPostRepository(), memory.NewCommentRepository(), memory.NewReportRepository()),
		service.NewListingService(memory.NewListingRepository()),
		service.NewLabourService(memory.NewLabourOrderRepository()),
		service.NewTrainingService(memory.NewCertificateRepository(), memory.NewCourseRepository(), memory.NewInstructorRepository(), memory.NewPilotRepository(nil)),
		service.NewTradingService(memory.NewProductRepository(), memory.NewRepairRepository()),
		service.NewInsuranceService(memory.NewPolicyRepository(), memory.NewInspectionRepository()),
		service.NewFinanceService(memory.NewLoanRepository()),
		service.NewHomeService(memory.NewDemandRepository(nil)),
		service.NewFileService("test_uploads/"),
		service.NewMessageService(memory.NewMessageRepository()),
		service.NewEnrollmentService(memory.NewEnrollmentRepository()),
		service.NewExpiryService(),
		service.NewTradeOrderService(memory.NewTradeOrderRepository()),
		service.NewEscrowService(memory.NewEscrowRepository()),
		service.NewNewsService(memory.NewArticleRepository()),
		service.NewReviewService(memory.NewReviewRepository()),
		service.NewVenueService(memory.NewVenueRepository()),
		memory.NewUserRepository(nil), memory.NewRefreshTokenRepository(), tokens,
	)
	srv.SetExpertService(service.NewExpertService(memory.NewExpertRepository()))
	srv.SetCaseService(service.NewCaseService(memory.NewCaseRepository()))
	srv.SetComplianceService(service.NewComplianceService(memory.NewComplianceRepository()))
	srv.SetReportService(service.NewReportService(memory.NewIndustryReportRepository()))
	srv.SetPortfolioService(service.NewPortfolioService(memory.NewPortfolioRepository()))
	srv.SetAchievementService(service.NewAchievementService(memory.NewAchievementRepository()))
	srv.SetRDChallengeService(service.NewRDChallengeService(memory.NewRDChallengeRepository()))
	srv.SetResearchProjectService(service.NewResearchProjectService(memory.NewResearchProjectRepository()))
	srv.SetProjectAppService(service.NewProjectAppService(memory.NewProjectAppRepository()))
	srv.SetCompetitionService(service.NewCompetitionService(memory.NewCompetitionRepository()))
	srv.SetEventService(service.NewEventService(memory.NewEventRepository()))
	srv.SetResourceService(service.NewResourceService(memory.NewResourceRepository()))
	srv.SetEmergencyService(service.NewEmergencyService(memory.NewEmergencyRepository()))
	srv.SetCollegeService(service.NewCollegeService(memory.NewCollegeRepository()))
	srv.SetExhibitionService(service.NewExhibitionService(memory.NewExhibitionRepository()))
	srv.SetTestSiteService(service.NewTestSiteService(memory.NewTestSiteRepository()))
	srv.SetTransformationService(service.NewTransformationService(memory.NewTransformationRepository()))
	srv.SetStudyTourRepo(memory.NewStudyTourRepository())
	srv.SetEmergencyService(service.NewEmergencyService(memory.NewEmergencyRepository()))
	srv.SetMatchingService(service.NewMatchingService(memory.NewDemandRepository(nil)))
	srv.SetStorage("memory")
	return srv.Router()
}

func TestBizExpertPublicList(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodGet, "/api/v1/experts", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK { t.Fatalf("GET experts: %d %s", w.Code, w.Body.String()) }
	t.Log("expert public list OK")
}

func TestBizCasePublicList(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodGet, "/api/v1/cases", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK { t.Fatalf("GET cases: %d", w.Code) }
}

func TestBizCompliancePublicList(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodGet, "/api/v1/compliance-docs", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK { t.Fatalf("GET docs: %d", w.Code) }
	w = request(t, app, http.MethodGet, "/api/v1/compliance-standards", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK { t.Fatalf("GET standards: %d", w.Code) }
}

func TestBizAchievementPublicList(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodGet, "/api/v1/achievements", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK { t.Fatalf("GET achievements: %d", w.Code) }
}

func TestBizCompetitionPublicList(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodGet, "/api/v1/competitions", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK { t.Fatalf("GET competitions: %d", w.Code) }
}

func TestBizEventPublicList(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodGet, "/api/v1/events", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK { t.Fatalf("GET events: %d", w.Code) }
}

func TestBizReportPublicList(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodGet, "/api/v1/industry-reports", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK { t.Fatalf("GET reports: %d", w.Code) }
}

func TestBizEmergencyPublicList(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodGet, "/api/v1/emergency-resources", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK { t.Fatalf("GET emergency: %d", w.Code) }
}

func TestBizResourcePublicList(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodGet, "/api/v1/industry-resources", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK { t.Fatalf("GET resources: %d", w.Code) }
}

func TestBizRecommendEndpoint(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodGet, "/api/v1/recommendations?biz_type=cable_inspection&limit=5", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK { t.Fatalf("GET recommend: %d %s", w.Code, w.Body.String()) }
}
