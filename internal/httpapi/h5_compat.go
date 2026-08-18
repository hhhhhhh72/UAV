package httpapi

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"

	"drone-platform/internal/config"
	"drone-platform/internal/domain"
	"drone-platform/internal/service"
)

// ============================================================================
// H5 Compatibility Layer
//
// The Vue 3 H5 frontend was built for the Node.js backend and expects specific
// API paths and response formats. This file bridges those expectations to the
// Go backend without modifying the frontend code.
// ============================================================================

// ─── JSON File Storage ──────────────────────────────────────────────────────

var (
	_appsMu       sync.RWMutex
	_appsFile     = "applications.json"
	_casesMu      sync.RWMutex
	_casesFile    = "cases.json"
	_catsMu       sync.RWMutex
	_catsFile     = "case_categories.json"
	_reviewsMu    sync.RWMutex
	_reviewsFile  = "reviews.json"
	_usersMu      sync.RWMutex
	_usersFile    = "users.json"
	_servicesMu   sync.RWMutex
	_servicesFile = "services_config.json"
)

func ensureFile(path string, fallback []byte) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.WriteFile(path, fallback, 0644)
	}
}

func readJSON(path string, mu *sync.RWMutex, target any) error {
	mu.RLock()
	defer mu.RUnlock()
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	return json.Unmarshal(data, target)
}

func writeJSON(path string, mu *sync.RWMutex, data any) error {
	mu.Lock()
	defer mu.Unlock()
	raw, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal for %s: %w", path, err)
	}
	// 原子写：临时文件 + rename，防崩溃产生半写文件
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, raw, 0644); err != nil {
		return fmt.Errorf("write %s tmp: %w", path, err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("rename %s: %w", path, err)
	}
	return nil
}

func init() {
	ensureFile(_appsFile, []byte("[]"))
	ensureFile(_casesFile, []byte("[]"))
	ensureFile(_catsFile, []byte(`[{"id":1,"name":"无人机物流","service":"无人机物流服务"},{"id":4,"name":"无人机吊运","service":"无人机吊运服务"},{"id":5,"name":"无人机表演","service":"航空表演"}]`))
	ensureFile(_reviewsFile, []byte("[]"))
	ensureFile(_usersFile, []byte("[]"))
	ssCfg, _ := os.ReadFile(_servicesFile)
	ensureFile(_servicesFile, ssCfg)
}

// ─── Services Config (the endpoint that was causing 500 errors) ─────────────

func (s *Server) h5GetServicesConfig(w http.ResponseWriter, r *http.Request) {
	var cfg map[string]any
	if err := readJSON(_servicesFile, &_servicesMu, &cfg); err != nil {
		respond(w, r, http.StatusOK, map[string]any{})
		return
	}
	// Inject baseUrl + rewrite relative image paths to full URLs for miniprogram
	origin := "http://" + r.Host
	if os.Getenv("BASE_URL") != "" {
		origin = os.Getenv("BASE_URL")
	}
	if home, ok := cfg["_home"].(map[string]any); ok {
		home["baseUrl"] = origin
		// Rewrite banner images to full URLs
		if banners, ok := home["banners"].([]any); ok {
			for i, b := range banners {
				if bm, ok := b.(map[string]any); ok {
					if img, ok := bm["image"].(string); ok && strings.HasPrefix(img, "/uploads/") {
						bm["image"] = origin + img
					}
					banners[i] = bm
				}
			}
			home["banners"] = banners
		}
		// Rewrite headerImage
		if hi, ok := home["headerImage"].(string); ok && strings.HasPrefix(hi, "/uploads/") {
			home["headerImage"] = origin + hi
		}
		cfg["_home"] = home
	}
	respond(w, r, http.StatusOK, cfg)
}

func (s *Server) h5SaveServicesConfig(w http.ResponseWriter, r *http.Request) {
	// 覆盖全局配置的写接口（含同步平台 banner/公告）：仅平台管理员可保存
	a, ok := authenticatedActor(r)
	if !ok || a.Role != domain.RolePlatformAdmin {
		fail(w, r, http.StatusForbidden, errors.New("platform admin permission required"))
		return
	}
	var body struct {
		Config map[string]any `json:"config"`
	}
	if err := decode(r, &body); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	if body.Config == nil {
		fail(w, r, http.StatusBadRequest, errBadRequest("config is required"))
		return
	}
	if err := writeJSON(_servicesFile, &_servicesMu, body.Config); err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	// 管理后台「首页配置」→ 平台全局配置：小程序 /api/v1/home 读 platform_config.json，
	// 后台的 _home 存 services_config.json，不打通则轮播 Banner/公告永远不生效
	if err := syncHomeConfigToPlatform(body.Config); err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, map[string]string{"status": "saved"})
}

// buildPlatformBanners 把管理后台 _home.banners（{image, link} 字段）映射为
// PlatformConfig.Banners（image_url/link_url），无图项跳过。
func buildPlatformBanners(raw []any) []domain.Banner {
	var out []domain.Banner
	for _, item := range raw {
		bm, ok := item.(map[string]any)
		if !ok {
			continue
		}
		img, _ := bm["image"].(string)
		if strings.TrimSpace(img) == "" {
			continue
		}
		link, _ := bm["link"].(string)
		out = append(out, domain.Banner{
			ID:        fmt.Sprintf("banner-%d", len(out)+1),
			ImageURL:  strings.TrimSpace(img),
			LinkURL:   strings.TrimSpace(link),
			SortOrder: len(out) + 1,
			Status:    "active",
		})
	}
	return out
}

