package service_test

import (
	"context"
	"errors"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

func seedUserWithPassword(t *testing.T, repo interface {
	Create(ctx context.Context, u domain.User) (domain.User, error)
}, id, password string, role domain.Role) {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if _, err := repo.Create(context.Background(), domain.User{
		ID: id, Name: id, Role: role, Status: domain.UserActive, PasswordHash: string(hash), Version: 1,
	}); err != nil {
		t.Fatalf("seed user %s: %v", id, err)
	}
}

// 本人改密：旧密码必须正确；弱口令/与原密码相同都拒绝；成功后 token_version 自增（旧令牌失效）。
func TestChangePassword(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewUserRepository(nil)
	seedUserWithPassword(t, repo, "u1", "InitPass123", domain.RolePlatformAdmin)
	svc := service.NewUserService(repo)
	act := domain.Actor{ID: "u1", Role: domain.RolePlatformAdmin}

	if err := svc.ChangePassword(ctx, act, "wrong-old", "NewPass1234"); !errors.Is(err, service.ErrOldPasswordWrong) {
		t.Fatalf("旧密码错误应拒绝，实际 %v", err)
	}
	if err := svc.ChangePassword(ctx, act, "InitPass123", "short"); !errors.Is(err, service.ErrWeakPassword) {
		t.Fatalf("短口令应拒绝，实际 %v", err)
	}
	if err := svc.ChangePassword(ctx, act, "InitPass123", "InitPass123"); !errors.Is(err, service.ErrWeakPassword) {
		t.Fatalf("与原密码相同应拒绝，实际 %v", err)
	}
	if err := svc.ChangePassword(ctx, act, "InitPass123", "aaaaaaaaaa"); !errors.Is(err, service.ErrWeakPassword) {
		t.Fatalf("全同字符应拒绝，实际 %v", err)
	}

	before, _ := repo.FindByID(ctx, "u1")
	if err := svc.ChangePassword(ctx, act, "InitPass123", "NewPass1234"); err != nil {
		t.Fatalf("正常改密应成功，实际 %v", err)
	}
	after, _ := repo.FindByID(ctx, "u1")
	if after.TokenVersion <= before.TokenVersion {
		t.Fatalf("改密应自增 token_version（旧令牌立即失效）：%d → %d", before.TokenVersion, after.TokenVersion)
	}
	if bcrypt.CompareHashAndPassword([]byte(after.PasswordHash), []byte("NewPass1234")) != nil {
		t.Fatal("新密码应可用于登录")
	}
	if bcrypt.CompareHashAndPassword([]byte(after.PasswordHash), []byte("InitPass123")) == nil {
		t.Fatal("旧密码不应再可用")
	}
}

// 无密码账号（只能用微信/验证码登录）改密 → 明确报错，引导找管理员重置。
func TestChangePasswordWithoutPasswordSet(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewUserRepository(nil)
	seedUser(t, repo, "u2", domain.RoleIndividual) // 该 helper 不设密码
	svc := service.NewUserService(repo)
	if err := svc.ChangePassword(ctx, domain.Actor{ID: "u2", Role: domain.RoleIndividual}, "x", "NewPass1234"); !errors.Is(err, service.ErrPasswordNotSet) {
		t.Fatalf("应报 ErrPasswordNotSet，实际 %v", err)
	}
}

// 管理员重置：仅平台管理员；弱口令拒绝；重置后旧密码失效且 token_version 自增。
func TestResetPasswordGuards(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewUserRepository(nil)
	seedUserWithPassword(t, repo, "target", "InitPass123", domain.RoleIndividual)
	svc := service.NewUserService(repo)

	if err := svc.ResetPassword(ctx, domain.Actor{ID: "ent-1", Role: domain.RoleEnterprise}, "target", "ResetPass789"); !errors.Is(err, service.ErrAdminOnly) {
		t.Fatalf("非平台管理员应被拒，实际 %v", err)
	}
	if err := svc.ResetPassword(ctx, adminUserActor(), "target", "short"); !errors.Is(err, service.ErrWeakPassword) {
		t.Fatalf("弱口令应被拒，实际 %v", err)
	}
	if err := svc.ResetPassword(ctx, adminUserActor(), "ghost", "ResetPass789"); !errors.Is(err, service.ErrUserNotFound) {
		t.Fatalf("账号不存在应报 ErrUserNotFound，实际 %v", err)
	}

	before, _ := repo.FindByID(ctx, "target")
	if err := svc.ResetPassword(ctx, adminUserActor(), "target", "ResetPass789"); err != nil {
		t.Fatalf("重置应成功，实际 %v", err)
	}
	after, _ := repo.FindByID(ctx, "target")
	if after.TokenVersion <= before.TokenVersion {
		t.Fatal("重置也应让旧令牌失效")
	}
	if bcrypt.CompareHashAndPassword([]byte(after.PasswordHash), []byte("ResetPass789")) != nil {
		t.Fatal("重置后的新密码应可用")
	}
	if bcrypt.CompareHashAndPassword([]byte(after.PasswordHash), []byte("InitPass123")) == nil {
		t.Fatal("旧密码应失效")
	}
}
