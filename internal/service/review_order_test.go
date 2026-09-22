package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

// 商城订单评价闭环（2026-09-22，用户报「商城订单评价完后为什么订单状态还是待评价」）。
//
// 两件事一起验：
//  ① 「已评价」这个事实必须由**服务端**回答。此前订单状态机的 completed 在小程序里被
//     展示成「待评价」，而"已评价"只写在手机本地存储（orderAdapter.js 的
//     order_reviewed_prod）—— 换设备、清缓存、或在开发者工具里评价而用真机看，
//     就立刻回到「待评价」。reviews 表里明明有记录，订单列表却拿不到。
//  ② order 这一类评价此前**完全没有校验**（只有 work_order 有）：不查订单存不存在、
//     不查是不是本人买的、不查订单是否完成 —— 任何登录用户都能对任意 target_id 造评价。
func TestReviewOrderValidationAndReviewedFlag(t *testing.T) {
	ctx := context.Background()
	orderRepo := memory.NewTradeOrderRepository()
	reviewRepo := memory.NewReviewRepository()

	// 直接种订单（不走"下单→支付→发货→收货"整条链，与本用例无关）
	seed := func(id, buyer, status string) {
		t.Helper()
		if _, err := orderRepo.Create(ctx, domain.TradeOrder{
			ID: id, ProductID: "prod-1", BuyerID: buyer, SellerID: "seller-1",
			AmountFen: 10000, Status: status, CreatedAt: time.Now(),
		}); err != nil {
			t.Fatalf("seed order %s: %v", id, err)
		}
	}
	seed("torder-done", "buyer-1", "completed")
	seed("torder-shipped", "buyer-1", "shipped")

	rv := service.NewReviewService(reviewRepo, memory.NewWorkOrderRepository(), orderRepo)

	// ① 不存在的订单（此前会 201 落一条孤儿评价）
	if _, err := rv.Submit(ctx, "buyer-1", "order", "torder-nope", 5, "x"); !errors.Is(err, service.ErrReviewTargetNotFound) {
		t.Fatalf("不存在的订单应 ErrReviewTargetNotFound，实际 %v", err)
	}
	// ② 不是买家
	if _, err := rv.Submit(ctx, "stranger", "order", "torder-done", 5, "刷好评"); !errors.Is(err, service.ErrReviewNotAllowed) {
		t.Fatalf("非买家应 ErrReviewNotAllowed，实际 %v", err)
	}
	// ③ 卖家也不能评价自己的订单（否则是刷自己好评的通道）
	if _, err := rv.Submit(ctx, "seller-1", "order", "torder-done", 5, "自评"); !errors.Is(err, service.ErrReviewNotAllowed) {
		t.Fatalf("卖家应 ErrReviewNotAllowed，实际 %v", err)
	}
	// ④ 还没完成的订单
	if _, err := rv.Submit(ctx, "buyer-1", "order", "torder-shipped", 5, "还没收到就评"); !errors.Is(err, service.ErrReviewTargetNotReady) {
		t.Fatalf("未完成订单应 ErrReviewTargetNotReady，实际 %v", err)
	}

	tradeSvc := service.NewTradeOrderService(orderRepo, memory.NewProductRepository(), reviewRepo)

	reviewedOf := func(user, orderID string) bool {
		t.Helper()
		orders, err := tradeSvc.ListMine(ctx, user)
		if err != nil {
			t.Fatalf("ListMine(%s): %v", user, err)
		}
		for _, o := range orders {
			if o.ID == orderID {
				return o.Reviewed
			}
		}
		t.Fatalf("%s 的订单列表里没有 %s", user, orderID)
		return false
	}

	// ⑤ 评价前：reviewed=false —— 这就是用户看到的「待评价」
	if reviewedOf("buyer-1", "torder-done") {
		t.Fatal("评价前 reviewed 就为 true？")
	}

	// ⑥ 买家评价已完成订单 → 成功
	rev, err := rv.Submit(ctx, "buyer-1", "order", "torder-done", 5, "非常可以")
	if err != nil {
		t.Fatalf("买家评价已完成订单应成功，实际 %v", err)
	}

	// ⑦ 评价后：被评价的那一单为 true，**同一用户另一单不受影响**
	if !reviewedOf("buyer-1", "torder-done") {
		t.Fatal("评价后 reviewed 仍为 false —— 这正是「评价完还是待评价」的根因")
	}
	if reviewedOf("buyer-1", "torder-shipped") {
		t.Fatal("没评价过的订单被标成了已评价")
	}

	// ⑧ 口径是"**当前请求者**是否评价过"，不是"这一单有没有人评价过"
	seed("torder-other", "buyer-2", "completed")
	if reviewedOf("buyer-2", "torder-other") {
		t.Fatal("别人评价过不影响我的 reviewed")
	}

	// ⑨ 被驳回的评价不算已评价，且允许重评（口径与 Submit 的幂等判定一致）
	if _, err := reviewRepo.UpdateStatus(ctx, rev.ID, "rejected"); err != nil {
		t.Fatalf("reject review: %v", err)
	}
	if reviewedOf("buyer-1", "torder-done") {
		t.Fatal("评价被驳回后不应再算已评价 —— 否则用户永远无法重评")
	}
	if _, err := rv.Submit(ctx, "buyer-1", "order", "torder-done", 4, "重评一次"); err != nil {
		t.Fatalf("驳回后应可重新评价，实际 %v", err)
	}
	if !reviewedOf("buyer-1", "torder-done") {
		t.Fatal("重评后应恢复为已评价")
	}
}