// syncHomeConfigToPlatform 把管理后台「首页配置」（services_config.json 的 _home）
// 同步到平台全局配置（platform_config.json）。只覆盖 banners/notices，
// 其余字段（费率等）保留当前值。
func syncHomeConfigToPlatform(cfg map[string]any) error {
	home, ok := cfg["_home"].(map[string]any)
	if !ok {
		return nil
	}
	platform := config.GetPlatformConfig()
	if raw, ok := home["banners"].([]any); ok {
		platform.Banners = buildPlatformBanners(raw)
	}
	if raw, ok := home["notices"].([]any); ok {
		var notices []string
		for _, item := range raw {
			if s, ok := item.(string); ok && strings.TrimSpace(s) != "" {
				notices = append(notices, strings.TrimSpace(s))
			}
		}
		platform.Notices = notices
	}
	return config.SavePlatformConfig(platform)
}

// ─── File Upload ────────────────────────────────────────────────────────────

func (s *Server) h5Upload(w http.ResponseWriter, r *http.Request) {
	s.uploadFile(w, r)
}

// ─── Applications (JSON file backed) ────────────────────────────────────────

type h5Application struct {
	ID                  string `json:"id"`
	OrderNo             string `json:"orderNo,omitempty"`
	ServiceID           string `json:"serviceId"`
	ServiceName         string `json:"serviceName"`
	Status              string `json:"status"`
	UserID              string `json:"userId"`
	ContactName         string `json:"contactName,omitempty"`
	ContactPhone        string `json:"contactPhone,omitempty"`
	TraineeName         string `json:"traineeName,omitempty"`
	TraineePhone        string `json:"traineePhone,omitempty"`
	Name                string `json:"name,omitempty"`
	Phone               string `json:"phone,omitempty"`
	CompanyName         string `json:"companyName,omitempty"`
	CompetitionRole     string `json:"competitionRole,omitempty"`
	CompetitionRoleText string `json:"competitionRoleText,omitempty"`
	CompetitionGroup    string `json:"competitionGroup,omitempty"`
	CompetitionProject  string `json:"competitionProject,omitempty"`
	Gender              string `json:"gender,omitempty"`
	IDCard              string `json:"idCard,omitempty"`
	RegNo               string `json:"regNo,omitempty"`
	Level               string `json:"level,omitempty"`
	ValidDate           string `json:"validDate,omitempty"`
	Manager             string `json:"manager,omitempty"`
	ManagerPhone        string `json:"managerPhone,omitempty"`
	ContactPerson       string `json:"contactPerson,omitempty"`
	CustomerType        string `json:"customerType,omitempty"`
	CargoType           string `json:"cargoType,omitempty"`
	StartAddress        string `json:"startAddress,omitempty"`
	EndAddress          string `json:"endAddress,omitempty"`
	TraineeGender       string `json:"traineeGender,omitempty"`
	TraineeIDCard       string `json:"traineeIdCard,omitempty"`
	LicenseLevel        string `json:"licenseLevel,omitempty"`
	ExamModel           string `json:"examModel,omitempty"`
	Remark              string `json:"remark,omitempty"`
	Email               string `json:"email,omitempty"`
	Location            string `json:"location,omitempty"`
	CreateTime          string `json:"createTime"`
	ApplyTime           string `json:"applyTime,omitempty"`
}

func (s *Server) h5ListApplications(w http.ResponseWriter, r *http.Request) {
	var apps []h5Application
	if err := readJSON(_appsFile, &_appsMu, &apps); err != nil {
		apps = []h5Application{}
	}
	userID := r.URL.Query().Get("userId")
	role := r.URL.Query().Get("role")
	// Filter: regular users see their own, admin sees all or by service
	if userID != "" && role != "admin" {
		filtered := make([]h5Application, 0)
		for _, a := range apps {
			if a.UserID == userID {
				filtered = append(filtered, a)
			}
		}
		apps = filtered
	}
	// Sort newest first
	sort.Slice(apps, func(i, j int) bool {
		return apps[i].CreateTime > apps[j].CreateTime
	})
	respond(w, r, http.StatusOK, apps)
}

func (s *Server) h5SubmitApplication(w http.ResponseWriter, r *http.Request) {
	if s.appSvc == nil {
		fail(w, r, http.StatusInternalServerError, errors.New("application service unavailable"))
		return
	}
	// 全量表单入库（form_data JSONB），关键字段抽列便于管理端查询
	var raw map[string]any
	if err := decode(r, &raw); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	// P1 修复：归属一律取已认证 actor，禁止客户端指定 userId 冒用他人身份。
	if a, ok := authenticatedActor(r); ok {
		raw["userId"] = a.ID
	}
	now := time.Now()
	app := domain.Application{
		ID:          now.Format("20060102150405") + randomSuffix(4),
		UserID:      strFromMap(raw, "userId"),
		ServiceID:   strFromMap(raw, "serviceId"),
		ServiceName: strFromMap(raw, "serviceName"),
		OrderNo:     strFromMap(raw, "orderNo"),
		Status:      "待处理",
		ApplyTime:   now.Format("2006-01-02 15:04:05"),
		FormData:    raw,
	}
	if v := strFromMap(raw, "status"); v != "" {
		app.Status = v
	}
	if v := strFromMap(raw, "applyTime"); v != "" {
		app.ApplyTime = v
	}
	if _, err := s.appSvc.Create(r.Context(), app); err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, map[string]any{"success": true, "id": app.ID})
}

