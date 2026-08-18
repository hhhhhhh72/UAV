package memory_test

import (
	"context"
	"testing"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
	"drone-platform/internal/repository/memory"
)

// === Demand Repository ===
func TestDemandRepoFull(t *testing.T) {
	r := memory.NewDemandRepository(nil)
	// Create
	d := domain.Demand{ID: "d-1", Title: "测试需求", Status: domain.DemandPublished, BizType: domain.BizCableInspection, PublisherID: "u-1"}
	c, err := r.Create(context.Background(), d)
	if err != nil || c.ID != "d-1" {
		t.Fatal("create failed")
	}
	// FindByID
	f, err := r.FindByID(context.Background(), "d-1")
	if err != nil || f.Title != "测试需求" {
		t.Fatal("find failed")
	}
	// List
	list, err := r.List(context.Background(), repository.DemandFilter{})
	_ = list
	if err != nil {
		t.Fatal(err)
	}
	// Search
	results, err := r.Search(context.Background(), "测试")
	if err != nil || len(results) == 0 {
		t.Fatal("search failed")
	}
	// SetStatus
	d2, err := r.SetStatus(context.Background(), "d-1", domain.DemandPublished)
	if err != nil || d2.Status != domain.DemandPublished {
		t.Fatal("set status failed")
	}
	// CompareAndSetStatus
	ok, d3, err := r.CompareAndSetStatus(context.Background(), "d-1", domain.DemandPublished, domain.DemandCompleted)
	if err != nil || !ok || d3.Status != domain.DemandCompleted {
		t.Fatal("CAS failed")
	}
	// CAS with wrong old status
	ok2, _, _ := r.CompareAndSetStatus(context.Background(), "d-1", domain.DemandPublished, domain.DemandCompleted)
	if ok2 {
		t.Fatal("CAS should fail with wrong old status")
	}
	// Update（乐观锁：必须基于最新版本——CAS 已把版本自增，旧对象会冲突）
	fresh, err := r.FindByID(context.Background(), "d-1")
	if err != nil {
		t.Fatal("find before update:", err)
	}
	fresh.Title = "updated"
	u, err := r.Update(context.Background(), fresh)
	if err != nil {
		t.Fatal("update failed:", err)
	}
	_ = u
	// 旧版本对象更新必须报并发冲突
	old := d
	old.Title = "stale"
	if _, err := r.Update(context.Background(), old); err == nil {
		t.Fatal("stale version update should conflict")
	}
}

// === Enterprise Repository ===
func TestEnterpriseRepoFull(t *testing.T) {
	r := memory.NewEnterpriseRepository(nil)
	e := domain.Enterprise{ID: "ent-1", OwnerUserID: "u-1", Name: "测试企业", Status: domain.EnterpriseSubmitted}
	c, _ := r.Create(context.Background(), e)
	_ = c
	r.FindByID(context.Background(), "ent-1")
	r.FindByOwner(context.Background(), "u-1")
	r.Pending(context.Background())
	r.Search(context.Background(), "测试")
	r.ListByStatus(context.Background(), "submitted", 0, 20)
	r.Update(context.Background(), "ent-1", domain.Enterprise{ID: "ent-1", OwnerUserID: "u-1", Name: "updated", Status: domain.EnterpriseApproved})
}

// === Employment Repository ===
func TestEmploymentRepoFull(t *testing.T) {
	r := memory.NewEmploymentRepository()
	er := domain.EmploymentRequest{ID: "emp-1", EnterpriseID: "ent-1", Position: "飞手", Headcount: 5}
	r.Create(context.Background(), er)
	r.ListByEnterprise(context.Background(), "ent-1", 0, 20)
	r.ListAll(context.Background(), 0, 20)
}

// === Contract Repository ===
func TestContractRepoFull(t *testing.T) {
	r := memory.NewContractRepository()
	c := domain.Contract{ID: "ctr-1", EnterpriseID: "ent-1", TemplateID: "tpl-1", Status: domain.ContractDraft}
	r.Create(context.Background(), c)
	r.FindByID(context.Background(), "ctr-1")
	r.ListByEnterprise(context.Background(), "ent-1", 0, 20)
	r.ListAll(context.Background(), 0, 20)
	r.UpdateStatus(context.Background(), "ctr-1", domain.ContractSent)
}

// === Bid Repository ===
func TestBidRepoFull(t *testing.T) {
	r := memory.NewBidRepository()
	b := domain.DemandBid{ID: "bid-1", DemandID: "d-1", BidderID: "u-2", AmountFen: 50000, Status: "pending"}
	r.Create(context.Background(), b)
	r.FindByID(context.Background(), "bid-1")
	r.ListByDemand(context.Background(), "d-1")
	r.UpdateStatus(context.Background(), "bid-1", "accepted")
}

