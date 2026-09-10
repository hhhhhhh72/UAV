package httpapi_test

import (
	"net/http"
	"strings"
	"testing"

	"drone-platform/internal/domain"
)

// BUG-010 另一半回归：参展商申请后，小程序侧必须能查到自己的申请。
//
// 历史缺陷：只有管理端缺审核入口、小程序也缺"我的申请"，用户提交后两边都看不到结果。
func TestMyBoothApplications(t *testing.T) {
	app := newBizServer(t)
	adminTok := authAs(t, "admin-1", domain.RolePlatformAdmin)
	userTok := authAs(t, "user-1", domain.RoleIndividual)
	otherTok := authAs(t, "user-2", domain.RoleIndividual)

	// 1) 未登录 → 401
	if w := doRaw(app, http.MethodGet, "/api/v1/exhibitions/booths/mine", "", ""); w.Code != http.StatusUnauthorized {
		t.Fatalf("未登录应 401，实际 %d %s", w.Code, w.Body.String())
	}

	// 2) 没申请过 → 200 且为空
	w := doRaw(app, http.MethodGet, "/api/v1/exhibitions/booths/mine", "", userTok)
	if w.Code != http.StatusOK {
		t.Fatalf("我的申请: %d %s", w.Code, w.Body.String())
	}
	if strings.Contains(w.Body.String(), "exbk-") {
		t.Fatalf("未申请却查到了记录: %s", w.Body.String())
	}

	// 3) 建展会 + 申请展位
	w = doRaw(app, http.MethodPost, "/api/v1/admin/exhibitions",
		"{\"title\":\"我的申请回归展\",\"category\":\"行业展会\",\"location\":\"重庆\",\"booth_count\":5,\"status\":\"recruiting\"}", adminTok)
	if w.Code != http.StatusCreated {
		t.Fatalf("create exhibition: %d %s", w.Code, w.Body.String())
	}
	expoID := dataID(t, w)
	w = doRaw(app, http.MethodPost, "/api/v1/exhibitions/"+expoID+"/booths",
		"{\"booth_number\":\"C1\",\"exhibit_name\":\"巡检无人机\",\"exhibit_desc\":\"整机\"}", userTok)
	if w.Code != http.StatusCreated {
		t.Fatalf("apply booth: %d %s", w.Code, w.Body.String())
	}
	boothID := dataID(t, w)

	// 4) 申请人自己能查到，且带展会 id（前端据此过滤到对应展会）
	w = doRaw(app, http.MethodGet, "/api/v1/exhibitions/booths/mine", "", userTok)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), boothID) {
		t.Fatalf("我的申请里应包含刚提交的记录: %d %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), expoID) {
		t.Fatalf("我的申请应带 exhibition_id 供前端过滤: %s", w.Body.String())
	}

	// 5) 别人看不到我的申请
	w = doRaw(app, http.MethodGet, "/api/v1/exhibitions/booths/mine", "", otherTok)
	if w.Code != http.StatusOK || strings.Contains(w.Body.String(), boothID) {
		t.Fatalf("他人不该看到我的申请: %d %s", w.Code, w.Body.String())
	}

	// 6) 审核通过后，我的申请里状态同步为 approved
	if w := doRaw(app, http.MethodPost, "/api/v1/admin/exhibitions/booths/"+boothID+"/review",
		"{\"action\":\"approve\"}", adminTok); w.Code != http.StatusOK {
		t.Fatalf("approve: %d %s", w.Code, w.Body.String())
	}
	w = doRaw(app, http.MethodGet, "/api/v1/exhibitions/booths/mine", "", userTok)
	if !strings.Contains(w.Body.String(), "approved") {
		t.Fatalf("审核通过后我的申请应显示 approved: %s", w.Body.String())
	}
}
