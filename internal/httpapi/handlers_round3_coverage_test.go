package httpapi_test

import (
	"net/http"
	"testing"

	"drone-platform/internal/config"
	"drone-platform/internal/domain"
)

// round3PlatformSnapshot 保存当前平台配置，测试结束后写回（含落盘文件），
// 避免 updateConfig 的 SavePlatformConfig 污染其他测试。
func round3PlatformSnapshot(t *testing.T) {
	t.Helper()
	snapshot := config.GetPlatformConfig()
	t.Cleanup(func() {
		if err := config.SavePlatformConfig(snapshot); err != nil {
			t.Logf("round3 restore platform config: %v", err)
		}
	})
}

// TestRound3Enterprise 覆盖 enterprise.go：企业入驻 → 编辑 → 提交 → 管理审核 →
// 公开名录/详情 → 批量审核闭环。
func TestRound3Enterprise(t *testing.T) {
	app := newBizServer(t)
	ownerTok := authAs(t, "enterprise-1", domain.RoleEnterprise)
	adminTok := authAs(t, "admin-1", domain.RolePlatformAdmin)

	// 公开名录（匿名）
	w := doRaw(app, http.MethodGet, "/api/v1/enterprises/public", "", "")
	assertStatus(t, http.MethodGet, "/api/v1/enterprises/public", w, http.StatusOK)

	// 不存在的企业详情 → 404
	w = doRaw(app, http.MethodGet, "/api/v1/enterprises/public/detail?id=ent-nope", "", "")
	assertStatus(t, http.MethodGet, "/api/v1/enterprises/public/detail?id=ent-nope", w, http.StatusNotFound)

	// 创建企业
	w = doRaw(app, http.MethodPost, "/api/v1/enterprises",
		`{"name":"round3企业","industry_category":"测绘","description":"测试企业"}`, ownerTok)
	assertStatus(t, http.MethodPost, "/api/v1/enterprises", w, http.StatusCreated)
	entID := dataID(t, w)

	// 编辑（草稿态）
	w = doRaw(app, http.MethodPatch, "/api/v1/enterprises/"+entID,
		`{"name":"round3企业-改","description":"更新描述"}`, ownerTok)
	assertStatus(t, http.MethodPatch, "/api/v1/enterprises/"+entID, w, http.StatusOK)

	// 我的企业列表
	w = doRaw(app, http.MethodGet, "/api/v1/enterprises", "", ownerTok)
	assertStatus(t, http.MethodGet, "/api/v1/enterprises", w, http.StatusOK)

	// 提交审核
	w = doRaw(app, http.MethodPost, "/api/v1/enterprises/"+entID+"/submit", "", ownerTok)
	assertStatus(t, http.MethodPost, "/api/v1/enterprises/"+entID+"/submit", w, http.StatusOK)

	// 管理端列表 / 待审核 / 搜索
	w = doRaw(app, http.MethodGet, "/api/v1/admin/enterprises", "", adminTok)
	assertStatus(t, http.MethodGet, "/api/v1/admin/enterprises", w, http.StatusOK)

	w = doRaw(app, http.MethodGet, "/api/v1/admin/enterprises/pending", "", adminTok)
	assertStatus(t, http.MethodGet, "/api/v1/admin/enterprises/pending", w, http.StatusOK)

	w = doRaw(app, http.MethodGet, "/api/v1/admin/enterprises/search?q=round3", "", adminTok)
	assertStatus(t, http.MethodGet, "/api/v1/admin/enterprises/search?q=round3", w, http.StatusOK)

	// 审核通过 → 公开详情 200
	w = doRaw(app, http.MethodPost, "/api/v1/admin/enterprises/"+entID+"/review",
		`{"action":"approve"}`, adminTok)
	assertStatus(t, http.MethodPost, "/api/v1/admin/enterprises/"+entID+"/review", w, http.StatusOK)

	w = doRaw(app, http.MethodGet, "/api/v1/enterprises/public/detail?id="+entID, "", "")
	assertStatus(t, http.MethodGet, "/api/v1/enterprises/public/detail?id="+entID, w, http.StatusOK)

	// 批量审核：第二条企业（另一企业账号）提交后批量通过
	owner2Tok := authAs(t, "enterprise-2", domain.RoleEnterprise)
	w = doRaw(app, http.MethodPost, "/api/v1/enterprises",
		`{"name":"round3企业2"}`, owner2Tok)
	assertStatus(t, http.MethodPost, "/api/v1/enterprises (2)", w, http.StatusCreated)
	entID2 := dataID(t, w)

	w = doRaw(app, http.MethodPost, "/api/v1/enterprises/"+entID2+"/submit", "", owner2Tok)
	assertStatus(t, http.MethodPost, "/api/v1/enterprises/"+entID2+"/submit", w, http.StatusOK)

	w = doRaw(app, http.MethodPost, "/api/v1/admin/enterprises/batch-review",
		`{"ids":["`+entID2+`"],"action":"approve"}`, adminTok)
	assertStatus(t, http.MethodPost, "/api/v1/admin/enterprises/batch-review", w, http.StatusOK)
}

