package service_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

func entActor() domain.Actor { return domain.Actor{ID: "ent-1", Role: domain.RoleEnterprise} }
func indActor() domain.Actor { return domain.Actor{ID: "user-1", Role: domain.RoleIndividual} }
func admActor() domain.Actor { return domain.Actor{ID: "admin-1", Role: domain.RolePlatformAdmin} }

// === Demand State Machine ===

func TestDemandFullLifecycle(t *testing.T) {
	svc := service.NewDemandService(memory.NewDemandRepository(nil))

	// Create → pending (awaiting admin review)
	d, err := svc.Create(context.Background(), entActor(), service.CreateDemandInput{
		PublisherName: "测试企业", Contact: "13800001111", Title: "巡检需求", BizType: "cable_inspection",
	})
	if err != nil {
		t.Fatal(err)
	}
	if d.Status != domain.DemandPending {
		t.Fatalf("expected pending, got %s", d.Status)
	}

	// Admin approve → published
	d, err = svc.Approve(context.Background(), admActor(), d.ID)
	if err != nil {
		t.Fatal(err)
	}
	if d.Status != domain.DemandPublished {
		t.Fatalf("expected published, got %s", d.Status)
	}

	// Publisher complete → completed
	d, err = svc.Complete(context.Background(), entActor(), d.ID)
	if err != nil {
		t.Fatal(err)
	}
	if d.Status != domain.DemandCompleted {
		t.Fatalf("expected completed, got %s", d.Status)
	}
}

func TestDemandUpdateDraft(t *testing.T) {
	svc := service.NewDemandService(memory.NewDemandRepository(nil))
	a := entActor()
	d, _ := svc.Create(context.Background(), a, service.CreateDemandInput{PublisherName: "P", Contact: "138", Title: "T"})
	// Update draft
	d2, err := svc.UpdateDraft(context.Background(), a, d.ID, "新标题", "新描述")
	if err != nil {
		t.Fatal(err)
	}
	if d2.Title != "新标题" {
		t.Fatal("title not updated")
	}
	// Non-owner cannot update
	if _, err := svc.UpdateDraft(context.Background(), indActor(), d.ID, "x", "x"); err == nil {
		t.Fatal("non-owner should not update")
	}
}

// === Contract State Machine ===

func TestContractStateMachine(t *testing.T) {
	svc := service.NewContractService(memory.NewContractRepository())
	a := entActor()
	c, err := svc.Create(context.Background(), a, domain.Contract{EnterpriseID: a.ID, TemplateID: "tpl-1"})
	if err != nil {
		t.Fatal(err)
	}
	if c.Status != domain.ContractDraft {
		t.Fatalf("expected draft, got %s", c.Status)
	}
	// Invalid transition: draft -> signed (should fail)
	if _, err := svc.UpdateStatus(context.Background(), a, c.ID, domain.ContractSigned); err == nil {
		t.Fatal("draft -> signed should fail")
	}
	// Valid: draft -> sent
	c2, err := svc.UpdateStatus(context.Background(), a, c.ID, domain.ContractSent)
	if err != nil {
		t.Fatal(err)
	}
	if c2.Status != domain.ContractSent {
		t.Fatal("status not sent")
	}
	// Valid: sent -> signing
	svc.UpdateStatus(context.Background(), a, c.ID, domain.ContractSigning)
	// Valid: signing -> signed
	svc.UpdateStatus(context.Background(), a, c.ID, domain.ContractSigned)
	// Valid: signed -> voided
	c5, err := svc.UpdateStatus(context.Background(), a, c.ID, domain.ContractVoided)
	if err != nil {
		t.Fatal(err)
	}
	if c5.Status != domain.ContractVoided {
		t.Fatal("status not voided")
	}
}

func TestContractOwnershipCheck(t *testing.T) {
	svc := service.NewContractService(memory.NewContractRepository())
	a := entActor()
	c, _ := svc.Create(context.Background(), a, domain.Contract{EnterpriseID: a.ID, TemplateID: "tpl-1"})
	// Another enterprise cannot update
	other := domain.Actor{ID: "ent-2", Role: domain.RoleEnterprise}
	if _, err := svc.UpdateStatus(context.Background(), other, c.ID, domain.ContractSent); err == nil {
		t.Fatal("other enterprise should not update contract")
	}
}

// === Enterprise State Machine ===

