package httpapi_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"drone-platform/internal/domain"
)

// 本文件为「业务流程打通」审计 P0 项的回归测试：
//   P0-1 派单（assignments）必须真实落库并可被雇主/工人双向查询
//   P0-2 企业资质文件（enterprise documents）必须落库并可查询
//   P0-3 协会会员批量导入必须逐行落库并返回导入明细
// 辅助函数 request()/requestAs()/listEnvelope() 来自同包既有测试文件。

// ── P0-1：派单闭环 ──────────────────────────────────────────────

func TestAssignmentPersistsAndLists(t *testing.T) {
	app := newBizServer(t)

	// 雇主（user-1, enterprise）创建用工订单
	w := request(t, app, http.MethodPost, "/api/v1/labour-orders",
		[]byte(`{"title":"航测外业","description":"需要两名飞手","worker_count":2,"start_date":"2026-08-20T00:00:00Z","end_date":"2026-08-22T00:00:00Z","budget_fen":500000}`),
		domain.RoleEnterprise)
	if w.Code != http.StatusCreated {
		t.Fatalf("create labour order: %d %s", w.Code, w.Body.String())
	}
	var created struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil || created.Data.ID == "" {
		t.Fatalf("labour order response: %s", w.Body.String())
	}
	orderID := created.Data.ID

	// 雇主派单 → 201
	w = request(t, app, http.MethodPost, "/api/v1/assignments",
		[]byte(`{"order_id":"`+orderID+`","worker_id":"user-2"}`), domain.RoleEnterprise)
	if w.Code != http.StatusCreated {
		t.Fatalf("assign worker: %d %s", w.Code, w.Body.String())
	}

	// 缺 worker_id → 400
	w = request(t, app, http.MethodPost, "/api/v1/assignments",
		[]byte(`{"order_id":"`+orderID+`"}`), domain.RoleEnterprise)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("assign missing worker: %d, want 400", w.Code)
	}

	// 非雇主（user-2 individual）派单 → 403
	w = requestAs(t, app, http.MethodPost, "/api/v1/assignments",
		[]byte(`{"order_id":"`+orderID+`","worker_id":"user-3"}`), "user-2", domain.RoleIndividual)
	if w.Code != http.StatusForbidden {
		t.Fatalf("assign by non-employer: %d, want 403", w.Code)
	}

	// 订单不存在 → 404
	w = request(t, app, http.MethodPost, "/api/v1/assignments",
		[]byte(`{"order_id":"labour-nope","worker_id":"user-2"}`), domain.RoleEnterprise)
	if w.Code != http.StatusNotFound {
		t.Fatalf("assign to missing order: %d, want 404", w.Code)
	}

	// 雇主查订单派单列表 → 1 条
	w = request(t, app, http.MethodGet, "/api/v1/labour-orders/"+orderID+"/assignments", nil, domain.RoleEnterprise)
	if w.Code != http.StatusOK {
		t.Fatalf("list order assignments: %d %s", w.Code, w.Body.String())
	}
	var list struct {
		Data []domain.Assignment `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &list); err != nil {
		t.Fatalf("parse assignments: %v", err)
	}
	if len(list.Data) != 1 || list.Data[0].WorkerID != "user-2" || list.Data[0].Status != "assigned" {
		t.Fatalf("order assignments: %+v, want 1 assigned item for user-2", list.Data)
	}

	// 无关用户查他人订单派单列表 → 403
	w = requestAs(t, app, http.MethodGet, "/api/v1/labour-orders/"+orderID+"/assignments", nil, "user-3", domain.RoleIndividual)
	if w.Code != http.StatusForbidden {
		t.Fatalf("list assignments by stranger: %d, want 403", w.Code)
	}

	// 工人查「我的派单」→ 1 条，订单号对得上
	w = requestAs(t, app, http.MethodGet, "/api/v1/assignments/mine", nil, "user-2", domain.RoleIndividual)
	if w.Code != http.StatusOK {
		t.Fatalf("my assignments: %d %s", w.Code, w.Body.String())
	}
	list.Data = nil
	if err := json.Unmarshal(w.Body.Bytes(), &list); err != nil {
		t.Fatalf("parse my assignments: %v", err)
	}
	if len(list.Data) != 1 || list.Data[0].OrderID != orderID {
		t.Fatalf("my assignments: %+v, want order %s", list.Data, orderID)
	}

	// 未派单的工人查「我的派单」→ 空列表
	w = requestAs(t, app, http.MethodGet, "/api/v1/assignments/mine", nil, "user-3", domain.RoleIndividual)
	if w.Code != http.StatusOK {
		t.Fatalf("empty my assignments: %d", w.Code)
	}
	list.Data = nil
	json.Unmarshal(w.Body.Bytes(), &list)
	if len(list.Data) != 0 {
		t.Fatalf("user-3 should have no assignments, got %+v", list.Data)
	}
}

// ── P0-2：企业资质文件闭环 ─────────────────────────────────────────

func TestEnterpriseDocumentsPersist(t *testing.T) {
	app := newBizServer(t)

	// user-1 创建企业
	w := request(t, app, http.MethodPost, "/api/v1/enterprises",
		[]byte(`{"name":"资质测试企业","account_name":"622200000001"}`), domain.RoleEnterprise)
	if w.Code != http.StatusCreated {
		t.Fatalf("create enterprise: %d %s", w.Code, w.Body.String())
	}
	var created struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil || created.Data.ID == "" {
		t.Fatalf("enterprise response: %s", w.Body.String())
	}
	entID := created.Data.ID

	// 企业主挂接资质文件 → 201
	w = request(t, app, http.MethodPost, "/api/v1/enterprises/"+entID+"/documents",
		[]byte(`{"file_id":"f-license","document_type":"business_license"}`), domain.RoleEnterprise)
	if w.Code != http.StatusCreated {
		t.Fatalf("attach document: %d %s", w.Code, w.Body.String())
	}

	// 缺 document_type → 400
	w = request(t, app, http.MethodPost, "/api/v1/enterprises/"+entID+"/documents",
		[]byte(`{"file_id":"f-2"}`), domain.RoleEnterprise)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("attach missing type: %d, want 400", w.Code)
	}

	// 非企业主（user-2）挂接 → 403
	w = requestAs(t, app, http.MethodPost, "/api/v1/enterprises/"+entID+"/documents",
		[]byte(`{"file_id":"f-x","document_type":"id_card"}`), "user-2", domain.RoleEnterprise)
	if w.Code != http.StatusForbidden {
		t.Fatalf("attach by non-owner: %d, want 403", w.Code)
	}

	// 企业主查文件列表 → 1 条
	w = request(t, app, http.MethodGet, "/api/v1/enterprises/"+entID+"/documents", nil, domain.RoleEnterprise)
	if w.Code != http.StatusOK {
		t.Fatalf("list documents: %d %s", w.Code, w.Body.String())
	}
	var docs struct {
		Data []domain.EnterpriseDocument `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &docs); err != nil {
		t.Fatalf("parse documents: %v", err)
	}
	if len(docs.Data) != 1 || docs.Data[0].DocumentType != "business_license" || docs.Data[0].ReviewStatus != "pending" {
		t.Fatalf("documents: %+v, want 1 pending business_license", docs.Data)
	}

	// 非企业主查询 → 403
	w = requestAs(t, app, http.MethodGet, "/api/v1/enterprises/"+entID+"/documents", nil, "user-2", domain.RoleEnterprise)
	if w.Code != http.StatusForbidden {
		t.Fatalf("list by non-owner: %d, want 403", w.Code)
	}

	// 企业不存在 → 404
	w = request(t, app, http.MethodGet, "/api/v1/enterprises/ent-nope/documents", nil, domain.RoleEnterprise)
	if w.Code != http.StatusNotFound {
		t.Fatalf("documents of missing enterprise: %d, want 404", w.Code)
	}
}

