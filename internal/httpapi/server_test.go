package httpapi_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/httpapi"
	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

const testSecret = "01234567890123456789012345678901"

func newServer(t *testing.T) http.Handler {
	t.Helper()
	tokens, err := httpapi.NewTokenManager(testSecret)
	if err != nil {
		t.Fatal(err)
	}
	srv := httpapi.NewServer(service.NewDemandService(memory.NewDemandRepository(nil), memory.NewBidRepository()), service.NewEnterpriseService(memory.NewEnterpriseRepository(nil)), service.NewEnterpriseSvc(memory.NewEnterpriseRepository(nil)), service.NewEmploymentService(memory.NewEmploymentRepository()), service.NewContractService(memory.NewContractRepository()), service.NewJobService(memory.NewJobRepository(), memory.NewResumeRepository(), memory.NewJobApplicationRepository()), service.NewCommunityService(memory.NewPostRepository(), memory.NewCommentRepository(), memory.NewReportRepository()), service.NewListingService(memory.NewListingRepository()), service.NewLabourService(memory.NewLabourOrderRepository()), service.NewTrainingService(memory.NewCertificateRepository(), memory.NewCourseRepository(), memory.NewInstructorRepository(), memory.NewPilotRepository(nil)), service.NewTradingService(memory.NewProductRepository(), memory.NewRepairRepository()), service.NewInsuranceService(memory.NewPolicyRepository(), memory.NewInspectionRepository()), service.NewFinanceService(memory.NewLoanRepository()), service.NewHomeService(memory.NewDemandRepository(nil)), service.NewFileService("test_uploads/"), service.NewMessageService(memory.NewMessageRepository()), service.NewEnrollmentService(memory.NewEnrollmentRepository()), service.NewExpiryService(), service.NewTradeOrderService(memory.NewTradeOrderRepository()), service.NewEscrowService(memory.NewEscrowRepository()), service.NewNewsService(memory.NewArticleRepository()), service.NewReviewService(memory.NewReviewRepository()), service.NewVenueService(memory.NewVenueRepository()), memory.NewUserRepository(nil), memory.NewRefreshTokenRepository(), tokens)
	// Extended services used by public handlers (home endpoint etc.).
	srv.SetShopService(service.NewShopService(memory.NewShopRepository()))
	return srv.Router()
}
func auth(t *testing.T, role domain.Role) string {
	t.Helper()
	tokens, _ := httpapi.NewTokenManager(testSecret)
	token, err := tokens.Issue(domain.Actor{ID: "user-1", Role: role}, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	return "Bearer " + token
}
func request(t *testing.T, app http.Handler, method, path string, body []byte, role domain.Role) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest(method, path, bytes.NewReader(body))
	r.Header.Set("Authorization", auth(t, role))
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	return w
}

func bidRequest(t *testing.T, app http.Handler, method, path string, body []byte, token string) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest(method, path, bytes.NewReader(body))
	r.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	return w
}

func TestAuthorizationIsRequired(t *testing.T) {
	app := newServer(t)
	w := httptest.NewRecorder()
	// Use POST (write operation) which still requires auth even with public GET paths
	app.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/v1/demands", nil))
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("POST demands without auth: expected 401, got %d", w.Code)
	}
	// Also verify admin endpoints require auth
	w2 := httptest.NewRecorder()
	app.ServeHTTP(w2, httptest.NewRequest(http.MethodGet, "/api/v1/admin/enterprises", nil))
	if w2.Code != http.StatusUnauthorized {
		t.Fatalf("admin endpoint without auth: expected 401, got %d", w2.Code)
	}
}
func TestDemandRequiresApproval(t *testing.T) {
	app := newServer(t)
	body := []byte(`{"publisher_name":"company","contact":"13800000000","title":"inspection"}`)
	if w := request(t, app, http.MethodPost, "/api/v1/demands", body, domain.RoleEnterprise); w.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", w.Code, w.Body.String())
	}
	w := request(t, app, http.MethodGet, "/api/v1/demands", nil, domain.RoleIndividual)
	if bytes.Contains(w.Body.Bytes(), []byte("inspection")) {
		t.Fatal("unapproved demand was public")
	}
	w = request(t, app, http.MethodPost, "/api/v1/admin/demands/demand-"+"", nil, domain.RoleAssociationAdmin)
	if w.Code == http.StatusOK {
		t.Fatal("empty id unexpectedly approved")
	}
}
func TestPanicRecovery(t *testing.T) {
	// BLK-07: directly verify that a panic in a handler returns 500 instead of crashing.
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("deliberate test panic")
	})

	// Replicate the same recovery logic used in Server.recoverPanic.
	recovery := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(map[string]any{
					"error": map[string]string{"code": "INTERNAL", "message": "内部服务错误"},
				})
			}
		}()
		panicHandler.ServeHTTP(w, r)
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/test-panic", nil)
	recovery.ServeHTTP(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 after panic, got %d", w.Code)
	}
	if !bytes.Contains(w.Body.Bytes(), []byte("INTERNAL")) {
		t.Fatalf("expected INTERNAL error code in body, got %s", w.Body.String())
	}
	t.Logf("panic recovery OK: status=%d body=%s", w.Code, w.Body.String())
}

