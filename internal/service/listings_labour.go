package service

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)

type ListingService struct{ repo repository.ListingRepository }

func NewListingService(r repository.ListingRepository) *ListingService { return &ListingService{repo: r} }

func (s *ListingService) Create(l domain.Listing) (domain.Listing, error) { slog.Info("listing created", "listing_id", l.ID)
	return s.repo.Create(l) }

func (s *ListingService) Close(a domain.Actor, id string) (domain.Listing, error) {
	l, err := s.repo.FindByID(id)
	if err != nil { return domain.Listing{}, err }
	if l.SellerID != a.ID { return domain.Listing{}, errors.New("only the seller can close") }
	l.Status = "removed"; l.UpdatedAt = time.Now()
	return s.repo.Update(id, l)
}

func (s *ListingService) ListListed(offset, limit int) ([]domain.Listing, int, error) {
	return s.repo.ListByStatus("listed", offset, limit)
}

func (s *ListingService) Favorite(listingID, userID string) error { return s.repo.AddFavorite(listingID, userID) }

type LabourService struct{ repo repository.LabourOrderRepository }

func NewLabourService(r repository.LabourOrderRepository) *LabourService { return &LabourService{repo: r} }

func (s *LabourService) CreateOrder(a domain.Actor, title, desc string, workers int, start, end time.Time, budget int64) (domain.LabourOrder, error) {
	now := time.Now()
	o := domain.LabourOrder{ID: fmt.Sprintf("labour-%d", now.UnixNano()), EmployerID: a.ID, Title: title,
		Description: desc, WorkerCount: workers, StartDate: start, EndDate: end, BudgetFen: budget, Status: "draft",
		Version: 1, CreatedAt: now, UpdatedAt: now}
	return s.repo.Create(o)
}

func (s *LabourService) ListOrders(a domain.Actor, offset, limit int) ([]domain.LabourOrder, int, error) {
	if a.Role == domain.RolePlatformAdmin { return s.repo.ListAll(offset, limit) }
	items, err := s.repo.ListByEmployer(a.ID)
	return items, len(items), err
}

func (s *LabourService) CreateQuote(a domain.Actor, orderID string, amount int64, proposal, name string) (domain.LabourQuote, error) {
	q := domain.LabourQuote{ID: fmt.Sprintf("quote-%d", time.Now().UnixNano()), OrderID: orderID, QuoterID: a.ID,
		QuoterName: name, AmountFen: amount, Proposal: proposal, Status: "pending", CreatedAt: time.Now()}
	return s.repo.CreateQuote(q)
}

func (s *LabourService) ListQuotes(a domain.Actor, orderID string) ([]domain.LabourQuote, error) {
	o, err := s.repo.FindByID(orderID)
	if err != nil { return nil, err }
	if o.EmployerID != a.ID && a.Role != domain.RolePlatformAdmin {
		return nil, errors.New("permission denied")
	}
	return s.repo.ListQuotes(orderID)
}

func (s *LabourService) CreateAssignment(a domain.Actor, orderID, workerID string) (domain.Assignment, error) {
	o, err := s.repo.FindByID(orderID)
	if err != nil { return domain.Assignment{}, err }
	if o.EmployerID != a.ID && a.Role != domain.RolePlatformAdmin {
		return domain.Assignment{}, errors.New("only the employer can assign workers")
	}
	now := time.Now()
	asgn := domain.Assignment{ID: fmt.Sprintf("assign-%d", now.UnixNano()), OrderID: orderID,
		WorkerID: workerID, Status: "assigned", CreatedAt: now}
	return s.repo.CreateAssignment(asgn)
}

func (s *LabourService) ListAssignments(a domain.Actor, orderID string) ([]domain.Assignment, error) {
	o, err := s.repo.FindByID(orderID)
	if err != nil { return nil, err }
	if o.EmployerID != a.ID && a.Role != domain.RolePlatformAdmin {
		return nil, errors.New("permission denied")
	}
	return s.repo.ListAssignmentsByOrder(orderID)
}

func (s *LabourService) ListMyAssignments(a domain.Actor) ([]domain.Assignment, error) {
	return s.repo.ListAssignmentsByWorker(a.ID)
}
