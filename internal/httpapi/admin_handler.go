package httpapi

import (
	"errors"
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
	// 固定 ID：dev 影子管理员不落 users 表，若每次登录换 ID，
	// 站内消息按 receiver 隔离，历史消息将永远不可见（消息互通靠固定 ID）
	uid := "admin-dev"
	actor := domain.Actor{ID: uid, Role: domain.Role(req.Role)}
	token, err := s.tokens.IssueJWT(actor, 2*time.Hour)
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
