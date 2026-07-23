package httpapi

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"drone-platform/internal/domain"
)

// adminDevMode returns true when the admin panel is enabled (dev/test only).
func adminDevMode() bool { return os.Getenv("ADMIN_DEV_MODE") == "true" }

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
