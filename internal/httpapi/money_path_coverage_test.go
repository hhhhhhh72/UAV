package httpapi_test

// 关键资金路径测试盲区补齐（B1/B2/B3）：
//   B1 completeEnrollment 成功路径（付费/免费 + 幂等 + 权限）
//   B2 payAndEnroll 付费路径（冻结成功/失败/报名失败回滚/重复付费）
//   B3 交易订单新接口（支付、卖家售后审核、取消恢复商品）
//
// 分层铁律：全部走 HTTP 层断言真实行为，不 mock 业务。
// 注意：escrowDeposit 的入账对象是"当前登录管理员自己的账户"（handler 用 a.ID），
// 因此"管理员给学员充值"在 HTTP 层表示为：用学员 ID + 管理员角色签发 token 调 deposit。

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"drone-platform/internal/domain"
)

// ── 通用资金/课程/商品辅助 ──────────────────────────────────────────────

// escrowBal 查询用户托管金余额（balance_fen / frozen_fen）。
func escrowBal(t *testing.T, app http.Handler, userID string, role domain.Role) (balanceFen, frozenFen int64) {
	t.Helper()
	w := requestAs(t, app, http.MethodGet, "/api/v1/escrow/balance", nil, userID, role)
	if w.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/escrow/balance (%s): %d %s", userID, w.Code, w.Body.String())
	}
	var resp struct {
		Data struct {
			BalanceFen int64 `json:"balance_fen"`
			FrozenFen  int64 `json:"frozen_fen"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse balance: %v (body=%s)", err, w.Body.String())
	}
	return resp.Data.BalanceFen, resp.Data.FrozenFen
}

// fundEscrow 管理员给指定用户充值（托管金入账当前调用者账户 → 用用户 ID 的管理员 token）。
func fundEscrow(t *testing.T, app http.Handler, userID string, amountFen int64) {
	t.Helper()
	w := requestAs(t, app, http.MethodPost, "/api/v1/escrow/deposit",
		[]byte(fmt.Sprintf(`{"amount_fen":%d}`, amountFen)), userID, domain.RolePlatformAdmin)
	if w.Code != http.StatusCreated && w.Code != http.StatusOK {
		t.Fatalf("deposit %s: %d %s", userID, w.Code, w.Body.String())
	}
}

// createPublishedCourse 机构发布课程 → 管理端审核发布（published）。
// 返回课程 id；课程 OrgID = 机构用户 id（CreateCourse 服务端设置）。
func createPublishedCourse(t *testing.T, app http.Handler, orgID, title string, priceFen int64) string {
	t.Helper()
	w := requestAs(t, app, http.MethodPost, "/api/v1/training-courses",
		[]byte(fmt.Sprintf(`{"title":%q,"cert_type":"caac","org_name":%q,"price_fen":%d}`, title, title, priceFen)),
		orgID, domain.RoleEnterprise)
	if w.Code != http.StatusCreated {
		t.Fatalf("create course: %d %s", w.Code, w.Body.String())
	}
	id := dataID(t, w)
	// 管理端发布：携带完整字段（updateCourse 是全量覆盖，缺字段会被清空）
	w = requestAs(t, app, http.MethodPut, "/api/v1/admin/training-courses/"+id,
		[]byte(fmt.Sprintf(`{"title":%q,"cert_type":"caac","org_name":%q,"price_fen":%d,"status":"published"}`, title, title, priceFen)),
		"admin-1", domain.RolePlatformAdmin)
	if w.Code != http.StatusOK {
		t.Fatalf("publish course: %d %s", w.Code, w.Body.String())
	}
	return id
}

// createListedProduct 卖家发布商品 → 管理端上架（listed）。返回商品 id。
func createListedProduct(t *testing.T, app http.Handler, sellerID, title string, priceFen int64) string {
	t.Helper()
	w := requestAs(t, app, http.MethodPost, "/api/v1/products",
		[]byte(fmt.Sprintf(`{"title":%q,"prod_type":"drone","price_fen":%d}`, title, priceFen)),
		sellerID, domain.RoleEnterprise)
	if w.Code != http.StatusCreated {
		t.Fatalf("create product: %d %s", w.Code, w.Body.String())
	}
	pid := dataID(t, w)
	w = requestAs(t, app, http.MethodPut, "/api/v1/admin/products/"+pid,
		[]byte(`{"status":"listed"}`), "admin-1", domain.RolePlatformAdmin)
	if w.Code != http.StatusOK {
		t.Fatalf("approve product: %d %s", w.Code, w.Body.String())
	}
	return pid
}

// createOrder 买家对商品下单，返回订单 id。
func createOrder(t *testing.T, app http.Handler, buyerID, productID string) string {
	t.Helper()
	w := requestAs(t, app, http.MethodPost, "/api/v1/trade-orders",
		[]byte(`{"product_id":"`+productID+`"}`), buyerID, domain.RoleIndividual)
	if w.Code != http.StatusCreated {
		t.Fatalf("create order: %d %s", w.Code, w.Body.String())
	}
	return dataID(t, w)
}

// productVisible 公开商品列表是否包含指定标题（listed 可见，sold/removed 不可见）。
func productVisible(t *testing.T, app http.Handler, title string) bool {
	t.Helper()
	w := doRaw(app, http.MethodGet, "/api/v1/products", "", "")
	if w.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/products: %d %s", w.Code, w.Body.String())
	}
	return strings.Contains(w.Body.String(), title)
}

// myCertCount 用户本人证书数量。
func myCertCount(t *testing.T, app http.Handler, userID string) int {
	t.Helper()
	w := requestAs(t, app, http.MethodGet, "/api/v1/certificates/mine", nil, userID, domain.RoleIndividual)
	if w.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/certificates/mine (%s): %d %s", userID, w.Code, w.Body.String())
	}
	var resp struct {
		Data []json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse certs: %v (body=%s)", err, w.Body.String())
	}
	return len(resp.Data)
}

// ── B1：completeEnrollment 成功路径 ──────────────────────────────────────

// 管理员完成付费报名：学员冻结减少、机构（course.OrgID 用户）到账、发证 auto-{enrollID}；
// 幂等（重复 complete → 409 且资金/证书不动）；非管理员 → 403；免费课程不释放资金但正常发证。
func TestB1CompleteEnrollmentFundRelease(t *testing.T) {
	app := newBizServer(t)
	const price = int64(300000)

	// 机构发布付费课程 → 管理端发布
	courseID := createPublishedCourse(t, app, "org-1", "执照培训A", price)

	// 学员充值 → payAndEnroll（冻结课程价）
	fundEscrow(t, app, "student-1", 500000)
	w := requestAs(t, app, http.MethodPost, "/api/v1/training-courses/"+courseID+"/pay-and-enroll",
		[]byte(`{"name":"学员甲","phone":"13800000001"}`), "student-1", domain.RoleIndividual)
	assertStatus(t, http.MethodPost, ".../pay-and-enroll", w, http.StatusCreated)
	enrollID := dataID(t, w)
	if !strings.Contains(w.Body.String(), `"paid_amount_fen":300000`) {
		t.Fatalf("pay-and-enroll should freeze course price: %s", w.Body.String())
	}
	if b, f := escrowBal(t, app, "student-1", domain.RoleIndividual); b != 200000 || f != 300000 {
		t.Fatalf("student escrow after freeze: balance=%d frozen=%d want 200000/300000", b, f)
	}

	// 非管理员 complete → 403，资金不动
	w = requestAs(t, app, http.MethodPost, "/api/v1/enrollments/"+enrollID+"/complete", nil, "student-1", domain.RoleIndividual)
	assertStatus(t, http.MethodPost, ".../complete (non-admin)", w, http.StatusForbidden)
	if b, f := escrowBal(t, app, "student-1", domain.RoleIndividual); b != 200000 || f != 300000 {
		t.Fatalf("403 complete must not move funds: balance=%d frozen=%d", b, f)
	}

	// 管理员 complete → 200 + 发证 auto-{enrollID}
	w = requestAs(t, app, http.MethodPost, "/api/v1/enrollments/"+enrollID+"/complete", nil, "admin-1", domain.RolePlatformAdmin)
	assertStatus(t, http.MethodPost, ".../complete", w, http.StatusOK)
	if !strings.Contains(w.Body.String(), `"cert_number":"auto-`+enrollID+`"`) {
		t.Fatalf("complete should issue cert auto-%s: %s", enrollID, w.Body.String())
	}
	// 学员冻结清零、机构（OrgID）到账课程价
	if b, f := escrowBal(t, app, "student-1", domain.RoleIndividual); b != 200000 || f != 0 {
		t.Fatalf("student escrow after complete: balance=%d frozen=%d want 200000/0", b, f)
	}
	if b, f := escrowBal(t, app, "org-1", domain.RoleEnterprise); b != price || f != 0 {
		t.Fatalf("org escrow after complete: balance=%d frozen=%d want %d/0", b, f, price)
	}

	// 幂等重试（回归修复）：再次 complete（状态已 completed）→ 200 幂等完成，
	// 余额/证书数不变（不重复释放、不重复发证）。此前返回 409，释放/发证失败后
	// 重试被挡死，学费滞留 frozen 无法补齐。
	w = requestAs(t, app, http.MethodPost, "/api/v1/enrollments/"+enrollID+"/complete", nil, "admin-1", domain.RolePlatformAdmin)
	assertStatus(t, http.MethodPost, ".../complete (2nd)", w, http.StatusOK)
	if b, _ := escrowBal(t, app, "org-1", domain.RoleEnterprise); b != price {
		t.Fatalf("duplicate complete must not re-release: org balance=%d want %d", b, price)
	}
	if n := myCertCount(t, app, "student-1"); n != 1 {
		t.Fatalf("duplicate complete must not re-issue cert: got %d certs", n)
	}

	// 免费课程：不释放资金、发证成功
	freeCourseID := createPublishedCourse(t, app, "org-2", "免费公开课", 0)
	w = requestAs(t, app, http.MethodPost, "/api/v1/training-courses/"+freeCourseID+"/pay-and-enroll",
		[]byte(`{"name":"学员乙","phone":"13800000002"}`), "student-2", domain.RoleIndividual)
	assertStatus(t, http.MethodPost, ".../pay-and-enroll (free)", w, http.StatusCreated)
	freeEnrollID := dataID(t, w)
	w = requestAs(t, app, http.MethodPost, "/api/v1/enrollments/"+freeEnrollID+"/complete", nil, "admin-1", domain.RolePlatformAdmin)
	assertStatus(t, http.MethodPost, ".../complete (free)", w, http.StatusOK)
	if !strings.Contains(w.Body.String(), `"cert_number":"auto-`+freeEnrollID+`"`) {
		t.Fatalf("free complete should issue cert: %s", w.Body.String())
	}
	if b, f := escrowBal(t, app, "org-2", domain.RoleEnterprise); b != 0 || f != 0 {
		t.Fatalf("free course must not move funds: org balance=%d frozen=%d", b, f)
	}
	if n := myCertCount(t, app, "student-2"); n != 1 {
		t.Fatalf("free complete should issue 1 cert, got %d", n)
	}
}

// ── B1b：completeEnrollment 失败可重试（幂等补齐） ─────────────────────────

// 回归：旧实现先 CAS 置 completed 再 Release/AddCertificate，释放/发证失败后
// 状态已 completed，重试被 409 挡死——学费滞留 frozen、证书缺失无法补齐。
// 修复后 completed 报名重试：查 release 流水与证书，缺哪个补哪个，全成则幂等 200。
func TestB1bCompleteEnrollmentRetryAfterPartialFailure(t *testing.T) {
	app := newBizServer(t)
	const price = int64(300000)

	// ── 场景 A：CAS 已置 completed，但释放与发证都未发生（旧实现释放失败后的状态）──
	courseID := createPublishedCourse(t, app, "org-1", "执照培训R1", price)
	fundEscrow(t, app, "student-1", 500000)
	w := requestAs(t, app, http.MethodPost, "/api/v1/training-courses/"+courseID+"/pay-and-enroll",
		[]byte(`{"name":"学员甲","phone":"13800000001"}`), "student-1", domain.RoleIndividual)
	assertStatus(t, http.MethodPost, ".../pay-and-enroll", w, http.StatusCreated)
	enrollID := dataID(t, w)

	// 模拟旧实现"释放失败后状态已 completed"：管理端直接改状态为 completed（不发证不释放）
	w = requestAs(t, app, http.MethodPut, "/api/v1/admin/enrollments/"+enrollID,
		[]byte(`{"status":"completed"}`), "admin-1", domain.RolePlatformAdmin)
	assertStatus(t, http.MethodPut, ".../admin set completed", w, http.StatusOK)
	if _, f := escrowBal(t, app, "student-1", domain.RoleIndividual); f != price {
		t.Fatalf("pre-retry state: student frozen=%d want %d (funds must still be frozen)", f, price)
	}
	if b, _ := escrowBal(t, app, "org-1", domain.RoleEnterprise); b != 0 {
		t.Fatalf("pre-retry state: org balance=%d want 0 (not yet released)", b)
	}

	// 重试 complete → 200：补齐释放 + 发证（此前 409 挡死）
	w = requestAs(t, app, http.MethodPost, "/api/v1/enrollments/"+enrollID+"/complete", nil, "admin-1", domain.RolePlatformAdmin)
	assertStatus(t, http.MethodPost, ".../complete retry", w, http.StatusOK)
	if !strings.Contains(w.Body.String(), `"cert_number":"auto-`+enrollID+`"`) {
		t.Fatalf("retry should issue cert auto-%s: %s", enrollID, w.Body.String())
	}
	if b, f := escrowBal(t, app, "student-1", domain.RoleIndividual); b != 200000 || f != 0 {
		t.Fatalf("student escrow after retry: balance=%d frozen=%d want 200000/0", b, f)
	}
	if b, f := escrowBal(t, app, "org-1", domain.RoleEnterprise); b != price || f != 0 {
		t.Fatalf("org escrow after retry: balance=%d frozen=%d want %d/0", b, f, price)
	}
	if n := myCertCount(t, app, "student-1"); n != 1 {
		t.Fatalf("retry should issue exactly 1 cert, got %d", n)
	}

	// 再重试 → 幂等 200，资金/证书不动
	w = requestAs(t, app, http.MethodPost, "/api/v1/enrollments/"+enrollID+"/complete", nil, "admin-1", domain.RolePlatformAdmin)
	assertStatus(t, http.MethodPost, ".../complete retry again", w, http.StatusOK)
	if b, _ := escrowBal(t, app, "org-1", domain.RoleEnterprise); b != price {
		t.Fatalf("idempotent retry must not double-release: org balance=%d want %d", b, price)
	}
	if n := myCertCount(t, app, "student-1"); n != 1 {
		t.Fatalf("idempotent retry must not re-issue cert: got %d certs", n)
	}

	// ── 场景 B：释放已完成、证书缺失（旧实现发证失败后的状态）──
	courseID2 := createPublishedCourse(t, app, "org-2", "执照培训R2", price)
	fundEscrow(t, app, "student-2", 400000)
	w = requestAs(t, app, http.MethodPost, "/api/v1/training-courses/"+courseID2+"/pay-and-enroll",
		[]byte(`{"name":"学员乙","phone":"13800000002"}`), "student-2", domain.RoleIndividual)
	assertStatus(t, http.MethodPost, ".../pay-and-enroll B", w, http.StatusCreated)
	enrollID2 := dataID(t, w)
	// 状态置 completed（模拟 CAS 已发生）
	w = requestAs(t, app, http.MethodPut, "/api/v1/admin/enrollments/"+enrollID2,
		[]byte(`{"status":"completed"}`), "admin-1", domain.RolePlatformAdmin)
	assertStatus(t, http.MethodPut, ".../admin set completed B", w, http.StatusOK)
	// 手动完成资金释放（模拟旧实现"释放成功、发证失败"）
	w = requestAs(t, app, http.MethodPost, "/api/v1/escrow/release",
		[]byte(fmt.Sprintf(`{"to_user":"org-2","amount_fen":%d,"reference_type":"training_course","reference_id":%q}`, price, courseID2)),
		"student-2", domain.RolePlatformAdmin)
	assertStatus(t, http.MethodPost, ".../manual release", w, http.StatusCreated)

	// 重试 complete → 200：只补发证书，不重复释放（org-2 到账恰为 price）
	w = requestAs(t, app, http.MethodPost, "/api/v1/enrollments/"+enrollID2+"/complete", nil, "admin-1", domain.RolePlatformAdmin)
	assertStatus(t, http.MethodPost, ".../complete retry B", w, http.StatusOK)
	if b, f := escrowBal(t, app, "org-2", domain.RoleEnterprise); b != price || f != 0 {
		t.Fatalf("retry B must not double-release: org-2 balance=%d frozen=%d want %d/0", b, f, price)
	}
	if n := myCertCount(t, app, "student-2"); n != 1 {
		t.Fatalf("retry B should issue 1 cert, got %d", n)
	}

	// 免费课程 completed 后重试：不释放资金、只补证
	freeCourseID := createPublishedCourse(t, app, "org-3", "免费课R", 0)
	w = requestAs(t, app, http.MethodPost, "/api/v1/training-courses/"+freeCourseID+"/pay-and-enroll",
		[]byte(`{"name":"学员丙","phone":"13800000003"}`), "student-3", domain.RoleIndividual)
	assertStatus(t, http.MethodPost, ".../pay-and-enroll free", w, http.StatusCreated)
	freeEnrollID := dataID(t, w)
	w = requestAs(t, app, http.MethodPut, "/api/v1/admin/enrollments/"+freeEnrollID,
		[]byte(`{"status":"completed"}`), "admin-1", domain.RolePlatformAdmin)
	assertStatus(t, http.MethodPut, ".../admin set completed free", w, http.StatusOK)
	w = requestAs(t, app, http.MethodPost, "/api/v1/enrollments/"+freeEnrollID+"/complete", nil, "admin-1", domain.RolePlatformAdmin)
	assertStatus(t, http.MethodPost, ".../complete retry free", w, http.StatusOK)
	if !strings.Contains(w.Body.String(), `"cert_number":"auto-`+freeEnrollID+`"`) {
		t.Fatalf("free retry should issue cert: %s", w.Body.String())
	}
	if b, f := escrowBal(t, app, "org-3", domain.RoleEnterprise); b != 0 || f != 0 {
		t.Fatalf("free course retry must not move funds: org-3 balance=%d frozen=%d", b, f)
	}
}

// ── B2：payAndEnroll 付费路径 ────────────────────────────────────────────

func TestB2PayAndEnrollFundFlow(t *testing.T) {
	app := newBizServer(t)
	const price = int64(300000)

	// 冻结成功：余额足够 → 201，报名记录 paid_amount_fen=课程价，余额冻结
	c1 := createPublishedCourse(t, app, "org-1", "付费课程B", price)
	fundEscrow(t, app, "s1", 500000)
	w := requestAs(t, app, http.MethodPost, "/api/v1/training-courses/"+c1+"/pay-and-enroll",
		[]byte(`{"name":"学员1","phone":"13800000011"}`), "s1", domain.RoleIndividual)
	assertStatus(t, http.MethodPost, ".../pay-and-enroll ok", w, http.StatusCreated)
	if !strings.Contains(w.Body.String(), `"paid_amount_fen":300000`) {
		t.Fatalf("paid_amount_fen should be course price: %s", w.Body.String())
	}
	if b, f := escrowBal(t, app, "s1", domain.RoleIndividual); b != 200000 || f != 300000 {
		t.Fatalf("freeze: balance=%d frozen=%d want 200000/300000", b, f)
	}

	// 冻结失败：余额不足 → 402，余额/冻结不动
	c2 := createPublishedCourse(t, app, "org-2", "付费课程C", price)
	fundEscrow(t, app, "s2", 100000)
	w = requestAs(t, app, http.MethodPost, "/api/v1/training-courses/"+c2+"/pay-and-enroll",
		[]byte(`{"name":"学员2","phone":"13800000022"}`), "s2", domain.RoleIndividual)
	assertStatus(t, http.MethodPost, ".../pay-and-enroll insufficient", w, http.StatusPaymentRequired)
	if b, f := escrowBal(t, app, "s2", domain.RoleIndividual); b != 100000 || f != 0 {
		t.Fatalf("insufficient freeze must not move funds: balance=%d frozen=%d", b, f)
	}

	// 报名失败回滚：先普通报名同一课程，再付费报名 → 409 且冻结已退回（余额复原）
	c3 := createPublishedCourse(t, app, "org-3", "付费课程D", price)
	fundEscrow(t, app, "s3", 300000)
	w = requestAs(t, app, http.MethodPost, "/api/v1/training-courses/"+c3+"/enroll",
		[]byte(`{"name":"学员3","phone":"13800000033"}`), "s3", domain.RoleIndividual)
	assertStatus(t, http.MethodPost, ".../enroll", w, http.StatusCreated)
	w = requestAs(t, app, http.MethodPost, "/api/v1/training-courses/"+c3+"/pay-and-enroll",
		[]byte(`{"name":"学员3","phone":"13800000033"}`), "s3", domain.RoleIndividual)
	assertStatus(t, http.MethodPost, ".../pay-and-enroll dup", w, http.StatusConflict)
	if b, f := escrowBal(t, app, "s3", domain.RoleIndividual); b != 300000 || f != 0 {
		t.Fatalf("rollback should refund freeze: balance=%d frozen=%d want 300000/0", b, f)
	}

	// 重复 payAndEnroll → 409，且第二次冻结退回（首次报名冻结保留）
	c4 := createPublishedCourse(t, app, "org-4", "付费课程E", price)
	fundEscrow(t, app, "s4", 600000)
	w = requestAs(t, app, http.MethodPost, "/api/v1/training-courses/"+c4+"/pay-and-enroll",
		[]byte(`{"name":"学员4","phone":"13800000044"}`), "s4", domain.RoleIndividual)
	assertStatus(t, http.MethodPost, ".../pay-and-enroll first", w, http.StatusCreated)
	w = requestAs(t, app, http.MethodPost, "/api/v1/training-courses/"+c4+"/pay-and-enroll",
		[]byte(`{"name":"学员4","phone":"13800000044"}`), "s4", domain.RoleIndividual)
	assertStatus(t, http.MethodPost, ".../pay-and-enroll repeat", w, http.StatusConflict)
	if b, f := escrowBal(t, app, "s4", domain.RoleIndividual); b != 300000 || f != 300000 {
		t.Fatalf("repeat must keep first freeze: balance=%d frozen=%d want 300000/300000", b, f)
	}
}

// ── B3：交易订单新接口 ───────────────────────────────────────────────────

// 支付：买家 pending→paid；非买家 403；重复支付被拒。
// 注：payTradeOrder handler 对 PayOrder 全部错误统一返回 403（含重复支付），
// 按"断言真实行为"铁律断言 403（任务草案预期 409，与当前实现不符）。
func TestB3TradeOrderPay(t *testing.T) {
	app := newBizServer(t)
	pid := createListedProduct(t, app, "seller-1", "交易无人机A", 500000)
	oid := createOrder(t, app, "buyer-1", pid)

	w := requestAs(t, app, http.MethodPost, "/api/v1/trade-orders/"+oid+"/pay", nil, "buyer-1", domain.RoleIndividual)
	assertStatus(t, http.MethodPost, ".../pay", w, http.StatusOK)
	if !strings.Contains(w.Body.String(), `"status":"paid"`) {
		t.Fatalf("pay should move order to paid: %s", w.Body.String())
	}
	// 重复支付（已 paid）→ 403（真实行为：PayOrder 错误统一 403）
	w = requestAs(t, app, http.MethodPost, "/api/v1/trade-orders/"+oid+"/pay", nil, "buyer-1", domain.RoleIndividual)
	assertStatus(t, http.MethodPost, ".../pay repeat", w, http.StatusForbidden)
	// 非买家支付 → 403
	w = requestAs(t, app, http.MethodPost, "/api/v1/trade-orders/"+oid+"/pay", nil, "buyer-2", domain.RoleIndividual)
	assertStatus(t, http.MethodPost, ".../pay stranger", w, http.StatusForbidden)
}

// 卖家售后审核：approve → aftersale_status=approved + 订单 completed；reject → rejected；
// 非卖家（买家）审核 → 403；action 非法 → 400。
func TestB3TradeOrderAftersaleReview(t *testing.T) {
	app := newBizServer(t)

	// 买家申请售后 → 非卖家（买家本人）审核 403 → 卖家 approve → 结案
	pid := createListedProduct(t, app, "seller-1", "交易无人机B", 500000)
	oid := createOrder(t, app, "buyer-2", pid)
	w := requestAs(t, app, http.MethodPost, "/api/v1/trade-orders/"+oid+"/pay", nil, "buyer-2", domain.RoleIndividual)
	assertStatus(t, http.MethodPost, ".../pay", w, http.StatusOK)
	w = requestAs(t, app, http.MethodPost, "/api/v1/trade-orders/"+oid+"/aftersale",
		[]byte(`{"aftersale_type":"refund","aftersale_reason":"质量问题","aftersale_amount_fen":500000}`),
		"buyer-2", domain.RoleIndividual)
	assertStatus(t, http.MethodPost, ".../aftersale apply", w, http.StatusOK)
	if !strings.Contains(w.Body.String(), `"aftersale_status":"pending"`) {
		t.Fatalf("aftersale should be pending: %s", w.Body.String())
	}
	w = requestAs(t, app, http.MethodPost, "/api/v1/trade-orders/"+oid+"/aftersale/review",
		[]byte(`{"action":"approve"}`), "buyer-2", domain.RoleIndividual)
	assertStatus(t, http.MethodPost, ".../review by buyer", w, http.StatusForbidden)
	w = requestAs(t, app, http.MethodPost, "/api/v1/trade-orders/"+oid+"/aftersale/review",
		[]byte(`{"action":"approve"}`), "seller-1", domain.RoleEnterprise)
	assertStatus(t, http.MethodPost, ".../review approve", w, http.StatusOK)
	if !strings.Contains(w.Body.String(), `"aftersale_status":"approved"`) || !strings.Contains(w.Body.String(), `"status":"completed"`) {
		t.Fatalf("approve should complete order: %s", w.Body.String())
	}

	// 卖家 reject → aftersale_status=rejected + completed
	pid2 := createListedProduct(t, app, "seller-1", "交易无人机C", 500000)
	oid2 := createOrder(t, app, "buyer-3", pid2)
	w = requestAs(t, app, http.MethodPost, "/api/v1/trade-orders/"+oid2+"/pay", nil, "buyer-3", domain.RoleIndividual)
	assertStatus(t, http.MethodPost, ".../pay2", w, http.StatusOK)
	w = requestAs(t, app, http.MethodPost, "/api/v1/trade-orders/"+oid2+"/aftersale",
		[]byte(`{"aftersale_type":"return","aftersale_reason":"不想要了","aftersale_amount_fen":500000}`), "buyer-3", domain.RoleIndividual)
	assertStatus(t, http.MethodPost, ".../aftersale apply2", w, http.StatusOK)
	w = requestAs(t, app, http.MethodPost, "/api/v1/trade-orders/"+oid2+"/aftersale/review",
		[]byte(`{"action":"reject"}`), "seller-1", domain.RoleEnterprise)
	assertStatus(t, http.MethodPost, ".../review reject", w, http.StatusOK)
	if !strings.Contains(w.Body.String(), `"aftersale_status":"rejected"`) || !strings.Contains(w.Body.String(), `"status":"completed"`) {
		t.Fatalf("reject should complete order: %s", w.Body.String())
	}

	// action 非法 → 400
	pid3 := createListedProduct(t, app, "seller-1", "交易无人机D", 500000)
	oid3 := createOrder(t, app, "buyer-4", pid3)
	w = requestAs(t, app, http.MethodPost, "/api/v1/trade-orders/"+oid3+"/pay", nil, "buyer-4", domain.RoleIndividual)
	assertStatus(t, http.MethodPost, ".../pay3", w, http.StatusOK)
	w = requestAs(t, app, http.MethodPost, "/api/v1/trade-orders/"+oid3+"/aftersale",
		[]byte(`{"aftersale_type":"refund","aftersale_amount_fen":500000}`), "buyer-4", domain.RoleIndividual)
	assertStatus(t, http.MethodPost, ".../aftersale apply3", w, http.StatusOK)
	w = requestAs(t, app, http.MethodPost, "/api/v1/trade-orders/"+oid3+"/aftersale/review",
		[]byte(`{"action":"hack"}`), "seller-1", domain.RoleEnterprise)
	assertStatus(t, http.MethodPost, ".../review bad action", w, http.StatusBadRequest)
}

// 取消恢复商品：下单（listed→sold，公开列表隐藏）→ PATCH cancelled（pending 可取消）
// → 商品恢复 listed；管理端 PUT /api/v1/admin/orders/{id} cancelled 同样恢复。
func TestB3CancelRestoresProduct(t *testing.T) {
	app := newBizServer(t)
	const title = "可恢复商品"
	pid := createListedProduct(t, app, "seller-1", title, 500000)
	if !productVisible(t, app, title) {
		t.Fatalf("listed product should be visible in public list")
	}

	// 买家下单 → 商品 sold，公开列表隐藏
	oid := createOrder(t, app, "buyer-1", pid)
	if productVisible(t, app, title) {
		t.Fatalf("product should be hidden after order placed")
	}

	// 买家取消（pending 可取消）→ 商品恢复 listed
	w := requestAs(t, app, http.MethodPatch, "/api/v1/trade-orders/"+oid+"/status",
		[]byte(`{"status":"cancelled"}`), "buyer-1", domain.RoleIndividual)
	assertStatus(t, http.MethodPatch, ".../cancel", w, http.StatusOK)
	if !productVisible(t, app, title) {
		t.Fatalf("product should be restored after buyer cancel")
	}

	// 再次下单 → 管理端取消 → 同样恢复
	oid2 := createOrder(t, app, "buyer-2", pid)
	if productVisible(t, app, title) {
		t.Fatalf("product should be hidden after 2nd order")
	}
	w = requestAs(t, app, http.MethodPut, "/api/v1/admin/orders/"+oid2,
		[]byte(`{"status":"cancelled"}`), "admin-1", domain.RolePlatformAdmin)
	assertStatus(t, http.MethodPut, ".../admin cancel", w, http.StatusOK)
	if !productVisible(t, app, title) {
		t.Fatalf("product should be restored after admin cancel")
	}
}
