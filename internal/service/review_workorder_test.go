package service_test

import (
	"context"
	"testing"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

// 工单评价闭环：仅已完成工单的双方（需求方/接单方）可评价。
// 回归：此前 ReviewService.Submit 零校验，任意用户可对任意 target 刷评价
// （POST /api/v1/reviews 全仓无业务调用即因缺校验而无处可用）。
func TestReviewWorkOrderValidation(t *testing.T) {
	orderSvc, _, pub, worker, d, it := newWorkOrderScenario(t)
	wo, err := orderSvc.AcceptIntent(context.Background(), pub, d.ID, it.ID, 100000)
	if err != nil {
		t.Fatalf("accept: %v", err)
	}
	// 订单流转到 completed：pending → ongoing → awaiting_accept → completed
	wo, err = orderSvc.StartWork(context.Background(), worker, wo.ID)
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	wo, err = orderSvc.CompleteWork(context.Background(), worker, wo.ID, nil)
	if err != nil {
		t.Fatalf("complete: %v", err)
	}
	wo, err = orderSvc.AcceptCompletion(context.Background(), pub, wo.ID)
	if err != nil {
		t.Fatalf("accept completion: %v", err)
	}
	if wo.Status != domain.WorkOrderCompleted {
		t.Fatalf("status: got %s want completed", wo.Status)
	}

	orderRepo := memory.NewWorkOrderRepository()
	// 重建工单仓储（newWorkOrderScenario 内部仓储未暴露），把 completed 工单写入
	if _, err := orderRepo.Create(context.Background(), wo); err != nil {
		t.Fatalf("seed order repo: %v", err)
	}
	rv := service.NewReviewService(memory.NewReviewRepository(), orderRepo)

	// 需求方（发布者）可评价
	if _, err := rv.Submit(context.Background(), pub.ID, "work_order", wo.ID, 5, "作业专业，完成及时"); err != nil {
		t.Fatalf("publisher review should succeed: %v", err)
	}
	// 接单方（飞手）可评价
	if _, err := rv.Submit(context.Background(), worker.ID, "work_order", wo.ID, 4, "沟通顺畅"); err != nil {
		t.Fatalf("worker review should succeed: %v", err)
	}
	// 无关用户不可评价
	if _, err := rv.Submit(context.Background(), "stranger", "work_order", wo.ID, 5, "刷评价"); err == nil {
		t.Fatal("stranger review should be rejected")
	}
	// 不存在的工单
	if _, err := rv.Submit(context.Background(), pub.ID, "work_order", "wo-nope", 5, "x"); err == nil {
		t.Fatal("review of unknown work order should be rejected")
	}
	// 非 work_order 目标不受影响（企业评价仍可提交）
	if _, err := rv.Submit(context.Background(), pub.ID, "enterprise", "ent-1", 4, "企业评价"); err != nil {
		t.Fatalf("enterprise review should still work: %v", err)
	}
}

// 未完成工单不可评价（状态机前置）。
func TestReviewWorkOrderRequiresCompleted(t *testing.T) {
	orderSvc, _, pub, worker, d, it := newWorkOrderScenario(t)
	wo, err := orderSvc.AcceptIntent(context.Background(), pub, d.ID, it.ID, 100000)
	if err != nil {
		t.Fatalf("accept: %v", err)
	}
	// 仅到 ongoing（未完成）
	if _, err := orderSvc.StartWork(context.Background(), worker, wo.ID); err != nil {
		t.Fatalf("start: %v", err)
	}

	orderRepo := memory.NewWorkOrderRepository()
	if _, err := orderRepo.Create(context.Background(), wo); err != nil {
		t.Fatalf("seed order repo: %v", err)
	}
	rv := service.NewReviewService(memory.NewReviewRepository(), orderRepo)

	if _, err := rv.Submit(context.Background(), pub.ID, "work_order", wo.ID, 5, "提前评价"); err == nil {
		t.Fatal("review of non-completed work order should be rejected")
	}
}