// === Job Repository ===
func TestJobRepoFull(t *testing.T) {
	r := memory.NewJobRepository()
	j := domain.Job{ID: "job-1", EnterpriseID: "ent-1", Title: "飞手招聘", Status: domain.JobPublished}
	r.Create(context.Background(), j)
	r.FindByID(context.Background(), "job-1")
	r.ListByEnterprise(context.Background(), "ent-1")
	r.ListPublished(context.Background(), 0, 20)
	r.Update(context.Background(), "job-1", domain.Job{ID: "job-1", EnterpriseID: "ent-1", Title: "updated", Status: domain.JobClosed})
}

// === Resume Repository ===
func TestResumeRepoFull(t *testing.T) {
	r := memory.NewResumeRepository()
	res := domain.Resume{ID: "res-1", UserID: "u-1", Title: "我的简历"}
	r.Create(context.Background(), res)
	r.FindByID(context.Background(), "res-1")
	r.ListByUser(context.Background(), "u-1")
	r.Update(context.Background(), "res-1", domain.Resume{ID: "res-1", UserID: "u-1", Title: "updated"})
}

// === Application Repository ===
func TestJobApplicationRepoFull(t *testing.T) {
	r := memory.NewJobApplicationRepository()
	a := domain.JobApplication{ID: "app-1", JobID: "job-1", ResumeID: "res-1", ApplicantID: "u-1", Status: domain.AppSubmitted}
	r.Create(context.Background(), a)
	r.ListByJob(context.Background(), "job-1")
	r.ListByApplicant(context.Background(), "u-1")
	r.UpdateStatus(context.Background(), "app-1", domain.AppViewed)
}

// === Post Repository ===
func TestPostRepoFull(t *testing.T) {
	r := memory.NewPostRepository()
	p := domain.Post{ID: "post-1", AuthorID: "u-1", Title: "帖子", Status: "published"}
	r.Create(context.Background(), p)
	r.FindByID(context.Background(), "post-1")
	r.ListPublished(context.Background(), 0, 20)
	r.ListByAuthor(context.Background(), "u-1")
	r.Update(context.Background(), "post-1", domain.Post{ID: "post-1", AuthorID: "u-1", Title: "updated", Status: "removed"})
}

// === Comment + Report Repository ===
func TestCommentAndReportRepo(t *testing.T) {
	cr := memory.NewCommentRepository()
	cr.Create(context.Background(), domain.Comment{ID: "cmt-1", PostID: "post-1", AuthorID: "u-1"})
	cr.ListByPost(context.Background(), "post-1")

	rr := memory.NewReportRepository()
	rr.Create(context.Background(), domain.Report{ID: "rpt-1", ReporterID: "u-1", ResourceType: "post", ResourceID: "post-1"})
	rr.ListPending(context.Background(), 0, 20)
}

// === Listing Repository ===
func TestListingRepoFull(t *testing.T) {
	r := memory.NewListingRepository()
	l := domain.Listing{ID: "lst-1", SellerID: "u-1", Title: "二手无人机", Status: "listed"}
	r.Create(context.Background(), l)
	r.FindByID(context.Background(), "lst-1")
	r.ListByStatus(context.Background(), "listed", 0, 20)
	r.ListBySeller(context.Background(), "u-1")
	r.Update(context.Background(), "lst-1", domain.Listing{ID: "lst-1", SellerID: "u-1", Title: "updated", Status: "sold"})
	r.AddFavorite(context.Background(), "lst-1", "u-2")
	r.RemoveFavorite(context.Background(), "lst-1", "u-2")
}

// === Labour Repository ===
func TestLabourRepoFull(t *testing.T) {
	r := memory.NewLabourOrderRepository()
	lo := domain.LabourOrder{ID: "lo-1", EmployerID: "u-1", Title: "用工需求", Status: "open"}
	r.Create(context.Background(), lo)
	r.FindByID(context.Background(), "lo-1")
	r.ListByEmployer(context.Background(), "u-1")
	r.ListAll(context.Background(), 0, 20)
	// Quote
	r.CreateQuote(context.Background(), domain.LabourQuote{ID: "q-1", OrderID: "lo-1", QuoterID: "u-2", AmountFen: 10000})
	r.ListQuotes(context.Background(), "lo-1")
	// Assignment
	r.CreateAssignment(context.Background(), domain.Assignment{ID: "asgn-1", OrderID: "lo-1", WorkerID: "u-3"})
}

