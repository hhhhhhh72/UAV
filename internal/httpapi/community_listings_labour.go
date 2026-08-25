package httpapi

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"drone-platform/internal/domain"
)

// ---- Posts ----

func (s *Server) createPost(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	var in struct {
		Title, Content string
		Images         []string `json:"images"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	p, err := s.communitySvc.CreatePost(r.Context(), a, in.Title, in.Content, in.Images)
	if err != nil {
		fail(w, r, http.StatusForbidden, err)
		return
	}
	s.audit(r.Context(), a.ID, "create_post", "post", p.ID, "created")
	respond(w, r, http.StatusCreated, p)
}

func (s *Server) publishPost(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	p, err := s.communitySvc.PublishPost(r.Context(), a, r.PathValue("id"))
	if err != nil {
		fail(w, r, http.StatusForbidden, err)
		return
	}
	respond(w, r, http.StatusOK, p)
}

func (s *Server) removePost(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	p, err := s.communitySvc.RemovePost(r.Context(), a, r.PathValue("id"))
	if err != nil {
		fail(w, r, http.StatusForbidden, err)
		return
	}
	respond(w, r, http.StatusOK, p)
}

func (s *Server) listPosts(w http.ResponseWriter, r *http.Request) {
	// 性能审查：分页下沉 SQL（repo COUNT+LIMIT/OFFSET），respondPage 不再二次切片。
	page, pageSize := paginationFromQuery(r)
	items, total, err := s.communitySvc.ListPublishedPosts(r.Context(), (page-1)*pageSize, pageSize)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	// 公开列表脱敏：post.author_id 是 user-<手机号>，匿名可批量抓取，换成稳定哈希假名。
	for i := range items {
		items[i].AuthorID = maskUserID(items[i].AuthorID)
	}
	respondPage(w, r, items, total, page, pageSize)
}

// GET /api/v1/admin/posts — 管理端帖子全量列表（含待审核 pending）。
// 审核闭环：CreatePost 默认 pending，管理端经此接口看到待审帖子，
// 调 POST /api/v1/posts/{id}/publish 上架（PublishPost 已有，仅管理员）。
func (s *Server) listAdminPosts(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		fail(w, r, http.StatusForbidden, errors.New("admin permission required"))
		return
	}
	// 对齐现有 admin 列表风格：全量拉取 + adminListFilter（keyword/status）+ paginatedRespond。
	all, _, err := s.communitySvc.ListAllPosts(r.Context(), a, 0, 100000)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	filtered, total := adminListFilter(all, r.URL.Query().Get("keyword"), r.URL.Query().Get("status"),
		func(p domain.Post) string { return p.Title + p.Content },
		func(p domain.Post) string { return p.Status })
	paginatedRespond(w, r, filtered, total)
}

// ---- Comments ----

func (s *Server) createComment(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	var in struct {
		PostID  string `json:"post_id"`
		Content string `json:"content"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	c, err := s.communitySvc.CreateComment(r.Context(), a, in.PostID, in.Content)
	if err != nil {
		fail(w, r, http.StatusForbidden, err)
		return
	}
	respond(w, r, http.StatusCreated, c)
}

func (s *Server) listComments(w http.ResponseWriter, r *http.Request) {
	postID := r.URL.Query().Get("post_id")
	if postID == "" {
		fail(w, r, http.StatusBadRequest, errors.New("post_id required"))
		return
	}
	items, err := s.communitySvc.ListComments(r.Context(), postID)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	// 公开列表脱敏：comment.author_id 是 user-<手机号>，匿名可批量抓取，换成稳定哈希假名。
	for i := range items {
		items[i].AuthorID = maskUserID(items[i].AuthorID)
	}
	respond(w, r, http.StatusOK, items)
}

// ---- Reports ----

func (s *Server) createReport(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	var in struct {
		ResourceType string `json:"resource_type"`
		ResourceID   string `json:"resource_id"`
		Reason       string `json:"reason"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	rp, err := s.communitySvc.CreateReport(r.Context(), a, in.ResourceType, in.ResourceID, in.Reason)
	if err != nil {
		fail(w, r, http.StatusForbidden, err)
		return
	}
	respond(w, r, http.StatusCreated, rp)
}

func (s *Server) listReports(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	// 性能审查：分页下沉 SQL（repo COUNT+LIMIT/OFFSET），respondPage 不再二次切片。
	page, pageSize := paginationFromQuery(r)
	items, total, err := s.communitySvc.ListPendingReports(r.Context(), a, (page-1)*pageSize, pageSize)
	if err != nil {
		fail(w, r, http.StatusForbidden, err)
		return
	}
	respondPage(w, r, items, total, page, pageSize)
}

// ---- Listings ----

func (s *Server) createListing(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	var in struct {
		Title, Description, Category string
		PriceFen                     int64 `json:"price_fen"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	now := time.Now()
	l := domain.Listing{ID: fmt.Sprintf("listing-%d", now.UnixNano()), SellerID: a.ID, Title: in.Title,
		Description: in.Description, Category: in.Category, PriceFen: in.PriceFen, Status: "listed", Version: 1, CreatedAt: now, UpdatedAt: now}
	l, err := s.listingSvc.Create(r.Context(), l)
	if err != nil {
		fail(w, r, http.StatusForbidden, err)
		return
	}
	respond(w, r, http.StatusCreated, l)
}

func (s *Server) closeListing(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	l, err := s.listingSvc.Close(r.Context(), a, r.PathValue("id"))
	if err != nil {
		fail(w, r, http.StatusForbidden, err)
		return
	}
	respond(w, r, http.StatusOK, l)
}

func (s *Server) listListings(w http.ResponseWriter, r *http.Request) {
	// 性能审查：分页下沉 SQL（repo COUNT+LIMIT/OFFSET），respondPage 不再二次切片。
	page, pageSize := paginationFromQuery(r)
	items, total, err := s.listingSvc.ListListed(r.Context(), (page-1)*pageSize, pageSize)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	// 公开列表脱敏：手机号注册用户 ID 不随响应泄露（与商品/服务/飞手口径一致）
	for i := range items {
		items[i].SellerID = maskUserID(items[i].SellerID)
	}
	respondPage(w, r, items, total, page, pageSize)
}

func (s *Server) favoriteListing(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	if err := s.listingSvc.Favorite(r.Context(), r.PathValue("id"), a.ID); err != nil {
		fail(w, r, http.StatusForbidden, err)
		return
	}
	respond(w, r, http.StatusOK, map[string]string{"status": "favorited"})
}

// ---- Labour Orders ----

func (s *Server) createLabourOrder(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	var in struct {
		Title, Description string
		WorkerCount        int       `json:"worker_count"`
		StartDate          time.Time `json:"start_date"`
		EndDate            time.Time `json:"end_date"`
		BudgetFen          int64     `json:"budget_fen"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	o, err := s.labourSvc.CreateOrder(r.Context(), a, in.Title, in.Description, in.WorkerCount, in.StartDate, in.EndDate, in.BudgetFen)
	if err != nil {
		fail(w, r, http.StatusForbidden, err)
		return
	}
	respond(w, r, http.StatusCreated, o)
}

func (s *Server) listLabourOrders(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	// 性能审查：分页下沉 SQL（管理员 ListAll / 雇主 ListByEmployer+服务层内存分页），
	// respondPage 不再二次切片。
	page, pageSize := paginationFromQuery(r)
	items, total, err := s.labourSvc.ListOrders(r.Context(), a, (page-1)*pageSize, pageSize)
	if err != nil {
		fail(w, r, http.StatusForbidden, err)
		return
	}
	respondPage(w, r, items, total, page, pageSize)
}

func (s *Server) createLabourQuote(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	var in struct {
		OrderID    string `json:"order_id"`
		AmountFen  int64  `json:"amount_fen"`
		Proposal   string `json:"proposal"`
		QuoterName string `json:"quoter_name"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	q, err := s.labourSvc.CreateQuote(r.Context(), a, in.OrderID, in.AmountFen, in.Proposal, in.QuoterName)
	if err != nil {
		fail(w, r, http.StatusForbidden, err)
		return
	}
	respond(w, r, http.StatusCreated, q)
}

func (s *Server) listLabourQuotes(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	orderID := r.URL.Query().Get("order_id")
	if orderID == "" {
		fail(w, r, http.StatusBadRequest, errors.New("order_id required"))
		return
	}
	items, err := s.labourSvc.ListQuotes(r.Context(), a, orderID)
	if err != nil {
		fail(w, r, http.StatusForbidden, err)
		return
	}
	respond(w, r, http.StatusOK, items)
}

// GET /api/v1/labour-orders/{id}/assignments
func (s *Server) listOrderAssignments(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	orderID := r.PathValue("id")
	items, err := s.labourSvc.ListAssignments(r.Context(), a, orderID)
	if err != nil {
		code := http.StatusForbidden
		if strings.Contains(err.Error(), "not found") {
			code = http.StatusNotFound
		}
		fail(w, r, code, err)
		return
	}
	respond(w, r, http.StatusOK, items)
}

// GET /api/v1/assignments/mine
func (s *Server) listMyAssignments(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	items, err := s.labourSvc.ListMyAssignments(r.Context(), a)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, items)
}
