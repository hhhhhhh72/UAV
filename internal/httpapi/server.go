// Package httpapi implements the drone platform's HTTP API.
//
// Architecture:
//   - Handlers parse requests, call services, and format responses (respond/fail/paginatedRespond).
//   - Handlers must NOT contain business logic, SQL queries, or JSON encoding.
//   - Middleware chain: rate limit → request ID → panic recovery → security headers → CORS → auth → idempotency.
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
	"reflect"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "drone-platform/docs"
	"drone-platform/internal/crypto"
	"drone-platform/internal/domain"
	"drone-platform/internal/middleware"
	"drone-platform/internal/repository"
	"drone-platform/internal/service"

	httpSwagger "github.com/swaggo/http-swagger"
)

type requestIDKey struct{}

// Server is the HTTP API server. It holds all service dependencies and
// middleware state. Create one with NewServer and call Router() to get
// an http.Handler ready for ListenAndServe.
type Server struct {
	demands           *service.DemandService
	enterprises       *service.EnterpriseService
	enterpriseSvc     *service.EnterpriseSvc
	employment        *service.EmploymentService
	contracts         *service.ContractService
	jobSvc            *service.JobService
	communitySvc      *service.CommunityService
	listingSvc        *service.ListingService
	labourSvc         *service.LabourService
	trainingSvc       *service.TrainingService
	tradingSvc        *service.TradingService
	insuranceSvc      *service.InsuranceService
	financeSvc        *service.FinanceService
	homeSvc           *service.HomeService
	fileSvc           *service.FileService
	msgSvc            *service.MessageService
	enrollSvc         *service.EnrollmentService
	expirySvc         *service.ExpiryService
	tradeSvc          *service.TradeOrderService
	escrowSvc         *service.EscrowService
	newsSvc           *service.NewsService
	reviewSvc         *service.ReviewService
	venueSvc          *service.VenueService
	expertSvc         *service.ExpertService
	caseSvc           *service.CaseService
	complianceSvc     *service.ComplianceService
	reportSvc         *service.ReportService
	portfolioSvc      *service.PortfolioService
	achievementSvc    *service.AchievementService
	rdService         *service.RDChallengeService
	researchSvc       *service.ResearchProjectService
	projectAppSvc     *service.ProjectAppService
	competitionSvc    *service.CompetitionService
	serviceListingSvc *service.ServiceListingService
	eventSvc          *service.EventService
	resourceSvc       *service.ResourceService
	emergencySvc      *service.EmergencyService
	matchingSvc       *service.MatchingService
	intentSvc         *service.IntentService
	workOrderSvc      *service.WorkOrderService
	poolSvc           *service.ResourcePoolService
	testSiteSvc       *service.TestSiteService
	exhibitionSvc     *service.ExhibitionService
	transSvc          *service.TransformationService
	collegeSvc        *service.CollegeService
	studyTourRepo     repository.StudyTourRepository
	coopSvc           *service.CooperationService
	rescueCaseSvc     *service.RescueCaseService
	emergDeptSvc      *service.EmergencyDeptService
	assocMemberSvc    *service.AssociationMemberService
	contractTplSvc    *service.ContractTemplateService
	appSvc            *service.ApplicationService
	userRepo          repository.UserRepository
	refreshRepo       repository.RefreshTokenRepository
	tokens            *TokenManager
	rateLimiter       *rateLimiter
	idempotency       *idempotencyStore
	auditWriter       repository.AuditWriter
	dbPinger          interface{ Ping(context.Context) error }
	storage           string
}

type idempotencyStore struct {
	mu      sync.RWMutex
	cache   map[string]idempotencyEntry
	flights map[string]*idempotencyFlight
}

type idempotencyEntry struct {
	status    int
	body      string
	expiresAt time.Time
}

// idempotencyFlight 并发同 key 单飞：首个请求执行，其余等待并复用其响应，
// 防止网络重试风暴期重复执行写操作（P1 修复）。
type idempotencyFlight struct {
	done   chan struct{}
	status int
	body   string
}

