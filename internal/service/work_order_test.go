package service_test

import (
	"context"
	"strings"
	"testing"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

// 企业工作台闭环测试：需求 → 意向 → 确认接单 → 订单 → 执行 → 验收

// 构造场景：企业 publisher 发需求并提交通过，飞手 worker 投意向
func newWorkOrderScenario(t *testing.T) (*service.WorkOrderService, *service.IntentService, domain.Actor, domain.Actor, domain.Demand, domain.DemandIntent) {
	t.Helper()
	demandRepo := memory.NewDemandRepository(nil)
	intentRepo := memory.NewIntentRepository()
	orderRepo := memory.NewWorkOrderRepository()

	demandSvc := service.NewDemandService(demandRepo)
	entRepo := memory.NewEnterpriseRepository(nil)
	seedEntCertRepo(t, entRepo, "worker-1")
	seedEntCertRepo(t, entRepo, "worker-2")
	intentSvc := service.NewIntentService(intentRepo, demandRepo, entRepo, memory.NewPilotRepository(nil))
	orderSvc := service.NewWorkOrderService(orderRepo, demandRepo, intentRepo)

	pub := domain.Actor{ID: "pub-1", Role: domain.RoleEnterprise}
	worker := domain.Actor{ID: "worker-1", Role: domain.RoleIndividual}

	d, err := demandSvc.Create(context.Background(), pub, service.CreateDemandInput{
		PublisherName: "测试企业", Contact: "13800000000",
		District: "渝北区", Title: "电力巡检需求", Description: "50公里线路巡检",
	})
	if err != nil {
		t.Fatalf("create demand: %v", err)
	}
	if _, err := demandSvc.Review(context.Background(), domain.Actor{ID: "admin", Role: domain.RolePlatformAdmin}, d.ID, "approve", ""); err != nil {
		t.Fatalf("approve demand: %v", err)
	}
	it, err := intentSvc.Create(context.Background(), worker, d.ID, service.CreateIntentInput{
		IntentorName: "飞手小张", Contact: "13900000000", Remark: "可完成巡检作业",
	})
	if err != nil {
		t.Fatalf("create intent: %v", err)
	}
	return orderSvc, intentSvc, pub, worker, d, it
}

func TestAcceptIntentCreatesOrder(t *testing.T) {
	orderSvc, _, pub, _, d, it := newWorkOrderScenario(t)

	wo, err := orderSvc.AcceptIntent(context.Background(), pub, d.ID, it.ID, 150000)
	if err != nil {
		t.Fatalf("accept intent: %v", err)
	}
	if wo.DemandID != d.ID || wo.PublisherID != "pub-1" || wo.WorkerID != "worker-1" {
		t.Fatalf("order fields wrong: %+v", wo)
	}
	if wo.AmountFen != 150000 {
		t.Fatalf("amount: got %d want 150000", wo.AmountFen)
	}
	if wo.Status != domain.WorkOrderPending {
		t.Fatalf("status: got %s want pending", wo.Status)
	}
	if !strings.HasPrefix(wo.OrderNo, "WO") {
		t.Fatalf("order no format: %s", wo.OrderNo)
	}
}

func TestAcceptIntentClosesOtherIntents(t *testing.T) {
	orderSvc, intentSvc, pub, _, d, it := newWorkOrderScenario(t)
	other := domain.Actor{ID: "worker-2", Role: domain.RoleIndividual}
	if _, err := intentSvc.Create(context.Background(), other, d.ID, service.CreateIntentInput{
		IntentorName: "飞手小李", Contact: "13700000000",
	}); err != nil {
		t.Fatalf("second intent: %v", err)
	}
	if _, err := orderSvc.AcceptIntent(context.Background(), pub, d.ID, it.ID, 0); err != nil {
		t.Fatalf("accept: %v", err)
	}
	intents, err := intentSvc.ListByDemand(context.Background(), pub, d.ID)
	if err != nil {
		t.Fatalf("list intents: %v", err)
	}
	for _, in := range intents {
		if in.ID == it.ID && in.Status != "contacted" {
			t.Fatalf("accepted intent should be contacted, got %s", in.Status)
		}
		if in.ID != it.ID && in.Status != "closed" {
			t.Fatalf("other intent should be closed, got %s", in.Status)
		}
	}
}

func TestAcceptIntentPermissionChecks(t *testing.T) {
	orderSvc, _, _, _, d, it := newWorkOrderScenario(t)
	// 非发布者不能确认
	if _, err := orderSvc.AcceptIntent(context.Background(), domain.Actor{ID: "other", Role: domain.RoleEnterprise}, d.ID, it.ID, 0); err == nil {
		t.Fatal("non-publisher should be rejected")
	}
	// 不存在的意向
	if _, err := orderSvc.AcceptIntent(context.Background(), domain.Actor{ID: "pub-1", Role: domain.RoleEnterprise}, d.ID, "nope", 0); err == nil {
		t.Fatal("unknown intent should fail")
	}
}

func TestRejectIntent(t *testing.T) {
	orderSvc, intentSvc, pub, _, d, it := newWorkOrderScenario(t)
	if err := orderSvc.RejectIntent(context.Background(), pub, d.ID, it.ID); err != nil {
		t.Fatalf("reject: %v", err)
	}
	intents, _ := intentSvc.ListByDemand(context.Background(), pub, d.ID)
	if intents[0].Status != "closed" {
		t.Fatalf("rejected intent should be closed, got %s", intents[0].Status)
	}
	// 二次拒绝/确认应失败
	if err := orderSvc.RejectIntent(context.Background(), pub, d.ID, it.ID); err == nil {
		t.Fatal("rejecting closed intent should fail")
	}
}

func TestWorkOrderLifecycle(t *testing.T) {
	orderSvc, _, pub, worker, d, it := newWorkOrderScenario(t)
	wo, err := orderSvc.AcceptIntent(context.Background(), pub, d.ID, it.ID, 100000)
	if err != nil {
		t.Fatalf("accept: %v", err)
	}

	// 飞手确认开始
	wo, err = orderSvc.StartWork(context.Background(), worker, wo.ID)
	if err != nil || wo.Status != domain.WorkOrderOngoing {
		t.Fatalf("start: %v status %s", err, wo.Status)
	}
	// 未开始前企业不能验收
	if _, err := orderSvc.AcceptCompletion(context.Background(), pub, wo.ID); err == nil {
		t.Fatal("accept before start should fail")
	}
	// 飞手确认完成（带成果照片）
	wo, err = orderSvc.CompleteWork(context.Background(), worker, wo.ID, []string{"a.jpg", "b.jpg"})
	if err != nil || wo.Status != domain.WorkOrderAwaitingAccept {
		t.Fatalf("complete: %v status %s", err, wo.Status)
	}
	if len(wo.ResultPhotos) != 2 {
		t.Fatalf("photos: %v", wo.ResultPhotos)
	}
	// 企业验收
	wo, err = orderSvc.AcceptCompletion(context.Background(), pub, wo.ID)
	if err != nil || wo.Status != domain.WorkOrderCompleted {
		t.Fatalf("accept completion: %v status %s", err, wo.Status)
	}
}

func TestWorkOrderRework(t *testing.T) {
	orderSvc, _, pub, worker, d, it := newWorkOrderScenario(t)
	wo, _ := orderSvc.AcceptIntent(context.Background(), pub, d.ID, it.ID, 100000)
	wo, _ = orderSvc.StartWork(context.Background(), worker, wo.ID)
	wo, _ = orderSvc.CompleteWork(context.Background(), worker, wo.ID, nil)

	// 企业提整改 → 回进行中
	wo, err := orderSvc.RequestRework(context.Background(), pub, wo.ID, "照片不清晰，请重新拍摄")
	if err != nil || wo.Status != domain.WorkOrderOngoing {
		t.Fatalf("rework: %v status %s", err, wo.Status)
	}
	if wo.ReworkNote != "照片不清晰，请重新拍摄" {
		t.Fatalf("rework note: %s", wo.ReworkNote)
	}
	// 飞手重新完成
	wo, err = orderSvc.CompleteWork(context.Background(), worker, wo.ID, []string{"c.jpg"})
	if err != nil || wo.Status != domain.WorkOrderAwaitingAccept {
		t.Fatalf("re-complete: %v", err)
	}
}

func TestWorkOrderCancel(t *testing.T) {
	orderSvc, _, pub, worker, d, it := newWorkOrderScenario(t)
	wo, _ := orderSvc.AcceptIntent(context.Background(), pub, d.ID, it.ID, 100000)

	// 企业取消（填原因）
	wo, err := orderSvc.RequestCancel(context.Background(), pub, wo.ID, "计划变更")
	if err != nil || wo.Status != domain.WorkOrderCancelled {
		t.Fatalf("cancel: %v status %s", err, wo.Status)
	}
	// 已取消订单不能再操作
	if _, err := orderSvc.StartWork(context.Background(), worker, wo.ID); err == nil {
		t.Fatal("start cancelled order should fail")
	}
}

func TestListMine(t *testing.T) {
	orderSvc, intentSvc, pub, worker, d, it := newWorkOrderScenario(t)
	// 业务规则：需求被接单后进入 assigned（一单一主）——锁新意向与再次接单
	if _, err := orderSvc.AcceptIntent(context.Background(), pub, d.ID, it.ID, 100000); err != nil {
		t.Fatalf("accept1: %v", err)
	}
	if _, err := intentSvc.Create(context.Background(), worker, d.ID, service.CreateIntentInput{
		IntentorName: "飞手小张", Contact: "13900000000",
	}); err == nil {
		t.Fatal("assigned demand should reject new intents")
	}
	if _, err := orderSvc.AcceptIntent(context.Background(), pub, d.ID, it.ID, 200000); err == nil {
		t.Fatal("assigned demand should reject re-accept")
	}

	mine, err := orderSvc.ListMine(context.Background(), pub)
	if err != nil || len(mine) != 1 {
		t.Fatalf("publisher mine: %d %v", len(mine), err)
	}
	mine2, err := orderSvc.ListMine(context.Background(), worker)
	if err != nil || len(mine2) != 1 {
		t.Fatalf("worker mine: %d %v", len(mine2), err)
	}
	// 无关用户看不到订单
	mine3, err := orderSvc.ListMine(context.Background(), domain.Actor{ID: "stranger", Role: domain.RoleIndividual})
	if err != nil || len(mine3) != 0 {
		t.Fatalf("stranger mine: %d", len(mine3))
	}
}
