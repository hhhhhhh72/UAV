package httpapi

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"drone-platform/internal/domain"
)

// ---- Posts ----

func (s *Server) createPost(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
	var in struct{ Title, Content string; Images []string `json:"images"` }
	if err := decode(r, &in); err != nil { fail(w, r, http.StatusBadRequest, err); return }
	p, err := s.communitySvc.CreatePost(a, in.Title, in.Content, in.Images)
	if err != nil { fail(w, r, http.StatusForbidden, err); return }
	s.audit(r.Context(), a.ID, "create_post", "post", p.ID, "created")
	respond(w, r, http.StatusCreated, p)
}

func (s *Server) publishPost(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
	p, err := s.communitySvc.PublishPost(a, r.PathValue("id"))
	if err != nil { fail(w, r, http.StatusForbidden, err); return }
	respond(w, r, http.StatusOK, p)
}

func (s *Server) removePost(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
	p, err := s.communitySvc.RemovePost(a, r.PathValue("id"))
	if err != nil { fail(w, r, http.StatusForbidden, err); return }
	respond(w, r, http.StatusOK, p)
}

func (s *Server) listPosts(w http.ResponseWriter, r *http.Request) {
	page, pageSize := paginationFromQuery(r)
	offset := (page - 1) * pageSize
	items, total, err := s.communitySvc.ListPublishedPosts(offset, pageSize)
	if err != nil { fail(w, r, http.StatusInternalServerError, err); return }
	paginatedRespond(w, r, items, total)
}

// ---- Comments ----

func (s *Server) createComment(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
	var in struct {
		PostID  string `json:"post_id"`
		Content string `json:"content"`
	}
	if err := decode(r, &in); err != nil { fail(w, r, http.StatusBadRequest, err); return }
	c, err := s.communitySvc.CreateComment(a, in.PostID, in.Content)
	if err != nil { fail(w, r, http.StatusForbidden, err); return }
	respond(w, r, http.StatusCreated, c)
}

func (s *Server) listComments(w http.ResponseWriter, r *http.Request) {
	postID := r.URL.Query().Get("post_id")
	if postID == "" { fail(w, r, http.StatusBadRequest, errors.New("post_id required")); return }
	items, err := s.communitySvc.ListComments(postID)
	if err != nil { fail(w, r, http.StatusInternalServerError, err); return }
	respond(w, r, http.StatusOK, items)
}

// ---- Reports ----

func (s *Server) createReport(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
	var in struct {
		ResourceType string `json:"resource_type"`
		ResourceID   string `json:"resource_id"`
		Reason       string `json:"reason"`
	}
	if err := decode(r, &in); err != nil { fail(w, r, http.StatusBadRequest, err); return }
	rp, err := s.communitySvc.CreateReport(a, in.ResourceType, in.ResourceID, in.Reason)
	if err != nil { fail(w, r, http.StatusForbidden, err); return }
	respond(w, r, http.StatusCreated, rp)
}

func (s *Server) listReports(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
	page, pageSize := paginationFromQuery(r)
	offset := (page - 1) * pageSize
	items, total, err := s.communitySvc.ListPendingReports(a, offset, pageSize)
	if err != nil { fail(w, r, http.StatusForbidden, err); return }
	paginatedRespond(w, r, items, total)
}

// ---- Listings ----

func (s *Server) createListing(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
	var in struct {
		Title, Description, Category string
		PriceFen                     int64 `json:"price_fen"`
	}
	if err := decode(r, &in); err != nil { fail(w, r, http.StatusBadRequest, err); return }
	now := time.Now()
	l := domain.Listing{ID: fmt.Sprintf("listing-%d", now.UnixNano()), SellerID: a.ID, Title: in.Title,
		Description: in.Description, Category: in.Category, PriceFen: in.PriceFen, Status: "listed", Version: 1, CreatedAt: now, UpdatedAt: now}
	l, err := s.listingSvc.Create(l)
	if err != nil { fail(w, r, http.StatusForbidden, err); return }
	respond(w, r, http.StatusCreated, l)
}

