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

// csvCell 防 CSV 公式注入：单元格以 = + - @ 或制表符开头时前缀单引号，
// 防止恶意内容（如 =HYPERLINK(...)）在 Excel/WPS 中作为公式执行。
func csvCell(s string) string {
	if s == "" {
		return s
	}
	switch s[0] {
	case '=', '+', '-', '@', '\t', '\r':
		return "'" + s
	}
	return s
}

// GET /api/v1/admin/export — exports demands as CSV (browser-compatible).
// 全量数据导出（含联系电话），仅平台管理员可操作。
func (s *Server) exportDemands(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok || a.Role != domain.RolePlatformAdmin {
		fail(w, r, http.StatusForbidden, fmt.Errorf("platform admin permission required"))
		return
	}

	demands, err := s.demands.List(r.Context(), repository.DemandFilter{})
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
			csvCell(d.ID), csvCell(d.Title), csvCell(bizLabel), csvCell(d.District), csvCell(budget),
			csvCell(statusLabel), csvCell(d.PublisherName), csvCell(d.CreatedAt.Format("2006-01-02 15:04")),
		})
	}
	writer.Flush()
}

// GET /api/v1/admin/enterprises/export — exports enterprises as CSV.
// 全量数据导出，仅平台管理员可操作。
func (s *Server) exportEnterprises(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok || a.Role != domain.RolePlatformAdmin {
		fail(w, r, http.StatusForbidden, fmt.Errorf("platform admin permission required"))
		return
	}

	status := r.URL.Query().Get("status")
	if status == "" {
		status = "submitted"
	}
	items, _, err := s.enterpriseSvc.ListByStatus(r.Context(), a, status, 0, 10000)
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
			csvCell(e.ID), csvCell(e.Name), csvCell(acct), csvCell(st), csvCell(member),
			csvCell(e.CreatedAt.Format("2006-01-02 15:04")),
		})
	}
	writer.Flush()
}

// POST /api/v1/admin/demands/batch-approve — batch approve demands.
// 批量审批全平台需求，仅平台管理员可操作。
func (s *Server) batchApproveDemands(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok || a.Role != domain.RolePlatformAdmin {
		fail(w, r, http.StatusForbidden, fmt.Errorf("platform admin permission required"))
		return
	}
	var req struct {
		IDs []string `json:"ids"`
	}
	if err := decode(r, &req); err != nil || len(req.IDs) == 0 {
		fail(w, r, http.StatusBadRequest, fmt.Errorf("ids array required"))
		return
	}
	if len(req.IDs) > 50 {
		fail(w, r, http.StatusBadRequest, fmt.Errorf("ids 数量不能超过 50"))
		return
	}
	approved, failed := 0, 0
	for _, id := range req.IDs {
		if _, err := s.demands.Approve(r.Context(), a, strings.TrimSpace(id)); err != nil {
			failed++
		} else {
			approved++
		}
	}
	s.audit(r.Context(), a.ID, "batch_approve_demands", "demand", "",
		fmt.Sprintf("approved=%d failed=%d total=%d", approved, failed, len(req.IDs)))
	respond(w, r, http.StatusOK, map[string]any{
		"approved": approved, "failed": failed, "total": len(req.IDs),
	})
}
