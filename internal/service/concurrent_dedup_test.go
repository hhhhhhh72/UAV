package service_test

import (
	"context"
	"sync"
	"testing"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

// 并发报名防重复回归：双并发 Enroll 只能成功一条（此前 check-then-insert
// 竞态会重复报名；付费报名场景还会重复扣冻结金额）。
func TestConcurrentEnrollSingle(t *testing.T) {
	svc := service.NewEnrollmentService(memory.NewEnrollmentRepository(), memory.NewCourseRepository())

	var wg sync.WaitGroup
	okCount := 0
	var mu sync.Mutex
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := svc.Enroll(context.Background(), "u-1", "c-1", service.EnrollmentForm{Name: "n"})
			if err == nil {
				mu.Lock()
				okCount++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	if okCount != 1 {
		t.Fatalf("concurrent enroll: want exactly 1 success, got %d", okCount)
	}
}

// 并发投递防重复回归：同一用户对同一职位并发投递只能成功一条。
func TestConcurrentApplySingle(t *testing.T) {
	jobRepo := memory.NewJobRepository()
	job, _ := jobRepo.Create(context.Background(), domain.Job{
		ID: "job-x", EnterpriseID: "ent-1", Title: "t", Status: domain.JobPublished, Version: 1,
	})
	svc := service.NewJobService(jobRepo, memory.NewResumeRepository(), memory.NewJobApplicationRepository())
	actor := domain.Actor{ID: "u-1", Role: domain.RoleIndividual}

	var wg sync.WaitGroup
	okCount := 0
	var mu sync.Mutex
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := svc.Apply(context.Background(), actor, job.ID, "res-1")
			if err == nil {
				mu.Lock()
				okCount++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	if okCount != 1 {
		t.Fatalf("concurrent apply: want exactly 1 success, got %d", okCount)
	}
}
