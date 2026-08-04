package httpapi

import (
	"errors"
	"net/http"
	"strconv"

	"drone-platform/internal/domain"
)

func (s *Server) registerBatch2Routes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/admin/transformations", s.createTransformation)
	mux.HandleFunc("GET /api/v1/transformations", s.listTransformations)
	mux.HandleFunc("POST /api/v1/transformations/{id}/advance", s.advanceStage)
	mux.HandleFunc("POST /api/v1/transformations/{id}/milestones", s.addMilestone)

	mux.HandleFunc("POST /api/v1/admin/colleges", s.createCollege)
	mux.HandleFunc("GET /api/v1/colleges", s.listColleges)

	mux.HandleFunc("POST /api/v1/cooperation-programs", s.createCooperation)
	mux.HandleFunc("GET /api/v1/cooperation-programs", s.listCooperations)
	mux.HandleFunc("POST /api/v1/cooperation-programs/{id}/status", s.updateCooperationStatus)
}

// ── Transformation ──
func (s *Server) createTransformation(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("auth required"))
		return
	}
	var in struct {
		Title         string `json:"title"`
		AchievementID string `json:"achievement_id"`
		PartnerID     string `json:"partner_id"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	t, err := s.transSvc.Create(in.Title, in.AchievementID, a.ID, in.PartnerID)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusCreated, t)
}
func (s *Server) listTransformations(w http.ResponseWriter, r *http.Request) {
	// 公开查询：按成果(achievement_id)或归属(owner_id)过滤，无参返回全部
	var (
		list []domain.Transformation
		err  error
	)
	if aid := r.URL.Query().Get("achievement_id"); aid != "" {
		list, err = s.transSvc.ListByAchievement(aid)
	} else {
		list, err = s.transSvc.List(r.URL.Query().Get("owner_id"))
	}
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	if list == nil {
		list = []domain.Transformation{}
	}
	respond(w, r, http.StatusOK, list)
}
func (s *Server) advanceStage(w http.ResponseWriter, r *http.Request) {
	var in struct{ Stage, Progress string }
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	t, err := s.transSvc.AdvanceStage(r.PathValue("id"), domain.TransformationStage(in.Stage), in.Progress)
	if err != nil {
		fail(w, r, http.StatusNotFound, err)
		return
	}
	respond(w, r, http.StatusOK, t)
}
func (s *Server) addMilestone(w http.ResponseWriter, r *http.Request) {
	var in struct{ Name, Evidence string }
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	t, err := s.transSvc.AddMilestone(r.PathValue("id"), in.Name, in.Evidence)
	if err != nil {
		fail(w, r, http.StatusNotFound, err)
		return
	}
	respond(w, r, http.StatusOK, t)
}

// ── College ──
func (s *Server) createCollege(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Name, Region, Description, LogoURL string
		CoopType                           string `json:"coop_type"` // research/talent/both
		Majors, Facilities                 []string
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	c, err := s.collegeSvc.Create(in.Name, in.Region, in.Description, in.LogoURL, in.CoopType, in.Majors, in.Facilities)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusCreated, c)
}
func (s *Server) listColleges(w http.ResponseWriter, r *http.Request) {
	list, err := s.collegeSvc.List(r.URL.Query().Get("region"))
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, list)
}

// ── Cooperation ──
func (s *Server) createCooperation(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Title, CollegeID, EnterpriseID, CoopType, Description, StartDate, EndDate string
		StudentQuota                                                              int
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	sd := domain.ParseTime(in.StartDate)
	ed := domain.ParseTime(in.EndDate)
	cp, err := s.coopSvc.Create(in.Title, in.CollegeID, in.EnterpriseID, in.CoopType, in.Description, sd, ed, in.StudentQuota)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusCreated, cp)
}
func (s *Server) listCooperations(w http.ResponseWriter, r *http.Request) {
	list, err := s.coopSvc.List(r.URL.Query().Get("enterprise_id"))
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, list)
}
func (s *Server) updateCooperationStatus(w http.ResponseWriter, r *http.Request) {
	_, _ = strconv.Atoi(r.URL.Query().Get("dummy")) // suppress unused import
	var in struct{ Status string }
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	cp, err := s.coopSvc.UpdateStatus(r.PathValue("id"), in.Status)
	if err != nil {
		fail(w, r, http.StatusNotFound, err)
		return
	}
	respond(w, r, http.StatusOK, cp)
}
