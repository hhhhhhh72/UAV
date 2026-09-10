package httpapi_test

import (
	"net/http"
	"strings"
	"testing"

	"drone-platform/internal/domain"
)

// 发布者删除自己的需求：只允许"已下架(cancelled)/未通过(rejected)"；
// 在架 → 409（状态不允许，引导先下架）；非属主 → 403；未登录 → 401。
//
// 状态门槛不可放宽的原因：需求下挂着 demand_intents（ON DELETE CASCADE）与 work_orders
// （外键 RESTRICT），硬删会连带清掉对接意向、有工单时直接报外键错。
func TestDemandOwnerDelete(t *testing.T) {
	app := newBizServer(t)

	dw := requestAs(t, app, http.MethodPost, "/api/v1/demands",
		[]byte(`{"title":"待删需求","contact":"13800000000","district":"渝北区"}`), "enterprise-1", domain.RoleEnterprise)
	if dw.Code != http.StatusCreated {
		t.Fatalf("create demand: %d %s", dw.Code, dw.Body.String())
	}
	id := dataID(t, dw)

	rw := requestAs(t, app, http.MethodPost, "/api/v1/admin/demands/"+id+"/review",
		[]byte(`{"action":"approve"}`), "admin-1", domain.RolePlatformAdmin)
	if rw.Code != http.StatusOK {
		t.Fatalf("review demand: %d %s", rw.Code, rw.Body.String())
	}

	// 1) 在架状态删除 → 409（资源状态冲突，不是权限问题）
	w := requestAs(t, app, http.MethodDelete, "/api/v1/demands/"+id, nil, "enterprise-1", domain.RoleEnterprise)
	if w.Code != http.StatusConflict {
		t.Fatalf("在架需求应 409，实际 %d %s", w.Code, w.Body.String())
	}

	// 2) 非属主删除 → 403
	w = requestAs(t, app, http.MethodDelete, "/api/v1/demands/"+id, nil, "worker-1", domain.RoleIndividual)
	if w.Code != http.StatusForbidden {
		t.Fatalf("非属主应 403，实际 %d %s", w.Code, w.Body.String())
	}

	// 3) 未登录 → 401（不带 token）
	w = doRaw(app, http.MethodDelete, "/api/v1/demands/"+id, "", "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("未登录应 401，实际 %d %s", w.Code, w.Body.String())
	}

	// 4) 发布者先下架 → 再删除 → 200
	cw := requestAs(t, app, http.MethodPost, "/api/v1/demands/"+id+"/cancel", nil, "enterprise-1", domain.RoleEnterprise)
	if cw.Code != http.StatusOK {
		t.Fatalf("cancel demand: %d %s", cw.Code, cw.Body.String())
	}
	w = requestAs(t, app, http.MethodDelete, "/api/v1/demands/"+id, nil, "enterprise-1", domain.RoleEnterprise)
	if w.Code != http.StatusOK {
		t.Fatalf("已下架需求应可删，实际 %d %s", w.Code, w.Body.String())
	}

	// 5) 删除后不再出现在"我的发布"
	lw := requestAs(t, app, http.MethodGet, "/api/v1/demands?mine=1&page_size=100", nil, "enterprise-1", domain.RoleEnterprise)
	if lw.Code != http.StatusOK {
		t.Fatalf("list mine: %d %s", lw.Code, lw.Body.String())
	}
	if strings.Contains(lw.Body.String(), id) {
		t.Fatalf("删除后仍出现在我的发布里: %s", lw.Body.String())
	}

	// 6) 重复删除 → 非 2xx（资源已不存在），不能静默成功
	w = requestAs(t, app, http.MethodDelete, "/api/v1/demands/"+id, nil, "enterprise-1", domain.RoleEnterprise)
	if w.Code >= 200 && w.Code < 300 {
		t.Fatalf("重复删除不应成功: %d %s", w.Code, w.Body.String())
	}
}

// 管理员通道不受影响：仍可删任意"已取消/已驳回"需求（含他人发布的）。
func TestAdminDemandDeleteUnchanged(t *testing.T) {
	app := newBizServer(t)

	dw := requestAs(t, app, http.MethodPost, "/api/v1/demands",
		[]byte(`{"title":"管理员待删","contact":"13800000000","district":"渝北区"}`), "enterprise-1", domain.RoleEnterprise)
	if dw.Code != http.StatusCreated {
		t.Fatalf("create demand: %d %s", dw.Code, dw.Body.String())
	}
	id := dataID(t, dw)
	if w := requestAs(t, app, http.MethodPost, "/api/v1/admin/demands/"+id+"/review",
		[]byte(`{"action":"approve"}`), "admin-1", domain.RolePlatformAdmin); w.Code != http.StatusOK {
		t.Fatalf("review: %d %s", w.Code, w.Body.String())
	}

	// 在架 → 管理员也删不了（状态门槛对管理员同样生效）
	if w := requestAs(t, app, http.MethodDelete, "/api/v1/admin/demands/"+id, nil, "admin-1", domain.RolePlatformAdmin); w.Code != http.StatusBadRequest {
		t.Fatalf("在架需求管理员通道应 400，实际 %d %s", w.Code, w.Body.String())
	}

	// 下架后 → 管理员可删
	if w := requestAs(t, app, http.MethodPost, "/api/v1/demands/"+id+"/cancel", nil, "enterprise-1", domain.RoleEnterprise); w.Code != http.StatusOK {
		t.Fatalf("cancel: %d %s", w.Code, w.Body.String())
	}
	if w := requestAs(t, app, http.MethodDelete, "/api/v1/admin/demands/"+id, nil, "admin-1", domain.RolePlatformAdmin); w.Code != http.StatusOK {
		t.Fatalf("管理员删已下架需求应 200，实际 %d %s", w.Code, w.Body.String())
	}
}
