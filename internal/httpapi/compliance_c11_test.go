package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"drone-platform/internal/domain"
)

// TestComplianceStandardsCategoryFilter: 回归 C11——标准分类列 + 筛选。
// 旧实现 PG 侧 standard_docs 无 category 列，带 category 查询直接报错；
// 内存实现忽略过滤，分类 tab 永远全量/空列表。
func TestComplianceStandardsCategoryFilter(t *testing.T) {
	app := newBizServer(t)
	body := []byte(`{"title":"低空飞行安全标准","category":"团体标准","standard_no":"T/CDA-001","publisher":"协会","effective_date":"2026-07-01","status":"published","scope":"范围"}`)
	w := request(t, app, http.MethodPost, "/api/v1/admin/compliance-standards", body, domain.RolePlatformAdmin)
	if w.Code != http.StatusCreated {
		t.Fatalf("create standard: %d %s", w.Code, w.Body.String())
	}
	// 中文分类筛选命中
	w = request(t, app, http.MethodGet, "/api/v1/compliance-standards?category="+url.QueryEscape("团体标准"), nil, domain.RoleIndividual)
	if w.Code != http.StatusOK {
		t.Fatalf("GET standards by category: %d %s", w.Code, w.Body.String())
	}
	items, total := listEnvelope(t, w)
	if total != 1 || len(items) != 1 {
		t.Fatalf("category=团体标准: total=%d items=%d, want 1/1", total, len(items))
	}
	// 其他分类为空
	w = request(t, app, http.MethodGet, "/api/v1/compliance-standards?category="+url.QueryEscape("国家标准"), nil, domain.RoleIndividual)
	_, total = listEnvelope(t, w)
	if total != 0 {
		t.Fatalf("category=国家标准: total=%d, want 0", total)
	}
	// 不带分类返回全部
	w = request(t, app, http.MethodGet, "/api/v1/compliance-standards", nil, domain.RoleIndividual)
	_, total = listEnvelope(t, w)
	if total != 1 {
		t.Fatalf("all standards: total=%d, want 1", total)
	}
}

// TestIndustryReportsCategoryKeywordFilter: 回归 C11——行业报告 keyword + category 筛选
// （reports/list.vue tabs 传 whitepaper/research/analysis/other）。
func TestIndustryReportsCategoryKeywordFilter(t *testing.T) {
	app := newBizServer(t)
	create := func(title, category string) {
		t.Helper()
		body := []byte(`{"title":"` + title + `","period":"2026H1","category":"` + category + `","summary":"摘要","content":"全文","file_url":"","author":"协会"}`)
		w := request(t, app, http.MethodPost, "/api/v1/admin/industry-reports", body, domain.RolePlatformAdmin)
		if w.Code != http.StatusCreated {
			t.Fatalf("create report %s: %d %s", title, w.Code, w.Body.String())
		}
	}
	create("低空经济白皮书", "whitepaper")
	create("植保市场调研", "research")

	get := func(query string) (int, int) {
		t.Helper()
		w := request(t, app, http.MethodGet, "/api/v1/industry-reports"+query, nil, domain.RoleIndividual)
		if w.Code != http.StatusOK {
			t.Fatalf("GET reports%s: %d %s", query, w.Code, w.Body.String())
		}
		items, total := listEnvelope(t, w)
		return len(items), total
	}
	if n, total := get("?category=whitepaper"); total != 1 || n != 1 {
		t.Fatalf("category=whitepaper: total=%d n=%d, want 1/1", total, n)
	}
	if n, total := get("?category=research"); total != 1 || n != 1 {
		t.Fatalf("category=research: total=%d n=%d, want 1/1", total, n)
	}
	if n, total := get("?keyword=" + url.QueryEscape("白皮书") + "&category=whitepaper"); total != 1 || n != 1 {
		t.Fatalf("keyword+category: total=%d n=%d, want 1/1", total, n)
	}
	if n, total := get("?keyword=" + url.QueryEscape("不存在的关键词")); total != 0 || n != 0 {
		t.Fatalf("missing keyword: total=%d n=%d, want 0/0", total, n)
	}
	if n, total := get(""); total != 2 || n != 2 {
		t.Fatalf("all reports: total=%d n=%d, want 2/2", total, n)
	}
}

