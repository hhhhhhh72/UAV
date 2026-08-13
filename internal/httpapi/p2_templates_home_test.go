package httpapi_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"drone-platform/internal/domain"
)

// P2-2 回归：合同模板从库读取（memory repo 预置种子，对齐 PG 迁移 000062）。
func TestContractTemplatesFromRepo(t *testing.T) {
	app := newBizServer(t)

	w := request(t, app, http.MethodGet, "/api/v1/contract-templates", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK {
		t.Fatalf("list contract templates: %d %s", w.Code, w.Body.String())
	}
	var resp struct {
		Data []domain.ContractTemplate `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse templates: %v", err)
	}
	byID := map[string]domain.ContractTemplate{}
	for _, tpl := range resp.Data {
		byID[tpl.ID] = tpl
	}
	for _, want := range []string{"tpl-001", "tpl-002"} {
		tpl, ok := byID[want]
		if !ok {
			t.Fatalf("template %s missing from repo-backed list: %+v", want, resp.Data)
		}
		if tpl.Name == "" || tpl.Status == "" {
			t.Fatalf("template %s missing name/status fields: %+v", want, tpl)
		}
	}
}

// P2-2 回归：服务未装配时（旧测试服务器等）列表兜底返回内置模板，不 500。
func TestContractTemplatesFallbackWithoutService(t *testing.T) {
	app := newServer(t) // newServer 不装配 ContractTemplateService

	w := request(t, app, http.MethodGet, "/api/v1/contract-templates", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK {
		t.Fatalf("fallback templates: %d %s", w.Code, w.Body.String())
	}
	for _, want := range []string{"tpl-001", "tpl-002"} {
		if !bytes.Contains(w.Body.Bytes(), []byte(want)) {
			t.Fatalf("fallback list missing %s: %s", want, w.Body.String())
		}
	}
}

// P2-2 回归：首页 stats.views 从商品浏览量实时汇总（不再返回硬编码 6690000）。
func TestHomeViewsFromProducts(t *testing.T) {
	app := newBizServer(t)

	homeViews := func() int {
		t.Helper()
		w := request(t, app, http.MethodGet, "/api/v1/home", nil, domain.RoleIndividual)
		if w.Code != http.StatusOK {
			t.Fatalf("home: %d %s", w.Code, w.Body.String())
		}
		var resp struct {
			Data struct {
				Stats struct {
					Views int `json:"views"`
				} `json:"stats"`
			} `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("parse home: %v", err)
		}
		return resp.Data.Stats.Views
	}

	base := homeViews()

	// 发布商品 → 待审核
	pw := request(t, app, http.MethodPost, "/api/v1/products",
		[]byte(`{"title":"浏览量测试机","prod_type":"drone","price_fen":100000,"brand":"DJI","model":"M350"}`),
		domain.RoleEnterprise)
	if pw.Code != http.StatusCreated {
		t.Fatalf("create product: %d %s", pw.Code, pw.Body.String())
	}
	var created struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(pw.Body.Bytes(), &created); err != nil {
		t.Fatalf("parse product: %v", err)
	}
	id := created.Data.ID
	if id == "" {
		t.Fatal("empty product id")
	}

	// 管理后台通过 → listed
	aw := request(t, app, http.MethodPut, "/api/v1/admin/products/"+id,
		[]byte(`{"status":"listed"}`), domain.RolePlatformAdmin)
	if aw.Code != http.StatusOK {
		t.Fatalf("approve product: %d %s", aw.Code, aw.Body.String())
	}

	// 详情访问 2 次，浏览量应 +2
	for i := 0; i < 2; i++ {
		dw := request(t, app, http.MethodGet, "/api/v1/products/"+id, nil, domain.RoleIndividual)
		if dw.Code != http.StatusOK {
			t.Fatalf("product detail hit %d: %d %s", i, dw.Code, dw.Body.String())
		}
	}

	got := homeViews()
	if got < base+2 {
		t.Fatalf("home views should reflect detail hits: base=%d got=%d", base, got)
	}
	if got == 6690000 {
		t.Fatal("home views still returns hardcoded 6690000")
	}
	if got == base {
		t.Fatal("home views did not change after detail hits")
	}
}
