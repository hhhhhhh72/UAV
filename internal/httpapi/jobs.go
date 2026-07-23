package httpapi

import (
	"errors"
	"net/http"

	"drone-platform/internal/domain"
)

// ---- Jobs ----

// POST /api/v1/jobs
func (s *Server) createJob(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
	var in struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Location    string `json:"location"`
		SalaryFen   int64  `json:"salary_fen"`
	}
	if err := decode(r, &in); err != nil { fail(w, r, http.StatusBadRequest, err); return }
	j, err := s.jobSvc.CreateJob(a, in.Title, in.Description, in.Location, in.SalaryFen)
	if err != nil { fail(w, r, http.StatusForbidden, err); return }
	s.audit(r.Context(), a.ID, "create_job", "job", j.ID, "created")
	respond(w, r, http.StatusCreated, j)
}

// POST /api/v1/jobs/{id}/publish
func (s *Server) publishJob(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
	j, err := s.jobSvc.PublishJob(a, r.PathValue("id"))
	if err != nil { fail(w, r, http.StatusForbidden, err); return }
	respond(w, r, http.StatusOK, j)
}

// POST /api/v1/jobs/{id}/close
func (s *Server) closeJob(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
	j, err := s.jobSvc.CloseJob(a, r.PathValue("id"))
	if err != nil { fail(w, r, http.StatusForbidden, err); return }
	respond(w, r, http.StatusOK, j)
}

// GET /api/v1/jobs
func (s *Server) listJobs(w http.ResponseWriter, r *http.Request) {
	page, pageSize := paginationFromQuery(r)
	offset := (page - 1) * pageSize
	items, total, err := s.jobSvc.ListPublishedJobs(offset, pageSize)
	if err != nil { fail(w, r, http.StatusInternalServerError, err); return }
	paginatedRespond(w, r, items, total)
}

// GET /api/v1/jobs/mine
func (s *Server) listMyJobs(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
	items, err := s.jobSvc.ListMyJobs(a)
	if err != nil { fail(w, r, http.StatusForbidden, err); return }
	respond(w, r, http.StatusOK, items)
}

// ---- Resumes ----

// POST /api/v1/resumes
func (s *Server) createResume(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
	var in struct {
		Title      string `json:"title"`
		Content    string `json:"content"`
		Visibility string `json:"visibility"`
	}
	if err := decode(r, &in); err != nil { fail(w, r, http.StatusBadRequest, err); return }
	res, err := s.jobSvc.CreateResume(a, in.Title, in.Content, in.Visibility)
	if err != nil { fail(w, r, http.StatusForbidden, err); return }
	respond(w, r, http.StatusCreated, res)
}

// PATCH /api/v1/resumes/{id}
func (s *Server) updateResume(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
	var in struct {
		Title      string `json:"title"`
		Content    string `json:"content"`
		Visibility string `json:"visibility"`
	}
	if err := decode(r, &in); err != nil { fail(w, r, http.StatusBadRequest, err); return }
	res, err := s.jobSvc.UpdateResume(a, r.PathValue("id"), in.Title, in.Content, in.Visibility)
	if err != nil { fail(w, r, http.StatusForbidden, err); return }
	respond(w, r, http.StatusOK, res)
}

// GET /api/v1/resumes/mine
func (s *Server) listMyResumes(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
	items, err := s.jobSvc.ListMyResumes(a)
	if err != nil { fail(w, r, http.StatusInternalServerError, err); return }
	respond(w, r, http.StatusOK, items)
}

// ---- Applications ----

// POST /api/v1/applications
func (s *Server) createApplication(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
	var in struct {
		JobID    string `json:"job_id"`
		ResumeID string `json:"resume_id"`
	}
	if err := decode(r, &in); err != nil { fail(w, r, http.StatusBadRequest, err); return }
	app, err := s.jobSvc.Apply(a, in.JobID, in.ResumeID)
	if err != nil { fail(w, r, http.StatusForbidden, err); return }
	respond(w, r, http.StatusCreated, app)
}

// PATCH /api/v1/applications/{id}/status
func (s *Server) updateApplicationStatus(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
	var in struct{ Status string `json:"status"` }
	if err := decode(r, &in); err != nil { fail(w, r, http.StatusBadRequest, err); return }
	st := domain.AppStatus(in.Status)
	switch st {
	case domain.AppViewed, domain.AppInterviewing, domain.AppOffered, domain.AppRejected, domain.AppWithdrawn:
	default:
		fail(w, r, http.StatusBadRequest, errors.New("invalid status"))
		return
	}
	app, err := s.jobSvc.UpdateApplicationStatus(a, r.PathValue("id"), st)
	if err != nil { fail(w, r, http.StatusForbidden, err); return }
	respond(w, r, http.StatusOK, app)
}

// GET /api/v1/applications?job_id=xxx
func (s *Server) listApplications(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
	jobID := r.URL.Query().Get("job_id")
	if jobID == "" {
		items, err := s.jobSvc.ListMyApplications(a)
		if err != nil { fail(w, r, http.StatusInternalServerError, err); return }
		respond(w, r, http.StatusOK, items)
		return
	}
	items, err := s.jobSvc.ListApplicationsForJob(a, jobID)
	if err != nil { fail(w, r, http.StatusForbidden, err); return }
	respond(w, r, http.StatusOK, items)
}

