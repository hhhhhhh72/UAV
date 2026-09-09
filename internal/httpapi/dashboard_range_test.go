package httpapi_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"drone-platform/internal/domain"
)

// /api/v1/admin/dashboard?range=：窗口/粒度回显正确，且报名/成交/工单三条业务趋势存在且长度与窗口一致。
func TestAdminDashboardRangeParam(t *testing.T) {
	app := newBizServer(t)
	cases := []struct {
		query      string
		wantRange  string
		wantBucket string
		wantPoints int
	}{
		{"?range=7d", "7d", "day", 7},
		{"?range=30d", "30d", "day", 30},
		{"?range=90d", "90d", "day", 90},
		{"", "12m", "month", 12},
		{"?range=bogus", "12m", "month", 12},
	}
	for _, c := range cases {
		w := request(t, app, http.MethodGet, "/api/v1/admin/dashboard"+c.query, nil, domain.RolePlatformAdmin)
		if w.Code != http.StatusOK {
			t.Fatalf("GET dashboard%s: %d %s", c.query, w.Code, w.Body.String())
		}
		var resp struct {
			Data struct {
				Range        string `json:"range"`
				Bucket       string `json:"bucket"`
				TrendsDetail map[string][]struct {
					Date  string `json:"date"`
					Count int    `json:"count"`
				} `json:"trends_detail"`
			} `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("decode dashboard%s: %v", c.query, err)
		}
		if resp.Data.Range != c.wantRange || resp.Data.Bucket != c.wantBucket {
			t.Errorf("dashboard%s: range=%q bucket=%q, want %q %q", c.query, resp.Data.Range, resp.Data.Bucket, c.wantRange, c.wantBucket)
		}
		for _, dim := range []string{"demand", "enrollment", "order", "work_order"} {
			series, ok := resp.Data.TrendsDetail[dim]
			if !ok {
				t.Fatalf("dashboard%s: trends_detail.%s 缺失", c.query, dim)
			}
			if len(series) != c.wantPoints {
				t.Errorf("dashboard%s: trends_detail.%s 长度 = %d, want %d", c.query, dim, len(series), c.wantPoints)
			}
		}
	}
}

// 非管理员不可访问仪表盘。
func TestAdminDashboardRequiresAdmin(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodGet, "/api/v1/admin/dashboard?range=30d", nil, domain.RoleIndividual)
	if w.Code != http.StatusForbidden {
		t.Fatalf("individual 访问 dashboard: %d, want 403", w.Code)
	}
}
