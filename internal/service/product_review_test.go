package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

func newReviewSvc() (*service.TradingService, context.Context) {
	return service.NewTradingService(memory.NewProductRepository(), memory.NewRepairRepository(), nil, nil), context.Background()
}

// 审核维度与上架维度必须正交——这是把 check_status 从 status 里拆出来的全部意义。
//
// 拆分前"驳回"只能写 status='removed'，与"卖家主动下架"同值：卖家在小程序
// "我的发布"里分不清自己是被驳回还是自己下的，也没有原因可看。
func TestReviewProductKeepsCheckAndStatusOrthogonal(t *testing.T) {
	ctx := context.Background()
	svc, _ := newReviewSvc()
	seller := domain.Actor{ID: "seller-1", Role: domain.RoleEnterprise}
	admin := domain.Actor{ID: "admin-1", Role: domain.RolePlatformAdmin}

	p, err := svc.CreateProduct(ctx, seller, domain.ProductDrone, "大疆M350", "", "DJI", "M350", "new", "", "", 8800000, nil, nil)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if p.CheckStatus != domain.ProductCheckPending {
		t.Fatalf("新发布商品应为待审核，实际 check_status=%q", p.CheckStatus)
	}
	if service.ProductVisibleInHall(p) {
		t.Fatal("待审核商品不得公开可见")
	}

	// 驳回：必须带原因；**不得**把 status 改成 removed（那是"下架"，是另一回事）
	rej, err := svc.ReviewProduct(ctx, admin, p.ID, domain.ProductCheckRejected, "  资料不全，请补充型号铭牌照片  ")
	if err != nil {
		t.Fatalf("reject: %v", err)
	}
	if rej.CheckStatus != domain.ProductCheckRejected {
		t.Fatalf("check_status 应为 rejected，实际 %q", rej.CheckStatus)
	}
	if rej.CheckReason != "资料不全，请补充型号铭牌照片" {
		t.Fatalf("驳回原因未落库或未 TrimSpace：%q", rej.CheckReason)
	}
	if rej.Status == "removed" {
		t.Fatal("驳回不得写成 status=removed——那与卖家主动下架同值，卖家分不清")
	}
	if rej.ReviewedBy != "admin-1" || rej.ReviewedAt == nil {
		t.Fatalf("审核留痕缺失：reviewed_by=%q reviewed_at=%v", rej.ReviewedBy, rej.ReviewedAt)
	}
	if service.ProductVisibleInHall(rej) {
		t.Fatal("被驳回商品不得公开可见")
	}

	// 驳回后重新审核通过：清掉旧驳回理由，并上架
	pass, err := svc.ReviewProduct(ctx, admin, p.ID, domain.ProductCheckPassed, "随便写的")
	if err != nil {
		t.Fatalf("re-review: %v", err)
	}
	if pass.CheckStatus != domain.ProductCheckPassed || pass.Status != "listed" {
		t.Fatalf("审核通过应同时 check_status=passed + status=listed，实际 %q/%q", pass.CheckStatus, pass.Status)
	}
	if pass.CheckReason != "" {
		t.Fatalf("重新通过后应清空上一次的驳回理由，实际 %q", pass.CheckReason)
	}
	if !service.ProductVisibleInHall(pass) {
		t.Fatal("审核通过且已上架的商品必须公开可见")
	}
}

