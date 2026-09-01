// Package service implements the business logic layer of the drone platform.
//
// Each service encapsulates a business domain: role checks, ownership validation,
// state machine enforcement, and coordination of multiple repositories.
// Services call repository interfaces only — they never touch HTTP, JSON, or SQL.
//
// Key services:
//   - DemandService — demand lifecycle with bid management and CAS-based selection
//   - EnterpriseSvc  — enterprise registration and admin review workflow
//   - ContractService — contract signing lifecycle with state machine validation
//   - EmploymentService, JobService, CommunityService, etc. — see individual docs
package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)

type CreateDemandInput struct {
	PublisherName string         `json:"publisher_name"`
	Contact       string         `json:"contact"`
	District      string         `json:"district"`
	BizType       string         `json:"biz_type"`
	Title         string         `json:"title"`
	Description   string         `json:"description"`
	Images        []string       `json:"images"`
	Latitude      float64        `json:"latitude"`
	Longitude     float64        `json:"longitude"`
	BudgetFen     int64          `json:"budget_fen"`
	Budget        int64          `json:"budget"` // 元（小程序发布表单），Create 时换算为分
	Deadline      string         `json:"deadline"`
	BizFields     map[string]any `json:"biz_fields"`
}

type DemandService struct {
	repo repository.DemandRepository
}

// ErrRoleNotAllowed 角色无权执行该操作（如非企业/个人发布需求）。
var ErrRoleNotAllowed = errors.New("only enterprise or individual users can publish demands")

func NewDemandService(r repository.DemandRepository) *DemandService {
	return &DemandService{repo: r}
}
func (s *DemandService) Create(ctx context.Context, a domain.Actor, in CreateDemandInput) (domain.Demand, error) {
	if a.Role != domain.RoleEnterprise && a.Role != domain.RoleIndividual {
		return domain.Demand{}, ErrRoleNotAllowed
	}
	if strings.TrimSpace(in.Title) == "" || strings.TrimSpace(in.Contact) == "" {
		return domain.Demand{}, errors.New("title and contact are required")
	}
	now := time.Now()
	bizType := domain.BizType(in.BizType)
	if bizType == "" {
		bizType = domain.BizOther
	}
	// 兼容小程序发布表单：提交 budget（元）时换算为分；显式 budget_fen 优先
	budgetFen := in.BudgetFen
	if budgetFen == 0 && in.Budget > 0 {
		budgetFen = in.Budget * 100
	}
	if budgetFen < 0 {
		return domain.Demand{}, errors.New("budget cannot be negative")
	}
	// 需求有效期（可选）：格式 YYYY-MM-DD / RFC3339；不得早于今天（发布时效校验）。
	deadline, err := validateDemandDeadline(in.Deadline)
	if err != nil {
		return domain.Demand{}, err
	}
	d := domain.Demand{ID: fmt.Sprintf("demand-%d-%d", now.UnixNano(), nextSeq()), PublisherID: a.ID, PublisherName: in.PublisherName, Contact: in.Contact, District: in.District, BizType: bizType, Title: in.Title, Description: in.Description, Images: in.Images, Latitude: in.Latitude, Longitude: in.Longitude, BudgetFen: budgetFen, Deadline: deadline, BizFields: in.BizFields, Status: domain.DemandPending, Version: 1, CreatedAt: now, UpdatedAt: now}
	slog.Info("demand created", "demand_id", d.ID, "publisher_id", a.ID, "biz_type", string(bizType))
	return s.repo.Create(ctx, d)
}

// validateDemandDeadline 校验需求截止日期：空串返回 ""（长期有效）；
// 否则接受 YYYY-MM-DD / "2006-01-02 15:04" / RFC3339，归一为 YYYY-MM-DD，
// 早于今天的日期拒绝（过期需求不允许发布）。
func validateDemandDeadline(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", nil
	}
	var t time.Time
	var err error
	for _, layout := range []string{time.RFC3339, "2006-01-02 15:04", "2006-01-02"} {
		if t, err = time.Parse(layout, s); err == nil {
			break
		}
	}
	if err != nil {
		return "", errors.New("无效的截止日期，格式应为 YYYY-MM-DD")
	}
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	if t.Before(today) {
		return "", errors.New("需求截止日期不能早于今天")
	}
	return t.Format("2006-01-02"), nil
}
func (s *DemandService) List(ctx context.Context, f repository.DemandFilter) ([]domain.Demand, error) {
	return s.repo.List(ctx, f)
}

