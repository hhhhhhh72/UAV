package httpapi_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"drone-platform/internal/domain"
)

// 需求删除必须一并清掉指向它的站内通知。
//
// 不清会留死链：收件箱里还挂着「您的需求《X》收到新的对接意向」，点进去需求已不存在。
// 与「需求删除时其对接意向随 ON DELETE CASCADE 一并清除」同一口径——
// 通知是行为提醒，不是凭证。生产上已实际出现过一条这样的孤儿。
func TestDeleteDemandPurgesItsNotifications(t *testing.T) {
	app := newBizServer(t)

	// 1. 企业发布需求
	dw := requestAs(t, app, http.MethodPost, "/api/v1/demands",
		[]byte(`{"title":"待删除需求","contact":"13800000000","district":"渝北区","biz_type":"cable_inspection","description":"x"}`),
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

	// 2. 管理员审核通过
	if rw := requestAs(t, app, http.MethodPost, "/api/v1/admin/demands/"+demandID+"/review",
		[]byte(`{"action":"approve"}`), "admin-1", domain.RolePlatformAdmin); rw.Code != http.StatusOK {
		t.Fatalf("review demand: %d %s", rw.Code, rw.Body.String())
	}

	// 3. 飞手投意向 → 发布方收到「新的对接意向」
	if iw := requestAs(t, app, http.MethodPost, "/api/v1/demands/"+demandID+"/intents",
		[]byte(`{"intentor_name":"飞手小张","contact":"13900000000","remark":"可做"}`),
		"worker-1", domain.RoleIndividual); iw.Code != http.StatusCreated {
		t.Fatalf("create intent: %d %s", iw.Code, iw.Body.String())
	}
	waitMessage(t, app, "enterprise-1", domain.RoleEnterprise, "新的对接意向")

	// 4. 先取消（published → cancelled）：删除只允许已取消/已驳回的需求
	if cw := requestAs(t, app, http.MethodPost, "/api/v1/demands/"+demandID+"/cancel", nil,
		"enterprise-1", domain.RoleEnterprise); cw.Code != http.StatusOK {
		t.Fatalf("cancel demand: %d %s", cw.Code, cw.Body.String())
	}

	// 5. 删除需求
	if delw := requestAs(t, app, http.MethodDelete, "/api/v1/demands/"+demandID, nil,
		"enterprise-1", domain.RoleEnterprise); delw.Code != http.StatusOK {
		t.Fatalf("delete demand: %d %s", delw.Code, delw.Body.String())
	}

	// 关键断言：通知必须一并消失，否则收件箱留死链
	mustNotHaveMessage(t, app, "enterprise-1", domain.RoleEnterprise, "新的对接意向")
}