// === User + RefreshToken Repository ===
func TestUserRepoFull(t *testing.T) {
	r := memory.NewUserRepository(nil)
	u := domain.User{ID: "usr-1", WechatOpenID: "wx-openid-1", Status: "active", Role: domain.RoleIndividual}
	r.Create(context.Background(), u)
	r.FindByOpenID(context.Background(), "wx-openid-1")
	r.FindByID(context.Background(), "usr-1")
	r.All(context.Background())
	r.UpdateRole(context.Background(), "usr-1", domain.RoleEnterprise)
}

func TestRefreshTokenRepoFull(t *testing.T) {
	r := memory.NewRefreshTokenRepository()
	r.Store(context.Background(), "usr-1", "hash-abc", time.Now().Add(time.Hour))
	uid, _, revoked, err := r.Find(context.Background(), "hash-abc")
	if err != nil || uid != "usr-1" || revoked {
		t.Fatal("find failed")
	}
	r.Revoke(context.Background(), "hash-abc")
	_, _, revoked2, _ := r.Find(context.Background(), "hash-abc")
	if !revoked2 {
		t.Fatal("should be revoked")
	}
}

// === Training Repos ===
func TestTrainingReposFull(t *testing.T) {
	certR := memory.NewCertificateRepository()
	certR.Create(context.Background(), domain.Certificate{ID: "cert-1", UserID: "u-1", CertType: domain.CertCAAC, Status: "pending"})
	certR.FindByID(context.Background(), "cert-1")
	certR.ListByUser(context.Background(), "u-1")
	certR.UpdateStatus(context.Background(), "cert-1", "approved")
	certR.ListAll(context.Background())

	crR := memory.NewCourseRepository()
	crR.Create(context.Background(), domain.TrainingCourse{ID: "crs-1", OrgID: "org-1", Title: "CAAC培训"})
	crR.List(context.Background())

	instR := memory.NewInstructorRepository()
	instR.Create(context.Background(), domain.Instructor{ID: "inst-1", UserID: "u-1", Name: "教练", Status: "pending"})
	instR.FindByID(context.Background(), "inst-1")
	instR.List(context.Background())
	instR.UpdateStatus(context.Background(), "inst-1", "approved")

	pilotR := memory.NewPilotRepository(nil)
	pilotR.Create(context.Background(), domain.CertifiedPilot{ID: "pilot-1", UserID: "u-1", RealName: "飞手", Status: "pending"})
	pilotR.FindByID(context.Background(), "pilot-1")
	pilotR.List(context.Background())
	pilotR.UpdateStatus(context.Background(), "pilot-1", "approved")
}

// === Trading Repos ===
func TestTradingReposFull(t *testing.T) {
	pr := memory.NewProductRepository()
	pr.Create(context.Background(), domain.DroneProduct{ID: "prod-1", SellerID: "u-1", ProdType: domain.ProductDrone, Title: "M300", Status: "listed"})
	pr.List(context.Background(), "")
	pr.List(context.Background(), "drone")

	rr := memory.NewRepairRepository()
	rr.Create(context.Background(), domain.RepairOrder{ID: "rep-1", CustomerID: "u-1", Status: "submitted"})
	rr.ListByUser(context.Background(), "u-1")
}

// === Insurance + Finance Repos ===
func TestInsuranceFinanceRepos(t *testing.T) {
	polR := memory.NewPolicyRepository()
	polR.Create(context.Background(), domain.InsurancePolicy{ID: "pol-1", UserID: "u-1", Status: "active"})
	polR.ListByUser(context.Background(), "u-1")

	inspR := memory.NewInspectionRepository()
	inspR.Create(context.Background(), domain.AnnualInspection{ID: "insp-1", UserID: "u-1", Status: "pending"})
	inspR.ListByUser(context.Background(), "u-1")
	inspR.ListAll(context.Background())

	loanR := memory.NewLoanRepository()
	loanR.Create(context.Background(), domain.LoanApplication{ID: "loan-1", UserID: "u-1", AmountFen: 100000, Status: "submitted"})
	loanR.ListByUser(context.Background(), "u-1")
}

