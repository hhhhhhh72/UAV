package service

import (
	"fmt"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)

type InsuranceService struct {
	policyRepo    repository.PolicyRepository
	inspectRepo   repository.InspectionRepository
}

func NewInsuranceService(pr repository.PolicyRepository, ir repository.InspectionRepository) *InsuranceService {
	return &InsuranceService{policyRepo: pr, inspectRepo: ir}
}

func (s *InsuranceService) CreatePolicy(a domain.Actor, droneModel, droneSN, policyType string, premiumFen, coverageFen int64, start, end time.Time) (domain.InsurancePolicy, error) {
	now := time.Now()
	p := domain.InsurancePolicy{ID: fmt.Sprintf("policy-%d", now.UnixNano()), UserID: a.ID, DroneModel: droneModel,
		DroneSN: droneSN, PolicyType: policyType, PremiumFen: premiumFen, CoverageFen: coverageFen,
		StartDate: start, EndDate: end, Insurer: "default", Status: "active", Version: 1, CreatedAt: now, UpdatedAt: now}
	return s.policyRepo.Create(p)
}

func (s *InsuranceService) ListMyPolicies(a domain.Actor) ([]domain.InsurancePolicy, error) {
	return s.policyRepo.ListByUser(a.ID)
}

func (s *InsuranceService) ListAllPolicies(offset, limit int) ([]domain.InsurancePolicy, int, error) {
	return s.policyRepo.ListAll(offset, limit)
}

func (s *InsuranceService) CreateInspection(a domain.Actor, droneModel, droneSN string, inspectDate, expireDate time.Time) (domain.AnnualInspection, error) {
	now := time.Now()
	i := domain.AnnualInspection{ID: fmt.Sprintf("inspect-%d", now.UnixNano()), UserID: a.ID, DroneModel: droneModel,
		DroneSN: droneSN, InspectDate: inspectDate, ExpireDate: expireDate, Status: "pending", Version: 1, CreatedAt: now, UpdatedAt: now}
	return s.inspectRepo.Create(i)
}

func (s *InsuranceService) ListAllInspections() ([]domain.AnnualInspection, error) {
	return s.inspectRepo.ListAll()
}
func (s *InsuranceService) ListMyInspections(a domain.Actor) ([]domain.AnnualInspection, error) {
	return s.inspectRepo.ListByUser(a.ID)
}

type FinanceService struct {
	loanRepo repository.LoanRepository
}

func NewFinanceService(lr repository.LoanRepository) *FinanceService {
	return &FinanceService{loanRepo: lr}
}

func (s *FinanceService) ApplyLoan(a domain.Actor, amountFen int64, termMonths int, purpose string) (domain.LoanApplication, error) {
	now := time.Now()
	l := domain.LoanApplication{ID: fmt.Sprintf("loan-%d", now.UnixNano()), UserID: a.ID, AmountFen: amountFen,
		TermMonths: termMonths, Purpose: purpose, Status: "submitted", Version: 1, CreatedAt: now, UpdatedAt: now}
	return s.loanRepo.Create(l)
}

func (s *FinanceService) ListMyLoans(a domain.Actor) ([]domain.LoanApplication, error) {
	return s.loanRepo.ListByUser(a.ID)
}

func (s *FinanceService) ListAllLoans(offset, limit int) ([]domain.LoanApplication, int, error) {
	return s.loanRepo.ListAll(offset, limit)
}
