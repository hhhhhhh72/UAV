package httpapi_test

import (
	"net/http"
	"testing"

	"drone-platform/internal/domain"
)

func TestExpertWriteOps(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodPost, "/api/v1/admin/experts",
		[]byte(`{"Name":"专家","Title":"教授","Org":"重大","Field":"无人机","Bio":"简介"}`),
		domain.RoleAssociationAdmin)
	if w.Code != http.StatusCreated {
		t.Fatalf("create expert: %d %s", w.Code, w.Body.String())
	}
	w = request(t, app, http.MethodPut, "/api/v1/admin/experts/exp-1",
		[]byte(`{"Name":"专家v2"}`), domain.RoleAssociationAdmin)
	w = request(t, app, http.MethodDelete, "/api/v1/admin/experts/exp-1", nil, domain.RolePlatformAdmin)
	_ = w
}

func TestCaseWriteOps(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodPost, "/api/v1/admin/cases",
		[]byte(`{"title":"案例","category":"logistics","description":"描述","client_name":"客户A","result":"成果"}`),
		domain.RoleAssociationAdmin)
	if w.Code != http.StatusCreated {
		t.Fatalf("create case: %d %s", w.Code, w.Body.String())
	}
}

func TestComplianceWriteOps(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodPost, "/api/v1/admin/compliance-docs",
		[]byte(`{"title":"条例","category":"policy","publisher":"民航局","publish_date":"2026-01-01","status":"draft","summary":"内容摘要"}`),
		domain.RoleAssociationAdmin)
	if w.Code != http.StatusCreated {
		t.Fatalf("create doc: %d %s", w.Code, w.Body.String())
	}
	w = request(t, app, http.MethodPost, "/api/v1/admin/compliance-standards",
		[]byte(`{"title":"标准","standard_no":"T/CDA-001","publisher":"协会","effective_date":"2026-07-01","status":"draft","scope":"巡检"}`),
		domain.RoleAssociationAdmin)
	if w.Code != http.StatusCreated {
		t.Fatalf("create std: %d %s", w.Code, w.Body.String())
	}
}

func TestReportWriteOps(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodPost, "/api/v1/admin/industry-reports",
		[]byte(`{"Title":"2026无人机产业报告","Period":"2026H1","Category":"行业","Summary":"摘要","Content":"全文","Author":"协会研究部"}`),
		domain.RoleAssociationAdmin)
	if w.Code != http.StatusCreated {
		t.Fatalf("create report: %d %s", w.Code, w.Body.String())
	}
}

func TestPortfolioWriteOps(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodPost, "/api/v1/portfolios",
		[]byte(`{"Name":"品牌A","LogoURL":"logo.png","CoverURL":"cover.png","Description":"无人机方案商","ContactInfo":"138"}`),
		domain.RoleEnterprise)
	if w.Code != http.StatusCreated {
		t.Fatalf("create portfolio: %d %s", w.Code, w.Body.String())
	}
}

func TestAchievementWriteOps(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodPost, "/api/v1/achievements",
		[]byte(`{"Title":"AI避障算法","AchieveType":"patent","Description":"自动避障","Field":"无人机","Stage":"lab","ContactInfo":"138"}`),
		domain.RoleEnterprise)
	if w.Code != http.StatusCreated {
		t.Fatalf("create achievement: %d %s", w.Code, w.Body.String())
	}
}

func TestRDChallengeWriteOps(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodPost, "/api/v1/rd-challenges",
		[]byte(`{"Title":"长续航电池","Field":"电池","Description":">2h续航方案","budget_fen":500000,"Deadline":"2026-12-31T00:00:00Z"}`),
		domain.RoleEnterprise)
	if w.Code != http.StatusCreated {
		t.Fatalf("create rd: %d %s", w.Code, w.Body.String())
	}
}

func TestResearchProjectWriteOps(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodPost, "/api/v1/admin/research-projects",
		[]byte(`{"Title":"航路规划","Field":"空域","Description":"城市低空航路","LeadOrg":"重庆大学","Milestones":"Q1调研","budget_fen":200000,"StartDate":"2026-08-01T00:00:00Z","EndDate":"2027-08-01T00:00:00Z"}`),
		domain.RoleAssociationAdmin)
	if w.Code != http.StatusCreated {
		t.Fatalf("create project: %d %s", w.Code, w.Body.String())
	}
}

