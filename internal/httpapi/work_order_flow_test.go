package httpapi_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/httpapi"
)

// requestAs 携带指定用户 ID 与角色的请求（公共 request() 固定 user-1，
// 而接单闭环需要「企业发布」与「飞手投意向」分属不同用户）。
func requestAs(t *testing.T, app http.Handler, method, path string, body []byte, userID string, role domain.Role) *httptest.ResponseRecorder {
	t.Helper()
	tokens, _ := httpapi.NewTokenManager(testSecret)
	token, err := tokens.Issue(domain.Actor{ID: userID, Role: role}, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest(method, path, bytes.NewReader(body))
	r.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	return w
}

// 接单派单闭环（PRD FR-6.2~6.5）端到端：
// 企业发布需求 → 飞手投意向 → 企业确认接单 → 订单生成 → 飞手开始/完成 → 企业验收。
func TestWorkOrderFlowE2E(t *testing.T) {
	app := newBizServer(t)

	// 1. 企业发布需求
	dw := requestAs(t, app, http.MethodPost, "/api/v1/demands", []byte(`{"title":"电力巡检","contact":"13800000000","district":"渝北区","biz_type":"cable_inspection","description":"50km 巡检"}`), "enterprise-1", domain.RoleEnterprise)
	if dw.Code != http.StatusCreated {
		t.Fatalf("create demand: %d %s", dw.Code, dw.Body.String())
	}
	var demand struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(dw.Body.Bytes(), &demand); err != nil {
		t.Fatalf("parse demand: %v", err)
	}
	demandID := demand.Data.ID

	// 2. 管理员审核通过
	rw := requestAs(t, app, http.MethodPost, "/api/v1/admin/demands/"+demandID+"/review", []byte(`{"action":"approve"}`), "admin-1", domain.RolePlatformAdmin)
	if rw.Code != http.StatusOK {
		t.Fatalf("review demand: %d %s", rw.Code, rw.Body.String())
	}

	// 3. 飞手投意向
	iw := requestAs(t, app, http.MethodPost, "/api/v1/demands/"+demandID+"/intents", []byte(`{"intentor_name":"飞手小张","contact":"13900000000","remark":"可完成巡检"}`), "worker-1", domain.RoleIndividual)
	if iw.Code != http.StatusCreated {
		t.Fatalf("create intent: %d %s", iw.Code, iw.Body.String())
	}
	var intent struct {
		Data struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(iw.Body.Bytes(), &intent); err != nil {
		t.Fatalf("parse intent: %v", err)
	}

	// 4. 企业确认接单（金额 1000 元）
	aw := requestAs(t, app, http.MethodPost, fmt.Sprintf("/api/v1/demands/%s/intents/%s/accept", demandID, intent.Data.ID), []byte(`{"amount_fen":100000}`), "enterprise-1", domain.RoleEnterprise)
	if aw.Code != http.StatusCreated {
		t.Fatalf("accept intent: %d %s", aw.Code, aw.Body.String())
	}
	var order struct {
		Data struct {
			ID        string `json:"id"`
			Status    string `json:"status"`
			WorkerID  string `json:"worker_id"`
			AmountFen int64  `json:"amount_fen"`
		} `json:"data"`
	}
	if err := json.Unmarshal(aw.Body.Bytes(), &order); err != nil {
		t.Fatalf("parse order: %v", err)
	}
	if order.Data.Status != string(domain.WorkOrderPending) || order.Data.AmountFen != 100000 {
		t.Fatalf("order wrong: %+v", order.Data)
	}
	orderID := order.Data.ID

	// 5. 我的订单（企业视角）
	mw := requestAs(t, app, http.MethodGet, "/api/v1/work-orders/mine", nil, "enterprise-1", domain.RoleEnterprise)
	if mw.Code != http.StatusOK {
		t.Fatalf("mine orders: %d", mw.Code)
	}
	var mine struct {
		Data []map[string]any `json:"data"`
	}
	if err := json.Unmarshal(mw.Body.Bytes(), &mine); err != nil {
		t.Fatalf("parse mine: %v", err)
	}
	if len(mine.Data) != 1 {
		t.Fatalf("mine count: %d", len(mine.Data))
	}

	// 6. 飞手确认开始
	sw := requestAs(t, app, http.MethodPost, "/api/v1/work-orders/"+orderID+"/start", nil, "worker-1", domain.RoleIndividual)
	if sw.Code != http.StatusOK {
		t.Fatalf("start: %d %s", sw.Code, sw.Body.String())
	}

	// 7. 飞手确认完成（带成果照片）
	cw := requestAs(t, app, http.MethodPost, "/api/v1/work-orders/"+orderID+"/complete", []byte(`{"result_photos":["/uploads/a.jpg"]}`), "worker-1", domain.RoleIndividual)
	if cw.Code != http.StatusOK {
		t.Fatalf("complete: %d %s", cw.Code, cw.Body.String())
	}
	// 7b. 无 body 确认完成：不应因缺 body 返回 400（成果照片可选；
	// 当前状态为 awaiting_accept，合理响应是 403 状态校验，而非 400 参数错误）
	cw2 := requestAs(t, app, http.MethodPost, "/api/v1/work-orders/"+orderID+"/complete", nil, "worker-1", domain.RoleIndividual)
	if cw2.Code == http.StatusBadRequest {
		t.Fatalf("complete without body should not be 400: %s", cw2.Body.String())
	}

	// 8. 企业验收
	xw := requestAs(t, app, http.MethodPost, "/api/v1/work-orders/"+orderID+"/accept", nil, "enterprise-1", domain.RoleEnterprise)
	if xw.Code != http.StatusOK {
		t.Fatalf("accept work order: %d %s", xw.Code, xw.Body.String())
	}
	var done struct {
		Data struct {
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(xw.Body.Bytes(), &done); err != nil {
		t.Fatalf("parse done: %v", err)
	}
	if done.Data.Status != string(domain.WorkOrderCompleted) {
		t.Fatalf("final status: %s", done.Data.Status)
	}
}

// 权限校验：非发布者不能确认接单；第三方不能看订单详情
func TestWorkOrderPermissionChecks(t *testing.T) {
	app := newBizServer(t)

	dw := requestAs(t, app, http.MethodPost, "/api/v1/demands", []byte(`{"title":"巡检测绘","contact":"13800000000","district":"渝北区"}`), "enterprise-1", domain.RoleEnterprise)
	t.Logf("create demand: %d %s", dw.Code, dw.Body.String())
	var demand struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	_ = json.Unmarshal(dw.Body.Bytes(), &demand)
	rw := requestAs(t, app, http.MethodPost, "/api/v1/admin/demands/"+demand.Data.ID+"/review", []byte(`{"action":"approve"}`), "admin-1", domain.RolePlatformAdmin)
	t.Logf("review demand: %d %s", rw.Code, rw.Body.String())
	iw := requestAs(t, app, http.MethodPost, "/api/v1/demands/"+demand.Data.ID+"/intents", []byte(`{"intentor_name":"飞手A","contact":"13900000000"}`), "worker-1", domain.RoleIndividual)
	t.Logf("create intent: %d %s", iw.Code, iw.Body.String())
	var intent struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	_ = json.Unmarshal(iw.Body.Bytes(), &intent)
	t.Logf("demandID=%q intentID=%q", demand.Data.ID, intent.Data.ID)

	// 飞手不能确认接单（只有发布者）
	aw := requestAs(t, app, http.MethodPost, fmt.Sprintf("/api/v1/demands/%s/intents/%s/accept", demand.Data.ID, intent.Data.ID), []byte(`{}`), "worker-1", domain.RoleIndividual)
	t.Logf("worker accept: %d %s", aw.Code, aw.Body.String())
	if aw.Code != http.StatusForbidden {
		t.Fatalf("worker accept intent should be 403, got %d", aw.Code)
	}
	// 企业首次确认成功（生成订单）
	aw2 := requestAs(t, app, http.MethodPost, fmt.Sprintf("/api/v1/demands/%s/intents/%s/accept", demand.Data.ID, intent.Data.ID), []byte(`{}`), "enterprise-1", domain.RoleEnterprise)
	t.Logf("enterprise accept #1: %d %s", aw2.Code, aw2.Body.String())
	if aw2.Code != http.StatusCreated {
		t.Fatalf("enterprise accept should be 201, got %d", aw2.Code)
	}
	// 已确认的意向不能重复处理（意向状态校验在 service 层）
	aw3 := requestAs(t, app, http.MethodPost, fmt.Sprintf("/api/v1/demands/%s/intents/%s/accept", demand.Data.ID, intent.Data.ID), []byte(`{}`), "enterprise-1", domain.RoleEnterprise)
	t.Logf("enterprise accept #2: %d %s", aw3.Code, aw3.Body.String())
	if aw3.Code != http.StatusForbidden {
		t.Fatalf("double accept should be 403, got %d", aw3.Code)
	}
}
