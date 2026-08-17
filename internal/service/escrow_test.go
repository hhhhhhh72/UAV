package service_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"drone-platform/internal/repository"
	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

// TestEscrowRejectsNonPositiveAmounts: 回归 C6——负/零金额一律拒绝，
// 旧实现 Freeze(-100) 会反向增加余额。
func TestEscrowRejectsNonPositiveAmounts(t *testing.T) {
	svc := service.NewEscrowService(memory.NewEscrowRepository())
	if _, err := svc.Deposit(context.Background(), "u", -1); err == nil {
		t.Fatal("negative deposit accepted")
	}
	if _, err := svc.Deposit(context.Background(), "u", 0); err == nil {
		t.Fatal("zero deposit accepted")
	}
	if _, err := svc.Freeze(context.Background(), "u", -100, "x", "x"); err == nil {
		t.Fatal("negative freeze accepted")
	}
	if _, err := svc.Release(context.Background(), "u", "v", -100, "x", "x"); err == nil {
		t.Fatal("negative release accepted")
	}
	if _, err := svc.Refund(context.Background(), "u", -100, "x", "x"); err == nil {
		t.Fatal("negative refund accepted")
	}
}

// TestEscrowFreezeConcurrent: 回归 C6——并发冻结不得丢更新。
// 入金 1000 分后 25 个 goroutine 各冻结 50 分：恰好 20 个成功，
// 其余收到 ErrInsufficientBalance，最终余额 0 / 冻结 1000。
func TestEscrowFreezeConcurrent(t *testing.T) {
	svc := service.NewEscrowService(memory.NewEscrowRepository())
	if _, err := svc.Deposit(context.Background(), "u", 1000); err != nil {
		t.Fatalf("deposit: %v", err)
	}

	const workers = 25
	var (
		ok, insufficient int
		mu               sync.Mutex
		wg               sync.WaitGroup
	)
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			_, err := svc.Freeze(context.Background(), "u", 50, "course", "c-1")
			mu.Lock()
			defer mu.Unlock()
			switch {
			case err == nil:
				ok++
			case errors.Is(err, repository.ErrInsufficientBalance):
				insufficient++
			default:
				t.Errorf("unexpected error: %v", err)
			}
		}()
	}
	wg.Wait()

	if ok != 20 || insufficient != 5 {
		t.Fatalf("expected 20 ok + 5 insufficient, got %d + %d", ok, insufficient)
	}
	acct, err := svc.Balance(context.Background(), "u")
	if err != nil {
		t.Fatalf("balance: %v", err)
	}
	if acct.BalanceFen != 0 || acct.FrozenFen != 1000 {
		t.Fatalf("final account: balance=%d frozen=%d, want 0/1000", acct.BalanceFen, acct.FrozenFen)
	}
}

// TestEscrowReleaseAtomicity: 回归 C6——解冻=付款方扣减+收款方入账+流水，三者在
// 同一临界区/事务内完成；收款方无账户时自动开户。
func TestEscrowReleaseAtomicity(t *testing.T) {
	svc := service.NewEscrowService(memory.NewEscrowRepository())
	svc.Deposit(context.Background(), "buyer", 1000)
	svc.Freeze(context.Background(), "buyer", 800, "course", "c-1")

	// 收款方 seller 从未开过户：Release 应自动创建并正确入账
	if _, err := svc.Release(context.Background(), "buyer", "seller", 800, "course", "c-1"); err != nil {
		t.Fatalf("release: %v", err)
	}
	seller, _ := svc.Balance(context.Background(), "seller")
	if seller.BalanceFen != 800 {
		t.Fatalf("seller balance: %d, want 800", seller.BalanceFen)
	}
	buyer, _ := svc.Balance(context.Background(), "buyer")
	if buyer.FrozenFen != 0 || buyer.BalanceFen != 200 {
		t.Fatalf("buyer after release: balance=%d frozen=%d", buyer.BalanceFen, buyer.FrozenFen)
	}

	// 冻结不足的解冻必须被拒绝且账户不变
	if _, err := svc.Release(context.Background(), "buyer", "seller", 1, "course", "c-1"); !errors.Is(err, repository.ErrInsufficientFrozenBalance) {
		t.Fatalf("expected ErrInsufficientFrozenBalance, got %v", err)
	}
	seller2, _ := svc.Balance(context.Background(), "seller")
	if seller2.BalanceFen != 800 {
		t.Fatalf("seller balance changed after failed release: %d", seller2.BalanceFen)
	}

	// 流水完整：deposit + freeze + release 共 3 条
	txs, _ := svc.Transactions(context.Background(), "buyer")
	if len(txs) != 3 {
		t.Fatalf("expected 3 transactions, got %d", len(txs))
	}
}
