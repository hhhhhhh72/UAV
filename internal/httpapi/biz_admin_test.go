package httpapi_test

import (
	"net/http"
	"testing"

	"drone-platform/internal/domain"
)

func TestBizAdminCreateExpert(t *testing.T) {
	app := newBizServer(t)
	body := []byte(`{"name":"张教授","title":"教授","org":"重庆大学","field":"无人机","bio":"专家简介"}`)
	w := request(t, app, http.MethodPost, "/api/v1/admin/experts", body, domain.RoleAssociationAdmin)
	if w.Code != http.StatusCreated { t.Fatalf("create expert: %d %s", w.Code, w.Body.String()) }
}

func TestBizAdminCreateCompetition(t *testing.T) {
	app := newBizServer(t)
	body := []byte(`{"Title":"竞速赛","Category":"racing","Description":"FPV","Location":"巴南","StartDate":"2026-09-01","EndDate":"2026-09-03","max_teams":50,"Sponsor":"协会"}`)
	w := request(t, app, http.MethodPost, "/api/v1/admin/competitions", body, domain.RoleAssociationAdmin)
	if w.Code != http.StatusCreated { t.Fatalf("create competition: %d %s", w.Code, w.Body.String()) }
}

func TestBizAdminCreateEvent(t *testing.T) {
	app := newBizServer(t)
	body := []byte(`{"Title":"产业论坛","EventType":"forum","Description":"年度论坛","Location":"博览中心","StartTime":"2026-10-01T10:00:00Z","EndTime":"2026-10-02T18:00:00Z","max_attendees":500}`)
	w := request(t, app, http.MethodPost, "/api/v1/admin/events", body, domain.RoleAssociationAdmin)
	if w.Code != http.StatusCreated { t.Fatalf("create event: %d %s", w.Code, w.Body.String()) }
}

func TestBizNonAdminCannotCreateExpert(t *testing.T) {
	app := newBizServer(t)
	body := []byte(`{"name":"hacker","title":"x","org":"x","field":"x","bio":"x"}`)
	w := request(t, app, http.MethodPost, "/api/v1/admin/experts", body, domain.RoleIndividual)
	if w.Code != http.StatusForbidden { t.Fatalf("individual should be forbidden: %d", w.Code) }
}

func TestBizRDChallengeList(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodGet, "/api/v1/rd-challenges", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK { t.Fatalf("GET challenges: %d", w.Code) }
}

func TestBizResearchProjectList(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodGet, "/api/v1/research-projects", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK { t.Fatalf("GET projects: %d", w.Code) }
}

func TestBizPortfolioList(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodGet, "/api/v1/portfolios", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK { t.Fatalf("GET portfolios: %d", w.Code) }
}

func TestBizSearchAndMatch(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodGet, "/api/v1/match?q=巡检&limit=5", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK { t.Fatalf("GET match: %d %s", w.Code, w.Body.String()) }
}