// TestRound3Batch1 覆盖 batch1_handlers.go：资源池 / 试飞场预约 / 展会展位。
func TestRound3Batch1(t *testing.T) {
	app := newBizServer(t)
	adminTok := authAs(t, "admin-1", domain.RolePlatformAdmin)
	userTok := authAs(t, "user-1", domain.RoleIndividual)

	// 资源池
	w := doRaw(app, http.MethodPost, "/api/v1/admin/resource-pools",
		`{"name":"round3应急资源池","pool_type":"emergency","description":"低空应急装备"}`, adminTok)
	assertStatus(t, http.MethodPost, "/api/v1/admin/resource-pools", w, http.StatusCreated)
	poolID := dataID(t, w)

	w = doRaw(app, http.MethodGet, "/api/v1/resource-pools", "", "")
	assertStatus(t, http.MethodGet, "/api/v1/resource-pools", w, http.StatusOK)

	// 资源池成员：仅池主(admin)或管理员可注入；普通用户应被拒绝
	w = doRaw(app, http.MethodPost, "/api/v1/resource-pools/"+poolID+"/members",
		`{"res_id":"res-1","res_type":"drone","quantity":3}`, adminTok)
	assertStatus(t, http.MethodPost, "/api/v1/resource-pools/"+poolID+"/members", w, http.StatusCreated)

	w = doRaw(app, http.MethodPost, "/api/v1/resource-pools/"+poolID+"/members",
		`{"res_id":"res-2","res_type":"drone","quantity":1}`, userTok)
	assertStatus(t, http.MethodPost, "/api/v1/resource-pools/"+poolID+"/members", w, http.StatusForbidden)

	w = doRaw(app, http.MethodGet, "/api/v1/resource-pools/"+poolID+"/members", "", userTok)
	assertStatus(t, http.MethodGet, "/api/v1/resource-pools/"+poolID+"/members", w, http.StatusOK)

	// 试飞场
	w = doRaw(app, http.MethodPost, "/api/v1/admin/test-sites",
		`{"name":"round3试飞场","site_type":"flying_field","location":"渝北区","status":"available"}`, adminTok)
	assertStatus(t, http.MethodPost, "/api/v1/admin/test-sites", w, http.StatusCreated)
	siteID := dataID(t, w)

	w = doRaw(app, http.MethodGet, "/api/v1/test-sites", "", "")
	assertStatus(t, http.MethodGet, "/api/v1/test-sites", w, http.StatusOK)

	w = doRaw(app, http.MethodPost, "/api/v1/test-sites/"+siteID+"/book",
		`{"purpose":"certification","date":"2026-08-20","time_slot":"09:00-11:00","contact_name":"张三","contact_phone":"13800000001"}`, userTok)
	assertStatus(t, http.MethodPost, "/api/v1/test-sites/"+siteID+"/book", w, http.StatusCreated)
	bookingID := dataID(t, w)

	w = doRaw(app, http.MethodGet, "/api/v1/test-sites/bookings/mine", "", userTok)
	assertStatus(t, http.MethodGet, "/api/v1/test-sites/bookings/mine", w, http.StatusOK)

	w = doRaw(app, http.MethodPost, "/api/v1/admin/test-sites/bookings/"+bookingID+"/review",
		`{"status":"approved","note":"已确认"}`, adminTok)
	assertStatus(t, http.MethodPost, "/api/v1/admin/test-sites/bookings/"+bookingID+"/review", w, http.StatusOK)

	// 展会
	w = doRaw(app, http.MethodPost, "/api/v1/admin/exhibitions",
		`{"title":"round3低空经济博览会","category":"展销","location":"重庆","booth_count":100,"status":"recruiting"}`, adminTok)
	assertStatus(t, http.MethodPost, "/api/v1/admin/exhibitions", w, http.StatusCreated)
	expoID := dataID(t, w)

	w = doRaw(app, http.MethodGet, "/api/v1/exhibitions", "", "")
	assertStatus(t, http.MethodGet, "/api/v1/exhibitions", w, http.StatusOK)

	w = doRaw(app, http.MethodPost, "/api/v1/exhibitions/"+expoID+"/booths",
		`{"BoothNumber":"A1","ExhibitName":"无人机展品","ExhibitDesc":"最新机型"}`, userTok)
	assertStatus(t, http.MethodPost, "/api/v1/exhibitions/"+expoID+"/booths", w, http.StatusCreated)

	w = doRaw(app, http.MethodGet, "/api/v1/exhibitions/"+expoID+"/booths", "", userTok)
	assertStatus(t, http.MethodGet, "/api/v1/exhibitions/"+expoID+"/booths", w, http.StatusOK)
}