// ListPage 公开语义 + 分页：透传 repo.ListPage（需求大厅高频路径 SQL 分页）。
func (s *DemandService) ListPage(ctx context.Context, f repository.DemandFilter, offset, limit int) ([]domain.Demand, int, error) {
	return s.repo.ListPage(ctx, f, offset, limit)
}

// ListAll 管理端全量（含待审核等全部状态）。
func (s *DemandService) ListAll(ctx context.Context, f repository.DemandFilter) ([]domain.Demand, error) {
	return s.repo.ListAll(ctx, f)
}

// Count 按 filter 统计需求条数（首页 stats 计数，避免全表拉取只为 len()）。
func (s *DemandService) Count(ctx context.Context, f repository.DemandFilter) (int, error) {
	return s.repo.Count(ctx, f)
}
func (s *DemandService) Search(ctx context.Context, q string) ([]domain.Demand, error) {
	return s.repo.Search(ctx, q)
}
func (s *DemandService) FindByID(ctx context.Context, id string) (domain.Demand, error) {
	return s.repo.FindByID(ctx, id)
}

// ListByPublisher 返回某发布者的全部需求（全状态），供"我的"页统计/查询。
func (s *DemandService) ListByPublisher(ctx context.Context, publisherID string) ([]domain.Demand, error) {
	return s.repo.ListByPublisher(ctx, publisherID)
}
func (s *DemandService) UpdateDraft(ctx context.Context, a domain.Actor, id, title, desc string) (domain.Demand, error) {
	d, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return domain.Demand{}, err
	}
	if d.PublisherID != a.ID {
		return domain.Demand{}, errors.New("only the publisher can edit")
	}
	if d.Status != domain.DemandPending {
		return domain.Demand{}, errors.New("only draft demands can be edited")
	}
	d.Title = title
	d.Description = desc
	d.UpdatedAt = time.Now()
	// 版本号自增由仓储层 Update 统一处理（乐观锁语义：WHERE version=$旧值）
	return s.repo.Update(ctx, d)
}

// Submit resubmits a rejected demand for admin review.
func (s *DemandService) Submit(ctx context.Context, a domain.Actor, id string) (domain.Demand, error) {
	d, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return domain.Demand{}, err
	}
	if d.PublisherID != a.ID {
		return domain.Demand{}, errors.New("only the publisher can submit")
	}
	if d.Status != domain.DemandRejected {
		return domain.Demand{}, fmt.Errorf("only rejected demands can be resubmitted, got %s", d.Status)
	}
	ok, _, err := s.repo.CompareAndSetStatus(ctx, id, domain.DemandRejected, domain.DemandPending)
	if err != nil {
		return domain.Demand{}, err
	}
	if !ok {
		return domain.Demand{}, errors.New("需求状态已变更，请刷新后重试")
	}
	d.Status = domain.DemandPending
	return d, nil
}

// Complete marks a published demand as done (publisher only, no bidding).
func (s *DemandService) Complete(ctx context.Context, a domain.Actor, id string) (domain.Demand, error) {
	d, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return domain.Demand{}, err
	}
	if d.PublisherID != a.ID {
		return domain.Demand{}, errors.New("only the publisher can complete the demand")
	}
	if d.Status != domain.DemandPublished {
		return domain.Demand{}, fmt.Errorf("only published demands can be completed, got %s", d.Status)
	}
	ok, _, err := s.repo.CompareAndSetStatus(ctx, id, domain.DemandPublished, domain.DemandCompleted)
	if err != nil {
		return domain.Demand{}, err
	}
	if !ok {
		return domain.Demand{}, errors.New("需求状态已变更，请刷新后重试")
	}
	d.Status = domain.DemandCompleted
	return d, nil
}

// Cancel withdraws a demand (pending or published) by the publisher.
func (s *DemandService) Cancel(ctx context.Context, a domain.Actor, id string) (domain.Demand, error) {
	d, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return domain.Demand{}, err
	}
	if d.PublisherID != a.ID {
		return domain.Demand{}, errors.New("only the publisher can cancel the demand")
	}
	if d.Status != domain.DemandPending && d.Status != domain.DemandPublished {
		return domain.Demand{}, fmt.Errorf("demand in status %s cannot be cancelled", d.Status)
	}
	// 两个合法前置状态：逐个 CAS，任一成功即完成；两者都失败说明状态竞态已变
	ok, _, err := s.repo.CompareAndSetStatus(ctx, id, domain.DemandPending, domain.DemandCancelled)
	if err != nil {
		return domain.Demand{}, err
	}
	if !ok {
		ok, _, err = s.repo.CompareAndSetStatus(ctx, id, domain.DemandPublished, domain.DemandCancelled)
		if err != nil {
			return domain.Demand{}, err
		}
	}
	if !ok {
		return domain.Demand{}, errors.New("需求状态已变更，请刷新后重试")
	}
	d.Status = domain.DemandCancelled
	return d, nil
}

