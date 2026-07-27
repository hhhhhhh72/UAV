// Package httpapi implements the drone platform's HTTP API.
//
// Architecture:
//   - Handlers parse requests, call services, and format responses (respond/fail/paginatedRespond).
//   - Handlers must NOT contain business logic, SQL queries, or JSON encoding.
//   - Middleware chain: idempotency → rate limit → request ID → panic recovery → security headers → CORS → auth.
//
// Key types:
//   - Server — holds all service dependencies and handles route registration.
//   - TokenManager — HMAC-SHA256 bearer token issuance and verification.
//   - rateLimiter — per-IP token bucket rate limiting.
//   - idempotencyStore — 24-hour idempotency key deduplication.
package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "drone-platform/docs"
	"drone-platform/internal/crypto"
	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
	"drone-platform/internal/service"

	httpSwagger "github.com/swaggo/http-swagger"
)

type requestIDKey struct{}

// Server is the HTTP API server. It holds all service dependencies and
// middleware state. Create one with NewServer and call Router() to get
// an http.Handler ready for ListenAndServe.
type Server struct {
	demands       *service.DemandService
	enterprises   *service.EnterpriseService
	enterpriseSvc *service.EnterpriseSvc
	employment    *service.EmploymentService
	contracts     *service.ContractService
	jobSvc        *service.JobService
	communitySvc  *service.CommunityService
	listingSvc    *service.ListingService
	labourSvc     *service.LabourService
	trainingSvc   *service.TrainingService
	tradingSvc    *service.TradingService
	insuranceSvc  *service.InsuranceService
	financeSvc    *service.FinanceService
	homeSvc       *service.HomeService
	fileSvc       *service.FileService
	msgSvc        *service.MessageService
	enrollSvc     *service.EnrollmentService
	expirySvc     *service.ExpiryService
	tradeSvc      *service.TradeOrderService
	escrowSvc     *service.EscrowService
	newsSvc       *service.NewsService
	reviewSvc     *service.ReviewService
	venueSvc      *service.VenueService
	expertSvc     *service.ExpertService
	caseSvc       *service.CaseService
	complianceSvc *service.ComplianceService
	reportSvc     *service.ReportService
	portfolioSvc  *service.PortfolioService
	achievementSvc *service.AchievementService
	rdService     *service.RDChallengeService
	researchSvc   *service.ResearchProjectService
	projectAppSvc *service.ProjectAppService
	competitionSvc *service.CompetitionService
	eventSvc      *service.EventService
	resourceSvc   *service.ResourceService
	emergencySvc  *service.EmergencyService
	matchingSvc   *service.MatchingService
poolSvc       *service.ResourcePoolService
	testSiteSvc   *service.TestSiteService
	exhibitionSvc *service.ExhibitionService
	transSvc       *service.TransformationService
	collegeSvc     *service.CollegeService
	coopSvc        *service.CooperationService
	rescueCaseSvc   *service.RescueCaseService
	emergDeptSvc   *service.EmergencyDeptService
	assocMemberSvc *service.AssociationMemberService
	userRepo      repository.UserRepository
	refreshRepo   repository.RefreshTokenRepository
	tokens        *TokenManager
	rateLimiter   *rateLimiter
	idempotency   *idempotencyStore
	auditWriter   repository.AuditWriter
	storage       string
}

type idempotencyStore struct {
	mu    sync.RWMutex
	cache map[string]idempotencyEntry
}

type idempotencyEntry struct {
	status   int
	body     string
	expiresAt time.Time
}

func newIdempotencyStore() *idempotencyStore {
	s := &idempotencyStore{cache: make(map[string]idempotencyEntry)}
	go s.cleanupLoop()
	return s
}

func (s *idempotencyStore) cleanupLoop() {
	for range time.Tick(time.Minute * 10) {
		s.mu.Lock()
		now := time.Now()
		for k, v := range s.cache {
			if now.After(v.expiresAt) {
				delete(s.cache, k)
			}
		}
		s.mu.Unlock()
	}
}

func (s *idempotencyStore) get(key string) (int, string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.cache[key]
	if !ok || time.Now().After(e.expiresAt) {
		return 0, "", false
	}
	return e.status, e.body, true
}

func (s *idempotencyStore) set(key string, status int, body string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cache[key] = idempotencyEntry{status: status, body: body, expiresAt: time.Now().Add(24 * time.Hour)}
}

type rateLimiter struct {
	mu       sync.Mutex
	buckets  map[string]*tokenBucket
	rate     int
	burst    int
	cleanupAt time.Time
}

type tokenBucket struct {
	tokens   float64
	lastTime time.Time
}

