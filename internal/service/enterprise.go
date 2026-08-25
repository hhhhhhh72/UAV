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

type EnterpriseSvc struct {
	repo  repository.EnterpriseRepository
	users repository.UserRepository
}

func NewEnterpriseSvc(r repository.EnterpriseRepository, users repository.UserRepository) *EnterpriseSvc {
	return &EnterpriseSvc{repo: r, users: users}
}

type CreateEnterpriseInput struct {
	Name             string `json:"name"`
	CreditCode       string `json:"credit_code"`
	LegalPerson      string `json:"legal_person"`
	ContactPhone     string `json:"contact_phone"`
	IndustryCategory string `json:"industry_category"`
	Scale            string `json:"scale"`
	Address          string `json:"address"`
	Description      string `json:"description"`
	BusinessHours    string `json:"business_hours"`
	Logo             string `json:"logo"`
	CoverImage       string `json:"cover_image"`
	AccountName      string `json:"account_name"`
	LicenseURL       string `json:"license_url"`
	ContactPerson    string `json:"contact_person"`
	Email            string `json:"email"`
	FoundedAt        string `json:"founded_at"`
	CapabilityTags   string `json:"capability_tags"`
}

func (s *EnterpriseSvc) Create(ctx context.Context, a domain.Actor, in CreateEnterpriseInput) (domain.Enterprise, error) {
	// 重复入驻限制：同一用户已有未删除企业档案 → 拒绝再建
	//（此前可无限创建，一家账号可同时维护多家'已认证'商户档案，
	//  认证收益 enterprise 角色对全部档案生效，身份重复/资源占用）。
	if existed, err := s.repo.FindByOwner(ctx, a.ID); err == nil && len(existed) > 0 {
		return domain.Enterprise{}, fmt.Errorf("您已入驻成功，每用户仅可维护一家企业档案")
	}
	now := time.Now()
	e := domain.Enterprise{
		ID:               nextID("ent"),
		OwnerUserID:      a.ID,
		Name:             in.Name,
		CreditCode:       in.CreditCode,
		LegalPerson:      in.LegalPerson,
		ContactPhone:     in.ContactPhone,
		IndustryCategory: in.IndustryCategory,
		Scale:            in.Scale,
		Address:          in.Address,
		Description:      in.Description,
		BusinessHours:    in.BusinessHours,
		Logo:             in.Logo,
		CoverImage:       in.CoverImage,
		LicenseURL:       in.LicenseURL,
		AccountName:      in.AccountName,
		ContactPerson:    in.ContactPerson,
		Email:            in.Email,
		FoundedAt:        in.FoundedAt,
		CapabilityTags:   in.CapabilityTags,
		Status:           domain.EnterpriseDraft,
		Version:          1,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	slog.Info("enterprise created", "enterprise_id", e.ID, "name", e.Name)
	return s.repo.Create(ctx, e)
}

func (s *EnterpriseSvc) Update(ctx context.Context, a domain.Actor, id string, in CreateEnterpriseInput) (domain.Enterprise, error) {
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return domain.Enterprise{}, err
	}
	// 越权校验：属主本人、平台管理员、协会管理员可编辑。
	// 协会管理员负责企业审核（本职），管理端企业档案编辑入口对其放行。
	if existing.OwnerUserID != a.ID && a.Role != domain.RolePlatformAdmin && a.Role != domain.RoleAssociationAdmin {
		return domain.Enterprise{}, errors.New("only the owner can edit")
	}
	// 状态限制：
	// - 企业主（owner）：草稿/需补充/已驳回可编辑（编辑后状态不变，走提交流程；
	//   已驳回可改后重新提交——PRD FR-2.2 驳回重提闭环）；已认证（approved）也可编辑，
	//   编辑后自动回待审（自助信息更新闭环——此前 approved 企业主无任何修改送审路径）
	// - 管理员（platform_admin / association_admin）：任意状态可编辑；编辑已审核/已驳回/审核中企业时，
	//   状态回退到「待审核」（PRD FR-2.2：信息修改后需重新审核）
	if existing.OwnerUserID == a.ID && existing.Status != domain.EnterpriseDraft &&
		existing.Status != domain.EnterpriseSupplementRequired && existing.Status != domain.EnterpriseRejected &&
		existing.Status != domain.EnterpriseApproved {
		return domain.Enterprise{}, fmt.Errorf("cannot edit enterprise in %s status", existing.Status)
	}
	isAdminEdit := a.Role == domain.RolePlatformAdmin || a.Role == domain.RoleAssociationAdmin
	wasApprovedOrReviewed := existing.Status == domain.EnterpriseApproved ||
		existing.Status == domain.EnterpriseRejected ||
		existing.Status == domain.EnterpriseSubmitted
	// PATCH 语义防清空：关键字段（识别/资质/联系方式）仅在传入非空时覆盖——
	// 此前全量替换，客户端局部提交（如仅改名称）会把 account_name/credit_code/
	// contact_phone 等写成空串（PII 与资质信息被静默抹掉）。
	if in.Name != "" {
		existing.Name = in.Name
	}
	if in.CreditCode != "" {
		existing.CreditCode = in.CreditCode
	}
	if in.LegalPerson != "" {
		existing.LegalPerson = in.LegalPerson
	}
	if in.ContactPhone != "" {
		existing.ContactPhone = in.ContactPhone
	}
	if in.AccountName != "" {
		existing.AccountName = in.AccountName
	}
	if in.Email != "" {
		existing.Email = in.Email
	}
	if in.ContactPerson != "" {
		existing.ContactPerson = in.ContactPerson
	}
	if in.IndustryCategory != "" {
		existing.IndustryCategory = in.IndustryCategory
	}
	if in.Address != "" {
		existing.Address = in.Address
	}
	if in.Scale != "" {
		existing.Scale = in.Scale
	}
	if in.Description != "" {
		existing.Description = in.Description
	}
	if in.BusinessHours != "" {
		existing.BusinessHours = in.BusinessHours
	}
	if in.Logo != "" {
		existing.Logo = in.Logo
	}
	if in.CoverImage != "" {
		existing.CoverImage = in.CoverImage
	}
	if in.LicenseURL != "" {
		existing.LicenseURL = in.LicenseURL
	}
	if len(in.CapabilityTags) > 0 {
		existing.CapabilityTags = in.CapabilityTags
	}
	if in.FoundedAt != "" {
		existing.FoundedAt = in.FoundedAt
	}
	existing.UpdatedAt = time.Now()
	if (isAdminEdit && wasApprovedOrReviewed) ||
		(existing.OwnerUserID == a.ID && existing.Status == domain.EnterpriseApproved) {
		// 管理端编辑已审企业 / 企业主更新已认证档案 → 重新进入审核队列（PRD FR-2.2）
		existing.Status = domain.EnterpriseSubmitted
		existing.ReviewComment = ""
	}
	return s.repo.Update(ctx, id, existing)
}