// TestRound3Batch3 覆盖 batch3_handlers.go：救援案例 / 应急部门 / 演练 / 协会会员。
func TestRound3Batch3(t *testing.T) {
	app := newBizServer(t)
	adminTok := authAs(t, "admin-1", domain.RolePlatformAdmin)
	userTok := authAs(t, "user-1", domain.RoleIndividual)

	// 救援案例
	w := doRaw(app, http.MethodPost, "/api/v1/admin/rescue-cases",
		`{"title":"round3山火救援","event_type":"mountain_fire","location":"重庆南山","summary":"火线侦察"}`, adminTok)
	assertStatus(t, http.MethodPost, "/api/v1/admin/rescue-cases", w, http.StatusCreated)

	w = doRaw(app, http.MethodGet, "/api/v1/rescue-cases", "", "")
	assertStatus(t, http.MethodGet, "/api/v1/rescue-cases", w, http.StatusOK)

	// 应急部门
	w = doRaw(app, http.MethodPost, "/api/v1/admin/emergency-depts",
		`{"name":"round3应急救援队","dept_type":"fire","region":"渝北区"}`, adminTok)
	assertStatus(t, http.MethodPost, "/api/v1/admin/emergency-depts", w, http.StatusCreated)

	w = doRaw(app, http.MethodGet, "/api/v1/emergency-depts", "", "")
	assertStatus(t, http.MethodGet, "/api/v1/emergency-depts", w, http.StatusOK)

	// 联合演练
	w = doRaw(app, http.MethodPost, "/api/v1/admin/emergency-drills",
		`{"dept_id":"dept-1","title":"round3联合演练","scenario":"山火","participants":20,"drone_count":5}`, adminTok)
	assertStatus(t, http.MethodPost, "/api/v1/admin/emergency-drills", w, http.StatusCreated)

	w = doRaw(app, http.MethodGet, "/api/v1/emergency-drills", "", "")
	assertStatus(t, http.MethodGet, "/api/v1/emergency-drills", w, http.StatusOK)

	// 协会会员：添加 user-1 → 列表 → 我的身份
	w = doRaw(app, http.MethodPost, "/api/v1/admin/association-members",
		`{"user_id":"user-1","enterprise_id":"ent-1","role":"member"}`, adminTok)
	assertStatus(t, http.MethodPost, "/api/v1/admin/association-members", w, http.StatusCreated)

	w = doRaw(app, http.MethodGet, "/api/v1/association-members", "", "")
	assertStatus(t, http.MethodGet, "/api/v1/association-members", w, http.StatusOK)

	w = doRaw(app, http.MethodGet, "/api/v1/association-members/me", "", userTok)
	assertStatus(t, http.MethodGet, "/api/v1/association-members/me", w, http.StatusOK)
}

// TestRound3News 覆盖 news.go：资讯创建 → 列表 → 编辑 → 发布。
func TestRound3News(t *testing.T) {
	app := newBizServer(t)
	adminTok := authAs(t, "admin-2", domain.RoleAssociationAdmin)

	w := doRaw(app, http.MethodPost, "/api/v1/articles",
		`{"title":"round3政策资讯","content":"低空经济新政发布","category":"policy","source":"协会"}`, adminTok)
	assertStatus(t, http.MethodPost, "/api/v1/articles", w, http.StatusCreated)
	artID := dataID(t, w)

	w = doRaw(app, http.MethodGet, "/api/v1/articles", "", "")
	assertStatus(t, http.MethodGet, "/api/v1/articles", w, http.StatusOK)

	w = doRaw(app, http.MethodPut, "/api/v1/articles/"+artID,
		`{"title":"round3政策资讯-改","content":"低空经济新政发布(修订)","category":"policy","source":"协会"}`, adminTok)
	assertStatus(t, http.MethodPut, "/api/v1/articles/"+artID, w, http.StatusOK)

	w = doRaw(app, http.MethodPost, "/api/v1/articles/"+artID+"/publish", "", adminTok)
	assertStatus(t, http.MethodPost, "/api/v1/articles/"+artID+"/publish", w, http.StatusOK)
}

// TestRound3Config 覆盖 config_handler.go：平台配置读取与更新（含快照恢复）。
func TestRound3Config(t *testing.T) {
	round3PlatformSnapshot(t)
	app := newBizServer(t)
	adminTok := authAs(t, "admin-1", domain.RolePlatformAdmin)

	w := doRaw(app, http.MethodGet, "/api/v1/admin/config", "", adminTok)
	assertStatus(t, http.MethodGet, "/api/v1/admin/config", w, http.StatusOK)

	w = doRaw(app, http.MethodPost, "/api/v1/admin/config",
		`{"match_fee_rate":2,"match_fee_note":"round3 撮合费率"}`, adminTok)
	assertStatus(t, http.MethodPost, "/api/v1/admin/config", w, http.StatusOK)
}

