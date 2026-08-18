package httpapi_test

import (
	"bytes"
	"encoding/json"
	"image"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"drone-platform/internal/domain"
)

// TestCompetitionRegisterExtendedFields: 回归 C13——小程序 register.vue 提交的
// name/phone/id_card/photo_url/id_card_image 必须落库并随响应返回；
// 非法手机号/身份证 → 400；未登录 → 401。
// 旧实现 Handler 只解析 team_name/member_count/contact_info，实名与证件字段被丢弃。
func TestCompetitionRegisterExtendedFields(t *testing.T) {
	app := newBizServer(t)
	// 管理端创建赛事
	w := request(t, app, http.MethodPost, "/api/v1/admin/competitions",
		[]byte(`{"title":"竞速赛","category":"racing","description":"FPV","location":"巴南","sponsor":"协会","start_date":"2026-09-01","end_date":"2026-09-03","max_teams":50}`),
		domain.RolePlatformAdmin)
	if w.Code != http.StatusCreated {
		t.Fatalf("create competition: %d %s", w.Code, w.Body.String())
	}
	var created struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
		ID string `json:"id"`
	}
	json.Unmarshal(w.Body.Bytes(), &created)
	compID := created.Data.ID
	if compID == "" {
		compID = created.ID
	}
	if compID == "" {
		t.Fatalf("no competition id in response: %s", w.Body.String())
	}

	// 未登录 → 401（/register 不在 publicPrefixes 白名单）
	anon := httptest.NewRecorder()
	app.ServeHTTP(anon, httptest.NewRequest(http.MethodPost, "/api/v1/competitions/"+compID+"/register",
		strings.NewReader(`{"team_name":"闪电队"}`)))
	if anon.Code != http.StatusUnauthorized {
		t.Fatalf("register without auth: %d, want 401", anon.Code)
	}

	// 手机号非法 → 400
	w = request(t, app, http.MethodPost, "/api/v1/competitions/"+compID+"/register",
		[]byte(`{"team_name":"闪电队","member_count":3,"name":"张三","phone":"12345"}`), domain.RoleIndividual)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("bad phone: %d, want 400", w.Code)
	}
	// 身份证非法 → 400
	w = request(t, app, http.MethodPost, "/api/v1/competitions/"+compID+"/register",
		[]byte(`{"team_name":"闪电队","member_count":3,"name":"张三","id_card":"123456"}`), domain.RoleIndividual)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("bad id_card: %d, want 400", w.Code)
	}
	// 完整报名 → 201 且扩展字段落库
	w = request(t, app, http.MethodPost, "/api/v1/competitions/"+compID+"/register",
		[]byte(`{"team_name":"闪电队","member_count":3,"contact_info":"13800000000","name":"张三","phone":"13800000000","id_card":"500101199001011234","photo_url":"/uploads/photo-a.jpg","id_card_image":"/uploads/idcard-a.jpg"}`),
		domain.RoleIndividual)
	if w.Code != http.StatusCreated {
		t.Fatalf("register: %d %s", w.Code, w.Body.String())
	}
	var reg struct {
		Data struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			Phone       string `json:"phone"`
			IDCard      string `json:"id_card"`
			PhotoURL    string `json:"photo_url"`
			IDCardImage string `json:"id_card_image"`
			UserID      string `json:"user_id"`
			TeamName    string `json:"team_name"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &reg)
	g := reg.Data
	if g.Name != "张三" || g.Phone != "13800000000" || g.IDCard != "500101199001011234" ||
		g.PhotoURL != "/uploads/photo-a.jpg" || g.IDCardImage != "/uploads/idcard-a.jpg" ||
		g.UserID != "user-1" || g.TeamName != "闪电队" {
		t.Fatalf("reg=%+v, want extended fields persisted for user-1", g)
	}
	if g.ID == "" {
		t.Fatalf("reg missing id: %s", w.Body.String())
	}
	// 旧客户端（只有 team_name/member_count/contact_info）仍可报名——向后兼容
	w = request(t, app, http.MethodPost, "/api/v1/competitions/"+compID+"/register",
		[]byte(`{"team_name":"飞鹰队","member_count":4,"contact_info":"13900000000"}`), domain.RoleIndividual)
	if w.Code != http.StatusCreated {
		t.Fatalf("legacy register: %d %s", w.Code, w.Body.String())
	}
}

// TestCompetitionRegPrivateUpload: 审查 HIGH 修复——身份证影像带 private=true 上传后
// URL 为 /uploads/private/，无 token 读取 401、携带 token 读取 200；普通上传仍公开前缀。
func TestCompetitionRegPrivateUpload(t *testing.T) {
	app := newBizServer(t)
	upload := func(private bool) string {
		t.Helper()
		var buf bytes.Buffer
		mw := multipart.NewWriter(&buf)
		h := textproto.MIMEHeader{}
		h.Set("Content-Disposition", `form-data; name="file"; filename="idcard.png"`)
		h.Set("Content-Type", "image/png")
		part, err := mw.CreatePart(h)
		if err != nil {
			t.Fatal(err)
		}
		// 魔数校验上线后必须上传真实 PNG 内容（声明 image/png 但传文本会被 400 拒绝）
		var pngBuf bytes.Buffer
		if err := png.Encode(&pngBuf, image.NewRGBA(image.Rect(0, 0, 1, 1))); err != nil {
			t.Fatal(err)
		}
		part.Write(pngBuf.Bytes())
		if private {
			mw.WriteField("private", "true")
		}
		mw.Close()
		r := httptest.NewRequest(http.MethodPost, "/api/v1/files/upload", &buf)
		r.Header.Set("Content-Type", mw.FormDataContentType())
		r.Header.Set("Authorization", auth(t, domain.RoleIndividual))
		w := httptest.NewRecorder()
		app.ServeHTTP(w, r)
		if w.Code != http.StatusCreated {
			t.Fatalf("upload private=%v: %d %s", private, w.Code, w.Body.String())
		}
		var out struct {
			Data struct {
				URL string `json:"url"`
			} `json:"data"`
		}
		json.Unmarshal(w.Body.Bytes(), &out)
		if out.Data.URL == "" {
			t.Fatalf("no url in response: %s", w.Body.String())
		}
		return out.Data.URL
	}
	privURL := upload(true)
	if !strings.HasPrefix(privURL, "/uploads/private/") {
		t.Fatalf("private upload url %q, want /uploads/private/ prefix", privURL)
	}
	pubURL := upload(false)
	if strings.HasPrefix(pubURL, "/uploads/private/") || !strings.HasPrefix(pubURL, "/uploads/") {
		t.Fatalf("public upload url %q, want /uploads/ prefix without private", pubURL)
	}

	// 读取鉴权：无 token → 401；非属主登录 → 403（归属校验）；属主/管理员 → 200。
	// 修复前只要登录即可读任意 private 文件（P1 越权读取）。
	// 测试环境上传目录（test_uploads/）与 FileServer 服务目录（uploads/）不同，
	// 先把真实上传的文件镜像到服务目录再验证读取鉴权。
	id := strings.TrimPrefix(privURL, "/uploads/private/")
	os.MkdirAll("uploads/private", 0755)
	imgData, err := os.ReadFile(filepath.Join("test_uploads/private", id))
	if err != nil {
		t.Fatalf("read uploaded file: %v", err)
	}
	dst := filepath.Join("uploads/private", id)
	if err := os.WriteFile(dst, imgData, 0644); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Remove(dst) })

	anon := httptest.NewRecorder()
	app.ServeHTTP(anon, httptest.NewRequest(http.MethodGet, privURL, nil))
	if anon.Code != http.StatusUnauthorized {
		t.Fatalf("GET private upload without auth: %d, want 401", anon.Code)
	}
	// 非属主（user-2）→ 403
	intruder := httptest.NewRequest(http.MethodGet, privURL, nil)
	intruder.Header.Set("Authorization", authAs(t, "user-2", domain.RoleIndividual))
	w2 := httptest.NewRecorder()
	app.ServeHTTP(w2, intruder)
	if w2.Code != http.StatusForbidden {
		t.Fatalf("GET private upload as other user: %d, want 403", w2.Code)
	}
	// 属主（user-1，上传者）→ 200
	owner := httptest.NewRequest(http.MethodGet, privURL, nil)
	owner.Header.Set("Authorization", auth(t, domain.RoleIndividual))
	w3 := httptest.NewRecorder()
	app.ServeHTTP(w3, owner)
	if w3.Code != http.StatusOK {
		t.Fatalf("GET private upload as owner: %d, want 200", w3.Code)
	}
	// 管理员 → 200
	adm := httptest.NewRequest(http.MethodGet, privURL, nil)
	adm.Header.Set("Authorization", authAs(t, "admin-1", domain.RolePlatformAdmin))
	w4 := httptest.NewRecorder()
	app.ServeHTTP(w4, adm)
	if w4.Code != http.StatusOK {
		t.Fatalf("GET private upload as admin: %d, want 200", w4.Code)
	}
}
