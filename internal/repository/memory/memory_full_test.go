package memory_test

import (
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
	c, err := r.Create(d)
	if err != nil || c.ID != "d-1" {
		t.Fatal("create failed")
	}
	// FindByID
	f, err := r.FindByID("d-1")
	if err != nil || f.Title != "测试需求" {
		t.Fatal("find failed")
	}
	// List
	list, err := r.List(repository.DemandFilter{})
	_ = list
	if err != nil {
		t.Fatal(err)
	}
	// Search
	results, err := r.Search("测试")
	if err != nil || len(results) == 0 {
		t.Fatal("search failed")
	}
	// SetStatus
	d2, err := r.SetStatus("d-1", domain.DemandPublished)
	if err != nil || d2.Status != domain.DemandPublished {
		t.Fatal("set status failed")
	}
	// CompareAndSetStatus
	ok, d3, err := r.CompareAndSetStatus("d-1", domain.DemandPublished, domain.DemandCompleted)
	if err != nil || !ok || d3.Status != domain.DemandCompleted {
		t.Fatal("CAS failed")
	}
	// CAS with wrong old status
	ok2, _, _ := r.CompareAndSetStatus("d-1", domain.DemandPublished, domain.DemandCompleted)
	if ok2 {
		t.Fatal("CAS should fail with wrong old status")
	}
	// Update
	d.Title = "updated"
	u, err := r.Update(d)
	if err != nil {
		t.Fatal("update failed:", err)
	}
	_ = u
}

// === Enterprise Repository ===
func TestEnterpriseRepoFull(t *testing.T) {
	r := memory.NewEnterpriseRepository(nil)
	e := domain.Enterprise{ID: "ent-1", OwnerUserID: "u-1", Name: "测试企业", Status: domain.EnterpriseSubmitted}
	c, _ := r.Create(e)
	_ = c
	r.FindByID("ent-1")
	r.FindByOwner("u-1")
	r.Pending()
	r.Search("测试")
	r.ListByStatus("submitted", 0, 20)
	r.Update("ent-1", domain.Enterprise{ID: "ent-1", OwnerUserID: "u-1", Name: "updated", Status: domain.EnterpriseApproved})
}

// === Employment Repository ===
func TestEmploymentRepoFull(t *testing.T) {
	r := memory.NewEmploymentRepository()
	er := domain.EmploymentRequest{ID: "emp-1", EnterpriseID: "ent-1", Position: "飞手", Headcount: 5}
	r.Create(er)
	r.ListByEnterprise("ent-1", 0, 20)
	r.ListAll(0, 20)
}

// === Contract Repository ===
func TestContractRepoFull(t *testing.T) {
	r := memory.NewContractRepository()
	c := domain.Contract{ID: "ctr-1", EnterpriseID: "ent-1", TemplateID: "tpl-1", Status: domain.ContractDraft}
	r.Create(c)
	r.FindByID("ctr-1")
	r.ListByEnterprise("ent-1", 0, 20)
	r.ListAll(0, 20)
	r.UpdateStatus("ctr-1", domain.ContractSent)
}

// === Bid Repository ===
func TestBidRepoFull(t *testing.T) {
	r := memory.NewBidRepository()
	b := domain.DemandBid{ID: "bid-1", DemandID: "d-1", BidderID: "u-2", AmountFen: 50000, Status: "pending"}
	r.Create(b)
	r.FindByID("bid-1")
	r.ListByDemand("d-1")
	r.UpdateStatus("bid-1", "accepted")
}

// === Job Repository ===
func TestJobRepoFull(t *testing.T) {
	r := memory.NewJobRepository()
	j := domain.Job{ID: "job-1", EnterpriseID: "ent-1", Title: "飞手招聘", Status: domain.JobPublished}
	r.Create(j)
	r.FindByID("job-1")
	r.ListByEnterprise("ent-1")
	r.ListPublished(0, 20)
	r.Update("job-1", domain.Job{ID: "job-1", EnterpriseID: "ent-1", Title: "updated", Status: domain.JobClosed})
}

// === Resume Repository ===
func TestResumeRepoFull(t *testing.T) {
	r := memory.NewResumeRepository()
	res := domain.Resume{ID: "res-1", UserID: "u-1", Title: "我的简历"}
	r.Create(res)
	r.FindByID("res-1")
	r.ListByUser("u-1")
	r.Update("res-1", domain.Resume{ID: "res-1", UserID: "u-1", Title: "updated"})
}

// === Application Repository ===
func TestJobApplicationRepoFull(t *testing.T) {
	r := memory.NewJobApplicationRepository()
	a := domain.JobApplication{ID: "app-1", JobID: "job-1", ResumeID: "res-1", ApplicantID: "u-1", Status: domain.AppSubmitted}
	r.Create(a)
	r.ListByJob("job-1")
	r.ListByApplicant("u-1")
	r.UpdateStatus("app-1", domain.AppViewed)
}

