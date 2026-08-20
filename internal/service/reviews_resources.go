package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)

// ---- Reviews ----

type ReviewService struct {
	repo      repository.ReviewRepository
	orderRepo repository.WorkOrderRepository // 工单评价校验（target_type=work_order）用
}

func NewReviewService(repo repository.ReviewRepository, orderRepo repository.WorkOrderRepository) *ReviewService {
	return &ReviewService{repo: repo, orderRepo: orderRepo}
}

func (s *ReviewService) Submit(ctx context.Context, reviewerID, targetType, targetID string, rating int, content string) (domain.Review, error) {
	// 工单评价闭环：仅已完成工单的双方（需求方/接单方）可评价（P0 死链修复——
	// 此前 POST /api/v1/reviews 零校验，任意登录用户可对任意目标刷评价）。
	if targetType == "work_order" {
		if s.orderRepo == nil {
			return domain.Review{}, errors.New("work order repository not available")
		}
		wo, err := s.orderRepo.FindByID(ctx, targetID)
		if err != nil {
			return domain.Review{}, errors.New("work order not found")
		}
		if wo.Status != domain.WorkOrderCompleted {
			return domain.Review{}, errors.New("only completed work orders can be reviewed")
		}
		if wo.PublisherID != reviewerID && wo.WorkerID != reviewerID {
			return domain.Review{}, errors.New("only the publisher or worker can review the work order")
		}
	}
	r := domain.Review{ID: nextID("review"), ReviewerID: reviewerID,
		TargetType: targetType, TargetID: targetID, Rating: rating, Content: content, Status: "pending", CreatedAt: time.Now()}
	return s.repo.Create(ctx, r)
}

func (s *ReviewService) ListByTarget(ctx context.Context, targetType, targetID string) ([]domain.Review, error) {
	return s.repo.ListByTarget(ctx, targetType, targetID)
}

func (s *ReviewService) ListAll(ctx context.Context, status string, offset, limit int) ([]domain.Review, int, error) {
	return s.repo.ListAll(ctx, status, offset, limit)
}

func (s *ReviewService) Approve(ctx context.Context, id string) error {
	_, err := s.repo.UpdateStatus(ctx, id, "approved")
	if err != nil {
		return fmt.Errorf("approve review %s: %w", id, err)
	}
	return nil
}

func (s *ReviewService) Reject(ctx context.Context, id string) error {
	_, err := s.repo.UpdateStatus(ctx, id, "rejected")
	if err != nil {
		return fmt.Errorf("reject review %s: %w", id, err)
	}
	return nil
}

func (s *ReviewService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// ---- Venues ----

type VenueService struct {
	repo repository.VenueRepository
}

func NewVenueService(repo repository.VenueRepository) *VenueService {
	return &VenueService{repo: repo}
}

func (s *VenueService) Create(ctx context.Context, ownerID, name, venueType, location string, priceFen int64) (domain.Venue, error) {
	if priceFen < 0 {
		return domain.Venue{}, errors.New("price cannot be negative")
	}
	v := domain.Venue{ID: nextID("venue"), OwnerID: ownerID, Name: name,
		VenueType: venueType, Location: location, PriceFen: priceFen, Status: "available", CreatedAt: time.Now()}
	return s.repo.Create(ctx, v)
}

func (s *VenueService) List(ctx context.Context, venueType string) ([]domain.Venue, error) {
	return s.repo.List(ctx, venueType)
}

func (s *VenueService) Book(ctx context.Context, venueID, userID string, start, end time.Time) (domain.VenueBooking, error) {
	// 时间范围校验：结束时间必须晚于开始时间（与 EventService 同款校验）
	if !end.After(start) {
		return domain.VenueBooking{}, errors.New("结束时间必须晚于开始时间")
	}
	// Check for time-slot conflicts.
	bookings, err := s.repo.ListBookings(ctx, venueID)
	if err != nil {
		return domain.VenueBooking{}, err
	}
	for _, b := range bookings {
		if b.Status == "booked" && !(end.Before(b.StartTime) || start.After(b.EndTime)) {
			return domain.VenueBooking{}, fmt.Errorf("time slot conflicted")
		}
	}
	bk := domain.VenueBooking{ID: nextID("booking"), VenueID: venueID,
		UserID: userID, StartTime: start, EndTime: end, Status: "booked", CreatedAt: time.Now()}
	return s.repo.CreateBooking(ctx, bk)
}
