package httpapi

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository/memory"
	"drone-platform/internal/service"
)

// r4MinServer 构造最小 Server（仅装配 publicMatch/listAllAdapter 等死代码 handler 所需服务）。
func r4MinServer(t *testing.T) *Server {
	t.Helper()
	tokens, err := NewTokenManager("01234567890123456789012345678901")
	if err != nil {
		t.Fatal(err)
	}
	srv := NewServer(
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		nil, nil, nil, nil, nil, nil, nil, nil, nil,
		memory.NewUserRepository(nil), memory.NewRefreshTokenRepository(), tokens,
	)
	srv.demands = service.NewDemandService(memory.NewDemandRepository(nil))
	srv.jobSvc = service.NewJobService(memory.NewJobRepository(), memory.NewResumeRepository(), memory.NewJobApplicationRepository())
	return srv
}

// 以下测试直接覆盖未通过 HTTP 路由触达的纯函数/helper，
// 补齐 0% 覆盖（hashString/mapContractStatus/contains/matchCollegeType/
// matchCompetitionStatus/jobMutationCode/mutationErrorCode 等）。

func TestR4HashString(t *testing.T) {
	if hashString("") == 0 {
		t.Fatal("hashString empty should be non-zero (FNV offset basis)")
	}
	if hashString("abc") != hashString("abc") {
		t.Fatal("hashString must be deterministic")
	}
	if hashString("abc") == hashString("abd") {
		t.Fatal("different input should yield different hash")
	}
}

func TestR4MapContractStatus(t *testing.T) {
	cases := map[string]domain.ContractStatus{
		"sent":        domain.ContractSent,
		"created":     domain.ContractSent,
		"signing":     domain.ContractSigning,
		"in_progress": domain.ContractSigning,
		"signed":      domain.ContractSigned,
		"completed":   domain.ContractSigned,
		"voided":      domain.ContractVoided,
		"cancelled":   domain.ContractVoided,
		"expired":     domain.ContractExpired,
		"weird-xyz":   domain.ContractDraft, // unknown → draft
	}
	for in, want := range cases {
		if got := mapContractStatus(in); got != want {
			t.Fatalf("mapContractStatus(%q)=%q want %q", in, got, want)
		}
	}
}

func TestR4Contains(t *testing.T) {
	if !contains("无人机巡检服务", "巡检") {
		t.Fatal("contains should find substring")
	}
	if contains("无人机", "缺失") {
		t.Fatal("contains should not find missing substring")
	}
}

func TestR4MatchCollegeType(t *testing.T) {
	vocational := domain.College{Tags: []string{"专科"}}
	highVoc := domain.College{Tags: []string{"高职", "示范"}}
	under := domain.College{Tags: []string{"本科", "重点"}}
	none := domain.College{Tags: nil}

	if !matchCollegeType("vocational", vocational) {
		t.Fatal("专科 should be vocational")
	}
	if !matchCollegeType("vocational", highVoc) {
		t.Fatal("高职 should be vocational")
	}
	if matchCollegeType("vocational", under) {
		t.Fatal("本科 should not be vocational")
	}
	if !matchCollegeType("undergraduate", under) {
		t.Fatal("本科 should be undergraduate")
	}
	if matchCollegeType("undergraduate", vocational) {
		t.Fatal("专科 should not be undergraduate")
	}
	if !matchCollegeType("undergraduate", none) {
		t.Fatal("无 tags 默认按非专科(undergraduate)处理")
	}
}

func TestR4MatchCompetitionStatus(t *testing.T) {
	cases := []struct {
		query, status string
		want          bool
	}{
		{"enrolling", "published", true},
		{"open", "upcoming", true},
		{"ongoing", "in_progress", true},
		{"ongoing", "active", true},
		{"closed", "ended", true},
		{"full", "finished", true},
		{"enrolling", "closed", false},
		{"custom", "custom", true},
		{"custom", "other", false},
	}
	for _, c := range cases {
		if got := matchCompetitionStatus(c.query, c.status); got != c.want {
			t.Fatalf("matchCompetitionStatus(%q,%q)=%v want %v", c.query, c.status, got, c.want)
		}
	}
}

func TestR4JobMutationCode(t *testing.T) {
	if jobMutationCode(service.ErrInvalidJobTransition) != http.StatusConflict {
		t.Fatal("ErrInvalidJobTransition should map to 409")
	}
	if jobMutationCode(errors.New("only the publisher")) != http.StatusForbidden {
		t.Fatal("generic error should map to 403")
	}
}

func TestR4MutationErrorCode(t *testing.T) {
	if mutationErrorCode(errors.New("only the owner can advance")) != http.StatusForbidden {
		t.Fatal("owner mismatch should map to 403")
	}
	if mutationErrorCode(errors.New("not found")) != http.StatusNotFound {
		t.Fatal("not found should map to 404")
	}
}

