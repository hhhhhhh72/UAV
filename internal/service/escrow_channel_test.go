package service_test

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

// 真实资金接入的资金侧契约：渠道标记 + 外部支付单号幂等 + 按渠道对账。
//
// 背景：资金此前全是内部记账（没有来源标记），接微信支付时最危险的两件事是
// "回调重试导致重复入账（印钞）"和"分不清账上的钱哪些是真钱"。这组用例把这两条钉死。
func TestDepositFromChannelRealFundsIdempotent(t *testing.T) {
	ctx := context.Background()
	e := service.NewEscrowService(memory.NewEscrowRepository())

	// 微信支付回调首次入账：100 元
	tx1, created, err := e.DepositFromChannel(ctx, "buyer-1", 10000, domain.ChannelWeChat, "wx-txn-0001")
	if err != nil || !created {
		t.Fatalf("首次入账应成功且 created=true，实际 created=%v err=%v", created, err)
	}
	if tx1.Channel != domain.ChannelWeChat || tx1.ExternalTxnID != "wx-txn-0001" {
		t.Fatalf("流水应带渠道与外部单号，实际 channel=%s external=%s", tx1.Channel, tx1.ExternalTxnID)
	}

	// 微信回调重试（同一 transaction_id）：不得再加一次钱
	tx2, created2, err := e.DepositFromChannel(ctx, "buyer-1", 10000, domain.ChannelWeChat, "wx-txn-0001")
	if err != nil {
		t.Fatalf("重复回调应幂等成功，实际 %v", err)
	}
	if created2 {
		t.Fatal("重复回调不得再次入账（created 应为 false）")
	}
	if tx2.ID != tx1.ID {
		t.Fatalf("重复回调应返回原流水，实际新流水 %s vs %s", tx2.ID, tx1.ID)
	}
	acc, err := e.Balance(ctx, "buyer-1")
	if err != nil {
		t.Fatalf("查余额: %v", err)
	}
	if acc.BalanceFen != 10000 {
		t.Fatalf("余额应为 10000（只入账一次），实际 %d", acc.BalanceFen)
	}
}

// 真实渠道必须有外部支付单号：没有凭证的"微信入账"是凭空造钱，直接拒绝。
func TestDepositFromChannelRealChannelNeedsExternalID(t *testing.T) {
	ctx := context.Background()
	e := service.NewEscrowService(memory.NewEscrowRepository())

	if _, _, err := e.DepositFromChannel(ctx, "buyer-1", 10000, domain.ChannelWeChat, ""); err == nil {
		t.Fatal("真实渠道缺少外部支付单号应被拒绝")
	}
	acc, _ := e.Balance(ctx, "buyer-1")
	if acc.BalanceFen != 0 {
		t.Fatalf("被拒绝的入账不得动钱，实际余额 %d", acc.BalanceFen)
	}

	// 内部记账渠道不需要外部单号（管理员线下收款后补记）
	if _, created, err := e.DepositFromChannel(ctx, "buyer-1", 5000, domain.ChannelInternalAdmin, ""); err != nil || !created {
		t.Fatalf("内部渠道入账应成功，实际 created=%v err=%v", created, err)
	}
	if acc, _ = e.Balance(ctx, "buyer-1"); acc.BalanceFen != 5000 {
		t.Fatalf("内部渠道入账应到账 5000，实际 %d", acc.BalanceFen)
	}
}

