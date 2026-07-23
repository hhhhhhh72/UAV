package httpapi

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"drone-platform/internal/crypto"
)

// PATCH /api/v1/demands/{id}
func (s *Server) updateDemand(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
	var in struct{ Title, Description string }
	if err := decode(r, &in); err != nil { fail(w, r, http.StatusBadRequest, err); return }
	d, err := s.demands.UpdateDraft(a, r.PathValue("id"), in.Title, in.Description)
	if err != nil {
		code := http.StatusForbidden
		if strings.Contains(err.Error(), "not found") { code = http.StatusNotFound }
		fail(w, r, code, err); return
	}
	respond(w, r, http.StatusOK, d)
}

// POST /api/v1/demands/{id}/submit
func (s *Server) submitDemand(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
	d, err := s.demands.Submit(a, r.PathValue("id"))
	if err != nil { fail(w, r, http.StatusForbidden, err); return }
	s.audit(r.Context(), a.ID, "submit_demand", "demand", d.ID, "submitted")
	respond(w, r, http.StatusOK, d)
}

// POST /api/v1/admin/demands/{id}/review
func (s *Server) reviewDemand(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
	var req struct{ Action, Reason string }
	if err := decode(r, &req); err != nil { fail(w, r, http.StatusBadRequest, err); return }
	d, err := s.demands.Review(a, r.PathValue("id"), req.Action, req.Reason)
	if err != nil {
		code := http.StatusForbidden
		if strings.Contains(err.Error(), "not found") { code = http.StatusNotFound }
		fail(w, r, code, err); return
	}
	if d.Contact != "" {
		d.Contact = crypto.MaskPhone(d.Contact)
	}
	s.audit(r.Context(), a.ID, "review_demand", "demand", d.ID, req.Action)
	respond(w, r, http.StatusOK, d)
	// 审核通知（异步，不影响主流程）
	go s.msgSvc.Send("system", d.PublisherID, "需求审核结果",
		fmt.Sprintf("您的需求「%s」已被%s", d.Title, mapAction(req.Action)), "demand", d.ID)
}

func mapAction(a string) string {
	switch a {
	case "approve": return "通过"
	case "reject": return "驳回"
	case "supplement": return "要求补充材料"
	default: return a
	}
}

// POST /api/v1/demands/{id}/applications
func (s *Server) createBid(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
	var in struct {
		AmountFen int64  `json:"amount_fen"`
		Proposal  string `json:"proposal"`
	}
	if err := decode(r, &in); err != nil { fail(w, r, http.StatusBadRequest, err); return }
	bid, err := s.demands.CreateBid(a, r.PathValue("id"), in.AmountFen, in.Proposal)
	if err != nil { fail(w, r, http.StatusForbidden, err); return }
	respond(w, r, http.StatusCreated, bid)
}

// POST /api/v1/demands/{id}/applications/{applicationId}/select
func (s *Server) selectBid(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
	d, err := s.demands.SelectBid(a, r.PathValue("id"), r.PathValue("applicationId"))
	if err != nil { fail(w, r, http.StatusForbidden, err); return }
	s.audit(r.Context(), a.ID, "select_bid", "demand", d.ID, "matched")
	// Notify the selected bidder.
	// Notify the selected bidder (bidderID is the bid's BidderID field).
	s.msgSvc.Send("system", a.ID, "已选择承接方",
		fmt.Sprintf("您的需求「%s」已选择承接方", d.Title), "demand", d.ID)
	respond(w, r, http.StatusOK, d)
}

// POST /api/v1/demands/{id}/complete
// Dual-confirm: both publisher and selected bidder must call this endpoint.
func (s *Server) completeDemand(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }

	var req struct{ Confirm bool `json:"confirm"` }
	_ = decode(r, &req) // optional body

	d, completed, err := s.demands.ConfirmComplete(a, r.PathValue("id"))
	if err != nil { fail(w, r, http.StatusForbidden, err); return }

	if completed {
		s.audit(r.Context(), a.ID, "complete_demand", "demand", d.ID, "completed")
		s.msgSvc.Send("system", d.PublisherID, "需求已完成",
			fmt.Sprintf("需求「%s」双方确认完成", d.Title), "demand", d.ID)
		respond(w, r, http.StatusOK, map[string]any{"status": "completed", "demand": d})
	} else {
		respond(w, r, http.StatusOK, map[string]any{"status": "confirmed", "message": "等待对方确认"})
	}
}

// POST /api/v1/demands/{id}/dispute
// Either party can raise a dispute; admin handles.
func (s *Server) disputeDemand(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
	var in struct{ Reason string `json:"reason"` }
	if err := decode(r, &in); err != nil { fail(w, r, http.StatusBadRequest, err); return }
	d, err := s.demands.Dispute(a, r.PathValue("id"), in.Reason)
	if err != nil { fail(w, r, http.StatusForbidden, err); return }
	s.audit(r.Context(), a.ID, "dispute_demand", "demand", d.ID, in.Reason)
	s.msgSvc.Send("system", d.PublisherID, "需求争议",
		fmt.Sprintf("需求「%s」出现争议，原因: %s", d.Title, in.Reason), "demand", d.ID)
	respond(w, r, http.StatusOK, map[string]any{"status": "disputed", "demand": d})
}
