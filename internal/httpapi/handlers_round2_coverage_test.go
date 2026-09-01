package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"drone-platform/internal/domain"
)

// assertStatus 断言响应状态码；失败时输出 method/path/code/body 前 200 字符。
func assertStatus(t *testing.T, method, path string, w *httptest.ResponseRecorder, want int) {
	t.Helper()
	if w.Code != want {
		t.Fatalf("%s %s: code=%d want=%d body=%.200s", method, path, w.Code, want, w.Body.String())
	}
}

// dataID 从响应体 {data:{id}} 中提取 id。
func dataID(t *testing.T, w *httptest.ResponseRecorder) string {
	t.Helper()
	var resp struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse data.id: %v (body=%.200s)", err, w.Body.String())
	}
	if resp.Data.ID == "" {
		t.Fatalf("missing data.id (body=%.200s)", w.Body.String())
	}
	return resp.Data.ID
}

// TestPublicHandlers 覆盖 public_handlers.go 的 10 个公开 GET 端点（匿名可访问）。
func TestPublicHandlers(t *testing.T) {
	app := newBizServer(t)

	paths := []string{
		"/api/v1/training/courses",
		"/api/v1/training/certificates",
		"/api/v1/study/tours",
		"/api/v1/rd/challenges",
		"/api/v1/research/projects",
		"/api/v1/test/sites",
		"/api/v1/emergency/resources",
		"/api/v1/industry/reports",
		"/api/v1/industry/resources",
		"/api/v1/services",
	}
	for _, p := range paths {
		w := doRaw(app, http.MethodGet, p, "", "")
		assertStatus(t, http.MethodGet, p, w, http.StatusOK)
	}
}

// TestExportCSV 覆盖 export_handler.go 的两个 CSV 导出 + 匿名 401。
func TestExportCSV(t *testing.T) {
	app := newBizServer(t)
	adminTok := authAs(t, "admin-1", domain.RolePlatformAdmin)

	w := doRaw(app, http.MethodGet, "/api/v1/admin/export/demands", "", adminTok)
	assertStatus(t, http.MethodGet, "/api/v1/admin/export/demands", w, http.StatusOK)
	if ct := w.Header().Get("Content-Type"); !strings.Contains(ct, "csv") {
		t.Fatalf("GET export demands: Content-Type=%q want csv (body=%.200s)", ct, w.Body.String())
	}

	w = doRaw(app, http.MethodGet, "/api/v1/admin/export/enterprises", "", adminTok)
	assertStatus(t, http.MethodGet, "/api/v1/admin/export/enterprises", w, http.StatusOK)
	if ct := w.Header().Get("Content-Type"); !strings.Contains(ct, "csv") {
		t.Fatalf("GET export enterprises: Content-Type=%q want csv (body=%.200s)", ct, w.Body.String())
	}

	// 匿名访问 admin 导出 → 401
	w = doRaw(app, http.MethodGet, "/api/v1/admin/export/demands", "", "")
	assertStatus(t, http.MethodGet, "/api/v1/admin/export/demands (anon)", w, http.StatusUnauthorized)
}

