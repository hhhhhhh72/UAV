package service

import (
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

func (s *ReviewService) Submit(reviewerID, targetType, targetID string, rating int, content string) (domain.Review, error) {
	r := domain.Review{ID: fmt.Sprintf("review-%d", time.Now().UnixNano()), ReviewerID: reviewerID,
		TargetType: targetType, TargetID: targetID, Rating: rating, Content: content, Status: "pending", CreatedAt: time.Now()}
	return s.repo.Create(r)
}

func (s *ReviewService) ListByTarget(targetType, targetID string) ([]domain.Review, error) {
	return s.repo.ListByTarget(targetType, targetID)
}

func (s *ReviewService) ListAll(status string, offset, limit int) ([]domain.Review, int, error) {
	return s.repo.ListAll(status, offset, limit)
}

func (s *ReviewService) Approve(id string) error {
	_, err := s.repo.UpdateStatus(id, "approved")
	return err
}

func (s *ReviewService) Reject(id string) error {
	_, err := s.repo.UpdateStatus(id, "rejected")
	return err
}

func (s *ReviewService) Delete(id string) error {
	return s.repo.Delete(id)
}

// ---- Venues ----

type VenueService struct {
	repo repository.VenueRepository
}

func NewVenueService(repo repository.VenueRepository) *VenueService {
	return &VenueService{repo: repo}
}

func (s *VenueService) Create(ownerID, name, venueType, location string, priceFen int64) (domain.Venue, error) {
	v := domain.Venue{ID: fmt.Sprintf("venue-%d", time.Now().UnixNano()), OwnerID: ownerID, Name: name,
		VenueType: venueType, Location: location, PriceFen: priceFen, Status: "available", CreatedAt: time.Now()}
	return s.repo.Create(v)
}

func (s *VenueService) List(venueType string) ([]domain.Venue, error) {
	return s.repo.List(venueType)
}

func (s *VenueService) Book(venueID, userID string, start, end time.Time) (domain.VenueBooking, error) {
	// Check for time-slot conflicts.
	bookings, err := s.repo.ListBookings(venueID)
	if err != nil {
		return domain.VenueBooking{}, err
	}
	for _, b := range bookings {
		if b.Status == "booked" && !(end.Before(b.StartTime) || start.After(b.EndTime)) {
			return domain.VenueBooking{}, fmt.Errorf("time slot conflicted")
		}
	}
	bk := domain.VenueBooking{ID: fmt.Sprintf("booking-%d", time.Now().UnixNano()), VenueID: venueID,
		UserID: userID, StartTime: start, EndTime: end, Status: "booked", CreatedAt: time.Now()}
	return s.repo.CreateBooking(bk)
}
