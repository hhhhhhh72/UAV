package httpapi_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"drone-platform/internal/domain"
)

// 越权防护回归（P0）：用户侧 PUT/DELETE 必须校验资源属主。
// 修复前 handler 丢弃 actor、service 无校验，任意登录用户可改删他人
// Achievement / RDChallenge / Portfolio。

// 成果（Achievement）：他人改删 → 403，属主改删 → 200。
func TestOwnershipAchievementForbiddenForOthers(t *testing.T) {
	app := newBizServer(t)

	// A 创建成果
	w := requestAs(t, app, http.MethodPost, "/api/v1/achievements",
		[]byte(`{"title":"A的避障算法","achieve_type":"patent","description":"d","field":"巡检"}`),
		"owner-a", domain.RoleIndividual)
	if w.Code != http.StatusCreated {
		t.Fatalf("create achievement: %d %s", w.Code, w.Body.String())
	}
	var created struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatalf("parse create: %v", err)
	}
	id := created.Data.ID

	// B 越权修改 → 403
	w = requestAs(t, app, http.MethodPut, "/api/v1/achievements/"+id,
		[]byte(`{"title":"B篡改标题"}`), "intruder-b", domain.RoleIndividual)
	if w.Code != http.StatusForbidden {
		t.Fatalf("intruder PUT achievement: want 403, got %d %s", w.Code, w.Body.String())
	}

	// B 越权删除 → 403
	w = requestAs(t, app, http.MethodDelete, "/api/v1/achievements/"+id, nil, "intruder-b", domain.RoleIndividual)
	if w.Code != http.StatusForbidden {
		t.Fatalf("intruder DELETE achievement: want 403, got %d %s", w.Code, w.Body.String())
	}

	// 资源仍存在（未被越权删除）
	w = requestAs(t, app, http.MethodGet, "/api/v1/achievements/"+id, nil, "owner-a", domain.RoleIndividual)
	if w.Code != http.StatusOK {
		t.Fatalf("achievement should still exist: %d %s", w.Code, w.Body.String())
	}

	// 属主修改 → 200
	w = requestAs(t, app, http.MethodPut, "/api/v1/achievements/"+id,
		[]byte(`{"title":"A的避障算法v2"}`), "owner-a", domain.RoleIndividual)
	if w.Code != http.StatusOK {
		t.Fatalf("owner PUT achievement: want 200, got %d %s", w.Code, w.Body.String())
	}

	// 属主删除 → 200
	w = requestAs(t, app, http.MethodDelete, "/api/v1/achievements/"+id, nil, "owner-a", domain.RoleIndividual)
	if w.Code != http.StatusOK {
		t.Fatalf("owner DELETE achievement: want 200, got %d %s", w.Code, w.Body.String())
	}
}

// 品牌（Portfolio）：他人修改 → 403，属主修改 → 200。
func TestOwnershipPortfolioForbiddenForOthers(t *testing.T) {
	app := newBizServer(t)

	w := requestAs(t, app, http.MethodPost, "/api/v1/portfolios",
		[]byte(`{"name":"A公司品牌","logo_url":"l.png","description":"d","contact_info":"138"}`),
		"owner-ent", domain.RoleEnterprise)
	if w.Code != http.StatusCreated {
		t.Fatalf("create portfolio: %d %s", w.Code, w.Body.String())
	}
	var created struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatalf("parse create: %v", err)
	}
	id := created.Data.ID

	w = requestAs(t, app, http.MethodPut, "/api/v1/portfolios/"+id,
		[]byte(`{"name":"B篡改品牌"}`), "intruder-ent", domain.RoleEnterprise)
	if w.Code != http.StatusForbidden {
		t.Fatalf("intruder PUT portfolio: want 403, got %d %s", w.Code, w.Body.String())
	}

	w = requestAs(t, app, http.MethodPut, "/api/v1/portfolios/"+id,
		[]byte(`{"name":"A公司品牌v2"}`), "owner-ent", domain.RoleEnterprise)
	if w.Code != http.StatusOK {
		t.Fatalf("owner PUT portfolio: want 200, got %d %s", w.Code, w.Body.String())
	}
}

// 研发难题（RDChallenge）：他人修改 → 403；平台管理员走 admin 路由删除 → 200。
func TestOwnershipRDChallengeForbiddenForOthers(t *testing.T) {
	app := newBizServer(t)

	w := requestAs(t, app, http.MethodPost, "/api/v1/rd-challenges",
		[]byte(`{"title":"A的电池难题","field":"电池","description":"d","deadline":"2026-12-31","budget_fen":500000}`),
		"owner-ent", domain.RoleEnterprise)
	if w.Code != http.StatusCreated {
		t.Fatalf("create rd-challenge: %d %s", w.Code, w.Body.String())
	}
	var created struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatalf("parse create: %v", err)
	}
	id := created.Data.ID

	// 他人修改 → 403
	w = requestAs(t, app, http.MethodPut, "/api/v1/rd-challenges/"+id,
		[]byte(`{"title":"B篡改难题","deadline":"2026-12-31"}`), "intruder-ent", domain.RoleEnterprise)
	if w.Code != http.StatusForbidden {
		t.Fatalf("intruder PUT rd-challenge: want 403, got %d %s", w.Code, w.Body.String())
	}

	// 属主修改 → 200
	w = requestAs(t, app, http.MethodPut, "/api/v1/rd-challenges/"+id,
		[]byte(`{"title":"A的电池难题v2","deadline":"2026-12-31"}`), "owner-ent", domain.RoleEnterprise)
	if w.Code != http.StatusOK {
		t.Fatalf("owner PUT rd-challenge: want 200, got %d %s", w.Code, w.Body.String())
	}

	// 平台管理员经 admin 路由删除 → 200（管理员放行）
	w = requestAs(t, app, http.MethodDelete, "/api/v1/admin/rd-challenges/"+id, nil, "admin-1", domain.RolePlatformAdmin)
	if w.Code != http.StatusOK {
		t.Fatalf("admin DELETE rd-challenge: want 200, got %d %s", w.Code, w.Body.String())
	}
}
