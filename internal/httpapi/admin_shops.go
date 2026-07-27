package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"drone-platform/internal/domain"
)

// ---- Admin Shops (重用 Enterprise 数据) ----

// listAdminShops GET /api/v1/admin/shops
func (s *Server) listAdminShops(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if page < 1 { page = 1 }
	if pageSize < 1 || pageSize > 100 { pageSize = 20 }
	offset := (page - 1) * pageSize

	all, total, err := s.enterprises.ListByStatus("", offset, pageSize)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, fmt.Errorf("list shops: %w", err))
		return
	}
	paginatedRespond(w, r, all, total)
}

// createAdminShop POST /api/v1/admin/shops
func (s *Server) createAdminShop(w http.ResponseWriter, r *http.Request) {
	var ent domain.Enterprise
	if err := json.NewDecoder(r.Body).Decode(&ent); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	created, err := s.enterprises.Create(ent)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, fmt.Errorf("create shop: %w", err))
		return
	}
	respond(w, r, http.StatusCreated, created)
}

// updateAdminShop PUT /api/v1/admin/shops/{id}
func (s *Server) updateAdminShop(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var ent domain.Enterprise
	if err := json.NewDecoder(r.Body).Decode(&ent); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	ent.ID = id
	updated, err := s.enterprises.Update(id, ent)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, fmt.Errorf("update shop: %w", err))
		return
	}
	respond(w, r, http.StatusOK, updated)
}

// deleteAdminShop DELETE /api/v1/admin/shops/{id}
func (s *Server) deleteAdminShop(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := s.enterprises.Delete(id); err != nil {
		fail(w, r, http.StatusInternalServerError, fmt.Errorf("delete shop: %w", err))
		return
	}
	respond(w, r, http.StatusOK, map[string]string{"id": id})
}
