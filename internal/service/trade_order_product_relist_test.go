package service_test

import (
	"testing"
	"time"

	"drone-platform/internal/service"
)

// 取消订单后商品必须重新上架（回到可售、重新出现在供给大厅）。
// 这条链路此前只在订单侧做恢复：只要那条路没走成，商品就永久留在 sold ——
// 大厅不展示、详情 404、卖家"我的商品"里也看不到，即用户报的"商品消失了"。
func TestCancelOrderRestoresProductToSale(t *testing.T) {
	tradeSvc, prodRepo, escrowSvc, ctx := newMoneyFlow(t)
	productID := seedProduct(t, ctx, prodRepo, "seller-1", 10000)
	if _, err := escrowSvc.Deposit(ctx, "buyer-1", 10000); err != nil {
		t.Fatalf("充值: %v", err)
	}
	o, err := tradeSvc.Create(ctx, "buyer-1", productID, "seller-1", 10000, service.OrderReceiver{}, false)
	if err != nil {
		t.Fatalf("下单: %v", err)
	}
	// 复现 handler 的下单占位（service.Create 本身不动商品状态）
	if err := prodRepo.MarkSold(ctx, productID); err != nil {
		t.Fatalf("占位: %v", err)
	}
	if p, _ := prodRepo.FindByID(ctx, productID); p.Status != "sold" {
		t.Fatalf("占位后应为 sold，实际 %s", p.Status)
	}

	if _, err := tradeSvc.UpdateStatus(ctx, o.ID, "buyer-1", "cancelled"); err != nil {
		t.Fatalf("取消订单: %v", err)
	}
	p, err := prodRepo.FindByID(ctx, productID)
	if err != nil {
		t.Fatalf("读商品: %v", err)
	}
	if p.Status != "listed" {
		t.Fatalf("取消订单后商品应重新上架(listed)，实际 %q", p.Status)
	}
}

// 孤儿已售商品回收：sold 但已无任何有效订单的商品会被重新上架；
// 仍有未取消订单的商品绝不能被动到。
func TestRelistOrphanSoldProducts(t *testing.T) {
	// 场景一：商品被占位后订单根本没落库（创建失败/进程崩溃）→ 必须回收
	t.Run("无订单的已售商品应重新上架", func(t *testing.T) {
		tradeSvc, prodRepo, _, ctx := newMoneyFlow(t)
		productID := seedProduct(t, ctx, prodRepo, "seller-1", 10000)
		if err := prodRepo.MarkSold(ctx, productID); err != nil {
			t.Fatalf("占位: %v", err)
		}

		n, err := tradeSvc.RelistOrphanSoldProducts(ctx, time.Now().Add(time.Hour), 100)
		if err != nil || n != 1 {
			t.Fatalf("应回收 1 件，实际 n=%d err=%v", n, err)
		}
		if p, _ := prodRepo.FindByID(ctx, productID); p.Status != "listed" {
			t.Fatalf("应回到 listed，实际 %s", p.Status)
		}
	})

	// 场景二：商品确实处于交易中（订单未取消）→ 绝不能上架（否则一物两卖）
	t.Run("仍有未取消订单的商品不得上架", func(t *testing.T) {
		tradeSvc, prodRepo, escrowSvc, ctx := newMoneyFlow(t)
		productID := seedProduct(t, ctx, prodRepo, "seller-1", 10000)
		if _, err := escrowSvc.Deposit(ctx, "buyer-1", 10000); err != nil {
			t.Fatalf("充值: %v", err)
		}
		o, err := tradeSvc.Create(ctx, "buyer-1", productID, "seller-1", 10000, service.OrderReceiver{}, false)
		if err != nil {
			t.Fatalf("下单: %v", err)
		}
		if err := prodRepo.MarkSold(ctx, productID); err != nil {
			t.Fatalf("占位: %v", err)
		}
		if _, err := tradeSvc.PayOrder(ctx, "buyer-1", o.ID); err != nil {
			t.Fatalf("付款: %v", err)
		}
		if p, _ := prodRepo.FindByID(ctx, productID); p.Status != "sold" {
			t.Fatalf("前置应为 sold，实际 %s", p.Status)
		}

		n, err := tradeSvc.RelistOrphanSoldProducts(ctx, time.Now().Add(time.Hour), 100)
		if err != nil {
			t.Fatalf("回收失败: %v", err)
		}
		if n != 0 {
			t.Fatalf("有有效订单时不应回收，实际 n=%d", n)
		}
		if p, _ := prodRepo.FindByID(ctx, productID); p.Status != "sold" {
			t.Fatalf("应保持 sold，实际 %s", p.Status)
		}
	})

	// 场景三：订单已取消但商品没被恢复（正是用户报的现象）→ 兜底回收
	t.Run("仅剩已取消订单的商品应重新上架", func(t *testing.T) {
		tradeSvc, prodRepo, escrowSvc, ctx := newMoneyFlow(t)
		productID := seedProduct(t, ctx, prodRepo, "seller-1", 10000)
		if _, err := escrowSvc.Deposit(ctx, "buyer-1", 10000); err != nil {
			t.Fatalf("充值: %v", err)
		}
		o, err := tradeSvc.Create(ctx, "buyer-1", productID, "seller-1", 10000, service.OrderReceiver{}, false)
		if err != nil {
			t.Fatalf("下单: %v", err)
		}
		if err := prodRepo.MarkSold(ctx, productID); err != nil {
			t.Fatalf("占位: %v", err)
		}
		// 正常取消（服务层会顺带恢复商品），随后再手工占位回去，
		// 复现"订单已取消、商品却仍停在 sold"的残留状态。
		if _, err := tradeSvc.UpdateStatus(ctx, o.ID, "buyer-1", "cancelled"); err != nil {
			t.Fatalf("取消: %v", err)
		}
		if err := prodRepo.MarkSold(ctx, productID); err != nil {
			t.Fatalf("再次占位: %v", err)
		}
		if p, _ := prodRepo.FindByID(ctx, productID); p.Status != "sold" {
			t.Fatalf("前置应为 sold，实际 %s", p.Status)
		}

		n, err := tradeSvc.RelistOrphanSoldProducts(ctx, time.Now().Add(time.Hour), 100)
		if err != nil || n != 1 {
			t.Fatalf("应回收 1 件，实际 n=%d err=%v", n, err)
		}
		if p, _ := prodRepo.FindByID(ctx, productID); p.Status != "listed" {
			t.Fatalf("应回到 listed，实际 %s", p.Status)
		}
	})

	// 场景四：宽限期内的已售商品不得被回收（避免与"占位→建单"的正常时序竞争）
	t.Run("宽限期内不回收", func(t *testing.T) {
		tradeSvc, prodRepo, _, ctx := newMoneyFlow(t)
		productID := seedProduct(t, ctx, prodRepo, "seller-1", 10000)
		if err := prodRepo.MarkSold(ctx, productID); err != nil {
			t.Fatalf("占位: %v", err)
		}

		n, err := tradeSvc.RelistOrphanSoldProducts(ctx, time.Now().Add(-time.Hour), 100)
		if err != nil || n != 0 {
			t.Fatalf("宽限期内不应回收，实际 n=%d err=%v", n, err)
		}
		if p, _ := prodRepo.FindByID(ctx, productID); p.Status != "sold" {
			t.Fatalf("应保持 sold，实际 %s", p.Status)
		}
	})
}
