package service_test

import (
	"context"
	"testing"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

// 金额负数校验回归：需求预算/商品价格/劳务报价/课程价格/服务供给价/
// 场地价格/预约场地价格/贷款金额 一律拒绝负数，防止负数金额进入业务数据
// （负数商品价此前会传导为负数订单金额）。
func TestNegativeAmountsRejected(t *testing.T) {
	actor := domain.Actor{ID: "u-1", Role: domain.RoleEnterprise}

	// 需求预算
	dSvc := service.NewDemandService(memory.NewDemandRepository(nil))
	if _, err := dSvc.Create(context.Background(), actor, service.CreateDemandInput{
		Title: "t", Contact: "138", BudgetFen: -1,
	}); err == nil {
		t.Fatal("negative demand budget accepted")
	}

	// 商品价格
	tSvc := service.NewTradingService(memory.NewProductRepository(), memory.NewRepairRepository())
	if _, err := tSvc.CreateProduct(context.Background(), actor, domain.ProductDrone, "t", "", "", "", "", -1, nil); err == nil {
		t.Fatal("negative product price accepted")
	}

	// 劳务订单预算 + 报价
	lSvc := service.NewLabourService(memory.NewLabourOrderRepository())
	if _, err := lSvc.CreateOrder(context.Background(), actor, "t", "d", 1, time.Now(), time.Now().Add(time.Hour), -1); err == nil {
		t.Fatal("negative labour budget accepted")
	}
	if _, err := lSvc.CreateQuote(context.Background(), actor, "wo-1", -1, "p", "n"); err == nil {
		t.Fatal("negative quote amount accepted")
	}

	// 课程价格
	trSvc := service.NewTrainingService(memory.NewCertificateRepository(), memory.NewCourseRepository(), memory.NewInstructorRepository(), memory.NewPilotRepository(nil))
	if _, err := trSvc.CreateCourse(context.Background(), actor, domain.TrainingCourse{Title: "t", PriceFen: -1}); err == nil {
		t.Fatal("negative course price accepted")
	}

	// 服务供给价格
	slSvc := service.NewServiceListingService(memory.NewServiceListingRepository())
	if _, err := slSvc.CreateListing(context.Background(), "p", "n", "t", "c", "d", "r", -1, "u", "", ""); err == nil {
		t.Fatal("negative service listing price accepted")
	}

	// 测试场地价格
	tsSvc := service.NewTestSiteService(memory.NewTestSiteRepository())
	if _, err := tsSvc.Create(context.Background(), "n", "t", "l", "r", "o", -1, nil, ""); err == nil {
		t.Fatal("negative test site price accepted")
	}

	// 预约场地价格
	vSvc := service.NewVenueService(memory.NewVenueRepository())
	if _, err := vSvc.Create(context.Background(), "o", "n", "t", "l", -1); err == nil {
		t.Fatal("negative venue price accepted")
	}

	// 贷款金额/期限
	fSvc := service.NewFinanceService(memory.NewLoanRepository())
	if _, err := fSvc.ApplyLoan(context.Background(), actor, -100, 12, "p"); err == nil {
		t.Fatal("negative loan amount accepted")
	}
	if _, err := fSvc.ApplyLoan(context.Background(), actor, 100, 0, "p"); err == nil {
		t.Fatal("zero loan term accepted")
	}
}
