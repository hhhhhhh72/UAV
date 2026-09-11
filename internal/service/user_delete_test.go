package service_test

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

func adminUserActor() domain.Actor { return domain.Actor{ID: "admin-1", Role: domain.RolePlatformAdmin} }
// 注销要把磁盘上的上传文件真删掉（不只是删台账行）：个人信息注销后不应留副本。
func TestDeleteUserRemovesUploadedFiles(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	uploadRepo := memory.NewUploadRepository()
	fileSvc := service.NewFileService(dir, service.WithUploadQuota(uploadRepo, 1<<30))

	f1, err := fileSvc.Upload(ctx, "user-f1", "a.txt", "text/plain", strings.NewReader("public file"))
	if err != nil {
		t.Fatalf("upload public: %v", err)
	}
	f2, err := fileSvc.UploadPrivate(ctx, "user-f1", "id-card.txt", "text/plain", strings.NewReader("private pii"))
	if err != nil {
		t.Fatalf("upload private: %v", err)
	}
	f3, err := fileSvc.Upload(ctx, "user-other", "keep.txt", "text/plain", strings.NewReader("another user"))
	if err != nil {
		t.Fatalf("upload other: %v", err)
	}

	userRepo := memory.NewUserRepository(nil)
	seedUser(t, userRepo, "user-f1", domain.RoleIndividual)
	svc := service.NewUserService(userRepo, service.WithUserFileCleaner(fileSvc))
	res, err := svc.DeleteUser(ctx, adminUserActor(), "user-f1")
	if err != nil {
		t.Fatalf("delete: %v", err)
	}
	if res.FilesRemoved != 2 {
		t.Fatalf("应删除 2 个上传文件，实际 %d", res.FilesRemoved)
	}
	for _, f := range []domain.FileRecord{f1, f2} {
		if _, err := os.Stat(f.StorageKey); !os.IsNotExist(err) {
			t.Fatalf("文件应已从磁盘删除: %s (err=%v)", f.StorageKey, err)
		}
	}
	if _, err := os.Stat(f3.StorageKey); err != nil {
		t.Fatalf("其他用户的文件不得被删: %s (err=%v)", f3.StorageKey, err)
	}
	// 台账里仍有 f1/f2 的键值可供兜底重试？——行由内容处置计划删除（内存实现不涉跨表），
	// 但物理文件必须先于台账行删除，这里断言删除器不会再报错（幂等）。
	if err := fileSvc.RemoveByStorageKey(f1.StorageKey); err != nil {
		t.Fatalf("重复删除应幂等: %v", err)
	}
	// 路径越界必须拒绝（台账里的 key 是历史数据，不能用来删上传目录外的文件）
	if err := fileSvc.RemoveByStorageKey("/etc/hosts"); err == nil {
		t.Fatal("越界路径必须被拒绝")
	}
}


func seedUser(t *testing.T, repo interface {
	Create(ctx context.Context, u domain.User) (domain.User, error)
}, id string, role domain.Role) {
	t.Helper()
	if _, err := repo.Create(context.Background(), domain.User{ID: id, Name: id, Role: role, Status: domain.UserActive, Version: 1}); err != nil {
		t.Fatalf("seed user %s: %v", id, err)
	}
}

// 删除是唯一动作（不可恢复）：账号立即失效 + 账号行进入 7 天缓冲期 + 发布内容保留。
func TestDeleteUserMarksPurgeWindowKeepsContents(t *testing.T) {
	ctx := context.Background()
	userRepo := memory.NewUserRepository(nil)
	seedUser(t, userRepo, "user-1", domain.RoleIndividual)
	demandRepo := memory.NewDemandRepository(nil)
	if _, err := demandRepo.Create(ctx, domain.Demand{
		ID: "demand-1", PublisherID: "user-1", Title: "巡检需求", Status: domain.DemandPublished, Version: 1,
	}); err != nil {
		t.Fatalf("seed demand: %v", err)
	}

	svc := service.NewUserService(userRepo)
	res, err := svc.DeleteUser(ctx, adminUserActor(), "user-1")
	if err != nil {
		t.Fatalf("delete: %v", err)
	}
	if res.Mode != "deleted" || res.Status != domain.UserDeleted || !res.KeptContents {
		t.Fatalf("unexpected result: %+v", res)
	}
	purgeAfter, perr := time.Parse(time.RFC3339, res.PurgeAfter)
	if perr != nil {
		t.Fatalf("purge_after 应为 RFC3339: %q", res.PurgeAfter)
	}
	if d := time.Until(purgeAfter); d < service.UserPurgeRetention-2*time.Minute || d > service.UserPurgeRetention+2*time.Minute {
		t.Fatalf("缓冲期应约 %v，实际 %v", service.UserPurgeRetention, d)
	}

	u, err := userRepo.FindByID(ctx, "user-1")
	if err != nil {
		t.Fatalf("缓冲期内账号行应保留（内容里的作者 ID 要能解析）: %v", err)
	}
	if u.Status != domain.UserDeleted || u.DeletedAt == nil || u.TokenVersion == 0 {
		t.Fatalf("删除标记不完整: status=%s deleted_at=%v tv=%d", u.Status, u.DeletedAt, u.TokenVersion)
	}
	items, err := demandRepo.ListByPublisher(ctx, "user-1")
	if err != nil || len(items) != 1 {
		t.Fatalf("删除账号不应删除其发布的内容: items=%d err=%v", len(items), err)
	}
	// 重复删除：缓冲期内仍可见 → 第二次按"不存在"处理（账号已删除）
	if _, err := svc.DeleteUser(ctx, adminUserActor(), "user-1"); !errors.Is(err, service.ErrUserNotFound) {
		t.Fatalf("重复删除应 404，实际 %v", err)
	}
}