// 并发回调：同一支付单号只允许入账一次（PG 由唯一索引兜底，内存实现同规则）。
func TestDepositFromChannelConcurrentCallbackCreditsOnce(t *testing.T) {
	ctx := context.Background()
	e := service.NewEscrowService(memory.NewEscrowRepository())

	var wg sync.WaitGroup
	createdCount := 0
	var mu sync.Mutex
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, created, err := e.DepositFromChannel(ctx, "buyer-1", 10000, domain.ChannelWeChat, "wx-txn-concurrent")
			if err != nil {
				t.Errorf("并发回调不应报错: %v", err)
				return
			}
			if created {
				mu.Lock()
				createdCount++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	if createdCount != 1 {
		t.Fatalf("同一支付单号只应有 1 次真正入账，实际 %d 次", createdCount)
	}
	acc, _ := e.Balance(ctx, "buyer-1")
	if acc.BalanceFen != 10000 {
		t.Fatalf("并发回调后余额应为 10000，实际 %d", acc.BalanceFen)
	}
}

// 对账：按渠道 + 区间汇总入金，真实资金与内部记账分得清。
func TestReconcileByChannel(t *testing.T) {
	ctx := context.Background()
	e := service.NewEscrowService(memory.NewEscrowRepository())

	if _, _, err := e.DepositFromChannel(ctx, "u1", 10000, domain.ChannelWeChat, "wx-1"); err != nil {
		t.Fatalf("入账: %v", err)
	}
	if _, _, err := e.DepositFromChannel(ctx, "u2", 25000, domain.ChannelWeChat, "wx-2"); err != nil {
		t.Fatalf("入账: %v", err)
	}
	if _, _, err := e.DepositFromChannel(ctx, "u3", 7000, domain.ChannelInternalSelf, ""); err != nil {
		t.Fatalf("入账: %v", err)
	}

	wechat, err := e.Reconcile(ctx, domain.ChannelWeChat, time.Time{}, time.Time{}, 0)
	if err != nil {
		t.Fatalf("对账: %v", err)
	}
	if !wechat.RealFunds {
		t.Fatal("wechat 应为真实资金渠道")
	}
	if wechat.DepositCount != 2 || wechat.DepositFen != 35000 {
		t.Fatalf("微信入金应为 2 笔 35000 分，实际 %d 笔 %d 分", wechat.DepositCount, wechat.DepositFen)
	}

	internal, err := e.Reconcile(ctx, domain.ChannelInternalSelf, time.Time{}, time.Time{}, 0)
	if err != nil {
		t.Fatalf("对账: %v", err)
	}
	if internal.RealFunds {
		t.Fatal("internal_self 不是真实资金渠道")
	}
	if internal.DepositCount != 1 || internal.DepositFen != 7000 {
		t.Fatalf("自助充值应为 1 笔 7000 分，实际 %d 笔 %d 分", internal.DepositCount, internal.DepositFen)
	}

	// 不限渠道：三笔全在，合计 42000
	all, err := e.Reconcile(ctx, "", time.Time{}, time.Time{}, 0)
	if err != nil {
		t.Fatalf("对账: %v", err)
	}
	if all.DepositCount != 3 || all.DepositFen != 42000 {
		t.Fatalf("全渠道应为 3 笔 42000 分，实际 %d 笔 %d 分", all.DepositCount, all.DepositFen)
	}

	// 区间过滤：未来区间应为空（证明 from/to 真的生效，而不是把全部流水都算进来）
	future := time.Now().Add(time.Hour)
	empty, err := e.Reconcile(ctx, "", future, time.Time{}, 0)
	if err != nil {
		t.Fatalf("对账: %v", err)
	}
	if empty.DepositCount != 0 || len(empty.Transactions) != 0 {
		t.Fatalf("未来区间应无流水，实际 %d 笔", empty.DepositCount)
	}
}

// 渠道标记语义：internal* 都是平台自己记的账，不是真钱。
func TestIsRealChannelSemantics(t *testing.T) {
	for _, ch := range []string{domain.ChannelInternal, domain.ChannelInternalAdmin, domain.ChannelInternalSelf, ""} {
		if domain.IsRealChannel(ch) {
			t.Fatalf("%q 不应被判定为真实资金渠道", ch)
		}
	}
	if !domain.IsRealChannel(domain.ChannelWeChat) {
		t.Fatal("wechat 应被判定为真实资金渠道")
	}
	// 新增渠道（如支付宝）自动视为真实资金，避免漏标
	if !domain.IsRealChannel("alipay") {
		t.Fatal("未知非 internal 渠道应保守视为真实资金渠道")
	}
	if !strings.HasPrefix(domain.ChannelInternalAdmin, "internal") {
		t.Fatal("内部渠道命名必须带 internal 前缀（判定规则依赖它）")
	}
}
