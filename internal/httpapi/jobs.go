package httpapi

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"drone-platform/internal/domain"
	"drone-platform/internal/service"
)

// ---- Jobs ----

// POST /api/v1/jobs
func (s *Server) createJob(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	var in struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Location    string `json:"location"`
		JobType     string `json:"job_type"`
		SalaryFen   int64  `json:"salary_fen"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	j, err := s.jobSvc.CreateJob(r.Context(), a, in.Title, in.Description, in.Location, in.SalaryFen, in.JobType)
	if err != nil {
		fail(w, r, http.StatusForbidden, err)
		return
	}
	s.audit(r.Context(), a.ID, "create_job", "job", j.ID, "created")
	respond(w, r, http.StatusCreated, j)
}

// POST /api/v1/jobs/{id}/publish
func (s *Server) publishJob(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	j, err := s.jobSvc.PublishJob(r.Context(), a, r.PathValue("id"))
	if err != nil {
		fail(w, r, jobMutationCode(err), err)
		return
	}
	respond(w, r, http.StatusOK, j)
}

// POST /api/v1/jobs/{id}/close
func (s *Server) closeJob(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	j, err := s.jobSvc.CloseJob(r.Context(), a, r.PathValue("id"))
	if err != nil {
		fail(w, r, jobMutationCode(err), err)
		return
	}
	respond(w, r, http.StatusOK, j)
}

// jobMutationCode 区分状态机流转违规（409）与归属/角色违规（403）。
func jobMutationCode(err error) int {
	if errors.Is(err, service.ErrInvalidJobTransition) {
		return http.StatusConflict
	}
	return http.StatusForbidden
}

// GET /api/v1/jobs?q=关键词&type=全职&page=1&page_size=10
func (s *Server) listJobs(w http.ResponseWriter, r *http.Request) {
	// 性能审查：repo 不支持 q/type 过滤，保持全量上限 2000 + 内存过滤；
	// TODO 下沉：JobRepository.ListPublished 增加 q/type 参数后改分页下沉 SQL + respondPage。
	items, _, err := s.jobSvc.ListPublishedJobs(r.Context(), 0, 2000)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	// q：标题/地点包含（大小写不敏感）；type：job_type 精确匹配
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	typ := r.URL.Query().Get("type")
	if q != "" || typ != "" {
		qs := strings.ToLower(q)
		filtered := make([]domain.Job, 0, len(items))
		for _, j := range items {
			if typ != "" && j.JobType != typ {
				continue
			}
			if q != "" && !strings.Contains(strings.ToLower(j.Title), qs) && !strings.Contains(strings.ToLower(j.Location), qs) {
				continue
			}
			filtered = append(filtered, j)
		}
		items = filtered
	}
	paginatedRespond(w, r, items, len(items))
}

// GET /api/v1/jobs/mine
func (s *Server) listMyJobs(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	items, err := s.jobSvc.ListMyJobs(r.Context(), a)
	if err != nil {
		fail(w, r, http.StatusForbidden, err)
		return
	}
	respond(w, r, http.StatusOK, items)
}

// ---- Resumes ----

// POST /api/v1/resumes
func (s *Server) createResume(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	var in struct {
		Title          string   `json:"title"`
		Name           string   `json:"name"`
		Phone          string   `json:"phone"`
		Email          string   `json:"email"`
		Education      string   `json:"education"`
		WorkExperience string   `json:"work_experience"`
		Skills         []string `json:"skills"`
		CertificateURL string   `json:"certificate_url"`
		Content        string   `json:"content"`
		Visibility     string   `json:"visibility"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	res, err := s.jobSvc.CreateResume(r.Context(), a, in.Title, in.Name, in.Phone, in.Email, in.Education, in.WorkExperience, in.Skills, in.CertificateURL, in.Content, in.Visibility)
	if err != nil {
		fail(w, r, http.StatusForbidden, err)
		return
	}
	respond(w, r, http.StatusCreated, res)
}

// PATCH /api/v1/resumes/{id}
func (s *Server) updateResume(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	var in struct {
		Title          string   `json:"title"`
		Name           string   `json:"name"`
		Phone          string   `json:"phone"`
		Email          string   `json:"email"`
		Education      string   `json:"education"`
		WorkExperience string   `json:"work_experience"`
		Skills         []string `json:"skills"`
		CertificateURL string   `json:"certificate_url"`
		Content        string   `json:"content"`
		Visibility     string   `json:"visibility"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	res, err := s.jobSvc.UpdateResume(r.Context(), a, r.PathValue("id"), in.Title, in.Name, in.Phone, in.Email, in.Education, in.WorkExperience, in.Skills, in.CertificateURL, in.Content, in.Visibility)
	if err != nil {
		fail(w, r, http.StatusForbidden, err)
		return
	}
	respond(w, r, http.StatusOK, res)
}

// GET /api/v1/resumes/mine
func (s *Server) listMyResumes(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	items, err := s.jobSvc.ListMyResumes(r.Context(), a)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, items)
}

// ---- Applications ----

// POST /api/v1/applications
func (s *Server) createApplication(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	var in struct {
		JobID    string `json:"job_id"`
		ResumeID string `json:"resume_id"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	app, err := s.jobSvc.Apply(r.Context(), a, in.JobID, in.ResumeID)
	if err != nil {
		fail(w, r, http.StatusForbidden, err)
		return
	}
	respond(w, r, http.StatusCreated, app)
}

// PATCH /api/v1/applications/{id}/status
func (s *Server) updateApplicationStatus(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	var in struct {
		Status string `json:"status"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	st := domain.AppStatus(in.Status)
	switch st {
	case domain.AppViewed, domain.AppInterviewing, domain.AppOffered, domain.AppRejected, domain.AppWithdrawn:
	default:
		fail(w, r, http.StatusBadRequest, errors.New("invalid status"))
		return
	}
	app, err := s.jobSvc.UpdateApplicationStatus(r.Context(), a, r.PathValue("id"), st)
	if err != nil {
		fail(w, r, http.StatusForbidden, err)
		return
	}
	// 状态流转通知求职者（面试/录用/拒绝）——异步 goroutine 用 Background()：
	// r.Context() 会在 handler 返回后取消，导致通知发送被中断。
	switch st {
	case domain.AppInterviewing:
		go s.msgSvc.Send(context.Background(), "system", app.ApplicantID, "面试通知", "您的求职投递已被企业查看，邀请进入面试环节，请留意后续沟通", "application", app.ID)
	case domain.AppOffered:
		go s.msgSvc.Send(context.Background(), "system", app.ApplicantID, "录用通知", "恭喜！企业已向您发出录用意向，请及时联系确认入职安排", "application", app.ID)
	case domain.AppRejected:
		go s.msgSvc.Send(context.Background(), "system", app.ApplicantID, "投递结果", "很遗憾，您的求职投递未通过筛选，感谢关注，欢迎继续投递其他职位", "application", app.ID)
	}
	respond(w, r, http.StatusOK, app)
}

// GET /api/v1/applications?job_id=xxx
func (s *Server) listApplications(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	jobID := r.URL.Query().Get("job_id")
	if jobID == "" {
		items, err := s.jobSvc.ListMyApplications(r.Context(), a)
		if err != nil {
			fail(w, r, http.StatusInternalServerError, err)
			return
		}
		respond(w, r, http.StatusOK, items)
		return
	}
	items, err := s.jobSvc.ListApplicantsForJob(r.Context(), a, jobID)
	if err != nil {
		fail(w, r, http.StatusForbidden, err)
		return
	}
	respond(w, r, http.StatusOK, items)
}
