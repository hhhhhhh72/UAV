package service_test

import (
	"testing"
	"time"

	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

// === Expert full CRUD ===
func TestExpertFullCRUD(t *testing.T) {
	svc := service.NewExpertService(memory.NewExpertRepository())
	e, _ := svc.Create("张三", "教授", "重大", "无人机", "bio", []string{"CAAC"})
	// Get
	got, err := svc.Get(e.ID)
	if err != nil || got.Name != "张三" { t.Fatal("Get failed") }
	// Update
	upd, err := svc.Update(e.ID, "李四", "副教授", "北大", "低空", "bio2", []string{"UTC"})
	if err != nil || upd.Name != "李四" { t.Fatal("Update failed") }
	// List by field
	list, _ := svc.List("低空")
	if len(list) != 1 { t.Fatal("List field filter failed") }
	// Delete
	svc.Delete(e.ID)
	if _, err := svc.Get(e.ID); err == nil { t.Fatal("should be deleted") }
}

// === Case full CRUD ===
func TestCaseFullCRUD(t *testing.T) {
	svc := service.NewCaseService(memory.NewCaseRepository())
	c, _ := svc.Create("案例1", "logistics", "desc", []string{"a.jpg"}, "客户A", "成果")
	got, _ := svc.Get(c.ID)
	if got.ClientName != "客户A" { t.Fatal("Get failed") }
	_, err := svc.Update(c.ID, "案例1v2", "agriculture", "desc2", []string{"b.jpg"}, "客户B", "成果2")
	if err != nil { t.Fatal(err) }
	// List all
	_, total, _ := svc.List("", 1, 20)
	if total != 1 { t.Fatal("list all") }
	// List by category
	_, total2, _ := svc.List("agriculture", 1, 20)
	if total2 != 1 { t.Fatal("list by cat") }
	svc.Delete(c.ID)
}

// === Compliance full CRUD ===
func TestComplianceFullCRUD(t *testing.T) {
	svc := service.NewComplianceService(memory.NewComplianceRepository())
	d, _ := svc.CreateDoc("条例", "policy", "内容", "摘要", "民航局", []string{"适航"})
	_, err := svc.UpdateDoc(d.ID, "条例v2", "regulation", "内容2", "摘要2", "民航局v2", []string{"通用"})
	if err != nil { t.Fatal(err) }
	svc.DeleteDoc(d.ID)
	// Standard
	s, _ := svc.CreateStandard("标准1", "T/CDA-001", "巡检", "1.0", "协会", "标准内容", "file.pdf", time.Now())
	_, total, _ := svc.ListStandards("", 1, 20)
	if total != 1 { t.Fatal("list standards") }
	_ = s
}

// === Report full CRUD ===
func TestReportFullCRUD(t *testing.T) {
	svc := service.NewReportService(memory.NewIndustryReportRepository())
	r, _ := svc.Create("报告", "2026H1", "行业", "摘要", "全文", "f.pdf", "作者")
	got, _ := svc.Get(r.ID)
	if got.Author != "作者" { t.Fatal("Get failed") }
	svc.Update(r.ID, "报告v2", "2026H2", "政策", "摘要2", "全文2", "f2.pdf", "作者2")
	svc.Delete(r.ID)
	_, total, _ := svc.List(1, 20)
	if total != 0 { t.Fatal("should be empty") }
}

// === Portfolio full CRUD ===
func TestPortfolioFullCRUD(t *testing.T) {
	svc := service.NewPortfolioService(memory.NewPortfolioRepository())
	p, _ := svc.Create("ent-1", "品牌A", "logo.png", "cover.png", "无人机方案", "138", []string{"巡检"}, []string{"优秀"})
	got, _ := svc.Get(p.ID)
	if got.Name != "品牌A" { t.Fatal("Get failed") }
	svc.Update(p.ID, "品牌A2", "logo2.png", "cover2.png", "方案2", "139", []string{"物流"}, []string{"十佳"})
	// ListByEnterprise
	mine, _ := svc.ListByEnterprise("ent-1")
	if len(mine) != 1 { t.Fatal("list by enterprise") }
	// ListPublished (draft status = not published)
	pub, total, _ := svc.ListPublished(1, 20)
	_ = pub
	_ = total
}

// === Achievement full CRUD ===
func TestAchievementFullCRUD(t *testing.T) {
	svc := service.NewAchievementService(memory.NewAchievementRepository())
	a, _ := svc.Create("user-1", "AI算法", "patent", "自动避障", "无人机", "lab", "138", []string{"d.jpg"})
	got, _ := svc.Get(a.ID)
	if got.AchieveType != "patent" { t.Fatal("Get failed") }
	svc.Update(a.ID, "AI算法v2", "software", "避障v2", "低空", "pilot", "139", []string{"d2.jpg"})
	// List by field
	_, total, _ := svc.List("低空", 1, 20)
	if total != 1 { t.Fatal("list by field") }
	svc.Delete(a.ID)
}

// === RDChallenge full CRUD ===
func TestRDChallengeFullCRUD(t *testing.T) {
	svc := service.NewRDChallengeService(memory.NewRDChallengeRepository())
	c, _ := svc.Create("ent-1", "电池技术", "电池", ">2h续航", 500000, time.Now().AddDate(0, 3, 0))
	got, _ := svc.Get(c.ID)
	if got.BudgetFen != 500000 { t.Fatal("Get failed") }
	svc.Update(c.ID, "电池v2", "电池", ">3h", 800000, time.Now().AddDate(0, 6, 0))
}

// === ResearchProject full CRUD ===
func TestResearchProjectFullCRUD(t *testing.T) {
	svc := service.NewResearchProjectService(memory.NewResearchProjectRepository())
	p, _ := svc.Create("航路规划", "空域", "城市航路", "重大", "Q1调研", []string{"巡航科技"}, 200000, time.Now(), time.Now().AddDate(1, 0, 0))
	got, _ := svc.Get(p.ID)
	if got.LeadOrg != "重大" { t.Fatal("Get failed") }
	svc.Update(p.ID, "航路规划v2", "空域", "v2", "重大", "Q2", []string{"新增企业"}, 300000, time.Now(), time.Now().AddDate(1, 0, 0))
}

// === ProjectApp full CRUD ===
func TestProjectAppFullCRUD(t *testing.T) {
	svc := service.NewProjectAppService(memory.NewProjectAppRepository())
	a, _ := svc.Create("user-1", "示范项目", "示范", "无人机产业", 1000000, []string{"计划书.pdf"})
	got, _ := svc.Get(a.ID)
	
	// List my
	mine, _ := svc.ListMy("user-1")
	if len(mine) != 1 { t.Fatal("list my") }
	// List all (admin)
	_, total, _ := svc.ListAll("", 1, 20)
	if total != 1 { t.Fatal("list all") }
	// Review approve
	reviewed, _ := svc.Review(a.ID, "同意立项", "approve")
	if reviewed.Status != "approved" { t.Fatal("should be approved") }
	// List all filtered by status
	_, total2, _ := svc.ListAll("approved", 1, 20)
	if total2 != 1 { t.Fatal("filter by status") }
	_ = got
}

// === Competition full CRUD ===
func TestCompetitionFullCRUD(t *testing.T) {
	svc := service.NewCompetitionService(memory.NewCompetitionRepository())
	c, _ := svc.Create("竞速赛", "racing", "FPV", "巴南", "协会", time.Now().AddDate(0, 1, 0), time.Now().AddDate(0, 1, 3), 50)
	got, _ := svc.Get(c.ID)
	if got.Sponsor != "协会" { t.Fatal("Get failed") }
	svc.Update(c.ID, "竞速赛v2", "fpv", "FPVv2", "渝北", "协会2", time.Now(), time.Now().AddDate(0, 2, 0), 100)
	// Register + ListRegs
	svc.Register(c.ID, "user-1", "闪电队", 3, "138")
	regs, _ := svc.ListRegs(c.ID)
	if len(regs) != 1 { t.Fatal("list regs") }
}

// === Event full CRUD ===
func TestEventFullCRUD(t *testing.T) {
	svc := service.NewEventService(memory.NewEventRepository())
	e, _ := svc.Create("论坛", "forum", "年度", "博览中心", "cover.jpg", time.Now().AddDate(0, 2, 0), time.Now().AddDate(0, 2, 1), 500)
	got, _ := svc.Get(e.ID)
	if got.EventType != "forum" { t.Fatal("Get failed") }
	svc.Update(e.ID, "论坛v2", "exhibition", "年度v2", "国际中心", "cover2.jpg", time.Now(), time.Now().AddDate(0, 3, 0), 1000)
	// Register + ListRegs
	svc.Register(e.ID, "user-1", "张三", "138", "巡航科技")
	regs, _ := svc.ListRegs(e.ID)
	if len(regs) != 1 { t.Fatal("list regs") }
}

// === Resource full CRUD ===
func TestResourceFullCRUD(t *testing.T) {
	svc := service.NewResourceService(memory.NewResourceRepository())
	r, _ := svc.Create("user-1", "无人机01", "drone", "M300", "RTK+热成像", "南岸", "9-18点", 100000)
	got, _ := svc.Get(r.ID)
	if got.Model != "M300" { t.Fatal("Get failed") }
	svc.Update(r.ID, "无人机02", "drone", "M350", "RTK+LiDAR", "江北", "全天", 150000)
	// List by type
	_, total, _ := svc.List("drone", 1, 20)
	if total != 1 { t.Fatal("list drone") }
	// List all
	_, total2, _ := svc.List("", 1, 20)
	if total2 != 1 { t.Fatal("list all") }
}

// === Emergency full CRUD ===
func TestEmergencyFullCRUD(t *testing.T) {
	svc := service.NewEmergencyService(memory.NewEmergencyRepository())
	r, _ := svc.CreateResource("user-1", "应急机01", "drone", "M300RTK", "南岸", "138", 2)
	got, _ := svc.GetResource(r.ID)
	if got.Quantity != 2 { t.Fatal("Get failed") }
	// Dispatch
	d, _ := svc.CreateDispatch(r.ID, "山火", "北碚", "张指挥", "成功", time.Now(), time.Now().Add(3*time.Hour))
	_ = d
	// List resources
	_, total, _ := svc.ListResources(1, 20)
	if total != 1 { t.Fatal("list resources") }
	// List dispatches
	_, total2, _ := svc.ListDispatches(1, 20)
	if total2 != 1 { t.Fatal("list dispatches") }
	_ = got
}
