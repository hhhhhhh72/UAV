package httpapi_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"os"
	"strings"
	"testing"

	"drone-platform/internal/config"
	"drone-platform/internal/domain"
	"drone-platform/internal/httpapi"
	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

// assertH5Code 断言状态码；失败时输出 method/path/code 与 body 前 200 字符。
func assertH5Code(t *testing.T, w *httptest.ResponseRecorder, method, path string, want int) {
	t.Helper()
	if w.Code == want {
		return
	}
	body := w.Body.String()
	if len(body) > 200 {
		body = body[:200]
	}
	t.Fatalf("%s %s: code=%d want=%d body=%s", method, path, w.Code, want, body)
}

// setDevMode 打开 ADMIN_DEV_MODE（注册 dev-only 的 /api/* JSON 文件路由），
// 测试结束后恢复原值（参考 auth_sms_test.go 的写法）。
func setDevMode(t *testing.T) {
	t.Helper()
	old := os.Getenv("ADMIN_DEV_MODE")
	if err := os.Setenv("ADMIN_DEV_MODE", "true"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Setenv("ADMIN_DEV_MODE", old) })
}

// newSubmitServer 构造一个装配了 ApplicationService 的最小服务器：
// newBizServer/newServer 均未调用 SetApplicationService，POST /api/submit 会 500
// "application service unavailable"，故需专用服务器。
func newSubmitServer(t *testing.T) http.Handler {
	t.Helper()
	tokens, err := httpapi.NewTokenManager(testSecret)
	if err != nil {
		t.Fatal(err)
	}
	srv := httpapi.NewServer(
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		nil, nil, nil, nil, nil, nil, nil, nil, nil,
		memory.NewUserRepository(nil), memory.NewRefreshTokenRepository(), tokens,
	)
	srv.SetApplicationService(service.NewApplicationService(memory.NewApplicationRepository()))
	return srv.Router()
}

