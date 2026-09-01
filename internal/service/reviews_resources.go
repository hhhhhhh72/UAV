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
	// 幂等：同一用户对同一目标已评价（pending/approved）则拒绝；被驳回后可重新评价。
	// 查询出错必须上抛（此前 _ 吞错，DB 故障时误判"未评价"继续创建 → 重复评价）。
	existing, err := s.repo.ListByTarget(ctx, targetType, targetID)
	if err != nil {
		return domain.Review{}, fmt.Errorf("list reviews for duplicate check: %w", err)
	}
	for _, e := range existing {
		if e.ReviewerID == reviewerID && e.Status != "rejected" {
			return domain.Review{}, errors.New("您已评价过该目标")
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
	// 状态机前置：当前状态为准（approved 幂等；rejected 已驳回不得再翻转为通过，
	// 需要重评请让用户重新提交产生新记录）。
	cur, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("review %s: %w", id, err)
	}
	if cur.Status == "approved" {
		return nil
	}
	if cur.Status == "rejected" {
		return fmt.Errorf("已驳回的评价不能改为通过")
	}
	if _, err := s.repo.UpdateStatus(ctx, id, "approved"); err != nil {
		return fmt.Errorf("approve review %s: %w", id, err)
	}
	return nil
}

func (s *ReviewService) Reject(ctx context.Context, id string) error {
	cur, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("review %s: %w", id, err)
	}
	if cur.Status == "rejected" {
		return nil
	}
	if cur.Status == "approved" {
		return fmt.Errorf("已通过的评价不能改为驳回")
	}
	if _, err := s.repo.UpdateStatus(ctx, id, "rejected"); err != nil {
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
	// 场地存在校验：防对幽灵场地造预约
	if _, err := s.repo.FindByID(ctx, venueID); err != nil {
		return domain.VenueBooking{}, fmt.Errorf("场地不存在")
	}
	// 冲突判定：未取消预约（booked/pending）均占位 + 站点锁串行化 check-then-insert
	unlock := lockByKey("venue-book|" + venueID)
	defer unlock()
	bookings, err := s.repo.ListBookings(ctx, venueID)
	if err != nil {
		return domain.VenueBooking{}, err
	}
	for _, b := range bookings {
		if (b.Status == "booked" || b.Status == "pending") && !(end.Before(b.StartTime) || start.After(b.EndTime)) {
			return domain.VenueBooking{}, fmt.Errorf("time slot conflicted")
		}
	}
	bk := domain.VenueBooking{ID: nextID("booking"), VenueID: venueID,
		UserID: userID, StartTime: start, EndTime: end, Status: "booked", CreatedAt: time.Now()}
	return s.repo.CreateBooking(ctx, bk)
}
