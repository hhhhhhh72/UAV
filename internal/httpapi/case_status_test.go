package httpapi_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"drone-platform/internal/domain"
)

// 案例状态的真实语义（回归三处缺陷）：
//  1. 新建时传的状态必须被尊重（此前 Create 写死 published，传 pending/archived 被静默丢弃）；
//  2. 公开列表只出 published —— 已下架/待审核的案例不能再对外曝光（曾经"下架"点了没用）；
//  3. 管理端看得到全部状态，并支持按状态筛选；非法状态一律 400。
func TestCaseStatusSemantics(t *testing.T) {
	app := newBizServer(t)
	adminTok := authAs(t, "admin-1", domain.RolePlatformAdmin)

	// 1) 新建：published（显式传）与 archived（下架态）
	pubID := createCaseWithStatus(t, app, adminTok, "公开案例", domain.CaseStatusPublished)
	archID := createCaseWithStatus(t, app, adminTok, "下架案例", domain.CaseStatusArchived)

	// 2) 公开列表只出 published
	pub := doRaw(app, http.MethodGet, "/api/v1/cases?page=1&page_size=50", "", "")
	if pub.Code != http.StatusOK {
		t.Fatalf("公开列表: %d %s", pub.Code, pub.Body.String())
	}
	body := pub.Body.String()
	if !strings.Contains(body, pubID) {
		t.Fatalf("已发布案例应从公开列表可见：%s", body)
	}
	if strings.Contains(body, archID) {
		t.Fatalf("已下架案例不应出现在公开列表：%s", body)
	}

	// 3) 管理端：全部可见 + 可按状态筛选
	all := doRaw(app, http.MethodGet, "/api/v1/admin/cases", "", adminTok)
	if all.Code != http.StatusOK {
		t.Fatalf("管理端列表: %d %s", all.Code, all.Body.String())
	}
	if !strings.Contains(all.Body.String(), archID) {
		t.Fatalf("管理端应能看到已下架案例：%s", all.Body.String())
	}
	onlyArch := doRaw(app, http.MethodGet, "/api/v1/admin/cases?status=archived", "", adminTok)
	if !strings.Contains(onlyArch.Body.String(), archID) || strings.Contains(onlyArch.Body.String(), pubID) {
		t.Fatalf("按 status=archived 筛选结果不对：%s", onlyArch.Body.String())
	}

	// 4) 下架 → 公开列表立刻看不到；重新发布 → 又能看到
	if w := doRaw(app, http.MethodPut, "/api/v1/admin/cases/"+pubID,
		`{"title":"公开案例","category":"测试","status":"archived"}`, adminTok); w.Code != http.StatusOK {
		t.Fatalf("下架: %d %s", w.Code, w.Body.String())
	}
	if now := doRaw(app, http.MethodGet, "/api/v1/cases?page=1&page_size=50", "", ""); strings.Contains(now.Body.String(), pubID) {
		t.Fatalf("下架后公开列表仍可见（下架无效）：%s", now.Body.String())
	}
	if w := doRaw(app, http.MethodPut, "/api/v1/admin/cases/"+pubID,
		`{"title":"公开案例","category":"测试","status":"published"}`, adminTok); w.Code != http.StatusOK {
		t.Fatalf("重新发布: %d %s", w.Code, w.Body.String())
	}
	if back := doRaw(app, http.MethodGet, "/api/v1/cases?page=1&page_size=50", "", ""); !strings.Contains(back.Body.String(), pubID) {
		t.Fatalf("重新发布后公开列表应可见：%s", back.Body.String())
	}

	// 5) 非法状态 → 400（创建与更新都拦）
	if w := doRaw(app, http.MethodPost, "/api/v1/admin/cases",
		`{"title":"非法状态","category":"测试","status":"whatever"}`, adminTok); w.Code != http.StatusBadRequest {
		t.Fatalf("非法状态建案例应 400，实际 %d %s", w.Code, w.Body.String())
	}
	if w := doRaw(app, http.MethodPut, "/api/v1/admin/cases/"+pubID,
		`{"title":"非法状态","category":"测试","status":"whatever"}`, adminTok); w.Code != http.StatusBadRequest {
		t.Fatalf("非法状态更新应 400，实际 %d %s", w.Code, w.Body.String())
	}
}

func createCaseWithStatus(t *testing.T, app http.Handler, tok, title, status string) string {
	t.Helper()
	w := doRaw(app, http.MethodPost, "/api/v1/admin/cases",
		`{"title":"`+title+`","category":"测试","status":"`+status+`"}`, tok)
	if w.Code != http.StatusCreated {
		t.Fatalf("建案例(%s): %d %s", status, w.Code, w.Body.String())
	}
	var resp struct {
		Data struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if resp.Data.Status != status {
		t.Fatalf("建案例时状态被丢弃：传 %s，落库 %s", status, resp.Data.Status)
	}
	return resp.Data.ID
}
