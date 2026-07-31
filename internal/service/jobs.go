package service

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)

type JobService struct {
	repo   repository.JobRepository
	resume repository.ResumeRepository
	app    repository.JobApplicationRepository
}

func NewJobService(j repository.JobRepository, r repository.ResumeRepository, a repository.JobApplicationRepository) *JobService {
	return &JobService{repo: j, resume: r, app: a}
}

// ---- Jobs ----

func (s *JobService) CreateJob(a domain.Actor, title, desc, location string, salaryFen int64) (domain.Job, error) {
	if a.Role != domain.RoleEnterprise && a.Role != domain.RolePlatformAdmin {
		return domain.Job{}, errors.New("only enterprise can post jobs")
	}
	now := time.Now()
	j := domain.Job{ID: fmt.Sprintf("job-%d", now.UnixNano()), EnterpriseID: a.ID, Title: title, Description: desc,
		Location: location, SalaryFen: salaryFen, Status: domain.JobDraft, Version: 1, CreatedAt: now, UpdatedAt: now}
	slog.Info("job created", "job_id", j.ID, "enterprise_id", j.EnterpriseID)
	return s.repo.Create(j)
}
func (s *JobService) PublishJob(a domain.Actor, id string) (domain.Job, error) {
	j, err := s.repo.FindByID(id)
	if err != nil { return domain.Job{}, err }
	if j.EnterpriseID != a.ID { return domain.Job{}, errors.New("only the owner can publish") }
	j.Status = domain.JobPublished; j.UpdatedAt = time.Now()
	slog.Info("job updated", "job_id", id)
	return s.repo.Update(id, j)
}
func (s *JobService) CloseJob(a domain.Actor, id string) (domain.Job, error) {
	j, err := s.repo.FindByID(id)
	if err != nil { return domain.Job{}, err }
	if j.EnterpriseID != a.ID { return domain.Job{}, errors.New("only the owner can close") }
	j.Status = domain.JobClosed; j.UpdatedAt = time.Now()
	slog.Info("job updated", "job_id", id)
	return s.repo.Update(id, j)
}
func (s *JobService) ListPublishedJobs(offset, limit int) ([]domain.Job, int, error) {
	return s.repo.ListPublished(offset, limit)
}
func (s *JobService) ListMyJobs(a domain.Actor) ([]domain.Job, error) {
	return s.repo.ListByEnterprise(a.ID)
}

func (s *JobService) GetJob(id string) (domain.Job, error) { return s.repo.FindByID(id) }

func (s *JobService) UpdateJob(id, title, desc, location, jobType string, salaryFen int64, status string) (domain.Job, error) {
	j, err := s.repo.FindByID(id)
	if err != nil { return domain.Job{}, err }
	j.Title = title; j.Description = desc; j.Location = location; j.SalaryFen = salaryFen
	j.Status = domain.JobStatus(status); j.JobType = jobType
	return s.repo.Update(id, j)
}

func (s *JobService) DeleteJob(id string) error {
	return s.repo.Delete(id)
}

// ---- Resumes ----

func (s *JobService) CreateResume(a domain.Actor, title, content, visibility string) (domain.Resume, error) {
	now := time.Now()
	r := domain.Resume{ID: fmt.Sprintf("resume-%d", now.UnixNano()), UserID: a.ID, Title: title,
		Content: content, Visibility: visibility, Version: 1, CreatedAt: now, UpdatedAt: now}
	return s.resume.Create(r)
}
func (s *JobService) UpdateResume(a domain.Actor, id, title, content, visibility string) (domain.Resume, error) {
	r, err := s.resume.FindByID(id)
	if err != nil { return domain.Resume{}, err }
	if r.UserID != a.ID { return domain.Resume{}, errors.New("only the owner can edit") }
	r.Title = title; r.Content = content; r.Visibility = visibility; r.UpdatedAt = time.Now()
	return s.resume.Update(id, r)
}
func (s *JobService) ListMyResumes(a domain.Actor) ([]domain.Resume, error) {
	return s.resume.ListByUser(a.ID)
}

// ---- Applications ----

func (s *JobService) Apply(a domain.Actor, jobID, resumeID string) (domain.JobApplication, error) {
	j, err := s.repo.FindByID(jobID)
	if err != nil { return domain.JobApplication{}, err }
	if j.EnterpriseID == a.ID { return domain.JobApplication{}, errors.New("cannot apply to your own job") }
	now := time.Now()
	app := domain.JobApplication{ID: fmt.Sprintf("app-%d", now.UnixNano()), JobID: jobID, ResumeID: resumeID,
		ApplicantID: a.ID, Status: domain.AppSubmitted, Version: 1, CreatedAt: now, UpdatedAt: now}
	return s.app.Create(app)
}
func (s *JobService) UpdateApplicationStatus(a domain.Actor, appID string, status domain.AppStatus) (domain.JobApplication, error) {
	ap, err := s.app.UpdateStatus(appID, status)
	if err != nil { return domain.JobApplication{}, err }
	j, err := s.repo.FindByID(ap.JobID)
	if err != nil { return domain.JobApplication{}, err }
	if j.EnterpriseID != a.ID && ap.ApplicantID != a.ID {
		return domain.JobApplication{}, errors.New("only the job owner or applicant can update")
	}
	return ap, nil
}
func (s *JobService) ListApplicationsForJob(a domain.Actor, jobID string) ([]domain.JobApplication, error) {
	j, err := s.repo.FindByID(jobID)
	if err != nil { return nil, err }
	if j.EnterpriseID != a.ID { return nil, errors.New("only the job owner can view applications") }
	return s.app.ListByJob(jobID)
}
func (s *JobService) ListMyApplications(a domain.Actor) ([]domain.JobApplication, error) {
	return s.app.ListByApplicant(a.ID)
}
