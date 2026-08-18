package httpapi_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"drone-platform/internal/domain"
)

// 专家可见性回归：公开列表/详情仅显示 published 专家；
// pending/archived 专家仅管理员（admin 路由/带管理员 token）可见。
// 修复前公开列表不过滤 status，未审核专家直接公开。
func TestExpertVisibilityPendingNotPublic(t *testing.T) {
	app := newBizServer(t)

	// 管理员创建两个专家：pending（待审核）+ published（已发布）
	create := func(status string) string {
		w := requestAs(t, app, http.MethodPost, "/api/v1/admin/experts",
			[]byte(`{"name":"专家`+status+`","title":"教授","org":"重大","field":"巡检","status":"`+status+`"}`),
			"admin-1", domain.RolePlatformAdmin)
		if w.Code != http.StatusCreated {
			t.Fatalf("create expert(%s): %d %s", status, w.Code, w.Body.String())
		}
		var envelope struct {
			Data struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
			t.Fatalf("parse create(%s): %v body=%s", status, err, w.Body.String())
		}
		return envelope.Data.ID
	}
	pendingID := create("pending")
	publishedID := create("published")

	// 1) 匿名公开列表：不含 pending，含 published
	w := request(t, app, http.MethodGet, "/api/v1/experts", nil, "")
	if w.Code != http.StatusOK {
		t.Fatalf("anon list: %d", w.Code)
	}
	if strings.Contains(w.Body.String(), "专家pending") {
		t.Fatal("public list must not expose pending expert")
	}
	if !strings.Contains(w.Body.String(), "专家published") {
		t.Fatal("public list should contain published expert")
	}

	// 2) 普通登录用户列表：同样不含 pending
	w = request(t, app, http.MethodGet, "/api/v1/experts", nil, domain.RoleIndividual)
	if strings.Contains(w.Body.String(), "专家pending") {
		t.Fatal("user list must not expose pending expert")
	}

	// 3) 匿名详情：pending → 404（不暴露存在性），published → 200
	w = request(t, app, http.MethodGet, "/api/v1/experts/"+pendingID, nil, "")
	if w.Code != http.StatusNotFound {
		t.Fatalf("anon get pending expert: want 404, got %d", w.Code)
	}
	w = request(t, app, http.MethodGet, "/api/v1/experts/"+publishedID, nil, "")
	if w.Code != http.StatusOK {
		t.Fatalf("anon get published expert: want 200, got %d", w.Code)
	}

	// 4) 管理员可见全部（admin 列表 + 带管理员 token 的公开详情）
	w = requestAs(t, app, http.MethodGet, "/api/v1/admin/experts", nil, "admin-1", domain.RolePlatformAdmin)
	if !strings.Contains(w.Body.String(), "专家pending") {
		t.Fatal("admin list should contain pending expert")
	}
	w = requestAs(t, app, http.MethodGet, "/api/v1/experts/"+pendingID, nil, "admin-1", domain.RolePlatformAdmin)
	if w.Code != http.StatusOK {
		t.Fatalf("admin get pending expert: want 200, got %d %s", w.Code, w.Body.String())
	}
}
