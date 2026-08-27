package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)

// ── StudyTourEnrollmentService (低空研学报名) ──
// 闭环：研学详情报名 → 我的报名(研学 tab) → 管理端审核（pending/approved/rejected/completed）

type StudyTourEnrollmentService struct {
	repo   repository.StudyTourEnrollmentRepository
	tours  repository.StudyTourRepository
}

func NewStudyTourEnrollmentService(repo repository.StudyTourEnrollmentRepository, tours repository.StudyTourRepository) *StudyTourEnrollmentService {
	return &StudyTourEnrollmentService{repo: repo, tours: tours}
}

var enrollPhoneRe = regexp.MustCompile(`^1[3-9]\d{9}$`)

// Create 研学报名：研学存在且招募中（active）、人数上限未满；手机号格式校验。
func (s *StudyTourEnrollmentService) Create(ctx context.Context, userID, tourID, name, phone string, adultCount, childCount int, remark string) (domain.StudyTourEnrollment, error) {
	tour, err := s.tours.FindByID(ctx, tourID)
	if err != nil {
		return domain.StudyTourEnrollment{}, errors.New("研学活动不存在或已下线")
	}
	if tour.Status != "active" && tour.Status != "published" {
		return domain.StudyTourEnrollment{}, errors.New("该研学活动已结束报名")
	}
	if name == "" {
		return domain.StudyTourEnrollment{}, errors.New("请填写报名人姓名")
	}
	if !enrollPhoneRe.MatchString(phone) {
		return domain.StudyTourEnrollment{}, errors.New("请填写正确的11位手机号")
	}
	if adultCount < 1 {
		adultCount = 1
	}
	if childCount < 0 {
		childCount = 0
	}
	// 容量校验（capacity>0 时生效）：已报名人数 + 本次人数不得超限
	if tour.Capacity > 0 {
		existing, err := s.repo.ListByTour(ctx, tourID)
		if err != nil {
			return domain.StudyTourEnrollment{}, fmt.Errorf("check tour capacity: %w", err)
		}
		taken := 0
		for _, e := range existing {
			if e.Status == "pending" || e.Status == "approved" {
				taken += e.AdultCount + e.ChildCount
			}
		}
		if taken+adultCount+childCount > tour.Capacity {
			return domain.StudyTourEnrollment{}, errors.New("该研学活动名额已满")
		}
	}
	now := time.Now()
	e := domain.StudyTourEnrollment{
		ID:         fmt.Sprintf("study-enroll-%d", now.UnixNano()),
		TourID:     tourID,
		UserID:     userID,
		Name:       name,
		Phone:      phone,
		AdultCount: adultCount,
		ChildCount: childCount,
		Remark:     remark,
		Status:     "pending",
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	return s.repo.Create(ctx, e)
}

// ListMyEnrollments 我的研学报名（附研学标题/出发日期）。
func (s *StudyTourEnrollmentService) ListMyEnrollments(ctx context.Context, userID string) ([]domain.StudyTourEnrollment, error) {
	items, err := s.repo.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	for i := range items {
		if t, err := s.tours.FindByID(ctx, items[i].TourID); err == nil {
			items[i].TourTitle = t.Title
			items[i].StartDate = t.StartDate
		}
	}
	return items, nil
}

// ListByTour 管理端：某研学的报名列表。
func (s *StudyTourEnrollmentService) ListByTour(ctx context.Context, tourID string) ([]domain.StudyTourEnrollment, error) {
	return s.repo.ListByTour(ctx, tourID)
}

// Review 管理端审核：approved/rejected/completed（pending 幂等守卫）。
func (s *StudyTourEnrollmentService) Review(ctx context.Context, id, status string) (domain.StudyTourEnrollment, error) {
	if status != "approved" && status != "rejected" && status != "completed" {
		return domain.StudyTourEnrollment{}, errors.New("状态仅支持 approved / rejected / completed")
	}
	cur, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return domain.StudyTourEnrollment{}, err
	}
	if cur.Status == status {
		return cur, nil
	}
	if cur.Status == "rejected" && status != "rejected" {
		return domain.StudyTourEnrollment{}, errors.New("已驳回的报名不能改为通过")
	}
	return s.repo.UpdateStatus(ctx, id, status)
}

// Get 单查（管理端详情/复核）。
func (s *StudyTourEnrollmentService) Get(ctx context.Context, id string) (domain.StudyTourEnrollment, error) {
	return s.repo.FindByID(ctx, id)
}
