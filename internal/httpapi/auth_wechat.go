package httpapi

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/service"
)

// phoneRe matches mainland China mobile numbers (11 digits starting with 13-19).
var phoneRe = regexp.MustCompile(`^1[3-9]\d{9}$`)

// ── 按 key 的进程内登录锁 ──
// loginLockByKey 为登录 find-or-create 提供按 key 的互斥（防并发首登重复建号）。
// 与 service 层 lockByKey 同构：引用计数归零即从池中删除条目，防止 key 无限累积。
var loginLocks sync.Map // key -> *loginLockEntry

// loginLockEntry 是登录锁条目：mu 是实际互斥锁；refs 引用计数（受 refMu 保护），
// 语义为"当前持有者 + 已递增但仍在等待 mu 的获取者"数量。
type loginLockEntry struct {
	mu    sync.Mutex
	refMu sync.Mutex
	refs  int
}

func loginLockByKey(key string) func() {
	for {
		created := false
		v, loaded := loginLocks.LoadOrStore(key, &loginLockEntry{refs: 1})
		if !loaded {
			created = true // 本次创建者自带一个引用，下面不再递增
		}
		e := v.(*loginLockEntry)
		e.refMu.Lock()
		if e.refs == 0 {
			// 条目刚被并发删除（refs 已归零），本 goroutine 拿到的指针已失效：重试。
			e.refMu.Unlock()
			continue
		}
		if !created {
			e.refs++
		}
		e.refMu.Unlock()

		e.mu.Lock()
		return func() {
			e.mu.Unlock()
			e.refMu.Lock()
			e.refs--
			if e.refs == 0 {
				loginLocks.Delete(key)
			}
			e.refMu.Unlock()
		}
	}
}

type wechatLoginRequest struct {
	Code string `json:"code"`
}

type authResponse struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	ExpiresIn    int64    `json:"expires_in"`
	User         userInfo `json:"user"`
}

type userInfo struct {
	ID        string      `json:"id"`
	Role      domain.Role `json:"role"`
	Status    string      `json:"status"`
	Name      string      `json:"name"`       // 昵称（users.name，微信账号可能为空）
	AvatarURL string      `json:"avatar_url"` // 头像
	HasWechat bool        `json:"has_wechat"` // 是否已绑定微信（微信登录恒为 true）
}