// TestRound3AdminUsers 覆盖 admin_users.go：创建用户 → 列表 → 改角色 → 删除。
func TestRound3AdminUsers(t *testing.T) {
	app := newBizServer(t)
	adminTok := authAs(t, "admin-1", domain.RolePlatformAdmin)

	w := doRaw(app, http.MethodPost, "/api/v1/admin/users",
		`{"id":"user-round3","role":"individual"}`, adminTok)
	assertStatus(t, http.MethodPost, "/api/v1/admin/users", w, http.StatusCreated)

	w = doRaw(app, http.MethodGet, "/api/v1/admin/users", "", adminTok)
	assertStatus(t, http.MethodGet, "/api/v1/admin/users", w, http.StatusOK)

	w = doRaw(app, http.MethodPost, "/api/v1/admin/users/user-round3/role",
		`{"role":"enterprise"}`, adminTok)
	assertStatus(t, http.MethodPost, "/api/v1/admin/users/user-round3/role", w, http.StatusOK)

	w = doRaw(app, http.MethodDelete, "/api/v1/admin/users/user-round3", "", adminTok)
	assertStatus(t, http.MethodDelete, "/api/v1/admin/users/user-round3", w, http.StatusOK)
}

// TestRound3Messages 覆盖 messages.go：定向消息标记已读 → 广播 → 列表 → 未读数。
func TestRound3Messages(t *testing.T) {
	app := newBizServer(t)
	adminTok := authAs(t, "admin-1", domain.RolePlatformAdmin)
	userTok := authAs(t, "user-1", domain.RoleIndividual)

	// 定向消息 → 收件人标记已读（先于广播，避免时间戳碰撞使 MarkRead 归属校验失配）
	w := doRaw(app, http.MethodPost, "/api/v1/admin/messages",
		`{"sender_id":"system","receiver_id":"user-1","title":"round3直发","content":"hello"}`, adminTok)
	assertStatus(t, http.MethodPost, "/api/v1/admin/messages (direct)", w, http.StatusCreated)
	msgID := dataID(t, w)

	w = doRaw(app, http.MethodPost, "/api/v1/messages/"+msgID+"/read", "", userTok)
	assertStatus(t, http.MethodPost, "/api/v1/messages/"+msgID+"/read", w, http.StatusOK)

	// 广播（receiver_id 留空）
	w = doRaw(app, http.MethodPost, "/api/v1/admin/messages",
		`{"sender_id":"system","receiver_id":"","title":"round3公告","content":"版本升级"}`, adminTok)
	assertStatus(t, http.MethodPost, "/api/v1/admin/messages", w, http.StatusCreated)

	w = doRaw(app, http.MethodGet, "/api/v1/messages", "", adminTok)
	assertStatus(t, http.MethodGet, "/api/v1/messages", w, http.StatusOK)

	w = doRaw(app, http.MethodGet, "/api/v1/messages/unread-count", "", adminTok)
	assertStatus(t, http.MethodGet, "/api/v1/messages/unread-count", w, http.StatusOK)
}

// TestRound3ServiceListings 覆盖 service_listing.go：公开列表/详情（200/404）。
func TestRound3ServiceListings(t *testing.T) {
	app := newBizServer(t)
	adminTok := authAs(t, "admin-1", domain.RolePlatformAdmin)
	userTok := authAs(t, "user-1", domain.RoleIndividual)

	// 公开列表（空列表）
	w := doRaw(app, http.MethodGet, "/api/v1/service-listings", "", userTok)
	assertStatus(t, http.MethodGet, "/api/v1/service-listings", w, http.StatusOK)

	// 不存在详情 → 404
	w = doRaw(app, http.MethodGet, "/api/v1/service-listings/sl-nope", "", userTok)
	assertStatus(t, http.MethodGet, "/api/v1/service-listings/sl-nope", w, http.StatusNotFound)

	// 管理端创建（默认上架）→ 公开详情 200
	w = doRaw(app, http.MethodPost, "/api/v1/admin/service-listings",
		`{"provider_name":"round3服务商","title":"航拍巡检服务","category":"巡检","region":"重庆"}`, adminTok)
	assertStatus(t, http.MethodPost, "/api/v1/admin/service-listings", w, http.StatusCreated)
	slID := dataID(t, w)

	w = doRaw(app, http.MethodGet, "/api/v1/service-listings/"+slID, "", userTok)
	assertStatus(t, http.MethodGet, "/api/v1/service-listings/"+slID, w, http.StatusOK)
}
