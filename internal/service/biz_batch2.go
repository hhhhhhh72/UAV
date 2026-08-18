package service

import (
	"context"
	"errors"
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

func (s *TransformationService) Create(ctx context.Context, title, achievementID, ownerID, partnerID string) (domain.Transformation, error) {
	t := domain.Transformation{ID: nextID("tran"),
		Title: title, AchievementID: achievementID, OwnerID: ownerID, PartnerID: partnerID,
		Stage: domain.StageLab, Status: "active", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	return s.repo.Create(ctx, t)
}

func (s *TransformationService) Get(ctx context.Context, id string) (domain.Transformation, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *TransformationService) DeleteTrans(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *TransformationService) UpdateTrans(ctx context.Context, id, title, stage, progress, partnerID, status string) (domain.Transformation, error) {
	t, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return domain.Transformation{}, err
	}
	t.Title = title
	t.Stage = domain.TransformationStage(stage)
	t.Progress = progress
	t.Status = status
	t.PartnerID = partnerID
	t.UpdatedAt = time.Now()
	return s.repo.Update(ctx, t)
}

func (s *TransformationService) List(ctx context.Context, ownerID string) ([]domain.Transformation, error) {
	return s.repo.List(ctx, ownerID)
}

// ListByAchievement 按成果查询转化记录（数据量小，内存过滤）
func (s *TransformationService) ListByAchievement(ctx context.Context, achievementID string) ([]domain.Transformation, error) {
	all, err := s.repo.List(ctx, "")
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

// canMutateTransformation 归属校验：仅转化负责人或管理员可推进阶段/添加里程碑（C2 修复）。
func canMutateTransformation(a domain.Actor, t domain.Transformation) bool {
	if t.OwnerID == a.ID {
		return true
	}
	return a.Role == domain.RolePlatformAdmin || a.Role == domain.RoleAssociationAdmin
}

func (s *TransformationService) AdvanceStage(ctx context.Context, a domain.Actor, id string, nextStage domain.TransformationStage, progress string) (domain.Transformation, error) {
	t, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return domain.Transformation{}, err
	}
	if !canMutateTransformation(a, t) {
		return domain.Transformation{}, errors.New("only the owner or admin can advance stage")
	}
	t.Stage = nextStage
	t.Progress = progress
	t.UpdatedAt = time.Now()
	return s.repo.Update(ctx, t)
}

func (s *TransformationService) AddMilestone(ctx context.Context, a domain.Actor, id, name, evidence string) (domain.Transformation, error) {
	t, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return domain.Transformation{}, err
	}
	if !canMutateTransformation(a, t) {
		return domain.Transformation{}, errors.New("only the owner or admin can add milestone")
	}
	t.Milestones = append(t.Milestones, domain.TransMilestone{
		Name: name, Completed: true, Date: time.Now(), Evidence: evidence,
	})
	t.UpdatedAt = time.Now()
	return s.repo.Update(ctx, t)
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

// Create 接收完整领域对象（含小程序页面字段 city/tags/major_count/cover 等）。
func (s *CollegeService) Create(ctx context.Context, c domain.College) (domain.College, error) {
	now := time.Now()
	if c.ID == "" {
		c.ID = nextID("col")
	}
	if c.Status == "" {
		c.Status = "active"
	}
	c.CreatedAt = now
	c.UpdatedAt = now
	return s.repo.Create(ctx, c)
}
func (s *CollegeService) List(ctx context.Context, region string) ([]domain.College, error) {
	return s.repo.List(ctx, region)
}
func (s *CollegeService) Get(ctx context.Context, id string) (domain.College, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *CollegeService) Update(ctx context.Context, c domain.College) (domain.College, error) {
	old, err := s.repo.FindByID(ctx, c.ID)
	if err != nil {
		return domain.College{}, err
	}
	c.CreatedAt = old.CreatedAt // 保留原创建时间
	c.UpdatedAt = time.Now()
	return s.repo.Update(ctx, c)
}

func (s *CollegeService) Delete(ctx context.Context, id string) error { return s.repo.Delete(ctx, id) }

// ── Cooperation Service ──

type CooperationService struct {
	repo repository.CooperationRepository
}

func NewCooperationService(r repository.CooperationRepository) *CooperationService {
	return &CooperationService{repo: r}
}

func (s *CooperationService) Create(ctx context.Context, title, collegeID, enterpriseID, coopType, description string, startDate, endDate time.Time, quota int) (domain.CooperationProgram, error) {
	cp := domain.CooperationProgram{ID: nextID("coop"),
		Title: title, CollegeID: collegeID, EnterpriseID: enterpriseID, CoopType: coopType,
		Description: description, StartDate: startDate, EndDate: endDate, StudentQuota: quota,
		Status: "proposed", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	return s.repo.Create(ctx, cp)
}
func (s *CooperationService) List(ctx context.Context, enterpriseID string) ([]domain.CooperationProgram, error) {
	return s.repo.List(ctx, enterpriseID)
}
func (s *CooperationService) Get(ctx context.Context, id string) (domain.CooperationProgram, error) {
	return s.repo.FindByID(ctx, id)
}
func (s *CooperationService) UpdateStatus(ctx context.Context, id, status string) (domain.CooperationProgram, error) {
	return s.repo.UpdateStatus(ctx, id, status)
}
