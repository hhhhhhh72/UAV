package service

import (
	"fmt"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)

// ── Transformation Service ──

type TransformationService struct {
	repo repository.TransformationRepository
}

func NewTransformationService(r repository.TransformationRepository) *TransformationService {
	return &TransformationService{repo: r}
}

func (s *TransformationService) Create(title, achievementID, ownerID, partnerID string) (domain.Transformation, error) {
	t := domain.Transformation{ID: fmt.Sprintf("tran-%d", time.Now().UnixNano()),
		Title: title, AchievementID: achievementID, OwnerID: ownerID, PartnerID: partnerID,
		Stage: domain.StageLab, Status: "active", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	return s.repo.Create(t)
}

func (s *TransformationService) Get(id string) (domain.Transformation, error) {
	return s.repo.FindByID(id)
}

func (s *TransformationService) DeleteTrans(id string) error { return s.repo.Delete(id) }

func (s *TransformationService) UpdateTrans(id, title, stage, progress, partnerID, status string) (domain.Transformation, error) {
	t, err := s.repo.FindByID(id)
	if err != nil {
		return domain.Transformation{}, err
	}
	t.Title = title
	t.Stage = domain.TransformationStage(stage)
	t.Progress = progress
	t.Status = status
	t.PartnerID = partnerID
	t.UpdatedAt = time.Now()
	return s.repo.Update(t)
}

func (s *TransformationService) List(ownerID string) ([]domain.Transformation, error) {
	return s.repo.List(ownerID)
}

// ListByAchievement 按成果查询转化记录（数据量小，内存过滤）
func (s *TransformationService) ListByAchievement(achievementID string) ([]domain.Transformation, error) {
	all, err := s.repo.List("")
	if err != nil {
		return nil, err
	}
	out := make([]domain.Transformation, 0, len(all))
	for _, t := range all {
		if t.AchievementID == achievementID {
			out = append(out, t)
		}
	}
	return out, nil
}

func (s *TransformationService) AdvanceStage(id string, nextStage domain.TransformationStage, progress string) (domain.Transformation, error) {
	t, err := s.repo.FindByID(id)
	if err != nil {
		return domain.Transformation{}, err
	}
	t.Stage = nextStage
	t.Progress = progress
	t.UpdatedAt = time.Now()
	return s.repo.Update(t)
}

func (s *TransformationService) AddMilestone(id, name, evidence string) (domain.Transformation, error) {
	t, err := s.repo.FindByID(id)
	if err != nil {
		return domain.Transformation{}, err
	}
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

// CoopTypeResearch / CoopTypeTalent / CoopTypeBoth 院校分域（功能方案修订版 三·五 分域）。
const (
	CoopTypeResearch = "research" // 科研合作（三系统）
	CoopTypeTalent   = "talent"   // 人才培养（五系统）
	CoopTypeBoth     = "both"     // 综合
)

// CoopTypeLabel 返回分域中文名。
func CoopTypeLabel(t string) string {
	switch t {
	case CoopTypeResearch:
		return "科研合作"
	case CoopTypeTalent:
		return "人才培养"
	case CoopTypeBoth:
		return "综合"
	default:
		return "综合"
	}
}

func (s *CollegeService) Create(name, region, description, logoURL, coopType string, majors, facilities []string) (domain.College, error) {
	c := domain.College{ID: fmt.Sprintf("col-%d", time.Now().UnixNano()),
		Name: name, Region: region, Description: description, LogoURL: logoURL,
		CoopType: coopType, Majors: majors, Facilities: facilities, Status: "active",
		CreatedAt: time.Now(), UpdatedAt: time.Now()}
	return s.repo.Create(c)
}
func (s *CollegeService) List(region string) ([]domain.College, error) { return s.repo.List(region) }
func (s *CollegeService) Get(id string) (domain.College, error)        { return s.repo.FindByID(id) }

func (s *CollegeService) Update(id, name, region, description, logoURL, status, coopType string, majors, facilities []string) (domain.College, error) {
	c, err := s.repo.FindByID(id)
	if err != nil {
		return domain.College{}, err
	}
	c.Name = name
	c.Region = region
	c.Description = description
	c.LogoURL = logoURL
	c.Status = status
	c.CoopType = coopType
	c.Majors = majors
	c.Facilities = facilities
	c.UpdatedAt = time.Now()
	return s.repo.Update(c)
}

func (s *CollegeService) Delete(id string) error { return s.repo.Delete(id) }

// ── Cooperation Service ──

type CooperationService struct {
	repo repository.CooperationRepository
}

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
func (s *CooperationService) Get(id string) (domain.CooperationProgram, error) {
	return s.repo.FindByID(id)
}
func (s *CooperationService) UpdateStatus(id, status string) (domain.CooperationProgram, error) {
	return s.repo.UpdateStatus(id, status)
}
