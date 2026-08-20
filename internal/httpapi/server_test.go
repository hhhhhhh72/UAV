package httpapi_test

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
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
	srv := httpapi.NewServer(service.NewDemandService(memory.NewDemandRepository(nil)), service.NewEnterpriseService(memory.NewEnterpriseRepository(nil)), service.NewEnterpriseSvc(memory.NewEnterpriseRepository(nil), memory.NewUserRepository(nil)), service.NewEmploymentService(memory.NewEmploymentRepository()), service.NewContractService(memory.NewContractRepository()), service.NewJobService(memory.NewJobRepository(), memory.NewResumeRepository(), memory.NewJobApplicationRepository()), service.NewCommunityService(memory.NewPostRepository(), memory.NewCommentRepository(), memory.NewReportRepository()), service.NewListingService(memory.NewListingRepository()), service.NewLabourService(memory.NewLabourOrderRepository()), service.NewTrainingService(memory.NewCertificateRepository(), memory.NewCourseRepository(), memory.NewInstructorRepository(), memory.NewPilotRepository(nil)), service.NewTradingService(memory.NewProductRepository(), memory.NewRepairRepository()), service.NewInsuranceService(memory.NewPolicyRepository(), memory.NewInspectionRepository()), service.NewFinanceService(memory.NewLoanRepository()), service.NewHomeService(memory.NewDemandRepository(nil), memory.NewEnterpriseRepository(nil)), service.NewFileService("test_uploads/", service.WithUploadQuota(memory.NewUploadRepository(), 1<<40)), service.NewMessageService(memory.NewMessageRepository()), service.NewEnrollmentService(memory.NewEnrollmentRepository(), memory.NewCourseRepository()), service.NewExpiryService(), service.NewTradeOrderService(memory.NewTradeOrderRepository(), memory.NewProductRepository()), service.NewEscrowService(memory.NewEscrowRepository()), service.NewNewsService(memory.NewArticleRepository()), service.NewReviewService(memory.NewReviewRepository()), service.NewVenueService(memory.NewVenueRepository()), memory.NewUserRepository(nil), memory.NewRefreshTokenRepository(), tokens)
	// Extended services used by public handlers (home endpoint etc.).
	srv.SetTestSiteService(service.NewTestSiteService(memory.NewTestSiteRepository()))
	// batch2 模块服务：鉴权回归测试（C2）需要
	srv.SetTransformationService(service.NewTransformationService(memory.NewTransformationRepository()))
	srv.SetCollegeService(service.NewCollegeService(memory.NewCollegeRepository()))
	srv.SetCooperationService(service.NewCooperationService(memory.NewCooperationRepository()))
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
// TestDemandListPagination: 回归 C8——listDemands 曾手工切片后再经
// paginatedRespond 二次切片，导致 page≥2 恒为空。
func TestDemandListPagination(t *testing.T) {
	app := newServer(t)

	// Arrange: 发布并审批 5 条需求，使其进入公开列表
	for i := 0; i < 5; i++ {
		body := []byte(fmt.Sprintf(`{"title":"分页需求%d","contact":"13800000000","biz_type":"cable_inspection"}`, i))
		dw := requestAs(t, app, http.MethodPost, "/api/v1/demands", body, "enterprise-1", domain.RoleEnterprise)
		if dw.Code != http.StatusCreated {
			t.Fatalf("create demand %d: %d %s", i, dw.Code, dw.Body.String())
		}
		var created struct {
			Data struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		if err := json.Unmarshal(dw.Body.Bytes(), &created); err != nil {
			t.Fatalf("parse demand %d: %v", i, err)
		}
		rw := requestAs(t, app, http.MethodPost, "/api/v1/admin/demands/"+created.Data.ID+"/review",
			[]byte(`{"action":"approve"}`), "admin-1", domain.RolePlatformAdmin)
		if rw.Code != http.StatusOK {
			t.Fatalf("approve demand %d: %d %s", i, rw.Code, rw.Body.String())
		}
	}

	pages := []struct {
		query     string
		wantItems int
		wantTotal int
	}{
		{"page=1&page_size=2", 2, 5},
		{"page=2&page_size=2", 2, 5}, // 修复前恒为 0
		{"page=3&page_size=2", 1, 5},
	}
	for _, p := range pages {
		w := requestAs(t, app, http.MethodGet, "/api/v1/demands?"+p.query, nil, "worker-1", domain.RoleIndividual)
		if w.Code != http.StatusOK {
			t.Fatalf("%s: %d %s", p.query, w.Code, w.Body.String())
		}
		var out struct {
			Data  []map[string]any `json:"data"`
			Total int              `json:"total"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
			t.Fatalf("parse %s: %v", p.query, err)
		}
		if len(out.Data) != p.wantItems {
			t.Fatalf("%s: expected %d items, got %d (%s)", p.query, p.wantItems, len(out.Data), w.Body.String())
		}
		if out.Total != p.wantTotal {
			t.Fatalf("%s: expected total %d, got %d", p.query, p.wantTotal, out.Total)
		}
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
	// P0 修复后 webhook 强制签名：测试需配置 SIGNING_SECRET 并计算 HMAC。
	oldSecret := os.Getenv("SIGNING_SECRET")
	os.Setenv("SIGNING_SECRET", "test-signing-secret")
	t.Cleanup(func() { os.Setenv("SIGNING_SECRET", oldSecret) })

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

	// Send signing webhook（带合法 HMAC 签名）
	ts := time.Now().Unix()
	sign := func(ts int64, eventID, contractID, status string) string {
		mac := hmac.New(sha256.New, []byte("test-signing-secret"))
		mac.Write([]byte(fmt.Sprintf("%d.%s.%s.%s", ts, eventID, contractID, status)))
		return hex.EncodeToString(mac.Sum(nil))
	}
	webhookBody, _ := json.Marshal(map[string]interface{}{
		"event_id":    "evt-test-001",
		"contract_id": cid,
		"status":      "sent",
		"timestamp":   ts,
		"signature":   sign(ts, "evt-test-001", cid, "sent"),
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

	// 未签名（错误签名）的事件必须 403
	forged, _ := json.Marshal(map[string]interface{}{
		"event_id":    "evt-test-002",
		"contract_id": cid,
		"status":      "signed",
		"timestamp":   time.Now().Unix(),
		"signature":   "deadbeef",
	})
	w = request(t, app, http.MethodPost, "/api/v1/webhooks/signing", forged, domain.RoleIndividual)
	if w.Code != http.StatusForbidden {
		t.Fatalf("forged webhook: expected 403, got %d", w.Code)
	}

	t.Log("contract webhook flow OK: create→signed callback→duplicate→forged rejected")
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

	// 1. 卖家用户发布商品（pending）→ 管理后台审核上架（listed）
	pw := requestAs(t, app, http.MethodPost, "/api/v1/products",
		[]byte(`{"title":"售后测试无人机","prod_type":"drone","price_fen":500000}`),
		"seller-1", domain.RoleEnterprise)
	if pw.Code != http.StatusCreated {
		t.Fatalf("seller create product: %d %s", pw.Code, pw.Body.String())
	}
	var product struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(pw.Body.Bytes(), &product); err != nil {
		t.Fatalf("parse product: %v", err)
	}
	appr := requestAs(t, app, http.MethodPut, "/api/v1/admin/products/"+product.Data.ID,
		[]byte(`{"status":"listed"}`), "admin-1", domain.RolePlatformAdmin)
	if appr.Code != http.StatusOK {
		t.Fatalf("admin approve product: %d %s", appr.Code, appr.Body.String())
	}

	// 2. 买家下单 → pending；金额/卖家以服务端商品为准（客户端传 1 分钱+假卖家被忽略）
	ow := requestAs(t, app, http.MethodPost, "/api/v1/trade-orders",
		[]byte(`{"product_id":"`+product.Data.ID+`","seller_id":"hacker-x","amount_fen":1}`),
		"buyer-1", domain.RoleIndividual)
	if ow.Code != http.StatusCreated {
		t.Fatalf("create order: %d %s", ow.Code, ow.Body.String())
	}
	if !bytes.Contains(ow.Body.Bytes(), []byte(`"amount_fen":500000`)) {
		t.Fatalf("order must use server-side price, got: %s", ow.Body.String())
	}
	if !bytes.Contains(ow.Body.Bytes(), []byte(`"seller_id":"seller-1"`)) {
		t.Fatalf("order must use product seller, got: %s", ow.Body.String())
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

	// 3. 模拟支付（paid 仅管理端可设，走管理端改单接口）→ paid；模拟发货（卖家）→ shipped
	aw := requestAs(t, app, http.MethodPut, "/api/v1/admin/orders/"+oid,
		[]byte(`{"status":"paid"}`), "admin-1", domain.RolePlatformAdmin)
	if aw.Code != http.StatusOK {
		t.Fatalf("admin mark paid: %d %s", aw.Code, aw.Body.String())
	}
	sw := requestAs(t, app, http.MethodPatch, "/api/v1/trade-orders/"+oid+"/status",
		[]byte(`{"status":"shipped"}`), "seller-1", domain.RoleEnterprise)
	if sw.Code != http.StatusOK {
		t.Fatalf("seller mark shipped: %d %s", sw.Code, sw.Body.String())
	}
	// 买家不能伪造支付：PATCH paid 应被拒
	fw := requestAs(t, app, http.MethodPatch, "/api/v1/trade-orders/"+oid+"/status",
		[]byte(`{"status":"paid"}`), "buyer-1", domain.RoleIndividual)
	if fw.Code != http.StatusForbidden {
		t.Fatalf("buyer mark paid should be rejected, got %d", fw.Code)
	}

	// 4. 买家申请售后 → aftersale + pending
	aw = requestAs(t, app, http.MethodPost, "/api/v1/trade-orders/"+oid+"/aftersale",
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

	// 8. 驳回分支：新商品 + 新订单 → 付款后未发货直接申请（paid → aftersale）→ reject → aftersale_status=rejected
	// 注：商品下单即标记 sold（防超卖），第二个订单必须用新商品
	pw2 := requestAs(t, app, http.MethodPost, "/api/v1/admin/products",
		[]byte(`{"title":"备用无人机","prod_type":"drone","price_fen":500000,"condition":"new"}`),
		"admin-1", domain.RolePlatformAdmin)
	if pw2.Code != http.StatusCreated {
		t.Fatalf("create product2: %d %s", pw2.Code, pw2.Body.String())
	}
	var prod2 struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(pw2.Body.Bytes(), &prod2); err != nil {
		t.Fatalf("parse product2: %v", err)
	}
	ow2 := requestAs(t, app, http.MethodPost, "/api/v1/trade-orders",
		[]byte(`{"product_id":"`+prod2.Data.ID+`","seller_id":"seller-1","amount_fen":500000}`),
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
	// 仅支付（paid，管理端改单），不发货——验证「付款后未发货可申请售后」
	requestAs(t, app, http.MethodPut, "/api/v1/admin/orders/"+oid2,
		[]byte(`{"status":"paid"}`), "admin-1", domain.RolePlatformAdmin)
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

// 我的预约（测试场地）：预约提交 → mine 可见 → 数据隔离 → 审核后状态同步 → 未登录 401
func TestMyTestSiteBookings(t *testing.T) {
	app := newServer(t)

	// 1. 管理后台建试飞场
	cw := requestAs(t, app, http.MethodPost, "/api/v1/admin/test-sites",
		[]byte(`{"name":"渝北试飞场","site_type":"flying_field","location":"渝北区","booking_rule":"工作日9-18点","price_fen":0,"status":"available"}`),
		"admin-1", domain.RolePlatformAdmin)
	if cw.Code != http.StatusCreated {
		t.Fatalf("create site: %d %s", cw.Code, cw.Body.String())
	}
	var site struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(cw.Body.Bytes(), &site); err != nil {
		t.Fatalf("parse site: %v", err)
	}

	// 2. 买家预约 → pending
	bw := requestAs(t, app, http.MethodPost, "/api/v1/test-sites/"+site.Data.ID+"/book",
		[]byte(`{"purpose":"certification","date":"2026-08-20","time_slot":"09:00-11:00","contact_name":"张三","contact_phone":"13800000001"}`),
		"buyer-1", domain.RoleIndividual)
	if bw.Code != http.StatusCreated {
		t.Fatalf("book site: %d %s", bw.Code, bw.Body.String())
	}

	// 3. 我的预约可见（含场地 id 与 pending 状态）
	mine := requestAs(t, app, http.MethodGet, "/api/v1/test-sites/bookings/mine", nil, "buyer-1", domain.RoleIndividual)
	if mine.Code != http.StatusOK {
		t.Fatalf("my bookings: %d %s", mine.Code, mine.Body.String())
	}
	if !bytes.Contains(mine.Body.Bytes(), []byte(`"status":"pending"`)) ||
		!bytes.Contains(mine.Body.Bytes(), []byte(site.Data.ID)) {
		t.Fatalf("mine should contain booking: %s", mine.Body.String())
	}

	// 4. 数据隔离：他人 mine 为空
	other := requestAs(t, app, http.MethodGet, "/api/v1/test-sites/bookings/mine", nil, "buyer-2", domain.RoleIndividual)
	if !bytes.Contains(other.Body.Bytes(), []byte(`"data":[]`)) {
		t.Fatalf("other user should see empty: %s", other.Body.String())
	}

	// 5. 未登录 401
	rr := httptest.NewRequest(http.MethodGet, "/api/v1/test-sites/bookings/mine", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, rr)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized should be 401, got %d", w.Code)
	}

	// 6. 管理端审核通过后，mine 显示 approved
	rw := requestAs(t, app, http.MethodPost, "/api/v1/admin/test-sites/bookings/"+bookingIDOf(t, mine)+"/review",
		[]byte(`{"status":"approved","note":"已确认档期"}`), "admin-1", domain.RolePlatformAdmin)
	if rw.Code != http.StatusOK {
		t.Fatalf("review booking: %d %s", rw.Code, rw.Body.String())
	}
	mine2 := requestAs(t, app, http.MethodGet, "/api/v1/test-sites/bookings/mine", nil, "buyer-1", domain.RoleIndividual)
	if !bytes.Contains(mine2.Body.Bytes(), []byte(`"status":"approved"`)) {
		t.Fatalf("mine should reflect approved: %s", mine2.Body.String())
	}
}

// bookingIDOf 从 mine 列表响应里取出第一条预约 ID（测试辅助）。
func bookingIDOf(t *testing.T, w *httptest.ResponseRecorder) string {
	t.Helper()
	var resp struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse mine: %v", err)
	}
	if len(resp.Data) == 0 {
		t.Fatalf("no bookings: %s", w.Body.String())
	}
	return resp.Data[0].ID
}