// strFromMap returns the string value at key, or "" if missing / not a string.
func strFromMap(m map[string]any, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func (s *Server) h5UpdateApplication(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	if err := decode(r, &body); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	var apps []h5Application
	if err := readJSON(_appsFile, &_appsMu, &apps); err != nil {
		fail(w, r, http.StatusNotFound, err)
		return
	}
	found := false
	for i := range apps {
		if apps[i].ID == body.ID || apps[i].OrderNo == body.ID {
			apps[i].Status = body.Status
			found = true
			break
		}
	}
	if !found {
		fail(w, r, http.StatusNotFound, errBadRequest("application not found"))
		return
	}
	writeJSON(_appsFile, &_appsMu, apps)
	respond(w, r, http.StatusOK, map[string]string{"status": "updated"})
}

func (s *Server) h5ExportApplications(w http.ResponseWriter, r *http.Request) {
	var apps []h5Application
	readJSON(_appsFile, &_appsMu, &apps)
	// Simple JSON export (frontend handles XLSX conversion if needed)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=export.json")
	json.NewEncoder(w).Encode(apps)
}

// ─── Dashboard Stats ────────────────────────────────────────────────────────

func (s *Server) h5AdminStats(w http.ResponseWriter, r *http.Request) {
	var apps []h5Application
	readJSON(_appsFile, &_appsMu, &apps)
	var users []map[string]any
	readJSON(_usersFile, &_usersMu, &users)

	totalOrders := len(apps)
	pendingOrders := 0
	processingOrders := 0
	completedOrders := 0
	for _, a := range apps {
		switch a.Status {
		case "待处理":
			pendingOrders++
		case "处理中":
			processingOrders++
		case "已完成":
			completedOrders++
		}
	}

	var comps []h5Application
	for _, a := range apps {
		if a.ServiceID == "13" {
			comps = append(comps, a)
		}
	}

	respond(w, r, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"overview": map[string]any{
				"totalOrders":      totalOrders,
				"pendingOrders":    pendingOrders,
				"processingOrders": processingOrders,
				"completedOrders":  completedOrders,
				"totalUsers":       len(users),
				"totalCases":       0,
				"totalCompetition": len(comps),
			},
			"orderTrend":        []any{},
			"competitionByRole": map[string]any{"athlete": 0, "coach": 0, "referee": 0, "club": 0},
			"userGrowth":        []any{},
			"statusDist":        map[string]any{"待处理": pendingOrders, "处理中": processingOrders, "已完成": completedOrders},
		},
	})
}

// ─── Admin Applications ─────────────────────────────────────────────────────

func (s *Server) h5AdminApplications(w http.ResponseWriter, r *http.Request) {
	var apps []h5Application
	readJSON(_appsFile, &_appsMu, &apps)
	respond(w, r, http.StatusOK, map[string]any{"success": true, "data": apps})
}

func (s *Server) h5AdminApplicationByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var apps []h5Application
	readJSON(_appsFile, &_appsMu, &apps)
	for _, a := range apps {
		if a.ID == id {
			respond(w, r, http.StatusOK, map[string]any{"success": true, "data": a})
			return
		}
	}
	fail(w, r, http.StatusNotFound, errBadRequest("not found"))
}

func (s *Server) h5AdminUpdateApplication(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body map[string]any
	if err := decode(r, &body); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	var apps []h5Application
	readJSON(_appsFile, &_appsMu, &apps)
	for i := range apps {
		if apps[i].ID == id {
			if s, ok := body["status"].(string); ok {
				apps[i].Status = s
			}
			writeJSON(_appsFile, &_appsMu, apps)
			respond(w, r, http.StatusOK, map[string]any{"success": true})
			return
		}
	}
	fail(w, r, http.StatusNotFound, errBadRequest("not found"))
}

// ─── Users ──────────────────────────────────────────────────────────────────

func (s *Server) h5Users(w http.ResponseWriter, r *http.Request) {
	var users []map[string]any
	readJSON(_usersFile, &_usersMu, &users)
	respond(w, r, http.StatusOK, users)
}

func (s *Server) h5UpdateUserRole(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ID   string `json:"id"`
		Role string `json:"role"`
	}
	if err := decode(r, &body); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	var users []map[string]any
	readJSON(_usersFile, &_usersMu, &users)
	for i, u := range users {
		if u["id"] == body.ID {
			users[i]["role"] = body.Role
			writeJSON(_usersFile, &_usersMu, users)
			respond(w, r, http.StatusOK, map[string]string{"status": "updated"})
			return
		}
	}
	fail(w, r, http.StatusNotFound, errBadRequest("user not found"))
}

func (s *Server) h5UpdateUserProfile(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Avatar string `json:"avatar"`
		Phone  string `json:"phone"`
	}
	if err := decode(r, &body); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	var users []map[string]any
	readJSON(_usersFile, &_usersMu, &users)
	for i, u := range users {
		if u["id"] == body.ID {
			if body.Name != "" {
				users[i]["name"] = body.Name
			}
			if body.Avatar != "" {
				users[i]["avatar"] = body.Avatar
			}
			if body.Phone != "" {
				users[i]["phone"] = body.Phone
			}
			writeJSON(_usersFile, &_usersMu, users)
			resp := map[string]any{"success": true, "user": users[i]}
			respond(w, r, http.StatusOK, resp)
			return
		}
	}
	fail(w, r, http.StatusNotFound, errBadRequest("user not found"))
}

// ─── Auth Compatibility ─────────────────────────────────────────────────────

// ── 密码登录失败锁定 ──
// bcrypt 慢哈希 + IP 限频挡不住多 IP 分布式爆破：连续失败达上限后
// 锁定该账号 15 分钟。表挂 Server 实例（测试互不干扰）。

const (
	pwMaxFailures  = 10
	pwLockDuration = 15 * time.Minute
)

type pwFailLog struct {
	mu          sync.Mutex
	count       int
	lockedUntil time.Time
}

