package service_test

import (
	"context"
	"strings"
	"testing"

	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

// 卖家不存在时，订单付款必须在**冻结之前**被拒。
//
// 回归背景：生产上有多件演示商品的 seller_id 是不存在的种子 ID（u-shop-1 之类）。
// 订单是「先冻结买家、确认收货才放款给卖家」，而放款侧有收款人守卫会 fail-closed
// 拒付——买家付了钱、拿不到货、也拿不回钱，资金永久卡死在冻结里。
// 宁可下单就失败，也不要把买家的钱冻进去。
func TestPayOrderRejectsNonexistentSellerBeforeFreezing(t *testing.T) {
	ctx := context.Background()
	escrowSvc := service.NewEscrowService(memory.NewEscrowRepository())
	// 只认 seller-real；下面那件商品挂的是 u-shop-1，不在其中
	escrowSvc.SetRecipientGuard(func(_ context.Context, id string) (bool, error) {
		return id == "seller-real", nil
	})
	prodRepo := memory.NewProductRepository()
	tradeSvc := service.NewTradeOrderService(memory.NewTradeOrderRepository(), prodRepo)
	tradeSvc.SetEscrow(escrowSvc)

	if _, err := escrowSvc.Deposit(ctx, "buyer-1", 500000); err != nil {
		t.Fatalf("充值: %v", err)
	}

	ghost := seedProduct(t, ctx, prodRepo, "u-shop-1", 100000)
	o, err := tradeSvc.Create(ctx, "buyer-1", ghost, "u-shop-1", 100000, service.OrderReceiver{}, false)
	if err != nil {
		t.Fatalf("建订单: %v", err)
	}

	if _, err := tradeSvc.PayOrder(ctx, "buyer-1", o.ID); err == nil {
		t.Fatal("卖家不存在时不该允许付款")
	} else if !strings.Contains(err.Error(), "不存在") {
		t.Fatalf("错误信息应说明原因，实际: %v", err)
	}

	// 关键：一分钱都不能被冻进去
	if b, f := balanceOf(t, ctx, escrowSvc, "buyer-1"); b != 500000 || f != 0 {
		t.Fatalf("被拒的付款不该动资金：balance=%d frozen=%d want 500000/0", b, f)
	}

	// 卖家存在时照常放行——护栏不能误伤
	real := seedProduct(t, ctx, prodRepo, "seller-real", 100000)
	o2, err := tradeSvc.Create(ctx, "buyer-1", real, "seller-real", 100000, service.OrderReceiver{}, false)
	if err != nil {
		t.Fatalf("建订单2: %v", err)
	}
	if _, err := tradeSvc.PayOrder(ctx, "buyer-1", o2.ID); err != nil {
		t.Fatalf("真实卖家应可付款: %v", err)
	}
	if b, f := balanceOf(t, ctx, escrowSvc, "buyer-1"); b != 400000 || f != 100000 {
		t.Fatalf("付款后应冻结 100000：balance=%d frozen=%d want 400000/100000", b, f)
	}
}
