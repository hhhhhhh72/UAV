package httpapi

import (
	"errors"
	"net/http"
	"strings"

	"drone-platform/internal/crypto"
	"drone-platform/internal/domain"
	"drone-platform/internal/service"
)

// POST /api/v1/enterprises
func (s *Server) createEnterprise(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	var in service.CreateEnterpriseInput
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	ent, err := s.enterpriseSvc.Create(r.Context(), a, in)
	if err != nil {
		fail(w, r, http.StatusForbidden, err)
		return
	}
	s.audit(r.Context(), a.ID, "create_enterprise", "enterprise", ent.ID, "created")
	respond(w, r, http.StatusCreated, ent)
}

// PATCH /api/v1/enterprises/{id}
func (s *Server) updateEnterprise(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	var in service.CreateEnterpriseInput
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	ent, err := s.enterpriseSvc.Update(r.Context(), a, r.PathValue("id"), in)
	if err != nil {
		fail(w, r, http.StatusForbidden, err)
		return
	}
	s.audit(r.Context(), a.ID, "update_enterprise", "enterprise", ent.ID, "updated")
	respond(w, r, http.StatusOK, ent)
}

// POST /api/v1/enterprises/{id}/submit
func (s *Server) submitEnterprise(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	ent, err := s.enterpriseSvc.Submit(r.Context(), a, r.PathValue("id"))
	if err != nil {
		fail(w, r, http.StatusForbidden, err)
		return
	}
	s.audit(r.Context(), a.ID, "submit_enterprise", "enterprise", ent.ID, "submitted")
	respond(w, r, http.StatusOK, ent)
}

// GET /api/v1/enterprises/public — 公开已认证企业列表（第一版企业入驻价值展示）
func (s *Server) listPublicEnterprises(w http.ResponseWriter, r *http.Request) {
	items, _, err := s.enterpriseSvc.ListByStatus(r.Context(), domain.Actor{Role: domain.RolePlatformAdmin}, "approved", 0, 100)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	// 仅暴露展示字段，脱敏敏感信息（PRD FR-2.3 列表展示：logo/名称/分类/核心能力/认证状态）
	out := make([]map[string]any, 0, len(items))
	for _, e := range items {
		out = append(out, map[string]any{
			"id":                e.ID,
			"name":              e.Name,
			"status":            e.Status,
			"is_member":         e.IsMember,
			"created_at":        e.CreatedAt,
			"logo":              e.Logo,
			"industry_category": e.IndustryCategory,
			"capability_tags":   e.CapabilityTags,
			"description":       e.Description,
		})
	}
	respond(w, r, http.StatusOK, out)
}

// GET /api/v1/enterprises/public/detail?id={id} — 公开企业详情。
// 仅已审核企业可见；未审核/不存在统一 404，不暴露存在性。
// 路径挂 public 前缀：auth.go publicPrefixes 的前缀匹配自动放行匿名访问。
func (s *Server) getPublicEnterprise(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		fail(w, r, http.StatusNotFound, errors.New("enterprise not found"))
		return
	}
	e, err := s.enterpriseSvc.FindByID(r.Context(), id)
	if err != nil || e.Status != domain.EnterpriseApproved {
		fail(w, r, http.StatusNotFound, errors.New("enterprise not found"))
		return
	}
	// 公开展示字段（PRD FR-2.3）：不含电话/信用代码等敏感信息
	respond(w, r, http.StatusOK, map[string]any{
		"id":                e.ID,
		"name":              e.Name,
		"status":            e.Status,
		"is_member":         e.IsMember,
		"created_at":        e.CreatedAt,
		"logo":              e.Logo,
		"cover_image":       e.CoverImage,
		"industry_category": e.IndustryCategory,
		"capability_tags":   e.CapabilityTags,
		"description":       e.Description,
		"address":           e.Address,
		"scale":             e.Scale,
		"business_hours":    e.BusinessHours,
		"founded_at":        e.FoundedAt,
	})
}