// dummyPasswordHash 用于不存在的账号也执行一次 bcrypt 校验，
// 抹平"账号不存在"与"密码错误"的响应时间差（防用户枚举时序侧信道）。
var dummyPasswordHash, _ = bcrypt.GenerateFromPassword([]byte("dummy-password-for-timing"), bcrypt.DefaultCost)

// passwordLoginLocked 报告账号是否处于失败锁定中。
func (s *Server) passwordLoginLocked(loginID string) bool {
	v, ok := s.pwLoginFailures.Load(loginID)
	if !ok {
		return false
	}
	log := v.(*pwFailLog)
	log.mu.Lock()
	defer log.mu.Unlock()
	return time.Now().Before(log.lockedUntil)
}

// recordPasswordFailure 记录一次失败；达到上限即锁定。
func (s *Server) recordPasswordFailure(loginID string) {
	v, _ := s.pwLoginFailures.LoadOrStore(loginID, &pwFailLog{})
	log := v.(*pwFailLog)
	log.mu.Lock()
	defer log.mu.Unlock()
	log.count++
	if log.count >= pwMaxFailures {
		log.lockedUntil = time.Now().Add(pwLockDuration)
	}
}

// clearPasswordFailures 登录成功后清零失败记录。
func (s *Server) clearPasswordFailures(loginID string) {
	s.pwLoginFailures.Delete(loginID)
}

func (s *Server) h5AuthLogin(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Phone    string `json:"phone"`
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := decode(r, &body); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	loginID := body.Phone
	if loginID == "" {
		loginID = body.Username
	}
	if loginID == "" || body.Password == "" {
		fail(w, r, http.StatusBadRequest, errBadRequest("phone/username and password required"))
		return
	}
	// 账号失败锁定：连续失败达上限后拒绝校验（防分布式爆破）。
	if s.passwordLoginLocked(loginID) {
		fail(w, r, http.StatusTooManyRequests, errBadRequest("尝试次数过多，请稍后再试"))
		return
	}

	// Resolve the credential record. The bcrypt hash lives in the database
	// (users.password_hash); users.json is a legacy fallback for accounts that
	// predate that column. Accounts without a stored hash cannot use password
	// login (WeChat-only users) — the login is rejected, never accepted.
	//
	// Two ID conventions are accepted: "user-"+loginID (app-registered
	// accounts) and the raw loginID (accounts created via the admin panel).
	uid := "user-" + loginID
	passwordHash := ""
	var user map[string]any
	for _, candidate := range []string{uid, loginID} {
		u, err := s.userRepo.FindByID(r.Context(), candidate)
		if err != nil {
			continue
		}
		passwordHash = u.PasswordHash
		if u.Role != "" {
			user = map[string]any{"id": u.ID, "phone": loginID, "role": string(u.Role), "status": u.Status}
		}
		break
	}
	if passwordHash == "" {
		// Legacy fallback: look up users.json (accounts registered before
		// password_hash was persisted to the database).
		var users []map[string]any
		readJSON(_usersFile, &_usersMu, &users)
		for _, ju := range users {
			if ju["phone"] == loginID || ju["username"] == loginID {
				if h, _ := ju["passwordHash"].(string); h != "" {
					passwordHash = h
				}
				user = ju
				break
			}
		}
	}
	if passwordHash == "" || bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(body.Password)) != nil {
		// 账号不存在也执行一次 dummy bcrypt，抹平时序差异（防用户枚举）。
		if passwordHash == "" {
			_ = bcrypt.CompareHashAndPassword(dummyPasswordHash, []byte(body.Password))
		}
		s.recordPasswordFailure(loginID)
		fail(w, r, http.StatusUnauthorized, errBadRequest("账号或密码错误"))
		return
	}
	s.clearPasswordFailures(loginID)

	// Issue Go backend tokens
	id, _ := user["id"].(string)
	if id == "" {
		id = uid
		user["id"] = id
	}
	role, _ := user["role"].(string)
	if role == "" {
		role = "individual"
	}

	accessToken, err := s.tokens.IssueJWT(actorFromMap(id, role), 15*time.Minute)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, fmt.Errorf("issue access token: %w", err))
		return
	}
	refreshToken, err := service.GenerateRefreshToken()
	if err != nil {
		fail(w, r, http.StatusInternalServerError, fmt.Errorf("generate refresh token: %w", err))
		return
	}
	tokenHash := service.HashToken(refreshToken)
	s.refreshRepo.Store(r.Context(), id, tokenHash, time.Now().Add(7*24*time.Hour))

	safeUser := map[string]any{}
	for k, v := range user {
		if k != "password" && k != "passwordHash" {
			safeUser[k] = v
		}
	}
	safeUser["id"] = id
	safeUser["role"] = role

	respond(w, r, http.StatusOK, map[string]any{
		"success":      true,
		"user":         safeUser,
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
	})
}

