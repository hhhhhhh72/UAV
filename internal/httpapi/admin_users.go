package httpapi

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"drone-platform/internal/domain"
)

// GET /api/v1/admin/users — list all users (admin only).
func (s *Server) listUsers(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok || (a.Role != domain.RolePlatformAdmin && a.Role != domain.RoleAssociationAdmin) {
		fail(w, r, http.StatusForbidden, errors.New("admin permission required"))
		return
	}
	roleLabel := func(s string) string {
		switch s {
		case "platform_admin": return "平台管理员"
		case "association_admin": return "协会管理员"
		case "enterprise": return "企业"
		default: return "个人"
		}
	}
	out := []map[string]string{{"id": "admin", "role": "platform_admin", "status": "active", "roleLabel": "平台管理员"}}
	users, err := s.userRepo.All()
	if err == nil {
		for _, u := range users {
			out = append(out, map[string]string{
				"id": u.ID, "role": "individual", "status": u.Status,
				"roleLabel": roleLabel("individual"),
			})
		}
	}
	respond(w, r, http.StatusOK, out)
}

// POST /api/v1/admin/users — create a new user.
func (s *Server) createUser(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok || (a.Role != domain.RolePlatformAdmin && a.Role != domain.RoleAssociationAdmin) {
		fail(w, r, http.StatusForbidden, errors.New("admin permission required"))
		return
	}
	var req struct {
		ID   string `json:"id"`
		Role string `json:"role"`
	}
	if err := decode(r, &req); err != nil || req.ID == "" {
		fail(w, r, http.StatusBadRequest, errors.New("user id required"))
		return
	}
	if req.Role == "" {
		req.Role = "individual"
	}
	now := time.Now()
	u := domain.User{
		ID:        req.ID,
		Status:    req.Role,
		Version:   1,
		CreatedAt: now,
		UpdatedAt: now,
	}
	_, err := s.userRepo.Create(u)
	if err != nil {
		fail(w, r, http.StatusConflict, fmt.Errorf("user '%s' already exists or create failed", req.ID))
		return
	}
	respond(w, r, http.StatusCreated, map[string]string{"id": req.ID, "role": req.Role, "status": "created"})
}

// POST /api/v1/admin/users/{id}/role — change user role.
func (s *Server) updateUserRole(w http.ResponseWriter, r *http.Request) {
	act, ok := authenticatedActor(r)
	if !ok || act.Role != domain.RolePlatformAdmin {
		fail(w, r, http.StatusForbidden, errors.New("only platform admin can change roles"))
		return
	}
	var req struct{ Role string `json:"role"` }
	if err := decode(r, &req); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	allowed := map[string]bool{"individual": true, "enterprise": true, "association_admin": true, "platform_admin": true}
	if !allowed[req.Role] {
		fail(w, r, http.StatusBadRequest, errors.New("invalid role"))
		return
	}
	if err := s.userRepo.UpdateRole(r.PathValue("id"), domain.Role(req.Role)); err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, map[string]string{"status": "updated", "role": req.Role})
}
