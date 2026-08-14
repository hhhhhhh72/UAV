package httpapi

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/service"
)

// phoneRe matches mainland China mobile numbers (11 digits starting with 13-19).
var phoneRe = regexp.MustCompile(`^1[3-9]\d{9}$`)

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

	u, err := s.userRepo.FindByOpenID(sess.OpenID)
	if err != nil {
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
		u, err = s.userRepo.Create(u)
		if err != nil {
			fail(w, r, http.StatusInternalServerError, fmt.Errorf("create user: %w", err))
			return
		}
	}

	role := u.Role
	if role == "" {
		role = domain.RoleIndividual
		// Super admin phone always gets platform_admin role.
		if superPhone := os.Getenv("SUPER_ADMIN_PHONE"); superPhone != "" && u.ID == superPhone {
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
	if err := s.refreshRepo.Store(u.ID, refreshHash, time.Now().Add(7*24*time.Hour)); err != nil {
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
	var req struct{ RefreshToken string `json:"refresh_token"` }
	if err := decode(r, &req); err != nil || req.RefreshToken == "" {
		fail(w, r, http.StatusBadRequest, errors.New("refresh_token is required"))
		return
	}
	tokenHash := service.HashToken(req.RefreshToken)
	userID, expiresAt, revoked, err := s.refreshRepo.Find(tokenHash)
	if err != nil || revoked || time.Now().After(expiresAt) {
		fail(w, r, http.StatusUnauthorized, errors.New("invalid or expired refresh token"))
		return
	}

	var u *domain.User
	role := domain.RoleIndividual
	if found, err := s.userRepo.FindByID(userID); err == nil && found.Role != "" {
		u = &found
		role = u.Role
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
	// 轮转顺序：先落库新令牌、成功后再撤销旧令牌。
	// 旧实现先 Revoke 后 Store——Store 失败会导致旧令牌已作废、新令牌未落库，
	// 用户被永久锁死（refresh 不可用）。新顺序下 Store 失败时旧令牌仍有效。
	newHash := service.HashToken(newRefresh)
	if err := s.refreshRepo.Store(userID, newHash, time.Now().Add(7*24*time.Hour)); err != nil {
		fail(w, r, http.StatusInternalServerError, fmt.Errorf("store refresh token: %w", err))
		return
	}
	s.refreshRepo.Revoke(tokenHash)

	hasWechat := u != nil && u.WechatOpenID != "" && !strings.HasPrefix(u.WechatOpenID, "phone:")
	ui := userInfo{ID: userID, Role: role, Status: ""}
	if u != nil {
		ui.Name = u.Name
		ui.AvatarURL = u.AvatarURL
		ui.HasWechat = hasWechat
	}
	respond(w, r, http.StatusOK, authResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefresh,
		ExpiresIn:    900,
		User:         ui,
	})
}

// POST /api/v1/auth/logout
func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	var req struct{ RefreshToken string `json:"refresh_token"` }
	if err := decode(r, &req); err == nil && req.RefreshToken != "" {
		s.refreshRepo.Revoke(service.HashToken(req.RefreshToken))
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
		if err := s.userRepo.UpdateAvatar(a.ID, in.AvatarURL); err != nil {
			fail(w, r, http.StatusInternalServerError, err)
			return
		}
	}
	if in.Name != "" {
		if err := s.userRepo.UpdateName(a.ID, in.Name); err != nil {
			fail(w, r, http.StatusInternalServerError, err)
			return
		}
	}
	if in.Phone != "" || in.Gender != "" || in.Birthday != "" || in.Region != "" || in.Bio != "" {
		if err := s.userRepo.UpdateProfile(a.ID, domain.UserProfile{
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
	if ds, err := s.demands.ListByPublisher(a.ID); err == nil {
		demandCount = len(ds)
	}
	var certCount int
	if s.trainingSvc != nil {
		certs, err := s.trainingSvc.ListMyCertificates(a)
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
	if u, err := s.userRepo.FindByID(a.ID); err == nil {
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
