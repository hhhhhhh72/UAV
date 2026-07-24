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
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
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
	BizFields     map[string]any `json:"biz_fields"`
}

type DemandService struct {
	repo    repository.DemandRepository
	bidRepo repository.BidRepository
}

func NewDemandService(r repository.DemandRepository, br repository.BidRepository) *DemandService {
	return &DemandService{repo: r, bidRepo: br}
}
func (s *DemandService) Create(a domain.Actor, in CreateDemandInput) (domain.Demand, error) {
	if a.Role != domain.RoleEnterprise && a.Role != domain.RoleIndividual {
		return domain.Demand{}, errors.New("only enterprise or individual users can publish demands")
	}
	if strings.TrimSpace(in.Title) == "" || strings.TrimSpace(in.Contact) == "" {
		return domain.Demand{}, errors.New("title and contact are required")
	}
	now := time.Now()
	bizType := domain.BizType(in.BizType)
	if bizType == "" {
		bizType = domain.BizOther
	}
	d := domain.Demand{ID: fmt.Sprintf("demand-%d", now.UnixNano()), PublisherID: a.ID, PublisherName: in.PublisherName, Contact: in.Contact, District: in.District, BizType: bizType, Title: in.Title, Description: in.Description, Images: in.Images, Latitude: in.Latitude, Longitude: in.Longitude, BudgetFen: in.BudgetFen, BizFields: in.BizFields, Status: domain.DemandPending, Version: 1, CreatedAt: now, UpdatedAt: now}
	slog.Info("demand created", "demand_id", d.ID, "publisher_id", a.ID, "biz_type", string(bizType))
	return s.repo.Create(d)
}
func (s *DemandService) List(f repository.DemandFilter) ([]domain.Demand, error) {
	return s.repo.List(f)
}
func (s *DemandService) Search(q string) ([]domain.Demand, error) { return s.repo.Search(q) }
func (s *DemandService) FindByID(id string) (domain.Demand, error) { return s.repo.FindByID(id) }
func (s *DemandService) ListBidsByDemand(demandID string) ([]domain.DemandBid, error) {
	return s.bidRepo.ListByDemand(demandID)
}
func (s *DemandService) ListBidsByBidder(bidderID string) ([]domain.DemandBid, error) {
	return s.bidRepo.ListByBidder(bidderID)
}
func (s *DemandService) UpdateDraft(a domain.Actor, id, title, desc string) (domain.Demand, error) {
	d, err := s.repo.FindByID(id)
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
	d.Version++
	return s.repo.Update(d)
}

func (s *DemandService) Submit(a domain.Actor, id string) (domain.Demand, error) {
	d, err := s.repo.SetStatus(id, domain.DemandPending)
	if err != nil {
		return domain.Demand{}, err
	}
	if d.PublisherID != a.ID {
		return domain.Demand{}, errors.New("only the publisher can submit")
	}
	return s.repo.SetStatus(id, domain.DemandPublished) // auto-publish for mini program flow
}

func (s *DemandService) Review(a domain.Actor, id, action, reason string) (domain.Demand, error) {
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		return domain.Demand{}, errors.New("admin permission required")
	}
	if action == "reject" && reason == "" {
		return domain.Demand{}, errors.New("reason is required for rejection")
	}
	switch action {
	case "approve":
		return s.repo.SetStatus(id, domain.DemandPublished)
	case "reject":
		return s.repo.SetStatus(id, domain.DemandRejected)
	case "supplement":
		return s.repo.SetStatus(id, domain.DemandPending)
	default:
		return domain.Demand{}, fmt.Errorf("unknown review action: %s", action)
	}
}

func (s *DemandService) Approve(a domain.Actor, id string) (domain.Demand, error) {
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		return domain.Demand{}, errors.New("admin permission required")
	}
	d, err := s.repo.SetStatus(id, domain.DemandPublished)
	if err != nil {
		return domain.Demand{}, fmt.Errorf("approve demand %s: %w", id, err)
	}
	return d, nil
}

