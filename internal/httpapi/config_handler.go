package httpapi

import (
	"errors"
	"net/http"

	"drone-platform/internal/config"
	"drone-platform/internal/domain"
)

// GET /api/v1/admin/config
// 平台级配置（banner/服务配置等影响全体用户），仅平台管理员可读写。
func (s *Server) getConfig(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok || a.Role != domain.RolePlatformAdmin {
		fail(w, r, http.StatusForbidden, errors.New("platform admin permission required"))
		return
	}
	respond(w, r, http.StatusOK, config.GetPlatformConfig())
}

// POST /api/v1/admin/config
func (s *Server) updateConfig(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok || a.Role != domain.RolePlatformAdmin {
		fail(w, r, http.StatusForbidden, errors.New("platform admin permission required"))
		return
	}
	var cfg config.PlatformConfig
	if err := decode(r, &cfg); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	if err := config.SavePlatformConfig(cfg); err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	s.audit(r.Context(), a.ID, "update_platform_config", "config", "platform", "saved")
	respond(w, r, http.StatusOK, map[string]string{"status": "saved"})
}
