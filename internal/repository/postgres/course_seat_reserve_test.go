package postgres_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"drone-platform/internal/domain"
)

// TestPG_CourseReserveSeatNoOversell 课程容量的**库级**兜底：直接调仓储层
// （绕过 service 的课程维度键锁），50 个并发抢 3 个座位。
//
// 为什么必须有这一层：service/phase3.go 的 lockByKey("enroll-course|…") 是
// **进程内**锁，多一个 API 实例就各锁各的 → 一起静默超卖（研学同类问题的实测：
// 容量 10 被 200 并发收下 87 人）。ReserveSeat 把容量判断与 +1 放进同一条
// UPDATE，RowsAffected=0 即已满，跨进程同样成立。
func TestPG_CourseReserveSeatNoOversell(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	ctx := context.Background()
	repo := store.NewCourseRepository()

	const (
		seats      = 3
		concurrent = 50
	)
	c, err := repo.Create(ctx, domain.TrainingCourse{
		ID: ug("course"), OrgID: "org-1", Title: "并发占座", MaxStudents: seats,
		Status: "published", CreatedAt: time.Now(), UpdatedAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("create course: %v", err)
	}

	var wg sync.WaitGroup
	var got int32
	for i := 0; i < concurrent; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ok, rerr := repo.ReserveSeat(ctx, c.ID)
			if rerr == nil && ok {
				atomic.AddInt32(&got, 1)
			}
		}()
	}
	wg.Wait()

	after, err := repo.FindByID(ctx, c.ID)
	if err != nil {
		t.Fatalf("find course: %v", err)
	}
	if got != seats || after.EnrolledCount != seats {
		t.Fatalf("超卖：座位 %d，ReserveSeat 成功 %d 次，enrolled_count=%d",
			seats, got, after.EnrolledCount)
	}
	if after.Remain != 0 {
		t.Fatalf("占满后 remain=%d，want 0", after.Remain)
	}
}

// TestPG_CourseReserveSeatUnlimited max_students=0 表示不限量：并发占座应全部成功，
// 且 remain 保持 0（不限额课程没有"剩余名额"概念）。
func TestPG_CourseReserveSeatUnlimited(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	ctx := context.Background()
	repo := store.NewCourseRepository()

	const concurrent = 30
	c, err := repo.Create(ctx, domain.TrainingCourse{
		ID: ug("course"), OrgID: "org-1", Title: "不限量", MaxStudents: 0,
		Status: "published", CreatedAt: time.Now(), UpdatedAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("create course: %v", err)
	}

	var wg sync.WaitGroup
	var got int32
	for i := 0; i < concurrent; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ok, rerr := repo.ReserveSeat(ctx, c.ID)
			if rerr == nil && ok {
				atomic.AddInt32(&got, 1)
			}
		}()
	}
	wg.Wait()

	after, err := repo.FindByID(ctx, c.ID)
	if err != nil {
		t.Fatalf("find course: %v", err)
	}
	if got != concurrent || after.EnrolledCount != concurrent {
		t.Fatalf("不限量课程占座成功 %d 次（enrolled_count=%d），want %d",
			got, after.EnrolledCount, concurrent)
	}
	if after.Remain != 0 {
		t.Fatalf("不限量课程 remain=%d，want 0", after.Remain)
	}
}

// TestPG_CourseReserveSeatMissingCourse 课程不存在时返回 (false, nil) ——
// 调用方须按「课程不存在」处理，不得当成「已满」（否则无课程仓储的历史路径会误报满员）。
func TestPG_CourseReserveSeatMissingCourse(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	ok, err := store.NewCourseRepository().ReserveSeat(context.Background(), ug("no-such-course"))
	if err != nil {
		t.Fatalf("missing course should not error: %v", err)
	}
	if ok {
		t.Fatal("不存在的课程不该占到座位")
	}
}