func newRateLimiter(ratePerSec, burst int) *rateLimiter {
	return &rateLimiter{
		buckets: make(map[string]*tokenBucket),
		rate:    ratePerSec,
		burst:   burst,
	}
}

func (rl *rateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Periodic cleanup of stale entries (every hour).
	if time.Now().After(rl.cleanupAt) {
		for k, b := range rl.buckets {
			if time.Since(b.lastTime) > time.Hour {
				delete(rl.buckets, k)
			}
		}
		rl.cleanupAt = time.Now().Add(time.Hour)
	}

	b, ok := rl.buckets[key]
	if !ok {
		b = &tokenBucket{tokens: float64(rl.burst), lastTime: time.Now()}
		rl.buckets[key] = b
	}

	elapsed := time.Since(b.lastTime).Seconds()
	b.tokens += elapsed * float64(rl.rate)
	if b.tokens > float64(rl.burst) {
		b.tokens = float64(rl.burst)
	}
	b.lastTime = time.Now()

	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

func NewServer(d *service.DemandService, e *service.EnterpriseService, es *service.EnterpriseSvc, h *service.EmploymentService, c *service.ContractService, js *service.JobService, cs *service.CommunityService, ls *service.ListingService, lbs *service.LabourService, ts *service.TrainingService, trs *service.TradingService, ins *service.InsuranceService, fin *service.FinanceService, hs *service.HomeService, fs *service.FileService, ms *service.MessageService, ens *service.EnrollmentService, exps *service.ExpiryService, trds *service.TradeOrderService, esc *service.EscrowService, nws *service.NewsService, rvs *service.ReviewService, vns *service.VenueService, ur repository.UserRepository, rr repository.RefreshTokenRepository, tokens *TokenManager) *Server {
	return &Server{demands: d, enterprises: e, enterpriseSvc: es, employment: h, contracts: c, jobSvc: js, communitySvc: cs, listingSvc: ls, labourSvc: lbs, trainingSvc: ts, tradingSvc: trs, insuranceSvc: ins, financeSvc: fin, homeSvc: hs, fileSvc: fs, msgSvc: ms, enrollSvc: ens, expirySvc: exps, tradeSvc: trds, escrowSvc: esc, newsSvc: nws, reviewSvc: rvs, venueSvc: vns, userRepo: ur, refreshRepo: rr, tokens: tokens, rateLimiter: newRateLimiter(100, 200), idempotency: newIdempotencyStore()}
}

// SetAuditWriter injects an audit log writer (typically the PG store).
func (s *Server) SetAuditWriter(w repository.AuditWriter) { s.auditWriter = w }

// SetStorage sets the storage backend name shown in health checks.
func (s *Server) SetStorage(name string) { s.storage = name }

// New business module service setters.
func (s *Server) SetExpertService(svc *service.ExpertService)               { s.expertSvc = svc }
func (s *Server) SetCaseService(svc *service.CaseService)                   { s.caseSvc = svc }
func (s *Server) SetComplianceService(svc *service.ComplianceService)       { s.complianceSvc = svc }
func (s *Server) SetReportService(svc *service.ReportService)               { s.reportSvc = svc }
func (s *Server) SetPortfolioService(svc *service.PortfolioService)         { s.portfolioSvc = svc }
func (s *Server) SetAchievementService(svc *service.AchievementService)     { s.achievementSvc = svc }
func (s *Server) SetRDChallengeService(svc *service.RDChallengeService)     { s.rdService = svc }
func (s *Server) SetResearchProjectService(svc *service.ResearchProjectService) { s.researchSvc = svc }
func (s *Server) SetProjectAppService(svc *service.ProjectAppService)       { s.projectAppSvc = svc }
func (s *Server) SetCompetitionService(svc *service.CompetitionService)     { s.competitionSvc = svc }
func (s *Server) SetEventService(svc *service.EventService)                 { s.eventSvc = svc }
func (s *Server) SetResourceService(svc *service.ResourceService)           { s.resourceSvc = svc }
func (s *Server) SetEmergencyService(svc *service.EmergencyService)         { s.emergencySvc = svc }
func (s *Server) SetMatchingService(svc *service.MatchingService)           { s.matchingSvc = svc }
func (s *Server) SetPoolService(svc *service.ResourcePoolService)         { s.poolSvc = svc }
func (s *Server) SetTestSiteService(svc *service.TestSiteService)       { s.testSiteSvc = svc }
func (s *Server) SetExhibitionService(svc *service.ExhibitionService)   { s.exhibitionSvc = svc }
func (s *Server) SetTransformationService(svc *service.TransformationService) { s.transSvc = svc }
func (s *Server) SetCollegeService(svc *service.CollegeService)           { s.collegeSvc = svc }
func (s *Server) SetCooperationService(svc *service.CooperationService)   { s.coopSvc = svc }
func (s *Server) SetRescueCaseService(svc *service.RescueCaseService)               { s.rescueCaseSvc = svc }
func (s *Server) SetEmergencyDeptService(svc *service.EmergencyDeptService)         { s.emergDeptSvc = svc }
func (s *Server) SetAssociationMemberService(svc *service.AssociationMemberService) { s.assocMemberSvc = svc }

// audit records a write operation to the audit log if a writer is configured.
func (s *Server) audit(ctx context.Context, actorID, action, resourceType, resourceID, result string) {
	if s.auditWriter == nil {
		return
	}
	rid := ""
	if v, ok := ctx.Value(requestIDKey{}).(string); ok {
		rid = v
	}
	_ = s.auditWriter.WriteAudit(context.Background(), repository.AuditEntry{
		ActorID:      actorID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Result:       result,
		RequestID:    rid,
	})
}

func (s *Server) Router() http.Handler {
	mux := http.NewServeMux()

	// ── Core routes (organized by module) ───────────────────────────
	s.registerCoreRoutes(mux)

	// ── Extended routes (batch / phase / biz modules) ───────────────
	s.registerPhase3Routes(mux)
	s.registerBatch3Routes(mux)
	s.registerBizRoutes(mux)
	s.registerBatch1Routes(mux)
	s.registerBatch2Routes(mux)

	// ── Legacy H5 /api/* compat routes — DEV ONLY ───────────────────
	// JSON file-backed storage. Disabled in production.
	// Remove after frontend migration to /api/v1/*.
	if adminDevMode() {
		slog.Warn("H5 compat routes enabled (ADMIN_DEV_MODE=true) — JSON file storage, NOT FOR PRODUCTION")
		s.registerCompatRoutes(mux)
		s.registerH5Compat(mux)
	}

	// ── Swagger UI ─────────────────────────────────────────────────
	if adminDevMode() {
		mux.HandleFunc("/swagger/", httpSwagger.Handler(
			httpSwagger.URL("/swagger/doc.json"),
		))
	}

	return s.idempotencyCheck(s.rateLimit(s.requestID(s.recoverPanic(s.securityHeaders(s.withCORS(s.authenticate(mux)))))))
}

func (s *Server) favicon(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Write([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32"><rect width="32" height="32" rx="6" fill="#2979FF"/><text x="16" y="23" text-anchor="middle" font-size="20">🚁</text></svg>`))
}

func (s *Server) serveUploads(w http.ResponseWriter, r *http.Request) {
	http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))).ServeHTTP(w, r)
}