// GET /api/v1/admin/enterprises
func (s *Server) listEnterprises(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	status := r.URL.Query().Get("status")
	// 默认待审核队列；'all' 或空串返回全部状态（空串由 repo 层处理为全部）
	if status == "" {
		status = "submitted"
	}
	if status == "all" {
		status = ""
	}
	// 性能审查：repo 支持 status 过滤 → 分页下沉 SQL（COUNT+LIMIT/OFFSET），
	// respondPage 不再二次切片。
	page, pageSize := paginationFromQuery(r)
	items, total, err := s.enterpriseSvc.ListByStatus(r.Context(), a, status, (page-1)*pageSize, pageSize)
	if err != nil {
		fail(w, r, http.StatusForbidden, err)
		return
	}
	// 脱敏敏感字段
	for i := range items {
		if items[i].AccountName != "" {
			items[i].AccountName = crypto.MaskPhone(items[i].AccountName)
		}
	}
	respondPage(w, r, items, total, page, pageSize)
}

// POST /api/v1/admin/enterprises/{id}/review
func (s *Server) reviewEnterprise(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	var req struct {
		Action string `json:"action"` // approve / reject / supplement
		Reason string `json:"reason"`
	}
	if err := decode(r, &req); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	ent, err := s.enterpriseSvc.Review(r.Context(), a, r.PathValue("id"), req.Action, req.Reason)
	if err != nil {
		code := http.StatusForbidden
		if strings.Contains(err.Error(), "not found") {
			code = http.StatusNotFound
		}
		fail(w, r, code, err)
		return
	}
	s.audit(r.Context(), a.ID, "review_enterprise", "enterprise", ent.ID, req.Action)
	respond(w, r, http.StatusOK, ent)
}

// GET /api/v1/enterprises (list my enterprises)
func (s *Server) listMyEnterprises(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	items, err := s.enterpriseSvc.ListMine(r.Context(), a)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	// 脱敏敏感字段
	for i := range items {
		if items[i].AccountName != "" {
			items[i].AccountName = crypto.MaskPhone(items[i].AccountName)
		}
		// LicenseURL 不脱敏（是文件URL，非证件号）
	}
	respond(w, r, http.StatusOK, items)
}

// POST /api/v1/admin/enterprises/batch-review
func (s *Server) batchReviewEnterprises(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	var req struct {
		IDs    []string `json:"ids"`
		Action string   `json:"action"`
		Reason string   `json:"reason"`
	}
	if err := decode(r, &req); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	if len(req.IDs) == 0 {
		fail(w, r, http.StatusBadRequest, errors.New("ids required"))
		return
	}
	if len(req.IDs) > 50 {
		fail(w, r, http.StatusBadRequest, errors.New("max 50 enterprises per batch"))
		return
	}

	results := make([]map[string]string, 0, len(req.IDs))
	for _, id := range req.IDs {
		_, err := s.enterpriseSvc.Review(r.Context(), a, id, req.Action, req.Reason)
		status := "ok"
		if err != nil {
			status = err.Error()
		}
		results = append(results, map[string]string{"id": id, "status": status})
		s.audit(r.Context(), a.ID, "batch_review_enterprise", "enterprise", id, req.Action)
	}
	respond(w, r, http.StatusOK, map[string]any{
		"total":   len(req.IDs),
		"results": results,
	})
}

// GET /api/v1/admin/enterprises/search?q=xxx
func (s *Server) searchEnterprises(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	q := r.URL.Query().Get("q")
	if q == "" {
		fail(w, r, http.StatusBadRequest, errors.New("q is required"))
		return
	}
	items, err := s.enterpriseSvc.Search(r.Context(), a, q)
	if err != nil {
		fail(w, r, http.StatusForbidden, err)
		return
	}
	for i := range items {
		if items[i].AccountName != "" {
			items[i].AccountName = crypto.MaskPhone(items[i].AccountName)
		}
	}
	respond(w, r, http.StatusOK, items)
}
