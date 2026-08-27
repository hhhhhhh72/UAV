package httpapi

import (
	"errors"
	"net/http"
)

// ── 低空研学报名（闭环：报名 → 我的报名 → 管理端审核） ──

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
	e, err := s.studyEnrollSvc.Review(r.Context(), r.PathValue("id"), in.Status)
	if err != nil {
		fail(w, r, http.StatusConflict, err)
		return
	}
	s.audit(r.Context(), "admin", "review_study_tour_enrollment", "study_tour_enrollment", e.ID, in.Status)
	respond(w, r, http.StatusOK, e)
}