func TestEnterpriseReviewFlow(t *testing.T) {
	users := memory.NewUserRepository(nil)
	// 审核通过需要将属主升级为企业角色（升级失败会回滚审批），测试需预置属主用户。
	_, _ = users.Create(context.Background(), domain.User{ID: entActor().ID, Role: domain.RoleEnterprise, Status: "active"})
	svc := service.NewEnterpriseSvc(memory.NewEnterpriseRepository(nil), users)
	a := entActor()
	e, err := svc.Create(context.Background(), a, service.CreateEnterpriseInput{Name: "测试企业", AccountName: "6222"})
	if err != nil {
		t.Fatal(err)
	}
	if e.Status != domain.EnterpriseDraft {
		t.Fatal("should be draft")
	}
	// Submit
	e2, err := svc.Submit(context.Background(), a, e.ID)
	if err != nil {
		t.Fatal(err)
	}
	if e2.Status != domain.EnterpriseSubmitted {
		t.Fatal("should be submitted")
	}
	// Admin review: approve
	e3, err := svc.Review(context.Background(), admActor(), e.ID, "approve", "")
	if err != nil {
		t.Fatal(err)
	}
	if e3.Status != domain.EnterpriseApproved {
		t.Fatal("should be approved")
	}
}

func TestEnterpriseReviewReject(t *testing.T) {
	svc := service.NewEnterpriseSvc(memory.NewEnterpriseRepository(nil), memory.NewUserRepository(nil))
	a := entActor()
	e, _ := svc.Create(context.Background(), a, service.CreateEnterpriseInput{Name: "测试企业"})
	svc.Submit(context.Background(), a, e.ID)
	// Admin reject
	e2, err := svc.Review(context.Background(), admActor(), e.ID, "reject", "资料不全")
	if err != nil {
		t.Fatal(err)
	}
	if e2.Status != domain.EnterpriseRejected {
		t.Fatal("should be rejected")
	}
}

func TestEnterpriseNonAdminCannotReview(t *testing.T) {
	svc := service.NewEnterpriseSvc(memory.NewEnterpriseRepository(nil), memory.NewUserRepository(nil))
	a := entActor()
	e, _ := svc.Create(context.Background(), a, service.CreateEnterpriseInput{Name: "测试企业"})
	svc.Submit(context.Background(), a, e.ID)
	if _, err := svc.Review(context.Background(), indActor(), e.ID, "approve", ""); err == nil {
		t.Fatal("individual should not review")
	}
}

// === Employment ===
func TestEmploymentCreateAndList(t *testing.T) {
	svc := service.NewEmploymentService(memory.NewEmploymentRepository())
	a := entActor()
	er, err := svc.Create(context.Background(), a, domain.EmploymentRequest{EnterpriseID: a.ID, Position: "飞手", Headcount: 5})
	if err != nil {
		t.Fatal(err)
	}
	_ = er
	// Enterprise lists own
	list, total, err := svc.List(context.Background(), a, 0, 20)
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 {
		t.Fatalf("expected 1, got %d", total)
	}
	_ = list
	// Individual cannot list
	if _, _, err := svc.List(context.Background(), indActor(), 0, 20); err == nil {
		t.Fatal("individual should not list employment")
	}
}

// === Escrow ===
func TestEscrowDepositFreezeReleaseRefund(t *testing.T) {
	svc := service.NewEscrowService(memory.NewEscrowRepository())
	// Deposit
	tx, err := svc.Deposit(context.Background(), "user-1", 100000)
	if err != nil {
		t.Fatal(err)
	}
	if tx.TxType != "deposit" {
		t.Fatal("wrong tx type")
	}
	// Balance
	acct, _ := svc.Balance(context.Background(), "user-1")
	if acct.BalanceFen != 100000 {
		t.Fatalf("balance: %d", acct.BalanceFen)
	}
	// Freeze
	svc.Freeze(context.Background(), "user-1", 30000, "course", "course-1")
	acct2, _ := svc.Balance(context.Background(), "user-1")
	if acct2.FrozenFen != 30000 {
		t.Fatalf("frozen: %d", acct2.FrozenFen)
	}
	if acct2.BalanceFen != 70000 {
		t.Fatalf("balance after freeze: %d", acct2.BalanceFen)
	}
	// Release
	svc.Release(context.Background(), "user-1", "org-1", 20000, "course", "course-1")
	acct3, _ := svc.Balance(context.Background(), "user-1")
	if acct3.FrozenFen != 10000 {
		t.Fatalf("frozen after release: %d", acct3.FrozenFen)
	}
	// Refund
	svc.Refund(context.Background(), "user-1", 10000, "course", "course-1")
	acct4, _ := svc.Balance(context.Background(), "user-1")
	if acct4.FrozenFen != 0 {
		t.Fatalf("frozen after refund: %d", acct4.FrozenFen)
	}
	if acct4.BalanceFen != 80000 {
		t.Fatalf("balance: %d", acct4.BalanceFen)
	}
	// Insufficient balance
	if _, err := svc.Freeze(context.Background(), "user-1", 200000, "x", "x"); err == nil {
		t.Fatal("should fail on insufficient balance")
	}
	// Transactions
	txs, _ := svc.Transactions(context.Background(), "user-1")
	if len(txs) != 4 {
		t.Fatalf("expected 4 txs, got %d", len(txs))
	}
}