func (s *EnterpriseSvc) Submit(ctx context.Context, a domain.Actor, id string) (domain.Enterprise, error) {
	e, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return domain.Enterprise{}, err
	}
	if e.OwnerUserID != a.ID {
		return domain.Enterprise{}, errors.New("only the owner can submit")
	}
	if e.Status != domain.EnterpriseDraft && e.Status != domain.EnterpriseSupplementRequired && e.Status != domain.EnterpriseRejected {
		return domain.Enterprise{}, fmt.Errorf("cannot submit enterprise in %s status", e.Status)
	}
	e.Status = domain.EnterpriseSubmitted
	e.UpdatedAt = time.Now()
	return s.repo.Update(ctx, id, e)
}

func (s *EnterpriseSvc) Review(ctx context.Context, a domain.Actor, id, action, reason string) (domain.Enterprise, error) {
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		return domain.Enterprise{}, errors.New("admin permission required")
	}
	if (action == "reject" || action == "rejected") && reason == "" {
		return domain.Enterprise{}, errors.New("reason is required for rejection")
	}
	e, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return domain.Enterprise{}, err
	}
	// 状态机前置：仅已提交/需补充的企业可审（防对草稿/已通过/已驳回重复翻转）
	if e.Status != domain.EnterpriseSubmitted && e.Status != domain.EnterpriseSupplementRequired {
		return domain.Enterprise{}, fmt.Errorf("只有已提交的企业可审核（当前状态 %s）", e.Status)
	}
	var newStatus domain.EnterpriseStatus
	// 兼容两种写法：动词 approve/reject/supplement 与过去式 approved/rejected
	switch action {
	case "approve", "approved":
		newStatus = domain.EnterpriseApproved
	case "reject", "rejected":
		newStatus = domain.EnterpriseRejected
	case "supplement", "supplement_required":
		newStatus = domain.EnterpriseSupplementRequired
	default:
		return domain.Enterprise{}, fmt.Errorf("unknown review action: %s", action)
	}
	e.Status = newStatus
	// 审核意见持久化：驳回/需补充必须附原因，通过时清空历史意见
	e.ReviewComment = reason
	e.UpdatedAt = time.Now()
	ent, err := s.repo.Update(ctx, id, e)
	if err != nil {
		return domain.Enterprise{}, err
	}
	// 审核通过：owner 用户升级为企业角色，否则用户仍是个体、无法获得企业权益（发招聘/合同等）
	if newStatus == domain.EnterpriseApproved {
		if err := s.users.UpdateRole(ctx, e.OwnerUserID, domain.RoleEnterprise); err != nil {
			// 升级失败回滚已批准状态（按最新版本 Read-Update 回 submitted），
			// 防"档案已认证但属主无企业权限"的卡死半态（此前仅 warn 不回滚）。
			slog.Warn("upgrade owner role failed, reverting approval", "user_id", e.OwnerUserID, "error", err)
			if latest, ferr := s.repo.FindByID(ctx, id); ferr == nil && latest.Status == domain.EnterpriseApproved {
				latest.Status = domain.EnterpriseSubmitted
				latest.ReviewComment = ""
				latest.UpdatedAt = time.Now()
				if _, uerr := s.repo.Update(ctx, id, latest); uerr != nil {
					slog.Warn("revert enterprise approval failed", "id", id, "error", uerr)
				}
			}
			return domain.Enterprise{}, fmt.Errorf("升级企业角色失败：%w", err)
		}
	}
	return ent, nil
}