// 清除只在缓冲期届满后发生：刚删除的账号 PurgeExpired 不动它，超过截止时间的才清掉。
func TestPurgeRespectsRetentionWindow(t *testing.T) {
	ctx := context.Background()
	userRepo := memory.NewUserRepository(nil)
	seedUser(t, userRepo, "user-p1", domain.RoleIndividual)
	svc := service.NewUserService(userRepo)
	if _, err := svc.DeleteUser(ctx, adminUserActor(), "user-p1"); err != nil {
		t.Fatalf("delete: %v", err)
	}

	// 1) 正常调用（截止时间 = now-7d）：缓冲期未到，不清除
	if n, err := svc.PurgeExpired(ctx); err != nil || n != 0 {
		t.Fatalf("缓冲期内不应清除: n=%d err=%v", n, err)
	}
	if _, err := userRepo.FindByID(ctx, "user-p1"); err != nil {
		t.Fatal("缓冲期内账号行被误删")
	}

	// 2) 缓冲期设为 0（等价于 7 天后再扫一次）→ 清除账号行（服务编排：清文件→兜底内容处置→删行）
	fast := service.NewUserService(userRepo, service.WithPurgeRetention(0))
	n, err := fast.PurgeExpired(ctx)
	if err != nil || n != 1 {
		t.Fatalf("到期应清除 1 个: n=%d err=%v", n, err)
	}
	if _, err := userRepo.FindByID(ctx, "user-p1"); err == nil {
		t.Fatal("清除后不该再查得到")
	}
}

// 内容处置策略的不变量（这是产品口径，改动必须落到测试里）：
//   - 下架：需求/商品/课程/动态/成果 在架内容；
//   - 擦除：报名、对接意向、预约、企业联系人等业务事实行里的个人信息列；
//   - 删除：简历/投递/站内信/文件/收藏/飞手名录档案；
//   - 绝不动：工单、合同、资金、保单、审计等履约与合规记录。
func TestUserContentCleanupPlanInvariants(t *testing.T) {
	plan := service.UserContentCleanupPlan()

	has := func(list []string, want string) bool {
		for _, it := range list {
			if it == want {
				return true
			}
		}
		return false
	}
	takeDown := make([]string, 0, len(plan.TakeDown))
	for _, r := range plan.TakeDown {
		takeDown = append(takeDown, r.Table)
		if r.OffValue == "" || r.OnValue == "" {
			t.Fatalf("%s 的下架规则缺少状态值: %+v", r.Table, r)
		}
	}
	for _, want := range []string{"demands", "drone_products", "training_courses", "posts", "achievements"} {
		if !has(takeDown, want) {
			t.Fatalf("下架计划缺少 %s（在架内容必须下架，否则留下僵尸条目）", want)
		}
	}
	wipe := make([]string, 0, len(plan.Wipe))
	for _, r := range plan.Wipe {
		wipe = append(wipe, r.Table)
		if len(r.Columns) == 0 {
			t.Fatalf("%s 的擦除规则没有列: %+v", r.Table, r)
		}
	}
	for _, want := range []string{"demands", "enterprises", "training_enrollments", "test_site_bookings", "competition_registrations"} {
		if !has(wipe, want) {
			t.Fatalf("擦除计划缺少 %s（个人信息列必须清空）", want)
		}
	}
	drop := make([]string, 0, len(plan.Drop))
	for _, r := range plan.Drop {
		drop = append(drop, r.Table)
	}
	for _, want := range []string{"resumes", "job_applications", "messages", "files", "uploads", "certified_pilots"} {
		if !has(drop, want) {
			t.Fatalf("删除计划缺少 %s（纯个人信息行应整行删除）", want)
		}
	}
	// 履约/合规记录一律不得出现在任何规则里
	frozen := []string{"work_orders", "contracts", "escrow_accounts", "loan_applications", "insurance_policies", "audit_logs", "trade_orders"}
	for _, table := range frozen {
		if has(takeDown, table) || has(wipe, table) || has(drop, table) {
			t.Fatalf("%s 是履约/合规记录，处置计划不得触碰", table)
		}
	}
}

