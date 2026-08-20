package httpapi_test

// 帖子审核闭环回归（修复）：CreatePost 默认 pending 后管理端无列表入口，
// API 发布的帖子永远无法上架。现新增 GET /api/v1/admin/posts 全量列表
// （含待审核 pending），配合既有 POST /api/v1/posts/{id}/publish 完成审核闭环。

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"drone-platform/internal/domain"
)

func TestAdminPostsListAuditClosedLoop(t *testing.T) {
	app := newBizServer(t)
	authorTok := authAs(t, "user-1", domain.RoleIndividual)
	adminTok := authAs(t, "admin-1", domain.RolePlatformAdmin)

	// 用户发两帖（默认 pending，不进公开列表）
	for _, title := range []string{"待审帖A", "待审帖B"} {
		w := doRaw(app, http.MethodPost, "/api/v1/posts",
			`{"title":"`+title+`","content":"内容"}`, authorTok)
		checkCode(t, http.MethodPost, "/api/v1/posts", w, http.StatusCreated)
	}

	// 非管理员访问管理端列表 → 403（adminGate）
	w := doRaw(app, http.MethodGet, "/api/v1/admin/posts", "", authorTok)
	checkCode(t, http.MethodGet, "/api/v1/admin/posts (non-admin)", w, http.StatusForbidden)

	// 管理员全量列表 → 200，含 2 条待审帖
	w = doRaw(app, http.MethodGet, "/api/v1/admin/posts", "", adminTok)
	checkCode(t, http.MethodGet, "/api/v1/admin/posts", w, http.StatusOK)
	if !strings.Contains(w.Body.String(), "待审帖A") || !strings.Contains(w.Body.String(), "待审帖B") {
		t.Fatalf("admin posts list must contain pending posts: %s", w.Body.String())
	}
	if !strings.Contains(w.Body.String(), `"total":2`) {
		t.Fatalf("admin posts list total should be 2: %s", w.Body.String())
	}
	// 状态筛选：pending 过滤
	w = doRaw(app, http.MethodGet, "/api/v1/admin/posts?status=pending", "", adminTok)
	checkCode(t, http.MethodGet, "/api/v1/admin/posts?status=pending", w, http.StatusOK)
	if !strings.Contains(w.Body.String(), `"total":2`) {
		t.Fatalf("pending filter should show 2: %s", w.Body.String())
	}

	// 公开列表（审核前）：不含待审帖
	w = doRaw(app, http.MethodGet, "/api/v1/posts", "", authorTok)
	checkCode(t, http.MethodGet, "/api/v1/posts", w, http.StatusOK)
	if strings.Contains(w.Body.String(), "待审帖A") {
		t.Fatalf("public posts must not show pending post before publish: %s", w.Body.String())
	}

	// 管理员从列表拿到 ID 后 publish → 上架（审核闭环）
	listResp := doRaw(app, http.MethodGet, "/api/v1/admin/posts?status=pending", "", adminTok)
	postID := firstIDFromList(t, listResp)
	if postID == "" {
		t.Fatalf("admin posts list should return ids, got: %s", listResp.Body.String())
	}
	w = doRaw(app, http.MethodPost, "/api/v1/posts/"+postID+"/publish", "", adminTok)
	checkCode(t, http.MethodPost, "/api/v1/posts/{id}/publish", w, http.StatusOK)

	// 公开列表现在可见该帖；管理端状态筛选 published 命中 1 条
	w = doRaw(app, http.MethodGet, "/api/v1/posts", "", authorTok)
	checkCode(t, http.MethodGet, "/api/v1/posts", w, http.StatusOK)
	if !strings.Contains(w.Body.String(), postID) {
		t.Fatalf("public posts should contain published post %s: %s", postID, w.Body.String())
	}
	w = doRaw(app, http.MethodGet, "/api/v1/admin/posts?status=published", "", adminTok)
	checkCode(t, http.MethodGet, "/api/v1/admin/posts?status=published", w, http.StatusOK)
	if !strings.Contains(w.Body.String(), `"total":1`) {
		t.Fatalf("published filter should show 1: %s", w.Body.String())
	}
}

// firstIDFromList 从分页响应 {data:[{id:...},...]} 中取第一条 id（管理端列表风格）。
func firstIDFromList(t *testing.T, w *httptest.ResponseRecorder) string {
	t.Helper()
	var out struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("parse list: %v (body=%s)", err, w.Body.String())
	}
	if len(out.Data) == 0 {
		return ""
	}
	return out.Data[0].ID
}
