package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"drone-platform/internal/domain"
)

// ── 字段契约核对修复的回归测试（2026-08-26） ──
// P1: challenges claims 路由 / testsites booking 扩展字段 / portfolios 公开详情
// P2: achievements 统计与状态 / challenges requirements / portfolios 分类排序 / transformations 联系方式 / testsites 场地参数
// P3: poster_name / portfolios 展示字段 / testsites image_url

// unmarshalData 解包 Go 响应信封 {data: ...} 到 v。
func unmarshalData(t *testing.T, w *httptest.ResponseRecorder, v any) {
	t.Helper()
	var env struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &env); err != nil {
		t.Fatalf("parse envelope: %v (%s)", err, w.Body.String())
	}
	if err := json.Unmarshal(env.Data, v); err != nil {
		t.Fatalf("parse data: %v (%s)", err, w.Body.String())
	}
}

// TestR5CertificateToPilotFullLoop 证书归档闭环：提交(含图片)→驳回→覆盖重提→通过→飞手认证申请成功。
func TestR5CertificateToPilotFullLoop(t *testing.T) {
	app := newBizServer(t)
	userTok := authAs(t, "user-1", domain.RoleIndividual)
	adminTok := authAs(t, "admin-1", domain.RolePlatformAdmin)

	// 1. 用户提交证书（含 image_url）
	w := doRaw(app, http.MethodPost, "/api/v1/certificates",
		`{"cert_type":"caac","cert_number":"CAAC-T-88888","level":"III","issuer_org":"民航局","image_url":"/uploads/cert1.jpg","issue_date":"2026-01-01","expire_date":"2029-01-01"}`, userTok)
	assertStatus(t, http.MethodPost, "/api/v1/certificates", w, http.StatusCreated)
	certID := dataID(t, w)

	// 2. 管理端驳回
	w = doRaw(app, http.MethodPost, "/api/v1/admin/certificates/"+certID+"/reject", "", adminTok)
	assertStatus(t, http.MethodPost, "/api/v1/admin/certificates/"+certID+"/reject", w, http.StatusOK)

	// 3. 同证书号覆盖重提（rejected → pending）
	w = doRaw(app, http.MethodPost, "/api/v1/certificates",
		`{"cert_type":"caac","cert_number":"CAAC-T-88888","level":"III","issuer_org":"民航局","image_url":"/uploads/cert1b.jpg","issue_date":"2026-01-01","expire_date":"2029-01-01"}`, userTok)
	assertStatus(t, http.MethodPost, "/api/v1/certificates", w, http.StatusCreated)

	// 4. 审批通过
	w = doRaw(app, http.MethodPost, "/api/v1/admin/certificates/"+certID+"/approve", "", adminTok)
	assertStatus(t, http.MethodPost, "/api/v1/admin/certificates/"+certID+"/approve", w, http.StatusOK)

	// 5. 飞手认证申请成功（有有效证书 → 不再 403）
	w = doRaw(app, http.MethodPost, "/api/v1/certified-pilots",
		`{"real_name":"张三","id_card":"500101199001011234","flight_hours":120,"bio":"测试","region":"渝北区"}`, userTok)
	assertStatus(t, http.MethodPost, "/api/v1/certified-pilots", w, http.StatusCreated)
}

// dataListLen 解析列表信封 {data:[...]} 的长度。
func dataListLen(t *testing.T, w *httptest.ResponseRecorder) int {
	t.Helper()
	var out struct {
		Data []json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("parse list: %v (%s)", err, w.Body.String())
	}
	return len(out.Data)
}