// TestJobsEndpoints 覆盖 jobs.go 的职位/简历/投递闭环 + 非企业发职位 403。
func TestJobsEndpoints(t *testing.T) {
	app := newBizServer(t)
	entTok := authAs(t, "ent-1", domain.RoleEnterprise)
	indTok := authAs(t, "user-2", domain.RoleIndividual)

	// 职位：创建 → 发布 → 关闭
	w := doRaw(app, http.MethodPost, "/api/v1/jobs",
		`{"title":"无人机飞手","description":"电力巡检","location":"重庆","salary_fen":800000}`, entTok)
	assertStatus(t, http.MethodPost, "/api/v1/jobs", w, http.StatusCreated)
	jobID := dataID(t, w)

	w = doRaw(app, http.MethodPost, "/api/v1/jobs/"+jobID+"/publish", "", entTok)
	assertStatus(t, http.MethodPost, "/api/v1/jobs/"+jobID+"/publish", w, http.StatusOK)

	w = doRaw(app, http.MethodPost, "/api/v1/jobs/"+jobID+"/close", "", entTok)
	assertStatus(t, http.MethodPost, "/api/v1/jobs/"+jobID+"/close", w, http.StatusOK)

	w = doRaw(app, http.MethodGet, "/api/v1/jobs", "", "")
	assertStatus(t, http.MethodGet, "/api/v1/jobs", w, http.StatusOK)

	w = doRaw(app, http.MethodGet, "/api/v1/jobs/mine", "", entTok)
	assertStatus(t, http.MethodGet, "/api/v1/jobs/mine", w, http.StatusOK)

	// 非企业（individual）发职位 → 403
	w = doRaw(app, http.MethodPost, "/api/v1/jobs", `{"title":"越权职位"}`, indTok)
	assertStatus(t, http.MethodPost, "/api/v1/jobs (individual)", w, http.StatusForbidden)

	// 简历：创建 → 更新 → mine
	w = doRaw(app, http.MethodPost, "/api/v1/resumes",
		`{"title":"飞手简历","name":"张三","phone":"13800000000","skills":["飞控","测绘"]}`, indTok)
	assertStatus(t, http.MethodPost, "/api/v1/resumes", w, http.StatusCreated)
	resumeID := dataID(t, w)

	w = doRaw(app, http.MethodPatch, "/api/v1/resumes/"+resumeID, `{"title":"更新后的简历"}`, indTok)
	assertStatus(t, http.MethodPatch, "/api/v1/resumes/"+resumeID, w, http.StatusOK)

	w = doRaw(app, http.MethodGet, "/api/v1/resumes/mine", "", indTok)
	assertStatus(t, http.MethodGet, "/api/v1/resumes/mine", w, http.StatusOK)

	// 投递：企业发第二个已发布职位 → 个人投递 → 企业改状态 → 列表
	w = doRaw(app, http.MethodPost, "/api/v1/jobs", `{"title":"测绘飞手","location":"成都"}`, entTok)
	assertStatus(t, http.MethodPost, "/api/v1/jobs (job2)", w, http.StatusCreated)
	job2ID := dataID(t, w)
	w = doRaw(app, http.MethodPost, "/api/v1/jobs/"+job2ID+"/publish", "", entTok)
	assertStatus(t, http.MethodPost, "/api/v1/jobs/"+job2ID+"/publish", w, http.StatusOK)

	w = doRaw(app, http.MethodPost, "/api/v1/applications",
		`{"job_id":"`+job2ID+`","resume_id":"`+resumeID+`"}`, indTok)
	assertStatus(t, http.MethodPost, "/api/v1/applications", w, http.StatusCreated)
	appID := dataID(t, w)

	// 投递状态机：submitted → viewed → interviewing（企业按序推进）
	w = doRaw(app, http.MethodPatch, "/api/v1/applications/"+appID+"/status", `{"status":"viewed"}`, entTok)
	assertStatus(t, http.MethodPatch, "/api/v1/applications/"+appID+"/status", w, http.StatusOK)

	w = doRaw(app, http.MethodPatch, "/api/v1/applications/"+appID+"/status", `{"status":"interviewing"}`, entTok)
	assertStatus(t, http.MethodPatch, "/api/v1/applications/"+appID+"/status", w, http.StatusOK)

	// 非法跳变（求职者自改 offered）应被拒绝
	w = doRaw(app, http.MethodPatch, "/api/v1/applications/"+appID+"/status", `{"status":"offered"}`, indTok)
	assertStatus(t, http.MethodPatch, "/api/v1/applications/"+appID+"/status", w, http.StatusForbidden)

	w = doRaw(app, http.MethodGet, "/api/v1/applications", "", entTok)
	assertStatus(t, http.MethodGet, "/api/v1/applications", w, http.StatusOK)

	w = doRaw(app, http.MethodGet, "/api/v1/applications?job_id="+job2ID, "", entTok)
	assertStatus(t, http.MethodGet, "/api/v1/applications?job_id="+job2ID, w, http.StatusOK)
}

