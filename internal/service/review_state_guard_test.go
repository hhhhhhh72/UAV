package service_test

import (
	"context"
	"testing"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

// 审核状态机回归：仅待审核/已提交资源可审——已通过/已驳回/草稿不可重复审核翻转。
func TestReviewStateGuard(t *testing.T) {
	admin := domain.Actor{ID: "admin-1", Role: domain.RolePlatformAdmin}
	entActor := domain.Actor{ID: "ent-1", Role: domain.RoleEnterprise}

	// ---- 需求：pending 可审，published 后不可再审 ----
	dSvc := service.NewDemandService(memory.NewDemandRepository(nil))
	d, err := dSvc.Create(context.Background(), entActor, service.CreateDemandInput{
		Title: "t", Contact: "138", BizType: "cable_inspection",
	})
	if err != nil {
		t.Fatalf("create demand: %v", err)
	}
	// 审核通过
	if _, err := dSvc.Review(context.Background(), admin, d.ID, "approve", ""); err != nil {
		t.Fatalf("review approve: %v", err)
	}
	// 已 published 再 reject → 拒绝（须用 close）
	if _, err := dSvc.Review(context.Background(), admin, d.ID, "reject", "r"); err == nil {
		t.Fatal("review on published demand must be rejected")
	}
	// 已 published 再 approve（重复审核）→ 拒绝
	if _, err := dSvc.Review(context.Background(), admin, d.ID, "approve", ""); err == nil {
		t.Fatal("duplicate approve must be rejected")
	}
	// 批量审批同样受约束
	if _, err := dSvc.Approve(context.Background(), admin, d.ID); err == nil {
		t.Fatal("batch approve on published demand must be rejected")
	}
	// 驳回后可重提再审（状态机闭环）
	d2, _ := dSvc.Create(context.Background(), entActor, service.CreateDemandInput{
		Title: "t2", Contact: "138", BizType: "cable_inspection",
	})
	if _, err := dSvc.Review(context.Background(), admin, d2.ID, "reject", "信息不全"); err != nil {
		t.Fatalf("review reject: %v", err)
	}
	if _, err := dSvc.Submit(context.Background(), entActor, d2.ID); err != nil {
		t.Fatalf("resubmit: %v", err)
	}
	if _, err := dSvc.Review(context.Background(), admin, d2.ID, "approve", ""); err != nil {
		t.Fatalf("re-review after resubmit: %v", err)
	}

	// ---- 企业：draft 不可审（须先提交），submitted 可审，approved 后不可再审 ----
	users := memory.NewUserRepository(nil)
	// 审核通过需将属主升级为企业角色（升级失败会回滚审批），测试需预置属主用户。
	_, _ = users.Create(context.Background(), domain.User{ID: entActor.ID, Role: domain.RoleEnterprise, Status: "active"})
	eSvc := service.NewEnterpriseSvc(memory.NewEnterpriseRepository(nil), users)
	ent, err := eSvc.Create(context.Background(), entActor, service.CreateEnterpriseInput{Name: "企业A"})
	if err != nil {
		t.Fatalf("create enterprise: %v", err)
	}
	// draft 直接审核 → 拒绝
	if _, err := eSvc.Review(context.Background(), admin, ent.ID, "approve", ""); err == nil {
		t.Fatal("review on draft enterprise must be rejected")
	}
	// 提交后可审
	if _, err := eSvc.Submit(context.Background(), entActor, ent.ID); err != nil {
		t.Fatalf("submit enterprise: %v", err)
	}
	if _, err := eSvc.Review(context.Background(), admin, ent.ID, "approve", ""); err != nil {
		t.Fatalf("review submitted enterprise: %v", err)
	}
	// 已 approved 再 reject → 拒绝
	if _, err := eSvc.Review(context.Background(), admin, ent.ID, "reject", "r"); err == nil {
		t.Fatal("review on approved enterprise must be rejected")
	}
}