// TestR5ChallengeClaimsFlow 揭榜全流程：发布 → 匿名看列表 → 揭榜 → 幂等 409 → 截止 409 → 不存在 404。
func TestR5ChallengeClaimsFlow(t *testing.T) {
	app := newBizServer(t)
	entTok := authAs(t, "user-1", domain.RoleEnterprise)

	// 发布难题（open 状态可揭榜）
	w := doRaw(app, http.MethodPost, "/api/v1/rd-challenges",
		`{"title":"长航时电池","field":"电池","description":"轻量大容量","requirements":"续航≥2h；循环≥500次","budget_fen":500000,"deadline":"2027-01-01","status":"open"}`, entTok)
	assertStatus(t, http.MethodPost, "/api/v1/rd-challenges", w, http.StatusCreated)
	chID := dataID(t, w)

	// 匿名 GET claims：items 空、claimed=false
	w = doRaw(app, http.MethodGet, "/api/v1/challenges/"+chID+"/claims", "", "")
	assertStatus(t, http.MethodGet, "/api/v1/challenges/"+chID+"/claims", w, http.StatusOK)
	var list0 struct {
		Items   []domain.ChallengeClaim `json:"items"`
		Total   int                     `json:"total"`
		Claimed bool                    `json:"claimed"`
	}
	unmarshalData(t, w, &list0)
	if list0.Total != 0 || list0.Claimed {
		t.Fatalf("initial claims: total=%d claimed=%v", list0.Total, list0.Claimed)
	}

	// user-1 揭榜
	w = doRaw(app, http.MethodPost, "/api/v1/challenges/"+chID+"/claims", `{}`, entTok)
	assertStatus(t, http.MethodPost, "/api/v1/challenges/"+chID+"/claims", w, http.StatusCreated)

	// 重复揭榜 → 409（幂等守卫）
	w = doRaw(app, http.MethodPost, "/api/v1/challenges/"+chID+"/claims", `{}`, entTok)
	assertStatus(t, http.MethodPost, "/api/v1/challenges/"+chID+"/claims", w, http.StatusConflict)

	// 带 token GET：claimed=true（测试用户不在 userRepo，claimer 兜底"匿名会员"）
	w = doRaw(app, http.MethodGet, "/api/v1/challenges/"+chID+"/claims", "", entTok)
	assertStatus(t, http.MethodGet, "/api/v1/challenges/"+chID+"/claims", w, http.StatusOK)
	var list1 struct {
		Items   []domain.ChallengeClaim `json:"items"`
		Total   int                     `json:"total"`
		Claimed bool                    `json:"claimed"`
	}
	unmarshalData(t, w, &list1)
	if list1.Total != 1 || !list1.Claimed {
		t.Fatalf("after claim: total=%d claimed=%v", list1.Total, list1.Claimed)
	}
	if len(list1.Items) != 1 || list1.Items[0].Status != "submitted" || list1.Items[0].Claimer == "" {
		t.Fatalf("claim item: %+v", list1.Items)
	}

	// 截止（closed）后他人揭榜 → 409
	w = doRaw(app, http.MethodPut, "/api/v1/rd-challenges/"+chID,
		`{"title":"长航时电池","field":"电池","description":"轻量大容量","requirements":"续航≥2h","budget_fen":500000,"deadline":"2027-01-01","status":"closed"}`, entTok)
	assertStatus(t, http.MethodPut, "/api/v1/rd-challenges/"+chID, w, http.StatusOK)
	otherTok := authAs(t, "user-2", domain.RoleIndividual)
	w = doRaw(app, http.MethodPost, "/api/v1/challenges/"+chID+"/claims", `{}`, otherTok)
	assertStatus(t, http.MethodPost, "/api/v1/challenges/"+chID+"/claims", w, http.StatusConflict)

	// 不存在的难题 → 404
	w = doRaw(app, http.MethodPost, "/api/v1/challenges/nope/claims", `{}`, otherTok)
	assertStatus(t, http.MethodPost, "/api/v1/challenges/nope/claims", w, http.StatusNotFound)
}