// TestTrainingEndpoints 覆盖 training.go 的证书/课程/讲师/飞手认证端点。
func TestTrainingEndpoints(t *testing.T) {
	app := newBizServer(t)
	userTok := authAs(t, "user-1", domain.RoleIndividual)
	adminTok := authAs(t, "admin-1", domain.RolePlatformAdmin)

	// 证书：创建 → mine
	w := doRaw(app, http.MethodPost, "/api/v1/certificates",
		`{"cert_type":"caac","cert_number":"C-001","level":"三级","issuer_org":"CAAC"}`, userTok)
	assertStatus(t, http.MethodPost, "/api/v1/certificates", w, http.StatusCreated)

	w = doRaw(app, http.MethodGet, "/api/v1/certificates/mine", "", userTok)
	assertStatus(t, http.MethodGet, "/api/v1/certificates/mine", w, http.StatusOK)

	// 课程：创建 → 列表
	w = doRaw(app, http.MethodPost, "/api/v1/training-courses",
		`{"title":"无人机驾驶培训","cert_type":"caac","org_name":"渝飞培训"}`, userTok)
	assertStatus(t, http.MethodPost, "/api/v1/training-courses", w, http.StatusCreated)

	w = doRaw(app, http.MethodGet, "/api/v1/training-courses", "", "")
	assertStatus(t, http.MethodGet, "/api/v1/training-courses", w, http.StatusOK)

	// 讲师：注册 → 管理员通过
	w = doRaw(app, http.MethodPost, "/api/v1/instructors",
		`{"name":"李教练","bio":"十年飞控经验","cert_types":["caac"]}`, userTok)
	assertStatus(t, http.MethodPost, "/api/v1/instructors", w, http.StatusCreated)
	instructorID := dataID(t, w)

	w = doRaw(app, http.MethodPost, "/api/v1/admin/instructors/"+instructorID+"/approve", "", adminTok)
	assertStatus(t, http.MethodPost, "/api/v1/admin/instructors/"+instructorID+"/approve", w, http.StatusOK)

	// 飞手认证：注册（缺 id_card → 400）→ 正常注册 → 通过 → 公开名录 → mine
	w = doRaw(app, http.MethodPost, "/api/v1/certified-pilots", `{"real_name":"王飞手"}`, userTok)
	assertStatus(t, http.MethodPost, "/api/v1/certified-pilots (missing id_card)", w, http.StatusBadRequest)

	// 新规则：申请飞手需至少一张已通过且未过期的证书——先为用户签发并批准一张
	w = doRaw(app, http.MethodPost, "/api/v1/certificates",
		`{"cert_type":"caac","cert_number":"CAAC-ROUND2-001","level":"III","issuer_org":"民航局","issue_date":"2026-01-01T00:00:00Z","expire_date":"2028-01-01T00:00:00Z"}`, userTok)
	assertStatus(t, http.MethodPost, "/api/v1/certificates", w, http.StatusCreated)
	certID := dataID(t, w)
	w = doRaw(app, http.MethodPost, "/api/v1/admin/certificates/"+certID+"/approve", "", adminTok)
	assertStatus(t, http.MethodPost, "/api/v1/admin/certificates/"+certID+"/approve", w, http.StatusOK)

	w = doRaw(app, http.MethodPost, "/api/v1/certified-pilots",
		`{"real_name":"王飞手","id_card":"110101199001011234","flight_hours":120}`, userTok)
	assertStatus(t, http.MethodPost, "/api/v1/certified-pilots", w, http.StatusCreated)
	pilotID := dataID(t, w)

	w = doRaw(app, http.MethodPost, "/api/v1/admin/certified-pilots/"+pilotID+"/approve", "", adminTok)
	assertStatus(t, http.MethodPost, "/api/v1/admin/certified-pilots/"+pilotID+"/approve", w, http.StatusOK)

	w = doRaw(app, http.MethodGet, "/api/v1/certified-pilots", "", "")
	assertStatus(t, http.MethodGet, "/api/v1/certified-pilots", w, http.StatusOK)

	w = doRaw(app, http.MethodGet, "/api/v1/certified-pilots/mine", "", userTok)
	assertStatus(t, http.MethodGet, "/api/v1/certified-pilots/mine", w, http.StatusOK)

	// 驳回分支：另一用户注册飞手 → 管理员驳回
	user2Tok := authAs(t, "user-2", domain.RoleIndividual)
	// user-2 同样需要一张已通过证书
	w = doRaw(app, http.MethodPost, "/api/v1/certificates",
		`{"cert_type":"caac","cert_number":"CAAC-ROUND2-002","level":"III","issuer_org":"民航局"}`, user2Tok)
	assertStatus(t, http.MethodPost, "/api/v1/certificates (user-2)", w, http.StatusCreated)
	cert2ID := dataID(t, w)
	w = doRaw(app, http.MethodPost, "/api/v1/admin/certificates/"+cert2ID+"/approve", "", adminTok)
	assertStatus(t, http.MethodPost, "/api/v1/admin/certificates/"+cert2ID+"/approve", w, http.StatusOK)

	w = doRaw(app, http.MethodPost, "/api/v1/certified-pilots",
		`{"real_name":"李飞手","id_card":"110101199002022345"}`, user2Tok)
	assertStatus(t, http.MethodPost, "/api/v1/certified-pilots (user-2)", w, http.StatusCreated)
	pilot2ID := dataID(t, w)

	w = doRaw(app, http.MethodPost, "/api/v1/admin/certified-pilots/"+pilot2ID+"/reject",
		`{"reason":"证书信息不完整"}`, adminTok)
	assertStatus(t, http.MethodPost, "/api/v1/admin/certified-pilots/"+pilot2ID+"/reject", w, http.StatusOK)
}