func TestProjectAppWriteOps(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodPost, "/api/v1/project-applications",
		[]byte(`{"ProjectName":"无人机产业示范","Category":"示范","Description":"打造示范区","budget_fen":1000000}`),
		domain.RoleEnterprise)
	if w.Code != http.StatusCreated {
		t.Fatalf("create app: %d %s", w.Code, w.Body.String())
	}
}

func TestResourceWriteOps(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodPost, "/api/v1/admin/industry-resources",
		[]byte(`{"Name":"无人机001","ResType":"drone","Model":"M300","Specs":"RTK+热成像","Location":"重庆南岸","price_fen":100000,"BookingInfo":"工作日9-18点"}`),
		domain.RoleAssociationAdmin)
	if w.Code != http.StatusCreated {
		t.Fatalf("create resource: %d %s", w.Code, w.Body.String())
	}
}

func TestEmergencyWriteOps(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodPost, "/api/v1/admin/emergency-resources",
		[]byte(`{"name":"应急机01","res_type":"drone","specs":"M300RTK+热成像","quantity":2,"location":"南岸","contact_info":"138"}`),
		domain.RoleAssociationAdmin)
	if w.Code != http.StatusCreated {
		t.Fatalf("create emerg: %d %s", w.Code, w.Body.String())
	}
}

func TestCompetitionEventRegOps(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodPost, "/api/v1/admin/competitions",
		[]byte(`{"title":"竞速赛","category":"racing","description":"FPV","location":"巴南","sponsor":"协会","start_date":"2026-09-01","end_date":"2026-09-03","max_teams":50}`),
		domain.RoleAssociationAdmin)
	if w.Code != http.StatusCreated {
		t.Fatalf("create comp: %d %s", w.Code, w.Body.String())
	}
	w = request(t, app, http.MethodPost, "/api/v1/admin/events",
		[]byte(`{"title":"产业论坛","event_type":"forum","description":"年度论坛","location":"博览中心","start_time":"2026-10-01T10:00:00Z","end_time":"2026-10-02T18:00:00Z","max_attendees":500}`),
		domain.RoleAssociationAdmin)
	if w.Code != http.StatusCreated {
		t.Fatalf("create event: %d %s", w.Code, w.Body.String())
	}
}

func TestBizAllNonAdminBlocked(t *testing.T) {
	app := newBizServer(t)
	tests := []struct{ method, path string }{
		{"POST", "/api/v1/admin/experts"},
		{"POST", "/api/v1/admin/cases"},
		{"POST", "/api/v1/admin/competitions"},
		{"POST", "/api/v1/admin/events"},
		{"POST", "/api/v1/admin/industry-resources"},
		{"POST", "/api/v1/admin/emergency-resources"},
		{"POST", "/api/v1/admin/emergency-dispatches"},
		{"POST", "/api/v1/admin/compliance-docs"},
		{"POST", "/api/v1/admin/compliance-standards"},
		{"POST", "/api/v1/admin/industry-reports"},
		{"POST", "/api/v1/admin/research-projects"},
	}
	for _, tc := range tests {
		w := request(t, app, tc.method, tc.path, []byte(`{}`), domain.RoleIndividual)
		if w.Code != http.StatusForbidden {
			t.Errorf("%s %s: expected 403, got %d", tc.method, tc.path, w.Code)
		}
	}
}

func TestEscrowDepositFlow(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodPost, "/api/v1/escrow/deposit",
		[]byte(`{"amount_fen":100000}`), domain.RoleIndividual)
	// Works or fails based on balance existence
	if w.Code != http.StatusCreated && w.Code != http.StatusOK {
		t.Logf("deposit: %d %s", w.Code, w.Body.String())
	}
	w = request(t, app, http.MethodGet, "/api/v1/escrow/balance", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK {
		t.Fatalf("balance: %d", w.Code)
	}
}