func (s *DemandService) Review(ctx context.Context, a domain.Actor, id, action, reason string) (domain.Demand, error) {
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		return domain.Demand{}, errors.New("admin permission required")
	}
	if action == "reject" && reason == "" {
		return domain.Demand{}, errors.New("reason is required for rejection")
	}
	// 状态机前置：仅待审核（pending）需求可审；已公开的用 close、已驳回的先重提
	d, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return domain.Demand{}, err
	}
	if d.Status != domain.DemandPending {
		return domain.Demand{}, fmt.Errorf("只有待审核的需求可审核（当前状态 %s）", d.Status)
	}
	switch action {
	case "approve":
		ok, _, err := s.repo.CompareAndSetStatus(ctx, id, domain.DemandPending, domain.DemandPublished)
		if err != nil {
			return domain.Demand{}, err
		}
		if !ok {
			return domain.Demand{}, errors.New("需求状态已变更，请刷新后重试")
		}
		d.Status = domain.DemandPublished
		return d, nil
	case "reject":
		// 原子迁移 pending → rejected（与 approve/取消并发时后到者失败，防"驳回后被翻回已发布"）
		ok, _, err := s.repo.CompareAndSetStatus(ctx, id, domain.DemandPending, domain.DemandRejected)
		if err != nil {
			return domain.Demand{}, err
		}
		if !ok {
			return domain.Demand{}, errors.New("需求状态已变更，请刷新后重试")
		}
		// 驳回理由落库（BizFields），供发布者查看
		d2, err := s.repo.FindByID(ctx, id)
		if err != nil {
			return domain.Demand{}, err
		}
		if d2.BizFields == nil {
			d2.BizFields = map[string]any{}
		}
		d2.BizFields["reject_reason"] = reason
		return s.repo.Update(ctx, d2)
	case "supplement":
		// 前端未使用该动作，且 Demand 无"需补充"状态/字段支撑该流转：
		// 原实现只是把已是 pending 的需求再置 pending（恒 no-op）。
		// 直接返回明确错误，避免调用方误以为已生效。
		return domain.Demand{}, errors.New("不支持该动作：需求审核仅支持 approve / reject")
	default:
		return domain.Demand{}, fmt.Errorf("unknown review action: %s", action)
	}
}

// CloseByAdmin 管理端关闭已公开需求（发布者失联/虚假信息/线下已成交等场景）。
func (s *DemandService) CloseByAdmin(ctx context.Context, a domain.Actor, id, reason string) (domain.Demand, error) {
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		return domain.Demand{}, errors.New("admin permission required")
	}
	d, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return domain.Demand{}, err
	}
	if d.Status != domain.DemandPublished {
		return domain.Demand{}, errors.New("只有已公开的需求可以关闭")
	}
	// 原子迁移 published → cancelled，防止与发布者操作竞态把非在售需求关闭
	ok, _, err := s.repo.CompareAndSetStatus(ctx, id, domain.DemandPublished, domain.DemandCancelled)
	if err != nil {
		return domain.Demand{}, err
	}
	if !ok {
		return domain.Demand{}, errors.New("需求状态已变更，请刷新后重试")
	}
	d2, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return domain.Demand{}, err
	}
	if d2.BizFields == nil {
		d2.BizFields = map[string]any{}
	}
	d2.BizFields["reject_reason"] = reason
	return s.repo.Update(ctx, d2)
}

// Delete 管理端删除需求（仅已取消/已关闭需求可删，防止误删在审/在售数据）。
func (s *DemandService) Delete(ctx context.Context, a domain.Actor, id string) error {
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		return errors.New("admin permission required")
	}
	d, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if d.Status != domain.DemandCancelled && d.Status != domain.DemandRejected {
		return errors.New("只有已取消或已驳回的需求可以删除")
	}
	return s.repo.Delete(ctx, id)
}

