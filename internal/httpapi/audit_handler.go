package httpapi

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)


// GET /api/v1/admin/audit-logs — 操作审计查询（仅平台管理员）。
// 支持 actor_id/action/resource_type/start/end 过滤，按时间倒序分页。
func (s *Server) listAuditLogs(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok || a.Role != domain.RolePlatformAdmin {
		fail(w, r, http.StatusForbidden, fmt.Errorf("platform admin permission required"))
		return
	}
	reader, ok := s.auditWriter.(repository.AuditReader)
	if !ok {
		fail(w, r, http.StatusNotImplemented, errors.New("audit query not supported by this storage backend"))
		return
	}
	page, pageSize := paginationFromQuery(r)
	offset := (page - 1) * pageSize

	f := repository.AuditFilter{
		ActorID:      strings.TrimSpace(r.URL.Query().Get("actor_id")),
		Action:       strings.TrimSpace(r.URL.Query().Get("action")),
		ResourceType: strings.TrimSpace(r.URL.Query().Get("resource_type")),
	}
	if v := strings.TrimSpace(r.URL.Query().Get("start")); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			f.Start = t
		}
	}
	if v := strings.TrimSpace(r.URL.Query().Get("end")); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			// end 为闭区间：含当天，故 +24h
			f.End = t.Add(24 * time.Hour)
		}
	}

	items, total, err := reader.ListAudit(r.Context(), f, offset, pageSize)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	paginatedRespond(w, r, items, total)
}