// 权限与保护：非管理员 / 超级管理员（SUPER_ADMIN_PHONE 指定）/ 不能删自己 / 不存在的账号；
// 被拒路径不得改动任何账号。
//
// 语义变更（2026-09-11）：此前保护的是字面 id "admin"——库里根本不存在的幽灵账号；
// 现在保护的是真实的超管（按手机号命中），并新增"不能删自己"的自锁保护。
func TestDeleteUserGuards(t *testing.T) {
	ctx := context.Background()
	userRepo := memory.NewUserRepository(nil)
	seedUser(t, userRepo, "user-1", domain.RoleIndividual)
	seedUser(t, userRepo, "user-19800000000", domain.RolePlatformAdmin)
	svc := service.NewUserService(userRepo, service.WithSuperAdminPhone("19800000000"))

	if _, err := svc.DeleteUser(ctx, domain.Actor{ID: "ent-1", Role: domain.RoleEnterprise}, "user-1"); !errors.Is(err, service.ErrAdminOnly) {
		t.Fatalf("非管理员应被拒，实际 %v", err)
	}
	// 超级管理员（配置里的手机号命中 user-19800000000）不可删除
	if _, err := svc.DeleteUser(ctx, adminUserActor(), "user-19800000000"); !errors.Is(err, service.ErrSuperAdminProtected) {
		t.Fatalf("超级管理员应受保护，实际 %v", err)
	}
	// 不能删自己（防自锁：唯一管理员删了自己，后台就没人能进）
	if _, err := svc.DeleteUser(ctx, adminUserActor(), adminUserActor().ID); !errors.Is(err, service.ErrCannotModifySelf) {
		t.Fatalf("删自己应被拒，实际 %v", err)
	}
	if _, err := svc.DeleteUser(ctx, adminUserActor(), "ghost"); !errors.Is(err, service.ErrUserNotFound) {
		t.Fatalf("不存在应报 ErrUserNotFound，实际 %v", err)
	}
	u, _ := userRepo.FindByID(ctx, "user-1")
	if u.Status != domain.UserActive || u.TokenVersion != 0 || u.DeletedAt != nil {
		t.Fatalf("拒绝路径不应改动账号: %+v", u)
	}
	admin, _ := userRepo.FindByID(ctx, "user-19800000000")
	if admin.Status != domain.UserActive || admin.DeletedAt != nil {
		t.Fatalf("超管拒绝路径不应改动账号: %+v", admin)
	}
}

// ChangeRole 的三道闸：非平台管理员 / 改自己 / 改超管 / 非法角色 / 目标不存在；
// 合法变更要真的改掉角色。
func TestChangeRoleGuards(t *testing.T) {
	ctx := context.Background()
	userRepo := memory.NewUserRepository(nil)
	seedUser(t, userRepo, "user-1", domain.RoleIndividual)
	seedUser(t, userRepo, "user-19800000000", domain.RolePlatformAdmin)
	svc := service.NewUserService(userRepo, service.WithSuperAdminPhone("19800000000"))

	if err := svc.ChangeRole(ctx, domain.Actor{ID: "admin-2", Role: domain.RoleAssociationAdmin}, "user-1", domain.RoleEnterprise); !errors.Is(err, service.ErrAdminOnly) {
		t.Fatalf("协会管理员改角色应被拒，实际 %v", err)
	}
	if err := svc.ChangeRole(ctx, adminUserActor(), adminUserActor().ID, domain.RoleIndividual); !errors.Is(err, service.ErrCannotModifySelf) {
		t.Fatalf("改自己应被拒，实际 %v", err)
	}
	if err := svc.ChangeRole(ctx, adminUserActor(), "user-19800000000", domain.RoleIndividual); !errors.Is(err, service.ErrSuperAdminProtected) {
		t.Fatalf("改超管应被拒，实际 %v", err)
	}
	if err := svc.ChangeRole(ctx, adminUserActor(), "user-1", domain.Role("superuser")); !errors.Is(err, service.ErrInvalidRole) {
		t.Fatalf("非法角色应被拒，实际 %v", err)
	}
	if err := svc.ChangeRole(ctx, adminUserActor(), "ghost", domain.RoleEnterprise); !errors.Is(err, service.ErrUserNotFound) {
		t.Fatalf("目标不存在应 404 语义，实际 %v", err)
	}
	if err := svc.ChangeRole(ctx, adminUserActor(), "user-1", domain.RoleEnterprise); err != nil {
		t.Fatalf("合法变更应成功，实际 %v", err)
	}
	u, _ := userRepo.FindByID(ctx, "user-1")
	if u.Role != domain.RoleEnterprise {
		t.Fatalf("角色未生效: %+v", u.Role)
	}
	// 降权/升级都要作废旧令牌（token_version 自增）
	if u.TokenVersion == 0 {
		t.Fatalf("角色变更应自增 token_version 使旧令牌失效，实际 %d", u.TokenVersion)
	}
}