// ── P0-7：公开企业详情闭环 ────────────────────────────────────────
// GET /api/v1/enterprises/public/detail?id= 匿名可访问，但仅已审核企业可见；
// 草稿/审核中/不存在统一 404；响应只含展示字段，不泄露电话/信用代码等敏感信息。

func TestEnterprisePublicDetailOnlyApproved(t *testing.T) {
	app := newBizServer(t)

	// user-1 创建企业（draft，带完整展示字段 + 敏感字段）
	w := request(t, app, http.MethodPost, "/api/v1/enterprises",
		[]byte(`{"name":"公开详情测试企业","credit_code":"91110108MA01TESTX","contact_phone":"13800001234","industry_category":"测绘","scale":"20-99人","address":"深圳市南山区科技园","description":"专注无人机测绘","business_hours":"9:00-18:00","founded_at":"2019-06-01","capability_tags":"测绘,巡检"}`), domain.RoleEnterprise)
	if w.Code != http.StatusCreated {
		t.Fatalf("create enterprise: %d %s", w.Code, w.Body.String())
	}
	var created struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil || created.Data.ID == "" {
		t.Fatalf("enterprise response: %s", w.Body.String())
	}
	entID := created.Data.ID

	// 草稿态：公开详情不可见（404，不暴露存在性）
	w = doRaw(app, http.MethodGet, "/api/v1/enterprises/public/detail?id="+entID, "", "")
	if w.Code != http.StatusNotFound {
		t.Fatalf("draft enterprise public detail: %d, want 404", w.Code)
	}

	// 提交审核 → 审核中仍不可见
	w = request(t, app, http.MethodPost, "/api/v1/enterprises/"+entID+"/submit", nil, domain.RoleEnterprise)
	if w.Code != http.StatusOK {
		t.Fatalf("submit enterprise: %d %s", w.Code, w.Body.String())
	}
	w = doRaw(app, http.MethodGet, "/api/v1/enterprises/public/detail?id="+entID, "", "")
	if w.Code != http.StatusNotFound {
		t.Fatalf("submitted enterprise public detail: %d, want 404", w.Code)
	}

	// 管理员审核通过
	w = request(t, app, http.MethodPost, "/api/v1/admin/enterprises/"+entID+"/review",
		[]byte(`{"action":"approve"}`), domain.RolePlatformAdmin)
	if w.Code != http.StatusOK {
		t.Fatalf("review enterprise: %d %s", w.Code, w.Body.String())
	}

	// 匿名访问已审核企业详情 → 200 + 展示字段齐全
	w = doRaw(app, http.MethodGet, "/api/v1/enterprises/public/detail?id="+entID, "", "")
	if w.Code != http.StatusOK {
		t.Fatalf("public detail: %d %s", w.Code, w.Body.String())
	}
	var detail struct {
		Data map[string]any `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &detail); err != nil {
		t.Fatalf("parse public detail: %v", err)
	}
	for _, field := range []string{"name", "industry_category", "scale", "address", "business_hours", "founded_at", "description", "capability_tags", "is_member", "created_at", "logo", "cover_image"} {
		if _, ok := detail.Data[field]; !ok {
			t.Fatalf("public detail missing field %s: %s", field, w.Body.String())
		}
	}
	if detail.Data["name"] != "公开详情测试企业" || detail.Data["address"] != "深圳市南山区科技园" {
		t.Fatalf("public detail values: %+v", detail.Data)
	}
	if detail.Data["status"] != string(domain.EnterpriseApproved) {
		t.Fatalf("public detail status: %v, want approved", detail.Data["status"])
	}
	// 敏感字段绝不暴露（电话/信用代码/账户等仅管理员与本人可见）
	raw := w.Body.String()
	for _, sensitive := range []string{"credit_code", "contact_phone", "account_name", "license_url", "legal_person", "email"} {
		if strings.Contains(raw, sensitive) {
			t.Fatalf("public detail leaks %s: %s", sensitive, raw)
		}
	}

	// 不存在的企业 → 404
	w = doRaw(app, http.MethodGet, "/api/v1/enterprises/public/detail?id=ent-nope", "", "")
	if w.Code != http.StatusNotFound {
		t.Fatalf("missing enterprise public detail: %d, want 404", w.Code)
	}
}

// ── P0-3：协会会员批量导入闭环 ─────────────────────────────────────

func TestImportMembersPersists(t *testing.T) {
	app := newBizServer(t)

	body := `{"members":[
		{"user_id":"u-a","role":"member"},
		{"user_id":"u-b","enterprise_id":"ent-1","role":"secretary"},
		{"role":"member"},
		{"user_id":"u-c","role":"boss"}
	]}`
	// 行 0/1 合法，行 2 缺 user_id，行 3 非法角色
	w := request(t, app, http.MethodPost, "/api/v1/admin/members/import",
		[]byte(body), domain.RoleAssociationAdmin)
	if w.Code != http.StatusOK {
		t.Fatalf("import members: %d %s", w.Code, w.Body.String())
	}
	var res struct {
		Data struct {
			Imported int `json:"imported"`
			Total    int `json:"total"`
			Failed   []struct {
				Index  int    `json:"index"`
				UserID string `json:"user_id"`
				Error  string `json:"error"`
			} `json:"failed"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("parse import result: %v", err)
	}
	if res.Data.Total != 4 || res.Data.Imported != 2 || len(res.Data.Failed) != 2 {
		t.Fatalf("import result: %+v, want total=4 imported=2 failed=2", res.Data)
	}
	if res.Data.Failed[0].Index != 2 || !strings.Contains(res.Data.Failed[0].Error, "user_id") {
		t.Fatalf("failed[0]: %+v, want index 2 missing user_id", res.Data.Failed[0])
	}
	if res.Data.Failed[1].Index != 3 || !strings.Contains(res.Data.Failed[1].Error, "invalid role") {
		t.Fatalf("failed[1]: %+v, want index 3 invalid role", res.Data.Failed[1])
	}

	// 空数组 → 400
	w = request(t, app, http.MethodPost, "/api/v1/admin/members/import",
		[]byte(`{"members":[]}`), domain.RoleAssociationAdmin)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("empty import: %d, want 400", w.Code)
	}

	// 非管理员 → 403（adminGate 拦截）
	w = request(t, app, http.MethodPost, "/api/v1/admin/members/import",
		[]byte(body), domain.RoleIndividual)
	if w.Code != http.StatusForbidden {
		t.Fatalf("import as individual: %d, want 403", w.Code)
	}

	// 导入结果可通过会员列表查询到（total=2）
	w = request(t, app, http.MethodGet, "/api/v1/association-members", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK {
		t.Fatalf("list members: %d %s", w.Code, w.Body.String())
	}
	_, total := listEnvelope(t, w)
	if total != 2 {
		t.Fatalf("members total: %d, want 2", total)
	}
}
