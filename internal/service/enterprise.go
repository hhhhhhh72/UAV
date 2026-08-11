package service

import (
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

func (s *EnterpriseSvc) Create(a domain.Actor, in CreateEnterpriseInput) (domain.Enterprise, error) {
	now := time.Now()
	e := domain.Enterprise{
		ID:               fmt.Sprintf("ent-%d", now.UnixNano()),
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
	return s.repo.Create(e)
}

func (s *EnterpriseSvc) Update(a domain.Actor, id string, in CreateEnterpriseInput) (domain.Enterprise, error) {
	existing, err := s.repo.FindByID(id)
	if err != nil {
		return domain.Enterprise{}, err
	}
	if existing.OwnerUserID != a.ID && a.Role != domain.RolePlatformAdmin {
		return domain.Enterprise{}, errors.New("only the owner can edit")
	}
	// 状态限制：
	// - 企业主（owner）：仅草稿/需补充可编辑（编辑后状态不变，走提交流程）
	// - 管理员（platform_admin）：任意状态可编辑；编辑已审核/已驳回/审核中企业时，
	//   状态回退到「待审核」（PRD FR-2.2：信息修改后需重新审核）
	if existing.OwnerUserID == a.ID && existing.Status != domain.EnterpriseDraft && existing.Status != domain.EnterpriseSupplementRequired {
		return domain.Enterprise{}, fmt.Errorf("cannot edit enterprise in %s status", existing.Status)
	}
	isAdminEdit := a.Role == domain.RolePlatformAdmin
	wasApprovedOrReviewed := existing.Status == domain.EnterpriseApproved ||
		existing.Status == domain.EnterpriseRejected ||
		existing.Status == domain.EnterpriseSubmitted
	existing.Name = in.Name
	existing.CreditCode = in.CreditCode
	existing.LegalPerson = in.LegalPerson
	existing.ContactPhone = in.ContactPhone
	existing.IndustryCategory = in.IndustryCategory
	existing.Scale = in.Scale
	existing.Address = in.Address
	existing.Description = in.Description
	existing.BusinessHours = in.BusinessHours
	existing.Logo = in.Logo
	existing.CoverImage = in.CoverImage
	existing.LicenseURL = in.LicenseURL
	existing.AccountName = in.AccountName
	existing.ContactPerson = in.ContactPerson
	existing.Email = in.Email
	existing.FoundedAt = in.FoundedAt
	existing.CapabilityTags = in.CapabilityTags
	existing.UpdatedAt = time.Now()
	if isAdminEdit && wasApprovedOrReviewed {
		// 管理员编辑已审企业 → 重新进入审核队列（PRD FR-2.2）
		existing.Status = domain.EnterpriseSubmitted
		existing.ReviewComment = ""
	}
	return s.repo.Update(id, existing)
}

func (s *EnterpriseSvc) Submit(a domain.Actor, id string) (domain.Enterprise, error) {
	e, err := s.repo.FindByID(id)
	if err != nil {
		return domain.Enterprise{}, err
	}
	if e.OwnerUserID != a.ID {
		return domain.Enterprise{}, errors.New("only the owner can submit")
	}
	if e.Status != domain.EnterpriseDraft && e.Status != domain.EnterpriseSupplementRequired {
		return domain.Enterprise{}, fmt.Errorf("cannot submit enterprise in %s status", e.Status)
	}
	e.Status = domain.EnterpriseSubmitted
	e.UpdatedAt = time.Now()
	return s.repo.Update(id, e)
}

func (s *EnterpriseSvc) Review(a domain.Actor, id, action, reason string) (domain.Enterprise, error) {
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		return domain.Enterprise{}, errors.New("admin permission required")
	}
	if (action == "reject" || action == "rejected") && reason == "" {
		return domain.Enterprise{}, errors.New("reason is required for rejection")
	}
	e, err := s.repo.FindByID(id)
	if err != nil {
		return domain.Enterprise{}, err
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
	ent, err := s.repo.Update(id, e)
	if err != nil {
		return domain.Enterprise{}, err
	}
	// 审核通过：owner 用户升级为企业角色，否则用户仍是个体、无法获得企业权益（发招聘/合同等）
	if newStatus == domain.EnterpriseApproved {
		if err := s.users.UpdateRole(e.OwnerUserID, domain.RoleEnterprise); err != nil {
			slog.Warn("upgrade owner role failed", "user_id", e.OwnerUserID, "error", err)
		}
	}
	return ent, nil
}

func (s *EnterpriseSvc) ListByStatus(a domain.Actor, status string, offset, limit int) ([]domain.Enterprise, int, error) {
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		return nil, 0, errors.New("admin permission required")
	}
	return s.repo.ListByStatus(status, offset, limit)
}

func (s *EnterpriseSvc) FindByID(id string) (domain.Enterprise, error) {
	return s.repo.FindByID(id)
}

func (s *EnterpriseSvc) ListMine(a domain.Actor) ([]domain.Enterprise, error) {
	return s.repo.FindByOwner(a.ID)
}

func (s *EnterpriseSvc) Search(a domain.Actor, q string) ([]domain.Enterprise, error) {
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		return nil, errors.New("admin permission required")
	}
	return s.repo.Search(q)
}
