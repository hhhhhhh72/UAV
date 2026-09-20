package service_test

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

// TestStudyTourRepoCapacityWithoutLock 直接压**仓储层**（绕过 service 的研学维度键锁）：
// 库级兜底必须自己站得住 —— 多实例部署时键锁各锁各的，只有仓储层这道才跨进程有效。
func TestStudyTourRepoCapacityWithoutLock(t *testing.T) {
	const (
		capacity   = 10
		concurrent = 200
	)
	ctx := context.Background()
	repo := memory.NewStudyTourEnrollmentRepository()

	var wg sync.WaitGroup
	var ok int32
	for i := 0; i < concurrent; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, err := repo.CreateWithCapacity(ctx, domain.StudyTourEnrollment{
				ID: fmt.Sprintf("se-%d", i), TourID: "tour-x", UserID: fmt.Sprintf("u-%d", i),
				Name: "张三", Phone: "13800000000", AdultCount: 1, Status: "pending",
			}, 1, capacity)
			if err == nil {
				atomic.AddInt32(&ok, 1)
			}
		}(i)
	}
	wg.Wait()

	items, err := repo.ListByTour(ctx, "tour-x")
	if err != nil {
		t.Fatal(err)
	}
	taken := 0
	for _, e := range items {
		if e.Status == "pending" || e.Status == "approved" {
			taken += e.AdultCount + e.ChildCount
		}
	}
	t.Logf("仓储层直压：容量 %d，并发 %d → 成功 %d 条，占用 %d 人", capacity, concurrent, ok, taken)
	if taken > capacity || int(ok) != capacity {
		t.Fatalf("仓储层容量兜底失效：容量 %d，成功 %d 条、占用 %d 人", capacity, ok, taken)
	}
}

// TestCourseEnrollmentNoOversell 课程报名容量的正向对照：这条链路**已经**有
// 课程维度键锁（phase3.go:137 lockByKey("enroll-course|"+courseID)），
// 所以 200 人抢 2 个座位应当只成功 2 条。它和上面的研学用例一起说明：
// 问题不是"键锁这个方案不行"，而是研学那一处漏了。
func TestCourseEnrollmentNoOversell(t *testing.T) {
	const (
		seats      = 2
		concurrent = 200
	)
	ctx := context.Background()

	courseRepo := memory.NewCourseRepository()
	// 建课在 TrainingService，报名在 EnrollmentService —— 共用同一个 courseRepo 实例，
	// 否则报名侧看不到课程、BumpEnrolled 也作用不到同一条记录。
	training := service.NewTrainingService(
		memory.NewCertificateRepository(), courseRepo,
		memory.NewInstructorRepository(), memory.NewPilotRepository(nil),
	)
	svc := service.NewEnrollmentService(memory.NewEnrollmentRepository(), courseRepo)
	admin := domain.Actor{ID: "admin-1", Role: domain.RolePlatformAdmin}
	c, err := training.CreateCourse(ctx, admin, domain.TrainingCourse{
		Title: "两座课", MaxStudents: seats, OrgID: "org-1", PriceFen: 0, Status: "published",
	})
	if err != nil {
		t.Fatalf("CreateCourse: %v", err)
	}

	var wg sync.WaitGroup
	var ok int32
	for i := 0; i < concurrent; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if _, err := svc.Enroll(ctx, fmt.Sprintf("u-%d", i), c.ID, service.EnrollmentForm{Name: "张三"}); err == nil {
				atomic.AddInt32(&ok, 1)
			}
		}(i)
	}
	wg.Wait()

	got, err := courseRepo.FindByID(ctx, c.ID)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("座位 %d，并发 %d 人报名 → 成功 %d 条，enrolled_count=%d，remain=%d",
		seats, concurrent, ok, got.EnrolledCount, got.Remain)
	if got.EnrolledCount > seats {
		t.Fatalf("课程超卖：座位 %d，却报满 %d 人", seats, got.EnrolledCount)
	}
	if ok != seats {
		t.Fatalf("成功条数 %d，want %d", ok, seats)
	}
}

// TestStudyTourEnrollmentNoOversell 研学报名的容量校验必须扛住并发。
//
// 回归背景：StudyTourEnrollmentService.Create（study_tour_enrollments.go:50-79）
// 是纯 check-then-act —— ListByTour 求和 → 比容量 → Create，**既没有 lockByKey
// 也没有库级唯一索引**（migrations 里 study_tour_enrollments 无任何唯一索引）。
// 并发报名时每一方都读到 taken=0，于是全部放行 → 超卖。
// 同仓的课程报名（phase3.go:137 lockByKey("enroll-course|"+courseID)）、
// 测试场地/场馆/赛事/活动报名都上了进程内键锁，研学漏了。
func TestStudyTourEnrollmentNoOversell(t *testing.T) {
	const (
		capacity   = 10
		concurrent = 200
	)
	ctx := context.Background()

	tourRepo := memory.NewStudyTourRepository()
	if _, err := tourRepo.Create(ctx, domain.StudyTour{
		ID: "tour-cap", Title: "低空研学", Status: "active", Capacity: capacity,
	}); err != nil {
		t.Fatalf("create tour: %v", err)
	}
	svc := service.NewStudyTourEnrollmentService(memory.NewStudyTourEnrollmentRepository(), tourRepo)

	var wg sync.WaitGroup
	var mu sync.Mutex
	ok := 0
	for i := 0; i < concurrent; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, err := svc.Create(ctx, fmt.Sprintf("u-%d", i), "tour-cap", "张三", "13800000000", 1, 0, "")
			if err == nil {
				mu.Lock()
				ok++
				mu.Unlock()
			}
		}(i)
	}
	wg.Wait()

	items, err := svc.ListByTour(ctx, "tour-cap")
	if err != nil {
		t.Fatal(err)
	}
	taken := 0
	for _, e := range items {
		if e.Status == "pending" || e.Status == "approved" {
			taken += e.AdultCount + e.ChildCount
		}
	}
	t.Logf("容量 %d，并发 %d 人报名 → 实际收下 %d 人（成功 %d 条）", capacity, concurrent, taken, ok)
	if taken > capacity {
		t.Fatalf("研学超卖：容量 %d，却收下 %d 人（多出 %d 人）", capacity, taken, taken-capacity)
	}
}
