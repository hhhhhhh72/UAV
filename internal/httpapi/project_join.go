package httpapi

import (
	"errors"
	"net/http"

	"drone-platform/internal/domain"
	"drone-platform/internal/service"
)

// POST /api/v1/projects/{id}/join
// 申请参与课题攻关：需登录；重复申请幂等返回；课题不存在 404。
func (s *Server) joinResearchProject(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	var in struct {
		OrgName string `json:"org_name"`
		Message string `json:"message"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	req, created, err := s.researchSvc.JoinProject(r.Context(), a.ID, r.PathValue("id"), in.OrgName, in.Message)
	if err != nil {
		if errors.Is(err, service.ErrProjectNotFound) {
			fail(w, r, http.StatusNotFound, err)
			return
		}
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	if created {
		s.audit(r.Context(), a.ID, "join_research_project", "project_join_request", req.ID, "created")
		respond(w, r, http.StatusCreated, req)
		return
	}
	respond(w, r, http.StatusOK, req)
}

// GET /api/v1/projects/{id}/join/mine
// 当前用户在该课题下的申请状态（未申请过返回 applied=false）。
func (s *Server) getMyProjectJoin(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	req, applied, err := s.researchSvc.GetMyJoin(r.Context(), r.PathValue("id"), a.ID)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, map[string]any{"applied": applied, "join": req})
}

// GET /api/v1/admin/projects/{id}/joins
// 后台查看课题全部参与申请（协会/平台管理员）。
func (s *Server) listProjectJoins(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok || (a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin) {
		fail(w, r, http.StatusForbidden, errors.New("admin permission required"))
		return
	}
	list, err := s.researchSvc.ListJoins(r.Context(), r.PathValue("id"))
	if err != nil {
		if errors.Is(err, service.ErrProjectNotFound) {
			fail(w, r, http.StatusNotFound, err)
			return
		}
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, map[string]any{"items": list, "total": len(list)})
}

// POST /api/v1/admin/projects/{id}/joins/{joinID}/status
// 后台流转申请状态：pending 待评估 / contacted 已对接 / closed 已关闭。
func (s *Server) updateProjectJoinStatus(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok || (a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin) {
		fail(w, r, http.StatusForbidden, errors.New("admin permission required"))
		return
	}
	var in struct {
		Status string `json:"status"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	req, err := s.researchSvc.UpdateJoinStatus(r.Context(), r.PathValue("joinID"), in.Status)
	if err != nil {
		if errors.Is(err, service.ErrJoinStatusInvalid) {
			fail(w, r, http.StatusBadRequest, err)
			return
		}
		fail(w, r, http.StatusNotFound, err)
		return
	}
	s.audit(r.Context(), a.ID, "update_join_status", "project_join_request", req.ID, in.Status)
	respond(w, r, http.StatusOK, req)
}
