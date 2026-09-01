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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"reflect"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	_ "drone-platform/docs"
	"drone-platform/internal/cache"
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
	claimSvc          *service.ChallengeClaimService
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
	studyEnrollSvc    *service.StudyTourEnrollmentService
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
	smsIPLimits       sync.Map // ip -> *smsIPLog（短信发送限频，实例级避免测试/多实例互扰）
	smsIPEntries      atomic.Int64
	pwLoginFailures   sync.Map // loginID|clientIP -> *pwFailLog（密码登录失败锁定，双维度键，实例级）
	accountFailCounts sync.Map // loginID -> *accountFailLog（账号级跨 IP 失败累计，实例级——换 IP 无法绕过账号上限）
	regLimits         sync.Map // "phone:"+phone / "ip:"+ip -> *regLimitLog（开放注册限频，实例级）
	regLimitEntries   atomic.Int64
	// adminOpLimits 管理端重型操作限频（广播/导出等）：key=clientIP -> *regLimitLog。
	adminOpLimits sync.Map
	// servicesCfgMu 序列化 services_config.json 的读-改-写（h5SaveServicesConfig 等）：
	// 并发保存此前会互相覆盖字段（readJSON 与 writeJSON 各自加锁，跨调用不原子）。
	servicesCfgMu sync.Mutex
	auditWriter       repository.AuditWriter
	dbPinger          interface{ Ping(context.Context) error }
	storage           string
	// homeCache 首页数据 60s 缓存（实例级：测试各自隔离，生产单实例内共享）。
	homeCache *cache.Cache
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
		func() {
			defer func() { _ = recover() }() // 批3：后台协程裸跑，panic 防护
			s.mu.Lock()
			now := time.Now()
			for k, v := range s.cache {
				if now.After(v.expiresAt) {
					delete(s.cache, k)
				}
			}
			s.mu.Unlock()
		}()
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
	// C 批：缓存容量上限——恶意构造海量唯一 key 时防止内存无限增长。
	// 超限时先清理过期项，仍超限则拒绝缓存新条目（仅本次响应不缓存，不影响功能）。
	if len(s.cache) >= maxIdempotencyEntries {
		now := time.Now()
		for k, v := range s.cache {
			if now.After(v.expiresAt) {
				delete(s.cache, k)
			}
		}
		if len(s.cache) >= maxIdempotencyEntries {
			return
		}
	}
	s.cache[key] = idempotencyEntry{status: status, body: body, expiresAt: time.Now().Add(24 * time.Hour)}
}

// maxIdempotencyEntries 幂等响应缓存条目上限（防内存 DoS）。
const maxIdempotencyEntries = 20000

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

// maxRateBuckets 限流桶上限（防内存 DoS），超限的新来源直接拒绝。
const maxRateBuckets = 100000

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

	// C 批：桶容量上限——海量来源 IP 下 map 持续增长的内存 DoS 防护。
	// 超限时新 IP 不建桶（按限流拒绝），不驱逐既有桶。
	if len(rl.buckets) >= maxRateBuckets {
		if _, ok := rl.buckets[key]; !ok {
			return false
		}
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
	return &Server{demands: d, enterprises: e, enterpriseSvc: es, employment: h, contracts: c, jobSvc: js, communitySvc: cs, listingSvc: ls, labourSvc: lbs, trainingSvc: ts, tradingSvc: trs, insuranceSvc: ins, financeSvc: fin, homeSvc: hs, fileSvc: fs, msgSvc: ms, enrollSvc: ens, expirySvc: exps, tradeSvc: trds, escrowSvc: esc, newsSvc: nws, reviewSvc: rvs, venueSvc: vns, userRepo: ur, refreshRepo: rr, tokens: tokens, rateLimiter: newRateLimiter(100, 200), idempotency: newIdempotencyStore(), homeCache: cache.New(60 * time.Second)}
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
func (s *Server) SetChallengeClaimService(svc *service.ChallengeClaimService)   { s.claimSvc = svc }
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
func (s *Server) SetStudyTourEnrollmentService(svc *service.StudyTourEnrollmentService) { s.studyEnrollSvc = svc }
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

	// ── C 批：性能观测 ─────────────────────────────────────────────
	// pprof 仅 dev 暴露（生产如需再经网关/内网注入）；/metrics 轻量公开。
	mux.HandleFunc("/metrics", s.metricsHandler)
	if adminDevMode() {
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}

	return s.rateLimit(s.accessLog(s.requestID(s.recoverPanic(s.securityHeaders(s.withCORS(s.authenticate(s.idempotencyCheck(s.adminGate(middleware.SanitizeBody(mux))))))))))
}

