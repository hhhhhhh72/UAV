package httpapi

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"drone-platform/internal/domain"
	"drone-platform/internal/repository"
	"drone-platform/internal/service"
)

// ---- Enrollments ----

// POST /api/v1/training-courses/{id}/enroll
func (s *Server) enrollCourse(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	// 未公开课程（待审核/草稿/已下架）不可报名——防绕过列表过滤直接报名
	if !isAdminRequest(r) {
		if c, err := s.trainingSvc.GetCourse(r.Context(), r.PathValue("id")); err == nil && isNonPublicStatus(c.Status) {
			fail(w, r, http.StatusNotFound, errors.New("course not found"))
			return
		}
	}
	var form service.EnrollmentForm
	if err := decode(r, &form); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	e, err := s.enrollSvc.Enroll(r.Context(), a.ID, r.PathValue("id"), form)
	if err != nil {
		fail(w, r, http.StatusConflict, err)
		return
	}
	respond(w, r, http.StatusCreated, e)
}

// POST /api/v1/training-courses/{id}/pay-and-enroll
// Freezes course fee from escrow balance, then enrolls student.
func (s *Server) payAndEnroll(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}

	course, err := s.trainingSvc.GetCourse(r.Context(), r.PathValue("id"))
	if err != nil {
		fail(w, r, http.StatusNotFound, errors.New("course not found"))
		return
	}
	// 未公开课程（待审核/草稿/已下架）不可付费报名——与 enrollCourse 一致
	if isNonPublicStatus(course.Status) {
		fail(w, r, http.StatusNotFound, errors.New("course not found"))
		return
	}

	var form service.EnrollmentForm
	if err := decode(r, &form); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	// 报名记录固化冻结金额：completeEnrollment 按此释放，与课程后续改价解耦
	form.PaidAmountFen = course.PriceFen

	if course.PriceFen > 0 {
		_, err := s.escrowSvc.Freeze(r.Context(), a.ID, course.PriceFen, "training_course", course.ID)
		if err != nil {
			fail(w, r, http.StatusPaymentRequired, fmt.Errorf("insufficient balance: %w", err))
			return
		}
	}

	e, err := s.enrollSvc.Enroll(r.Context(), a.ID, course.ID, form)
	if err != nil {
		// 报名失败必须回滚冻结（重复报名/参数错误等），否则用户资金滞留 frozen
		// 且无法自助解冻——托管金写接口仅管理员可用。
		if course.PriceFen > 0 {
			// 资金逻辑：回滚退款失败不得静默——无日志则资金滞留无人知晓；
			// RefundOrphanFreezes 补偿任务会兜底，但此处必须留痕。
			if _, rerr := s.escrowSvc.Refund(r.Context(), a.ID, course.PriceFen, "training_course", course.ID); rerr != nil {
				slog.Error("refund rollback after enroll failure failed",
					"user_id", a.ID, "course_id", course.ID, "amount_fen", course.PriceFen, "error", rerr,
					"enroll_error", err)
			}
		}
		fail(w, r, http.StatusConflict, err)
		return
	}
	s.audit(r.Context(), a.ID, "pay_and_enroll", "enrollment", e.ID, "enrolled")
	respond(w, r, http.StatusCreated, e)
}