// === Content Repos (Message, Article, Review) ===
func TestContentReposFull(t *testing.T) {
	msgR := memory.NewMessageRepository()
	msgR.Create(context.Background(), domain.Message{ID: "msg-1", ReceiverID: "u-1", SenderID: "sys", IsRead: false})
	msgR.ListByUser(context.Background(), "u-1", false)
	msgR.ListByUser(context.Background(), "u-1", true)
	msgR.MarkRead(context.Background(), "msg-1")
	cnt, _ := msgR.UnreadCount(context.Background(), "u-1")
	if cnt != 0 {
		t.Fatalf("unread: %d", cnt)
	}

	artR := memory.NewArticleRepository()
	artR.Create(context.Background(), domain.Article{ID: "art-1", Title: "新闻", Category: "policy", Status: "draft"})
	artR.FindByID(context.Background(), "art-1")
	artR.ListByCategory(context.Background(), "", 0, 20)
	artR.Update(context.Background(), domain.Article{ID: "art-1", Title: "updated", Status: "published"})

	revR := memory.NewReviewRepository()
	revR.Create(context.Background(), domain.Review{ID: "rev-1", ReviewerID: "u-1", TargetType: "enterprise", TargetID: "ent-1", Rating: 5, Status: "pending"})
	revR.ListByTarget(context.Background(), "enterprise", "ent-1")
	revR.ListAll(context.Background(), "", 0, 20)
	revR.UpdateStatus(context.Background(), "rev-1", "approved")
	revR.Delete(context.Background(), "rev-1")
}

// === Venue + Enrollment + TradeOrder + Escrow ===
func TestVenueEnrollTradeEscrowRepos(t *testing.T) {
	vR := memory.NewVenueRepository()
	vR.Create(context.Background(), domain.Venue{ID: "v-1", OwnerID: "u-1", Name: "飞行基地", Status: "available"})
	vR.List(context.Background(), "")
	vR.FindByID(context.Background(), "v-1")
	vR.CreateBooking(context.Background(), domain.VenueBooking{ID: "bk-1", VenueID: "v-1", UserID: "u-2", Status: "booked"})
	vR.ListBookings(context.Background(), "v-1")

	eR := memory.NewEnrollmentRepository()
	eR.Create(context.Background(), domain.Enrollment{ID: "enr-1", CourseID: "crs-1", UserID: "u-1"})
	eR.ListByCourse(context.Background(), "crs-1")
	eR.FindByUserAndCourse(context.Background(), "u-1", "crs-1")

	tR := memory.NewTradeOrderRepository()
	tR.Create(context.Background(), domain.TradeOrder{ID: "to-1", ProductID: "prod-1", BuyerID: "u-1", SellerID: "u-2", Status: "pending"})
	tR.FindByID(context.Background(), "to-1")
	tR.UpdateStatus(context.Background(), "to-1", "paid")
	tR.ListByUser(context.Background(), "u-1")

	escR := memory.NewEscrowRepository()
	escR.GetAccount(context.Background(), "u-1")
	// C6 接口变更：资金操作原子化，Deposit 即开户
	escR.Deposit(context.Background(), "u-1", 100000, domain.EscrowTransaction{ID: "tx-1", FromUser: "sys", ToUser: "u-1", AmountFen: 100000, TxType: "deposit"})
	escR.ListTransactions(context.Background(), "u-1")
}

