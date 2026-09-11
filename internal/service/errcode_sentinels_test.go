package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

// 服务层哨兵语义：仓储的「记录不存在」必须被翻成各聚合自己的 not-found 哨兵
// （Handler 才能稳定映射 404），状态冲突则是另一个哨兵（409）。
// 此前这些错误一路裸奔到 Handler，被一律当成 403/500。
func TestNotFoundAndStateConflictSentinels(t *testing.T) {
	ctx := context.Background()

	// 评价：不存在 → ErrReviewNotFound；已通过再驳回 → ErrReviewStateConflict
	reviewRepo := memory.NewReviewRepository()
	reviewSvc := service.NewReviewService(reviewRepo, memory.NewWorkOrderRepository())
	if err := reviewSvc.Approve(ctx, "ghost"); !errors.Is(err, service.ErrReviewNotFound) {
		t.Fatalf("通过不存在的评价应 ErrReviewNotFound，实际 %v", err)
	}
	if err := reviewSvc.Delete(ctx, "ghost"); !errors.Is(err, service.ErrReviewNotFound) {
		t.Fatalf("删除不存在的评价应 ErrReviewNotFound，实际 %v", err)
	}
	rv, err := reviewRepo.Create(ctx, domain.Review{
		ID: "review-1", ReviewerID: "user-1", TargetType: "demand", TargetID: "demand-1",
		Rating: 5, Content: "好", Status: "pending", CreatedAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("建评价: %v", err)
	}
	if err := reviewSvc.Approve(ctx, rv.ID); err != nil {
		t.Fatalf("通过评价: %v", err)
	}
	if err := reviewSvc.Reject(ctx, rv.ID); !errors.Is(err, service.ErrReviewStateConflict) {
		t.Fatalf("已通过的评价再驳回应 ErrReviewStateConflict，实际 %v", err)
	}

	// 证书：不存在 → ErrTrainingNotFound（不再因为"找不到"被说成"没权限"）
	trainingSvc := service.NewTrainingService(
		memory.NewCertificateRepository(), memory.NewCourseRepository(),
		memory.NewInstructorRepository(), memory.NewPilotRepository(nil))
	if _, err := trainingSvc.ApproveCertificate(ctx, adminActor(), "ghost"); !errors.Is(err, service.ErrTrainingNotFound) {
		t.Fatalf("通过不存在的证书应 ErrTrainingNotFound，实际 %v", err)
	}
	if _, err := trainingSvc.ApproveInstructor(ctx, adminActor(), "ghost"); !errors.Is(err, service.ErrTrainingNotFound) {
		t.Fatalf("通过不存在的培训师应 ErrTrainingNotFound，实际 %v", err)
	}
	if _, err := trainingSvc.ApprovePilot(ctx, adminActor(), "ghost"); !errors.Is(err, service.ErrTrainingNotFound) {
		t.Fatalf("通过不存在的飞手应 ErrTrainingNotFound，实际 %v", err)
	}

	// 需求：不存在 → ErrDemandNotFound
	demandSvc := service.NewDemandService(memory.NewDemandRepository(nil))
	if err := demandSvc.Delete(ctx, adminActor(), "ghost"); !errors.Is(err, service.ErrDemandNotFound) {
		t.Fatalf("删除不存在的需求应 ErrDemandNotFound，实际 %v", err)
	}
}

func adminActor() domain.Actor {
	return domain.Actor{ID: "admin-1", Role: domain.RolePlatformAdmin}
}
