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
	var in struct {
		AmountFen int64  `json:"amount_fen"`
		ToUser    string `json:"to_user"`
	}
	if err := decode(r, &in); err != nil || in.AmountFen <= 0 {
		fail(w, r, http.StatusBadRequest, errors.New("amount_fen > 0 required"))
		return
	}
	// 自充值（模拟托管通道）：登录用户仅可为自己入账；管理员可指定 to_user 代充。
	// 单笔上限 20_000_000 分（¥200000）：覆盖高单价课程/商品；模拟通道限额定闸，真实支付接入后由支付校验替代。
	if in.AmountFen > 20000000 {
		fail(w, r, http.StatusBadRequest, errors.New("单笔充值上限 200000 元"))
		return
	}
	target := in.ToUser
	if target == "" {
		target = a.ID
	} else if target != a.ID && !requireEscrowAdmin(a) {
		fail(w, r, http.StatusForbidden, errors.New("仅管理员可为他人的托管金入账"))
		return
	}
	tx, err := s.escrowSvc.Deposit(r.Context(), target, in.AmountFen)
	if err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	s.audit(r.Context(), a.ID, "escrow_deposit", "escrow", tx.ID, "deposited")
	respond(w, r, http.StatusCreated, tx)
}

// GET /api/v1/escrow/mine — 我的托管金：余额 + 近期流水（用户自查，admin 版 balance/transactions 保留）
func (s *Server) escrowMine(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	acc, err := s.escrowSvc.Balance(r.Context(), a.ID)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	txs, err := s.escrowSvc.Transactions(r.Context(), a.ID)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, map[string]any{"account": acc, "transactions": txs})
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
		FromUser      string `json:"from_user"`
		AmountFen     int64  `json:"amount_fen"`
		ReferenceType string `json:"reference_type"`
		ReferenceID   string `json:"reference_id"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	// 释放账户：默认管理员本人；可指定 from_user（平台代学员/商户结算给收款方）。
	src := in.FromUser
	if src == "" {
		src = a.ID
	}
	tx, err := s.escrowSvc.Release(r.Context(), src, in.ToUser, in.AmountFen, in.ReferenceType, in.ReferenceID)
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
