package service

import (
	"context"
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

var (
	// ErrInvalidJobStatus 管理端更新职位时传入了白名单（draft/published/closed）之外的状态。
	ErrInvalidJobStatus = errors.New("invalid job status")
	// ErrInvalidJobTransition 状态机不允许的流转（如发布已发布的职位、关闭草稿职位）。
	ErrInvalidJobTransition = errors.New("invalid job status transition")
)

// validJobStatus 职位状态白名单：draft → published → closed，无其他状态。
func validJobStatus(s domain.JobStatus) bool {
	return s == domain.JobDraft || s == domain.JobPublished || s == domain.JobClosed
}

func NewJobService(j repository.JobRepository, r repository.ResumeRepository, a repository.JobApplicationRepository) *JobService {
	return &JobService{repo: j, resume: r, app: a}
}

// ---- Jobs ----

func (s *JobService) CreateJob(ctx context.Context, a domain.Actor, title, desc, location string, salaryFen int64) (domain.Job, error) {
	if a.Role != domain.RoleEnterprise && a.Role != domain.RolePlatformAdmin {
		return domain.Job{}, errors.New("only enterprise can post jobs")
	}
	now := time.Now()
	// nextSeq 保证同纳秒连续创建时 ID 唯一（Windows 时钟精度约 100ns），
	// 否则内存/PG 按 ID 更新会错配到前一条记录。
	j := domain.Job{ID: fmt.Sprintf("job-%d-%d", now.UnixNano(), nextSeq()), EnterpriseID: a.ID, Title: title, Description: desc,
		Location: location, SalaryFen: salaryFen, Status: domain.JobDraft, Version: 1, CreatedAt: now, UpdatedAt: now}
	slog.Info("job created", "job_id", j.ID, "enterprise_id", j.EnterpriseID)
	return s.repo.Create(ctx, j)
}
func (s *JobService) PublishJob(ctx context.Context, a domain.Actor, id string) (domain.Job, error) {
	j, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return domain.Job{}, err
	}
	if j.EnterpriseID != a.ID {
		return domain.Job{}, errors.New("only the owner can publish")
	}
	if j.Status != domain.JobDraft {
		return domain.Job{}, fmt.Errorf("%w: cannot publish job in %q status", ErrInvalidJobTransition, j.Status)
	}
	j.Status = domain.JobPublished
	j.UpdatedAt = time.Now()
	slog.Info("job updated", "job_id", id)
	return s.repo.Update(ctx, id, j)
}
func (s *JobService) CloseJob(ctx context.Context, a domain.Actor, id string) (domain.Job, error) {
	j, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return domain.Job{}, err
	}
	if j.EnterpriseID != a.ID {
		return domain.Job{}, errors.New("only the owner can close")
	}
	if j.Status != domain.JobPublished {
		return domain.Job{}, fmt.Errorf("%w: cannot close job in %q status", ErrInvalidJobTransition, j.Status)
	}
	j.Status = domain.JobClosed
	j.UpdatedAt = time.Now()
	slog.Info("job updated", "job_id", id)
	return s.repo.Update(ctx, id, j)
}

// ListAllJobs 管理端全量列表（含草稿），供 admin 列表页使用。
func (s *JobService) ListAllJobs(ctx context.Context, offset, limit int) ([]domain.Job, int, error) {
	return s.repo.ListAll(ctx, offset, limit)
}

func (s *JobService) ListPublishedJobs(ctx context.Context, offset, limit int) ([]domain.Job, int, error) {
	return s.repo.ListPublished(ctx, offset, limit)
}
func (s *JobService) ListMyJobs(ctx context.Context, a domain.Actor) ([]domain.Job, error) {
	return s.repo.ListByEnterprise(ctx, a.ID)
}

func (s *JobService) GetJob(ctx context.Context, id string) (domain.Job, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *JobService) UpdateJob(ctx context.Context, id, title, desc, location, jobType string, salaryFen int64, status string) (domain.Job, error) {
	if !validJobStatus(domain.JobStatus(status)) {
		return domain.Job{}, fmt.Errorf("%w: %q", ErrInvalidJobStatus, status)
	}
	j, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return domain.Job{}, err
	}
	j.Title = title
	j.Description = desc
	j.Location = location
	j.SalaryFen = salaryFen
	j.Status = domain.JobStatus(status)
	j.JobType = jobType
	return s.repo.Update(ctx, id, j)
}

