package httpapi

import (
	"errors"
	"net/http"

	"drone-platform/internal/domain"
)

// ---- 内容收藏：商品 / 服务能力 / 培训课程（"我的收藏"列表数据源，与需求收藏同构） ----

// POST /api/v1/products/{id}/favorite — 收藏/取消收藏商品
func (s *Server) toggleProductFavorite(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	var in struct {
		Favorite bool `json:"favorite"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	if err := s.tradingSvc.ToggleProductFavorite(r.Context(), a.ID, r.PathValue("id"), in.Favorite); err != nil {
		fail(w, r, http.StatusNotFound, err)
		return
	}
	respond(w, r, http.StatusOK, map[string]bool{"favorite": in.Favorite})
}

// GET /api/v1/products/favorites/mine — 我的收藏商品（按收藏时间倒序）
func (s *Server) listMyProductFavorites(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		respond(w, r, http.StatusOK, []domain.DroneProduct{})
		return
	}
	items, err := s.tradingSvc.ListFavoriteProducts(r.Context(), a.ID)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, items)
}

// POST /api/v1/service-listings/{id}/favorite — 收藏/取消收藏服务能力
func (s *Server) toggleServiceListingFavorite(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	var in struct {
		Favorite bool `json:"favorite"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	if err := s.serviceListingSvc.ToggleFavorite(r.Context(), a.ID, r.PathValue("id"), in.Favorite); err != nil {
		fail(w, r, http.StatusNotFound, err)
		return
	}
	respond(w, r, http.StatusOK, map[string]bool{"favorite": in.Favorite})
}

// GET /api/v1/service-listings/favorites/mine — 我的收藏服务（按收藏时间倒序，公开视图脱敏）
func (s *Server) listMyServiceListingFavorites(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		respond(w, r, http.StatusOK, []domain.ServiceListing{})
		return
	}
	items, err := s.serviceListingSvc.ListFavorites(r.Context(), a.ID)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	// P1 脱敏：与公开列表一致，替换手机号注册用户的 provider_id。
	for i := range items {
		items[i].ProviderID = maskUserID(items[i].ProviderID)
	}
	respond(w, r, http.StatusOK, items)
}

// POST /api/v1/training-courses/{id}/favorite — 收藏/取消收藏培训课程
func (s *Server) toggleTrainingCourseFavorite(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	var in struct {
		Favorite bool `json:"favorite"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	if err := s.trainingSvc.ToggleCourseFavorite(r.Context(), a.ID, r.PathValue("id"), in.Favorite); err != nil {
		fail(w, r, http.StatusNotFound, err)
		return
	}
	respond(w, r, http.StatusOK, map[string]bool{"favorite": in.Favorite})
}

// GET /api/v1/training-courses/favorites/mine — 我的收藏课程（按收藏时间倒序）
func (s *Server) listMyTrainingCourseFavorites(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		respond(w, r, http.StatusOK, []domain.TrainingCourse{})
		return
	}
	items, err := s.trainingSvc.ListFavoriteCourses(r.Context(), a.ID)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, items)
}
