package httpapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"drone-platform/internal/crypto"
	"drone-platform/internal/domain"
)

// GET /api/v1/demands/{id}
// 信息公告模式：公开详情含完整联系方式（供直接联系）；未公开状态（pending/rejected）
// 仅发布者与管理员可见。
func (s *Server) demandDetail(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	d, err := s.demands.FindByID(r.Context(), id)
	if err != nil {
		fail(w, r, http.StatusNotFound, err)
		return
	}
	if d.Status != domain.DemandPublished && d.Status != domain.DemandCompleted && d.Status != domain.DemandCancelled && d.Status != domain.DemandAssigned {
		a, ok := authenticatedActor(r)
		if !ok || (a.ID != d.PublisherID && a.Role != domain.RolePlatformAdmin && a.Role != domain.RoleAssociationAdmin) {
			fail(w, r, http.StatusNotFound, errors.New("demand not found"))
			return
		}
	}
	// 发布者企业摘要：有 approved 认证企业才返回（无则前端按个人发布展示，杜绝虚假认证声明）
	if ents, err := s.enterprises.ListByOwner(r.Context(), d.PublisherID); err == nil {
		for _, e := range ents {
			if e.Status == domain.EnterpriseApproved {
				d.PublisherEnterprise = &domain.EnterpriseBrief{ID: e.ID, Name: e.Name, Logo: e.Logo}
				break
			}
		}
	}
	// 标记"我的需求"（请求者即发布者）——前端据此禁用自登记入口
	if a, ok := authenticatedActor(r); ok && a.ID == d.PublisherID {
		d.IsMine = true
	}
	// 公开完整联系方式（公告目的，仅已登录用户；未认证只给脱敏号码），隐藏发布者 ID 与坐标
	d.PublisherID = ""
	d.Latitude = 0
	d.Longitude = 0
	if _, ok := authenticatedActor(r); !ok {
		// 未认证：脱敏，防爬取 PII（安全审计 P1）
		d.Contact = crypto.MaskPhone(d.Contact)
	}
	respond(w, r, http.StatusOK, d)
}

// PATCH /api/v1/demands/{id}
func (s *Server) updateDemand(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	var in struct{ Title, Description string }
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	d, err := s.demands.UpdateDraft(r.Context(), a, r.PathValue("id"), in.Title, in.Description)
	if err != nil {
		code := http.StatusForbidden
		if strings.Contains(err.Error(), "not found") {
			code = http.StatusNotFound
		}
		fail(w, r, code, err)
		return
	}
	respond(w, r, http.StatusOK, d)
}

// POST /api/v1/demands/{id}/submit
func (s *Server) submitDemand(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	d, err := s.demands.Submit(r.Context(), a, r.PathValue("id"))
	if err != nil {
		fail(w, r, http.StatusForbidden, err)
		return
	}
	s.audit(r.Context(), a.ID, "submit_demand", "demand", d.ID, "submitted")
	respond(w, r, http.StatusOK, d)
}

// POST /api/v1/admin/demands/{id}/review
func (s *Server) reviewDemand(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	var req struct{ Action, Reason string }
	if err := decode(r, &req); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	d, err := s.demands.Review(r.Context(), a, r.PathValue("id"), req.Action, req.Reason)
	if err != nil {
		code := http.StatusForbidden
		if strings.Contains(err.Error(), "not found") {
			code = http.StatusNotFound
		}
		fail(w, r, code, err)
		return
	}
	if d.Contact != "" {
		d.Contact = crypto.MaskPhone(d.Contact)
	}
	s.audit(r.Context(), a.ID, "review_demand", "demand", d.ID, req.Action)
	respond(w, r, http.StatusOK, d)
	// 审核通知（异步，不影响主流程）；驳回/补充材料附理由
	reason := ""
	if (req.Action == "reject" || req.Action == "supplement") && req.Reason != "" {
		reason = "：" + req.Reason
	}
	// 异步通知使用 context.Background()：goroutine 生命周期超出请求，
	// r.Context() 会在 handler 返回后取消，导致通知发送被中断。
	go s.msgSvc.Send(context.Background(), "system", d.PublisherID, "需求审核结果",
		fmt.Sprintf("您的需求「%s」已被%s%s", d.Title, mapAction(req.Action), reason), "demand", d.ID)
}

func mapAction(a string) string {
	switch a {
	case "approve":
		return "通过"
	case "reject":
		return "驳回"
	case "supplement":
		return "要求补充材料"
	default:
		return a
	}
}

// POST /api/v1/demands/{id}/complete — publisher marks a published demand done.
func (s *Server) completeDemand(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}

	d, err := s.demands.Complete(r.Context(), a, r.PathValue("id"))
	if err != nil {
		code := http.StatusForbidden
		if strings.Contains(err.Error(), "not found") {
			code = http.StatusNotFound
		}
		fail(w, r, code, err)
		return
	}
	s.audit(r.Context(), a.ID, "complete_demand", "demand", d.ID, "completed")
	s.msgSvc.Send(r.Context(), "system", d.PublisherID, "需求已完成",
		fmt.Sprintf("需求「%s」已标记完成", d.Title), "demand", d.ID)
	respond(w, r, http.StatusOK, map[string]any{"status": "completed", "demand": d})
}

// POST /api/v1/demands/{id}/cancel — publisher withdraws a pending/published demand.
func (s *Server) cancelDemand(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}

	d, err := s.demands.Cancel(r.Context(), a, r.PathValue("id"))
	if err != nil {
		code := http.StatusForbidden
		if strings.Contains(err.Error(), "not found") {
			code = http.StatusNotFound
		}
		fail(w, r, code, err)
		return
	}
	s.audit(r.Context(), a.ID, "cancel_demand", "demand", d.ID, "cancelled")
	respond(w, r, http.StatusOK, map[string]any{"status": "cancelled", "demand": d})
}
