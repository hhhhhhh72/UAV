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
	// 性能审查：repo 支持 category 过滤，但公开路由的 status=published 过滤在内存
	// （ListByCategory 需同时服务管理端全量）——保持全量上限 2000 + 内存过滤；
	// TODO 下沉：ListByCategory 增加 status 参数后改分页下沉 SQL + respondPage。
	items, _, err := s.newsSvc.ListByCategory(r.Context(), r.URL.Query().Get("category"), 1, 2000)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	filtered := make([]domain.Article, 0, len(items))
	for _, a := range items {
		if a.Status == "published" {
			filtered = append(filtered, a)
		}
	}
	paginatedRespond(w, r, filtered, len(filtered))
}

// GET /api/v1/admin/articles — 管理端全量列表（含草稿），与 listArticles 共用分页逻辑。
func (s *Server) listAdminArticles(w http.ResponseWriter, r *http.Request) {
	items, total, err := s.newsSvc.ListByCategory(r.Context(), r.URL.Query().Get("category"), 1, 100000)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	paginatedRespond(w, r, items, total)
}