// POST /api/v1/auth/wechat/login
func (s *Server) wechatLogin(w http.ResponseWriter, r *http.Request) {
	appID, appSecret := os.Getenv("WECHAT_APPID"), os.Getenv("WECHAT_APPSECRET")
	var req wechatLoginRequest
	if err := decode(r, &req); err != nil || req.Code == "" {
		fail(w, r, http.StatusBadRequest, errors.New("code is required"))
		return
	}

	sess, err := service.WeChatLogin(req.Code, appID, appSecret)
	if err != nil && adminDevMode() {
		// Dev mode: every code maps to one fixed openid so silent login reuses
		// a single dev user instead of creating a new row per login attempt.
		sess = service.WeChatSession{OpenID: "dev-fixed", SessionKey: "dev"}
	} else if err != nil {
		fail(w, r, http.StatusUnauthorized, err)
		return
	}

	u, err := s.userRepo.FindByOpenID(r.Context(), sess.OpenID)
	if err != nil {
		// P2 修复：同 openid 并发首登会双双 miss 再各 Create 一个用户（重复账号）。
		// 按 openid 加进程内锁，锁内复查一次再创建；后到者复用先建好的用户。
		unlock := loginLockByKey("openid|" + sess.OpenID)
		u, findErr := s.userRepo.FindByOpenID(r.Context(), sess.OpenID)
		if findErr != nil {
			now := time.Now()
			u = domain.User{
				ID:           fmt.Sprintf("user-%d", now.UnixNano()),
				WechatOpenID: sess.OpenID,
				Role:         domain.RoleIndividual,
				Status:       "active",
				Version:      1,
				CreatedAt:    now,
				UpdatedAt:    now,
			}
			u, findErr = s.userRepo.Create(r.Context(), u)
		}
		unlock()
		if findErr != nil {
			fail(w, r, http.StatusInternalServerError, fmt.Errorf("create user: %w", findErr))
			return
		}
	}

	role := u.Role
	if role == "" {
		role = domain.RoleIndividual
		// Super admin phone always gets platform_admin role.
		// 用户 ID 形如 "user-<手机号>"（手机号注册）或 "user-<unixnano>"（微信首登），
		// 微信注册用户 openid 为 "phone:<手机号>"——环境变量为裸手机号，需带前缀比较。
		if superPhone := os.Getenv("SUPER_ADMIN_PHONE"); superPhone != "" &&
			(u.ID == superPhone || u.ID == "user-"+superPhone || u.WechatOpenID == "phone:"+superPhone) {
			role = domain.RolePlatformAdmin
		}
	}
	actor := domain.Actor{ID: u.ID, Role: role}
	accessToken, err := s.tokens.IssueJWT(actor, 15*time.Minute)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}

	refreshToken, err := service.GenerateRefreshToken()
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	refreshHash := service.HashToken(refreshToken)
	if err := s.refreshRepo.Store(r.Context(), u.ID, refreshHash, time.Now().Add(7*24*time.Hour)); err != nil {
		fail(w, r, http.StatusInternalServerError, fmt.Errorf("store refresh token: %w", err))
		return
	}

	s.audit(r.Context(), u.ID, "login_wechat", "auth", u.ID, "success")
	respond(w, r, http.StatusOK, authResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    900,
		User:         userInfo{ID: u.ID, Role: role, Status: u.Status, Name: u.Name, AvatarURL: u.AvatarURL, HasWechat: true},
	})
}

// POST /api/v1/auth/refresh
func (s *Server) refreshToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := decode(r, &req); err != nil || req.RefreshToken == "" {
		fail(w, r, http.StatusBadRequest, errors.New("refresh_token is required"))
		return
	}
	tokenHash := service.HashToken(req.RefreshToken)
	// 原子消费旧令牌：并发同一令牌二次刷新时仅一个成功（防 TOCTOU 双签发）
	found, userID, _, err := s.refreshRepo.Consume(r.Context(), tokenHash)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	if !found {
		fail(w, r, http.StatusUnauthorized, errors.New("invalid or expired refresh token"))
		return
	}

	// P3 修复：用户已删除/不存在时拒绝续期——此前回退为 individual 继续签发，
	// 被删除账号的 refresh token 仍可无限续期（deleteUser 后会话不失效）。
	u, err := s.userRepo.FindByID(r.Context(), userID)
	if err != nil {
		fail(w, r, http.StatusUnauthorized, errors.New("user not found"))
		return
	}
	role := u.Role
	if role == "" {
		role = domain.RoleIndividual
	}
	// 与登录路径保持一致：统一签发标准 JWT（此前这里用旧式两段 Issue）。
	accessToken, err := s.tokens.IssueJWT(domain.Actor{ID: userID, Role: role}, 15*time.Minute)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}

	newRefresh, err := service.GenerateRefreshToken()
	if err != nil {
		fail(w, r, http.StatusInternalServerError, fmt.Errorf("generate refresh token: %w", err))
		return
	}
	// 新令牌落库（旧令牌已在上方原子消费，无需再 Revoke）
	newHash := service.HashToken(newRefresh)
	if err := s.refreshRepo.Store(r.Context(), userID, newHash, time.Now().Add(7*24*time.Hour)); err != nil {
		fail(w, r, http.StatusInternalServerError, fmt.Errorf("store refresh token: %w", err))
		return
	}

	hasWechat := u.WechatOpenID != "" && !strings.HasPrefix(u.WechatOpenID, "phone:")
	ui := userInfo{ID: userID, Role: role, Status: ""}
	ui.Name = u.Name
	ui.AvatarURL = u.AvatarURL
	ui.HasWechat = hasWechat
	respond(w, r, http.StatusOK, authResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefresh,
		ExpiresIn:    900,
		User:         ui,
	})
}