func newIdempotencyStore() *idempotencyStore {
	s := &idempotencyStore{cache: make(map[string]idempotencyEntry), flights: make(map[string]*idempotencyFlight)}
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
	mu        sync.Mutex
	buckets   map[string]*tokenBucket
	rate      int
	burst     int
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

// SetDBPinger injects a database prober for /healthz (typically the pgx pool).
func (s *Server) SetDBPinger(p interface{ Ping(context.Context) error }) { s.dbPinger = p }

// SetStorage sets the storage backend name shown in health checks.
func (s *Server) SetStorage(name string) { s.storage = name }

// New business module service setters.
func (s *Server) SetExpertService(svc *service.ExpertService)                   { s.expertSvc = svc }
func (s *Server) SetCaseService(svc *service.CaseService)                       { s.caseSvc = svc }
func (s *Server) SetComplianceService(svc *service.ComplianceService)           { s.complianceSvc = svc }
func (s *Server) SetReportService(svc *service.ReportService)                   { s.reportSvc = svc }
func (s *Server) SetPortfolioService(svc *service.PortfolioService)             { s.portfolioSvc = svc }
func (s *Server) SetAchievementService(svc *service.AchievementService)         { s.achievementSvc = svc }
func (s *Server) SetRDChallengeService(svc *service.RDChallengeService)         { s.rdService = svc }
func (s *Server) SetResearchProjectService(svc *service.ResearchProjectService) { s.researchSvc = svc }
func (s *Server) SetProjectAppService(svc *service.ProjectAppService)           { s.projectAppSvc = svc }
func (s *Server) SetCompetitionService(svc *service.CompetitionService)         { s.competitionSvc = svc }
func (s *Server) SetServiceListingService(svc *service.ServiceListingService) {
	s.serviceListingSvc = svc
}
func (s *Server) SetEventService(svc *service.EventService)                   { s.eventSvc = svc }
func (s *Server) SetResourceService(svc *service.ResourceService)             { s.resourceSvc = svc }
func (s *Server) SetEmergencyService(svc *service.EmergencyService)           { s.emergencySvc = svc }
func (s *Server) SetMatchingService(svc *service.MatchingService)             { s.matchingSvc = svc }
func (s *Server) SetIntentService(svc *service.IntentService)                 { s.intentSvc = svc }
func (s *Server) SetWorkOrderService(svc *service.WorkOrderService)           { s.workOrderSvc = svc }
func (s *Server) SetPoolService(svc *service.ResourcePoolService)             { s.poolSvc = svc }
func (s *Server) SetTestSiteService(svc *service.TestSiteService)             { s.testSiteSvc = svc }
func (s *Server) SetExhibitionService(svc *service.ExhibitionService)         { s.exhibitionSvc = svc }
func (s *Server) SetTransformationService(svc *service.TransformationService) { s.transSvc = svc }
func (s *Server) SetCollegeService(svc *service.CollegeService)               { s.collegeSvc = svc }
func (s *Server) SetStudyTourRepo(r repository.StudyTourRepository)           { s.studyTourRepo = r }
func (s *Server) SetCooperationService(svc *service.CooperationService)       { s.coopSvc = svc }
func (s *Server) SetRescueCaseService(svc *service.RescueCaseService)         { s.rescueCaseSvc = svc }
func (s *Server) SetEmergencyDeptService(svc *service.EmergencyDeptService)   { s.emergDeptSvc = svc }
func (s *Server) SetAssociationMemberService(svc *service.AssociationMemberService) {
	s.assocMemberSvc = svc
}
func (s *Server) SetApplicationService(svc *service.ApplicationService) { s.appSvc = svc }
func (s *Server) SetContractTemplateService(svc *service.ContractTemplateService) {
	s.contractTplSvc = svc
}

// audit records a write operation to the audit log if a writer is configured.
// P1 修复：带 2s 超时（此前同步无超时，审计慢/池耗尽会拖死写请求），
// 失败落日志（此前静默吞掉，审计链断裂无感知）。
func (s *Server) audit(ctx context.Context, actorID, action, resourceType, resourceID, result string) {
	if s.auditWriter == nil {
		return
	}
	rid := ""
	if v, ok := ctx.Value(requestIDKey{}).(string); ok {
		rid = v
	}
	actx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := s.auditWriter.WriteAudit(actx, repository.AuditEntry{
		ActorID:      actorID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Result:       result,
		RequestID:    rid,
	}); err != nil {
		slog.Warn("audit write failed", "action", action, "resource", resourceType,
			"resource_id", resourceID, "err", err)
	}
}

func (s *Server) Router() http.Handler {
	mux := http.NewServeMux()

	// ── Core routes (organized by module) ───────────────────────────
	s.registerCoreRoutes(mux)

	// ── Extended routes (batch / phase / biz modules) ───────────────
	s.registerPhase3Routes(mux)
	s.registerBatch3Routes(mux)
	s.registerBizRoutes(mux)
	s.registerAdminListRoutes(mux) // batch admin GET list routes
	s.registerBatch1Routes(mux)
	s.registerBatch2Routes(mux)
	s.registerPublicAPIRoutes(mux) // mini-program public routes

	// ── File upload ────────────────────────────────────────────────
	mux.HandleFunc("POST /api/v1/upload", s.handleUpload)

	// ── H5 auth + services config — production (PG/memory backed) ──
	s.registerH5AuthRoutes(mux)

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

	return s.rateLimit(s.requestID(s.recoverPanic(s.securityHeaders(s.withCORS(s.authenticate(s.idempotencyCheck(s.adminGate(middleware.SanitizeBody(mux)))))))))
}

func (s *Server) favicon(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Write([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32"><rect width="32" height="32" rx="6" fill="#2979FF"/><text x="16" y="23" text-anchor="middle" font-size="20">🚁</text></svg>`))
}

func (s *Server) serveUploads(w http.ResponseWriter, r *http.Request) {
	// withCORS 中间件预设了 application/json（供 JSON 响应嗅探），
	// 此处必须清除，否则 FileServer 不会按扩展名推导类型，
	// 加上 nosniff 头后浏览器会拒绝渲染图片。
	w.Header().Del("Content-Type")
	http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))).ServeHTTP(w, r)
}

