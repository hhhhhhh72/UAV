package httpapi

import (
	"crypto/rand"
	"crypto/subtle"
	"errors"
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
	mu sync.Mutex // 串行化同号并发校验的读-改-写，防错误计数丢更新（爆破防护）
	Code      string
	ExpiresAt time.Time
	// Attempts 记录错误校验次数：达到 maxSMSCodeAttempts 即作废验证码，
	// 防止 6 位码被在线爆破（此前无尝试限制）。
	Attempts int
}

var smsCodes sync.Map // phone -> *smsRecord（指针 + 条目锁，并发安全递增）

// startSMSCodesJanitor 启动验证码条目周期清理（进程内只启动一次）：
// 未使用/未删除的验证码条目（用户获取后从不校验）会永久留存，内存无限增长。
func startSMSCodesJanitor() {
	smsCodesJanitorOnce.Do(func() {
		go smsCodesCleanupLoop()
	})
}

var smsCodesJanitorOnce sync.Once

// smsCodesCleanupLoop 每分钟遍历 smsCodes，删除已过期的验证码条目。
// 无退出通道，与 internal/cache 的 cleanupLoop 模式一致；panic 由 recover 兜底。
func smsCodesCleanupLoop() {
	defer func() { _ = recover() }()
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		func() {
			defer func() { _ = recover() }() // 后台协程 panic 防护
			now := time.Now()
			smsCodes.Range(func(k, v any) bool {
				rec := v.(*smsRecord)
				// ExpiresAt 仅在 Store 前写入、之后不再修改，无需加锁读取。
				if now.After(rec.ExpiresAt) {
					smsCodes.Delete(k)
				}
				return true
			})
		}()
	}
}

const (
	smsCodeTTL         = 5 * time.Minute
	smsResendWait      = 60 * time.Second
	maxSMSCodeAttempts = 5
	// smsIPMaxPerMinute 同 IP 每分钟发送上限：60s 重发限制只按手机号，
	// 攻击者可换号轰炸（接入真实短信商后有成本风险），须按 IP 再加一道闸。
	smsIPMaxPerMinute = 5
	// smsIPLimitMaxEntries 限频表上限（防内存 DoS），超限清空重建（粗糙但有效）。
	smsIPLimitMaxEntries = 10000
)

// smsIPLog 记录某 IP 在窗口内的发送次数。
type smsIPLog struct {
	mu          sync.Mutex
	count       int
	windowStart time.Time
}

// smsIPAllowed 报告该 IP 是否仍允许发送短信，并累计本次发送。
// 限频表挂在 Server 实例上（单实例足够，测试实例互不干扰）。
// 表大小上限用惰性清理：条目过多时随机清空（sync.Map 无 Len，用计数近似）。
func (s *Server) smsIPAllowed(ip string) bool {
	if s.smsIPEntries.Load() >= smsIPLimitMaxEntries {
		s.smsIPLimits.Range(func(k, _ any) bool { s.smsIPLimits.Delete(k); return true })
		s.smsIPEntries.Store(int64(0))
	}
	s.smsIPEntries.Add(1)
	v, _ := s.smsIPLimits.LoadOrStore(ip, &smsIPLog{windowStart: time.Now()})
	log := v.(*smsIPLog)
	log.mu.Lock()
	defer log.mu.Unlock()
	now := time.Now()
	if now.Sub(log.windowStart) >= time.Minute {
		log.windowStart = now
		log.count = 0
	}
	log.count++
	return log.count <= smsIPMaxPerMinute
}

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
	// IP 限频（防换号轰炸）：同 IP 每分钟最多 smsIPMaxPerMinute 次。
	if !s.smsIPAllowed(clientIP(r)) {
		fail(w, r, http.StatusTooManyRequests, errBadRequest("发送过于频繁，请稍后再试"))
		return
	}
	if rec, ok := smsCodes.Load(body.Phone); ok {
		if time.Until(rec.(*smsRecord).ExpiresAt) > smsCodeTTL-smsResendWait {
			fail(w, r, http.StatusTooManyRequests, errBadRequest("验证码已发送，请 60 秒后再试"))
			return
		}
	}
	code, err := genSMSCode()
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	startSMSCodesJanitor() // 首次发送时启动过期条目清理（sync.Once 幂等）
	smsCodes.Store(body.Phone, &smsRecord{Code: code, ExpiresAt: time.Now().Add(smsCodeTTL)})

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
	recv := rec.(*smsRecord)
	recv.mu.Lock()
	defer recv.mu.Unlock()
	if time.Now().After(recv.ExpiresAt) {
		smsCodes.Delete(body.Phone)
		fail(w, r, http.StatusUnauthorized, errBadRequest("验证码错误或已过期"))
		return
	}
	// 常量时间比较防时序侧信道；错误累计 5 次后作废验证码，须重新获取。
	// 锁内递增：并发请求串行化，杜绝计数丢更新绕过爆破防护。
	if subtle.ConstantTimeCompare([]byte(recv.Code), []byte(body.Code)) != 1 {
		recv.Attempts++
		if recv.Attempts >= maxSMSCodeAttempts {
			smsCodes.Delete(body.Phone)
			fail(w, r, http.StatusUnauthorized, errBadRequest("验证码错误次数过多，请重新获取验证码"))
			return
		}
		fail(w, r, http.StatusUnauthorized, errBadRequest("验证码错误或已过期"))
		return
	}
	smsCodes.Delete(body.Phone)

	uid := "user-" + body.Phone
	u, err := s.userRepo.FindByID(r.Context(), uid)
	if err != nil {
		// P2 修复：同手机号并发首登（连点登录）会双双 miss 再各 Create 一个用户。
		// 按 uid 加进程内锁，锁内复查一次再创建；后到者复用先建好的用户。
		unlock := loginLockByKey("uid|" + uid)
		u, findErr := s.userRepo.FindByID(r.Context(), uid)
		if findErr != nil {
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
			_, findErr = s.userRepo.Create(r.Context(), u)
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
	}
	if u.Status != "" && u.Status != "active" {
		fail(w, r, http.StatusUnauthorized, errors.New("账号已停用"))
		return
	}
	accessToken, err := s.tokens.IssueJWT(domain.Actor{ID: u.ID, Role: role, TokenVersion: u.TokenVersion}, 15*time.Minute)
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
	if err := s.refreshRepo.Store(r.Context(), u.ID, tokenHash, time.Now().Add(7*24*time.Hour)); err != nil {
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
