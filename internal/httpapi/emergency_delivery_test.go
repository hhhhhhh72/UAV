package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"drone-platform/internal/domain"
)

// 应急模块交付包契约回归：
// 1) createEmergencyDept 接收 dept_type/contact_name（此前无 json tag 被丢弃）
// 2) createEmergencyDrill 接收 dept_id/drone_count
// 3) createEmergencyDispatch 接收 status
// 4) listEmergencyDispatches 支持 status 值域归并（ongoing 匹配 dispatched）与 resource_id 过滤
func TestEmergencyDeliveryContract(t *testing.T) {
	app := newBizServer(t)
	admin := "admin-1"

	// 1) 部门：snake_case 字段必须落库
	dw := requestAs(t, app, http.MethodPost, "/api/v1/admin/emergency-depts",
		[]byte(`{"name":"南岸应急办","dept_type":"government","region":"南岸区","contact_name":"李科","contact_phone":"13800001111","protocol_url":"/uploads/p.pdf"}`),
		admin, domain.RolePlatformAdmin)
	if dw.Code != http.StatusCreated {
		t.Fatalf("create dept: %d %s", dw.Code, dw.Body.String())
	}
	if !jsonContains(dw.Body.String(), `"dept_type":"government"`) || !jsonContains(dw.Body.String(), `"contact_name":"李科"`) {
		t.Fatalf("dept snake_case fields must persist, got: %s", dw.Body.String())
	}

	// 2) 演练：dept_id/drone_count 必须落库
	dw = requestAs(t, app, http.MethodPost, "/api/v1/admin/emergency-drills",
		[]byte(`{"dept_id":"dept-1","title":"防汛演练","scenario":"洪涝","participants":20,"drone_count":5,"date":"2026-08-20T00:00:00Z"}`),
		admin, domain.RolePlatformAdmin)
	if dw.Code != http.StatusCreated {
		t.Fatalf("create drill: %d %s", dw.Code, dw.Body.String())
	}
	if !jsonContains(dw.Body.String(), `"dept_id":"dept-1"`) || !jsonContains(dw.Body.String(), `"drone_count":5`) {
		t.Fatalf("drill snake_case fields must persist, got: %s", dw.Body.String())
	}

	// 3) 创建资源 + 调度（status=pending）
	rw := requestAs(t, app, http.MethodPost, "/api/v1/admin/emergency-resources",
		[]byte(`{"name":"侦察机01","res_type":"drone","specs":"M300","location":"南岸","contact_info":"138","quantity":2}`),
		admin, domain.RolePlatformAdmin)
	if rw.Code != http.StatusCreated {
		t.Fatalf("create resource: %d %s", rw.Code, rw.Body.String())
	}
	var res struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rw.Body.Bytes(), &res); err != nil {
		t.Fatalf("parse resource: %v", err)
	}
	dw = requestAs(t, app, http.MethodPost, "/api/v1/admin/emergency-dispatches",
		[]byte(`{"resource_id":"`+res.Data.ID+`","event_desc":"山火","location":"北碚","commander":"张","status":"pending","start_time":"2026-08-18T10:00:00Z"}`),
		admin, domain.RolePlatformAdmin)
	if dw.Code != http.StatusCreated {
		t.Fatalf("create dispatch: %d %s", dw.Code, dw.Body.String())
	}
	if !jsonContains(dw.Body.String(), `"status":"pending"`) {
		t.Fatalf("dispatch status must persist, got: %s", dw.Body.String())
	}

	// 4) 状态筛选：resource_id 过滤 + status 精确匹配（pending 命中）
	req := httptest.NewRequest(http.MethodGet, "/api/v1/emergency-dispatches?status=pending&resource_id="+res.Data.ID, nil)
	lw := httptest.NewRecorder()
	app.ServeHTTP(lw, req)
	if lw.Code != http.StatusOK {
		t.Fatalf("list dispatches: %d", lw.Code)
	}
	if !jsonContains(lw.Body.String(), `"resource_id":"`+res.Data.ID+`"`) {
		t.Fatalf("resource_id filter must include target, got: %s", lw.Body.String())
	}

	// 5) 值域归并：默认创建（dispatched）后，status=ongoing 应命中
	dw = requestAs(t, app, http.MethodPost, "/api/v1/admin/emergency-dispatches",
		[]byte(`{"resource_id":"`+res.Data.ID+`","event_desc":"演练","location":"南岸","start_time":"2026-08-18T10:00:00Z"}`),
		admin, domain.RolePlatformAdmin)
	if dw.Code != http.StatusCreated {
		t.Fatalf("create dispatch (default): %d %s", dw.Code, dw.Body.String())
	}
	req = httptest.NewRequest(http.MethodGet, "/api/v1/emergency-dispatches?status=ongoing", nil)
	lw = httptest.NewRecorder()
	app.ServeHTTP(lw, req)
	if lw.Code != http.StatusOK {
		t.Fatalf("list dispatches ongoing: %d", lw.Code)
	}
	if !jsonContains(lw.Body.String(), `"event_desc":"演练"`) {
		t.Fatalf("ongoing must match dispatched records, got: %s", lw.Body.String())
	}
}
