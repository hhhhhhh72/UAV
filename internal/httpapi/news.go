package httpapi

import (
	"errors"
	"net/http"

	"drone-platform/internal/domain"
)

// POST /api/v1/articles
func (s *Server) createArticle(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		fail(w, r, http.StatusForbidden, errors.New("admin permission required"))
		return
	}
	var in struct {
		Title    string `json:"title"`
		Content  string `json:"content"`
		Category string `json:"category"`
		Source   string `json:"source"`
		Author   string `json:"author"`
		IsPinned bool   `json:"is_pinned"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	art, err := s.newsSvc.Create(r.Context(), in.Title, in.Content, in.Category, in.Source, in.Author, in.IsPinned)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	s.audit(r.Context(), a.ID, "create_article", "article", art.ID, "created")
	respond(w, r, http.StatusCreated, art)
}

// PUT /api/v1/articles/{id} — 编辑资讯（标题/分类/来源/正文）
func (s *Server) updateArticle(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		fail(w, r, http.StatusForbidden, errors.New("admin permission required"))
		return
	}
	var in struct {
		Title    string `json:"title"`
		Content  string `json:"content"`
		Category string `json:"category"`
		Source   string `json:"source"`
		Author   string `json:"author"`
		IsPinned bool   `json:"is_pinned"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	art, err := s.newsSvc.Update(r.Context(), r.PathValue("id"), in.Title, in.Content, in.Category, in.Source, in.Author, in.IsPinned)
	if err != nil {
		fail(w, r, http.StatusNotFound, err)
		return
	}
	s.audit(r.Context(), a.ID, "update_article", "article", art.ID, "updated")
	respond(w, r, http.StatusOK, art)
}

// POST /api/v1/articles/{id}/publish
func (s *Server) publishArticle(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		fail(w, r, http.StatusForbidden, errors.New("admin permission required"))
		return
	}
	art, err := s.newsSvc.Publish(r.Context(), r.PathValue("id"))
	if err != nil {
		fail(w, r, http.StatusNotFound, err)
		return
	}
	respond(w, r, http.StatusOK, art)
}

// DELETE /api/v1/articles/{id} — 删除资讯（草稿/已发布均可；发布后删除即下线）。
func (s *Server) deleteArticle(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		fail(w, r, http.StatusForbidden, errors.New("admin permission required"))
		return
	}
	if err := s.newsSvc.Delete(r.Context(), r.PathValue("id")); err != nil {
		fail(w, r, http.StatusNotFound, err)
		return
	}
	s.audit(r.Context(), a.ID, "delete_article", "article", r.PathValue("id"), "deleted")
	respond(w, r, http.StatusOK, map[string]any{"deleted": true, "id": r.PathValue("id")})
}

// GET /api/v1/articles?category=policy&page=1&page_size=10
// 公开路由（匿名可读）：只返回已发布（published）资讯——草稿不得对公众可见。
// 管理端列表走 GET /api/v1/admin/articles（listAdminArticles，不过滤）。
func (s *Server) listArticles(w http.ResponseWriter, r *http.Request) {
	// status=published 已下沉到 SQL，与分页一起做。此前的写法是「拉全量 2000 行 →
	// 内存筛 published → 再切片」：total 报的是**当前这一批**的条数（前端分页器因此算错），
	// 而且文章超过 2000 条之后，第一页之后的内容永远看不到。
	page, pageSize := paginationFromQuery(r)
	items, total, err := s.newsSvc.ListByCategory(r.Context(), r.URL.Query().Get("category"), "published", page, pageSize)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respondPage(w, r, items, total, page, pageSize)
}

// GET /api/v1/admin/articles — 管理端全量列表（含草稿），与 listArticles 共用分页逻辑。
// status 传空 = 不过滤；分页同样下沉到 SQL（此前是拉 100000 行再由 respond 切片）。
func (s *Server) listAdminArticles(w http.ResponseWriter, r *http.Request) {
	page, pageSize := paginationFromQuery(r)
	items, total, err := s.newsSvc.ListByCategory(r.Context(), r.URL.Query().Get("category"), "", page, pageSize)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respondPage(w, r, items, total, page, pageSize)
}
