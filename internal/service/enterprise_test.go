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
