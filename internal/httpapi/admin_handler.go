package httpapi

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"drone-platform/internal/domain"
)

//go:embed admin.html
var adminHTML []byte

// adminDevMode returns true when the admin panel is enabled (dev/test only).
func adminDevMode() bool { return os.Getenv("ADMIN_DEV_MODE") == "true" }

func (s *Server) adminPage(w http.ResponseWriter, r *http.Request) {
	if !adminDevMode() {
		fail(w, r, http.StatusForbidden, errors.New("admin panel disabled in production"))
		return
	}
	// Generate a fresh admin token on every page load.
	now := time.Now()
	uid := fmt.Sprintf("admin-%d", now.UnixMilli())
	actor := domain.Actor{ID: uid, Role: domain.RolePlatformAdmin}
	token, err := s.tokens.Issue(actor, 2*time.Hour)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	// Inject token into the HTML before serving.
	html := bytes.Replace(adminHTML, []byte("__ADMIN_TOKEN__"), []byte(token), 1)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	w.Write(html)
}

// POST /api/v1/admin/token — dev login, issues an admin token without WeChat.
// Protected by ADMIN_DEV_MODE: must NOT be enabled in production.
func (s *Server) adminDevLogin(w http.ResponseWriter, r *http.Request) {
	if !adminDevMode() {
		fail(w, r, http.StatusForbidden, errors.New("dev token disabled in production"))
		return
	}
	var req struct{ Role string `json:"role"` }
	if err := decode(r, &req); err != nil {
		req.Role = "platform_admin"
	}
	if req.Role != "platform_admin" && req.Role != "association_admin" && req.Role != "enterprise" && req.Role != "individual" {
		req.Role = "platform_admin"
	}
	now := time.Now()
	uid := fmt.Sprintf("admin-%d", now.UnixMilli())
	actor := domain.Actor{ID: uid, Role: domain.Role(req.Role)}
	token, err := s.tokens.Issue(actor, 2*time.Hour)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, map[string]any{
		"access_token": token,
		"expires_in":   7200,
		"user":         map[string]any{"id": uid, "role": req.Role},
	})
}
