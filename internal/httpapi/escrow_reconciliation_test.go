package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"drone-platform/internal/domain"
)

// 托管金对账接口：真实资金接入后的"钱从哪来"核对入口。
//
// 断言三件事：
//  1. 只有平台管理员能看（协会管理员/企业/个人/匿名一律挡住）——这是平台财务数据；
//  2. 渠道由服务端判定，不由前端传：用户自助充值记 internal_self、管理员代充记 internal_admin，
//     真实资金渠道 wechat 谁也不能自己造；
//  3. 汇总口径正确（入金笔数/金额），渠道与时间区间过滤真的生效。
func TestAdminEscrowReconciliation(t *testing.T) {
	app := newBizServer(t)

	// 用户自助充值 100 元（模拟通道 → internal_self）
	w := requestAs(t, app, http.MethodPost, "/api/v1/escrow/deposit",
		[]byte(`{"amount_fen":10000}`), "user-1", domain.RoleIndividual)
	assertStatus(t, http.MethodPost, "/api/v1/escrow/deposit", w, http.StatusCreated)
	if body := w.Body.String(); !strings.Contains(body, `"channel":"internal_self"`) {
		t.Fatalf("自助充值应记 channel=internal_self: %s", body)
	}

	// 管理员代充 250 元给 user-2（internal_admin）
	w = requestAs(t, app, http.MethodPost, "/api/v1/escrow/deposit",
		[]byte(`{"amount_fen":25000,"to_user":"user-2"}`), "admin-1", domain.RolePlatformAdmin)
	assertStatus(t, http.MethodPost, "/api/v1/escrow/deposit (admin)", w, http.StatusCreated)
	if body := w.Body.String(); !strings.Contains(body, `"channel":"internal_admin"`) {
		t.Fatalf("管理员代充应记 channel=internal_admin: %s", body)
	}

	// 平台管理员对账 internal_self：1 笔 10000 分，非真实资金
	recon := fetchReconcile(t, app, "/api/v1/admin/escrow/reconciliation?channel=internal_self", "admin-1", domain.RolePlatformAdmin, http.StatusOK)
	if recon.RealFunds || recon.DepositCount != 1 || recon.DepositFen != 10000 {
		t.Fatalf("internal_self 对账错误: %+v", recon)
	}
	if len(recon.Transactions) != 1 || recon.Transactions[0].Channel != domain.ChannelInternalSelf {
		t.Fatalf("明细应带渠道标记: %+v", recon.Transactions)
	}

	// internal_admin：1 笔 25000 分
	recon = fetchReconcile(t, app, "/api/v1/admin/escrow/reconciliation?channel=internal_admin", "admin-1", domain.RolePlatformAdmin, http.StatusOK)
	if recon.DepositCount != 1 || recon.DepositFen != 25000 {
		t.Fatalf("internal_admin 对账错误: %+v", recon)
	}

	// 微信渠道：还没有任何真实资金入账 → 0 笔（数字诚实：不能凭空显示收入）
	recon = fetchReconcile(t, app, "/api/v1/admin/escrow/reconciliation?channel=wechat", "admin-1", domain.RolePlatformAdmin, http.StatusOK)
	if !recon.RealFunds || recon.DepositCount != 0 || recon.DepositFen != 0 {
		t.Fatalf("未接入微信支付时应为 0 笔真实入金: %+v", recon)
	}

	// 不限渠道（channel=all）：两笔都在，合计 35000
	recon = fetchReconcile(t, app, "/api/v1/admin/escrow/reconciliation?channel=all", "admin-1", domain.RolePlatformAdmin, http.StatusOK)
	if recon.DepositCount != 2 || recon.DepositFen != 35000 {
		t.Fatalf("全渠道对账错误: %+v", recon)
	}

	// 日期形式：今天应为 2 笔，2099 年应为 0 笔（证明区间过滤生效）
	today := time.Now().Format("2006-01-02")
	recon = fetchReconcile(t, app, "/api/v1/admin/escrow/reconciliation?channel=all&from="+today+"&to="+today, "admin-1", domain.RolePlatformAdmin, http.StatusOK)
	if recon.DepositCount != 2 {
		t.Fatalf("今日区间应为 2 笔: %+v", recon)
	}
	recon = fetchReconcile(t, app, "/api/v1/admin/escrow/reconciliation?channel=all&from=2099-01-01", "admin-1", domain.RolePlatformAdmin, http.StatusOK)
	if recon.DepositCount != 0 || len(recon.Transactions) != 0 {
		t.Fatalf("未来区间应为空: %+v", recon)
	}

	// 日期格式错误 → 400（不能把拼错的日期当成"不限"而返回全量）
	w = requestAs(t, app, http.MethodGet, "/api/v1/admin/escrow/reconciliation?from=oops", nil, "admin-1", domain.RolePlatformAdmin)
	assertStatus(t, http.MethodGet, ".../reconciliation?from=oops", w, http.StatusBadRequest)

	// 财务数据门禁：协会管理员 403、企业/个人 403、匿名 401
	w = requestAs(t, app, http.MethodGet, "/api/v1/admin/escrow/reconciliation", nil, "admin-2", domain.RoleAssociationAdmin)
	assertStatus(t, http.MethodGet, ".../reconciliation (association_admin)", w, http.StatusForbidden)
	w = requestAs(t, app, http.MethodGet, "/api/v1/admin/escrow/reconciliation", nil, "enterprise-1", domain.RoleEnterprise)
	assertStatus(t, http.MethodGet, ".../reconciliation (enterprise)", w, http.StatusForbidden)
	w = requestAs(t, app, http.MethodGet, "/api/v1/admin/escrow/reconciliation", nil, "user-1", domain.RoleIndividual)
	assertStatus(t, http.MethodGet, ".../reconciliation (individual)", w, http.StatusForbidden)
	r := httptest.NewRequest(http.MethodGet, "/api/v1/admin/escrow/reconciliation", nil)
	w = httptest.NewRecorder()
	app.ServeHTTP(w, r)
	assertStatus(t, http.MethodGet, ".../reconciliation (anonymous)", w, http.StatusUnauthorized)
}

// reconData 对账响应体（与服务端 domain.EscrowReconcile 对齐）。
type reconData struct {
	Channel      string `json:"channel"`
	RealFunds    bool   `json:"real_funds"`
	DepositCount int    `json:"deposit_count"`
	DepositFen   int64  `json:"deposit_fen"`
	Transactions []struct {
		TxType  string `json:"tx_type"`
		Channel string `json:"channel"`
		Amount  int64  `json:"amount_fen"`
	} `json:"transactions"`
}

// fetchReconcile 调对账接口并按期望状态码解析 data。
func fetchReconcile(t *testing.T, app http.Handler, path, userID string, role domain.Role, want int) reconData {
	t.Helper()
	w := requestAs(t, app, http.MethodGet, path, nil, userID, role)
	assertStatus(t, http.MethodGet, path, w, want)
	var resp struct {
		Data reconData `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("解析对账响应: %v (body=%.200s)", err, w.Body.String())
	}
	return resp.Data
}
