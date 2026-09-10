package httpapi_test

import (
	"net/http"
	"strings"
	"testing"

	"drone-platform/internal/domain"
)

// BUG-010 回归：展位申请落库后，管理端必须能列出来并审核。
//
// 历史缺陷：申请确实写进了 exhibition_booths，但管理端没有任何展位申请入口
// （只有展会的增删改查），所以测试同学看到的是"后台未见待审核"。
func TestAdminReviewBoothApplication(t *testing.T) {
	app := newBizServer(t)
	adminTok := authAs(t, "admin-1", domain.RolePlatformAdmin)
	userTok := authAs(t, "user-1", domain.RoleIndividual)
	otherTok := authAs(t, "user-2", domain.RoleIndividual)

	// 建一个招募中的展会
	w := doRaw(app, http.MethodPost, "/api/v1/admin/exhibitions",
		"{\"title\":\"展位审核回归展\",\"category\":\"行业展会\",\"location\":\"重庆\",\"booth_count\":10,\"status\":\"recruiting\"}", adminTok)
	if w.Code != http.StatusCreated {
		t.Fatalf("create exhibition: %d %s", w.Code, w.Body.String())
	}
	expoID := dataID(t, w)

	// 参展商申请展位 → 落库为 applied
	w = doRaw(app, http.MethodPost, "/api/v1/exhibitions/"+expoID+"/booths",
		"{\"booth_number\":\"A9\",\"exhibit_name\":\"巡检无人机\",\"exhibit_desc\":\"最新机型\"}", userTok)
	if w.Code != http.StatusCreated {
		t.Fatalf("apply booth: %d %s", w.Code, w.Body.String())
	}
	boothID := dataID(t, w)

	// 1) 管理端审核列表必须能查到这条申请
	w = doRaw(app, http.MethodGet, "/api/v1/admin/exhibitions/booths", "", adminTok)
	if w.Code != http.StatusOK {
		t.Fatalf("admin list booths: %d %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), boothID) {
		t.Fatalf("审核列表里没有这条申请: %s", w.Body.String())
	}

	// 2) 普通用户不能看全平台申请
	if w := doRaw(app, http.MethodGet, "/api/v1/admin/exhibitions/booths", "", userTok); w.Code != http.StatusForbidden {
		t.Fatalf("非管理员访问审核列表应 403，实际 %d %s", w.Code, w.Body.String())
	}

	// 3) 审核通过
	w = doRaw(app, http.MethodPost, "/api/v1/admin/exhibitions/booths/"+boothID+"/review",
		"{\"action\":\"approve\"}", adminTok)
	if w.Code != http.StatusOK {
		t.Fatalf("approve booth: %d %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "approved") {
		t.Fatalf("审核通过后状态应为 approved: %s", w.Body.String())
	}

	// 4) 按状态过滤：approved 能查到、rejected 查不到
	w = doRaw(app, http.MethodGet, "/api/v1/admin/exhibitions/booths?status=approved", "", adminTok)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), boothID) {
		t.Fatalf("status=approved 应能查到: %d %s", w.Code, w.Body.String())
	}
	w = doRaw(app, http.MethodGet, "/api/v1/admin/exhibitions/booths?status=rejected", "", adminTok)
	if w.Code != http.StatusOK || strings.Contains(w.Body.String(), boothID) {
		t.Fatalf("status=rejected 不该包含已通过的申请: %d %s", w.Code, w.Body.String())
	}

	// 5) 只允许 approve/reject：写内部态（paid）必须被拒
	if w := doRaw(app, http.MethodPost, "/api/v1/admin/exhibitions/booths/"+boothID+"/review",
		"{\"action\":\"paid\"}", adminTok); w.Code != http.StatusBadRequest {
		t.Fatalf("非法 action 应 400，实际 %d %s", w.Code, w.Body.String())
	}

	// 6) 非管理员不能审核
	if w := doRaw(app, http.MethodPost, "/api/v1/admin/exhibitions/booths/"+boothID+"/review",
		"{\"action\":\"reject\"}", otherTok); w.Code != http.StatusForbidden {
		t.Fatalf("非管理员审核应 403，实际 %d %s", w.Code, w.Body.String())
	}

	// 7) 驳回
	w = doRaw(app, http.MethodPost, "/api/v1/admin/exhibitions/booths/"+boothID+"/review",
		"{\"action\":\"reject\"}", adminTok)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "rejected") {
		t.Fatalf("reject booth: %d %s", w.Code, w.Body.String())
	}
}
