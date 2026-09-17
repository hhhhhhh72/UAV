package service_test

import (
	"context"
	"testing"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

// 课程归属（course.OrgID）决定学费结算方向：
// completeEnrollment 用 Release(学员, course.OrgID, 学费) 放款。
// 所以这里必须钉死两条：
//
//  1. 个人/企业**不能**指定他人归属——否则可以把别人的学费结算指向任意账户；
//  2. 平台/协会管理员**可以**显式指定——协会的课要挂协会，协会也可能替商家建课。
func TestCreateCourseOrgIDRules(t *testing.T) {
	ctx := context.Background()
	svc := service.NewTrainingService(
		memory.NewCertificateRepository(), memory.NewCourseRepository(),
		memory.NewInstructorRepository(), memory.NewPilotRepository(nil),
	)
	const victim = "user-victim"

	cases := []struct {
		name  string
		actor domain.Actor
		give  string // 调用方尝试指定的归属（"" = 不指定）
		want  string
	}{
		{"个人指定他人归属 → 必须回落本人", domain.Actor{ID: "user-a", Role: domain.RoleIndividual}, victim, "user-a"},
		{"个人不指定 → 本人", domain.Actor{ID: "user-a", Role: domain.RoleIndividual}, "", "user-a"},
		{"企业指定他人归属 → 必须回落本人", domain.Actor{ID: "user-ent", Role: domain.RoleEnterprise}, victim, "user-ent"},
		{"协会管理员指定机构 → 尊重指定", domain.Actor{ID: "user-assocadmin", Role: domain.RoleAssociationAdmin}, "ent-association", "ent-association"},
		{"协会管理员不指定 → 回落本人", domain.Actor{ID: "user-assocadmin", Role: domain.RoleAssociationAdmin}, "", "user-assocadmin"},
		{"平台管理员指定机构 → 尊重指定", domain.Actor{ID: "user-padmin", Role: domain.RolePlatformAdmin}, "ent-association", "ent-association"},
	}
	for i, c := range cases {
		got, err := svc.CreateCourse(ctx, c.actor, domain.TrainingCourse{
			Title: "课", OrgID: c.give,
		})
		if err != nil {
			t.Fatalf("%s: CreateCourse: %v", c.name, err)
		}
		if got.OrgID != c.want {
			t.Fatalf("%s: OrgID=%q，want %q", c.name, got.OrgID, c.want)
		}
		if got.Status != "draft" {
			t.Fatalf("%s: 未指定 status 时默认应为 draft，实际 %q", c.name, got.Status)
		}
		_ = i
	}
}
