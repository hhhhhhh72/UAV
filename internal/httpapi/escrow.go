package httpapi

import (
	"errors"
	"net/http"

	"drone-platform/internal/domain"
)

// requireEscrowAdmin 托管金写操作门禁：托管金为内部记账（无外部支付网关），
// 充值/冻结/释放/退款仅管理员可操作；业务侧由服务端状态机内部调用
// （pay-and-enroll / completeEnrollment），前端无公开调用。
// 此前任意登录用户可无限充值（deposit 无资金来源校验）再转账，属 P0 印钞漏洞。
func requireEscrowAdmin(a domain.Actor) bool {
	return a.Role == domain.RolePlatformAdmin || a.Role == domain.RoleAssociationAdmin
}

// POST /api/v1/escrow/deposit
func (s *Server) escrowDeposit(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	if !requireEscrowAdmin(a) {
		fail(w, r, http.StatusForbidden, errors.New("admin permission required"))
		return
	}
	var in struct {
		AmountFen int64 `json:"amount_fen"`
	}
	if err := decode(r, &in); err != nil || in.AmountFen <= 0 {
		fail(w, r, http.StatusBadRequest, errors.New("amount_fen > 0 required"))
		return
	}
	tx, err := s.escrowSvc.Deposit(r.Context(), a.ID, in.AmountFen)
	if err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	s.audit(r.Context(), a.ID, "escrow_deposit", "escrow", tx.ID, "deposited")
	respond(w, r, http.StatusCreated, tx)
}

// POST /api/v1/escrow/freeze
func (s *Server) escrowFreeze(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	if !requireEscrowAdmin(a) {
		fail(w, r, http.StatusForbidden, errors.New("admin permission required"))
		return
	}
	var in struct {
		AmountFen     int64  `json:"amount_fen"`
		ReferenceType string `json:"reference_type"`
		ReferenceID   string `json:"reference_id"`
	}
	if err := decode(r, &in); err != nil || in.AmountFen <= 0 {
		fail(w, r, http.StatusBadRequest, errors.New("amount_fen > 0 required"))
		return
	}
	tx, err := s.escrowSvc.Freeze(r.Context(), a.ID, in.AmountFen, in.ReferenceType, in.ReferenceID)
	if err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	s.audit(r.Context(), a.ID, "escrow_freeze", in.ReferenceType, in.ReferenceID, "frozen")
	respond(w, r, http.StatusCreated, tx)
}

// POST /api/v1/escrow/release
func (s *Server) escrowRelease(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	if !requireEscrowAdmin(a) {
		fail(w, r, http.StatusForbidden, errors.New("admin permission required"))
		return
	}
	var in struct {
		ToUser        string `json:"to_user"`
		AmountFen     int64  `json:"amount_fen"`
		ReferenceType string `json:"reference_type"`
		ReferenceID   string `json:"reference_id"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	tx, err := s.escrowSvc.Release(r.Context(), a.ID, in.ToUser, in.AmountFen, in.ReferenceType, in.ReferenceID)
	if err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	s.audit(r.Context(), a.ID, "escrow_release", in.ReferenceType, in.ReferenceID, "released")
	respond(w, r, http.StatusCreated, tx)
}

// POST /api/v1/escrow/refund
func (s *Server) escrowRefund(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	if !requireEscrowAdmin(a) {
		fail(w, r, http.StatusForbidden, errors.New("admin permission required"))
		return
	}
	var in struct {
		AmountFen     int64  `json:"amount_fen"`
		ReferenceType string `json:"reference_type"`
		ReferenceID   string `json:"reference_id"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	tx, err := s.escrowSvc.Refund(r.Context(), a.ID, in.AmountFen, in.ReferenceType, in.ReferenceID)
	if err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	s.audit(r.Context(), a.ID, "escrow_refund", in.ReferenceType, in.ReferenceID, "refunded")
	respond(w, r, http.StatusCreated, tx)
}

// GET /api/v1/escrow/balance
func (s *Server) escrowBalance(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	bal, err := s.escrowSvc.Balance(r.Context(), a.ID)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, bal)
}

// GET /api/v1/escrow/transactions
func (s *Server) escrowTransactions(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	txs, err := s.escrowSvc.Transactions(r.Context(), a.ID)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, txs)
}
