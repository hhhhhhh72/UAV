package httpapi

import (
	"errors"
	"net/http"
)

// POST /api/v1/escrow/deposit
func (s *Server) escrowDeposit(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
	var in struct{ AmountFen int64 `json:"amount_fen"` }
	if err := decode(r, &in); err != nil || in.AmountFen <= 0 {
		fail(w, r, http.StatusBadRequest, errors.New("amount_fen > 0 required")); return
	}
	tx, err := s.escrowSvc.Deposit(a.ID, in.AmountFen)
	if err != nil { fail(w, r, http.StatusBadRequest, err); return }
	s.audit(r.Context(), a.ID, "escrow_deposit", "escrow", tx.ID, "deposited")
	respond(w, r, http.StatusCreated, tx)
}

// POST /api/v1/escrow/freeze
func (s *Server) escrowFreeze(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
	var in struct {
		AmountFen     int64  `json:"amount_fen"`
		ReferenceType string `json:"reference_type"`
		ReferenceID   string `json:"reference_id"`
	}
	if err := decode(r, &in); err != nil || in.AmountFen <= 0 {
		fail(w, r, http.StatusBadRequest, errors.New("amount_fen > 0 required")); return
	}
	tx, err := s.escrowSvc.Freeze(a.ID, in.AmountFen, in.ReferenceType, in.ReferenceID)
	if err != nil { fail(w, r, http.StatusBadRequest, err); return }
	s.audit(r.Context(), a.ID, "escrow_freeze", in.ReferenceType, in.ReferenceID, "frozen")
	respond(w, r, http.StatusCreated, tx)
}

// POST /api/v1/escrow/release
func (s *Server) escrowRelease(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
	var in struct {
		ToUser        string `json:"to_user"`
		AmountFen     int64  `json:"amount_fen"`
		ReferenceType string `json:"reference_type"`
		ReferenceID   string `json:"reference_id"`
	}
	if err := decode(r, &in); err != nil { fail(w, r, http.StatusBadRequest, err); return }
	tx, err := s.escrowSvc.Release(a.ID, in.ToUser, in.AmountFen, in.ReferenceType, in.ReferenceID)
	if err != nil { fail(w, r, http.StatusBadRequest, err); return }
	s.audit(r.Context(), a.ID, "escrow_release", in.ReferenceType, in.ReferenceID, "released")
	respond(w, r, http.StatusCreated, tx)
}

// POST /api/v1/escrow/refund
func (s *Server) escrowRefund(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
	var in struct {
		AmountFen     int64  `json:"amount_fen"`
		ReferenceType string `json:"reference_type"`
		ReferenceID   string `json:"reference_id"`
	}
	if err := decode(r, &in); err != nil { fail(w, r, http.StatusBadRequest, err); return }
	tx, err := s.escrowSvc.Refund(a.ID, in.AmountFen, in.ReferenceType, in.ReferenceID)
	if err != nil { fail(w, r, http.StatusBadRequest, err); return }
	s.audit(r.Context(), a.ID, "escrow_refund", in.ReferenceType, in.ReferenceID, "refunded")
	respond(w, r, http.StatusCreated, tx)
}

// GET /api/v1/escrow/balance
func (s *Server) escrowBalance(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
	bal, err := s.escrowSvc.Balance(a.ID)
	if err != nil { fail(w, r, http.StatusInternalServerError, err); return }
	respond(w, r, http.StatusOK, bal)
}

// GET /api/v1/escrow/transactions
func (s *Server) escrowTransactions(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
	txs, err := s.escrowSvc.Transactions(a.ID)
	if err != nil { fail(w, r, http.StatusInternalServerError, err); return }
	respond(w, r, http.StatusOK, txs)
}
