package service

import (
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

func (s *AchievementService) Create(ownerID, title, achieveType, description, field, stage, contactInfo string, images []string) (domain.Achievement, error) {
	now := time.Now()
	a := domain.Achievement{
		ID:          fmt.Sprintf("achieve-%d", now.UnixNano()),
		OwnerID:     ownerID,
		Title:       title,
		AchieveType: achieveType,
		Description: description,
		Field:       field,
		Stage:       stage,
		Images:      images,
		ContactInfo: contactInfo,
		Status:      "published",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	return s.repo.Create(a)
}

func (s *AchievementService) List(field string, page, pageSize int) ([]domain.Achievement, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(field, offset, pageSize)
}

func (s *AchievementService) Get(id string) (domain.Achievement, error) {
	return s.repo.FindByID(id)
}

func (s *AchievementService) Update(id, title, achieveType, description, field, stage, contactInfo string, images []string) (domain.Achievement, error) {
	a, err := s.repo.FindByID(id)
	if err != nil {
		return domain.Achievement{}, err
	}
	a.Title = title
	a.AchieveType = achieveType
	a.Description = description
	a.Field = field
	a.Stage = stage
	a.Images = images
	a.ContactInfo = contactInfo
	a.UpdatedAt = time.Now()
	return s.repo.Update(a)
}

func (s *AchievementService) Delete(id string) error {
	return s.repo.Delete(id)
}

// ---- RDChallengeService (技术攻关/揭榜挂帅) ----

type RDChallengeService struct {
	repo repository.RDChallengeRepository
}

func NewRDChallengeService(repo repository.RDChallengeRepository) *RDChallengeService {
	return &RDChallengeService{repo: repo}
}

func (s *RDChallengeService) Create(posterID, title, field, description string, budgetFen int64, deadline time.Time) (domain.RDChallenge, error) {
	now := time.Now()
	c := domain.RDChallenge{
		ID:          fmt.Sprintf("challenge-%d", now.UnixNano()),
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
	return s.repo.Create(c)
}

func (s *RDChallengeService) List(field string, page, pageSize int) ([]domain.RDChallenge, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(field, offset, pageSize)
}

func (s *RDChallengeService) Get(id string) (domain.RDChallenge, error) {
	return s.repo.FindByID(id)
}

func (s *RDChallengeService) Update(id, title, field, description string, budgetFen int64, deadline time.Time) (domain.RDChallenge, error) {
	c, err := s.repo.FindByID(id)
	if err != nil {
		return domain.RDChallenge{}, err
	}
	c.Title = title
	c.Field = field
	c.Description = description
	c.BudgetFen = budgetFen
	c.Deadline = deadline
	c.UpdatedAt = time.Now()
	return s.repo.Update(c)
}

// ---- ResearchProjectService (联合研发项目) ----

type ResearchProjectService struct {
	repo repository.ResearchProjectRepository
}

func NewResearchProjectService(repo repository.ResearchProjectRepository) *ResearchProjectService {
	return &ResearchProjectService{repo: repo}
}

func (s *ResearchProjectService) Create(title, field, description, leadOrg, milestones string, members []string, budgetFen int64, startDate, endDate time.Time) (domain.ResearchProject, error) {
	now := time.Now()
	p := domain.ResearchProject{
		ID:          fmt.Sprintf("proj-%d", now.UnixNano()),
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
	return s.repo.Create(p)
}

func (s *ResearchProjectService) List(page, pageSize int) ([]domain.ResearchProject, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(offset, pageSize)
}

func (s *ResearchProjectService) Get(id string) (domain.ResearchProject, error) {
	return s.repo.FindByID(id)
}

func (s *ResearchProjectService) Update(id, title, field, description, leadOrg, milestones string, members []string, budgetFen int64, startDate, endDate time.Time) (domain.ResearchProject, error) {
	p, err := s.repo.FindByID(id)
	if err != nil {
		return domain.ResearchProject{}, err
	}
	p.Title = title
	p.Field = field
	p.Description = description
	p.LeadOrg = leadOrg
	p.Members = members
	p.BudgetFen = budgetFen
	p.StartDate = startDate
	p.EndDate = endDate
	p.Milestones = milestones
	p.UpdatedAt = time.Now()
	return s.repo.Update(p)
}

// ---- ProjectAppService (项目申报) ----

type ProjectAppService struct {
	repo repository.ProjectAppRepository
}

func NewProjectAppService(repo repository.ProjectAppRepository) *ProjectAppService {
	return &ProjectAppService{repo: repo}
}

func (s *ProjectAppService) Create(applicantID, projectName, category, description string, budgetFen int64, attachments []string) (domain.ProjectApplication, error) {
	now := time.Now()
	a := domain.ProjectApplication{
		ID:          fmt.Sprintf("app-%d", now.UnixNano()),
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
	return s.repo.Create(a)
}

func (s *ProjectAppService) ListMy(userID string) ([]domain.ProjectApplication, error) {
	return s.repo.ListByUser(userID)
}

func (s *ProjectAppService) ListAll(status string, page, pageSize int) ([]domain.ProjectApplication, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListAll(status, offset, pageSize)
}

func (s *ProjectAppService) Get(id string) (domain.ProjectApplication, error) {
	return s.repo.FindByID(id)
}

func (s *ProjectAppService) Review(id, reviewNote, action string) (domain.ProjectApplication, error) {
	a, err := s.repo.FindByID(id)
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
	return s.repo.Update(a)
}