// ToggleFavorite 收藏/取消收藏需求（登录用户可收藏任意公开需求）。
func (s *DemandService) ToggleFavorite(ctx context.Context, userID, demandID string, favorite bool) error {
	d, err := s.repo.FindByID(ctx, demandID)
	if err != nil {
		return err
	}
	// 门禁：仅在售需求可被收藏（与意向登记的 published 门禁一致，未发布需求不入收藏）
	if favorite && d.Status != domain.DemandPublished {
		return fmt.Errorf("only published demands can be favorited, got %s", d.Status)
	}
	if favorite {
		return s.repo.FavoriteDemand(ctx, userID, demandID)
	}
	return s.repo.UnfavoriteDemand(ctx, userID, demandID)
}

// ListFavoriteDemandIDs 当前用户已收藏的需求 ID 列表。
func (s *DemandService) ListFavoriteDemandIDs(ctx context.Context, userID string) ([]string, error) {
	return s.repo.ListFavoriteDemandIDs(ctx, userID)
}

// ListFavoriteDemands 当前用户收藏的完整需求列表（按收藏时间倒序）。
func (s *DemandService) ListFavoriteDemands(ctx context.Context, userID string) ([]domain.Demand, error) {
	return s.repo.ListFavoriteDemands(ctx, userID)
}

// SetOfflineAmount 登记线下成交金额（联系对接模式：平台撮合价值度量）。
// 仅已公开/已完成需求可登记；管理端补登或发布者完成时登记。
func (s *DemandService) SetOfflineAmount(ctx context.Context, a domain.Actor, id string, amountFen int64) (domain.Demand, error) {
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		return domain.Demand{}, errors.New("admin permission required")
	}
	if amountFen < 0 {
		return domain.Demand{}, errors.New("成交金额不能为负数")
	}
	d, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return domain.Demand{}, err
	}
	if d.Status != domain.DemandPublished && d.Status != domain.DemandCompleted {
		return domain.Demand{}, errors.New("只有已公开或已完成的需求可以登记成交金额")
	}
	d.OfflineAmountFen = amountFen
	d.UpdatedAt = time.Now()
	return s.repo.Update(ctx, d)
}

func (s *DemandService) Approve(ctx context.Context, a domain.Actor, id string) (domain.Demand, error) {
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		return domain.Demand{}, errors.New("admin permission required")
	}
	// CAS 原子迁移：仅 pending → published（与单条 Review 一致）。
	// 此前先读后 SetStatus（无旧状态谓词），并发取消/驳回会被覆盖回 published（需求"复活"）。
	ok, d, err := s.repo.CompareAndSetStatus(ctx, id, domain.DemandPending, domain.DemandPublished)
	if err != nil {
		return domain.Demand{}, fmt.Errorf("approve demand %s: %w", id, err)
	}
	if !ok {
		return domain.Demand{}, fmt.Errorf("只有待审核的需求可审核（状态已变更，请刷新后重试）")
	}
	return d, nil
}

type EnterpriseService struct {
	repo repository.EnterpriseRepository
}

func NewEnterpriseService(r repository.EnterpriseRepository) *EnterpriseService {
	return &EnterpriseService{r}
}
func (s *EnterpriseService) ListByStatus(ctx context.Context, status string, offset, limit int) ([]domain.Enterprise, int, error) {
	return s.repo.ListByStatus(ctx, status, offset, limit)
}
func (s *EnterpriseService) Create(ctx context.Context, e domain.Enterprise) (domain.Enterprise, error) {
	return s.repo.Create(ctx, e)
}
func (s *EnterpriseService) Update(ctx context.Context, id string, e domain.Enterprise) (domain.Enterprise, error) {
	return s.repo.Update(ctx, id, e)
}
func (s *EnterpriseService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
func (s *EnterpriseService) Pending(ctx context.Context, a domain.Actor) ([]domain.Enterprise, error) {
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		return nil, errors.New("association admin permission required")
	}
	return s.repo.Pending(ctx)
}
func (s *EnterpriseService) Search(ctx context.Context, q string) ([]domain.Enterprise, error) {
	return s.repo.Search(ctx, q)
}

// ListByOwner 按用户 ID 查企业（详情页发布者企业摘要用，无权限检查）
func (s *EnterpriseService) ListByOwner(ctx context.Context, userID string) ([]domain.Enterprise, error) {
	return s.repo.FindByOwner(ctx, userID)
}

