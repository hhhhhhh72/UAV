package postgres_test

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"drone-platform/internal/domain"
)

// TestPG_StudyTourCapacityAtomic 研学容量是**求和式**的（没有计数列），
// 所以库级兜底只能靠「同一事务里 SELECT … FOR UPDATE 锁住研学行 → 求和 → 插入」。
// 这里直接调仓储层、**绕过 service 的研学维度键锁**，50 个并发抢 3 个名额。
func TestPG_StudyTourCapacityAtomic(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	ctx := context.Background()
	tourRepo := store.NewStudyTourRepository()
	repo := store.NewStudyTourEnrollmentRepository()

	const (
		capacity   = 3
		concurrent = 50
	)
	tour, err := tourRepo.Create(ctx, domain.StudyTour{
		ID: ug("tour"), Title: "容量并发", Status: "active", Capacity: capacity,
	})
	if err != nil {
		t.Fatalf("create tour: %v", err)
	}

	var wg sync.WaitGroup
	var ok int32
	for i := 0; i < concurrent; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, cerr := repo.CreateWithCapacity(ctx, domain.StudyTourEnrollment{
				ID: ug("se"), TourID: tour.ID, UserID: fmt.Sprintf("u-%d", i),
				Name: "张三", Phone: "13800000000", AdultCount: 1, ChildCount: 0, Status: "pending",
			}, 1, tour.Capacity)
			if cerr == nil {
				atomic.AddInt32(&ok, 1)
			}
		}(i)
	}
	wg.Wait()

	items, err := repo.ListByTour(ctx, tour.ID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	taken := 0
	for _, e := range items {
		if e.Status == "pending" || e.Status == "approved" {
			taken += e.AdultCount + e.ChildCount
		}
	}
	if int(ok) != capacity || taken != capacity {
		t.Fatalf("超卖：容量 %d，CreateWithCapacity 成功 %d 次，占用 %d 人", capacity, ok, taken)
	}
}