func (s *JobService) DeleteJob(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// ---- Resumes ----

func (s *JobService) CreateResume(ctx context.Context, a domain.Actor, title, name, phone, email, education, workExperience string, skills []string, certificateURL, content, visibility string) (domain.Resume, error) {
	now := time.Now()
	r := domain.Resume{ID: nextID("resume"), UserID: a.ID, Title: title,
		Name: name, Phone: phone, Email: email, Education: education, WorkExperience: workExperience,
		Skills: skills, CertificateURL: certificateURL, Content: content, Visibility: visibility,
		Version: 1, CreatedAt: now, UpdatedAt: now}
	return s.resume.Create(ctx, r)
}
func (s *JobService) UpdateResume(ctx context.Context, a domain.Actor, id, title, name, phone, email, education, workExperience string, skills []string, certificateURL, content, visibility string) (domain.Resume, error) {
	r, err := s.resume.FindByID(ctx, id)
	if err != nil {
		return domain.Resume{}, err
	}
	if r.UserID != a.ID {
		return domain.Resume{}, errors.New("only the owner can edit")
	}
	r.Title = title
	r.Name = name
	r.Phone = phone
	r.Email = email
	r.Education = education
	r.WorkExperience = workExperience
	r.Skills = skills
	r.CertificateURL = certificateURL
	r.Content = content
	r.Visibility = visibility
	r.UpdatedAt = time.Now()
	return s.resume.Update(ctx, id, r)
}
func (s *JobService) ListMyResumes(ctx context.Context, a domain.Actor) ([]domain.Resume, error) {
	return s.resume.ListByUser(ctx, a.ID)
}

func (s *JobService) ListAllResumes(ctx context.Context, offset, limit int) ([]domain.Resume, int, error) {
	return s.resume.ListAll(ctx, offset, limit)
}

// ---- Applications ----

func (s *JobService) Apply(ctx context.Context, a domain.Actor, jobID, resumeID string) (domain.JobApplication, error) {
	j, err := s.repo.FindByID(ctx, jobID)
	if err != nil {
		return domain.JobApplication{}, err
	}
	// 仅招聘中的职位可投递：草稿/已关闭一律拒绝（旧实现草稿也可投递）
	if j.Status != domain.JobPublished {
		return domain.JobApplication{}, fmt.Errorf("job %q is not open for applications", j.Status)
	}
	if j.EnterpriseID == a.ID {
		return domain.JobApplication{}, errors.New("cannot apply to your own job")
	}
	// 并发防重复投递：check-then-insert 加进程内锁。
	unlock := lockByKey("apply|" + a.ID + "|" + jobID)
	defer unlock()
	// 防重复投递：同一职位已有有效投递（未撤回）则拒绝
	existing, err := s.app.ListByJob(ctx, jobID)
	if err == nil {
		for _, e := range existing {
			if e.ApplicantID == a.ID && e.Status != domain.AppWithdrawn {
				return domain.JobApplication{}, errors.New("you have already applied to this job")
			}
		}
	}
	now := time.Now()
	app := domain.JobApplication{ID: nextID("app"), JobID: jobID, ResumeID: resumeID,
		ApplicantID: a.ID, Status: domain.AppSubmitted, Version: 1, CreatedAt: now, UpdatedAt: now}
	return s.app.Create(ctx, app)
}
func (s *JobService) UpdateApplicationStatus(ctx context.Context, a domain.Actor, appID string, status domain.AppStatus) (domain.JobApplication, error) {
	// 归属校验前置（C5 修复）：必须先确认操作者身份再落库，
	// 旧实现先 UpdateStatus 后校验，越权写入已生效才返回 403。
	ap, err := s.app.FindByID(ctx, appID)
	if err != nil {
		return domain.JobApplication{}, err
	}
	j, err := s.repo.FindByID(ctx, ap.JobID)
	if err != nil {
		return domain.JobApplication{}, err
	}
	if j.EnterpriseID != a.ID && ap.ApplicantID != a.ID {
		return domain.JobApplication{}, errors.New("only the job owner or applicant can update")
	}
	return s.app.UpdateStatus(ctx, appID, status)
}
func (s *JobService) ListApplicationsForJob(ctx context.Context, a domain.Actor, jobID string) ([]domain.JobApplication, error) {
	j, err := s.repo.FindByID(ctx, jobID)
	if err != nil {
		return nil, err
	}
	if j.EnterpriseID != a.ID {
		return nil, errors.New("only the job owner can view applications")
	}
	return s.app.ListByJob(ctx, jobID)
}

// ApplicantView 企业视角的投递 + 简历快照。
type ApplicantView struct {
	Application domain.JobApplication `json:"application"`
	Resume      domain.Resume         `json:"resume"`
}

// ListApplicantsForJob 企业查看某职位的投递者（含简历快照）。
func (s *JobService) ListApplicantsForJob(ctx context.Context, a domain.Actor, jobID string) ([]ApplicantView, error) {
	apps, err := s.ListApplicationsForJob(ctx, a, jobID)
	if err != nil {
		return nil, err
	}
	out := make([]ApplicantView, 0, len(apps))
	for _, ap := range apps {
		rs, err := s.resume.FindByID(ctx, ap.ResumeID)
		if err != nil {
			continue // 简历已删：跳过该投递
		}
		out = append(out, ApplicantView{Application: ap, Resume: rs})
	}
	return out, nil
}
func (s *JobService) ListMyApplications(ctx context.Context, a domain.Actor) ([]domain.JobApplication, error) {
	return s.app.ListByApplicant(ctx, a.ID)
}
