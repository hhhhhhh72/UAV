package httpapi

import "net/http"

// registerCompatRoutes maps old frontend /api/* paths to our /api/v1/* handlers.
func (s *Server) registerCompatRoutes(mux *http.ServeMux) {
	// Auth (old → new)
	mux.HandleFunc("POST /api/auth/login", s.wechatLogin)        // → /api/v1/auth/wechat/login
	mux.HandleFunc("POST /api/auth/wechat/login", s.wechatLogin)  // same handler
	mux.HandleFunc("POST /api/auth/wx-login", s.wechatLogin)
	mux.HandleFunc("POST /api/auth/register", s.wechatLogin)      // 简化:注册走微信登录
	mux.HandleFunc("GET /api/auth/me", s.me)                       // → /api/v1/me
	mux.HandleFunc("POST /api/auth/refresh", s.refreshToken)
	mux.HandleFunc("POST /api/auth/logout", s.logout)

	// Cases (old → new)
	mux.HandleFunc("GET /api/cases", s.listCases)

	// Admin (old → new)
	mux.HandleFunc("GET /api/users", s.listUsers)                 // → /api/v1/admin/users
	mux.HandleFunc("POST /api/user/role", s.updateUserRole)       // → /api/v1/admin/users/{id}/role
	mux.HandleFunc("GET /api/list", s.listAllAdapter)             // 通用列表适配
	mux.HandleFunc("POST /api/submit", s.submitAdapter)            // 通用提交适配
	mux.HandleFunc("POST /api/update", s.updateAdapter)            // 通用更新适配

	// Config (old → new)
	mux.HandleFunc("GET /api/services/config", s.getConfig)
}

// listAllAdapter handles old /api/list?role=admin /api/list?userId=X
func (s *Server) listAllAdapter(w http.ResponseWriter, r *http.Request) {
	role := r.URL.Query().Get("role")
	userID := r.URL.Query().Get("userId")
	if role == "admin" {
		// Admin dashboard redirect
		s.adminDashboard(w, r)
		return
	}
	if userID != "" {
		// User's applications
		r.SetPathValue("dummy", "")
		s.listMyProjectApps(w, r)
		return
	}
	respond(w, r, http.StatusOK, []any{})
}

// submitAdapter handles old /api/submit → create demand or enterprise
func (s *Server) submitAdapter(w http.ResponseWriter, r *http.Request) {
	s.createDemand(w, r)
}

// updateAdapter handles old /api/update
func (s *Server) updateAdapter(w http.ResponseWriter, r *http.Request) {
	respond(w, r, http.StatusOK, map[string]string{"status": "ok"})
}