func TestR4HTTPStatusToCode(t *testing.T) {
	cases := map[int]string{
		http.StatusBadRequest:          "VALIDATION_ERROR",
		http.StatusUnauthorized:        "UNAUTHENTICATED",
		http.StatusForbidden:           "FORBIDDEN",
		http.StatusNotFound:            "NOT_FOUND",
		http.StatusConflict:            "CONFLICT",
		http.StatusUnprocessableEntity: "STATE_INVALID",
		http.StatusTooManyRequests:     "RATE_LIMITED",
		http.StatusInternalServerError: "INTERNAL",
		999:                            "INTERNAL",
	}
	for code, want := range cases {
		if got := httpStatusToCode(code); got != want {
			t.Fatalf("httpStatusToCode(%d)=%q want %q", code, got, want)
		}
	}
}

func TestR4PaginationFromQuery(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	page, size := paginationFromQuery(r)
	if page != 1 || size != 20 {
		t.Fatalf("default pagination = (%d,%d) want (1,20)", page, size)
	}

	r = httptest.NewRequest(http.MethodGet, "/?page=3&page_size=50", nil)
	page, size = paginationFromQuery(r)
	if page != 3 || size != 50 {
		t.Fatalf("pagination = (%d,%d) want (3,50)", page, size)
	}

	r = httptest.NewRequest(http.MethodGet, "/?page=0&page_size=9999", nil)
	page, size = paginationFromQuery(r)
	if page != 1 || size != 100 {
		t.Fatalf("pagination clamp = (%d,%d) want (1,100)", page, size)
	}

	r = httptest.NewRequest(http.MethodGet, "/?page=abc&page_size=-1", nil)
	page, size = paginationFromQuery(r)
	if page != 1 || size != 20 {
		t.Fatalf("invalid pagination = (%d,%d) want (1,20)", page, size)
	}
}

func TestR4SlicePage(t *testing.T) {
	in := []int{1, 2, 3, 4, 5}
	got := slicePage(in, 2, 2).([]int)
	if len(got) != 2 || got[0] != 3 || got[1] != 4 {
		t.Fatalf("slicePage page2 size2 = %v", got)
	}

	got = slicePage(in, 9, 2).([]int)
	if len(got) != 0 {
		t.Fatalf("slicePage out-of-range should be empty, got %v", got)
	}

	got = slicePage(in, 1, 100).([]int)
	if len(got) != 5 {
		t.Fatalf("slicePage oversized page should clamp, got %v", got)
	}

	// 非 slice 输入原样返回
	if v := slicePage(42, 1, 2); v != 42 {
		t.Fatalf("slicePage non-slice should pass through, got %v", v)
	}
}

// TestR4DeadHandlers 覆盖未注册的死代码 handler：publicMatch/listAllAdapter/
// submitAdapter/updateAdapter 以及 SetAuditWriter/SetDBPinger setter。
func TestR4DeadHandlers(t *testing.T) {
	srv := r4MinServer(t)

	// publicMatch（跨源搜索：q 命中/缺省分页）
	r := httptest.NewRequest(http.MethodGet, "/?q=巡检", nil)
	w := httptest.NewRecorder()
	srv.publicMatch(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("publicMatch: %d %s", w.Code, w.Body.String())
	}

	// listAllAdapter：无 role/userId → 空数组
	r = httptest.NewRequest(http.MethodGet, "/", nil)
	w = httptest.NewRecorder()
	srv.listAllAdapter(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("listAllAdapter: %d %s", w.Code, w.Body.String())
	}

	// listAllAdapter：带 userId → listMyProjectApps
	srv.projectAppSvc = service.NewProjectAppService(memory.NewProjectAppRepository())
	r = httptest.NewRequest(http.MethodGet, "/?userId=u1", nil)
	r = r.WithContext(contextWithActor(r, domain.Actor{ID: "u1", Role: domain.RoleEnterprise}))
	w = httptest.NewRecorder()
	srv.listAllAdapter(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("listAllAdapter userId: %d %s", w.Code, w.Body.String())
	}

	// updateAdapter：恒 200
	r = httptest.NewRequest(http.MethodPost, "/", nil)
	w = httptest.NewRecorder()
	srv.updateAdapter(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("updateAdapter: %d", w.Code)
	}

	// submitAdapter → createDemand（需登录 actor）
	r = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"title":"x","contact":"13800000000"}`))
	r = r.WithContext(contextWithActor(r, domain.Actor{ID: "u1", Role: domain.RoleEnterprise}))
	w = httptest.NewRecorder()
	srv.submitAdapter(w, r)
	if w.Code != http.StatusCreated {
		t.Fatalf("submitAdapter: %d %s", w.Code, w.Body.String())
	}

	// SetAuditWriter / SetDBPinger setter（nil 安全）
	srv.SetAuditWriter(nil)
	srv.SetDBPinger(nil)
}
