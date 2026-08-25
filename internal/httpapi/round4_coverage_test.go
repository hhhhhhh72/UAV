package httpapi_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"drone-platform/internal/domain"
)

// r4CreateDemand 发布一条需求并返回 id（enterprise 角色）。
func r4CreateDemand(t *testing.T, app http.Handler, tok, title string) string {
	t.Helper()
	w := doRaw(app, http.MethodPost, "/api/v1/demands",
		`{"title":"`+title+`","contact":"13800000000","biz_type":"cable_inspection"}`, tok)
	assertStatus(t, http.MethodPost, "/api/v1/demands", w, http.StatusCreated)
	return dataID(t, w)
}

// r4ApproveDemand 管理端审核通过一条需求。
func r4ApproveDemand(t *testing.T, app http.Handler, adminTok, id string) {
	t.Helper()
	w := doRaw(app, http.MethodPost, "/api/v1/admin/demands/"+id+"/review",
		`{"action":"approve"}`, adminTok)
	assertStatus(t, http.MethodPost, "/api/v1/admin/demands/"+id+"/review", w, http.StatusOK)
}

// doRawWithKey 类似 doRaw，但额外携带 Idempotency-Key 头。
func doRawWithKey(app http.Handler, method, path, body, token, idemKey string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	if token != "" {
		r.Header.Set("Authorization", token)
	}
	if idemKey != "" {
		r.Header.Set("Idempotency-Key", idemKey)
	}
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	return w
}

