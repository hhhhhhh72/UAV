package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)

// ── ChallengeClaimService (研发难题揭榜意向) ──
// 契约对齐小程序 challenges/detail.vue：
//   GET  /api/v1/challenges/{id}/claims → { items:[{id,claimer,status,created_at}], total, claimed }
//   POST /api/v1/challenges/{id}/claims（body 空，登录身份即揭榜人）

type ChallengeClaimService struct {
	claims     repository.ChallengeClaimRepository
	challenges repository.RDChallengeRepository
}

func NewChallengeClaimService(claims repository.ChallengeClaimRepository, challenges repository.RDChallengeRepository) *ChallengeClaimService {
	return &ChallengeClaimService{claims: claims, challenges: challenges}
}

// 揭榜守卫的稳定错误哨兵（handler 按 errors.Is 映射 404/409 与中文提示）。
var (
	ErrChallengeNotFound = errors.New("难题不存在或已下架")
	ErrChallengeClosed   = errors.New("该难题已截止揭榜")
	ErrClaimExists       = errors.New("已提交过揭榜意向，请等待协会审核")
)

// isPublicChallengeStatus 难题可揭榜的公开状态（与 httpapi.isPublicRDStatus 同一语义：
// published/open/in_progress 之外不可揭榜）。
func isPublicChallengeStatus(s string) bool {
	return s == "published" || s == "open" || s == "in_progress"
}

// ListByChallenge 揭榜动态：items（最新在前）+ 当前用户是否已揭榜。
// userID 为空时（匿名浏览）claimed 恒 false。
func (s *ChallengeClaimService) ListByChallenge(ctx context.Context, challengeID, userID string) ([]domain.ChallengeClaim, bool, error) {
	items, err := s.claims.ListByChallenge(ctx, challengeID)
	if err != nil {
		return nil, false, err
	}
	claimed := false
	if userID != "" {
		if _, found, err := s.claims.FindByChallengeAndUser(ctx, challengeID, userID); err == nil && found {
			claimed = true
		}
	}
	return items, claimed, nil
}

// Submit 提交揭榜意向（幂等 + 状态机守卫）：
//   - 难题不存在或未公开（draft/pending 等）→ "难题不存在或已下架"（404）
//   - 已截止（closed/resolved/deadline 已过）→ "该难题已截止揭榜"（409）
//   - 本用户已提交 → "已提交过揭榜意向，请等待协会审核"（409）
func (s *ChallengeClaimService) Submit(ctx context.Context, challengeID, userID string) (domain.ChallengeClaim, error) {
	c, err := s.challenges.FindByID(ctx, challengeID)
	if err != nil {
		return domain.ChallengeClaim{}, ErrChallengeNotFound
	}
	// 已结题/已解决 → 截止（先于公开判定：closed/resolved 是「曾公开但已截止」的语义）
	if c.Status == "closed" || c.Status == "resolved" {
		return domain.ChallengeClaim{}, ErrChallengeClosed
	}
	if !isPublicChallengeStatus(c.Status) {
		return domain.ChallengeClaim{}, ErrChallengeNotFound
	}
	if !c.Deadline.IsZero() && time.Now().After(c.Deadline) {
		return domain.ChallengeClaim{}, ErrChallengeClosed
	}
	if _, found, err := s.claims.FindByChallengeAndUser(ctx, challengeID, userID); err == nil && found {
		return domain.ChallengeClaim{}, ErrClaimExists
	}
	claim := domain.ChallengeClaim{
		ID:          nextID("claim"),
		ChallengeID: challengeID,
		UserID:      userID,
		Status:      "submitted", // submitted 待审核 / reviewing 审核中 / matched 已对接
		CreatedAt:   time.Now(),
	}
	created, err := s.claims.Create(ctx, claim)
	if err != nil {
		return domain.ChallengeClaim{}, fmt.Errorf("submit claim: %w", err)
	}
	return created, nil
}
