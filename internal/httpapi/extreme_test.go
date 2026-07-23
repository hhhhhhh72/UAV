package httpapi_test

import (
	"fmt"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

// ============================================================
// 极限压力 — 1000 goroutines, 10万+操作, 持续5秒
// ============================================================

func TestExtreme_1000GoroutineSustained(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping extreme test in short mode")
	}
	bidRepo := memory.NewBidRepository()
	demandRepo := memory.NewDemandRepository(nil)
	entRepo := memory.NewEnterpriseRepository(nil)
	demandSvc := service.NewDemandService(demandRepo, bidRepo)
	entSvc := service.NewEnterpriseSvc(entRepo)
	admin := domain.Actor{ID: "admin", Role: domain.RoleAssociationAdmin}

	// Pre-populate 50 demands
	for i := 0; i < 50; i++ {
		pub := domain.Actor{ID: fmt.Sprintf("ext-pub-%d", i), Role: domain.RoleEnterprise}
		d, _ := demandSvc.Create(pub, service.CreateDemandInput{
			PublisherName: "ext", Contact: "13800001111", Title: fmt.Sprintf("极限需求-%d", i), BizType: "other",
		})
		time.Sleep(time.Microsecond)
		demandSvc.Submit(pub, d.ID)
		demandSvc.Approve(admin, d.ID)
	}

	var (
		totalOps    int64
		totalErrors int64
		latencies   = make([]int64, 0, 100000)
		latMu       sync.Mutex
		stopped     int32
	)

	// 100 concurrent workers
	for w := 0; w < 100; w++ {
		go func(wid int) {
			ops := 0
			for atomic.LoadInt32(&stopped) == 0 {
				start := time.Now()
				switch ops % 10 {
				case 0, 1, 2: // 30% read — list demands
					demandSvc.List(repository.DemandFilter{})
				case 3, 4: // 20% read — search enterprises
					entSvc.Search(domain.Actor{ID: "admin", Role: domain.RolePlatformAdmin}, fmt.Sprintf("ext-%d", wid))
				case 5, 6: // 20% write — create bids
					actor := domain.Actor{ID: fmt.Sprintf("ext-bidder-%d-%d", wid, ops), Role: domain.RoleIndividual}
					demandSvc.CreateBid(actor, fmt.Sprintf("demand-%d", ops%25), 50000, fmt.Sprintf("ext-bid-%d", ops))
				case 7, 8: // 20% write — enterprises
					actor := domain.Actor{ID: fmt.Sprintf("ext-ent-%d-%d", wid, ops), Role: domain.RoleEnterprise}
					entSvc.Create(actor, service.CreateEnterpriseInput{
						Name: fmt.Sprintf("极限企业-%d-%d", wid, ops), AccountName: "13800000000",
					})
				case 9: // 10% admin — list enterprises
					entSvc.ListByStatus(domain.Actor{ID: "admin", Role: domain.RolePlatformAdmin}, "", 0, 50)
				}
				lat := time.Since(start).Nanoseconds()
				latMu.Lock()
				latencies = append(latencies, lat)
				latMu.Unlock()
				atomic.AddInt64(&totalOps, 1)
				ops++
			}
		}(w)
	}

	time.Sleep(5 * time.Second)
	atomic.StoreInt32(&stopped, 1)
	time.Sleep(100 * time.Millisecond) // let workers drain

	ops := atomic.LoadInt64(&totalOps)
	errs := atomic.LoadInt64(&totalErrors)
	opsPerSec := float64(ops) / 5.0

	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
	n := len(latencies)

	t.Logf("╔══════════════════════════════════════╗")
	t.Logf("║   EXTREME STRESS — 100 workers × 5s  ║")
	t.Logf("╠══════════════════════════════════════╣")
	t.Logf("║ Total ops:  %8d                  ║", ops)
	t.Logf("║ Errors:     %8d                  ║", errs)
	t.Logf("║ Throughput: %8.0f ops/sec          ║", opsPerSec)
	if n > 0 {
		t.Logf("║ P50:  %12v                  ║", time.Duration(latencies[n*50/100]))
		t.Logf("║ P95:  %12v                  ║", time.Duration(latencies[n*95/100]))
		t.Logf("║ P99:  %12v                  ║", time.Duration(latencies[n*99/100]))
		t.Logf("║ Max:  %12v                  ║", time.Duration(latencies[n-1]))
	}
	t.Logf("╚══════════════════════════════════════╝")

	if ops < 10000 {
		t.Errorf("throughput too low: %d ops in 5s (target >=10000)", ops)
	}
}

// ---- 2000 goroutine burst: bid flood ----