// TestIndustryResourceBooking: 回归 C11——POST /api/v1/industry-resources/{id}/book 端点。
// 旧后端无该路由（必 404），小程序资源详情页预约弹窗不可用。
func TestIndustryResourceBooking(t *testing.T) {
	app := newBizServer(t)
	// 管理端创建资源
	w := request(t, app, http.MethodPost, "/api/v1/admin/industry-resources",
		[]byte(`{"name":"M300 航测机","res_type":"drone","model":"M300","specs":"RTK","location":"重庆","booking_info":"提前一天预约","visibility_level":"public","price_fen":50000}`),
		domain.RolePlatformAdmin)
	if w.Code != http.StatusCreated {
		t.Fatalf("create resource: %d %s", w.Code, w.Body.String())
	}
	var created struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
		ID string `json:"id"`
	}
	json.Unmarshal(w.Body.Bytes(), &created)
	resID := created.Data.ID
	if resID == "" {
		resID = created.ID
	}
	if resID == "" {
		t.Fatalf("no resource id in response: %s", w.Body.String())
	}

	// 未登录 → 401（/book 不在 publicPrefixes 白名单）
	anon := httptest.NewRecorder()
	app.ServeHTTP(anon, httptest.NewRequest(http.MethodPost, "/api/v1/industry-resources/"+resID+"/book",
		strings.NewReader(`{"date":"2026-08-20","purpose":"测试","contact_name":"张三","contact_phone":"13800000000"}`)))
	if anon.Code != http.StatusUnauthorized {
		t.Fatalf("book without auth: %d, want 401", anon.Code)
	}

	valid := `{"date":"2026-08-20","purpose":"航测项目","contact_name":"张三","contact_phone":"13800000000"}`
	// 手机号非法 → 400
	w = request(t, app, http.MethodPost, "/api/v1/industry-resources/"+resID+"/book",
		[]byte(`{"date":"2026-08-20","purpose":"测试","contact_name":"张三","contact_phone":"123"}`), domain.RoleIndividual)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("bad phone: %d, want 400", w.Code)
	}
	// 日期格式非法 → 400
	w = request(t, app, http.MethodPost, "/api/v1/industry-resources/"+resID+"/book",
		[]byte(`{"date":"2026/08/20","purpose":"测试","contact_name":"张三","contact_phone":"13800000000"}`), domain.RoleIndividual)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("bad date: %d, want 400", w.Code)
	}
	// 资源不存在 → 404
	w = request(t, app, http.MethodPost, "/api/v1/industry-resources/res-nope/book",
		[]byte(valid), domain.RoleIndividual)
	if w.Code != http.StatusNotFound {
		t.Fatalf("missing resource: %d, want 404", w.Code)
	}
	// 正常预约 → 201 + pending
	w = request(t, app, http.MethodPost, "/api/v1/industry-resources/"+resID+"/book",
		[]byte(valid), domain.RoleIndividual)
	if w.Code != http.StatusCreated {
		t.Fatalf("book: %d %s", w.Code, w.Body.String())
	}
	var booked struct {
		Data struct {
			ID           string `json:"id"`
			ResourceID   string `json:"resource_id"`
			UserID       string `json:"user_id"`
			BookingDate  string `json:"date"`
			Status       string `json:"status"`
			ContactPhone string `json:"contact_phone"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &booked)
	b := booked.Data
	if b.Status != "pending" || b.ResourceID != resID || b.BookingDate != "2026-08-20" || b.UserID != "user-1" {
		t.Fatalf("booking=%+v, want pending/%s/2026-08-20/user-1", b, resID)
	}
	if b.ID == "" {
		t.Fatalf("booking missing id: %s", w.Body.String())
	}

	// 可见级别校验（审查 LOW 修复）：partner 级资源对 individual 不可见 → 预约 403
	w = request(t, app, http.MethodPost, "/api/v1/admin/industry-resources",
		[]byte(`{"name":"内部高端机","res_type":"drone","model":"M600","specs":"RTK","location":"重庆","booking_info":"仅合作方","visibility_level":"partner","price_fen":90000}`),
		domain.RolePlatformAdmin)
	if w.Code != http.StatusCreated {
		t.Fatalf("create partner resource: %d %s", w.Code, w.Body.String())
	}
	var partner struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
		ID string `json:"id"`
	}
	json.Unmarshal(w.Body.Bytes(), &partner)
	partnerID := partner.Data.ID
	if partnerID == "" {
		partnerID = partner.ID
	}
	if partnerID == resID {
		t.Fatalf("resource ID collision: two creates returned %s (UnixNano tick)", partnerID)
	}
	w = request(t, app, http.MethodPost, "/api/v1/industry-resources/"+partnerID+"/book",
		[]byte(valid), domain.RoleIndividual)
	if w.Code != http.StatusForbidden {
		t.Fatalf("book partner-tier resource as individual: %d, want 403", w.Code)
	}
}
