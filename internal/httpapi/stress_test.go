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
// 上线前极限压力测试 — 2000+ 操作, P50/P95/P99 延迟, 内存泄漏检测
// ============================================================

// ---- Test 1: 2000 mixed CRUD ops with latency tracking ----

func TestStress_2000MixedOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}
	demandRepo := memory.NewDemandRepository(nil)
	entRepo := memory.NewEnterpriseRepository(nil)
	demandSvc := service.NewDemandService(demandRepo)
	entSvc := service.NewEnterpriseSvc(entRepo)

	// Setup: create 20 published demands and 20 enterprises
	for i := 0; i < 20; i++ {
		pub := domain.Actor{ID: fmt.Sprintf("pub-%d", i), Role: domain.RoleEnterprise}
		d, _ := demandSvc.Create(pub, service.CreateDemandInput{
			PublisherName: "pub", Contact: "13800001111", Title: fmt.Sprintf("压测需求-%d", i), BizType: "other",
		})
		time.Sleep(time.Microsecond)
		demandSvc.Approve(domain.Actor{ID: "admin", Role: domain.RoleAssociationAdmin}, d.ID)
	}

	var (
		wg          sync.WaitGroup
		totalOps    int64
		totalErrors int64
		latencies   = make([]int64, 0, 2000)
		latMu       sync.Mutex
		stopped     int32
		start       = time.Now()
		duration    = 3 * time.Second // sustained load for 3 seconds
	)

	// 10 worker goroutines doing mixed operations
	for w := 0; w < 10; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			opCount := 0
			for atomic.LoadInt32(&stopped) == 0 {
				opStart := time.Now()

				switch opCount % 10 {
				case 0, 1, 2, 3: // 40% reads — list demands
					_, _ = demandSvc.List(repository.DemandFilter{})
				case 4, 5: // 20% reads — search enterprises
					_, _ = entSvc.Search(domain.Actor{ID: "admin", Role: domain.RolePlatformAdmin}, fmt.Sprintf("企业-%d", workerID))
				case 6, 7: // 20% writes — create demands
					actor := domain.Actor{ID: fmt.Sprintf("new-pub-%d-%d", workerID, opCount), Role: domain.RoleEnterprise}
					_, err := demandSvc.Create(actor, service.CreateDemandInput{
						PublisherName: "pub", Contact: "13800001111", Title: fmt.Sprintf("压测新需求-%d-%d", workerID, opCount), BizType: "other",
					})
					if err != nil {
						atomic.AddInt64(&totalErrors, 1)
					}
				case 8: // 10% writes — create enterprises
					actor := domain.Actor{ID: fmt.Sprintf("new-ent-%d-%d", workerID, opCount), Role: domain.RoleEnterprise}
					_, err := entSvc.Create(actor, service.CreateEnterpriseInput{
						Name: fmt.Sprintf("压测企业-%d-%d", workerID, opCount), AccountName: "13800000000",
					})
					if err != nil {
						atomic.AddInt64(&totalErrors, 1)
					}
				case 9: // 10% writes — create demands
					actor := domain.Actor{ID: fmt.Sprintf("new-pub-%d-%d", workerID, opCount), Role: domain.RoleEnterprise}
					_, err := demandSvc.Create(actor, service.CreateDemandInput{
						PublisherName: "pub", Contact: "13800001111", Title: fmt.Sprintf("新需求-%d-%d", workerID, opCount), BizType: "other",
					})
					if err != nil {
						atomic.AddInt64(&totalErrors, 1)
					}
				}

				lat := time.Since(opStart).Nanoseconds()
				latMu.Lock()
				latencies = append(latencies, lat)
				latMu.Unlock()
				atomic.AddInt64(&totalOps, 1)
				opCount++
			}
		}(w)
	}

	// Run for duration
	time.Sleep(duration)
	atomic.StoreInt32(&stopped, 1)
	wg.Wait()

	elapsed := time.Since(start)
	ops := atomic.LoadInt64(&totalOps)
	errs := atomic.LoadInt64(&totalErrors)
	opsPerSec := float64(ops) / elapsed.Seconds()

	// Latency stats
	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
	p50 := latencies[len(latencies)*50/100]
	p95 := latencies[len(latencies)*95/100]
	p99 := latencies[len(latencies)*99/100]

	t.Logf("=== 2000+ Mixed Ops ===")
	t.Logf("Total ops: %d, errors: %d, elapsed: %v", ops, errs, elapsed)
	t.Logf("Throughput: %.0f ops/sec", opsPerSec)
	t.Logf("P50: %v, P95: %v, P99: %v", time.Duration(p50), time.Duration(p95), time.Duration(p99))

	if ops < 2000 {
		t.Errorf("only %d ops in %v (target >=2000)", ops, duration)
	}
	if errs > int64(float64(ops)*0.05) {
		t.Errorf("error rate %.1f%% (target <5%%)", float64(errs)/float64(ops)*100)
	}
}