type EmploymentService struct {
	repo repository.EmploymentRepository
}

func NewEmploymentService(r repository.EmploymentRepository) *EmploymentService {
	return &EmploymentService{repo: r}
}
func (s *EmploymentService) Create(ctx context.Context, a domain.Actor, v domain.EmploymentRequest) (domain.EmploymentRequest, error) {
	if a.Role != domain.RoleEnterprise {
		return v, errors.New("enterprise permission required")
	}
	now := time.Now()
	v.ID = nextID("employment")
	v.EnterpriseID = a.ID
	v.Status = domain.EmploymentPending
	v.Version = 1
	v.CreatedAt = now
	v.UpdatedAt = now
	return s.repo.Create(ctx, v)
}
func (s *EmploymentService) List(ctx context.Context, a domain.Actor, offset, limit int) ([]domain.EmploymentRequest, int, error) {
	if a.Role != domain.RolePlatformAdmin && a.Role != domain.RoleEnterprise {
		return nil, 0, errors.New("employment permission required")
	}
	if a.Role == domain.RolePlatformAdmin {
		return s.repo.ListAll(ctx, offset, limit)
	}
	return s.repo.ListByEnterprise(ctx, a.ID, offset, limit)
}

type ContractService struct {
	repo repository.ContractRepository
}

func NewContractService(r repository.ContractRepository) *ContractService {
	return &ContractService{repo: r}
}

// ContractTemplateService 提供合同模板列表（contract_templates 表）。
type ContractTemplateService struct {
	repo repository.ContractTemplateRepository
}

func NewContractTemplateService(r repository.ContractTemplateRepository) *ContractTemplateService {
	return &ContractTemplateService{repo: r}
}

func (s *ContractTemplateService) List(ctx context.Context) ([]domain.ContractTemplate, error) {
	return s.repo.List(ctx)
}
func (s *ContractService) Create(ctx context.Context, a domain.Actor, v domain.Contract) (domain.Contract, error) {
	if a.Role != domain.RolePlatformAdmin && a.Role != domain.RoleEnterprise {
		return v, errors.New("platform admin or enterprise permission required")
	}
	if a.Role == domain.RoleEnterprise {
		v.EnterpriseID = a.ID
	}
	now := time.Now()
	v.ID = nextID("contract")
	v.Status = domain.ContractDraft
	v.Version = 1
	v.CreatedAt = now
	v.UpdatedAt = now
	slog.Info("contract created", "contract_id", v.ID, "enterprise_id", v.EnterpriseID)
	return s.repo.Create(ctx, v)
}
func (s *ContractService) List(ctx context.Context, a domain.Actor, offset, limit int) ([]domain.Contract, int, error) {
	if a.Role != domain.RolePlatformAdmin && a.Role != domain.RoleEnterprise {
		return nil, 0, errors.New("contract permission required")
	}
	if a.Role == domain.RolePlatformAdmin {
		return s.repo.ListAll(ctx, offset, limit)
	}
	return s.repo.ListByEnterprise(ctx, a.ID, offset, limit)
}

func (s *ContractService) UpdateStatus(ctx context.Context, a domain.Actor, id string, newStatus domain.ContractStatus) (domain.Contract, error) {
	if a.Role != domain.RolePlatformAdmin && a.Role != domain.RoleEnterprise {
		return domain.Contract{}, errors.New("platform admin or enterprise permission required")
	}
	c, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return domain.Contract{}, err
	}
	// Enterprise users can only modify their own contracts
	if a.Role != domain.RolePlatformAdmin && c.EnterpriseID != a.ID {
		return domain.Contract{}, fmt.Errorf("you can only modify your own contracts")
	}
	// Basic state machine validation
	validTransitions := map[domain.ContractStatus][]domain.ContractStatus{
		domain.ContractDraft:   {domain.ContractSent},
		domain.ContractSent:    {domain.ContractSigning, domain.ContractVoided, domain.ContractExpired},
		domain.ContractSigning: {domain.ContractSigned, domain.ContractVoided, domain.ContractExpired},
		domain.ContractSigned:  {domain.ContractVoided},
	}
	allowed := validTransitions[c.Status]
	for _, st := range allowed {
		if st == newStatus {
			return s.repo.UpdateStatus(ctx, id, newStatus)
		}
	}
	return domain.Contract{}, fmt.Errorf("invalid contract status transition: %s -> %s", c.Status, newStatus)
}