// POST /api/v1/enrollments/{id}/complete
// Admin marks enrollment complete: releases escrow to course org + issues certificate.
func (s *Server) completeEnrollment(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		fail(w, r, http.StatusForbidden, errors.New("admin permission required"))
		return
	}

	// 按报名 ID 定位（原实现误用 ListByCourse 按课程 ID 匹配报名 ID，恒 404 → 学费释放/发证闭环瘫痪）
	enrollment, err := s.enrollSvc.FindByID(r.Context(), r.PathValue("id"))
	if err != nil {
		fail(w, r, http.StatusNotFound, errors.New("enrollment not found"))
		return
	}

	// Find course for org/type context（价格不再用于资金操作，仅取机构与证书类型）
	course, err := s.trainingSvc.GetCourse(r.Context(), enrollment.CourseID)
	if err != nil {
		fail(w, r, http.StatusNotFound, fmt.Errorf("course not found: %w", err))
		return
	}

	// completed 幂等重试（回归修复）：此前先 CAS 置 completed 再 Release/AddCertificate，
	// 释放/发证失败后状态已 completed，重试被下方 409 "already completed" 挡死——
	// 学费滞留 frozen、证书缺失且无法补齐。现允许 completed 报名补齐缺失步骤：
	// 资金（release 流水缺失则按 PaidAmountFen 释放）与证书（cert_number='auto-'+ID 缺失则补发）。
	// 两者都完成 → 幂等返回 200（不重复操作）。
	if enrollment.Status == "completed" {
		// 查证：cert_number='auto-'+enrollment.ID 存在即已发证（err==nil 即存在）
		cert, certErr := s.trainingSvc.FindByNumber(r.Context(), "auto-"+enrollment.ID)
		certFound := certErr == nil
		// 免费报名（PaidAmountFen=0）无需释放，视为已完成该步骤。
		releaseDone := enrollment.PaidAmountFen <= 0
		if !releaseDone {
			hasReleased, herr := s.escrowSvc.HasReleased(r.Context(), enrollment.UserID, "training_course", course.ID)
			if herr != nil {
				fail(w, r, http.StatusInternalServerError, fmt.Errorf("check escrow release: %w", herr))
				return
			}
			releaseDone = hasReleased
		}
		if !releaseDone {
			if _, rerr := s.escrowSvc.Release(r.Context(), enrollment.UserID, course.OrgID, enrollment.PaidAmountFen, "training_course", course.ID); rerr != nil {
				s.audit(r.Context(), a.ID, "complete_enrollment_release_failed", "enrollment", enrollment.ID, rerr.Error())
				fail(w, r, http.StatusInternalServerError, fmt.Errorf("release escrow: %w", rerr))
				return
			}
		}
		if !certFound {
			cert, err = s.trainingSvc.AddCertificate(r.Context(),
				domain.Actor{ID: enrollment.UserID, Role: domain.RoleIndividual},
				course.CertType, "auto-"+enrollment.ID, "passed", course.OrgID, "",
				time.Now(), time.Now().AddDate(3, 0, 0),
			)
			if err != nil {
				s.audit(r.Context(), a.ID, "complete_enrollment_cert_failed", "enrollment", enrollment.ID, err.Error())
				fail(w, r, http.StatusInternalServerError, fmt.Errorf("issue certificate: %w", err))
				return
			}
			// 完成即认可：系统签发的证书直接批准（否则长期 pending 不作为有效资质）
			if ap, aerr := s.trainingSvc.ApproveCertificate(r.Context(), a, cert.ID); aerr == nil {
				cert = ap
			}
		} else if cert.Status != "approved" {
			// 证书已存在但未获批（初次审批失败被静默忽略的历史半态）：补审批，
			// 否则证书永久 pending 不构成有效资质。
			if ap, aerr := s.trainingSvc.ApproveCertificate(r.Context(), a, cert.ID); aerr == nil {
				cert = ap
			}
		}
		s.audit(r.Context(), a.ID, "complete_enrollment", "enrollment", enrollment.ID, "completed(retry)")
		respond(w, r, http.StatusOK, map[string]any{
			"enrollment":  enrollment,
			"certificate": cert,
			"status":      "completed",
		})
		return
	}

	// 状态校验：仅 enrolled/paid 可完成；pending/rejected 不可完成
	if enrollment.Status != "enrolled" && enrollment.Status != "paid" {
		fail(w, r, http.StatusConflict, fmt.Errorf("enrollment status %q cannot be completed", enrollment.Status))
		return
	}

	// ① 原子置 completed（CAS：仅 enrolled/paid 可改，WHERE status= 谓词）——并发/重试
	//    只会有一方成功；置完成失败（状态已被并发变更）直接 409，资金不会动。
	//    （此前是盲写 Update，两个并发请求都能把状态写成 completed → 双释放学费。）
	ok, casErr := s.enrollSvc.UpdateStatusCas(r.Context(), enrollment.ID, string(enrollment.Status), "completed")
	if casErr != nil {
		fail(w, r, http.StatusInternalServerError, fmt.Errorf("mark enrollment completed: %w", casErr))
		return
	}
	if !ok {
		fail(w, r, http.StatusConflict, errors.New("报名状态已变更，请刷新后重试"))
		return
	}
	enrollment.Status = "completed"

	// ② 按报名时冻结金额释放（与课程实时价格解耦；免费报名 paid_amount_fen=0 不释放）
	released := false
	if enrollment.PaidAmountFen > 0 {
		if _, err := s.escrowSvc.Release(r.Context(), enrollment.UserID, course.OrgID, enrollment.PaidAmountFen, "training_course", course.ID); err != nil {
			// 释放失败：状态已置 completed（幂等锚点，不会重复释放），
			// 资金滞留 frozen —— 记录审计供管理员人工处理（escrow 管理员接口可解）；
			// 重试走上方 completed 幂等补齐分支，不再被 409 挡死。
			s.audit(r.Context(), a.ID, "complete_enrollment_release_failed", "enrollment", enrollment.ID, err.Error())
			fail(w, r, http.StatusInternalServerError, fmt.Errorf("release escrow: %w", err))
			return
		}
		released = true
	}

	// ③ 发证（幂等：同报名已发过则跳过，防重试重复发证）
	cert, err := s.trainingSvc.AddCertificate(r.Context(),
		domain.Actor{ID: enrollment.UserID, Role: domain.RoleIndividual},
		course.CertType, "auto-"+enrollment.ID, "passed", course.OrgID, "",
		time.Now(), time.Now().AddDate(3, 0, 0),
	)
	if err != nil {
		s.audit(r.Context(), a.ID, "complete_enrollment_cert_failed", "enrollment", enrollment.ID, err.Error())
		fail(w, r, http.StatusInternalServerError, fmt.Errorf("issue certificate: %w", err))
		return
	}
	// 完成即认可：系统签发的证书直接批准（否则长期 pending 不作为有效资质）
	if ap, aerr := s.trainingSvc.ApproveCertificate(r.Context(), a, cert.ID); aerr == nil {
		cert = ap
	}
	_ = released

	s.audit(r.Context(), a.ID, "complete_enrollment", "enrollment", enrollment.ID, "completed+cert_issued")
	respond(w, r, http.StatusOK, map[string]any{
		"enrollment":  enrollment,
		"certificate": cert,
		"status":      "completed",
	})
}

