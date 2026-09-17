package service_test

import (
	"context"
	"errors"
	"testing"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

// 需求发布权的有效规则：企业/个人是市场主体，平台管理员与协会管理员是「运营方」，
// 其中协会账号只有一个、二合一（登后台审核、登小程序就是协会这个机构本身）。
//
// 四种角色全部放行后，服务层那条判断退化为**默认拒绝**兜底——只挡将来新增却没想清楚
// 发布权的角色。这条兜底在 HTTP 层测不到（authenticate 中间件会先以 401 拒掉未知角色），
// 所以必须在服务层钉住。
func TestDemandCreateRejectsUnknownRole(t *testing.T) {
	ctx := context.Background()
	svc := service.NewDemandService(memory.NewDemandRepository(nil))
	in := service.CreateDemandInput{Title: "测试需求", Contact: "13800000000"}

	// 四种已知角色都能发
	for _, role := range []domain.Role{
		domain.RoleEnterprise, domain.RoleIndividual,
		domain.RolePlatformAdmin, domain.RoleAssociationAdmin,
	} {
		if _, err := svc.Create(ctx, domain.Actor{ID: "u-" + string(role), Role: role}, in); err != nil {
			t.Fatalf("角色 %s 应可发布需求，实际 %v", role, err)
		}
	}

	// 未知角色：默认拒绝
	_, err := svc.Create(ctx, domain.Actor{ID: "u-ghost", Role: domain.Role("ghost_role")}, in)
	if err == nil {
		t.Fatal("未知角色应被拒绝（默认拒绝兜底），实际放行了")
	}
	if !errors.Is(err, service.ErrRoleNotAllowed) {
		t.Fatalf("未知角色应返回 ErrRoleNotAllowed，实际 %v", err)
	}
}