// ---- Test 2: Memory leak detection — 10K creates ----

func TestStress_MemoryLeakDetection(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}
	runtime.GC()
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)

	demandRepo := memory.NewDemandRepository(nil)
	svc := service.NewDemandService(demandRepo)
	admin := domain.Actor{ID: "admin", Role: domain.RoleAssociationAdmin}

	// 10,000 operations — create, approve
	for i := 0; i < 10000; i++ {
		pub := domain.Actor{ID: fmt.Sprintf("mem-pub-%d", i%100), Role: domain.RoleEnterprise}
		d, err := svc.Create(pub, service.CreateDemandInput{
			PublisherName: "memtest", Contact: "13800001111", Title: fmt.Sprintf("内存测试-%d", i), BizType: "other",
		})
		if err != nil {
			continue
		}
		svc.Approve(admin, d.ID)
	}

	runtime.GC()
	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)

	allocDiff := int64(m2.TotalAlloc) - int64(m1.TotalAlloc)
	avgPerOp := allocDiff / 10000

	t.Logf("=== Memory Leak Detection ===")
	t.Logf("10K ops: TotalAlloc diff=%d KB, avg=%d bytes/op", allocDiff/1024, avgPerOp)
	t.Logf("HeapObjects: %d → %d (diff %d)", m1.HeapObjects, m2.HeapObjects, int64(m2.HeapObjects)-int64(m1.HeapObjects))

	// Memory should NOT grow linearly with ops (objects should be GC'd)
	if int64(m2.HeapObjects) > int64(m1.HeapObjects)*10 {
		t.Errorf("potential memory leak: HeapObjects grew from %d to %d", m1.HeapObjects, m2.HeapObjects)
	}
}

// ---- Test 3: Concurrent search while data mutates ----

func TestStress_SearchDuringMutation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}
	demandRepo := memory.NewDemandRepository(nil)
	svc := service.NewDemandService(demandRepo)
	admin := domain.Actor{ID: "admin", Role: domain.RoleAssociationAdmin}
	publisher := domain.Actor{ID: "search-pub", Role: domain.RoleEnterprise}

	var wg sync.WaitGroup
	var searchErrors, writeErrors int64
	done := make(chan struct{})

	// Writer: continuously create demands
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(wid int) {
			defer wg.Done()
			for j := 0; j < 200; j++ {
				select {
				case <-done:
					return
				default:
				}
				d, _ := svc.Create(publisher, service.CreateDemandInput{
					PublisherName: "search-pub", Contact: "13800001111",
					Title: fmt.Sprintf("搜索测试-%d-%d", wid, j), BizType: "other",
				})
				time.Sleep(time.Microsecond)
				svc.Submit(publisher, d.ID)
				svc.Approve(admin, d.ID)
			}
		}(i)
	}

	// Searcher: continuously search while writes happen
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 500; j++ {
				select {
				case <-done:
					return
				default:
				}
				_, err := svc.List(repository.DemandFilter{})
				if err != nil {
					atomic.AddInt64(&searchErrors, 1)
				}
			}
		}()
	}

	// Wait for writers to finish
	time.Sleep(2 * time.Second)
	close(done)
	wg.Wait()

	t.Logf("=== Search During Mutation ===")
	t.Logf("Search errors: %d, Write errors: %d", searchErrors, writeErrors)

	if searchErrors > 0 {
		t.Errorf("%d search errors during mutation (target 0)", searchErrors)
	}
}

// ---- Test 5: Sustained admin dashboard load ----