// PUT /api/v1/admin/enrollments/{id} — 管理端编辑报名记录
func (s *Server) updateEnrollment(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok || (a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin) {
		fail(w, r, http.StatusForbidden, errors.New("admin permission required"))
		return
	}
	var in struct {
		Name        string `json:"name"`
		Phone       string `json:"phone"`
		IDCard      string `json:"id_card"`
		Gender      string `json:"gender"`
		Birthday    string `json:"birthday"`
		Email       string `json:"email"`
		Education   string `json:"education"`
		Experience  string `json:"experience"`
		PhotoURL    string `json:"photo_url"`
		IDCardImage string `json:"id_card_image"`
		NoCrime     string `json:"no_crime"`
		Status      string `json:"status"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	// 按报名 ID 定位原记录（保留 course_id/user_id/created_at）——替代全表扫描
	found, err := s.enrollSvc.FindByID(r.Context(), r.PathValue("id"))
	if err != nil {
		fail(w, r, http.StatusNotFound, errors.New("enrollment not found"))
		return
	}
	found.Name = in.Name
	found.Phone = in.Phone
	found.IDCard = in.IDCard
	found.Gender = in.Gender
	if in.Birthday == "" {
		found.Birthday = time.Time{} // 清空生日
	} else {
		t, err := time.Parse("2006-01-02", in.Birthday)
		if err != nil {
			fail(w, r, http.StatusBadRequest, errors.New("invalid birthday format"))
			return
		}
		found.Birthday = t
	}
	found.Email = in.Email
	found.Education = in.Education
	found.Experience = in.Experience
	found.PhotoURL = in.PhotoURL
	found.IDCardImage = in.IDCardImage
	found.NoCrime = in.NoCrime
	found.Status = in.Status
	updated, err := s.enrollSvc.Update(r.Context(), a, found)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "not found"):
			fail(w, r, http.StatusNotFound, err)
		case strings.Contains(err.Error(), "permission"):
			fail(w, r, http.StatusForbidden, err)
		default:
			fail(w, r, http.StatusBadRequest, err) // 非法状态 / 防回退
		}
		return
	}
	respond(w, r, http.StatusOK, updated)
}

// GET /api/v1/enrollments/mine
func (s *Server) listMyEnrollments(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}

	// Collect all enrollments for this user across all courses.
	courses, err := s.trainingSvc.ListCourses(r.Context())
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}

	type enrollmentWithCourse struct {
		domain.Enrollment
		CourseTitle string `json:"course_title"`
	}

	// 性能审查（N+1 修复）：一次 ListByUser 拿全部报名（替代按课程 ListByCourse
	// 循环 + 用户过滤），课程名用一次 ListCourses 建 map 批量补齐（替代逐报名
	// GetCourse），两次查询取代 N+1。
	enrolls, err := s.enrollSvc.ListByUser(r.Context(), a.ID)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	courseByID := make(map[string]domain.TrainingCourse, len(courses))
	for _, c := range courses {
		courseByID[c.ID] = c
	}
	out := make([]enrollmentWithCourse, 0, len(enrolls))
	for _, e := range enrolls {
		out = append(out, enrollmentWithCourse{Enrollment: e, CourseTitle: courseByID[e.CourseID].Title})
	}
	respond(w, r, http.StatusOK, out)
}

// GET /api/v1/training-courses/{id}/enrollments — 管理端按课程查报名（含 PII），限管理员
func (s *Server) listEnrollments(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		fail(w, r, http.StatusForbidden, errors.New("admin permission required"))
		return
	}
	enrolls, err := s.enrollSvc.ListByCourse(r.Context(), r.PathValue("id"))
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, enrolls)
}

// ---- Expiry Alerts ----

// GET /api/v1/certificates/expiring?days=30
func (s *Server) listExpiringCerts(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	// P2 修复：该接口返回全平台证书台账（证书编号等），此前任意登录用户可看，限管理员。
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		fail(w, r, http.StatusForbidden, errors.New("admin permission required"))
		return
	}
	days := 30
	if d, err := fmt.Sscanf(r.URL.Query().Get("days"), "%d", &days); d != 1 || err != nil {
		days = 30
	}
	certs, err := s.trainingSvc.ListAllCertificates(r.Context())
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, s.expirySvc.GetExpiringCerts(certs, days))
}

// GET /api/v1/inspections/expiring?days=30
func (s *Server) listExpiringInspections(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	// P2 修复：该接口返回全平台年检台账，此前任意登录用户可看，限管理员。
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		fail(w, r, http.StatusForbidden, errors.New("admin permission required"))
		return
	}
	days := 30
	if d, err := fmt.Sscanf(r.URL.Query().Get("days"), "%d", &days); d != 1 || err != nil {
		days = 30
	}
	inspections, err := s.insuranceSvc.ListAllInspections(r.Context())
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	respond(w, r, http.StatusOK, s.expirySvc.GetExpiringInspections(inspections, days))
}

// ---- Trade Orders ----

// POST /api/v1/trade-orders
func (s *Server) createTradeOrder(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	var in struct {
		ProductID string `json:"product_id"`
	}
	if err := decode(r, &in); err != nil || in.ProductID == "" {
		fail(w, r, http.StatusBadRequest, errors.New("product_id required"))
		return
	}
	// P0 修复：订单金额与卖家一律以服务端商品为准——此前 amount_fen/seller_id
	// 客户端自报，任意买家可 1 分钱下单或把订单挂到任意卖家名下。
	product, err := s.tradingSvc.GetProduct(r.Context(), in.ProductID)
	if err != nil {
		fail(w, r, http.StatusNotFound, errors.New("product not found"))
		return
	}
	if product.Status != "" && product.Status != "listed" {
		fail(w, r, http.StatusConflict, errors.New("product not available"))
		return
	}
	// 防自买自卖：卖家不能购买自己发布的商品
	if product.SellerID != "" && product.SellerID == a.ID {
		fail(w, r, http.StatusConflict, errors.New("cannot buy your own product"))
		return
	}
	// 下单抢占：先原子标记 sold（仅 listed 可改，防一物多卖/超卖），
	// 供给大厅公开列表只展示 listed，售出后自动不再显示。
	if err := s.tradingSvc.MarkProductSold(r.Context(), product.ID); err != nil {
		fail(w, r, http.StatusConflict, errors.New("product not available"))
		return
	}
	o, err := s.tradeSvc.Create(r.Context(), a.ID, product.ID, product.SellerID, product.PriceFen)
	if err != nil {
		// 订单创建失败：回滚商品为 listed，允许继续售卖——
		// 回滚失败导致商品滞留 sold 无法再售，必须留痕（此前静默吞错）。
		if rerr := s.tradingSvc.RestoreProduct(r.Context(), product.ID); rerr != nil {
			slog.Error("restore product after order creation failed failed",
				"product_id", product.ID, "order_error", err, "restore_error", rerr)
		}
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	s.audit(r.Context(), a.ID, "create_trade_order", "trade_order", o.ID, "created")
	respond(w, r, http.StatusCreated, o)
}

// POST /api/v1/trade-orders/{id}/pay — 买家支付（模拟：pending → paid；真实支付接入后由回调替代）
func (s *Server) payTradeOrder(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	o, err := s.tradeSvc.PayOrder(r.Context(), a.ID, r.PathValue("id"))
	if err != nil {
		fail(w, r, http.StatusForbidden, err)
		return
	}
	s.audit(r.Context(), a.ID, "pay_trade_order", "trade_order", o.ID, "paid")
	respond(w, r, http.StatusOK, o)
}

// PATCH /api/v1/trade-orders/{id}/status
func (s *Server) updateTradeOrderStatus(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	var in struct {
		Status string `json:"status"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	o, err := s.tradeSvc.UpdateStatus(r.Context(), r.PathValue("id"), a.ID, in.Status)
	if err != nil {
		fail(w, r, http.StatusForbidden, err)
		return
	}
	respond(w, r, http.StatusOK, o)
}

// POST /api/v1/trade-orders/{id}/aftersale/review — 卖家/管理员审核售后单（同意退款 / 驳回）
func (s *Server) reviewAftersaleBySeller(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	var in struct {
		Action string `json:"action"` // approve=同意退款 / reject=驳回
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	switch in.Action {
	case "approve", "reject":
	default:
		fail(w, r, http.StatusBadRequest, errors.New("action 仅支持 approve / reject"))
		return
	}
	o, err := s.tradeSvc.ReviewAftersaleAsSeller(r.Context(), a, r.PathValue("id"), in.Action == "approve")
	if err != nil {
		fail(w, r, http.StatusForbidden, err)
		return
	}
	s.audit(r.Context(), a.ID, "review_aftersale", "trade_order", o.ID, o.AftersaleStatus)
	respond(w, r, http.StatusOK, o)
}

// POST /api/v1/trade-orders/{id}/aftersale — 买家申请售后（仅买家可调，shipped/completed → aftersale）
func (s *Server) applyAftersale(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	var in struct {
		AftersaleType      string `json:"aftersale_type"`
		AftersaleReason    string `json:"aftersale_reason"`
		AftersaleDesc      string `json:"aftersale_desc"`
		AftersaleAmountFen int64  `json:"aftersale_amount_fen"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	o, err := s.tradeSvc.ApplyAftersale(r.Context(), a.ID, r.PathValue("id"), in.AftersaleType, in.AftersaleReason, in.AftersaleDesc, in.AftersaleAmountFen)
	if err != nil {
		// 权限/状态机/重复申请错误统一 400，不泄露细节
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	s.audit(r.Context(), a.ID, "apply_aftersale", "trade_order", o.ID, "pending")
	respond(w, r, http.StatusOK, o)
}

// GET /api/v1/trade-orders/mine
func (s *Server) listMyTradeOrders(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	orders, err := s.tradeSvc.ListMine(r.Context(), a.ID)
	if err != nil {
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	// 补商品名（product_id → title）；商品已删除/下架时忽略，保持订单可读。
	// 性能审查（N+1 修复）：收集去重 product_id 一次 ListByIDs 批量查询 +
	// map 填充，替代逐订单 GetProduct。
	ids := make([]string, 0, len(orders))
	seen := make(map[string]bool, len(orders))
	for _, o := range orders {
		if o.ProductID == "" || seen[o.ProductID] {
			continue
		}
		seen[o.ProductID] = true
		ids = append(ids, o.ProductID)
	}
	if len(ids) > 0 {
		if prods, err := s.tradingSvc.ListProductsByIDs(r.Context(), ids); err == nil {
			byID := make(map[string]domain.DroneProduct, len(prods))
			for _, p := range prods {
				byID[p.ID] = p
			}
			for i := range orders {
				if p, ok := byID[orders[i].ProductID]; ok {
					orders[i].ProductName = p.Title
				}
			}
		}
	}
	respond(w, r, http.StatusOK, orders)
}

// ---- Admin Dashboard ----

// GET /api/v1/admin/dashboard
func (s *Server) adminDashboard(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	if a.Role != domain.RolePlatformAdmin && a.Role != domain.RoleAssociationAdmin {
		fail(w, r, http.StatusForbidden, errors.New("admin permission required"))
		return
	}

	ent, err := s.enterprises.Pending(r.Context(), a)
	if err != nil {
		slog.Warn("admin dashboard: load pending enterprises", "err", err)
		ent = nil
	}
	entPending := len(ent)

	dem, err := s.demands.ListAll(r.Context(), repository.DemandFilter{})
	if err != nil {
		slog.Warn("admin dashboard: load demands", "err", err)
		dem = nil
	}
	// dem 全量保留：trends/category_dist/status_dist/offline_amount_total 四个聚合依赖它
	//（转 SQL 聚合属后续优化，本轮保持响应值不变）。
	totalDemands := len(dem)

	// posts 全量保留：trends_detail.post 需按创建时间做月度桶（COUNT 无法替代）。
	posts, _, err := s.communitySvc.ListPublishedPosts(r.Context(), 0, 10000)
	if err != nil {
		slog.Warn("admin dashboard: load posts", "err", err)
		posts = nil
	}
	totalPosts := len(posts)

	// pending_reports：社区举报待处理数（Report.status=pending）。
	// 此前误用 ReviewService.ListAll（企业评价审核）且 status="" 全量，字段名与实体不符。
	// 性能审查：只取 total，不物化行。
	_, totalReports, err := s.communitySvc.ListPendingReports(r.Context(), a, 0, 1)
	if err != nil {
		slog.Warn("admin dashboard: count pending reports", "err", err)
	}

	// users 全量保留：trends_detail.user 需按创建时间做月度桶（All 已 LIMIT 200）。
	users, err := s.userRepo.All(r.Context())
	if err != nil {
		slog.Warn("admin dashboard: load users", "err", err)
	}
	totalUsers := len(users)

	// msgs 全量保留：trends_detail.message 需按创建时间做月度桶。
	msgs, _, err := s.msgSvc.ListAll(r.Context(), 0, 10000)
	if err != nil {
		slog.Warn("admin dashboard: load messages", "err", err)
	}
	totalMessages := len(msgs)

	// Trends: monthly counts (last 12 months) — 4 维趋势（需求/帖子/用户/消息）
	trends := buildDemandTrends(dem)
	trendsDetail := map[string][]map[string]any{
		"demand":  trends,
		"post":    buildMonthlyTrends(posts, func(p domain.Post) time.Time { return p.CreatedAt }),
		"user":    buildMonthlyTrends(users, func(u domain.User) time.Time { return u.CreatedAt }),
		"message": buildMonthlyTrends(msgs, func(m domain.Message) time.Time { return m.CreatedAt }),
	}

	// 线下成交金额汇总（联系对接模式撮合价值）
	var offlineAmountTotal int64
	for _, dd := range dem {
		offlineAmountTotal += dd.OfflineAmountFen
	}

	// Category distribution
	categoryDist := buildCategoryDist(dem)

	// Module stats
	modules := map[string]map[string]int{"talent": {}, "events": {}, "industry": {}}
	// Talent
	certs, err := s.trainingSvc.ListAllCertificates(r.Context())
	if err != nil {
		slog.Warn("admin dashboard: load certificates", "err", err)
	}
	modules["talent"]["certificates"] = len(certs)
	cols, err := s.collegeSvc.List(r.Context(), "")
	if err != nil {
		slog.Warn("admin dashboard: load colleges", "err", err)
	}
	modules["talent"]["colleges"] = len(cols)
	// 性能审查：以下计数只取 total（List 的 COUNT 返回值），不再物化行。
	_, jobsTotal, err := s.jobSvc.ListPublishedJobs(r.Context(), 0, 1)
	if err != nil {
		slog.Warn("admin dashboard: count jobs", "err", err)
	}
	modules["talent"]["jobs"] = jobsTotal
	tours, err := s.studyTourRepo.List(r.Context())
	if err != nil {
		slog.Warn("admin dashboard: load study tours", "err", err)
	}
	modules["talent"]["study_tours"] = len(tours)
	courses, err := s.trainingSvc.ListCourses(r.Context())
	if err != nil {
		slog.Warn("admin dashboard: load training courses", "err", err)
	}
	modules["talent"]["training_courses"] = len(courses)

	// Events
	_, competitionsTotal, err := s.competitionSvc.List(r.Context(), 1, 1)
	if err != nil {
		slog.Warn("admin dashboard: count competitions", "err", err)
	}
	modules["events"]["competitions"] = competitionsTotal
	_, evsTotal, err := s.eventSvc.List(r.Context(), 1, 1)
	if err != nil {
		slog.Warn("admin dashboard: count events", "err", err)
	}
	modules["events"]["events"] = evsTotal
	_, exhsTotal, err := s.exhibitionSvc.List(r.Context(), 1, 1)
	if err != nil {
		slog.Warn("admin dashboard: count exhibitions", "err", err)
	}
	modules["events"]["exhibitions"] = exhsTotal
	_, emgResTotal, err := s.emergencySvc.ListResources(r.Context(), "", "", 1, 1)
	if err != nil {
		slog.Warn("admin dashboard: count emergency resources", "err", err)
	}
	modules["events"]["emergency_resources"] = emgResTotal
	_, dispTotal, err := s.emergencySvc.ListDispatches(r.Context(), "", 1, 1)
	if err != nil {
		slog.Warn("admin dashboard: count emergency dispatches", "err", err)
	}
	modules["events"]["emergency_dispatches"] = dispTotal

	// Industry
	_, achsTotal, err := s.achievementSvc.List(r.Context(), "", 1, 1)
	if err != nil {
		slog.Warn("admin dashboard: count achievements", "err", err)
	}
	modules["industry"]["achievements"] = achsTotal
	_, casesTotal, err := s.caseSvc.List(r.Context(), "", 1, 1)
	if err != nil {
		slog.Warn("admin dashboard: count cases", "err", err)
	}
	modules["industry"]["cases"] = casesTotal
	exps, err := s.expertSvc.List(r.Context(), "")
	if err != nil {
		slog.Warn("admin dashboard: load experts", "err", err)
	}
	modules["industry"]["experts"] = len(exps)
	_, rptsTotal, err := s.reportSvc.List(r.Context(), 1, 1)
	if err != nil {
		slog.Warn("admin dashboard: count industry reports", "err", err)
	}
	modules["industry"]["industry_reports"] = rptsTotal
	_, resTotal, err := s.resourceSvc.List(r.Context(), "", 1, 1)
	if err != nil {
		slog.Warn("admin dashboard: count industry resources", "err", err)
	}
	modules["industry"]["industry_resources"] = resTotal
	_, portsTotal, err := s.portfolioSvc.ListPublished(r.Context(), 1, 1)
	if err != nil {
		slog.Warn("admin dashboard: count portfolios", "err", err)
	}
	modules["industry"]["portfolios"] = portsTotal
	_, rdsTotal, err := s.rdService.List(r.Context(), "", 1, 1)
	if err != nil {
		slog.Warn("admin dashboard: count rd challenges", "err", err)
	}
	modules["industry"]["rd_challenges"] = rdsTotal
	_, projsTotal, err := s.researchSvc.List(r.Context(), 1, 1)
	if err != nil {
		slog.Warn("admin dashboard: count research projects", "err", err)
	}
	modules["industry"]["research_projects"] = projsTotal
	sites, err := s.testSiteSvc.List(r.Context(), "")
	if err != nil {
		slog.Warn("admin dashboard: load test sites", "err", err)
	}
	modules["industry"]["test_sites"] = len(sites)

	// Status distribution
	statusDist := buildStatusDist(dem)

	respond(w, r, http.StatusOK, map[string]any{
		"pending_enterprises":  entPending,
		"total_demands":        totalDemands,
		"total_posts":          totalPosts,
		"pending_reports":      totalReports,
		"total_users":          totalUsers,
		"total_messages":       totalMessages,
		"offline_amount_total": offlineAmountTotal,
		"trends":               trends,
		"trends_detail":        trendsDetail,
		"category_dist":        categoryDist,
		"status_dist":          statusDist,
		"modules":              modules,
		"server_time":          time.Now().UTC().Format(time.RFC3339),
	})
}

func buildDemandTrends(dem []domain.Demand) []map[string]any {
	return buildMonthlyTrends(dem, func(d domain.Demand) time.Time { return d.CreatedAt })
}

// buildMonthlyTrends 近 12 个月月度计数（泛型：需求/帖子/用户/消息等任意带 CreatedAt 的实体）
func buildMonthlyTrends[T any](items []T, getTime func(T) time.Time) []map[string]any {
	now := time.Now()
	counts := map[string]int{}
	for i := 0; i < 12; i++ {
		m := now.AddDate(0, -i, 0).Format("2006-01")
		counts[m] = 0
	}
	for _, it := range items {
		m := getTime(it).Format("2006-01")
		if _, ok := counts[m]; ok {
			counts[m]++
		}
	}
	out := make([]map[string]any, 12)
	for i := 0; i < 12; i++ {
		m := now.AddDate(0, -11+i, 0).Format("2006-01")
		out[i] = map[string]any{"date": m, "count": counts[m]}
	}
	return out
}

func buildCategoryDist(dem []domain.Demand) map[string]int {
	bizLabel := map[domain.BizType]string{
		domain.BizCableInspection: "工业巡检",
		domain.BizPlantTransport:  "植保运输",
		domain.BizSprayPesticide:  "农药喷洒",
		domain.BizCleanPaint:      "清洗保洁",
		domain.BizTradeLease:      "租赁服务",
		domain.BizOther:           "其他服务",
	}
	dist := map[string]int{}
	for _, d := range dem {
		label := bizLabel[d.BizType]
		if label == "" {
			label = string(d.BizType)
		}
		if label != "" {
			dist[label]++
		}
	}
	return dist
}

func buildStatusDist(dem []domain.Demand) map[string]int {
	statusLabel := map[string]string{
		"published": "已发布", "pending": "待审核", "completed": "已完成",
		"cancelled": "已取消", "rejected": "已驳回", "draft": "草稿", "processing": "进行中",
	}
	dist := map[string]int{}
	for _, d := range dem {
		st := string(d.Status)
		label := statusLabel[st]
		if label == "" {
			label = st
		}
		dist[label] = dist[label] + 1
	}
	return dist
}
