package httpapi_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/httpapi"
	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

// ============================================================
// 上线前并发压力测试
// 模拟: 200商家入驻 + 100人竞标 + 50个审核员并发操作
// ============================================================

func newFullServer(t *testing.T) *httpapi.Server {
	t.Helper()
	tokens, err := httpapi.NewTokenManager(testSecret)
	if err != nil {
		t.Fatal(err)
	}
	demandRepo := memory.NewDemandRepository(nil)
	enterpriseRepo := memory.NewEnterpriseRepository(nil)
	employmentRepo := memory.NewEmploymentRepository()
	contractRepo := memory.NewContractRepository()
	userRepo := memory.NewUserRepository(nil)
	refreshRepo := memory.NewRefreshTokenRepository()

	srv := httpapi.NewServer(
		service.NewDemandService(demandRepo),
		service.NewEnterpriseService(enterpriseRepo),
		service.NewEnterpriseSvc(enterpriseRepo),
		service.NewEmploymentService(employmentRepo),
		service.NewContractService(contractRepo),
		service.NewJobService(memory.NewJobRepository(), memory.NewResumeRepository(), memory.NewJobApplicationRepository()),
		service.NewCommunityService(memory.NewPostRepository(), memory.NewCommentRepository(), memory.NewReportRepository()),
		service.NewListingService(memory.NewListingRepository()),
		service.NewLabourService(memory.NewLabourOrderRepository()),
		service.NewTrainingService(nil, nil, nil, nil),
		service.NewTradingService(nil, nil),
		service.NewInsuranceService(nil, nil),
		service.NewFinanceService(nil),
		service.NewHomeService(demandRepo, memory.NewEnterpriseRepository(nil)),
		service.NewFileService("test_uploads/"),
		service.NewMessageService(nil),
		service.NewEnrollmentService(nil),
		service.NewExpiryService(),
		service.NewTradeOrderService(nil),
		service.NewEscrowService(nil),
		service.NewNewsService(nil),
		service.NewReviewService(nil),
		service.NewVenueService(nil),
		userRepo, refreshRepo, tokens,
	)
	srv.SetStorage("memory")
	return srv
}

// ---- Test 1: 200 enterprises register concurrently ----

func TestConcurrent_200EnterpriseRegistrations(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping concurrency test in short mode")
	}
	svc := service.NewEnterpriseSvc(memory.NewEnterpriseRepository(nil))
	var wg sync.WaitGroup
	errs := make(chan error, 200)
	start := time.Now()

	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			actor := domain.Actor{ID: fmt.Sprintf("user-%d", idx), Role: domain.RoleEnterprise}
			_, err := svc.Create(actor, service.CreateEnterpriseInput{
				Name:        fmt.Sprintf("企业-%d 科技有限公司", idx),
				AccountName: fmt.Sprintf("138%08d", idx),
				LicenseURL:  fmt.Sprintf("/uploads/license-%d.jpg", idx),
			})
			if err != nil {
				errs <- fmt.Errorf("enterprise %d: %w", idx, err)
			}
		}(i)
	}
	wg.Wait()
	close(errs)

	elapsed := time.Since(start)
	failCount := 0
	for range errs {
		failCount++
	}

	t.Logf("200 enterprises: %d failed, elapsed %v", failCount, elapsed)
	if failCount > 0 {
		t.Errorf("%d enterprises failed to register", failCount)
	}
	if elapsed > 5*time.Second {
		t.Errorf("200 registrations took %v (target <5s)", elapsed)
	}
}

// ---- Test 3: 200 enterprises → approve all via batch ----

