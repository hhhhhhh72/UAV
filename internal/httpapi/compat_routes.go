package httpapi

import (
	"errors"
	"net/http"
)

func (s *Server) registerCompatRoutes(mux *http.ServeMux) {
	// Only non-overlapping routes not covered by registerH5Compat.
	mux.HandleFunc("POST /api/auth/wechat/login", s.wechatLogin)
	mux.HandleFunc("POST /api/auth/wx-login", s.wechatLogin)
	mux.HandleFunc("POST /api/auth/wx-phone", s.wxPhone)
}

// C9 修复说明：此处原有无注册的 passwordLogin / passwordRegister / getMeLegacy
// 死代码——passwordLogin 不校验密码即签发 token 且 refresh 明文入库，
// 一旦被重新挂载即成安全事故，已整体删除。
// 生产密码登录走 h5_compat.go 的 h5AuthLogin（bcrypt 校验，已注册）。

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
