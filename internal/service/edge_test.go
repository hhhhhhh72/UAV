package service_test

import (
	"testing"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

// === Community ===
func TestCommunityAll(t *testing.T) {
	svc := service.NewCommunityService(memory.NewPostRepository(), memory.NewCommentRepository(), memory.NewReportRepository())
	// CreatePost
	p, _ := svc.CreatePost(entActor(), "帖子标题", "内容", nil)
	if p.Status != "published" { t.Fatal("post should be published") }
	// PublishPost
	svc.PublishPost(admActor(), p.ID)
	// RemovePost
	svc.RemovePost(entActor(), p.ID)
	// ListPublishedPosts
	svc.ListPublishedPosts(0, 20)
	// CreateComment
	cmt, _ := svc.CreateComment(entActor(), p.ID, "评论")
	_ = cmt
	// ListComments
	svc.ListComments(p.ID)
	// CreateReport
	svc.CreateReport(entActor(), "post", p.ID, "违规")
	// ListPendingReports
	svc.ListPendingReports(admActor(), 0, 20)
}

// === Jobs ===
func TestJobServiceAll(t *testing.T) {
	svc := service.NewJobService(memory.NewJobRepository(), memory.NewResumeRepository(), memory.NewJobApplicationRepository())
	// CreateJob
	j, _ := svc.CreateJob(entActor(), "飞手招聘", "描述", "重庆", 100000)
	// PublishJob
	svc.PublishJob(entActor(), j.ID)
	// CloseJob
	svc.CloseJob(entActor(), j.ID)
	// ListPublishedJobs
	svc.ListPublishedJobs(0, 20)
	// ListMyJobs
	svc.ListMyJobs(entActor())
	// CreateResume
	r, _ := svc.CreateResume(indActor(), "我的简历", "经验...", "public")
	// UpdateResume
	svc.UpdateResume(indActor(), r.ID, "更新简历", "新经验...", "public")
	// ListMyResumes
	svc.ListMyResumes(indActor())
	// Apply
	app, _ := svc.Apply(indActor(), j.ID, r.ID)
	// UpdateApplicationStatus
	svc.UpdateApplicationStatus(entActor(), app.ID, domain.AppInterviewing)
	// ListApplicationsForJob
	svc.ListApplicationsForJob(entActor(), j.ID)
	// ListMyApplications
	svc.ListMyApplications(indActor())
}

// === Enterprise ===
func TestEnterpriseSvcAll(t *testing.T) {
	svc := service.NewEnterpriseSvc(memory.NewEnterpriseRepository(nil))
	e, _ := svc.Create(entActor(), service.CreateEnterpriseInput{Name: "企业"})
	// Update
	svc.Update(entActor(), e.ID, service.CreateEnterpriseInput{Name: "更新企业"})
	// FindByID
	svc.FindByID(e.ID)
	// ListMine
	svc.ListMine(entActor())
	// ListByStatus (admin)
	svc.ListByStatus(admActor(), "", 0, 20)
	// Search (admin)
	svc.Search(admActor(), "企业")
	// Submit
	svc.Submit(entActor(), e.ID)
}

// === Listings + Labour ===
func TestListingsLabourAll(t *testing.T) {
	ls := service.NewListingService(memory.NewListingRepository())
	l, _ := ls.Create(domain.Listing{ID: "lst-1", SellerID: entActor().ID, Title: "二手无人机", Description: "描述", Category: "drone", PriceFen: 50000, District: "渝北", Status: "listed"})
	ls.Close(entActor(), l.ID)
	ls.ListListed(0, 20)
	ls.Favorite(l.ID, indActor().ID)

	lbr := service.NewLabourService(memory.NewLabourOrderRepository())
	lo, _ := lbr.CreateOrder(entActor(), "用工", "描述", 5, time.Now(), time.Now().AddDate(0, 1, 0), 100000)
	lbr.ListOrders(entActor(), 0, 20)
	lbr.CreateQuote(indActor(), lo.ID, 80000, "报价", "飞手团队")
	lbr.ListQuotes(entActor(), lo.ID)
}

// === Insurance + Finance ===
func TestInsuranceFinanceAll(t *testing.T) {
	ins := service.NewInsuranceService(memory.NewPolicyRepository(), memory.NewInspectionRepository())
	ins.CreatePolicy(indActor(), "M300", "SN001", "liability", 5000, 500000, time.Now(), time.Now().AddDate(1, 0, 0))
	ins.ListMyPolicies(indActor())
	ins.CreateInspection(indActor(), "M300", "SN001", time.Now(), time.Now().AddDate(1, 0, 0))
	ins.ListAllInspections()
	ins.ListMyInspections(indActor())

	fin := service.NewFinanceService(memory.NewLoanRepository())
	fin.ApplyLoan(indActor(), 200000, 12, "采购无人机")
	fin.ListMyLoans(indActor())
}

// === Trading ===
func TestTradingAll(t *testing.T) {
	svc := service.NewTradingService(memory.NewProductRepository(), memory.NewRepairRepository())
	svc.CreateProduct(indActor(), domain.ProductDrone, "M300", "RTK版", "DJI", "M300", "new", 5000000)
	svc.ListProducts("")
	svc.CreateRepair(indActor(), "M300", "云台故障")
	svc.ListMyRepairs(indActor())
}

// === Phase3 (Enrollment + Expiry + TradeOrder) ===
func TestPhase3All(t *testing.T) {
	enr := service.NewEnrollmentService(memory.NewEnrollmentRepository())
	enr.Enroll("u-1", "crs-1")
	enr.ListByCourse("crs-1")

	exp := service.NewExpiryService()
	certs := []domain.Certificate{{ID: "c1", ExpireDate: time.Now().AddDate(0, 0, 10), Status: "approved"}}
	exp.GetExpiringCerts(certs, 30)
	inspections := []domain.AnnualInspection{{ID: "i1", ExpireDate: time.Now().AddDate(0, 0, 10), Status: "approved"}}
	exp.GetExpiringInspections(inspections, 30)

	to := service.NewTradeOrderService(memory.NewTradeOrderRepository())
	o, _ := to.Create("u-1", "prod-1", "u-2", 100000)
	to.UpdateStatus(o.ID, "u-1", "paid")
	to.ListMine("u-1")
}

// === Reviews + Venues ===
func TestReviewsVenuesAll(t *testing.T) {
	rv := service.NewReviewService(memory.NewReviewRepository())
	rv.Submit("u-1", "enterprise", "ent-1", 4, "不错")
	rv.ListByTarget("enterprise", "ent-1")
	rv.ListAll("", 0, 20)
	rv.Approve("rev-1")
	rv.Reject("rev-2")
	rv.Delete("rev-3")

	vn := service.NewVenueService(memory.NewVenueRepository())
	v, _ := vn.Create("u-1", "飞行基地", "training_field", "重庆", 50000)
	vn.List("")
	vn.Book(v.ID, "u-2", time.Now().Add(time.Hour), time.Now().Add(3*time.Hour))
}

// === Home ===
func TestHomeService(t *testing.T) {
	svc := service.NewHomeService(memory.NewDemandRepository(nil), memory.NewEnterpriseRepository(nil))
	data := svc.GetHome("重庆", 29.5, 106.5)
	if data.City != "重庆" { t.Fatal("city mismatch") }
	if len(data.Banners) == 0 { t.Log("no banners configured") }
}

// === Files ===
func TestFileService(t *testing.T) {
	svc := service.NewFileService("test_uploads/")
	_ = svc
	// Upload requires io.Reader, skip for now
}

// === Enterprise Service (legacy) ===
func TestEnterpriseServiceLegacy(t *testing.T) {
	svc := service.NewEnterpriseService(memory.NewEnterpriseRepository(nil))
	pending, _ := svc.Pending(admActor())
	_ = pending
	results, _ := svc.Search("测试")
	_ = results
}

// === Demand Search ===
func TestDemandServiceSearch(t *testing.T) {
	svc := service.NewDemandService(memory.NewDemandRepository(nil))
	results, _ := svc.Search("巡检")
	_ = results
}

// === EnterpriseService Search ===
func TestEnterpriseServiceSearch(t *testing.T) {
	svc := service.NewEnterpriseService(memory.NewEnterpriseRepository(nil))
	results, _ := svc.Search("企业")
	_ = results
}

// === Employment List ===
func TestEmploymentServiceList(t *testing.T) {
	svc := service.NewEmploymentService(memory.NewEmploymentRepository())
	svc.Create(entActor(), domain.EmploymentRequest{ID: "emp-1", EnterpriseID: entActor().ID, Position: "飞手", Headcount: 5})
	list, total, _ := svc.List(entActor(), 0, 20)
	_ = list
	_ = total
}

// === Contract List ===
func TestContractServiceList(t *testing.T) {
	svc := service.NewContractService(memory.NewContractRepository())
	svc.Create(entActor(), domain.Contract{ID: "ctr-1", EnterpriseID: entActor().ID, TemplateID: "tpl-1"})
	list, total, _ := svc.List(entActor(), 0, 20)
	_ = list
	_ = total
}

// === Training ListAllCertificates ===
func TestTrainingListAllCertificates(t *testing.T) {
	svc := service.NewTrainingService(
		memory.NewCertificateRepository(), memory.NewCourseRepository(),
		memory.NewInstructorRepository(), memory.NewPilotRepository(nil),
	)
	svc.AddCertificate(indActor(), domain.CertCAAC, "CAAC-001", "III", "民航局", time.Now(), time.Now().AddDate(2, 0, 0))
	certs, _ := svc.ListAllCertificates()
	_ = certs
}