// === Post Repository ===
func TestPostRepoFull(t *testing.T) {
	r := memory.NewPostRepository()
	p := domain.Post{ID: "post-1", AuthorID: "u-1", Title: "帖子", Status: "published"}
	r.Create(p)
	r.FindByID("post-1")
	r.ListPublished(0, 20)
	r.ListByAuthor("u-1")
	r.Update("post-1", domain.Post{ID: "post-1", AuthorID: "u-1", Title: "updated", Status: "removed"})
}

// === Comment + Report Repository ===
func TestCommentAndReportRepo(t *testing.T) {
	cr := memory.NewCommentRepository()
	cr.Create(domain.Comment{ID: "cmt-1", PostID: "post-1", AuthorID: "u-1"})
	cr.ListByPost("post-1")

	rr := memory.NewReportRepository()
	rr.Create(domain.Report{ID: "rpt-1", ReporterID: "u-1", ResourceType: "post", ResourceID: "post-1"})
	rr.ListPending(0, 20)
}

// === Listing Repository ===
func TestListingRepoFull(t *testing.T) {
	r := memory.NewListingRepository()
	l := domain.Listing{ID: "lst-1", SellerID: "u-1", Title: "二手无人机", Status: "listed"}
	r.Create(l)
	r.FindByID("lst-1")
	r.ListByStatus("listed", 0, 20)
	r.ListBySeller("u-1")
	r.Update("lst-1", domain.Listing{ID: "lst-1", SellerID: "u-1", Title: "updated", Status: "sold"})
	r.AddFavorite("lst-1", "u-2")
	r.RemoveFavorite("lst-1", "u-2")
}

// === Labour Repository ===
func TestLabourRepoFull(t *testing.T) {
	r := memory.NewLabourOrderRepository()
	lo := domain.LabourOrder{ID: "lo-1", EmployerID: "u-1", Title: "用工需求", Status: "open"}
	r.Create(lo)
	r.FindByID("lo-1")
	r.ListByEmployer("u-1")
	r.ListAll(0, 20)
	// Quote
	r.CreateQuote(domain.LabourQuote{ID: "q-1", OrderID: "lo-1", QuoterID: "u-2", AmountFen: 10000})
	r.ListQuotes("lo-1")
	// Assignment
	r.CreateAssignment(domain.Assignment{ID: "asgn-1", OrderID: "lo-1", WorkerID: "u-3"})
}

// === User + RefreshToken Repository ===
func TestUserRepoFull(t *testing.T) {
	r := memory.NewUserRepository(nil)
	u := domain.User{ID: "usr-1", WechatOpenID: "wx-openid-1", Status: "active", Role: domain.RoleIndividual}
	r.Create(u)
	r.FindByOpenID("wx-openid-1")
	r.FindByID("usr-1")
	r.All()
	r.UpdateRole("usr-1", domain.RoleEnterprise)
}

func TestRefreshTokenRepoFull(t *testing.T) {
	r := memory.NewRefreshTokenRepository()
	r.Store("usr-1", "hash-abc", time.Now().Add(time.Hour))
	uid, _, revoked, err := r.Find("hash-abc")
	if err != nil || uid != "usr-1" || revoked {
		t.Fatal("find failed")
	}
	r.Revoke("hash-abc")
	_, _, revoked2, _ := r.Find("hash-abc")
	if !revoked2 {
		t.Fatal("should be revoked")
	}
}

// === Training Repos ===
func TestTrainingReposFull(t *testing.T) {
	certR := memory.NewCertificateRepository()
	certR.Create(domain.Certificate{ID: "cert-1", UserID: "u-1", CertType: domain.CertCAAC, Status: "pending"})
	certR.FindByID("cert-1")
	certR.ListByUser("u-1")
	certR.UpdateStatus("cert-1", "approved")
	certR.ListAll()

	crR := memory.NewCourseRepository()
	crR.Create(domain.TrainingCourse{ID: "crs-1", OrgID: "org-1", Title: "CAAC培训"})
	crR.List()

	instR := memory.NewInstructorRepository()
	instR.Create(domain.Instructor{ID: "inst-1", UserID: "u-1", Name: "教练", Status: "pending"})
	instR.FindByID("inst-1")
	instR.List()
	instR.UpdateStatus("inst-1", "approved")

	pilotR := memory.NewPilotRepository(nil)
	pilotR.Create(domain.CertifiedPilot{ID: "pilot-1", UserID: "u-1", RealName: "飞手", Status: "pending"})
	pilotR.FindByID("pilot-1")
	pilotR.List()
	pilotR.UpdateStatus("pilot-1", "approved")
}