func (s *Server) index(w http.ResponseWriter, r *http.Request) {
	respond(w, r, http.StatusOK, map[string]any{
		"name":      "Drone Industry Service Platform API",
		"status":    "running",
		"endpoints": []string{"/healthz", "/api/v1/home", "/api/v1/demands", "/api/v1/search"},
	})
}

func (s *Server) home(w http.ResponseWriter, r *http.Request) {
	city := r.URL.Query().Get("city")
	lat, _ := strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
	lng, _ := strconv.ParseFloat(r.URL.Query().Get("lng"), 64)
	data := s.homeSvc.GetHome(city, lat, lng)
	respond(w, r, http.StatusOK, map[string]any{
		"city":           data.City,
		"banners":        data.Banners,
		"quick_entries":  data.QuickEntries,
		"latest_demands": data.HotDemands,
			"notices":        data.Notices,
	})
}
func (s *Server) search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	ds, err := s.demands.Search(q)
	if err != nil {
		slog.Warn("search: demands failed", "error", err)
		ds = nil
	}
	es, err := s.enterprises.Search(q)
	if err != nil {
		slog.Warn("search: enterprises failed", "error", err)
		es = nil
	}
	public := make([]domain.Demand, len(ds))
	for i, d := range ds {
		public[i] = publicDemand(d)
	}
	for i := range es {
		if es[i].AccountName != "" {
			es[i].AccountName = crypto.MaskPhone(es[i].AccountName)
		}
	}
	respond(w, r, http.StatusOK, map[string]any{"demands": public, "enterprises": es})
}
func (s *Server) listDemands(w http.ResponseWriter, r *http.Request) {
	page, pageSize := paginationFromQuery(r)
	result, err := s.demands.List(repository.DemandFilter{District: r.URL.Query().Get("district"), BizType: r.URL.Query().Get("biz_type"), Sort: r.URL.Query().Get("sort")})
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	// Keyword search filter
	if q := r.URL.Query().Get("q"); q != "" {
		qs := strings.ToLower(q)
		filtered := make([]domain.Demand, 0, len(result))
		for _, d := range result {
			if strings.Contains(strings.ToLower(d.Title), qs) ||
				strings.Contains(strings.ToLower(d.Description), qs) ||
				strings.Contains(strings.ToLower(d.PublisherName), qs) {
				filtered = append(filtered, d)
			}
		}
		result = filtered
	}
	// Mine filter: only demands published by current user
	if r.URL.Query().Get("mine") == "1" {
		if a, ok := authenticatedActor(r); ok {
			filtered := make([]domain.Demand, 0, len(result))
			for _, d := range result {
				if d.PublisherID == a.ID {
					filtered = append(filtered, d)
				}
			}
			result = filtered
		}
	}
	start := (page - 1) * pageSize
	if start > len(result) {
		start = len(result)
	}
	end := start + pageSize
	if end > len(result) {
		end = len(result)
	}
	paged := result[start:end]
	public := make([]domain.Demand, len(paged))
	for i, d := range paged {
		public[i] = publicDemand(d)
	}
	paginatedRespond(w, r, public, len(result))
}
// GET /api/v1/admin/demands — 管理员查看所有状态需求
func (s *Server) listAdminDemands(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok || (a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin) {
		fail(w, r, http.StatusForbidden, errors.New("admin permission required"))
		return
	}
	st := r.URL.Query().Get("status")
	if st == "" { st = "all" }
	result, err := s.demands.List(repository.DemandFilter{Status: st})
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	page, pageSize := paginationFromQuery(r)
	start := (page - 1) * pageSize
	if start > len(result) { start = len(result) }
	end := start + pageSize
	if end > len(result) { end = len(result) }
	paginatedRespond(w, r, result[start:end], len(result))
}

