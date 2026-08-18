package httpapi

import (
	"errors"
	"net/http"

	"drone-platform/internal/service"
)

// POST /api/v1/demands/{id}/intents — 意向方登记对接意向（联系对接模式）
func (s *Server) createIntent(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	var in service.CreateIntentInput
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	it, err := s.intentSvc.Create(r.Context(), a, r.PathValue("id"), in)
	if err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	// 站内消息通知需求发布方：收到新的对接意向（仅真实发布者；自己发自己的已被 Service 拒绝）
	if d, err := s.demands.FindByID(r.Context(), r.PathValue("id")); err == nil && d.PublisherID != a.ID {
		s.msgSvc.Send(r.Context(), "system", d.PublisherID, "新的对接意向",
			"您的需求《"+d.Title+"》收到一条新的对接意向，请及时查看处理",
			"demand", d.ID)
	}
	respond(w, r, http.StatusCreated, it)
}

// GET /api/v1/demands/{id}/intents — 发布方查看该需求的意向列表
func (s *Server) listDemandIntents(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	intents, err := s.intentSvc.ListByDemand(r.Context(), a, r.PathValue("id"))
	if err != nil {
		fail(w, r, http.StatusForbidden, err)
		return
	}
	respond(w, r, http.StatusOK, intents)
}

// GET /api/v1/intents/mine — 我的意向登记记录
func (s *Server) listMyIntents(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	intents, err := s.intentSvc.ListMine(r.Context(), a)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, intents)
}

// POST /api/v1/intents/{id}/cancel — 意向方取消自己的登记（仅待处理可取消）
func (s *Server) cancelIntent(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	if err := s.intentSvc.Cancel(r.Context(), a, r.PathValue("id")); err != nil {
		fail(w, r, http.StatusForbidden, err)
		return
	}
	s.audit(r.Context(), a.ID, "cancel_intent", "intent", r.PathValue("id"), "cancelled")
	respond(w, r, http.StatusOK, map[string]any{"status": "cancelled"})
}
