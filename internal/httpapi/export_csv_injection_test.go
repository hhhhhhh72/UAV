package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"drone-platform/internal/domain"
)

// CSV 导出公式注入回归：需求标题含 =HYPERLINK(...) 时导出内容必须前缀单引号，
// 防止管理员用 Excel 打开 CSV 时执行恶意公式。
func TestExportDemandsCsvInjectionNeutralized(t *testing.T) {
	app := newBizServer(t)

	// 发布一个恶意标题需求 → 管理员审批通过（进入导出数据集）
	dw := requestAs(t, app, http.MethodPost, "/api/v1/demands",
		[]byte(`{"title":"=HYPERLINK(\"http://evil.example\")","contact":"13800000000","biz_type":"other"}`),
		"ent-1", domain.RoleEnterprise)
	if dw.Code != http.StatusCreated {
		t.Fatalf("create demand: %d %s", dw.Code, dw.Body.String())
	}
	var created struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := jsonUnmarshal(dw.Body.Bytes(), &created); err != nil {
		t.Fatalf("parse demand: %v", err)
	}
	rw := requestAs(t, app, http.MethodPost, "/api/v1/admin/demands/"+created.Data.ID+"/review",
		[]byte(`{"action":"approve"}`), "admin-1", domain.RolePlatformAdmin)
	if rw.Code != http.StatusOK {
		t.Fatalf("approve demand: %d %s", rw.Code, rw.Body.String())
	}

	// 平台管理员导出 CSV
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/export/demands", nil)
	req.Header.Set("Authorization", authAs(t, "admin-1", domain.RolePlatformAdmin))
	app.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("export: %d", w.Code)
	}
	body := w.Body.String()
	// 中和生效：恶意单元格被前缀单引号（CSV 转义后表现为 "'=HYPERLINK），
	// Excel 打开时作为文本而非公式执行。
	if !strings.Contains(body, "'=HYPERLINK") {
		t.Fatalf("CSV export should prefix quote to formula cells, got: %s", body)
	}
	// 严禁出现"单元格起始处即裸公式"（,=HYPERLINK 或 \n=HYPERLINK）
	if strings.Contains(body, ",=HYPERLINK") || strings.Contains(body, "\n=HYPERLINK") ||
		strings.Contains(body, "\r=HYPERLINK") {
		t.Fatalf("CSV export must neutralize formula injection, got: %s", body)
	}
}

// 辅助：避免与既有 helper 命名冲突
func jsonUnmarshal(b []byte, v any) error { return json.Unmarshal(b, v) }