// TestReviewsEndpoints 覆盖 reviews_resources.go 的评价提交/列表/管理端审批。
func TestReviewsEndpoints(t *testing.T) {
	app := newBizServer(t)
	userTok := authAs(t, "user-1", domain.RoleIndividual)
	adminTok := authAs(t, "admin-1", domain.RolePlatformAdmin)

	// 评分越界 → 400
	w := doRaw(app, http.MethodPost, "/api/v1/reviews",
		`{"target_type":"enterprise","target_id":"ent-1","rating":0,"content":"差"}`, userTok)
	assertStatus(t, http.MethodPost, "/api/v1/reviews (rating 0)", w, http.StatusBadRequest)

	// 正常提交 → 201
	w = doRaw(app, http.MethodPost, "/api/v1/reviews",
		`{"target_type":"enterprise","target_id":"ent-1","rating":5,"content":"服务专业"}`, userTok)
	assertStatus(t, http.MethodPost, "/api/v1/reviews", w, http.StatusCreated)
	reviewID := dataID(t, w)

	w = doRaw(app, http.MethodGet, "/api/v1/reviews?target_type=enterprise&target_id=ent-1", "", "")
	assertStatus(t, http.MethodGet, "/api/v1/reviews", w, http.StatusOK)

	w = doRaw(app, http.MethodGet, "/api/v1/admin/reviews", "", adminTok)
	assertStatus(t, http.MethodGet, "/api/v1/admin/reviews", w, http.StatusOK)

	w = doRaw(app, http.MethodPost, "/api/v1/admin/reviews/"+reviewID+"/approve", "", adminTok)
	assertStatus(t, http.MethodPost, "/api/v1/admin/reviews/"+reviewID+"/approve", w, http.StatusOK)
}

// TestDemandFulfillment 覆盖 demand_fulfillment.go 的重提/完成/取消 + 批量审批。
func TestDemandFulfillment(t *testing.T) {
	app := newBizServer(t)
	entTok := authAs(t, "ent-1", domain.RoleEnterprise)
	adminTok := authAs(t, "admin-1", domain.RolePlatformAdmin)

	createDemand := func(title string) string {
		t.Helper()
		w := doRaw(app, http.MethodPost, "/api/v1/demands",
			`{"title":"`+title+`","contact":"13800000000","biz_type":"cable_inspection"}`, entTok)
		assertStatus(t, http.MethodPost, "/api/v1/demands", w, http.StatusCreated)
		return dataID(t, w)
	}

	// 1. 驳回后重提：建需求 → 管理员驳回 → 发布者重新提交 → 200
	d1 := createDemand("驳回重提需求")
	w := doRaw(app, http.MethodPost, "/api/v1/admin/demands/"+d1+"/review",
		`{"action":"reject","reason":"信息不全"}`, adminTok)
	assertStatus(t, http.MethodPost, "/api/v1/admin/demands/"+d1+"/review", w, http.StatusOK)

	w = doRaw(app, http.MethodPost, "/api/v1/demands/"+d1+"/submit", "", entTok)
	assertStatus(t, http.MethodPost, "/api/v1/demands/"+d1+"/submit", w, http.StatusOK)

	// 2. 完成：建需求 → 审批通过 → 发布者标记完成 → 200
	d2 := createDemand("完成需求")
	w = doRaw(app, http.MethodPost, "/api/v1/admin/demands/"+d2+"/review", `{"action":"approve"}`, adminTok)
	assertStatus(t, http.MethodPost, "/api/v1/admin/demands/"+d2+"/review", w, http.StatusOK)

	w = doRaw(app, http.MethodPost, "/api/v1/demands/"+d2+"/complete", "", entTok)
	assertStatus(t, http.MethodPost, "/api/v1/demands/"+d2+"/complete", w, http.StatusOK)

	// 3. 取消：建需求（pending）→ 发布者取消 → 200
	d3 := createDemand("取消需求")
	w = doRaw(app, http.MethodPost, "/api/v1/demands/"+d3+"/cancel", "", entTok)
	assertStatus(t, http.MethodPost, "/api/v1/demands/"+d3+"/cancel", w, http.StatusOK)

	// 4. 批量审批：两条需求一次性通过 → 200
	d4 := createDemand("批量审批需求A")
	d5 := createDemand("批量审批需求B")
	w = doRaw(app, http.MethodPost, "/api/v1/admin/demands/batch-approve",
		`{"ids":["`+d4+`","`+d5+`"]}`, adminTok)
	assertStatus(t, http.MethodPost, "/api/v1/admin/demands/batch-approve", w, http.StatusOK)
}