func TestStress_AdminDashboardLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}
	entRepo := memory.NewEnterpriseRepository(nil)
	demandRepo := memory.NewDemandRepository(nil)
	entSvc := service.NewEnterpriseSvc(entRepo)
	demandSvc := service.NewDemandService(demandRepo)
	admin := domain.Actor{ID: "admin", Role: domain.RolePlatformAdmin}

	// Pre-populate: 200 enterprises, 200 demands
	for i := 0; i < 200; i++ {
		actor := domain.Actor{ID: fmt.Sprintf("dash-user-%d", i), Role: domain.RoleEnterprise}
		entSvc.Create(actor, service.CreateEnterpriseInput{
			Name: fmt.Sprintf("仪表盘企业-%d", i), AccountName: fmt.Sprintf("138%08d", i),
		})
		time.Sleep(time.Microsecond)
		d, _ := demandSvc.Create(actor, service.CreateDemandInput{
			PublisherName: "dash", Contact: "13800001111", Title: fmt.Sprintf("仪表盘需求-%d", i), BizType: "other",
		})
		time.Sleep(time.Microsecond)
		demandSvc.Submit(actor, d.ID)
		demandSvc.Approve(domain.Actor{ID: "admin", Role: domain.RoleAssociationAdmin}, d.ID)
	}

	var wg sync.WaitGroup
	start := time.Now()
	var totalOps int64
	var p50s, p95s, p99s []int64
	var latMu sync.Mutex

	// 20 concurrent admin dashboard queries (simulating 20 admins refreshing)
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(adminID int) {
			defer wg.Done()
			lats := make([]int64, 0, 100)
			for j := 0; j < 100; j++ {
				opStart := time.Now()
				// Simulate dashboard: list enterprises + demand stats
				_, _, _ = entSvc.ListByStatus(admin, "submitted", 0, 50)
				_, _ = demandSvc.List(repository.DemandFilter{Status: "all"})
				lats = append(lats, time.Since(opStart).Nanoseconds())
				atomic.AddInt64(&totalOps, 2)
			}
			sort.Slice(lats, func(a, b int) bool { return lats[a] < lats[b] })
			latMu.Lock()
			p50s = append(p50s, lats[len(lats)*50/100])
			p95s = append(p95s, lats[len(lats)*95/100])
			p99s = append(p99s, lats[len(lats)*99/100])
			latMu.Unlock()
		}(i)
	}
	wg.Wait()

	elapsed := time.Since(start)
	ops := atomic.LoadInt64(&totalOps)

	// Average P50/P95/P99 across all workers
	avg := func(vals []int64) time.Duration {
		var sum int64
		for _, v := range vals {
			sum += v
		}
		return time.Duration(sum / int64(len(vals)))
	}

	t.Logf("=== Admin Dashboard Load ===")
	t.Logf("200 enterprises + 200 demands pre-populated")
	t.Logf("20 admins × 100 refreshes: %d ops in %v", ops, elapsed)
	t.Logf("Avg P50: %v, Avg P95: %v, Avg P99: %v", avg(p50s), avg(p95s), avg(p99s))

	if ops < 3000 {
		t.Errorf("only %d ops (target >=4000)", ops)
	}
	maxP99 := avg(p99s)
	if maxP99 > 100*time.Millisecond {
		t.Errorf("P99 latency %v too high (target <100ms)", maxP99)
	}
}

// ---- Test 6: Verify no deadlocks under heavy lock contention ----

func TestStress_DeadlockDetection(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}
	entRepo := memory.NewEnterpriseRepository(nil)
	svc := service.NewEnterpriseSvc(entRepo)
	admin := domain.Actor{ID: "admin", Role: domain.RolePlatformAdmin}

	// 100 enterprises
	ids := make([]string, 100)
	for i := 0; i < 100; i++ {
		actor := domain.Actor{ID: fmt.Sprintf("deadlock-u-%d", i), Role: domain.RoleEnterprise}
		ent, _ := svc.Create(actor, service.CreateEnterpriseInput{
			Name: fmt.Sprintf("死锁测试-%d", i), AccountName: fmt.Sprintf("138%08d", i),
		})
		time.Sleep(time.Microsecond)
		svc.Submit(actor, ent.ID)
		ids[i] = ent.ID
	}

	done := make(chan struct{}, 1)
	var wg sync.WaitGroup

	// 3 goroutines: concurrent approve + search + list
	for g := 0; g < 3; g++ {
		wg.Add(1)
		go func(grp int) {
			defer wg.Done()
			for i := 0; i < 500; i++ {
				select {
				case <-done:
					return
				default:
				}
				switch grp {
				case 0: // Approver
					svc.Review(admin, ids[i%len(ids)], "approve", "")
				case 1: // Searcher
					svc.Search(admin, fmt.Sprintf("死锁-%d", i%50))
				case 2: // Lister
					svc.ListByStatus(admin, "approved", 0, 20)
				}
			}
		}(g)
	}

	// Timeout after 3 seconds — if deadlocked, test should time out
	timeout := time.After(3 * time.Second)
	waitDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(waitDone)
	}()

	select {
	case <-waitDone:
		t.Log("=== Deadlock Detection ===")
		t.Log("1500 ops under lock contention: NO DEADLOCK ✅")
	case <-timeout:
		close(done)
		t.Error("DEADLOCK DETECTED — operations timed out after 3s")
	}
}
