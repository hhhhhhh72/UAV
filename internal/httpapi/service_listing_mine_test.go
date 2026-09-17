package httpapi_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"drone-platform/internal/domain"
)

// 用户自助发布服务能力：待审核（pending）、不进公开列表、mine=1 可见、匿名 mine 为空。
func TestCreateServiceListingPendingAndMine(t *testing.T) {
	app := newBizServer(t)

	// 1) 用户发布服务能力 → 201 + status=pending
	body := []byte(`{"provider_name":"测试机构","title":"低空巡检服务","category":"巡检","description":"电力巡检作业","region":"南岸区","price_fen":10000,"unit":"按次"}`)
	w := request(t, app, http.MethodPost, "/api/v1/service-listings", body, domain.RoleIndividual)
	if w.Code != http.StatusCreated {
		t.Fatalf("create service listing: %d %s", w.Code, w.Body.String())
	}
	var envelope struct {
		Data struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	created := envelope.Data
	if created.ID == "" {
		t.Fatal("created listing id is empty")
	}
	if created.Status != "pending" {
		t.Fatalf("status = %q, want pending", created.Status)
	}

	// 2) 公开列表不包含（待审核不公开），且匿名可访问（白名单回归：此前 401）
	w = request(t, app, http.MethodGet, "/api/v1/service-listings", nil, "")
	if w.Code != http.StatusOK {
		t.Fatalf("public list: %d (was 401 before whitelist fix)", w.Code)
	}
	if strings.Contains(w.Body.String(), created.ID) {
		t.Fatal("pending listing must not appear in public list")
	}

	// 3) mine=1 未登录 → 空列表
	w = request(t, app, http.MethodGet, "/api/v1/service-listings?mine=1", nil, "")
	if w.Code != http.StatusOK {
		t.Fatalf("anon mine: %d", w.Code)
	}
	if strings.Contains(w.Body.String(), created.ID) {
		t.Fatal("anonymous mine must not leak listing")
	}

	// 4) mine=1 登录 → 包含自己的待审核记录
	w = request(t, app, http.MethodGet, "/api/v1/service-listings?mine=1", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK {
		t.Fatalf("mine list: %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), created.ID) {
		t.Fatal("mine=1 should include own pending listing")
	}
}

// 用户发布课程待审核（pending），不进公开列表；管理端审核通过（published）后公开；
// mine=1 只看自己（机构）的课程，匿名为空。
func TestCreateCoursePendingAndMine(t *testing.T) {
	app := newBizServer(t)

	body := []byte(`{"title":"CAAC 多旋翼执照班","cert_type":"caac","description":"执照培训","org_name":"测试航校","district":"南岸区","location":"金开大道68号","price_fen":980000,"duration_days":25,"max_students":20}`)
	w := request(t, app, http.MethodPost, "/api/v1/training-courses", body, domain.RoleIndividual)
	if w.Code != http.StatusCreated {
		t.Fatalf("create course: %d %s", w.Code, w.Body.String())
	}
	var envelope struct {
		Data struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	created := envelope.Data
	if created.Status != "pending" {
		t.Fatalf("course status = %q, want pending", created.Status)
	}

	// 公开列表不含待审核课程
	w = request(t, app, http.MethodGet, "/api/v1/training-courses", nil, "")
	if strings.Contains(w.Body.String(), created.ID) {
		t.Fatal("pending course must not appear in public list")
	}

	// mine=1 未登录 → 空
	w = request(t, app, http.MethodGet, "/api/v1/training-courses?mine=1", nil, "")
	if strings.Contains(w.Body.String(), created.ID) {
		t.Fatal("anonymous course mine must not leak")
	}

	// mine=1 登录 → 包含（含待审核）
	w = request(t, app, http.MethodGet, "/api/v1/training-courses?mine=1", nil, domain.RoleIndividual)
	if !strings.Contains(w.Body.String(), created.ID) {
		t.Fatal("mine=1 should include own course")
	}

	// 管理端审核通过（status → published）后进入公开列表
	aw := request(t, app, http.MethodPut, "/api/v1/admin/training-courses/"+created.ID,
		[]byte(`{"status":"published"}`), domain.RolePlatformAdmin)
	if aw.Code != http.StatusOK {
		t.Fatalf("admin approve course: %d %s", aw.Code, aw.Body.String())
	}
	w = request(t, app, http.MethodGet, "/api/v1/training-courses", nil, "")
	if !strings.Contains(w.Body.String(), created.ID) {
		t.Fatal("approved course should appear in public list")
	}

	// 管理端关闭（closed）后不再公开
	cw := request(t, app, http.MethodPut, "/api/v1/admin/training-courses/"+created.ID,
		[]byte(`{"status":"closed"}`), domain.RolePlatformAdmin)
	if cw.Code != http.StatusOK {
		t.Fatalf("admin close course: %d %s", cw.Code, cw.Body.String())
	}
	w = request(t, app, http.MethodGet, "/api/v1/training-courses", nil, "")
	if strings.Contains(w.Body.String(), created.ID) {
		t.Fatal("closed course must not appear in public list")
	}
}

// 公开课程列表对**管理员**也必须过滤未公开课程。
//
// 回归：公开路由 /api/v1/training-courses 此前与管理端路由
// /api/v1/admin/training-courses 共用同一个 handler，靠 authenticatedActor 的角色
// 放行非公开课程。于是 platform_admin / association_admin 在小程序里打开培训课程列表，
// 会看到全部待审核/草稿/已下架课程——实测现象是"后台设成三种状态的三门课，
// 管理员登录的小程序里一门不落地全都在"。
//
// 原有用例只测了匿名视角（request(..., "", ...)），这条旁路一直没有覆盖。
func TestPublicCourseListFiltersEvenForAdmins(t *testing.T) {
	app := newBizServer(t)

	body := []byte(`{"title":"管理员可见性测试班","cert_type":"caac","description":"过滤测试","org_name":"测试航校","district":"南岸区","location":"金开大道68号","price_fen":980000,"duration_days":25,"max_students":20}`)
	w := request(t, app, http.MethodPost, "/api/v1/training-courses", body, domain.RoleIndividual)
	if w.Code != http.StatusCreated {
		t.Fatalf("create course: %d %s", w.Code, w.Body.String())
	}
	var envelope struct {
		Data struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	id := envelope.Data.ID
	if id == "" {
		t.Fatal("created course id is empty")
	}
	if envelope.Data.Status != "pending" {
		t.Fatalf("course status = %q, want pending", envelope.Data.Status)
	}

	// 1) 待审核：管理员走**公开**接口也看不到——本次修复点
	w = request(t, app, http.MethodGet, "/api/v1/training-courses", nil, domain.RolePlatformAdmin)
	if w.Code != http.StatusOK {
		t.Fatalf("public list as admin: %d", w.Code)
	}
	if strings.Contains(w.Body.String(), id) {
		t.Fatal("pending course must not appear in PUBLIC list even for admin")
	}

	// 2) 同一时刻管理端接口必须能看到，否则后台没法审核
	w = request(t, app, http.MethodGet, "/api/v1/admin/training-courses", nil, domain.RolePlatformAdmin)
	if w.Code != http.StatusOK {
		t.Fatalf("admin list: %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), id) {
		t.Fatal("admin list must include pending course")
	}

	// 3) 审核通过后公开接口可见
	w = request(t, app, http.MethodPut, "/api/v1/admin/training-courses/"+id,
		[]byte(`{"status":"published"}`), domain.RolePlatformAdmin)
	if w.Code != http.StatusOK {
		t.Fatalf("admin publish: %d %s", w.Code, w.Body.String())
	}
	w = request(t, app, http.MethodGet, "/api/v1/training-courses", nil, domain.RolePlatformAdmin)
	if !strings.Contains(w.Body.String(), id) {
		t.Fatal("published course should appear in public list")
	}

	// 4) 下架后公开接口又不可见，但管理端仍能看到（后台要能重新上架）
	w = request(t, app, http.MethodPut, "/api/v1/admin/training-courses/"+id,
		[]byte(`{"status":"closed"}`), domain.RolePlatformAdmin)
	if w.Code != http.StatusOK {
		t.Fatalf("admin close: %d %s", w.Code, w.Body.String())
	}
	w = request(t, app, http.MethodGet, "/api/v1/training-courses", nil, domain.RolePlatformAdmin)
	if strings.Contains(w.Body.String(), id) {
		t.Fatal("closed course must not appear in PUBLIC list even for admin")
	}
	w = request(t, app, http.MethodGet, "/api/v1/admin/training-courses", nil, domain.RolePlatformAdmin)
	if !strings.Contains(w.Body.String(), id) {
		t.Fatal("admin list must still include closed course")
	}
}

// GET /api/v1/admin/certified-pilots/{id} — 审核详情必须给出"申请人提交了什么"。
//
// 回归：管理端此前只有列表接口，证书仅以 cert_ids（一串 ID）呈现，审核人看不到
// 证书类型/编号/发证机构/有效期/照片，无从核对；身份证也在前端被打了码。
//
// 另注意它与公开详情 GetPilotDetail 的口径差异：公开详情按 certValid 过滤
//（只给已通过且未过期的），审核详情必须**含待审证书**——否则刚随申请提交的
// 那批 pending 证书根本看不到，审核就成了盲审。
func TestAdminPilotReviewDetail(t *testing.T) {
	app := newBizServer(t)

	body := []byte(`{"real_name":"李四","id_card":"500101199001011234","flight_hours":120,"bio":"电力巡检","region":"渝北区","certs":[{"cert_type":"caac","cert_number":"CAAC-REVIEW-1","issuer_org":"民航局","issue_date":"2024-01-01","expire_date":"2030-01-01"}]}`)
	w := request(t, app, http.MethodPost, "/api/v1/certified-pilots", body, domain.RoleIndividual)
	if w.Code != http.StatusCreated {
		t.Fatalf("register pilot with certs: %d %s", w.Code, w.Body.String())
	}
	var env struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &env); err != nil || env.Data.ID == "" {
		t.Fatalf("unmarshal pilot id: %v %s", err, w.Body.String())
	}
	id := env.Data.ID

	// 非管理员不得拿到审核资料
	w = request(t, app, http.MethodGet, "/api/v1/admin/certified-pilots/"+id, nil, domain.RoleIndividual)
	if w.Code == http.StatusOK {
		t.Fatal("非管理员不应拿到审核资料")
	}

	// 管理员：完整档案 + 随附证书明细
	w = request(t, app, http.MethodGet, "/api/v1/admin/certified-pilots/"+id, nil, domain.RolePlatformAdmin)
	if w.Code != http.StatusOK {
		t.Fatalf("admin pilot review detail: %d %s", w.Code, w.Body.String())
	}
	got := w.Body.String()
	for _, want := range []string{"CAAC-REVIEW-1", "民航局", "500101199001011234", "certificates"} {
		if !strings.Contains(got, want) {
			t.Fatalf("审核资料缺少 %q: %s", want, got)
		}
	}
	// 关键口径：必须含**待审**证书（公开详情会把它过滤掉）
	if !strings.Contains(got, `"status":"pending"`) {
		t.Fatalf("审核资料应包含待审状态的证书: %s", got)
	}
}

// 完整身份证号只下发给平台管理员，协会管理员拿脱敏值（详情与列表**两个**接口都算）。
//
// 回归：管理端列表接口本来就整表下发完整证号，任何 association_admin 都能一次性
// 批量导出全平台飞手的身份证号；后加的审核详情接口延续了同一口径。
// 只锁详情不锁列表等于没锁——列表才是批量导出面。
func TestAdminPilotPIIMaskingByRole(t *testing.T) {
	app := newBizServer(t)
	const fullID = "500101199001019999"

	body := []byte(`{"real_name":"证件分级测试","id_card":"` + fullID + `","flight_hours":66,"certs":[{"cert_type":"caac","cert_number":"PII-MASK-1","issue_date":"2024-01-01"}]}`)
	w := request(t, app, http.MethodPost, "/api/v1/certified-pilots", body, domain.RoleIndividual)
	if w.Code != http.StatusCreated {
		t.Fatalf("register pilot: %d %s", w.Code, w.Body.String())
	}
	var env struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &env); err != nil || env.Data.ID == "" {
		t.Fatalf("unmarshal pilot id: %v %s", err, w.Body.String())
	}
	id := env.Data.ID

	// ── 详情接口 ──
	wp := request(t, app, http.MethodGet, "/api/v1/admin/certified-pilots/"+id, nil, domain.RolePlatformAdmin)
	if wp.Code != http.StatusOK || !strings.Contains(wp.Body.String(), fullID) {
		t.Fatalf("平台管理员应拿到完整身份证: %d %s", wp.Code, wp.Body.String())
	}
	wa := request(t, app, http.MethodGet, "/api/v1/admin/certified-pilots/"+id, nil, domain.RoleAssociationAdmin)
	if wa.Code != http.StatusOK {
		t.Fatalf("协会管理员应能打开详情: %d %s", wa.Code, wa.Body.String())
	}
	if strings.Contains(wa.Body.String(), fullID) {
		t.Fatalf("协会管理员不得拿到完整身份证: %s", wa.Body.String())
	}
	if !strings.Contains(wa.Body.String(), "***********") {
		t.Fatalf("协会管理员应拿到脱敏值: %s", wa.Body.String())
	}

	// ── 列表接口（批量导出面）──
	wp = request(t, app, http.MethodGet, "/api/v1/admin/certified-pilots", nil, domain.RolePlatformAdmin)
	if !strings.Contains(wp.Body.String(), fullID) {
		t.Fatalf("平台管理员列表应含完整身份证: %s", wp.Body.String())
	}
	wa = request(t, app, http.MethodGet, "/api/v1/admin/certified-pilots", nil, domain.RoleAssociationAdmin)
	if strings.Contains(wa.Body.String(), fullID) {
		t.Fatalf("协会管理员列表不得含完整身份证: %s", wa.Body.String())
	}
}