// TestIntentsEndpoints 覆盖 intent.go 的意向登记/列表 + 接单/拒单。
func TestIntentsEndpoints(t *testing.T) {
	app := newBizServer(t)
	entTok := authAs(t, "ent-1", domain.RoleEnterprise)
	workerTok := authAs(t, "worker-1", domain.RoleIndividual)
	adminTok := authAs(t, "admin-1", domain.RolePlatformAdmin)

	// 建需求并审批通过
	w := doRaw(app, http.MethodPost, "/api/v1/demands",
		`{"title":"巡检需求","contact":"13800000000","biz_type":"cable_inspection"}`, entTok)
	assertStatus(t, http.MethodPost, "/api/v1/demands", w, http.StatusCreated)
	demandID := dataID(t, w)

	w = doRaw(app, http.MethodPost, "/api/v1/admin/demands/"+demandID+"/review", `{"action":"approve"}`, adminTok)
	assertStatus(t, http.MethodPost, "/api/v1/admin/demands/"+demandID+"/review", w, http.StatusOK)

	// 飞手登记意向 → 201
	w = doRaw(app, http.MethodPost, "/api/v1/demands/"+demandID+"/intents",
		`{"intentor_name":"飞手小王","contact":"13900000000","remark":"可接"}`, workerTok)
	assertStatus(t, http.MethodPost, "/api/v1/demands/"+demandID+"/intents", w, http.StatusCreated)
	intentID := dataID(t, w)

	// 发布方查看意向列表 → 200
	w = doRaw(app, http.MethodGet, "/api/v1/demands/"+demandID+"/intents", "", entTok)
	assertStatus(t, http.MethodGet, "/api/v1/demands/"+demandID+"/intents", w, http.StatusOK)

	// 意向方查看我的意向 → 200
	w = doRaw(app, http.MethodGet, "/api/v1/intents/mine", "", workerTok)
	assertStatus(t, http.MethodGet, "/api/v1/intents/mine", w, http.StatusOK)

	// 企业确认接单 → 201（acceptIntent 实际返回 StatusCreated）
	w = doRaw(app, http.MethodPost, "/api/v1/demands/"+demandID+"/intents/"+intentID+"/accept",
		`{"amount_fen":100000}`, entTok)
	assertStatus(t, http.MethodPost, "/api/v1/demands/"+demandID+"/intents/"+intentID+"/accept", w, http.StatusCreated)

	// 拒绝分支：新建需求 + 意向 → 企业拒绝 → 200
	w = doRaw(app, http.MethodPost, "/api/v1/demands",
		`{"title":"测绘需求","contact":"13800000000","biz_type":"cable_inspection"}`, entTok)
	assertStatus(t, http.MethodPost, "/api/v1/demands (demand2)", w, http.StatusCreated)
	demand2ID := dataID(t, w)
	w = doRaw(app, http.MethodPost, "/api/v1/admin/demands/"+demand2ID+"/review", `{"action":"approve"}`, adminTok)
	assertStatus(t, http.MethodPost, "/api/v1/admin/demands/"+demand2ID+"/review", w, http.StatusOK)

	w = doRaw(app, http.MethodPost, "/api/v1/demands/"+demand2ID+"/intents",
		`{"intentor_name":"飞手小李","contact":"13900000001"}`, workerTok)
	assertStatus(t, http.MethodPost, "/api/v1/demands/"+demand2ID+"/intents", w, http.StatusCreated)
	intent2ID := dataID(t, w)

	w = doRaw(app, http.MethodPost, "/api/v1/demands/"+demand2ID+"/intents/"+intent2ID+"/reject", "", entTok)
	assertStatus(t, http.MethodPost, "/api/v1/demands/"+demand2ID+"/intents/"+intent2ID+"/reject", w, http.StatusOK)

	// 取消登记分支：新建需求 + 意向 → 意向方取消 → 200；重复取消 → 403
	w = doRaw(app, http.MethodPost, "/api/v1/demands",
		`{"title":"航拍需求","contact":"13800000000","biz_type":"cable_inspection"}`, entTok)
	assertStatus(t, http.MethodPost, "/api/v1/demands (demand3)", w, http.StatusCreated)
	demand3ID := dataID(t, w)
	w = doRaw(app, http.MethodPost, "/api/v1/admin/demands/"+demand3ID+"/review", `{"action":"approve"}`, adminTok)
	assertStatus(t, http.MethodPost, "/api/v1/admin/demands/"+demand3ID+"/review", w, http.StatusOK)

	w = doRaw(app, http.MethodPost, "/api/v1/demands/"+demand3ID+"/intents",
		`{"intentor_name":"飞手小赵","contact":"13900000002"}`, workerTok)
	assertStatus(t, http.MethodPost, "/api/v1/demands/"+demand3ID+"/intents", w, http.StatusCreated)
	intent3ID := dataID(t, w)

	w = doRaw(app, http.MethodPost, "/api/v1/intents/"+intent3ID+"/cancel", "", workerTok)
	assertStatus(t, http.MethodPost, "/api/v1/intents/"+intent3ID+"/cancel", w, http.StatusOK)

	// 已取消 → 再取消 403（已处理）；发布方视角该意向 closed
	w = doRaw(app, http.MethodPost, "/api/v1/intents/"+intent3ID+"/cancel", "", workerTok)
	assertStatus(t, http.MethodPost, "/api/v1/intents/"+intent3ID+"/cancel (again)", w, http.StatusForbidden)
	w = doRaw(app, http.MethodGet, "/api/v1/demands/"+demand3ID+"/intents", "", entTok)
	if !strings.Contains(w.Body.String(), `"closed"`) {
		t.Fatalf("cancelled intent should be closed, got: %s", w.Body.String())
	}

	// 非本人取消他人意向 → 403
	otherTok := authAs(t, "worker-2", domain.RoleIndividual)
	w = doRaw(app, http.MethodPost, "/api/v1/intents/"+intent2ID+"/cancel", "", otherTok)
	assertStatus(t, http.MethodPost, "/api/v1/intents/"+intent2ID+"/cancel (other)", w, http.StatusForbidden)
}

