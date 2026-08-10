package httpapi

import (
	"errors"
	"io"
	"net/http"
)

// POST /api/v1/demands/{id}/intents/{intentID}/accept — 企业确认接单，生成订单
func (s *Server) acceptIntent(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	var req struct {
		AmountFen int64 `json:"amount_fen"` // 订单金额（分），面议为 0
	}
	// 无 body 视为面议（金额 0）
	if err := decode(r, &req); err != nil && !errors.Is(err, io.EOF) {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	wo, err := s.workOrderSvc.AcceptIntent(a, r.PathValue("id"), r.PathValue("intentID"), req.AmountFen)
	if err != nil {
		fail(w, r, http.StatusForbidden, err)
		return
	}
	s.audit(r.Context(), a.ID, "accept_intent", "work_order", wo.ID, "created")
	respond(w, r, http.StatusCreated, wo)
}

// POST /api/v1/demands/{id}/intents/{intentID}/reject — 企业拒绝接单
func (s *Server) rejectIntent(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	if err := s.workOrderSvc.RejectIntent(a, r.PathValue("id"), r.PathValue("intentID")); err != nil {
		fail(w, r, http.StatusForbidden, err)
		return
	}
	s.audit(r.Context(), a.ID, "reject_intent", "work_order", r.PathValue("intentID"), "rejected")
	respond(w, r, http.StatusOK, map[string]any{"status": "rejected"})
}

// GET /api/v1/work-orders/mine — 我的订单（企业=发出的，飞手=接到的）
func (s *Server) listMyWorkOrders(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	orders, err := s.workOrderSvc.ListMine(a)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, orders)
}

// GET /api/v1/work-orders/{id} — 订单详情（仅订单双方）
func (s *Server) workOrderDetail(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	wo, err := s.workOrderSvc.FindByID(a, r.PathValue("id"))
	if err != nil {
		fail(w, r, http.StatusForbidden, err)
		return
	}
	respond(w, r, http.StatusOK, wo)
}

// POST /api/v1/work-orders/{id}/start — 飞手确认开始作业
func (s *Server) startWorkOrder(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	wo, err := s.workOrderSvc.StartWork(a, r.PathValue("id"))
	if err != nil {
		fail(w, r, http.StatusForbidden, err)
		return
	}
	s.audit(r.Context(), a.ID, "start_work_order", "work_order", wo.ID, "started")
	respond(w, r, http.StatusOK, wo)
}

// POST /api/v1/work-orders/{id}/complete — 飞手确认完成（可上传成果照片）
func (s *Server) completeWorkOrder(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	var req struct {
		ResultPhotos []string `json:"result_photos"`
	}
	// 空 body 视为无成果照片（确认完成不强制上传）
	if err := decode(r, &req); err != nil && !errors.Is(err, io.EOF) {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	wo, err := s.workOrderSvc.CompleteWork(a, r.PathValue("id"), req.ResultPhotos)
	if err != nil {
		fail(w, r, http.StatusForbidden, err)
		return
	}
	s.audit(r.Context(), a.ID, "complete_work_order", "work_order", wo.ID, "completed")
	respond(w, r, http.StatusOK, wo)
}

// POST /api/v1/work-orders/{id}/accept — 企业验收通过
func (s *Server) acceptWorkOrder(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	wo, err := s.workOrderSvc.AcceptCompletion(a, r.PathValue("id"))
	if err != nil {
		fail(w, r, http.StatusForbidden, err)
		return
	}
	s.audit(r.Context(), a.ID, "accept_work_order", "work_order", wo.ID, "accepted")
	respond(w, r, http.StatusOK, wo)
}

// POST /api/v1/work-orders/{id}/rework — 企业提出整改要求
func (s *Server) reworkWorkOrder(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	var req struct {
		Note string `json:"note"`
	}
	if err := decode(r, &req); err != nil && !errors.Is(err, io.EOF) {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	wo, err := s.workOrderSvc.RequestRework(a, r.PathValue("id"), req.Note)
	if err != nil {
		fail(w, r, http.StatusForbidden, err)
		return
	}
	s.audit(r.Context(), a.ID, "rework_work_order", "work_order", wo.ID, "reworked")
	respond(w, r, http.StatusOK, wo)
}

// POST /api/v1/work-orders/{id}/cancel — 任意一方取消订单（填写原因）
func (s *Server) cancelWorkOrder(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	var req struct {
		Reason string `json:"reason"`
	}
	if err := decode(r, &req); err != nil && !errors.Is(err, io.EOF) {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	wo, err := s.workOrderSvc.RequestCancel(a, r.PathValue("id"), req.Reason)
	if err != nil {
		fail(w, r, http.StatusForbidden, err)
		return
	}
	s.audit(r.Context(), a.ID, "cancel_work_order", "work_order", wo.ID, "cancelled")
	respond(w, r, http.StatusOK, wo)
}
