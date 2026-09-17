package service_test

import (
	"context"
	"testing"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

// 服务能力并入商品表后的适配层语义。
//
// 合并前：同一个"航拍服务"可以走 service_listings（纯展示）或 drone_products（可下单），
// 两条路字段不同、状态机不同。现在数据统一在商品表，本组用例钉住映射不被改回去。
func TestServiceListingAdapterOverProducts(t *testing.T) {
	ctx := context.Background()
	prodRepo := memory.NewProductRepository()
	svc := service.NewServiceListingService(prodRepo)

	// ① 创建：管理端录入默认已上架
	sl, err := svc.CreateListing(ctx, "ent-1", "重庆低空科技", "无人机电力巡检", "巡检", "杆塔巡检", "重庆", 10000, "次", "img.jpg", "")
	if err != nil {
		t.Fatalf("CreateListing: %v", err)
	}
	if sl.Status != "published" {
		t.Fatalf("管理端录入应直接 published，实际 %q", sl.Status)
	}

	// 数据必须真的落在商品表（这才是"合并"的含义）
	p, err := prodRepo.FindByID(ctx, sl.ID)
	if err != nil {
		t.Fatalf("服务能力应落在商品表: %v", err)
	}
	if p.SellerID != "ent-1" || p.SellerName != "重庆低空科技" {
		t.Fatalf("服务商映射错误: seller=%q/%q", p.SellerID, p.SellerName)
	}
	if p.CheckStatus != domain.ProductCheckPassed || p.Status != "listed" {
		t.Fatalf("published 应映射为 check_status=passed + status=listed，实际 %q/%q", p.CheckStatus, p.Status)
	}
	if p.Category != "巡检" || p.Region != "重庆" || p.Unit != "次" {
		t.Fatalf("服务类字段未落库: category=%q region=%q unit=%q", p.Category, p.Region, p.Unit)
	}
	if p.PriceMode != domain.PriceModeFixed {
		t.Fatalf("有报价应为明码标价，实际 %q", p.PriceMode)
	}
	// 分类 → 商品类型按迁移规则推断（"巡检"不在关键词里 → 服务类兜底 repair）
	if p.ProdType != domain.ProductRepair {
		t.Fatalf("巡检应兜底为 repair，实际 %q", p.ProdType)
	}
	// 服务不涉及寄送：下单时不得强制收货地址
	if service.OrderNeedsReceiver(p) {
		t.Fatal("服务能力不应要求收货地址")
	}

	// ② 分类 → prod_type 的关键词推断
	for _, c := range []struct {
		category string
		want     domain.ProductType
	}{
		{"航拍测绘", domain.ProductAerial},
		{"试飞测试", domain.ProductTestFly},
		{"设备检测", domain.ProductCalibration},
		{"空域协调", domain.ProductAirspace},
		{"应急巡检", domain.ProductRepair},
	} {
		got, err := svc.CreateListing(ctx, "ent-1", "服务商", c.category+"服务", c.category, "", "", 5000, "次", "", "")
		if err != nil {
			t.Fatalf("建 %s: %v", c.category, err)
		}
		pp, _ := prodRepo.FindByID(ctx, got.ID)
		if pp.ProdType != c.want {
			t.Fatalf("分类 %q 应推断为 %q，实际 %q", c.category, c.want, pp.ProdType)
		}
	}

	// ③ 零报价 → 面议
	free, _ := svc.CreateListing(ctx, "ent-1", "服务商", "面议服务", "巡检", "", "", 0, "", "", "")
	fp, _ := prodRepo.FindByID(ctx, free.ID)
	if fp.PriceMode != domain.PriceModeNegotiable {
		t.Fatalf("0 报价应映射为面议，实际 %q", fp.PriceMode)
	}

	// ④ 列表隔离：实物商品不得混进服务能力
	if _, err := prodRepo.Create(ctx, domain.DroneProduct{
		ID: "p-drone", SellerID: "s", SellerName: "卖家", ProdType: domain.ProductDrone,
		Title: "整机", CheckStatus: domain.ProductCheckPassed, Status: "listed",
	}); err != nil {
		t.Fatalf("建整机: %v", err)
	}
	pub, err := svc.ListPublished(ctx)
	if err != nil {
		t.Fatalf("ListPublished: %v", err)
	}
	for _, x := range pub {
		if x.ID == "p-drone" {
			t.Fatal("实物商品不得出现在服务能力列表里")
		}
	}
	admin, _ := svc.ListAdmin(ctx, "", "")
	for _, x := range admin {
		if x.ID == "p-drone" {
			t.Fatal("实物商品不得出现在管理端服务能力列表里")
		}
	}

	// ⑤ 按 ID 取实物商品 → 按"不存在"处理，不泄漏
	if _, err := svc.Get(ctx, "p-drone"); err == nil {
		t.Fatal("服务能力端点不应能读到实物商品")
	}

	// ⑥ 用户自助发布 → 待审核（两个维度都进待审）
	pend, err := svc.CreateListingPending(ctx, "ent-2", "另一家", "待审服务", "航拍", "", "", 100, "次", "")
	if err != nil {
		t.Fatalf("CreateListingPending: %v", err)
	}
	if pend.Status != "pending" {
		t.Fatalf("自助发布应为 pending，实际 %q", pend.Status)
	}
	pp2, _ := prodRepo.FindByID(ctx, pend.ID)
	if pp2.CheckStatus != domain.ProductCheckPending || pp2.Status != "pending" {
		t.Fatalf("自助发布两个维度都应待审，实际 %q/%q", pp2.CheckStatus, pp2.Status)
	}
	// 待审的不得出现在公开列表
	pub2, _ := svc.ListPublished(ctx)
	for _, x := range pub2 {
		if x.ID == pend.ID {
			t.Fatal("待审核的服务能力不得公开")
		}
	}

	// ⑥b 关键：status=listed 但**未过审**的商品不得出现在公开服务列表。
	// 拆分审核维度后如果这里只判 status，任何把 status 置成 listed 的路径都会让未审内容上架。
	if _, err := prodRepo.Create(ctx, domain.DroneProduct{
		ID: "sl-unchecked", Title: "未审却在售", ProdType: domain.ProductRepair,
		CheckStatus: domain.ProductCheckPending, Status: "listed",
	}); err != nil {
		t.Fatalf("建未审上架商品: %v", err)
	}
	pub3, _ := svc.ListPublished(ctx)
	for _, x := range pub3 {
		if x.ID == "sl-unchecked" {
			t.Fatal("未过审但 status=listed 的服务不得出现在公开列表")
		}
	}

	// ⑦ 下架：published → offline → removed
	sl.Status = "offline"
	if _, err := svc.UpdateListing(ctx, sl); err != nil {
		t.Fatalf("UpdateListing: %v", err)
	}
	off, _ := prodRepo.FindByID(ctx, sl.ID)
	if off.Status != "removed" {
		t.Fatalf("offline 应映射为 status=removed，实际 %q", off.Status)
	}
	// 审核维度不该被这个入口改掉
	if off.CheckStatus != domain.ProductCheckPassed {
		t.Fatalf("更新服务能力不应改动审核状态，实际 %q", off.CheckStatus)
	}

	// ⑧ 删除走回收站，不是物理删除（订单还要靠这行显示名字）
	if err := svc.DeleteListing(ctx, sl.ID); err != nil {
		t.Fatalf("DeleteListing: %v", err)
	}
	if _, err := prodRepo.FindByID(ctx, sl.ID); err == nil {
		t.Fatal("删除后不应再查得到")
	}
	if del, _ := prodRepo.ListDeleted(ctx); len(del) != 1 {
		t.Fatalf("删除应进回收站，实际 %d 行", len(del))
	}

	// ⑨ 收藏与商品收藏共用一张表
	if err := svc.ToggleFavorite(ctx, "user-1", free.ID, true); err != nil {
		t.Fatalf("ToggleFavorite: %v", err)
	}
	favs, err := svc.ListFavorites(ctx, "user-1")
	if err != nil || len(favs) != 1 || favs[0].ID != free.ID {
		t.Fatalf("收藏未走商品收藏表: %v %+v", err, favs)
	}
	if err := svc.ToggleFavorite(ctx, "user-1", free.ID, false); err != nil {
		t.Fatalf("取消收藏: %v", err)
	}
	if favs, _ := svc.ListFavorites(ctx, "user-1"); len(favs) != 0 {
		t.Fatalf("取消收藏后应为空，实际 %d", len(favs))
	}

	// ⑩ 合并的核心收益：服务能力现在能走商城下单
	orderSvc, _, escrowSvc, octx := newMoneyFlow(t)
	if _, err := escrowSvc.Deposit(octx, "buyer-1", 100000); err != nil {
		t.Fatalf("充值: %v", err)
	}
	o, err := orderSvc.Create(octx, "buyer-1", free.ID, "ent-1", 0, service.OrderReceiver{}, false)
	if err != nil {
		t.Fatalf("服务能力应可下单（合并的核心收益）: %v", err)
	}
	if o.Status != "pending" {
		t.Fatalf("订单状态异常: %q", o.Status)
	}
}
