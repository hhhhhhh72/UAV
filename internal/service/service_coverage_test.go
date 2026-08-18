package service_test

import (
	"context"
	"math"
	"strings"
	"testing"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

// ─────────────────────────────────────────────────────────────────────────────
// 辅助函数
// ─────────────────────────────────────────────────────────────────────────────

func closeScore(got, want float64) bool {
	return math.Abs(got-want) < 1e-6
}

func hasReason(reasons []string, want string) bool {
	for _, r := range reasons {
		if strings.Contains(r, want) {
			return true
		}
	}
	return false
}

// publishDemandForTest 创建并通过审核发布一条需求，供 work_order / intent 测试使用。
func publishDemandForTest(t *testing.T, demandSvc *service.DemandService, pub domain.Actor, title string) domain.Demand {
	t.Helper()
	d, err := demandSvc.Create(context.Background(), pub, service.CreateDemandInput{
		PublisherName: "测试企业", Contact: "13800000000", Title: title,
	})
	if err != nil {
		t.Fatalf("publishDemandForTest create: %v", err)
	}
	if _, err := demandSvc.Approve(context.Background(), domain.Actor{ID: "admin", Role: domain.RolePlatformAdmin}, d.ID); err != nil {
		t.Fatalf("publishDemandForTest approve: %v", err)
	}
	return d
}

// ─────────────────────────────────────────────────────────────────────────────
// applications.go — ApplicationService
// ─────────────────────────────────────────────────────────────────────────────

func TestApplicationService_CreateGetListMineListAll(t *testing.T) {
	svc := service.NewApplicationService(memory.NewApplicationRepository())

	app := domain.Application{
		ID: "app-1", UserID: "user-1", ServiceID: "svc-1",
		ServiceName: "测试服务", OrderNo: "NO-1", Status: "pending",
		FormData: map[string]any{"field": "value"},
	}
	created, err := svc.Create(context.Background(), app)
	if err != nil {
		t.Fatalf("ApplicationService.Create: %v", err)
	}
	if created.ID != "app-1" {
		t.Fatalf("ApplicationService.Create: got id %q, want app-1", created.ID)
	}

	got, err := svc.Get(context.Background(), created.ID)
	if err != nil || got.ID != "app-1" {
		t.Fatalf("ApplicationService.Get: id=%q err=%v", got.ID, err)
	}
	if _, err := svc.Get(context.Background(), "nope"); err == nil {
		t.Fatal("ApplicationService.Get: expected error for unknown id")
	}

	mine, total, err := svc.ListMine(context.Background(), "user-1", 1, 10)
	if err != nil || total != 1 || len(mine) != 1 {
		t.Fatalf("ApplicationService.ListMine: total=%d len=%d err=%v", total, len(mine), err)
	}
	none, total, err := svc.ListMine(context.Background(), "other", 1, 10)
	if err != nil || total != 0 || len(none) != 0 {
		t.Fatalf("ApplicationService.ListMine(other): total=%d len=%d err=%v", total, len(none), err)
	}

	all, total, err := svc.ListAll(context.Background(), 1, 10)
	if err != nil || total != 1 || len(all) != 1 {
		t.Fatalf("ApplicationService.ListAll: total=%d len=%d err=%v", total, len(all), err)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// auth.go — WeChatLogin / HashToken / GenerateRefreshToken
// ─────────────────────────────────────────────────────────────────────────────

func TestWeChatLogin_MissingCredentials(t *testing.T) {
	// 空 appID / appSecret 在任何网络调用前就报错，无需拦截 HTTPS。
	if _, err := service.WeChatLogin("code", "", "secret"); err == nil {
		t.Fatal("WeChatLogin: expected error for empty appID")
	}
	if _, err := service.WeChatLogin("code", "appid", ""); err == nil {
		t.Fatal("WeChatLogin: expected error for empty appSecret")
	}
}

func TestHashToken(t *testing.T) {
	h := service.HashToken("token-abc")
	if h == "" {
		t.Fatal("HashToken: empty hash")
	}
	if h != service.HashToken("token-abc") {
		t.Fatal("HashToken: not deterministic for same input")
	}
	if h == service.HashToken("token-def") {
		t.Fatal("HashToken: different tokens produced identical hash")
	}
	// SHA-256 (32 bytes) base64 RawURLEncoding → 43 chars
	if len(h) != 43 {
		t.Fatalf("HashToken: expected 43 chars, got %d", len(h))
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	t1, err := service.GenerateRefreshToken()
	if err != nil {
		t.Fatalf("GenerateRefreshToken: %v", err)
	}
	if len(t1) != 43 {
		t.Fatalf("GenerateRefreshToken: expected 43 chars, got %d", len(t1))
	}
	t2, err := service.GenerateRefreshToken()
	if err != nil {
		t.Fatalf("GenerateRefreshToken (2nd): %v", err)
	}
	if t1 == t2 {
		t.Fatal("GenerateRefreshToken: consecutive tokens should be unique")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// biz_batch1.go — ResourcePool / TestSite / Exhibition
// ─────────────────────────────────────────────────────────────────────────────

func TestResourcePoolService_CRUD(t *testing.T) {
	svc := service.NewResourcePoolService(memory.NewResourcePoolRepository())

	p, err := svc.Create(context.Background(), "无人机资源池", "equipment", "巡检设备池", "owner-1")
	if err != nil {
		t.Fatalf("ResourcePoolService.Create: %v", err)
	}
	if p.Status != "active" {
		t.Fatalf("ResourcePoolService.Create: status=%q, want active", p.Status)
	}

	list, err := svc.List(context.Background(), "equipment")
	if err != nil || len(list) != 1 {
		t.Fatalf("ResourcePoolService.List: len=%d err=%v", len(list), err)
	}
	if empty, _ := svc.List(context.Background(), "team"); len(empty) != 0 {
		t.Fatalf("ResourcePoolService.List(team): expected 0, got %d", len(empty))
	}

	got, err := svc.Get(context.Background(), p.ID)
	if err != nil || got.ID != p.ID {
		t.Fatalf("ResourcePoolService.Get: id=%q err=%v", got.ID, err)
	}
	if _, err := svc.Get(context.Background(), "nope"); err == nil {
		t.Fatal("ResourcePoolService.Get: expected error for unknown id")
	}

	m, err := svc.AddMember(context.Background(), p.ID, "res-1", "drone", 3)
	if err != nil {
		t.Fatalf("ResourcePoolService.AddMember: %v", err)
	}
	if m.Status != "standby" {
		t.Fatalf("ResourcePoolService.AddMember: status=%q, want standby", m.Status)
	}
	members, err := svc.ListMembers(context.Background(), p.ID)
	if err != nil || len(members) != 1 {
		t.Fatalf("ResourcePoolService.ListMembers: len=%d err=%v", len(members), err)
	}
	if none, _ := svc.ListMembers(context.Background(), "nope"); len(none) != 0 {
		t.Fatalf("ResourcePoolService.ListMembers(nope): expected 0, got %d", len(none))
	}
}

func TestTestSiteService_CRUD_BookingConflict(t *testing.T) {
	svc := service.NewTestSiteService(memory.NewTestSiteRepository())

	site, err := svc.Create(context.Background(), "试飞场", "flying_field", "重庆", "工作日9-18点", "owner-1", 10000, []string{"5G", "RTK"}, "")
	if err != nil {
		t.Fatalf("TestSiteService.Create: %v", err)
	}
	if site.Status != "available" {
		t.Fatalf("TestSiteService.Create: status=%q, want available (empty default)", site.Status)
	}

	if list, _ := svc.List(context.Background(), "flying_field"); len(list) != 1 {
		t.Fatalf("TestSiteService.List: len=%d, want 1", len(list))
	}
	if _, err := svc.Get(context.Background(), site.ID); err != nil {
		t.Fatalf("TestSiteService.Get: %v", err)
	}
	if _, err := svc.Get(context.Background(), "nope"); err == nil {
		t.Fatal("TestSiteService.Get: expected error for unknown id")
	}

	upd, err := svc.UpdateSite(context.Background(), site.ID, "试飞场2", "lab", "南岸", "提前3天", "maintenance", 20000, []string{"RTK"})
	if err != nil || upd.Name != "试飞场2" || upd.Status != "maintenance" {
		t.Fatalf("TestSiteService.UpdateSite: name=%q status=%q err=%v", upd.Name, upd.Status, err)
	}
	if _, err := svc.UpdateSite(context.Background(), "nope", "", "", "", "", "", 0, nil); err == nil {
		t.Fatal("TestSiteService.UpdateSite: expected error for unknown id")
	}

	// 预约 + 冲突校验
	base := time.Now().Add(48 * time.Hour)
	start := base
	end := base.Add(2 * time.Hour)

	bk, err := svc.Book(context.Background(), site.ID, "user-1", "R&D", "张三", "138", start, end)
	if err != nil {
		t.Fatalf("TestSiteService.Book: %v", err)
	}
	if bk.Status != "pending" {
		t.Fatalf("TestSiteService.Book: status=%q, want pending", bk.Status)
	}
	if _, err := svc.ReviewBooking(context.Background(), bk.ID, "approved", "同意"); err != nil {
		t.Fatalf("TestSiteService.ReviewBooking: %v", err)
	}

	// 与已批准预约重叠 → 冲突报错
	if _, err := svc.Book(context.Background(), site.ID, "user-2", "demo", "李四", "139", start.Add(time.Hour), end.Add(time.Hour)); err == nil {
		t.Fatal("TestSiteService.Book: expected time slot conflict error")
	}
	// 不重叠 → 成功
	if _, err := svc.Book(context.Background(), site.ID, "user-2", "demo", "李四", "139", start.Add(3*time.Hour), end.Add(4*time.Hour)); err != nil {
		t.Fatalf("TestSiteService.Book (non-overlap): %v", err)
	}

	if list, _ := svc.ListBookings(context.Background(), site.ID); len(list) != 2 {
		t.Fatalf("TestSiteService.ListBookings: len=%d, want 2", len(list))
	}
	if _, total, err := svc.ListAllBookings(context.Background(), 0, 10); err != nil || total != 2 {
		t.Fatalf("TestSiteService.ListAllBookings: total=%d err=%v", total, err)
	}
	if mine, _ := svc.ListMyBookings(context.Background(), "user-1"); len(mine) != 1 {
		t.Fatalf("TestSiteService.ListMyBookings(user-1): len=%d, want 1", len(mine))
	}
	if mine, _ := svc.ListMyBookings(context.Background(), "user-2"); len(mine) != 1 {
		t.Fatalf("TestSiteService.ListMyBookings(user-2): len=%d, want 1", len(mine))
	}

	if err := svc.DeleteSite(context.Background(), site.ID); err != nil {
		t.Fatalf("TestSiteService.DeleteSite: %v", err)
	}
	if err := svc.DeleteSite(context.Background(), "nope"); err == nil {
		t.Fatal("TestSiteService.DeleteSite: expected error for unknown id")
	}
}

func TestExhibitionService_CRUD_Booths(t *testing.T) {
	svc := service.NewExhibitionService(memory.NewExhibitionRepository())
	start := time.Now().AddDate(0, 1, 0)
	end := start.AddDate(0, 0, 3)

	e, err := svc.Create(context.Background(), "无人机展", "drone_show", "描述", "重庆", "协会", "", start, end, 10, 5000, "")
	if err != nil {
		t.Fatalf("ExhibitionService.Create: %v", err)
	}
	if e.Status != "draft" {
		t.Fatalf("ExhibitionService.Create: status=%q, want draft (empty default)", e.Status)
	}

	if _, total, err := svc.List(context.Background(), 1, 10); err != nil || total != 1 {
		t.Fatalf("ExhibitionService.List: total=%d err=%v", total, err)
	}
	if _, err := svc.Get(context.Background(), e.ID); err != nil {
		t.Fatalf("ExhibitionService.Get: %v", err)
	}
	if _, err := svc.Get(context.Background(), "nope"); err == nil {
		t.Fatal("ExhibitionService.Get: expected error for unknown id")
	}

	upd, err := svc.Update(context.Background(), e.ID, "无人机展2", "equipment_expo", "新描述", "北京", "主办", start, end, 20, 8000, "recruiting")
	if err != nil || upd.Title != "无人机展2" || upd.Status != "recruiting" {
		t.Fatalf("ExhibitionService.Update: title=%q status=%q err=%v", upd.Title, upd.Status, err)
	}
	if _, err := svc.Update(context.Background(), "nope", "", "", "", "", "", start, end, 0, 0, ""); err == nil {
		t.Fatal("ExhibitionService.Update: expected error for unknown id")
	}

	b, err := svc.ApplyBooth(context.Background(), e.ID, "ent-1", "A1", "展品", "展品说明")
	if err != nil {
		t.Fatalf("ExhibitionService.ApplyBooth: %v", err)
	}
	if b.Status != "applied" {
		t.Fatalf("ExhibitionService.ApplyBooth: status=%q, want applied", b.Status)
	}
	if list, _ := svc.ListBooths(context.Background(), e.ID); len(list) != 1 {
		t.Fatalf("ExhibitionService.ListBooths: len=%d, want 1", len(list))
	}
	rb, err := svc.ReviewBooth(context.Background(), b.ID, "approved")
	if err != nil || rb.Status != "approved" {
		t.Fatalf("ExhibitionService.ReviewBooth: status=%q err=%v", rb.Status, err)
	}
	if _, err := svc.ReviewBooth(context.Background(), "nope", "approved"); err == nil {
		t.Fatal("ExhibitionService.ReviewBooth: expected error for unknown booth")
	}

	if err := svc.Delete(context.Background(), e.ID); err != nil {
		t.Fatalf("ExhibitionService.Delete: %v", err)
	}
	if err := svc.Delete(context.Background(), "nope"); err == nil {
		t.Fatal("ExhibitionService.Delete: expected error for unknown id")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// biz_batch2.go — Transformation / College / Cooperation
// ─────────────────────────────────────────────────────────────────────────────

func TestTransformationService_CRUD_StageMilestone(t *testing.T) {
	svc := service.NewTransformationService(memory.NewTransformationRepository())
	owner := domain.Actor{ID: "owner-1", Role: domain.RoleEnterprise}
	admin := domain.Actor{ID: "admin", Role: domain.RolePlatformAdmin}

	tr, err := svc.Create(context.Background(), "成果转化", "ach-1", "owner-1", "partner-1")
	if err != nil {
		t.Fatalf("TransformationService.Create: %v", err)
	}
	if tr.Stage != domain.StageLab {
		t.Fatalf("TransformationService.Create: stage=%q, want lab", tr.Stage)
	}

	if _, err := svc.Get(context.Background(), tr.ID); err != nil {
		t.Fatalf("TransformationService.Get: %v", err)
	}
	if _, err := svc.Get(context.Background(), "nope"); err == nil {
		t.Fatal("TransformationService.Get: expected error for unknown id")
	}
	if list, _ := svc.List(context.Background(), "owner-1"); len(list) != 1 {
		t.Fatalf("TransformationService.List: len=%d, want 1", len(list))
	}
	if byAch, _ := svc.ListByAchievement(context.Background(), "ach-1"); len(byAch) != 1 {
		t.Fatalf("TransformationService.ListByAchievement: len=%d, want 1", len(byAch))
	}
	if byAch, _ := svc.ListByAchievement(context.Background(), "nope"); len(byAch) != 0 {
		t.Fatalf("TransformationService.ListByAchievement(nope): len=%d, want 0", len(byAch))
	}

	upd, err := svc.UpdateTrans(context.Background(), tr.ID, "新标题", "pilot", "50%", "partner-2", "active")
	if err != nil || upd.Title != "新标题" || upd.Stage != domain.StagePilot {
		t.Fatalf("TransformationService.UpdateTrans: title=%q stage=%q err=%v", upd.Title, upd.Stage, err)
	}
	if _, err := svc.UpdateTrans(context.Background(), "nope", "", "", "", "", ""); err == nil {
		t.Fatal("TransformationService.UpdateTrans: expected error for unknown id")
	}

	// AdvanceStage：不存在的转化 → 报错
	if _, err := svc.AdvanceStage(context.Background(), owner, "nope", domain.StagePilot, ""); err == nil {
		t.Fatal("TransformationService.AdvanceStage: expected error for unknown id")
	}
	// AdvanceStage：负责人可推进
	adv, err := svc.AdvanceStage(context.Background(), owner, tr.ID, domain.StageIndustrialized, "产业化")
	if err != nil || adv.Stage != domain.StageIndustrialized {
		t.Fatalf("TransformationService.AdvanceStage(owner): stage=%q err=%v", adv.Stage, err)
	}
	// 非负责人/非管理员 → 报错
	if _, err := svc.AdvanceStage(context.Background(), domain.Actor{ID: "other", Role: domain.RoleEnterprise}, tr.ID, domain.StageListed, ""); err == nil {
		t.Fatal("TransformationService.AdvanceStage: expected error for non-owner")
	}
	// 管理员可推进
	adv2, err := svc.AdvanceStage(context.Background(), admin, tr.ID, domain.StageListed, "上市")
	if err != nil || adv2.Stage != domain.StageListed {
		t.Fatalf("TransformationService.AdvanceStage(admin): stage=%q err=%v", adv2.Stage, err)
	}

	// AddMilestone：不存在的转化 → 报错
	if _, err := svc.AddMilestone(context.Background(), owner, "nope", "里程碑", ""); err == nil {
		t.Fatal("TransformationService.AddMilestone: expected error for unknown id")
	}
	// AddMilestone：负责人可添加
	m, err := svc.AddMilestone(context.Background(), owner, tr.ID, "里程碑1", "证据URL")
	if err != nil || len(m.Milestones) != 1 || !m.Milestones[0].Completed {
		t.Fatalf("TransformationService.AddMilestone: milestones=%d err=%v", len(m.Milestones), err)
	}
	// 非负责人 → 报错
	if _, err := svc.AddMilestone(context.Background(), domain.Actor{ID: "other", Role: domain.RoleEnterprise}, tr.ID, "里程碑2", ""); err == nil {
		t.Fatal("TransformationService.AddMilestone: expected error for non-owner")
	}

	if err := svc.DeleteTrans(context.Background(), tr.ID); err != nil {
		t.Fatalf("TransformationService.DeleteTrans: %v", err)
	}
	if err := svc.DeleteTrans(context.Background(), "nope"); err == nil {
		t.Fatal("TransformationService.DeleteTrans: expected error for unknown id")
	}
}

func TestCollegeService_CRUD(t *testing.T) {
	svc := service.NewCollegeService(memory.NewCollegeRepository())

	c, err := svc.Create(context.Background(), domain.College{Name: "重庆大学", Region: "重庆", City: "重庆"})
	if err != nil {
		t.Fatalf("CollegeService.Create: %v", err)
	}
	if c.ID == "" {
		t.Fatal("CollegeService.Create: expected auto-generated ID")
	}
	if c.Status != "active" {
		t.Fatalf("CollegeService.Create: status=%q, want active (empty default)", c.Status)
	}

	if list, _ := svc.List(context.Background(), "重庆"); len(list) != 1 {
		t.Fatalf("CollegeService.List: len=%d, want 1", len(list))
	}
	if list, _ := svc.List(context.Background(), "北京"); len(list) != 0 {
		t.Fatalf("CollegeService.List(北京): len=%d, want 0", len(list))
	}
	if _, err := svc.Get(context.Background(), c.ID); err != nil {
		t.Fatalf("CollegeService.Get: %v", err)
	}
	if _, err := svc.Get(context.Background(), "nope"); err == nil {
		t.Fatal("CollegeService.Get: expected error for unknown id")
	}

	c.Name = "重庆大学(更新)"
	upd, err := svc.Update(context.Background(), c)
	if err != nil || upd.Name != "重庆大学(更新)" {
		t.Fatalf("CollegeService.Update: name=%q err=%v", upd.Name, err)
	}
	if _, err := svc.Update(context.Background(), domain.College{ID: "nope"}); err == nil {
		t.Fatal("CollegeService.Update: expected error for unknown id")
	}

	if err := svc.Delete(context.Background(), c.ID); err != nil {
		t.Fatalf("CollegeService.Delete: %v", err)
	}
	if err := svc.Delete(context.Background(), "nope"); err == nil {
		t.Fatal("CollegeService.Delete: expected error for unknown id")
	}
}

func TestCoopTypeLabel(t *testing.T) {
	cases := map[string]string{
		service.CoopTypeResearch: "科研合作",
		service.CoopTypeTalent:   "人才培养",
		service.CoopTypeBoth:     "综合",
		"unknown":                "综合",
	}
	for in, want := range cases {
		if got := service.CoopTypeLabel(in); got != want {
			t.Fatalf("CoopTypeLabel(%q): got %q, want %q", in, got, want)
		}
	}
}

func TestCooperationService_CRUD(t *testing.T) {
	svc := service.NewCooperationService(memory.NewCooperationRepository())
	start := time.Now().AddDate(0, 1, 0)
	end := start.AddDate(0, 6, 0)

	cp, err := svc.Create(context.Background(), "校企共建", "col-1", "ent-1", "joint_lab", "描述", start, end, 30)
	if err != nil {
		t.Fatalf("CooperationService.Create: %v", err)
	}
	if cp.Status != "proposed" {
		t.Fatalf("CooperationService.Create: status=%q, want proposed", cp.Status)
	}

	if list, _ := svc.List(context.Background(), "ent-1"); len(list) != 1 {
		t.Fatalf("CooperationService.List: len=%d, want 1", len(list))
	}
	if list, _ := svc.List(context.Background(), "other"); len(list) != 0 {
		t.Fatalf("CooperationService.List(other): len=%d, want 0", len(list))
	}
	if _, err := svc.Get(context.Background(), cp.ID); err != nil {
		t.Fatalf("CooperationService.Get: %v", err)
	}
	if _, err := svc.Get(context.Background(), "nope"); err == nil {
		t.Fatal("CooperationService.Get: expected error for unknown id")
	}

	upd, err := svc.UpdateStatus(context.Background(), cp.ID, "active")
	if err != nil || upd.Status != "active" {
		t.Fatalf("CooperationService.UpdateStatus: status=%q err=%v", upd.Status, err)
	}
	if _, err := svc.UpdateStatus(context.Background(), "nope", "active"); err == nil {
		t.Fatal("CooperationService.UpdateStatus: expected error for unknown id")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// biz_batch3.go — RescueCase / EmergencyDept / AssociationMember
// ─────────────────────────────────────────────────────────────────────────────

func TestRescueCaseService_Get(t *testing.T) {
	svc := service.NewRescueCaseService(memory.NewRescueCaseRepository())

	rc, err := svc.Create(context.Background(), "救援案例", "山火", "重庆", "M300", "应急队", "摘要", "结果", "教训", "协会", time.Now())
	if err != nil {
		t.Fatalf("RescueCaseService.Create: %v", err)
	}
	got, err := svc.Get(context.Background(), rc.ID)
	if err != nil || got.ID != rc.ID {
		t.Fatalf("RescueCaseService.Get: id=%q err=%v", got.ID, err)
	}
	if _, err := svc.Get(context.Background(), "nope"); err == nil {
		t.Fatal("RescueCaseService.Get: expected error for unknown id")
	}
}

func TestEmergencyDeptService_CRUD(t *testing.T) {
	svc := service.NewEmergencyDeptService(memory.NewEmergencyDeptRepository())

	d, err := svc.CreateDept(context.Background(), "消防支队", "fire", "重庆", "张三", "138", "http://protocol")
	if err != nil {
		t.Fatalf("EmergencyDeptService.CreateDept: %v", err)
	}
	if d.Status != "active" {
		t.Fatalf("EmergencyDeptService.CreateDept: status=%q, want active", d.Status)
	}
	if depts, _ := svc.ListDepts(context.Background()); len(depts) != 1 {
		t.Fatalf("EmergencyDeptService.ListDepts: len=%d, want 1", len(depts))
	}

	drill, err := svc.CreateDrill(context.Background(), d.ID, "联合演练", "山火", time.Now(), 10, 5, "成功")
	if err != nil {
		t.Fatalf("EmergencyDeptService.CreateDrill: %v", err)
	}
	if drills, _ := svc.ListDrills(context.Background(), d.ID); len(drills) != 1 {
		t.Fatalf("EmergencyDeptService.ListDrills: len=%d, want 1", len(drills))
	}
	if drills, _ := svc.ListDrills(context.Background(), "nope"); len(drills) != 0 {
		t.Fatalf("EmergencyDeptService.ListDrills(nope): len=%d, want 0", len(drills))
	}
	_ = drill
}

func TestAssociationMemberService_Roles(t *testing.T) {
	svc := service.NewAssociationMemberService(memory.NewAssociationMemberRepository())

	m, err := svc.AddMember(context.Background(), "user-1", "ent-1", domain.AssocMember)
	if err != nil {
		t.Fatalf("AssociationMemberService.AddMember: %v", err)
	}
	if m.Role != domain.AssocMember {
		t.Fatalf("AssociationMemberService.AddMember: role=%q, want member", m.Role)
	}

	if _, total, err := svc.ListMembers(context.Background(), "member", 1, 10); err != nil || total != 1 {
		t.Fatalf("AssociationMemberService.ListMembers: total=%d err=%v", total, err)
	}
	got, err := svc.GetByUserID(context.Background(), "user-1")
	if err != nil || got.ID != m.ID {
		t.Fatalf("AssociationMemberService.GetByUserID: id=%q err=%v", got.ID, err)
	}
	if _, err := svc.GetByUserID(context.Background(), "nope"); err == nil {
		t.Fatal("AssociationMemberService.GetByUserID: expected error for unknown user")
	}

	// 8 级角色逐一 UpdateRole 校验
	roles := []domain.AssociationRole{
		domain.AssocPresident, domain.AssocVicePresident, domain.AssocSecretary,
		domain.AssocDeptHead, domain.AssocMember, domain.AssocPartner,
		domain.AssocCollege, domain.AssocGuest,
	}
	for _, r := range roles {
		up, err := svc.UpdateRole(context.Background(), m.ID, r)
		if err != nil || up.Role != r {
			t.Fatalf("AssociationMemberService.UpdateRole(%q): role=%q err=%v", r, up.Role, err)
		}
	}
	if _, err := svc.UpdateRole(context.Background(), "nope", domain.AssocMember); err == nil {
		t.Fatal("AssociationMemberService.UpdateRole: expected error for unknown id")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// biz_operations.go — Competition / Event / Resource / Emergency 补充分支
// ─────────────────────────────────────────────────────────────────────────────

func TestCompetitionService_DeleteUpdateRegister(t *testing.T) {
	svc := service.NewCompetitionService(memory.NewCompetitionRepository(nil))

	c, err := svc.Create(context.Background(), domain.Competition{Title: "竞速赛", Category: "racing"})
	if err != nil {
		t.Fatalf("CompetitionService.Create: %v", err)
	}
	// Register 不存在的赛事 → 报错
	if _, err := svc.Register(context.Background(), "nope", "user-1", "队", 1, "138", "张三", "138", "id", "", ""); err == nil {
		t.Fatal("CompetitionService.Register: expected error for unknown competition")
	}
	// Update 不存在的赛事 → 报错
	if _, err := svc.Update(context.Background(), domain.Competition{ID: "nope"}); err == nil {
		t.Fatal("CompetitionService.Update: expected error for unknown id")
	}
	// Delete
	if err := svc.Delete(context.Background(), c.ID); err != nil {
		t.Fatalf("CompetitionService.Delete: %v", err)
	}
	if err := svc.Delete(context.Background(), "nope"); err == nil {
		t.Fatal("CompetitionService.Delete: expected error for unknown id")
	}
}

func TestEventService_DeleteUpdateRegister(t *testing.T) {
	svc := service.NewEventService(memory.NewEventRepository())
	start := time.Now().AddDate(0, 1, 0)
	end := start.AddDate(0, 0, 1)

	e, err := svc.Create(context.Background(), "产业论坛", "forum", "描述", "地点", "", start, end, 100)
	if err != nil {
		t.Fatalf("EventService.Create: %v", err)
	}
	// Register 不存在的活动 → 报错
	if _, err := svc.Register(context.Background(), "nope", "user-1", "张三", "138", "公司"); err == nil {
		t.Fatal("EventService.Register: expected error for unknown event")
	}
	// Update 不存在的活动 → 报错
	if _, err := svc.Update(context.Background(), "nope", "t", "", "", "", "", "", start, end, 10); err == nil {
		t.Fatal("EventService.Update: expected error for unknown id")
	}
	if err := svc.Delete(context.Background(), e.ID); err != nil {
		t.Fatalf("EventService.Delete: %v", err)
	}
	if err := svc.Delete(context.Background(), "nope"); err == nil {
		t.Fatal("EventService.Delete: expected error for unknown id")
	}
}

func TestResourceService_DeleteBookAndVisibility(t *testing.T) {
	svc := service.NewResourceService(memory.NewResourceRepository())

	// 空 visibilityLevel 默认 public
	r, err := svc.Create(context.Background(), "owner-1", "无人机", "drone", "M300", "规格", "地点", "预约信息", 10000, "")
	if err != nil {
		t.Fatalf("ResourceService.Create: %v", err)
	}
	if r.VisibilityLevel != "public" {
		t.Fatalf("ResourceService.Create: visibility=%q, want public (empty default)", r.VisibilityLevel)
	}
	// Book 不存在的资源 → ErrResourceNotFound
	if _, err := svc.Book(context.Background(), "user-1", "nope", "2026-01-01", "用途", "张三", "138"); err == nil {
		t.Fatal("ResourceService.Book: expected error for unknown resource")
	}
	if err := svc.Delete(context.Background(), r.ID); err != nil {
		t.Fatalf("ResourceService.Delete: %v", err)
	}
	if err := svc.Delete(context.Background(), "nope"); err == nil {
		t.Fatal("ResourceService.Delete: expected error for unknown id")
	}
}

func TestEmergencyService_FullCRUD(t *testing.T) {
	svc := service.NewEmergencyService(memory.NewEmergencyRepository())

	r, err := svc.CreateResource(context.Background(), "owner-1", "应急无人机", "drone", "M300", "南岸", "138", 2)
	if err != nil {
		t.Fatalf("EmergencyService.CreateResource: %v", err)
	}

	// UpdateResource
	up, err := svc.UpdateResource(context.Background(), r.ID, "应急无人机2", "无人机", "M300RTK", "渝中", "139", "engaged", 5)
	if err != nil || up.Name != "应急无人机2" || up.ResType != "drone" || up.Status != "engaged" {
		t.Fatalf("EmergencyService.UpdateResource: name=%q type=%q status=%q err=%v", up.Name, up.ResType, up.Status, err)
	}
	if _, err := svc.UpdateResource(context.Background(), "nope", "", "", "", "", "", "", 0); err == nil {
		t.Fatal("EmergencyService.UpdateResource: expected error for unknown id")
	}

	// CreateDispatch 不存在的资源 → 报错
	now := time.Now()
	if _, err := svc.CreateDispatch(context.Background(), "nope", "事件", "地点", "指挥", "结果", now, now.Add(time.Hour)); err == nil {
		t.Fatal("EmergencyService.CreateDispatch: expected error for unknown resource")
	}
	d, err := svc.CreateDispatch(context.Background(), r.ID, "山火应急", "北碚", "指挥", "结果", now, now.Add(time.Hour))
	if err != nil {
		t.Fatalf("EmergencyService.CreateDispatch: %v", err)
	}

	// FindDispatchByID
	fd, err := svc.FindDispatchByID(context.Background(), d.ID)
	if err != nil || fd.ID != d.ID {
		t.Fatalf("EmergencyService.FindDispatchByID: id=%q err=%v", fd.ID, err)
	}
	if _, err := svc.FindDispatchByID(context.Background(), "nope"); err == nil {
		t.Fatal("EmergencyService.FindDispatchByID: expected error for unknown id")
	}

	// UpdateDispatch
	ud, err := svc.UpdateDispatch(context.Background(), d.ID, r.ID, "新事件", "新地点", "新指挥", "新结果", "completed", now, now.Add(2*time.Hour))
	if err != nil || ud.Status != "completed" || ud.EventDesc != "新事件" {
		t.Fatalf("EmergencyService.UpdateDispatch: status=%q desc=%q err=%v", ud.Status, ud.EventDesc, err)
	}
	if _, err := svc.UpdateDispatch(context.Background(), "nope", r.ID, "", "", "", "", "", now, now); err == nil {
		t.Fatal("EmergencyService.UpdateDispatch: expected error for unknown id")
	}

	if err := svc.DeleteDispatch(context.Background(), d.ID); err != nil {
		t.Fatalf("EmergencyService.DeleteDispatch: %v", err)
	}
	if err := svc.DeleteDispatch(context.Background(), "nope"); err == nil {
		t.Fatal("EmergencyService.DeleteDispatch: expected error for unknown id")
	}
	if err := svc.DeleteResource(context.Background(), r.ID); err != nil {
		t.Fatalf("EmergencyService.DeleteResource: %v", err)
	}
	if err := svc.DeleteResource(context.Background(), "nope"); err == nil {
		t.Fatal("EmergencyService.DeleteResource: expected error for unknown id")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// biz_innovation.go — Achievement / RDChallenge / ResearchProject / ProjectApp
// ─────────────────────────────────────────────────────────────────────────────

func TestAchievementService_UpdateError(t *testing.T) {
	svc := service.NewAchievementService(memory.NewAchievementRepository())
	if _, err := svc.Update(context.Background(), domain.Actor{ID: "u-x", Role: domain.RoleIndividual}, "nope", "t", "", "", "", "", "", nil, nil); err == nil {
		t.Fatal("AchievementService.Update: expected error for unknown id")
	}
}

func TestRDChallengeService_DeleteUpdate(t *testing.T) {
	svc := service.NewRDChallengeService(memory.NewRDChallengeRepository())
	c, err := svc.Create(context.Background(), "ent-1", "长续航电池", "电池", "描述", 500000, time.Now().AddDate(0, 3, 0))
	if err != nil {
		t.Fatalf("RDChallengeService.Create: %v", err)
	}
	if _, err := svc.Update(context.Background(), domain.Actor{ID: "ent-1", Role: domain.RoleEnterprise}, "nope", "", "", "", "", 0, time.Now()); err == nil {
		t.Fatal("RDChallengeService.Update: expected error for unknown id")
	}
	if err := svc.Delete(context.Background(), domain.Actor{ID: "ent-1", Role: domain.RoleEnterprise}, c.ID); err != nil {
		t.Fatalf("RDChallengeService.Delete: %v", err)
	}
	if err := svc.Delete(context.Background(), domain.Actor{ID: "ent-1", Role: domain.RoleEnterprise}, "nope"); err == nil {
		t.Fatal("RDChallengeService.Delete: expected error for unknown id")
	}
}

func TestResearchProjectService_DeleteUpdate(t *testing.T) {
	svc := service.NewResearchProjectService(memory.NewResearchProjectRepository())
	p, err := svc.Create(context.Background(), "航路规划", "空域", "描述", "重庆大学", "Q1调研", []string{"巡航科技"}, 200000, time.Now(), time.Now().AddDate(1, 0, 0))
	if err != nil {
		t.Fatalf("ResearchProjectService.Create: %v", err)
	}
	if _, err := svc.Update(context.Background(), "nope", "", "", "", "", "", "", nil, 0, time.Now(), time.Now()); err == nil {
		t.Fatal("ResearchProjectService.Update: expected error for unknown id")
	}
	if err := svc.Delete(context.Background(), p.ID); err != nil {
		t.Fatalf("ResearchProjectService.Delete: %v", err)
	}
	if err := svc.Delete(context.Background(), "nope"); err == nil {
		t.Fatal("ResearchProjectService.Delete: expected error for unknown id")
	}
}

func TestProjectAppService_ReviewBranches(t *testing.T) {
	svc := service.NewProjectAppService(memory.NewProjectAppRepository())
	a, err := svc.Create(context.Background(), "user-1", "示范项目", "示范", "描述", 1000000, nil)
	if err != nil {
		t.Fatalf("ProjectAppService.Create: %v", err)
	}
	// approve
	ap, err := svc.Review(context.Background(), a.ID, "同意", "approve")
	if err != nil || ap.Status != "approved" {
		t.Fatalf("ProjectAppService.Review(approve): status=%q err=%v", ap.Status, err)
	}
	// reject
	rj, err := svc.Review(context.Background(), a.ID, "材料不足", "reject")
	if err != nil || rj.Status != "rejected" {
		t.Fatalf("ProjectAppService.Review(reject): status=%q err=%v", rj.Status, err)
	}
	// invalid action
	if _, err := svc.Review(context.Background(), a.ID, "", "bogus"); err == nil {
		t.Fatal("ProjectAppService.Review: expected error for invalid action")
	}
	// unknown id
	if _, err := svc.Review(context.Background(), "nope", "", "approve"); err == nil {
		t.Fatal("ProjectAppService.Review: expected error for unknown id")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// work_order.go / intent.go / matching.go 未覆盖分支
// ─────────────────────────────────────────────────────────────────────────────

func newWOScenario(t *testing.T) (*service.WorkOrderService, *service.IntentService, *service.DemandService, domain.Actor, domain.Actor) {
	t.Helper()
	demandRepo := memory.NewDemandRepository(nil)
	intentRepo := memory.NewIntentRepository()
	orderRepo := memory.NewWorkOrderRepository()
	demandSvc := service.NewDemandService(demandRepo)
	intentSvc := service.NewIntentService(intentRepo, demandRepo)
	orderSvc := service.NewWorkOrderService(orderRepo, demandRepo, intentRepo)
	pub := domain.Actor{ID: "pub-1", Role: domain.RoleEnterprise}
	worker := domain.Actor{ID: "worker-1", Role: domain.RoleIndividual}
	return orderSvc, intentSvc, demandSvc, pub, worker
}

func TestWorkOrder_AcceptIntentErrors(t *testing.T) {
	orderSvc, intentSvc, demandSvc, pub, worker := newWOScenario(t)

	// 需求不存在
	if _, err := orderSvc.AcceptIntent(context.Background(), pub, "nope", "intent-1", 0); err == nil {
		t.Fatal("AcceptIntent: expected error for unknown demand")
	}
	// 需求未发布（pending）
	pd, _ := demandSvc.Create(context.Background(), pub, service.CreateDemandInput{PublisherName: "企业", Contact: "138", Title: "待审核需求"})
	if _, err := orderSvc.AcceptIntent(context.Background(), pub, pd.ID, "intent-1", 0); err == nil {
		t.Fatal("AcceptIntent: expected error for unpublished demand")
	}
	// 正常场景 + 重复确认报错
	d := publishDemandForTest(t, demandSvc, pub, "正常需求")
	it, err := intentSvc.Create(context.Background(), worker, d.ID, service.CreateIntentInput{IntentorName: "飞手", Contact: "139"})
	if err != nil {
		t.Fatalf("IntentService.Create: %v", err)
	}
	wo, err := orderSvc.AcceptIntent(context.Background(), pub, d.ID, it.ID, 100000)
	if err != nil || wo.Status != domain.WorkOrderPending {
		t.Fatalf("AcceptIntent: status=%q err=%v", wo.Status, err)
	}
	// 意向已被确认 → 重复确认报错
	if _, err := orderSvc.AcceptIntent(context.Background(), pub, d.ID, it.ID, 200000); err == nil {
		t.Fatal("AcceptIntent: expected error when intent already processed")
	}
}

func TestWorkOrder_StatusPreconditions(t *testing.T) {
	orderSvc, intentSvc, demandSvc, pub, worker := newWOScenario(t)
	d := publishDemandForTest(t, demandSvc, pub, "状态机需求")
	it, _ := intentSvc.Create(context.Background(), worker, d.ID, service.CreateIntentInput{IntentorName: "飞手", Contact: "139"})
	wo, _ := orderSvc.AcceptIntent(context.Background(), pub, d.ID, it.ID, 100000)

	// StartWork：订单不存在 / 非飞手
	if _, err := orderSvc.StartWork(context.Background(), worker, "nope"); err == nil {
		t.Fatal("StartWork: expected error for unknown order")
	}
	if _, err := orderSvc.StartWork(context.Background(), pub, wo.ID); err == nil {
		t.Fatal("StartWork: expected error when publisher (non-worker) starts")
	}
	// CompleteWork：订单不存在
	if _, err := orderSvc.CompleteWork(context.Background(), worker, "nope", nil); err == nil {
		t.Fatal("CompleteWork: expected error for unknown order")
	}
	// CompleteWork：pending 状态不允许（状态前置校验）/ 非飞手
	if _, err := orderSvc.CompleteWork(context.Background(), worker, wo.ID, nil); err == nil {
		t.Fatal("CompleteWork: expected error when order not ongoing")
	}
	if _, err := orderSvc.CompleteWork(context.Background(), pub, wo.ID, nil); err == nil {
		t.Fatal("CompleteWork: expected error when non-worker completes")
	}
	// AcceptCompletion：pending 状态不允许
	if _, err := orderSvc.AcceptCompletion(context.Background(), pub, wo.ID); err == nil {
		t.Fatal("AcceptCompletion: expected error when order not awaiting_accept")
	}
	// RequestRework：pending 状态不允许
	if _, err := orderSvc.RequestRework(context.Background(), pub, wo.ID, "note"); err == nil {
		t.Fatal("RequestRework: expected error when order not awaiting_accept")
	}

	// 推进到 awaiting_accept
	wo, _ = orderSvc.StartWork(context.Background(), worker, wo.ID)
	wo, _ = orderSvc.CompleteWork(context.Background(), worker, wo.ID, []string{"a.jpg"})
	if wo.Status != domain.WorkOrderAwaitingAccept {
		t.Fatalf("CompleteWork: status=%q, want awaiting_accept", wo.Status)
	}
	// AcceptCompletion：非发布者
	if _, err := orderSvc.AcceptCompletion(context.Background(), worker, wo.ID); err == nil {
		t.Fatal("AcceptCompletion: expected error when non-publisher accepts")
	}
	// RequestRework：非发布者
	if _, err := orderSvc.RequestRework(context.Background(), worker, wo.ID, "note"); err == nil {
		t.Fatal("RequestRework: expected error when non-publisher reworks")
	}

	// 验收通过 → completed
	wo, _ = orderSvc.AcceptCompletion(context.Background(), pub, wo.ID)
	if wo.Status != domain.WorkOrderCompleted {
		t.Fatalf("AcceptCompletion: status=%q, want completed", wo.Status)
	}
	// completed 订单不能取消
	if _, err := orderSvc.RequestCancel(context.Background(), pub, wo.ID, "reason"); err == nil {
		t.Fatal("RequestCancel: expected error when order completed")
	}
}

func TestWorkOrder_RejectIntentErrors(t *testing.T) {
	orderSvc, intentSvc, demandSvc, pub, worker := newWOScenario(t)

	// 需求不存在
	if err := orderSvc.RejectIntent(context.Background(), pub, "nope", "intent-1"); err == nil {
		t.Fatal("RejectIntent: expected error for unknown demand")
	}
	d := publishDemandForTest(t, demandSvc, pub, "拒绝意向需求")
	it, _ := intentSvc.Create(context.Background(), worker, d.ID, service.CreateIntentInput{IntentorName: "飞手", Contact: "139"})
	// 非发布者
	if err := orderSvc.RejectIntent(context.Background(), worker, d.ID, it.ID); err == nil {
		t.Fatal("RejectIntent: expected error for non-publisher")
	}
	// 意向不存在
	if err := orderSvc.RejectIntent(context.Background(), pub, d.ID, "nope"); err == nil {
		t.Fatal("RejectIntent: expected error for unknown intent")
	}
	// 正常拒绝
	if err := orderSvc.RejectIntent(context.Background(), pub, d.ID, it.ID); err != nil {
		t.Fatalf("RejectIntent: %v", err)
	}
}

func TestWorkOrder_FindByID(t *testing.T) {
	orderSvc, intentSvc, demandSvc, pub, worker := newWOScenario(t)
	d := publishDemandForTest(t, demandSvc, pub, "订单详情需求")
	it, _ := intentSvc.Create(context.Background(), worker, d.ID, service.CreateIntentInput{IntentorName: "飞手", Contact: "139"})
	wo, _ := orderSvc.AcceptIntent(context.Background(), pub, d.ID, it.ID, 100000)

	if _, err := orderSvc.FindByID(context.Background(), pub, wo.ID); err != nil {
		t.Fatalf("FindByID(publisher): %v", err)
	}
	if _, err := orderSvc.FindByID(context.Background(), worker, wo.ID); err != nil {
		t.Fatalf("FindByID(worker): %v", err)
	}
	if _, err := orderSvc.FindByID(context.Background(), domain.Actor{ID: "stranger", Role: domain.RoleIndividual}, wo.ID); err == nil {
		t.Fatal("FindByID: expected error for non-party user")
	}
	if _, err := orderSvc.FindByID(context.Background(), pub, "nope"); err == nil {
		t.Fatal("FindByID: expected error for unknown order")
	}

	// RequestCancel：非订单双方
	if _, err := orderSvc.RequestCancel(context.Background(), domain.Actor{ID: "stranger", Role: domain.RoleIndividual}, wo.ID, "reason"); err == nil {
		t.Fatal("RequestCancel: expected error for non-party user")
	}
}

func TestIntentService_CreateErrors(t *testing.T) {
	demandRepo := memory.NewDemandRepository(nil)
	intentRepo := memory.NewIntentRepository()
	demandSvc := service.NewDemandService(demandRepo)
	intentSvc := service.NewIntentService(intentRepo, demandRepo)
	pub := domain.Actor{ID: "pub-1", Role: domain.RoleEnterprise}
	worker := domain.Actor{ID: "worker-1", Role: domain.RoleIndividual}

	// 空 demand_id
	if _, err := intentSvc.Create(context.Background(), worker, "", service.CreateIntentInput{Contact: "138"}); err == nil {
		t.Fatal("IntentService.Create: expected error for empty demand_id")
	}
	// 空 contact
	if _, err := intentSvc.Create(context.Background(), worker, "d", service.CreateIntentInput{Contact: ""}); err == nil {
		t.Fatal("IntentService.Create: expected error for empty contact")
	}
	// 需求不存在
	if _, err := intentSvc.Create(context.Background(), worker, "nope", service.CreateIntentInput{Contact: "138"}); err == nil {
		t.Fatal("IntentService.Create: expected error for unknown demand")
	}
	// 需求未发布
	pd, _ := demandSvc.Create(context.Background(), pub, service.CreateDemandInput{PublisherName: "企业", Contact: "138", Title: "待审核"})
	if _, err := intentSvc.Create(context.Background(), worker, pd.ID, service.CreateIntentInput{Contact: "138"}); err == nil {
		t.Fatal("IntentService.Create: expected error for unpublished demand")
	}
	// 自己发布的需求
	d := publishDemandForTest(t, demandSvc, pub, "已发布需求")
	if _, err := intentSvc.Create(context.Background(), pub, d.ID, service.CreateIntentInput{Contact: "138"}); err == nil {
		t.Fatal("IntentService.Create: expected error for self intent")
	}
	// 正常创建（IntentorName 为空 → 默认 a.ID）
	it, err := intentSvc.Create(context.Background(), worker, d.ID, service.CreateIntentInput{Contact: "138"})
	if err != nil {
		t.Fatalf("IntentService.Create: %v", err)
	}
	if it.IntentorName != worker.ID {
		t.Fatalf("IntentService.Create: name=%q, want default %q", it.IntentorName, worker.ID)
	}
	// 重复 pending 意向
	if _, err := intentSvc.Create(context.Background(), worker, d.ID, service.CreateIntentInput{Contact: "138"}); err == nil {
		t.Fatal("IntentService.Create: expected error for duplicate pending intent")
	}
}

func TestIntentService_ListByDemandListMine(t *testing.T) {
	demandRepo := memory.NewDemandRepository(nil)
	intentRepo := memory.NewIntentRepository()
	demandSvc := service.NewDemandService(demandRepo)
	intentSvc := service.NewIntentService(intentRepo, demandRepo)
	pub := domain.Actor{ID: "pub-1", Role: domain.RoleEnterprise}
	worker := domain.Actor{ID: "worker-1", Role: domain.RoleIndividual}

	d := publishDemandForTest(t, demandSvc, pub, "意向列表需求")
	if _, err := intentSvc.Create(context.Background(), worker, d.ID, service.CreateIntentInput{IntentorName: "飞手", Contact: "139"}); err != nil {
		t.Fatalf("IntentService.Create: %v", err)
	}

	// 非发布者/非管理员 → 报错
	if _, err := intentSvc.ListByDemand(context.Background(), worker, d.ID); err == nil {
		t.Fatal("ListByDemand: expected error for non-publisher non-admin")
	}
	// 发布者可见
	if list, err := intentSvc.ListByDemand(context.Background(), pub, d.ID); err != nil || len(list) != 1 {
		t.Fatalf("ListByDemand(publisher): len=%d err=%v", len(list), err)
	}
	// 管理员可见
	if list, err := intentSvc.ListByDemand(context.Background(), domain.Actor{ID: "admin", Role: domain.RolePlatformAdmin}, d.ID); err != nil || len(list) != 1 {
		t.Fatalf("ListByDemand(admin): len=%d err=%v", len(list), err)
	}
	// 需求不存在
	if _, err := intentSvc.ListByDemand(context.Background(), pub, "nope"); err == nil {
		t.Fatal("ListByDemand: expected error for unknown demand")
	}
	// ListMine
	if mine, err := intentSvc.ListMine(context.Background(), worker); err != nil || len(mine) != 1 {
		t.Fatalf("ListMine: len=%d err=%v", len(mine), err)
	}
	if mine, _ := intentSvc.ListMine(context.Background(), pub); len(mine) != 0 {
		t.Fatalf("ListMine(pub): len=%d, want 0", len(mine))
	}
}

func TestMatchingService_ScoreBranches(t *testing.T) {
	repo := memory.NewDemandRepository(nil)
	svc := service.NewMatchingService(repo)

	mk := func(id string, biz domain.BizType, district string, lat, lng float64, budget int64, title string) {
		if _, err := repo.Create(context.Background(), domain.Demand{
			ID: id, Status: domain.DemandPublished, BizType: biz, District: district,
			Latitude: lat, Longitude: lng, BudgetFen: budget, Title: title,
			PublisherName: "测试企业", Description: "线路巡检",
		}); err != nil {
			t.Fatalf("seed demand %s: %v", id, err)
		}
	}
	// 近距（<10km）+ 预算 → 最高分
	mk("d-near", domain.BizCableInspection, "渝北区", 29.5, 106.5, 10000, "近距巡检")
	// 中距（10~30km）
	mk("d-mid", domain.BizCableInspection, "渝北区", 29.7, 106.5, 0, "中距巡检")
	// 远距（30~50km）
	mk("d-far", domain.BizCableInspection, "渝北区", 29.9, 106.5, 0, "远距巡检")
	// 完全不匹配（类型/区域/距离/预算全不中 → score 0，被过滤）
	mk("d-nomatch", domain.BizOther, "南岸区", 0, 0, 0, "不匹配")

	results, err := svc.Recommend(context.Background(), "user-1", 29.5, 106.5, "cable_inspection", "渝北", 10)
	if err != nil {
		t.Fatalf("Recommend: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("Recommend: expected 3 results (score 0 filtered), got %d", len(results))
	}
	if results[0].Demand.ID != "d-near" || !closeScore(results[0].Score, 0.925) {
		t.Fatalf("Recommend[0]: id=%q score=%v, want d-near 0.925", results[0].Demand.ID, results[0].Score)
	}
	if !hasReason(results[0].Reasons, "业务类型匹配") || !hasReason(results[0].Reasons, "区域匹配") || !hasReason(results[0].Reasons, "距离近(<10km)") {
		t.Fatalf("Recommend[0]: reasons=%v, want biz/district/distance reasons", results[0].Reasons)
	}
	if results[1].Demand.ID != "d-mid" || !hasReason(results[1].Reasons, "距离适中(<30km)") {
		t.Fatalf("Recommend[1]: id=%q reasons=%v, want d-mid with 距离适中", results[1].Demand.ID, results[1].Reasons)
	}
	if results[2].Demand.ID != "d-far" {
		t.Fatalf("Recommend[2]: id=%q, want d-far", results[2].Demand.ID)
	}

	// limit 截断
	limited, _ := svc.Recommend(context.Background(), "user-1", 29.5, 106.5, "cable_inspection", "渝北", 2)
	if len(limited) != 2 {
		t.Fatalf("Recommend(limit=2): len=%d, want 2", len(limited))
	}
	// limit 0 → 不截断
	all, _ := svc.Recommend(context.Background(), "user-1", 29.5, 106.5, "cable_inspection", "渝北", 0)
	if len(all) != 3 {
		t.Fatalf("Recommend(limit=0): len=%d, want 3", len(all))
	}
}

func TestMatchingService_SearchAndMatch_Branches(t *testing.T) {
	repo := memory.NewDemandRepository(nil)
	svc := service.NewMatchingService(repo)

	repo.Create(context.Background(), domain.Demand{
		ID: "s1", Status: domain.DemandPublished, BizType: domain.BizCableInspection,
		Title: "电力巡检", PublisherName: "测试企业", Description: "线路巡检作业",
	})
	// 未发布需求应被过滤
	repo.Create(context.Background(), domain.Demand{
		ID: "s2", Status: domain.DemandPending, BizType: domain.BizCableInspection,
		Title: "电力巡检待审", PublisherName: "测试企业", Description: "待审",
	})

	results, err := svc.SearchAndMatch(context.Background(), "巡检", 29.5, 106.5, "cable_inspection", 10)
	if err != nil {
		t.Fatalf("SearchAndMatch: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("SearchAndMatch: expected 1 published result, got %d", len(results))
	}
	if results[0].Demand.ID != "s1" {
		t.Fatalf("SearchAndMatch[0]: id=%q, want s1", results[0].Demand.ID)
	}
	// 0.35(biz) + 0.3(boost) = 0.65
	if !closeScore(results[0].Score, 0.65) {
		t.Fatalf("SearchAndMatch[0]: score=%v, want 0.65", results[0].Score)
	}

	// 无命中关键词
	none, err := svc.SearchAndMatch(context.Background(), "不存在", 0, 0, "", 10)
	if err != nil || len(none) != 0 {
		t.Fatalf("SearchAndMatch(no-hit): len=%d err=%v", len(none), err)
	}
}
