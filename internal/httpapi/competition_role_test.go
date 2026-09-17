package httpapi_test

import (
	"net/http"
	"testing"

	"drone-platform/internal/domain"
)

// 赛事发布权规则：企业是办赛主体；协会是办赛的**天然主体**——协会账号是
// 「运营方 + 市场主体」二合一，既登后台审核内容，又下场办赛事；
// 平台管理员同样放行。个人是参赛/报名方，被挡住。
// 三者发布后一律进 pending，等管理端审核通过才公开。
func TestCreateCompetitionRoleRules(t *testing.T) {
	app := newBizServer(t)
	body := `{"title":"无人机竞速赛","start_date":"2026-10-01","end_date":"2026-10-02"}`

	for _, role := range []domain.Role{
		domain.RoleEnterprise, domain.RoleAssociationAdmin, domain.RolePlatformAdmin,
	} {
		w := doRaw(app, http.MethodPost, "/api/v1/competitions", body, authAs(t, authRoleUserID(role), role))
		if w.Code != http.StatusCreated {
			t.Fatalf("角色 %s 应可发布赛事，code=%d body=%s", role, w.Code, w.Body.String())
		}
	}

	// 个人被挡：参赛者是报名方，不是办赛方
	w := doRaw(app, http.MethodPost, "/api/v1/competitions", body,
		authAs(t, authRoleUserID(domain.RoleIndividual), domain.RoleIndividual))
	if w.Code != http.StatusForbidden {
		t.Fatalf("个人不应能发布赛事，code=%d body=%s", w.Code, w.Body.String())
	}
}