func (s *DemandService) CreateBid(a domain.Actor, demandID string, amountFen int64, proposal string) (domain.DemandBid, error) {
	if a.Role != domain.RoleEnterprise && a.Role != domain.RoleIndividual {
		return domain.DemandBid{}, errors.New("only enterprise/individual can bid")
	}
	d, err := s.repo.FindByID(demandID)
	if err != nil {
		return domain.DemandBid{}, fmt.Errorf("demand not found: %w", err)
	}
	if d.Status != domain.DemandPublished {
		return domain.DemandBid{}, fmt.Errorf("cannot bid on demand in status %q", d.Status)
	}
	if d.PublisherID == a.ID {
		return domain.DemandBid{}, errors.New("cannot bid on your own demand")
	}
	now := time.Now()
	bid := domain.DemandBid{
		ID:         fmt.Sprintf("bid-%d", now.UnixNano()),
		DemandID:   demandID,
		BidderID:   a.ID,
		BidderName: a.ID,
		AmountFen:  amountFen,
		Proposal:   proposal,
		Status:     "pending",
		Version:    1,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	saved, err := s.bidRepo.Create(bid)
	if err != nil {
		slog.Error("failed to persist bid", "error", err, "demand_id", demandID)
		return domain.DemandBid{}, fmt.Errorf("persist bid: %w", err)
	}
	slog.Info("bid created", "bid_id", saved.ID, "demand_id", demandID, "bidder_id", a.ID)
	return saved, nil
}

func (s *DemandService) SelectBid(a domain.Actor, demandID, bidID string) (domain.Demand, error) {
	d, err := s.repo.FindByID(demandID)
	if err != nil {
		return domain.Demand{}, fmt.Errorf("demand not found: %w", err)
	}
	if d.PublisherID != a.ID {
		return domain.Demand{}, errors.New("only the publisher can select a bid")
	}
	if d.Status != domain.DemandPublished {
		return domain.Demand{}, fmt.Errorf("cannot select bid for demand in status %q", d.Status)
	}
	bid, err := s.bidRepo.FindByID(bidID)
	if err != nil {
		return domain.Demand{}, fmt.Errorf("bid not found: %w", err)
	}
	if bid.DemandID != demandID {
		return domain.Demand{}, fmt.Errorf("bid %s does not belong to demand %s", bidID, demandID)
	}
	if _, err := s.bidRepo.UpdateStatus(bidID, "accepted"); err != nil {
		slog.Error("failed to accept bid", "error", err, "bid_id", bidID)
		return domain.Demand{}, fmt.Errorf("accept bid: %w", err)
	}
	swapped, _, err := s.repo.CompareAndSetStatus(demandID, domain.DemandPublished, domain.DemandMatched)
	if err != nil {
		return domain.Demand{}, fmt.Errorf("match demand: %w", err)
	}
	if !swapped {
		// Another concurrent request already matched this demand — rollback bid acceptance
		s.bidRepo.UpdateStatus(bidID, "pending")
		return domain.Demand{}, fmt.Errorf("demand %s is no longer published, bid selection lost race", demandID)
	}
	slog.Info("bid accepted, demand matched", "demand_id", demandID, "bid_id", bidID)
	d, _ = s.repo.FindByID(demandID)
	return d, nil
}

// ConfirmComplete tracks partial confirmations for dual-confirm flow.
var demandConfirms = struct {
	m map[string]map[string]bool // demandID -> userID -> true
	mu sync.Mutex
}{m: make(map[string]map[string]bool)}

func (s *DemandService) ConfirmComplete(a domain.Actor, id string) (domain.Demand, bool, error) {
	d, err := s.repo.FindByID(id)
	if err != nil {
		return domain.Demand{}, false, err
	}
	if d.Status != domain.DemandMatched {
		return domain.Demand{}, false, fmt.Errorf("demand must be in matched status, got %s", d.Status)
	}
	// Only publisher or matched bidder can confirm.
	demandConfirms.mu.Lock()
	if demandConfirms.m[id] == nil {
		demandConfirms.m[id] = make(map[string]bool)
	}
	demandConfirms.m[id][a.ID] = true
	count := len(demandConfirms.m[id])
	demandConfirms.mu.Unlock()

	if count >= 2 {
		d, err = s.repo.SetStatus(id, domain.DemandCompleted)
		return d, true, err
	}
	return d, false, nil
}

func (s *DemandService) Dispute(a domain.Actor, id, reason string) (domain.Demand, error) {
	d, err := s.repo.FindByID(id)
	if err != nil {
		return domain.Demand{}, err
	}
	if d.PublisherID != a.ID {
		return domain.Demand{}, errors.New("only the publisher can raise a dispute")
	}
	if d.Status != domain.DemandPublished && d.Status != domain.DemandMatched && d.Status != domain.DemandCompleted {
		return domain.Demand{}, fmt.Errorf("cannot dispute demand in status %s", d.Status)
	}
	_ = reason
	return s.repo.SetStatus(id, domain.DemandPending)
}

type EnterpriseService struct {
	repo repository.EnterpriseRepository
}

func NewEnterpriseService(r repository.EnterpriseRepository) *EnterpriseService {
	return &EnterpriseService{r}
}
func (s *EnterpriseService) Pending(a domain.Actor) ([]domain.Enterprise, error) {
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		return nil, errors.New("association admin permission required")
	}
	return s.repo.Pending()
}
func (s *EnterpriseService) Search(q string) ([]domain.Enterprise, error) { return s.repo.Search(q) }