// TestH5CompatAuthPasswordFlow 覆盖 H5 账号密码登录闭环（生产注册的 /api/auth/* 路由，
// 无需 dev 模式）：注册 → 登录 → 错误密码 → me → refresh（snake_case）→ logout。
func TestH5CompatAuthPasswordFlow(t *testing.T) {
	app := newBizServer(t)
	phone := "13800007777"
	password := "secret-pass-1"

	// 1) 注册 → 200（返回 accessToken/refreshToken）
	w := doRaw(app, http.MethodPost, "/api/auth/register",
		fmt.Sprintf(`{"phone":%q,"password":%q,"name":"测试用户"}`, phone, password), "")
	assertH5Code(t, w, http.MethodPost, "/api/auth/register", http.StatusOK)

	// 2) 用刚注册的账号密码登录 → 200 且含 access_token
	w = doRaw(app, http.MethodPost, "/api/auth/login",
		fmt.Sprintf(`{"phone":%q,"password":%q}`, phone, password), "")
	assertH5Code(t, w, http.MethodPost, "/api/auth/login", http.StatusOK)
	var login struct {
		Data struct {
			AccessToken  string `json:"accessToken"`
			RefreshToken string `json:"refreshToken"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &login); err != nil {
		t.Fatalf("parse login: %v", err)
	}
	if login.Data.AccessToken == "" || login.Data.RefreshToken == "" {
		t.Fatalf("login missing tokens: %s", w.Body.String())
	}

	// 3) 错误密码 → 401
	w = doRaw(app, http.MethodPost, "/api/auth/login",
		fmt.Sprintf(`{"phone":%q,"password":"wrong-password"}`, phone), "")
	assertH5Code(t, w, http.MethodPost, "/api/auth/login", http.StatusUnauthorized)

	// 4) 带 token 读取 me → 200
	w = doRaw(app, http.MethodGet, "/api/auth/me", "", "Bearer "+login.Data.AccessToken)
	assertH5Code(t, w, http.MethodGet, "/api/auth/me", http.StatusOK)
	if !strings.Contains(w.Body.String(), "user") {
		t.Fatalf("me response missing user: %s", w.Body.String())
	}

	// 5) refresh 用登录返回的 refresh_token（snake_case）→ 200 且返回新令牌
	w = doRaw(app, http.MethodPost, "/api/auth/refresh",
		fmt.Sprintf(`{"refresh_token":%q}`, login.Data.RefreshToken), "")
	assertH5Code(t, w, http.MethodPost, "/api/auth/refresh", http.StatusOK)
	var refreshed struct {
		Data struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &refreshed); err != nil {
		t.Fatalf("parse refresh: %v", err)
	}
	if refreshed.Data.AccessToken == "" || refreshed.Data.RefreshToken == "" {
		t.Fatalf("refresh missing tokens: %s", w.Body.String())
	}

	// 6) logout → 200
	w = doRaw(app, http.MethodPost, "/api/auth/logout",
		fmt.Sprintf(`{"refresh_token":%q}`, refreshed.Data.RefreshToken), "")
	assertH5Code(t, w, http.MethodPost, "/api/auth/logout", http.StatusOK)
}

// TestH5CompatSubmitApplication 覆盖服务申请提交（生产注册的 POST /api/submit）。
// handler 本身不强求登录（仅在有 actor 时用 actor 覆盖 userId），但 authenticate
// 中间件对非白名单 POST 强制 Bearer 鉴权，因此匿名提交实际返回 401。
func TestH5CompatSubmitApplication(t *testing.T) {
	app := newSubmitServer(t)
	token := authAs(t, "user-1", domain.RoleIndividual)

	// 带 token 提交 → 200 且返回 id
	w := doRaw(app, http.MethodPost, "/api/submit",
		`{"serviceId":"13","serviceName":"无人机培训报名","contactName":"张三","contactPhone":"13800000001"}`, token)
	assertH5Code(t, w, http.MethodPost, "/api/submit", http.StatusOK)
	var out struct {
		Data struct {
			Success bool   `json:"success"`
			ID      string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("parse submit: %v", err)
	}
	if !out.Data.Success || out.Data.ID == "" {
		t.Fatalf("submit missing success/id: %s", w.Body.String())
	}

	// 匿名提交：handler 可工作（不硬性要求登录），但中间件鉴权拦截 → 401
	w = doRaw(app, http.MethodPost, "/api/submit",
		`{"serviceId":"13","serviceName":"无人机培训报名"}`, "")
	assertH5Code(t, w, http.MethodPost, "/api/submit", http.StatusUnauthorized)
}

// TestH5CompatServicesConfig 覆盖 services config 读写：GET 公开 200；
// POST 需 platform_admin（200），匿名 403。POST 会同步 platform_config.json，
// 故先快照平台配置并在 cleanup 恢复。
func TestH5CompatServicesConfig(t *testing.T) {
	snapshot := config.GetPlatformConfig()
	t.Cleanup(func() {
		_ = config.SavePlatformConfig(snapshot)
	})

	app := newBizServer(t)

	// GET → 200（公开）
	w := doRaw(app, http.MethodGet, "/api/services/config", "", "")
	assertH5Code(t, w, http.MethodGet, "/api/services/config", http.StatusOK)

	// POST 带 platform_admin token → 200
	adminTok := authAs(t, "user-1", domain.RolePlatformAdmin)
	body := `{"config":{"_home":{"banners":[{"image":"/uploads/ban.jpg","link":"/pages/demand/list"}],"notices":["测试公告"]}}}`
	w = doRaw(app, http.MethodPost, "/api/services/config", body, adminTok)
	assertH5Code(t, w, http.MethodPost, "/api/services/config", http.StatusOK)

	// POST 匿名 → 403
	w = doRaw(app, http.MethodPost, "/api/services/config", body, "")
	assertH5Code(t, w, http.MethodPost, "/api/services/config", http.StatusForbidden)
}

// TestH5CompatDevJSONRoutes 覆盖 dev-only 的 JSON 文件路由（需 ADMIN_DEV_MODE）：
// cases 增查改删、case-categories、list。这些路径不在公开白名单，需携带 token。
func TestH5CompatDevJSONRoutes(t *testing.T) {
	setDevMode(t)
	app := newBizServer(t)
	token := authAs(t, "user-1", domain.RoleIndividual)

	// 1) 创建用例 → 200，记录 id
	title := "无人机巡检案例"
	w := doRaw(app, http.MethodPost, "/api/cases/create",
		fmt.Sprintf(`{"title":%q,"content":"某园区日常巡检","categoryId":1}`, title), token)
	assertH5Code(t, w, http.MethodPost, "/api/cases/create", http.StatusOK)
	var created struct {
		Data struct {
			ID float64 `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatalf("parse cases/create: %v", err)
	}
	if created.Data.ID == 0 {
		t.Fatalf("cases/create missing id: %s", w.Body.String())
	}
	idStr := fmt.Sprintf("%.0f", created.Data.ID)

	// 2) 列表包含刚创建的 id 与标题
	w = doRaw(app, http.MethodGet, "/api/cases", "", token)
	assertH5Code(t, w, http.MethodGet, "/api/cases", http.StatusOK)
	if !strings.Contains(w.Body.String(), idStr) {
		t.Fatalf("cases list missing created id %s: %s", idStr, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), title) {
		t.Fatalf("cases list missing title %q: %s", title, w.Body.String())
	}

	// 3) 更新用例 → 200
	w = doRaw(app, http.MethodPost, "/api/cases/update",
		fmt.Sprintf(`{"id":%s,"title":"无人机巡检案例-已更新"}`, idStr), token)
	assertH5Code(t, w, http.MethodPost, "/api/cases/update", http.StatusOK)

	// 4) 删除用例 → 200
	w = doRaw(app, http.MethodPost, "/api/cases/delete",
		fmt.Sprintf(`{"id":%s}`, idStr), token)
	assertH5Code(t, w, http.MethodPost, "/api/cases/delete", http.StatusOK)

	// 5) case-categories → 200 且包含种子分类
	w = doRaw(app, http.MethodGet, "/api/case-categories", "", token)
	assertH5Code(t, w, http.MethodGet, "/api/case-categories", http.StatusOK)
	if !strings.Contains(w.Body.String(), "无人机物流") {
		t.Fatalf("case-categories missing seeded category: %s", w.Body.String())
	}

	// 6) list（服务申请列表）→ 200
	w = doRaw(app, http.MethodGet, "/api/list", "", token)
	assertH5Code(t, w, http.MethodGet, "/api/list", http.StatusOK)
}

// TestH5CompatDevUploadAndBoundary 覆盖 dev-only 的上传与边界路由：
// multipart 真 PNG 上传（201）、微信 OAuth URL（200，/api/auth/* 免鉴权）、
// SSO 登录（恒 400，但路径非公开需带 token 才能到达 handler）。
func TestH5CompatDevUploadAndBoundary(t *testing.T) {
	setDevMode(t)
	app := newBizServer(t)

	// 1) multipart 上传真 PNG → 201
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	h := textproto.MIMEHeader{}
	h.Set("Content-Disposition", `form-data; name="file"; filename="cover.png"`)
	h.Set("Content-Type", "image/png")
	part, err := mw.CreatePart(h)
	if err != nil {
		t.Fatal(err)
	}
	var pngBuf bytes.Buffer
	if err := png.Encode(&pngBuf, image.NewRGBA(image.Rect(0, 0, 1, 1))); err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(pngBuf.Bytes()); err != nil {
		t.Fatal(err)
	}
	if err := mw.Close(); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/upload", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("Authorization", authAs(t, "user-1", domain.RoleIndividual))
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)
	assertH5Code(t, w, http.MethodPost, "/api/upload", http.StatusCreated)
	if !strings.Contains(w.Body.String(), "file_id") {
		t.Fatalf("upload response missing file_id: %s", w.Body.String())
	}

	// 2) 微信 OAuth URL → 200（/api/auth/* 前缀免鉴权，无 token 也 200）
	w = doRaw(app, http.MethodGet, "/api/auth/wechat-oauth-url", "", "")
	assertH5Code(t, w, http.MethodGet, "/api/auth/wechat-oauth-url", http.StatusOK)
	if !strings.Contains(w.Body.String(), "authUrl") {
		t.Fatalf("wechat-oauth-url missing authUrl: %s", w.Body.String())
	}

	// 3) SSO 登录恒失败 → 400（路径非公开，需 token 才到达 handler）
	w = doRaw(app, http.MethodPost, "/api/sso/login", `{}`, authAs(t, "user-1", domain.RoleIndividual))
	assertH5Code(t, w, http.MethodPost, "/api/sso/login", http.StatusBadRequest)
}