// === Trading Repos ===
func TestTradingReposFull(t *testing.T) {
	pr := memory.NewProductRepository()
	pr.Create(domain.DroneProduct{ID: "prod-1", SellerID: "u-1", ProdType: domain.ProductDrone, Title: "M300", Status: "listed"})
	pr.List("")
	pr.List("drone")

	rr := memory.NewRepairRepository()
	rr.Create(domain.RepairOrder{ID: "rep-1", CustomerID: "u-1", Status: "submitted"})
	rr.ListByUser("u-1")
}

// === Insurance + Finance Repos ===
func TestInsuranceFinanceRepos(t *testing.T) {
	polR := memory.NewPolicyRepository()
	polR.Create(domain.InsurancePolicy{ID: "pol-1", UserID: "u-1", Status: "active"})
	polR.ListByUser("u-1")

	inspR := memory.NewInspectionRepository()
	inspR.Create(domain.AnnualInspection{ID: "insp-1", UserID: "u-1", Status: "pending"})
	inspR.ListByUser("u-1")
	inspR.ListAll()

	loanR := memory.NewLoanRepository()
	loanR.Create(domain.LoanApplication{ID: "loan-1", UserID: "u-1", AmountFen: 100000, Status: "submitted"})
	loanR.ListByUser("u-1")
}

// === Content Repos (Message, Article, Review) ===
func TestContentReposFull(t *testing.T) {
	msgR := memory.NewMessageRepository()
	msgR.Create(domain.Message{ID: "msg-1", ReceiverID: "u-1", SenderID: "sys", IsRead: false})
	msgR.ListByUser("u-1", false)
	msgR.ListByUser("u-1", true)
	msgR.MarkRead("msg-1")
	cnt, _ := msgR.UnreadCount("u-1")
	if cnt != 0 {
		t.Fatalf("unread: %d", cnt)
	}

	artR := memory.NewArticleRepository()
	artR.Create(domain.Article{ID: "art-1", Title: "新闻", Category: "policy", Status: "draft"})
	artR.FindByID("art-1")
	artR.ListByCategory("", 0, 20)
	artR.Update(domain.Article{ID: "art-1", Title: "updated", Status: "published"})

	revR := memory.NewReviewRepository()
	revR.Create(domain.Review{ID: "rev-1", ReviewerID: "u-1", TargetType: "enterprise", TargetID: "ent-1", Rating: 5, Status: "pending"})
	revR.ListByTarget("enterprise", "ent-1")
	revR.ListAll("", 0, 20)
	revR.UpdateStatus("rev-1", "approved")
	revR.Delete("rev-1")
}

// === Venue + Enrollment + TradeOrder + Escrow ===
func TestVenueEnrollTradeEscrowRepos(t *testing.T) {
	vR := memory.NewVenueRepository()
	vR.Create(domain.Venue{ID: "v-1", OwnerID: "u-1", Name: "飞行基地", Status: "available"})
	vR.List("")
	vR.FindByID("v-1")
	vR.CreateBooking(domain.VenueBooking{ID: "bk-1", VenueID: "v-1", UserID: "u-2", Status: "booked"})
	vR.ListBookings("v-1")

	eR := memory.NewEnrollmentRepository()
	eR.Create(domain.Enrollment{ID: "enr-1", CourseID: "crs-1", UserID: "u-1"})
	eR.ListByCourse("crs-1")
	eR.FindByUserAndCourse("u-1", "crs-1")

	tR := memory.NewTradeOrderRepository()
	tR.Create(domain.TradeOrder{ID: "to-1", ProductID: "prod-1", BuyerID: "u-1", SellerID: "u-2", Status: "pending"})
	tR.FindByID("to-1")
	tR.UpdateStatus("to-1", "paid")
	tR.ListByUser("u-1")

	escR := memory.NewEscrowRepository()
	escR.GetAccount("u-1")
	// C6 接口变更：资金操作原子化，Deposit 即开户
	escR.Deposit("u-1", 100000, domain.EscrowTransaction{ID: "tx-1", FromUser: "sys", ToUser: "u-1", AmountFen: 100000, TxType: "deposit"})
	escR.ListTransactions("u-1")
}

