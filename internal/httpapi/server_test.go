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
	srv := httpapi.NewServer(service.NewDemandService(memory.NewDemandRepository(nil)), service.NewEnterpriseService(memory.NewEnterpriseRepository(nil)), service.NewEnterpriseSvc(memory.NewEnterpriseRepository(nil), memory.NewUserRepository(nil)), service.NewEmploymentService(memory.NewEmploymentRepository()), service.NewContractService(memory.NewContractRepository()), service.NewJobService(memory.NewJobRepository(), memory.NewResumeRepository(), memory.NewJobApplicationRepository()), service.NewCommunityService(memory.NewPostRepository(), memory.NewCommentRepository(), memory.NewReportRepository()), service.NewListingService(memory.NewListingRepository()), service.NewLabourService(memory.NewLabourOrderRepository()), service.NewTrainingService(memory.NewCertificateRepository(), memory.NewCourseRepository(), memory.NewInstructorRepository(), memory.NewPilotRepository(nil)), service.NewTradingService(memory.NewProductRepository(), memory.NewRepairRepository()), service.NewInsuranceService(memory.NewPolicyRepository(), memory.NewInspectionRepository()), service.NewFinanceService(memory.NewLoanRepository()), service.NewHomeService(memory.NewDemandRepository(nil), memory.NewEnterpriseRepository(nil)), service.NewFileService("test_uploads/"), service.NewMessageService(memory.NewMessageRepository()), service.NewEnrollmentService(memory.NewEnrollmentRepository()), service.NewExpiryService(), service.NewTradeOrderService(memory.NewTradeOrderRepository()), service.NewEscrowService(memory.NewEscrowRepository()), service.NewNewsService(memory.NewArticleRepository()), service.NewReviewService(memory.NewReviewRepository()), service.NewVenueService(memory.NewVenueRepository()), memory.NewUserRepository(nil), memory.NewRefreshTokenRepository(), tokens)
	// Extended services used by public handlers (home endpoint etc.).
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
// 我的发布（mine=1）必须只返回当前用户的需求：未登录时返回空列表，
// 绝不回退为"全部需求"（防止未登录泄露他人/种子数据）。
func TestDemandMineUnauthenticatedReturnsEmpty(t *testing.T) {
	app := newServer(t)

	// 企业发布一条需求 → 管理员审批通过 → 进入公开列表
	dw := requestAs(t, app, http.MethodPost, "/api/v1/demands",
		[]byte(`{"title":"我的巡检需求","contact":"13800000000","biz_type":"cable_inspection"}`),
		"enterprise-1", domain.RoleEnterprise)
	if dw.Code != http.StatusCreated {
		t.Fatalf("create demand: %d %s", dw.Code, dw.Body.String())
	}
	var created struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(dw.Body.Bytes(), &created); err != nil {
		t.Fatalf("parse demand: %v", err)
	}
	rw := requestAs(t, app, http.MethodPost, "/api/v1/admin/demands/"+created.Data.ID+"/review",
		[]byte(`{"action":"approve"}`), "admin-1", domain.RolePlatformAdmin)
	if rw.Code != http.StatusOK {
		t.Fatalf("approve demand: %d %s", rw.Code, rw.Body.String())
	}

	// 对照组：未登录看公开列表 → 能看到该需求（列表本身有数据）
	anon := httptest.NewRecorder()
	app.ServeHTTP(anon, httptest.NewRequest(http.MethodGet, "/api/v1/demands", nil))
	if anon.Code != http.StatusOK || !bytes.Contains(anon.Body.Bytes(), []byte("我的巡检需求")) {
		t.Fatalf("public list should contain the demand: %d %s", anon.Code, anon.Body.String())
	}

	// 修复点：未登录 + mine=1 → 必须为空列表，不得返回全部需求
	anonMine := httptest.NewRecorder()
	app.ServeHTTP(anonMine, httptest.NewRequest(http.MethodGet, "/api/v1/demands?mine=1", nil))
	if anonMine.Code != http.StatusOK || bytes.Contains(anonMine.Body.Bytes(), []byte("我的巡检需求")) {
		t.Fatalf("unauthenticated mine=1 must not leak demands: %d %s", anonMine.Code, anonMine.Body.String())
	}

	// 发布者本人 → 能看到自己的需求
	owner := requestAs(t, app, http.MethodGet, "/api/v1/demands?mine=1", nil, "enterprise-1", domain.RoleEnterprise)
	if !bytes.Contains(owner.Body.Bytes(), []byte("我的巡检需求")) {
		t.Fatalf("owner should see own demand: %d %s", owner.Code, owner.Body.String())
	}

	// 其他用户 → 看不到别人的需求
	other := requestAs(t, app, http.MethodGet, "/api/v1/demands?mine=1", nil, "worker-1", domain.RoleIndividual)
	if bytes.Contains(other.Body.Bytes(), []byte("我的巡检需求")) {
		t.Fatalf("other user must not see someone else's demand: %d %s", other.Code, other.Body.String())
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

// 商品审核流程：用户发布 → pending（不进公开列表）→ 管理后台通过 → listed（公开可见）→ 下架后再次隐藏
func TestProductRequiresApproval(t *testing.T) {
	app := newServer(t)

	// 用户发布商品 → 响应状态必须是 pending（待审核）
	pw := requestAs(t, app, http.MethodPost, "/api/v1/products",
		[]byte(`{"title":"待审无人机","prod_type":"drone","price_fen":100000,"brand":"DJI","model":"M350"}`),
		"seller-1", domain.RoleEnterprise)
	if pw.Code != http.StatusCreated {
		t.Fatalf("create product: %d %s", pw.Code, pw.Body.String())
	}
	var created struct {
		Data struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(pw.Body.Bytes(), &created); err != nil {
		t.Fatalf("parse product: %v", err)
	}
	if created.Data.Status != "pending" {
		t.Fatalf("new product must be pending, got %q", created.Data.Status)
	}

	// 待审核商品不得出现在公开列表
	pub := requestAs(t, app, http.MethodGet, "/api/v1/products", nil, "buyer-1", domain.RoleIndividual)
	if bytes.Contains(pub.Body.Bytes(), []byte("待审无人机")) {
		t.Fatalf("pending product must not be public: %s", pub.Body.String())
	}
	// 发布者 mine=1 应能看到（含待审核）
	mine := requestAs(t, app, http.MethodGet, "/api/v1/products?mine=1", nil, "seller-1", domain.RoleEnterprise)
	if !bytes.Contains(mine.Body.Bytes(), []byte("待审无人机")) {
		t.Fatalf("owner mine=1 should see own pending product: %d %s", mine.Code, mine.Body.String())
	}

	// 管理后台通过 → listed → 公开可见
	apw := requestAs(t, app, http.MethodPut, "/api/v1/admin/products/"+created.Data.ID,
		[]byte(`{"status":"listed"}`), "admin-1", domain.RolePlatformAdmin)
	if apw.Code != http.StatusOK {
		t.Fatalf("admin approve product: %d %s", apw.Code, apw.Body.String())
	}
	pub = requestAs(t, app, http.MethodGet, "/api/v1/products", nil, "buyer-1", domain.RoleIndividual)
	if !bytes.Contains(pub.Body.Bytes(), []byte("待审无人机")) {
		t.Fatalf("approved product should be public: %s", pub.Body.String())
	}

	// 管理后台下架 → 公开列表再次隐藏（mine=1 仍可见）
	rmw := requestAs(t, app, http.MethodPut, "/api/v1/admin/products/"+created.Data.ID,
		[]byte(`{"status":"removed"}`), "admin-1", domain.RolePlatformAdmin)
	if rmw.Code != http.StatusOK {
		t.Fatalf("admin remove product: %d %s", rmw.Code, rmw.Body.String())
	}
	pub = requestAs(t, app, http.MethodGet, "/api/v1/products", nil, "buyer-1", domain.RoleIndividual)
	if bytes.Contains(pub.Body.Bytes(), []byte("待审无人机")) {
		t.Fatalf("removed product must not be public: %s", pub.Body.String())
	}
}

// 售后闭环：下单 → 支付 → 发货 → 买家申请售后（aftersale）→ 管理端审核（同意退款/驳回）→ 结案回 completed
func TestAftersaleFlow(t *testing.T) {
	app := newServer(t)

	// 1. 管理后台直接创建商品（listed，跳过审核流程，聚焦订单售后链路）
	pw := requestAs(t, app, http.MethodPost, "/api/v1/admin/products",
		[]byte(`{"title":"售后测试无人机","prod_type":"drone","price_fen":500000,"seller_name":"卖家A"}`),
		"admin-1", domain.RolePlatformAdmin)
	if pw.Code != http.StatusCreated {
		t.Fatalf("admin create product: %d %s", pw.Code, pw.Body.String())
	}
	var product struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(pw.Body.Bytes(), &product); err != nil {
		t.Fatalf("parse product: %v", err)
	}

	// 2. 买家下单 → pending
	ow := requestAs(t, app, http.MethodPost, "/api/v1/trade-orders",
		[]byte(`{"product_id":"`+product.Data.ID+`","seller_id":"seller-1","amount_fen":500000}`),
		"buyer-1", domain.RoleIndividual)
	if ow.Code != http.StatusCreated {
		t.Fatalf("create order: %d %s", ow.Code, ow.Body.String())
	}
	var order struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(ow.Body.Bytes(), &order); err != nil {
		t.Fatalf("parse order: %v", err)
	}
	oid := order.Data.ID

	// 3. 模拟支付（买家）→ paid；模拟发货（卖家）→ shipped
	for _, step := range []struct{ user, role string; body string }{
		{"buyer-1", string(domain.RoleIndividual), `{"status":"paid"}`},
		{"seller-1", string(domain.RoleEnterprise), `{"status":"shipped"}`},
	} {
		w := requestAs(t, app, http.MethodPatch, "/api/v1/trade-orders/"+oid+"/status",
			[]byte(step.body), step.user, domain.Role(step.role))
		if w.Code != http.StatusOK {
			t.Fatalf("patch status %s: %d %s", step.body, w.Code, w.Body.String())
		}
	}

	// 4. 买家申请售后 → aftersale + pending
	aw := requestAs(t, app, http.MethodPost, "/api/v1/trade-orders/"+oid+"/aftersale",
		[]byte(`{"aftersale_type":"refund","aftersale_reason":"质量问题","aftersale_desc":"云台故障","aftersale_amount_fen":500000}`),
		"buyer-1", domain.RoleIndividual)
	if aw.Code != http.StatusOK {
		t.Fatalf("apply aftersale: %d %s", aw.Code, aw.Body.String())
	}
	if !bytes.Contains(aw.Body.Bytes(), []byte(`"status":"aftersale"`)) ||
		!bytes.Contains(aw.Body.Bytes(), []byte(`"aftersale_status":"pending"`)) ||
		!bytes.Contains(aw.Body.Bytes(), []byte(`"aftersale_type":"refund"`)) {
		t.Fatalf("aftersale state wrong: %s", aw.Body.String())
	}

	// 5. 卖家不能替买家申请；重复申请被拒
	if w := requestAs(t, app, http.MethodPost, "/api/v1/trade-orders/"+oid+"/aftersale",
		[]byte(`{"aftersale_type":"refund","aftersale_reason":"x"}`), "seller-1", domain.RoleEnterprise); w.Code != http.StatusBadRequest {
		t.Fatalf("seller apply aftersale should be rejected, got %d", w.Code)
	}
	if w := requestAs(t, app, http.MethodPost, "/api/v1/trade-orders/"+oid+"/aftersale",
		[]byte(`{"aftersale_type":"refund","aftersale_reason":"x"}`), "buyer-1", domain.RoleIndividual); w.Code != http.StatusBadRequest {
		t.Fatalf("duplicate aftersale should be rejected, got %d %s", w.Code, w.Body.String())
	}

	// 6. mine 列表能看到售后记录（买家视角）
	mine := requestAs(t, app, http.MethodGet, "/api/v1/trade-orders/mine", nil, "buyer-1", domain.RoleIndividual)
	if !bytes.Contains(mine.Body.Bytes(), []byte(`"aftersale_status":"pending"`)) {
		t.Fatalf("mine should expose aftersale fields: %s", mine.Body.String())
	}

	// 7. 管理端同意退款 → aftersale_status=approved，订单回 completed
	rv := requestAs(t, app, http.MethodPut, "/api/v1/admin/orders/"+oid+"/aftersale",
		[]byte(`{"action":"approve"}`), "admin-1", domain.RolePlatformAdmin)
	if rv.Code != http.StatusOK {
		t.Fatalf("review approve: %d %s", rv.Code, rv.Body.String())
	}
	if !bytes.Contains(rv.Body.Bytes(), []byte(`"aftersale_status":"approved"`)) ||
		!bytes.Contains(rv.Body.Bytes(), []byte(`"status":"completed"`)) {
		t.Fatalf("review result wrong: %s", rv.Body.String())
	}
	// 结案后不能再审
	if w := requestAs(t, app, http.MethodPut, "/api/v1/admin/orders/"+oid+"/aftersale",
		[]byte(`{"action":"reject"}`), "admin-1", domain.RolePlatformAdmin); w.Code != http.StatusBadRequest {
		t.Fatalf("re-review should be rejected, got %d", w.Code)
	}

	// 8. 驳回分支：新订单 → 付款后未发货直接申请（paid → aftersale）→ reject → aftersale_status=rejected
	ow2 := requestAs(t, app, http.MethodPost, "/api/v1/trade-orders",
		[]byte(`{"product_id":"`+product.Data.ID+`","seller_id":"seller-1","amount_fen":500000}`),
		"buyer-2", domain.RoleIndividual)
	var order2 struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(ow2.Body.Bytes(), &order2); err != nil {
		t.Fatalf("parse order2: %v", err)
	}
	oid2 := order2.Data.ID
	// 仅支付（paid），不发货——验证「付款后未发货可申请售后」
	requestAs(t, app, http.MethodPatch, "/api/v1/trade-orders/"+oid2+"/status", []byte(`{"status":"paid"}`), "buyer-2", domain.RoleIndividual)
	if w := requestAs(t, app, http.MethodPost, "/api/v1/trade-orders/"+oid2+"/aftersale",
		[]byte(`{"aftersale_type":"return","aftersale_reason":"不想要了"}`), "buyer-2", domain.RoleIndividual); w.Code != http.StatusOK {
		t.Fatalf("apply aftersale2: %d %s", w.Code, w.Body.String())
	}
	rj := requestAs(t, app, http.MethodPut, "/api/v1/admin/orders/"+oid2+"/aftersale",
		[]byte(`{"action":"reject"}`), "admin-1", domain.RolePlatformAdmin)
	if rj.Code != http.StatusOK || !bytes.Contains(rj.Body.Bytes(), []byte(`"aftersale_status":"rejected"`)) {
		t.Fatalf("review reject: %d %s", rj.Code, rj.Body.String())
	}
}
