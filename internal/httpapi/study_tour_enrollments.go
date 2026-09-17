package httpapi

import (
	"context"
	"errors"
	"net/http"
	"strings"
)

// ── 低空研学报名（闭环：报名 → 我的报名 → 管理端审核） ──

// studyTourLabel 研学标题的展示形态：《标题》。取不到标题时退回通称「研学活动」，
// 避免通知里出现空的《》。顺带把主办方 ID 带出来，用于机构侧的待办通知。
//
// 研学线与培训报名是同一个问题：此前报名成功与审核结果**全程静默**，
// 学员只能自己在「我的报名-研学」里反复刷新。
func (s *Server) studyTourLabel(ctx context.Context, tourID string) (label, organizerID string) {
	label = "研学活动"
	if s.studyTourRepo == nil {
		return label, ""
	}
	t, err := s.studyTourRepo.FindByID(ctx, tourID)
	if err != nil {
		return label, ""
	}
	if strings.TrimSpace(t.Title) != "" {
		label = "《" + t.Title + "》"
	}
	return label, t.OrganizerID
}

// POST /api/v1/study/tours/{id}/enroll — 研学报名
func (s *Server) createStudyTourEnrollment(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("auth required"))
		return
	}
	var in struct {
		Name       string `json:"name"`
		Phone      string `json:"phone"`
		AdultCount int    `json:"adult_count"`
		ChildCount int    `json:"child_count"`
		Remark     string `json:"remark"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	e, err := s.studyEnrollSvc.Create(r.Context(), a.ID, r.PathValue("id"), in.Name, in.Phone, in.AdultCount, in.ChildCount, in.Remark)
	if err != nil {
		fail(w, r, http.StatusConflict, err)
		return
	}
	s.audit(r.Context(), a.ID, "study_tour_enroll", "study_tour_enrollment", e.ID, "created")
	// 报名成功即通知双方（取不到研学只是少一条通知，不影响报名结果）
	label, organizerID := s.studyTourLabel(r.Context(), e.TourID)
	s.notify(a.ID, "研学报名成功",
		"您已提交"+label+"的报名，等待工作人员确认。", "study_tour_enrollment", e.TourID)
	s.notify(organizerID, "新的研学报名待审核",
		label+"收到一条新的报名，审核通过后学员即可参加。", "study_tour_enrollment", e.TourID)
	respond(w, r, http.StatusCreated, e)
}

// GET /api/v1/study-tours/enrollments/mine — 我的研学报名
func (s *Server) listMyStudyTourEnrollments(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("auth required"))
		return
	}
	items, err := s.studyEnrollSvc.ListMyEnrollments(r.Context(), a.ID)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, items)
}

// GET /api/v1/admin/study-tours/{id}/enrollments — 管理端报名列表
func (s *Server) listStudyTourEnrollments(w http.ResponseWriter, r *http.Request) {
	items, err := s.studyEnrollSvc.ListByTour(r.Context(), r.PathValue("id"))
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, items)
}

// POST /api/v1/admin/study-tours/enrollments/{id}/review — 管理端审核
func (s *Server) reviewStudyTourEnrollment(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Status string `json:"status"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	// 先读原状态：Review 对「同状态重复审核」是幂等的（cur.Status == status 时直接回原行），
	// 此时 e.Status 与之前相同，不该再发一条一模一样的通知去打扰学员。
	before, beforeErr := s.studyEnrollSvc.Get(r.Context(), r.PathValue("id"))
	e, err := s.studyEnrollSvc.Review(r.Context(), r.PathValue("id"), in.Status)
	if err != nil {
		fail(w, r, http.StatusConflict, err)
		return
	}
	s.audit(r.Context(), "admin", "review_study_tour_enrollment", "study_tour_enrollment", e.ID, in.Status)
	// 只在状态真的变了时通知（幂等重审不重复打扰）
	if beforeErr != nil || before.Status != e.Status {
		label, _ := s.studyTourLabel(r.Context(), e.TourID)
		switch e.Status {
		case "approved":
			s.notify(e.UserID, "研学报名审核结果",
				"恭喜！"+label+"的报名已通过审核，请留意活动安排。", "study_tour_enrollment", e.TourID)
		case "rejected":
			s.notify(e.UserID, "研学报名审核结果",
				"很遗憾，"+label+"的报名未通过审核。", "study_tour_enrollment", e.TourID)
		case "completed":
			s.notify(e.UserID, "研学活动结业",
				"您参加的"+label+"已完成，感谢参与。", "study_tour_enrollment", e.TourID)
		}
	}
	respond(w, r, http.StatusOK, e)
}
