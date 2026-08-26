package httpapi

import (
	"context"
	"errors"
	"net/http"

	"drone-platform/internal/service"
)

// ── Challenge Claims (研发难题揭榜意向) ──
// 契约对齐小程序 challenges/detail.vue：
//   GET  /api/v1/challenges/{id}/claims → { items:[{id,claimer,status,created_at}], total, claimed }
//   POST /api/v1/challenges/{id}/claims（body 空，登录身份即揭榜人）

// GET /api/v1/challenges/{id}/claims — 揭榜动态（聚合数 + 脱敏条目 + 当前用户是否已揭榜）
func (s *Server) getChallengeClaims(w http.ResponseWriter, r *http.Request) {
	userID := ""
	if a, ok := authenticatedActor(r); ok {
		userID = a.ID
	}
	items, claimed, err := s.claimSvc.ListByChallenge(r.Context(), r.PathValue("id"), userID)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	for i := range items {
		items[i].Claimer = s.fillPosterName(r.Context(), items[i].UserID)
		if items[i].Claimer == "" {
			items[i].Claimer = "匿名会员"
		}
	}
	respond(w, r, http.StatusOK, map[string]any{"items": items, "total": len(items), "claimed": claimed})
}

// POST /api/v1/challenges/{id}/claims — 提交揭榜意向
func (s *Server) createChallengeClaim(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("auth required"))
		return
	}
	claim, err := s.claimSvc.Submit(r.Context(), r.PathValue("id"), a.ID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrChallengeNotFound):
			fail(w, r, http.StatusNotFound, err)
		case errors.Is(err, service.ErrChallengeClosed), errors.Is(err, service.ErrClaimExists):
			fail(w, r, http.StatusConflict, err)
		default:
			fail(w, r, http.StatusInternalServerError, err)
		}
		return
	}
	claim.Claimer = s.fillPosterName(r.Context(), claim.UserID)
	respond(w, r, http.StatusCreated, claim)
}

// POST /api/v1/achievements/{id}/favorite — 收藏/取消收藏成果（计数联动：收藏 +1 / 取消 -1）
func (s *Server) toggleAchievementFavorite(w http.ResponseWriter, r *http.Request) {
	_, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("auth required"))
		return
	}
	var in struct {
		Favorite *bool `json:"favorite"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	if in.Favorite == nil {
		fail(w, r, http.StatusBadRequest, errors.New("favorite 必填"))
		return
	}
	delta := 1
	if !*in.Favorite {
		delta = -1
	}
	if err := s.achievementSvc.AdjustStats(r.Context(), r.PathValue("id"), 0, delta); err != nil {
		writeMutationErr(w, r, err)
		return
	}
	respond(w, r, http.StatusOK, map[string]bool{"favorite": *in.Favorite})
}

// fillPosterName 响应层填充发布方展示名（userRepo 查询；失败返回空串，前端已有兜底文案）。
func (s *Server) fillPosterName(ctx context.Context, userID string) string {
	if userID == "" {
		return ""
	}
	u, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return ""
	}
	return u.Name
}

// fillPosterNames 列表批量填充（map 缓存去重，避免同一人重复查询）。
func (s *Server) fillPosterNames(ctx context.Context, ownerIDs []string) map[string]string {
	out := make(map[string]string, len(ownerIDs))
	seen := make(map[string]bool, len(ownerIDs))
	for _, id := range ownerIDs {
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		u, err := s.userRepo.FindByID(ctx, id)
		if err == nil && u.Name != "" {
			out[id] = u.Name
		}
	}
	return out
}