func (s *Server) favicon(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Write([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32"><rect width="32" height="32" rx="6" fill="#2979FF"/><text x="16" y="23" text-anchor="middle" font-size="20">🚁</text></svg>`))
}

func (s *Server) serveUploads(w http.ResponseWriter, r *http.Request) {
	// withCORS 中间件预设了 application/json（供 JSON 响应嗅探），
	// 此处必须清除，否则 FileServer 不会按扩展名推导类型，
	// 加上 nosniff 头后浏览器会拒绝渲染图片。
	// P2 修复：目录列举 + 大小写变体绕过。
	// ① 指向 uploads/private 子树（大小写不敏感）的请求一律 404——私有文件只经
	//    servePrivateUploads 鉴权读取，公开 FileServer 不服务该子树（/uploads/Private/xx
	//    此前可绕过鉴权）；
	// ② 磁盘上为目录的路径一律 404——GET /uploads/private（无尾斜杠）此前被 FileServer
	//    重定向后列出私有文件 ID。
	if isPrivateUploadsPath(r.URL.Path) || isUploadDirPath(r.URL.Path) {
		http.NotFound(w, r)
		return
	}
	w.Header().Del("Content-Type")
	http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))).ServeHTTP(w, r)
}

// isPrivateUploadsPath 报告路径是否指向 uploads/private 子树（大小写不敏感）。
func isPrivateUploadsPath(p string) bool {
	for _, seg := range strings.Split(strings.TrimPrefix(p, "/"), "/") {
		if strings.EqualFold(seg, "private") {
			return true
		}
	}
	return false
}

// isUploadDirPath 报告路径在磁盘上对应 uploads 下的一个目录（目录列举拒绝）。
// 经 http.Dir 打开，路径穿越（../）由 http.Dir 自身拒绝。
func isUploadDirPath(p string) bool {
	f, err := http.Dir("uploads").Open(strings.TrimPrefix(p, "/uploads/"))
	if err != nil {
		return false
	}
	defer f.Close()
	fi, err := f.Stat()
	return err == nil && fi.IsDir()
}

