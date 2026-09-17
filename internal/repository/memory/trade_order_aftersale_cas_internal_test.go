package memory_test

import (
	"context"
	"testing"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository/memory"
)

// 售后审核的 CAS 原子性（仓储层）：两个并发审批者都读到 aftersale_status="pending" 时，
// 只能有一个写入成功。这里用"两个调用者持有同一份 pending 快照"复现该交错，
// 不依赖 goroutine 调度，因此是确定性用例。
//
// 若失去这道 CAS，售后 B 分支（货款已放给卖家）会各自调用 Transfer 把卖家余额扣两次。
func TestTradeOrderAftersaleApproveIsAtomic(t *testing.T) {
	ctx := context.Background()
	r := memory.NewTradeOrderRepository()
	if _, err := r.Create(ctx, domain.TradeOrder{
		ID: "to-cas-1", BuyerID: "u-buyer", SellerID: "u-seller",
		AmountFen: 100, Status: "aftersale", AftersaleStatus: "pending",
	}); err != nil {
		t.Fatalf("create: %v", err)
	}

	// 两个审批者各自基于同一份 "pending" 快照构造写入
	first := domain.TradeOrder{ID: "to-cas-1", Status: "completed", AftersaleStatus: "approved"}
	second := domain.TradeOrder{ID: "to-cas-1", Status: "completed", AftersaleStatus: "approved"}

	if _, err := r.UpdateAftersale(ctx, first, "pending"); err != nil {
		t.Fatalf("第一个审批应成功: %v", err)
	}
	if _, err := r.UpdateAftersale(ctx, second, "pending"); err == nil {
		t.Fatal("第二个并发审批必须被 CAS 拒绝（否则卖家余额会被扣两次）")
	}

	got, err := r.FindByID(ctx, "to-cas-1")
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if got.AftersaleStatus != "approved" || got.Status != "completed" {
		t.Fatalf("售后状态应停在首个成功结果，实际 status=%s aftersale=%s", got.Status, got.AftersaleStatus)
	}
}
