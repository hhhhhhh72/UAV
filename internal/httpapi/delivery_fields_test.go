package httpapi_test

import (
	"net/http"
	"testing"

	"drone-platform/internal/domain"
)

// 交付包字段回归：培训课程（划线原价/通过率/年限/规模/简介/Hero大图）与
// 赛事（划线原价）创建后回读必须非空——前端已兼容兜底，后端补齐后真实显示。
func TestDeliveryFieldsPersist(t *testing.T) {
	app := newBizServer(t)
	admin := "admin-1"

	// 课程：管理端创建含新字段
	cw := requestAs(t, app, http.MethodPost, "/api/v1/admin/training-courses",
		[]byte(`{"title":"CAAC多旋翼班","cert_type":"caac","price_fen":980000,"original_fee":1280000,
			"pass_rate":96.8,"years":5,"scale":"规模大","intro":"机构整体简介内容内容内容内容内容内容内容内容内容内容",
			"banner":"https://cdn.example.com/banner.jpg","status":"published"}`),
		admin, domain.RolePlatformAdmin)
	if cw.Code != http.StatusCreated {
		t.Fatalf("create course: %d %s", cw.Code, cw.Body.String())
	}
	for _, needle := range []string{`"original_fee":1280000`, `"pass_rate":96.8`, `"years":5`, `"scale":"规模大"`, `"banner":"https://cdn.example.com/banner.jpg"`} {
		if !jsonContains(cw.Body.String(), needle) {
			t.Fatalf("course missing %s, got: %s", needle, cw.Body.String())
		}
	}

	// 赛事：创建含划线原价
	w := requestAs(t, app, http.MethodPost, "/api/v1/admin/competitions",
		[]byte(`{"title":"2026全国无人机大赛","fee":38000,"original_fee":48000,"start_date":"2026-10-01","end_date":"2026-10-03"}`),
		admin, domain.RolePlatformAdmin)
	if w.Code != http.StatusCreated {
		t.Fatalf("create competition: %d %s", w.Code, w.Body.String())
	}
	if !jsonContains(w.Body.String(), `"original_fee":48000`) {
		t.Fatalf("competition missing original_fee, got: %s", w.Body.String())
	}

	// 列表回读确认（用户侧接口透传）
	lw := requestAs(t, app, http.MethodGet, "/api/v1/training-courses", nil, "user-1", domain.RoleIndividual)
	if !jsonContains(lw.Body.String(), `"original_fee":1280000`) {
		t.Fatalf("course list must include original_fee, got: %s", lw.Body.String())
	}
	clw := requestAs(t, app, http.MethodGet, "/api/v1/competitions", nil, "user-1", domain.RoleIndividual)
	if !jsonContains(clw.Body.String(), `"original_fee":48000`) {
		t.Fatalf("competition list must include original_fee, got: %s", clw.Body.String())
	}
}
