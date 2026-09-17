package service_test

import (
	"context"
	"testing"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

// ListReceived 是发布方消息页「一键同意」的判断依据：
// 必须跨该发布者名下**所有**需求一次取回（消息页据此判断"是否只有唯一一条待处理申请"），
// 且绝不能串到别人的需求上——串了就会把不相干的人认成唯一申请人。
func TestIntentService_ListReceived(t *testing.T) {
	ctx := context.Background()
	demandRepo := memory.NewDemandRepository(nil)
	intentRepo := memory.NewIntentRepository(demandRepo)
	demandSvc := service.NewDemandService(demandRepo)

	entRepo := memory.NewEnterpriseRepository(nil)
	seedEntCertRepo(t, entRepo, "worker-1")

	intentSvc := service.NewIntentService(intentRepo, demandRepo, entRepo, memory.NewPilotRepository(nil))
	pub := domain.Actor{ID: "pub-1", Role: domain.RoleEnterprise}
	other := domain.Actor{ID: "pub-2", Role: domain.RoleEnterprise}
	worker := domain.Actor{ID: "worker-1", Role: domain.RoleIndividual}

	d1 := publishDemandForTest(t, demandSvc, pub, "需求一")
	d2 := publishDemandForTest(t, demandSvc, pub, "需求二")
	dOther := publishDemandForTest(t, demandSvc, other, "别人的需求")

	for _, id := range []string{d1.ID, d2.ID, dOther.ID} {
		if _, err := intentSvc.Create(ctx, worker, id, service.CreateIntentInput{IntentorName: "飞手", Contact: "13900000000"}); err != nil {
			t.Fatalf("Create intent on %s: %v", id, err)
		}
	}

	// 跨需求聚合：pub 名下两条需求各一条意向，一次全取回
	got, err := intentSvc.ListReceived(ctx, pub)
	if err != nil {
		t.Fatalf("ListReceived(pub): %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("ListReceived(pub): len=%d, want 2（跨需求聚合）", len(got))
	}
	for _, it := range got {
		if it.DemandID == dOther.ID {
			t.Fatalf("ListReceived(pub) 串到了他人需求 %s 上的意向", dOther.ID)
		}
		if it.DemandTitle == "" {
			t.Fatalf("ListReceived(pub) 的意向缺少 demand_title，聚合页会显示不出「申请项目」: %+v", it)
		}
	}

	// 归属隔离：other 只该看到自己那条需求上的意向
	otherGot, err := intentSvc.ListReceived(ctx, other)
	if err != nil {
		t.Fatalf("ListReceived(other): %v", err)
	}
	if len(otherGot) != 1 || otherGot[0].DemandID != dOther.ID {
		t.Fatalf("ListReceived(other): len=%d want 1 且只含 %s", len(otherGot), dOther.ID)
	}

	// 没发布过需求的人 → 空
	if none, err := intentSvc.ListReceived(ctx, domain.Actor{ID: "nobody", Role: domain.RoleIndividual}); err != nil || len(none) != 0 {
		t.Fatalf("ListReceived(nobody): len=%d err=%v, want 0/nil", len(none), err)
	}

	// 空 actor（未认证）→ 拒绝，堵住"空 ID 查出全表"这类越权
	if _, err := intentSvc.ListReceived(ctx, domain.Actor{}); err == nil {
		t.Fatal("ListReceived(空 actor): 应拒绝")
	}
}
