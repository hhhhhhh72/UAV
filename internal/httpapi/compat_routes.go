package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"drone-platform/internal/domain"
)

func (s *Server) registerCompatRoutes(mux *http.ServeMux) {
	// Only non-overlapping routes not covered by registerH5Compat.
	mux.HandleFunc("POST /api/auth/wechat/login", s.wechatLogin)
	mux.HandleFunc("POST /api/auth/wx-login", s.wechatLogin)
	mux.HandleFunc("POST /api/auth/wx-phone", s.wxPhone)
}

// passwordLogin handles old /api/auth/login with phone+password.
func (s *Server) passwordLogin(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Phone    string `json:"phone"`
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	loginID := in.Phone
	if loginID == "" {
		loginID = in.Username
	}
	if loginID == "" || in.Password == "" {
		fail(w, r, http.StatusBadRequest, fmt.Errorf("账号或密码不能为空"))
		return
	}

	// Find user by phone or username
	users, _ := s.userRepo.All()
	var found *domain.User
	for _, u := range users {
		if u.ID == loginID || u.WechatOpenID == loginID {
			found = &u
			break
		}
	}

	// Dev mode: accept any user, any password. Create user if not exists.
	inDev := adminDevMode()
	if inDev {
		if found == nil {
			now := time.Now()
			newUser := domain.User{
				ID:           loginID,
				WechatOpenID: loginID,
				Role:         domain.RolePlatformAdmin,
				Status:       "active",
				Version:      1,
				CreatedAt:    now,
				UpdatedAt:    now,
			}
			if _, err := s.userRepo.Create(newUser); err == nil {
				found = &newUser
			}
		}
		// In dev mode, skip password check
	} else {
		if found == nil {
			fail(w, r, http.StatusUnauthorized, fmt.Errorf("账号或密码错误"))
			return
		}
		// Production would check bcrypt(password, user.passwordHash)
	}

	if found == nil {
		fail(w, r, http.StatusUnauthorized, fmt.Errorf("账号或密码错误"))
		return
	}

	actor := domain.Actor{ID: found.ID, Role: found.Role}
	access, _ := s.tokens.Issue(actor, 15*time.Minute)
	refresh, _ := s.tokens.Issue(actor, 7*24*time.Hour)
	s.refreshRepo.Store(found.ID, refresh, time.Now().Add(7*24*time.Hour))

	legacyJSON(w, r, http.StatusOK, map[string]any{
		"success":      true,
		"accessToken":  access,
		"refreshToken": refresh,
		"expiresIn":    900,
		"user": map[string]any{
			"id":       found.ID,
			"username": found.ID,
			"phone":    "",
			"role":     string(found.Role),
			"status":   found.Status,
		},
	})
}

// passwordRegister handles old /api/auth/register.
func (s *Server) passwordRegister(w http.ResponseWriter, r *http.Request) {
	// Forward to wechatLogin for simplicity — creates user in dev mode
	s.wechatLogin(w, r)
}

// getMeLegacy returns user info in legacy format for /api/auth/me.
// Does its own token parsing since /api/auth/* skips the auth middleware.
func (s *Server) getMeLegacy(w http.ResponseWriter, r *http.Request) {
	h := r.Header.Get("Authorization")
	token := ""
	if len(h) > 7 && h[:7] == "Bearer " {
		token = h[7:]
	} else {
		fail(w, r, http.StatusUnauthorized, fmt.Errorf("未登录"))
		return
	}
	actor, err := s.tokens.Verify(token)
	if err != nil {
		fail(w, r, http.StatusUnauthorized, fmt.Errorf("未登录"))
		return
	}
	u, err := s.userRepo.FindByID(actor.ID)
	if err != nil {
		fail(w, r, http.StatusNotFound, err)
		return
	}
	legacyJSON(w, r, http.StatusOK, map[string]any{
		"success": true,
		"user": map[string]any{
			"id":     u.ID,
			"role":   string(u.Role),
			"status": u.Status,
		},
	})
}

// legacyJSON writes a structured JSON response (compat format).
func legacyJSON(w http.ResponseWriter, r *http.Request, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// listAllAdapter / submitAdapter / updateAdapter
func (s *Server) listAllAdapter(w http.ResponseWriter, r *http.Request) {
	role := r.URL.Query().Get("role")
	userID := r.URL.Query().Get("userId")
	if role == "admin" {
		s.adminDashboard(w, r)
		return
	}
	if userID != "" {
		s.listMyProjectApps(w, r)
		return
	}
	respond(w, r, http.StatusOK, []any{})
}

func (s *Server) submitAdapter(w http.ResponseWriter, r *http.Request) {
	s.createDemand(w, r)
}

func (s *Server) updateAdapter(w http.ResponseWriter, r *http.Request) {
	respond(w, r, http.StatusOK, map[string]string{"status": "ok"})
}

// POST /api/auth/wx-phone — 绑定微信手机号 (dev mode: 直接接收 phone)
func (s *Server) wxPhone(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	if in.Phone != "" {
		// Dev mode: directly accept phone number
		respond(w, r, http.StatusOK, map[string]any{
			"phone": in.Phone,
			"msg":   "phone bound (dev mode)",
		})
		return
	}
	if in.Code != "" {
		// WeChat phone code — in production, exchange via WeChat API
		respond(w, r, http.StatusOK, map[string]any{
			"phone": "138****" + in.Code[len(in.Code)-4:],
			"msg":   "phone bound (wx code)",
		})
		return
	}
	fail(w, r, http.StatusBadRequest, errors.New("phone or code required"))
}
