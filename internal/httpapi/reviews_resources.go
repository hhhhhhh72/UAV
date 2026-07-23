package httpapi

import (
	"errors"
	"net/http"
	"time"

	"drone-platform/internal/domain"
)

// ---- Reviews ----

// POST /api/v1/reviews
func (s *Server) submitReview(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
	var in struct {
		TargetType string `json:"target_type"`
		TargetID   string `json:"target_id"`
		Rating     int    `json:"rating"`
		Content    string `json:"content"`
	}
	if err := decode(r, &in); err != nil || in.Rating < 1 || in.Rating > 5 {
		fail(w, r, http.StatusBadRequest, errors.New("rating must be 1-5")); return
	}
	rev, err := s.reviewSvc.Submit(a.ID, in.TargetType, in.TargetID, in.Rating, in.Content)
	if err != nil { fail(w, r, http.StatusInternalServerError, err); return }
	respond(w, r, http.StatusCreated, rev)
}

// GET /api/v1/reviews?target_type=enterprise&target_id=xxx
func (s *Server) listReviews(w http.ResponseWriter, r *http.Request) {
	ttype := r.URL.Query().Get("target_type")
	tid := r.URL.Query().Get("target_id")
	if ttype == "" || tid == "" {
		fail(w, r, http.StatusBadRequest, errors.New("target_type and target_id required")); return
	}
	reviews, err := s.reviewSvc.ListByTarget(ttype, tid)
	if err != nil { fail(w, r, http.StatusInternalServerError, err); return }
	respond(w, r, http.StatusOK, reviews)
}

// ---- Venues ----

// POST /api/v1/venues
func (s *Server) createVenue(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
	var in struct {
		Name      string `json:"name"`
		VenueType string `json:"venue_type"`
		Location  string `json:"location"`
		PriceFen  int64  `json:"price_fen"`
	}
	if err := decode(r, &in); err != nil { fail(w, r, http.StatusBadRequest, err); return }
	v, err := s.venueSvc.Create(a.ID, in.Name, in.VenueType, in.Location, in.PriceFen)
	if err != nil { fail(w, r, http.StatusInternalServerError, err); return }
	respond(w, r, http.StatusCreated, v)
}

// GET /api/v1/venues?type=training_field
func (s *Server) listVenues(w http.ResponseWriter, r *http.Request) {
	venues, err := s.venueSvc.List(r.URL.Query().Get("type"))
	if err != nil { fail(w, r, http.StatusInternalServerError, err); return }
	respond(w, r, http.StatusOK, venues)
}

// POST /api/v1/venues/{id}/book
func (s *Server) bookVenue(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
	var in struct {
		StartTime time.Time `json:"start_time"`
		EndTime   time.Time `json:"end_time"`
	}
	if err := decode(r, &in); err != nil { fail(w, r, http.StatusBadRequest, err); return }
	bk, err := s.venueSvc.Book(r.PathValue("id"), a.ID, in.StartTime, in.EndTime)
	if err != nil { fail(w, r, http.StatusConflict, err); return }
	respond(w, r, http.StatusCreated, bk)
}

// ---- Admin Review Management ----

// GET /api/v1/admin/reviews — list all reviews with optional status filter.
func (s *Server) listAllReviews(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok || (a.Role != domain.RolePlatformAdmin && a.Role != domain.RoleAssociationAdmin) {
		fail(w, r, http.StatusForbidden, errors.New("admin permission required"))
		return
	}
	status := r.URL.Query().Get("status")
	page, pageSize := paginationFromQuery(r)
	offset := (page - 1) * pageSize
	reviews, total, err := s.reviewSvc.ListAll(status, offset, pageSize)
	if err != nil { fail(w, r, http.StatusInternalServerError, err); return }
	paginatedRespond(w, r, reviews, total)
}

// POST /api/v1/admin/reviews/{id}/approve
func (s *Server) approveReview(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok || (a.Role != domain.RolePlatformAdmin && a.Role != domain.RoleAssociationAdmin) {
		fail(w, r, http.StatusForbidden, errors.New("admin permission required"))
		return
	}
	if err := s.reviewSvc.Approve(r.PathValue("id")); err != nil {
		code := http.StatusForbidden
		if err.Error() == "not found" {
			code = http.StatusNotFound
		}
		fail(w, r, code, err)
		return
	}
	respond(w, r, http.StatusOK, map[string]string{"status": "approved"})
}

// POST /api/v1/admin/reviews/{id}/reject
func (s *Server) rejectReview(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok || (a.Role != domain.RolePlatformAdmin && a.Role != domain.RoleAssociationAdmin) {
		fail(w, r, http.StatusForbidden, errors.New("admin permission required"))
		return
	}
	if err := s.reviewSvc.Reject(r.PathValue("id")); err != nil {
		code := http.StatusForbidden
		if err.Error() == "not found" {
			code = http.StatusNotFound
		}
		fail(w, r, code, err)
		return
	}
	respond(w, r, http.StatusOK, map[string]string{"status": "rejected"})
}

// DELETE /api/v1/admin/reviews/{id}
func (s *Server) deleteReview(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok || (a.Role != domain.RolePlatformAdmin && a.Role != domain.RoleAssociationAdmin) {
		fail(w, r, http.StatusForbidden, errors.New("admin permission required"))
		return
	}
	if err := s.reviewSvc.Delete(r.PathValue("id")); err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, map[string]string{"status": "deleted"})
}