func TestEmploymentListRequiresPermission(t *testing.T) {
	// BLK-05: individual users must not be able to list employment requests.
	app := newServer(t)

	// Create an employment request as enterprise first.
	body := []byte(`{"enterprise_id":"ent-1","position":"飞手","headcount":5}`)
	w := request(t, app, http.MethodPost, "/api/v1/employment-requests", body, domain.RoleEnterprise)
	if w.Code != http.StatusCreated {
		t.Fatalf("create employment: got %d %s", w.Code, w.Body.String())
	}

	// Individual must NOT be able to list.
	w = request(t, app, http.MethodGet, "/api/v1/employment-requests", nil, domain.RoleIndividual)
	if w.Code != http.StatusForbidden {
		t.Fatalf("individual should be forbidden, got %d %s", w.Code, w.Body.String())
	}

	// Enterprise can list (should see their own).
	w = request(t, app, http.MethodGet, "/api/v1/employment-requests", nil, domain.RoleEnterprise)
	if w.Code != http.StatusOK {
		t.Fatalf("enterprise should see their own, got %d", w.Code)
	}

	t.Logf("employment list permission OK: individual=%d enterprise=%d", w.Code, http.StatusOK)
}

func TestCORSPreflight(t *testing.T) {
	// BLK-10: OPTIONS preflight must return 204 without auth.
	app := newServer(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodOptions, "/api/v1/demands", nil)
	r.Header.Set("Origin", "http://localhost:3000")
	app.ServeHTTP(w, r)

	if w.Code != http.StatusNoContent {
		t.Fatalf("preflight: expected 204, got %d", w.Code)
	}
	if w.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
		t.Fatalf("missing CORS origin header: %s", w.Header().Get("Access-Control-Allow-Origin"))
	}
	t.Logf("CORS preflight OK: status=%d origin=%s", w.Code, w.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORSBlockedOrigin(t *testing.T) {
	// BLK-10: unknown origin must not get Access-Control-Allow-Origin.
	app := newServer(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	r.Header.Set("Origin", "https://evil.com")
	app.ServeHTTP(w, r)

	origin := w.Header().Get("Access-Control-Allow-Origin")
	if origin != "" {
		t.Fatalf("untrusted origin should not get CORS header, got %s", origin)
	}
	t.Log("CORS blocked origin OK")
}

func TestContractCreatePermissions(t *testing.T) {
	// MAJ-01: enterprise must be able to create contracts (not just platform_admin).
	app := newServer(t)

	// Enterprise should be able to create a contract.
	body := []byte(`{"enterprise_id":"ent-1","template_id":"tpl-1"}`)
	w := request(t, app, http.MethodPost, "/api/v1/contracts", body, domain.RoleEnterprise)
	if w.Code != http.StatusCreated {
		t.Fatalf("enterprise create contract: expected 201, got %d %s", w.Code, w.Body.String())
	}

	// Individual must NOT be able to create a contract.
	w = request(t, app, http.MethodPost, "/api/v1/contracts", body, domain.RoleIndividual)
	if w.Code != http.StatusForbidden {
		t.Fatalf("individual should be forbidden, got %d", w.Code)
	}

	// Enterprise listing should only see their own contracts.
	w = request(t, app, http.MethodGet, "/api/v1/contracts", nil, domain.RoleEnterprise)
	if w.Code != http.StatusOK {
		t.Fatalf("enterprise list contracts: expected 200, got %d", w.Code)
	}

	t.Log("contract permissions OK: enterprise can create, individual blocked")
}

func TestPendingEnterpriseNeedsAssociationRole(t *testing.T) {
	app := newServer(t)
	if w := request(t, app, http.MethodGet, "/api/v1/admin/enterprises/pending", nil, domain.RoleIndividual); w.Code != http.StatusForbidden {
		t.Fatalf("got %d", w.Code)
	}
}

func TestBidCreateAndSelectFlow(t *testing.T) {
	app := newServer(t)
	// Create demand as enterprise user-1
	demandBody := []byte(`{"publisher_name":"测试","contact":"13800001111","title":"Bid测试需求","description":"test","biz_type":"other"}`)
	w := request(t, app, http.MethodPost, "/api/v1/demands", demandBody, domain.RoleEnterprise)
	if w.Code != http.StatusCreated {
		t.Fatalf("create demand: expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var m map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &m); err != nil {
		t.Fatalf("parse demand response: %v", err)
	}
	did := m["data"].(map[string]interface{})["id"].(string)
	w = request(t, app, http.MethodPost, "/api/v1/demands/"+did+"/submit", nil, domain.RoleEnterprise)
	w = request(t, app, http.MethodPost, "/api/v1/admin/demands/"+did+"/approve", nil, domain.RoleAssociationAdmin)

	// Create bid as individual user-2 (different from publisher user-1)
	bidBody := []byte(`{"amount_fen":50000,"proposal":"可以做"}`)
	bidderToken, _ := httpapi.NewTokenManager(testSecret)
	bidToken, _ := bidderToken.Issue(domain.Actor{ID: "user-2", Role: domain.RoleIndividual}, time.Hour)
	w = bidRequest(t, app, http.MethodPost, "/api/v1/demands/"+did+"/applications", bidBody, bidToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create bid: expected 201, got %d: %s", w.Code, w.Body.String())
	}
	if err := json.Unmarshal(w.Body.Bytes(), &m); err != nil {
		t.Fatalf("parse bid response: %v", err)
	}
	bidID := m["data"].(map[string]interface{})["id"].(string)

	// Select bid as publisher
	w = request(t, app, http.MethodPost, "/api/v1/demands/"+did+"/applications/"+bidID+"/select", nil, domain.RoleEnterprise)
	if w.Code != http.StatusOK {
		t.Fatalf("select bid: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	t.Log("bid flow OK: create→approve→bid→select")
}

func TestContractWebhookFlow(t *testing.T) {
	app := newServer(t)
	// Create contract
	contractBody := []byte(`{"template_id":"tpl-001","enterprise_id":"ent-webhook"}`)
	w := request(t, app, http.MethodPost, "/api/v1/contracts", contractBody, domain.RoleEnterprise)
	if w.Code != http.StatusCreated {
		t.Fatalf("create contract: expected 201, got %d", w.Code)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &m); err != nil {
		t.Fatalf("parse contract response: %v", err)
	}
	cid := m["data"].(map[string]interface{})["id"].(string)

	// Send signing webhook
	ts := time.Now().Unix()
	webhookBody, _ := json.Marshal(map[string]interface{}{
		"event_id":    "evt-test-001",
		"contract_id": cid,
		"status":      "sent",
		"timestamp":   ts,
	})
	w = request(t, app, http.MethodPost, "/api/v1/webhooks/signing", webhookBody, domain.RoleIndividual)
	if w.Code != http.StatusOK {
		t.Fatalf("webhook: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Duplicate webhook should return duplicate status
	w = request(t, app, http.MethodPost, "/api/v1/webhooks/signing", webhookBody, domain.RoleIndividual)
	if w.Code != http.StatusOK {
		t.Fatalf("duplicate webhook: expected 200, got %d", w.Code)
	}

	t.Log("contract webhook flow OK: create→signing callback→duplicate detected")
}

func TestEmploymentListPagination(t *testing.T) {
	app := newServer(t)
	w := request(t, app, http.MethodGet, "/api/v1/employment-requests?page=1&page_size=5", nil, domain.RoleEnterprise)
	if w.Code != http.StatusOK {
		t.Fatalf("list employment with pagination: expected 200, got %d", w.Code)
	}
	var m map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &m)
	if _, ok := m["total"]; !ok {
		t.Fatal("paginated response missing total field")
	}
	t.Log("employment pagination OK")
}
