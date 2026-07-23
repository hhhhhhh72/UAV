package httpapi

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strings"
	"time"

	"drone-platform/internal/crypto"
	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
)

// GET /api/v1/admin/export — exports demands as CSV (browser-compatible).
func (s *Server) exportDemands(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok || (a.Role != domain.RolePlatformAdmin && a.Role != domain.RoleAssociationAdmin) {
		fail(w, r, http.StatusForbidden, fmt.Errorf("admin permission required"))
		return
	}

	demands, err := s.demands.List(repository.DemandFilter{})
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}

	filename := fmt.Sprintf("demands_export_%s.csv", time.Now().Format("20060102_150405"))
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	// BOM for Excel UTF-8 compatibility
	w.Write([]byte{0xEF, 0xBB, 0xBF})

	writer := csv.NewWriter(w)
	writer.Write([]string{"ID", "标题", "业务类型", "区域", "预算(元)", "状态", "发布者", "创建时间"})

	for _, d := range demands {
		bizLabel := map[string]string{
			"cable_inspection": "电缆巡检", "plant_transport": "植保运输",
			"spray_pesticide": "喷洒农药", "clean_paint": "清洗喷绘",
			"trade_lease": "买卖租赁", "other": "其他",
		}[string(d.BizType)]
		if bizLabel == "" {
			bizLabel = string(d.BizType)
		}
		statusLabel := map[string]string{
			"pending": "待审核", "published": "已发布", "matched": "已匹配",
			"completed": "已完成", "cancelled": "已取消", "rejected": "已驳回",
		}[string(d.Status)]
		if statusLabel == "" {
			statusLabel = string(d.Status)
		}
		budget := fmt.Sprintf("%.2f", float64(d.BudgetFen)/100.0)
		writer.Write([]string{
			d.ID, d.Title, bizLabel, d.District, budget,
			statusLabel, d.PublisherName, d.CreatedAt.Format("2006-01-02 15:04"),
		})
	}
	writer.Flush()
}

// GET /api/v1/admin/enterprises/export — exports enterprises as CSV.
func (s *Server) exportEnterprises(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok || (a.Role != domain.RolePlatformAdmin && a.Role != domain.RoleAssociationAdmin) {
		fail(w, r, http.StatusForbidden, fmt.Errorf("admin permission required"))
		return
	}

	status := r.URL.Query().Get("status")
	if status == "" {
		status = "submitted"
	}
	items, _, err := s.enterpriseSvc.ListByStatus(a, status, 0, 10000)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}

	filename := fmt.Sprintf("enterprises_export_%s.csv", time.Now().Format("20060102_150405"))
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	w.Write([]byte{0xEF, 0xBB, 0xBF})

	writer := csv.NewWriter(w)
	writer.Write([]string{"ID", "企业名称", "开户账号", "状态", "协会成员", "创建时间"})
	statusLabels := map[string]string{"draft": "草稿", "submitted": "待审核", "approved": "已通过", "rejected": "已驳回", "supplement_required": "需补件"}
	for _, e := range items {
		st := statusLabels[string(e.Status)]
		if st == "" {
			st = string(e.Status)
		}
		member := "否"
		if e.IsMember {
			member = "是"
		}
		acct := crypto.MaskPhone(e.AccountName)
		if acct == "" {
			acct = "-"
		}
		writer.Write([]string{
			e.ID, e.Name, acct, st, member,
			e.CreatedAt.Format("2006-01-02 15:04"),
		})
	}
	writer.Flush()
}

// POST /api/v1/admin/demands/batch-approve — batch approve demands.
func (s *Server) batchApproveDemands(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok || (a.Role != domain.RolePlatformAdmin && a.Role != domain.RoleAssociationAdmin) {
		fail(w, r, http.StatusForbidden, fmt.Errorf("admin permission required"))
		return
	}
	var req struct {
		IDs []string `json:"ids"`
	}
	if err := decode(r, &req); err != nil || len(req.IDs) == 0 {
		fail(w, r, http.StatusBadRequest, fmt.Errorf("ids array required"))
		return
	}
	approved, failed := 0, 0
	for _, id := range req.IDs {
		if _, err := s.demands.Approve(a, strings.TrimSpace(id)); err != nil {
			failed++
		} else {
			approved++
		}
	}
	respond(w, r, http.StatusOK, map[string]any{
		"approved": approved, "failed": failed, "total": len(req.IDs),
	})
}
