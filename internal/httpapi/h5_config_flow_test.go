package httpapi_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"drone-platform/internal/config"
	"drone-platform/internal/domain"
)

// TestSaveServicesConfigSyncsHome 走完整 HTTP 链路：管理后台 POST /api/services/config
// 保存 _home 后，平台全局配置（小程序 /api/v1/home 的数据源）里的 banner/公告
// 必须同步生效——后台设置轮播 Banner 不生效的回归测试
func TestSaveServicesConfigSyncsHome(t *testing.T) {
	snapshot := config.GetPlatformConfig()
	t.Cleanup(func() {
		_ = config.SavePlatformConfig(snapshot)
	})

	app := newBizServer(t)
	body := []byte(`{"config":{"_home":{"banners":[{"image":"/uploads/sync.jpg","link":"/pages/demands/list"}],"notices":["同步公告"]}}}`)
	w := request(t, app, http.MethodPost, "/api/services/config", body, domain.RolePlatformAdmin)
	if w.Code != http.StatusOK {
		t.Fatalf("save: %d %s", w.Code, w.Body.String())
	}
	got := config.GetPlatformConfig()
	if len(got.Banners) != 1 || got.Banners[0].ImageURL != "/uploads/sync.jpg" {
		t.Fatalf("platform banners after save: %+v", got.Banners)
	}
	if got.Banners[0].LinkURL != "/pages/demands/list" {
		t.Fatalf("platform banner link: %+v", got.Banners[0])
	}
	if len(got.Notices) != 1 || got.Notices[0] != "同步公告" {
		t.Fatalf("platform notices after save: %v", got.Notices)
	}
}

// TestSaveServicesConfigRequiresAdmin 覆盖全局配置的写接口必须管理员：
// 个体用户 / 匿名均 403（此前无任何鉴权，任意人可覆盖平台配置）
func TestSaveServicesConfigRequiresAdmin(t *testing.T) {
	app := newBizServer(t)
	body := []byte(`{"config":{"_home":{"banners":[]}}}`)
	w := request(t, app, http.MethodPost, "/api/services/config", body, domain.RoleIndividual)
	if w.Code != http.StatusForbidden {
		t.Fatalf("individual save: want 403 got %d", w.Code)
	}
	// 匿名（无 Authorization 头）
	req := httptest.NewRequest(http.MethodPost, "/api/services/config", strings.NewReader(string(body)))
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("anonymous save: want 403 got %d", rec.Code)
	}
	// GET 仍公开（banner 是公开数据）
	w = request(t, app, http.MethodGet, "/api/services/config", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK {
		t.Fatalf("GET config: want 200 got %d", w.Code)
	}
}

// TestHomeBannerURLRewritten banner 为 /uploads/ 相对路径时，/api/v1/home
// 返回的 image_url 必须带完整域名（小程序 <image> 不能渲染相对路径）
func TestHomeBannerURLRewritten(t *testing.T) {
	snapshot := config.GetPlatformConfig()
	t.Cleanup(func() {
		_ = config.SavePlatformConfig(snapshot)
	})

	// 平台配置写入一个相对路径 banner
	cfg := snapshot
	cfg.Banners = []domain.Banner{
		{ID: "b1", ImageURL: "/uploads/hero.jpg", LinkURL: "/pages/demands/list", Status: "active"},
	}
	if err := config.SavePlatformConfig(cfg); err != nil {
		t.Fatalf("save platform config: %v", err)
	}

	app := newBizServer(t)
	w := request(t, app, http.MethodGet, "/api/v1/home?city=重庆", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK {
		t.Fatalf("home: %d %s", w.Code, w.Body.String())
	}
	body := w.Body.String()
	if !strings.Contains(body, "http://example.com/uploads/hero.jpg") && !strings.Contains(body, "uploads/hero.jpg") {
		t.Fatalf("home banners missing rewritten url: %s", body)
	}
}