func publicDemand(d domain.Demand) domain.Demand {
	d.PublisherID = ""
	d.Contact = ""
	d.Latitude = 0
	d.Longitude = 0
	return d
}
func (s *Server) createDemand(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	var in service.CreateDemandInput
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	d, err := s.demands.Create(a, in)
	if err != nil {
		fail(w, r, http.StatusForbidden, err)
		return
	}
	s.audit(r.Context(), a.ID, "create_demand", "demand", d.ID, "created")
	respond(w, r, http.StatusCreated, d)
}
func (s *Server) approveDemand(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	d, err := s.demands.Approve(a, r.PathValue("id"))
	if err != nil {
		code := http.StatusForbidden
		if strings.Contains(err.Error(), "not found") {
			code = http.StatusNotFound
		}
		fail(w, r, code, err)
		return
	}
	if d.Contact != "" {
		d.Contact = crypto.MaskPhone(d.Contact)
	}
	s.audit(r.Context(), a.ID, "approve_demand", "demand", d.ID, "approved")
	respond(w, r, http.StatusOK, d)
}
func (s *Server) listEmployment(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	page, pageSize := paginationFromQuery(r)
	offset := (page - 1) * pageSize
	out, total, err := s.employment.List(a, offset, pageSize)
	if err != nil {
		fail(w, r, http.StatusForbidden, err)
		return
	}
	paginatedRespond(w, r, out, total)
}
func (s *Server) createEmployment(w http.ResponseWriter, r *http.Request) {
	var v domain.EmploymentRequest
	if err := decode(r, &v); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	out, err := s.employment.Create(a, v)
	if err != nil {
		fail(w, r, http.StatusForbidden, err)
		return
	}
	respond(w, r, http.StatusCreated, out)
}
func (s *Server) listContracts(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	page, pageSize := paginationFromQuery(r)
	offset := (page - 1) * pageSize
	out, total, err := s.contracts.List(a, offset, pageSize)
	if err != nil {
		fail(w, r, http.StatusForbidden, err)
		return
	}
	paginatedRespond(w, r, out, total)
}
func (s *Server) createContract(w http.ResponseWriter, r *http.Request) {
	var v domain.Contract
	if err := decode(r, &v); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	out, err := s.contracts.Create(a, v)
	if err != nil {
		fail(w, r, http.StatusForbidden, err)
		return
	}
	respond(w, r, http.StatusCreated, out)
}
func (s *Server) pendingEnterprises(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	out, err := s.enterprises.Pending(a)
	if err != nil {
		fail(w, r, http.StatusForbidden, err)
		return
	}
	respond(w, r, http.StatusOK, out)
}

