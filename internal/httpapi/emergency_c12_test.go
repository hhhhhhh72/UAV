package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"drone-platform/internal/domain"
)

// listEnvelope 解析分页响应包 {data:[...], total:N}
func listEnvelope(t *testing.T, w *httptest.ResponseRecorder) ([]json.RawMessage, int) {
	t.Helper()
	var env struct {
		Data  []json.RawMessage `json:"data"`
		Total int               `json:"total"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &env); err != nil {
		t.Fatalf("parse envelope: %v; body=%s", err, w.Body.String())
	}
	return env.Data, env.Total
}

// TestRescueCaseC12Filters: 回归 C12——救援案例经 HTTP 全链路：
// 英文 event_type 创建归一为中文，中文/英文筛选与 q 搜索均生效。
func TestRescueCaseC12Filters(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodPost, "/api/v1/admin/rescue-cases",
		[]byte(`{"title":"南山森林火情","event_type":"mountain_fire","location":"重庆南山","drone_model":"M300热成像","summary":"火线侦察"}`),
		domain.RolePlatformAdmin)
	if w.Code != http.StatusCreated {
		t.Fatalf("create rescue case: %d %s", w.Code, w.Body.String())
	}

	// 小程序契约：中文 event_type + q
	w = request(t, app, http.MethodGet, "/api/v1/rescue-cases?event_type=山火&q=热成像", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK {
		t.Fatalf("list with 中文筛选: %d %s", w.Code, w.Body.String())
	}
	data, total := listEnvelope(t, w)
	if total != 1 || len(data) != 1 {
		t.Fatalf("中文筛选: total=%d len=%d, want 1/1; body=%s", total, len(data), w.Body.String())
	}

	// 英文键等价
	w = request(t, app, http.MethodGet, "/api/v1/rescue-cases?event_type=mountain_fire", nil, domain.RoleIndividual)
	data, total = listEnvelope(t, w)
	if total != 1 || len(data) != 1 {
		t.Fatalf("英文筛选: total=%d len=%d, want 1/1", total, len(data))
	}

	// q 不命中时为空
	w = request(t, app, http.MethodGet, "/api/v1/rescue-cases?q=不存在的关键词", nil, domain.RoleIndividual)
	data, total = listEnvelope(t, w)
	if total != 0 || len(data) != 0 {
		t.Fatalf("无命中搜索: total=%d len=%d, want 0/0", total, len(data))
	}
}

// TestEmergencyResourceC12Filters: 回归 C12——应急资源 res_type/q 参数经 HTTP 全链路生效，
// 中文 res_type 创建归一为英文。
func TestEmergencyResourceC12Filters(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodPost, "/api/v1/admin/emergency-resources",
		[]byte(`{"name":"应急无人机01","res_type":"无人机","specs":"M300RTK+热成像","quantity":2,"location":"南岸","contact_info":"138"}`),
		domain.RoleAssociationAdmin)
	if w.Code != http.StatusCreated {
		t.Fatalf("create resource: %d %s", w.Code, w.Body.String())
	}

	// 小程序契约：res_type=drone + q
	w = request(t, app, http.MethodGet, "/api/v1/emergency-resources?res_type=drone&q=热成像", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK {
		t.Fatalf("list filtered: %d %s", w.Code, w.Body.String())
	}
	data, total := listEnvelope(t, w)
	if total != 1 || len(data) != 1 {
		t.Fatalf("filtered list: total=%d len=%d, want 1/1; body=%s", total, len(data), w.Body.String())
	}

	// 类型不匹配 → 空
	w = request(t, app, http.MethodGet, "/api/v1/emergency-resources?res_type=vehicle", nil, domain.RoleIndividual)
	data, total = listEnvelope(t, w)
	if total != 0 || len(data) != 0 {
		t.Fatalf("类型不匹配: total=%d len=%d, want 0/0", total, len(data))
	}
}
