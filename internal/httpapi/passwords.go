package httpapi

import (
	"errors"
	"net/http"

	"drone-platform/internal/service"
)

// POST /api/v1/auth/password — 本人修改密码（任何已登录账号，含管理员）。
//
// 安全设计：必须提供旧密码；旧密码错误复用登录那套账号级失败计数（连续错多次锁定），
// 防止把改密接口当成在线爆破旧密码的入口；成功后 token_version 自增 + 刷新令牌全撤，
// 所有设备必须重新登录。
func (s *Server) changePassword(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	if s.passwordAccountLocked(a.ID) {
		fail(w, r, http.StatusTooManyRequests, errors.New("密码错误次数过多，请稍后再试"))
		return
	}
	var in struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	err := s.userSvc.ChangePassword(r.Context(), a, in.OldPassword, in.NewPassword)
	switch {
	case err == nil:
		s.clearAccountFailures(a.ID)
		s.audit(r.Context(), a.ID, "change_password", "user", a.ID, "changed")
		respond(w, r, http.StatusOK, map[string]any{
			"changed": true,
			"note":    "密码已修改，请用新密码重新登录（其他设备上的登录已失效）",
		})
	case errors.Is(err, service.ErrOldPasswordWrong):
		s.recordAccountFailure(a.ID) // 与登录同一套账号级失败锁定
		fail(w, r, http.StatusBadRequest, err)
	case errors.Is(err, service.ErrWeakPassword):
		fail(w, r, http.StatusBadRequest, err)
	case errors.Is(err, service.ErrPasswordNotSet):
		fail(w, r, http.StatusConflict, err)
	case errors.Is(err, service.ErrUserNotFound):
		fail(w, r, http.StatusNotFound, err)
	default:
		fail(w, r, http.StatusInternalServerError, err)
	}
}

// POST /api/v1/admin/users/{id}/password — 平台管理员重置他人密码（无需旧密码）。
// 使用场景：本人忘了密码；此前平台只有"新建用户时设一次密码"，之后谁都改不了。
func (s *Server) resetUserPassword(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	var in struct {
		Password string `json:"password"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	id := r.PathValue("id")
	err := s.userSvc.ResetPassword(r.Context(), a, id, in.Password)
	switch {
	case err == nil:
		s.audit(r.Context(), a.ID, "reset_user_password", "user", id, "reset")
		respond(w, r, http.StatusOK, map[string]any{
			"id": id, "reset": true,
			"note": "已重置密码：该账号此前的登录全部失效，请把新密码告知本人",
		})
	case errors.Is(err, service.ErrAdminOnly):
		fail(w, r, http.StatusForbidden, err)
	case errors.Is(err, service.ErrWeakPassword):
		fail(w, r, http.StatusBadRequest, err)
	case errors.Is(err, service.ErrUserNotFound):
		fail(w, r, http.StatusNotFound, err)
	default:
		fail(w, r, http.StatusInternalServerError, err)
	}
}
