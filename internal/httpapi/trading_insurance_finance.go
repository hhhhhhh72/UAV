package httpapi

import (
	"errors"
	"net/http"
	"time"

	"drone-platform/internal/domain"
)

// ---- Trading ----

func (s *Server) createProduct(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
	var in struct {
		Title, Description, Brand, Model, Condition string
		ProdType string `json:"prod_type"`
		PriceFen int64  `json:"price_fen"`
	}
	if err := decode(r, &in); err != nil { fail(w, r, http.StatusBadRequest, err); return }
	p, err := s.tradingSvc.CreateProduct(a, domain.ProductType(in.ProdType), in.Title, in.Description, in.Brand, in.Model, in.Condition, in.PriceFen)
	if err != nil { fail(w, r, http.StatusForbidden, err); return }
	respond(w, r, http.StatusCreated, p)
}

func (s *Server) listProducts(w http.ResponseWriter, r *http.Request) {
	products, err := s.tradingSvc.ListProducts(r.URL.Query().Get("prod_type"))
	if err != nil { fail(w, r, http.StatusInternalServerError, err); return }
	respond(w, r, http.StatusOK, products)
}

func (s *Server) createRepair(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
	var in struct {
		ProductDesc string `json:"product_desc"`
		FaultDesc   string `json:"fault_desc"`
	}
	if err := decode(r, &in); err != nil { fail(w, r, http.StatusBadRequest, err); return }
	rp, err := s.tradingSvc.CreateRepair(a, in.ProductDesc, in.FaultDesc)
	if err != nil { fail(w, r, http.StatusForbidden, err); return }
	respond(w, r, http.StatusCreated, rp)
}

func (s *Server) listMyRepairs(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
	repairs, err := s.tradingSvc.ListMyRepairs(a)
	if err != nil { fail(w, r, http.StatusInternalServerError, err); return }
	respond(w, r, http.StatusOK, repairs)
}

// ---- Insurance ----

func (s *Server) createPolicy(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
	var in struct {
		DroneModel  string    `json:"drone_model"`
		DroneSN     string    `json:"drone_sn"`
		PolicyType  string    `json:"policy_type"`
		PremiumFen  int64     `json:"premium_fen"`
		CoverageFen int64     `json:"coverage_fen"`
		StartDate   time.Time `json:"start_date"`
		EndDate     time.Time `json:"end_date"`
	}
	if err := decode(r, &in); err != nil { fail(w, r, http.StatusBadRequest, err); return }
	p, err := s.insuranceSvc.CreatePolicy(a, in.DroneModel, in.DroneSN, in.PolicyType, in.PremiumFen, in.CoverageFen, in.StartDate, in.EndDate)
	if err != nil { fail(w, r, http.StatusForbidden, err); return }
	respond(w, r, http.StatusCreated, p)
}

func (s *Server) listMyPolicies(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
	policies, err := s.insuranceSvc.ListMyPolicies(a)
	if err != nil { fail(w, r, http.StatusInternalServerError, err); return }
	respond(w, r, http.StatusOK, policies)
}

func (s *Server) createInspection(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
	var in struct {
		DroneModel  string    `json:"drone_model"`
		DroneSN     string    `json:"drone_sn"`
		InspectDate time.Time `json:"inspect_date"`
		ExpireDate  time.Time `json:"expire_date"`
	}
	if err := decode(r, &in); err != nil { fail(w, r, http.StatusBadRequest, err); return }
	i, err := s.insuranceSvc.CreateInspection(a, in.DroneModel, in.DroneSN, in.InspectDate, in.ExpireDate)
	if err != nil { fail(w, r, http.StatusForbidden, err); return }
	respond(w, r, http.StatusCreated, i)
}

func (s *Server) listMyInspections(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
	inspections, err := s.insuranceSvc.ListMyInspections(a)
	if err != nil { fail(w, r, http.StatusInternalServerError, err); return }
	respond(w, r, http.StatusOK, inspections)
}

// ---- Finance ----

func (s *Server) applyLoan(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
	var in struct {
		AmountFen  int64  `json:"amount_fen"`
		TermMonths int    `json:"term_months"`
		Purpose    string `json:"purpose"`
	}
	if err := decode(r, &in); err != nil { fail(w, r, http.StatusBadRequest, err); return }
	l, err := s.financeSvc.ApplyLoan(a, in.AmountFen, in.TermMonths, in.Purpose)
	if err != nil { fail(w, r, http.StatusForbidden, err); return }
	respond(w, r, http.StatusCreated, l)
}

func (s *Server) listMyLoans(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
	loans, err := s.financeSvc.ListMyLoans(a)
	if err != nil { fail(w, r, http.StatusInternalServerError, err); return }
	respond(w, r, http.StatusOK, loans)
}