// POST /api/v1/auth/logout
func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := decode(r, &req); err == nil && req.RefreshToken != "" {
		s.refreshRepo.Revoke(r.Context(), service.HashToken(req.RefreshToken))
	}
	a, _ := authenticatedActor(r)
	if a.ID != "" {
		s.audit(r.Context(), a.ID, "logout", "auth", a.ID, "success")
	}
	respond(w, r, http.StatusOK, map[string]string{"status": "logged_out"})
}

// PATCH /api/v1/me — update profile fields (avatar_url/name/phone/gender/birthday/region/bio).
// Phone is encrypted before persistence by the repository; empty values are left unchanged.
func (s *Server) updateMe(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	var in struct {
		Name      string `json:"name"`
		AvatarURL string `json:"avatar_url"`
		Phone     string `json:"phone"`
		Gender    string `json:"gender"`
		Birthday  string `json:"birthday"`
		Region    string `json:"region"`
		Bio       string `json:"bio"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	if in.Phone != "" && !phoneRe.MatchString(in.Phone) {
		fail(w, r, http.StatusBadRequest, errors.New("invalid phone number"))
		return
	}
	if in.AvatarURL != "" {
		if err := s.userRepo.UpdateAvatar(r.Context(), a.ID, in.AvatarURL); err != nil {
			fail(w, r, http.StatusInternalServerError, err)
			return
		}
	}
	if in.Name != "" {
		if err := s.userRepo.UpdateName(r.Context(), a.ID, in.Name); err != nil {
			fail(w, r, http.StatusInternalServerError, err)
			return
		}
	}
	if in.Phone != "" || in.Gender != "" || in.Birthday != "" || in.Region != "" || in.Bio != "" {
		if err := s.userRepo.UpdateProfile(r.Context(), a.ID, domain.UserProfile{
			Phone:    in.Phone,
			Gender:   in.Gender,
			Birthday: in.Birthday,
			Region:   in.Region,
			Bio:      in.Bio,
		}); err != nil {
			fail(w, r, http.StatusInternalServerError, err)
			return
		}
	}
	respond(w, r, http.StatusOK, map[string]any{
		"id": a.ID, "role": a.Role, "status": "active",
		"name": in.Name, "avatar_url": in.AvatarURL, "phone": in.Phone,
		"gender": in.Gender, "birthday": in.Birthday, "region": in.Region, "bio": in.Bio,
	})
}

// GET /api/v1/me
func (s *Server) me(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	// demand_count 只统计当前用户发布的需求（此前误用 List(空 filter) 返回平台全量）。
	demandCount := 0
	if ds, err := s.demands.ListByPublisher(r.Context(), a.ID); err == nil {
		demandCount = len(ds)
	}
	var certCount int
	if s.trainingSvc != nil {
		certs, err := s.trainingSvc.ListMyCertificates(r.Context(), a)
		if err != nil {
			certCount = 0
		} else {
			certCount = len(certs)
		}
	}
	name := ""
	avatarURL := ""
	phone := ""
	gender := ""
	birthday := ""
	region := ""
	bio := ""
	status := "active"
	if u, err := s.userRepo.FindByID(r.Context(), a.ID); err == nil {
		name = u.Name
		avatarURL = u.AvatarURL
		// PhoneCipher holds the decrypted plaintext after FindByID (see repository)
		phone = u.PhoneCipher
		gender = u.Gender
		birthday = u.Birthday
		region = u.Region
		bio = u.Bio
		if u.Status != "" {
			status = u.Status
		}
	}
	respond(w, r, http.StatusOK, map[string]any{
		"id":           a.ID,
		"role":         string(a.Role),
		"status":       status,
		"name":         name,
		"avatar_url":   avatarURL,
		"phone":        phone,
		"gender":       gender,
		"birthday":     birthday,
		"region":       region,
		"bio":          bio,
		"demand_count": demandCount,
		"cert_count":   certCount,
	})
}
