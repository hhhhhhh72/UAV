package httpapi

import (
	"errors"
	"net/http"

	"drone-platform/internal/domain"
)

// ── ResourcePool Handlers ──

func (s *Server) registerBatch1Routes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/admin/resource-pools", s.createResourcePool)
	mux.HandleFunc("GET /api/v1/resource-pools", s.listResourcePools)
	mux.HandleFunc("POST /api/v1/resource-pools/{id}/members", s.addPoolMember)
	mux.HandleFunc("GET /api/v1/resource-pools/{id}/members", s.listPoolMembers)

	mux.HandleFunc("POST /api/v1/admin/test-sites", s.createTestSite)
	mux.HandleFunc("GET /api/v1/test-sites", s.listTestSites)
	mux.HandleFunc("POST /api/v1/test-sites/{id}/book", s.bookTestSite)
	mux.HandleFunc("POST /api/v1/admin/test-sites/bookings/{id}/review", s.reviewTestSiteBooking)

	mux.HandleFunc("POST /api/v1/admin/exhibitions", s.createExhibition)
	mux.HandleFunc("GET /api/v1/exhibitions", s.listExhibitions)
	mux.HandleFunc("POST /api/v1/exhibitions/{id}/booths", s.applyBooth)
	mux.HandleFunc("GET /api/v1/exhibitions/{id}/booths", s.listBooths)
}

// GET /api/v1/resource-pools?type=emergency
func (s *Server) listResourcePools(w http.ResponseWriter, r *http.Request) {
	pools, err := s.poolSvc.List(r.URL.Query().Get("type"))
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, pools)
}

func (s *Server) createResourcePool(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("auth required"))
		return
	}
	var in struct {
		Name        string `json:"name"`
		PoolType    string `json:"pool_type"`
		Description string `json:"description"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	p, err := s.poolSvc.Create(in.Name, in.PoolType, in.Description, a.ID)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusCreated, p)
}

func (s *Server) addPoolMember(w http.ResponseWriter, r *http.Request) {
	_, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("auth required"))
		return
	}
	var in struct {
		ResID    string `json:"res_id"`
		ResType  string `json:"res_type"`
		Quantity int    `json:"quantity"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	m, err := s.poolSvc.AddMember(r.PathValue("id"), in.ResID, in.ResType, in.Quantity)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusCreated, m)
}

func (s *Server) listPoolMembers(w http.ResponseWriter, r *http.Request) {
	members, err := s.poolSvc.ListMembers(r.PathValue("id"))
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, members)
}

// ── TestSite ──

func (s *Server) createTestSite(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("auth required"))
		return
	}
	var in struct {
		Name        string   `json:"name"`
		SiteType    string   `json:"site_type"`
		Location    string   `json:"location"`
		BookingRule string   `json:"booking_rule"`
		PriceFen    int64    `json:"price_fen"`
		Facilities  []string `json:"facilities"`
		Status      string   `json:"status"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	ts, err := s.testSiteSvc.Create(in.Name, in.SiteType, in.Location, in.BookingRule, a.ID, in.PriceFen, in.Facilities, in.Status)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusCreated, ts)
}

func (s *Server) listTestSites(w http.ResponseWriter, r *http.Request) {
	list, err := s.testSiteSvc.List(r.URL.Query().Get("site_type"))
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, list)
}

func (s *Server) bookTestSite(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("auth required"))
		return
	}
	var in struct {
		Purpose   string `json:"purpose"`
		StartTime string `json:"start_time"`
		EndTime   string `json:"end_time"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	st := domain.ParseTime(in.StartTime)
	et := domain.ParseTime(in.EndTime)
	bk, err := s.testSiteSvc.Book(r.PathValue("id"), a.ID, in.Purpose, st, et)
	if err != nil {
		fail(w, r, http.StatusConflict, err)
		return
	}
	respond(w, r, http.StatusCreated, bk)
}

func (s *Server) reviewTestSiteBooking(w http.ResponseWriter, r *http.Request) {
	_, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("auth required"))
		return
	}
	var in struct {
		Status string `json:"status"`
		Note   string `json:"note"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	bk, err := s.testSiteSvc.ReviewBooking(r.PathValue("id"), in.Status, in.Note)
	if err != nil {
		fail(w, r, http.StatusNotFound, err)
		return
	}
	respond(w, r, http.StatusOK, bk)
}

// ── Exhibition ──

func (s *Server) createExhibition(w http.ResponseWriter, r *http.Request) {
	_, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("auth required"))
		return
	}
	var in struct {
		Title       string `json:"title"`
		Category    string `json:"category"`
		Description string `json:"description"`
		Location    string `json:"location"`
		Organizer   string `json:"organizer"`
		CoverURL    string `json:"cover_url"`
		StartDate   string `json:"start_date"`
		EndDate     string `json:"end_date"`
		BoothCount  int    `json:"booth_count"`
		BoothPrice  int64  `json:"booth_price_fen"`
		Status      string `json:"status"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	sd := domain.ParseTime(in.StartDate)
	ed := domain.ParseTime(in.EndDate)
	e, err := s.exhibitionSvc.Create(in.Title, in.Category, in.Description, in.Location, in.Organizer, in.CoverURL, sd, ed, in.BoothCount, in.BoothPrice, in.Status)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusCreated, e)
}

func (s *Server) listExhibitions(w http.ResponseWriter, r *http.Request) {
	page, pageSize := paginationFromQuery(r)
	list, total, err := s.exhibitionSvc.List(page, pageSize)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	paginatedRespond(w, r, list, total)
}

func (s *Server) applyBooth(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("auth required"))
		return
	}
	var in struct{ BoothNumber, ExhibitName, ExhibitDesc string }
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	b, err := s.exhibitionSvc.ApplyBooth(r.PathValue("id"), a.ID, in.BoothNumber, in.ExhibitName, in.ExhibitDesc)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusCreated, b)
}

func (s *Server) listBooths(w http.ResponseWriter, r *http.Request) {
	booths, err := s.exhibitionSvc.ListBooths(r.PathValue("id"))
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, booths)
}
