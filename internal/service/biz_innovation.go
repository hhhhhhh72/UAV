package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
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

func (s *AchievementService) Create(ctx context.Context, ownerID, title, achieveType, description, field, stage, contactInfo string, images []string, attachments []domain.Attachment, status string) (domain.Achievement, error) {
	now := time.Now()
	if status == "" {
		status = "published"
	}
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
		Status:      status,
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

func (s *AchievementService) Update(ctx context.Context, a domain.Actor, id, title, achieveType, description, field, stage, contactInfo string, images []string, attachments []domain.Attachment, status string) (domain.Achievement, error) {
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
	if status != "" {
		ach.Status = status
	}
	ach.UpdatedAt = time.Now()
	return s.repo.Update(ctx, ach)
}

// AdjustStats 浏览/收藏计数增量（负值下限 0）。
func (s *AchievementService) AdjustStats(ctx context.Context, id string, viewsDelta, favsDelta int) error {
	return s.repo.AdjustStats(ctx, id, viewsDelta, favsDelta)
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

func (s *RDChallengeService) Create(ctx context.Context, posterID, title, field, description, requirements string, budgetFen int64, deadline time.Time, status string) (domain.RDChallenge, error) {
	if budgetFen < 0 {
		return domain.RDChallenge{}, fmt.Errorf("budget cannot be negative")
	}
	now := time.Now()
	if status == "" {
		status = "published"
	}
	c := domain.RDChallenge{
		ID:           nextID("challenge"),
		PosterID:     posterID,
		Title:        title,
		Field:        field,
		Description:  description,
		Requirements: requirements,
		BudgetFen:    budgetFen,
		Deadline:     deadline,
		Status:       status,
		CreatedAt:    now,
		UpdatedAt:    now,
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

func (s *RDChallengeService) Update(ctx context.Context, a domain.Actor, id, title, field, description, requirements, status string, budgetFen int64, deadline time.Time) (domain.RDChallenge, error) {
	if budgetFen < 0 {
		return domain.RDChallenge{}, fmt.Errorf("budget cannot be negative")
	}
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
	c.Requirements = requirements
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
	if budgetFen < 0 {
		return domain.ResearchProject{}, fmt.Errorf("budget cannot be negative")
	}
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
	if budgetFen < 0 {
		return domain.ResearchProject{}, fmt.Errorf("budget cannot be negative")
	}
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

// ---- 参与申请（课题攻关） ----

// ErrProjectNotFound 课题不存在（handler 映射 404）。
var ErrProjectNotFound = errors.New("research project not found")

// ErrJoinStatusInvalid 后台流转状态非法。
var ErrJoinStatusInvalid = errors.New("invalid join status")

// JoinProject 用户申请参与课题：每用户每课题一条记录，幂等——
// 已存在且未关闭时原样返回（重复点击不重复提交）；已关闭的记录允许同一用户重新申请（更新原记录）。
// 返回 (申请记录, 是否新建, error)。
func (s *ResearchProjectService) JoinProject(ctx context.Context, userID, projectID, orgName, message string) (domain.ProjectJoinRequest, bool, error) {
	if _, err := s.repo.FindByID(ctx, projectID); err != nil {
		return domain.ProjectJoinRequest{}, false, ErrProjectNotFound
	}
	orgName = strings.TrimSpace(orgName)
	message = strings.TrimSpace(message)
	existing, err := s.repo.FindJoinByProjectUser(ctx, projectID, userID)
	if err == nil {
		if existing.Status != "closed" {
			return existing, false, nil // 重复申请：幂等返回
		}
		// closed 后重新申请：同一记录重置为 pending，更新内容
		existing.OrgName = orgName
		existing.Message = message
		existing.Status = "pending"
		upd, uerr := s.repo.UpdateJoinRequest(ctx, existing)
		if uerr != nil {
			return domain.ProjectJoinRequest{}, false, fmt.Errorf("renew join request: %w", uerr)
		}
		return upd, false, nil
	}
	now := time.Now()
	v := domain.ProjectJoinRequest{
		ID:        nextID("pjreq"),
		ProjectID: projectID,
		UserID:    userID,
		OrgName:   orgName,
		Message:   message,
		Status:    "pending",
		CreatedAt: now,
		UpdatedAt: now,
	}
	created, cerr := s.repo.CreateJoinRequest(ctx, v)
	if cerr != nil {
		return domain.ProjectJoinRequest{}, false, fmt.Errorf("create join request: %w", cerr)
	}
	return created, true, nil
}

// GetMyJoin 查询当前用户在某课题下的申请（applied=false 表示未申请过）。
func (s *ResearchProjectService) GetMyJoin(ctx context.Context, projectID, userID string) (domain.ProjectJoinRequest, bool, error) {
	v, err := s.repo.FindJoinByProjectUser(ctx, projectID, userID)
	if err != nil {
		return domain.ProjectJoinRequest{}, false, nil
	}
	return v, true, nil
}

// ListJoins 后台查看课题全部申请（新→旧）。
func (s *ResearchProjectService) ListJoins(ctx context.Context, projectID string) ([]domain.ProjectJoinRequest, error) {
	if _, err := s.repo.FindByID(ctx, projectID); err != nil {
		return nil, ErrProjectNotFound
	}
	return s.repo.ListJoinRequests(ctx, projectID)
}

// UpdateJoinStatus 后台流转申请状态：pending 待评估 / contacted 已对接 / closed 已关闭。
func (s *ResearchProjectService) UpdateJoinStatus(ctx context.Context, joinID, status string) (domain.ProjectJoinRequest, error) {
	if status != "pending" && status != "contacted" && status != "closed" {
		return domain.ProjectJoinRequest{}, ErrJoinStatusInvalid
	}
	v, err := s.repo.FindJoinByID(ctx, joinID)
	if err != nil {
		return domain.ProjectJoinRequest{}, err
	}
	v.Status = status
	return s.repo.UpdateJoinRequest(ctx, v)
}

// ---- ProjectAppService (项目申报) ----

type ProjectAppService struct {
	repo repository.ProjectAppRepository
}

func NewProjectAppService(repo repository.ProjectAppRepository) *ProjectAppService {
	return &ProjectAppService{repo: repo}
}

func (s *ProjectAppService) Create(ctx context.Context, applicantID, projectName, category, description string, budgetFen int64, attachments []string) (domain.ProjectApplication, error) {
	if budgetFen < 0 {
		return domain.ProjectApplication{}, fmt.Errorf("budget cannot be negative")
	}
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
	// 状态机门禁：仅「已提交」可审——获批/驳回后不得再翻转（防重复审核/反复改判）。
	if a.Status != "submitted" {
		return domain.ProjectApplication{}, fmt.Errorf("only submitted applications can be reviewed (current status %s)", a.Status)
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
