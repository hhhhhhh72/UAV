package httpapi

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"
)

// notifyTimeout 单条通知的写入上限：DB 卡住时不能让后台 goroutine 无限堆积。
const notifyTimeout = 5 * time.Second

// notify 站内消息统一入口（异步，不阻塞主流程）。
//
// 三个必须守住的点：
//  1. 用 context.Background() 而非 r.Context()——goroutine 生命周期超出请求，
//     handler 返回后 r.Context() 立刻被取消，通知会被静默丢掉；
//  2. 固定超时，避免下游变慢时 goroutine 越积越多；
//  3. recover 兜底，后台 goroutine panic 不能拖垮进程。
func (s *Server) notify(receiverID, title, content, resType, resID string) {
	if s == nil || s.msgSvc == nil || receiverID == "" {
		return
	}
	go func() {
		defer func() { _ = recover() }()
		ctx, cancel := context.WithTimeout(context.Background(), notifyTimeout)
		defer cancel()
		if _, err := s.msgSvc.Send(ctx, "system", receiverID, title, content, resType, resID); err != nil {
			slog.Warn("notify send failed", "receiver", receiverID, "title", title, "resource", resType, "error", err)
		}
	}()
}

// POST /api/v1/messages/{id}/read
func (s *Server) markMessageRead(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	if _, err := s.msgSvc.MarkRead(r.Context(), a.ID, r.PathValue("id")); err != nil {
		fail(w, r, http.StatusNotFound, err)
		return
	}
	respond(w, r, http.StatusOK, map[string]string{"status": "read"})
}

// GET /api/v1/messages
func (s *Server) listMessages(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	unread := r.URL.Query().Get("unread") == "1"
	msgs, err := s.msgSvc.ListForUser(r.Context(), a.ID, unread)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, msgs)
}

// GET /api/v1/messages/unread-count
func (s *Server) unreadCount(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	count, err := s.msgSvc.UnreadCount(r.Context(), a.ID)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, map[string]int{"count": count})
}
