package service_test

import (
	"context"
	"testing"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

// 发布职位的角色规则：企业是招聘主体，平台/协会管理员同样放行
// （协会账号是「运营方 + 市场主体」二合一——协会自己也要招人、也要替会员单位代发）。
// 个人被挡住：求职者是应聘方，不是招聘方。
func TestJobCreateRoleRules(t *testing.T) {
	ctx := context.Background()
	svc := service.NewJobService(memory.NewJobRepository(), memory.NewResumeRepository(), memory.NewJobApplicationRepository())

	allowed := []domain.Role{domain.RoleEnterprise, domain.RolePlatformAdmin, domain.RoleAssociationAdmin}
	for _, role := range allowed {
		if _, err := svc.CreateJob(ctx, domain.Actor{ID: "u-" + string(role), Role: role}, "巡检飞手", "描述", "渝北区", 100000); err != nil {
			t.Fatalf("角色 %s 应可发布职位，实际 %v", role, err)
		}
	}
	if _, err := svc.CreateJob(ctx, domain.Actor{ID: "u-ind", Role: domain.RoleIndividual}, "巡检飞手", "描述", "渝北区", 100000); err == nil {
		t.Fatal("个人不应能发布职位（求职者是应聘方）")
	}
}
