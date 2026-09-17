package service_test

import (
	"context"
	"strings"
	"testing"

	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

// Release 收款人守卫（生产装配）：收款 ID 非空但**用户不存在**时必须拒绝。
//
// 回归背景：仓储层 Release 是 upsert，收款人不存在时会凭空建出一个谁也登不上的
// 账户——钱等于蒸发，对账时只看到一笔「正常」的放款流水。空串已被 Release 自身
// fail-closed，这里补的是「非空但悬空」那一类（历史课程 org_id 指向已删除账号）。
//
// 注意守卫校验的是用户是否存在，不是收款方有没有托管账户——
// 真实用户首次收款自动开户的语义不变。
func TestReleaseRejectsNonexistentRecipient(t *testing.T) {
	ctx := context.Background()
	svc := service.NewEscrowService(memory.NewEscrowRepository())
	known := map[string]bool{"seller-ok": true}
	svc.SetRecipientGuard(func(_ context.Context, id string) (bool, error) { return known[id], nil })

	if _, err := svc.Deposit(ctx, "buyer", 1000); err != nil {
		t.Fatalf("deposit: %v", err)
	}
	if _, err := svc.Freeze(ctx, "buyer", 800, "training_course", "c-1"); err != nil {
		t.Fatalf("freeze: %v", err)
	}

	// 收款人不存在 → 拒绝，且资金一分未动、不给幽灵开户
	_, err := svc.Release(ctx, "buyer", "ghost-seller", 800, "training_course", "c-1")
	if err == nil {
		t.Fatal("收款人不存在时必须拒绝 release")
	}
	if !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("错误信息应说明收款人不存在，实际：%v", err)
	}
	b, _ := svc.Balance(ctx, "buyer")
	if b.BalanceFen != 200 || b.FrozenFen != 800 {
		t.Fatalf("被拒的 release 不该动资金：balance=%d frozen=%d want 200/800", b.BalanceFen, b.FrozenFen)
	}
	if ghost, _ := svc.Balance(ctx, "ghost-seller"); ghost.BalanceFen != 0 || ghost.FrozenFen != 0 {
		t.Fatalf("不该为不存在的收款人开户，实际 balance=%d frozen=%d", ghost.BalanceFen, ghost.FrozenFen)
	}

	// 真实用户（即使还没有托管账户）→ 放行并自动开户
	if _, err := svc.Release(ctx, "buyer", "seller-ok", 800, "training_course", "c-1"); err != nil {
		t.Fatalf("真实收款人应可 release: %v", err)
	}
	ok, _ := svc.Balance(ctx, "seller-ok")
	if ok.BalanceFen != 800 {
		t.Fatalf("seller-ok balance=%d want 800", ok.BalanceFen)
	}

	// 守卫查询失败必须 fail-closed（宁可拒绝放款，也不能把钱打给一个查不实的人）
	broken := service.NewEscrowService(memory.NewEscrowRepository())
	broken.SetRecipientGuard(func(_ context.Context, _ string) (bool, error) {
		return false, context.DeadlineExceeded
	})
	broken.Deposit(ctx, "buyer", 1000)
	broken.Freeze(ctx, "buyer", 800, "training_course", "c-2")
	if _, err := broken.Release(ctx, "buyer", "seller-ok", 800, "training_course", "c-2"); err == nil {
		t.Fatal("守卫查询失败时必须 fail-closed")
	}
}
