package postgres_test

import (
	"context"
	"fmt"
	"testing"
	"time"

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
	if pgStore == nil { t.Skip("PG not available") }
	r := pgStore.NewExpertRepository()
	id := uid("exp")
	e, err := r.Create(domain.Expert{ID: id, Name: "PG专家", Title: "教授", Field: "无人机", Status: "active"})
	if err != nil { t.Fatal(err) }
	if _, err := r.FindByID(e.ID); err != nil { t.Fatal(err) }
	if list, _ := r.List("无人机"); len(list) == 0 { t.Fatal("list empty") }
	r.Update(domain.Expert{ID: id, Name: "updated", Title: "副教授", Field: "低空", Status: "active"})
	r.Delete(id)
}

// === Case PG ===
func TestPG_CaseRepo(t *testing.T) {
	if pgStore == nil { t.Skip("PG not available") }
	r := pgStore.NewCaseRepository()
	id := uid("case")
	r.Create(domain.CaseEntry{ID: id, Title: "案例", Category: "logistics", Status: "draft"})
	r.FindByID(id)
	r.List("", 0, 20)
	r.Delete(id)
}

// === Compliance PG ===
func TestPG_ComplianceRepo(t *testing.T) {
	if pgStore == nil { t.Skip("PG not available") }
	r := pgStore.NewComplianceRepository()
	id := uid("cdoc")
	r.CreateDoc(domain.ComplianceDoc{ID: id, Title: "条例", Category: "policy", Status: "draft"})
	r.FindDocByID(id)
	r.ListDocs("", 0, 20)
	r.UpdateDoc(domain.ComplianceDoc{ID: id, Title: "updated", Category: "regulation", Status: "published"})
	r.DeleteDoc(id)
	stdID := uid("std")
	r.CreateStandard(domain.StandardDoc{ID: stdID, Title: "标准", StandardNo: "T/CDA-001", Status: "draft"})
	r.ListStandards("", 0, 20)
}

// === Report PG ===
func TestPG_ReportRepo(t *testing.T) {
	if pgStore == nil { t.Skip("PG not available") }
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
	if pgStore == nil { t.Skip("PG not available") }
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
	if pgStore == nil { t.Skip("PG not available") }
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
	if pgStore == nil { t.Skip("PG not available") }
	r := pgStore.NewRDChallengeRepository()
	id := uid("rd")
	r.Create(domain.RDChallenge{ID: id, PosterID: "u-1", Title: "难题", Status: "open"})
	r.FindByID(id)
	r.List("", 0, 20)
	r.Update(domain.RDChallenge{ID: id, Title: "updated", Status: "in_progress"})
}

// === ResearchProject PG ===
func TestPG_ResearchProjRepo(t *testing.T) {
	if pgStore == nil { t.Skip("PG not available") }
	r := pgStore.NewResearchProjectRepository()
	id := uid("rp")
	r.Create(domain.ResearchProject{ID: id, Title: "课题", Status: "planning"})
	r.FindByID(id)
	r.List(0, 20)
	r.Update(domain.ResearchProject{ID: id, Title: "updated", Status: "active"})
}

// === ProjectApp PG ===
func TestPG_ProjectAppRepo(t *testing.T) {
	if pgStore == nil { t.Skip("PG not available") }
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
	if pgStore == nil { t.Skip("PG not available") }
	r := pgStore.NewCompetitionRepository()
	id := uid("comp")
	r.Create(domain.Competition{ID: id, Title: "比赛", Status: "draft"})
	r.FindByID(id)
	r.List(0, 20)
	r.Update(domain.Competition{ID: id, Title: "updated", Status: "open"})
	r.CreateReg(domain.CompetitionReg{ID: uid("creg"), CompetitionID: id, UserID: "u-1", Status: "submitted"})
	r.ListRegs(id)
}

// === Event PG ===
func TestPG_EventRepo(t *testing.T) {
	if pgStore == nil { t.Skip("PG not available") }
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
	if pgStore == nil { t.Skip("PG not available") }
	r := pgStore.NewResourceRepository()
	id := uid("res")
	r.Create(domain.IndustryResource{ID: id, OwnerID: "u-1", Name: "无人机", ResType: "drone", Status: "available"})
	r.FindByID(id)
	r.List("drone", 0, 20)
	r.List("", 0, 20)
	r.Update(domain.IndustryResource{ID: id, Name: "updated", Status: "in_use"})
}

