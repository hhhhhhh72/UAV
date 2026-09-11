package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"drone-platform/internal/domain"
)

// fakeMP4 造一个带 ftyp box 的最小 mp4：http.DetectContentType 靠 data[4:8]=="ftyp"
// 判定 video/mp4，所以这几十个字节就足以穿过魔数校验（不用于真实播放）。
func fakeMP4() []byte {
	b := make([]byte, 0, 64)
	b = append(b, 0x00, 0x00, 0x00, 0x18) // box size
	b = append(b, []byte("ftyp")...)
	b = append(b, []byte("isom")...)     // major brand
	b = append(b, 0x00, 0x00, 0x02, 0x00) // minor version
	b = append(b, []byte("isomiso2avc1mp41")...)
	return b
}

// 上传接口放行 mp4（案例视频用）：真 ftyp 头 → 200 且 content_type=video/mp4；
// 文本伪装成 .mp4 → 400（魔数说了算，不看扩展名）。
func TestUploadAcceptsMP4(t *testing.T) {
	app := newBizServer(t)
	tok := authAs(t, "user-1", domain.RoleIndividual)

	body, ct := multipartBody(t, "case.mp4", "video/mp4", fakeMP4())
	r := httptest.NewRequest(http.MethodPost, "/api/v1/files/upload", body)
	r.Header.Set("Content-Type", ct)
	r.Header.Set("Authorization", tok)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Code != http.StatusCreated {
		t.Fatalf("mp4 上传应 201，实际 %d %s", w.Code, w.Body.String())
	}
	var out struct {
		Data struct {
			URL         string `json:"url"`
			ContentType string `json:"content_type"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("parse upload: %v", err)
	}
	if !strings.HasPrefix(out.Data.URL, "/uploads/") {
		t.Fatalf("url 应指向 /uploads/，实际 %q", out.Data.URL)
	}
	if out.Data.ContentType != "video/mp4" {
		t.Fatalf("content_type 应为 video/mp4（按魔数落库），实际 %q", out.Data.ContentType)
	}

	// 注意：取回 /uploads/<id> 的类型头在这里验不了——测试服的 FileService 落在 test_uploads/，
	// 公开静态路由读的是 uploads/。类型判定逻辑由 TestLooksLikeISOBMFF 覆盖，类型头线上实测。

	// 后台案例表单实际调的是 /api/v1/upload（只回 url）——同一条通道也要放行 mp4
	bodyOK, ctOK := multipartBody(t, "case2.mp4", "video/mp4", fakeMP4())
	rOK := httptest.NewRequest(http.MethodPost, "/api/v1/upload", bodyOK)
	rOK.Header.Set("Content-Type", ctOK)
	rOK.Header.Set("Authorization", tok)
	wOK := httptest.NewRecorder()
	app.ServeHTTP(wOK, rOK)
	if wOK.Code != http.StatusOK {
		t.Fatalf("/api/v1/upload 也应放行 mp4，实际 %d %s", wOK.Code, wOK.Body.String())
	}
	if !strings.Contains(wOK.Body.String(), "/uploads/") {
		t.Fatalf("未返回上传地址：%s", wOK.Body.String())
	}

	// 伪造：文本改名 .mp4 → 仍 400，错误信息里应提到允许的类型
	body2, ct2 := multipartBody(t, "fake.mp4", "video/mp4", []byte("this is definitely not a video"))
	r2 := httptest.NewRequest(http.MethodPost, "/api/v1/files/upload", body2)
	r2.Header.Set("Content-Type", ct2)
	r2.Header.Set("Authorization", tok)
	w2 := httptest.NewRecorder()
	app.ServeHTTP(w2, r2)
	if w2.Code != http.StatusBadRequest {
		t.Fatalf("伪装 mp4 应 400，实际 %d %s", w2.Code, w2.Body.String())
	}
	if !strings.Contains(w2.Body.String(), "mp4") {
		t.Fatalf("拒绝信息应说明允许的类型（含 mp4）：%s", w2.Body.String())
	}
}

// 案例带视频：建 → 读 → 改，video_url 三个环节都不丢。
// 回归背景：案例模型原本只有 images，视频字段是 2026-09-11 补的（迁移 000102）。
// 视频封面不做独立字段——直接用 images[0] 当封面（产品确认不需要单独上传）。
func TestCaseVideoRoundTrip(t *testing.T) {
	app := newBizServer(t)
	adminTok := authAs(t, "admin-1", domain.RolePlatformAdmin)

	w := doRaw(app, http.MethodPost, "/api/v1/admin/cases",
		`{"title":"带视频的案例","category":"电力巡检","description":"说明","client_name":"甲方",`+
			`"result":"成效","images":["/uploads/cover.jpg"],`+
			`"video_url":"/uploads/case-video.mp4"}`, adminTok)
	if w.Code != http.StatusCreated {
		t.Fatalf("建案例: %d %s", w.Code, w.Body.String())
	}
	var created struct {
		Data struct {
			ID       string `json:"id"`
			VideoURL string `json:"video_url"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatalf("parse create: %v", err)
	}
	if created.Data.VideoURL != "/uploads/case-video.mp4" {
		t.Fatalf("建案例时 video_url 丢了：%+v", created.Data)
	}
	if strings.Contains(w.Body.String(), "video_poster_url") {
		t.Fatalf("视频封面字段已删除（改用封面图当 poster），不应再出现在响应里：%s", w.Body.String())
	}

	// 公开列表接口（小程序读的就是它）必须带出视频字段
	got := doRaw(app, http.MethodGet, "/api/v1/cases?page=1&page_size=50", "", "")
	if got.Code != http.StatusOK {
		t.Fatalf("公开详情: %d %s", got.Code, got.Body.String())
	}
	if !strings.Contains(got.Body.String(), "case-video.mp4") {
		t.Fatalf("公开详情未返回 video_url：%s", got.Body.String())
	}

	// 更新路径也要带上（换视频 / 清空视频）
	upd := doRaw(app, http.MethodPut, "/api/v1/admin/cases/"+created.Data.ID,
		`{"title":"带视频的案例","category":"电力巡检","description":"说明","client_name":"甲方",`+
			`"result":"成效","status":"published","images":[],"video_url":"/uploads/new.mp4"}`, adminTok)
	if upd.Code != http.StatusOK {
		t.Fatalf("更新案例: %d %s", upd.Code, upd.Body.String())
	}
	if !strings.Contains(upd.Body.String(), "new.mp4") {
		t.Fatalf("更新未生效：%s", upd.Body.String())
	}
}