func (s *Server) closeListing(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
	l, err := s.listingSvc.Close(a, r.PathValue("id"))
	if err != nil { fail(w, r, http.StatusForbidden, err); return }
	respond(w, r, http.StatusOK, l)
}

func (s *Server) listListings(w http.ResponseWriter, r *http.Request) {
	page, pageSize := paginationFromQuery(r)
	offset := (page - 1) * pageSize
	items, total, err := s.listingSvc.ListListed(offset, pageSize)
	if err != nil { fail(w, r, http.StatusInternalServerError, err); return }
	paginatedRespond(w, r, items, total)
}

func (s *Server) favoriteListing(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
	if err := s.listingSvc.Favorite(r.PathValue("id"), a.ID); err != nil {
		fail(w, r, http.StatusForbidden, err); return
	}
	respond(w, r, http.StatusOK, map[string]string{"status": "favorited"})
}

// ---- Labour Orders ----

func (s *Server) createLabourOrder(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
	var in struct {
		Title, Description string
		WorkerCount        int       `json:"worker_count"`
		StartDate  time.Time `json:"start_date"`
		EndDate    time.Time `json:"end_date"`
		BudgetFen          int64     `json:"budget_fen"`
	}
	if err := decode(r, &in); err != nil { fail(w, r, http.StatusBadRequest, err); return }
	o, err := s.labourSvc.CreateOrder(a, in.Title, in.Description, in.WorkerCount, in.StartDate, in.EndDate, in.BudgetFen)
	if err != nil { fail(w, r, http.StatusForbidden, err); return }
	respond(w, r, http.StatusCreated, o)
}

func (s *Server) listLabourOrders(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
	page, pageSize := paginationFromQuery(r)
	offset := (page - 1) * pageSize
	items, total, err := s.labourSvc.ListOrders(a, offset, pageSize)
	if err != nil { fail(w, r, http.StatusForbidden, err); return }
	paginatedRespond(w, r, items, total)
}

func (s *Server) createLabourQuote(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
	var in struct {
		OrderID    string `json:"order_id"`
		AmountFen  int64  `json:"amount_fen"`
		Proposal   string `json:"proposal"`
		QuoterName string `json:"quoter_name"`
	}
	if err := decode(r, &in); err != nil { fail(w, r, http.StatusBadRequest, err); return }
	q, err := s.labourSvc.CreateQuote(a, in.OrderID, in.AmountFen, in.Proposal, in.QuoterName)
	if err != nil { fail(w, r, http.StatusForbidden, err); return }
	respond(w, r, http.StatusCreated, q)
}

func (s *Server) listLabourQuotes(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
	orderID := r.URL.Query().Get("order_id")
	if orderID == "" { fail(w, r, http.StatusBadRequest, errors.New("order_id required")); return }
	items, err := s.labourSvc.ListQuotes(a, orderID)
	if err != nil { fail(w, r, http.StatusForbidden, err); return }
	respond(w, r, http.StatusOK, items)
}

// GET /api/v1/labour-orders/{id}/assignments
func (s *Server) listOrderAssignments(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
	orderID := r.PathValue("id")
	items, err := s.labourSvc.ListAssignments(a, orderID)
	if err != nil {
		code := http.StatusForbidden
		if strings.Contains(err.Error(), "not found") {
			code = http.StatusNotFound
		}
		fail(w, r, code, err)
		return
	}
	respond(w, r, http.StatusOK, items)
}

// GET /api/v1/assignments/mine
func (s *Server) listMyAssignments(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
	items, err := s.labourSvc.ListMyAssignments(a)
	if err != nil { fail(w, r, http.StatusInternalServerError, err); return }
	respond(w, r, http.StatusOK, items)
}
