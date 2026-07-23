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
	repo repository.EnterpriseRepository
}

func NewEnterpriseSvc(r repository.EnterpriseRepository) *EnterpriseSvc {
	return &EnterpriseSvc{repo: r}
}

type CreateEnterpriseInput struct {
	Name        string `json:"name"`
	AccountName string `json:"account_name"`
	LicenseURL  string `json:"license_url"`
}

func (s *EnterpriseSvc) Create(a domain.Actor, in CreateEnterpriseInput) (domain.Enterprise, error) {
	now := time.Now()
	e := domain.Enterprise{
		ID:          fmt.Sprintf("ent-%d", now.UnixNano()),
		OwnerUserID: a.ID,
		Name:        in.Name,
		LicenseURL:  in.LicenseURL,
		AccountName: in.AccountName,
		Status:      domain.EnterpriseDraft,
		Version:     1,
		CreatedAt:   now,
		UpdatedAt:   now,
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
	if existing.Status != domain.EnterpriseDraft && existing.Status != domain.EnterpriseSupplementRequired {
		return domain.Enterprise{}, fmt.Errorf("cannot edit enterprise in %s status", existing.Status)
	}
	existing.Name = in.Name
	existing.LicenseURL = in.LicenseURL
	existing.AccountName = in.AccountName
	existing.UpdatedAt = time.Now()
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
	if action == "reject" && reason == "" {
		return domain.Enterprise{}, errors.New("reason is required for rejection")
	}
	e, err := s.repo.FindByID(id)
	if err != nil {
		return domain.Enterprise{}, err
	}
	var newStatus domain.EnterpriseStatus
	switch action {
	case "approve":
		newStatus = domain.EnterpriseApproved
	case "reject":
		newStatus = domain.EnterpriseRejected
	case "supplement":
		newStatus = domain.EnterpriseSupplementRequired
	default:
		return domain.Enterprise{}, fmt.Errorf("unknown review action: %s", action)
	}
	e.Status = newStatus
	e.UpdatedAt = time.Now()
	return s.repo.Update(id, e)
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