// === Emergency PG ===
func TestPG_EmergencyRepo(t *testing.T) {
	if pgStore == nil { t.Skip("PG not available") }
	r := pgStore.NewEmergencyRepository()
	id := uid("er")
	r.CreateResource(domain.EmergencyResource{ID: id, OwnerID: "u-1", Name: "应急机", ResType: "drone", Status: "standby"})
	r.FindResourceByID(id)
	r.ListResources(0, 20)
	r.UpdateResource(domain.EmergencyResource{ID: id, Name: "updated", Status: "deployed"})
	r.CreateDispatch(domain.EmergencyDispatch{ID: uid("ed"), ResourceID: id, Status: "active"})
	r.ListDispatches(0, 20)
}

// === Message PG ===
func TestPG_MessageRepo(t *testing.T) {
	if pgStore == nil { t.Skip("PG not available") }
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
	if pgStore == nil { t.Skip("PG not available") }
	r := pgStore.NewArticleRepository()
	id := uid("art")
	r.Create(domain.Article{ID: id, Title: "新闻", Category: "policy", Status: "draft"})
	r.FindByID(id)
	r.ListByCategory("", 0, 20)
	r.Update(domain.Article{ID: id, Title: "updated", Status: "published"})
}

// === Review PG ===
func TestPG_ReviewRepo(t *testing.T) {
	if pgStore == nil { t.Skip("PG not available") }
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
	if pgStore == nil { t.Skip("PG not available") }
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
	if pgStore == nil { t.Skip("PG not available") }
	r := pgStore.NewCourseRepository()
	id := uid("crs")
	r.Create(domain.TrainingCourse{ID: id, OrgID: "org-1", Title: "培训", Status: "draft"})
	r.List()
}

// === Pilot PG ===
func TestPG_PilotRepo(t *testing.T) {
	if pgStore == nil { t.Skip("PG not available") }
	r := pgStore.NewPilotRepository(nil)
	id := uid("pilot")
	r.Create(domain.CertifiedPilot{ID: id, UserID: "u-1", RealName: "飞手", Status: "pending"})
	r.FindByID(id)
	r.List()
	r.UpdateStatus(id, "approved")
}

// === Product PG ===
func TestPG_ProductRepo(t *testing.T) {
	if pgStore == nil { t.Skip("PG not available") }
	r := pgStore.NewProductRepository()
	id := uid("prod")
	r.Create(domain.DroneProduct{ID: id, SellerID: "u-1", ProdType: domain.ProductDrone, Title: "M300", Status: "listed"})
	r.List("")
	r.List("drone")
}

// === Policy PG ===
func TestPG_PolicyRepo(t *testing.T) {
	if pgStore == nil { t.Skip("PG not available") }
	r := pgStore.NewPolicyRepository()
	id := uid("pol")
	r.Create(domain.InsurancePolicy{ID: id, UserID: "u-1", DroneModel: "M300", Status: "active"})
	r.ListByUser("u-1")
}

// === Escrow PG ===
func TestPG_EscrowRepo(t *testing.T) {
	if pgStore == nil { t.Skip("PG not available") }
	r := pgStore.NewEscrowRepository()
	uid := uid("eu")
	r.UpsertAccount(domain.EscrowAccount{UserID: uid, BalanceFen: 100000})
	r.GetAccount(uid)
	r.CreateTransaction(domain.EscrowTransaction{ID: "tx-" + uid, FromUser: "sys", ToUser: uid, AmountFen: 100000, TxType: "deposit"})
	r.ListTransactions(uid)
}

// === Enrollment PG ===
func TestPG_EnrollmentRepo(t *testing.T) {
	if pgStore == nil { t.Skip("PG not available") }
	r := pgStore.NewEnrollmentRepository()
	id := uid("enr")
	r.Create(domain.Enrollment{ID: id, CourseID: "crs-1", UserID: "u-1"})
	r.ListByCourse("crs-1")
	r.FindByUserAndCourse("u-1", "crs-1")
}
