package service

import (
	"fmt"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)

// ── Transformation Service ──

type TransformationService struct{ repo repository.TransformationRepository }

func NewTransformationService(r repository.TransformationRepository) *TransformationService {
	return &TransformationService{repo: r}
}

func (s *TransformationService) Create(title, achievementID, ownerID, partnerID string) (domain.Transformation, error) {
	t := domain.Transformation{ID: fmt.Sprintf("tran-%d", time.Now().UnixNano()),
		Title: title, AchievementID: achievementID, OwnerID: ownerID, PartnerID: partnerID,
		Stage: domain.StageLab, Status: "active", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	return s.repo.Create(t)
}

func (s *TransformationService) Get(id string) (domain.Transformation, error) { return s.repo.FindByID(id) }

func (s *TransformationService) List(ownerID string) ([]domain.Transformation, error) {
	return s.repo.List(ownerID)
}

func (s *TransformationService) AdvanceStage(id string, nextStage domain.TransformationStage, progress string) (domain.Transformation, error) {
	t, err := s.repo.FindByID(id)
	if err != nil { return domain.Transformation{}, err }
	t.Stage = nextStage
	t.Progress = progress
	t.UpdatedAt = time.Now()
	return s.repo.Update(t)
}

func (s *TransformationService) AddMilestone(id, name, evidence string) (domain.Transformation, error) {
	t, err := s.repo.FindByID(id)
	if err != nil { return domain.Transformation{}, err }
	t.Milestones = append(t.Milestones, domain.TransMilestone{
		Name: name, Completed: true, Date: time.Now(), Evidence: evidence,
	})
	t.UpdatedAt = time.Now()
	return s.repo.Update(t)
}

// ── College Service ──

type CollegeService struct{ repo repository.CollegeRepository }

func NewCollegeService(r repository.CollegeRepository) *CollegeService {
	return &CollegeService{repo: r}
}

func (s *CollegeService) Create(name, region, description, logoURL string, majors, facilities []string) (domain.College, error) {
	c := domain.College{ID: fmt.Sprintf("col-%d", time.Now().UnixNano()),
		Name: name, Region: region, Description: description, LogoURL: logoURL,
		Majors: majors, Facilities: facilities, Status: "active",
		CreatedAt: time.Now(), UpdatedAt: time.Now()}
	return s.repo.Create(c)
}
func (s *CollegeService) List(region string) ([]domain.College, error) { return s.repo.List(region) }
func (s *CollegeService) Get(id string) (domain.College, error)          { return s.repo.FindByID(id) }

// ── Cooperation Service ──

type CooperationService struct{ repo repository.CooperationRepository }

func NewCooperationService(r repository.CooperationRepository) *CooperationService {
	return &CooperationService{repo: r}
}

func (s *CooperationService) Create(title, collegeID, enterpriseID, coopType, description string, startDate, endDate time.Time, quota int) (domain.CooperationProgram, error) {
	cp := domain.CooperationProgram{ID: fmt.Sprintf("coop-%d", time.Now().UnixNano()),
		Title: title, CollegeID: collegeID, EnterpriseID: enterpriseID, CoopType: coopType,
		Description: description, StartDate: startDate, EndDate: endDate, StudentQuota: quota,
		Status: "proposed", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	return s.repo.Create(cp)
}
func (s *CooperationService) List(enterpriseID string) ([]domain.CooperationProgram, error) {
	return s.repo.List(enterpriseID)
}
func (s *CooperationService) Get(id string) (domain.CooperationProgram, error) { return s.repo.FindByID(id) }
func (s *CooperationService) UpdateStatus(id, status string) (domain.CooperationProgram, error) {
	return s.repo.UpdateStatus(id, status)
}
