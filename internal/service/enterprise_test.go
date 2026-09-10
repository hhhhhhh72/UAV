package service_test

import (
	"context"
	"strconv"
	"strings"
	"testing"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

// validEntInput 资料齐全的入驻入参（提审档要求全部必填项，缺失会被 Submit 拒绝）。
func validEntInput(name string) service.CreateEnterpriseInput {
	return service.CreateEnterpriseInput{
		Name:             name,
		CreditCode:       "91500108MA5U1234XY",
		LegalPerson:      "张三",
		ContactPerson:    "李四",
		ContactPhone:     "13800138000",
		Email:            "contact@example.com",
		IndustryCategory: "整机研发",
		Scale:            "50-100人",
		LicenseURL:       "/uploads/private/lic-test-file",
	}
}

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
	e, err := svc.Create(context.Background(), a, validEntInput("测试企业"))
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
	e, err := svc.Create(context.Background(), a, validEntInput("测试企业"))
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
	e, err := svc.Create(context.Background(), a, validEntInput("测试企业"))
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
	e, err := svc.Create(context.Background(), a, validEntInput("测试企业"))
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

// —— 字段合法性（权威校验在 Service 层：前端校验只影响体验，绕过前端不得写入脏数据）——

// TestEnterpriseCreateRejectsMalformedFields: 格式非法的字段必须在 Create 就被拒。
func TestEnterpriseCreateRejectsMalformedFields(t *testing.T) {
	cases := []struct {
		label string
		mut   func(*service.CreateEnterpriseInput)
	}{
		{"企业名称全是空格", func(in *service.CreateEnterpriseInput) { in.Name = "   " }},
		{"企业名称过短", func(in *service.CreateEnterpriseInput) { in.Name = "甲" }},
		{"信用代码不足18位", func(in *service.CreateEnterpriseInput) { in.CreditCode = "12345" }},
		{"信用代码含非法字符", func(in *service.CreateEnterpriseInput) { in.CreditCode = "9150-0108MA5U1234X" }},
		{"手机号非法", func(in *service.CreateEnterpriseInput) { in.ContactPhone = "12345" }},
		{"邮箱非法", func(in *service.CreateEnterpriseInput) { in.Email = "not-an-email" }},
		{"成立时间晚于今天", func(in *service.CreateEnterpriseInput) { in.FoundedAt = "2099-01-01" }},
		{"成立时间格式非法", func(in *service.CreateEnterpriseInput) { in.FoundedAt = "去年" }},
		{"企业简介超长", func(in *service.CreateEnterpriseInput) { in.Description = strings.Repeat("长", 501) }},
	}
	for i, c := range cases {
		t.Run(c.label, func(t *testing.T) {
			users := memory.NewUserRepository(nil)
			svc := service.NewEnterpriseSvc(memory.NewEnterpriseRepository(nil), users)
			uid := "user-bad-" + strconv.Itoa(i)
			if _, err := users.Create(context.Background(), domain.User{ID: uid, Role: domain.RoleIndividual, Status: "active"}); err != nil {
				t.Fatal(err)
			}
			in := validEntInput("测试企业")
			c.mut(&in)
			if _, err := svc.Create(context.Background(), domain.Actor{ID: uid, Role: domain.RoleIndividual}, in); err == nil {
				t.Fatalf("%s: 期望被拒绝，实际通过", c.label)
			}
		})
	}
}

// TestEnterpriseCreditCodeNormalized: 小写信用代码自动转大写后应被接受（不能因大小写误拒）。
func TestEnterpriseCreditCodeNormalized(t *testing.T) {
	users := memory.NewUserRepository(nil)
	svc := service.NewEnterpriseSvc(memory.NewEnterpriseRepository(nil), users)
	if _, err := users.Create(context.Background(), domain.User{ID: "user-lower", Role: domain.RoleIndividual, Status: "active"}); err != nil {
		t.Fatal(err)
	}
	in := validEntInput("  测试企业  ")
	in.CreditCode = "91500108ma5u1234xy"
	e, err := svc.Create(context.Background(), domain.Actor{ID: "user-lower", Role: domain.RoleIndividual}, in)
	if err != nil {
		t.Fatalf("小写信用代码应被规范化后接受: %v", err)
	}
	if e.CreditCode != "91500108MA5U1234XY" {
		t.Fatalf("credit_code: expected 大写, got %q", e.CreditCode)
	}
	if strings.TrimSpace(e.Name) != "测试企业" {
		t.Fatalf("name 未去首尾空白: %q", e.Name)
	}
}

// TestEnterpriseSubmitRequiresCompleteProfile: 资料不全（缺营业执照等）不得进入审核队列。
func TestEnterpriseSubmitRequiresCompleteProfile(t *testing.T) {
	cases := []struct {
		label string
		mut   func(*service.CreateEnterpriseInput)
	}{
		{"缺营业执照", func(in *service.CreateEnterpriseInput) { in.LicenseURL = "" }},
		{"缺法人代表", func(in *service.CreateEnterpriseInput) { in.LegalPerson = "" }},
		{"缺联系人", func(in *service.CreateEnterpriseInput) { in.ContactPerson = "" }},
		{"缺联系电话", func(in *service.CreateEnterpriseInput) { in.ContactPhone = "" }},
		{"缺企业分类", func(in *service.CreateEnterpriseInput) { in.IndustryCategory = "" }},
		{"缺企业规模", func(in *service.CreateEnterpriseInput) { in.Scale = "" }},
	}
	for i, c := range cases {
		t.Run(c.label, func(t *testing.T) {
			users := memory.NewUserRepository(nil)
			svc := service.NewEnterpriseSvc(memory.NewEnterpriseRepository(nil), users)
			uid := "user-inc-" + strconv.Itoa(i)
			if _, err := users.Create(context.Background(), domain.User{ID: uid, Role: domain.RoleIndividual, Status: "active"}); err != nil {
				t.Fatal(err)
			}
			a := domain.Actor{ID: uid, Role: domain.RoleIndividual}
			in := validEntInput("测试企业")
			c.mut(&in)
			e, err := svc.Create(context.Background(), a, in)
			if err != nil {
				t.Fatalf("Create(草稿档) 不应因必填项缺失而失败: %v", err)
			}
			if _, err := svc.Submit(context.Background(), a, e.ID); err == nil {
				t.Fatalf("%s: 提审应被拒绝，实际通过", c.label)
			}
		})
	}
}

// TestEnterpriseSubmitAcceptsCompleteProfile: 资料齐全可正常提审。
func TestEnterpriseSubmitAcceptsCompleteProfile(t *testing.T) {
	users := memory.NewUserRepository(nil)
	svc := service.NewEnterpriseSvc(memory.NewEnterpriseRepository(nil), users)
	if _, err := users.Create(context.Background(), domain.User{ID: "user-full", Role: domain.RoleIndividual, Status: "active"}); err != nil {
		t.Fatal(err)
	}
	a := domain.Actor{ID: "user-full", Role: domain.RoleIndividual}
	e, err := svc.Create(context.Background(), a, validEntInput("重庆测试无人机有限公司"))
	if err != nil {
		t.Fatal(err)
	}
	got, err := svc.Submit(context.Background(), a, e.ID)
	if err != nil {
		t.Fatalf("资料齐全应可提审: %v", err)
	}
	if got.Status != domain.EnterpriseSubmitted {
		t.Fatalf("status: expected %s, got %s", domain.EnterpriseSubmitted, got.Status)
	}
}
