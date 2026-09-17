package memory_test

import (
	"context"
	"testing"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository/memory"
)

// ListByPublisher 的内存实现要与 PG 的 JOIN 语义对齐：跨需求按发布者聚合意向。
// 内存实现没有 JOIN，靠注入的需求仓储解析归属；未注入时必须**明确报错**——
// 静默返回空会让消息页的"一键同意"按钮该出而不出，这种静默失效最难排查。
func TestIntentRepoListByPublisher(t *testing.T) {
	ctx := context.Background()
	demandRepo := memory.NewDemandRepository(nil)
	repo := memory.NewIntentRepository(demandRepo)

	for _, d := range []domain.Demand{
		{ID: "d-1", PublisherID: "pub-1", Title: "需求一", Status: domain.DemandPublished},
		{ID: "d-2", PublisherID: "pub-2", Title: "需求二", Status: domain.DemandPublished},
	} {
		if _, err := demandRepo.Create(ctx, d); err != nil {
			t.Fatalf("seed demand %s: %v", d.ID, err)
		}
	}
	for _, it := range []domain.DemandIntent{
		{ID: "i-1", DemandID: "d-1", IntentorID: "u-1", Status: "pending"},
		{ID: "i-2", DemandID: "d-2", IntentorID: "u-1", Status: "pending"},
	} {
		if _, err := repo.Create(ctx, it); err != nil {
			t.Fatalf("seed intent %s: %v", it.ID, err)
		}
	}

	// 只拿到自己需求上的意向
	got, err := repo.ListByPublisher(ctx, "pub-1")
	if err != nil {
		t.Fatalf("ListByPublisher: %v", err)
	}
	if len(got) != 1 || got[0].ID != "i-1" {
		t.Fatalf("ListByPublisher(pub-1)=%+v, want 仅 i-1", got)
	}
	// 聚合页要显示「申请项目」，标题必须随查询带出（与 PG 的 JOIN d.title 对齐）
	if got[0].DemandTitle != "需求一" {
		t.Fatalf("DemandTitle=%q, want 需求一", got[0].DemandTitle)
	}
	// 没有需求的人 → 空
	if none, err := repo.ListByPublisher(ctx, "pub-none"); err != nil || len(none) != 0 {
		t.Fatalf("ListByPublisher(pub-none): len=%d err=%v, want 0/nil", len(none), err)
	}
	// 未注入需求仓储 → 报错而非静默空
	if _, err := memory.NewIntentRepository(nil).ListByPublisher(ctx, "pub-1"); err == nil {
		t.Fatal("未注入需求仓储时应报错，而不是静默返回空")
	}
}