// === New Biz Repos ===
func TestNewBizReposFull(t *testing.T) {
	expR := memory.NewExpertRepository()
	expR.Create(context.Background(), domain.Expert{ID: "exp-1", Name: "专家", Field: "无人机"})
	expR.FindByID(context.Background(), "exp-1")
	expR.List(context.Background(), "")
	expR.Update(context.Background(), domain.Expert{ID: "exp-1", Name: "updated", Field: "低空"})
	expR.Delete(context.Background(), "exp-1")

	caseR := memory.NewCaseRepository()
	caseR.Create(context.Background(), domain.CaseEntry{ID: "case-1", Title: "案例", Category: "logistics"})
	caseR.FindByID(context.Background(), "case-1")
	caseR.List(context.Background(), "", 0, 20)
	caseR.Update(context.Background(), domain.CaseEntry{ID: "case-1", Title: "updated"})
	caseR.Delete(context.Background(), "case-1")

	compR := memory.NewComplianceRepository()
	compR.CreateDoc(context.Background(), domain.ComplianceDoc{ID: "cd-1", Title: "条例"})
	compR.FindDocByID(context.Background(), "cd-1")
	compR.ListDocs(context.Background(), "", 0, 20)
	compR.UpdateDoc(context.Background(), domain.ComplianceDoc{ID: "cd-1", Title: "updated"})
	compR.DeleteDoc(context.Background(), "cd-1")
	compR.CreateStandard(context.Background(), domain.StandardDoc{ID: "std-1", Title: "标准"})
	compR.ListStandards(context.Background(), "", 0, 20)

	achR := memory.NewAchievementRepository()
	achR.Create(context.Background(), domain.Achievement{ID: "ach-1", OwnerID: "u-1", Title: "专利"})
	achR.FindByID(context.Background(), "ach-1")
	achR.List(context.Background(), "", 0, 20)
	achR.Update(context.Background(), domain.Achievement{ID: "ach-1", Title: "updated"})
	achR.Delete(context.Background(), "ach-1")

	rdR := memory.NewRDChallengeRepository()
	rdR.Create(context.Background(), domain.RDChallenge{ID: "rd-1", PosterID: "u-1", Title: "难题"})
	rdR.FindByID(context.Background(), "rd-1")
	rdR.List(context.Background(), "", 0, 20)
	rdR.Update(context.Background(), domain.RDChallenge{ID: "rd-1", Title: "updated"})

	rpR := memory.NewResearchProjectRepository()
	rpR.Create(context.Background(), domain.ResearchProject{ID: "rp-1", Title: "课题"})
	rpR.FindByID(context.Background(), "rp-1")
	rpR.List(context.Background(), 0, 20)
	rpR.Update(context.Background(), domain.ResearchProject{ID: "rp-1", Title: "updated"})

	paR := memory.NewProjectAppRepository()
	paR.Create(context.Background(), domain.ProjectApplication{ID: "pa-1", ApplicantID: "u-1", ProjectName: "申报"})
	paR.FindByID(context.Background(), "pa-1")
	paR.ListByUser(context.Background(), "u-1")
	paR.ListAll(context.Background(), "", 0, 20)
	paR.Update(context.Background(), domain.ProjectApplication{ID: "pa-1", Status: "submitted"})

	compR2 := memory.NewCompetitionRepository(nil)
	compR2.Create(context.Background(), domain.Competition{ID: "comp-1", Title: "比赛"})
	compR2.FindByID(context.Background(), "comp-1")
	compR2.List(context.Background(), 0, 20)
	compR2.Update(context.Background(), domain.Competition{ID: "comp-1", Title: "updated"})
	compR2.CreateReg(context.Background(), domain.CompetitionReg{ID: "reg-1", CompetitionID: "comp-1", UserID: "u-1"})
	compR2.ListRegs(context.Background(), "comp-1")

	evR := memory.NewEventRepository()
	evR.Create(context.Background(), domain.AssociationEvent{ID: "evt-1", Title: "活动"})
	evR.FindByID(context.Background(), "evt-1")
	evR.List(context.Background(), 0, 20)
	evR.Update(context.Background(), domain.AssociationEvent{ID: "evt-1", Title: "updated"})
	evR.CreateReg(context.Background(), domain.EventRegistration{ID: "erg-1", EventID: "evt-1", UserID: "u-1"})
	evR.ListRegs(context.Background(), "evt-1")

	portR := memory.NewPortfolioRepository()
	portR.Create(context.Background(), domain.MemberPortfolio{ID: "port-1", EnterpriseID: "ent-1", Name: "品牌"})
	portR.FindByID(context.Background(), "port-1")
	portR.ListByEnterprise(context.Background(), "ent-1")
	portR.ListPublished(context.Background(), 0, 20)
	portR.Update(context.Background(), domain.MemberPortfolio{ID: "port-1", Name: "updated"})

	irR := memory.NewIndustryReportRepository()
	irR.Create(context.Background(), domain.IndustryReport{ID: "ir-1", Title: "报告"})
	irR.FindByID(context.Background(), "ir-1")
	irR.List(context.Background(), 0, 20)
	irR.Update(context.Background(), domain.IndustryReport{ID: "ir-1", Title: "updated"})
	irR.Delete(context.Background(), "ir-1")

	resR := memory.NewResourceRepository()
	resR.Create(context.Background(), domain.IndustryResource{ID: "res-1", OwnerID: "u-1", Name: "无人机"})
	resR.FindByID(context.Background(), "res-1")
	resR.List(context.Background(), "", 0, 20)
	resR.Update(context.Background(), domain.IndustryResource{ID: "res-1", Name: "updated"})

	emR := memory.NewEmergencyRepository()
	emR.CreateResource(context.Background(), domain.EmergencyResource{ID: "er-1", OwnerID: "u-1", Name: "应急机"})
	emR.FindResourceByID(context.Background(), "er-1")
	emR.ListResources(context.Background(), "", "", 0, 20)
	emR.UpdateResource(context.Background(), domain.EmergencyResource{ID: "er-1", Name: "updated"})
	emR.CreateDispatch(context.Background(), domain.EmergencyDispatch{ID: "ed-1", ResourceID: "er-1"})
	emR.ListDispatches(context.Background(), 0, 20)
}