type EmploymentService struct {
	repo repository.EmploymentRepository
}

func NewEmploymentService(r repository.EmploymentRepository) *EmploymentService {
	return &EmploymentService{repo: r}
}
func (s *EmploymentService) Create(a domain.Actor, v domain.EmploymentRequest) (domain.EmploymentRequest, error) {
	if a.Role != domain.RoleEnterprise {
		return v, errors.New("enterprise permission required")
	}
	now := time.Now()
	v.ID = fmt.Sprintf("employment-%d", now.UnixNano())
	v.EnterpriseID = a.ID
	v.Status = domain.EmploymentPending
	v.Version = 1
	v.CreatedAt = now
	v.UpdatedAt = now
	return s.repo.Create(v)
}
func (s *EmploymentService) List(a domain.Actor, offset, limit int) ([]domain.EmploymentRequest, int, error) {
	if a.Role != domain.RolePlatformAdmin && a.Role != domain.RoleEnterprise {
		return nil, 0, errors.New("employment permission required")
	}
	if a.Role == domain.RolePlatformAdmin {
		return s.repo.ListAll(offset, limit)
	}
	return s.repo.ListByEnterprise(a.ID, offset, limit)
}

type ContractService struct {
	repo repository.ContractRepository
}

func NewContractService(r repository.ContractRepository) *ContractService {
	return &ContractService{repo: r}
}
func (s *ContractService) Create(a domain.Actor, v domain.Contract) (domain.Contract, error) {
	if a.Role != domain.RolePlatformAdmin && a.Role != domain.RoleEnterprise {
		return v, errors.New("platform admin or enterprise permission required")
	}
	if a.Role == domain.RoleEnterprise {
		v.EnterpriseID = a.ID
	}
	now := time.Now()
	v.ID = fmt.Sprintf("contract-%d", now.UnixNano())
	v.Status = domain.ContractDraft
	v.Version = 1
	v.CreatedAt = now
	v.UpdatedAt = now
	slog.Info("contract created", "contract_id", v.ID, "enterprise_id", v.EnterpriseID)
	return s.repo.Create(v)
}
func (s *ContractService) List(a domain.Actor, offset, limit int) ([]domain.Contract, int, error) {
	if a.Role != domain.RolePlatformAdmin && a.Role != domain.RoleEnterprise {
		return nil, 0, errors.New("contract permission required")
	}
	if a.Role == domain.RolePlatformAdmin {
		return s.repo.ListAll(offset, limit)
	}
	return s.repo.ListByEnterprise(a.ID, offset, limit)
}

func (s *ContractService) UpdateStatus(a domain.Actor, id string, newStatus domain.ContractStatus) (domain.Contract, error) {
	if a.Role != domain.RolePlatformAdmin && a.Role != domain.RoleEnterprise {
		return domain.Contract{}, errors.New("platform admin or enterprise permission required")
	}
	c, err := s.repo.FindByID(id)
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
			return s.repo.UpdateStatus(id, newStatus)
		}
	}
	return domain.Contract{}, fmt.Errorf("invalid contract status transition: %s -> %s", c.Status, newStatus)
}
