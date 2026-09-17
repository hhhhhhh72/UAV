package service_test

import (
	"context"
	"testing"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

// resourceID 为空必须被拒：空条件会命中管理端广播（resource_type='' 且 resource_id=''），
// 把全站公告一次删光。这道守卫不能只靠仓储层兜底。
func TestMessageServiceDeleteByReferenceGuardsEmptyResource(t *testing.T) {
	repo := memory.NewMessageRepository()
	svc := service.NewMessageService(repo)
	ctx := context.Background()

	if _, err := repo.Create(ctx, domain.Message{
		ID: "b-1", ReceiverID: "u-1", SenderID: "sys", Title: "全站公告",
	}); err != nil {
		t.Fatalf("seed broadcast: %v", err)
	}

	if _, err := svc.DeleteByReference(ctx, "", []string{"demand"}); err == nil {
		t.Fatal("空 resourceID 应被拒绝")
	}
	if _, err := repo.FindByID(ctx, "b-1"); err != nil {
		t.Fatal("空 resourceID 的调用把管理端广播删掉了")
	}

	// 正常路径：命中即删
	if _, err := repo.Create(ctx, domain.Message{
		ID: "n-1", ReceiverID: "u-1", SenderID: "sys", ResourceType: "demand", ResourceID: "d-1",
	}); err != nil {
		t.Fatalf("seed notice: %v", err)
	}
	n, err := svc.DeleteByReference(ctx, "d-1", []string{"demand", "demand_intent"})
	if err != nil || n != 1 {
		t.Fatalf("DeleteByReference(d-1): n=%d err=%v，want 1/nil", n, err)
	}
}
