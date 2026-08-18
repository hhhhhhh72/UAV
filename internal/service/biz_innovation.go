package service

import (
	"context"
	"fmt"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)

// ---- AchievementService (科技成果) ----

type AchievementService struct {
	repo repository.AchievementRepository
}

func NewAchievementService(repo repository.AchievementRepository) *AchievementService {
	return &AchievementService{repo: repo}
}

func (s *AchievementService) Create(ctx context.Context, ownerID, title, achieveType, description, field, stage, contactInfo string, images []string, attachments []domain.Attachment) (domain.Achievement, error) {
	now := time.Now()
	a := domain.Achievement{
		ID:          nextID("achieve"),
		OwnerID:     ownerID,
		Title:       title,
		AchieveType: achieveType,
		Description: description,
		Field:       field,
		Stage:       stage,
		Images:      images,
		Attachments: attachments,
		ContactInfo: contactInfo,
		Status:      "published",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	return s.repo.Create(ctx, a)
}

func (s *AchievementService) List(ctx context.Context, field string, page, pageSize int) ([]domain.Achievement, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(ctx, field, offset, pageSize)
}

func (s *AchievementService) Get(ctx context.Context, id string) (domain.Achievement, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *AchievementService) Update(ctx context.Context, a domain.Actor, id, title, achieveType, description, field, stage, contactInfo string, images []string, attachments []domain.Attachment) (domain.Achievement, error) {
	ach, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return domain.Achievement{}, err
	}
	// 越权防护：仅属主或管理员可修改
	if !canMutate(a, ach.OwnerID) {
		return domain.Achievement{}, ErrNotOwner
	}
	ach.Title = title
	ach.AchieveType = achieveType
	ach.Description = description
	ach.Field = field
	ach.Stage = stage
	ach.Images = images
	ach.Attachments = attachments
	ach.ContactInfo = contactInfo
	ach.UpdatedAt = time.Now()
	return s.repo.Update(ctx, ach)
}

func (s *AchievementService) Delete(ctx context.Context, a domain.Actor, id string) error {
	ach, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	// 越权防护：仅属主或管理员可删除
	if !canMutate(a, ach.OwnerID) {
		return ErrNotOwner
	}
	return s.repo.Delete(ctx, id)
}

// ---- RDChallengeService (技术攻关/揭榜挂帅) ----

type RDChallengeService struct {
	repo repository.RDChallengeRepository
}

func NewRDChallengeService(repo repository.RDChallengeRepository) *RDChallengeService {
	return &RDChallengeService{repo: repo}
}

func (s *RDChallengeService) Create(ctx context.Context, posterID, title, field, description string, budgetFen int64, deadline time.Time) (domain.RDChallenge, error) {
	now := time.Now()
	c := domain.RDChallenge{
		ID:          nextID("challenge"),
		PosterID:    posterID,
		Title:       title,
		Field:       field,
		Description: description,
		BudgetFen:   budgetFen,
		Deadline:    deadline,
		Status:      "published",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	return s.repo.Create(ctx, c)
}

func (s *RDChallengeService) List(ctx context.Context, field string, page, pageSize int) ([]domain.RDChallenge, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(ctx, field, offset, pageSize)
}

func (s *RDChallengeService) Get(ctx context.Context, id string) (domain.RDChallenge, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *RDChallengeService) Update(ctx context.Context, a domain.Actor, id, title, field, description, status string, budgetFen int64, deadline time.Time) (domain.RDChallenge, error) {
	c, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return domain.RDChallenge{}, err
	}
	// 越权防护：仅发布者或管理员可修改
	if !canMutate(a, c.PosterID) {
		return domain.RDChallenge{}, ErrNotOwner
	}
	c.Title = title
	c.Field = field
	c.Description = description
	c.Status = status
	c.BudgetFen = budgetFen
	c.Deadline = deadline
	c.UpdatedAt = time.Now()
	return s.repo.Update(ctx, c)
}

func (s *RDChallengeService) Delete(ctx context.Context, a domain.Actor, id string) error {
	c, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	// 越权防护：仅发布者或管理员可删除
	if !canMutate(a, c.PosterID) {
		return ErrNotOwner
	}
	return s.repo.Delete(ctx, id)
}

// ---- ResearchProjectService (联合研发项目) ----

type ResearchProjectService struct {
	repo repository.ResearchProjectRepository
}

func NewResearchProjectService(repo repository.ResearchProjectRepository) *ResearchProjectService {
	return &ResearchProjectService{repo: repo}
}

func (s *ResearchProjectService) Create(ctx context.Context, title, field, description, leadOrg, milestones string, members []string, budgetFen int64, startDate, endDate time.Time) (domain.ResearchProject, error) {
	now := time.Now()
	p := domain.ResearchProject{
		ID:          nextID("proj"),
		Title:       title,
		Field:       field,
		Description: description,
		LeadOrg:     leadOrg,
		Members:     members,
		BudgetFen:   budgetFen,
		StartDate:   startDate,
		EndDate:     endDate,
		Milestones:  milestones,
		Status:      "active",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	return s.repo.Create(ctx, p)
}

func (s *ResearchProjectService) List(ctx context.Context, page, pageSize int) ([]domain.ResearchProject, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(ctx, offset, pageSize)
}

func (s *ResearchProjectService) Get(ctx context.Context, id string) (domain.ResearchProject, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *ResearchProjectService) Update(ctx context.Context, id, title, field, description, status, leadOrg, milestones string, members []string, budgetFen int64, startDate, endDate time.Time) (domain.ResearchProject, error) {
	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return domain.ResearchProject{}, err
	}
	p.Title = title
	p.Field = field
	p.Description = description
	p.Status = status
	p.LeadOrg = leadOrg
	p.Members = members
	p.BudgetFen = budgetFen
	p.StartDate = startDate
	p.EndDate = endDate
	p.Milestones = milestones
	p.UpdatedAt = time.Now()
	return s.repo.Update(ctx, p)
}

func (s *ResearchProjectService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// ---- ProjectAppService (项目申报) ----

type ProjectAppService struct {
	repo repository.ProjectAppRepository
}

func NewProjectAppService(repo repository.ProjectAppRepository) *ProjectAppService {
	return &ProjectAppService{repo: repo}
}

func (s *ProjectAppService) Create(ctx context.Context, applicantID, projectName, category, description string, budgetFen int64, attachments []string) (domain.ProjectApplication, error) {
	now := time.Now()
	a := domain.ProjectApplication{
		ID:          nextID("app"),
		ApplicantID: applicantID,
		ProjectName: projectName,
		Category:    category,
		BudgetFen:   budgetFen,
		Description: description,
		Attachments: attachments,
		Status:      "submitted",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	return s.repo.Create(ctx, a)
}

func (s *ProjectAppService) ListMy(ctx context.Context, userID string) ([]domain.ProjectApplication, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *ProjectAppService) ListAll(ctx context.Context, status string, page, pageSize int) ([]domain.ProjectApplication, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListAll(ctx, status, offset, pageSize)
}

func (s *ProjectAppService) Get(ctx context.Context, id string) (domain.ProjectApplication, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *ProjectAppService) Review(ctx context.Context, id, reviewNote, action string) (domain.ProjectApplication, error) {
	a, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return domain.ProjectApplication{}, err
	}
	if action == "approve" {
		a.Status = "approved"
	} else if action == "reject" {
		a.Status = "rejected"
	} else {
		return domain.ProjectApplication{}, fmt.Errorf("invalid action: %s", action)
	}
	a.ReviewNote = reviewNote
	a.UpdatedAt = time.Now()
	return s.repo.Update(ctx, a)
}
