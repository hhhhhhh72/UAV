package httpapi

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"

	"drone-platform/internal/domain"
	"drone-platform/internal/service"
)

// GET /api/v1/admin/users — list users with pagination (admin only).
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
	out := []map[string]any{{"id": "admin", "role": "platform_admin", "status": "active", "roleLabel": "平台管理员", "created_at": "—", "has_password": true}}
	// 管理端列表含已注销账号（All 会过滤 deleted_at，注销后管理员就再也看不到、也就无法恢复）
	users, err := s.userRepo.AllWithDeleted(r.Context())
	if err == nil {
		for _, u := range users {
			rl := roleLabel(string(u.Role))
			// 密码状态：仅暴露"是否设置过密码"，绝不返回 hash 本身
			hasPassword := u.PasswordHash != ""
			// 已删除账号在缓冲期内：给出到期自动清除的日期，后台据此提示（注销不可恢复）
			deletedAt, purgeAfter := "", ""
			if u.DeletedAt != nil {
				deletedAt = u.DeletedAt.Format("2006-01-02 15:04")
				purgeAfter = u.DeletedAt.Add(service.UserPurgeRetention).Format("2006-01-02")
			}
			out = append(out, map[string]any{
				"id":           u.ID,
				"role":         string(u.Role),
				"status":       u.Status,
				"roleLabel":    rl,
				"name":         u.Name,
				"avatar_url":   u.AvatarURL,
				"created_at":   u.CreatedAt.Format("2006-01-02 15:04"),
				"has_password": hasPassword,
				"deleted_at":   deletedAt,
				"purge_after":  purgeAfter,
			})
		}
	}
	// paginatedRespond 内部会按 query 的 page/page_size 自动切片，此处传全量
	paginatedRespond(w, r, out, len(out))
}

// POST /api/v1/admin/users — create a new user.
func (s *Server) createUser(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok || (a.Role != domain.RolePlatformAdmin && a.Role != domain.RoleAssociationAdmin) {
		fail(w, r, http.StatusForbidden, errors.New("admin permission required"))
		return
	}
	var req struct {
		ID       string `json:"id"`
		Role     string `json:"role"`
		Password string `json:"password"` // optional: sets a login password (bcrypt)
	}
	if err := decode(r, &req); err != nil || req.ID == "" {
		fail(w, r, http.StatusBadRequest, errors.New("user id required"))
		return
	}
	if req.Role == "" {
		req.Role = "individual"
	}
	// 角色白名单：与 updateUserRole 一致，拒绝任意字符串
	allowed := map[string]bool{"individual": true, "enterprise": true, "association_admin": true, "platform_admin": true}
	if !allowed[req.Role] {
		fail(w, r, http.StatusBadRequest, errors.New("invalid role"))
		return
	}
	// 防提权：协会管理员只能创建 individual/enterprise 账号，
	// 不得创建 association_admin / platform_admin（C1 修复）
	if a.Role == domain.RoleAssociationAdmin && (req.Role == "association_admin" || req.Role == "platform_admin") {
		fail(w, r, http.StatusForbidden, errors.New("association admin cannot create admin accounts"))
		return
	}
	now := time.Now()
	u := domain.User{
		ID:        req.ID,
		Role:      domain.Role(req.Role),
		Status:    "active",
		Version:   1,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if req.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			fail(w, r, http.StatusInternalServerError, fmt.Errorf("hash password: %w", err))
			return
		}
		u.PasswordHash = string(hash)
	}
	_, err := s.userRepo.Create(r.Context(), u)
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
	if err := s.userRepo.UpdateRole(r.Context(), r.PathValue("id"), domain.Role(req.Role)); err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	s.audit(r.Context(), act.ID, "update_user_role", "user", r.PathValue("id"), req.Role)
	respond(w, r, http.StatusOK, map[string]string{"status": "updated", "role": req.Role})
}

// DELETE /api/v1/admin/users/{id} — 删除用户账号（仅平台管理员，唯一动作、不可恢复）。
//
// 语义：账号立即失效（令牌作废 + 角色回收）并从平台消失；账号行保留 7 天缓冲期，
// 到期由后台任务自动物理清除；期间**不可恢复**；其发布的内容一律保留（无外键级联）。
// 策略与权限判定在 service.UserService（Handler 只做参数解析与状态码映射）。
func (s *Server) deleteUser(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	id := r.PathValue("id")
	res, err := s.userSvc.DeleteUser(r.Context(), a, id)
	switch {
	case err == nil:
		s.audit(r.Context(), a.ID, "delete_user", "user", id, res.Mode)
		respond(w, r, http.StatusOK, res)
	case errors.Is(err, service.ErrAdminOnly), errors.Is(err, service.ErrSuperAdminProtected):
		fail(w, r, http.StatusForbidden, err)
	case errors.Is(err, service.ErrUserNotFound):
		fail(w, r, http.StatusNotFound, err)
	default:
		fail(w, r, http.StatusInternalServerError, err)
	}
}
