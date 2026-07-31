package httpapi

import (
	"fmt"
	"net/http"
	"strconv"
)

// ---- Admin Shops ----

// listAdminShops GET /api/v1/admin/shops
func (s *Server) listAdminShops(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if page < 1 { page = 1 }
	if pageSize < 1 || pageSize > 100 { pageSize = 20 }
	offset := (page - 1) * pageSize

	all, total, err := s.shopSvc.List(offset, pageSize)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, fmt.Errorf("list shops: %w", err))
		return
	}
	paginatedRespond(w, r, all, total)
}

// createAdminShop POST /api/v1/admin/shops
func (s *Server) createAdminShop(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Name         string `json:"name"`
		LicenseURL   string `json:"license_url"`
		AccountName  string `json:"account_name"`
		ContactPhone string `json:"contact_phone"`
		Address      string `json:"address"`
		Status       string `json:"status"`
		IsMember     bool   `json:"is_member"`
		CreatedAt    string `json:"created_at"`
		UpdatedAt    string `json:"updated_at"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	shop, err := s.shopSvc.Create(in.Name, in.LicenseURL, in.AccountName, in.ContactPhone, in.Address, in.Status, in.IsMember)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, fmt.Errorf("create shop: %w", err))
		return
	}
	respond(w, r, http.StatusCreated, shop)
}

// updateAdminShop PUT /api/v1/admin/shops/{id}
func (s *Server) updateAdminShop(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var in struct {
		ID           string `json:"id"`
		Name         string `json:"name"`
		LicenseURL   string `json:"license_url"`
		AccountName  string `json:"account_name"`
		ContactPhone string `json:"contact_phone"`
		Address      string `json:"address"`
		Status       string `json:"status"`
		IsMember     bool   `json:"is_member"`
		CreatedAt    string `json:"created_at"`
		UpdatedAt    string `json:"updated_at"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	shop, err := s.shopSvc.Update(id, in.Name, in.LicenseURL, in.AccountName, in.ContactPhone, in.Address, in.Status, in.IsMember)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, fmt.Errorf("update shop: %w", err))
		return
	}
	respond(w, r, http.StatusOK, shop)
}

// deleteAdminShop DELETE /api/v1/admin/shops/{id}
func (s *Server) deleteAdminShop(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := s.shopSvc.Delete(id); err != nil {
		fail(w, r, http.StatusInternalServerError, fmt.Errorf("delete shop: %w", err))
		return
	}
	respond(w, r, http.StatusOK, map[string]string{"id": id})
}
