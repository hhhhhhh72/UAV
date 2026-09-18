package postgres_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)

// TestPG_EscrowReleaseIsOncePerReference 回归「同一付款方对同一业务单只能放款一次」这条资金不变量**由数据库保证**。
//
// 场景刻意构造成「就算没有库级约束、应用层也没拦，也一定会重复放款」：
// 付款方先冻结 8 笔各 amount 的资金（total = 8*amount），于是单看余额条件，
// 对同一笔业务单连放 8 次款都会成功——每次都能通过 frozen_fen >= amount 的条件更新。
//
// 修复前：escrow_transactions 上唯一的唯一索引是 idx_escrow_external
// （WHERE external_txn_id <> ''），而 release 走内部渠道、external_txn_id 恒为空串，
// 这条索引完全不生效；唯一的防线是 EscrowService.Release 里那句
// 「先查 HasReleased 再落库」——两个并发请求可以同时通过查询。
// 修复后：migration 000116 的 idx_escrow_once_per_ref 让第二次插入带着整个事务回滚。
//
// 这里**直接调仓储层**（绕过 service 的 HasReleased 短路），测的就是库那一层。
func TestPG_EscrowReleaseIsOncePerReference(t *testing.T) {
	store := setupStore(t)
	if store == nil {
		return
	}
	ctx := context.Background()
	repo := store.NewEscrowRepository()

	const amount int64 = 1000
	const n = 8
	buyer, seller := ug("buyer"), ug("seller")
	refID := ug("course")
	now := time.Now()

	mkTx := func(kind, ref string, amt int64, from, to string) domain.EscrowTransaction {
		return domain.EscrowTransaction{
			ID: ug("tx"), FromUser: from, ToUser: to, AmountFen: amt, TxType: kind,
			ReferenceType: "training_course", ReferenceID: ref,
			Status: "completed", CreatedAt: now,
		}
	}

	// 买家备足资金：8 笔独立冻结，业务单彼此不同。
	if _, err := repo.Deposit(ctx, buyer, amount*n, mkTx("deposit", "", amount*n, "system", buyer)); err != nil {
		t.Fatalf("deposit: %v", err)
	}
	for i := 0; i < n; i++ {
		if _, err := repo.Freeze(ctx, buyer, amount, mkTx("freeze", ug("other"), amount, buyer, "escrow")); err != nil {
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

	// 并发对**同一个业务单**放款 n 次。
	var wg sync.WaitGroup
	errs := make([]error, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, errs[i] = repo.Release(ctx, buyer, seller, amount, mkTx("release", refID, amount, buyer, seller))
		}(i)
	}
	wg.Wait()

	// 1) 流水只有一条——唯一索引生效的直接证据。
	var txCount int
	if err := store.Pool().QueryRow(ctx,
		`SELECT count(*) FROM escrow_transactions WHERE from_user=$1 AND tx_type='release' AND reference_id=$2`,
		buyer, refID).Scan(&txCount); err != nil {
		t.Fatalf("count release tx: %v", err)
	}
	if txCount != 1 {
		t.Fatalf("同一业务单的 release 流水应为 1 条，实际 %d 条——库级唯一约束没生效，机构被重复放款", txCount)
	}

	// 2) 收款方只该收到一份。
	sellerAcc, err := repo.GetAccount(ctx, seller)
	if err != nil {
		t.Fatalf("get seller account: %v", err)
	}
	if sellerAcc.BalanceFen != amount {
		t.Fatalf("收款方应只收到 %d，实际 %d——重复放款", amount, sellerAcc.BalanceFen)
	}

	// 3) 付款方的冻结只该被扣掉一份；被索引挡下的那几次必须整体回滚，不能白扣。
	buyerAcc, err := repo.GetAccount(ctx, buyer)
	if err != nil {
		t.Fatalf("get buyer account: %v", err)
	}
	if want := amount * (n - 1); buyerAcc.FrozenFen != want {
		t.Fatalf("买家冻结应为 %d（只扣一次释放），实际 %d", want, buyerAcc.FrozenFen)
	}

	// 4) 幂等：被索引挡下的那些调用不该把错误抛给业务方——它们拿到的是「已放款」。
	for i, err := range errs {
		if err != nil && err != repository.ErrInsufficientFrozenBalance {
			t.Errorf("第 %d 次并发放款返回了非幂等错误：%v", i, err)
		}
	}
}