// servePrivateUploads 鉴权读取 uploads/private/（身份证影像等敏感文件）。
// authenticate 中间件不放行 /uploads/private/ 前缀，此处 actor 必然已注入。
func (s *Server) servePrivateUploads(w http.ResponseWriter, r *http.Request) {
	if _, ok := authenticatedActor(r); !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	w.Header().Del("Content-Type")
	http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))).ServeHTTP(w, r)
}

func (s *Server) index(w http.ResponseWriter, r *http.Request) {
	respond(w, r, http.StatusOK, map[string]any{
		"name":      "Drone Industry Service Platform API",
		"status":    "running",
		"endpoints": []string{"/healthz", "/api/v1/home", "/api/v1/demands", "/api/v1/search"},
	})
}

// resolveBannerImageURL 给 banner 图补全域名/校正协议：
//   - 相对路径 /uploads/xxx：管理后台存的是相对路径，小程序 <image> 直接
//     渲染相对路径会当本地资源 → 白图，需拼完整域名（BASE_URL 优先，
//     其次 X-Forwarded-Proto + Host，适配 nginx https 反代）
//   - http://本站/uploads/xxx 存量数据：管理后台早期版本存过 http 完整 URL，
//     微信小程序强制 https，http 图白屏 → 统一升为 https
func resolveBannerImageURL(r *http.Request, url string) string {
	if strings.HasPrefix(url, "/uploads/") {
		if origin := os.Getenv("BASE_URL"); origin != "" {
			return origin + url
		}
		proto := r.Header.Get("X-Forwarded-Proto")
		if proto == "" {
			proto = "http"
		}
		return proto + "://" + r.Host + url
	}
	if strings.HasPrefix(url, "http://") && strings.Contains(url, r.Host) {
		return "https://" + strings.TrimPrefix(url, "http://")
	}
	return url
}

