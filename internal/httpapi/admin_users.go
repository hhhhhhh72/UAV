package httpapi

import (
	"errors"
	"net/http"
	"regexp"

	"drone-platform/internal/crypto"
	"drone-platform/internal/domain"
	"drone-platform/internal/service"
)

// listPhoneIDRe 匹配"登录名=手机号"约定下由系统生成的账号 ID（user-1xxxxxxxxxx）。
var listPhoneIDRe = regexp.MustCompile(`^user-(1[3-9][0-9]{9})$`)

// listPhoneMasked 列表回显的手机号（脱敏）。
//
// 优先取库里的手机号（PhoneCipher 在读库时已解密为明文）；早于"登录名=手机号"约定
// 建立的账号（后台旧表单让运营手填用户 ID，如 user-18623249541）没有手机号密文，
// 但 ID 里就是手机号——直接按 ID 回显，免得列表显示"未绑定"而实际能用该号码登录。
func listPhoneMasked(u domain.User) string {
	if u.PhoneCipher != "" {
		return crypto.MaskPhone(u.PhoneCipher)
	}
	if m := listPhoneIDRe.FindStringSubmatch(u.ID); m != nil {
		return crypto.MaskPhone(m[1])
	}
	return ""
}

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
				"phone_masked": listPhoneMasked(u),
				"deleted_at":   deletedAt,
				"purge_after":  purgeAfter,
			})
		}
	}
	// paginatedRespond 内部会按 query 的 page/page_size 自动切片，此处传全量
	paginatedRespond(w, r, out, len(out))
}

// POST /api/v1/admin/users — 管理员建号。
//
// 以**手机号作为登录名**（与用户端注册同一套约定）：id 由系统生成 user-<手机号>，
// 不再让运营自己填 ID；密码必填（≥8 位）；填手机号同时写加密手机号与占位 openid，
// 使账号能手机号登录/找回。策略与权限判定在 service.UserService.CreateUser。
func (s *Server) createUser(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	var req struct {
		Phone    string `json:"phone"`    // 登录名（必填，11 位手机号）
		Name     string `json:"name"`     // 昵称（可选，缺省按手机号后四位生成）
		Role     string `json:"role"`     // 缺省 individual
		Password string `json:"password"` // 初始密码（必填，≥8 位）
	}
	if err := decode(r, &req); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	if req.Role == "" {
		req.Role = string(domain.RoleIndividual)
	}
	u, err := s.userSvc.CreateUser(r.Context(), a, req.Phone, encryptPhone(req.Phone), req.Name, domain.Role(req.Role), req.Password)
	switch {
	case err == nil:
		s.audit(r.Context(), a.ID, "create_user", "user", u.ID, string(u.Role))
		respond(w, r, http.StatusCreated, map[string]any{
			"id": u.ID, "role": string(u.Role), "status": "created",
			"name": u.Name, "phone_masked": crypto.MaskPhone(req.Phone),
			"note": "登录名就是手机号；已设置初始密码，请把账号与密码告知本人并提醒其尽快修改",
		})
	case errors.Is(err, service.ErrAdminOnly):
		fail(w, r, http.StatusForbidden, err)
	case errors.Is(err, service.ErrInvalidPhone), errors.Is(err, service.ErrInvalidRole), errors.Is(err, service.ErrWeakPassword):
		fail(w, r, http.StatusBadRequest, err)
	case errors.Is(err, service.ErrUserExists):
		fail(w, r, http.StatusConflict, err)
	default:
		fail(w, r, http.StatusInternalServerError, err)
	}
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
