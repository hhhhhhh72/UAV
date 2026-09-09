package httpapi

import (
	"encoding/csv"
	"errors"
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
	// 独立限频：全量无界导出 + 敏感字段，每 IP 60s 内限 5 次（防连环抓取/拖库）。
	if !s.adminOpAllowed(r, "export", 5) {
		fail(w, r, http.StatusTooManyRequests, errors.New("导出过于频繁，请稍后再试"))
		return
	}

	demands, err := s.demands.List(r.Context(), repository.DemandFilter{})
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	s.audit(r.Context(), a.ID, "export_demands", "csv", "demands", "exported")

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
// GET /api/v1/admin/export/{resource} — 通用全量导出（CSV）。
// 覆盖企业/需求/课程/证书/飞手/报名/赛事等模块：一次请求导出全量数据（不受分页
// page_size≤100 限制），统一平台管理员鉴权、独立限频（防拖库）、CSV 公式注入防护与审计。
func (s *Server) exportResource(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok || a.Role != domain.RolePlatformAdmin {
		fail(w, r, http.StatusForbidden, fmt.Errorf("platform admin permission required"))
		return
	}
	// 独立限频：全量无界导出 + 敏感字段，每 IP 60s 内限 5 次。
	if !s.adminOpAllowed(r, "export", 5) {
		fail(w, r, http.StatusTooManyRequests, errors.New("导出过于频繁，请稍后再试"))
		return
	}

	resource := r.PathValue("resource")
	var header []string
	var rows [][]string

	switch resource {
	case "enterprises":
		header = []string{"ID", "企业名称", "开户账号", "状态", "协会成员", "创建时间"}
		items, _, err := s.enterpriseSvc.ListByStatus(r.Context(), a, "", 0, 10000)
		if err != nil {
			fail(w, r, http.StatusInternalServerError, err)
			return
		}
		labels := map[string]string{"draft": "草稿", "submitted": "待审核", "approved": "已通过", "rejected": "已驳回", "supplement_required": "需补件"}
		for _, e := range items {
			st := labels[string(e.Status)]
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
			rows = append(rows, []string{csvCell(e.ID), csvCell(e.Name), csvCell(acct), csvCell(st), csvCell(member), csvCell(e.CreatedAt.Format("2006-01-02 15:04"))})
		}

	case "demands":
		header = []string{"ID", "标题", "业务类型", "区域", "预算下限(元)", "预算上限(元)", "状态", "发布者", "创建时间"}
		// ListAll：管理端全量（含 pending/rejected 等所有状态）；List 只返回 published，
		// 会导致导出遗漏待审/已驳回数据。
		items, err := s.demands.ListAll(r.Context(), repository.DemandFilter{})
		if err != nil {
			fail(w, r, http.StatusInternalServerError, err)
			return
		}
		for _, d := range items {
			rows = append(rows, []string{
				csvCell(d.ID), csvCell(d.Title), csvCell(string(d.BizType)), csvCell(d.District),
				csvCell(fmt.Sprintf("%.2f", float64(d.BudgetMinFen)/100)),
				csvCell(fmt.Sprintf("%.2f", float64(d.BudgetFen)/100)),
				csvCell(string(d.Status)), csvCell(d.PublisherName), csvCell(d.CreatedAt.Format("2006-01-02 15:04")),
			})
		}

	case "training-courses":
		header = []string{"ID", "课程名称", "机构", "证书类型", "价格(元)", "名额", "已报名", "状态", "创建时间"}
		items, err := s.trainingSvc.ListCourses(r.Context())
		if err != nil {
			fail(w, r, http.StatusInternalServerError, err)
			return
		}
		for _, c := range items {
			rows = append(rows, []string{
				csvCell(c.ID), csvCell(c.Title), csvCell(c.OrgName), csvCell(string(c.CertType)),
				csvCell(fmt.Sprintf("%.2f", float64(c.PriceFen)/100)),
				csvCell(fmt.Sprintf("%d", c.MaxStudents)), csvCell(fmt.Sprintf("%d", c.EnrolledCount)),
				csvCell(c.Status), csvCell(c.CreatedAt.Format("2006-01-02 15:04")),
			})
		}

	case "certificates":
		header = []string{"ID", "用户ID", "证书类型", "证书编号", "等级", "状态", "发证日期", "到期日期"}
		items, err := s.trainingSvc.ListAllCertificates(r.Context())
		if err != nil {
			fail(w, r, http.StatusInternalServerError, err)
			return
		}
		for _, c := range items {
			rows = append(rows, []string{
				csvCell(c.ID), csvCell(c.UserID), csvCell(string(c.CertType)), csvCell(c.CertNumber),
				csvCell(c.Level), csvCell(c.Status),
				csvCell(c.IssueDate.Format("2006-01-02")), csvCell(c.ExpireDate.Format("2006-01-02")),
			})
		}

	case "certified-pilots":
		header = []string{"ID", "真实姓名", "所在地区", "飞行小时", "完成工单", "状态", "驳回理由", "创建时间"}
		items, err := s.trainingSvc.ListPilots(r.Context())
		if err != nil {
			fail(w, r, http.StatusInternalServerError, err)
			return
		}
		for _, p := range items {
			rows = append(rows, []string{
				csvCell(p.ID), csvCell(p.RealName), csvCell(p.Region),
				csvCell(fmt.Sprintf("%d", p.FlightHours)), csvCell(fmt.Sprintf("%d", p.CompletedJobs)),
				csvCell(p.Status), csvCell(p.RejectReason), csvCell(p.CreatedAt.Format("2006-01-02 15:04")),
			})
		}

	case "enrollments":
		header = []string{"ID", "课程ID", "学员", "电话", "状态", "缴费(元)", "创建时间"}
		items, _, err := s.enrollSvc.All(r.Context(), 0, 10000)
		if err != nil {
			fail(w, r, http.StatusInternalServerError, err)
			return
		}
		for _, e := range items {
			rows = append(rows, []string{
				csvCell(e.ID), csvCell(e.CourseID), csvCell(e.Name), csvCell(crypto.MaskPhone(e.Phone)),
				csvCell(e.Status), csvCell(fmt.Sprintf("%.2f", float64(e.PaidAmountFen)/100)),
				csvCell(e.CreatedAt.Format("2006-01-02 15:04")),
			})
		}

	case "competitions":
		header = []string{"ID", "赛事名称", "分类", "地点", "报名费(元)", "报名人数", "状态", "开始日期", "结束日期"}
		items, _, err := s.competitionSvc.List(r.Context(), 1, 10000)
		if err != nil {
			fail(w, r, http.StatusInternalServerError, err)
			return
		}
		for _, c := range items {
			rows = append(rows, []string{
				csvCell(c.ID), csvCell(c.Title), csvCell(c.Category), csvCell(c.Location),
				csvCell(fmt.Sprintf("%d", c.Fee)), csvCell(fmt.Sprintf("%d", c.RegCount)),
				csvCell(c.Status),
				csvCell(c.StartDate.Format("2006-01-02")), csvCell(c.EndDate.Format("2006-01-02")),
			})
		}

	default:
		fail(w, r, http.StatusNotFound, fmt.Errorf("resource %s export not supported", resource))
		return
	}

	s.audit(r.Context(), a.ID, "export_resource", "csv", resource, fmt.Sprintf("rows=%d", len(rows)))

	filename := fmt.Sprintf("%s_export_%s.csv", strings.ReplaceAll(resource, "/", "-"), time.Now().Format("20060102_150405"))
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	w.Write([]byte{0xEF, 0xBB, 0xBF}) // Excel UTF-8 BOM

	writer := csv.NewWriter(w)
	writer.Write(header)
	for _, row := range rows {
		writer.Write(row)
	}
	writer.Flush()
}

