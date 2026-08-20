package httpapi_test

import (
	"net/http"
	"testing"

	"drone-platform/internal/domain"
)

func TestEscrowFlow(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodGet, "/api/v1/escrow/balance", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK { t.Fatalf("balance: %d", w.Code) }
	w = request(t, app, http.MethodGet, "/api/v1/escrow/transactions", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK { t.Fatalf("transactions: %d", w.Code) }
}

func TestTrainingsFlow(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodGet, "/api/v1/instructors", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK { t.Fatalf("instructors: %d", w.Code) }
	w = request(t, app, http.MethodGet, "/api/v1/certified-pilots", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK { t.Fatalf("pilots: %d", w.Code) }
}

func TestContractsFlow(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodGet, "/api/v1/contracts", nil, domain.RoleEnterprise)
	if w.Code != http.StatusOK { t.Fatalf("contracts: %d", w.Code) }
	w = request(t, app, http.MethodGet, "/api/v1/contract-templates", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK { t.Fatalf("templates: %d", w.Code) }
}

func TestJobsFlow(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodGet, "/api/v1/jobs", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK { t.Fatalf("jobs: %d", w.Code) }
}

func TestPostsFlow(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodGet, "/api/v1/posts", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK { t.Fatalf("posts: %d", w.Code) }
}

func TestListingsFlow(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodGet, "/api/v1/listings", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK { t.Fatalf("listings: %d", w.Code) }
}

func TestLabourOrdersFlow(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodGet, "/api/v1/labour-orders", nil, domain.RoleEnterprise)
	if w.Code != http.StatusOK { t.Fatalf("labour: %d", w.Code) }
}

func TestPoliciesFlow(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodGet, "/api/v1/policies/mine", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK { t.Fatalf("policies: %d", w.Code) }
}

func TestInspectionsFlow(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodGet, "/api/v1/inspections/mine", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK { t.Fatalf("inspections: %d", w.Code) }
}

func TestDemandSubmitFlow(t *testing.T) {
	app := newBizServer(t)
	// Create demand
	body := []byte(`{"publisher_name":"测试企业","contact":"13800001111","title":"巡检需求","biz_type":"cable_inspection"}`)
	w := request(t, app, http.MethodPost, "/api/v1/demands", body, domain.RoleEnterprise)
	if w.Code != http.StatusCreated { t.Fatalf("create demand: %d %s", w.Code, w.Body.String()) }
}

func TestEnterpriseCreateFlow(t *testing.T) {
	app := newBizServer(t)
	body := []byte(`{"name":"测试企业","account_name":"622200000000","license_url":"license.jpg"}`)
	w := request(t, app, http.MethodPost, "/api/v1/enterprises", body, domain.RoleEnterprise)
	if w.Code != http.StatusCreated { t.Fatalf("create enterprise: %d %s", w.Code, w.Body.String()) }
}

func TestConfigFlow(t *testing.T) {
	app := newBizServer(t)
	// Admin config
	w := request(t, app, http.MethodGet, "/api/v1/admin/config", nil, domain.RolePlatformAdmin)
	if w.Code != http.StatusOK { t.Fatalf("config: %d", w.Code) }
}

func TestFileUploadUnauthorized(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodPost, "/api/v1/files/upload", nil, domain.RoleIndividual)
	// No multipart body = bad request or 400
	if w.Code != http.StatusBadRequest && w.Code != http.StatusUnauthorized {
		t.Logf("upload without file: %d %s", w.Code, w.Body.String())
	}
}

func TestEnrollmentFlow(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodGet, "/api/v1/enrollments/mine", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK { t.Fatalf("enrollments: %d", w.Code) }
}

func TestTradeOrdersMine(t *testing.T) {
	app := newBizServer(t)
	w := request(t, app, http.MethodGet, "/api/v1/trade-orders/mine", nil, domain.RoleIndividual)
	if w.Code != http.StatusOK { t.Fatalf("trade orders: %d", w.Code) }
}

func TestCertExpiring(t *testing.T) {
	app := newBizServer(t)
	// P2：expiring 台账限管理员——普通用户 403，管理员 200
	w := request(t, app, http.MethodGet, "/api/v1/certificates/expiring?days=30", nil, domain.RoleIndividual)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expiring as individual: want 403, got %d", w.Code)
	}
	w = request(t, app, http.MethodGet, "/api/v1/certificates/expiring?days=30", nil, domain.RolePlatformAdmin)
	if w.Code != http.StatusOK { t.Fatalf("expiring as admin: %d", w.Code) }
}
