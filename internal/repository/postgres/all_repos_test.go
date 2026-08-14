package postgres_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"drone-platform/internal/crypto"
	"drone-platform/internal/domain"
	"drone-platform/internal/repository/postgres"
)

// pgStore is a shared store for all tests (reduces connection overhead).
var pgStore *postgres.Store

func init() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	url := databaseURL()
	s, err := postgres.NewStore(ctx, url, nil)
	if err != nil {
		return
	}
	if err := s.RunMigrationsFromDir(ctx, postgres.MigrationsDir()); err != nil {
		s.Close()
		return
	}
	pgStore = s
}

func uid(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}

// === Expert PG ===
func TestPG_ExpertRepo(t *testing.T) {
	if pgStore == nil {
		t.Skip("PG not available")
	}
	r := pgStore.NewExpertRepository()
	id := uid("exp")
	e, err := r.Create(domain.Expert{ID: id, Name: "PG专家", Title: "教授", Field: "无人机", Status: "active"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.FindByID(e.ID); err != nil {
		t.Fatal(err)
	}
	if list, _ := r.List("无人机"); len(list) == 0 {
		t.Fatal("list empty")
	}
	r.Update(domain.Expert{ID: id, Name: "updated", Title: "副教授", Field: "低空", Status: "active"})
	r.Delete(id)
}

// === Case PG ===
func TestPG_CaseRepo(t *testing.T) {
	if pgStore == nil {
		t.Skip("PG not available")
	}
	r := pgStore.NewCaseRepository()
	id := uid("case")
	r.Create(domain.CaseEntry{ID: id, Title: "案例", Category: "logistics", Status: "draft"})
	r.FindByID(id)
	r.List("", 0, 20)
	r.Delete(id)
}

// containsID reports whether items contains a StandardDoc with the given id.
func containsID(items []domain.StandardDoc, id string) bool {
	for _, it := range items {
		if it.ID == id {
			return true
		}
	}
	return false
}

// containsBookingID reports whether bookings contains one with the given id.
func containsBookingID(bookings []domain.IndustryResourceBooking, id string) bool {
	for _, b := range bookings {
		if b.ID == id {
			return true
		}
	}
	return false
}

// === Compliance PG ===
func TestPG_ComplianceRepo(t *testing.T) {
	if pgStore == nil {
		t.Skip("PG not available")
	}
	r := pgStore.NewComplianceRepository()
	id := uid("cdoc")
	r.CreateDoc(domain.ComplianceDoc{ID: id, Title: "条例", Category: "policy", Status: "draft"})
	r.FindDocByID(id)
	r.ListDocs("", 0, 20)
	r.UpdateDoc(domain.ComplianceDoc{ID: id, Title: "updated", Category: "regulation", Status: "published"})
	r.DeleteDoc(id)
	stdID := uid("std")
	r.CreateStandard(domain.StandardDoc{ID: stdID, Title: "标准", Category: "团体标准", StandardNo: "T/CDA-001", Status: "draft"})
	r.ListStandards("", 0, 20)
	// C11: 分类列 + 筛选（旧表无 category 列，WHERE category=$1 会报错）
	// 本地测试库有历史残留行，断言只针对本次创建的记录。
	items, _, err := r.ListStandards("团体标准", 0, 20)
	if err != nil {
		t.Fatalf("ListStandards(团体标准): %v", err)
	}
	found := false
	for _, it := range items {
		if it.ID == stdID && it.Category == "团体标准" {
			found = true
		}
	}
	if !found {
		t.Fatalf("ListStandards(团体标准): %d items, stdID %s not found with category 团体标准", len(items), stdID)
	}
	if others, _, _ := r.ListStandards("国家标准", 0, 20); containsID(others, stdID) {
		t.Fatalf("ListStandards(国家标准) unexpectedly contains stdID %s", stdID)
	}
	r.UpdateStandard(domain.StandardDoc{ID: stdID, Title: "标准", Category: "企业标准", StandardNo: "T/CDA-001", Status: "published"})
	if got, _ := r.FindStandardByID(stdID); got.Category != "企业标准" {
		t.Fatalf("updated category %q, want 企业标准", got.Category)
	}
	r.DeleteStandard(stdID)
}

// === Report PG ===
func TestPG_ReportRepo(t *testing.T) {
	if pgStore == nil {
		t.Skip("PG not available")
	}
	r := pgStore.NewIndustryReportRepository()
	id := uid("ir")
	r.Create(domain.IndustryReport{ID: id, Title: "报告", Period: "2026H1", Status: "draft"})
	r.FindByID(id)
	r.List(0, 20)
	r.Update(domain.IndustryReport{ID: id, Title: "updated", Status: "published"})
	r.Delete(id)
}

// === Portfolio PG ===
func TestPG_PortfolioRepo(t *testing.T) {
	if pgStore == nil {
		t.Skip("PG not available")
	}
	r := pgStore.NewPortfolioRepository()
	id := uid("port")
	r.Create(domain.MemberPortfolio{ID: id, EnterpriseID: "ent-1", Name: "品牌", Status: "draft"})
	r.FindByID(id)
	r.ListByEnterprise("ent-1")
	r.ListPublished(0, 20)
	r.Update(domain.MemberPortfolio{ID: id, Name: "updated", Status: "published"})
}

// === Achievement PG ===
func TestPG_AchievementRepo(t *testing.T) {
	if pgStore == nil {
		t.Skip("PG not available")
	}
	r := pgStore.NewAchievementRepository()
	id := uid("ach")
	r.Create(domain.Achievement{ID: id, OwnerID: "u-1", Title: "专利", AchieveType: "patent", Status: "pending"})
	r.FindByID(id)
	r.List("", 0, 20)
	r.Update(domain.Achievement{ID: id, Title: "updated", Status: "published"})
	r.Delete(id)
}

// === RDChallenge PG ===
func TestPG_RDChallengeRepo(t *testing.T) {
	if pgStore == nil {
		t.Skip("PG not available")
	}
	r := pgStore.NewRDChallengeRepository()
	id := uid("rd")
	r.Create(domain.RDChallenge{ID: id, PosterID: "u-1", Title: "难题", Status: "open"})
	r.FindByID(id)
	r.List("", 0, 20)
	r.Update(domain.RDChallenge{ID: id, Title: "updated", Status: "in_progress"})
}

// === ResearchProject PG ===
func TestPG_ResearchProjRepo(t *testing.T) {
	if pgStore == nil {
		t.Skip("PG not available")
	}
	r := pgStore.NewResearchProjectRepository()
	id := uid("rp")
	r.Create(domain.ResearchProject{ID: id, Title: "课题", Status: "planning"})
	r.FindByID(id)
	r.List(0, 20)
	r.Update(domain.ResearchProject{ID: id, Title: "updated", Status: "active"})
}

// === ProjectApp PG ===
func TestPG_ProjectAppRepo(t *testing.T) {
	if pgStore == nil {
		t.Skip("PG not available")
	}
	r := pgStore.NewProjectAppRepository()
	id := uid("pa")
	r.Create(domain.ProjectApplication{ID: id, ApplicantID: "u-1", ProjectName: "示范项目", Status: "draft"})
	r.FindByID(id)
	r.ListByUser("u-1")
	r.ListAll("", 0, 20)
	r.Update(domain.ProjectApplication{ID: id, Status: "submitted", ReviewNote: "ok"})
}

// === Competition PG ===
func TestPG_CompetitionRepo(t *testing.T) {
	if pgStore == nil {
		t.Skip("PG not available")
	}
	r := pgStore.NewCompetitionRepository(nil)
	id := uid("comp")
	r.Create(domain.Competition{ID: id, Title: "比赛", Status: "draft"})
	r.FindByID(id)
	r.List(0, 20)
	r.Update(domain.Competition{ID: id, Title: "updated", Status: "open"})
	// C13: 扩展字段（name/phone/id_card/photo_url/id_card_image）落库
	cregID := uid("creg")
	r.CreateReg(domain.CompetitionReg{ID: cregID, CompetitionID: id, UserID: "u-1",
		Name: "张三", Phone: "13800000000", IDCard: "500101199001011234",
		PhotoURL: "/uploads/p.jpg", IDCardImage: "/uploads/i.jpg", Status: "submitted"})
	regs, err := r.ListRegs(id)
	if err != nil {
		t.Fatalf("list regs: %v", err)
	}
	// 历史测试残留会累积报名行，断言只针对本次创建的记录
	found := false
	for _, rg := range regs {
		if rg.ID == cregID {
			found = true
			if rg.Name != "张三" || rg.Phone != "13800000000" || rg.IDCard != "500101199001011234" ||
				rg.PhotoURL != "/uploads/p.jpg" || rg.IDCardImage != "/uploads/i.jpg" {
				t.Fatalf("reg=%+v, want extended fields persisted", rg)
			}
		}
	}
	if !found {
		t.Fatalf("reg %s not found in %d rows", cregID, len(regs))
	}
}

// === CompetitionReg PII encryption PG ===
// 审查 HIGH 修复：id_card/phone 静态加密——库内为密文，读取还原明文。
func TestPG_CompetitionRegPIIEncryption(t *testing.T) {
	if pgStore == nil {
		t.Skip("PG not available")
	}
	cipher, err := crypto.NewCipher("MTIzNDU2Nzg5MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTI=")
	if err != nil {
		t.Fatal(err)
	}
	r := pgStore.NewCompetitionRepository(cipher)
	compID := uid("comp-enc")
	r.Create(domain.Competition{ID: compID, Title: "加密测试赛", Status: "draft"})
	cregID := uid("creg-enc")
	if _, err := r.CreateReg(domain.CompetitionReg{ID: cregID, CompetitionID: compID, UserID: "u-1",
		Name: "张三", Phone: "13800000000", IDCard: "500101199001011234", Status: "submitted"}); err != nil {
		t.Fatalf("create reg: %v", err)
	}
	// 库内为密文
	var stored string
	if err := pgStore.Pool().QueryRow(context.Background(),
		`SELECT id_card FROM competition_registrations WHERE id=$1`, cregID).Scan(&stored); err != nil {
		t.Fatalf("read raw id_card: %v", err)
	}
	if stored == "500101199001011234" {
		t.Fatalf("id_card stored in plaintext")
	}
	if dec, err := cipher.Decrypt(stored); err != nil || dec != "500101199001011234" {
		t.Fatalf("decrypt stored id_card: %q err=%v", dec, err)
	}
	// 读取还原明文
	regs, err := r.ListRegs(compID)
	if err != nil {
		t.Fatalf("list regs: %v", err)
	}
	found := false
	for _, rg := range regs {
		if rg.ID == cregID {
			found = true
			if rg.IDCard != "500101199001011234" || rg.Phone != "13800000000" {
				t.Fatalf("reg=%+v, want decrypted PII", rg)
			}
		}
	}
	if !found {
		t.Fatalf("reg %s not found in %d rows", cregID, len(regs))
	}
}

// === Event PG ===
func TestPG_EventRepo(t *testing.T) {
	if pgStore == nil {
		t.Skip("PG not available")
	}
	r := pgStore.NewEventRepository()
	id := uid("evt")
	r.Create(domain.AssociationEvent{ID: id, Title: "活动", Status: "draft"})
	r.FindByID(id)
	r.List(0, 20)
	r.Update(domain.AssociationEvent{ID: id, Title: "updated", Status: "published"})
	r.CreateReg(domain.EventRegistration{ID: uid("ereg"), EventID: id, UserID: "u-1", Status: "registered"})
	r.ListRegs(id)
}

// === Resource PG ===
func TestPG_ResourceRepo(t *testing.T) {
	if pgStore == nil {
		t.Skip("PG not available")
	}
	r := pgStore.NewResourceRepository()
	id := uid("res")
	r.Create(domain.IndustryResource{ID: id, OwnerID: "u-1", Name: "无人机", ResType: "drone", Status: "available"})
	r.FindByID(id)
	r.List("drone", 0, 20)
	r.List("", 0, 20)
	r.Update(domain.IndustryResource{ID: id, Name: "updated", Status: "in_use"})
	// C11: 资源预约
	bk, err := r.CreateBooking(domain.IndustryResourceBooking{ID: uid("rb"), ResourceID: id, UserID: "u-2", BookingDate: "2026-08-20", ContactName: "张三", ContactPhone: "13800000000", Status: "pending"})
	if err != nil {
		t.Fatalf("create booking: %v", err)
	}
	if bk.Status != "pending" || bk.ResourceID != id {
		t.Fatalf("booking=%+v, want pending/%s", bk, id)
	}
	// 历史测试残留会累积预约行，断言只针对本次创建的记录
	if byRes, _ := r.ListBookingsByResource(id); !containsBookingID(byRes, bk.ID) {
		t.Fatalf("bookings by resource: booking %s not found in %d rows", bk.ID, len(byRes))
	}
	if byUser, _ := r.ListBookingsByUser("u-2"); !containsBookingID(byUser, bk.ID) {
		t.Fatalf("bookings by user: booking %s not found in %d rows", bk.ID, len(byUser))
	}
	r.Delete(id)
}

// === Emergency PG ===
func TestPG_EmergencyRepo(t *testing.T) {
	if pgStore == nil {
		t.Skip("PG not available")
	}
	r := pgStore.NewEmergencyRepository()
	id := uid("er")
	r.CreateResource(domain.EmergencyResource{ID: id, OwnerID: "u-1", Name: "应急机", ResType: "drone", Status: "standby"})
	r.FindResourceByID(id)
	r.ListResources("", "", 0, 20)
	r.UpdateResource(domain.EmergencyResource{ID: id, Name: "updated", Status: "deployed"})
	r.CreateDispatch(domain.EmergencyDispatch{ID: uid("ed"), ResourceID: id, Status: "active"})
	r.ListDispatches(0, 20)
}

// 回归：end_time 为 NULL（进行中/待响应调度）时 List/FindByID 不能崩溃
// 曾把 NULL 直接 Scan 进 time.Time 导致 500
func TestPG_EmergencyRepo_NullEndTime(t *testing.T) {
	if pgStore == nil {
		t.Skip("PG not available")
	}
	r := pgStore.NewEmergencyRepository()
	resID := uid("er-null")
	r.CreateResource(domain.EmergencyResource{ID: resID, OwnerID: "u-1", Name: "中继站", ResType: "comm", Status: "standby"})

	dID := uid("ed-null")
	// EndTime 零值 → 应存 NULL；CreateDispatch 本身不能报错
	if _, err := r.CreateDispatch(domain.EmergencyDispatch{
		ID: dID, ResourceID: resID, EventDesc: "山区图传中断，正在调度中继站",
		Location: "北碚区缙云山", Status: "ongoing", Commander: "张队",
	}); err != nil {
		t.Fatalf("create dispatch with NULL end_time: %v", err)
	}

	// List 不崩溃且该条 EndTime 为零值
	got, _, err := r.ListDispatches(0, 100)
	if err != nil {
		t.Fatalf("list dispatches with NULL end_time: %v", err)
	}
	var found *domain.EmergencyDispatch
	for i := range got {
		if got[i].ID == dID {
			found = &got[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("dispatch %s not found in list (%d rows)", dID, len(got))
	}
	if !found.EndTime.IsZero() {
		t.Fatalf("NULL end_time should map to zero time, got %v", found.EndTime)
	}

	// FindByID 不崩溃且零值一致
	one, err := r.FindDispatchByID(dID)
	if err != nil {
		t.Fatalf("find dispatch with NULL end_time: %v", err)
	}
	if !one.EndTime.IsZero() {
		t.Fatalf("find: NULL end_time should map to zero time, got %v", one.EndTime)
	}

	// Update 零值 EndTime → 仍为 NULL，不报错
	one.EventDesc = "updated"
	if _, err := r.UpdateDispatch(one); err != nil {
		t.Fatalf("update dispatch keeping NULL end_time: %v", err)
	}
	if after, err := r.FindDispatchByID(dID); err != nil || !after.EndTime.IsZero() {
		t.Fatalf("after update: err=%v end_time=%v, want zero", err, after.EndTime)
	}
}

// === Message PG ===
func TestPG_MessageRepo(t *testing.T) {
	if pgStore == nil {
		t.Skip("PG not available")
	}
	r := pgStore.NewMessageRepository()
	id := uid("msg")
	r.Create(domain.Message{ID: id, ReceiverID: "u-1", SenderID: "sys", IsRead: false})
	r.ListByUser("u-1", false)
	r.MarkRead(id)
	cnt, _ := r.UnreadCount("u-1")
	_ = cnt
}

// === Article PG ===
func TestPG_ArticleRepo(t *testing.T) {
	if pgStore == nil {
		t.Skip("PG not available")
	}
	r := pgStore.NewArticleRepository()
	id := uid("art")
	r.Create(domain.Article{ID: id, Title: "新闻", Category: "policy", Status: "draft"})
	r.FindByID(id)
	r.ListByCategory("", 0, 20)
	r.Update(domain.Article{ID: id, Title: "updated", Status: "published"})
}

// === Review PG ===
func TestPG_ReviewRepo(t *testing.T) {
	if pgStore == nil {
		t.Skip("PG not available")
	}
	r := pgStore.NewReviewRepository()
	id := uid("rev")
	r.Create(domain.Review{ID: id, ReviewerID: "u-1", TargetType: "enterprise", TargetID: "ent-1", Status: "pending"})
	r.ListByTarget("enterprise", "ent-1")
	r.ListAll("", 0, 20)
	r.UpdateStatus(id, "approved")
	r.Delete(id)
}

// === Certificate PG ===
func TestPG_CertificateRepo(t *testing.T) {
	if pgStore == nil {
		t.Skip("PG not available")
	}
	r := pgStore.NewCertificateRepository()
	id := uid("cert")
	r.Create(domain.Certificate{ID: id, UserID: "u-1", CertType: domain.CertCAAC, Status: "pending"})
	r.FindByID(id)
	r.ListByUser("u-1")
	r.UpdateStatus(id, "approved")
	r.ListAll()
}

// === Course PG ===
func TestPG_CourseRepo(t *testing.T) {
	if pgStore == nil {
		t.Skip("PG not available")
	}
	r := pgStore.NewCourseRepository()
	id := uid("crs")
	r.Create(domain.TrainingCourse{ID: id, OrgID: "org-1", Title: "培训", Status: "draft"})
	r.List()
}

// === Pilot PG ===
func TestPG_PilotRepo(t *testing.T) {
	if pgStore == nil {
		t.Skip("PG not available")
	}
	r := pgStore.NewPilotRepository(nil)
	id := uid("pilot")
	r.Create(domain.CertifiedPilot{ID: id, UserID: "u-1", RealName: "飞手", Status: "pending"})
	r.FindByID(id)
	r.List()
	r.UpdateStatus(id, "approved")
}

// === Product PG ===
func TestPG_ProductRepo(t *testing.T) {
	if pgStore == nil {
		t.Skip("PG not available")
	}
	r := pgStore.NewProductRepository()
	id := uid("prod")
	r.Create(domain.DroneProduct{ID: id, SellerID: "u-1", ProdType: domain.ProductDrone, Title: "M300", Status: "listed"})
	r.List("")
	r.List("drone")
}

// === Policy PG ===
func TestPG_PolicyRepo(t *testing.T) {
	if pgStore == nil {
		t.Skip("PG not available")
	}
	r := pgStore.NewPolicyRepository()
	id := uid("pol")
	r.Create(domain.InsurancePolicy{ID: id, UserID: "u-1", DroneModel: "M300", Status: "active"})
	r.ListByUser("u-1")
}

// === Escrow PG ===
func TestPG_EscrowRepo(t *testing.T) {
	if pgStore == nil {
		t.Skip("PG not available")
	}
	r := pgStore.NewEscrowRepository()
	uid := uid("eu")
	// C6 接口变更：资金操作原子化，Deposit 即开户
	r.Deposit(uid, 100000, domain.EscrowTransaction{ID: "tx-" + uid, FromUser: "sys", ToUser: uid, AmountFen: 100000, TxType: "deposit", Status: "completed", CreatedAt: time.Now()})
	r.GetAccount(uid)
	r.Freeze(uid, 30000, domain.EscrowTransaction{ID: "tx-f-" + uid, FromUser: uid, ToUser: "escrow", AmountFen: 30000, TxType: "freeze", Status: "completed", CreatedAt: time.Now()})
	r.Refund(uid, 30000, domain.EscrowTransaction{ID: "tx-r-" + uid, FromUser: "escrow", ToUser: uid, AmountFen: 30000, TxType: "refund", Status: "completed", CreatedAt: time.Now()})
	r.ListTransactions(uid)
}

// === Enrollment PG ===
func TestPG_EnrollmentRepo(t *testing.T) {
	if pgStore == nil {
		t.Skip("PG not available")
	}
	r := pgStore.NewEnrollmentRepository()
	id := uid("enr")
	r.Create(domain.Enrollment{ID: id, CourseID: "crs-1", UserID: "u-1"})
	r.ListByCourse("crs-1")
	r.FindByUserAndCourse("u-1", "crs-1")
}
