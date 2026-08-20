package httpapi

import (
	"errors"
	"net/http"

	"drone-platform/internal/domain"
)

// ---- Service Listings (PRD ②-2 供给能力展示) ----

// GET /api/v1/service-listings — 公开列表（仅上架中）；mine=1 时返回当前用户发布的全部（含待审核，未登录返回空）
func (s *Server) listServiceListings(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("mine") == "1" {
		a, ok := authenticatedActor(r)
		if !ok {
			paginatedRespond(w, r, []domain.ServiceListing{}, 0)
			return
		}
		all, err := s.serviceListingSvc.ListAdmin(r.Context(), "", "")
		if err != nil {
			fail(w, r, http.StatusInternalServerError, err)
			return
		}
		mine := make([]domain.ServiceListing, 0, len(all))
		for _, sl := range all {
			if sl.ProviderID == a.ID {
				mine = append(mine, sl)
			}
		}
		paginatedRespond(w, r, mine, len(mine))
		return
	}
	items, err := s.serviceListingSvc.ListPublished(r.Context())
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	// P1 脱敏：公开列表返回前替换手机号注册用户的 provider_id，防止手机号泄露。
	// mine=1 分支返回本人发布的服务不脱敏。
	for i := range items {
		items[i].ProviderID = maskUserID(items[i].ProviderID)
	}
	paginatedRespond(w, r, items, len(items))
}

// POST /api/v1/service-listings — 用户自助发布服务能力（登录；待审核，管理端审核后上架）
func (s *Server) createServiceListing(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	var in struct {
		ProviderName string `json:"provider_name"`
		Title        string `json:"title"`
		Category     string `json:"category"`
		Description  string `json:"description"`
		Region       string `json:"region"`
		PriceFen     int64  `json:"price_fen"`
		Unit         string `json:"unit"`
		Image        string `json:"image"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	if in.Title == "" {
		fail(w, r, http.StatusBadRequest, errors.New("title is required"))
		return
	}
	sl, err := s.serviceListingSvc.CreateListingPending(r.Context(), a.ID, in.ProviderName, in.Title, in.Category, in.Description, in.Region, in.PriceFen, in.Unit, in.Image)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	s.audit(r.Context(), a.ID, "create_service_listing", "service_listing", sl.ID, "created")
	respond(w, r, http.StatusCreated, sl)
}

// GET /api/v1/service-listings/{id} — 公开详情（仅上架中，下架视为不存在）
func (s *Server) getServiceListing(w http.ResponseWriter, r *http.Request) {
	sl, err := s.serviceListingSvc.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		fail(w, r, http.StatusNotFound, err)
		return
	}
	if sl.Status != "" && sl.Status != "published" {
		fail(w, r, http.StatusNotFound, errors.New("service listing not found"))
		return
	}
	// P1 脱敏：公开详情返回前替换手机号注册用户的 provider_id，防止手机号泄露。
	sl.ProviderID = maskUserID(sl.ProviderID)
	respond(w, r, http.StatusOK, sl)
}

// POST /api/v1/admin/service-listings — 管理端创建
func (s *Server) adminCreateServiceListing(w http.ResponseWriter, r *http.Request) {
	var in struct {
		ProviderName string `json:"provider_name"`
		Title        string `json:"title"`
		Category     string `json:"category"`
		Description  string `json:"description"`
		Region       string `json:"region"`
		PriceFen     int64  `json:"price_fen"`
		Unit         string `json:"unit"`
		Image        string `json:"image"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	if in.Title == "" {
		fail(w, r, http.StatusBadRequest, errors.New("title is required"))
		return
	}
	sl, err := s.serviceListingSvc.CreateListing(r.Context(), "", in.ProviderName, in.Title, in.Category, in.Description, in.Region, in.PriceFen, in.Unit, in.Image)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusCreated, sl)
}

// GET /api/v1/admin/service-listings — 管理端全部列表（含下架），支持 keyword/category 过滤
func (s *Server) adminListServiceListings(w http.ResponseWriter, r *http.Request) {
	all, err := s.serviceListingSvc.ListAdmin(r.Context(), r.URL.Query().Get("keyword"), r.URL.Query().Get("category"))
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	paginatedRespond(w, r, all, len(all))
}

// PUT /api/v1/admin/service-listings/{id} — 管理端更新
func (s *Server) adminUpdateServiceListing(w http.ResponseWriter, r *http.Request) {
	var in struct {
		ProviderName string `json:"provider_name"`
		Title        string `json:"title"`
		Category     string `json:"category"`
		Description  string `json:"description"`
		Region       string `json:"region"`
		PriceFen     int64  `json:"price_fen"`
		Unit         string `json:"unit"`
		Image        string `json:"image"`
		Status       string `json:"status"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	existing, err := s.serviceListingSvc.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		fail(w, r, http.StatusNotFound, err)
		return
	}
	existing.ProviderName = in.ProviderName
	existing.Title = in.Title
	existing.Category = in.Category
	existing.Description = in.Description
	existing.Region = in.Region
	existing.PriceFen = in.PriceFen
	existing.Unit = in.Unit
	existing.Image = in.Image
	if in.Status != "" {
		existing.Status = in.Status
	}
	sl, err := s.serviceListingSvc.UpdateListing(r.Context(), existing)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, sl)
}

// DELETE /api/v1/admin/service-listings/{id} — 管理端删除
func (s *Server) adminDeleteServiceListing(w http.ResponseWriter, r *http.Request) {
	if err := s.serviceListingSvc.DeleteListing(r.Context(), r.PathValue("id")); err != nil {
		fail(w, r, http.StatusNotFound, err)
		return
	}
	respond(w, r, http.StatusOK, map[string]string{"deleted": "ok"})
}