func (s *Server) home(w http.ResponseWriter, r *http.Request) {
	city := r.URL.Query().Get("city")
	lat, _ := strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
	lng, _ := strconv.ParseFloat(r.URL.Query().Get("lng"), 64)
	data := s.homeSvc.GetHome(city, lat, lng)

	// Banner 图补全域名（拷贝 slice，不污染全局配置的底层数组）
	banners := make([]domain.Banner, len(data.Banners))
	for i, b := range data.Banners {
		b.ImageURL = resolveBannerImageURL(r, b.ImageURL)
		banners[i] = b
	}

	// Stats: demand count, user count — 单源失败不阻断首页，记日志后按零值继续
	demands, err := s.demands.List(repository.DemandFilter{})
	if err != nil {
		slog.Warn("home: list demands", "err", err)
	}
	demandTotal := len(demands)
	users, err := s.userRepo.All()
	if err != nil {
		slog.Warn("home: list users", "err", err)
	}
	userTotal := len(users)

	products, err := s.tradingSvc.ListProducts("")
	if err != nil {
		slog.Warn("home: list products", "err", err)
	}

	// 平台累计浏览量：商品浏览量之和（需求无浏览统计字段，只能按商品口径汇总）
	productViews := 0
	for _, p := range products {
		productViews += p.Views
	}

	respond(w, r, http.StatusOK, map[string]any{
		"city":           data.City,
		"banners":        banners,
		"quick_entries":  data.QuickEntries,
		"latest_demands": data.HotDemands,
		"notices":        data.Notices,
		"shops":          data.Shops, // 商家/企业合一：已审核企业即商家（PRD 企业展示页）
		"products":       products,
		"stats": map[string]int{
			"demands": demandTotal,
			"users":   userTotal,
			"views":   productViews, // 商品累计浏览量之和
		},
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
	// Mine filter: only demands published by current user.
	// 未登录时返回空列表，绝不回退为"全部需求"（防止未登录泄露他人/种子数据）。
	if r.URL.Query().Get("mine") == "1" {
		a, ok := authenticatedActor(r)
		if !ok {
			result = nil
		} else {
			filtered := make([]domain.Demand, 0, len(result))
			for _, d := range result {
				if d.PublisherID == a.ID {
					filtered = append(filtered, d)
				}
			}
			result = filtered
		}
	}
	// C8 修复：不再手工切片——paginatedRespond 内部已按 page/page_size 分页，
	// 双重切片导致 page≥2 恒为空。此处只做公开字段脱敏，分页交给响应层。
	public := make([]domain.Demand, len(result))
	for i, d := range result {
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
	if st == "" {
		st = "all"
	}
	result, err := s.demands.ListAll(repository.DemandFilter{Status: st})
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	paginatedRespond(w, r, result, len(result))
}

// DELETE /api/v1/admin/demands/{id} — 管理端删除需求（仅已取消/已驳回）
func (s *Server) deleteDemand(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok || (a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin) {
		fail(w, r, http.StatusForbidden, errors.New("admin permission required"))
		return
	}
	if err := s.demands.Delete(a, r.PathValue("id")); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	respond(w, r, http.StatusOK, map[string]string{"deleted": "ok"})
}

// GET /api/v1/admin/demands/stats — 需求全量统计（独立于列表分页，
// 管理端统计条不随翻页变化：状态分布基于全量数据）
func (s *Server) adminDemandStats(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok || (a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin) {
		fail(w, r, http.StatusForbidden, errors.New("admin permission required"))
		return
	}
	result, err := s.demands.ListAll(repository.DemandFilter{Status: "all"})
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	stats := map[string]int64{
		"total":   int64(len(result)),
		"pending": 0, "published": 0, "completed": 0,
		"cancelled": 0, "rejected": 0,
	}
	var offlineFen int64
	for _, d := range result {
		switch d.Status {
		case domain.DemandPending:
			stats["pending"]++
		case domain.DemandPublished:
			stats["published"]++
		case domain.DemandCompleted:
			stats["completed"]++
		case domain.DemandCancelled:
			stats["cancelled"]++
		case domain.DemandRejected:
			stats["rejected"]++
		}
		offlineFen += d.OfflineAmountFen
	}
	rate := 0
	if stats["completed"]+stats["published"] > 0 {
		rate = int(stats["completed"] * 100 / (stats["completed"] + stats["published"]))
	}
	respond(w, r, http.StatusOK, map[string]any{
		"total":          stats["total"],
		"pending":        stats["pending"],
		"published":      stats["published"],
		"completed":      stats["completed"],
		"cancelled":      stats["cancelled"],
		"rejected":       stats["rejected"],
		"rate":           rate,
		"offline_amount": float64(offlineFen) / 100,
	})
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
		// 参数/校验错误 → 400；角色/权限错误 → 403（区分错误码，避免掩盖真实原因）
		code := http.StatusForbidden
		if strings.Contains(err.Error(), "required") || strings.Contains(err.Error(), "cannot") {
			code = http.StatusBadRequest
		}
		fail(w, r, code, err)
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

// POST /api/v1/admin/demands/{id}/close — 管理端关闭已公开需求
func (s *Server) closeDemand(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	var in struct {
		Reason string `json:"reason"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	d, err := s.demands.CloseByAdmin(a, r.PathValue("id"), in.Reason)
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
	s.audit(r.Context(), a.ID, "close_demand", "demand", d.ID, "closed")
	respond(w, r, http.StatusOK, d)
}

// POST /api/v1/admin/demands/{id}/amount — 登记线下成交金额（联系对接模式撮合价值）
func (s *Server) setDemandOfflineAmount(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	var in struct {
		OfflineAmountFen int64 `json:"offline_amount_fen"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	d, err := s.demands.SetOfflineAmount(a, r.PathValue("id"), in.OfflineAmountFen)
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
	s.audit(r.Context(), a.ID, "set_offline_amount", "demand", d.ID, "amount")
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
		// Validate the client-supplied key length.
		if len(key) < 8 || len(key) > 128 {
			fail(w, r, http.StatusBadRequest, errors.New("Idempotency-Key must be 8-128 characters"))
			return
		}
		// Namespace the key by the authenticated actor so one user's key
		// cannot replay another user's response.
		if a, ok := authenticatedActor(r); ok {
			key = a.ID + ":" + key
		}
		// Check for previously completed request.
		replay := func(status int, body string) {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(status)
			w.Write([]byte(body))
		}
		if status, body, ok := s.idempotency.get(key); ok {
			replay(status, body)
			return
		}
		// P1 修复：并发同 key 单飞——首个请求执行写操作，其余等待并复用
		// 其响应（此前并发同 key 全部穿透执行，重复创建订单/报名）。
		s.idempotency.mu.Lock()
		if f, ok := s.idempotency.flights[key]; ok {
			s.idempotency.mu.Unlock()
			<-f.done
			replay(f.status, f.body)
			return
		}
		f := &idempotencyFlight{done: make(chan struct{})}
		s.idempotency.flights[key] = f
		s.idempotency.mu.Unlock()

		// 用 defer 保证 panic 时也会关闭 flight，等待者不会永久挂起。
		rec := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		defer func() {
			s.idempotency.mu.Lock()
			f.status, f.body = rec.statusCode, rec.body.String()
			delete(s.idempotency.flights, key)
			s.idempotency.mu.Unlock()
			close(f.done)
		}()
		// Capture the response via a response recorder.
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
		// Rate-limit by the direct peer address. X-Forwarded-For is client
		// controlled and would let attackers rotate it to bypass the limit.
		// (If a trusted proxy is added, parse the last hop it appends instead.)
		key := r.RemoteAddr
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
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
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
	return d.Decode(v)
}

// parseDateInput accepts both RFC3339 timestamps and plain "2006-01-02" dates.
// Returns a zero time (no error) for empty input.
func parseDateInput(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	return time.Parse("2006-01-02", s)
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
	// Slice items for the requested page
	sliced := slicePage(items, page, pageSize)
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]any{
		"data":       sliced,
		"page":       page,
		"page_size":  pageSize,
		"total":      total,
		"request_id": requestIDFromCtx(r),
	}); err != nil {
		slog.Warn("encode paginated response", "error", err)
	}
}

// slicePage returns a page slice from any slice using reflection.
func slicePage(items any, page, pageSize int) any {
	v := reflect.ValueOf(items)
	if v.Kind() != reflect.Slice {
		return items
	}
	total := v.Len()
	start := (page - 1) * pageSize
	if start >= total {
		return reflect.MakeSlice(v.Type(), 0, 0).Interface()
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return v.Slice(start, end).Interface()
}

func requestIDFromCtx(r *http.Request) string {
	if rid, ok := r.Context().Value(requestIDKey{}).(string); ok {
		return rid
	}
	return ""
}

func respond(w http.ResponseWriter, r *http.Request, status int, v any) {
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(map[string]any{"data": v, "request_id": requestIDFromCtx(r)}); err != nil {
		slog.Warn("encode response", "error", err)
	}
}
func fail(w http.ResponseWriter, r *http.Request, status int, err error) {
	if errors.Is(err, io.EOF) {
		err = errors.New("request body is required")
	}
	// P1 修复：5xx 必须落日志——此前所有业务错误仅返回客户端，服务端完全不可见。
	if status >= http.StatusInternalServerError {
		slog.Error("request failed", "method", r.Method, "path", r.URL.Path,
			"status", status, "err", err, "request_id", requestIDFromCtx(r))
	}
	w.WriteHeader(status)
	if e := json.NewEncoder(w).Encode(map[string]any{
		"error":      map[string]string{"code": httpStatusToCode(status), "message": strings.TrimSpace(err.Error())},
		"request_id": requestIDFromCtx(r),
	}); e != nil {
		slog.Warn("encode error response", "error", e)
	}
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