func (s *Server) idempotencyCheck(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("Idempotency-Key")
		if key == "" || (r.Method != http.MethodPost && r.Method != http.MethodPatch) {
			next.ServeHTTP(w, r)
			return
		}
		// Check for previously completed request.
		if status, body, ok := s.idempotency.get(key); ok {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(status)
			w.Write([]byte(body))
			return
		}
		// lenKey := len(key); if lenKey < 8 || lenKey > 128 { fail(w, r, http.StatusBadRequest, errors.New("Idempotency-Key must be 8-128 characters")); return }
			// Capture the response via a response recorder.
		rec := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rec, r)
		// Store the result for future idempotent requests.
		s.idempotency.set(key, rec.statusCode, rec.body.String())
		// Write the actual response (already written to rec.wrapped).
	})
}

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	body       strings.Builder
}

func (r *responseRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

func (s *Server) rateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Use X-Forwarded-For or RemoteAddr as the rate-limit key.
		key := r.Header.Get("X-Forwarded-For")
		if key == "" {
			key = r.RemoteAddr
		}
		if !s.rateLimiter.allow(key) {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			rid := ""
			if v, ok := r.Context().Value(requestIDKey{}).(string); ok {
				rid = v
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error":      map[string]string{"code": "RATE_LIMITED", "message": "请求过于频繁，请稍后重试"},
				"request_id": rid,
			})
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rid := r.Header.Get("X-Request-ID")
		if rid == "" {
			rid = fmt.Sprintf("req_%s%d", strings.ReplaceAll(time.Now().UTC().Format("20060102T150405"), "", ""), rand.Intn(10000))
		}
		w.Header().Set("X-Request-ID", rid)
		ctx := context.WithValue(r.Context(), requestIDKey{}, rid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Server) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				slog.Error("panic recovered", "method", r.Method, "path", r.URL.Path, "panic", rec, "stack", string(debug.Stack()))
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(map[string]any{
					"error": map[string]string{"code": "INTERNAL", "message": "internal server error"},
				})
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (s *Server) securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		next.ServeHTTP(w, r)
	})
}

func (s *Server) withCORS(next http.Handler) http.Handler {
	allowedOrigins := s.allowedCORSOrigins()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		allowed := ""
		for _, o := range allowedOrigins {
			if o == origin {
				allowed = origin
				break
			}
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if allowed != "" {
			w.Header().Set("Access-Control-Allow-Origin", allowed)
		}
		w.Header().Set("Vary", "Origin")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
func (s *Server) allowedCORSOrigins() []string {
	raw := os.Getenv("CORS_ORIGINS")
	if raw == "" {
		return []string{"http://localhost:3000"}
	}
	return strings.Split(raw, ",")
}

func decode(r *http.Request, v any) error {
	d := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	d.DisallowUnknownFields()
	return d.Decode(v)
}
func paginationFromQuery(r *http.Request) (page, pageSize int) {
	page, pageSize = 1, 20
	if p, err := strconv.Atoi(r.URL.Query().Get("page")); err == nil && p > 0 {
		page = p
	}
	if ps, err := strconv.Atoi(r.URL.Query().Get("page_size")); err == nil && ps > 0 {
		pageSize = ps
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return
}

func paginatedRespond(w http.ResponseWriter, r *http.Request, items any, total int) {
	page, pageSize := paginationFromQuery(r)
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"data":       items,
		"page":       page,
		"page_size":  pageSize,
		"total":      total,
		"request_id": requestIDFromCtx(r),
	})
}

func requestIDFromCtx(r *http.Request) string {
	if rid, ok := r.Context().Value(requestIDKey{}).(string); ok {
		return rid
	}
	return ""
}

func respond(w http.ResponseWriter, r *http.Request, status int, v any) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"data": v, "request_id": requestIDFromCtx(r)})
}
func fail(w http.ResponseWriter, r *http.Request, status int, err error) {
	if errors.Is(err, io.EOF) {
		err = errors.New("request body is required")
	}
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error":      map[string]string{"code": httpStatusToCode(status), "message": strings.TrimSpace(err.Error())},
		"request_id": requestIDFromCtx(r),
	})
}

func httpStatusToCode(status int) string {
	switch status {
	case http.StatusBadRequest:
		return "VALIDATION_ERROR"
	case http.StatusUnauthorized:
		return "UNAUTHENTICATED"
	case http.StatusForbidden:
		return "FORBIDDEN"
	case http.StatusNotFound:
		return "NOT_FOUND"
	case http.StatusConflict:
		return "CONFLICT"
	case http.StatusUnprocessableEntity:
		return "STATE_INVALID"
	case http.StatusTooManyRequests:
		return "RATE_LIMITED"
	default:
		return "INTERNAL"
	}
}
