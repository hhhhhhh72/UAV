package service_test

import (
	"context"
	"testing"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

// TestReviewApprovedUpgradesOwnerRole: 入驻审核通过后，owner 用户角色必须升级为 enterprise。
// 回归：此前 Review 只改企业状态不升角色，导致已通过企业用户无法发招聘、身份永远是个体。
func TestReviewApprovedUpgradesOwnerRole(t *testing.T) {
	users := memory.NewUserRepository(nil)
	svc := service.NewEnterpriseSvc(memory.NewEnterpriseRepository(nil), users)

	// Arrange: 个体用户 + 提交入驻申请
	owner := domain.User{ID: "user-ent-owner", Role: domain.RoleIndividual, Status: "active"}
	if _, err := users.Create(context.Background(), owner); err != nil {
		t.Fatal(err)
	}
	a := domain.Actor{ID: owner.ID, Role: domain.RoleIndividual}
	e, err := svc.Create(context.Background(), a, service.CreateEnterpriseInput{Name: "测试企业"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Submit(context.Background(), a, e.ID); err != nil {
		t.Fatal(err)
	}

	// Act: 管理员审核通过
	admin := domain.Actor{ID: "admin-1", Role: domain.RoleAssociationAdmin}
	if _, err := svc.Review(context.Background(), admin, e.ID, "approve", ""); err != nil {
		t.Fatal(err)
	}

	// Assert: 用户角色升级为 enterprise
	u, err := users.FindByID(context.Background(), owner.ID)
	if err != nil {
		t.Fatal(err)
	}
	if u.Role != domain.RoleEnterprise {
		t.Fatalf("role: expected %s, got %s", domain.RoleEnterprise, u.Role)
	}
}

// TestReviewStoresReason: 驳回/需补充必须保留审核意见，用户端才能看到原因。
// 回归：此前 reason 只做必填校验后即丢弃，用户被驳回却看不到任何理由。
func TestReviewStoresReason(t *testing.T) {
	users := memory.NewUserRepository(nil)
	svc := service.NewEnterpriseSvc(memory.NewEnterpriseRepository(nil), users)

	owner := domain.User{ID: "user-ent-2", Role: domain.RoleIndividual, Status: "active"}
	if _, err := users.Create(context.Background(), owner); err != nil {
		t.Fatal(err)
	}
	a := domain.Actor{ID: owner.ID, Role: domain.RoleIndividual}
	e, err := svc.Create(context.Background(), a, service.CreateEnterpriseInput{Name: "测试企业"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Submit(context.Background(), a, e.ID); err != nil {
		t.Fatal(err)
	}

	// Act: 驳回并附理由
	admin := domain.Actor{ID: "admin-2", Role: domain.RolePlatformAdmin}
	reason := "营业执照不清晰，请重新拍摄上传"
	got, err := svc.Review(context.Background(), admin, e.ID, "reject", reason)
	if err != nil {
		t.Fatal(err)
	}

	// Assert: 理由持久化在审核意见字段
	if got.ReviewComment != reason {
		t.Fatalf("review_comment: expected %q, got %q", reason, got.ReviewComment)
	}
	if got.Status != domain.EnterpriseRejected {
		t.Fatalf("status: expected %s, got %s", domain.EnterpriseRejected, got.Status)
	}
}

// TestRejectedEnterpriseCanEditAndResubmit: 驳回后企业主可编辑并重新提交（PRD FR-2.2 重提闭环）。
// 回归：此前 Update/Submit 的 owner 门禁拒绝 rejected，前端"重新编辑并提交"按钮形同虚设。
func TestRejectedEnterpriseCanEditAndResubmit(t *testing.T) {
	users := memory.NewUserRepository(nil)
	svc := service.NewEnterpriseSvc(memory.NewEnterpriseRepository(nil), users)

	owner := domain.User{ID: "user-ent-3", Role: domain.RoleIndividual, Status: "active"}
	if _, err := users.Create(context.Background(), owner); err != nil {
		t.Fatal(err)
	}
	a := domain.Actor{ID: owner.ID, Role: domain.RoleIndividual}
	e, err := svc.Create(context.Background(), a, service.CreateEnterpriseInput{Name: "测试企业"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Submit(context.Background(), a, e.ID); err != nil {
		t.Fatal(err)
	}
	admin := domain.Actor{ID: "admin-3", Role: domain.RoleAssociationAdmin}
	if _, err := svc.Review(context.Background(), admin, e.ID, "reject", "资料不全"); err != nil {
		t.Fatal(err)
	}

	// 驳回后：owner 可编辑（改企业名）
	updated, err := svc.Update(context.Background(), a, e.ID, service.CreateEnterpriseInput{Name: "测试企业-修订版"})
	if err != nil {
		t.Fatalf("owner edit rejected enterprise should succeed: %v", err)
	}
	if updated.Name != "测试企业-修订版" {
		t.Fatalf("name: expected %q, got %q", "测试企业-修订版", updated.Name)
	}
	// 编辑后状态保持 rejected（走重新提交）
	if updated.Status != domain.EnterpriseRejected {
		t.Fatalf("status after edit: expected %s, got %s", domain.EnterpriseRejected, updated.Status)
	}

	// 重新提交 → submitted，可再次进入审核队列
	submitted, err := svc.Submit(context.Background(), a, e.ID)
	if err != nil {
		t.Fatalf("resubmit rejected enterprise should succeed: %v", err)
	}
	if submitted.Status != domain.EnterpriseSubmitted {
		t.Fatalf("status after resubmit: expected %s, got %s", domain.EnterpriseSubmitted, submitted.Status)
	}
}

// TestApprovedEnterpriseOwnerEditResubmits: 已通过企业 owner 可编辑（P1：自助更新送审），
// 编辑后自动回「已提交」进入重新审核；不允许 owner 直接提交已通过的企业。
func TestApprovedEnterpriseOwnerEditResubmits(t *testing.T) {
	users := memory.NewUserRepository(nil)
	svc := service.NewEnterpriseSvc(memory.NewEnterpriseRepository(nil), users)

	owner := domain.User{ID: "user-ent-4", Role: domain.RoleIndividual, Status: "active"}
	if _, err := users.Create(context.Background(), owner); err != nil {
		t.Fatal(err)
	}
	a := domain.Actor{ID: owner.ID, Role: domain.RoleIndividual}
	e, err := svc.Create(context.Background(), a, service.CreateEnterpriseInput{Name: "测试企业"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Submit(context.Background(), a, e.ID); err != nil {
		t.Fatal(err)
	}
	admin := domain.Actor{ID: "admin-4", Role: domain.RolePlatformAdmin}
	if _, err := svc.Review(context.Background(), admin, e.ID, "approve", ""); err != nil {
		t.Fatal(err)
	}
	// 企业主更新已认证档案 → 允许，状态回「已提交」（重新审核闭环）
	updated, err := svc.Update(context.Background(), a, e.ID, service.CreateEnterpriseInput{Name: "改" + e.Name})
	if err != nil {
		t.Fatalf("approved owner edit should succeed (自助更新送审): %v", err)
	}
	if updated.Status != domain.EnterpriseSubmitted {
		t.Fatalf("status after approved owner edit: expected %s, got %s", domain.EnterpriseSubmitted, updated.Status)
	}
	if updated.ReviewComment != "" {
		t.Fatalf("review comment should be cleared on resubmit, got %q", updated.ReviewComment)
	}
	if _, err := svc.Submit(context.Background(), a, e.ID); err == nil {
		t.Fatal("owner must not resubmit a non-draft enterprise via submit")
	}
}
