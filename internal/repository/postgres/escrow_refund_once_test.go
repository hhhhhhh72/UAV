package postgres_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"drone-platform/internal/domain"
)

// TestPG_EscrowRefundIsOncePerTradeOrder 回归售后「确认收到退货」的重复退款。
//
// ConfirmReturnReceived 是**先退款再改状态**的（钱动不了就保持 returned 供重试），
// 所以并发确认时两边都会读到 AftersaleStatus='returned'，并双双通过
// refundForAftersale 的 HasFrozen + HasRefunded 两道 check-then-act 查询。
//
// 场景同样构造成「就算没有库级约束也一定会重复退款」：买家先冻结 n 笔各 amount
// （total = n*amount），于是单看 frozen_fen >= amount 的条件更新，对同一订单
// 连退 n 次款都会成功。后果不是凭空生钱，而是把同一笔冻结款重复退回去——
// 其余仍处于「已付款冻结」的订单再也释放不出来（release 会因 frozen 不足失败）。
//
// 直接调仓储层，绕开 service 的查询，测的就是库那一层。
func TestPG_EscrowRefundIsOncePerTradeOrder(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	ctx := context.Background()
	repo := store.NewEscrowRepository()

	const amount int64 = 1000
	const n = 8
	buyer := ug("buyer")
	orderID := ug("torder")
	now := time.Now()

	mkTx := func(kind, refType, ref string, amt int64, from, to string) domain.EscrowTransaction {
		return domain.EscrowTransaction{
			ID: ug("tx"), FromUser: from, ToUser: to, AmountFen: amt, TxType: kind,
			ReferenceType: refType, ReferenceID: ref,
			Status: "completed", CreatedAt: now,
		}
	}

	if _, err := repo.Deposit(ctx, buyer, amount*n, mkTx("deposit", "", "", amount*n, "system", buyer)); err != nil {
		t.Fatalf("deposit: %v", err)
	}
	// n 笔彼此独立的订单冻结，制造「还有其它冻结资金」的局面。
	for i := 0; i < n; i++ {
		if _, err := repo.Freeze(ctx, buyer, amount, mkTx("freeze", "trade_order", ug("other-order"), amount, buyer, "escrow")); err != nil {
			t.Fatalf("freeze %d: %v", i, err)
		}
	}
	acc, err := repo.GetAccount(ctx, buyer)
	if err != nil {
		t.Fatalf("get buyer account: %v", err)
	}
	if acc.FrozenFen != amount*n {
		t.Fatalf("前置条件不成立：买家冻结应为 %d，实际 %d", amount*n, acc.FrozenFen)
	}

	// 并发对**同一个订单**退款 n 次。
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, _ = repo.Refund(ctx, buyer, amount, mkTx("refund", "trade_order", orderID, amount, "escrow", buyer))
		}(i)
	}
	wg.Wait()

	// 1) 流水只有一条。
	var txCount int
	if err := store.Pool().QueryRow(ctx,
		`SELECT count(*) FROM escrow_transactions WHERE tx_type='refund' AND reference_type='trade_order' AND reference_id=$1`,
		orderID).Scan(&txCount); err != nil {
		t.Fatalf("count refund tx: %v", err)
	}
	if txCount != 1 {
		t.Fatalf("同一订单的 refund 流水应为 1 条，实际 %d 条——库级唯一约束没生效，买家被重复退款", txCount)
	}

	// 2) 买家只该被退回一份：冻结剩 (n-1)*amount，可用余额恰为 amount。
	after, err := repo.GetAccount(ctx, buyer)
	if err != nil {
		t.Fatalf("get buyer account: %v", err)
	}
	if after.BalanceFen != amount {
		t.Fatalf("买家可用余额应为 %d（只退一份），实际 %d——重复退款", amount, after.BalanceFen)
	}
	if want := amount * (n - 1); after.FrozenFen != want {
		t.Fatalf("买家冻结应为 %d（只扣一笔），实际 %d", want, after.FrozenFen)
	}
}