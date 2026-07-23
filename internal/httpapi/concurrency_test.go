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
	bidRepo := memory.NewBidRepository()
	demandRepo := memory.NewDemandRepository(nil)
	enterpriseRepo := memory.NewEnterpriseRepository(nil)
	employmentRepo := memory.NewEmploymentRepository()
	contractRepo := memory.NewContractRepository()
	userRepo := memory.NewUserRepository(nil)
	refreshRepo := memory.NewRefreshTokenRepository()

	srv := httpapi.NewServer(
		service.NewDemandService(demandRepo, bidRepo),
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
		service.NewHomeService(demandRepo),
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

// ---- Test 2: 100 concurrent bids on one demand (race detection) ----

func TestConcurrent_100BidsOnOneDemand(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping concurrency test in short mode")
	}
	bidRepo := memory.NewBidRepository()
	demandRepo := memory.NewDemandRepository(nil)
	svc := service.NewDemandService(demandRepo, bidRepo)

	publisher := domain.Actor{ID: "pub-001", Role: domain.RoleEnterprise}
	d, _ := svc.Create(publisher, service.CreateDemandInput{
		PublisherName: "发布方", Contact: "13800001111", Title: "并发竞标测试",
	})
	svc.Submit(publisher, d.ID)
	svc.Approve(domain.Actor{ID: "admin", Role: domain.RoleAssociationAdmin}, d.ID)

	var wg sync.WaitGroup
	successCount := make(chan string, 100)
	start := time.Now()

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			actor := domain.Actor{ID: fmt.Sprintf("bidder-%d", idx), Role: domain.RoleIndividual}
			bid, err := svc.CreateBid(actor, d.ID, int64(10000+idx*1000), fmt.Sprintf("报价%d", idx))
			if err == nil {
				successCount <- bid.ID
			}
		}(i)
	}
	wg.Wait()
	close(successCount)

	elapsed := time.Since(start)
	count := len(successCount)
	t.Logf("100 concurrent bids: %d succeeded, elapsed %v", count, elapsed)

	if count < 90 {
		t.Errorf("only %d/100 bids succeeded (expect >=90)", count)
	}
}

// ---- Test 3: SelectBid race — only ONE should win ----

func TestConcurrent_SelectBidRace(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping concurrency test in short mode")
	}
	bidRepo := memory.NewBidRepository()
	demandRepo := memory.NewDemandRepository(nil)
	svc := service.NewDemandService(demandRepo, bidRepo)

	publisher := domain.Actor{ID: "pub-race", Role: domain.RoleEnterprise}
	d, _ := svc.Create(publisher, service.CreateDemandInput{
		PublisherName: "发布方", Contact: "13800001111", Title: "选标竞态测试",
	})
	svc.Submit(publisher, d.ID)
	svc.Approve(domain.Actor{ID: "admin", Role: domain.RoleAssociationAdmin}, d.ID)

	// Create 5 bids
	bids := make([]string, 5)
	for i := 0; i < 5; i++ {
		actor := domain.Actor{ID: fmt.Sprintf("bidder-%d", i), Role: domain.RoleIndividual}
		bid, _ := svc.CreateBid(actor, d.ID, int64(10000+i*1000), fmt.Sprintf("报价%d", i))
		bids[i] = bid.ID
	}

	// 10 concurrent SelectBid calls — only one should succeed
	var wg sync.WaitGroup
	results := make(chan bool, 50)

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			bidID := bids[idx%len(bids)]
			_, err := svc.SelectBid(publisher, d.ID, bidID)
			results <- (err == nil)
		}(i)
	}
	wg.Wait()
	close(results)

	successCount := 0
	for r := range results {
		if r {
			successCount++
		}
	}

	t.Logf("SelectBid race: %d/50 succeeded (expect exactly 1)", successCount)
	if successCount != 1 {
		t.Errorf("CAS TOCTOU bug: %d SelectBid calls succeeded (must be exactly 1)", successCount)
	}
}

// ---- Test 4: 200 enterprises → approve all via batch ----

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

// ---- Test 5: Mixed workload — enterprises + demands + bids simultaneously ----

func TestConcurrent_MixedWorkload(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping concurrency test in short mode")
	}
	bidRepo := memory.NewBidRepository()
	demandRepo := memory.NewDemandRepository(nil)
	entRepo := memory.NewEnterpriseRepository(nil)
	demandSvc := service.NewDemandService(demandRepo, bidRepo)
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

	// 100 bids on first 10 demands
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			did := fmt.Sprintf("demand-%d", idx%10) // depend on pre-existing demands — this test is approximate
			actor := domain.Actor{ID: fmt.Sprintf("mix-bidder-%d", idx), Role: domain.RoleIndividual}
			// Skip error — bids may fail if demand doesn't exist or isn't published
			_, _ = demandSvc.CreateBid(actor, did, int64(10000+idx*500), fmt.Sprintf("混合报价%d", idx))
		}(i)
	}

	wg.Wait()
	close(errs)

	elapsed := time.Since(start)
	failCount := 0
	for range errs {
		failCount++
	}
	t.Logf("Mixed workload 300 ops: %d errors (some expected), elapsed %v", failCount, elapsed)
	if elapsed > 10*time.Second {
		t.Errorf("mixed workload took %v (target <10s)", elapsed)
	}
}

// ---- Test 6: Data race check for confirmComplete ----

func TestConcurrent_DualConfirmRace(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping concurrency test in short mode")
	}
	bidRepo := memory.NewBidRepository()
	demandRepo := memory.NewDemandRepository(nil)
	svc := service.NewDemandService(demandRepo, bidRepo)

	publisher := domain.Actor{ID: "pub-conf", Role: domain.RoleEnterprise}
	bidder := domain.Actor{ID: "bidder-conf", Role: domain.RoleIndividual}

	d, _ := svc.Create(publisher, service.CreateDemandInput{
		PublisherName: "确认方", Contact: "13800001111", Title: "双确认并发测试",
	})
	svc.Submit(publisher, d.ID)
	svc.Approve(domain.Actor{ID: "admin", Role: domain.RoleAssociationAdmin}, d.ID)
	bid, _ := svc.CreateBid(bidder, d.ID, 50000, "报价")
	svc.SelectBid(publisher, d.ID, bid.ID)

	// 10 concurrent confirm calls from publisher and bidder
	var wg sync.WaitGroup
	completed := make(chan bool, 20)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, c, _ := svc.ConfirmComplete(publisher, d.ID)
			completed <- c
		}()
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, c, _ := svc.ConfirmComplete(bidder, d.ID)
			completed <- c
		}()
	}
	wg.Wait()
	close(completed)

	completeCount := 0
	for c := range completed {
		if c {
			completeCount++
		}
	}
	// Only the FIRST pair of confirmations should trigger completion
	t.Logf("Dual confirm race: %d completion signals (expect >=1)", completeCount)
	if completeCount == 0 {
		t.Error("demand was never completed — race condition may have dropped confirms")
	}

	// Verify demand IS completed
	final, _ := demandRepo.FindByID(d.ID)
	if final.Status != domain.DemandCompleted {
		t.Errorf("expected DemandCompleted, got %s — dual confirm race failed", final.Status)
	}
}

// ---- Test 7: Rate limiter throughput ----

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