// 公开可见 = 审核通过 AND 在售，缺一不可。表格覆盖两条维度的所有组合。
func TestProductVisibleInHallRequiresBothDimensions(t *testing.T) {
	cases := []struct {
		check  string
		status string
		want   bool
	}{
		{domain.ProductCheckPassed, "listed", true},
		{domain.ProductCheckPassed, "sold", false},
		{domain.ProductCheckPassed, "removed", false},
		{domain.ProductCheckPassed, "pending", false},
		{domain.ProductCheckPending, "listed", false},   // ← 拆分后最容易漏的一格
		{domain.ProductCheckPending, "pending", false},
		{domain.ProductCheckRejected, "listed", false},  // ← 被驳回的商品即便被改成 listed 也不得公开
		{domain.ProductCheckRejected, "removed", false},
		{"", "listed", false},                           // 历史脏数据：无审核状态一律不公开
	}
	for _, c := range cases {
		got := service.ProductVisibleInHall(domain.DroneProduct{CheckStatus: c.check, Status: c.status})
		if got != c.want {
			t.Errorf("check=%q status=%q: 期望 %v 实际 %v", c.check, c.status, c.want, got)
		}
	}
}

// 审核入参校验：驳回必须给原因；审核结果只能是 passed/rejected。
func TestReviewProductValidatesInput(t *testing.T) {
	ctx := context.Background()
	svc, _ := newReviewSvc()
	seller := domain.Actor{ID: "seller-1", Role: domain.RoleEnterprise}
	admin := domain.Actor{ID: "admin-1", Role: domain.RolePlatformAdmin}

	newPending := func(t *testing.T) string {
		t.Helper()
		p, err := svc.CreateProduct(ctx, seller, domain.ProductDrone, "大疆M350", "", "", "", "new", "", "", 100, nil, nil)
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		return p.ID
	}

	// 驳回不给原因
	if _, err := svc.ReviewProduct(ctx, admin, newPending(t), domain.ProductCheckRejected, "   "); !errors.Is(err, service.ErrProductReviewInvalid) {
		t.Fatalf("驳回不带原因应返回 ErrProductReviewInvalid，实际 %v", err)
	}
	// 审核结果乱填
	if _, err := svc.ReviewProduct(ctx, admin, newPending(t), "looks-good", ""); !errors.Is(err, service.ErrProductReviewInvalid) {
		t.Fatalf("非法审核结果应返回 ErrProductReviewInvalid，实际 %v", err)
	}
	if _, err := svc.ReviewProduct(ctx, admin, newPending(t), "", ""); !errors.Is(err, service.ErrProductReviewInvalid) {
		t.Fatalf("空审核结果应返回 ErrProductReviewInvalid，实际 %v", err)
	}
	// 原因超长
	long := strings.Repeat("理", 201)
	if _, err := svc.ReviewProduct(ctx, admin, newPending(t), domain.ProductCheckRejected, long); !errors.Is(err, service.ErrProductReviewInvalid) {
		t.Fatalf("超长驳回原因应返回 ErrProductReviewInvalid，实际 %v", err)
	}
}

// 条件更新语义：已终审（passed）的商品不可重复审核，防并发重复审核把状态改回去。
func TestReviewProductRejectsDoubleReview(t *testing.T) {
	ctx := context.Background()
	svc, _ := newReviewSvc()
	seller := domain.Actor{ID: "seller-1", Role: domain.RoleEnterprise}
	admin := domain.Actor{ID: "admin-1", Role: domain.RolePlatformAdmin}

	p, err := svc.CreateProduct(ctx, seller, domain.ProductDrone, "大疆M350", "", "", "", "new", "", "", 100, nil, nil)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := svc.ReviewProduct(ctx, admin, p.ID, domain.ProductCheckPassed, ""); err != nil {
		t.Fatalf("first review: %v", err)
	}
	// 第二个管理员基于陈旧页面再点一次"驳回"：不得把已通过的商品改回驳回
	if _, err := svc.ReviewProduct(ctx, admin, p.ID, domain.ProductCheckRejected, "重复审核"); err == nil {
		t.Fatal("已审核通过的商品不应能被再次审核（防并发重复审核）")
	}
	got, err := svc.GetProduct(ctx, p.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.CheckStatus != domain.ProductCheckPassed || got.Status != "listed" {
		t.Fatalf("重复审核不得改动已终审商品，实际 %q/%q", got.CheckStatus, got.Status)
	}
}
