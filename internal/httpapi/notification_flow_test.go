package httpapi_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"drone-platform/internal/domain"
)

// 站内通知是异步发送的（notify 里起 goroutine），直接断言会 flaky，
// 因此统一用轮询等待：最多 2 秒，等不到即失败。
func waitMessage(t *testing.T, app http.Handler, userID string, role domain.Role, title string) map[string]any {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for {
		w := requestAs(t, app, http.MethodGet, "/api/v1/messages", nil, userID, role)
		if w.Code != http.StatusOK {
			t.Fatalf("list messages for %s: %d %s", userID, w.Code, w.Body.String())
		}
		var out struct {
			Data []map[string]any `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
			t.Fatalf("parse messages: %v", err)
		}
		for _, m := range out.Data {
			if m["title"] == title {
				return m
			}
		}
		if time.Now().After(deadline) {
			t.Fatalf("等不到通知 %q（user=%s，现有 %d 条）", title, userID, len(out.Data))
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// mustNotHaveMessage 断言某人没有收到某标题的通知（用于验证"只通知对方"）。
func mustNotHaveMessage(t *testing.T, app http.Handler, userID string, role domain.Role, title string) {
	t.Helper()
	w := requestAs(t, app, http.MethodGet, "/api/v1/messages", nil, userID, role)
	if w.Code != http.StatusOK {
		t.Fatalf("list messages for %s: %d", userID, w.Code)
	}
	var out struct {
		Data []map[string]any `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("parse messages: %v", err)
	}
	for _, m := range out.Data {
		if m["title"] == title {
			t.Fatalf("%s 不该收到 %q，但收到了", userID, title)
		}
	}
}

// createWorkOrder 走完"发布需求 → 审核 → 投意向 → 企业接单"，返回工单 ID。
func createWorkOrder(t *testing.T, app http.Handler, title string) string {
	t.Helper()
	dw := requestAs(t, app, http.MethodPost, "/api/v1/demands",
		[]byte(`{"title":"`+title+`","contact":"13800000000","district":"渝北区"}`), "enterprise-1", domain.RoleEnterprise)
	if dw.Code != http.StatusCreated {
		t.Fatalf("create demand: %d %s", dw.Code, dw.Body.String())
	}
	demandID := dataID(t, dw)

	rw := requestAs(t, app, http.MethodPost, "/api/v1/admin/demands/"+demandID+"/review",
		[]byte(`{"action":"approve"}`), "admin-1", domain.RolePlatformAdmin)
	if rw.Code != http.StatusOK {
		t.Fatalf("review demand: %d %s", rw.Code, rw.Body.String())
	}

	iw := requestAs(t, app, http.MethodPost, "/api/v1/demands/"+demandID+"/intents",
		[]byte(`{"intentor_name":"飞手小张","contact":"13900000000"}`), "worker-1", domain.RoleIndividual)
	if iw.Code != http.StatusCreated {
		t.Fatalf("create intent: %d %s", iw.Code, iw.Body.String())
	}
	intentID := dataID(t, iw)

	aw := requestAs(t, app, http.MethodPost, "/api/v1/demands/"+demandID+"/intents/"+intentID+"/accept",
		[]byte(`{"amount_fen":100000}`), "enterprise-1", domain.RoleEnterprise)
	if aw.Code != http.StatusCreated {
		t.Fatalf("accept intent: %d %s", aw.Code, aw.Body.String())
	}
	return dataID(t, aw)
}

// 工单闭环每一步都必须通知到"对方"：接单 / 开始 / 提交 / 验收。
func TestWorkOrderLifecycleSendsNotifications(t *testing.T) {
	app := newBizServer(t)
	orderID := createWorkOrder(t, app, "电力巡检")

	m := waitMessage(t, app, "worker-1", domain.RoleIndividual, "接单成功")
	if m["resource_type"] != "work_order" || m["resource_id"] != orderID {
		t.Fatalf("接单成功未挂到工单上: resource_type=%v resource_id=%v want=%s", m["resource_type"], m["resource_id"], orderID)
	}

	sw := requestAs(t, app, http.MethodPost, "/api/v1/work-orders/"+orderID+"/start", nil, "worker-1", domain.RoleIndividual)
	if sw.Code != http.StatusOK {
		t.Fatalf("start: %d %s", sw.Code, sw.Body.String())
	}
	waitMessage(t, app, "enterprise-1", domain.RoleEnterprise, "作业已开始")

	cw := requestAs(t, app, http.MethodPost, "/api/v1/work-orders/"+orderID+"/complete",
		[]byte(`{"result_photos":["/uploads/a.jpg"]}`), "worker-1", domain.RoleIndividual)
	if cw.Code != http.StatusOK {
		t.Fatalf("complete: %d %s", cw.Code, cw.Body.String())
	}
	waitMessage(t, app, "enterprise-1", domain.RoleEnterprise, "作业待验收")

	xw := requestAs(t, app, http.MethodPost, "/api/v1/work-orders/"+orderID+"/accept", nil, "enterprise-1", domain.RoleEnterprise)
	if xw.Code != http.StatusOK {
		t.Fatalf("accept work order: %d %s", xw.Code, xw.Body.String())
	}
	waitMessage(t, app, "worker-1", domain.RoleIndividual, "验收通过")
}

// 整改与取消：整改要把备注带到飞手；取消只通知"另一方"，取消人自己不收。
func TestWorkOrderReworkAndCancelNotify(t *testing.T) {
	app := newBizServer(t)
	orderID := createWorkOrder(t, app, "测绘复核")

	if w := requestAs(t, app, http.MethodPost, "/api/v1/work-orders/"+orderID+"/start", nil, "worker-1", domain.RoleIndividual); w.Code != http.StatusOK {
		t.Fatalf("start: %d %s", w.Code, w.Body.String())
	}
	if w := requestAs(t, app, http.MethodPost, "/api/v1/work-orders/"+orderID+"/complete", nil, "worker-1", domain.RoleIndividual); w.Code != http.StatusOK {
		t.Fatalf("complete: %d %s", w.Code, w.Body.String())
	}

	rw := requestAs(t, app, http.MethodPost, "/api/v1/work-orders/"+orderID+"/rework",
		[]byte(`{"note":"成果照片模糊"}`), "enterprise-1", domain.RoleEnterprise)
	if rw.Code != http.StatusOK {
		t.Fatalf("rework: %d %s", rw.Code, rw.Body.String())
	}
	m := waitMessage(t, app, "worker-1", domain.RoleIndividual, "整改通知")
	if content, _ := m["content"].(string); !strings.Contains(content, "成果照片模糊") {
		t.Fatalf("整改通知未带上企业备注: %q", content)
	}

	// 飞手取消 → 企业收到；飞手自己不再收到一条
	cw := requestAs(t, app, http.MethodPost, "/api/v1/work-orders/"+orderID+"/cancel",
		[]byte(`{"reason":"设备故障"}`), "worker-1", domain.RoleIndividual)
	if cw.Code != http.StatusOK {
		t.Fatalf("cancel: %d %s", cw.Code, cw.Body.String())
	}
	e := waitMessage(t, app, "enterprise-1", domain.RoleEnterprise, "订单已取消")
	if content, _ := e["content"].(string); !strings.Contains(content, "设备故障") {
		t.Fatalf("取消通知未带原因: %q", content)
	}
	mustNotHaveMessage(t, app, "worker-1", domain.RoleIndividual, "订单已取消")
}

// 飞手认证审核：通过/驳回都要通知申请人本人；驳回要带理由。
// 前置：无有效证书不允许申请飞手认证（产品口径），所以先走"提交证书 → 审核通过"。
func TestPilotReviewSendsNotification(t *testing.T) {
	app := newBizServer(t)
	adminTok := authAs(t, "admin-1", domain.RolePlatformAdmin)
	_ = adminTok

	for _, tc := range []struct {
		user     string
		certNo   string
		name     string
		approve  bool
		reason   string
		wantWord string
	}{
		{user: "worker-1", certNo: "NOTI-001", name: "通知飞手甲", approve: true, wantWord: "已通过"},
		{user: "worker-2", certNo: "NOTI-002", name: "通知飞手乙", approve: false, reason: "证件照片不清晰", wantWord: "未通过"},
	} {
		cw := requestAs(t, app, http.MethodPost, "/api/v1/certificates",
			[]byte(`{"cert_type":"caac","cert_number":"`+tc.certNo+`","level":"三级","issuer_org":"CAAC"}`), tc.user, domain.RoleIndividual)
		if cw.Code != http.StatusCreated {
			t.Fatalf("[%s] create cert: %d %s", tc.user, cw.Code, cw.Body.String())
		}
		certID := dataID(t, cw)
		aw := requestAs(t, app, http.MethodPost, "/api/v1/admin/certificates/"+certID+"/approve", nil, "admin-1", domain.RolePlatformAdmin)
		if aw.Code != http.StatusOK {
			t.Fatalf("[%s] approve cert: %d %s", tc.user, aw.Code, aw.Body.String())
		}

		pw := requestAs(t, app, http.MethodPost, "/api/v1/certified-pilots",
			[]byte(`{"real_name":"`+tc.name+`","id_card":"110101199001011234","flight_hours":88}`), tc.user, domain.RoleIndividual)
		if pw.Code != http.StatusCreated {
			t.Fatalf("[%s] register pilot: %d %s", tc.user, pw.Code, pw.Body.String())
		}
		pilotID := dataID(t, pw)

		if tc.approve {
			w := requestAs(t, app, http.MethodPost, "/api/v1/admin/certified-pilots/"+pilotID+"/approve", nil, "admin-1", domain.RolePlatformAdmin)
			if w.Code != http.StatusOK {
				t.Fatalf("[%s] approve pilot: %d %s", tc.user, w.Code, w.Body.String())
			}
		} else {
			w := requestAs(t, app, http.MethodPost, "/api/v1/admin/certified-pilots/"+pilotID+"/reject",
				[]byte(`{"reason":"`+tc.reason+`"}`), "admin-1", domain.RolePlatformAdmin)
			if w.Code != http.StatusOK {
				t.Fatalf("[%s] reject pilot: %d %s", tc.user, w.Code, w.Body.String())
			}
		}

		m := waitMessage(t, app, tc.user, domain.RoleIndividual, "飞手认证结果")
		content, _ := m["content"].(string)
		if !strings.Contains(content, tc.wantWord) {
			t.Fatalf("[%s] 通知文案不对: %q 应含 %q", tc.user, content, tc.wantWord)
		}
		if tc.reason != "" && !strings.Contains(content, tc.reason) {
			t.Fatalf("[%s] 驳回理由未带进通知: %q", tc.user, content)
		}
		if m["resource_type"] != "pilot" || m["resource_id"] != pilotID {
			t.Fatalf("[%s] 未挂到飞手档案上: %v/%v", tc.user, m["resource_type"], m["resource_id"])
		}
	}
}