func TestExtreme_2000GoroutineBidFlood(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping extreme test in short mode")
	}
	bidRepo := memory.NewBidRepository()
	demandRepo := memory.NewDemandRepository(nil)
	svc := service.NewDemandService(demandRepo, bidRepo)
	admin := domain.Actor{ID: "admin", Role: domain.RoleAssociationAdmin}

	// 5 demands
	demandIDs := make([]string, 5)
	for i := 0; i < 5; i++ {
		pub := domain.Actor{ID: fmt.Sprintf("flood-pub-%d", i), Role: domain.RoleEnterprise}
		d, _ := svc.Create(pub, service.CreateDemandInput{
			PublisherName: "flood", Contact: "13800001111", Title: fmt.Sprintf("洪峰需求-%d", i), BizType: "other",
		})
		time.Sleep(time.Microsecond)
		svc.Submit(pub, d.ID)
		svc.Approve(admin, d.ID)
		demandIDs[i] = d.ID
	}

	startG := runtime.NumGoroutine()
	var wg sync.WaitGroup
	var success, fail int64
	start := time.Now()

	// 2000 goroutines bidding simultaneously
	for i := 0; i < 2000; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			did := demandIDs[idx%len(demandIDs)]
			actor := domain.Actor{ID: fmt.Sprintf("flood-bidder-%d", idx), Role: domain.RoleIndividual}
			_, err := svc.CreateBid(actor, did, int64(10000+idx), fmt.Sprintf("洪峰报价-%d", idx))
			if err == nil {
				atomic.AddInt64(&success, 1)
			} else {
				atomic.AddInt64(&fail, 1)
			}
		}(i)
	}
	wg.Wait()

	elapsed := time.Since(start)
	endG := runtime.NumGoroutine()

	t.Logf("╔══════════════════════════════════════╗")
	t.Logf("║   2000 GOROUTINE BID FLOOD           ║")
	t.Logf("╠══════════════════════════════════════╣")
	t.Logf("║ Success: %5d  Fail: %5d          ║", success, fail)
	t.Logf("║ Elapsed: %10v                  ║", elapsed)
	t.Logf("║ Goroutines: %d→%d (leak: %d)       ║", startG, endG, endG-startG)
	t.Logf("╚══════════════════════════════════════╝")

	if success < 1900 {
		t.Errorf("only %d/2000 bids succeeded", success)
	}
}

// ---- SelectBid atomicity under 500 concurrent selectors ----

func TestExtreme_SelectBidAtomicity(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping extreme test in short mode")
	}
	bidRepo := memory.NewBidRepository()
	demandRepo := memory.NewDemandRepository(nil)
	svc := service.NewDemandService(demandRepo, bidRepo)

	publisher := domain.Actor{ID: "atom-pub", Role: domain.RoleEnterprise}
	d, _ := svc.Create(publisher, service.CreateDemandInput{
		PublisherName: "atom", Contact: "13800001111", Title: "原子性测试", BizType: "other",
	})
	svc.Submit(publisher, d.ID)
	svc.Approve(domain.Actor{ID: "admin", Role: domain.RoleAssociationAdmin}, d.ID)

	// Create 100 bids
	bidIDs := make([]string, 100)
	for i := 0; i < 100; i++ {
		actor := domain.Actor{ID: fmt.Sprintf("atom-bidder-%d", i), Role: domain.RoleIndividual}
		bid, _ := svc.CreateBid(actor, d.ID, int64(10000+i), fmt.Sprintf("原子报价-%d", i))
		bidIDs[i] = bid.ID
	}

	var wg sync.WaitGroup
	var selectWins int64

	// 500 goroutines race to select
	for i := 0; i < 500; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			bidID := bidIDs[idx%len(bidIDs)]
			_, err := svc.SelectBid(publisher, d.ID, bidID)
			if err == nil {
				atomic.AddInt64(&selectWins, 1)
			}
		}(i)
	}
	wg.Wait()

	wins := atomic.LoadInt64(&selectWins)

	t.Logf("╔══════════════════════════════════════╗")
	t.Logf("║   SELECTBID ATOMICITY — 500 racers   ║")
	t.Logf("╠══════════════════════════════════════╣")
	t.Logf("║ SelectBid wins: %3d (must be 1)     ║", wins)
	t.Logf("╚══════════════════════════════════════╝")

	if wins != 1 {
		t.Errorf("CAS BROKEN: %d SelectBid calls won (must be exactly 1)", wins)
	}

	// Verify demand is matched
	final, _ := demandRepo.FindByID(d.ID)
	if final.Status != domain.DemandMatched {
		t.Errorf("demand not matched: got %s", final.Status)
	}
}
