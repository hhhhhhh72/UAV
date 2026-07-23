package httpapi

import (
	"errors"
	"net/http"

	"drone-platform/internal/domain"
)

// POST /api/v1/articles
func (s *Server) createArticle(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		fail(w, r, http.StatusForbidden, errors.New("admin permission required")); return
	}
	var in struct {
		Title, Content, Category, Source string
	}
	if err := decode(r, &in); err != nil { fail(w, r, http.StatusBadRequest, err); return }
	art, err := s.newsSvc.Create(in.Title, in.Content, in.Category, in.Source)
	if err != nil { fail(w, r, http.StatusInternalServerError, err); return }
	s.audit(r.Context(), a.ID, "create_article", "article", art.ID, "created")
	respond(w, r, http.StatusCreated, art)
}

// POST /api/v1/articles/{id}/publish
func (s *Server) publishArticle(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		fail(w, r, http.StatusForbidden, errors.New("admin permission required")); return
	}
	art, err := s.newsSvc.Publish(r.PathValue("id"))
	if err != nil { fail(w, r, http.StatusNotFound, err); return }
	respond(w, r, http.StatusOK, art)
}

// GET /api/v1/articles?category=policy&page=1&page_size=10
func (s *Server) listArticles(w http.ResponseWriter, r *http.Request) {
	page, pageSize := paginationFromQuery(r)
	items, total, err := s.newsSvc.ListByCategory(r.URL.Query().Get("category"), page, pageSize)
	if err != nil { fail(w, r, http.StatusInternalServerError, err); return }
	paginatedRespond(w, r, items, total)
}
