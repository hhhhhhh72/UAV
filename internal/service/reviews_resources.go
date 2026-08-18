package service

import (
	"context"
	"fmt"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)

// ---- Reviews ----

type ReviewService struct {
	repo repository.ReviewRepository
}

func NewReviewService(repo repository.ReviewRepository) *ReviewService {
	return &ReviewService{repo: repo}
}

func (s *ReviewService) Submit(ctx context.Context, reviewerID, targetType, targetID string, rating int, content string) (domain.Review, error) {
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
	v := domain.Venue{ID: nextID("venue"), OwnerID: ownerID, Name: name,
		VenueType: venueType, Location: location, PriceFen: priceFen, Status: "available", CreatedAt: time.Now()}
	return s.repo.Create(ctx, v)
}

func (s *VenueService) List(ctx context.Context, venueType string) ([]domain.Venue, error) {
	return s.repo.List(ctx, venueType)
}

func (s *VenueService) Book(ctx context.Context, venueID, userID string, start, end time.Time) (domain.VenueBooking, error) {
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
