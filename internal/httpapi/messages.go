package httpapi

import (
	"errors"
	"net/http"
)

// POST /api/v1/messages/{id}/read
func (s *Server) markMessageRead(w http.ResponseWriter, r *http.Request) {
	_, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	if _, err := s.msgSvc.MarkRead(r.PathValue("id")); err != nil {
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
	msgs, err := s.msgSvc.ListForUser(a.ID, unread)
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
	count, err := s.msgSvc.UnreadCount(a.ID)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, map[string]int{"count": count})
}
