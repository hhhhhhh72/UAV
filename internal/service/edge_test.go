package service_test

import (
	"context"
	"testing"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

// === Community ===
func TestCommunityAll(t *testing.T) {
	svc := service.NewCommunityService(memory.NewPostRepository(), memory.NewCommentRepository(), memory.NewReportRepository())
	// CreatePost — 默认待审核（pending），管理端 PublishPost 后公开
	p, _ := svc.CreatePost(context.Background(), entActor(), "帖子标题", "内容", nil)
	if p.Status != "pending" {
		t.Fatal("post should be pending until admin publishes")
	}
	// PublishPost
	svc.PublishPost(context.Background(), admActor(), p.ID)
	// RemovePost
	svc.RemovePost(context.Background(), entActor(), p.ID)
	// ListPublishedPosts
	svc.ListPublishedPosts(context.Background(), 0, 20)
	// CreateComment
	cmt, _ := svc.CreateComment(context.Background(), entActor(), p.ID, "评论")
	_ = cmt
	// ListComments
	svc.ListComments(context.Background(), p.ID)
	// CreateReport
	svc.CreateReport(context.Background(), entActor(), "post", p.ID, "违规")
	// ListPendingReports
	svc.ListPendingReports(context.Background(), admActor(), 0, 20)
}

// === Jobs ===
func TestJobServiceAll(t *testing.T) {
	svc := service.NewJobService(memory.NewJobRepository(), memory.NewResumeRepository(), memory.NewJobApplicationRepository())
	// CreateJob
	j, _ := svc.CreateJob(context.Background(), entActor(), "飞手招聘", "描述", "重庆", 100000)
	// PublishJob
	svc.PublishJob(context.Background(), entActor(), j.ID)
	// ListPublishedJobs
	svc.ListPublishedJobs(context.Background(), 0, 20)
	// ListMyJobs
	svc.ListMyJobs(context.Background(), entActor())
	// CreateResume
	r, _ := svc.CreateResume(context.Background(), indActor(), "我的简历", "张三", "13800000000", "a@b.com", "本科", "经验...", []string{"巡检"}, "", "经验...", "public")
	// UpdateResume
	svc.UpdateResume(context.Background(), indActor(), r.ID, "更新简历", "张三", "13800000000", "a@b.com", "本科", "新经验...", []string{"巡检"}, "", "新经验...", "public")
	// ListMyResumes
	svc.ListMyResumes(context.Background(), indActor())
	// Apply（招聘中状态方可投递）
	app, err := svc.Apply(context.Background(), indActor(), j.ID, r.ID)
	if err != nil {
		t.Fatalf("apply to published job: %v", err)
	}
	// CloseJob
	svc.CloseJob(context.Background(), entActor(), j.ID)
	// UpdateApplicationStatus
	svc.UpdateApplicationStatus(context.Background(), entActor(), app.ID, domain.AppInterviewing)
	// ListApplicationsForJob
	svc.ListApplicationsForJob(context.Background(), entActor(), j.ID)
	// ListMyApplications
	svc.ListMyApplications(context.Background(), indActor())
}

// === Enterprise ===
func TestEnterpriseSvcAll(t *testing.T) {
	svc := service.NewEnterpriseSvc(memory.NewEnterpriseRepository(nil), memory.NewUserRepository(nil))
	e, _ := svc.Create(context.Background(), entActor(), service.CreateEnterpriseInput{Name: "企业"})
	// Update
	svc.Update(context.Background(), entActor(), e.ID, service.CreateEnterpriseInput{Name: "更新企业"})
	// FindByID
	svc.FindByID(context.Background(), e.ID)
	// ListMine
	svc.ListMine(context.Background(), entActor())
	// ListByStatus (admin)
	svc.ListByStatus(context.Background(), admActor(), "", 0, 20)
	// Search (admin)
	svc.Search(context.Background(), admActor(), "企业")
	// Submit
	svc.Submit(context.Background(), entActor(), e.ID)
}

// === Listings + Labour ===
func TestListingsLabourAll(t *testing.T) {
	ls := service.NewListingService(memory.NewListingRepository())
	l, _ := ls.Create(context.Background(), domain.Listing{ID: "lst-1", SellerID: entActor().ID, Title: "二手无人机", Description: "描述", Category: "drone", PriceFen: 50000, District: "渝北", Status: "listed"})
	ls.Close(context.Background(), entActor(), l.ID)
	ls.ListListed(context.Background(), 0, 20)
	ls.Favorite(context.Background(), l.ID, indActor().ID)

	lbr := service.NewLabourService(memory.NewLabourOrderRepository())
	lo, _ := lbr.CreateOrder(context.Background(), entActor(), "用工", "描述", 5, time.Now(), time.Now().AddDate(0, 1, 0), 100000)
	lbr.ListOrders(context.Background(), entActor(), 0, 20)
	lbr.CreateQuote(context.Background(), indActor(), lo.ID, 80000, "报价", "飞手团队")
	lbr.ListQuotes(context.Background(), entActor(), lo.ID)
}

// === Insurance + Finance ===
func TestInsuranceFinanceAll(t *testing.T) {
	ins := service.NewInsuranceService(memory.NewPolicyRepository(), memory.NewInspectionRepository())
	ins.CreatePolicy(context.Background(), indActor(), "M300", "SN001", "liability", 5000, 500000, time.Now(), time.Now().AddDate(1, 0, 0))
	ins.ListMyPolicies(context.Background(), indActor())
	ins.CreateInspection(context.Background(), indActor(), "M300", "SN001", time.Now(), time.Now().AddDate(1, 0, 0))
	ins.ListAllInspections(context.Background())
	ins.ListMyInspections(context.Background(), indActor())

	fin := service.NewFinanceService(memory.NewLoanRepository())
	fin.ApplyLoan(context.Background(), indActor(), 200000, 12, "采购无人机")
	fin.ListMyLoans(context.Background(), indActor())
}

// === Trading ===
func TestTradingAll(t *testing.T) {
	svc := service.NewTradingService(memory.NewProductRepository(), memory.NewRepairRepository())
	svc.CreateProduct(context.Background(), indActor(), domain.ProductDrone, "M300", "RTK版", "DJI", "M300", "new", 5000000, nil)
	svc.ListProducts(context.Background(), "")
	svc.CreateRepair(context.Background(), indActor(), "M300", "云台故障")
	svc.ListMyRepairs(context.Background(), indActor())
}

// === Phase3 (Enrollment + Expiry + TradeOrder) ===
func TestPhase3All(t *testing.T) {
	enr := service.NewEnrollmentService(memory.NewEnrollmentRepository(), memory.NewCourseRepository())
	enr.Enroll(context.Background(), "u-1", "crs-1", service.EnrollmentForm{Name: "张三", Phone: "13800000000"})
	enr.ListByCourse(context.Background(), "crs-1")

	exp := service.NewExpiryService()
	certs := []domain.Certificate{{ID: "c1", ExpireDate: time.Now().AddDate(0, 0, 10), Status: "approved"}}
	exp.GetExpiringCerts(certs, 30)
	inspections := []domain.AnnualInspection{{ID: "i1", ExpireDate: time.Now().AddDate(0, 0, 10), Status: "approved"}}
	exp.GetExpiringInspections(inspections, 30)

	to := service.NewTradeOrderService(memory.NewTradeOrderRepository(), memory.NewProductRepository())
	o, _ := to.Create(context.Background(), "u-1", "prod-1", "u-2", 100000)
	// paid 仅管理端可设（UpdateStatusAdmin）；买家直接改 paid 应被拒
	if _, err := to.UpdateStatus(context.Background(), o.ID, "u-1", "paid"); err == nil {
		t.Fatal("buyer must not mark order paid")
	}
	to.UpdateStatusAdmin(context.Background(), o.ID, "paid")
	to.ListMine(context.Background(), "u-1")
}

// === Reviews + Venues ===
func TestReviewsVenuesAll(t *testing.T) {
	rv := service.NewReviewService(memory.NewReviewRepository())
	rv.Submit(context.Background(), "u-1", "enterprise", "ent-1", 4, "不错")
	rv.ListByTarget(context.Background(), "enterprise", "ent-1")
	rv.ListAll(context.Background(), "", 0, 20)
	rv.Approve(context.Background(), "rev-1")
	rv.Reject(context.Background(), "rev-2")
	rv.Delete(context.Background(), "rev-3")

	vn := service.NewVenueService(memory.NewVenueRepository())
	v, _ := vn.Create(context.Background(), "u-1", "飞行基地", "training_field", "重庆", 50000)
	vn.List(context.Background(), "")
	vn.Book(context.Background(), v.ID, "u-2", time.Now().Add(time.Hour), time.Now().Add(3*time.Hour))
}

// === Home ===
func TestHomeService(t *testing.T) {
	svc := service.NewHomeService(memory.NewDemandRepository(nil), memory.NewEnterpriseRepository(nil))
	data := svc.GetHome(context.Background(), "重庆", 29.5, 106.5)
	if data.City != "重庆" {
		t.Fatal("city mismatch")
	}
	if len(data.Banners) == 0 {
		t.Log("no banners configured")
	}
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
	pending, _ := svc.Pending(context.Background(), admActor())
	_ = pending
	results, _ := svc.Search(context.Background(), "测试")
	_ = results
}

// === Demand Search ===
func TestDemandServiceSearch(t *testing.T) {
	svc := service.NewDemandService(memory.NewDemandRepository(nil))
	results, _ := svc.Search(context.Background(), "巡检")
	_ = results
}

// === EnterpriseService Search ===
func TestEnterpriseServiceSearch(t *testing.T) {
	svc := service.NewEnterpriseService(memory.NewEnterpriseRepository(nil))
	results, _ := svc.Search(context.Background(), "企业")
	_ = results
}

// === Employment List ===
func TestEmploymentServiceList(t *testing.T) {
	svc := service.NewEmploymentService(memory.NewEmploymentRepository())
	svc.Create(context.Background(), entActor(), domain.EmploymentRequest{ID: "emp-1", EnterpriseID: entActor().ID, Position: "飞手", Headcount: 5})
	list, total, _ := svc.List(context.Background(), entActor(), 0, 20)
	_ = list
	_ = total
}

// === Contract List ===
func TestContractServiceList(t *testing.T) {
	svc := service.NewContractService(memory.NewContractRepository())
	svc.Create(context.Background(), entActor(), domain.Contract{ID: "ctr-1", EnterpriseID: entActor().ID, TemplateID: "tpl-1"})
	list, total, _ := svc.List(context.Background(), entActor(), 0, 20)
	_ = list
	_ = total
}

// === Training ListAllCertificates ===
func TestTrainingListAllCertificates(t *testing.T) {
	svc := service.NewTrainingService(
		memory.NewCertificateRepository(), memory.NewCourseRepository(),
		memory.NewInstructorRepository(), memory.NewPilotRepository(nil),
	)
	svc.AddCertificate(context.Background(), indActor(), domain.CertCAAC, "CAAC-001", "III", "民航局", time.Now(), time.Now().AddDate(2, 0, 0))
	certs, _ := svc.ListAllCertificates(context.Background())
	_ = certs
}