// TestCompatRoutes 覆盖 compat_routes.go 的旧版 /api/auth/* 兼容路由（dev 模式）。
func TestCompatRoutes(t *testing.T) {
	oldDev := os.Getenv("ADMIN_DEV_MODE")
	oldAppID := os.Getenv("WECHAT_APPID")
	oldSecret := os.Getenv("WECHAT_APPSECRET")
	os.Setenv("ADMIN_DEV_MODE", "true")
	os.Unsetenv("WECHAT_APPID")
	os.Unsetenv("WECHAT_APPSECRET")
	t.Cleanup(func() {
		os.Setenv("ADMIN_DEV_MODE", oldDev)
		os.Setenv("WECHAT_APPID", oldAppID)
		os.Setenv("WECHAT_APPSECRET", oldSecret)
	})

	app := newBizServer(t)

	w := doRaw(app, http.MethodPost, "/api/auth/wechat/login", `{"code":"any-code"}`, "")
	assertStatus(t, http.MethodPost, "/api/auth/wechat/login", w, http.StatusOK)

	w = doRaw(app, http.MethodPost, "/api/auth/wx-login", `{"code":"any-code"}`, "")
	assertStatus(t, http.MethodPost, "/api/auth/wx-login", w, http.StatusOK)

	w = doRaw(app, http.MethodPost, "/api/auth/wx-phone", `{"phone":"13800000000"}`, "")
	assertStatus(t, http.MethodPost, "/api/auth/wx-phone", w, http.StatusOK)
}