// TestR5TestSiteBookingExtFields booking 扩展字段完整落库并可回读。
func TestR5TestSiteBookingExtFields(t *testing.T) {
	app := newBizServer(t)
	adminTok := authAs(t, "admin-1", domain.RolePlatformAdmin)
	userTok := authAs(t, "user-1", domain.RoleIndividual)

	// 管理员建场地（带场地参数）
	w := doRaw(app, http.MethodPost, "/api/v1/admin/test-sites",
		`{"name":"渝北试飞场","site_type":"flying_field","location":"渝北区","booking_rule":"工作日9-18点","price_fen":0,"status":"available","facilities":["RTK"],"airspace_range":"10km","max_takeoff_weight":"25kg","runway_length":"800m","max_flight_height":"500m","compatible_models":"多旋翼","image_url":"/uploads/site1.jpg"}`, adminTok)
	assertStatus(t, http.MethodPost, "/api/v1/admin/test-sites", w, http.StatusCreated)
	siteID := dataID(t, w)

	// 场地参数回读
	w = doRaw(app, http.MethodGet, "/api/v1/test-sites/"+siteID, "", "")
	assertStatus(t, http.MethodGet, "/api/v1/test-sites/"+siteID, w, http.StatusOK)
	var site domain.TestSite
	unmarshalData(t, w, &site)
	if site.AirspaceRange != "10km" || site.MaxTakeoffWeight != "25kg" || site.RunwayLength != "800m" ||
		site.MaxFlightHeight != "500m" || site.CompatibleModels != "多旋翼" || site.ImageURL != "/uploads/site1.jpg" {
		t.Fatalf("site params not persisted: %+v", site)
	}

	// 预约提交 9 扩展字段
	w = doRaw(app, http.MethodPost, "/api/v1/test-sites/"+siteID+"/book",
		`{"purpose":"R&D","date":"2026-09-01","time_slot":"09:00-11:00","time_slots":"09:00-11:00,14:00-16:00","booking_type":"group","model":"M300","license_url":"/uploads/lic.pdf","team_name":"飞鹰队","people_count":3,"equipment_list":"M300两台","qualification_url":"/uploads/quali.pdf","equipment_note":"外场测试","contact_name":"张三","contact_phone":"13800000000"}`, userTok)
	assertStatus(t, http.MethodPost, "/api/v1/test-sites/"+siteID+"/book", w, http.StatusCreated)

	// 我的预约回读：9 字段与 2 个基础字段一致
	w = doRaw(app, http.MethodGet, "/api/v1/test-sites/bookings/mine", "", userTok)
	assertStatus(t, http.MethodGet, "/api/v1/test-sites/bookings/mine", w, http.StatusOK)
	var mine struct {
		Data []domain.TestSiteBooking `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &mine); err != nil {
		t.Fatalf("parse mine: %v (%s)", err, w.Body.String())
	}
	if len(mine.Data) == 0 {
		t.Fatalf("mine empty: %s", w.Body.String())
	}
	b := mine.Data[0]
	if b.BookingType != "group" || b.Model != "M300" || b.LicenseURL != "/uploads/lic.pdf" ||
		b.TeamName != "飞鹰队" || b.PeopleCount != 3 || b.EquipmentList != "M300两台" ||
		b.QualificationURL != "/uploads/quali.pdf" || b.EquipmentNote != "外场测试" ||
		b.TimeSlots != "09:00-11:00,14:00-16:00" {
		t.Fatalf("booking ext fields lost: %+v", b)
	}
}

// TestR5PortfolioPublicDetailAndStats 公开详情路由 + 草稿过滤 + 浏览计数 + featured + 分类筛选。
func TestR5PortfolioPublicDetailAndStats(t *testing.T) {
	app := newBizServer(t)
	entTok := authAs(t, "user-1", domain.RoleEnterprise)
	adminTok := authAs(t, "admin-1", domain.RolePlatformAdmin)

	// 建品牌（草稿）
	w := doRaw(app, http.MethodPost, "/api/v1/portfolios",
		`{"name":"巡航科技","logo_url":"/uploads/logo.png","cover_url":"/uploads/cover.jpg","description":"无人机方案商","category":"整机","industry":"巡检","video_url":"/uploads/v.mp4","products":["M300"],"honors":["优秀"],"contact_info":"138","status":"draft"}`, entTok)
	assertStatus(t, http.MethodPost, "/api/v1/portfolios", w, http.StatusCreated)
	pID := dataID(t, w)

	// 草稿 → 公开详情 404
	w = doRaw(app, http.MethodGet, "/api/v1/portfolios/"+pID, "", "")
	assertStatus(t, http.MethodGet, "/api/v1/portfolios/"+pID, w, http.StatusNotFound)

	// 管理端发布（featured/verified）
	w = doRaw(app, http.MethodPut, "/api/v1/admin/portfolios/"+pID,
		`{"name":"巡航科技","logo_url":"/uploads/logo.png","cover_url":"/uploads/cover.jpg","description":"无人机方案商","category":"整机","industry":"巡检","video_url":"/uploads/v.mp4","products":["M300"],"honors":["优秀"],"contact_info":"138","status":"published","featured":true,"verified":true}`, adminTok)
	assertStatus(t, http.MethodPut, "/api/v1/admin/portfolios/"+pID, w, http.StatusOK)

	// 公开详情 200 + 字段回显 + 浏览计数（响应值 = 递增后）
	w = doRaw(app, http.MethodGet, "/api/v1/portfolios/"+pID, "", "")
	assertStatus(t, http.MethodGet, "/api/v1/portfolios/"+pID, w, http.StatusOK)
	var p domain.MemberPortfolio
	unmarshalData(t, w, &p)
	if p.Category != "整机" || !p.Featured || !p.Verified || p.VideoURL != "/uploads/v.mp4" {
		t.Fatalf("portfolio ext fields lost: %+v", p)
	}
	if p.Views < 1 {
		t.Fatalf("portfolio views not incremented: %d", p.Views)
	}

	// 再次浏览持续递增
	w = doRaw(app, http.MethodGet, "/api/v1/portfolios/"+pID, "", "")
	assertStatus(t, http.MethodGet, "/api/v1/portfolios/"+pID, w, http.StatusOK)
	unmarshalData(t, w, &p)
	if p.Views < 2 {
		t.Fatalf("portfolio views did not grow: %d", p.Views)
	}

	// featured 接口含此品牌
	w = doRaw(app, http.MethodGet, "/api/v1/portfolios/featured", "", "")
	assertStatus(t, http.MethodGet, "/api/v1/portfolios/featured", w, http.StatusOK)
	var feats []domain.MemberPortfolio
	unmarshalData(t, w, &feats)
	if len(feats) != 1 || feats[0].ID != pID {
		t.Fatalf("featured: %+v", feats)
	}

	// category 筛选
	w = doRaw(app, http.MethodGet, "/api/v1/portfolios?category=整机", "", "")
	assertStatus(t, http.MethodGet, "/api/v1/portfolios?category=整机", w, http.StatusOK)
	if got := dataListLen(t, w); got != 1 {
		t.Fatalf("category filter: got %d items, want 1 (%s)", got, w.Body.String())
	}
}

// TestR5AchievementStatsAndStatus 浏览量/收藏量/状态值域（transformed 透传）。
func TestR5AchievementStatsAndStatus(t *testing.T) {
	app := newBizServer(t)
	entTok := authAs(t, "user-1", domain.RoleEnterprise)
	userTok := authAs(t, "user-2", domain.RoleIndividual)

	w := doRaw(app, http.MethodPost, "/api/v1/achievements",
		`{"title":"智巡算法","achieve_type":"patent","description":"自动避障","field":"无人机","stage":"pilot","contact_info":"138","images":["a.jpg"],"status":"published"}`, entTok)
	assertStatus(t, http.MethodPost, "/api/v1/achievements", w, http.StatusCreated)
	achID := dataID(t, w)

	// 公开详情浏览（首次为计数前快照，第二次可见 +1）
	doRaw(app, http.MethodGet, "/api/v1/achievements/"+achID, "", "")
	w = doRaw(app, http.MethodGet, "/api/v1/achievements/"+achID, "", "")
	assertStatus(t, http.MethodGet, "/api/v1/achievements/"+achID, w, http.StatusOK)
	var a domain.Achievement
	unmarshalData(t, w, &a)
	if a.Views < 1 {
		t.Fatalf("achievement views not incremented: %d", a.Views)
	}

	// 收藏 +1
	w = doRaw(app, http.MethodPost, "/api/v1/achievements/"+achID+"/favorite", `{"favorite":true}`, userTok)
	assertStatus(t, http.MethodPost, "/api/v1/achievements/"+achID+"/favorite", w, http.StatusOK)
	w = doRaw(app, http.MethodGet, "/api/v1/achievements/"+achID, "", userTok)
	assertStatus(t, http.MethodGet, "/api/v1/achievements/"+achID, w, http.StatusOK)
	unmarshalData(t, w, &a)
	if a.Favs != 1 {
		t.Fatalf("achievement favs: %d, want 1", a.Favs)
	}

	// 取消收藏 → 0
	w = doRaw(app, http.MethodPost, "/api/v1/achievements/"+achID+"/favorite", `{"favorite":false}`, userTok)
	assertStatus(t, http.MethodPost, "/api/v1/achievements/"+achID+"/favorite", w, http.StatusOK)
	w = doRaw(app, http.MethodGet, "/api/v1/achievements/"+achID, "", userTok)
	unmarshalData(t, w, &a)
	if a.Favs != 0 {
		t.Fatalf("achievement favs after unfavorite: %d, want 0", a.Favs)
	}

	// 未登录收藏 → 401
	w = doRaw(app, http.MethodPost, "/api/v1/achievements/"+achID+"/favorite", `{"favorite":true}`, "")
	assertStatus(t, http.MethodPost, "/api/v1/achievements/"+achID+"/favorite", w, http.StatusUnauthorized)

	// 状态值 transformed（管理端更新）→ 公开详情返回 transformed
	w = doRaw(app, http.MethodPut, "/api/v1/admin/achievements/"+achID,
		`{"title":"智巡算法","achieve_type":"patent","description":"自动避障","field":"无人机","stage":"listed","contact_info":"138","images":["a.jpg"],"status":"transformed"}`, authAs(t, "admin-1", domain.RolePlatformAdmin))
	assertStatus(t, http.MethodPut, "/api/v1/admin/achievements/"+achID, w, http.StatusOK)
	w = doRaw(app, http.MethodGet, "/api/v1/achievements/"+achID, "", "")
	assertStatus(t, http.MethodGet, "/api/v1/achievements/"+achID, w, http.StatusOK)
	unmarshalData(t, w, &a)
	if a.Status != "transformed" {
		t.Fatalf("achievement status: %q, want transformed", a.Status)
	}
}

// TestR5ChallengeRequirementsAndPoster requirements 落库回显。
func TestR5ChallengeRequirementsAndPoster(t *testing.T) {
	app := newBizServer(t)
	entTok := authAs(t, "user-1", domain.RoleEnterprise)

	w := doRaw(app, http.MethodPost, "/api/v1/rd-challenges",
		`{"title":"飞控抗干扰","field":"飞控","description":"强电磁环境","requirements":"抗干扰≥50dB；续航≥1h","budget_fen":0,"deadline":"2027-06-01","status":"open"}`, entTok)
	assertStatus(t, http.MethodPost, "/api/v1/rd-challenges", w, http.StatusCreated)
	chID := dataID(t, w)

	w = doRaw(app, http.MethodGet, "/api/v1/challenges/"+chID, "", "")
	assertStatus(t, http.MethodGet, "/api/v1/challenges/"+chID, w, http.StatusOK)
	var c domain.RDChallenge
	unmarshalData(t, w, &c)
	if c.Requirements != "抗干扰≥50dB；续航≥1h" {
		t.Fatalf("requirements lost: %q", c.Requirements)
	}
}

// TestR5TransformationContactInfo 转化记录联系方式回显。
func TestR5TransformationContactInfo(t *testing.T) {
	app := newBizServer(t)
	entTok := authAs(t, "user-1", domain.RoleEnterprise)

	// 建成果 → 建转化（contact_info 透传）
	w := doRaw(app, http.MethodPost, "/api/v1/achievements",
		`{"title":"成果A","achieve_type":"patent","description":"d","field":"无人机","stage":"lab","images":[]}`, entTok)
	assertStatus(t, http.MethodPost, "/api/v1/achievements", w, http.StatusCreated)
	achID := dataID(t, w)

	w = doRaw(app, http.MethodPost, "/api/v1/admin/transformations",
		`{"title":"成果转化A","achievement_id":"`+achID+`","partner_id":"partner-1","contact_info":"138****1234"}`, authAs(t, "admin-1", domain.RolePlatformAdmin))
	assertStatus(t, http.MethodPost, "/api/v1/admin/transformations", w, http.StatusCreated)

	w = doRaw(app, http.MethodGet, "/api/v1/transformations?achievement_id="+achID, "", "")
	assertStatus(t, http.MethodGet, "/api/v1/transformations?achievement_id="+achID, w, http.StatusOK)
	var list struct {
		Data []domain.Transformation `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &list); err != nil {
		t.Fatalf("parse transformations: %v (%s)", err, w.Body.String())
	}
	if len(list.Data) != 1 || list.Data[0].ContactInfo != "138****1234" {
		t.Fatalf("contact_info lost: %+v", list.Data)
	}
}