func (s *Server) h5AuthRegister(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Phone    string `json:"phone"`
		Password string `json:"password"`
		Name     string `json:"name"`
	}
	if err := decode(r, &body); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	if body.Phone == "" || body.Password == "" {
		fail(w, r, http.StatusBadRequest, errBadRequest("phone and password required"))
		return
	}

	uid := "user-" + body.Phone

	// Check if user already exists in PG
	if _, err := s.userRepo.FindByID(r.Context(), uid); err == nil {
		fail(w, r, http.StatusConflict, errBadRequest("user already exists"))
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}

	name := body.Name
	if name == "" {
		name = "User" + body.Phone[len(body.Phone)-4:]
	}

	// Save to PG users table — the bcrypt hash is persisted here so password
	// login works from the database (no reliance on the JSON compat file).
	now := time.Now()
	user := domain.User{
		ID:           uid,
		WechatOpenID: "phone:" + body.Phone, // non-WeChat users get unique openid to avoid UNIQUE violation
		PasswordHash: string(hashedPassword),
		Role:         domain.RoleIndividual,
		Status:       "active",
		Version:      1,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if _, err := s.userRepo.Create(r.Context(), user); err != nil {
		fail(w, r, http.StatusInternalServerError, fmt.Errorf("create user: %w", err))
		return
	}

	// Also keep in users.json for legacy compatibility
	var users []map[string]any
	readJSON(_usersFile, &_usersMu, &users)
	jsonUser := map[string]any{
		"id":           uid,
		"phone":        body.Phone,
		"passwordHash": string(hashedPassword),
		"name":         name,
		"role":         "individual",
		"avatar":       "",
		"createTime":   now.Format(time.RFC3339),
	}
	users = append(users, jsonUser)
	writeJSON(_usersFile, &_usersMu, users)

	accessToken, err := s.tokens.Issue(domain.Actor{ID: uid, Role: domain.RoleIndividual}, 15*time.Minute)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, fmt.Errorf("issue access token: %w", err))
		return
	}
	refreshToken, err := service.GenerateRefreshToken()
	if err != nil {
		fail(w, r, http.StatusInternalServerError, fmt.Errorf("generate refresh token: %w", err))
		return
	}
	s.refreshRepo.Store(r.Context(), uid, service.HashToken(refreshToken), time.Now().Add(7*24*time.Hour))

	safeUser := map[string]any{}
	for k, v := range jsonUser {
		if k != "password" && k != "passwordHash" {
			safeUser[k] = v
		}
	}

	respond(w, r, http.StatusOK, map[string]any{
		"success":      true,
		"user":         safeUser,
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
	})
}

func (s *Server) h5AuthMe(w http.ResponseWriter, r *http.Request) {
	// Parse token manually: /api/auth/* paths skip auth middleware.
	h := r.Header.Get("Authorization")
	token := strings.TrimPrefix(h, "Bearer ")
	if token == "" || token == h {
		fail(w, r, http.StatusUnauthorized, errBadRequest("not authenticated"))
		return
	}
	actor, err := s.tokens.Verify(token)
	if err != nil {
		fail(w, r, http.StatusUnauthorized, errBadRequest("not authenticated"))
		return
	}
	// Try real user repo first, fall back to users.json.
	u, repoErr := s.userRepo.FindByID(r.Context(), actor.ID)
	if repoErr == nil {
		// name/phone live in the legacy users.json for password accounts;
		// phone is derivable from the "phone:" openid prefix.
		name := ""
		var users []map[string]any
		readJSON(_usersFile, &_usersMu, &users)
		for _, ju := range users {
			if ju["id"] == u.ID {
				if n, ok := ju["name"].(string); ok {
					name = n
				}
				break
			}
		}
		phone := ""
		if strings.HasPrefix(u.WechatOpenID, "phone:") {
			phone = strings.TrimPrefix(u.WechatOpenID, "phone:")
		}
		respond(w, r, http.StatusOK, map[string]any{
			"success": true,
			"user": map[string]any{
				"id":         u.ID,
				"role":       string(u.Role),
				"status":     u.Status,
				"name":       name,
				"phone":      phone,
				"avatar_url": u.AvatarURL,
			},
		})
		return
	}
	var users []map[string]any
	readJSON(_usersFile, &_usersMu, &users)
	var user map[string]any
	for _, ju := range users {
		if ju["id"] == actor.ID {
			user = ju
			break
		}
	}
	if user == nil {
		user = map[string]any{"id": actor.ID, "role": string(actor.Role)}
	}
	safe := map[string]any{}
	for k, v := range user {
		if k != "password" && k != "passwordHash" {
			safe[k] = v
		}
	}
	respond(w, r, http.StatusOK, map[string]any{"success": true, "user": safe})
}

func (s *Server) h5AuthRefresh(w http.ResponseWriter, r *http.Request) {
	s.refreshToken(w, r)
}

func (s *Server) h5AuthLogout(w http.ResponseWriter, r *http.Request) {
	s.logout(w, r)
}

func (s *Server) h5AuthWechatOAuthURL(w http.ResponseWriter, r *http.Request) {
	// Go backend doesn't have WeChat MP OAuth configured
	_ = config.GetPlatformConfig() // reference for future config
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	respond(w, r, http.StatusOK, map[string]any{
		"success": true,
		"authUrl": baseURL + "/api/auth/wechat-oauth/callback?mock=1",
	})
}

func (s *Server) h5SSOLogin(w http.ResponseWriter, r *http.Request) {
	fail(w, r, http.StatusBadRequest, errBadRequest("SSO login not available in standalone mode"))
}

func (s *Server) h5SSOVerify(w http.ResponseWriter, r *http.Request) {
	fail(w, r, http.StatusBadRequest, errBadRequest("SSO verify not available in standalone mode"))
}

// ─── Cases ──────────────────────────────────────────────────────────────────

func (s *Server) h5Cases(w http.ResponseWriter, r *http.Request) {
	var cases []map[string]any
	if err := readJSON(_casesFile, &_casesMu, &cases); err != nil {
		cases = []map[string]any{}
	}
	sort.Slice(cases, func(i, j int) bool {
		di, _ := cases[i]["id"].(float64)
		dj, _ := cases[j]["id"].(float64)
		return di > dj
	})
	// 分类筛选（前端传 categoryId；"全部分类"为 null，axios 会省略该参数）
	if q := r.URL.Query().Get("categoryId"); q != "" && q != "null" {
		if want, err := strconv.ParseFloat(q, 64); err == nil {
			filtered := make([]map[string]any, 0, len(cases))
			for _, c := range cases {
				if cid, _ := c["categoryId"].(float64); cid == want {
					filtered = append(filtered, c)
				}
			}
			cases = filtered
		}
	}
	// 关键词筛选（标题）
	if kw := strings.TrimSpace(r.URL.Query().Get("keyword")); kw != "" {
		filtered := make([]map[string]any, 0, len(cases))
		for _, c := range cases {
			if title, _ := c["title"].(string); strings.Contains(title, kw) {
				filtered = append(filtered, c)
			}
		}
		cases = filtered
	}
	respond(w, r, http.StatusOK, cases)
}