// === New Biz Repos ===
func TestNewBizReposFull(t *testing.T) {
	expR := memory.NewExpertRepository()
	expR.Create(domain.Expert{ID: "exp-1", Name: "专家", Field: "无人机"})
	expR.FindByID("exp-1")
	expR.List("")
	expR.Update(domain.Expert{ID: "exp-1", Name: "updated", Field: "低空"})
	expR.Delete("exp-1")

	caseR := memory.NewCaseRepository()
	caseR.Create(domain.CaseEntry{ID: "case-1", Title: "案例", Category: "logistics"})
	caseR.FindByID("case-1")
	caseR.List("", 0, 20)
	caseR.Update(domain.CaseEntry{ID: "case-1", Title: "updated"})
	caseR.Delete("case-1")

	compR := memory.NewComplianceRepository()
	compR.CreateDoc(domain.ComplianceDoc{ID: "cd-1", Title: "条例"})
	compR.FindDocByID("cd-1")
	compR.ListDocs("", 0, 20)
	compR.UpdateDoc(domain.ComplianceDoc{ID: "cd-1", Title: "updated"})
	compR.DeleteDoc("cd-1")
	compR.CreateStandard(domain.StandardDoc{ID: "std-1", Title: "标准"})
	compR.ListStandards("", 0, 20)

	achR := memory.NewAchievementRepository()
	achR.Create(domain.Achievement{ID: "ach-1", OwnerID: "u-1", Title: "专利"})
	achR.FindByID("ach-1")
	achR.List("", 0, 20)
	achR.Update(domain.Achievement{ID: "ach-1", Title: "updated"})
	achR.Delete("ach-1")

	rdR := memory.NewRDChallengeRepository()
	rdR.Create(domain.RDChallenge{ID: "rd-1", PosterID: "u-1", Title: "难题"})
	rdR.FindByID("rd-1")
	rdR.List("", 0, 20)
	rdR.Update(domain.RDChallenge{ID: "rd-1", Title: "updated"})

	rpR := memory.NewResearchProjectRepository()
	rpR.Create(domain.ResearchProject{ID: "rp-1", Title: "课题"})
	rpR.FindByID("rp-1")
	rpR.List(0, 20)
	rpR.Update(domain.ResearchProject{ID: "rp-1", Title: "updated"})

	paR := memory.NewProjectAppRepository()
	paR.Create(domain.ProjectApplication{ID: "pa-1", ApplicantID: "u-1", ProjectName: "申报"})
	paR.FindByID("pa-1")
	paR.ListByUser("u-1")
	paR.ListAll("", 0, 20)
	paR.Update(domain.ProjectApplication{ID: "pa-1", Status: "submitted"})

	compR2 := memory.NewCompetitionRepository(nil)
	compR2.Create(domain.Competition{ID: "comp-1", Title: "比赛"})
	compR2.FindByID("comp-1")
	compR2.List(0, 20)
	compR2.Update(domain.Competition{ID: "comp-1", Title: "updated"})
	compR2.CreateReg(domain.CompetitionReg{ID: "reg-1", CompetitionID: "comp-1", UserID: "u-1"})
	compR2.ListRegs("comp-1")

	evR := memory.NewEventRepository()
	evR.Create(domain.AssociationEvent{ID: "evt-1", Title: "活动"})
	evR.FindByID("evt-1")
	evR.List(0, 20)
	evR.Update(domain.AssociationEvent{ID: "evt-1", Title: "updated"})
	evR.CreateReg(domain.EventRegistration{ID: "erg-1", EventID: "evt-1", UserID: "u-1"})
	evR.ListRegs("evt-1")

	portR := memory.NewPortfolioRepository()
	portR.Create(domain.MemberPortfolio{ID: "port-1", EnterpriseID: "ent-1", Name: "品牌"})
	portR.FindByID("port-1")
	portR.ListByEnterprise("ent-1")
	portR.ListPublished(0, 20)
	portR.Update(domain.MemberPortfolio{ID: "port-1", Name: "updated"})

	irR := memory.NewIndustryReportRepository()
	irR.Create(domain.IndustryReport{ID: "ir-1", Title: "报告"})
	irR.FindByID("ir-1")
	irR.List(0, 20)
	irR.Update(domain.IndustryReport{ID: "ir-1", Title: "updated"})
	irR.Delete("ir-1")

	resR := memory.NewResourceRepository()
	resR.Create(domain.IndustryResource{ID: "res-1", OwnerID: "u-1", Name: "无人机"})
	resR.FindByID("res-1")
	resR.List("", 0, 20)
	resR.Update(domain.IndustryResource{ID: "res-1", Name: "updated"})

	emR := memory.NewEmergencyRepository()
	emR.CreateResource(domain.EmergencyResource{ID: "er-1", OwnerID: "u-1", Name: "应急机"})
	emR.FindResourceByID("er-1")
	emR.ListResources("", "", 0, 20)
	emR.UpdateResource(domain.EmergencyResource{ID: "er-1", Name: "updated"})
	emR.CreateDispatch(domain.EmergencyDispatch{ID: "ed-1", ResourceID: "er-1"})
	emR.ListDispatches(0, 20)
}
