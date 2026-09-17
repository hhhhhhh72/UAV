package httpapi_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"drone-platform/internal/domain"
)

// 「消息里一键同意」所依赖的服务端契约（端到端）：
// 企业发布需求 → 飞手登记意向 → 发布方用 GET /api/v1/intents/received 一次拿到
// 名下所有需求收到的意向（消息页据此判断"是否只有唯一一条待处理申请"），
// 再以 amount_fen=0（面议）直接确认接单并生成工单。
func TestIntentReceivedAndOneTapAccept(t *testing.T) {
	app := newBizServer(t)

	// 1. 企业发布需求
	dw := requestAs(t, app, http.MethodPost, "/api/v1/demands",
		[]byte(`{"title":"一键同意需求","contact":"13800000000","district":"渝北区","biz_type":"cable_inspection","description":"50km 巡检"}`),
		"enterprise-1", domain.RoleEnterprise)
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

	// 2. 管理员审核通过（未通过的需求不收意向）
	rw := requestAs(t, app, http.MethodPost, "/api/v1/admin/demands/"+demandID+"/review",
		[]byte(`{"action":"approve"}`), "admin-1", domain.RolePlatformAdmin)
	if rw.Code != http.StatusOK {
		t.Fatalf("review demand: %d %s", rw.Code, rw.Body.String())
	}

	// 3. 飞手投意向
	iw := requestAs(t, app, http.MethodPost, "/api/v1/demands/"+demandID+"/intents",
		[]byte(`{"intentor_name":"飞手小张","contact":"13900000000","remark":"可完成巡检"}`), "worker-1", domain.RoleIndividual)
	if iw.Code != http.StatusCreated {
		t.Fatalf("create intent: %d %s", iw.Code, iw.Body.String())
	}
	var intent struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(iw.Body.Bytes(), &intent); err != nil {
		t.Fatalf("parse intent: %v", err)
	}

	// 4. 发布方一次取回"我收到的意向"——消息页一键同意的判断依据
	lw := requestAs(t, app, http.MethodGet, "/api/v1/intents/received", nil, "enterprise-1", domain.RoleEnterprise)
	if lw.Code != http.StatusOK {
		t.Fatalf("list received intents: %d %s", lw.Code, lw.Body.String())
	}
	var received struct {
		Data []struct {
			ID          string `json:"id"`
			DemandID    string `json:"demand_id"`
			Status      string `json:"status"`
			DemandTitle string `json:"demand_title"`
		} `json:"data"`
	}
	if err := json.Unmarshal(lw.Body.Bytes(), &received); err != nil {
		t.Fatalf("parse received: %v", err)
	}
	if len(received.Data) != 1 {
		t.Fatalf("received len=%d, want 1（唯一一条待处理申请才允许一键同意）", len(received.Data))
	}
	if received.Data[0].ID != intent.Data.ID || received.Data[0].DemandID != demandID || received.Data[0].Status != "pending" {
		t.Fatalf("received[0]=%+v, want id=%s demand=%s status=pending", received.Data[0], intent.Data.ID, demandID)
	}
	if received.Data[0].DemandTitle != "一键同意需求" {
		t.Fatalf("demand_title=%q, want 一键同意需求（聚合页的「申请项目」靠它显示）", received.Data[0].DemandTitle)
	}

	// 5. 归属隔离：投意向的飞手自己在"我收到的"里应是空的
	ww := requestAs(t, app, http.MethodGet, "/api/v1/intents/received", nil, "worker-1", domain.RoleIndividual)
	if ww.Code != http.StatusOK {
		t.Fatalf("worker list received: %d %s", ww.Code, ww.Body.String())
	}
	var workerGot struct {
		Data []map[string]any `json:"data"`
	}
	if err := json.Unmarshal(ww.Body.Bytes(), &workerGot); err != nil {
		t.Fatalf("parse worker received: %v", err)
	}
	if len(workerGot.Data) != 0 {
		t.Fatalf("飞手不该在「我收到的意向」里看到数据，len=%d", len(workerGot.Data))
	}

	// 6. 完全不带令牌 → 401
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/intents/received", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated: %d, want 401", rec.Code)
	}

	// 7. 一键同意：只带 amount_fen=0（面议），直接生成工单
	aw := requestAs(t, app, http.MethodPost,
		fmt.Sprintf("/api/v1/demands/%s/intents/%s/accept", demandID, intent.Data.ID),
		[]byte(`{"amount_fen":0}`), "enterprise-1", domain.RoleEnterprise)
	if aw.Code != http.StatusCreated {
		t.Fatalf("one-tap accept: %d %s", aw.Code, aw.Body.String())
	}
	var order struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(aw.Body.Bytes(), &order); err != nil {
		t.Fatalf("parse work order: %v", err)
	}
	if order.Data.ID == "" {
		t.Fatal("一键同意后未生成工单")
	}

	// 8. 接单后该意向不再 pending —— 消息页的一键同意按钮随之消失，不会重复成单
	lw2 := requestAs(t, app, http.MethodGet, "/api/v1/intents/received", nil, "enterprise-1", domain.RoleEnterprise)
	var after struct {
		Data []struct {
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(lw2.Body.Bytes(), &after); err != nil {
		t.Fatalf("parse after: %v", err)
	}
	if len(after.Data) != 1 || after.Data[0].Status == "pending" {
		t.Fatalf("接单后意向状态应已流转，实际=%+v", after.Data)
	}
}