func (s *Server) h5CasesCreate(w http.ResponseWriter, r *http.Request) {
	var c map[string]any
	if err := decode(r, &c); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	c["id"] = float64(time.Now().UnixMilli())
	var cases []map[string]any
	readJSON(_casesFile, &_casesMu, &cases)
	cases = append([]map[string]any{c}, cases...)
	writeJSON(_casesFile, &_casesMu, cases)
	respond(w, r, http.StatusOK, map[string]any{"success": true, "id": c["id"]})
}

func (s *Server) h5CasesUpdate(w http.ResponseWriter, r *http.Request) {
	var c map[string]any
	if err := decode(r, &c); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	id, _ := c["id"].(float64)
	var cases []map[string]any
	readJSON(_casesFile, &_casesMu, &cases)
	for i, existing := range cases {
		if eid, _ := existing["id"].(float64); eid == id {
			for k, v := range c {
				cases[i][k] = v
			}
			writeJSON(_casesFile, &_casesMu, cases)
			respond(w, r, http.StatusOK, map[string]string{"status": "updated"})
			return
		}
	}
	fail(w, r, http.StatusNotFound, errBadRequest("case not found"))
}

func (s *Server) h5CasesDelete(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ID float64 `json:"id"`
	}
	if err := decode(r, &body); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	var cases []map[string]any
	readJSON(_casesFile, &_casesMu, &cases)
	for i, c := range cases {
		if eid, _ := c["id"].(float64); eid == body.ID {
			cases = append(cases[:i], cases[i+1:]...)
			writeJSON(_casesFile, &_casesMu, cases)
			respond(w, r, http.StatusOK, map[string]string{"status": "deleted"})
			return
		}
	}
	fail(w, r, http.StatusNotFound, errBadRequest("case not found"))
}

// ─── Case Categories ────────────────────────────────────────────────────────

func (s *Server) h5CaseCategories(w http.ResponseWriter, r *http.Request) {
	var cats []map[string]any
	readJSON(_catsFile, &_catsMu, &cats)
	respond(w, r, http.StatusOK, cats)
}

func (s *Server) h5CaseCategoriesCreate(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name    string `json:"name"`
		Service string `json:"service"`
	}
	if err := decode(r, &body); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	var cats []map[string]any
	readJSON(_catsFile, &_catsMu, &cats)
	maxID := 0.0
	for _, c := range cats {
		if id, ok := c["id"].(float64); ok && id > maxID {
			maxID = id
		}
	}
	newCat := map[string]any{
		"id":      maxID + 1,
		"name":    body.Name,
		"service": body.Service,
	}
	cats = append(cats, newCat)
	writeJSON(_catsFile, &_catsMu, cats)
	respond(w, r, http.StatusOK, map[string]any{"success": true, "data": newCat})
}

func (s *Server) h5CaseCategoriesUpdate(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ID      float64 `json:"id"`
		Name    string  `json:"name"`
		Service string  `json:"service"`
	}
	if err := decode(r, &body); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	var cats []map[string]any
	readJSON(_catsFile, &_catsMu, &cats)
	for i, c := range cats {
		if eid, _ := c["id"].(float64); eid == body.ID {
			if body.Name != "" {
				cats[i]["name"] = body.Name
			}
			if body.Service != "" {
				cats[i]["service"] = body.Service
			}
			writeJSON(_catsFile, &_catsMu, cats)
			respond(w, r, http.StatusOK, map[string]any{"success": true, "data": cats[i]})
			return
		}
	}
	fail(w, r, http.StatusNotFound, errBadRequest("category not found"))
}

func (s *Server) h5CaseCategoriesDelete(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ID float64 `json:"id"`
	}
	if err := decode(r, &body); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	var cats []map[string]any
	readJSON(_catsFile, &_catsMu, &cats)
	for i, c := range cats {
		if eid, _ := c["id"].(float64); eid == body.ID {
			cats = append(cats[:i], cats[i+1:]...)
			writeJSON(_catsFile, &_catsMu, cats)
			respond(w, r, http.StatusOK, map[string]string{"status": "deleted"})
			return
		}
	}
	fail(w, r, http.StatusNotFound, errBadRequest("category not found"))
}

// ─── Reviews ────────────────────────────────────────────────────────────────

func (s *Server) h5Reviews(w http.ResponseWriter, r *http.Request) {
	var reviews []map[string]any
	readJSON(_reviewsFile, &_reviewsMu, &reviews)
	section := r.URL.Query().Get("section")
	filtered := make([]map[string]any, 0)
	for _, rv := range reviews {
		if s, _ := rv["status"].(string); s == "approved" {
			if section == "" || rv["section"] == section {
				filtered = append(filtered, rv)
			}
		}
	}
	sort.Slice(filtered, func(i, j int) bool {
		di, _ := filtered[i]["createTime"].(string)
		dj, _ := filtered[j]["createTime"].(string)
		return di > dj
	})
	respond(w, r, http.StatusOK, map[string]any{"success": true, "data": filtered})
}

