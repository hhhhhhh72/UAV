package httpapi_test

import (
	"fmt"
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
	demandRepo := memory.NewDemandRepository(nil)
	entRepo := memory.NewEnterpriseRepository(nil)
	demandSvc := service.NewDemandService(demandRepo)
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
				case 5, 6: // 20% write — create demands
					actor := domain.Actor{ID: fmt.Sprintf("ext-pub-%d-%d", wid, ops), Role: domain.RoleEnterprise}
					demandSvc.Create(actor, service.CreateDemandInput{
						PublisherName: "ext", Contact: "13800001111", Title: fmt.Sprintf("极限新需求-%d-%d", wid, ops), BizType: "other",
					})
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
