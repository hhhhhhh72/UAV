package service

import (
	"context"
	"errors"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)

type InsuranceService struct {
	policyRepo  repository.PolicyRepository
	inspectRepo repository.InspectionRepository
}

func NewInsuranceService(pr repository.PolicyRepository, ir repository.InspectionRepository) *InsuranceService {
	return &InsuranceService{policyRepo: pr, inspectRepo: ir}
}

func (s *InsuranceService) CreatePolicy(ctx context.Context, a domain.Actor, droneModel, droneSN, policyType string, premiumFen, coverageFen int64, start, end time.Time) (domain.InsurancePolicy, error) {
	now := time.Now()
	p := domain.InsurancePolicy{ID: nextID("policy"), UserID: a.ID, DroneModel: droneModel,
		DroneSN: droneSN, PolicyType: policyType, PremiumFen: premiumFen, CoverageFen: coverageFen,
		StartDate: start, EndDate: end, Insurer: "default", Status: "active", Version: 1, CreatedAt: now, UpdatedAt: now}
	return s.policyRepo.Create(ctx, p)
}

func (s *InsuranceService) ListMyPolicies(ctx context.Context, a domain.Actor) ([]domain.InsurancePolicy, error) {
	return s.policyRepo.ListByUser(ctx, a.ID)
}

func (s *InsuranceService) ListAllPolicies(ctx context.Context, offset, limit int) ([]domain.InsurancePolicy, int, error) {
	return s.policyRepo.ListAll(ctx, offset, limit)
}

func (s *InsuranceService) CreateInspection(ctx context.Context, a domain.Actor, droneModel, droneSN string, inspectDate, expireDate time.Time) (domain.AnnualInspection, error) {
	now := time.Now()
	i := domain.AnnualInspection{ID: nextID("inspect"), UserID: a.ID, DroneModel: droneModel,
		DroneSN: droneSN, InspectDate: inspectDate, ExpireDate: expireDate, Status: "pending", Version: 1, CreatedAt: now, UpdatedAt: now}
	return s.inspectRepo.Create(ctx, i)
}

func (s *InsuranceService) ListAllInspections(ctx context.Context) ([]domain.AnnualInspection, error) {
	return s.inspectRepo.ListAll(ctx)
}
func (s *InsuranceService) ListMyInspections(ctx context.Context, a domain.Actor) ([]domain.AnnualInspection, error) {
	return s.inspectRepo.ListByUser(ctx, a.ID)
}

type FinanceService struct {
	loanRepo repository.LoanRepository
}

func NewFinanceService(lr repository.LoanRepository) *FinanceService {
	return &FinanceService{loanRepo: lr}
}

func (s *FinanceService) ApplyLoan(ctx context.Context, a domain.Actor, amountFen int64, termMonths int, purpose string) (domain.LoanApplication, error) {
	if amountFen <= 0 {
		return domain.LoanApplication{}, errors.New("loan amount must be positive")
	}
	if termMonths <= 0 {
		return domain.LoanApplication{}, errors.New("loan term must be positive")
	}
	now := time.Now()
	l := domain.LoanApplication{ID: nextID("loan"), UserID: a.ID, AmountFen: amountFen,
		TermMonths: termMonths, Purpose: purpose, Status: "submitted", Version: 1, CreatedAt: now, UpdatedAt: now}
	return s.loanRepo.Create(ctx, l)
}

func (s *FinanceService) ListMyLoans(ctx context.Context, a domain.Actor) ([]domain.LoanApplication, error) {
	return s.loanRepo.ListByUser(ctx, a.ID)
}

func (s *FinanceService) ListAllLoans(ctx context.Context, offset, limit int) ([]domain.LoanApplication, int, error) {
	return s.loanRepo.ListAll(ctx, offset, limit)
}