func (s *EnterpriseSvc) ListByStatus(ctx context.Context, a domain.Actor, status string, offset, limit int) ([]domain.Enterprise, int, error) {
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		return nil, 0, errors.New("admin permission required")
	}
	return s.repo.ListByStatus(ctx, status, offset, limit)
}

func (s *EnterpriseSvc) FindByID(ctx context.Context, id string) (domain.Enterprise, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *EnterpriseSvc) ListMine(ctx context.Context, a domain.Actor) ([]domain.Enterprise, error) {
	return s.repo.FindByOwner(ctx, a.ID)
}

func (s *EnterpriseSvc) Search(ctx context.Context, a domain.Actor, q string) ([]domain.Enterprise, error) {
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		return nil, errors.New("admin permission required")
	}
	return s.repo.Search(ctx, q)
}

// AttachDocument links an uploaded file to an enterprise (business license, ID card, ...).
// Only the enterprise owner or admins may attach.
func (s *EnterpriseSvc) AttachDocument(ctx context.Context, a domain.Actor, enterpriseID, fileID, documentType string) (domain.EnterpriseDocument, error) {
	e, err := s.repo.FindByID(ctx, enterpriseID)
	if err != nil {
		return domain.EnterpriseDocument{}, err
	}
	if e.OwnerUserID != a.ID && a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		return domain.EnterpriseDocument{}, errors.New("permission denied")
	}
	now := time.Now()
	doc := domain.EnterpriseDocument{ID: nextID("edoc"), EnterpriseID: enterpriseID,
		FileID: fileID, DocumentType: documentType, ReviewStatus: "pending", CreatedAt: now}
	return s.repo.AddDocument(ctx, doc)
}

// ListDocuments returns documents of an enterprise for the owner or admins.
func (s *EnterpriseSvc) ListDocuments(ctx context.Context, a domain.Actor, enterpriseID string) ([]domain.EnterpriseDocument, error) {
	e, err := s.repo.FindByID(ctx, enterpriseID)
	if err != nil {
		return nil, err
	}
	if e.OwnerUserID != a.ID && a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		return nil, errors.New("permission denied")
	}
	return s.repo.ListDocuments(ctx, enterpriseID)
}