// jsonFloatID 从 {data:{data:{key:float}}} 双层包裹响应中取数值型 id（H5 JSON 路由）。
func jsonFloatID(t *testing.T, w *httptest.ResponseRecorder, key string) string {
	t.Helper()
	var outer struct {
		Data struct {
			Data map[string]any `json:"data"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &outer); err != nil {
		t.Fatalf("parse jsonFloatID: %v (body=%s)", err, w.Body.String())
	}
	v, _ := outer.Data.Data[key].(float64)
	if v == 0 {
		t.Fatalf("missing %s: %s", key, w.Body.String())
	}
	return fmt.Sprintf("%.0f", v)
}

// TestR4ServerMetaAndCore 覆盖 server.go 的 meta 路由与 search/listDemands/listAdminDemands/adminDemandStats。
func TestR4ServerMetaAndCore(t *testing.T) {
	app := newBizServer(t)
	adminTok := authAs(t, "admin-1", domain.RolePlatformAdmin)

	// index / favicon / uploads（静态文件服务）
	if w := doRaw(app, http.MethodGet, "/", "", ""); w.Code != http.StatusOK {
		t.Fatalf("GET /: %d", w.Code)
	}
	if w := doRaw(app, http.MethodGet, "/favicon.ico", "", ""); w.Code != http.StatusOK {
		t.Fatalf("GET /favicon.ico: %d", w.Code)
	}
	if w := doRaw(app, http.MethodGet, "/uploads/", "", ""); !contains([]int{http.StatusOK, http.StatusNotFound}, w.Code) {
		t.Fatalf("GET /uploads/: %d", w.Code)
	}

	// search（失败容错 + 企业账号脱敏分支）
	if w := doRaw(app, http.MethodGet, "/api/v1/search?q=巡检", "", ""); w.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/search: %d %s", w.Code, w.Body.String())
	}

	// listDemands 的 keyword/district/biz_type 过滤分支
	if w := doRaw(app, http.MethodGet, "/api/v1/demands?q=xx&district=渝北&biz_type=cable_inspection&sort=latest", "", ""); w.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/demands filtered: %d %s", w.Code, w.Body.String())
	}

	// listAdminDemands：status 默认 all 与显式 status
	if w := doRaw(app, http.MethodGet, "/api/v1/admin/demands", "", adminTok); w.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/admin/demands: %d %s", w.Code, w.Body.String())
	}
	if w := doRaw(app, http.MethodGet, "/api/v1/admin/demands?status=pending", "", adminTok); w.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/admin/demands?status=pending: %d %s", w.Code, w.Body.String())
	}

	// adminDemandStats：统计聚合（含 rate/offline_amount 汇总）
	if w := doRaw(app, http.MethodGet, "/api/v1/admin/demands/stats", "", adminTok); w.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/admin/demands/stats: %d %s", w.Code, w.Body.String())
	}
}

// TestR4DemandDetailAndUpdate 覆盖 demandDetail / updateDemand / cancelDemand 的 404 与脱敏分支。
func TestR4DemandDetailAndUpdate(t *testing.T) {
	app := newBizServer(t)
	entTok := authAs(t, "ent-1", domain.RoleEnterprise)

	id := r4CreateDemand(t, app, entTok, "详情与编辑需求")

	// 发布者看自己 pending 需求 → 200（含脱敏：PublisherID 被清空）
	w := doRaw(app, http.MethodGet, "/api/v1/demands/"+id, "", entTok)
	assertStatus(t, http.MethodGet, "/api/v1/demands/"+id, w, http.StatusOK)
	if strings.Contains(w.Body.String(), `"publisher_id":"ent-1"`) {
		t.Fatalf("demandDetail should strip publisher_id: %s", w.Body.String())
	}

	// 匿名看 pending → 404
	w = doRaw(app, http.MethodGet, "/api/v1/demands/"+id, "", "")
	assertStatus(t, http.MethodGet, "/api/v1/demands/"+id+" (anon)", w, http.StatusNotFound)

	// 不存在 → 404
	w = doRaw(app, http.MethodGet, "/api/v1/demands/"+nonexistentID, "", "")
	assertStatus(t, http.MethodGet, "/api/v1/demands/zzz", w, http.StatusNotFound)

	// updateDemand 草稿编辑 → 200；不存在 → 404
	w = doRaw(app, http.MethodPatch, "/api/v1/demands/"+id, `{"title":"编辑后标题","description":"补充描述"}`, entTok)
	assertStatus(t, http.MethodPatch, "/api/v1/demands/"+id, w, http.StatusOK)
	w = doRaw(app, http.MethodPatch, "/api/v1/demands/"+nonexistentID, `{"title":"x"}`, entTok)
	assertStatus(t, http.MethodPatch, "/api/v1/demands/zzz", w, http.StatusNotFound)

	// cancelDemand 不存在 → 404（not-found 分支）
	w = doRaw(app, http.MethodPost, "/api/v1/demands/"+nonexistentID+"/cancel", "", entTok)
	assertStatus(t, http.MethodPost, "/api/v1/demands/zzz/cancel", w, http.StatusNotFound)
}

// TestR4DemandAdminLifecycle 覆盖 approveDemand/closeDemand/setDemandOfflineAmount/deleteDemand。
func TestR4DemandAdminLifecycle(t *testing.T) {
	app := newBizServer(t)
	entTok := authAs(t, "ent-1", domain.RoleEnterprise)
	adminTok := authAs(t, "admin-1", domain.RolePlatformAdmin)

	// approveDemand（/approve 端点，含 contact 脱敏）
	d1 := r4CreateDemand(t, app, entTok, "审批关闭需求")
	w := doRaw(app, http.MethodPost, "/api/v1/admin/demands/"+d1+"/approve", "", adminTok)
	assertStatus(t, http.MethodPost, "/api/v1/admin/demands/"+d1+"/approve", w, http.StatusOK)
	// closeDemand → 200（published → cancelled）
	w = doRaw(app, http.MethodPost, "/api/v1/admin/demands/"+d1+"/close", `{"reason":"虚假信息"}`, adminTok)
	assertStatus(t, http.MethodPost, "/api/v1/admin/demands/"+d1+"/close", w, http.StatusOK)
	// closeDemand 不存在 → 404
	w = doRaw(app, http.MethodPost, "/api/v1/admin/demands/"+nonexistentID+"/close", `{"reason":"x"}`, adminTok)
	assertStatus(t, http.MethodPost, "/api/v1/admin/demands/zzz/close", w, http.StatusNotFound)

	// setDemandOfflineAmount：发布→审批→登记金额→200；不存在→404
	d2 := r4CreateDemand(t, app, entTok, "线下成交需求")
	r4ApproveDemand(t, app, adminTok, d2)
	w = doRaw(app, http.MethodPost, "/api/v1/admin/demands/"+d2+"/amount", `{"offline_amount_fen":123456}`, adminTok)
	assertStatus(t, http.MethodPost, "/api/v1/admin/demands/"+d2+"/amount", w, http.StatusOK)
	w = doRaw(app, http.MethodPost, "/api/v1/admin/demands/"+nonexistentID+"/amount", `{"offline_amount_fen":1}`, adminTok)
	assertStatus(t, http.MethodPost, "/api/v1/admin/demands/zzz/amount", w, http.StatusNotFound)

	// deleteDemand：取消后可删 → 200；不存在 → 400（Delete 统一 400）
	d3 := r4CreateDemand(t, app, entTok, "待删除需求")
	w = doRaw(app, http.MethodPost, "/api/v1/demands/"+d3+"/cancel", "", entTok)
	assertStatus(t, http.MethodPost, "/api/v1/demands/"+d3+"/cancel", w, http.StatusOK)
	w = doRaw(app, http.MethodDelete, "/api/v1/admin/demands/"+d3, "", adminTok)
	assertStatus(t, http.MethodDelete, "/api/v1/admin/demands/"+d3, w, http.StatusOK)
	w = doRaw(app, http.MethodDelete, "/api/v1/admin/demands/"+nonexistentID, "", adminTok)
	assertStatus(t, http.MethodDelete, "/api/v1/admin/demands/zzz", w, http.StatusBadRequest)
}

// TestR4CreateDemandErrors 覆盖 createDemand 的 400（缺字段）与 403（角色不符）分支。
func TestR4CreateDemandErrors(t *testing.T) {
	app := newBizServer(t)
	// 缺 title/contact → 400（"required"）
	w := doRaw(app, http.MethodPost, "/api/v1/demands", `{}`, authAs(t, "ent-1", domain.RoleEnterprise))
	assertStatus(t, http.MethodPost, "/api/v1/demands (empty)", w, http.StatusBadRequest)
	// 协会管理员不可发布 → 403
	w = doRaw(app, http.MethodPost, "/api/v1/demands",
		`{"title":"越权","contact":"13800000000"}`, authAs(t, "admin-2", domain.RoleAssociationAdmin))
	assertStatus(t, http.MethodPost, "/api/v1/demands (admin)", w, http.StatusForbidden)
}

// TestR4IdempotencyMiddleware 覆盖 idempotencyStore 的 get/set + 幂等重放 + 短 key 校验。
func TestR4IdempotencyMiddleware(t *testing.T) {
	app := newBizServer(t)
	entTok := authAs(t, "ent-1", domain.RoleEnterprise)

	// 短 key（<8）→ 400
	w := doRawWithKey(app, http.MethodPost, "/api/v1/demands", `{"title":"幂等需求","contact":"13800000000"}`, entTok, "short")
	assertStatus(t, http.MethodPost, "/api/v1/demands (short key)", w, http.StatusBadRequest)

	// 合法 key 首次执行 → 201
	w = doRawWithKey(app, http.MethodPost, "/api/v1/demands", `{"title":"幂等需求","contact":"13800000000"}`, entTok, "r4-idem-key-1")
	assertStatus(t, http.MethodPost, "/api/v1/demands (idem)", w, http.StatusCreated)
	firstBody := w.Body.String()

	// 同 key 重放 → 命中缓存，返回与首次一致的状态码与响应体
	w = doRawWithKey(app, http.MethodPost, "/api/v1/demands", `{"title":"幂等需求","contact":"13800000000"}`, entTok, "r4-idem-key-1")
	assertStatus(t, http.MethodPost, "/api/v1/demands (replay)", w, http.StatusCreated)
	if w.Body.String() != firstBody {
		t.Fatalf("idempotent replay body mismatch:\nfirst=%s\nreplay=%s", firstBody, w.Body.String())
	}

	// 不同用户同 key → 不互相去重（actor 命名空间），应正常创建
	w = doRawWithKey(app, http.MethodPost, "/api/v1/demands", `{"title":"幂等需求2","contact":"13800000000"}`, authAs(t, "ent-2", domain.RoleEnterprise), "r4-idem-key-1")
	assertStatus(t, http.MethodPost, "/api/v1/demands (other actor)", w, http.StatusCreated)
	if w.Body.String() == firstBody {
		t.Fatal("different actor with same key must not replay another user's response")
	}
}

// TestR4Batch2ListAndCollege 覆盖 listTransformations/listColleges/matchCollegeType/createCollege。
func TestR4Batch2ListAndCollege(t *testing.T) {
	app := newBizServer(t)
	adminTok := authAs(t, "admin-1", domain.RolePlatformAdmin)

	// 转化成果列表（公开）
	if w := doRaw(app, http.MethodGet, "/api/v1/transformations", "", ""); w.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/transformations: %d", w.Code)
	}
	if w := doRaw(app, http.MethodGet, "/api/v1/transformations?achievement_id=ach-1", "", ""); w.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/transformations?achievement_id: %d", w.Code)
	}

	// 创建院校（专科 + 本科）以覆盖 matchCollegeType 两侧
	w := doRaw(app, http.MethodPost, "/api/v1/admin/colleges",
		`{"name":"渝北职业学院","city":"重庆","tags":["专科"],"short_name":"渝职院"}`, adminTok)
	assertStatus(t, http.MethodPost, "/api/v1/admin/colleges (vocational)", w, http.StatusCreated)
	w = doRaw(app, http.MethodPost, "/api/v1/admin/colleges",
		`{"name":"重庆大学","city":"重庆","tags":["本科","双一流"],"short_name":"重大"}`, adminTok)
	assertStatus(t, http.MethodPost, "/api/v1/admin/colleges (undergrad)", w, http.StatusCreated)

	// 院校创建须接收 logo_url（回归：曾因缺 json tag 导致新增院校 Logo 恒为空）
	w = doRaw(app, http.MethodPost, "/api/v1/admin/colleges",
		`{"name":"重庆理工职院","city":"重庆","logo_url":"/uploads/logo-a.png"}`, adminTok)
	assertStatus(t, http.MethodPost, "/api/v1/admin/colleges (logo)", w, http.StatusCreated)
	if !strings.Contains(w.Body.String(), "/uploads/logo-a.png") {
		t.Fatalf("create college should persist logo_url, got: %s", w.Body.String())
	}

	// listColleges 各种 type/keyword 分支
	for _, q := range []string{"", "?type=vocational", "?type=undergraduate", "?type=undergraduate&keyword=重庆", "?type=vocational&keyword=不存在"} {
		if w := doRaw(app, http.MethodGet, "/api/v1/colleges"+q, "", ""); w.Code != http.StatusOK {
			t.Fatalf("GET /api/v1/colleges%s: %d", q, w.Code)
		}
	}
}

// TestR4CooperationStatus 覆盖 createCooperation/updateCooperationStatus。
func TestR4CooperationStatus(t *testing.T) {
	app := newBizServer(t)
	adminTok := authAs(t, "admin-1", domain.RolePlatformAdmin)

	w := doRaw(app, http.MethodPost, "/api/v1/cooperation-programs",
		`{"title":"校企共建无人机实训基地","college_id":"c1","enterprise_id":"e1","coop_type":"talent"}`, adminTok)
	assertStatus(t, http.MethodPost, "/api/v1/cooperation-programs", w, http.StatusCreated)
	cpID := dataID(t, w)

	w = doRaw(app, http.MethodPost, "/api/v1/cooperation-programs/"+cpID+"/status", `{"status":"active"}`, adminTok)
	assertStatus(t, http.MethodPost, "/api/v1/cooperation-programs/"+cpID+"/status", w, http.StatusOK)

	// 不存在 → 404
	w = doRaw(app, http.MethodPost, "/api/v1/cooperation-programs/zzz/status", `{"status":"active"}`, adminTok)
	assertStatus(t, http.MethodPost, "/api/v1/cooperation-programs/zzz/status", w, http.StatusNotFound)
}

// TestR4TrainingCoverage 覆盖 approveCertificate/listPilots/getPilot/listAdminPilots。
func TestR4TrainingCoverage(t *testing.T) {
	app := newBizServer(t)
	userTok := authAs(t, "user-1", domain.RoleIndividual)
	adminTok := authAs(t, "admin-1", domain.RolePlatformAdmin)

	// 证书：创建 → 管理员批准
	w := doRaw(app, http.MethodPost, "/api/v1/certificates",
		`{"cert_type":"caac","cert_number":"R4-001","level":"三级","issuer_org":"CAAC"}`, userTok)
	assertStatus(t, http.MethodPost, "/api/v1/certificates", w, http.StatusCreated)
	certID := dataID(t, w)
	w = doRaw(app, http.MethodPost, "/api/v1/admin/certificates/"+certID+"/approve", "", adminTok)
	assertStatus(t, http.MethodPost, "/api/v1/admin/certificates/"+certID+"/approve", w, http.StatusOK)
	// 不存在证书 → 403（ApproveCertificate 错误统一 403）
	w = doRaw(app, http.MethodPost, "/api/v1/admin/certificates/zzz/approve", "", adminTok)
	assertStatus(t, http.MethodPost, "/api/v1/admin/certificates/zzz/approve", w, http.StatusForbidden)

	// 飞手：注册 → 批准 → 名录/详情（脱敏）
	w = doRaw(app, http.MethodPost, "/api/v1/certified-pilots",
		`{"real_name":"王飞手","id_card":"110101199001011234","flight_hours":120}`, userTok)
	assertStatus(t, http.MethodPost, "/api/v1/certified-pilots", w, http.StatusCreated)
	pilotID := dataID(t, w)
	w = doRaw(app, http.MethodPost, "/api/v1/admin/certified-pilots/"+pilotID+"/approve", "", adminTok)
	assertStatus(t, http.MethodPost, "/api/v1/admin/certified-pilots/"+pilotID+"/approve", w, http.StatusOK)

	// listPilots：keyword + page_size 分页分支
	w = doRaw(app, http.MethodGet, "/api/v1/certified-pilots?keyword=王&page_size=5", "", "")
	assertStatus(t, http.MethodGet, "/api/v1/certified-pilots?keyword", w, http.StatusOK)
	w = doRaw(app, http.MethodGet, "/api/v1/certified-pilots?page=9&page_size=5", "", "")
	assertStatus(t, http.MethodGet, "/api/v1/certified-pilots?page=9", w, http.StatusOK)

	// getPilot 详情 → 200；不存在 → 404
	w = doRaw(app, http.MethodGet, "/api/v1/certified-pilots/"+pilotID, "", "")
	assertStatus(t, http.MethodGet, "/api/v1/certified-pilots/"+pilotID, w, http.StatusOK)
	w = doRaw(app, http.MethodGet, "/api/v1/certified-pilots/zzz", "", "")
	assertStatus(t, http.MethodGet, "/api/v1/certified-pilots/zzz", w, http.StatusNotFound)

	// listAdminPilots：管理员全量（含 status 过滤）→ 200
	w = doRaw(app, http.MethodGet, "/api/v1/admin/certified-pilots?status=all", "", adminTok)
	assertStatus(t, http.MethodGet, "/api/v1/admin/certified-pilots", w, http.StatusOK)
}

// TestR4JobsMutationCode 覆盖 jobMutationCode 的分支：不存在职位 → 404
//（资源不存在语义；此前默认 403 会把未知错误与越权混淆）。
func TestR4JobsMutationCode(t *testing.T) {
	app := newBizServer(t)
	entTok := authAs(t, "ent-1", domain.RoleEnterprise)

	w := doRaw(app, http.MethodPost, "/api/v1/jobs/zzz/publish", "", entTok)
	assertStatus(t, http.MethodPost, "/api/v1/jobs/zzz/publish", w, http.StatusNotFound)
	w = doRaw(app, http.MethodPost, "/api/v1/jobs/zzz/close", "", entTok)
	assertStatus(t, http.MethodPost, "/api/v1/jobs/zzz/close", w, http.StatusNotFound)
}

// TestR4CommunityListings 覆盖 removePost/closeListing/favoriteListing。
func TestR4CommunityListings(t *testing.T) {
	app := newBizServer(t)
	userTok := authAs(t, "user-1", domain.RoleIndividual)

	// 帖子：创建 → 移除 → 200；不存在 → 403
	w := doRaw(app, http.MethodPost, "/api/v1/posts", `{"title":"测试帖","content":"内容"}`, userTok)
	assertStatus(t, http.MethodPost, "/api/v1/posts", w, http.StatusCreated)
	postID := dataID(t, w)
	w = doRaw(app, http.MethodPost, "/api/v1/posts/"+postID+"/remove", "", userTok)
	assertStatus(t, http.MethodPost, "/api/v1/posts/"+postID+"/remove", w, http.StatusOK)
	w = doRaw(app, http.MethodPost, "/api/v1/posts/zzz/remove", "", userTok)
	assertStatus(t, http.MethodPost, "/api/v1/posts/zzz/remove", w, http.StatusForbidden)

	// 挂牌：创建 → 收藏 → 关闭
	w = doRaw(app, http.MethodPost, "/api/v1/listings",
		`{"title":"无人机出租","description":"大疆M350","category":"rent","price_fen":500000}`, userTok)
	assertStatus(t, http.MethodPost, "/api/v1/listings", w, http.StatusCreated)
	listingID := dataID(t, w)

	w = doRaw(app, http.MethodPost, "/api/v1/listings/"+listingID+"/favorites", "", userTok)
	assertStatus(t, http.MethodPost, "/api/v1/listings/"+listingID+"/favorites", w, http.StatusOK)
	// Favorite 对不存在 listing 亦不报错（AddFavorite 幂等）
	w = doRaw(app, http.MethodPost, "/api/v1/listings/zzz/favorites", "", userTok)
	assertStatus(t, http.MethodPost, "/api/v1/listings/zzz/favorites", w, http.StatusOK)

	w = doRaw(app, http.MethodPost, "/api/v1/listings/"+listingID+"/close", "", userTok)
	assertStatus(t, http.MethodPost, "/api/v1/listings/"+listingID+"/close", w, http.StatusOK)
	w = doRaw(app, http.MethodPost, "/api/v1/listings/zzz/close", "", userTok)
	assertStatus(t, http.MethodPost, "/api/v1/listings/zzz/close", w, http.StatusForbidden)
}

// TestR4WorkOrderDetailReworkCancel 覆盖 workOrderDetail/reworkWorkOrder/cancelWorkOrder 完整闭环。
func TestR4WorkOrderDetailReworkCancel(t *testing.T) {
	app := newBizServer(t)
	entTok := authAs(t, "ent-1", domain.RoleEnterprise)
	workerTok := authAs(t, "worker-1", domain.RoleIndividual)
	strangerTok := authAs(t, "worker-2", domain.RoleIndividual)
	adminTok := authAs(t, "admin-1", domain.RolePlatformAdmin)

	// 需求 → 审核 → 意向 → 接单生成订单
	dID := r4CreateDemand(t, app, entTok, "工单整改需求")
	r4ApproveDemand(t, app, adminTok, dID)
	w := doRaw(app, http.MethodPost, "/api/v1/demands/"+dID+"/intents",
		`{"intentor_name":"飞手小王","contact":"13900000000"}`, workerTok)
	assertStatus(t, http.MethodPost, "/api/v1/demands/"+dID+"/intents", w, http.StatusCreated)
	intentID := dataID(t, w)
	w = doRaw(app, http.MethodPost, "/api/v1/demands/"+dID+"/intents/"+intentID+"/accept",
		`{"amount_fen":100000}`, entTok)
	assertStatus(t, http.MethodPost, ".../accept", w, http.StatusCreated)
	orderID := dataID(t, w)

	// workOrderDetail：双方可见 → 200
	w = doRaw(app, http.MethodGet, "/api/v1/work-orders/"+orderID, "", entTok)
	assertStatus(t, http.MethodGet, "/api/v1/work-orders/"+orderID, w, http.StatusOK)
	// 第三方不可见 → 403
	w = doRaw(app, http.MethodGet, "/api/v1/work-orders/"+orderID, "", strangerTok)
	assertStatus(t, http.MethodGet, "/api/v1/work-orders/"+orderID+" (stranger)", w, http.StatusForbidden)
	// 不存在 → 403
	w = doRaw(app, http.MethodGet, "/api/v1/work-orders/zzz", "", entTok)
	assertStatus(t, http.MethodGet, "/api/v1/work-orders/zzz", w, http.StatusForbidden)

	// 开始 → 完成 → 整改（awaiting_accept → ongoing）
	w = doRaw(app, http.MethodPost, "/api/v1/work-orders/"+orderID+"/start", "", workerTok)
	assertStatus(t, http.MethodPost, ".../start", w, http.StatusOK)
	w = doRaw(app, http.MethodPost, "/api/v1/work-orders/"+orderID+"/complete", `{"result_photos":["/uploads/a.jpg"]}`, workerTok)
	assertStatus(t, http.MethodPost, ".../complete", w, http.StatusOK)
	w = doRaw(app, http.MethodPost, "/api/v1/work-orders/"+orderID+"/rework", `{"note":"重新返工"}`, entTok)
	assertStatus(t, http.MethodPost, ".../rework", w, http.StatusOK)

	// 再次完成 → 取消（任意一方）
	w = doRaw(app, http.MethodPost, "/api/v1/work-orders/"+orderID+"/complete", `{}`, workerTok)
	assertStatus(t, http.MethodPost, ".../complete (2)", w, http.StatusOK)
	w = doRaw(app, http.MethodPost, "/api/v1/work-orders/"+orderID+"/cancel", `{"reason":"需求变更"}`, entTok)
	assertStatus(t, http.MethodPost, ".../cancel", w, http.StatusOK)

	// 不存在 → 403
	w = doRaw(app, http.MethodPost, "/api/v1/work-orders/zzz/rework", `{}`, entTok)
	assertStatus(t, http.MethodPost, ".../rework zzz", w, http.StatusForbidden)
	w = doRaw(app, http.MethodPost, "/api/v1/work-orders/zzz/cancel", `{}`, entTok)
	assertStatus(t, http.MethodPost, ".../cancel zzz", w, http.StatusForbidden)
}

// TestR4Phase3Enrollments 覆盖 enrollCourse/payAndEnroll/listEnrollments/listMyEnrollments/
// updateEnrollment/completeEnrollment/listExpiringInspections。
func TestR4Phase3Enrollments(t *testing.T) {
	app := newBizServer(t)
	userTok := authAs(t, "user-1", domain.RoleIndividual)
	user2Tok := authAs(t, "user-2", domain.RoleIndividual)
	adminTok := authAs(t, "admin-1", domain.RolePlatformAdmin)

	// 免费课程（price_fen=0，payAndEnroll 走免托管分支）
	w := doRaw(app, http.MethodPost, "/api/v1/training-courses",
		`{"title":"无人机执照培训","cert_type":"caac","org_name":"渝飞培训","price_fen":0}`, userTok)
	assertStatus(t, http.MethodPost, "/api/v1/training-courses", w, http.StatusCreated)
	courseID := dataID(t, w)

	// 课程发布默认待审核（pending）：先管理端审核通过，否则报名 404
	w = doRaw(app, http.MethodPut, "/api/v1/admin/training-courses/"+courseID,
		`{"title":"无人机执照培训","status":"published"}`, adminTok)
	assertStatus(t, http.MethodPut, "/api/v1/admin/training-courses/"+courseID, w, http.StatusOK)

	// enrollCourse → 201
	w = doRaw(app, http.MethodPost, "/api/v1/training-courses/"+courseID+"/enroll",
		`{"name":"学员A","phone":"13800000001","gender":"男"}`, userTok)
	assertStatus(t, http.MethodPost, ".../enroll", w, http.StatusCreated)
	enrollID := dataID(t, w)

	// payAndEnroll（免费课程，另一用户避免 already-enrolled）→ 201
	w = doRaw(app, http.MethodPost, "/api/v1/training-courses/"+courseID+"/pay-and-enroll",
		`{"name":"学员B","phone":"13800000002"}`, user2Tok)
	assertStatus(t, http.MethodPost, ".../pay-and-enroll", w, http.StatusCreated)

	// listEnrollments 限管理员（含 PII）；普通用户 403 / listMyEnrollments 本人可查
	w = doRaw(app, http.MethodGet, "/api/v1/training-courses/"+courseID+"/enrollments", "", userTok)
	assertStatus(t, http.MethodGet, ".../enrollments", w, http.StatusForbidden)
	w = doRaw(app, http.MethodGet, "/api/v1/training-courses/"+courseID+"/enrollments", "", adminTok)
	assertStatus(t, http.MethodGet, ".../enrollments", w, http.StatusOK)
	w = doRaw(app, http.MethodGet, "/api/v1/enrollments/mine", "", userTok)
	assertStatus(t, http.MethodGet, "/api/v1/enrollments/mine", w, http.StatusOK)

	// updateEnrollment（管理员编辑）→ 200
	w = doRaw(app, http.MethodPut, "/api/v1/admin/enrollments/"+enrollID,
		`{"name":"学员A改","phone":"13800000009","status":"approved"}`, adminTok)
	assertStatus(t, http.MethodPut, "/api/v1/admin/enrollments/"+enrollID, w, http.StatusOK)

	// updateEnrollment 不存在 → 404
	w = doRaw(app, http.MethodPut, "/api/v1/admin/enrollments/zzz", `{"name":"x","status":"approved"}`, adminTok)
	assertStatus(t, http.MethodPut, "/api/v1/admin/enrollments/zzz", w, http.StatusNotFound)

	// completeEnrollment 不存在 → 404
	w = doRaw(app, http.MethodPost, "/api/v1/enrollments/zzz/complete", "", adminTok)
	assertStatus(t, http.MethodPost, "/api/v1/enrollments/zzz/complete", w, http.StatusNotFound)

	// listExpiringInspections 限管理员——普通用户 403，管理员 200（days 默认 30）
	w = doRaw(app, http.MethodGet, "/api/v1/inspections/expiring", "", userTok)
	assertStatus(t, http.MethodGet, ".../inspections/expiring (user)", w, http.StatusForbidden)
	w = doRaw(app, http.MethodGet, "/api/v1/inspections/expiring", "", adminTok)
	assertStatus(t, http.MethodGet, ".../inspections/expiring (admin)", w, http.StatusOK)
}

// TestR4BizHandlersZero 覆盖 biz_handlers.go 的 0% 端点。
func TestR4BizHandlersZero(t *testing.T) {
	app := newBizServer(t)
	userTok := authAs(t, "enterprise-1", domain.RoleEnterprise)
	adminTok := authAs(t, "admin-1", domain.RolePlatformAdmin)

	// listMyPortfolios → 200
	w := doRaw(app, http.MethodGet, "/api/v1/portfolios/mine", "", userTok)
	assertStatus(t, http.MethodGet, "/api/v1/portfolios/mine", w, http.StatusOK)

	// listAllProjectApps（管理员）→ 200
	w = doRaw(app, http.MethodGet, "/api/v1/admin/project-applications?status=submitted", "", adminTok)
	assertStatus(t, http.MethodGet, "/api/v1/admin/project-applications", w, http.StatusOK)

	// 创建课题申报 → reviewProjectApp（approve）
	w = doRaw(app, http.MethodPost, "/api/v1/project-applications",
		`{"project_name":"低空物流课题","category":"物流","description":"课题说明"}`, userTok)
	assertStatus(t, http.MethodPost, "/api/v1/project-applications", w, http.StatusCreated)
	appID := dataID(t, w)
	w = doRaw(app, http.MethodPost, "/api/v1/admin/project-applications/"+appID+"/review",
		`{"action":"approve","review_note":"通过"}`, adminTok)
	assertStatus(t, http.MethodPost, ".../review", w, http.StatusOK)
	// reviewProjectApp 不存在 → 404
	w = doRaw(app, http.MethodPost, "/api/v1/admin/project-applications/zzz/review", `{"action":"reject"}`, adminTok)
	assertStatus(t, http.MethodPost, ".../review zzz", w, http.StatusNotFound)

	// 创建赛事 → listCompetitionRegs（管理员）→ 200
	w = doRaw(app, http.MethodPost, "/api/v1/admin/competitions",
		`{"title":"全国无人机大赛","category":"竞技"}`, adminTok)
	assertStatus(t, http.MethodPost, "/api/v1/admin/competitions", w, http.StatusCreated)
	compID := dataID(t, w)
	w = doRaw(app, http.MethodGet, "/api/v1/competitions/"+compID+"/registrations", "", adminTok)
	assertStatus(t, http.MethodGet, ".../registrations", w, http.StatusOK)

	// getIndustryResourcePublic：创建资源 → 公开详情 → 200；不存在 → 404
	w = doRaw(app, http.MethodPost, "/api/v1/admin/industry-resources",
		`{"name":"测试机库","res_type":"facility","visibility_level":"public"}`, adminTok)
	assertStatus(t, http.MethodPost, "/api/v1/admin/industry-resources", w, http.StatusCreated)
	resID := dataID(t, w)
	w = doRaw(app, http.MethodGet, "/api/v1/industry-resources/"+resID, "", "")
	assertStatus(t, http.MethodGet, "/api/v1/industry-resources/"+resID, w, http.StatusOK)
	w = doRaw(app, http.MethodGet, "/api/v1/industry-resources/zzz", "", "")
	assertStatus(t, http.MethodGet, "/api/v1/industry-resources/zzz", w, http.StatusNotFound)

	// listEmergencyDispatches（公开，含 status 过滤）
	w = doRaw(app, http.MethodGet, "/api/v1/emergency-dispatches", "", "")
	assertStatus(t, http.MethodGet, "/api/v1/emergency-dispatches", w, http.StatusOK)
	w = doRaw(app, http.MethodGet, "/api/v1/emergency-dispatches?status=pending", "", "")
	assertStatus(t, http.MethodGet, "/api/v1/emergency-dispatches?status=pending", w, http.StatusOK)
}

// TestR4ReviewsVenues 覆盖 createVenue/bookVenue/rejectReview/deleteReview/approveReview。
func TestR4ReviewsVenues(t *testing.T) {
	app := newBizServer(t)
	userTok := authAs(t, "user-1", domain.RoleIndividual)
	adminTok := authAs(t, "admin-1", domain.RolePlatformAdmin)

	// 提交评价 → 201（供 approve/delete 使用）
	w := doRaw(app, http.MethodPost, "/api/v1/reviews",
		`{"target_type":"enterprise","target_id":"ent-1","rating":5,"content":"服务专业"}`, userTok)
	assertStatus(t, http.MethodPost, "/api/v1/reviews", w, http.StatusCreated)
	reviewID := dataID(t, w)

	// approveReview → 200；deleteReview → 200
	w = doRaw(app, http.MethodPost, "/api/v1/admin/reviews/"+reviewID+"/approve", "", adminTok)
	assertStatus(t, http.MethodPost, ".../approve", w, http.StatusOK)
	w = doRaw(app, http.MethodDelete, "/api/v1/admin/reviews/"+reviewID, "", adminTok)
	assertStatus(t, http.MethodDelete, ".../delete", w, http.StatusOK)

	// rejectReview：提交新评价 → 驳回 → 200；不存在 → 403（错误非精确 "not found"）
	w = doRaw(app, http.MethodPost, "/api/v1/reviews",
		`{"target_type":"enterprise","target_id":"ent-1","rating":3,"content":"一般"}`, userTok)
	assertStatus(t, http.MethodPost, "/api/v1/reviews (2)", w, http.StatusCreated)
	review2ID := dataID(t, w)
	w = doRaw(app, http.MethodPost, "/api/v1/admin/reviews/"+review2ID+"/reject", "", adminTok)
	assertStatus(t, http.MethodPost, ".../reject", w, http.StatusOK)
	w = doRaw(app, http.MethodPost, "/api/v1/admin/reviews/zzz/reject", "", adminTok)
	assertStatus(t, http.MethodPost, ".../reject zzz", w, http.StatusForbidden)

	// 场地：创建 → 预约 → 201
	w = doRaw(app, http.MethodPost, "/api/v1/venues",
		`{"name":"无人机训练场","venue_type":"training_field","location":"渝北区","price_fen":10000}`, userTok)
	assertStatus(t, http.MethodPost, "/api/v1/venues", w, http.StatusCreated)
	venueID := dataID(t, w)
	w = doRaw(app, http.MethodPost, "/api/v1/venues/"+venueID+"/book",
		`{"start_time":"2026-08-20T09:00:00Z","end_time":"2026-08-20T11:00:00Z"}`, userTok)
	assertStatus(t, http.MethodPost, "/api/v1/venues/"+venueID+"/book", w, http.StatusCreated)
}

// TestR4ImportMembers 覆盖 importMembers 的逐行失败明细（空 user_id / 非法 role / 成功行）。
func TestR4ImportMembers(t *testing.T) {
	app := newBizServer(t)
	adminTok := authAs(t, "admin-1", domain.RolePlatformAdmin)

	body := `{"members":[
		{"user_id":"m1","enterprise_id":"e1","role":"member"},
		{"user_id":"","enterprise_id":"e1","role":"member"},
		{"user_id":"m2","enterprise_id":"e1","role":"superhero"}
	]}`
	w := doRaw(app, http.MethodPost, "/api/v1/admin/members/import", body, adminTok)
	assertStatus(t, http.MethodPost, "/api/v1/admin/members/import", w, http.StatusOK)
	if !strings.Contains(w.Body.String(), "imported") || !strings.Contains(w.Body.String(), "failed") {
		t.Fatalf("import response missing imported/failed: %s", w.Body.String())
	}

	// 空 members → 400
	w = doRaw(app, http.MethodPost, "/api/v1/admin/members/import", `{"members":[]}`, adminTok)
	assertStatus(t, http.MethodPost, "/api/v1/admin/members/import (empty)", w, http.StatusBadRequest)
}

// TestR4BatchApproveErrors 覆盖 batchApproveDemands 的空 ids 400 与逐条失败计数。
func TestR4BatchApproveErrors(t *testing.T) {
	app := newBizServer(t)
	adminTok := authAs(t, "admin-1", domain.RolePlatformAdmin)

	w := doRaw(app, http.MethodPost, "/api/v1/admin/demands/batch-approve", `{}`, adminTok)
	assertStatus(t, http.MethodPost, "/api/v1/admin/demands/batch-approve (empty)", w, http.StatusBadRequest)

	// 不存在 id → failed 计数
	w = doRaw(app, http.MethodPost, "/api/v1/admin/demands/batch-approve", `{"ids":["zzz-none"]}`, adminTok)
	assertStatus(t, http.MethodPost, "/api/v1/admin/demands/batch-approve (missing)", w, http.StatusOK)
	if !strings.Contains(w.Body.String(), `"failed":1`) {
		t.Fatalf("batch-approve missing id should report failed=1: %s", w.Body.String())
	}
}

// TestR4CompatMissingCode 覆盖 compat_routes.go 与 auth_wechat.go 的缺 code/缺 phone 分支。
func TestR4CompatMissingCode(t *testing.T) {
	setDevMode(t)
	app := newBizServer(t)

	// 缺 code → 400（两条微信登录路由）
	w := doRaw(app, http.MethodPost, "/api/auth/wechat/login", `{}`, "")
	assertStatus(t, http.MethodPost, "/api/auth/wechat/login (no code)", w, http.StatusBadRequest)
	w = doRaw(app, http.MethodPost, "/api/auth/wx-login", `{}`, "")
	assertStatus(t, http.MethodPost, "/api/auth/wx-login (no code)", w, http.StatusBadRequest)

	// wx-phone：code 分支 → 200；两者皆缺 → 400
	w = doRaw(app, http.MethodPost, "/api/auth/wx-phone", `{"code":"1234567890"}`, "")
	assertStatus(t, http.MethodPost, "/api/auth/wx-phone (code)", w, http.StatusOK)
	w = doRaw(app, http.MethodPost, "/api/auth/wx-phone", `{}`, "")
	assertStatus(t, http.MethodPost, "/api/auth/wx-phone (empty)", w, http.StatusBadRequest)
}

// TestR4H5AuthLoginErrors 覆盖 h5AuthLogin 的缺字段 400 与用户不存在 401。
func TestR4H5AuthLoginErrors(t *testing.T) {
	app := newBizServer(t)

	w := doRaw(app, http.MethodPost, "/api/auth/login", `{}`, "")
	assertStatus(t, http.MethodPost, "/api/auth/login (empty)", w, http.StatusBadRequest)

	// 用户不存在（DB + users.json 均无）→ 401
	w = doRaw(app, http.MethodPost, "/api/auth/login", `{"phone":"19999999999","password":"x"}`, "")
	assertStatus(t, http.MethodPost, "/api/auth/login (missing user)", w, http.StatusUnauthorized)
}

// TestR4H5AdminAndUsers 覆盖 h5_compat.go 的管理端/用户 JSON 路由（dev-only）。
func TestR4H5AdminAndUsers(t *testing.T) {
	setDevMode(t)
	app := newBizServer(t)
	token := authAs(t, "user-1", domain.RoleIndividual)

	// h5UpdateApplication：不存在 → 404；非法 JSON → 400
	w := doRaw(app, http.MethodPost, "/api/update", `{"id":"r4-nope","status":"done"}`, token)
	assertStatus(t, http.MethodPost, "/api/update", w, http.StatusNotFound)
	w = doRaw(app, http.MethodPost, "/api/update", `not-json`, token)
	assertStatus(t, http.MethodPost, "/api/update (bad json)", w, http.StatusBadRequest)

	// h5ExportApplications → 200
	w = doRaw(app, http.MethodGet, "/api/export", "", token)
	assertStatus(t, http.MethodGet, "/api/export", w, http.StatusOK)

	// h5AdminStats / h5AdminApplications / h5AdminApplicationByID / h5AdminUpdateApplication
	w = doRaw(app, http.MethodGet, "/api/admin/stats", "", token)
	assertStatus(t, http.MethodGet, "/api/admin/stats", w, http.StatusOK)
	w = doRaw(app, http.MethodGet, "/api/admin/applications", "", token)
	assertStatus(t, http.MethodGet, "/api/admin/applications", w, http.StatusOK)
	w = doRaw(app, http.MethodGet, "/api/admin/applications/r4-nope", "", token)
	assertStatus(t, http.MethodGet, "/api/admin/applications/r4-nope", w, http.StatusNotFound)
	w = doRaw(app, http.MethodPost, "/api/admin/applications/r4-nope", `{"status":"done"}`, token)
	assertStatus(t, http.MethodPost, "/api/admin/applications/r4-nope", w, http.StatusNotFound)

	// h5Users / h5UpdateUserRole / h5UpdateUserProfile
	w = doRaw(app, http.MethodGet, "/api/users", "", token)
	assertStatus(t, http.MethodGet, "/api/users", w, http.StatusOK)
	w = doRaw(app, http.MethodPost, "/api/user/role", `{"id":"r4-nope","role":"admin"}`, token)
	assertStatus(t, http.MethodPost, "/api/user/role", w, http.StatusNotFound)
	w = doRaw(app, http.MethodPost, "/api/user/update", `{"id":"r4-nope","name":"x"}`, token)
	assertStatus(t, http.MethodPost, "/api/user/update", w, http.StatusNotFound)

	// h5SSOVerify → 400
	w = doRaw(app, http.MethodPost, "/api/sso/verify", `{}`, token)
	assertStatus(t, http.MethodPost, "/api/sso/verify", w, http.StatusBadRequest)
}

// TestR4H5ReviewsShowcaseCategories 覆盖 reviews/showcase/categories 的 JSON 路由。
func TestR4H5ReviewsShowcaseCategories(t *testing.T) {
	setDevMode(t)
	app := newBizServer(t)
	token := authAs(t, "user-1", domain.RoleIndividual)

	// h5Reviews（含 section 过滤）
	w := doRaw(app, http.MethodGet, "/api/reviews", "", token)
	assertStatus(t, http.MethodGet, "/api/reviews", w, http.StatusOK)
	w = doRaw(app, http.MethodGet, "/api/reviews?section=training", "", token)
	assertStatus(t, http.MethodGet, "/api/reviews?section", w, http.StatusOK)

	// h5ReviewsSubmit → 200
	w = doRaw(app, http.MethodPost, "/api/reviews", `{"section":"training","content":"好评"}`, token)
	assertStatus(t, http.MethodPost, "/api/reviews", w, http.StatusOK)

	w = doRaw(app, http.MethodGet, "/api/reviews/courses", "", token)
	assertStatus(t, http.MethodGet, "/api/reviews/courses", w, http.StatusOK)
	w = doRaw(app, http.MethodGet, "/api/admin/reviews", "", token)
	assertStatus(t, http.MethodGet, "/api/admin/reviews", w, http.StatusOK)
	w = doRaw(app, http.MethodPost, "/api/admin/reviews/r4-nope", `{"status":"approved"}`, token)
	assertStatus(t, http.MethodPost, "/api/admin/reviews/r4-nope", w, http.StatusNotFound)
	w = doRaw(app, http.MethodDelete, "/api/admin/reviews/r4-nope", "", token)
	assertStatus(t, http.MethodDelete, "/api/admin/reviews/r4-nope", w, http.StatusNotFound)

	// h5StudyShowcase / h5StudyShowcaseSave
	w = doRaw(app, http.MethodGet, "/api/study/showcase", "", token)
	assertStatus(t, http.MethodGet, "/api/study/showcase", w, http.StatusOK)
	w = doRaw(app, http.MethodPost, "/api/study/showcase", `{"showcase":[{"title":"成果展"}]}`, token)
	assertStatus(t, http.MethodPost, "/api/study/showcase", w, http.StatusOK)

	// 分类：创建 → 更新 → 删除（用返回的 id 闭环）
	w = doRaw(app, http.MethodPost, "/api/case-categories/create", `{"name":"新分类","service":"测试服务"}`, token)
	assertStatus(t, http.MethodPost, "/api/case-categories/create", w, http.StatusOK)
	catID := jsonFloatID(t, w, "id")
	w = doRaw(app, http.MethodPost, "/api/case-categories/update", `{"id":`+catID+`,"name":"新分类改"}`, token)
	assertStatus(t, http.MethodPost, "/api/case-categories/update", w, http.StatusOK)
	w = doRaw(app, http.MethodPost, "/api/case-categories/delete", `{"id":`+catID+`}`, token)
	assertStatus(t, http.MethodPost, "/api/case-categories/delete", w, http.StatusOK)
	// 更新/删除不存在 → 404
	w = doRaw(app, http.MethodPost, "/api/case-categories/update", `{"id":999999,"name":"x"}`, token)
	assertStatus(t, http.MethodPost, "/api/case-categories/update zzz", w, http.StatusNotFound)
	w = doRaw(app, http.MethodPost, "/api/case-categories/delete", `{"id":999999}`, token)
	assertStatus(t, http.MethodPost, "/api/case-categories/delete zzz", w, http.StatusNotFound)

	// h5Cases 分类/关键词过滤分支
	w = doRaw(app, http.MethodGet, "/api/cases?categoryId=1&keyword=巡检", "", token)
	assertStatus(t, http.MethodGet, "/api/cases?categoryId&keyword", w, http.StatusOK)
}

// TestR4VoidContract 覆盖 voidContract 的 404/409 错误分支。
func TestR4VoidContract(t *testing.T) {
	app := newBizServer(t)
	entTok := authAs(t, "ent-1", domain.RoleEnterprise)

	// 不存在合同 → 404
	w := doRaw(app, http.MethodPost, "/api/v1/contracts/zzz/void", "", entTok)
	assertStatus(t, http.MethodPost, "/api/v1/contracts/zzz/void", w, http.StatusNotFound)

	// 新建草稿合同 → void 触发状态机校验（draft→voided 非法）→ 409
	w = doRaw(app, http.MethodPost, "/api/v1/contracts", `{"enterprise_id":"ent-1","template_id":"tpl-1"}`, entTok)
	assertStatus(t, http.MethodPost, "/api/v1/contracts", w, http.StatusCreated)
	cid := dataID(t, w)
	w = doRaw(app, http.MethodPost, "/api/v1/contracts/"+cid+"/void", "", entTok)
	assertStatus(t, http.MethodPost, "/api/v1/contracts/"+cid+"/void", w, http.StatusConflict)
}

// TestR4ListMyRepairs 覆盖 trading_insurance_finance.go 的 listMyRepairs。
func TestR4ListMyRepairs(t *testing.T) {
	app := newBizServer(t)
	userTok := authAs(t, "user-1", domain.RoleIndividual)

	// 创建维修单 → mine 列表 → 200
	w := doRaw(app, http.MethodPost, "/api/v1/repairs", `{"product_desc":"M350","fault_desc":"云台抖动"}`, userTok)
	assertStatus(t, http.MethodPost, "/api/v1/repairs", w, http.StatusCreated)
	w = doRaw(app, http.MethodGet, "/api/v1/repairs/mine", "", userTok)
	assertStatus(t, http.MethodGet, "/api/v1/repairs/mine", w, http.StatusOK)
}

// TestR4ServeImagePng 覆盖 serveImage 的 png 输出与 width/quality 边界分支。
func TestR4ServeImagePng(t *testing.T) {
	app := newBizServer(t)

	const name = "r4-png-test.png"
	if err := os.MkdirAll("uploads", 0755); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(filepath.Join("uploads", name))
	var pngBuf bytes.Buffer
	if err := png.Encode(&pngBuf, image.NewRGBA(image.Rect(0, 0, 64, 48))); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join("uploads", name), pngBuf.Bytes(), 0644); err != nil {
		t.Fatal(err)
	}

	// png 输出 → 200 + image/png（覆盖 png.Encode 分支）
	w := doRaw(app, http.MethodGet, "/api/v1/image?url="+name+"&width=32&quality=70&format=png", "", "")
	if w.Code != http.StatusOK {
		t.Fatalf("png serve: %d %s", w.Code, w.Body.String())
	}
	if ct := w.Header().Get("Content-Type"); ct != "image/png" {
		t.Fatalf("content-type = %q", ct)
	}

	// 边界参数：width<=0 默认 800；quality>100 默认 75；up-scaling 分支（width > 原宽）
	w = doRaw(app, http.MethodGet, "/api/v1/image?url="+name+"&width=-1&quality=200&format=jpeg", "", "")
	if w.Code != http.StatusOK {
		t.Fatalf("boundary params: %d %s", w.Code, w.Body.String())
	}
	w = doRaw(app, http.MethodGet, "/api/v1/image?url="+name+"&width=800&format=jpeg", "", "")
	if w.Code != http.StatusOK {
		t.Fatalf("upscale branch: %d %s", w.Code, w.Body.String())
	}

	// 清理缓存文件
	for _, p := range []string{filepath.Join("uploads", ".image-cache", "*.png"), filepath.Join("uploads", ".image-cache", "*.jpeg")} {
		entries, _ := filepath.Glob(p)
		for _, e := range entries {
			os.Remove(e)
		}
	}
}

// TestR4AdminDevLogin 覆盖 admin_handler.go 的 adminDevLogin（dev 令牌签发）。
func TestR4AdminDevLogin(t *testing.T) {
	setDevMode(t)
	app := newBizServer(t)

	w := doRaw(app, http.MethodPost, "/api/v1/admin/token", `{"role":"platform_admin"}`, "")
	assertStatus(t, http.MethodPost, "/api/v1/admin/token", w, http.StatusOK)
	if !strings.Contains(w.Body.String(), "access_token") {
		t.Fatalf("admin token missing access_token: %s", w.Body.String())
	}

	// 非法 role 回落 platform_admin
	w = doRaw(app, http.MethodPost, "/api/v1/admin/token", `{"role":"superhero"}`, "")
	assertStatus(t, http.MethodPost, "/api/v1/admin/token (bad role)", w, http.StatusOK)
}