// === Message ===
func TestMessageSendMarkRead(t *testing.T) {
	svc := service.NewMessageService(memory.NewMessageRepository())
	m, _ := svc.Send(context.Background(), "sys", "user-1", "通知", "内容", "demand", "d-1")
	if m.IsRead {
		t.Fatal("should be unread")
	}
	// C10 回归：非收件人标已读必须被拒绝
	if _, err := svc.MarkRead(context.Background(), "stranger", m.ID); err == nil {
		t.Fatal("stranger should not mark read")
	}
	svc.MarkRead(context.Background(), "user-1", m.ID)
	// Unread count
	count, _ := svc.UnreadCount(context.Background(), "user-1")
	if count != 0 {
		t.Fatalf("unread: %d", count)
	}
	// List
	msgs, _ := svc.ListForUser(context.Background(), "user-1", false)
	if len(msgs) != 1 {
		t.Fatal("list all")
	}
	msgs2, _ := svc.ListForUser(context.Background(), "user-1", true)
	if len(msgs2) != 0 {
		t.Fatal("list unread should be empty")
	}
}

// === News ===
func TestNewsCreatePublish(t *testing.T) {
	svc := service.NewNewsService(memory.NewArticleRepository())
	a, _ := svc.Create(context.Background(), "新闻", "内容...", "policy", "来源", "", false)
	if a.Status != "draft" {
		t.Fatal("should be draft")
	}
	a2, _ := svc.Publish(context.Background(), a.ID)
	if a2.Status != "published" {
		t.Fatal("should be published")
	}
	_, total, _ := svc.ListByCategory(context.Background(), "policy", 1, 20)
	if total != 1 {
		t.Fatal("list by category")
	}
}

func TestNewsUpdate(t *testing.T) {
	svc := service.NewNewsService(memory.NewArticleRepository())
	a, _ := svc.Create(context.Background(), "原标题", "原内容", "policy", "来源", "", false)
	// 编辑：保留 ID 与状态，内容更新，摘要重算
	u, err := svc.Update(context.Background(), a.ID, "新标题", strings.Repeat("新内容", 60), "uav_regulation", "新来源", "", false)
	if err != nil {
		t.Fatal("update should succeed")
	}
	if u.ID != a.ID || u.Status != "draft" || u.Title != "新标题" || u.Category != "uav_regulation" || u.Source != "新来源" {
		t.Fatal("update should keep id/status and change fields")
	}
	if n := len([]rune(u.Summary)); n > 103 { // 100 字 + "..."
		t.Fatal("summary should be truncated to 100 chars")
	}
	// 编辑不存在的文章应报错
	if _, err := svc.Update(context.Background(), "article-missing", "x", "y", "", "", "", false); err == nil {
		t.Fatal("update missing article should error")
	}
}

