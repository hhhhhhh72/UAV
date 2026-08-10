package httpapi

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/service"
)

// ── 短信验证码登录 ──
// 验证码存内存（map + 5 分钟 TTL，60s 内不可重发）。
// 接入短信服务商（腾讯云/阿里云 SMS）后，在 sendSMSCode 中发送真实短信并移除 dev_code。

type smsRecord struct {
	Code      string
	ExpiresAt time.Time
}

var smsCodes sync.Map // phone -> smsRecord

const (
	smsCodeTTL    = 5 * time.Minute
	smsResendWait = 60 * time.Second
)

func genSMSCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

func isChinaPhone(p string) bool {
	if len(p) != 11 || p[0] != '1' {
		return false
	}
	for i := 1; i < len(p); i++ {
		if p[i] < '0' || p[i] > '9' {
			return false
		}
	}
	return true
}

// POST /api/auth/send-code
func (s *Server) sendSMSCode(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Phone string `json:"phone"`
	}
	if err := decode(r, &body); err != nil || !isChinaPhone(body.Phone) {
		fail(w, r, http.StatusBadRequest, errBadRequest("请输入正确的手机号"))
		return
	}
	if rec, ok := smsCodes.Load(body.Phone); ok {
		if time.Until(rec.(smsRecord).ExpiresAt) > smsCodeTTL-smsResendWait {
			fail(w, r, http.StatusTooManyRequests, errBadRequest("验证码已发送，请 60 秒后再试"))
			return
		}
	}
	code, err := genSMSCode()
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	smsCodes.Store(body.Phone, smsRecord{Code: code, ExpiresAt: time.Now().Add(smsCodeTTL)})

	// TODO(短信服务商): 接入腾讯云/阿里云 SMS 后在此发送真实短信，并删除 dev_code 回显。
	resp := map[string]any{"success": true, "expires_in": int(smsCodeTTL.Seconds())}
	if adminDevMode() {
		resp["dev_code"] = code // 开发态回显验证码，便于调试
	}
	respond(w, r, http.StatusOK, resp)
}

// POST /api/auth/login-code — 验证码登录；手机号未注册时自动注册（与 h5AuthRegister 同一 ID 约定）。
func (s *Server) loginWithSMS(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}
	if err := decode(r, &body); err != nil || !isChinaPhone(body.Phone) || body.Code == "" {
		fail(w, r, http.StatusBadRequest, errBadRequest("手机号和验证码必填"))
		return
	}
	rec, ok := smsCodes.Load(body.Phone)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errBadRequest("请先获取验证码"))
		return
	}
	recv := rec.(smsRecord)
	if time.Now().After(recv.ExpiresAt) || recv.Code != body.Code {
		fail(w, r, http.StatusUnauthorized, errBadRequest("验证码错误或已过期"))
		return
	}
	smsCodes.Delete(body.Phone)

	uid := "user-" + body.Phone
	u, err := s.userRepo.FindByID(uid)
	if err != nil {
		now := time.Now()
		u = domain.User{
			ID:           uid,
			WechatOpenID: "phone:" + body.Phone, // 非微信用户唯一 openid，避免 UNIQUE 冲突
			Role:         domain.RoleIndividual,
			Status:       "active",
			Version:      1,
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		if _, err := s.userRepo.Create(u); err != nil {
			fail(w, r, http.StatusInternalServerError, fmt.Errorf("create user: %w", err))
			return
		}
	}
	role := u.Role
	if role == "" {
		role = domain.RoleIndividual
	}

	accessToken, err := s.tokens.IssueJWT(domain.Actor{ID: u.ID, Role: role}, 15*time.Minute)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	refreshToken, err := service.GenerateRefreshToken()
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	tokenHash := service.HashToken(refreshToken)
	if err := s.refreshRepo.Store(u.ID, tokenHash, time.Now().Add(7*24*time.Hour)); err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}

	s.audit(r.Context(), u.ID, "login_sms", "auth", u.ID, "success")
	respond(w, r, http.StatusOK, map[string]any{
		"success":      true,
		"user":         map[string]any{"id": u.ID, "phone": body.Phone, "role": string(role), "status": u.Status},
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
	})
}