func (s *Server) h5ReviewsSubmit(w http.ResponseWriter, r *http.Request) {
	var body map[string]any
	if err := decode(r, &body); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	body["id"] = "review-" + time.Now().Format("20060102150405")
	body["status"] = "pending"
	body["createTime"] = time.Now().Format(time.RFC3339)
	var reviews []map[string]any
	readJSON(_reviewsFile, &_reviewsMu, &reviews)
	reviews = append(reviews, body)
	writeJSON(_reviewsFile, &_reviewsMu, reviews)
	respond(w, r, http.StatusOK, map[string]any{"success": true, "message": "评价提交成功"})
}

func (s *Server) h5ReviewsCourses(w http.ResponseWriter, r *http.Request) {
	respond(w, r, http.StatusOK, map[string]any{"success": true, "data": []any{}})
}

func (s *Server) h5AdminReviews(w http.ResponseWriter, r *http.Request) {
	var reviews []map[string]any
	readJSON(_reviewsFile, &_reviewsMu, &reviews)
	respond(w, r, http.StatusOK, map[string]any{"success": true, "data": reviews})
}

func (s *Server) h5AdminReviewUpdate(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body struct {
		Status string `json:"status"`
	}
	if err := decode(r, &body); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	var reviews []map[string]any
	readJSON(_reviewsFile, &_reviewsMu, &reviews)
	for i, rv := range reviews {
		if rv["id"] == id {
			reviews[i]["status"] = body.Status
			writeJSON(_reviewsFile, &_reviewsMu, reviews)
			respond(w, r, http.StatusOK, map[string]string{"status": "updated"})
			return
		}
	}
	fail(w, r, http.StatusNotFound, errBadRequest("review not found"))
}

func (s *Server) h5AdminReviewDelete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var reviews []map[string]any
	readJSON(_reviewsFile, &_reviewsMu, &reviews)
	for i, rv := range reviews {
		if rv["id"] == id {
			reviews = append(reviews[:i], reviews[i+1:]...)
			writeJSON(_reviewsFile, &_reviewsMu, reviews)
			respond(w, r, http.StatusOK, map[string]string{"status": "deleted"})
			return
		}
	}
	fail(w, r, http.StatusNotFound, errBadRequest("review not found"))
}

// ─── Study Showcase ─────────────────────────────────────────────────────────

func (s *Server) h5StudyShowcase(w http.ResponseWriter, r *http.Request) {
	var cfg map[string]any
	readJSON(_servicesFile, &_servicesMu, &cfg)
	if s9, ok := cfg["9"].(map[string]any); ok {
		if ss, ok := s9["studyShowcase"]; ok {
			respond(w, r, http.StatusOK, map[string]any{"success": true, "data": ss})
			return
		}
	}
	respond(w, r, http.StatusOK, map[string]any{"success": true, "data": []any{}})
}

func (s *Server) h5StudyShowcaseSave(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Showcase []map[string]any `json:"showcase"`
		Items    []map[string]any `json:"items"`
	}
	if err := decode(r, &body); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	items := body.Showcase
	if len(items) == 0 {
		items = body.Items
	}
	var cfg map[string]any
	readJSON(_servicesFile, &_servicesMu, &cfg)
	if cfg["9"] == nil {
		cfg["9"] = map[string]any{}
	}
	s9 := cfg["9"].(map[string]any)
	s9["studyShowcase"] = items
	writeJSON(_servicesFile, &_servicesMu, cfg)
	respond(w, r, http.StatusOK, map[string]string{"status": "saved"})
}

// ─── Image Proxy ────────────────────────────────────────────────────────────

func (s *Server) h5ImageProxy(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Query().Get("url")
	if target == "" {
		fail(w, r, http.StatusBadRequest, errBadRequest("url required"))
		return
	}
	// Serve the image directly if it's in uploads
	if strings.HasPrefix(target, "/uploads/") || strings.HasPrefix(target, "uploads/") {
		clean := strings.TrimPrefix(target, "/")
		resolved := filepath.Clean(filepath.Join(".", clean))
		// 防止路径穿越：确保解析后的路径在 uploads/ 目录内
		if !strings.HasPrefix(resolved, "uploads"+string(filepath.Separator)) && resolved != "uploads" {
			fail(w, r, http.StatusForbidden, errBadRequest("invalid path"))
			return
		}
		http.ServeFile(w, r, resolved)
		return
	}
	// 仅允许 http/https 且域名在白名单内的跳转（本机 + BASE_URL 域名），
	// 防止任意 URL 302 跳转被用作开放重定向（钓鱼/绕过 Referer 检查）。
	u, err := url.Parse(target)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		fail(w, r, http.StatusBadRequest, errBadRequest("unsupported url scheme"))
		return
	}
	host := strings.ToLower(u.Hostname())
	allowed := map[string]bool{"localhost": true, "127.0.0.1": true}
	if base := os.Getenv("BASE_URL"); base != "" {
		if bu, err := url.Parse(base); err == nil && bu.Hostname() != "" {
			allowed[strings.ToLower(bu.Hostname())] = true
		}
	}
	if !allowed[host] {
		fail(w, r, http.StatusForbidden, errBadRequest("external image redirect not allowed"))
		return
	}
	http.Redirect(w, r, target, http.StatusFound)
}

// ─── Route Registration ─────────────────────────────────────────────────────