// === Training ===
func TestTrainingCertCourseFlow(t *testing.T) {
	svc := service.NewTrainingService(
		memory.NewCertificateRepository(), memory.NewCourseRepository(),
		memory.NewInstructorRepository(), memory.NewPilotRepository(nil),
	)
	// Certificate
	cert, _ := svc.AddCertificate(context.Background(), indActor(), domain.CertCAAC, "CAAC-001", "III", "民航局", "", time.Now(), time.Now().AddDate(2, 0, 0))
	if cert.Status != "pending" {
		t.Fatal("should be pending")
	}
	cert2, _ := svc.ApproveCertificate(context.Background(), admActor(), cert.ID)
	if cert2.Status != "approved" {
		t.Fatal("should be approved")
	}
	// Non-admin cannot approve
	if _, err := svc.ApproveCertificate(context.Background(), indActor(), cert.ID); err == nil {
		t.Fatal("individual should not approve cert")
	}
	// Course
	course, _ := svc.CreateCourse(context.Background(), entActor(), domain.TrainingCourse{
		Title: "课程", CertType: domain.CertCAAC, Description: "描述", Location: "重庆",
		StartDate: time.Now(), EndDate: time.Now().AddDate(0, 1, 0), MaxStudents: 30, PriceFen: 50000,
	})
	if course.Status != "draft" {
		t.Fatal("course should be draft")
	}
	// Instructor
	inst, _ := svc.RegisterInstructor(context.Background(), indActor(), "教练", "", "bio", "org-1", []string{"CAAC"})
	inst2, _ := svc.ApproveInstructor(context.Background(), admActor(), inst.ID)
	_ = inst2
	// Pilot
	pilot, _ := svc.RegisterPilot(context.Background(), indActor(), "飞行员", "500101199001011234", 120, "电力巡检", "avatar.jpg", "重庆")
	pilot2, _ := svc.ApprovePilot(context.Background(), admActor(), pilot.ID)
	_ = pilot2
	// Lists
	certs, _ := svc.ListMyCertificates(context.Background(), indActor())
	if len(certs) != 1 {
		t.Fatal("list my certs")
	}
	courses, _ := svc.ListCourses(context.Background())
	if len(courses) != 1 {
		t.Fatal("list courses")
	}
	insts, _ := svc.ListInstructors(context.Background())
	if len(insts) != 1 {
		t.Fatal("list instructors")
	}
	pilots, _ := svc.ListPilots(context.Background())
	if len(pilots) != 1 {
		t.Fatal("list pilots")
	}
}

// === Pilot 详情字段：avatar/region 落库 + 证书关联 + 单查 ===
func TestPilotDetailFields(t *testing.T) {
	svc := service.NewTrainingService(
		memory.NewCertificateRepository(), memory.NewCourseRepository(),
		memory.NewInstructorRepository(), memory.NewPilotRepository(nil),
	)
	// 证书：一张 approved，一张 pending（后者不应出现在名录关联里）
	caac, _ := svc.AddCertificate(context.Background(), indActor(), domain.CertCAAC, "CAAC-001", "III", "民航局", "", time.Now(), time.Now().AddDate(2, 0, 0))
	svc.ApproveCertificate(context.Background(), admActor(), caac.ID)
	svc.AddCertificate(context.Background(), indActor(), domain.CertUTCDJI, "UTC-001", "", "大疆", "", time.Now(), time.Now().AddDate(2, 0, 0))

	// 注册带 avatar/region
	pilot, err := svc.RegisterPilot(context.Background(), indActor(), "王飞", "500101199001011234", 120, "电力巡检", "/uploads/a.jpg", "重庆渝北")
	if err != nil {
		t.Fatal(err)
	}
	if pilot.Avatar != "/uploads/a.jpg" || pilot.Region != "重庆渝北" {
		t.Fatalf("avatar/region 未落库: %+v", pilot)
	}
	svc.ApprovePilot(context.Background(), admActor(), pilot.ID)

	// 单查
	got, err := svc.GetPilot(context.Background(), pilot.ID)
	if err != nil || got.ID == "" {
		t.Fatalf("GetPilot failed: %v", err)
	}
	if got.Avatar != "/uploads/a.jpg" || got.Region != "重庆渝北" {
		t.Fatalf("GetPilot 新字段缺失: %+v", got)
	}

	// 名录详情：cert_ids 扩展为 certificates（仅 approved）
	details, err := svc.ListPilotsDetailed(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(details) != 1 {
		t.Fatalf("expected 1 pilot, got %d", len(details))
	}
	if len(details[0].Certificates) != 1 {
		t.Fatalf("expected 1 approved certificate, got %d", len(details[0].Certificates))
	}
	c := details[0].Certificates[0]
	if c.CertType != "caac" || c.CertName != "CAAC无人机驾驶员执照" {
		t.Fatalf("cert brief wrong: %+v", c)
	}
	// 管理员视角不含敏感字段时不可见 IDCard 明文
	if details[0].IDCard == "" {
		t.Fatal("名录应带身份证号（handler 层才脱敏）")
	}
}
