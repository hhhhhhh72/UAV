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

// ErrReviewAlreadyExists 同一评价人对同一目标已有评价（待审或已通过）。
// 单独成哨兵错误是为了让 Handler 能映射成 409 而不是 500——这是业务冲突，不是服务故障。
var ErrReviewAlreadyExists = errors.New("您已评价过该目标")

// 评价目标校验的哨兵错误。单独成哨兵是为了让 Handler 映射成 4xx 而不是笼统 500 ——
// 「订单不存在」「不是你的订单」「订单还没完成」都是用户能看懂、也改得动的状态。
var (
	// ErrReviewTargetNotFound 评价目标不存在（Handler → 404）。
	ErrReviewTargetNotFound = errors.New("评价目标不存在")
	// ErrReviewNotAllowed 无权评价该目标（Handler → 403）。
	ErrReviewNotAllowed = errors.New("无权评价该目标")
	// ErrReviewTargetNotReady 目标尚未进入可评价状态（Handler → 409）。
	ErrReviewTargetNotReady = errors.New("该目标当前状态不可评价")
)

type ReviewService struct {
	repo      repository.ReviewRepository
	orderRepo repository.WorkOrderRepository  // 工单评价校验（target_type=work_order）用
	tradeRepo repository.TradeOrderRepository // 商城订单评价校验（target_type=order）用
}

func NewReviewService(repo repository.ReviewRepository, orderRepo repository.WorkOrderRepository, tradeRepo repository.TradeOrderRepository) *ReviewService {
	return &ReviewService{repo: repo, orderRepo: orderRepo, tradeRepo: tradeRepo}
}

func (s *ReviewService) Submit(ctx context.Context, reviewerID, targetType, targetID string, rating int, content string) (domain.Review, error) {
	// 评分范围校验：1-5，越界（999/-5）拒绝入库（此前零校验）。
	if rating < 1 || rating > 5 {
		return domain.Review{}, errors.New("rating must be between 1 and 5")
	}
	// 工单评价闭环：仅已完成工单的双方（需求方/接单方）可评价（P0 死链修复——
	// 此前 POST /api/v1/reviews 零校验，任意登录用户可对任意目标刷评价）。
	if targetType == "work_order" {
		if s.orderRepo == nil {
			return domain.Review{}, errors.New("work order repository not available")
		}
		wo, err := s.orderRepo.FindByID(ctx, targetID)
		if err != nil {
			return domain.Review{}, fmt.Errorf("%w：工单 %s", ErrReviewTargetNotFound, targetID)
		}
		if wo.Status != domain.WorkOrderCompleted {
			return domain.Review{}, fmt.Errorf("%w（工单状态 %s）", ErrReviewTargetNotReady, wo.Status)
		}
		if wo.PublisherID != reviewerID && wo.WorkerID != reviewerID {
			return domain.Review{}, fmt.Errorf("%w：只有需求方或接单方可以评价该工单", ErrReviewNotAllowed)
		}
	}
	// 商城订单评价闭环（2026-09-22）：此前**只有 work_order 走校验**，order 这一类
	// 零校验 —— 不查订单存不存在、不查是不是本人买的、不查订单是否已完成，
	// 任何登录用户都能对任意 target_id 造一条评价。
	if targetType == "order" {
		if s.tradeRepo == nil {
			return domain.Review{}, errors.New("trade order repository not available")
		}
		o, err := s.tradeRepo.FindByID(ctx, targetID)
		if err != nil {
			return domain.Review{}, fmt.Errorf("%w：订单 %s", ErrReviewTargetNotFound, targetID)
		}
		// 只有买家可以评价：卖家评价自己的商品没有业务意义，且是刷自己好评的通道。
		if o.BuyerID != reviewerID {
			return domain.Review{}, fmt.Errorf("%w：只有买家可以评价该订单", ErrReviewNotAllowed)
		}
		// 状态机前置：只有 completed 才拿到货。pending/paid/shipped 都还没收货，
		// aftersale/cancelled 更不该评价（订单表的 status 是纯字符串，见 phase3.go 的状态机）。
		if o.Status != "completed" {
			return domain.Review{}, fmt.Errorf("%w（订单状态 %s）", ErrReviewTargetNotReady, o.Status)
		}
	}
	// 幂等：同一用户对同一目标已评价（pending/approved）则拒绝；被驳回后可重新评价。
	//
	// 必须走 ListByReviewerTarget（不按 status 过滤）而不是 ListByTarget：
	// 后者是"对外展示口径"仅返回 approved，而新评价是 pending —— 用它会查空，
	// 同一人就能反复提交同一目标的评价（实测连发 3 次全部 201）。
	existing, err := s.repo.ListByReviewerTarget(ctx, reviewerID, targetType, targetID)
	if err != nil {
		return domain.Review{}, fmt.Errorf("list reviews for duplicate check: %w", err)
	}
	for _, e := range existing {
		if e.Status != "rejected" {
			return domain.Review{}, ErrReviewAlreadyExists
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

var (
	// ErrReviewNotFound 评价不存在（Handler → 404）。
	ErrReviewNotFound = errors.New("评价不存在")
	// ErrReviewStateConflict 评价当前状态不允许该操作（Handler → 409）。
	ErrReviewStateConflict = errors.New("评价当前状态不允许该操作")
)

func (s *ReviewService) Approve(ctx context.Context, id string) error {
	// 状态机前置：当前状态为准（approved 幂等；rejected 已驳回不得再翻转为通过，
	// 需要重评请让用户重新提交产生新记录）。
	cur, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return notFoundErr(ErrReviewNotFound, "review", id, err)
	}
	if cur.Status == "approved" {
		return nil
	}
	if cur.Status == "rejected" {
		return fmt.Errorf("%w：已驳回的评价不能改为通过", ErrReviewStateConflict)
	}
	if _, err := s.repo.UpdateStatus(ctx, id, "approved"); err != nil {
		return notFoundErr(ErrReviewNotFound, "review", id, err)
	}
	return nil
}

func (s *ReviewService) Reject(ctx context.Context, id string) error {
	cur, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return notFoundErr(ErrReviewNotFound, "review", id, err)
	}
	if cur.Status == "rejected" {
		return nil
	}
	if cur.Status == "approved" {
		return fmt.Errorf("%w：已通过的评价不能改为驳回", ErrReviewStateConflict)
	}
	if _, err := s.repo.UpdateStatus(ctx, id, "rejected"); err != nil {
		return notFoundErr(ErrReviewNotFound, "review", id, err)
	}
	return nil
}

func (s *ReviewService) Delete(ctx context.Context, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return notFoundErr(ErrReviewNotFound, "review", id, err)
	}
	return nil
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
	created, cerr := s.repo.CreateBooking(ctx, bk)
	if cerr != nil {
		return domain.VenueBooking{}, bookingSlotErr(cerr)
	}
	return created, nil
}
