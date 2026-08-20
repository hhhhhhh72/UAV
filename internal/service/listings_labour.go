package service

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)

type ListingService struct{ repo repository.ListingRepository }

func NewListingService(r repository.ListingRepository) *ListingService {
	return &ListingService{repo: r}
}

func (s *ListingService) Create(ctx context.Context, l domain.Listing) (domain.Listing, error) {
	slog.Info("listing created", "listing_id", l.ID)
	return s.repo.Create(ctx, l)
}

func (s *ListingService) Close(ctx context.Context, a domain.Actor, id string) (domain.Listing, error) {
	l, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return domain.Listing{}, err
	}
	if l.SellerID != a.ID {
		return domain.Listing{}, errors.New("only the seller can close")
	}
	l.Status = "removed"
	l.UpdatedAt = time.Now()
	return s.repo.Update(ctx, id, l)
}

func (s *ListingService) ListListed(ctx context.Context, offset, limit int) ([]domain.Listing, int, error) {
	return s.repo.ListByStatus(ctx, "listed", offset, limit)
}

func (s *ListingService) Favorite(ctx context.Context, listingID, userID string) error {
	return s.repo.AddFavorite(ctx, listingID, userID)
}

type LabourService struct {
	repo repository.LabourOrderRepository
}

func NewLabourService(r repository.LabourOrderRepository) *LabourService {
	return &LabourService{repo: r}
}

func (s *LabourService) CreateOrder(ctx context.Context, a domain.Actor, title, desc string, workers int, start, end time.Time, budget int64) (domain.LabourOrder, error) {
	if budget < 0 {
		return domain.LabourOrder{}, errors.New("budget cannot be negative")
	}
	// 时间范围校验：结束时间必须晚于开始时间（与 EventService 同款校验）
	if !end.After(start) {
		return domain.LabourOrder{}, errors.New("结束时间必须晚于开始时间")
	}
	now := time.Now()
	o := domain.LabourOrder{ID: nextID("labour"), EmployerID: a.ID, Title: title,
		Description: desc, WorkerCount: workers, StartDate: start, EndDate: end, BudgetFen: budget, Status: "draft",
		Version: 1, CreatedAt: now, UpdatedAt: now}
	return s.repo.Create(ctx, o)
}

func (s *LabourService) ListOrders(ctx context.Context, a domain.Actor, offset, limit int) ([]domain.LabourOrder, int, error) {
	if a.Role == domain.RolePlatformAdmin {
		return s.repo.ListAll(ctx, offset, limit)
	}
	items, err := s.repo.ListByEmployer(ctx, a.ID)
	return items, len(items), err
}

func (s *LabourService) CreateQuote(ctx context.Context, a domain.Actor, orderID string, amount int64, proposal, name string) (domain.LabourQuote, error) {
	if amount < 0 {
		return domain.LabourQuote{}, errors.New("quote amount cannot be negative")
	}
	q := domain.LabourQuote{ID: nextID("quote"), OrderID: orderID, QuoterID: a.ID,
		QuoterName: name, AmountFen: amount, Proposal: proposal, Status: "pending", CreatedAt: time.Now()}
	return s.repo.CreateQuote(ctx, q)
}

func (s *LabourService) ListQuotes(ctx context.Context, a domain.Actor, orderID string) ([]domain.LabourQuote, error) {
	o, err := s.repo.FindByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if o.EmployerID != a.ID && a.Role != domain.RolePlatformAdmin {
		return nil, errors.New("permission denied")
	}
	return s.repo.ListQuotes(ctx, orderID)
}

func (s *LabourService) CreateAssignment(ctx context.Context, a domain.Actor, orderID, workerID string) (domain.Assignment, error) {
	o, err := s.repo.FindByID(ctx, orderID)
	if err != nil {
		return domain.Assignment{}, err
	}
	if o.EmployerID != a.ID && a.Role != domain.RolePlatformAdmin {
		return domain.Assignment{}, errors.New("only the employer can assign workers")
	}
	now := time.Now()
	asgn := domain.Assignment{ID: nextID("assign"), OrderID: orderID,
		WorkerID: workerID, Status: "assigned", CreatedAt: now}
	return s.repo.CreateAssignment(ctx, asgn)
}

func (s *LabourService) ListAssignments(ctx context.Context, a domain.Actor, orderID string) ([]domain.Assignment, error) {
	o, err := s.repo.FindByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if o.EmployerID != a.ID && a.Role != domain.RolePlatformAdmin {
		return nil, errors.New("permission denied")
	}
	return s.repo.ListAssignmentsByOrder(ctx, orderID)
}

func (s *LabourService) ListMyAssignments(ctx context.Context, a domain.Actor) ([]domain.Assignment, error) {
	return s.repo.ListAssignmentsByWorker(ctx, a.ID)
}