// servePrivateUploads 鉴权读取 uploads/private/（身份证影像等敏感文件）。
// authenticate 中间件不放行 /uploads/private/ 前缀，此处 actor 必然已注入。
// P1 修复：文件 ID 虽为 128 位随机（不可枚举），仍须校验台账归属——
// 仅上传者本人或平台/协会管理员可读；台账无记录（配额上线前的旧文件）一律 404。
func (s *Server) servePrivateUploads(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/uploads/private/")
	rec, err := s.fileSvc.FindUpload(r.Context(), id)
	if err != nil || rec.Visibility != "private" {
		fail(w, r, http.StatusNotFound, errors.New("file not found"))
		return
	}
	if rec.OwnerID != a.ID &&
		a.Role != domain.RolePlatformAdmin && a.Role != domain.RoleAssociationAdmin {
		fail(w, r, http.StatusForbidden, errors.New("permission denied"))
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

// homeCacheKey 首页缓存键：city + 经纬度（用于距离排序的热门需求，含坐标防串）。
func homeCacheKey(city string, lat, lng float64) string {
	return fmt.Sprintf("home:%s:%.3f:%.3f", city, lat, lng)
}

func (s *Server) home(w http.ResponseWriter, r *http.Request) {
	city := r.URL.Query().Get("city")
	lat, _ := strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
	lng, _ := strconv.ParseFloat(r.URL.Query().Get("lng"), 64)

	// 性能审查：首页数据（banners/快捷入口/公告/热门需求/商家）60s 缓存 +
	// GetOrSet 单飞防击穿；stats 与商品列表不缓存，各自走轻量 COUNT/SUM/LIMIT 查询。
	key := homeCacheKey(city, lat, lng)
	cached, err := s.homeCache.GetOrSet(key, func() (any, error) {
		return s.homeSvc.GetHome(r.Context(), city, lat, lng), nil
	}, 60*time.Second)
	if err != nil {
		slog.Warn("home: cache fill", "err", err)
	}
	data, _ := cached.(service.HomeData)

	// Banner 图补全域名（拷贝 slice，不污染全局配置的底层数组）
	banners := make([]domain.Banner, len(data.Banners))
	for i, b := range data.Banners {
		b.ImageURL = resolveBannerImageURL(r, b.ImageURL)
		banners[i] = b
	}

	// Stats: demand count, user count — 单源失败不阻断首页，记日志后按零值继续。
	// 审查修复：全量拉取只为 len() → COUNT 聚合查询。
	demandTotal, err := s.demands.Count(r.Context(), repository.DemandFilter{Status: string(domain.DemandPublished)})
	if err != nil {
		slog.Warn("home: count demands", "err", err)
	}
	userTotal, err := s.userRepo.Count(r.Context())
	if err != nil {
		slog.Warn("home: count users", "err", err)
	}

	// 商品只取 Top-N（LIMIT 10）不整表；浏览量用 SUM 聚合（不物化行）。
	products, err := s.tradingSvc.ListTopProducts(r.Context(), "", 10)
	if err != nil {
		slog.Warn("home: list products", "err", err)
	}
	productViews, err := s.tradingSvc.SumProductViews(r.Context(), "")
	if err != nil {
		slog.Warn("home: sum product views", "err", err)
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
			"views":   productViews, // 商品累计浏览量之和（SUM 聚合）
		},
	})
}
func (s *Server) search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	ds, err := s.demands.Search(r.Context(), q)
	if err != nil {
		slog.Warn("search: demands failed", "error", err)
		ds = nil
	}
	es, err := s.enterprises.Search(r.Context(), q)
	if err != nil {
		slog.Warn("search: enterprises failed", "error", err)
		es = nil
	}
	public := make([]domain.Demand, len(ds))
	for i, d := range ds {
		public[i] = publicDemand(d)
	}
	// 公开搜索仅返回已审核企业（与首页商家列表/公开企业列表口径一致），
	// 并裁剪敏感字段（联系电话/信用代码/法人等不随搜索结果公开）。
	publicEnts := make([]domain.Enterprise, 0, len(es))
	for _, e := range es {
		if e.Status != "approved" {
			continue
		}
		e.AccountName = crypto.MaskPhone(e.AccountName)
		e.ContactPhone = ""
		e.CreditCode = ""
		e.LegalPerson = ""
		e.Email = ""
		publicEnts = append(publicEnts, e)
	}
	respond(w, r, http.StatusOK, map[string]any{"demands": public, "enterprises": publicEnts})
}
func (s *Server) listDemands(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	mine := r.URL.Query().Get("mine") == "1"
	var result []domain.Demand
	if mine {
		// mine=1：按发布者全量拉取（含 pending/rejected/draft 等全部状态），
		// 此前先经 List（仅 published）再按发布者过滤，导致待审核/已驳回需求在"我的发布"页丢失。
		// 未登录时返回空列表，绝不回退为"全部需求"（防止未登录泄露他人/种子数据）。
		a, ok := authenticatedActor(r)
		if ok {
			var err error
			result, err = s.demands.ListByPublisher(r.Context(), a.ID)
			if err != nil {
				fail(w, r, http.StatusInternalServerError, err)
				return
			}
		}
	} else {
		var err error
		result, err = s.demands.List(r.Context(), repository.DemandFilter{District: r.URL.Query().Get("district"), BizType: r.URL.Query().Get("biz_type"), Sort: r.URL.Query().Get("sort")})
		if err != nil {
			fail(w, r, http.StatusInternalServerError, err)
			return
		}
	}
	// Keyword search filter
	if q != "" {
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
	// C8 修复：不再手工切片——paginatedRespond 内部已按 page/page_size 分页，
	// 双重切片导致 page≥2 恒为空。此处只做公开字段脱敏，分页交给响应层。
	public := make([]domain.Demand, len(result))
	for i, d := range result {
		public[i] = publicDemand(d)
	}
	paginatedRespond(w, r, public, len(result))
}

// POST /api/v1/demands/{id}/favorite — 收藏/取消收藏需求
func (s *Server) toggleDemandFavorite(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	var in struct {
		Favorite bool `json:"favorite"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	if err := s.demands.ToggleFavorite(r.Context(), a.ID, r.PathValue("id"), in.Favorite); err != nil {
		// 区分 404（需求不存在/无权收藏）与 500（DB/加密故障）
		code := http.StatusNotFound
		if strings.Contains(err.Error(), "not found") {
			code = http.StatusNotFound
		} else if strings.Contains(err.Error(), "only published") {
			code = http.StatusBadRequest
		} else {
			slog.Warn("toggle demand favorite", "error", err)
			code = http.StatusInternalServerError
		}
		fail(w, r, code, err)
		return
	}
	respond(w, r, http.StatusOK, map[string]bool{"favorite": in.Favorite})
}

// GET /api/v1/demands/favorites/mine — 我的收藏需求列表（按收藏时间倒序，公开视图脱敏）
func (s *Server) listMyDemandFavorites(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		respond(w, r, http.StatusOK, []domain.Demand{})
		return
	}
	demands, err := s.demands.ListFavoriteDemands(r.Context(), a.ID)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	public := make([]domain.Demand, len(demands))
	for i, d := range demands {
		public[i] = publicDemand(d)
	}
	respond(w, r, http.StatusOK, public)
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
	result, err := s.demands.ListAll(r.Context(), repository.DemandFilter{Status: st})
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
	if err := s.demands.Delete(r.Context(), a, r.PathValue("id")); err != nil {
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
	result, err := s.demands.ListAll(r.Context(), repository.DemandFilter{Status: "all"})
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
	d, err := s.demands.Create(r.Context(), a, in)
	if err != nil {
		// 参数/校验错误 → 400；角色/权限错误 → 403；数据库等系统错误 → 500
		// （此前一律 403 会把 DB 故障伪装成权限拒绝）
		code := http.StatusForbidden
		if errors.Is(err, service.ErrRoleNotAllowed) || errors.Is(err, service.ErrNotOwner) || strings.Contains(err.Error(), "permission") {
			code = http.StatusForbidden
		} else if strings.Contains(err.Error(), "required") || strings.Contains(err.Error(), "cannot") || strings.Contains(err.Error(), "only") {
			code = http.StatusBadRequest
		} else {
			slog.Warn("create demand failed", "error", err)
			code = http.StatusInternalServerError
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
	d, err := s.demands.Approve(r.Context(), a, r.PathValue("id"))
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
	d, err := s.demands.CloseByAdmin(r.Context(), a, r.PathValue("id"), in.Reason)
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
	d, err := s.demands.SetOfflineAmount(r.Context(), a, r.PathValue("id"), in.OfflineAmountFen)
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
	// 性能审查：分页下沉 SQL（repo COUNT+LIMIT/OFFSET），响应层不再二次切片。
	page, pageSize := paginationFromQuery(r)
	out, total, err := s.employment.List(r.Context(), a, (page-1)*pageSize, pageSize)
	if err != nil {
		fail(w, r, http.StatusForbidden, err)
		return
	}
	respondPage(w, r, out, total, page, pageSize)
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
	out, err := s.employment.Create(r.Context(), a, v)
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
	// 性能审查：分页下沉 SQL（repo COUNT+LIMIT/OFFSET），响应层不再二次切片。
	page, pageSize := paginationFromQuery(r)
	out, total, err := s.contracts.List(r.Context(), a, (page-1)*pageSize, pageSize)
	if err != nil {
		fail(w, r, http.StatusForbidden, err)
		return
	}
	respondPage(w, r, out, total, page, pageSize)
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
	out, err := s.contracts.Create(r.Context(), a, v)
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
	out, err := s.enterprises.Pending(r.Context(), a)
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
		// cannot replay another user's response. 未认证请求同样加固定前缀，
		// 避免匿名写路径（如登录/注册）的响应被跨请求回放（B 批加固）。
		if a, ok := authenticatedActor(r); ok {
			key = a.ID + ":" + key
		} else {
			key = "anon:" + key
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
		// statusCode 初始为 0（未写响应头）：此前默认 200，handler panic 时 defer 会把
		// 一个不存在的 200 空响应缓存为"成功"，同 key 重放得到假成功（评审 P2）。
		rec := &responseRecorder{ResponseWriter: w, statusCode: 0}
		defer func() {
			s.idempotency.mu.Lock()
			if rec.statusCode == 0 {
				// panic 且未写响应头：以 500 记账，等待者不误以为成功。
				rec.statusCode = http.StatusInternalServerError
			}
			f.status, f.body = rec.statusCode, rec.body.String()
			delete(s.idempotency.flights, key)
			s.idempotency.mu.Unlock()
			close(f.done)
		}()
		// Capture the response via a response recorder.
		next.ServeHTTP(rec, r)
		// Store the result for future idempotent requests.
		// P2 修复：仅缓存 2xx（200-299）与 4xx 校验类错误（400/409/422 等确定性结果）；
		// 5xx 不缓存——服务端错误重试须重新执行写操作，而非回放错误响应。
		if rec.statusCode >= http.StatusOK && rec.statusCode < http.StatusInternalServerError {
			s.idempotency.set(key, rec.statusCode, rec.body.String())
		}
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

// appStartTime 记录进程启动时刻，供 /metrics 计算运行时长。
var appStartTime = time.Now()

// metricsHandler 输出轻量运行时指标（内存/协程/GC/运行时长），JSON 格式。
func (s *Server) metricsHandler(w http.ResponseWriter, r *http.Request) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"goroutines":       runtime.NumGoroutine(),
		"heap_alloc_bytes": m.HeapAlloc,
		"heap_sys_bytes":   m.HeapSys,
		"heap_objects":     m.HeapObjects,
		"gc_cycles":        m.NumGC,
		"uptime_seconds":   int64(time.Since(appStartTime).Seconds()),
	})
}

// statusWriter 透传写操作并记录状态码，供访问日志中间件使用。
type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

// accessLog 记录每个请求的方法/路径/状态码/耗时/来源 IP/request_id。
// 批3 P1：此前 RequestLog 定义但零调用，生产完全无访问日志；挂在最外层
// （rateLimit 之后），panic 恢复写入的状态也能被记录。
func (s *Server) accessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)
		if sw.status == 0 {
			sw.status = http.StatusOK
		}
		slog.Info("http", "method", r.Method, "path", r.URL.Path,
			"status", sw.status, "dur_ms", time.Since(start).Milliseconds(),
			"ip", clientIP(r), "request_id", requestIDFromCtx(r))
	})
}

// clientIP 提取真实客户端 IP。
// nginx 反代场景下 RemoteAddr 恒为 127.0.0.1/网桥（所有用户共享一个限流桶），
// 须取 X-Forwarded-For：nginx 按 $proxy_add_x_forwarded_for 把远端地址追加在
// 末尾，取最后一项即最接近真实客户端的一跳（nginx 标准做法）。
// P2 修复：仅当直连方是受信代理（回环/私网/链路本地）时才采信 XFF——
// 此前直连公网部署时也信任 XFF，攻击者伪造 X-Forwarded-For 即可绕过
// 全局限流/短信 5/min/注册 3/10min 限频；直连场景一律用 RemoteAddr。
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	if isTrustedProxyIP(host) {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			// 取最后一项（代理追加的远端在末尾，最接近真实客户端）。
			if i := strings.LastIndexByte(xff, ','); i >= 0 {
				xff = xff[i+1:]
			}
			if ip := strings.TrimSpace(xff); ip != "" && net.ParseIP(ip) != nil {
				return ip
			}
		}
	}
	return host
}

// isTrustedProxyIP 报告直连 IP 是否来自受信代理（回环 / RFC1918 私网 / 链路本地）。
func isTrustedProxyIP(ip string) bool {
	p := net.ParseIP(ip)
	if p == nil {
		return false
	}
	return p.IsLoopback() || p.IsPrivate() || p.IsLinkLocalUnicast()
}

func (s *Server) rateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 按真实客户端 IP 限流（nginx 反代后取 X-Forwarded-For 首跳；
		// API 仅回环监听，不存在绕过 nginx 伪造 XFF 的路径）。
		key := clientIP(r)
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
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID, Idempotency-Key")
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
		// 开发默认（未配置 CORS_ORIGINS 时）：前端 Vite dev server(5173) + 历史 3000。
		return []string{
			"http://localhost:3000",
			"http://localhost:5173",
			"http://127.0.0.1:5173",
		}
	}
	return strings.Split(raw, ",")
}

// maxBodyBytes JSON 请求体上限（与 SanitizeBody 一致）。
const maxBodyBytes = 1 << 20

func decode(r *http.Request, v any) error {
	// C 批：超限不再静默截断——1MiB+1 探测，超限返回明确错误（由 handler 映射 413）。
	body, err := io.ReadAll(io.LimitReader(r.Body, maxBodyBytes+1))
	if err != nil {
		return fmt.Errorf("read request body: %w", err)
	}
	if len(body) > maxBodyBytes {
		return errors.New("request body too large (max 1MiB)")
	}
	// 空 body 保持 io.EOF 语义：handler 常用 errors.Is(err, io.EOF) 区分"无 body 但可选"，
	// fail() 也将其映射为统一文案。
	if len(bytes.TrimSpace(body)) == 0 {
		return io.EOF
	}
	// XSS 兜底：无论 Content-Type 声明是什么，只要是可解析的 JSON 一律过白名单消毒
	// （攻击者可伪造 text/plain 等绕过 SanitizeBody 中间件，此处为最后防线）。
	clean, serr := middleware.SanitizeJSONBody(body)
	if serr != nil {
		return serr
	}
	if clean != nil {
		body = clean
	}
	return json.Unmarshal(body, v)
}

// parseDateInput accepts RFC3339 timestamps, "2006-01-02 15:04" (管理端活动/调度
// 表单的 placeholder 格式) and plain "2006-01-02" dates.
// Returns a zero time (no error) for empty input.
func parseDateInput(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	if t, err := time.Parse("2006-01-02 15:04", s); err == nil {
		return t, nil
	}
	return time.Parse("2006-01-02", s)
}

// normalizeCreateStatus 校验创建请求透传的状态值：仅白名单内值透传，
// 其余（含空）返回 ""，由 service 层按各自默认状态兜底——防止前端传非法
// 状态值直接落库（创建入口只透传合法值，非法值用默认）。
func normalizeCreateStatus(s string, allowed ...string) string {
	for _, a := range allowed {
		if s == a {
			return s
		}
	}
	return ""
}

// strictDate 严格解析用户提交的日期字段（P2 修复）：非法日期（非 RFC3339 /
// "2006-01-02 15:04:05" / "2006-01-02" 格式）直接 400，防止此前 ParseTime 把
// 非法日期静默写成当前时间落库；空串视为"未设置"，返回零值时间不报错。
func strictDate(w http.ResponseWriter, r *http.Request, v string) (time.Time, bool) {
	t, err := domain.ParseTimeStrict(v)
	if err != nil {
		fail(w, r, http.StatusBadRequest, errors.New("invalid date format"))
		return time.Time{}, false
	}
	return t, true
}
func paginationFromQuery(r *http.Request) (page, pageSize int) {
	page, pageSize = 1, 20
	if p, err := strconv.Atoi(r.URL.Query().Get("page")); err == nil && p > 0 {
		page = p
	}
	if ps, err := strconv.Atoi(r.URL.Query().Get("page_size")); err == nil && ps > 0 {
		pageSize = ps
	}
	// C 批：page 上界——防超大 OFFSET 深分页慢查询。
	if page > 10000 {
		page = 10000
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

// respondPage 输出「已由 service/repo 按真实 offset/limit 分页好」的数据：
// 与 paginatedRespond 同信封（data/page/page_size/total/request_id），但不再
// 二次切片——调用方负责把分页下沉到 SQL（COUNT+LIMIT/OFFSET），total 为
// 过滤后的总条数（COUNT 或 len(filtered)）。
func respondPage(w http.ResponseWriter, r *http.Request, items any, total, page, pageSize int) {
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]any{
		"data":       items,
		"page":       page,
		"page_size":  pageSize,
		"total":      total,
		"request_id": requestIDFromCtx(r),
	}); err != nil {
		slog.Warn("encode page response", "error", err)
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
	// B 批加固：生产环境 5xx 不再向客户端回显内部错误细节（可能含 SQL/实现信息），
	// 统一文案，原始错误只留在服务端日志；4xx 保留业务提示语。
	// 生产判定与 config 一致：ENV=production 或 DATABASE_URL 已设（此前仅认 ENV 字符串，
	// go run/二进制直连 PG 部署漏设 ENV 时 500 会裸奔内部细节）。
	message := strings.TrimSpace(err.Error())
	if status >= http.StatusInternalServerError && (os.Getenv("ENV") == "production" || os.Getenv("DATABASE_URL") != "") {
		message = "internal server error"
	}
	w.WriteHeader(status)
	if e := json.NewEncoder(w).Encode(map[string]any{
		"error":      map[string]string{"code": httpStatusToCode(status), "message": message},
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