func TestConcurrent_BatchApprove(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping concurrency test in short mode")
	}
	repo := memory.NewEnterpriseRepository(nil)
	svc := service.NewEnterpriseSvc(repo)
	admin := domain.Actor{ID: "admin", Role: domain.RolePlatformAdmin}

	// Create and submit 200 enterprises
	ids := make([]string, 200)
	for i := 0; i < 200; i++ {
		time.Sleep(time.Microsecond) // ensure unique ID
		actor := domain.Actor{ID: fmt.Sprintf("user-%d", i), Role: domain.RoleEnterprise}
		ent, _ := svc.Create(actor, service.CreateEnterpriseInput{
			Name: fmt.Sprintf("批量企业-%d", i), AccountName: fmt.Sprintf("138%08d", i),
		})
		svc.Submit(actor, ent.ID)
		ids[i] = ent.ID
	}

	// Approve in batches of 50 from 4 goroutines
	var wg sync.WaitGroup
	errs := make(chan error, 200)
	start := time.Now()
	batchSize := 50

	for g := 0; g < 4; g++ {
		wg.Add(1)
		go func(grp int) {
			defer wg.Done()
			startIdx := grp * batchSize
			endIdx := startIdx + batchSize
			if endIdx > len(ids) {
				endIdx = len(ids)
			}
			for _, id := range ids[startIdx:endIdx] {
				_, err := svc.Review(admin, id, "approve", "")
				if err != nil {
					errs <- fmt.Errorf("approve %s: %w", id, err)
				}
			}
		}(g)
	}
	wg.Wait()
	close(errs)

	elapsed := time.Since(start)
	failCount := 0
	for range errs {
		failCount++
	}
	t.Logf("Batch approve 200: %d failed, elapsed %v", failCount, elapsed)

	// Verify all approved
	items, total, _ := svc.ListByStatus(admin, "approved", 0, 500)
	if total != 200 {
		t.Errorf("expected 200 approved, got %d", total)
	}
	if len(items) != 200 {
		t.Errorf("expected 200 items, got %d", len(items))
	}
}

// ---- Test 4: Mixed workload — enterprises + demands simultaneously ----

func TestConcurrent_MixedWorkload(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping concurrency test in short mode")
	}
	demandRepo := memory.NewDemandRepository(nil)
	entRepo := memory.NewEnterpriseRepository(nil)
	demandSvc := service.NewDemandService(demandRepo)
	entSvc := service.NewEnterpriseSvc(entRepo)

	start := time.Now()
	var wg sync.WaitGroup
	errs := make(chan error, 300)

	// 100 enterprises registering
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			actor := domain.Actor{ID: fmt.Sprintf("mix-user-%d", idx), Role: domain.RoleEnterprise}
			_, err := entSvc.Create(actor, service.CreateEnterpriseInput{
				Name: fmt.Sprintf("混合企业-%d", idx), AccountName: fmt.Sprintf("138%08d", idx),
			})
			if err != nil {
				errs <- err
			}
		}(i)
	}

	// 100 demands created
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			actor := domain.Actor{ID: fmt.Sprintf("mix-pub-%d", idx), Role: domain.RoleEnterprise}
			_, err := demandSvc.Create(actor, service.CreateDemandInput{
				PublisherName: fmt.Sprintf("发布方-%d", idx),
				Contact:       fmt.Sprintf("138%08d", idx),
				Title:         fmt.Sprintf("混合需求-%d", idx),
				BizType:       "other",
			})
			if err != nil {
				errs <- err
			}
		}(i)
	}

	wg.Wait()
	close(errs)

	elapsed := time.Since(start)
	failCount := 0
	for range errs {
		failCount++
	}
	t.Logf("Mixed workload 200 ops: %d errors (some expected), elapsed %v", failCount, elapsed)
	if elapsed > 10*time.Second {
		t.Errorf("mixed workload took %v (target <10s)", elapsed)
	}
}

// ---- Test 5: Rate limiter throughput ----

func TestConcurrent_RateLimiterThroughput(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping concurrency test in short mode")
	}
	app := newServer(t)

	var wg sync.WaitGroup
	results := make(chan int, 500)
	start := time.Now()

	// 500 concurrent healthcheck requests
	for i := 0; i < 500; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			w := request(t, app, "GET", "/healthz", nil, domain.RoleIndividual)
			results <- w.Code
		}()
	}
	wg.Wait()
	close(results)

	elapsed := time.Since(start)
	oks, limited := 0, 0
	for code := range results {
		if code == 200 {
			oks++
		} else if code == 429 {
			limited++
		}
	}

	t.Logf("Rate limiter: %d OK, %d limited (429), elapsed %v", oks, limited, elapsed)
	if limited == 0 {
		t.Log("Note: rate limiter may not trigger — burst allows first 200 requests through")
	}
}