// registerH5AuthRoutes registers the H5 auth + services-config routes
// unconditionally (production needs them: the H5/Admin SPA logs in through
// /api/auth/* and the home page reads /api/services/config). The handlers are
// backed by the real PG/memory repositories (bcrypt password check included).
// The remaining JSON-file-backed legacy routes stay dev-only in registerH5Compat.
func (s *Server) registerH5AuthRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/auth/login", s.h5AuthLogin)
	mux.HandleFunc("POST /api/auth/register", s.h5AuthRegister)
	mux.HandleFunc("GET /api/auth/me", s.h5AuthMe)
	mux.HandleFunc("POST /api/auth/refresh", s.h5AuthRefresh)
	mux.HandleFunc("POST /api/auth/logout", s.h5AuthLogout)
	mux.HandleFunc("POST /api/auth/send-code", s.sendSMSCode)
	mux.HandleFunc("POST /api/auth/login-code", s.loginWithSMS)
	// 系统配置读写均无条件注册（管理后台 ServiceConfigList 在生产环境需要保存配置）
	mux.HandleFunc("GET /api/services/config", s.h5GetServicesConfig)
	mux.HandleFunc("POST /api/services/config", s.h5SaveServicesConfig)
	// 服务申请提交生产注册（写入 service_applications 表；列表/更新等仍 dev-only）
	mux.HandleFunc("POST /api/submit", s.h5SubmitApplication)
}

func (s *Server) registerH5Compat(mux *http.ServeMux) {
	// Admin Services Config (same handlers; dev-only — 管理后台用 /api/services/config)
	mux.HandleFunc("GET /api/admin/services/config", s.h5GetServicesConfig)
	mux.HandleFunc("POST /api/admin/services/config", s.h5SaveServicesConfig)

	// Upload
	mux.HandleFunc("POST /api/upload", s.h5Upload)

	// Applications（提交已生产注册，其余 JSON 文件路由 dev-only）
	mux.HandleFunc("GET /api/list", s.h5ListApplications)
	mux.HandleFunc("POST /api/update", s.h5UpdateApplication)
	mux.HandleFunc("GET /api/export", s.h5ExportApplications)

	// Admin Applications
	mux.HandleFunc("GET /api/admin/applications", s.h5AdminApplications)
	mux.HandleFunc("GET /api/admin/applications/{id}", s.h5AdminApplicationByID)
	mux.HandleFunc("POST /api/admin/applications/{id}", s.h5AdminUpdateApplication)

	// Dashboard
	mux.HandleFunc("GET /api/admin/stats", s.h5AdminStats)

	// Users
	mux.HandleFunc("GET /api/users", s.h5Users)
	mux.HandleFunc("POST /api/user/role", s.h5UpdateUserRole)
	mux.HandleFunc("POST /api/user/update", s.h5UpdateUserProfile)

	// Auth compatibility — login/register/me/refresh/logout are registered
	// unconditionally in registerH5AuthRoutes (production login depends on them).
	// WeChat OAuth + SSO remain dev-only.
	mux.HandleFunc("GET /api/auth/wechat-oauth-url", s.h5AuthWechatOAuthURL)
	mux.HandleFunc("POST /api/sso/login", s.h5SSOLogin)
	mux.HandleFunc("POST /api/sso/verify", s.h5SSOVerify)

	// Cases
	mux.HandleFunc("GET /api/cases", s.h5Cases)
	mux.HandleFunc("POST /api/cases/create", s.h5CasesCreate)
	mux.HandleFunc("POST /api/cases/update", s.h5CasesUpdate)
	mux.HandleFunc("POST /api/cases/delete", s.h5CasesDelete)

	// Case Categories
	mux.HandleFunc("GET /api/case-categories", s.h5CaseCategories)
	mux.HandleFunc("POST /api/case-categories/create", s.h5CaseCategoriesCreate)
	mux.HandleFunc("POST /api/case-categories/update", s.h5CaseCategoriesUpdate)
	mux.HandleFunc("POST /api/case-categories/delete", s.h5CaseCategoriesDelete)

	// Reviews
	mux.HandleFunc("GET /api/reviews", s.h5Reviews)
	mux.HandleFunc("POST /api/reviews", s.h5ReviewsSubmit)
	mux.HandleFunc("GET /api/reviews/courses", s.h5ReviewsCourses)
	mux.HandleFunc("GET /api/admin/reviews", s.h5AdminReviews)
	mux.HandleFunc("POST /api/admin/reviews/{id}", s.h5AdminReviewUpdate)
	mux.HandleFunc("DELETE /api/admin/reviews/{id}", s.h5AdminReviewDelete)

	// Study Showcase
	mux.HandleFunc("GET /api/study/showcase", s.h5StudyShowcase)
	mux.HandleFunc("POST /api/study/showcase", s.h5StudyShowcaseSave)
	mux.HandleFunc("GET /api/admin/study/showcase", s.h5StudyShowcase)
	mux.HandleFunc("POST /api/admin/study/showcase", s.h5StudyShowcaseSave)

	// Image proxy
	mux.HandleFunc("GET /api/image", s.h5ImageProxy)

	// Client IP
	mux.HandleFunc("GET /api/client-ip", func(w http.ResponseWriter, r *http.Request) {
		respond(w, r, http.StatusOK, map[string]any{"success": true, "ip": r.RemoteAddr})
	})
}

// ─── Helpers ────────────────────────────────────────────────────────────────

func errBadRequest(msg string) error {
	return &h5CompatError{msg: msg}
}

type h5CompatError struct{ msg string }

func (e *h5CompatError) Error() string { return e.msg }

func hashString(s string) uint32 {
	h := uint32(2166136261)
	for i := 0; i < len(s); i++ {
		h ^= uint32(s[i])
		h *= 16777619
	}
	return h
}

func actorFromMap(id, role string) domain.Actor {
	return domain.Actor{ID: id, Role: domain.Role(role)}
}

func randomSuffix(n int) string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			idx = big.NewInt(0)
		}
		b[i] = chars[idx.Int64()]
	}
	return string(b)
}
