package httpapi_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"drone-platform/internal/domain"
)

// 回归系统性双重分页修复：约 18 个列表接口的 handler 曾把「已按 offset 分页的
// 单页结果」又交给 paginatedRespond 二次切片，导致 page≥2 恒为空（前端触底加载翻不了页）。
// 修复模式：handler 全量拉取（offset=0/limit 大值），分页只由 paginatedRespond 做一次。

// createVia POST/PUT 并返回信封 data.id。
func createVia(t *testing.T, app http.Handler, method, path string, body []byte, userID string, role domain.Role) string {
	t.Helper()
	w := requestAs(t, app, method, path, body, userID, role)
	if w.Code != http.StatusCreated && w.Code != http.StatusOK {
		t.Fatalf("%s %s: %d %s", method, path, w.Code, w.Body.String())
	}
	var out struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("parse %s response: %v (%s)", path, err, w.Body.String())
	}
	return out.Data.ID
}

// assertPagination 校验 page=1 与 page=2 均有正确数据（旧实现 page≥2 恒空）。
func assertPagination(t *testing.T, app http.Handler, path string, userID string, role domain.Role, wantTotal int) {
	t.Helper()
	checks := []struct {
		query     string
		wantItems int
	}{
		{"page=1&page_size=2", 2},
		{"page=2&page_size=2", 2}, // 修复前恒为 0
	}
	for _, c := range checks {
		sep := "?"
		if containsRune(path, '?') {
			sep = "&"
		}
		url := path + sep + c.query
		w := requestAs(t, app, http.MethodGet, url, nil, userID, role)
		if w.Code != http.StatusOK {
			t.Fatalf("%s: %d %s", url, w.Code, w.Body.String())
		}
		var out struct {
			Data  []json.RawMessage `json:"data"`
			Total int               `json:"total"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
			t.Fatalf("parse %s: %v (%s)", url, err, w.Body.String())
		}
		if len(out.Data) != c.wantItems {
			t.Fatalf("%s: want %d items, got %d (%s)", url, c.wantItems, len(out.Data), w.Body.String())
		}
		if out.Total != wantTotal {
			t.Fatalf("%s: want total %d, got %d", url, wantTotal, out.Total)
		}
	}
}

func containsRune(s string, r rune) bool {
	for _, c := range s {
		if c == r {
			return true
		}
	}
	return false
}

func TestSystematicDoublePaginationFixed(t *testing.T) {
	app := newBizServer(t)
	const n = 5
	admin := "admin-1"
	adminRole := domain.RolePlatformAdmin
	ent := "enterprise-1"
	entRole := domain.RoleEnterprise

	// 1. cases：管理员建 5 条 → 公开列表翻页
	for i := 0; i < n; i++ {
		createVia(t, app, http.MethodPost, "/api/v1/admin/cases",
			[]byte(fmt.Sprintf(`{"title":"案例%d","category":"农业","description":"d","client_name":"c","result":"r"}`, i)),
			admin, adminRole)
	}
	assertPagination(t, app, "/api/v1/cases", "user-1", domain.RoleIndividual, n)

	// 2. compliance docs
	for i := 0; i < n; i++ {
		createVia(t, app, http.MethodPost, "/api/v1/admin/compliance-docs",
			[]byte(fmt.Sprintf(`{"title":"文档%d","category":"政策","publisher":"p","publish_date":"2024-01-01","status":"published","summary":"s","file_url":"","tags":[]}`, i)),
			admin, adminRole)
	}
	assertPagination(t, app, "/api/v1/compliance-docs", "user-1", domain.RoleIndividual, n)

	// 3. compliance standards
	for i := 0; i < n; i++ {
		createVia(t, app, http.MethodPost, "/api/v1/admin/compliance-standards",
			[]byte(fmt.Sprintf(`{"title":"标准%d","category":"团体标准","standard_no":"T/00%d","publisher":"p","effective_date":"2024-01-01","status":"published","scope":"s","file_url":""}`, i, i)),
			admin, adminRole)
	}
	assertPagination(t, app, "/api/v1/compliance-standards", "user-1", domain.RoleIndividual, n)

	// 4. emergency resources
	for i := 0; i < n; i++ {
		createVia(t, app, http.MethodPost, "/api/v1/admin/emergency-resources",
			[]byte(fmt.Sprintf(`{"name":"应急资源%d","res_type":"drone","specs":"s","location":"l","contact_info":"c","quantity":1}`, i)),
			admin, adminRole)
	}
	assertPagination(t, app, "/api/v1/emergency-resources", "user-1", domain.RoleIndividual, n)

	// 5. emergency dispatches（含 status 筛选在分页前的回归）
	resID := createVia(t, app, http.MethodPost, "/api/v1/admin/emergency-resources",
		[]byte(`{"name":"调度资源","res_type":"drone","specs":"s","location":"l","contact_info":"c","quantity":1}`), admin, adminRole)
	for i := 0; i < n; i++ {
		createVia(t, app, http.MethodPost, "/api/v1/admin/emergency-dispatches",
			[]byte(fmt.Sprintf(`{"resource_id":%q,"event_desc":"事件%d","location":"l","commander":"c","result":"r","status":"dispatched"}`, resID, i)),
			admin, adminRole)
	}
	assertPagination(t, app, "/api/v1/emergency-dispatches", "user-1", domain.RoleIndividual, n)
	// status 筛选必须在分页前：全部 5 条均为 dispatched → status 过滤后 total 仍为 5，page=2 有数据
	assertPagination(t, app, "/api/v1/emergency-dispatches?status=dispatched", "user-1", domain.RoleIndividual, n)

	// 6. rescue cases
	for i := 0; i < n; i++ {
		createVia(t, app, http.MethodPost, "/api/v1/admin/rescue-cases",
			[]byte(fmt.Sprintf(`{"title":"救援案例%d","event_type":"山火","location":"l","summary":"s","result":"r","date":"2024-01-01"}`, i)),
			admin, adminRole)
	}
	assertPagination(t, app, "/api/v1/rescue-cases", "user-1", domain.RoleIndividual, n)

	// 7. articles（公开列表仅 published：创建后显式发布，草稿不进公开列表）
	for i := 0; i < n; i++ {
		id := createVia(t, app, http.MethodPost, "/api/v1/articles",
			[]byte(fmt.Sprintf(`{"title":"资讯%d","content":"c","category":"policy","source":"s"}`, i)),
			admin, adminRole)
		w := requestAs(t, app, http.MethodPost, "/api/v1/articles/"+id+"/publish", nil, admin, adminRole)
		if w.Code != http.StatusOK {
			t.Fatalf("publish article %s: %d %s", id, w.Code, w.Body.String())
		}
	}
	assertPagination(t, app, "/api/v1/articles", "user-1", domain.RoleIndividual, n)

	// 8. jobs：企业发布 + 发布上架
	for i := 0; i < n; i++ {
		id := createVia(t, app, http.MethodPost, "/api/v1/jobs",
			[]byte(fmt.Sprintf(`{"title":"岗位%d","description":"d","location":"l","job_type":"full_time","salary_fen":10000}`, i)),
			ent, entRole)
		w := requestAs(t, app, http.MethodPost, "/api/v1/jobs/"+id+"/publish", nil, ent, entRole)
		if w.Code != http.StatusOK {
			t.Fatalf("publish job %s: %d %s", id, w.Code, w.Body.String())
		}
	}
	assertPagination(t, app, "/api/v1/jobs", "user-1", domain.RoleIndividual, n)

	// 9. portfolios：企业发布（draft）→ 更新为 published
	for i := 0; i < n; i++ {
		id := createVia(t, app, http.MethodPost, "/api/v1/portfolios",
			[]byte(fmt.Sprintf(`{"name":"品牌%d","description":"d","contact_info":"c"}`, i)),
			ent, entRole)
		createVia(t, app, http.MethodPut, "/api/v1/portfolios/"+id,
			[]byte(fmt.Sprintf(`{"name":"品牌%d","description":"d","contact_info":"c","status":"published"}`, i)),
			ent, entRole)
	}
	assertPagination(t, app, "/api/v1/portfolios", "user-1", domain.RoleIndividual, n)

	// 10. association members
	for i := 0; i < n; i++ {
		createVia(t, app, http.MethodPost, "/api/v1/admin/association-members",
			[]byte(fmt.Sprintf(`{"user_id":"member-%d","enterprise_id":"ent-%d","role":"member"}`, i, i)),
			admin, adminRole)
	}
	assertPagination(t, app, "/api/v1/association-members", "user-1", domain.RoleIndividual, n)

	// 11. employment-requests：企业创建 → 本人列表翻页（非管理员走 ListByEnterprise）
	for i := 0; i < n; i++ {
		createVia(t, app, http.MethodPost, "/api/v1/employment-requests",
			[]byte(fmt.Sprintf(`{"position":"飞手%d","headcount":2}`, i)),
			ent, entRole)
	}
	assertPagination(t, app, "/api/v1/employment-requests", ent, entRole, n)

	// 12. contracts：企业创建 → 本人列表翻页
	for i := 0; i < n; i++ {
		createVia(t, app, http.MethodPost, "/api/v1/contracts", []byte(`{}`), ent, entRole)
	}
	assertPagination(t, app, "/api/v1/contracts", ent, entRole, n)

	// 13. certified pilots：个人注册（不同用户）→ 管理员审批 → 公开名录翻页
	for i := 0; i < n; i++ {
		// 新规则：申请飞手需至少一张已通过且未过期的证书
		cid := createVia(t, app, http.MethodPost, "/api/v1/certificates",
			[]byte(fmt.Sprintf(`{"cert_type":"caac","cert_number":"PAG-CERT-%d","level":"III","issuer_org":"民航局","expire_date":"2028-01-01T00:00:00Z"}`, i)),
			fmt.Sprintf("worker-%d", i), domain.RoleIndividual)
		w := requestAs(t, app, http.MethodPost, "/api/v1/admin/certificates/"+cid+"/approve", nil, admin, adminRole)
		if w.Code != http.StatusOK {
			t.Fatalf("approve cert %s: %d %s", cid, w.Code, w.Body.String())
		}
		id := createVia(t, app, http.MethodPost, "/api/v1/certified-pilots",
			[]byte(fmt.Sprintf(`{"real_name":"飞手%d","id_card":"51010719900101%04d"}`, i, i)),
			fmt.Sprintf("worker-%d", i), domain.RoleIndividual)
		w = requestAs(t, app, http.MethodPost, "/api/v1/admin/certified-pilots/"+id+"/approve", nil, admin, adminRole)
		if w.Code != http.StatusOK {
			t.Fatalf("approve pilot %s: %d %s", id, w.Code, w.Body.String())
		}
	}
	assertPagination(t, app, "/api/v1/certified-pilots", "user-1", domain.RoleIndividual, n)
}
