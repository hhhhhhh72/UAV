package httpapi

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"sort"
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
	_appsMu        sync.RWMutex
	_appsFile      = "applications.json"
	_casesMu       sync.RWMutex
	_casesFile     = "cases.json"
	_catsMu        sync.RWMutex
	_catsFile      = "case_categories.json"
	_reviewsMu     sync.RWMutex
	_reviewsFile   = "reviews.json"
	_usersMu       sync.RWMutex
	_usersFile     = "users.json"
	_servicesMu    sync.RWMutex
	_servicesFile  = "services_config.json"
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
	return os.WriteFile(path, raw, 0644)
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
	respond(w, r, http.StatusOK, map[string]string{"status": "saved"})
}

// ─── File Upload ────────────────────────────────────────────────────────────

func (s *Server) h5Upload(w http.ResponseWriter, r *http.Request) {
	s.uploadFile(w, r)
}

// ─── Applications (JSON file backed) ────────────────────────────────────────

type h5Application struct {
	ID        string `json:"id"`
	OrderNo   string `json:"orderNo,omitempty"`
	ServiceID string `json:"serviceId"`
	ServiceName string `json:"serviceName"`
	Status    string `json:"status"`
	UserID    string `json:"userId"`
	ContactName  string `json:"contactName,omitempty"`
	ContactPhone string `json:"contactPhone,omitempty"`
	TraineeName  string `json:"traineeName,omitempty"`
	TraineePhone string `json:"traineePhone,omitempty"`
	Name       string `json:"name,omitempty"`
	Phone      string `json:"phone,omitempty"`
	CompanyName string `json:"companyName,omitempty"`
	CompetitionRole string `json:"competitionRole,omitempty"`
	CompetitionRoleText string `json:"competitionRoleText,omitempty"`
	CompetitionGroup string `json:"competitionGroup,omitempty"`
	CompetitionProject string `json:"competitionProject,omitempty"`
	Gender     string `json:"gender,omitempty"`
	IDCard     string `json:"idCard,omitempty"`
	RegNo      string `json:"regNo,omitempty"`
	Level      string `json:"level,omitempty"`
	ValidDate  string `json:"validDate,omitempty"`
	Manager    string `json:"manager,omitempty"`
	ManagerPhone string `json:"managerPhone,omitempty"`
	ContactPerson string `json:"contactPerson,omitempty"`
	CustomerType string `json:"customerType,omitempty"`
	CargoType  string `json:"cargoType,omitempty"`
	StartAddress string `json:"startAddress,omitempty"`
	EndAddress  string `json:"endAddress,omitempty"`
	TraineeGender string `json:"traineeGender,omitempty"`
	TraineeIDCard string `json:"traineeIdCard,omitempty"`
	LicenseLevel string `json:"licenseLevel,omitempty"`
	ExamModel  string `json:"examModel,omitempty"`
	Remark     string `json:"remark,omitempty"`
	Email      string `json:"email,omitempty"`
	Location   string `json:"location,omitempty"`
	CreateTime string `json:"createTime"`
	ApplyTime  string `json:"applyTime,omitempty"`
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
	var app h5Application
	if err := decode(r, &app); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	app.ID = time.Now().Format("20060102150405") + randomSuffix(4)
	app.CreateTime = time.Now().Format(time.RFC3339)
	if app.ApplyTime == "" {
		app.ApplyTime = time.Now().Format("2006-01-02 15:04:05")
	}
	if app.Status == "" {
		app.Status = "待处理"
	}

	var apps []h5Application
	readJSON(_appsFile, &_appsMu, &apps)
	apps = append(apps, app)
	if err := writeJSON(_appsFile, &_appsMu, apps); err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, map[string]any{"success": true, "id": app.ID})
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

	var user map[string]any
	var users []map[string]any

	// 1) Check real user repo (Go backend users).
	if u, err := s.userRepo.FindByID(loginID); err == nil {
		user = map[string]any{"id": u.ID, "username": u.ID, "phone": u.ID, "role": string(u.Role), "status": u.Status}
	} else {
		// 2) Fallback to users.json (legacy compat).
		readJSON(_usersFile, &_usersMu, &users)
		for _, ju := range users {
			if ju["phone"] == loginID || ju["username"] == loginID {
				user = ju
				break
			}
		}
	}

	// Dev mode: super admin phone can login with any password.
	if user == nil && adminDevMode() && os.Getenv("SUPER_ADMIN_PHONE") != "" && loginID == os.Getenv("SUPER_ADMIN_PHONE") {
		user = map[string]any{"id": "user-" + loginID, "username": loginID, "phone": loginID, "role": "platform_admin", "status": "active"}
		users = append(users, user)
		writeJSON(_usersFile, &_usersMu, users)
	}
	if user == nil {
		fail(w, r, http.StatusUnauthorized, errBadRequest("账号或密码错误"))
		return
	}

	// Issue Go backend tokens via dev-mode WeChat login
	id, _ := user["id"].(string)
	if id == "" {
		id = "user-" + loginID
		user["id"] = id
	}
	role, _ := user["role"].(string)
	if role == "" {
		role = "individual"
	}

	accessToken, _ := s.tokens.IssueJWT(actorFromMap(id, role), 15*time.Minute)
	refreshToken, _ := service.GenerateRefreshToken()
	tokenHash := service.HashToken(refreshToken)
	s.refreshRepo.Store(id, tokenHash, time.Now().Add(7*24*time.Hour))

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

	var users []map[string]any
	readJSON(_usersFile, &_usersMu, &users)
	for _, u := range users {
		if u["phone"] == body.Phone {
			fail(w, r, http.StatusConflict, errBadRequest("user already exists"))
			return
		}
	}

	name := body.Name
	if name == "" {
		name = "User" + body.Phone[len(body.Phone)-4:]
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	newUser := map[string]any{
		"id":           "user-" + body.Phone,
		"phone":        body.Phone,
		"passwordHash": string(hashedPassword),
		"name":         name,
		"role":         "user",
		"avatar":       "",
		"createTime":   time.Now().Format(time.RFC3339),
	}
	users = append(users, newUser)
	writeJSON(_usersFile, &_usersMu, users)

	accessToken, _ := s.tokens.Issue(actorFromMap(newUser["id"].(string), "individual"), 15*time.Minute)
	refreshToken, _ := service.GenerateRefreshToken()
	s.refreshRepo.Store(newUser["id"].(string), service.HashToken(refreshToken), time.Now().Add(7*24*time.Hour))

	safeUser := map[string]any{}
	for k, v := range newUser {
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
	u, repoErr := s.userRepo.FindByID(actor.ID)
	if repoErr == nil {
		respond(w, r, http.StatusOK, map[string]any{
			"success": true,
			"user":    map[string]any{"id": u.ID, "role": string(u.Role), "status": u.Status},
		})
		return
	}
	var users []map[string]any
	readJSON(_usersFile, &_usersMu, &users)
	var user map[string]any
	for _, ju := range users {
		if ju["id"] == actor.ID { user = ju; break }
	}
	if user == nil {
		user = map[string]any{"id": actor.ID, "role": string(actor.Role)}
	}
	safe := map[string]any{}
	for k, v := range user {
		if k != "password" && k != "passwordHash" { safe[k] = v }
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
	var body struct{ ID float64 `json:"id"` }
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
	var body struct{ ID float64 `json:"id"` }
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
	var body struct{ Status string `json:"status"` }
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
	url := r.URL.Query().Get("url")
	if url == "" {
		fail(w, r, http.StatusBadRequest, errBadRequest("url required"))
		return
	}
	// Serve the image directly if it's in uploads
	if strings.HasPrefix(url, "/uploads/") || strings.HasPrefix(url, "uploads/") {
		clean := strings.TrimPrefix(url, "/")
		resolved := filepath.Clean(filepath.Join(".", clean))
		// 防止路径穿越：确保解析后的路径在 uploads/ 目录内
		if !strings.HasPrefix(resolved, "uploads"+string(filepath.Separator)) && resolved != "uploads" {
			fail(w, r, http.StatusForbidden, errBadRequest("invalid path"))
			return
		}
		http.ServeFile(w, r, resolved)
		return
	}
	// 仅允许 http/https 跳转，防止 javascript: 等协议注入
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		fail(w, r, http.StatusBadRequest, errBadRequest("unsupported url scheme"))
		return
	}
	http.Redirect(w, r, url, http.StatusFound)
}

// ─── Route Registration ─────────────────────────────────────────────────────

func (s *Server) registerH5Compat(mux *http.ServeMux) {
	// Services Config
	mux.HandleFunc("GET /api/services/config", s.h5GetServicesConfig)
	mux.HandleFunc("POST /api/services/config", s.h5SaveServicesConfig)

	// Admin Services Config (same handlers)
	mux.HandleFunc("GET /api/admin/services/config", s.h5GetServicesConfig)
	mux.HandleFunc("POST /api/admin/services/config", s.h5SaveServicesConfig)

	// Upload
	mux.HandleFunc("POST /api/upload", s.h5Upload)

	// Applications
	mux.HandleFunc("GET /api/list", s.h5ListApplications)
	mux.HandleFunc("POST /api/submit", s.h5SubmitApplication)
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

	// Auth compatibility
	mux.HandleFunc("POST /api/auth/login", s.h5AuthLogin)
	mux.HandleFunc("POST /api/auth/register", s.h5AuthRegister)
	mux.HandleFunc("GET /api/auth/me", s.h5AuthMe)
	mux.HandleFunc("POST /api/auth/refresh", s.h5AuthRefresh)
	mux.HandleFunc("POST /api/auth/logout", s.h5AuthLogout)
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
