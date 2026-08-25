package service_test

import (
	"context"
	"testing"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

// 需求截止日期校验（C2）：合法未来日期归一为 YYYY-MM-DD；过去日期/非法格式拒绝。
func TestDemandDeadlineValidation(t *testing.T) {
	ctx := context.Background()
	ds := service.NewDemandService(memory.NewDemandRepository(nil))
	pub := domain.Actor{ID: "pub-deadline", Role: domain.RoleEnterprise}
	base := service.CreateDemandInput{PublisherName: "企业", Contact: "13800000000", Title: "时效需求"}

	// 未来日期 → 创建成功且归一 YYYY-MM-DD
	future := time.Now().AddDate(0, 0, 3).Format("2006-01-02")
	in := base
	in.Deadline = future
	d, err := ds.Create(ctx, pub, in)
	if err != nil {
		t.Fatalf("create with future deadline: %v", err)
	}
	if d.Deadline != future {
		t.Fatalf("deadline normalized: want %s, got %q", future, d.Deadline)
	}

	// 今天（当天有效）→ 允许
	in = base
	in.Deadline = time.Now().Format("2006-01-02")
	if _, err := ds.Create(ctx, pub, in); err != nil {
		t.Fatalf("create with today deadline: %v", err)
	}

	// 过去日期 → 拒绝
	in = base
	in.Deadline = time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	if _, err := ds.Create(ctx, pub, in); err == nil {
		t.Fatal("past deadline should be rejected")
	}

	// 非法格式 → 拒绝
	in = base
	in.Deadline = "2026/13/99"
	if _, err := ds.Create(ctx, pub, in); err == nil {
		t.Fatal("invalid deadline format should be rejected")
	}

	// 空串 → 长期有效
	in = base
	in.Deadline = ""
	d2, err := ds.Create(ctx, pub, in)
	if err != nil || d2.Deadline != "" {
		t.Fatalf("empty deadline should mean no expiry, got %q err=%v", d2.Deadline, err)
	}
}

// 工单金额不得超过需求预算（C2）：预算 0 = 面议/未填不设限。
func TestAcceptIntentAmountWithinBudget(t *testing.T) {
	ctx := context.Background()
	demandRepo := memory.NewDemandRepository(nil)
	intentRepo := memory.NewIntentRepository()
	orderRepo := memory.NewWorkOrderRepository()
	demandSvc := service.NewDemandService(demandRepo)
	intentSvc := service.NewIntentService(intentRepo, demandRepo)
	orderSvc := service.NewWorkOrderService(orderRepo, demandRepo, intentRepo)

	pub := domain.Actor{ID: "pub-budget", Role: domain.RoleEnterprise}
	worker := domain.Actor{ID: "worker-budget", Role: domain.RoleIndividual}

	d, err := demandSvc.Create(ctx, pub, service.CreateDemandInput{
		PublisherName: "预算企业", Contact: "13800000000", Title: "预算需求", BudgetFen: 100000,
	})
	if err != nil {
		t.Fatalf("create demand: %v", err)
	}
	if _, err := demandSvc.Review(ctx, domain.Actor{ID: "admin", Role: domain.RolePlatformAdmin}, d.ID, "approve", ""); err != nil {
		t.Fatalf("approve demand: %v", err)
	}
	it, err := intentSvc.Create(ctx, worker, d.ID, service.CreateIntentInput{
		IntentorName: "飞手预算", Contact: "13900000000",
	})
	if err != nil {
		t.Fatalf("create intent: %v", err)
	}

	// 超预算 → 拒绝
	if _, err := orderSvc.AcceptIntent(ctx, pub, d.ID, it.ID, 100001); err == nil {
		t.Fatal("amount above budget should be rejected")
	}
	// 等额 → 成功
	if _, err := orderSvc.AcceptIntent(ctx, pub, d.ID, it.ID, 100000); err != nil {
		t.Fatalf("amount equal to budget should be accepted: %v", err)
	}

	// 面议需求（预算 0）：任意金额可确认
	d2, err := demandSvc.Create(ctx, pub, service.CreateDemandInput{
		PublisherName: "面议企业", Contact: "13800000000", Title: "面议需求",
	})
	if err != nil {
		t.Fatalf("create demand2: %v", err)
	}
	if _, err := demandSvc.Review(ctx, domain.Actor{ID: "admin", Role: domain.RolePlatformAdmin}, d2.ID, "approve", ""); err != nil {
		t.Fatalf("approve demand2: %v", err)
	}
	it2, err := intentSvc.Create(ctx, worker, d2.ID, service.CreateIntentInput{
		IntentorName: "飞手面议", Contact: "13900000001",
	})
	if err != nil {
		t.Fatalf("create intent2: %v", err)
	}
	if _, err := orderSvc.AcceptIntent(ctx, pub, d2.ID, it2.ID, 500000); err != nil {
		t.Fatalf("negotiable demand should accept any amount: %v", err)
	}
}
